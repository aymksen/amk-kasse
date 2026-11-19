<script>
  import { onMount } from 'svelte';
  import ArrowRight from 'lucide-svelte/icons/arrow-right';
  import LedNumber from '../components/LedNumber.svelte';
  import NumPad from '../components/NumPad.svelte';
  import Stepper from '../components/Stepper.svelte';
  import { euro, press } from '../lib/format.js';
  import { t } from '../lib/i18n.svelte.js';
  import { listenForScans } from '../lib/scanner.js';

  // The money at the end of a shift, two ways. The quick one is picked from the start: type the total on the
  // keypad, which only ever types the total. The careful one is note by note: tap a row and type how many on
  // the keyboard, and Enter or Tab goes on to the next one, 50 € to 20 € to 10 € and down to the cents.
  // − and + work on every row. Whatever was touched last counts: the typed total, or the notes and coins.
  // Unlike the bar it stays blind: nothing here says what should be in the box.
  // counts maps a note or coin (in cents) to how many there are. typed is the total in cents as typed, or null
  // while the notes and coins count. Both live in the screen around this one, so a recount starts from the
  // first count. ondone gets the total in cents.
  let { counts = $bindable({}), typed = $bindable(''), ondone } = $props();

  const groups = [
    { label: () => t('money.notes'), values: [5000, 2000, 1000, 500] },
    { label: () => t('money.coins'), values: [200, 100, 50, 20, 10, 5, 2, 1] },
  ];
  const all = groups.flatMap((g) => g.values);

  // The note or coin the keyboard types into, while the notes and coins count. A recount by notes and coins
  // starts at the top again.
  // svelte-ignore state_referenced_locally
  let target = $state(typed === null ? all[0] : null);
  let fresh = $state(true); // the next digit starts the figure anew instead of adding to it
  let goButton = $state();

  const byTotal = $derived(typed !== null);
  const coinSum = $derived(all.reduce((sum, d) => sum + d * (counts[d] ?? 0), 0));
  const total = $derived(byTotal ? Number(typed || 0) : coinSum);

  function set(value, n) {
    n = Math.max(0, Math.min(9999, n));
    if (n === 0) delete counts[value];
    else counts[value] = n;
  }

  function figure(value) {
    return document.querySelector(`[data-count-figure="${value}"]`);
  }

  // Counting note by note from here on.
  function select(value, focus = false) {
    if (target !== value || byTotal) fresh = true;
    typed = null;
    target = value;
    if (focus) {
      figure(value)?.focus();
      figure(value)?.scrollIntoView({ block: 'nearest' });
    }
  }

  // Typing the total from here on. It starts from what the notes and coins came to, and the first digit
  // replaces that.
  function pickTotal() {
    if (byTotal) return;
    typed = coinSum ? String(coinSum) : '';
    target = null;
    fresh = true;
    if (document.activeElement?.hasAttribute('data-count-figure')) document.activeElement.blur();
  }

  function typeTotal(key) {
    pickTotal();
    const from = fresh && key !== 'del' ? '' : typed;
    typed = key === '00' ? press(press(from, '0', 7), '0', 7) : press(from, key, 7);
    fresh = false;
  }

  function typeCount(key) {
    if (target === null) select(all[0], true);
    const now = counts[target] ? String(counts[target]) : '';
    set(target, Number(press(fresh && key !== 'del' ? '' : now, key, 4) || 0));
    fresh = false;
  }

  // Enter and Tab go down the notes and coins, Shift+Tab back up. After the last coin comes the button that
  // goes on.
  function step(by) {
    const following = all[all.indexOf(target) + by];
    if (following !== undefined) select(following, true);
    else if (by > 0) goButton?.focus();
  }

  function tab(e) {
    if (e.key !== 'Tab' || byTotal || target === null || !document.activeElement?.hasAttribute('data-count-figure')) return;
    if (e.shiftKey && target === all[0]) return;
    e.preventDefault();
    step(e.shiftKey ? -1 : 1);
  }

  onMount(() => {
    const stop = listenForScans({
      onScan() {},
      onKey(key) {
        const onButton = document.activeElement instanceof HTMLButtonElement && !document.activeElement.hasAttribute('data-count-figure');
        if (key === 'Enter') {
          // The scanner listener takes Enter from every button, so a focused button gets it back here.
          if (onButton) document.activeElement.click();
          else if (!byTotal) step(1);
        } else if (key === 'Backspace') {
          byTotal ? typeTotal('del') : typeCount('del');
        } else if (/\d/.test(key)) {
          byTotal ? typeTotal(key) : typeCount(key);
        }
      },
    });
    window.addEventListener('keydown', tab, true);
    return () => {
      stop();
      window.removeEventListener('keydown', tab, true);
    };
  });
</script>

