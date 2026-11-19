<script>
  import AmountDialog from '../components/AmountDialog.svelte';
  import Modal from '../components/Modal.svelte';
  import { get, post } from '../lib/api.js';
  import { euro } from '../lib/format.js';
  import { t } from '../lib/i18n.svelte.js';
  import { toastError } from '../lib/state.svelte.js';

  let { item, onclose, ondone } = $props();

  let room = $state('');
  let name = $state('');
  // svelte-ignore state_referenced_locally
  let deposit = $state(item.depositCents);
  let note = $state('');
  let editing = $state(false);
  let busy = $state(false);
  let known = $state([]); // names used for this room before
  let lookupTimer;

  const ready = $derived(room.trim() !== '' && name.trim() !== '');

  // Names from this room are only offered, never filled in, because a room changes hands.
  function lookup() {
    clearTimeout(lookupTimer);
    known = [];
    const wanted = room.trim();
    if (!wanted) return;
    lookupTimer = setTimeout(async () => {
      try {
        const found = await get('/api/workers/' + encodeURIComponent(wanted));
        if (wanted === room.trim()) known = found.names ?? [];
      } catch {
        // A guest who never worked a shift has no name to suggest.
      }
    }, 250);
  }

  async function submit(e) {
    e.preventDefault();
    if (!ready || busy) return;
    busy = true;
    try {
      await post('/api/rentals', { itemId: item.id, name: name.trim(), room: room.trim(), depositCents: deposit, depositNote: note.trim() });
      ondone();
    } catch (err) {
      toastError(err);
    } finally {
      busy = false;
    }
  }
</script>

{#if editing}
  <AmountDialog
    title={t('rent.deposit')}
    confirmLabel={t('common.save')}
    initialCents={deposit}
    onconfirm={(cents) => {
      deposit = cents;
      editing = false;
    }}
    onclose={() => (editing = false)}
  />
{:else}
  <Modal title={t('rent.lendTitle', { item: item.name })} {onclose}>
    <form class="stack" onsubmit={submit}>
      <div class="pair">
        <label class="field">
          <span>{t('common.room')}</span>
          <input class="input big" bind:value={room} oninput={lookup} inputmode="numeric" autocomplete="off" />
        </label>
        <label class="field">
          <span>{t('common.name')}</span>
          <input class="input big" bind:value={name} autocomplete="off" />
        </label>
      </div>
      {#if known.length && !name.trim()}
        <div class="known">
          <span class="muted">{t('rent.suggest')}</span>
          {#each known as person (person)}
            <button type="button" class="btn" onclick={() => (name = person)}>{person}</button>
          {/each}
        </div>
      {/if}
      <div class="field">
        <span>{t('rent.deposit')}</span>
        <button type="button" class="btn big block num" onclick={() => (editing = true)}>{euro(deposit)}</button>
      </div>
      <label class="field">
        <span>{t('rent.depositOr')}</span>
        <input class="input" bind:value={note} placeholder={t('rent.depositHint')} autocomplete="off" />
      </label>
      {#if deposit > 0}
        <p class="good">{t('rent.takeDeposit', { amount: euro(deposit) })}</p>
      {/if}
      <button class="btn big primary block" disabled={!ready || busy}>{t('rent.lend')}</button>
    </form>
  </Modal>
{/if}

<style>
  .pair { display: grid; grid-template-columns: 110px minmax(0, 1fr); gap: 12px; }
  .known { display: flex; flex-wrap: wrap; align-items: center; gap: 8px; }
</style>
