<script>
  import { onMount } from 'svelte';
  import NumPad from '../components/NumPad.svelte';
  import Wordmark from '../components/Wordmark.svelte';
  import tinSign from '../assets/tin-sign.webp';
  import { post } from '../lib/api.js';
  import { errorText, t } from '../lib/i18n.svelte.js';
  import { listenForScans } from '../lib/scanner.js';

  let { session, onlogin, onback } = $props();

  const creating = $derived(!session.configured);
  let pin = $state('');
  let first = $state(null); // the PIN from step one while confirming a new PIN
  let error = $state('');
  let busy = $state(false);

  const hint = $derived(creating ? (first === null ? t('login.create') : t('login.confirm')) : t('login.enter'));

  function onkey(key) {
    error = '';
    // Not the shared keypad helper: that one drops leading zeros, a PIN may start with 0.
    if (key === 'del') pin = pin.slice(0, -1);
    else if (key === 'clear') pin = '';
    else if (pin.length < 12) pin += key;
  }

  async function submit() {
    if (busy || pin.length < 4) return;
    error = '';
    if (creating && first === null) {
      first = pin;
      pin = '';
      return;
    }
    if (creating && pin !== first) {
      error = t('login.mismatch');
      first = null;
      pin = '';
      return;
    }
    busy = true;
    try {
      await post(creating ? '/api/admin/setup' : '/api/admin/login', { pin });
      onlogin();
    } catch (err) {
      error = errorText(err);
      pin = '';
    } finally {
      busy = false;
    }
  }

  onMount(() =>
    listenForScans({
      onScan() {},
      onKey(key) {
        if (key === 'Enter') submit();
        else if (key === 'Backspace') onkey('del');
        else if (/\d/.test(key)) onkey(key);
      },
    }),
  );
</script>

<main class="wrap">
  <img class="sign" src={tinSign} alt="" width="474" height="651" />
  <div class="panel">
    <Wordmark />
    {#if creating && !session.local}
      <h1>{t('login.createTitle')}</h1>
      <p class="muted">{t('login.remoteSetup')}</p>
    {:else}
      <h1>{creating ? t('login.createTitle') : t('login.title')}</h1>
      <p class="muted hint">{hint}</p>
      <div class="dots" aria-label={`${pin.length}`}>
        {#each { length: Math.max(pin.length, 4) } as _, i}
          <span class="dot" class:filled={i < pin.length}></span>
        {/each}
      </div>
      <p class="bad error" role="alert">{error}</p>
      <NumPad {onkey} compact />
      <button class="btn big primary block" disabled={busy || pin.length < 4} onclick={submit}>
        {creating && first === null ? t('common.next') : t('login.button')}
      </button>
    {/if}
    {#if onback}
      <button class="btn ghost" onclick={onback}>{t('admin.backToTill')}</button>
    {/if}
  </div>
</main>

<style>
  .wrap { position: relative; display: flex; align-items: center; justify-content: center; height: 100%; padding: 16px; overflow-y: auto; }
  /* Hangs on the wall left of the panel, and leaves when there is no wall left. */
  .sign { position: absolute; top: 50%; left: calc(50% - 170px - 32px); width: 270px; height: auto; pointer-events: none; transform: translate(-100%, -50%) rotate(-2deg); }
  @media (max-width: 960px), (max-height: 520px) {
    .sign { display: none; }
  }
  .panel { display: flex; flex-direction: column; gap: 12px; width: 100%; max-width: 340px; }
  .hint { min-height: 2.8em; }
  .dots { display: flex; justify-content: center; gap: 12px; padding: 16px; border-radius: var(--radius); background: var(--led-bg); box-shadow: var(--dent); }
  .dot { width: 14px; height: 14px; border-radius: 50%; background: var(--led-amber-off); }
  .dot.filled { background: var(--led-amber); }
  .error { min-height: 1.4em; font-size: 0.875rem; text-align: center; }
</style>
