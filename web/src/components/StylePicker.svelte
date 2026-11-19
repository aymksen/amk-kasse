<script>
  import Check from 'lucide-svelte/icons/check';
  import Moon from 'lucide-svelte/icons/moon';
  import Palette from 'lucide-svelte/icons/palette';
  import Sun from 'lucide-svelte/icons/sun';
  import { t } from '../lib/i18n.svelte.js';
  import { basicLooks, glassModes, glassTints, pickGlass, pickTheme, theme, themeEntry, visibleThemes } from '../lib/theme.svelte.js';

  // `up` opens the list above the button, for the dock at the bottom of the screen. `beside` opens it to the right, out
  // of the OBK's side bar, which would otherwise cut it off.
  let { up = false, beside = false } = $props();

  let open = $state(false);
  let root = $state();
  let spot = $state(undefined);

  function toggle() {
    open = !open;
    if (open && beside) {
      const box = root.getBoundingClientRect();
      spot = `left: ${box.right + 12}px; bottom: ${window.innerHeight - box.bottom}px`;
    }
  }

  function closeOutside(e) {
    if (open && !root.contains(e.target)) open = false;
  }

  const icons = { machine: Moon, light: Sun };

  // Basic and Glass stay open, so their choices are right there.
  function pick(id) {
    pickTheme(id);
    open = id === 'glass' || themeEntry() === 'basic';
  }
</script>

<svelte:window
  onpointerdown={closeOutside}
  onkeydown={(e) => {
    if (e.key === 'Escape') open = false;
  }}
/>

<div class="picker" bind:this={root}>
  <button
    class="btn ghost trigger"
    aria-label={t('style.title')}
    title={t('style.title')}
    aria-haspopup="menu"
    aria-expanded={open}
    onclick={toggle}
  >
    <Palette size={20} />
  </button>
  {#if open}
    <div class="menu" class:up class:beside style={spot}>
      <div class="options" role="menu" aria-label={t('style.title')}>
        <span class="head">{t('style.title')}</span>
        {#each visibleThemes() as item (item.id)}
          <button class="option" role="menuitemradio" aria-checked={themeEntry() === item.id} onclick={() => pick(item.id)}>
            <span class="swatch" style:background={item.colors[0]}><i style:background={item.colors[1]}></i></span>
            <span class="grow">{t(item.label)}</span>
            {#if themeEntry() === item.id}<Check size={18} />{/if}
          </button>
        {/each}
      </div>
      {#if themeEntry() === 'basic'}
        <div class="looks" role="radiogroup" aria-label={t('style.basic')}>
          {#each basicLooks as look (look.id)}
            {@const Icon = icons[look.id]}
            <button class="look" role="radio" aria-checked={theme.id === look.id} onclick={() => pick(look.id)}>
              <Icon size={18} />
              {t(look.label)}
            </button>
          {/each}
        </div>
      {:else if theme.id === 'glass'}
        <div class="colors" role="group" aria-label={t('style.colors')}>
          <span class="head">{t('style.colors')}</span>
          {#each glassModes as mode (mode.id)}
            <div class="dots" role="radiogroup" aria-label={t(mode.label)}>
              {#each glassTints as tint (tint.id)}
                <button
                  class="dot"
                  role="radio"
                  aria-checked={theme.mode === mode.id && theme.tint === tint.id}
                  aria-label={`${t(mode.label)}, ${t(tint.label)}`}
                  title={`${t(mode.label)}, ${t(tint.label)}`}
                  onclick={() => pickGlass(mode.id, tint.id)}
                >
                  <i style:background={`linear-gradient(135deg, ${mode.base} 50%, ${tint.color} 50%)`}></i>
                </button>
              {/each}
            </div>
          {/each}
        </div>
      {/if}
    </div>
  {/if}
</div>

<style>
  .picker { position: relative; flex-shrink: 0; }
  .trigger { width: 44px; margin-bottom: 0; padding: 0; color: var(--muted); }
  .trigger:hover, .trigger[aria-expanded='true'] { color: var(--text); }
  .menu {
    position: absolute;
    top: calc(100% + 8px);
    right: 0;
    z-index: 30;
    width: 196px;
    padding: 8px;
    border: var(--modal-edge);
    border-radius: var(--radius-lg);
    background: var(--surface);
    box-shadow: var(--modal-shadow);
  }
  .menu.up { top: auto; bottom: calc(100% + 16px); }
  .menu.beside { position: fixed; top: auto; right: auto; }
  .options { display: flex; flex-direction: column; gap: 2px; }
  .head { padding: 4px 8px 6px; color: var(--muted); font-size: 0.8125rem; font-weight: 600; }
  .option {
    display: flex;
    align-items: center;
    gap: 10px;
    min-height: 48px;
    padding: 0 8px;
    border: 0;
    border-radius: 6px;
    background: transparent;
    font-weight: 620;
    text-align: left;
  }
  .option:hover { background: var(--surface-2); }
  .option[aria-checked='true'] { background: var(--surface-3); }
  .swatch {
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    width: 28px;
    height: 28px;
    border-radius: 6px;
    box-shadow: 0 0 0 1px var(--line);
  }
  .swatch i { display: block; width: 12px; height: 12px; border-radius: 3px; }
  .looks { display: grid; grid-template-columns: 1fr 1fr; gap: 2px; margin-top: 6px; padding-top: 8px; border-top: 1px solid var(--line); }
  .look {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    min-height: 44px;
    padding: 0 6px;
    border: 0;
    border-radius: 6px;
    background: transparent;
    font-weight: 620;
  }
  .look:hover { background: var(--surface-2); }
  .look[aria-checked='true'] { background: var(--surface-3); }
  .colors { display: flex; flex-direction: column; gap: 6px; margin-top: 6px; padding-top: 8px; border-top: 1px solid var(--line); }
  .dots { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); }
  .dot { display: flex; align-items: center; justify-content: center; height: 44px; padding: 0; border: 0; background: transparent; }
  .dot i { display: block; width: 30px; height: 30px; border-radius: 50%; box-shadow: 0 0 0 1px var(--line); }
  .dot[aria-checked='true'] i { box-shadow: 0 0 0 2px var(--solid), 0 0 0 4px var(--focus); }
</style>
