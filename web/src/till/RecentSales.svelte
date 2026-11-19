<script>
  import { onMount } from 'svelte';
  import Modal from '../components/Modal.svelte';
  import { get, post } from '../lib/api.js';
  import { euro, time } from '../lib/format.js';
  import { t } from '../lib/i18n.svelte.js';
  import { refresh, till, toast, toastError } from '../lib/state.svelte.js';

  let { onclose } = $props();

  const reasons = ['reason.typo', 'reason.changed', 'reason.payment', 'reason.other'];
  let sales = $state(null);
  let voiding = $state(null);
  let busy = $state(false);

  async function load() {
    try {
      sales = (await get('/api/shift/sales')).reverse();
    } catch (err) {
      toastError(err);
      sales = [];
    }
  }

  async function confirmVoid(reasonKey) {
    if (busy) return;
    busy = true;
    try {
      await post(`/api/sales/${voiding.id}/void`, { reason: t(reasonKey) });
      voiding = null;
      toast(t('recent.done'), 'good', 6000);
      await Promise.all([load(), refresh()]);
    } catch (err) {
      toastError(err);
    } finally {
      busy = false;
    }
  }

  const summary = (sale) => sale.lines.map((l) => `${Math.abs(l.qty)}× ${l.name}`).join(', ');
  const paymentLabel = (sale) => (sale.payment === 'cash' ? t('common.cash') : till.settings.otherPayment);

  onMount(load);
</script>

<Modal title={voiding ? t('recent.whyVoid') : t('recent.title')} wide onclose={voiding ? () => (voiding = null) : onclose}>
  {#if voiding}
    <div class="stack">
      <p>
        <span class="num">{time(voiding.createdAt)}</span> · {summary(voiding)} ·
        <strong class="num">{euro(voiding.totalCents)}</strong>
      </p>
      <div class="reasons">
        {#each reasons as reason}
          <button class="btn big" disabled={busy} onclick={() => confirmVoid(reason)}>{t(reason)}</button>
        {/each}
      </div>
    </div>
  {:else if sales === null}
    <p class="muted">{t('common.loading')}</p>
  {:else if sales.length === 0}
    <p class="empty">{t('recent.none')}</p>
  {:else}
    <ul class="list">
      {#each sales as sale (sale.id)}
        <li class="sale">
          <span class="num muted">{time(sale.createdAt)}</span>
          <div class="grow">
            <div class:struck={sale.voided}>{summary(sale)}</div>
            <div class="muted small">
              {paymentLabel(sale)}
              {#if sale.voidsSaleId}· {t('recent.reversal')}: {sale.reason}{/if}
            </div>
          </div>
          <strong class="num" class:struck={sale.voided}>{euro(sale.totalCents)}</strong>
          <div class="action">
            {#if sale.voided}
              <span class="tag">{t('recent.voided')}</span>
            {:else if !sale.voidsSaleId}
              <button class="btn" onclick={() => (voiding = sale)}>{t('recent.void')}</button>
            {/if}
          </div>
        </li>
      {/each}
    </ul>
  {/if}
</Modal>

<style>
  .reasons { display: grid; grid-template-columns: repeat(2, 1fr); gap: 8px; }
  .reasons .btn { white-space: normal; }
  .list { display: flex; flex-direction: column; margin: 0; padding: 0; list-style: none; }
  .sale { display: flex; align-items: center; gap: 12px; padding: 10px 0; border-bottom: var(--hair) solid var(--line); }
  .sale:last-child { border-bottom: 0; }
  .action { display: flex; justify-content: flex-end; width: 92px; }
  .struck { color: var(--muted); text-decoration: line-through; }
</style>
