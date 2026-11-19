import { get, post } from '../lib/api.js';
import { till } from '../lib/state.svelte.js';

/** The games screen's way to the till. The server takes room and name from the open shift, not from here. */
export const tillSource = () => ({
  player: { name: till.shift.workerName, room: till.shift.workerRoom },
  load: () => get('/api/games'),
  save: (game, score) => post(`/api/games/${game}/scores`, { score }),
});
