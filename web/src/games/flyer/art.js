// Everything the flyer draws, painted once into small canvases whenever the size changes. The game then only copies
// these pictures each frame (no per-frame shadows or filters), which keeps it smooth on the bar's PC.
// All sizes are in world units: the room is 100 units from ceiling (0) to floor (100). S is device pixels per unit.

import { HELI, BODY, HUB, ROTOR_HALF, TAIL_R, LAMP, BODY_BOX, MONKEY, BODY_TOP, BODY_MID, BODY_LOW, STRIPE, CRATE_COLORS } from './heli.js';

export { CRATE_COLORS };

export const CW = 12; // a beer crate, side on
export const CH = 8;
export const NECK = 2.9; // how far the bottles stand out of a crate
export const FAR_PERIOD = 160;
export const MID_PERIOD = 128;
export const CEIL_PERIOD = 24;
export const FLOOR_PERIOD = 36;
export const FG_PERIOD = 84;
export const LAMPS = [26, 90]; // x of the pendant lamps on the middle layer
export const LAMP_Y = 13;

const FONT = '"Barlow Condensed", "Barlow Semi Condensed", "Segoe UI", sans-serif';

// The fairy lights on the middle layer: two loose strings across each period, a bulb every four units.
const BULB_COLORS = ['#ffd27a', '#ff9b5a', '#ffe9b0', '#ffb224'];
export const BULBS = [];
for (let x = 2; x < MID_PERIOD; x += 4) {
  const local = (x % 64) / 64;
  BULBS.push({ x, y: 2.5 + 6 * 4 * local * (1 - local), color: BULB_COLORS[(x / 4) % 4 | 0], phase: x * 1.7 });
}

