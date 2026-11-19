// The flyer's world: the room, the crate stacks, the helicopter and the bits that fly about. It knows nothing of
// Svelte; Game.svelte feeds it time and taps and hears back through `hooks`.
//
// The room is 100 units from ceiling (0) to floor (100). The camera always shows all of it, plus a band above and
// below, and at least MIN_W units across, so an upright screen sees about one and a half stacks ahead and a lying
// one sees more. Physics runs in fixed steps of 1/120 s and drawing blends the last two steps, so 60 Hz and 144 Hz
// play the same.

import { calmNow } from '../../lib/moments.svelte.js';
import { paint, CW, CH, NECK, CRATE_COLORS, FAR_PERIOD, MID_PERIOD, CEIL_PERIOD, FLOOR_PERIOD, FG_PERIOD, LAMPS, LAMP_Y, BULBS, FLOOR_TOP } from './art.js';
import { HUB, ROTOR_HALF, TAIL, TAIL_R, LAMP, BEACON, BODY_BOX } from './heli.js';

const P = 100;
const MIN_W = 74;
const BAND = 9;
const STEP = 1 / 120;
const G = 230; // units/s², down
const FLAP = -68; // units/s, the climb a tap gives
const VMAX = 92;
const HOVER = 46;
const IDLE_SPEED = 9;
const FIRST = 56; // how far ahead the first stack stands
const SKID = 3.1; // from the cabin's middle down to the skids

// The parts of the helicopter that count as a hit: x, y, radius in world units. Smaller than the picture on purpose:
// rotor, tail and skids may brush a crate.
const HIT = [
  [2.5, 0.6, 1.75],
  [0, 0, 2.05],
  [-2.6, -0.4, 1.55],
];

const FAR = 0.12;
const MID = 0.38;
const DUST = 0.55;
const FORE = 1.5;

const clamp = (v, a, b) => (v < a ? a : v > b ? b : v);
const rand = (a, b) => a + Math.random() * (b - a);
const ease = (dt, rate) => 1 - Math.exp(-dt * rate);
const easeBack = (t) => 1 + 2.4 * (t - 1) ** 3 + 1.4 * (t - 1) ** 2;
const easeIn = (t) => t * t * t;

// Difficulty, by the number of stacks: a roomy start that slowly tightens.
const gapAt = (k) => (k < 40 ? 34 - (9 * k) / 40 : Math.max(22.5, 25 - (k - 40) * 0.05));
const spacingAt = (k) => Math.max(38, 46 - k * 0.16);
const speedAt = (k) => (k < 50 ? 32 + (12 * k) / 50 : Math.min(50, 44 + (k - 50) * 0.08));
const fallIn = (t) => (t <= 0.4 ? 0.5 * G * t * t : 18.4 + VMAX * (t - 0.4));

function crates(first) {
  const list = [];
  for (let i = 0; i < 18; i++) {
    list.push({
      c: (Math.random() * CRATE_COLORS.length) | 0,
      l: Math.random() < 0.4,
      dx: i === 0 && first ? rand(-0.25, 0.25) : rand(-0.45, 0.45),
      rot: i === 0 && first ? 0 : rand(-0.018, 0.018),
    });
  }
  return list;
}

function freshHeli(entering) {
  return {
    x: 0, y: HOVER, vx: 0, vy: 0, angle: 0, av: 0, sq: 0, sqv: 0,
    rotor: 0, tail: 0, spin: 26, alpha: 1, beam: 1,
    enter: entering ? 0 : 1, down: false, bounces: 0,
  };
}

/**
 * @param {HTMLCanvasElement} canvas
 * @param {{ score: (n: number) => void, crash: (n: number) => void, over: () => void, ready: () => void }} hooks
 */
