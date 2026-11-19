<script>
  import { onMount } from 'svelte';
  import Modal from '../components/Modal.svelte';
  import ProductThumb from '../components/ProductThumb.svelte';
  import SlideTabs from '../components/SlideTabs.svelte';
  import RefillDialog from './RefillDialog.svelte';
  import StatTile from './StatTile.svelte';
  import StorageCount from './StorageCount.svelte';
  import { get, post } from '../lib/api.js';
  import { dateTime, euro } from '../lib/format.js';
  import { t } from '../lib/i18n.svelte.js';
  import { refresh, toast, toastError } from '../lib/state.svelte.js';

  let products = $state(null);
  let categories = $state([]);
  let counts = $state([]);
  let refills = $state([]);
  let detail = $state(null);
  let fetching = $state(null);
  let confirming = $state(false);
  let counting = $state(false);
  let view = $state('all');
  let search = $state('');
  let busy = $state(false);

  const total = (p) => p.barStock + p.storageStock;
  const worth = (p, n = total(p)) => Math.max(n, 0) * p.costCents;
  const sum = (list, pick) => list.reduce((acc, p) => acc + pick(p), 0);
  const who = (count) => (count.byRoom ? `${count.byName} (${count.byRoom})` : count.byName);

  const stocked = $derived((products ?? []).filter((p) => p.active && p.trackStock));
  const toRefill = $derived(stocked.filter((p) => p.refillQty > 0));
  const toBuy = $derived(stocked.filter((p) => total(p) <= p.minStock));
  const views = $derived([
    { value: 'all', label: t('common.all') },
    { value: 'refill', label: `${t('st.toRefill')} (${toRefill.length})` },
    { value: 'buy', label: `${t('st.toBuy')} (${toBuy.length})` },
  ]);
  const shown = $derived(
    { all: stocked, refill: toRefill, buy: toBuy }[view].filter((p) => p.name.toLowerCase().includes(search.trim().toLowerCase())),
  );
  const groups = $derived(
    categories.map((c) => ({ name: c.name, items: shown.filter((p) => p.categoryId === c.id) })).filter((g) => g.items.length),
  );
  const plan = $derived({ n: shown.length, units: sum(shown, (p) => p.refillQty) });
  const nothingText = $derived(search.trim() || view === 'all' ? 'prod.none' : view === 'refill' ? 'st.noneRefill' : 'st.noneBuy');

  function inCrates(p) {
    if (p.crateSize <= 1 || p.storageStock < p.crateSize) return '';
    const c = Math.floor(p.storageStock / p.crateSize);
    const r = p.storageStock % p.crateSize;
    const crates = c === 1 ? t('st.crate') : t('st.crates', { c });
    return r ? t('st.more', { crates, r }) : crates;
  }

  async function open(count) {
    try {
      detail = { count, lines: await get(`/api/admin/stocktakes/${count.id}`) };
    } catch (err) {
      toastError(err);
    }
  }

  async function bookAll() {
    if (busy) return;
    busy = true;
    try {
      await post('/api/admin/refills', { lines: shown.map((p) => ({ productId: p.id, qty: p.refillQty })) });
      toast(t('st.booked'), 'good');
      confirming = false;
      await changed();
    } catch (err) {
      toastError(err);
    } finally {
      busy = false;
    }
  }

  async function changed() {
    fetching = null;
    counting = false;
    await Promise.all([load(), refresh()]);
  }

  async function load() {
    try {
      [products, categories, counts, refills] = await Promise.all([
        get('/api/admin/products'),
        get('/api/admin/categories'),
        get('/api/admin/stocktakes'),
        get('/api/admin/refills'),
      ]);
    } catch (err) {
      toastError(err);
    }
  }

  onMount(load);
</script>