export function rng(seed) {
  return () => {
    seed = (seed + 0x6d2b79f5) | 0;
    let t = Math.imul(seed ^ (seed >>> 15), 1 | seed);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

function canvas(w, h) {
  const c = document.createElement('canvas');
  c.width = Math.max(1, Math.ceil(w));
  c.height = Math.max(1, Math.ceil(h));
  return c;
}

// A tile that repeats must be a whole number of pixels wide, or a hairline seam shows where two copies meet. So
// tiles are drawn a hair wider or narrower than S across: this many pixels per unit.
function across(period, S) {
  return Math.max(1, Math.round(period * S)) / period;
}

function rr(g, x, y, w, h, r) {
  g.beginPath();
  g.moveTo(x + r, y);
  g.arcTo(x + w, y, x + w, y + h, r);
  g.arcTo(x + w, y + h, x, y + h, r);
  g.arcTo(x, y + h, x, y, r);
  g.arcTo(x, y, x + w, y, r);
  g.closePath();
}

function circle(g, x, y, r) {
  g.beginPath();
  g.arc(x, y, r, 0, Math.PI * 2);
}

let heliPaths = null;
function paths() {
  if (!heliPaths) {
    heliPaths = {};
    for (const key in HELI) heliPaths[key] = new Path2D(HELI[key]);
  }
  return heliPaths;
}

/** Draws the monkey (or any list of simple shapes from heli.js) in the current transform. */
export function drawShapes(g, shapes) {
  for (const s of shapes) {
    const path = s.p ? new Path2D(s.p) : null;
    if (s.e) {
      g.beginPath();
      g.ellipse(s.e[0], s.e[1], s.e[2], s.e[3], 0, 0, Math.PI * 2);
    }
    if (s.fill) {
      g.fillStyle = s.fill;
      if (path) g.fill(path);
      else g.fill();
    }
    if (s.stroke) {
      g.strokeStyle = s.stroke;
      g.lineWidth = s.w;
      g.lineCap = 'round';
      if (path) g.stroke(path);
      else g.stroke();
    }
  }
}

// ---------------------------------------------------------------------------------------------------------------
// The helicopter

function heliBody(S) {
  const u = S * 0.1;
  const [bx, by, bw, bh] = BODY_BOX;
  const c = canvas(bw * u, bh * u);
  const g = c.getContext('2d');
  g.scale(u, u);
  g.translate(-bx, -by);
  const p = paths();
  const body = BODY.map((k) => p[k]);
  g.lineJoin = 'round';
  g.lineCap = 'round';

  // One dark outline around the whole shape, then the fills on top of it.
  g.strokeStyle = '#050404';
  g.lineWidth = 4.4;
  for (const q of body) g.stroke(q);
  g.lineWidth = 6.6;
  g.stroke(p.struts);
  g.stroke(p.skid);
  g.strokeStyle = '#2a2622';
  g.lineWidth = 3.2;
  g.stroke(p.struts);
  g.stroke(p.skid);
  g.strokeStyle = 'rgba(255,196,120,0.28)';
  g.lineWidth = 0.9;
  g.save();
  g.translate(0, -0.9);
  g.stroke(p.skid);
  g.restore();

  const shade = g.createLinearGradient(0, -40, 0, 24);
  shade.addColorStop(0, BODY_TOP);
  shade.addColorStop(0.35, BODY_MID);
  shade.addColorStop(0.75, '#1b1917');
  shade.addColorStop(1, BODY_LOW);
  g.fillStyle = shade;
  for (const q of body) g.fill(q);

  const all = new Path2D();
  for (const q of body) all.addPath(q);
  g.save();
  g.clip(all);
  g.fillStyle = STRIPE;
  g.fill(p.stripe);
  g.fillStyle = 'rgba(110,60,0,0.55)';
  g.fillRect(-130, 10.6, 200, 1.6);
  // Light from the bar's lamps above: the top of every part catches it.
  const top = g.createLinearGradient(0, -36, 0, -6);
  top.addColorStop(0, 'rgba(255,196,110,0.42)');
  top.addColorStop(1, 'rgba(255,196,110,0)');
  g.fillStyle = top;
  g.fillRect(-130, -50, 200, 46);
  // and a warm bounce from below
  const low = g.createLinearGradient(0, 12, 0, 24);
  low.addColorStop(0, 'rgba(255,150,60,0)');
  low.addColorStop(1, 'rgba(255,150,60,0.16)');
  g.fillStyle = low;
  g.fillRect(-130, 10, 200, 16);
  g.strokeStyle = 'rgba(0,0,0,0.55)';
  g.lineWidth = 0.8;
  g.stroke(p.door);
  g.restore();

  // Windscreen
  let glass = g.createLinearGradient(19, -21, 50, 6);
  glass.addColorStop(0, '#d8f1f7');
  glass.addColorStop(0.45, '#79b8cf');
  glass.addColorStop(1, '#2c6680');
  g.fillStyle = glass;
  g.fill(p.windshield);
  g.save();
  g.clip(p.windshield);
  g.fillStyle = 'rgba(255,255,255,0.55)';
  g.beginPath();
  g.moveTo(27, -23);
  g.lineTo(33, -23);
  g.lineTo(23, 8);
  g.lineTo(17, 8);
  g.fill();
  g.restore();

  // Side window with the monkey
  g.save();
  g.clip(p.window);
  glass = g.createLinearGradient(0, -20, 0, 7);
  glass.addColorStop(0, '#3a6880');
  glass.addColorStop(1, '#15262f');
  g.fillStyle = glass;
  g.fillRect(-21, -21, 36, 29);
  drawShapes(g, MONKEY);
  g.fillStyle = 'rgba(255,255,255,0.14)';
  g.beginPath();
  g.moveTo(-8, -21);
  g.lineTo(0, -21);
  g.lineTo(-12, 8);
  g.lineTo(-20, 8);
  g.fill();
  g.restore();
  g.strokeStyle = '#070606';
  g.lineWidth = 1.6;
  g.stroke(p.window);
  g.stroke(p.windshield);

  // Rotor hub and the searchlight in the nose
  g.fillStyle = '#1d1a17';
  circle(g, HUB[0], HUB[1], 3.4);
  g.fill();
  g.strokeStyle = '#050404';
  g.lineWidth = 1.3;
  g.stroke();
  g.fillStyle = '#fff3cc';
  circle(g, LAMP[0], LAMP[1], 2.4);
  g.fill();
  return c;
}

function rotorDisc(S) {
  const u = S * 0.1;
  const c = canvas(ROTOR_HALF * 2 * u + 4, 9 * u + 4);
  const g = c.getContext('2d');
  g.translate(2, 2);
  g.scale(u, u);
  g.beginPath();
  g.ellipse(ROTOR_HALF, 4.5, ROTOR_HALF - 0.5, 3.6, 0, 0, Math.PI * 2);
  g.fillStyle = 'rgba(16,14,12,0.24)';
  g.fill();
  g.strokeStyle = 'rgba(255,205,140,0.16)';
  g.lineWidth = 0.9;
  g.stroke();
  return c;
}

function tailDisc(S) {
  const u = S * 0.1;
  const c = canvas(TAIL_R * 2 * u + 4, TAIL_R * 2 * u + 4);
  const g = c.getContext('2d');
  g.translate(2, 2);
  g.scale(u, u);
  circle(g, TAIL_R, TAIL_R, TAIL_R - 0.4);
  g.fillStyle = 'rgba(16,14,12,0.28)';
  g.fill();
  g.strokeStyle = 'rgba(255,205,140,0.18)';
  g.lineWidth = 0.8;
  g.stroke();
  return c;
}

// The searchlight's beam, pointing right from its left middle. Low resolution: it is soft anyway.
function beam() {
  const w = 256;
  const h = 128;
  const c = canvas(w, h);
  const g = c.getContext('2d');
  for (const [spread, a] of [[0.95, 0.3], [0.7, 0.35], [0.45, 0.4]]) {
    const gr = g.createRadialGradient(0, h / 2, 0, 0, h / 2, w);
    gr.addColorStop(0, `rgba(255,222,160,${a})`);
    gr.addColorStop(0.4, `rgba(255,196,120,${a * 0.45})`);
    gr.addColorStop(1, 'rgba(255,180,90,0)');
    g.fillStyle = gr;
    g.beginPath();
    g.moveTo(0, h / 2 - 3);
    g.lineTo(w, h / 2 - (h / 2) * spread);
    g.lineTo(w, h / 2 + (h / 2) * spread);
    g.lineTo(0, h / 2 + 3);
    g.fill();
  }
  return c;
}

// ---------------------------------------------------------------------------------------------------------------
// Crates

function shade(hex, amount) {
  const n = parseInt(hex.slice(1), 16);
  const mix = (v) => Math.round(amount < 0 ? v * (1 + amount) : v + (255 - v) * amount);
  return `rgb(${mix(n >> 16)},${mix((n >> 8) & 255)},${mix(n & 255)})`;
}

function crate(S, base, label) {
  const c = canvas(CW * S + 2, CH * S + 2);
  const g = c.getContext('2d');
  g.translate(1, 1);
  g.save();
  g.scale(S, S);
  rr(g, 0, 0, CW, CH, 0.6);
  g.fillStyle = base;
  g.fill();
  g.save();
  g.clip();
  // The bar's warm light pulls every colour a little towards amber.
  g.fillStyle = 'rgba(255,160,60,0.10)';
  g.fillRect(0, 0, CW, CH);
  let gr = g.createLinearGradient(0, 0, 0, CH);
  gr.addColorStop(0, 'rgba(255,214,150,0.24)');
  gr.addColorStop(0.45, 'rgba(255,214,150,0.03)');
  gr.addColorStop(1, 'rgba(0,0,0,0.34)');
  g.fillStyle = gr;
  g.fillRect(0, 0, CW, CH);
  gr = g.createLinearGradient(CW * 0.68, 0, CW, 0);
  gr.addColorStop(0, 'rgba(0,0,0,0)');
  gr.addColorStop(1, 'rgba(0,0,0,0.3)');
  g.fillStyle = gr;
  g.fillRect(0, 0, CW, CH);
  // rim
  g.fillStyle = shade(base, 0.16);
  g.fillRect(0, 0, CW, 1.05);
  g.fillStyle = 'rgba(0,0,0,0.3)';
  g.fillRect(0, 1.05, CW, 0.2);
  // handle
  rr(g, CW / 2 - 2, 1.75, 4, 1.3, 0.65);
  g.fillStyle = 'rgba(8,5,3,0.82)';
  g.fill();
  g.fillStyle = 'rgba(255,230,190,0.18)';
  g.fillRect(CW / 2 - 1.5, 3.02, 3, 0.13);
  // ribs
  for (const x of [1.15, 2.45, CW - 2.79, CW - 1.49]) {
    g.fillStyle = 'rgba(0,0,0,0.24)';
    g.fillRect(x, 3.5, 0.34, 3.7);
    g.fillStyle = 'rgba(255,230,190,0.09)';
    g.fillRect(x + 0.34, 3.5, 0.14, 3.7);
  }
  if (label) {
    rr(g, CW / 2 - 2.1, 3.75, 4.2, 2.6, 0.3);
    g.fillStyle = shade(base, 0.32);
    g.fill();
    g.fillStyle = 'rgba(0,0,0,0.12)';
    g.fillRect(CW / 2 - 2.1, 5.9, 4.2, 0.45);
  }
  g.fillStyle = 'rgba(0,0,0,0.32)';
  g.fillRect(0, CH - 0.7, CW, 0.7);
  g.restore();
  rr(g, 0, 0, CW, CH, 0.6);
  g.lineWidth = 0.16;
  g.strokeStyle = 'rgba(0,0,0,0.62)';
  g.stroke();
  g.beginPath();
  g.moveTo(0.55, 0.09);
  g.lineTo(CW - 0.55, 0.09);
  g.strokeStyle = 'rgba(255,205,130,0.5)';
  g.lineWidth = 0.13;
  g.stroke();
  g.restore();
  if (label) {
    // The label in device pixels, so small text stays crisp.
    g.font = `900 ${(1.75 * S).toFixed(1)}px ${FONT}`;
    g.textAlign = 'center';
    g.textBaseline = 'middle';
    g.fillStyle = shade(base, -0.45);
    g.fillText('AMK', (CW / 2) * S, 5.0 * S);
  }
  return c;
}

function bottles(S, seed) {
  const r = rng(seed);
  const h = NECK + 0.6;
  const c = canvas(CW * S + 2, h * S + 2);
  const g = c.getContext('2d');
  g.translate(1, 1);
  g.scale(S, S);
  const glass = ['#5c2c0d', '#6f3b12', '#24502a', '#5c2c0d'];
  for (let i = 0; i < 5; i++) {
    const x = 1.5 + i * 2.25 + (r() - 0.5) * 0.25;
    const lift = (r() - 0.5) * 0.35;
    const col = glass[(r() * glass.length) | 0];
    const gr = g.createLinearGradient(x - 0.55, 0, x + 0.55, 0);
    gr.addColorStop(0, shade(col, 0.35));
    gr.addColorStop(0.35, col);
    gr.addColorStop(1, shade(col, -0.5));
    g.fillStyle = gr;
    rr(g, x - 0.52, 0.6 + lift, 1.04, h, 0.25);
    g.fill();
    g.fillStyle = 'rgba(255,232,190,0.38)';
    g.fillRect(x - 0.34, 1.0 + lift, 0.16, h - 1.2);
    // crown cap
    const cap = g.createLinearGradient(0, lift, 0, 0.85 + lift);
    cap.addColorStop(0, '#ffe7a0');
    cap.addColorStop(0.5, '#e2a93a');
    cap.addColorStop(1, '#8c5c10');
    g.fillStyle = cap;
    rr(g, x - 0.68, lift, 1.36, 0.85, 0.18);
    g.fill();
    g.fillStyle = 'rgba(70,40,5,0.6)';
    for (let k = -2; k <= 2; k++) g.fillRect(x + k * 0.27 - 0.05, 0.62 + lift, 0.1, 0.22);
  }
  return c;
}

// The soft shadow a stack throws on the wall to its right.
function stackShadow() {
  const c = canvas(32, 1);
  const g = c.getContext('2d');
  const gr = g.createLinearGradient(0, 0, 32, 0);
  gr.addColorStop(0, 'rgba(0,0,0,0.45)');
  gr.addColorStop(1, 'rgba(0,0,0,0)');
  g.fillStyle = gr;
  g.fillRect(0, 0, 32, 1);
  return c;
}

// ---------------------------------------------------------------------------------------------------------------
// Small things: glows, particles, the banana for a record

function glow(r, gr, b, px = 64) {
  const c = canvas(px, px);
  const g = c.getContext('2d');
  const grad = g.createRadialGradient(px / 2, px / 2, 0, px / 2, px / 2, px / 2);
  grad.addColorStop(0, `rgba(${r},${gr},${b},1)`);
  grad.addColorStop(0.22, `rgba(${r},${gr},${b},0.5)`);
  grad.addColorStop(1, `rgba(${r},${gr},${b},0)`);
  g.fillStyle = grad;
  g.fillRect(0, 0, px, px);
  return c;
}

function puff(r, gr, b, a, px = 48) {
  const c = canvas(px, px);
  const g = c.getContext('2d');
  const grad = g.createRadialGradient(px / 2, px / 2, 0, px / 2, px / 2, px / 2);
  grad.addColorStop(0, `rgba(${r},${gr},${b},${a})`);
  grad.addColorStop(0.55, `rgba(${r},${gr},${b},${a * 0.6})`);
  grad.addColorStop(1, `rgba(${r},${gr},${b},0)`);
  g.fillStyle = grad;
  g.fillRect(0, 0, px, px);
  return c;
}

function bubble(px = 32) {
  const c = canvas(px, px);
  const g = c.getContext('2d');
  const h = px / 2;
  circle(g, h, h, h - 1);
  const grad = g.createRadialGradient(h * 0.7, h * 0.7, 0, h, h, h);
  grad.addColorStop(0, '#fffaf0');
  grad.addColorStop(0.7, '#f4e3c2');
  grad.addColorStop(1, '#cdb48a');
  g.fillStyle = grad;
  g.fill();
  return c;
}

function banana(px = 64) {
  const c = canvas(px, px);
  const g = c.getContext('2d');
  g.translate(px / 2, px / 2);
  g.scale(px / 26, px / 26);
  g.rotate(-0.5);
  const p = new Path2D('M-9 4 C-6 10 6 10 10 -2 C7 2 -2 6 -9 1 Z');
  g.lineJoin = 'round';
  g.strokeStyle = '#5a3d06';
  g.lineWidth = 1.4;
  g.stroke(p);
  g.fillStyle = '#f6c41f';
  g.fill(p);
  g.strokeStyle = '#fff1a6';
  g.lineWidth = 0.8;
  g.beginPath();
  g.moveTo(-7, 4.5);
  g.quadraticCurveTo(0, 8, 7, 1);
  g.stroke();
  g.fillStyle = '#4a3410';
  g.fillRect(9, -3.6, 2, 2);
  return c;
}

// ---------------------------------------------------------------------------------------------------------------
// The room: back wall (far), shelves and lamps (middle), ceiling and floor (front), glasses in the foreground

function wall(S, camTop, vh) {
  const sx = across(FAR_PERIOD, S);
  const c = canvas(Math.round(FAR_PERIOD * sx), vh * S);
  const g = c.getContext('2d');
  g.scale(sx, S);
  g.translate(0, -camTop);
  const top = camTop;
  const bottom = camTop + vh;
  const W = FAR_PERIOD;
  const r = rng(11);

  let gr = g.createLinearGradient(0, -20, 0, 120);
  gr.addColorStop(0, '#17110d');
  gr.addColorStop(0.42, '#261c15');
  gr.addColorStop(1, '#100b08');
  g.fillStyle = gr;
  g.fillRect(0, top, W, vh);

  // Brick, very quiet, above the rail.
  const tint = [];
  for (let i = 0; i < 400; i++) tint.push(r());
  for (let row = 0, y = 63.2; y > top - 3; row++, y -= 3) {
    const off = row % 2 ? 4 : 0;
    for (let col = -1; col <= W / 8; col++) {
      const x = col * 8 - off;
      const v = tint[(row * 20 + ((col + 20) % 20)) % 400];
      g.fillStyle = `rgba(150,86,48,${0.05 + v * 0.09})`;
      g.fillRect(x + 0.2, y - 2.8, 7.6, 2.6);
    }
  }

  // Wood panelling below the rail.
  g.fillStyle = '#140e0a';
  g.fillRect(0, 64.6, W, bottom - 64.6);
  for (let x = 0; x < W; x += 16) {
    rr(g, x + 1.4, 68, 13.2, 28, 0.6);
    g.strokeStyle = 'rgba(255,190,120,0.06)';
    g.lineWidth = 0.35;
    g.stroke();
    g.fillStyle = 'rgba(0,0,0,0.18)';
    g.fillRect(x + 1.6, 68.2, 12.8, 0.6);
  }
  g.fillStyle = '#2c1e14';
  g.fillRect(0, 63.2, W, 1.6);
  g.fillStyle = 'rgba(255,190,110,0.2)';
  g.fillRect(0, 63.2, W, 0.25);

  // Window with the night over Münster
  const wx = 12;
  const ww = 26;
  const winPath = new Path2D();
  winPath.moveTo(wx, 52);
  winPath.lineTo(wx, 23);
  winPath.arc(wx + ww / 2, 23, ww / 2, Math.PI, 0);
  winPath.lineTo(wx + ww, 52);
  winPath.closePath();
  g.save();
  g.clip(winPath);
  gr = g.createLinearGradient(0, 10, 0, 52);
  gr.addColorStop(0, '#0a1322');
  gr.addColorStop(1, '#1d2c48');
  g.fillStyle = gr;
  g.fillRect(wx, 8, ww, 46);
  for (let i = 0; i < 22; i++) {
    g.fillStyle = `rgba(255,248,225,${0.35 + r() * 0.5})`;
    const s = 0.18 + r() * 0.22;
    g.fillRect(wx + r() * ww, 11 + r() * 26, s, s);
  }
  const moon = g.createRadialGradient(31, 19, 0, 31, 19, 9);
  moon.addColorStop(0, 'rgba(244,231,198,0.35)');
  moon.addColorStop(1, 'rgba(244,231,198,0)');
  g.fillStyle = moon;
  g.fillRect(20, 8, 22, 22);
  g.drawImage(crescent(S), 31 - 3, 19 - 3, 6, 6);
  // roofs and a church spire
  g.fillStyle = '#070a12';
  g.beginPath();
  g.moveTo(wx, 53);
  g.lineTo(wx, 46);
  g.lineTo(wx + 3, 43.5);
  g.lineTo(wx + 6, 46);
  g.lineTo(wx + 6, 44);
  g.lineTo(wx + 10, 40.5);
  g.lineTo(wx + 14, 44);
  g.lineTo(wx + 15.2, 44);
  g.lineTo(wx + 16.6, 28);
  g.lineTo(wx + 18, 44);
  g.lineTo(wx + 19, 44);
  g.lineTo(wx + 22, 41.5);
  g.lineTo(wx + 26, 45);
  g.lineTo(wx + 26, 53);
  g.fill();
  g.fillStyle = 'rgba(255,190,90,0.75)';
  for (const [x, y] of [[wx + 2.4, 47.5], [wx + 9.4, 46], [wx + 11, 48.6], [wx + 22.6, 47.4]]) g.fillRect(x, y, 0.7, 0.9);
  g.fillStyle = 'rgba(255,190,110,0.05)';
  g.beginPath();
  g.moveTo(wx + 4, 52);
  g.lineTo(wx + 14, 10);
  g.lineTo(wx + 19, 10);
  g.lineTo(wx + 9, 52);
  g.fill();
  g.restore();
  g.lineWidth = 1.4;
  g.strokeStyle = '#0c0805';
  g.stroke(winPath);
  g.fillStyle = '#0c0805';
  g.fillRect(wx + ww / 2 - 0.45, 10, 0.9, 42);
  g.fillRect(wx, 33, ww, 0.9);
  g.fillStyle = '#2b1d13';
  g.fillRect(wx - 1.5, 52, ww + 3, 1.6);
  g.fillStyle = 'rgba(255,190,110,0.22)';
  g.fillRect(wx - 1.5, 52, ww + 3, 0.25);

  // Neon sign: BAR over AMK. Its glow is painted once here, never per frame.
  g.save();
  g.setTransform(1, 0, 0, 1, 0, 0);
  g.textAlign = 'center';
  g.textBaseline = 'alphabetic';
  g.lineJoin = 'round';
  const neon = (text, x, y, size, tube, halo) => {
    g.font = `800 ${(size * S).toFixed(1)}px ${FONT}`;
    g.lineWidth = Math.max(1, 0.5 * S);
    g.shadowColor = halo;
    for (const blur of [2.4, 1.1]) {
      g.shadowBlur = blur * S;
      g.strokeStyle = tube;
      g.strokeText(text, x * sx, (y - camTop) * S);
    }
    g.shadowBlur = 0;
    g.lineWidth = Math.max(0.6, 0.18 * S);
    g.strokeStyle = 'rgba(255,250,235,0.9)';
    g.strokeText(text, x * sx, (y - camTop) * S);
  };
  neon('BAR', 66, 24, 5, '#ff7766', '#ff3b2a');
  neon('AMK', 66, 36, 11, '#ffd27a', '#ff9d1c');
  g.restore();

  // Poster with a banana
  g.fillStyle = '#3b2a1b';
  g.fillRect(90, 17, 16, 23);
  g.fillStyle = '#5e211b';
  g.fillRect(91, 18, 14, 21);
  g.save();
  g.translate(98, 25.5);
  g.rotate(-0.35);
  g.scale(0.42, 0.42);
  g.fillStyle = '#d9a52a';
  g.fill(new Path2D('M-9 4 C-6 10 6 10 10 -2 C7 2 -2 6 -9 1 Z'));
  g.restore();
  g.fillStyle = 'rgba(240,220,180,0.45)';
  g.fillRect(93, 32, 10, 1);
  g.fillRect(94.5, 34.4, 7, 0.7);
  g.fillRect(94, 36.4, 8, 0.7);
  g.fillStyle = 'rgba(0,0,0,0.22)';
  g.fillRect(91, 18, 14, 21);

  // Dartboard
  const dx = 122;
  const dy = 33;
  circle(g, dx, dy, 6.6);
  g.fillStyle = '#0d0907';
  g.fill();
  for (let i = 0; i < 20; i++) {
    const a0 = (i / 20) * Math.PI * 2 - Math.PI / 20;
    const a1 = a0 + Math.PI / 10;
    for (const [r0, r1, col] of [
      [0.9, 3.0, i % 2 ? '#a8956f' : '#1a1511'],
      [3.0, 3.4, i % 2 ? '#2c5a34' : '#8c2620'],
      [3.4, 4.9, i % 2 ? '#a8956f' : '#1a1511'],
      [4.9, 5.3, i % 2 ? '#2c5a34' : '#8c2620'],
    ]) {
      g.beginPath();
      g.arc(dx, dy, r1, a0, a1);
      g.arc(dx, dy, r0, a1, a0, true);
      g.fillStyle = col;
      g.fill();
    }
  }
  circle(g, dx, dy, 0.9);
  g.fillStyle = '#2c5a34';
  g.fill();
  circle(g, dx, dy, 0.45);
  g.fillStyle = '#8c2620';
  g.fill();
  circle(g, dx, dy, 5.4);
  g.fillStyle = 'rgba(25,14,6,0.4)';
  g.fill();
  for (const [ax, ay, ang] of [[dx + 1.6, dy - 1.2, -0.5], [dx - 2.2, dy + 2.4, -0.2]]) {
    g.save();
    g.translate(ax, ay);
    g.rotate(ang);
    g.fillStyle = '#cfc6b4';
    g.fillRect(0, -0.12, 2.6, 0.24);
    g.fillStyle = '#c0392b';
    g.beginPath();
    g.moveTo(2.4, 0);
    g.lineTo(3.8, -0.9);
    g.lineTo(3.8, 0.9);
    g.fill();
    g.restore();
  }

  // Chalkboard with the prices
  g.fillStyle = '#3b2a1b';
  g.fillRect(137, 19, 19, 24);
  g.fillStyle = '#1b231f';
  g.fillRect(138, 20, 17, 22);
  g.fillStyle = 'rgba(235,232,220,0.5)';
  g.fillRect(140.5, 22.2, 12, 0.8);
  for (let i = 0; i < 6; i++) {
    const y = 25.4 + i * 2.6;
    g.fillStyle = 'rgba(235,232,220,0.32)';
    g.fillRect(140, y, 5 + r() * 5, 0.5);
    g.fillRect(150.4, y, 2.6, 0.5);
  }

  // Shadow from the ceiling and the floor, so the middle of the room is the lightest.
  gr = g.createLinearGradient(0, top, 0, 22);
  gr.addColorStop(0, 'rgba(0,0,0,0.55)');
  gr.addColorStop(1, 'rgba(0,0,0,0)');
  g.fillStyle = gr;
  g.fillRect(0, top, W, 22 - top);
  gr = g.createLinearGradient(0, 78, 0, bottom);
  gr.addColorStop(0, 'rgba(0,0,0,0)');
  gr.addColorStop(1, 'rgba(0,0,0,0.5)');
  g.fillStyle = gr;
  g.fillRect(0, 78, W, bottom - 78);
  return c;
}

// The moon as a real crescent: a disc with a bite taken out, so the sky shows through the bite.
function crescent(S) {
  const px = Math.max(8, Math.ceil(6 * S));
  const c = canvas(px, px);
  const g = c.getContext('2d');
  g.scale(px / 6, px / 6);
  circle(g, 3, 3, 2.8);
  g.fillStyle = '#f2e5c2';
  g.fill();
  g.globalCompositeOperation = 'destination-out';
  circle(g, 4.3, 2.2, 2.5);
  g.fill();
  return c;
}

function middle(S, camTop, vh) {
  const sx = across(MID_PERIOD, S);
  const c = canvas(Math.round(MID_PERIOD * sx), vh * S);
  const g = c.getContext('2d');
  g.scale(sx, S);
  g.translate(0, -camTop);
  const r = rng(23);

  // Shelf of spirits
  g.fillStyle = '#24170e';
  g.fillRect(38, 47, 42, 1.2);
  g.fillStyle = 'rgba(255,190,110,0.22)';
  g.fillRect(38, 47, 42, 0.22);
  g.fillStyle = '#1a110a';
  g.fillRect(42, 48.2, 0.8, 2.4);
  g.fillRect(75, 48.2, 0.8, 2.4);
  const glassy = ['rgba(46,88,52,0.92)', 'rgba(108,54,18,0.92)', 'rgba(150,110,52,0.88)', 'rgba(190,196,186,0.4)', 'rgba(122,32,30,0.9)'];
  for (let x = 40; x < 78; ) {
    const kind = (r() * 3) | 0;
    const col = glassy[(r() * glassy.length) | 0];
    const w = kind === 1 ? 2.6 : 2;
    const h = kind === 0 ? 8 : kind === 1 ? 5.4 : 6.4;
    g.fillStyle = col;
    if (kind === 0) {
      rr(g, x, 47 - h + 3, w, h - 3, 0.5);
      g.fill();
      g.fillRect(x + w / 2 - 0.35, 47 - h, 0.7, 3.2);
    } else {
      rr(g, x, 47 - h, w, h, kind === 1 ? 0.2 : 0.8);
      g.fill();
      g.fillRect(x + w / 2 - 0.35, 47 - h - 1.2, 0.7, 1.4);
    }
    g.fillStyle = 'rgba(255,225,175,0.28)';
    g.fillRect(x + 0.35, 47 - h + 3.2, 0.25, h - 3.8);
    g.fillStyle = 'rgba(240,226,196,0.4)';
    g.fillRect(x + 0.2, 47 - h * 0.5, w - 0.4, 1.4);
    x += w + 0.7 + r() * 0.9;
  }

  // Shelf of glasses
  g.fillStyle = '#24170e';
  g.fillRect(96, 58, 28, 1.1);
  g.fillStyle = 'rgba(255,190,110,0.2)';
  g.fillRect(96, 58, 28, 0.2);
  for (let x = 98; x < 122; x += 4.6) {
    g.fillStyle = 'rgba(230,210,170,0.16)';
    rr(g, x, 52.4, 2.8, 5.6, 0.3);
    g.fill();
    g.strokeStyle = 'rgba(255,230,190,0.3)';
    g.lineWidth = 0.18;
    g.stroke();
    g.beginPath();
    g.moveTo(x + 2.8, 53.6);
    g.quadraticCurveTo(x + 4.3, 55.2, x + 2.8, 56.8);
    g.stroke();
  }

  // Bar stools
  for (const sx of [10, 62]) {
    g.strokeStyle = '#120c08';
    g.lineWidth = 0.5;
    g.beginPath();
    g.moveTo(sx - 2.1, 84.6);
    g.lineTo(sx - 3, 101);
    g.moveTo(sx + 2.1, 84.6);
    g.lineTo(sx + 3, 101);
    g.moveTo(sx, 84.6);
    g.lineTo(sx, 101);
    g.moveTo(sx - 2.6, 93);
    g.lineTo(sx + 2.6, 93);
    g.stroke();
    g.beginPath();
    g.ellipse(sx, 84, 3.3, 1, 0, 0, Math.PI * 2);
    g.fillStyle = '#1b130c';
    g.fill();
    g.fillStyle = 'rgba(255,190,110,0.16)';
    g.fillRect(sx - 2.6, 83.1, 5.2, 0.25);
  }

  // Pendant lamps: cord, shade and the lit rim
  for (const lx of LAMPS) {
    g.strokeStyle = '#0b0806';
    g.lineWidth = 0.25;
    g.beginPath();
    g.moveTo(lx, camTop);
    g.lineTo(lx, 9);
    g.stroke();
    g.fillStyle = '#2b1d10';
    g.fillRect(lx - 0.6, 8.4, 1.2, 1.1);
    const sh = g.createLinearGradient(lx - 3.4, 0, lx + 3.4, 0);
    sh.addColorStop(0, '#5a3a14');
    sh.addColorStop(0.45, '#d39a3c');
    sh.addColorStop(1, '#4a3012');
    g.fillStyle = sh;
    g.beginPath();
    g.moveTo(lx - 3.5, LAMP_Y);
    g.quadraticCurveTo(lx - 3.2, 9.4, lx, 9.3);
    g.quadraticCurveTo(lx + 3.2, 9.4, lx + 3.5, LAMP_Y);
    g.closePath();
    g.fill();
    g.beginPath();
    g.ellipse(lx, LAMP_Y, 3.4, 0.55, 0, 0, Math.PI * 2);
    g.fillStyle = '#fff0c0';
    g.fill();
  }

  // Fairy lights
  g.strokeStyle = 'rgba(16,11,8,0.95)';
  g.lineWidth = 0.18;
  g.beginPath();
  for (let x = 0; x <= MID_PERIOD; x += 1) {
    const local = (x % 64) / 64;
    const y = 2.5 + 6 * 4 * local * (1 - local);
    if (x === 0) g.moveTo(x, y);
    else g.lineTo(x, y);
  }
  g.stroke();
  for (const b of BULBS) {
    circle(g, b.x, b.y + 0.5, 0.45);
    g.fillStyle = b.color;
    g.fill();
  }
  return c;
}

function ceiling(S, camTop) {
  const h = 3 - camTop;
  const sx = across(CEIL_PERIOD, S);
  const c = canvas(Math.round(CEIL_PERIOD * sx), h * S);
  const g = c.getContext('2d');
  g.scale(sx, S);
  g.translate(0, -camTop);
  const W = CEIL_PERIOD;
  g.fillStyle = '#110b08';
  g.fillRect(0, camTop, W, -camTop);
  for (let y = -3.2; y > camTop; y -= 2.2) {
    g.fillStyle = 'rgba(255,200,140,0.035)';
    g.fillRect(0, y, W, 0.18);
  }
  const gr = g.createLinearGradient(0, -2.8, 0, 0);
  gr.addColorStop(0, '#3a2717');
  gr.addColorStop(1, '#22170e');
  g.fillStyle = gr;
  g.fillRect(0, -2.8, W, 2.8);
  g.fillStyle = 'rgba(0,0,0,0.18)';
  g.fillRect(0, -2.0, W, 0.12);
  g.fillRect(0, -1.1, W, 0.1);
  g.fillStyle = 'rgba(255,184,96,0.4)';
  g.fillRect(0, -0.32, W, 0.32);
  for (const bx of [6, 18]) {
    circle(g, bx, -1.45, 0.4);
    g.fillStyle = '#6b5a48';
    g.fill();
    circle(g, bx - 0.12, -1.57, 0.14);
    g.fillStyle = 'rgba(255,230,190,0.6)';
    g.fill();
  }
  const sh = g.createLinearGradient(0, 0, 0, 3);
  sh.addColorStop(0, 'rgba(0,0,0,0.45)');
  sh.addColorStop(1, 'rgba(0,0,0,0)');
  g.fillStyle = sh;
  g.fillRect(0, 0, W, 3);
  return c;
}

export const FLOOR_TOP = 99.4;

function floor(S, bottom) {
  const h = bottom - FLOOR_TOP;
  const sx = across(FLOOR_PERIOD, S);
  const c = canvas(Math.round(FLOOR_PERIOD * sx), h * S);
  const g = c.getContext('2d');
  g.scale(sx, S);
  g.translate(0, -FLOOR_TOP);
  const W = FLOOR_PERIOD;
  const sh = g.createLinearGradient(0, FLOOR_TOP, 0, 100);
  sh.addColorStop(0, 'rgba(0,0,0,0)');
  sh.addColorStop(1, 'rgba(0,0,0,0.35)');
  g.fillStyle = sh;
  g.fillRect(0, FLOOR_TOP, W, 0.6);
  let gr = g.createLinearGradient(0, 100, 0, 103);
  gr.addColorStop(0, '#4a3220');
  gr.addColorStop(1, '#2a1c12');
  g.fillStyle = gr;
  g.fillRect(0, 100, W, 3);
  g.fillStyle = 'rgba(255,190,110,0.45)';
  g.fillRect(0, 100, W, 0.25);
  g.strokeStyle = 'rgba(0,0,0,0.42)';
  g.lineWidth = 0.18;
  g.beginPath();
  for (let x = 2; x < W + 2; x += 6) {
    g.moveTo(x, 100.25);
    g.lineTo(x - 0.9, 103);
  }
  g.stroke();
  g.fillStyle = 'rgba(0,0,0,0.25)';
  g.fillRect(0, 101.5, W, 0.12);
  g.fillStyle = '#1f140c';
  g.fillRect(0, 103, W, 0.7);
  g.fillStyle = 'rgba(255,190,110,0.18)';
  g.fillRect(0, 103, W, 0.15);
  gr = g.createLinearGradient(0, 103.7, 0, Math.max(bottom, 104));
  gr.addColorStop(0, '#0f0906');
  gr.addColorStop(1, '#070403');
  g.fillStyle = gr;
  g.fillRect(0, 103.7, W, bottom - 103.7);
  for (let x = 0; x < W; x += 12) {
    g.fillStyle = 'rgba(255,190,120,0.035)';
    g.fillRect(x, 104.6, 0.2, bottom - 104.6);
  }
  if (bottom > 108.5) {
    // the brass foot rail along the counter
    gr = g.createLinearGradient(0, 105.6, 0, 106.6);
    gr.addColorStop(0, '#ffe2a0');
    gr.addColorStop(0.4, '#c8892c');
    gr.addColorStop(1, '#5a3a10');
    g.fillStyle = gr;
    g.fillRect(0, 105.6, W, 1);
    g.fillStyle = '#3a2610';
    g.fillRect(8, 106.6, 0.8, 2);
    g.fillRect(26, 106.6, 0.8, 2);
  }
  return c;
}

// Glasses and a bottle on the counter, passing close to the eye.
function foreground(S, height) {
  const sx = across(FG_PERIOD, S);
  const c = canvas(Math.round(FG_PERIOD * sx), height * S);
  const g = c.getContext('2d');
  g.scale(sx, S);
  const base = height + 0.6;
  const pint = (x, w, h, fill) => {
    const y = base - h;
    g.beginPath();
    g.moveTo(x, y);
    g.lineTo(x + w, y);
    g.lineTo(x + w * 0.88, base);
    g.lineTo(x + w * 0.12, base);
    g.closePath();
    g.fillStyle = 'rgba(30,18,8,0.92)';
    g.fill();
    const beer = g.createLinearGradient(x, 0, x + w, 0);
    beer.addColorStop(0, '#7a4a10');
    beer.addColorStop(0.3, '#c98a22');
    beer.addColorStop(1, '#5a3008');
    g.fillStyle = beer;
    g.beginPath();
    const fy = y + h * (1 - fill);
    g.moveTo(x + 0.3, fy);
    g.lineTo(x + w - 0.3, fy);
    g.lineTo(x + w * 0.88 - 0.2, base);
    g.lineTo(x + w * 0.12 + 0.2, base);
    g.closePath();
    g.fill();
    rr(g, x + 0.1, fy - 1.1, w - 0.2, 1.4, 0.6);
    g.fillStyle = '#f1e2c4';
    g.fill();
    g.fillStyle = 'rgba(255,240,210,0.4)';
    g.fillRect(x + 0.45, y + 0.6, 0.3, h - 1);
  };
  pint(12, 4.4, 7.2, 0.8);
  pint(17.6, 4.0, 6.4, 0.55);
  // bottle
  const bx = 52;
  g.fillStyle = 'rgba(40,22,8,0.95)';
  rr(g, bx, base - 6.5, 2.8, 6.5, 0.7);
  g.fill();
  g.fillRect(bx + 0.9, base - 9.5, 1, 3.3);
  g.fillStyle = '#d9a23a';
  g.fillRect(bx + 0.75, base - 9.9, 1.3, 0.6);
  g.fillStyle = 'rgba(255,225,170,0.35)';
  g.fillRect(bx + 0.4, base - 6, 0.26, 4.8);
  g.fillStyle = 'rgba(240,226,196,0.5)';
  g.fillRect(bx + 0.2, base - 4.4, 2.4, 1.6);
  return c;
}

/** Paints every picture the game needs for one size. */
export function paint(S, camTop, vh) {
  const bottom = camTop + vh;
  const fgHeight = Math.min(10, Math.max(0, bottom - 101.2));
  const crates = CRATE_COLORS.map((col) => [crate(S, col, false), crate(S, col, true)]);
  return {
    S,
    camTop,
    bottom,
    heli: heliBody(S),
    rotor: rotorDisc(S),
    tail: tailDisc(S),
    beam: beam(),
    crates,
    bottles: [bottles(S, 3), bottles(S, 8), bottles(S, 21)],
    shadow: stackShadow(),
    glow: glow(255, 196, 110),
    red: glow(255, 60, 40),
    spark: glow(255, 200, 110, 32),
    smoke: puff(38, 32, 28, 0.55),
    dust: puff(255, 236, 205, 0.35, 32),
    bubble: bubble(),
    banana: banana(),
    lampCone: lampCone(),
    wall: wall(S, camTop, vh),
    middle: middle(S, camTop, vh),
    ceiling: ceiling(S, camTop),
    floor: floor(S, bottom),
    fg: fgHeight > 4 ? foreground(S, fgHeight) : null,
    fgHeight,
  };
}

// A pendant lamp's light falling down the wall, top middle at the shade.
function lampCone() {
  const w = 128;
  const h = 256;
  const c = canvas(w, h);
  const g = c.getContext('2d');
  for (const [spread, a] of [[1, 0.25], [0.7, 0.3], [0.45, 0.35]]) {
    const gr = g.createLinearGradient(0, 0, 0, h);
    gr.addColorStop(0, `rgba(255,200,120,${a})`);
    gr.addColorStop(0.5, `rgba(255,180,90,${a * 0.35})`);
    gr.addColorStop(1, 'rgba(255,170,80,0)');
    g.fillStyle = gr;
    g.beginPath();
    g.moveTo(w / 2 - 6, 0);
    g.lineTo(w / 2 + 6, 0);
    g.lineTo(w / 2 + (w / 2) * spread, h);
    g.lineTo(w / 2 - (w / 2) * spread, h);
    g.closePath();
    g.fill();
  }
  return c;
}
