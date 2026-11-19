// The bar's black helicopter from the lending sticker, seen from the side and facing right, with the monkey from the
// logo at the window. Path data is in heli units: the middle of the cabin is 0,0 and ten heli units are one unit of
// the game world. The game turns the paths into Path2D and paints them once into sprites (art.js).

export const HELI = {
  finLow: 'M-104 1 L-113 13 L-106 14 L-98 1 Z',
  boom: 'M-36 -13 L-110 -6 L-113 0 L-110 3 L-36 11 Z',
  stab: 'M-90 -1 L-74 -1 L-72 2 L-90 3 Z',
  fin: 'M-101 -3 L-117 -31 Q-118 -33 -115 -33 L-108 -33 Q-106 -33 -105 -31 L-92 -4 Z',
  hump: 'M-30 -22 C-28 -33 14 -35 20 -23 Z',
  mast: 'M-6 -39 L2 -39 L3 -30 L-7 -30 Z',
  cabin: 'M-44 -8 C-40 -20 -22 -25 2 -25 C27 -25 45 -18 52 -6 C58 5 55 18 42 22 L-26 22 C-37 22 -45 14 -44 4 Z',
  windshield: 'M19 -21 C34 -20 45 -13 50 -5 C53 1 52 4 49 6 L19 6 Z',
  window: 'M-15 -20 L9 -20 Q14 -20 14 -15 L14 2 Q14 7 9 7 L-15 7 Q-20 7 -20 2 L-20 -15 Q-20 -20 -15 -20 Z',
  stripe: 'M-116 -1 L-40 7 L60 7 L60 12 L-40 12 L-116 3 Z',
  door: 'M-24 -19 L-24 19 M16 -21 L16 19',
  struts: 'M-18 21 L-22 33 M30 21 L34 33',
  skid: 'M-36 33 L46 33 Q54 33 55 26',
};

// The parts of the body that make its outline, back to front.
export const BODY = ['finLow', 'boom', 'stab', 'fin', 'hump', 'mast', 'cabin'];

export const HUB = [-2, -40];
export const ROTOR_HALF = 70;
export const TAIL = [-110, -19];
export const TAIL_R = 11;
export const LAMP = [47, 15];
export const BEACON = [-112, -33];

// The box the body picture covers, without the spinning rotors: x, y, width, height.
export const BODY_BOX = [-124, -46, 186, 86];

export const BODY_TOP = '#5b534b';
export const BODY_MID = '#2e2a26';
export const BODY_LOW = '#141210';
export const STRIPE = '#ffb224';

// Beer crates come in these colours.
export const CRATE_COLORS = ['#a9561d', '#2f6b3f', '#9b2b22', '#2b4b78', '#c48a17'];

// The monkey at the side window, drawn inside the window's outline: ellipses [x, y, rx, ry] and paths.
export const MONKEY = [
  { e: [-5, 10, 10, 6], fill: '#2e1b0f' },
  { e: [-11.5, -3, 3.2, 3.5], fill: '#3a2313' },
  { e: [-11.5, -3, 1.6, 1.9], fill: '#b9805a' },
  { e: [-3, -2, 8, 7.6], fill: '#3a2313' },
  { e: [0.6, -0.6, 6, 5.6], fill: '#e3ad84' },
  { e: [3.2, 2, 4.2, 2.8], fill: '#f0c7a0' },
  { e: [-0.4, -3.4, 1.15, 1.4], fill: '#1a0f08' },
  { e: [4.4, -3.4, 1.05, 1.35], fill: '#1a0f08' },
  { e: [-0.1, -3.9, 0.4, 0.45], fill: '#ffffff' },
  { e: [4.8, -3.9, 0.36, 0.42], fill: '#ffffff' },
  { p: 'M1.4 3.4 Q4 5.8 6.6 3.2', stroke: '#5a2e18', w: 0.9 },
  { p: 'M-7 -8 C-9.5 -13 -6.5 -19 0.5 -21 C0 -18 -1.5 -13 -2 -8.5 Z', fill: '#f6c41f' },
  { p: 'M-6.2 -9 C-8 -13 -6 -17 -1.5 -19.4', stroke: '#fff1a6', w: 0.7 },
  { p: 'M0.5 -21 L2.2 -22.6 L2.8 -21.8 L1.2 -20.2 Z', fill: '#4a3410' },
];
