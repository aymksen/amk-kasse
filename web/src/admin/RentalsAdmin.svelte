<script>
  import { onMount } from 'svelte';
  import Plus from 'lucide-svelte/icons/plus';
  import MoneyInput from './MoneyInput.svelte';
  import Modal from '../components/Modal.svelte';
  import { get, post } from '../lib/api.js';
  import { dateTime, euro } from '../lib/format.js';
  import { t } from '../lib/i18n.svelte.js';
  import { toast, toastError } from '../lib/state.svelte.js';

  let items = $state(null);
  let history = $state([]);
  let form = $state(null);
  let busy = $state(false);

  async function load() {
    try {
      [items, history] = await Promise.all([get('/api/admin/rental-items'), get('/api/admin/rental-history')]);
    } catch (err) {
      toastError(err);
    }
  }

  async function save(e) {
    e.preventDefault();
    if (busy) return;
    busy = true;
    try {
      await post('/api/admin/rental-items', form);
      toast(t('common.saved'), 'good');
      form = null;
      await load();
    } catch (err) {
      toastError(err);
    } finally {
      busy = false;
    }
  }

  const feeText = (item) => `${euro(item.feeCents)} ${item.feePer === 'hour' ? '/ h' : ''}`;

  onMount(load);
</script>

<div class="stack gap">
  <header class="spread wrap">
    <h1>{t('ri.title')}</h1>
    <button class="btn primary" onclick={() => (form = { id: 0, name: '', code: '', feeCents: 0, feePer: 'flat', depositCents: 0, active: true })}>
      <Plus size={18} /> {t('ri.new')}
    </button>
  </header>

  <section class="card">
    {#if items === null}
      <p class="empty">{t('common.loading')}</p>
    {:else}
      <div class="table-wrap">
        <table class="table">
          <thead>
            <tr>
              <th>{t('common.name')}</th>
              <th>{t('ri.code')}</th>
              <th class="right">{t('ri.fee')}</th>
              <th class="right">{t('ri.deposit')}</th>
            </tr>
          </thead>
          <tbody>
            {#each items as item (item.id)}
              <tr class="click" onclick={() => (form = { ...item })}>
                <td>
                  {item.name}
                  {#if !item.active}<span class="tag">{t('common.inactive')}</span>{/if}
                </td>
                <td class="muted num">{item.code || '–'}</td>
                <td class="right num">{feeText(item)}</td>
                <td class="right num">{euro(item.depositCents)}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  </section>

  <section class="card">
    <div class="head"><h2>{t('ri.history')}</h2></div>
    {#if history.length}
      <div class="table-wrap">
        <table class="table">
          <thead>
            <tr>
              <th>{t('ri.item')}</th>
              <th>{t('ri.renter')}</th>
              <th>{t('common.from')}</th>
              <th>{t('ri.returned')}</th>
              <th class="right">{t('ri.fee')}</th>
            </tr>
          </thead>
          <tbody>
            {#each history as rental (rental.id)}
              <tr>
                <td>{rental.itemName}</td>
                <td>{rental.renterName} ({rental.renterRoom})</td>
                <td class="num">{dateTime(rental.startedAt)}</td>
                <td class="num">{rental.returnedAt ? dateTime(rental.returnedAt) : ''}</td>
                <td class="right num">{rental.returnedAt ? euro(rental.feeCents) : ''}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {:else}
      <p class="empty">{t('ri.none')}</p>
    {/if}
  </section>
</div>

{#if form}
  <Modal title={form.id ? t('ri.edit') : t('ri.new')} onclose={() => (form = null)}>
    <form class="stack" onsubmit={save}>
      <label class="field">
        <span>{t('common.name')}</span>
        <input class="input" bind:value={form.name} required autocomplete="off" />
      </label>
      <label class="field">
        <span>{t('ri.code')}</span>
        <input class="input num" bind:value={form.code} inputmode="numeric" autocomplete="off" onkeydown={(e) => e.key === 'Enter' && e.preventDefault()} />
      </label>
      <p class="muted small">{t('ri.codeHint')}</p>
      <div class="pair">
        <label class="field">
          <span>{t('ri.fee')} (€)</span>
          <MoneyInput bind:cents={form.feeCents} />
        </label>
        <label class="field">
          <span>{t('ri.per')}</span>
          <select class="input" bind:value={form.feePer}>
            <option value="flat">{t('ri.flat')}</option>
            <option value="hour">{t('ri.hour')}</option>
          </select>
        </label>
      </div>
      <label class="field">
        <span>{t('ri.deposit')} (€)</span>
        <MoneyInput bind:cents={form.depositCents} />
      </label>
      <label class="check"><input type="checkbox" bind:checked={form.active} /> {t('common.active')}</label>
      <button class="btn big primary" disabled={busy}>{t('common.save')}</button>
    </form>
  </Modal>
{/if}

<style>
  .gap { gap: 16px; }
  .head { padding: 16px 16px 4px; }
  .pair { display: grid; grid-template-columns: 110px minmax(0, 1fr); gap: 12px; }
</style>
