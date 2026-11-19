// Puts the OBK's page online. Run through einrichten.cmd. It needs a free Cloudflare account, nothing else.
// Running it again is fine: the page keeps its numbers, and you get a new key to paste into the till.

import { spawnSync } from 'node:child_process';
import { randomBytes } from 'node:crypto';
import { existsSync, mkdtempSync, readFileSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const MIN_PASSWORD = 8;
const here = dirname(fileURLToPath(import.meta.url));
const wranglerFile = join(here, 'node_modules', 'wrangler', 'bin', 'wrangler.js');

function stop(message) {
  console.error(`\n${message}`);
  process.exit(1);
}

const wrangler = (args, options = {}) => spawnSync(process.execPath, [wranglerFile, ...args], { cwd: here, stdio: 'inherit', ...options });

function askHidden(question) {
  return new Promise((resolve) => {
    process.stdout.write(question);
    let text = '';
    const onData = (chunk) => {
      for (const character of chunk) {
        if (character === '\u0003') process.exit(1); // Strg+C
        if (character === '\r' || character === '\n') {
          process.stdin.off('data', onData);
          process.stdin.setRawMode?.(false);
          process.stdin.pause();
          process.stdout.write('\n');
          resolve(text);
          return;
        }
        text = character === '\b' || character === '\u007f' ? text.slice(0, -1) : text + character;
      }
    };
    process.stdin.setEncoding('utf8');
    process.stdin.setRawMode?.(true);
    process.stdin.resume();
    process.stdin.on('data', onData);
  });
}

async function askPassword() {
  for (;;) {
    const password = await askHidden(`Passwort für die Webseite (mindestens ${MIN_PASSWORD} Zeichen): `);
    if (password.length < MIN_PASSWORD) {
      console.log('Das ist zu kurz.');
      continue;
    }
    if ((await askHidden('Noch einmal: ')) === password) return password;
    console.log('Die beiden stimmen nicht überein.');
  }
}

/** Sends the page to Cloudflare and returns its address. */
function deploy() {
  const folder = mkdtempSync(join(tmpdir(), 'amk-bar-'));
  const report = join(folder, 'report.ndjson');
  try {
    const result = wrangler(['deploy'], { env: { ...process.env, WRANGLER_OUTPUT_FILE_PATH: report } });
    if (result.status !== 0) stop('Das Hochladen hat nicht geklappt. Die Meldung steht oben.');
    if (!existsSync(report)) return null;
    const entries = readFileSync(report, 'utf8').split('\n').filter(Boolean).map((row) => JSON.parse(row));
    return entries.find((entry) => entry.type === 'deploy')?.targets?.[0] ?? null;
  } finally {
    rmSync(folder, { recursive: true, force: true });
  }
}

function setSecret(name, value) {
  const result = wrangler(['secret', 'put', name], { input: value, stdio: ['pipe', 'inherit', 'inherit'] });
  if (result.status !== 0) stop(`${name} ließ sich nicht speichern. Die Meldung steht oben.`);
}

console.log('AMK Bar: Online-Dashboard einrichten\n');

if (!existsSync(wranglerFile)) {
  console.log('Lade die Werkzeuge. Das dauert eine Minute ...');
  const result = spawnSync('npm', ['install', '--no-audit', '--no-fund'], { cwd: here, stdio: 'inherit', shell: true });
  if (result.status !== 0) stop('Das Laden hat nicht geklappt. Hat dieser PC Internet?');
}

const loggedIn = wrangler(['whoami'], { stdio: 'pipe', encoding: 'utf8' });
if (loggedIn.status !== 0 || /not authenticated/i.test(loggedIn.stdout)) {
  console.log('Gleich öffnet sich der Browser. Bitte bei Cloudflare anmelden und den Zugriff erlauben.\n');
  if (wrangler(['login']).status !== 0) stop('Die Anmeldung hat nicht geklappt.');
}

const password = await askPassword();
const key = randomBytes(32).toString('base64url');

console.log('\nLade die Seite hoch ...');
const address = deploy();
setSecret('PASSWORD', password);
setSecret('SYNC_KEY', key);

console.log('\nFertig. Jetzt in der Kasse unter OBK, Einstellungen, Online-Dashboard eintragen:\n');
console.log(`  Adresse:    ${address ?? '(die workers.dev-Adresse, die oben steht)'}`);
console.log(`  Schlüssel:  ${key}\n`);
console.log('Den Schlüssel braucht nur die Kasse. Das OBK meldet sich auf der Webseite mit dem Passwort an.');
