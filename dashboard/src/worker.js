import { DurableObject } from 'cloudflare:workers';

// Two doors. The till comes in with the key and leaves its latest numbers, and the OBK comes in with the password and
// reads them. Nothing that arrives here ever goes back to the till, so this page cannot change anything at the bar.

const COOKIE = 'amk_session';
const SESSION_MS = 30 * 24 * 60 * 60 * 1000;
const MAX_SNAPSHOT = 512 * 1024; // a real one is about 10 KB
const MAX_LOGIN = 1024;

// Four wrong passwords are free, then the wait doubles with each one.
const FREE_TRIES = 4;
const FIRST_WAIT = 60 * 1000;
const LONGEST_WAIT = 60 * 60 * 1000;

const encoder = new TextEncoder();

function reply(body, status = 200, headers = {}) {
  return new Response(JSON.stringify(body), {
    status,
    headers: {
      'Content-Type': 'application/json; charset=utf-8',
      'Cache-Control': 'no-store',
      'X-Content-Type-Options': 'nosniff',
      ...headers,
    },
  });
}

/** Compares without giving away how much of the guess was right. */
async function same(given, expected) {
  const [a, b] = await Promise.all([given, expected].map((text) => crypto.subtle.digest('SHA-256', encoder.encode(text))));
  return crypto.subtle.timingSafeEqual(a, b);
}

// Changing the password or the key ends every session.
const sessionKey = (env) =>
  crypto.subtle.importKey('raw', encoder.encode(`${env.SYNC_KEY}:${env.PASSWORD}`), { name: 'HMAC', hash: 'SHA-256' }, false, ['sign', 'verify']);

const toBase64 = (buffer) => btoa(String.fromCharCode(...new Uint8Array(buffer))).replaceAll('+', '-').replaceAll('/', '_').replaceAll('=', '');
const fromBase64 = (text) => Uint8Array.from(atob(text.replaceAll('-', '+').replaceAll('_', '/')), (c) => c.charCodeAt(0));

async function newSession(env) {
  const body = `v1.${Date.now() + SESSION_MS}`;
  const signature = await crypto.subtle.sign('HMAC', await sessionKey(env), encoder.encode(body));
  return `${body}.${toBase64(signature)}`;
}

async function hasSession(request, env) {
  const cookie = (request.headers.get('Cookie') ?? '').split(/;\s*/).find((part) => part.startsWith(`${COOKIE}=`));
  const [version, expires, signature] = cookie?.slice(COOKIE.length + 1).split('.') ?? [];
  if (version !== 'v1' || !(Number(expires) > Date.now()) || !signature) return false;
  try {
    return await crypto.subtle.verify('HMAC', await sessionKey(env), fromBase64(signature), encoder.encode(`${version}.${expires}`));
  } catch {
    return false;
  }
}

const sessionCookie = (value, maxAge) => `${COOKIE}=${value}; Path=/; HttpOnly; Secure; SameSite=Strict; Max-Age=${maxAge}`;

/** The one place that remembers anything: the last numbers from the till and the wrong passwords. */
export class Bar extends DurableObject {
  async keep(snapshot) {
    await this.ctx.storage.put('latest', { receivedAt: Date.now(), snapshot });
  }

  async latest() {
    return (await this.ctx.storage.get('latest')) ?? { receivedAt: 0, snapshot: null };
  }

  async login(password) {
    const now = Date.now();
    const tries = (await this.ctx.storage.get('tries')) ?? { fails: 0, until: 0 };
    if (tries.until > now) return { wait: Math.ceil((tries.until - now) / 1000) };

    // Written down before the comparison, so tries sent at the same moment cannot slip past it.
    const fails = tries.fails + 1;
    const wait = fails <= FREE_TRIES ? 0 : Math.min(FIRST_WAIT * 2 ** (fails - FREE_TRIES - 1), LONGEST_WAIT);
    await this.ctx.storage.put('tries', { fails, until: now + wait });

    if (await same(password, this.env.PASSWORD)) {
      await this.ctx.storage.delete('tries');
      return { ok: true };
    }
    return { wait: Math.ceil(wait / 1000) };
  }
}

async function push(request, env, bar) {
  const given = request.headers.get('Authorization')?.replace(/^Bearer /, '') ?? '';
  if (!(await same(given, env.SYNC_KEY))) return reply({ error: 'wrong_key' }, 401);

  const text = await request.text();
  if (text.length > MAX_SNAPSHOT) return reply({ error: 'too_big' }, 413);
  let snapshot;
  try {
    snapshot = JSON.parse(text);
  } catch {
    return reply({ error: 'bad_data' }, 400);
  }
  if (typeof snapshot?.at !== 'number') return reply({ error: 'bad_data' }, 400);
  await bar.keep(snapshot);
  return reply({ ok: true });
}

async function login(request, env, bar) {
  if (Number(request.headers.get('Content-Length') ?? 0) > MAX_LOGIN) return reply({ error: 'too_big' }, 413);
  const text = await request.text();
  if (text.length > MAX_LOGIN) return reply({ error: 'too_big' }, 413);
  let password;
  try {
    ({ password } = JSON.parse(text));
  } catch {
    return reply({ error: 'bad_data' }, 400);
  }
  if (typeof password !== 'string') return reply({ error: 'bad_data' }, 400);

  const result = await bar.login(password);
  if (!result.ok) return reply({ error: result.wait ? 'locked' : 'wrong_password', wait: result.wait }, result.wait ? 429 : 401);
  return reply({ ok: true }, 200, { 'Set-Cookie': sessionCookie(await newSession(env), SESSION_MS / 1000) });
}

export default {
  async fetch(request, env) {
    // Without these an empty password would fit, so nothing opens until the setup has set both.
    if (!env.SYNC_KEY || !env.PASSWORD) return reply({ error: 'not_set_up' }, 503);

    const bar = env.BAR.get(env.BAR.idFromName('bar'));
    const { pathname } = new URL(request.url);
    switch (`${request.method} ${pathname}`) {
      case 'POST /api/push':
        return push(request, env, bar);
      case 'POST /api/login':
        return login(request, env, bar);
      case 'POST /api/logout':
        return reply({ ok: true }, 200, { 'Set-Cookie': sessionCookie('', 0) });
      case 'GET /api/data':
        if (!(await hasSession(request, env))) return reply({ error: 'unauthorized' }, 401);
        return reply(await bar.latest());
      default:
        return reply({ error: 'not_found' }, 404);
    }
  },
};
