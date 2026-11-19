<script>
  import LedNumber from './LedNumber.svelte';
  import Modal from './Modal.svelte';
  import NumPad from './NumPad.svelte';
  import { press } from '../lib/format.js';
  import { t } from '../lib/i18n.svelte.js';
  import { listenForScans } from '../lib/scanner.js';

  let { title, confirmLabel, initialCents = 0, onconfirm, onclose } = $props();

  // svelte-ignore state_referenced_locally
  let digits = $state(initialCents > 0 ? String(initialCents) : '');
  const cents = $derived(Number(digits || 0));

  function onkey(key) {
    digits = press(digits, key === '00' ? '0' : key, 7);
    if (key === '00') digits = press(digits, '0', 7);
  }

  function keyboard(node) {
    return {
      destroy: listenForScans({
        onScan() {},
        onKey(key) {
          if (key === 'Enter') onconfirm(cents);
          else if (key === 'Backspace') onkey('del');
          else if (/\d/.test(key)) onkey(key);
        },
      }),
    };
  }
</script>

<Modal {title} {onclose}>
  <div class="stack" use:keyboard>
    <LedNumber {cents} digits={7} size={56} unit="EUR" />
    <NumPad {onkey} doubleZero />
    <div class="row">
      <button class="btn big grow" onclick={() => onkey('del')}>{t('common.back')}</button>
      <button class="btn big primary grow" onclick={() => onconfirm(cents)}>{confirmLabel}</button>
    </div>
  </div>
</Modal>
