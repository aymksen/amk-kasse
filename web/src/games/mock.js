// A pretend till for the test page: the same answers the real one gives, kept in this browser only.
// The arcade takes either this or the real one as its `source`.

const KEY = 'amk-games-sandbox';
const LIMIT = 10;
export const TEST_PLAYER = { name: 'Testperson', room: '101' };

const DAY = 24 * 60 * 60 * 1000;
const STARTERS = [
  ['Lena', '112', 41], ['Jonas', '204', 33], ['Mehmet', '318', 27], ['Sophie', '007', 22], ['Paul', '415', 18],
  ['Clara', '226', 15], ['Finn', '301', 11], ['Aylin', '119', 9], ['Ben', '402', 6], ['Mia', '210', 4], ['Tom', '333', 2],
];

function load() {
  try {
    const saved = JSON.parse(localStorage.getItem(KEY));
    if (Array.isArray(saved)) return saved;
  } catch {}
  return null;
}

function store(runs) {
  try {
    localStorage.setItem(KEY, JSON.stringify(runs));
  } catch {}
}

function startingRuns(gameIds) {
  const now = Date.now();
  const runs = [];
  gameIds.forEach((game, g) => {
    STARTERS.forEach(([name, room, score], i) => {
      runs.push({ game, name, room, score: Math.round(score * (g ? 1.6 : 1)), at: now - (i + 1) * 2.5 * DAY });
    });
  });
  return runs;
}

const sameWorker = (a, b) => a.room === b.room && a.name.toLowerCase() === b.name.toLowerCase();

/** Each worker's best run in one game, best first; at a tie, whoever got there first. */
function ranking(runs, game) {
  const best = [];
  for (const run of runs) {
    if (run.game !== game) continue;
    const seen = best.find((b) => sameWorker(b, run));
    if (!seen) best.push({ ...run });
    else if (run.score > seen.score || (run.score === seen.score && run.at < seen.at)) Object.assign(seen, run);
  }
  return best.sort((a, b) => b.score - a.score || a.at - b.at).map(({ name, room, score, at }) => ({ name, room, score, at }));
}

function boardOf(runs, game, player) {
  const all = ranking(runs, game);
  const i = all.findIndex((row) => sameWorker(row, player));
  return { top: all.slice(0, LIMIT), mine: i < 0 ? null : { ...all[i], place: i + 1 } };
}

/**
 * @param {string[]} gameIds
 * @param {{ fail?: boolean, best?: number, record?: number, player?: {name: string, room: string} }} options
 */
export function mockSource(gameIds, { fail = false, best = 0, record = 0, player = TEST_PLAYER } = {}) {
  let runs = load() ?? startingRuns(gameIds);
  const now = Date.now();
  for (const game of gameIds) {
    if (record) runs.push({ game, name: 'Rekordhalter', room: '999', score: record, at: now - DAY });
    if (best) runs.push({ game, ...player, score: best, at: now - DAY });
  }
  store(runs);
  const wait = () => new Promise((done) => setTimeout(done, 350));

  return {
    player,
    async load() {
      await wait();
      return Object.fromEntries(gameIds.map((game) => [game, boardOf(runs, game, player)]));
    },
    async save(game, score) {
      await wait();
      if (fail) throw new Error('Die Kasse ist gerade nicht erreichbar (Test).');
      const before = boardOf(runs, game, player);
      if (score > 0) {
        runs.push({ game, ...player, score, at: Date.now() });
        store(runs);
      }
      const after = boardOf(runs, game, player);
      return {
        saved: score > 0,
        newBest: score > 0 && score > (before.mine?.score ?? 0),
        best: after.mine?.score ?? 0,
        place: score > 0 ? after.mine.place : null,
        record: after.top[0]?.score ?? 0,
        board: after,
      };
    },
    reset() {
      runs = startingRuns(gameIds);
      store(runs);
    },
  };
}
