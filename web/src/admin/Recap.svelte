<script>
  import { onMount } from 'svelte';
  import StatTile from './StatTile.svelte';
  import BoardButton from '../components/BoardButton.svelte';
  import SlideTabs from '../components/SlideTabs.svelte';
  import ProductThumb from '../components/ProductThumb.svelte';
  import { get } from '../lib/api.js';
  import { day, time } from '../lib/format.js';
  import { getLanguage, t } from '../lib/i18n.svelte.js';
  import { toastError } from '../lib/state.svelte.js';
  import { nightName, periodName } from '../moments/recapText.js';

  let recap = $state(null);
  let period = $state('semester');

  const data = $derived(recap?.[period]);
  const tabs = $derived(
    recap
      ? [
          { value: 'week', label: t('recap.week') },
          { value: 'month', label: t('recap.month') },
          { value: 'lastMonth', label: t('recap.lastMonth') },
          { value: 'semester', label: periodName(recap, 'semester') },
          { value: 'lastSemester', label: periodName(recap, 'lastSemester') },
        ]
      : [],
  );
  const format = new Intl.NumberFormat('de-DE');

  onMount(async () => {
    try {
      recap = await get('/api/recap');
    } catch (err) {
      toastError(err);
    }
  });
</script>

<div class="stack gap">
  <div class="spread">
    <h1>{t('admin.recap')}</h1>
    <BoardButton labelled primary />
  </div>

  {#if recap}
    <SlideTabs options={tabs} value={period} onpick={(id) => (period = id)} label={t('admin.recap')} />
    <p class="small muted">{day(data.from)} – {day(data.to)}. {t('recap.hint')}</p>

    {#if data.qty}
      <div class="stats tiles">
        <StatTile label={t('recap.pieces')} value={format.format(data.qty)} />
        <StatTile label={t('recap.orders')} value={format.format(data.orders)} />
        <StatTile label={t('recap.nights')} value={format.format(data.nights)} />
        {#if data.lastAt}<StatTile label={t('recap.latest')} value={time(data.lastAt)} />{/if}
      </div>

      <div class="columns">
        <section class="card highlight">
          <h3>{t('recap.topDrink')}</h3>
          {#if data.drinks.length}
            <div class="row"><ProductThumb product={data.drinks[0]} /><strong>{data.drinks[0].name}</strong></div>
            <p class="muted">{format.format(data.drinks[0].qty)}{t('recap.times')}</p>
          {:else}
            <p class="muted">{t('rep.none')}</p>
          {/if}
        </section>
        <section class="card highlight">
          <h3>{t('recap.bestseller')}</h3>
          <div class="row"><ProductThumb product={data.products[0]} /><strong>{data.products[0].name}</strong></div>
          <p class="muted">{format.format(data.products[0].qty)}{t('recap.times')}</p>
        </section>
        <section class="card highlight">
          <h3>{t('recap.mostSold')}</h3>
          <strong class="big">{data.workers[0].name}</strong>
          <p class="muted">{t('common.room')} {data.workers[0].room}</p>
          <p class="muted">{t(data.workers[0].shifts === 1 ? 'recap.workerNoteOne' : 'recap.workerNote', { qty: format.format(data.workers[0].qty), shifts: data.workers[0].shifts })}</p>
        </section>
        <section class="card highlight">
          <h3>{t('recap.busiest')}</h3>
          <strong class="big">{nightName(data.busiest.day, getLanguage())}</strong>
          <p class="muted">{format.format(data.busiest.qty)} {t('recap.piecesShort')}, {data.busiest.workers.join(', ')}</p>
        </section>
      </div>

      <div class="columns two">
        <section class="card">
          <div class="head"><h2>{t('recap.ranking')}</h2></div>
          <div class="table-wrap">
            <table class="table">
              <thead>
                <tr><th>{t('common.product')}</th><th class="right">{t('rep.units')}</th></tr>
              </thead>
              <tbody>
                {#each data.products as row}
                  <tr><td>{row.name}</td><td class="right num">{format.format(row.qty)}</td></tr>
                {/each}
              </tbody>
            </table>
          </div>
        </section>
        <section class="card">
          <div class="head"><h2>{t('recap.crew')}</h2></div>
          <div class="table-wrap">
            <table class="table">
              <thead>
                <tr><th>{t('common.name')}</th><th>{t('common.room')}</th><th class="right">{t('admin.shifts')}</th><th class="right">{t('rep.units')}</th></tr>
              </thead>
              <tbody>
                {#each data.workers.slice(0, 10) as row}
                  <tr><td>{row.name}</td><td class="num">{row.room}</td><td class="right num">{row.shifts}</td><td class="right num">{format.format(row.qty)}</td></tr>
                {/each}
              </tbody>
            </table>
          </div>
        </section>
      </div>
    {:else}
      <p class="card empty">{t('rep.none')}</p>
    {/if}
  {/if}
</div>

<style>
  .gap { gap: 16px; }
  .tiles { grid-template-columns: repeat(auto-fit, minmax(136px, 1fr)); }
  .columns { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 16px; }
  .columns.two { grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); }
  .highlight { display: flex; flex-direction: column; gap: 8px; padding: 14px 16px; }
  .highlight strong { font-size: 1.0625rem; line-height: 1.2; }
  .highlight strong.big { font-size: 1.375rem; font-stretch: 112%; font-weight: 800; }
  .head { padding: 16px 16px 4px; }
</style>
