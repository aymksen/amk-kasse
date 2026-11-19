<script>
  import { onMount } from 'svelte';
  import Arcade from './Arcade.svelte';
  import { GAMES, pick } from './index.js';
  import { mockSource } from './mock.js';
  import { getLanguage } from '../lib/i18n.svelte.js';

  // The test page (games.html). ?game=<id> plays one game alone, and without it the page shows the games window
  // as the till does, over a pretend leaderboard, growing out of where the logo sits in the dock. Closing it leads
  // to a short list of links.
  const address = new URLSearchParams(location.search);
  const gameId = address.get('game');
  const number = (name) => Number(address.get(name)) || 0;
  const source = mockSource(GAMES.map((g) => g.id), { fail: address.has('fail'), best: number('best'), record: number('record') });
  const language = getLanguage();

  const game = GAMES.find((g) => g.id === gameId);
  // Where the logo sits in the till's dock, bottom left.
  const logo = () => new DOMRect(14, innerHeight - 58, 170, 48);

  let best = $state(0);
  let record = $state(0);
  let round = $state(0);
  let quit = $state(false);
  let arcade = $state(true);

  async function refresh() {
    const boards = await source.load();
    best = boards[gameId]?.mine?.score ?? 0;
    record = boards[gameId]?.top[0]?.score ?? 0;
  }

  async function finish(score) {
    const result = await source.save(gameId, score);
    best = result.best;
    record = result.record;
    const { board, ...rest } = result;
    console.info('[sandbox] saved', score, rest);
    return rest;
  }

  function again() {
    quit = false;
    round++;
  }

  onMount(() => {
    if (!game) return;
    refresh();
    const escape = (e) => {
      if (e.key === 'Escape') quit = true;
    };
    window.addEventListener('keydown', escape);
    return () => window.removeEventListener('keydown', escape);
  });
</script>

{#if game}
  <div class="stage">
    {#if quit}
      <div class="quit">
        <p>Zurück zum Spielebildschirm (onquit oder Escape).</p>
        <button class="btn primary" onclick={again}>Wieder ins Spiel</button>
      </div>
    {:else}
      {#key round}
        {#await game.game() then { default: Game }}
          <Game player={source.player} {best} {record} onfinish={finish} onquit={() => (quit = true)} />
        {/await}
      {/key}
    {/if}
  </div>
{:else if arcade}
  <Arcade {source} from={logo()} onclose={() => (arcade = false)} />
{:else}
  <nav class="index">
    <h1>AMK Spiele – Testseite</h1>
    {#each GAMES as g (g.id)}
      <p>
        <strong>{pick(g.title, language)}</strong> – {pick(g.blurb, language)}
        <a href="?game={g.id}">spielen</a>
      </p>
    {/each}
    <button class="btn primary" onclick={() => (arcade = true)}>Spielefenster öffnen</button>
    <button class="btn" onclick={() => source.reset()}>Testbestenliste zurücksetzen</button>
  </nav>
{/if}

<style>
  .stage { position: fixed; inset: 0; overflow: hidden; background: var(--bg); }
  .quit { display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 16px; height: 100%; }
  .index { display: flex; flex-direction: column; gap: 12px; max-width: 720px; padding: 24px; }
  .index a { color: var(--accent-text); }
</style>
