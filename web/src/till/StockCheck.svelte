<script>
  import { onMount } from 'svelte';
  import ArrowRight from 'lucide-svelte/icons/arrow-right';
  import Check from 'lucide-svelte/icons/check';
  import Undo2 from 'lucide-svelte/icons/undo-2';
  import CountDialog from '../components/CountDialog.svelte';
  import LedNumber from '../components/LedNumber.svelte';
  import ProductThumb from '../components/ProductThumb.svelte';
  import Stepper from '../components/Stepper.svelte';
  import { time } from '../lib/format.js';
  import { getLanguage, t } from '../lib/i18n.svelte.js';
  import { listenForScans } from '../lib/scanner.js';
  import { findByCode, till, toast } from '../lib/state.svelte.js';

  // The check of the bar, the one way the till counts stock: at the start of a shift, from Zählen in the dock,
  // and at the end. Every product the bar counts is listed with what should be there, and the worker walks
  // past each spot and only changes what is wrong: with − and +, the arrow keys, or by tapping the figure.
  // A changed figure keeps what it was beside it, so nobody loses track of what they changed.
  // when: 'start', 'during' or 'end'. data is { last, since, lines } from /api/stock/check.
  // counts maps a product id to the figure the worker changed it to. It can live in the screen around this one,
  // so going back and forth loses nothing. ondone gets every line, changed or not, and how many changed.
  let { when, data, counts = $bindable({}), busy = false, ondone } = $props();

  // The figures come from the check, the names and pictures from the till's product list.
  const items = $derived.by(() => {
    const byId = new Map(till.products.map((p) => [p.id, p]));
    return data.lines
      .filter((line) => byId.has(line.productId))
      .map((line) => ({
        ...byId.get(line.productId),
        start: line.start,
        expected: Math.max(0, line.expected),
        sold: line.sold,
        added: line.added,
        corrected: line.corrected,
      }));
  });
  // Ordered by place, so the list follows the walk through the bar.
  const groups = $derived.by(() => {
    const byPlace = new Map();
    for (const item of items) {
      const place = item.location || t('count.other');
      if (!byPlace.has(place)) byPlace.set(place, []);
      byPlace.get(place).push(item);
    }
    return [...byPlace.entries()].map(([place, list]) => ({ place, list }));
  });
  const ordered = $derived(groups.flatMap((g) => g.list));

  const value = (item) => counts[item.id] ?? item.expected;
  // Each row shows how the bar stood (at the start of this shift, or as the last one left it) beside what it
  // should hold now. A new till has nothing to compare with.
  const showStart = $derived(data.since > 0);
  const startLabel = $derived(when === 'start' ? t('check.startLast') : t('check.startShift'));
  const changes = $derived(ordered.filter((item) => value(item) !== item.expected).length);

  const whenFormat = $derived(
    new Intl.DateTimeFormat(getLanguage() === 'de' ? 'de-DE' : 'en-GB', { weekday: 'long', day: 'numeric', month: 'long', hour: '2-digit', minute: '2-digit' }),
  );
  // Where the figures come from, so the worker knows what they are checking.
  const intro = $derived.by(() => {
    if (when !== 'start') return t('check.sinceStart', { time: time(data.since) });
    const last = data.last;
    if (!last) return t('check.first');
    const vars = { name: last.name, room: last.room, when: whenFormat.format(last.closedAt) };
    return last.counted ? t('check.fromLast', vars) : t('check.notCounted', vars);
  });

  const doneLabel = $derived.by(() => {
    if (changes === 0) return t('check.allRight');
    if (when === 'during') return changes === 1 ? t('check.saveChange') : t('check.saveChanges', { n: changes });
    return changes === 1 ? t('check.withChange') : t('check.withChanges', { n: changes });
  });

  function set(item, n) {
    n = Math.max(0, Math.min(9999, n));
    if (n === item.expected) delete counts[item.id];
    else counts[item.id] = n;
  }

  // Something happened to the product since: sold, filled in, or corrected by a count in between.
  const moved = (item) => item.sold > 0 || item.added > 0 || item.corrected !== 0;

  // What happened to a product since the figure was last right, under its name.
  function news(item) {
    const parts = [];
    if (item.sold > 0) parts.push(t('check.sold', { n: item.sold }));
    if (item.added > 0) parts.push(t('check.added', { n: item.added }));
    if (item.corrected !== 0) parts.push(t('check.corrected', { n: (item.corrected > 0 ? '+' : '−') + Math.abs(item.corrected) }));
    return parts.join(' · ');
  }

  let flash = $state(null);
  let typing = $state(null); // { item, first }

  function finish() {
    if (busy) return;
    ondone(
      ordered.map((item) => ({ productId: item.id, counted: value(item) })),
      changes,
    );
  }

  const focused = () => {
    const id = Number(document.activeElement?.getAttribute('data-count-figure'));
    return ordered.find((item) => item.id === id);
  };

  onMount(() =>
    listenForScans({
      // A scanned bottle jumps to its row, which helps on a long list.
      onScan(code) {
        if (typing) return;
        const found = findByCode(code);
        const item = found && ordered.find((i) => i.id === found.product.id);
        if (!item) {
          toast(t('sell.unknownCode', { code }), 'bad');
          return;
        }
        const figure = document.querySelector(`[data-count-figure="${item.id}"]`);
        figure?.focus();
        figure?.scrollIntoView({ block: 'nearest' });
        flash = item.id;
        setTimeout(() => flash === item.id && (flash = null), 1200);
      },
      // Typing on a focused figure asks for the whole number, starting with that digit.
      onKey(key) {
        const item = !typing && focused();
        if (!item) return;
        if (/\d/.test(key)) typing = { item, first: key };
        else if (key === 'Enter') typing = { item, first: '' };
      },
    }),
  );
