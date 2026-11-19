<script>
  import ReturnDialog from './ReturnDialog.svelte';
  import { dateTime, euro, time } from '../lib/format.js';
  import { t } from '../lib/i18n.svelte.js';

  // The first step of closing a shift, and only when something is out. It comes before the cash on purpose: a deposit
  // handed back after counting would make the count wrong. Nothing has to be back, an item can stay out overnight.
  let { items, onreturned, onnext } = $props();

  const HALF_A_DAY = 12 * 60 * 60 * 1000;

  let returning = $state(null);

  const deposits = $derived(items.reduce((sum, item) => sum + item.rental.depositCents, 0));
  const since = (ms) => (Date.now() - ms > HALF_A_DAY ? dateTime(ms) : time(ms));
  const details = ({ startedAt, depositCents, depositNote }) =>
    [
      t('rent.since', { time: since(startedAt) }),
      depositCents > 0 && t('rent.depositOnly', { amount: euro(depositCents) }),
      depositNote,
    ].filter(Boolean).join(' · ');

  function done() {
    returning = null;
    onreturned();
  }
</script>

<div class="lent">
  <div class="sheet">
    <h1>{t('close.lentTitle')}</h1>
    <p class="muted">{t('close.lentHint')}</p>
    <ul>
      {#each items as item (item.id)}
        <li class="card">
          <div class="who">
            <strong>{item.name}</strong>
            <span>{item.rental.renterName} · {t('common.room')} {item.rental.renterRoom}</span>
            <span class="muted small num">{details(item.rental)}</span>
          </div>
          <button class="btn" onclick={() => (returning = item)}>{t('rent.return')}</button>
        </li>
      {/each}
    </ul>
    {#if deposits}<p class="stays">{t('close.lentStays', { amount: euro(deposits) })}</p>{/if}
    <button class="btn big primary block" onclick={onnext}>{t('common.next')}</button>
  </div>
</div>

{#if returning}
  <ReturnDialog item={returning} onclose={() => (returning = null)} ondone={done} />
{/if}

<style>
  .lent { display: flex; justify-content: center; height: 100%; padding: 16px; overflow-y: auto; }
  .sheet { display: flex; flex-direction: column; gap: 12px; width: 100%; max-width: 620px; margin: auto; }
  ul { display: flex; flex-direction: column; gap: 10px; margin: 0; padding: 0; list-style: none; }
  li { display: flex; align-items: center; justify-content: space-between; gap: 12px; padding: 10px 12px 10px 16px; }
  .who { display: flex; flex-direction: column; gap: 2px; min-width: 0; }
  .who strong { font-size: 1.0625rem; font-stretch: 92%; font-weight: 700; }
  li .btn { flex: none; }
  .stays { color: var(--muted); font-weight: 600; }
</style>
