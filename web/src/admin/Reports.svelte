<script>
  import { onMount } from 'svelte';
  import Download from 'lucide-svelte/icons/download';
  import DateRange from './DateRange.svelte';
  import StatTile from './StatTile.svelte';
  import { get } from '../lib/api.js';
  import { businessToday, euro } from '../lib/format.js';
  import { t } from '../lib/i18n.svelte.js';
  import { till, toastError } from '../lib/state.svelte.js';

  let from = $state(businessToday(-29));
  let to = $state(businessToday());
  let report = $state(null);

  const margin = $derived(report ? report.revenueCents - report.costCents : 0);
  const percent = $derived(report?.revenueCents ? `${Math.round((margin / report.revenueCents) * 100)} %` : '');
  const range = $derived(`?from=${from}&to=${to}`);

  async function load() {
    if (!from || !to) return;
    try {
      report = await get(`/api/admin/report${range}`);
    } catch (err) {
      toastError(err);
    }
  }

  const share = (cents) => `${Math.round((cents / (report.cashCents + report.otherCents)) * 100)} %`;

  onMount(load);
</script>

<div class="stack gap">
  <h1>{t('admin.reports')}</h1>
  <DateRange bind:from bind:to onchange={load} />

  {#if report}
    <div class="stats tiles">
      <StatTile label={t('rep.revenue')} value={euro(report.revenueCents)} note={`${report.salesCount} ${t('rep.salesCount')}`} />
      <StatTile label={t('rep.cost')} value={euro(report.costCents)} />
      <StatTile label={t('rep.margin')} value={euro(margin)} note={percent} />
      <StatTile label={t('rep.cash')} value={euro(report.cashCents)} />
      {#if report.otherCents || till.settings.otherPayment}
        <StatTile label={till.settings.otherPayment || t('common.other')} value={euro(report.otherCents)} />
      {/if}
    </div>

    {#if report.cashCents > 0 && report.otherCents > 0}
      <div class="split" role="img" aria-label={`${t('common.cash')} ${share(report.cashCents)}, ${till.settings.otherPayment} ${share(report.otherCents)}`}>
        <div class="part" style:flex-grow={report.cashCents}>
          <span class="bar cash"></span>
          <span class="small">{t('common.cash')} <span class="num muted">{share(report.cashCents)}</span></span>
        </div>
        <div class="part" style:flex-grow={report.otherCents}>
          <span class="bar other"></span>
          <span class="small">{till.settings.otherPayment} <span class="num muted">{share(report.otherCents)}</span></span>
        </div>
      </div>
    {/if}

    <div class="columns">
      <section class="card">
        <div class="head"><h2>{t('rep.byCategory')}</h2></div>
        {#if report.categories.length}
          <div class="table-wrap">
            <table class="table">
              <thead>
                <tr><th>{t('common.category')}</th><th class="right">{t('rep.units')}</th><th class="right">{t('rep.revenue')}</th></tr>
              </thead>
              <tbody>
                {#each report.categories as row}
                  <tr><td>{row.name}</td><td class="right num">{row.qty}</td><td class="right num">{euro(row.revenueCents)}</td></tr>
                {/each}
              </tbody>
            </table>
          </div>
        {:else}
          <p class="empty">{t('rep.none')}</p>
        {/if}
      </section>

      <section class="card">
        <div class="head"><h2>{t('rep.losses')}</h2></div>
        {#if report.losses.length}
          <div class="table-wrap">
            <table class="table">
              <thead>
                <tr><th>{t('common.product')}</th><th class="right">{t('rep.units')}</th><th class="right">{t('prod.cost')}</th></tr>
              </thead>
              <tbody>
                {#each report.losses as row}
                  <tr><td>{row.name}</td><td class="right num">{row.units}</td><td class="right num">{euro(row.costCents)}</td></tr>
                {/each}
              </tbody>
            </table>
          </div>
        {:else}
          <p class="empty">{t('rep.lossesNone')}</p>
        {/if}
      </section>
    </div>

    <section class="card">
      <div class="head"><h2>{t('rep.byProduct')}</h2></div>
      {#if report.products.length}
        <div class="table-wrap">
          <table class="table">
            <thead>
              <tr>
                <th>{t('common.product')}</th>
                <th>{t('common.category')}</th>
                <th class="right">{t('rep.units')}</th>
                <th class="right">{t('rep.revenue')}</th>
                <th class="right">{t('rep.margin')}</th>
              </tr>
            </thead>
            <tbody>
              {#each report.products as row}
                <tr>
                  <td>{row.name}</td>
                  <td class="muted">{row.category}</td>
                  <td class="right num">{row.qty}</td>
                  <td class="right num">{euro(row.revenueCents)}</td>
                  <td class="right num">{euro(row.revenueCents - row.costCents)}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {:else}
        <p class="empty">{t('rep.none')}</p>
      {/if}
    </section>
  {/if}

  <section class="card pad stack">
    <h2>{t('rep.export')}</h2>
    <div class="row wrap">
      <a class="btn" href={`/api/admin/export/sales.csv${range}`} download><Download size={18} /> {t('rep.exSales')}</a>
      <a class="btn" href={`/api/admin/export/products.csv${range}`} download><Download size={18} /> {t('rep.exProducts')}</a>
      <a class="btn" href="/api/admin/export/shifts.csv" download><Download size={18} /> {t('rep.exShifts')}</a>
      <a class="btn" href="/api/admin/export/stock.csv" download><Download size={18} /> {t('rep.exStock')}</a>
    </div>
  </section>
</div>

<style>
  .gap { gap: 16px; }
  .tiles { grid-template-columns: repeat(auto-fit, minmax(136px, 1fr)); }
  .split { display: flex; gap: 2px; }
  .part { display: flex; flex-direction: column; gap: 4px; flex-basis: 0; min-width: 96px; white-space: nowrap; }
  .bar { height: 10px; border-radius: 2px; }
  .bar.cash { background: var(--chart); }
  .bar.other { background: var(--text); }
  .columns { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 16px; }
  .head { padding: 16px 16px 4px; }
  a.btn { color: inherit; text-decoration: none; }
</style>
