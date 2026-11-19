<script>
  import Modal from '../components/Modal.svelte';
  import ProductThumb from '../components/ProductThumb.svelte';
  import { post } from '../lib/api.js';
  import { t } from '../lib/i18n.svelte.js';
  import { toast, toastError } from '../lib/state.svelte.js';

  let { product, onclose, onsaved } = $props();

  // svelte-ignore state_referenced_locally
  const size = product.crateSize;
  // Before the bar is half empty there is no suggestion, so offer what still fits.
  // svelte-ignore state_referenced_locally
  const offer = product.refillQty || Math.max(Math.min(product.barTarget - product.barStock, product.storageStock), 0);
  let crates = $state(size > 1 ? Math.floor(offer / size) : 0);
  let units = $state(size > 1 ? offer % size : offer);
  let busy = $state(false);

  const qty = $derived(Number(crates || 0) * size + Number(units || 0));

  async function submit(e) {
    e.preventDefault();
    if (busy || qty <= 0) return;
    busy = true;
    try {
      await post('/api/admin/refills', { lines: [{ productId: product.id, qty }] });
      toast(t('st.booked'), 'good');
      onsaved();
    } catch (err) {
      toastError(err);
    } finally {
      busy = false;
    }
  }
</script>

<Modal title={t('fetch.title')} {onclose}>
  <form class="stack" onsubmit={submit}>
    <div class="row">
      <ProductThumb {product} size={40} />
      <strong>{product.name}</strong>
    </div>
    <div class="amounts">
      {#if size > 1}
        <label class="field">
          <span>{t('rs.crates')} ({t('count.perCrate', { n: size })})</span>
          <input class="input num" type="number" min="0" bind:value={crates} />
        </label>
      {/if}
      <label class="field">
        <span>{t('rs.units')}</span>
        <input class="input num" type="number" min="0" bind:value={units} />
      </label>
    </div>
    <div class="stats after">
      <div class="side">
        <span class="muted small">{t('place.storage')}</span>
        <span class="num" class:bad={product.storageStock - qty < 0}>{t('fetch.after', { from: product.storageStock, to: product.storageStock - qty })}</span>
      </div>
      <div class="side">
        <span class="muted small">{t('place.bar')}</span>
        <span class="num">{t('fetch.after', { from: product.barStock, to: product.barStock + qty })}</span>
      </div>
    </div>
    {#if qty > product.storageStock}<p class="muted small">{t('fetch.short')}</p>{/if}
    <div class="row">
      <button type="button" class="btn big grow" onclick={onclose}>{t('common.cancel')}</button>
      <button class="btn big primary grow" disabled={busy || qty <= 0}>{t('fetch.book')}</button>
    </div>
  </form>
</Modal>

<style>
  .amounts { display: grid; grid-auto-flow: column; grid-auto-columns: 1fr; gap: 12px; }
  .after { grid-template-columns: 1fr 1fr; }
  .side { display: flex; flex-direction: column; gap: 2px; padding: 10px 12px; background: var(--surface); }
</style>
