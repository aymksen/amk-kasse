<script>
  import { onMount } from 'svelte';
  import Download from 'lucide-svelte/icons/download';
  import DateRange from './DateRange.svelte';
  import SalesList from './SalesList.svelte';
  import { get } from '../lib/api.js';
  import { businessToday } from '../lib/format.js';
  import { t } from '../lib/i18n.svelte.js';
  import { toastError } from '../lib/state.svelte.js';

  let from = $state(businessToday(-6));
  let to = $state(businessToday());
  let sales = $state(null);

  async function load() {
    if (!from || !to) return;
    try {
      sales = await get(`/api/admin/sales?from=${from}&to=${to}`);
    } catch (err) {
      toastError(err);
    }
  }

  onMount(load);
</script>

<div class="stack gap">
  <header class="spread wrap">
    <h1>{t('admin.sales')}</h1>
    <a class="btn" href={`/api/admin/export/sales.csv?from=${from}&to=${to}`} download><Download size={18} /> CSV</a>
  </header>
  <DateRange bind:from bind:to onchange={load} />
  <section class="card">
    {#if sales === null}
      <p class="empty">{t('common.loading')}</p>
    {:else if sales.length === 0}
      <p class="empty">{t('sales.none')}</p>
    {:else}
      <SalesList {sales} />
    {/if}
  </section>
  {#if sales?.length === 500}
    <p class="muted small">{t('sales.limit')}</p>
  {/if}
</div>

<style>
  .gap { gap: 16px; }
  a.btn { color: inherit; text-decoration: none; }
</style>
