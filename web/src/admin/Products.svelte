<script>
  import { onMount } from 'svelte';
  import Plus from 'lucide-svelte/icons/plus';
  import ProductThumb from '../components/ProductThumb.svelte';
  import SlideTabs from '../components/SlideTabs.svelte';
  import CategoriesDialog from './CategoriesDialog.svelte';
  import ProductDialog from './ProductDialog.svelte';
  import { get } from '../lib/api.js';
  import { euro } from '../lib/format.js';
  import { t } from '../lib/i18n.svelte.js';
  import { refresh, toastError } from '../lib/state.svelte.js';

  let products = $state(null);
  let categories = $state([]);
  let search = $state('');
  let category = $state(null);
  let editing = $state(null);
  let showCategories = $state(false);

  const tabs = $derived([{ value: null, label: t('common.all') }, ...categories.map((c) => ({ value: c.id, label: c.name }))]);
  const total = (p) => p.barStock + p.storageStock;
  const categoryName = (id) => categories.find((c) => c.id === id)?.name ?? '';
  const visible = $derived(
    (products ?? []).filter(
      (p) =>
        (category === null || p.categoryId === category) &&
        (p.name.toLowerCase().includes(search.toLowerCase()) || (search && (p.barcodes.includes(search) || p.crateBarcode === search))),
    ),
  );

  async function load() {
    try {
      [products, categories] = await Promise.all([get('/api/admin/products'), get('/api/admin/categories')]);
    } catch (err) {
      toastError(err);
    }
  }

  function saved() {
    editing = null;
    load();
    refresh().catch(toastError);
  }

  function newProduct() {
    editing = {
      id: 0, name: '', categoryId: categories[0]?.id ?? 0, priceCents: 0, costCents: 0, barcodes: [], crateBarcode: '',
      crateSize: 1, trackStock: true, barTarget: 10, minStock: 0, location: '', quickCount: false, sort: products.length, active: true,
    };
  }

  onMount(load);
</script>

<div class="stack gap">
  <header class="spread wrap">
    <h1>{t('admin.products')}</h1>
    <div class="row">
      <button class="btn" onclick={() => (showCategories = true)}>{t('prod.categories')}</button>
      <button class="btn primary" onclick={newProduct} disabled={!products}><Plus size={18} /> {t('prod.new')}</button>
    </div>
  </header>

  <div class="row wrap">
    <input class="input search" type="search" bind:value={search} placeholder={t('prod.search')} aria-label={t('prod.search')} autocomplete="off" />
    <SlideTabs options={tabs} value={category} onpick={(id) => (category = id)} label={t('common.category')} />
  </div>

  <section class="card">
    {#if products === null}
      <p class="empty">{t('common.loading')}</p>
    {:else if visible.length === 0}
      <p class="empty">{t('prod.none')}</p>
    {:else}
      <div class="table-wrap">
        <table class="table">
          <thead>
            <tr>
              <th>{t('common.name')}</th>
              <th>{t('common.category')}</th>
              <th class="right">{t('prod.price')}</th>
              <th class="right">{t('prod.cost')}</th>
              <th class="right">{t('prod.stock')}</th>
              <th>{t('prod.barcodes')}</th>
            </tr>
          </thead>
          <tbody>
            {#each visible as product (product.id)}
              <tr class="click" onclick={() => (editing = { ...product })}>
                <td>
                  <ProductThumb {product} />
                  <span class="name">{product.name}</span>
                  {#if !product.active}<span class="tag">{t('common.inactive')}</span>{/if}
                </td>
                <td class="muted">{categoryName(product.categoryId)}</td>
                <td class="right num">{euro(product.priceCents)}</td>
                <td class="right num muted">{euro(product.costCents)}</td>
                <td class="right num" class:bad={product.trackStock && total(product) <= 0}>{product.trackStock ? total(product) : '–'}</td>
                <td class="muted num small">
                  {product.barcodes[0] ?? '–'}
                  {#if product.barcodes.length > 1}<span class="more" title={product.barcodes.slice(1).join(', ')}>+{product.barcodes.length - 1}</span>{/if}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  </section>
</div>

{#if editing}
  <ProductDialog product={editing} {categories} onclose={() => (editing = null)} onsaved={saved} />
{/if}
{#if showCategories}
  <CategoriesDialog
    {categories}
    onclose={() => (showCategories = false)}
    onchanged={() => {
      load();
      refresh().catch(toastError);
    }}
  />
{/if}

<style>
  .gap { gap: 16px; }
  .search { width: 240px; }
  .name { margin-left: 8px; }
  .more { margin-left: 4px; padding: 0 5px; border-radius: 4px; background: var(--surface-2); color: var(--text); font-weight: 600; }
</style>
