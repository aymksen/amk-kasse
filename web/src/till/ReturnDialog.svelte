<script>
  import { onMount } from 'svelte';
  import Modal from '../components/Modal.svelte';
  import { get, post } from '../lib/api.js';
  import { euro } from '../lib/format.js';
  import { t } from '../lib/i18n.svelte.js';
  import { refresh, till, toastError } from '../lib/state.svelte.js';

  let { item, onclose, ondone } = $props();

  let quote = $state(null);
  let busy = $state(false);

  const net = $derived(quote ? quote.feeCents - quote.depositCents : 0);
  const other = $derived(till.settings.otherPayment);

  async function finish(payment) {
    if (busy) return;
    busy = true;
    try {
      await post(`/api/rentals/${item.rental.id}/return`, { payment });
      await refresh();
      ondone();
    } catch (err) {
      toastError(err);
    } finally {
      busy = false;
    }
  }

  onMount(async () => {
    try {
      quote = await get(`/api/rentals/${item.rental.id}/quote`);
    } catch (err) {
      toastError(err);
      onclose();
    }
  });
</script>

<Modal title={t('rent.returnTitle', { item: item.name })} {onclose}>
  {#if quote}
    <div class="stack">
      <p>{item.rental.renterName} · {t('common.room')} {item.rental.renterRoom}</p>
      <dl>
        <dt>{t('rent.duration')}</dt>
        <dd class="num">{t('rent.minutes', { n: quote.minutes })}</dd>
        {#if quote.feeCents > 0}
          <dt>{t('rent.fee')}</dt>
          <dd class="num">{euro(quote.feeCents)}</dd>
        {/if}
        <dt>{t('rent.depositBack')}</dt>
        <dd class="num">{euro(quote.depositCents)}</dd>
      </dl>
      {#if item.rental.depositNote}
        <p class="note">{t('rent.itemBack', { note: item.rental.depositNote })}</p>
      {/if}
      <div class="net">
        <span class="muted">{net > 0 ? t('rent.guestPays') : net < 0 ? t('rent.youPay') : t('rent.even')}</span>
        {#if net !== 0}<strong class="num">{euro(Math.abs(net))}</strong>{/if}
      </div>
      <button class="btn big primary block" data-autofocus disabled={busy} onclick={() => finish('cash')}>
        {t('rent.return')}{net > 0 ? ` · ${t('common.cash')}` : ''}
      </button>
      {#if quote.feeCents > 0 && other}
        <button class="btn big block" disabled={busy} onclick={() => finish('other')}>
          {quote.depositCents > 0 ? t('rent.feeOtherBack', { name: other, amount: euro(quote.depositCents) }) : t('rent.feeOther', { name: other })}
        </button>
      {/if}
    </div>
  {:else}
    <p class="muted">{t('common.loading')}</p>
  {/if}
</Modal>

<style>
  dl { display: grid; grid-template-columns: 1fr auto; gap: 6px 16px; margin: 0; }
  dt { color: var(--muted); }
  dd { margin: 0; text-align: right; }
  .note { padding: 10px 12px; border-radius: var(--radius); background: var(--warn-bg); color: var(--warn); font-weight: 600; }
  .net { display: flex; align-items: baseline; justify-content: space-between; gap: 12px; padding-top: 12px; border-top: var(--hair) solid var(--line); }
  .net strong { font-size: 2rem; font-weight: 700; letter-spacing: -0.01em; }
</style>