</script>

<div class="check-screen">
  <div class="scroll">
    <div class="inner">
      <div class="intro">
        <h1>{t('check.heading')}</h1>
        <p class="muted">{intro} {t('check.howTo')}</p>
      </div>

      {#if ordered.length === 0}
        <p class="muted none">{t('check.none')}</p>
      {/if}

      {#each groups as group (group.place)}
        <section class="place">
          <h2><span>{group.place}</span><span class="muted small num">{group.list.length}</span></h2>
          <ul>
            {#each group.list as item (item.id)}
              {@const now = value(item)}
              {@const diff = now - item.expected}
              <li class="line" class:moved={moved(item)} class:changed={diff !== 0} class:flash={flash === item.id}>
                <ProductThumb product={item} size={44} />
                <div class="what">
                  <strong>{item.name}</strong>
                  {#if news(item)}<span class="news muted small">{news(item)}</span>{/if}
                </div>
                {#if showStart}
                  <div class="start" title={startLabel}>
                    <span class="muted">{startLabel}</span>
                    <LedNumber text={String(item.start)} digits={3} size={22} tone="dim" />
                  </div>
                  <ArrowRight size={18} class="muted" aria-hidden="true" />
                {/if}
                <Stepper
                  id={item.id}
                  value={now}
                  name={item.name}
                  tone={diff === 0 ? 'amber' : diff > 0 ? 'green' : 'red'}
                  onchange={(n) => set(item, n)}
                  ontype={() => (typing = { item, first: '' })}
                />
                <!-- What it was, right beside the new figure. A tap puts it back. The space stays when nothing
                     changed, so the figures line up from row to row. -->
                <div class="was-slot">
                  {#if diff !== 0}
                    <button class="was" onclick={() => set(item, item.expected)} title={t('check.reset', { n: item.expected })}>
                      <span class="old"><Undo2 size={14} /> {t('check.was', { n: item.expected })}</span>
                      <span class="diff num" class:up={diff > 0}>{diff > 0 ? '+' : '−'}{Math.abs(diff)}</span>
                    </button>
                  {/if}
                </div>
              </li>
            {/each}
          </ul>
        </section>
      {/each}
    </div>
  </div>

  <footer class="foot">
    <div class="inner spread">
      <span class="summary">
        {#if changes === 0}
          {t('check.allSame', { n: ordered.length })}
        {:else}
          {t('check.someChanged', { changed: changes, n: ordered.length })}
        {/if}
      </span>
      <button class="btn big primary" disabled={busy} onclick={finish}>
        <Check size={20} strokeWidth={2.6} />
        {doneLabel}
      </button>
    </div>
  </footer>
</div>

{#if typing}
  {@const item = typing.item}
  <CountDialog
    title={t('check.howMany', { name: item.name })}
    before={value(item)}
    first={typing.first}
    hint={item.crateSize > 1 ? t('check.crate', { n: item.crateSize }) : ''}
    onpick={(n) => {
      set(item, n);
      typing = null;
    }}
    onclose={() => (typing = null)}
  >
    <ProductThumb product={item} size={56} />
  </CountDialog>
{/if}

<style>
  .check-screen { display: grid; grid-template-rows: minmax(0, 1fr) auto; height: 100%; }
  .scroll { overflow-y: auto; overscroll-behavior: contain; }
  .inner { width: min(100%, 860px); margin: 0 auto; padding: 0 16px; }
  .intro { display: flex; flex-direction: column; gap: 6px; padding: 20px 0 8px; }
  .intro p { max-width: 62ch; margin: 0; }
  .none { padding: 24px 0; }
  .place { padding-bottom: 8px; }
  /* The spot stays on top while its products scroll past. */
  .place h2 {
    position: sticky;
    top: 0;
    z-index: 1;
    display: flex;
    align-items: baseline;
    gap: 8px;
    margin: 0;
    padding: 14px 4px 8px;
    background: var(--bg);
    font-size: 1rem;
    font-weight: 750;
  }
  ul { display: flex; flex-direction: column; gap: 6px; margin: 0; padding: 0; list-style: none; }
  .line {
    display: flex;
    align-items: center;
    gap: 12px;
    min-height: 64px;
    padding: 8px 8px 8px 10px;
    border-radius: var(--radius);
    background: var(--surface);
    box-shadow: 0 0 0 1px var(--line);
    transition: box-shadow 160ms ease, background 160ms ease;
  }
  /* A product that sold or was filled in gets a faint wash and its news in the accent, so what moved during
     the shift is easy to find. A row the worker changed gets a stronger one, with a bar on the left. */
  .line.moved { background: color-mix(in srgb, var(--accent) 6%, var(--surface)); }
  .line.moved .news { color: var(--accent-text); font-weight: 650; }
  .line.changed {
    background: color-mix(in srgb, var(--accent) 12%, var(--surface));
    box-shadow: 0 0 0 1px color-mix(in srgb, var(--accent) 35%, var(--line)), inset 4px 0 0 var(--accent);
  }
  .line.flash { box-shadow: 0 0 0 2px var(--accent); }
  .what { display: flex; flex: 1; flex-direction: column; align-items: flex-start; gap: 2px; min-width: 0; }
  .what strong { overflow: hidden; max-width: 100%; text-overflow: ellipsis; white-space: nowrap; }
  .start { display: flex; flex: none; align-items: center; gap: 8px; }
  .start span { font-size: 0.8125rem; font-weight: 600; }
  .start :global(.led) { padding: 7px 9px; }
  .was-slot { display: flex; flex: none; justify-content: flex-start; width: 92px; }
  .was {
    display: flex;
    flex: none;
    flex-direction: column;
    align-items: flex-start;
    gap: 3px;
    min-height: 48px;
    padding: 2px 8px;
    border: 0;
    border-radius: 6px;
    background: none;
    color: var(--muted);
    font-size: 0.8125rem;
    font-weight: 600;
    cursor: pointer;
  }
  .was:hover { background: var(--surface-2); color: var(--text); }
  .old { display: inline-flex; align-items: center; gap: 4px; white-space: nowrap; }
  .diff { padding: 1px 8px; border-radius: 5px; background: var(--bad-bg); color: var(--bad); font-size: 0.875rem; font-weight: 750; }
  .diff.up { background: var(--good-bg); color: var(--good); }
  .foot { padding: 12px 0; background: var(--surface); box-shadow: inset 0 1px 0 var(--line); }
  .summary { font-weight: 650; }
  .foot .btn { min-width: 240px; }
  /* Not much room: the word goes, the figure stays, and its tooltip still names it. */
  @media (max-width: 760px) {
    .start span { display: none; }
  }
  @media (max-width: 560px) {
    .line { flex-wrap: wrap; }
    .was-slot { width: auto; }
    .foot .btn { min-width: 0; }
  }
  @media (prefers-reduced-motion: reduce) {
    .line { transition: none; }
  }
</style>