export function createWorld(canvas, hooks) {
  const g = canvas.getContext('2d', { alpha: false });
  let S = 10;
  let Wpx = 0;
  let Hpx = 0;
  let vw = MIN_W;
  let vh = P + 2 * BAND;
  let camTop = -BAND;
  let flyX = 20;
  let art = null;

  let mode = 'intro';
  let modeT = 0;
  let time = 0;
  let acc = 0;
  let scroll = 0;
  let prevScroll = 0;
  let speed = IDLE_SPEED;
  let heli = freshHeli(true);
  heli.x = -60;
  const prev = { x: heli.x, y: heli.y, a: 0 };
  let stacks = [];
  const parts = [];
  let spawned = 0;
  let lastWx = 0;
  let lastCenter = 50;
  let gaps = 0;
  let shake = 0;
  let smokeT = 0;
  let still = 0;

  const motes = [];
  for (let i = 0; i < 34; i++) {
    motes.push({ x: rand(0, 400), y: rand(6, 94), s: rand(0.16, 0.34), a: rand(0.08, 0.3), f: rand(0.3, 0.9), p: rand(0, 6) });
  }

  // ------------------------------------------------------------------------------------------------------------
  // Size

  /** Fits the canvas to its box. Answers where the room's ceiling and floor lie, in CSS pixels from the top and bottom. */
  function resize(cssW, cssH, dpr) {
    const w = Math.max(1, Math.round(cssW * dpr));
    const h = Math.max(1, Math.round(cssH * dpr));
    if (w !== Wpx || h !== Hpx || !art) refit(w, h);
    return { ceiling: (-camTop * S * cssH) / h, floor: ((camTop + vh - P) * S * cssH) / h };
  }

  function refit(w, h) {
    Wpx = w;
    Hpx = h;
    canvas.width = w;
    canvas.height = h;
    S = Math.min(h / (P + 2 * BAND), w / MIN_W);
    vw = w / S;
    vh = h / S;
    camTop = -(vh - P) / 2;
    flyX = clamp(vw * 0.28, 19, 58);
    art = paint(S, camTop, vh);
    still = 0;
    draw(1);
  }

  function repaint() {
    if (!Wpx) return;
    art = paint(S, camTop, vh);
    still = 0;
  }

  // ------------------------------------------------------------------------------------------------------------
  // Particles

  function emit(kind, x, y, vx, vy, life, size, more) {
    if (parts.length > 280) return;
    const p = { kind, x, y, vx, vy, life, max: life, size, rot: 0, vr: 0, grav: 0, drag: 0, grow: 0, color: '' };
    if (more) Object.assign(p, more);
    parts.push(p);
  }

  function updateParts(dt) {
    for (let i = parts.length - 1; i >= 0; i--) {
      const p = parts[i];
      p.life -= dt;
      if (p.life <= 0) {
        parts[i] = parts[parts.length - 1];
        parts.pop();
        continue;
      }
      p.vy += p.grav * dt;
      if (p.drag) {
        const d = Math.exp(-p.drag * dt);
        p.vx *= d;
        p.vy *= d;
      }
      p.x += p.vx * dt;
      p.y += p.vy * dt;
      p.rot += p.vr * dt;
      p.size += p.grow * dt;
      if ((p.kind === 'splinter' || p.kind === 'banana') && p.y > P + 1.2 && p.vy > 0) {
        p.y = P + 1.2;
        p.vy *= -0.32;
        p.vx *= 0.55;
        p.vr *= 0.5;
      }
    }
  }

  function sparks(x, y, dir, n) {
    for (let i = 0; i < n; i++) {
      emit('spark', x + rand(-CW / 2, CW / 2), y, rand(-12, 6), dir * rand(4, 15), rand(0.3, 0.55), rand(0.9, 1.7), { drag: 2.5 });
    }
  }

  // ------------------------------------------------------------------------------------------------------------
  // Stacks

  function spawn() {
    const k = spawned++;
    const gap = gapAt(k);
    const spacing = spacingAt(k);
    const t = (spacing - CW * 0.5) / speedAt(k);
    const variety = k < 25 ? 0.55 + (0.45 * k) / 25 : 1;
    const up = Math.min(30, 0.55 * 33 * t) * variety;
    const down = Math.min(34, 0.6 * fallIn(t)) * variety;
    const lo = gap / 2 + 9;
    const hi = P - gap / 2 - 9;
    const goal = rand(lo, hi);
    const center = clamp(lastCenter + clamp(goal - lastCenter, -up, down), lo, hi);
    const wx = lastWx + spacing;
    lastWx = wx;
    lastCenter = center;
    const s = {
      wx,
      top: center - gap / 2,
      bottom: center + gap / 2,
      passed: false,
      appear: 1,
      delay: 0,
      target: 1,
      floor: crates(true),
      ceil: crates(true),
      bottles: (Math.random() * 3) | 0,
    };
    stacks.push(s);
    return s;
  }

  function circleRect(cx, cy, r, x0, y0, x1, y1) {
    const nx = cx < x0 ? x0 : cx > x1 ? x1 : cx;
    const ny = cy < y0 ? y0 : cy > y1 ? y1 : cy;
    const dx = cx - nx;
    const dy = cy - ny;
    return dx * dx + dy * dy < r * r ? { x: nx, y: ny } : null;
  }

  function collide(hx, hy) {
    const c = Math.cos(heli.angle);
    const s = Math.sin(heli.angle);
    for (const [lx, ly, r] of HIT) {
      const px = hx + lx * c - ly * s;
      const py = hy + lx * s + ly * c;
      if (py - r < 0.2) return { x: px, y: 0, what: 'ceiling' };
      if (py + r > P - 0.6) return { x: px, y: P, what: 'floor' };
      for (const st of stacks) {
        if (Math.abs(st.wx - px) > CW / 2 + r) continue;
        const x0 = st.wx - CW / 2 + 0.35;
        const x1 = st.wx + CW / 2 - 0.35;
        let at = circleRect(px, py, r, x0, -1e3, x1, st.top - 0.5);
        if (at) return { ...at, what: 'crate', stack: st, side: at.x === x0, upper: true };
        at = circleRect(px, py, r, x0, st.bottom + 1.4, x1, 1e3);
        if (at) return { ...at, what: 'crate', stack: st, side: at.x === x0, upper: false };
      }
    }
    return null;
  }

  function groundAt(x, y) {
    let ground = P + 1.5;
    for (const s of stacks) {
      if (Math.abs(s.wx - x) < CW / 2 + 1.5 && s.bottom + 0.3 >= y + SKID - 0.8) ground = Math.min(ground, s.bottom + 0.3);
    }
    return ground;
  }

  // ------------------------------------------------------------------------------------------------------------
  // Steps

  function aim(dt, target) {
    heli.angle += (clamp(target, -0.38, 0.55) - heli.angle) * ease(dt, 10);
  }

  function hover(dt) {
    speed += (IDLE_SPEED - speed) * ease(dt, 2);
    scroll += speed * dt;
    heli.enter = Math.min(1, heli.enter + dt / 0.95);
    const e = 1 - (1 - heli.enter) ** 3;
    heli.x = -(1 - e) * (flyX + 20);
    const target = HOVER + Math.sin(time * 2.1) * 1.8;
    const y = heli.y + (target - heli.y) * ease(dt, 5);
    heli.vy = (y - heli.y) / dt;
    heli.y = y;
    aim(dt, heli.vy * 0.012 + (1 - e) * 0.3);
    heli.spin += (26 - heli.spin) * ease(dt, 3);
  }

  function fly(dt) {
    speed += (speedAt(gaps) - speed) * ease(dt, 2.4);
    heli.x += (0 - heli.x) * ease(dt, 6);
    heli.vy = Math.min(heli.vy + G * dt, VMAX);
    heli.y += heli.vy * dt;
    scroll += speed * dt;
    aim(dt, 0.06 + heli.vy * 0.0055);
    heli.spin += (30 - heli.spin) * ease(dt, 3);

    while (lastWx - scroll < vw - flyX + CW) spawn();
    const hx = scroll + heli.x;
    const hy = heli.y;
    for (const s of stacks) {
      if (!s.passed && s.wx <= hx) {
        s.passed = true;
        gaps++;
        sparks(s.wx, s.top, -1, 6);
        sparks(s.wx, s.bottom, 1, 6);
        if (gaps % 10 === 0) {
          for (let i = 0; i < 18; i++) {
            const a = (i / 18) * Math.PI * 2;
            emit('spark', hx, hy, Math.cos(a) * 26, Math.sin(a) * 26, 0.6, 1.8, { drag: 3.5 });
          }
        }
        hooks.score(gaps);
      }
    }
    const hit = collide(hx, hy);
    if (hit) crash(hit);
    while (stacks.length && stacks[0].wx - scroll < -flyX - CW * 2) stacks.shift();
  }

  function crash(hit) {
    mode = 'crash';
    modeT = 0;
    if (hit.what === 'floor') {
      heli.vy = -26;
      heli.vx = -4;
    } else if (hit.what === 'ceiling') {
      heli.vy = 12;
      heli.vx = -4;
    } else if (hit.side) {
      heli.vy = Math.min(heli.vy, 0) * 0.3 - 10;
      heli.vx = -16;
    } else {
      heli.vy = hit.upper ? 14 : -22;
      heli.vx = -6;
    }
    heli.av = (heli.vy < 0 ? -1 : 1) * rand(4.5, 7);
    heli.sq = -0.16;
    heli.sqv = 0;
    const colors = hit.stack
      ? [CRATE_COLORS[hit.stack.floor[0].c], CRATE_COLORS[hit.stack.ceil[0].c], CRATE_COLORS[hit.stack.floor[1].c]]
      : ['#5a3d26', '#3a2717', '#2a1c12'];
    for (let i = 0; i < 18; i++) {
      emit('splinter', hit.x, hit.y, rand(-26, 10), rand(-30, 8), rand(0.9, 1.5), rand(0.5, 1.2), {
        color: colors[i % colors.length], rot: rand(0, 6), vr: rand(-14, 14), grav: 120, drag: 0.6,
      });
    }
    for (let i = 0; i < 14; i++) {
      emit('bubble', hit.x, hit.y, rand(-16, 10), rand(-24, -2), rand(0.6, 1.1), rand(0.35, 0.95), { grav: 28, drag: 1.4 });
    }
    for (let i = 0; i < 12; i++) {
      emit('spark', hit.x, hit.y, rand(-30, 18), rand(-30, 20), rand(0.25, 0.5), rand(1, 2), { drag: 3 });
    }
    if (!calmNow()) shake = 1;
    hooks.crash(gaps);
  }

  function fall(dt) {
    speed *= Math.exp(-dt * 5);
    scroll += speed * dt;
    heli.vx *= Math.exp(-dt * 2.5);
    heli.x += heli.vx * dt;
    heli.spin *= Math.exp(-dt * 0.9);
    heli.beam = Math.max(0, heli.beam - dt * 1.4);
    if (!heli.down) {
      heli.vy = Math.min(heli.vy + G * dt, VMAX * 1.3);
      heli.y += heli.vy * dt;
      heli.angle += heli.av * dt;
      const ground = groundAt(scroll + heli.x, heli.y) - SKID;
      if (heli.y >= ground && heli.vy > 0) {
        heli.y = ground;
        if (heli.vy > 28 && heli.bounces < 2) {
          heli.vy *= -0.3;
          heli.av *= 0.4;
          heli.bounces++;
          for (let i = 0; i < 8; i++) {
            emit('dust', scroll + heli.x + rand(-4, 4), ground + SKID, rand(-10, 10), rand(-6, -1), rand(0.5, 0.8), rand(1, 1.6), { grow: 2.5, drag: 3 });
          }
          if (!calmNow()) shake = Math.max(shake, 0.4);
        } else {
          heli.vy = 0;
          heli.down = true;
          let a = heli.angle % (Math.PI * 2);
          if (a > Math.PI) a -= Math.PI * 2;
          if (a < -Math.PI) a += Math.PI * 2;
          heli.angle = a;
        }
      }
    } else {
      heli.angle += (clamp(heli.angle, -0.3, 0.3) - heli.angle) * ease(dt, 7);
    }
    smokeT -= dt;
    if (smokeT <= 0 && modeT < 2.6) {
      smokeT = 0.075;
      emit('smoke', scroll + heli.x - 1.5, heli.y - 1.6, rand(-5, 1), rand(-9, -4), rand(1, 1.5), rand(0.9, 1.4), { grow: 2.4, drag: 0.6 });
    }
    if (mode === 'crash' && modeT > 0.95) {
      mode = 'over';
      hooks.over();
    }
  }

  function leave(dt) {
    speed += (IDLE_SPEED - speed) * ease(dt, 2);
    scroll += speed * dt;
    heli.alpha = Math.max(0, 1 - modeT / 0.22);
    if (modeT >= 0.5) {
      stacks = [];
      heli = freshHeli(true);
      heli.x = -(flyX + 20);
      prev.x = heli.x;
      prev.y = heli.y;
      prev.a = 0;
      mode = 'ready';
      modeT = 0;
      hooks.ready();
    }
  }

  function step(dt) {
    prevScroll = scroll;
    prev.x = heli.x;
    prev.y = heli.y;
    prev.a = heli.angle;
    time += dt;
    modeT += dt;
    heli.rotor += heli.spin * dt;
    heli.tail += heli.spin * 1.6 * dt;
    heli.sqv += (-180 * heli.sq - 14 * heli.sqv) * dt;
    heli.sq += heli.sqv * dt;
    if (mode === 'intro' || mode === 'ready') hover(dt);
    else if (mode === 'play') fly(dt);
    else if (mode === 'crash' || mode === 'over') fall(dt);
    else if (mode === 'leave') leave(dt);
    for (const s of stacks) {
      if (s.delay > 0) s.delay -= dt;
      else if (s.target) s.appear = Math.min(1, s.appear + dt / 0.6);
      else s.appear = Math.max(0, s.appear - dt / 0.38);
    }
    updateParts(dt);
    shake *= Math.exp(-dt * 6);
    if (shake < 0.01) shake = 0;
  }

  // ------------------------------------------------------------------------------------------------------------
  // Drawing

  // Repeats a tile across the screen. `units` is how far it has moved, `period` how many units one tile holds.
  function tile(img, units, period, y, shx = 0) {
    if (!img) return;
    const w = img.width;
    let x = -Math.round((units * (w / period)) % w) + Math.round(shx);
    while (x > 0) x -= w;
    for (; x < Wpx; x += w) g.drawImage(img, x, y);
  }

  function draw(alpha) {
    if (!art) return;
    const sc = prevScroll + (scroll - prevScroll) * alpha;
    const hxo = prev.x + (heli.x - prev.x) * alpha;
    const hy = prev.y + (heli.y - prev.y) * alpha;
    const ha = prev.a + (heli.angle - prev.a) * alpha;
    let shx = 0;
    let shy = 0;
    if (shake) {
      shx = (Math.random() * 2 - 1) * shake * S * 0.9;
      shy = (Math.random() * 2 - 1) * shake * S * 0.9;
    }

    g.setTransform(1, 0, 0, 1, 0, 0);
    g.globalCompositeOperation = 'source-over';
    g.globalAlpha = 1;
    tile(art.wall, sc * FAR, FAR_PERIOD, 0);

    // Light from the pendant lamps on the wall, then the shelves and lamps themselves.
    const mo = sc * MID;
    const k0 = Math.floor(mo / MID_PERIOD) - 1;
    const k1 = k0 + Math.ceil(vw / MID_PERIOD) + 2;
    const lampY = (LAMP_Y - camTop) * S;
    const mx = art.middle.width / MID_PERIOD;
    g.globalCompositeOperation = 'lighter';
    for (let k = k0; k <= k1; k++) {
      for (const lx of LAMPS) {
        const X = (lx + k * MID_PERIOD - mo) * mx;
        if (X < -24 * S || X > Wpx + 24 * S) continue;
        g.globalAlpha = 0.5;
        g.drawImage(art.lampCone, X - 22 * S, lampY - 0.5 * S, 44 * S, 72 * S);
        g.globalAlpha = 0.55;
        g.drawImage(art.glow, X - 8 * S, lampY - 7 * S, 16 * S, 14 * S);
      }
    }
    g.globalCompositeOperation = 'source-over';
    g.globalAlpha = 1;
    tile(art.middle, mo, MID_PERIOD, 0);

    g.globalCompositeOperation = 'lighter';
    for (let k = k0; k <= k1; k++) {
      for (const b of BULBS) {
        const X = (b.x + k * MID_PERIOD - mo) * mx;
        if (X < -2 * S || X > Wpx + 2 * S) continue;
        const tw = 0.5 + 0.5 * Math.sin(time * 2.3 + b.phase);
        g.globalAlpha = 0.18 + 0.42 * tw;
        g.drawImage(art.spark, X - 1.7 * S, (b.y + 0.5 - camTop) * S - 1.7 * S, 3.4 * S, 3.4 * S);
      }
    }
    g.globalCompositeOperation = 'source-over';

    // Dust in the light
    g.fillStyle = '#ffe6bf';
    const span = vw + 8;
    for (const m of motes) {
      const X = ((((m.x - sc * DUST) % span) + span) % span) - 4;
      const Y = m.y + Math.sin(time * m.f + m.p) * 1.4;
      g.globalAlpha = m.a * (0.6 + 0.4 * Math.sin(time * m.f * 1.7 + m.p));
      g.fillRect(X * S, (Y - camTop) * S, m.s * S, m.s * S);
    }
    g.globalAlpha = 1;

    // The world: crate stacks
    const ox = (flyX - sc) * S + shx;
    const oy = -camTop * S + shy;
    const bottom = camTop + vh;
    const crateAt = (img, cx, cy, rot) => {
      if (rot) {
        const c = Math.cos(rot) * S;
        const s = Math.sin(rot) * S;
        g.setTransform(c, s, -s, c, ox + cx * S, oy + cy * S);
      } else g.setTransform(S, 0, 0, S, ox + cx * S, oy + cy * S);
      g.drawImage(img, -CW / 2 - 1 / S, -CH / 2 - 1 / S, CW + 2 / S, CH + 2 / S);
    };
    for (const s of stacks) {
      const sx = s.wx - sc + flyX;
      if (sx < -CW * 1.5 || sx > vw + CW * 1.5) continue;
      const e = s.target ? easeBack(s.appear) : easeIn(s.appear);
      const dropF = (1 - e) * (bottom - s.bottom + 4);
      const dropC = (1 - e) * (s.top - camTop + 4);
      const fy = s.bottom + dropF;
      const cy = s.top - dropC;
      g.setTransform(S, 0, 0, S, ox, oy);
      g.drawImage(art.shadow, s.wx + CW / 2 - 0.3, fy + 1, 3.8, bottom - fy);
      g.drawImage(art.shadow, s.wx + CW / 2 - 0.3, camTop, 3.8, cy - camTop);
      const f0 = s.floor[0];
      g.drawImage(art.bottles[s.bottles], s.wx - CW / 2 + f0.dx - 1 / S, fy - 1 / S, CW + 2 / S, NECK + 0.6 + 2 / S);
      let y = fy + NECK;
      for (let i = 0; y < bottom + 0.5 && i < s.floor.length; i++, y += CH) {
        const k = s.floor[i];
        crateAt(art.crates[k.c][k.l ? 1 : 0], s.wx + k.dx, y + CH / 2, k.rot);
      }
      y = cy;
      for (let i = 0; y > camTop - 0.5 && i < s.ceil.length; i++, y -= CH) {
        const k = s.ceil[i];
        crateAt(art.crates[k.c][k.l ? 1 : 0], s.wx + k.dx, y - CH / 2, k.rot);
      }
      // the straps that hold the hanging stack
      g.setTransform(S, 0, 0, S, ox, oy);
      for (const x of [s.wx - CW / 2 + 2.3, s.wx + CW / 2 - 3.2]) {
        g.fillStyle = '#b8771a';
        g.fillRect(x, camTop, 0.9, cy - camTop - 0.25);
        g.fillStyle = 'rgba(255,224,160,0.4)';
        g.fillRect(x, camTop, 0.16, cy - camTop - 0.25);
        g.fillStyle = 'rgba(0,0,0,0.3)';
        g.fillRect(x + 0.74, camTop, 0.16, cy - camTop - 0.25);
        g.fillStyle = '#8f949a';
        g.fillRect(x - 0.22, cy - 2.6, 1.34, 1.15);
        g.fillStyle = 'rgba(255,255,255,0.45)';
        g.fillRect(x - 0.22, cy - 2.6, 1.34, 0.22);
      }
    }

    // Ceiling and floor pass in front of the stacks.
    g.setTransform(1, 0, 0, 1, 0, 0);
    tile(art.ceiling, sc, CEIL_PERIOD, 0, shx);
    tile(art.floor, sc, FLOOR_PERIOD, Math.round((FLOOR_TOP - camTop) * S), shx);

    drawHeli(ox + (sc + hxo) * S, oy + hy * S, ha);

    // Particles
    g.globalCompositeOperation = 'source-over';
    for (const p of parts) {
      const t = p.life / p.max;
      if (p.kind === 'spark') continue;
      if (p.kind === 'splinter' || p.kind === 'banana') {
        const c = Math.cos(p.rot) * S;
        const s = Math.sin(p.rot) * S;
        g.setTransform(c, s, -s, c, ox + p.x * S, oy + p.y * S);
        g.globalAlpha = Math.min(1, t * 3);
        if (p.kind === 'banana') g.drawImage(art.banana, -p.size, -p.size, p.size * 2, p.size * 2);
        else {
          g.fillStyle = p.color;
          g.fillRect(-p.size / 2, -p.size * 0.18, p.size, p.size * 0.36);
        }
        continue;
      }
      g.setTransform(S, 0, 0, S, ox, oy);
      let img = art.bubble;
      let a = Math.min(1, t * 2.5);
      if (p.kind === 'smoke') {
        img = art.smoke;
        a = t * 0.85;
      } else if (p.kind === 'dust') {
        img = art.dust;
        a = t * 0.7;
      }
      g.globalAlpha = a;
      g.drawImage(img, p.x - p.size, p.y - p.size, p.size * 2, p.size * 2);
    }
    g.globalCompositeOperation = 'lighter';
    g.setTransform(S, 0, 0, S, ox, oy);
    for (const p of parts) {
      if (p.kind !== 'spark') continue;
      const t = p.life / p.max;
      g.globalAlpha = t;
      const r = p.size * (0.5 + 0.5 * t);
      g.drawImage(art.spark, p.x - r, p.y - r, r * 2, r * 2);
    }
    g.globalCompositeOperation = 'source-over';
    g.globalAlpha = 1;

    g.setTransform(1, 0, 0, 1, 0, 0);
    if (art.fg) tile(art.fg, sc * FORE, FG_PERIOD, Math.round((vh - art.fgHeight) * S), shx);
  }

  function drawHeli(px, py, angle) {
    const k = S * 0.1;
    const stretch = mode === 'play' ? clamp(heli.vy / VMAX, 0, 1) * 0.06 : 0;
    const sx = 1 + heli.sq - stretch * 0.5;
    const sy = 1 - heli.sq + stretch;
    const c = Math.cos(angle);
    const s = Math.sin(angle);
    g.setTransform(c * k * sx, s * k * sx, -s * k * sy, c * k * sy, px, py);
    const alpha = heli.alpha;
    if (alpha <= 0) return;

    if (heli.beam > 0.02) {
      const flicker = mode === 'crash' || mode === 'over' ? (Math.random() < 0.3 ? 0.3 : 1) : 1;
      g.globalCompositeOperation = 'lighter';
      g.globalAlpha = 0.34 * heli.beam * alpha * flicker;
      g.save();
      g.translate(LAMP[0], LAMP[1]);
      g.rotate(0.36);
      g.drawImage(art.beam, -2, -70, 320, 140);
      g.restore();
      g.globalAlpha = 0.6 * heli.beam * alpha * flicker;
      g.drawImage(art.glow, LAMP[0] - 9, LAMP[1] - 9, 18, 18);
      g.globalCompositeOperation = 'source-over';
    }
    g.globalAlpha = alpha;
    const pad = 2 / k;
    g.drawImage(art.tail, TAIL[0] - TAIL_R - pad, TAIL[1] - TAIL_R - pad, TAIL_R * 2 + pad * 2, TAIL_R * 2 + pad * 2);
    g.save();
    g.translate(TAIL[0], TAIL[1]);
    g.rotate(heli.tail);
    g.fillStyle = '#16130f';
    g.fillRect(-TAIL_R, -1.1, TAIL_R * 2, 2.2);
    g.restore();
    const [bx, by, bw, bh] = BODY_BOX;
    g.drawImage(art.heli, bx, by, bw, bh);
    g.drawImage(art.rotor, HUB[0] - ROTOR_HALF - pad, HUB[1] - 4.5 - pad, ROTOR_HALF * 2 + pad * 2, 9 + pad * 2);
    const la = Math.abs(Math.cos(heli.rotor)) * ROTOR_HALF;
    const lb = Math.abs(Math.sin(heli.rotor)) * ROTOR_HALF;
    g.fillStyle = '#141110';
    g.fillRect(HUB[0] - la, HUB[1] - 1.3, la * 2, 2.6);
    g.globalAlpha = alpha * 0.45;
    g.fillRect(HUB[0] - lb, HUB[1] - 1.1, lb * 2, 2.2);
    g.globalAlpha = alpha;
    g.fillStyle = '#2a2622';
    g.beginPath();
    g.arc(HUB[0], HUB[1], 2.6, 0, Math.PI * 2);
    g.fill();
    // the red beacon on the fin blinks while the engine runs
    if (mode !== 'crash' && mode !== 'over' && time % 1.2 < 0.13) {
      g.globalCompositeOperation = 'lighter';
      g.drawImage(art.red, BEACON[0] - 9, BEACON[1] - 9, 18, 18);
      g.globalCompositeOperation = 'source-over';
    }
    g.fillStyle = '#ff5a44';
    g.beginPath();
    g.arc(BEACON[0], BEACON[1], 1.6, 0, Math.PI * 2);
    g.fill();
    g.globalAlpha = 1;
  }

  // ------------------------------------------------------------------------------------------------------------
  // What Game.svelte calls

  /** Moves time on by dt seconds (already clamped) and draws. Frozen: draws the last picture once and waits. */
  function advance(dt, frozen) {
    if (!art) return;
    if (frozen) {
      if (still++ === 0) draw(1);
      return;
    }
    acc += dt;
    let n = 0;
    while (acc >= STEP && n < 30) {
      step(STEP);
      acc -= STEP;
      n++;
    }
    if (n === 30) acc = 0;
    // Once the wreck lies still and every bit has landed, the picture stops changing: draw it once and rest.
    const settled = mode === 'over' && heli.down && parts.length === 0 && speed < 0.05 && !shake;
    if (!settled) still = 0;
    else if (still++ > 1) return;
    draw(acc / STEP);
  }

  function begin() {
    mode = 'play';
    modeT = 0;
    gaps = 0;
    spawned = 0;
    stacks = [];
    heli.enter = 1;
    lastCenter = clamp(heli.y, 30, 70);
    lastWx = scroll + heli.x + FIRST - spacingAt(0);
    let i = 0;
    while (lastWx - scroll < vw - flyX + CW) {
      const s = spawn();
      if (s.wx - scroll < vw - flyX + CW / 2 + 2) {
        s.appear = 0;
        s.delay = 0.05 + i * 0.07;
        i++;
      }
    }
    flap();
  }

  function flap() {
    if (mode !== 'play') return;
    heli.vy = FLAP;
    heli.sqv += 5;
    heli.spin = 44;
    const hx = scroll + heli.x;
    for (let i = 0; i < 3; i++) {
      emit('dust', hx + rand(-4, 2), heli.y + 3.6, rand(-10, -2), rand(10, 18), 0.45, rand(0.7, 1.2), { grow: 2.2, drag: 3 });
    }
  }

  /** From the result screen back to a hovering helicopter: the stacks sink away, a new one flies in. */
  function again() {
    if (mode !== 'over' && mode !== 'crash') return;
    for (const s of stacks) {
      s.target = 0;
      s.delay = 0;
    }
    const hx = scroll + heli.x;
    for (let i = 0; i < 10; i++) {
      emit('smoke', hx + rand(-5, 5), heli.y + rand(-2, 2), rand(-6, 6), rand(-8, 2), rand(0.5, 0.8), rand(1.2, 2), { grow: 3, drag: 2 });
    }
    parts.length = Math.min(parts.length, 200);
    mode = 'leave';
    modeT = 0;
    still = 0;
  }

  /** Bananas rain down for a new record (many) or a new personal best (a few). */
  function celebrate(count) {
    const left = scroll - flyX;
    for (let i = 0; i < count; i++) {
      emit('banana', left + rand(0, vw), camTop - rand(2, 40), rand(-3, 3), rand(4, 22), rand(3.2, 4.2), rand(2.4, 3.6), {
        rot: rand(0, 6), vr: rand(-4, 4), grav: 34, drag: 0.25,
      });
    }
    still = 0;
  }

  return { resize, repaint, advance, begin, flap, again, celebrate };
}
