import heli from '../assets/sticker-heli.webp';
import flyer from './flyer/meta.js';

// The games behind the logo. The id is what the till stores with each score, so it never changes once scores
// exist. `picture` stands above the Play button in the games window. A game's code only loads when someone
// presses Play.
export const GAMES = [
  { id: 'flyer', ...flyer, picture: heli, game: () => import('./flyer/Game.svelte') },
];

/** Picks the German or English half of a [de, en] pair. */
export const pick = (pair, language) => (language === 'en' ? pair[1] : pair[0]);
