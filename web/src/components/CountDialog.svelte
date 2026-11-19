<script>
  import LedNumber from './LedNumber.svelte';
  import Modal from './Modal.svelte';
  import NumPad from './NumPad.svelte';
  import { press } from '../lib/format.js';
  import { t } from '../lib/i18n.svelte.js';
  import { listenForScans } from '../lib/scanner.js';

  // Typing a whole figure, for when more than one or two are different. before is the figure on the row now,
  // first a digit already typed on the row, hint a line under the figure. children goes beside the display,
  // like the product's picture.
  let { title, before, first = '', hint = '', onpick, onclose, children } = $props();

  // svelte-ignore state_referenced_locally
  let digits = $state(first);

  function take() {
    if (digits === '') onclose();
    else onpick(Number(digits));
  }

  function keyboard() {
    return {
      destroy: listenForScans({
        onScan() {},
        onKey(key) {
          if (key === 'Enter') take();
          else if (key === 'Backspace') digits = press(digits, 'del', 4);
          else if (/\d/.test(key)) digits = press(digits, key, 4);
        },
      }),
    };
  }
</script>

<Modal {title} {onclose}>
  <div class="stack" use:keyboard>
    <div class="typed">
      {@render children?.()}
      <LedNumber text={digits || '0'} digits={4} size={56} />
    </div>
    <p class="muted small">{t('check.before', { n: before })}{#if hint}. {hint}{/if}</p>
    <NumPad onkey={(key) => (digits = press(digits, key, 4))} />
    <div class="buttons">
      <button class="btn big" onclick={onclose}>{t('common.cancel')}</button>
      <button class="btn big primary" disabled={digits === ''} onclick={take}>{t('check.take')}</button>
    </div>
  </div>
</Modal>

<style>
  .typed { display: flex; align-items: center; justify-content: space-between; gap: 16px; }
  .typed > :global(:only-child) { margin-left: auto; }
  .buttons { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; }
</style>
