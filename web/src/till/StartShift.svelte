<script>
  import { onMount } from 'svelte';
  import ShieldCheck from 'lucide-svelte/icons/shield-check';
  import AmountDialog from '../components/AmountDialog.svelte';
  import BoardButton from '../components/BoardButton.svelte';
  import LedNumber from '../components/LedNumber.svelte';
  import SeasonHat from '../components/SeasonHat.svelte';
  import StepBar from '../components/StepBar.svelte';
  import StylePicker from '../components/StylePicker.svelte';
  import StockCheck from './StockCheck.svelte';
  import logoBare from '../assets/logo-bare.png';
  import logoPlain from '../assets/logo.png';
  import { get, post } from '../lib/api.js';
  import { time } from '../lib/format.js';
  import { getLanguage, t } from '../lib/i18n.svelte.js';
  import { wearsHat } from '../lib/moments.svelte.js';
  import { route } from '../lib/route.svelte.js';
  import { pickLanguage, refresh, till, toastError } from '../lib/state.svelte.js';
  import { leaveHiddenStyle } from '../lib/theme.svelte.js';

  let step = $state('who'); // who, stock (the check of the bar), cash
  let room = $state('');
  let name = $state('');
  let editing = $state(false);
  let busy = $state(false);
  let known = $state([]); // names used for this room before
  let lookupTimer;
  // The check of the bar: what it should hold, what the worker changed, and the figures they settled on.
  let opening = $state(null);
  let counts = $state({});
  let checked = $state([]);

  const ready = $derived(room.trim() !== '' && name.trim() !== '');
  const steps = $derived([
    { id: 'who', label: t('open.stepWho') },
    { id: 'stock', label: t('open.stepStock') },
    { id: 'cash', label: t('open.stepCash') },
  ]);

  let clock = $state(time(Date.now()));
  onMount(() => {
    leaveHiddenStyle();
    const timer = setInterval(() => (clock = time(Date.now())), 10000);
    return () => clearInterval(timer);
  });

  // Returning workers can tap their name. It is only offered, never filled in, because a room changes hands.
  function lookup() {
    clearTimeout(lookupTimer);
    known = [];
    const wanted = room.trim();
    if (!wanted) return;
    lookupTimer = setTimeout(async () => {
      try {
        const found = await get('/api/workers/' + encodeURIComponent(wanted));
        if (wanted === room.trim()) known = found.names;
      } catch {
        // No suggestion is fine; the worker just types their name.
      }
    }, 250);
  }

  async function open(floatCents) {
    if (busy) return;
    busy = true;
    try {
      const opened = await post('/api/shifts', { room: room.trim(), name: name.trim(), floatCents, count: checked });
      till.message = opened.message;
      await refresh();
    } catch (err) {
      toastError(err);
    } finally {
      busy = false;
    }
  }

  // The figures are fetched fresh each time, because the till can sit on this screen for a day while the
  // OBK fills the fridges.
  async function next(e) {
    e.preventDefault();
    if (!ready || busy) return;
    busy = true;
    try {
      opening = await get('/api/stock/check');
      checked = [];
      step = opening.lines.length ? 'stock' : 'cash';
    } catch (err) {
      toastError(err);
    } finally {
      busy = false;
    }
  }
</script>

