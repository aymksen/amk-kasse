<script>
  import { onMount } from 'svelte';
  import Crown from 'lucide-svelte/icons/crown';
  import Trophy from 'lucide-svelte/icons/trophy';
  import X from 'lucide-svelte/icons/x';
  import LedNumber from '../components/LedNumber.svelte';
  import ProductThumb from '../components/ProductThumb.svelte';
  import SlideTabs from '../components/SlideTabs.svelte';
  import { time } from '../lib/format.js';
  import { getLanguage, t } from '../lib/i18n.svelte.js';
  import { nightName } from './recapText.js';

  // The month at one glance: who sold the most, with their rooms, what sold best and the biggest night. Pieces, names
  // and rooms only, because this is on the bar's screen. Any key closes it, so a scan is never lost behind it. In the
  // screensaver (`auto`) it stays for a moment by itself, and a tap falls through to the screensaver and ends it all.
  let { recap, auto = false, onclose } = $props();

  const OPEN_MS = 5 * 60 * 1000;
  const AUTO_MS = 18 * 1000;
  const ROWS_BELOW_PODIUM = 7;
  const PODIUM_ORDER = [2, 1, 3];

  const firstPeriod = () => (recap.month.qty || !recap.lastMonth.qty ? 'month' : 'lastMonth');
  let period = $state(firstPeriod());

  const language = getLanguage();
  const pieces = new Intl.NumberFormat('de-DE');
  const monthFormat = new Intl.DateTimeFormat(language === 'en' ? 'en-GB' : 'de-DE', { month: 'long', year: 'numeric' });

  const data = $derived(recap[period]);
  const month = $derived(monthFormat.format(new Date(`${data.from}T12:00:00`)));
  const below = $derived(data.workers.slice(3, 3 + ROWS_BELOW_PODIUM));
  const notShown = $derived(data.workers.length - 3 - below.length);
  const most = $derived(data.workers[0]?.qty ?? 1);
  const bestseller = $derived(data.products[0]?.name);
  const tabs = $derived([
    { value: 'month', label: t('recap.month') },
    { value: 'lastMonth', label: t('recap.lastMonth') },
  ]);

  function key(e) {
    if (['Tab', 'Shift', 'Control', 'Alt', 'Meta'].includes(e.key)) return;
    onclose();
  }

  onMount(() => {
    const timer = setTimeout(onclose, auto ? AUTO_MS : OPEN_MS);
    if (!auto) window.addEventListener('keydown', key);
    return () => {
      clearTimeout(timer);
      window.removeEventListener('keydown', key);
    };
  });
</script>

