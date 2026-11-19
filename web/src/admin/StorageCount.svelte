<script>
  import { onMount } from 'svelte';
  import ProductThumb from '../components/ProductThumb.svelte';
  import { post } from '../lib/api.js';
  import { t } from '../lib/i18n.svelte.js';
  import { listenForScans } from '../lib/scanner.js';
  import { toast, toastError } from '../lib/state.svelte.js';

  let { products, onclose, onsaved } = $props();

  let search = $state('');
  let entries = $state({}); // product id -> { crates: '', units: '' }
  let busy = $state(false);

  const shown = $derived(products.filter((p) => p.name.toLowerCase().includes(search.trim().toLowerCase())));

  function counted(product) {
    const e = entries[product.id];
    if (!e || (e.crates === '' && e.units === '')) return null;
    return Number(e.crates || 0) * product.crateSize + Number(e.units || 0);
  }

  const lines = $derived(products.filter((p) => counted(p) !== null).map((p) => ({ productId: p.id, counted: counted(p) })));

  function entry(product) {
    entries[product.id] ??= { crates: '', units: '' };
    return entries[product.id];
  }

  function scan(code) {
    const product = products.find((p) => p.barcodes.includes(code) || p.crateBarcode === code);
    if (!product) {
      toast(t('sell.unknownCode', { code }), 'bad');
      return;
    }
    const field = product.crateBarcode === code ? 'crates' : 'units';
    entry(product)[field] = Number(entry(product)[field] || 0) + 1;
    search = product.name; // shows the row the scan just changed
  }

  function searchKey(e) {
    if (e.key !== 'Enter') return;
    e.preventDefault();
    if (/^\d{4,}$/.test(search.trim())) scan(search.trim());
  }

  async function save() {
    if (busy || lines.length === 0) return;
    busy = true;
    try {
      await post('/api/admin/stocktakes', { lines });
      toast(t('count.saved', { n: lines.length }), 'good');
      onsaved();
    } catch (err) {
      toastError(err);
    } finally {
      busy = false;
    }
  }

  onMount(() => listenForScans({ onScan: scan }));
</script>

<div class="stack gap">
  <header class="stack tight">
    <h1>{t('st.countStorage')}</h1>
    <p class="muted small">{t('sc.hint')}</p>
  </header>

  <input class="input search" type="search" bind:value={search} onkeydown={searchKey} placeholder={t('rs.scan')} aria-label={t('rs.scan')} autocomplete="off" />

  <section class="card">
    {#if shown.length === 0}
      <p class="empty">{t('prod.none')}</p>
    {:else}
      <div class="table-wrap">
        <table class="table">
          <thead>
            <tr>
              <th>{t('common.product')}</th>
              <th class="right">{t('sc.system')}</th>
              <th>{t('rs.crates')}</th>
              <th>{t('rs.units')}</th>
              <th class="right">{t('st.counted')}</th>
              <th class="right">{t('st.diff')}</th>
            </tr>
          </thead>
          <tbody>
            {#each shown as product (product.id)}
              <tr>
                <td><ProductThumb {product} size={28} /> {product.name}</td>
                <td class="right num muted">{product.storageStock}</td>
                <td>
                  <input
                    class="input num short"
                    type="number"
                    min="0"
                    value={entries[product.id]?.crates ?? ''}
                    oninput={(e) => (entry(product).crates = e.target.value)}
                    disabled={product.crateSize <= 1}
                    aria-label="{t('rs.crates')}, {product.name}"
                  />
                </td>
                <td>
                  <input
                    class="input num short"
                    type="number"
                    min="0"
                    value={entries[product.id]?.units ?? ''}
                    oninput={(e) => (entry(product).units = e.target.value)}
                    aria-label="{t('rs.units')}, {product.name}"
                  />
                </td>
                <td class="right num">{counted(product) ?? '–'}</td>
                <td class="right num" class:bad={counted(product) !== null && counted(product) < product.storageStock}>
                  {counted(product) === null ? '–' : counted(product) - product.storageStock}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
    <div class="row foot">
      <button class="btn big grow" onclick={onclose}>{t('common.cancel')}</button>
      <button class="btn big primary grow" disabled={busy || lines.length === 0} onclick={save}>
        {lines.length ? t('sc.save', { n: lines.length }) : t('common.save')}
      </button>
    </div>
  </section>
</div>

<style>
  .gap { gap: 16px; }
  .tight { gap: 2px; }
  .short { width: 84px; }
  .foot { position: sticky; bottom: 0; padding: 12px; border-top: var(--hair) solid var(--line); background: var(--surface); }
</style>
