<script>
  import { onMount } from 'svelte';
  import X from 'lucide-svelte/icons/x';
  import MoneyInput from './MoneyInput.svelte';
  import ProductThumb from '../components/ProductThumb.svelte';
  import SlideTabs from '../components/SlideTabs.svelte';
  import { get, post } from '../lib/api.js';
  import { dateTime, euro } from '../lib/format.js';
  import { t } from '../lib/i18n.svelte.js';
  import { listenForScans } from '../lib/scanner.js';
  import { refresh, toast, toastError } from '../lib/state.svelte.js';

  let products = $state([]);
  let history = $state([]);
  let supplier = $state('');
  let note = $state('');
  let place = $state('storage');
  let query = $state('');
  let lines = $state([]);
  let busy = $state(false);

  const places = $derived([
    { value: 'storage', label: t('rs.toStorage') },
    { value: 'bar', label: t('rs.toBar') },
  ]);
  const stocked = $derived(products.filter((p) => p.active && p.trackStock));
  const matches = $derived(
    query.trim().length < 2 ? [] : stocked.filter((p) => p.name.toLowerCase().includes(query.trim().toLowerCase())).slice(0, 6),
  );
  const pieces = (line) => Number(line.crates || 0) * line.product.crateSize + Number(line.units || 0);
  const total = $derived(lines.reduce((sum, l) => sum + pieces(l) * l.unitCostCents, 0));
  const ready = $derived(lines.length > 0 && lines.every((l) => pieces(l) > 0));

  function addLine(product, crates, units) {
    query = '';
    const line = lines.find((l) => l.product.id === product.id);
    if (line) {
      line.crates = Number(line.crates || 0) + crates;
      line.units = Number(line.units || 0) + units;
    } else {
      lines.push({ product, crates, units, unitCostCents: product.costCents });
    }
  }

  function scan(code) {
    const product = stocked.find((p) => p.barcodes.includes(code) || p.crateBarcode === code);
    if (!product) {
      toast(t('sell.unknownCode', { code }), 'bad');
    } else if (product.crateBarcode === code) {
      addLine(product, 1, 0);
    } else {
      addLine(product, 0, 1);
    }
  }

  function searchKey(e) {
    if (e.key !== 'Enter') return;
    e.preventDefault();
    if (matches.length === 1) addLine(matches[0], matches[0].crateSize > 1 ? 1 : 0, matches[0].crateSize > 1 ? 0 : 1);
    else if (/^\d{4,}$/.test(query.trim())) {
      scan(query.trim());
      query = '';
    }
  }

  async function book() {
    if (busy || !ready) return;
    busy = true;
    try {
      await post('/api/admin/restocks', {
        supplier: supplier.trim(),
        note: note.trim(),
        place,
        lines: lines.map((l) => ({ productId: l.product.id, qty: pieces(l), unitCostCents: l.unitCostCents })),
      });
      toast(t('rs.booked'), 'good');
      lines = [];
      supplier = '';
      note = '';
      place = 'storage';
      await Promise.all([load(), refresh()]);
    } catch (err) {
      toastError(err);
    } finally {
      busy = false;
    }
  }

  async function load() {
    try {
      [products, history] = await Promise.all([get('/api/admin/products'), get('/api/admin/restocks')]);
    } catch (err) {
      toastError(err);
    }
  }

  onMount(() => {
    load();
    return listenForScans({ onScan: scan });
  });
</script>

