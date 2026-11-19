<script>
  import { onMount } from 'svelte';
  import SalesList from './SalesList.svelte';
  import Modal from '../components/Modal.svelte';
  import { get } from '../lib/api.js';
  import { dateTime, euro } from '../lib/format.js';
  import { t } from '../lib/i18n.svelte.js';
  import { till, toastError } from '../lib/state.svelte.js';

  let shifts = $state(null);
  let detail = $state(null);

  const difference = (shift) => (shift.countedCents ?? 0) - (shift.expectedCents ?? 0);

  async function open(shift) {
    try {
      detail = await get(`/api/admin/shifts/${shift.id}`);
    } catch (err) {
      toastError(err);
    }
  }

  onMount(async () => {
    try {
      shifts = await get('/api/admin/shifts');
    } catch (err) {
      toastError(err);
    }
  });
</script>

<div class="stack gap">
  <h1>{t('admin.shifts')}</h1>
  <section class="card">
    {#if shifts === null}
      <p class="empty">{t('common.loading')}</p>
    {:else if shifts.length === 0}
      <p class="empty">{t('sh.none')}</p>
    {:else}
      <div class="table-wrap">
        <table class="table">
          <thead>
            <tr>
              <th>{t('sh.start')}</th>
              <th>{t('sh.worker')}</th>
              <th class="right">{t('sh.revenue')}</th>
              <th class="right">{t('sh.expected')}</th>
              <th class="right">{t('sh.counted')}</th>
              <th class="right">{t('sh.diff')}</th>
            </tr>
          </thead>
          <tbody>
            {#each shifts as shift (shift.id)}
              <tr class="click" onclick={() => open(shift)}>
                <td class="num">{dateTime(shift.openedAt)}</td>
                <td>{shift.workerName} ({shift.workerRoom})</td>
                <td class="right num">{euro(shift.totalCents)}</td>
                {#if shift.closedAt}
                  <td class="right num muted">{euro(shift.expectedCents)}</td>
                  <td class="right num muted">{euro(shift.countedCents)}</td>
                  <td class="right num" class:bad={difference(shift) < 0}>{euro(difference(shift))}</td>
                {:else}
                  <td class="right" colspan="3"><span class="tag good">{t('sh.running')}</span></td>
                {/if}
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  </section>
</div>

{#if detail}
  <Modal title={t('sh.detail', { name: detail.shift.workerName })} wide onclose={() => (detail = null)}>
    <div class="stack">
      <dl>
        <dt>{t('sh.start')}</dt>
        <dd class="num">{dateTime(detail.shift.openedAt)}</dd>
        <dt>{t('sh.end')}</dt>
        <dd class="num">{detail.shift.closedAt ? dateTime(detail.shift.closedAt) : t('sh.running')}</dd>
        <dt>{t('sh.float')}</dt>
        <dd class="num">{euro(detail.shift.floatCents)}</dd>
        <dt>{t('sh.floatExpected')}</dt>
        <dd class="num">{euro(detail.shift.expectedFloatCents)}</dd>
        <dt>{t('close.cashSales')}</dt>
        <dd class="num">{euro(detail.totals.cashCents)}</dd>
        <dt>{till.settings.otherPayment || t('common.other')}</dt>
        <dd class="num">{euro(detail.totals.otherCents)}</dd>
        {#if detail.shift.note}
          <dt>{t('common.note')}</dt>
          <dd>{detail.shift.note}</dd>
        {/if}
        {#if detail.shift.handover}
          <dt>{t('sh.handover')}</dt>
          <dd>{detail.shift.handover}</dd>
        {/if}
      </dl>
      {#if detail.sales.length}
        <SalesList sales={[...detail.sales].reverse()} />
      {/if}
    </div>
  </Modal>
{/if}

<style>
  .gap { gap: 16px; }
  dl { display: grid; grid-template-columns: 1fr auto; gap: 6px 16px; margin: 0; }
  dt { color: var(--muted); }
  dd { margin: 0; text-align: right; }
</style>
