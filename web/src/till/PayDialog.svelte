<script>
  import AmountDialog from '../components/AmountDialog.svelte';
  import LedNumber from '../components/LedNumber.svelte';
  import Modal from '../components/Modal.svelte';
  import { euro } from '../lib/format.js';
  import { keyColors } from '../lib/keycolor.js';
  import { t } from '../lib/i18n.svelte.js';
  import { till, toast } from '../lib/state.svelte.js';

  let { total, onpay, onclose } = $props();

  const notes = [
    { cents: 500, key: keyColors('#9c9a93') },
    { cents: 1000, key: keyColors('#dc6a50') },
    { cents: 2000, key: keyColors('#5a8fd0') },
    { cents: 5000, key: keyColors('#e8933a') },
  ];
  let custom = $state(false);

  const payout = $derived(total < 0);
  const usefulNotes = $derived(notes.filter((n) => n.cents > total));
  const other = $derived(till.settings.otherPayment);
</script>

{#if custom}
  <AmountDialog
    title={t('sell.received2')}
    confirmLabel={t('sell.pay')}
    initialCents={0}
    onconfirm={(cents) => {
      if (cents >= total) onpay('cash', cents);
      else toast(t('sell.tooLittle', { amount: euro(total) }), 'bad');
    }}
    onclose={() => (custom = false)}
  />
{:else}
  <Modal title={payout ? t('sell.toPayOut') : t('sell.toPay')} {onclose}>
    <div class="stack">
      <LedNumber cents={Math.abs(total)} digits={6} size={60} unit="EUR" />
      {#if payout}
        <button class="btn big primary block" data-autofocus onclick={() => onpay('cash', 0)}>{t('sell.payout')}</button>
      {:else}
        <button class="btn big primary block" data-autofocus onclick={() => onpay('cash', total)}>{t('sell.exact')}</button>
        <div class="notes">
          {#each usefulNotes as note}
            <button class="btn big num note" style:--face={note.key.face} style:--face-hover={note.key.hover} style:--lip={note.key.lip} onclick={() => onpay('cash', note.cents)}>{euro(note.cents)}</button>
          {/each}
        </div>
        <button class="btn big block" onclick={() => (custom = true)}>{t('sell.otherAmount')}</button>
        {#if other}
          <button class="btn big block" onclick={() => onpay('other', 0)}>{t('sell.confirmOther', { name: other })}</button>
        {/if}
      {/if}
    </div>
  </Modal>
{/if}

<style>
  .notes { display: grid; grid-auto-flow: column; grid-auto-columns: 1fr; gap: 8px; }
  .note { padding: 0 6px; color: var(--tile-ink); }
</style>