<div class="stack gap">
  <h1>{t('rs.title')}</h1>

  <section class="card pad stack">
    <div class="meta">
      <label class="field">
        <span>{t('rs.supplier')}</span>
        <input class="input" bind:value={supplier} placeholder={t('rs.supplierHint')} autocomplete="off" />
      </label>
      <label class="field">
        <span>{t('common.note')} ({t('common.optional')})</span>
        <input class="input" bind:value={note} autocomplete="off" />
      </label>
      <div class="field place">
        <span>{t('rs.place')}</span>
        <SlideTabs options={places} value={place} onpick={(value) => (place = value)} label={t('rs.place')} />
      </div>
    </div>

    <div class="search">
      <input class="input big" bind:value={query} onkeydown={searchKey} placeholder={t('rs.scan')} aria-label={t('rs.scan')} autocomplete="off" />
      {#if matches.length}
        <ul class="results">
          {#each matches as product}
            <li>
              <button onclick={() => addLine(product, product.crateSize > 1 ? 1 : 0, product.crateSize > 1 ? 0 : 1)}>
                <ProductThumb {product} size={28} /> {product.name}
              </button>
            </li>
          {/each}
        </ul>
      {/if}
    </div>

    {#if lines.length}
      <div class="table-wrap">
        <table class="table">
          <thead>
            <tr>
              <th>{t('common.product')}</th>
              <th>{t('rs.crates')}</th>
              <th>{t('rs.units')}</th>
              <th>{t('rs.unitCost')}</th>
              <th class="right">{t('rs.pieces')}</th>
              <th class="right">{t('rs.sum')}</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {#each lines as line (line.product.id)}
              <tr>
                <td><ProductThumb product={line.product} size={28} /> {line.product.name}</td>
                <td><input class="input num short" type="number" min="0" bind:value={line.crates} disabled={line.product.crateSize <= 1} aria-label={t('rs.crates')} /></td>
                <td><input class="input num short" type="number" min="0" bind:value={line.units} aria-label={t('rs.units')} /></td>
                <td class="cost"><MoneyInput bind:cents={line.unitCostCents} aria-label={t('rs.unitCost')} /></td>
                <td class="right num">{pieces(line)}</td>
                <td class="right num">{euro(pieces(line) * line.unitCostCents)}</td>
                <td class="right">
                  <button class="icon-btn" onclick={() => lines.splice(lines.indexOf(line), 1)} aria-label={t('common.remove')}><X size={18} /></button>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
      <div class="spread">
        <strong class="num sum">{euro(total)}</strong>
        <button class="btn big primary" disabled={!ready || busy} onclick={book}>{t('rs.book')}</button>
      </div>
    {:else}
      <p class="empty">{t('rs.empty')}</p>
    {/if}
  </section>

  <section class="card">
    <div class="head"><h2>{t('rs.history')}</h2></div>
    {#if history.length}
      <div class="table-wrap">
        <table class="table">
          <thead>
            <tr>
              <th>{t('common.time')}</th>
              <th>{t('rs.supplier')}</th>
              <th>{t('rs.where')}</th>
              <th>{t('common.product')}</th>
              <th class="right">{t('rs.sum')}</th>
            </tr>
          </thead>
          <tbody>
            {#each history as restock (restock.id)}
              <tr>
                <td class="num">{dateTime(restock.createdAt)}</td>
                <td>{restock.supplier || '–'}</td>
                <td class="muted">{t(`place.${restock.place}`)}</td>
                <td class="muted small">{restock.lines.map((l) => `${l.qty}× ${l.name}`).join(', ')}</td>
                <td class="right num">{euro(restock.totalCostCents)}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {:else}
      <p class="empty">{t('rs.none')}</p>
    {/if}
  </section>
</div>

<style>
  .gap { gap: 16px; }
  .meta { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
  .place { grid-column: span 2; align-items: flex-start; }
  .search { position: relative; }
  .results {
    position: absolute;
    inset: 100% 0 auto 0;
    z-index: 5;
    margin: 4px 0 0;
    padding: 4px;
    border: var(--hair) solid var(--line);
    border-radius: var(--radius);
    background: var(--surface-3);
    list-style: none;
  }
  .results button { width: 100%; min-height: 44px; padding: 0 12px; border: 0; border-radius: 8px; background: transparent; text-align: left; }
  .results button:hover { background: var(--line); }
  .short { width: 84px; }
  .cost { width: 112px; }
  .sum { font-size: 1.5rem; font-weight: 680; }
  .head { padding: 16px 16px 4px; }
</style>