{#if step === 'stock'}
  <div class="wizard">
    <StepBar title={t('open.title')} {steps} index={1} backLabel={t('common.back')} onback={() => (step = 'who')} />
    <div class="stage">
      <StockCheck
        when="start"
        data={opening}
        bind:counts
        ondone={(lines) => {
          checked = lines;
          step = 'cash';
        }}
      />
    </div>
  </div>
{:else}
<main class="page start">
  <div class="top">
    <LedNumber text={clock} digits={4} size={30} />
    <div class="row">
      {#if till.demo}<span class="tag warn">Demo</span>{/if}
      <button class="btn ghost" onclick={() => pickLanguage(getLanguage() === 'de' ? 'en' : 'de')}>
        {getLanguage() === 'de' ? 'English' : 'Deutsch'}
      </button>
      <BoardButton labelled />
      <StylePicker />
      <button class="btn ghost" onclick={() => route.go('/admin')}><ShieldCheck size={18} /> {t('nav.admin')}</button>
    </div>
  </div>

  <div class="center">
    <div class="logo">
      <img src={wearsHat() ? logoBare : logoPlain} alt="AMK Bar, Albertus-Magnus-Kolleg Studentenwohnheim Münster" width="512" height="499" />
      <SeasonHat />
    </div>

    <div class="sheet slip ledger">
      {#if step === 'who'}
        <form class="stack" onsubmit={next}>
          <h1>{t('start.who')}</h1>
          <div class="pair">
            <label class="field">
              <span>{t('start.room')}</span>
              <!-- svelte-ignore a11y_autofocus -->
              <input class="input big" bind:value={room} oninput={lookup} inputmode="numeric" autocomplete="off" autofocus />
            </label>
            <label class="field">
              <span>{t('start.name')}</span>
              <input class="input big" bind:value={name} autocomplete="off" />
            </label>
          </div>
          <div class="known">
            {#if known.length && !name.trim()}
              <span class="muted">{t('start.suggest')}</span>
              {#each known as person (person)}
                <button type="button" class="btn" onclick={() => (name = person)}>{person}</button>
              {/each}
            {/if}
          </div>
          <button class="btn big primary block" disabled={!ready || busy}>{t('common.next')}</button>
        </form>
      {:else}
        <div class="stack">
          <h1>{t('start.checkCash')}</h1>
          <p class="muted">{t('start.shouldBe')}</p>
          <LedNumber cents={till.expectedFloatCents} digits={6} size={48} unit="EUR" />
          <div class="pair even">
            <button class="btn big" onclick={() => (editing = true)}>{t('start.different')}</button>
            <button class="btn big primary" disabled={busy} onclick={() => open(till.expectedFloatCents)}>{t('start.correct')}</button>
          </div>
          <button class="btn ghost" onclick={() => (step = opening?.lines.length ? 'stock' : 'who')}>{t('common.back')}</button>
        </div>
      {/if}
    </div>
  </div>
</main>
{/if}

{#if editing}
  <AmountDialog
    title={t('start.enterAmount')}
    confirmLabel={t('start.openShift')}
    onconfirm={(cents) => {
      editing = false;
      open(cents);
    }}
    onclose={() => (editing = false)}
  />
{/if}

<style>
  .wizard { display: flex; flex-direction: column; height: 100%; }
  .stage { flex: 1; min-height: 0; }
  /* Logo and form are one centred piece: the badge sits on the top edge of the sheet like a seal. */
  .page {
    --logo: min(36vh, 340px);
    --tuck: calc(var(--logo) * 0.2);
    display: flex;
    height: 100%;
    padding: 64px 16px 16px;
    overflow-y: auto;
  }
  .top {
    position: fixed;
    inset: 0 0 auto 0;
    z-index: 2;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    pointer-events: none;
  }
  .top > :global(*) { pointer-events: auto; }
  .center { position: relative; display: flex; flex-direction: column; align-items: center; width: min(100%, 520px); margin: auto; }
  /* A light above the badge. One still picture, so it costs nothing after the first paint. */
  .center::before {
    position: absolute;
    top: calc(var(--logo) * -0.75);
    left: 50%;
    width: calc(var(--logo) * 2.5);
    height: calc(var(--logo) * 2.5);
    margin-left: calc(var(--logo) * -1.25);
    background: radial-gradient(closest-side, var(--accent), transparent);
    content: '';
    opacity: 0.2;
    pointer-events: none;
  }
  :global(:root:not([data-theme])) .center::before { display: none; }
  .logo { position: relative; z-index: 1; height: var(--logo); margin-bottom: calc(var(--tuck) * -1); }
  .logo img { display: block; width: auto; height: 100%; }
  .sheet { width: 100%; padding: 28px; --surface: var(--paper); }
  .sheet .stack { padding-top: var(--tuck); }
  h1 { text-align: center; }
  .sheet .muted { text-align: center; }
  .pair { display: grid; grid-template-columns: 148px minmax(0, 1fr); gap: 12px; }
  .pair.even { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  /* Always this tall, so the form does not jump when a name appears. */
  .known { display: flex; flex-wrap: wrap; align-items: center; gap: 8px; min-height: 48px; }
  @media (max-width: 560px) {
    .pair { grid-template-columns: minmax(0, 1fr); }
    .top > :global(.led) { display: none; }
    .top { justify-content: flex-end; }
  }
</style>
