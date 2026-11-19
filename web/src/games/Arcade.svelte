<script>
  import { onMount, tick } from 'svelte';
  import { flip } from 'svelte/animate';
  import Gamepad from 'lucide-svelte/icons/gamepad-2';
  import Play from 'lucide-svelte/icons/play';
  import Trophy from 'lucide-svelte/icons/trophy';
  import X from 'lucide-svelte/icons/x';
  import { GAMES, pick } from './index.js';
  import { getLanguage, t } from '../lib/i18n.svelte.js';
  import { calmNow, moments } from '../lib/moments.svelte.js';
  import { listenForScans } from '../lib/scanner.js';

  // The games behind the logo, in a window over the till: the leaderboard and a Play button, and the game itself
  // plays inside the same window. The X in the top-right corner (or Escape) leaves the game for the leaderboard,
  // and on the leaderboard closes the window. A scan closes everything, and the sales screen underneath takes the
  // scan as it always does.
  // `source` is the till, or the test page's pretend till: { player, load(), save(game, score) }.
  // `from` is where the logo sits on the screen, so the window grows out of it and goes back into it.
  let { source, from = null, onclose } = $props();

  // Someone who walked off in the middle of a menu gets the till back by itself.
  const IDLE_MS = 4 * 60 * 1000;
  const FULL = 'inset(0px 0px 0px 0px round 0px)';
  const language = getLanguage();
  const dateFormat = new Intl.DateTimeFormat(language === 'en' ? 'en-GB' : 'de-DE', { day: '2-digit', month: '2-digit', year: '2-digit' });
  const pieces = new Intl.NumberFormat('de-DE');
  // One game for now. A second one needs a way to pick it in the window.
  const game = GAMES[0];
  const title = pick(game.title, language);
  const unit = pick(game.unit, language);
  const me = $derived(source.player);

  let boards = $state.raw(null); // what the window shows: { [game]: { top, mine } }
  let latest = $state.raw({}); // boards a run has changed, shown once the game is left
  let failed = $state(false);
  let playing = $state(false);
  let Game = $state.raw(null);
  let leaving = $state(false);
  let covered = $state(false); // the game fills the window, so the leaderboard can stop drawing
  let fresh = $state(false); // the board just took a new best, for a moment
  let shade = $state();
  let frame = $state();
  let stage = $state();
  let playButton = $state();
  let opening = $state(true); // a finger still tapping the logo must not land on the window while it grows
  let closing = false;
  let lastInput = Date.now();

  const board = $derived(boards?.[game.id]);
  const shown = $derived(latest[game.id] ?? board); // what the game is told while it runs
  const leader = $derived(board?.top[0]);

  const isMe = (row) => row.room === me.room && row.name.toLowerCase() === me.name.toLowerCase();
  const rowKey = (row) => row.room + '\u0000' + row.name.toLowerCase();
  // A clip that shows only `rect` of the box the game plays in.
  const clipTo = (rect, box) =>
    `inset(${rect.top - box.top}px ${box.right - rect.right}px ${box.bottom - rect.bottom}px ${rect.left - box.left}px round 14px)`;

  // The window shrunk onto the logo, or, without one, just a little smaller.
  function folded() {
    if (!from || !frame) return 'scale(0.96)';
    const r = frame.getBoundingClientRect();
    const scale = Math.max(0.05, Math.min(from.width / r.width, from.height / r.height));
    const dx = from.left + from.width / 2 - (r.left + r.width / 2);
    const dy = from.top + from.height / 2 - (r.top + r.height / 2);
    return `translate(${dx}px, ${dy}px) scale(${scale})`;
  }

  async function load() {
    failed = false;
    try {
      boards = await source.load();
    } catch {
      failed = true;
    }
  }

  async function start() {
    if (playing || closing) return;
    let module;
    try {
      module = await game.game();
    } catch {
      return;
    }
    if (closing) return;
    Game = module.default;
    playing = true;
    await tick();
    // Space or an arrow key must reach the game, never the button that was just pressed.
    if (document.activeElement instanceof HTMLElement) document.activeElement.blur();
    if (calmNow() || !playButton) {
      covered = true;
      return;
    }
    const grow = stage.animate([{ clipPath: clipTo(playButton.getBoundingClientRect(), stage.getBoundingClientRect()) }, { clipPath: FULL }], {
      duration: 560,
      easing: 'cubic-bezier(0.35, 0.65, 0.1, 1)',
    });
    grow.finished.then(() => playing && !leaving && (covered = true)).catch(() => {});
  }

  async function leave() {
    if (!playing || leaving) return;
    leaving = true;
    covered = false;
    if (!calmNow() && stage && playButton) {
      await tick();
      const into = clipTo(playButton.getBoundingClientRect(), stage.getBoundingClientRect());
      const shrink = stage.animate(
        [{ clipPath: FULL, opacity: 1 }, { clipPath: into, opacity: 1, offset: 0.75 }, { clipPath: into, opacity: 0 }],
        { duration: 480, easing: 'cubic-bezier(0.45, 0, 0.2, 1)', fill: 'forwards' },
      );
      await shrink.finished.catch(() => {});
    }
    playing = false;
    Game = null;
    leaving = false;
    showLatest();
  }

  // The new order slides into place where everyone can see it, rather than behind the game.
  function showLatest() {
    const next = latest[game.id];
    if (!next) return;
    const before = board?.mine?.score ?? 0;
    boards = { ...(boards ?? {}), ...latest };
    latest = {};
    if ((next.mine?.score ?? 0) > before) {
      fresh = true;
      setTimeout(() => (fresh = false), 2400);
    }
  }

  async function finish(score) {
    const { board: after, ...result } = await source.save(game.id, score);
    latest = { ...latest, [game.id]: after };
    if (!boards) load();
    return result;
  }

  async function close() {
    if (closing) return;
    closing = true;
    if (!calmNow() && frame) {
      const out = { duration: 280, easing: 'cubic-bezier(0.5, 0, 0.75, 0)', fill: 'forwards' };
      shade?.animate([{ opacity: 1 }, { opacity: 0 }], out);
      await frame.animate([{ transform: 'none', opacity: 1 }, { transform: folded(), opacity: 0 }], out).finished.catch(() => {});
    }
    onclose();
  }

  // The X and Escape: out of the game first, then out of the window.
  function back() {
    if (playing) leave();
    else close();
  }

  function key(e) {
    lastInput = Date.now();
    if (e.key !== 'Escape') return;
    e.preventDefault();
    back();
  }

  // A tap beside the window closes it, but never while a game runs, where a finger can easily slip.
  function besides(e) {
    if (e.target === e.currentTarget && !playing) close();
  }

  onMount(() => {
    load();
    moments.busy = true;
    if (calmNow()) setTimeout(() => (opening = false), 500);
    else {
      frame
        .animate([{ transform: folded(), opacity: 0 }, { opacity: 1, offset: 0.3 }, { transform: 'none', opacity: 1 }], {
          duration: 520,
          easing: 'cubic-bezier(0.2, 0.9, 0.25, 1)',
        })
        .finished.catch(() => {})
        .then(() => (opening = false));
    }
    const touched = () => (lastInput = Date.now());
    window.addEventListener('keydown', key);
    window.addEventListener('pointerdown', touched, true);
    const idle = setInterval(() => Date.now() - lastInput > IDLE_MS && close(), 15_000);
    const stopScans = listenForScans({ onScan: close });
    return () => {
      moments.busy = false;
      window.removeEventListener('keydown', key);
      window.removeEventListener('pointerdown', touched, true);
      clearInterval(idle);
      stopScans();
    };
  });
