<script>
  import Minus from 'lucide-svelte/icons/minus';
  import Plus from 'lucide-svelte/icons/plus';
  import LedNumber from './LedNumber.svelte';
  import { t } from '../lib/i18n.svelte.js';

  // − figure +: how every count at the till changes a number. Tapping the figure asks for the whole number.
  // With a keyboard, up and down go to the figure of the row above or below, left and right change it by one.
  // Digits and Enter on a focused figure are left to the screen, which tells them apart from a scan.
  // id marks the figure, so the screen can find its row. name labels the buttons for a screen reader.
  // tone colours the figure: amber as expected, green for more, red for less. onfocus hears when the figure
  // gets the focus, by a tap or by the arrow keys, for a screen that marks the row being typed into.
  let { id, value, name, onchange, ontype, onfocus = null, tone = 'amber', digits = 3, max = 9999 } = $props();

  function keys(e) {
    if (e.key === 'ArrowUp' || e.key === 'ArrowDown') {
      e.preventDefault();
      const all = [...document.querySelectorAll('[data-count-figure]')];
      const next = all[all.indexOf(e.currentTarget) + (e.key === 'ArrowUp' ? -1 : 1)];
      next?.focus();
      next?.scrollIntoView({ block: 'nearest' });
    } else if (e.key === 'ArrowLeft' || e.key === '-') {
      e.preventDefault();
      if (value > 0) onchange(value - 1);
    } else if (e.key === 'ArrowRight' || e.key === '+') {
      e.preventDefault();
      if (value < max) onchange(value + 1);
    }
  }
</script>

<div class="stepper">
  <button class="keycap step" aria-label={t('check.less', { name })} disabled={value <= 0} onclick={() => onchange(value - 1)}>
    <Minus size={22} strokeWidth={2.6} />
  </button>
  <button class="figure" data-count-figure={id} aria-label={t('check.type', { name, n: value })} onclick={() => ontype()} onkeydown={keys} onfocus={() => onfocus?.()}>
    <LedNumber text={String(value)} {digits} size={30} plain {tone} />
  </button>
  <button class="keycap step" aria-label={t('check.more', { name })} disabled={value >= max} onclick={() => onchange(value + 1)}>
    <Plus size={22} strokeWidth={2.6} />
  </button>
</div>

<style>
  .stepper { display: flex; flex: none; align-items: center; gap: 6px; }
  .step { display: grid; place-items: center; width: 56px; height: 48px; padding: 0; }
  .figure {
    display: grid;
    place-items: center;
    min-width: 76px;
    height: 48px;
    padding: 0 6px;
    border: 0;
    border-radius: var(--radius);
    background: var(--led-bg);
    box-shadow: var(--dent);
    cursor: pointer;
  }
  .figure:focus-visible { outline: 3px solid var(--focus); outline-offset: 2px; }
</style>
