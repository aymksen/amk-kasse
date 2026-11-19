<script>
  import { onMount } from 'svelte';
  import ScanBarcode from 'lucide-svelte/icons/scan-barcode';
  import LendDialog from './LendDialog.svelte';
  import ReturnDialog from './ReturnDialog.svelte';
  import helicopter from '../assets/sticker-heli.webp';
  import airhockey from '../assets/lend-airhockey.webp';
  import cue from '../assets/lend-cue.webp';
  import darts from '../assets/lend-darts.webp';
  import { get } from '../lib/api.js';
  import { euro, time } from '../lib/format.js';
  import { t } from '../lib/i18n.svelte.js';
  import { listenForScans } from '../lib/scanner.js';
  import { till, toast, toastError } from '../lib/state.svelte.js';
  import { revealHiddenStyle } from '../lib/theme.svelte.js';

  let { initialCode = null, onconsumed } = $props();

  let board = $state(null);
  let lending = $state(null);
  let returning = $state(null);
  let chopper = $state();

  // The picture follows the name the OBK gives an item, so a new kind of thing simply shows none.
  const PICTURES = [[/billard|queue/i, cue], [/dart/i, darts], [/air.?hockey/i, airhockey]];
  const pictureOf = (name) => PICTURES.find(([pattern]) => pattern.test(name))?.[1];

  // Nobody is told about this. Five quick taps on the helicopter put on the next hidden style.
  const TAPS_TO_UNLOCK = 5;
  const TAP_GAP_MS = 1500;
  let taps = 0;
  let lastTap = 0;

  function bob(keyframes, duration) {
    if (matchMedia('(prefers-reduced-motion: reduce)').matches) return;
    chopper.animate(keyframes, { duration, easing: 'cubic-bezier(0.2, 1.3, 0.4, 1)' });
  }

  function tapHelicopter() {
    const now = performance.now();
    taps = now - lastTap < TAP_GAP_MS ? taps + 1 : 1;
    lastTap = now;
    if (taps < TAPS_TO_UNLOCK) {
      bob([{ transform: `rotate(${taps % 2 ? 3 : -3}deg) translateY(-6px)` }, { transform: 'none' }], 300);
      return;
    }
    taps = 0;
    if (revealHiddenStyle()) bob([{ transform: 'none' }, { transform: 'translate(36px, -64px) rotate(-7deg)', offset: 0.45 }, { transform: 'none' }], 1100);
  }

  async function load() {
    try {
      board = await get('/api/rentals');
    } catch (err) {
      toastError(err);
      board = [];
    }
  }

  function open(item) {
    if (item.rental) returning = item;
    else lending = item;
  }

  function scanned(code) {
    if (lending || returning) return;
    const item = board?.find((i) => i.code === code);
    if (item) open(item);
    else toast(t('sell.unknownCode', { code }), 'bad');
  }

  function feeLabel(item) {
    if (item.feeCents === 0) {
      return item.depositCents > 0 ? t('rent.depositOnly', { amount: euro(item.depositCents) }) : t('rent.free2');
    }
    const fee = euro(item.feeCents);
    return item.feePer === 'hour' ? t('rent.perHour', { fee }) : t('rent.flat', { fee });
  }

  function finished(message) {
    lending = null;
    returning = null;
    toast(message, 'good');
    load();
  }

  onMount(() => {
    load().then(() => {
      if (initialCode) {
        scanned(initialCode);
        onconsumed?.();
      }
    });
    return listenForScans({ onScan: scanned });
  });
</script>

<div class="page">
  <header class="spread">
    <h1>{t('rent.title')}</h1>
    <span class="muted row small"><ScanBarcode size={18} /> {t('rent.hint')}</span>
  </header>

  {#if board === null}
    <p class="muted">{t('common.loading')}</p>
  {:else if board.length === 0}
    <p class="empty">{t('rent.none')}</p>
  {:else}
    <div class="grid">
      {#each board as item (item.id)}
        {@const picture = pictureOf(item.name)}
        <button class="item keycap" class:out={item.rental} onclick={() => open(item)}>
          {#if picture}
            <span class="window"><img src={picture} alt="" /></span>
          {/if}
          <span class="body">
            <span class="spread">
              <span class="name">{item.name}</span>
              <span class="tag" class:good={!item.rental} class:warn={item.rental}>{item.rental ? t('rent.out') : t('rent.free')}</span>
            </span>
            {#if item.rental}
              <span>{item.rental.renterName} · {t('common.room')} {item.rental.renterRoom}</span>
              <span class="muted small num">{t('rent.since', { time: time(item.rental.startedAt) })}</span>
            {:else}
              <span class="muted small">{feeLabel(item)}</span>
            {/if}
          </span>
        </button>
      {/each}
    </div>
  {/if}
  <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
  <div class="sticker heli" onclick={tapHelicopter}>
    <img bind:this={chopper} src={helicopter} alt="" width="660" height="503" />
  </div>
</div>

{#if lending}
  <LendDialog item={lending} onclose={() => (lending = null)} ondone={() => finished(t('rent.lent'))} />
{/if}
{#if returning}
  <ReturnDialog item={returning} onclose={() => (returning = null)} ondone={() => finished(t('rent.returned'))} />
{/if}

<style>
  .page { display: flex; flex-direction: column; gap: 16px; height: 100%; padding: 16px; overflow-x: hidden; overflow-y: auto; }
  .grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(260px, 1fr)); gap: 12px 10px; padding: 1px 1px 4px; }
  .item {
    display: flex;
    align-items: center;
    gap: 14px;
    min-height: 104px;
    padding: 12px 16px 12px 12px;
    text-align: left;
  }
  .body { display: flex; flex: 1; flex-direction: column; gap: 4px; min-width: 0; align-self: stretch; justify-content: center; }
  .window {
    flex: none;
    width: 76px;
    height: 76px;
    padding: 4px;
    border-radius: 4px;
    background: #fff;
    box-shadow: 0 0 0 1px rgb(0 0 0 / 0.28);
  }
  .window img { display: block; width: 100%; height: 100%; object-fit: contain; }
  .item.out { --face: var(--warn-bg); --face-hover: var(--warn-hover); }
  .name { font-size: 1.0625rem; font-stretch: 92%; font-weight: 700; }
  .sticker { flex: none; align-self: flex-end; width: min(44vw, 430px); margin-top: auto; transform: rotate(-2deg); touch-action: manipulation; user-select: none; }
  .sticker img { display: block; width: 100%; height: auto; pointer-events: none; }
  @media (max-height: 640px) {
    .sticker { display: none; }
  }
</style>
