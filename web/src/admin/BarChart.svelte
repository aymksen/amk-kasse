<script>
  import { day, euro } from '../lib/format.js';
  import { t } from '../lib/i18n.svelte.js';

  // series: [{ day: 'YYYY-MM-DD', totalCents }]
  let { series, label } = $props();

  const HEIGHT = 220;
  const LEFT = 52;
  const RIGHT = 20;
  const TOP = 8;
  const BOTTOM = 26;
  const BAR_MAX = 24;
  const RADIUS = 4;

  let width = $state(640);
  let hovered = $state(null);
  let asTable = $state(false);

  // Rounds up to 1, 2 or 5 times a power of ten, so the axis shows clean numbers.
  function niceMax(cents) {
    const magnitude = 10 ** Math.floor(Math.log10(Math.max(cents, 100)));
    return [1, 2, 5, 10].map((m) => m * magnitude).find((v) => v >= cents);
  }

  const plotWidth = $derived(Math.max(width - LEFT - RIGHT, 0));
  const plotHeight = HEIGHT - TOP - BOTTOM;
  const baseline = TOP + plotHeight;
  const max = $derived(niceMax(Math.max(...series.map((s) => s.totalCents), 1)));
  const slot = $derived(plotWidth / series.length);
  const barWidth = $derived(Math.max(Math.min(slot - 2, BAR_MAX), 2));
  const ticks = $derived([0, max / 2, max]);

  const y = (cents) => baseline - (Math.max(cents, 0) / max) * plotHeight;
  const barX = (i) => LEFT + i * slot + (slot - barWidth) / 2;

  // Rounded at the data end, square at the baseline.
  function barPath(i, cents) {
    const x = barX(i);
    const top = y(cents);
    const r = Math.min(RADIUS, barWidth / 2, baseline - top);
    return `M${x} ${baseline}V${top + r}Q${x} ${top} ${x + r} ${top}H${x + barWidth - r}Q${x + barWidth} ${top} ${x + barWidth} ${top + r}V${baseline}Z`;
  }

  function move(e) {
    const box = e.currentTarget.getBoundingClientRect();
    const index = Math.floor((e.clientX - box.left - LEFT) / slot);
    hovered = index >= 0 && index < series.length ? index : null;
  }

  const shortDay = (iso) => iso.slice(8) + '.' + iso.slice(5, 7) + '.';
  const shortEuro = (cents) => `${Math.round(cents / 100)} €`;
  const tipLeft = $derived(hovered === null ? 0 : Math.min(barX(hovered) + barWidth / 2, width - 96));
</script>

<figure>
  <figcaption class="spread">
    <h2>{label}</h2>
    <button class="btn ghost small" onclick={() => (asTable = !asTable)}>{asTable ? t('ov.showChart') : t('ov.showTable')}</button>
  </figcaption>

  {#if asTable}
    <div class="table-wrap scroll">
      <table class="table">
        <thead><tr><th>{t('ov.day')}</th><th class="right">{t('ov.revenue')}</th></tr></thead>
        <tbody>
          {#each [...series].reverse() as row}
            <tr><td class="num">{day(row.day)}</td><td class="right num">{euro(row.totalCents)}</td></tr>
          {/each}
        </tbody>
      </table>
    </div>
  {:else}
    <div class="plot" bind:clientWidth={width}>
      <svg {width} height={HEIGHT} role="img" aria-label={label} onpointermove={move} onpointerleave={() => (hovered = null)}>
        {#each ticks as tick}
          <line x1={LEFT} x2={width - RIGHT} y1={y(tick)} y2={y(tick)} class="grid" />
          <text x={LEFT - 8} y={y(tick) + 4} text-anchor="end" class="axis">{shortEuro(tick)}</text>
        {/each}
        {#each series as row, i}
          {#if row.totalCents > 0}
            <path d={barPath(i, row.totalCents)} class="bar" class:active={hovered === i} />
          {/if}
          {#if i % 7 === (series.length - 1) % 7}
            <text x={barX(i) + barWidth / 2} y={HEIGHT - 8} text-anchor="middle" class="axis">{shortDay(row.day)}</text>
          {/if}
        {/each}
      </svg>
      {#if hovered !== null}
        <div class="tip" style:left="{tipLeft}px" style:top="{Math.min(y(series[hovered].totalCents), baseline - 56)}px">
          <span class="muted small">{day(series[hovered].day)}</span>
          <strong class="num">{euro(series[hovered].totalCents)}</strong>
        </div>
      {/if}
    </div>
  {/if}
</figure>

<style>
  figure { display: flex; flex-direction: column; gap: 8px; margin: 0; min-width: 0; }
  .plot { position: relative; height: 220px; }
  svg { display: block; touch-action: none; }
  .grid { stroke: var(--line); stroke-width: 1; }
  .axis { fill: var(--muted); font-size: 12px; font-variant-numeric: tabular-nums; }
  .bar { fill: var(--chart); }
  .bar.active { fill: var(--accent); }
  .tip {
    position: absolute;
    display: flex;
    flex-direction: column;
    padding: 6px 10px;
    border: var(--hair) solid var(--line);
    border-radius: 8px;
    background: var(--surface-3);
    transform: translate(-50%, calc(-100% - 8px));
    pointer-events: none;
    white-space: nowrap;
  }
  .scroll { max-height: 220px; overflow-y: auto; }
</style>