</script>

<div class="arcade">
  <div class="shade" bind:this={shade}></div>
  <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
  <div class="spot" class:opening onclick={besides}>
    <div class="arcade-window" bind:this={frame} role="dialog" aria-modal="true" aria-label={t('games.title')}>
      <header class="bar">
        <span class="badge"><Gamepad size={18} /></span>
        <strong class="where">{playing ? title : t('games.title')}</strong>
        <span class="who">{t('games.player', { name: me.name, room: me.room })}</span>
        <button class="x" aria-label={playing ? t('games.toBoard') : t('common.close')} title={playing ? t('games.toBoard') : t('common.close')} onclick={back}>
          <X size={24} strokeWidth={2.6} />
        </button>
      </header>

      <div class="body">
        <div class="home" class:behind={playing && !leaving} class:covered inert={playing ? true : undefined}>
          <section class="intro">
            {#if game.picture}<img class="picture" src={game.picture} alt="" draggable="false" />{/if}
            <div class="words">
              <h1>{title}</h1>
              <p>{pick(game.blurb, language)}</p>
            </div>
            <dl class="figures">
              <div class="figure" class:fresh>
                <dt>{t('games.yourBest')}</dt>
                {#if board?.mine}
                  <dd><strong class="num">{pieces.format(board.mine.score)}</strong> <span class="tag">{t('games.place', { place: board.mine.place })}</span></dd>
                {:else}
                  <dd class="none">{board ? t('games.noRun') : '…'}</dd>
                {/if}
              </div>
              <div class="figure">
                <dt>{t('games.record')}</dt>
                {#if leader}
                  <dd><strong class="num">{pieces.format(leader.score)}</strong> <span class="holder">{t('games.recordBy', { name: leader.name, room: leader.room })}</span></dd>
                {:else}
                  <dd class="none">{board ? t('games.noRecord') : '…'}</dd>
                {/if}
              </div>
            </dl>
            <button class="btn primary play" bind:this={playButton} onclick={start}>
              <Play size={30} fill="currentColor" />{t('games.play')}
            </button>
          </section>

          <section class="ranking">
            <h2><Trophy size={18} />{t('games.leaderboard')}</h2>
            <table class="card board" aria-label={t('games.board', { game: title })}>
              <thead>
                <tr>
                  <th class="place">{t('games.colPlace')}</th>
                  <th class="who-col">{t('games.colName')}</th>
                  <th>{t('games.colRoom')}</th>
                  <th>{t('games.colDate')}</th>
                  <th class="score">{unit}</th>
                </tr>
              </thead>
              <tbody>
                {#if failed && !board}
                  <tr>
                    <td colspan="5" class="note">
                      {t('games.unreachable')}
                      <button class="btn ghost" onclick={load}>{t('games.retry')}</button>
                    </td>
                  </tr>
                {:else if !board}
                  {#each [0, 1, 2, 3, 4] as n (n)}
                    <tr class="waiting"><td colspan="5"><i></i></td></tr>
                  {/each}
                {:else if !board.top.length}
                  <tr><td colspan="5" class="note">{t('games.empty')}</td></tr>
                {:else}
                  {#each board.top as row, n (rowKey(row))}
                    <tr class="p{n + 1}" class:me={isMe(row)} class:fresh={fresh && isMe(row)} style:--n={n} animate:flip={{ duration: 520 }}>
                      <td class="place"><span class="medal num">{n + 1}</span></td>
                      <td class="who-col">{row.name}</td>
                      <td class="num quiet">{row.room}</td>
                      <td class="num quiet">{dateFormat.format(row.at)}</td>
                      <td class="score num">{pieces.format(row.score)}</td>
                    </tr>
                  {/each}
                  {#if board.mine && board.mine.place > board.top.length}
                    <tr class="me below" style:--n={board.top.length}>
                      <td class="place"><span class="medal num">{board.mine.place}</span></td>
                      <td class="who-col">{board.mine.name}</td>
                      <td class="num quiet">{board.mine.room}</td>
                      <td class="num quiet">{dateFormat.format(board.mine.at)}</td>
                      <td class="score num">{pieces.format(board.mine.score)}</td>
                    </tr>
                  {/if}
                {/if}
              </tbody>
            </table>
          </section>
        </div>

        {#if playing && Game}
          <div class="stage" bind:this={stage}>
            <Game player={me} best={shown?.mine?.score ?? 0} record={shown?.top[0]?.score ?? 0} onfinish={finish} onquit={leave} />
          </div>
        {/if}
      </div>
    </div>
  </div>
</div>

<style>
  .arcade { position: fixed; inset: 0; z-index: 56; color: var(--text); cursor: default; }
  .shade { position: absolute; inset: 0; background: rgb(6 4 2 / 0.62); animation: dim 260ms ease-out; }
  .spot { position: absolute; inset: 0; display: grid; place-items: center; padding: 24px; }
  .spot.opening { pointer-events: none; }

  /* The same size for the leaderboard and the game, so going in and out never jumps. */
  .arcade-window {
    display: flex;
    flex-direction: column;
    width: min(1560px, 100%);
    height: min(940px, 100%);
    overflow: hidden;
    border-radius: 16px;
    background: var(--bg);
    box-shadow: 0 0 0 1px var(--line), 0 30px 80px -20px rgb(0 0 0 / 0.75);
  }

  /* The title bar: the X sits in the top-right corner, like on any window, and never moves or hides while a game runs. */
  .bar {
    display: flex;
    flex: none;
    align-items: center;
    gap: 0.625em;
    min-height: 60px;
    padding: 6px 6px 6px 1em;
    background: var(--surface);
    box-shadow: inset 0 -1px 0 var(--line);
  }
  .x {
    display: grid;
    flex: none;
    place-items: center;
    width: 52px;
    height: 48px;
    border: 0;
    border-radius: 10px;
    background: var(--surface-2);
    color: var(--text);
    cursor: pointer;
    touch-action: manipulation;
    transition: background 120ms ease, transform 120ms ease;
  }
  @media (hover: hover) {
    .x:hover { background: var(--bad); color: #fff; }
  }
  .x:active { transform: scale(0.92); }
  .x:focus-visible { outline: 3px solid var(--focus); outline-offset: 2px; }
  .badge { display: grid; flex: none; place-items: center; width: 30px; height: 30px; border-radius: 8px; background: var(--accent); color: var(--on-accent); }
  .where { min-width: 0; overflow: hidden; font-family: var(--display); font-size: 1.25em; font-weight: var(--display-weight); text-overflow: ellipsis; white-space: nowrap; }
  .who { flex: 1; min-width: 0; overflow: hidden; color: var(--muted); font-weight: 600; text-align: right; text-overflow: ellipsis; white-space: nowrap; }

  .body { position: relative; flex: 1; min-height: 0; }
  .home {
    position: absolute;
    inset: 0;
    display: grid;
    grid-template-columns: minmax(0, 0.85fr) minmax(0, 1.15fr);
    gap: 1.5em;
    padding: 1.25em 1.5em 1.5em;
    overflow-x: hidden;
    overflow-y: auto;
    transform-origin: 30% 60%;
    transition: transform 520ms cubic-bezier(0.2, 0.85, 0.15, 1), opacity 520ms ease;
  }
  .home.behind { transform: scale(0.95); opacity: 0.25; }
  .home.covered { visibility: hidden; }

  /* The game's side: its picture, what it is, the two numbers that matter and the big button. */
  .intro {
    display: flex;
    flex-direction: column;
    justify-content: center;
    gap: 1em;
    min-width: 0;
    animation: rise 560ms 120ms cubic-bezier(0.2, 1.1, 0.35, 1) both;
  }
  .picture {
    flex: 1 1 0;
    width: 100%;
    min-height: 5em;
    max-height: 15em;
    object-fit: contain;
    filter: drop-shadow(0 14px 18px rgb(0 0 0 / 0.45));
    user-select: none;
    animation: hover 3.6s ease-in-out infinite;
  }
  h1 { border: 0; font-family: var(--display); font-size: 2.25em; font-weight: var(--display-weight); line-height: 1; }
  .words p { margin-top: 0.35em; color: var(--muted); font-weight: 600; }
  .figures { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0.75em; }
  .figure { display: flex; flex-direction: column; gap: 0.25em; min-width: 0; padding: 0.75em 0.9em; border-radius: 12px; background: var(--surface); box-shadow: 0 0 0 1px var(--line); }
  dt { color: var(--muted); font-size: 0.75em; font-weight: 800; letter-spacing: 0.05em; text-transform: uppercase; }
  dd { display: flex; flex-wrap: wrap; align-items: baseline; gap: 0.15em 0.5em; min-width: 0; margin: 0; }
  dd strong { font-size: 1.75em; font-weight: 800; line-height: 1.1; }
  dd.none { color: var(--muted); font-weight: 600; }
  .holder { min-width: 0; overflow: hidden; color: var(--muted); font-weight: 600; text-overflow: ellipsis; white-space: nowrap; }
  .figure.fresh strong { animation: pop 700ms cubic-bezier(0.3, 1.8, 0.5, 1); color: var(--accent-text); }
  .play {
    flex: none;
    gap: 0.4em;
    width: 100%;
    min-height: 84px;
    border-radius: 14px;
    font-size: 1.625em;
    font-weight: 800;
    touch-action: manipulation;
    animation: breathe 2.6s 900ms ease-in-out infinite;
  }
  .play:active:not(:disabled) { animation: none; }

  .ranking { display: flex; flex-direction: column; gap: 0.5em; min-width: 0; animation: rise 560ms 220ms cubic-bezier(0.2, 1.1, 0.35, 1) both; }
  h2 { display: flex; align-items: center; gap: 0.4em; border: 0; color: var(--muted); font-size: 0.8125em; font-weight: 800; letter-spacing: 0.06em; text-transform: uppercase; }
  .board { width: 100%; border-collapse: separate; border-spacing: 0; overflow: hidden; table-layout: auto; }
  th {
    padding: 0.6em 0.75em;
    border-bottom: 1px solid var(--line);
    color: var(--muted);
    font-size: 0.75em;
    font-weight: 800;
    letter-spacing: 0.05em;
    text-align: left;
    text-transform: uppercase;
    white-space: nowrap;
  }
  td { height: 2.5em; padding: 0 0.75em; border-bottom: 1px solid var(--line); white-space: nowrap; }
  tbody tr:last-child td { border-bottom: 0; }
  tbody tr { animation: row 420ms ease-out both; animation-delay: calc(300ms + var(--n, 0) * 40ms); }
  .place { width: 1%; }
  .who-col { width: 100%; max-width: 0; overflow: hidden; font-weight: 700; text-overflow: ellipsis; }
  .score { width: 1%; text-align: right; }
  td.score { font-size: 1.0625em; font-weight: 800; }
  .quiet { color: var(--muted); }
  .medal {
    display: inline-grid;
    place-items: center;
    min-width: 1.9em;
    height: 1.9em;
    padding: 0 0.3em;
    border-radius: 999px;
    color: var(--muted);
    font-weight: 800;
  }
  .p1 .medal { background: var(--accent); color: var(--on-accent); }
  .p2 .medal { background: #cfc8bb; color: #1d1a16; }
  .p3 .medal { background: #c98a52; color: #1d1a16; }
  .p1 td { height: 2.9em; }
  .p1 .who-col, .p1 .score { font-size: 1.125em; }
  tr.me td { background: var(--warn-bg); }
  tr.below td { border-top: 2px dashed var(--line); }
  tr.fresh td { animation: glow 2.2s ease-out; }
  .note { height: auto; padding: 1.5em 1em; color: var(--muted); font-weight: 600; text-align: center; white-space: normal; }
  .note .btn { margin-left: 0.5em; }
  .waiting i {
    display: block;
    height: 0.75em;
    border-radius: 6px;
    background: linear-gradient(90deg, var(--surface-2) 0%, var(--line) 50%, var(--surface-2) 100%) 0 0 / 200% 100%;
    animation: shimmer 1.2s linear infinite;
  }

  .stage { position: absolute; inset: 0; z-index: 1; overflow: hidden; background: var(--bg); }

  @media (max-height: 820px) and (orientation: landscape) {
    .spot { padding: 16px; }
    td { height: 2.2em; }
    .p1 td { height: 2.5em; }
    .home { padding: 1em 1.25em 1.25em; }
  }
  /* Upright screen: a tall window, the game's side on top with its picture beside the words, the board below. */
  @media (orientation: portrait), (max-width: 760px) {
    .arcade-window { height: min(1700px, 100%); }
    .home { grid-template-columns: minmax(0, 1fr); grid-template-rows: auto auto; align-content: center; row-gap: 2em; }
    .intro { display: grid; grid-template-columns: minmax(0, 0.8fr) minmax(0, 1fr); grid-template-areas: 'picture words' 'figures figures' 'play play'; align-items: center; }
    .picture { grid-area: picture; max-height: 13em; }
    .words { grid-area: words; }
    .figures { grid-area: figures; }
    .play { grid-area: play; }
  }
  /* It is read from across the bar, so a big screen shows the same window bigger, not emptier. */
  @media (min-width: 1500px) and (min-height: 900px) {
    .arcade-window { font-size: 1.15rem; }
  }
  @media (orientation: portrait) and (min-height: 1400px) {
    .arcade-window { font-size: 1.35rem; }
  }

  @media (prefers-reduced-motion: reduce) {
    .shade, .intro, .ranking, .picture, .play, tbody tr, tr.fresh td, .figure.fresh strong, .waiting i { animation: none; }
    .home, .x { transition: none; }
  }

  @keyframes dim {
    from { opacity: 0; }
  }
  @keyframes rise {
    from { opacity: 0; transform: translateY(18px); }
  }
  @keyframes row {
    from { opacity: 0; transform: translateX(-10px); }
  }
  @keyframes hover {
    50% { transform: translateY(-8px) rotate(-1.5deg); }
  }
  @keyframes breathe {
    50% { transform: scale(1.025); }
  }
  @keyframes pop {
    40% { transform: scale(1.4); }
  }
  @keyframes glow {
    0%, 30% { background: var(--accent); color: var(--on-accent); }
  }
  @keyframes shimmer {
    to { background-position: -200% 0; }
  }
</style>
