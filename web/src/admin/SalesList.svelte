<script>
  import { dateTime, euro } from '../lib/format.js';
  import { t } from '../lib/i18n.svelte.js';
  import { till } from '../lib/state.svelte.js';

  let { sales } = $props();

  const items = (sale) => sale.lines.map((l) => `${Math.abs(l.qty)}× ${l.name}`).join(', ');
</script>

<div class="table-wrap">
  <table class="table">
    <thead>
      <tr>
        <th>{t('common.time')}</th>
        <th>{t('sales.items')}</th>
        <th>{t('sales.payment')}</th>
        <th class="right">{t('common.total')}</th>
      </tr>
    </thead>
    <tbody>
      {#each sales as sale (sale.id)}
        <tr>
          <td class="num nowrap">{dateTime(sale.createdAt)}</td>
          <td>
            {items(sale)}
            {#if sale.voided}<span class="tag">{t('recent.voided')}</span>{/if}
            {#if sale.voidsSaleId}<span class="tag warn">{t('recent.reversal')}</span> <span class="muted small">{sale.reason}</span>{/if}
          </td>
          <td class="muted">{sale.payment === 'cash' ? t('common.cash') : till.settings.otherPayment || sale.payment}</td>
          <td class="right num">{euro(sale.totalCents)}</td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>

<style>
  .nowrap { white-space: nowrap; }
</style>