<div class="cash-screen">
  <div class="scroll">
    <div class="inner">
      <div class="intro">
        <h1>{t('close.cashTitle')}</h1>
        <p class="muted">{t('close.cashHint')}</p>
      </div>

      {#each groups as group (group.values[0])}
        <section class="kind">
          <h2>{group.label()}</h2>
          <ul>
            {#each group.values as value (value)}
              {@const n = counts[value] ?? 0}
              <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_noninteractive_element_interactions -->
              <li class="line" class:has={n > 0} class:target={target === value} onclick={() => select(value, true)}>
                <strong class="face num">{euro(value)}</strong>
                <span class="sum num muted">{n > 0 ? euro(value * n) : ''}</span>
                <Stepper
                  id={value}
                  value={n}
                  name={euro(value)}
                  onchange={(next) => {
                    select(value);
                    set(value, next);
                    fresh = true;
                  }}
                  ontype={() => select(value, true)}
                  onfocus={() => select(value)}
                />
              </li>
            {/each}
          </ul>
        </section>
      {/each}
    </div>
  </div>

  <aside class="panel">
    <div class="sums">
      <button class="total" class:target={byTotal} onclick={pickTotal} aria-label={t('money.typeTotal')}>
        <LedNumber label={byTotal ? t('money.total') : t('money.sum')} cents={total} digits={7} size={40} unit="EUR" />
      </button>
      <p class="muted small">{byTotal ? t('money.totalHint') : t('money.coinHint')}</p>
    </div>
    <div class="pad"><NumPad onkey={typeTotal} doubleZero compact /></div>
    <div class="go">
      <!-- With nothing counted, the button says what going on would mean, so nobody skips the count by mistake. -->
      {#if total > 0}
        <button class="btn big primary block" bind:this={goButton} onclick={() => ondone(total)}>{t('common.next')} <ArrowRight size={20} /></button>
      {:else}
        <button class="btn big block" bind:this={goButton} onclick={() => ondone(0)}>{t('close.empty')}</button>
      {/if}
    </div>
  </aside>
</div>

<style>
  .cash-screen { display: grid; grid-template-columns: minmax(0, 1fr) minmax(320px, 380px); height: 100%; }
  .scroll { min-height: 0; overflow-y: auto; overscroll-behavior: contain; }
  .inner { width: min(100%, 720px); margin: 0 auto; padding: 0 16px 16px; }
  .intro { display: flex; flex-direction: column; gap: 4px; padding: 20px 0 8px; }
  .intro p { max-width: 60ch; margin: 0; }
  .kind { padding-bottom: 8px; }
  .kind h2 { margin: 0; padding: 14px 4px 8px; font-size: 1rem; font-weight: 750; }
  ul { display: flex; flex-direction: column; gap: 6px; margin: 0; padding: 0; list-style: none; }
  .line {
    display: flex;
    align-items: center;
    gap: 12px;
    min-height: 60px;
    padding: 6px 8px 6px 16px;
    border-radius: var(--radius);
    background: var(--surface);
    box-shadow: 0 0 0 1px var(--line);
    cursor: pointer;
  }
  .line.has {
    background: color-mix(in srgb, var(--accent) 12%, var(--surface));
    box-shadow: 0 0 0 1px color-mix(in srgb, var(--accent) 35%, var(--line)), inset 4px 0 0 var(--accent);
  }
  .line.target { box-shadow: 0 0 0 2px var(--accent-text); }
  .line.has.target { box-shadow: 0 0 0 2px var(--accent-text), inset 4px 0 0 var(--accent); }
  .face { min-width: 92px; font-size: 1.125rem; font-weight: 750; }
  .sum { flex: 1; text-align: right; }

  .panel {
    display: flex;
    flex-direction: column;
    gap: 12px;
    min-height: 0;
    padding: 16px;
    background: var(--surface);
    box-shadow: inset 1px 0 0 var(--line);
    overflow-y: auto;
  }
  .sums { display: flex; flex-direction: column; gap: 6px; }
  .sums p { margin: 0; }
  .total { display: block; width: 100%; padding: 0; border: 0; border-radius: var(--radius); background: none; text-align: left; cursor: pointer; }
  .total.target { outline: 2px solid var(--accent-text); outline-offset: 2px; }
  .total:focus-visible { outline: 3px solid var(--focus); outline-offset: 2px; }
  .go { display: flex; flex-direction: column; gap: 8px; margin-top: auto; }

  /* Upright screen: the notes and coins on top, the keypad under them within reach, beside the total. */
  @media (orientation: portrait) {
    .cash-screen { grid-template-columns: minmax(0, 1fr); grid-template-rows: minmax(0, 1fr) auto; }
    .panel {
      display: grid;
      grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
      grid-template-areas: 'sums pad' 'go pad';
      align-items: start;
      box-shadow: inset 0 1px 0 var(--line);
    }
    .sums { grid-area: sums; }
    .pad { grid-area: pad; }
    .go { grid-area: go; margin-top: 0; }
  }
  @media (orientation: portrait) and (max-width: 560px) {
    .panel { grid-template-columns: minmax(0, 1fr); grid-template-areas: 'sums' 'pad' 'go'; }
  }
</style>
