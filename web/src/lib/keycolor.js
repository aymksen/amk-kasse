import { theme } from './theme.svelte.js';

const CREAM = [255, 246, 229];
const INK = [34, 28, 21];
const BLACK = [0, 0, 0];
const WHITE = [255, 255, 255];
const NAVY = [11, 18, 56];

function luminance(rgb) {
  const [r, g, b] = rgb.map((v) => {
    const c = v / 255;
    return c <= 0.04045 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4;
  });
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

function contrast(a, b) {
  const [high, low] = [luminance(a), luminance(b)].sort((x, y) => y - x);
  return (high + 0.05) / (low + 0.05);
}

const css = (rgb) => `rgb(${rgb.map(Math.round).join(' ')})`;
const glass = (rgb, alpha) => `rgb(${rgb.map(Math.round).join(' ')} / ${alpha})`;
const mix = (rgb, to, share) => rgb.map((c, i) => c + (to[i] - c) * share);

/** Pushes a colour away from grey, so a shade of it does not look dull. */
function vivid(rgb) {
  const mean = (rgb[0] + rgb[1] + rgb[2]) / 3;
  return rgb.map((c) => Math.min(Math.max(mean + (c - mean) * 1.3, 0), 255));
}

/** Turns a colour picked in the OBK area into a key face that the style's text stays readable on. */
export function keyColors(hex) {
  const value = parseInt(String(hex).replace('#', ''), 16) || 0x8b8375;
  const rgb = [value >> 16, (value >> 8) & 255, value & 255];

  if (theme.id === 'light') {
    let face = mix(rgb, WHITE, 0.6);
    while (contrast(face, INK) < 7) face = mix(face, WHITE, 0.15);
    return { face: css(face), hover: css(mix(face, WHITE, 0.3)), lip: css(rgb.map((c) => c * 0.75)) };
  }
  if (theme.id === 'pop') {
    let face = vivid(rgb);
    while (contrast(face, BLACK) < 7) face = mix(face, WHITE, 0.12);
    return { face: css(face), hover: css(mix(face, WHITE, 0.2)), lip: '#000' };
  }
  if (theme.id === 'glass' && theme.mode === 'light') {
    let face = mix(rgb, WHITE, 0.5);
    while (contrast(face, INK) < 7) face = mix(face, WHITE, 0.15);
    return { face: glass(face, 0.74), hover: glass(face, 0.9), lip: glass(rgb, 0.5) };
  }
  if (theme.id === 'glass') {
    return { face: glass(rgb, 0.3), hover: glass(rgb, 0.42), lip: glass(rgb, 0.6) };
  }
  if (theme.id === 'y2k') {
    let face = mix(vivid(rgb), WHITE, 0.25);
    while (contrast(face, NAVY) < 7) face = mix(face, WHITE, 0.12);
    return { face: css(face), hover: css(mix(face, WHITE, 0.25)), lip: css(mix(rgb, BLACK, 0.38)) };
  }

  let face = vivid(rgb).map((c) => c * 0.62);
  while (contrast(face, CREAM) < 4.5) face = face.map((c) => c * 0.92);
  return { face: css(face), hover: css(face.map((c) => c * 1.12)), lip: css(face.map((c) => c * 0.5)) };
}
