<script>
  import { onMount } from 'svelte';
  import X from 'lucide-svelte/icons/x';
  import Screensaver from './Screensaver.svelte';
  import { t } from '../lib/i18n.svelte.js';
  import { allowed, endPreview, moments, updateSeason } from '../lib/moments.svelte.js';
  import { till } from '../lib/state.svelte.js';

  // Five minutes with nothing going on and an empty cart, so it never gets in the way of a slow evening.
  const QUIET_MS = 5 * 60 * 1000;
  // A mouse knocked by a sleeve is not someone wanting the till back.
  const PUSH_PX = 40;

  let lastUse = Date.now();
  let pointer = null;

  const touched = () => (lastUse = Date.now());

  // A tap on the screensaver must not reach the key under it, so a tap only counts here. The screensaver ends it itself.
  function woken() {
    touched();
    if (moments.preview !== 'screensaver') moments.screensaver = false;
  }

  function dismissed() {
    touched();
    moments.screensaver = false;
    if (moments.preview === 'screensaver') endPreview();
  }

  function moved(e) {
    if (e.pointerType !== 'mouse') return;
    if (pointer && Math.hypot(e.clientX - pointer.x, e.clientY - pointer.y) < PUSH_PX) return;
    if (pointer) woken();
    pointer = { x: e.clientX, y: e.clientY };
  }

  function watch() {
    updateSeason();
    const open = document.querySelector('.backdrop, .board');
    const quiet = Date.now() - lastUse > QUIET_MS;
    const midnight = moments.season?.phase === 'final' || moments.season?.phase === 'show';
    if (midnight) moments.screensaver = false;
    else if (quiet && allowed() && !till.cart.length && !moments.busy && !open) moments.screensaver = true;
  }

  $effect(() => {
    updateSeason();
  });

  // Lets a stylesheet make room, for the garland at the top.
  $effect(() => {
    const root = document.documentElement;
    if (moments.season) root.dataset.season = moments.season.kind;
    else delete root.dataset.season;
    return () => delete root.dataset.season;
  });

  onMount(() => {
    const timer = setInterval(watch, 5000);
    window.addEventListener('pointermove', moved, { passive: true });
    window.addEventListener('pointerdown', touched, { capture: true, passive: true });
    window.addEventListener('keydown', woken, { capture: true, passive: true });
    return () => {
      clearInterval(timer);
      window.removeEventListener('pointermove', moved);
      window.removeEventListener('pointerdown', touched, { capture: true });
      window.removeEventListener('keydown', woken, { capture: true });
      moments.season = null; // the OBK area stays plain
    };
  });
</script>

{#if moments.season?.kind === 'christmas'}
  {#await import('./Christmas.svelte') then { default: Christmas }}
    <Christmas />
  {/await}
{:else if moments.season?.kind === 'halloween'}
  {#await import('./Halloween.svelte') then { default: Halloween }}
    <Halloween />
  {/await}
{:else if moments.season?.kind === 'newyear'}
  {#await import('./NewYear.svelte') then { default: NewYear }}
    <NewYear />
  {/await}
{/if}

{#if moments.screensaver}
  <Screensaver onwake={dismissed} />
{/if}

{#if moments.preview && moments.preview !== 'date'}
  <button class="btn end" onclick={endPreview}><X size={18} /> {t('moments.endPreview')}</button>
{/if}

<style>
  .end {
    position: fixed;
    top: 8px;
    left: 50%;
    z-index: 80;
    transform: translateX(-50%);
  }
</style>
