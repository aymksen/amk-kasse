<script>
  import Beer from 'lucide-svelte/icons/beer';
  import ClipboardList from 'lucide-svelte/icons/clipboard-list';
  import LogOut from 'lucide-svelte/icons/log-out';
  import ShieldCheck from 'lucide-svelte/icons/shield-check';
  import Target from 'lucide-svelte/icons/target';
  import SlideTabs from '../components/SlideTabs.svelte';
  import BoardButton from '../components/BoardButton.svelte';
  import StylePicker from '../components/StylePicker.svelte';
  import Wordmark from '../components/Wordmark.svelte';
  import Sell from './Sell.svelte';
  import Rentals from './Rentals.svelte';
  import BarCount from './BarCount.svelte';
  import CloseShift from './CloseShift.svelte';
  import HandoverDialog from './HandoverDialog.svelte';
  import { tillSource } from '../games/tillSource.js';
  import { getLanguage, t } from '../lib/i18n.svelte.js';
  import { moments } from '../lib/moments.svelte.js';
  import { route } from '../lib/route.svelte.js';
  import { pickLanguage, till } from '../lib/state.svelte.js';

  const views = $derived([
    { value: 'sell', icon: Beer, label: t('nav.sell') },
    { value: 'rentals', icon: Target, label: t('nav.rentals') },
    { value: 'count', icon: ClipboardList, label: t('nav.count') },
  ]);
  let view = $state('sell');
  // Set when a rental sticker is scanned on the sales screen.
  let rentalCode = $state(null);
  // The games hide behind the logo: seven quick taps open them. From the third tap on, the logo gives a small
  // wobble, so whoever knows can tell it is counting, while a stray tap shows nothing.
  const TAPS_TO_OPEN = 7;
  const TAP_GAP_MS = 1500;
  let taps = 0;
  let lastTap = 0;
  let logo = $state();
  let games = $state.raw(null); // while the games are open: where the logo sits, so their window grows out of it

  function tapLogo() {
    if (games) return;
    const now = performance.now();
    taps = now - lastTap < TAP_GAP_MS ? taps + 1 : 1;
    lastTap = now;
    if (taps >= TAPS_TO_OPEN) {
      taps = 0;
      games = logo.getBoundingClientRect();
      return;
    }
    if (taps < 3 || matchMedia('(prefers-reduced-motion: reduce)').matches) return;
    const tilt = (taps - 2) * 1.5;
    logo.animate([{ transform: `rotate(${taps % 2 ? tilt : -tilt}deg) scale(${1 + (taps - 2) * 0.02})` }, { transform: 'none' }], {
      duration: 280,
      easing: 'cubic-bezier(0.2, 1.3, 0.4, 1)',
    });
  }

  // The dock has its own layer, and the games have to lie over everything, so they move to the page itself.
  function onThePage(node) {
    document.body.append(node);
    return { destroy: () => node.remove() };
  }

  // Counting takes a while without touching the screen, and the screensaver would only get in the way.
  $effect(() => {
    moments.busy = view === 'count' || view === 'close';
    return () => (moments.busy = false);
  });

  function openRental(code) {
    rentalCode = code;
    view = 'rentals';
  }
</script>

<div class="till">
  <main class="content">
    {#if view === 'sell'}
      <Sell onrental={openRental} />
    {:else if view === 'rentals'}
      <Rentals initialCode={rentalCode} onconsumed={() => (rentalCode = null)} />
    {:else if view === 'count'}
      <BarCount ondone={() => (view = 'sell')} />
    {:else}
      <CloseShift oncancel={() => (view = 'sell')} />
    {/if}
  </main>

  <!-- Ending the shift has its own bar on top, with the way back, so nothing down here pulls elsewhere. -->
  {#if view !== 'close'}
  <footer class="dock">
    <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
    <div class="logo" bind:this={logo} onclick={tapLogo}><Wordmark /></div>
    <SlideTabs options={views} value={view} onpick={(id) => (view = id)} label="Kasse" big />
    <div class="grow"></div>
    <div class="worker">
      <strong class="name">{till.shift.workerName}</strong>
      <span class="muted">{t('common.room')} {till.shift.workerRoom}</span>
    </div>
    {#if till.demo}<span class="tag warn">Demo</span>{/if}
    <button class="btn ghost" onclick={() => pickLanguage(getLanguage() === 'de' ? 'en' : 'de')}>{getLanguage() === 'de' ? 'EN' : 'DE'}</button>
    <BoardButton />
    <StylePicker up />
    <button class="btn ghost" onclick={() => route.go('/admin')}><ShieldCheck size={20} /> {t('nav.admin')}</button>
    <button class="btn" onclick={() => (view = 'close')}><LogOut size={20} /> {t('nav.close')}</button>
  </footer>
  {/if}
</div>

{#if games}
  <div use:onThePage>
    {#await import('../games/Arcade.svelte') then { default: Arcade }}
      <Arcade source={tillSource()} from={games} onclose={() => (games = null)} />
    {/await}
  </div>
{/if}

{#if till.message}
  <HandoverDialog message={till.message} onclose={() => (till.message = null)} />
{/if}

<style>
  .till { display: grid; grid-template-columns: minmax(0, 1fr); grid-template-rows: minmax(0, 1fr) auto; height: 100%; }
  .content { min-width: 0; min-height: 0; overflow: hidden; }
  .dock {
    display: flex;
    align-items: center;
    gap: 14px;
    min-height: 68px;
    padding: 8px 14px;
    background: var(--surface);
    box-shadow: inset 0 1px 0 var(--seam), inset 0 2px 0 var(--seam-light);
  }
  /* A long name gives way before anything else in the dock does. */
  .worker { display: flex; flex-direction: column; flex-shrink: 1; min-width: 0; font-size: 0.8125rem; line-height: 1.2; text-align: right; white-space: nowrap; }
  .name { overflow: hidden; text-overflow: ellipsis; }
  /* On a small screen the logo alone names the bar, and the room number keeps its place. */
  @media (max-width: 1140px) {
    .dock :global(.mark .words) { display: none; }
  }
  /* Too narrow for one row: where to go on top, who is working and the rest below. */
  @media (max-width: 900px) {
    .dock { flex-wrap: wrap; row-gap: 4px; }
    .dock .grow { flex: 0 0 100%; height: 0; }
    .worker { margin-right: auto; text-align: left; }
  }
  /* Looks like the plain logo: no pointer, no hover, nothing that gives the games away. */
  .logo { flex: none; touch-action: manipulation; user-select: none; -webkit-tap-highlight-color: transparent; }
  .dock .btn { margin-bottom: 0; }
  .dock .btn.ghost { padding: 0 10px; color: var(--muted); }
  .dock .btn.ghost:hover { color: var(--text); }
</style>
