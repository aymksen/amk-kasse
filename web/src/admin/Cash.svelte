<script>
  import { onMount } from 'svelte';
  import SlideTabs from '../components/SlideTabs.svelte';
  import MoneyInput from './MoneyInput.svelte';
  import { get, post } from '../lib/api.js';
  import { dateTime, euro } from '../lib/format.js';
  import { t } from '../lib/i18n.svelte.js';
  import { toast, toastError } from '../lib/state.svelte.js';

  const kinds = $derived([
    { value: 'withdrawal', label: t('cash.withdraw') },
    { value: 'topup', label: t('cash.topup') },
  ]);

  let data = $state(null);
  let kind = $state('withdrawal');
  let amount = $state(0);
  let note = $state('');
  let busy = $state(false);

  async function load() {
    try {
      data = await get('/api/admin/cash');
    } catch (err) {
      toastError(err);
    }
  }

  async function submit(e) {
    e.preventDefault();
    if (busy || amount <= 0) return;
    busy = true;
    try {
      await post('/api/admin/cash', { kind, amountCents: amount, note: note.trim() });
      toast(t('cash.done'), 'good');
      amount = 0;
      note = '';
      await load();
    } catch (err) {
      toastError(err);
    } finally {
      busy = false;
    }
  }

  onMount(load);
</script>

<div class="stack gap">
  <h1>{t('admin.cash')}</h1>

  <div class="top">
    <section class="card pad box">
      <span class="muted small">{t('cash.box')}</span>
      <strong class="hero num">{data ? euro(data.boxCents) : '…'}</strong>
      <span class="muted small">{t('cash.boxHint')}</span>
    </section>

    <form class="card pad stack" onsubmit={submit}>
      <SlideTabs options={kinds} value={kind} onpick={(value) => (kind = value)} label={t('cash.withdraw')} />
      <div class="pair">
        <label class="field">
          <span>{t('common.amount')} (€)</span>
          <MoneyInput bind:cents={amount} />
        </label>
        <label class="field">
          <span>{t('common.note')}</span>
          <input class="input" bind:value={note} placeholder={t('cash.noteHint')} autocomplete="off" />
        </label>
      </div>
      <button class="btn big primary" disabled={busy || amount <= 0}>{kind === 'withdrawal' ? t('cash.withdraw') : t('cash.topup')}</button>
    </form>
  </div>

  <section class="card">
    <div class="head"><h2>{t('cash.history')}</h2></div>
    {#if data?.moves.length}
      <div class="table-wrap">
        <table class="table">
          <thead>
            <tr>
              <th>{t('common.time')}</th>
              <th>{t('st.kind')}</th>
              <th>{t('common.note')}</th>
              <th class="right">{t('common.amount')}</th>
            </tr>
          </thead>
          <tbody>
            {#each data.moves as move (move.id)}
              <tr>
                <td class="num">{dateTime(move.createdAt)}</td>
                <td>{t('cash.kind.' + move.kind)}</td>
                <td class="muted">{move.note || '–'}</td>
                <td class="right num">{euro(move.amountCents)}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {:else}
      <p class="empty">{t('cash.none')}</p>
    {/if}
  </section>
</div>

<style>
  .gap { gap: 16px; }
  .top { display: grid; grid-template-columns: repeat(auto-fit, minmax(320px, 1fr)); gap: 16px; }
  .box { display: flex; flex-direction: column; justify-content: center; gap: 4px; }
  .hero { font-size: 3rem; font-weight: 720; letter-spacing: -0.02em; line-height: 1.1; }
  .pair { display: grid; grid-template-columns: 130px minmax(0, 1fr); gap: 12px; }
  .head { padding: 16px 16px 4px; }
</style>