{#if counting}
  <StorageCount products={stocked} onclose={() => (counting = false)} onsaved={changed} />
{:else}
<div class="stack gap">
  <header class="spread wrap">
    <div class="stack tight">
      <h1>{t('st.title')}</h1>
      <p class="muted small">{t('st.hint')}</p>
    </div>
    <button class="btn" disabled={!products} onclick={() => (counting = true)}>{t('st.countStorage')}</button>
  </header>

  <div class="stats tiles">
    <StatTile
      label={t('st.bar')}
      value={t('st.pieces', { n: sum(stocked, (p) => p.barStock) })}
      note={t('st.valueOf', { value: euro(sum(stocked, (p) => worth(p, p.barStock))) })}
    />
    <StatTile
      label={t('st.storage')}
      value={t('st.pieces', { n: sum(stocked, (p) => p.storageStock) })}
      note={t('st.valueOf', { value: euro(sum(stocked, (p) => worth(p, p.storageStock))) })}
    />
    <StatTile
      label={t('st.total')}
      value={t('st.pieces', { n: sum(stocked, total) })}
      note={t('st.valueOf', { value: euro(sum(stocked, (p) => worth(p))) })}
    />
  </div>

  <div class="row wrap">
    <input class="input search" type="search" bind:value={search} placeholder={t('prod.search')} aria-label={t('prod.search')} autocomplete="off" />
    <SlideTabs options={views} value={view} onpick={(id) => (view = id)} label={t('admin.stock')} />
  </div>

  <section class="card">
    {#if products === null}
      <p class="empty">{t('common.loading')}</p>
    {:else if groups.length === 0}
      <p class="empty">{t(nothingText)}</p>
    {:else}
      {#if view === 'refill'}
        <div class="spread wrap plan">
          <span class="muted">{t('st.plan', plan)}</span>
          <button class="btn primary" onclick={() => (confirming = true)}>{t('st.bookAll')}</button>
        </div>
      {/if}
      <div class="table-wrap">
        <table class="table">
          <thead>
            <tr>
              <th>{t('common.product')}</th>
              <th>{t('st.bar')}</th>
              <th class="right">{t('st.storage')}</th>
              <th class="right">{t('st.total')}</th>
              <th class="right">{t('st.value')}</th>
              <th></th>
            </tr>
          </thead>
          {#each groups as group (group.name)}
            <tbody>
              <tr class="subtotal">
                <th scope="rowgroup">{group.name}</th>
                <td class="num">{sum(group.items, (p) => p.barStock)}</td>
                <td class="right num">{sum(group.items, (p) => p.storageStock)}</td>
                <td class="right num">{sum(group.items, total)}</td>
                <td class="right num">{euro(sum(group.items, (p) => worth(p)))}</td>
                <td></td>
              </tr>
              {#each group.items as product (product.id)}
                <tr>
                  <td>
                    <div class="product">
                      <ProductThumb {product} size={28} />
                      <span>{product.name}</span>
                      {#if product.barStock <= 0}<span class="tag bad">{t('st.barEmpty')}</span>{/if}
                      {#if total(product) <= product.minStock}<span class="tag warn">{t('st.toBuy')}</span>{/if}
                    </div>
                  </td>
                  <td>
                    <div class="at-bar">
                      <span class="num">{product.barStock} <span class="muted">/ {product.barTarget}</span></span>
                      <div class="meter">
                        <div class="level" class:low={product.barStock * 2 <= product.barTarget} style:width="{Math.min(Math.max(product.barStock, 0) / (product.barTarget || 1), 1) * 100}%"></div>
                      </div>
                    </div>
                  </td>
                  <td class="right">
                    <span class="num" class:bad={product.storageStock < 0}>{product.storageStock}</span>
                    {#if inCrates(product)}<div class="muted small crates">{inCrates(product)}</div>{/if}
                  </td>
                  <td class="right num">{total(product)}</td>
                  <td class="right num">{euro(worth(product))}</td>
                  <td class="right">
                    <button class="btn" class:primary={product.refillQty > 0} onclick={() => (fetching = product)}>
                      {product.refillQty > 0 ? t('st.fetchN', { n: product.refillQty }) : t('st.fetch')}
                    </button>
                  </td>
                </tr>
              {/each}
            </tbody>
          {/each}
        </table>
      </div>
    {/if}
  </section>

  <section class="card">
    <div class="head"><h2>{t('st.refills')}</h2></div>
    {#if refills.length}
      <div class="table-wrap recent">
        <table class="table">
          <thead>
            <tr>
              <th>{t('common.time')}</th>
              <th>{t('common.product')}</th>
              <th class="right">{t('rs.pieces')}</th>
            </tr>
          </thead>
          <tbody>
            {#each refills as refill (refill.id)}
              <tr>
                <td class="num">{dateTime(refill.createdAt)}</td>
                <td>{refill.name}</td>
                <td class="right num">{refill.qty}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {:else}
      <p class="empty">{t('st.refillsNone')}</p>
    {/if}
  </section>

  <section class="card">
    <div class="head"><h2>{t('st.counts')}</h2></div>
    {#if counts.length}
      <div class="table-wrap">
        <table class="table">
          <thead>
            <tr>
              <th>{t('common.time')}</th>
              <th>{t('st.by')}</th>
              <th>{t('st.where')}</th>
              <th>{t('st.kind')}</th>
              <th class="right">{t('st.lines')}</th>
              <th class="right">{t('st.diffs')}</th>
              <th class="right">{t('st.diffValue')}</th>
            </tr>
          </thead>
          <tbody>
            {#each counts as count (count.id)}
              <tr class="click" onclick={() => open(count)}>
                <td class="num">{dateTime(count.createdAt)}</td>
                <td>{who(count)}</td>
                <td class="muted">{t(`place.${count.place}`)}</td>
                <td class="muted">{count.opening ? t('st.opening') : count.kind === 'quick' ? t('st.quick') : t('st.full')}</td>
                <td class="right num">{count.lines}</td>
                <td class="right num">{count.diffs}</td>
                <td class="right num" class:bad={count.diffCents < 0}>{euro(count.diffCents)}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {:else}
      <p class="empty">{t('st.none')}</p>
    {/if}
  </section>
</div>
{/if}

{#if fetching}
  <RefillDialog product={fetching} onclose={() => (fetching = null)} onsaved={changed} />
{/if}

{#if confirming}
  <Modal title={t('st.bookAllTitle')} onclose={() => (confirming = false)}>
    <div class="stack">
      <p>{t('st.bookAllText', plan)}</p>
      <div class="row">
        <button class="btn big grow" onclick={() => (confirming = false)}>{t('common.cancel')}</button>
        <button class="btn big primary grow" disabled={busy} onclick={bookAll}>{t('fetch.book')}</button>
      </div>
    </div>
  </Modal>
{/if}

{#if detail}
  <Modal title={t('st.detail', { date: dateTime(detail.count.createdAt) })} wide onclose={() => (detail = null)}>
    <div class="table-wrap">
      <table class="table">
        <thead>
          <tr>
            <th>{t('common.product')}</th>
            <th class="right">{t('st.expected')}</th>
            <th class="right">{t('st.counted')}</th>
            <th class="right">{t('st.diff')}</th>
          </tr>
        </thead>
        <tbody>
          {#each detail.lines as line}
            <tr>
              <td>{line.name}</td>
              <td class="right num">{line.expected}</td>
              <td class="right num">{line.counted}</td>
              <td class="right num" class:bad={line.counted < line.expected}>{line.counted - line.expected}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  </Modal>
{/if}

<style>
  .gap { gap: 16px; }
  .tight { gap: 2px; }
  .tiles { grid-template-columns: repeat(3, minmax(0, 1fr)); }
  .search { width: 240px; }
  .plan { padding: 12px; border-bottom: var(--hair) solid var(--line); }
  .head { padding: 16px 16px 4px; }
  .product { display: flex; flex-wrap: wrap; align-items: center; gap: 4px 8px; min-width: 200px; }
  .subtotal th, .subtotal td { border-top: var(--hair) solid var(--line); border-bottom: var(--hair) solid var(--line); background: var(--surface-2); font-weight: 600; }
  .at-bar { display: flex; flex-direction: column; gap: 4px; min-width: 88px; }
  .meter { width: 88px; height: 6px; border-radius: 2px; background: var(--surface-3); }
  .level { height: 100%; border-radius: 2px; background: var(--text); }
  .level.low { background: var(--bad); }
  .crates { white-space: nowrap; }
  /* Booking everything at once adds a row per product, which must not bury the counts below. */
  .recent { max-height: 288px; overflow-y: auto; }
</style>
