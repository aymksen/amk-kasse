<script>
  import { onMount } from 'svelte';
  import MessageCircle from 'lucide-svelte/icons/message-circle';
  import TriangleAlert from 'lucide-svelte/icons/triangle-alert';
  import BarChart from './BarChart.svelte';
  import StatTile from './StatTile.svelte';
  import { get } from '../lib/api.js';
  import { dateTime, euro, time } from '../lib/format.js';
  import { t } from '../lib/i18n.svelte.js';
  import { route } from '../lib/route.svelte.js';
  import { toastError } from '../lib/state.svelte.js';

  let data = $state(null);
  const topMax = $derived(Math.max(...(data?.topProducts ?? []).map((p) => p.revenueCents), 1));

  onMount(async () => {
    try {
      data = await get('/api/admin/dashboard');
    } catch (err) {
      toastError(err);
    }
  });

  const difference = (shift) => (shift.countedCents ?? 0) - (shift.expectedCents ?? 0);
  const crates = (item) => (item.crateSize > 1 ? ` · ${(item.stock / item.crateSize).toFixed(1).replace('.', ',')} ${t('rs.crates')}` : '');
</script>

{#if data}
  <div class="stack page">
    <header class="spread wrap">
      <h1>{t('admin.overview')}</h1>
      <div class="row wrap small muted">
        {#if data.openShift}
          <span class="tag good">{t('ov.openShift', { name: data.openShift.workerName, time: time(data.openShift.openedAt) })}</span>
        {/if}
        {#if data.rentalsOut}
          <span class="tag">{t('ov.rentalsOut', { n: data.rentalsOut })}</span>
        {/if}
        {#if data.toRefill}
          <button class="tag warn jump" onclick={() => route.go('/admin/stock')}>{t('ov.toRefill', { n: data.toRefill })}</button>
        {/if}
      </div>
    </header>

    {#if data.message}
      <section class="card pad message">
        <MessageCircle size={20} />
        <p><strong>{t('ov.message', { name: data.message.from })}</strong> {data.message.text}</p>
        <span class="muted small num">{dateTime(data.message.at)}</span>
      </section>
    {/if}

    <div class="stats tiles">
      <StatTile label={t('ov.today')} value={euro(data.todayCents)} note={t('ov.salesN', { n: data.todaySales })} />
      <StatTile label={t('ov.week')} value={euro(data.weekCents)} />
      <StatTile label={t('ov.month')} value={euro(data.monthCents)} />
      <StatTile label={t('ov.box')} value={euro(data.cashInBoxCents)} />
      <StatTile label={t('ov.stockValue')} value={euro(data.stockValueCents)} />
      <StatTile label={t('ov.missing')} value={euro(data.missingCents30)} note={t('ov.days30')} tone={data.missingCents30 < 0 ? 'bad' : ''} />
    </div>

    <section class="card pad">
      <BarChart series={data.series} label={t('ov.chart')} />
    </section>

    <div class="columns">
      <section class="card pad stack">
        <h2>{t('ov.top')}</h2>
        {#if data.topProducts.length}
          <ul class="bars">
            {#each data.topProducts as product}
              <li>
                <div class="spread">
                  <span>{product.name}</span>
                  <span class="muted num">{product.qty} · {euro(product.revenueCents)}</span>
                </div>
                <div class="track"><div class="fill" style:width="{(product.revenueCents / topMax) * 100}%"></div></div>
              </li>
            {/each}
          </ul>
        {:else}
          <p class="muted">{t('rep.none')}</p>
        {/if}
      </section>

      <section class="card pad stack">
        <h2>{t('ov.low')}</h2>
        {#if data.lowStock.length}
          <ul class="low">
            {#each data.lowStock as item}
              <li class="spread">
                <span class="row"><TriangleAlert size={16} class={item.stock <= 0 ? 'bad' : 'warn-icon'} /> {item.name}</span>
                <span class="muted num">{item.stock} / {item.minStock}{crates(item)}</span>
              </li>
            {/each}
          </ul>
        {:else}
          <p class="muted">{t('ov.lowNone')}</p>
        {/if}
      </section>
    </div>

    <section class="card">
      <div class="head"><h2>{t('ov.recent')}</h2></div>
      {#if data.recentShifts.length}
        <div class="table-wrap">
          <table class="table">
            <thead>
              <tr>
                <th>{t('sh.start')}</th>
                <th>{t('sh.worker')}</th>
                <th class="right">{t('sh.revenue')}</th>
                <th class="right">{t('sh.diff')}</th>
              </tr>
            </thead>
            <tbody>
              {#each data.recentShifts as shift}
                <tr>
                  <td class="num">{dateTime(shift.openedAt)}</td>
                  <td>{shift.workerName} ({shift.workerRoom})</td>
                  <td class="right num">{euro(shift.totalCents)}</td>
                  <td class="right num" class:bad={difference(shift) < 0}>{shift.closedAt ? euro(difference(shift)) : t('sh.running')}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {:else}
        <p class="empty">{t('sh.none')}</p>
      {/if}
    </section>
  </div>
{:else}
  <p class="muted">{t('common.loading')}</p>
{/if}

<style>
  .jump { border: 0; cursor: pointer; }
  .message { display: flex; align-items: center; gap: 12px; }
  .message > :global(svg) { flex: none; color: var(--accent-text); }
  .message p { flex: 1; min-width: 0; overflow-wrap: anywhere; }
  .message span { flex: none; }
  .page { gap: 16px; max-width: 1200px; }
  .tiles { grid-template-columns: repeat(3, minmax(0, 1fr)); }
  @media (min-width: 1400px) {
    .tiles { grid-template-columns: repeat(6, minmax(0, 1fr)); }
  }
  .columns { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 16px; }
  .bars, .low { display: flex; flex-direction: column; gap: 10px; margin: 0; padding: 0; list-style: none; }
  .track { height: 6px; margin-top: 4px; border-radius: 2px; background: var(--surface-3); }
  .fill { height: 100%; border-radius: 2px; background: var(--chart); }
  .head { padding: 16px 16px 4px; }
  :global(.warn-icon) { color: var(--warn); }
</style>
