<script>
  import { onMount } from 'svelte';
  import LedNumber from '../components/LedNumber.svelte';
  import { t } from '../lib/i18n.svelte.js';
  import { clock, moments } from '../lib/moments.svelte.js';
  import { seasonAt } from '../lib/seasons.js';

  // What the empty receipt shows at this time of year. It is gone with the first item, so it is never in the way.
  const season = $derived(moments.season);
  const advent = $derived(season?.kind === 'christmas' ? season.advent : null);
  const halloween = $derived(season?.kind === 'halloween' ? season : null);
  const toMidnight = $derived(season?.kind === 'newyear' && (season.phase === 'eve' || season.phase === 'final'));

  let left = $state('');

  function hms(ms) {
    const seconds = Math.ceil(ms / 1000);
    const pad = (n) => String(n).padStart(2, '0');
    return `${pad(Math.floor(seconds / 3600))}:${pad(Math.floor(seconds / 60) % 60)}:${pad(seconds % 60)}`;
  }

  onMount(() => {
    const tick = () => {
      const now = seasonAt(clock());
      if (now?.kind === 'newyear' && now.left > 0) left = hms(now.left);
    };
    tick();
    const timer = setInterval(tick, 250);
    return () => clearInterval(timer);
  });
</script>

{#if advent}
  <div class="slot">
    {#if advent.daysLeft > 0}
      <strong>{t('moments.advent', { n: advent.sunday })}</strong>
      <span>{t(advent.daysLeft === 1 ? 'moments.oneDay' : 'moments.days', { n: advent.daysLeft })}</span>
    {:else}
      <strong>{t('moments.merry')}</strong>
    {/if}
  </div>
{:else if halloween}
  <div class="slot">
    {#if halloween.daysLeft > 0}
      <span>{t(halloween.daysLeft === 1 ? 'moments.halloweenOne' : 'moments.halloweenDays', { n: halloween.daysLeft })}</span>
    {:else}
      <strong>{t('moments.halloween')}</strong>
    {/if}
  </div>
{:else if toMidnight}
  <div class="slot">
    <span>{t('moments.until', { year: season.year })}</span>
    <div class="clock"><LedNumber text={left} digits={6} size={34} /></div>
  </div>
{:else if season?.kind === 'newyear'}
  <div class="slot"><strong>{t('moments.happy', { year: season.year })}</strong></div>
{/if}

<style>
  .slot {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;
    padding: 0 6px 12px;
    color: var(--ink-soft);
    font-size: 0.875rem;
    line-height: 1.3;
    text-align: center;
    animation: fade 400ms ease-out;
  }
  strong { color: var(--ink); font-size: 1rem; }
  .clock { margin-top: 6px; }
</style>
