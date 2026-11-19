<script>
  import { onMount, untrack } from 'svelte';

  // options: [{ value, label, icon?, color? }]. The marker takes the option's colour if it has one.
  let { options, value, onpick, label, big = false } = $props();

  let track = $state();
  let x = $state(0);
  let width = $state(0);
  let shown = $state(false);
  let placed = $state(false);

  const active = $derived(options.find((o) => o.value === value));

  // Plain numbers on purpose: setting the same value twice wakes nothing, and `shown` is read
  // untracked, so measuring can never re-trigger itself.
  function measure() {
    const tab = track?.querySelector('[aria-pressed="true"]');
    placed = Boolean(tab) && untrack(() => shown);
    shown = Boolean(tab);
    if (tab) {
      x = tab.offsetLeft;
      width = tab.offsetWidth;
    }
  }

  $effect(() => {
    // Labels too: switching language changes the width of a tab.
    value;
    options.map((o) => o.label);
    measure();
  });

  onMount(() => {
    // Tab widths change once the font has loaded and when the window is resized.
    document.fonts?.ready.then(measure);
    const observer = new ResizeObserver(measure);
    observer.observe(track);
    return () => observer.disconnect();
  });
</script>

<div
  class="tabs"
  class:big
  role="group"
  aria-label={label}
  bind:this={track}
  style:--marker={active?.color ?? 'var(--text)'}
  style:--marker-text={active?.color ? 'var(--tile-ink)' : 'var(--bg)'}
>
  <span
    class="marker"
    class:shown
    class:placed
    style:width="{width}px"
    style:transform="translateX({x}px)"
  ></span>
  {#each options as option (option.value)}
    <button type="button" class="tab" aria-pressed={option.value === value} onclick={() => onpick(option.value)}>
      {#if option.icon}<option.icon size={20} />{/if}
      {option.label}
    </button>
  {/each}
</div>

<style>
  .tabs {
    position: relative;
    display: flex;
    flex-shrink: 0;
    gap: 2px;
    max-width: 100%;
    padding: 4px;
    border-radius: 10px;
    background: var(--well);
    box-shadow: var(--dent), var(--dent-edge);
    overflow-x: auto;
    scrollbar-width: none;
  }
  :global(.ledger) .tabs { background: var(--surface-3); box-shadow: inset 0 1px 0 var(--line); }
  .marker {
    position: absolute;
    top: 4px;
    bottom: 4px;
    left: 0;
    border-radius: 7px;
    background: var(--marker);
    box-shadow: inset 0 1px 0 rgb(255 255 255 / 0.14);
    opacity: 0;
  }
  .marker.shown { opacity: 1; }
  .marker.placed { transition: transform 130ms ease-out; }
  .tab {
    position: relative;
    display: inline-flex;
    align-items: center;
    gap: 8px;
    min-height: 40px;
    /* Tighter on an upright screen, so the last category is not hidden behind a scroll. */
    padding: 0 clamp(8px, 1.25vw, 16px);
    border: 0;
    border-radius: 7px;
    background: transparent;
    color: var(--muted);
    font-weight: 620;
    white-space: nowrap;
  }
  .big .tab { min-height: 44px; padding: 0 clamp(10px, 1.4vw, 18px); }
  .tab:hover { color: var(--text); }
  .tab[aria-pressed='true'] { color: var(--marker-text); }
</style>