{#snippet ranking(title, items)}
  {#if items.length}
    <section class="card top-list">
      <h2>{title}</h2>
      <ol>
        {#each items as item, i (item.name)}
          <li>
            <span class="place num">{i + 1}</span>
            <ProductThumb product={item} size={32} />
            <span class="what">{item.name}</span>
            {#if item.name === bestseller}<span class="tag good">{t('recap.bestseller')}</span>{/if}
            <span class="num amount">{pieces.format(item.qty)}{t('recap.times')}</span>
          </li>
        {/each}
      </ol>
    </section>
  {/if}
{/snippet}

<div class="board" role="dialog" aria-label={t('board.title')}>
  <header>
    <Trophy size={30} />
    <div class="title">
      <h1>{t('board.title')}</h1>
      <p>{month}</p>
    </div>
    {#if !auto}
      <SlideTabs options={tabs} value={period} onpick={(id) => (period = id)} label={t('board.title')} />
      <button class="btn close" aria-label={t('common.close')} title={t('common.close')} onclick={onclose}><X size={22} /></button>
    {/if}
  </header>

  {#if data.qty}
    <div class="columns">
      <section class="crew">
        <div class="podium">
          {#each PODIUM_ORDER as place (place)}
            {@const worker = data.workers[place - 1]}
            <div class="stand place-{place}" style:--i={3 - place}>
              {#if worker}
                {#if place === 1}<Crown size={30} />{/if}
                <strong>{worker.name}</strong>
                <span class="room">{t('common.room')} {worker.room}</span>
                <div class="plinth">
                  <span class="rank num">{place}</span>
                  <span class="sold">
                    <LedNumber text={String(worker.qty)} digits={Math.max(String(worker.qty).length, 2)} size={place === 1 ? 44 : 32} plain />
                    <small>{t('recap.piecesShort')}</small>
                  </span>
                </div>
              {/if}
            </div>
          {/each}
        </div>

        {#if below.length}
          <ol class="rows">
            {#each below as worker, i (worker.room + worker.name)}
              <li>
                <span class="place num">{i + 4}</span>
                <span class="what">{worker.name}</span>
                <span class="room">{t('common.room')} {worker.room}</span>
                <span class="bar" aria-hidden="true"><i style:width="{(worker.qty / most) * 100}%"></i></span>
                <span class="num amount">{pieces.format(worker.qty)}</span>
              </li>
            {/each}
          </ol>
        {/if}
        {#if notShown > 0}<p class="more">{t('board.more', { n: notShown })}</p>{/if}
      </section>

      <div class="side">
        {@render ranking(t('board.drinks'), data.drinks)}
        {@render ranking(t('board.snacks'), data.snacks)}
        <div class="facts">
          {#if data.busiest}
            <section class="card fact">
              <span>{t('recap.busiest')}</span>
              <strong>{nightName(data.busiest.day, language)}</strong>
              <small class="num">{pieces.format(data.busiest.qty)} {t('recap.piecesShort')}</small>
            </section>
          {/if}
          {#if data.lastAt}
            <section class="card fact">
              <span>{t('recap.latest')}</span>
              <strong class="num">{time(data.lastAt)}</strong>
            </section>
          {/if}
        </div>
      </div>
    </div>
  {:else}
    <p class="none">{t('board.empty')}</p>
  {/if}
</div>

<style>
  .board {
    position: fixed;
    inset: 0;
    z-index: 55;
    display: flex;
    flex-direction: column;
    gap: 14px;
    padding: 14px 20px;
    overflow-y: auto;
    background: var(--wallpaper, var(--bg));
    color: var(--text);
    cursor: default;
    animation: fade 300ms ease-out;
  }
  header { display: flex; flex: none; align-items: center; gap: 12px; }
  header > :global(svg) { flex: none; color: var(--accent-text); }
  .title { flex: 1; min-width: 0; }
  h1 { border: 0; font-size: 1.5rem; font-stretch: 112%; font-weight: 800; line-height: 1.1; }
  .title p { color: var(--muted); font-weight: 600; }
  .close { flex: none; width: 46px; padding: 0; }

  .columns { display: grid; flex: 1; grid-template-columns: minmax(0, 1.25fr) minmax(0, 1fr); gap: 16px; min-height: 0; }
  .crew, .side { display: flex; flex-direction: column; gap: 12px; min-width: 0; min-height: 0; }
  .crew { justify-content: center; } /* early in a month there are only a few names */

  .podium { display: grid; flex: none; grid-template-columns: repeat(3, minmax(0, 1fr)); align-items: end; gap: 8px; }
  .stand {
    display: flex;
    flex-direction: column;
    align-items: center;
    min-width: 0;
    text-align: center;
    animation: rise 520ms cubic-bezier(0.2, 1.2, 0.4, 1) both;
    animation-delay: calc(var(--i) * 110ms);
  }
  .stand > :global(svg) { margin-bottom: 2px; color: var(--accent-text); }
  .stand strong {
    max-width: 100%;
    overflow: hidden;
    font-size: 1.1875rem;
    font-weight: 800;
    line-height: 1.2;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .place-1 strong { font-size: 1.5rem; }
  .room { color: var(--muted); font-size: 0.875rem; font-weight: 600; white-space: nowrap; }
  .plinth {
    --led-amber: var(--text);
    --led-amber-off: transparent;
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: flex-start;
    gap: 2px;
    width: 100%;
    margin-top: 6px;
    padding: 8px 6px 0;
    border-radius: var(--radius) var(--radius) 0 0;
    background: var(--surface-2);
    box-shadow: inset 0 0 0 1px var(--line);
  }
  .place-1 .plinth { height: 142px; background: var(--accent); box-shadow: none; color: var(--on-accent); --led-amber: var(--on-accent); }
  .place-2 .plinth { height: 104px; }
  .place-3 .plinth { height: 80px; }
  .rank { font-size: 1.5rem; font-weight: 800; line-height: 1; }
  .sold { display: inline-flex; align-items: baseline; gap: 5px; }
  .sold small { font-size: 0.8125rem; font-weight: 700; }

  ol { display: flex; flex-direction: column; margin: 0; padding: 0; list-style: none; }
  li { display: flex; align-items: center; gap: 8px; min-width: 0; }
  .place { flex: none; width: 1.5em; color: var(--muted); font-weight: 800; text-align: right; }
  .what { flex: 1 1 0; min-width: 0; overflow: hidden; font-weight: 700; text-overflow: ellipsis; white-space: nowrap; }
  .amount { flex: none; font-weight: 800; text-align: right; }

  .rows li { min-height: 36px; padding: 0 6px; border-bottom: 1px solid var(--line); }
  .rows li:last-child { border-bottom: 0; }
  .rows .room { flex: none; width: 5.6em; }
  .bar { flex: 0 0 22%; height: 8px; border-radius: 4px; background: var(--surface-2); overflow: hidden; }
  .bar i { display: block; height: 100%; border-radius: 4px; background: var(--accent); }
  .rows .amount { min-width: 2.6em; }
  .more { color: var(--muted); font-size: 0.875rem; text-align: center; }

  .top-list { flex: none; padding: 8px 12px; }
  h2 { margin-bottom: 4px; font-size: 1rem; font-weight: 800; }
  .top-list li { min-height: 38px; }
  .top-list .tag { flex: none; }

  .facts { display: grid; flex: none; grid-template-columns: minmax(0, 1.5fr) minmax(0, 1fr); gap: 12px; }
  .fact { display: flex; flex-direction: column; gap: 1px; padding: 8px 12px; }
  .fact span { color: var(--muted); font-size: 0.8125rem; font-weight: 700; }
  .fact strong { font-size: 1.0625rem; font-weight: 800; line-height: 1.2; }
  .fact small { color: var(--muted); font-size: 0.8125rem; font-weight: 600; }

  .none { margin: auto; color: var(--muted); font-size: 1.25rem; font-weight: 600; text-align: center; }

  /* A low wide screen shows fewer rows, an upright one stacks everything. */
  @media (max-height: 700px) {
    .place-1 .plinth { height: 112px; }
    .place-2 .plinth { height: 86px; }
    .place-3 .plinth { height: 70px; }
  }
  @media (orientation: portrait) {
    .columns { display: flex; flex-direction: column; }
  }
  /* It is read from across the bar, so a big screen shows the same board bigger, not emptier. */
  @media (min-width: 1500px) and (min-height: 900px) {
    .board { zoom: 1.35; }
  }
  @media (orientation: portrait) and (min-height: 1400px) {
    .board { zoom: 1.4; }
  }
  @media (prefers-reduced-motion: reduce) {
    .stand { animation: none; }
  }
  @keyframes rise {
    from { opacity: 0; transform: translateY(26px); }
  }
</style>
