<script>
  import { onMount } from 'svelte';
  import CircleCheck from 'lucide-svelte/icons/circle-check';
  import TriangleAlert from 'lucide-svelte/icons/triangle-alert';
  import CashCount from './CashCount.svelte';
  import LentReminder from './LentReminder.svelte';
  import StockCheck from './StockCheck.svelte';
  import StepBar from '../components/StepBar.svelte';
  import farmer from '../assets/sticker-farmer.webp';
  import { get, post } from '../lib/api.js';
  import { euro } from '../lib/format.js';
  import { t } from '../lib/i18n.svelte.js';
  import { refresh, till, toast, toastError } from '../lib/state.svelte.js';
  import { leaveHiddenStyle } from '../lib/theme.svelte.js';

  let { oncancel } = $props();

  const workerName = till.shift.workerName;

  let step = $state(null); // null until it is clear whether anything is still lent out
  let lent = $state([]);
  let remind = $state(false); // something was out when the wizard opened, so the reminder is one of its steps
  // The bar, checked the same way as at the start: what should be there, and what the worker changed.
  let check = $state(null);
  let stockCounts = $state({});
  let checked = $state([]); // saved together with the end of the shift
  let cashCounts = $state({}); // note or coin in cents -> how many
  let cashTyped = $state(''); // the total in cents as typed, or null when the notes and coins were counted
  let countedCents = $state(0);
  let totals = $state(null);
  let note = $state('');
  let handover = $state('');
  let busy = $state(false);

  const difference = $derived(totals ? countedCents - totals.expectedCents : 0);
  // A till that took nothing cannot be off, so the stamp has to be earned.
  const calm = matchMedia('(prefers-reduced-motion: reduce)');
  const perfect = $derived(difference === 0 && totals?.salesCount > 0 && !calm.matches);
  const steps = $derived([
    ...(remind ? [{ id: 'lent', label: t('close.stepLent') }] : []),
    ...(check?.lines.length ? [{ id: 'count', label: t('close.step1') }] : []),
    { id: 'cash', label: t('close.step2') },
    { id: 'result', label: t('close.step3') },
  ]);
  const index = $derived(steps.findIndex((item) => item.id === step));

  async function lentOut() {
    const board = await get('/api/rentals');
    return board.filter((item) => item.rental);
  }

  // Neither the reminder nor the check of the bar may stand between a worker and the end of their shift:
  // when they cannot be loaded, the shift ends without them.
  onMount(async () => {
    [lent, check] = await Promise.all([lentOut().catch(() => []), get('/api/stock/check').catch(() => null)]);
    remind = lent.length > 0;
    step = remind ? 'lent' : afterLent();
  });

  const afterLent = () => (check?.lines.length ? 'count' : 'cash');

  async function returned() {
    toast(t('rent.returned'), 'good');
    lent = await lentOut().catch(() => []);
    if (!lent.length) step = afterLent();
  }

  async function showResult(cents) {
    countedCents = cents;
    try {
      totals = await get('/api/shift/totals');
      step = 'result';
    } catch (err) {
      toastError(err);
    }
  }

  async function finish() {
    if (busy) return;
    busy = true;
    try {
      await post('/api/shift/close', { countedCents, note: note.trim(), handover: handover.trim(), count: checked });
      leaveHiddenStyle(); // A style found with the helicopter lasts one shift, and this one is over.
      step = 'done';
    } catch (err) {
      toastError(err);
    } finally {
      busy = false;
    }
  }
</script>

{#if step !== 'done'}
  <div class="wizard">
    <StepBar title={t('close.title')} {steps} {index} backLabel={t('close.back')} onback={oncancel} />

    <div class="stage">
      {#if step === 'lent'}
        <LentReminder items={lent} onreturned={returned} onnext={() => (step = afterLent())} />
      {:else if step === 'count'}
        <StockCheck
          when="end"
          data={check}
          bind:counts={stockCounts}
          ondone={(lines) => {
            checked = lines;
            step = 'cash';
          }}
        />
      {:else if step === 'cash'}
        <CashCount bind:counts={cashCounts} bind:typed={cashTyped} ondone={showResult} />
      {:else if step === 'result'}
        <div class="result">
          <div class="summary slip ledger" class:slammed={perfect}>
            {#if perfect}<span class="perfect" aria-hidden="true">{t('close.perfect')}</span>{/if}
            {#if difference === 0}
              <div class="verdict good"><CircleCheck size={28} /> {t('close.exact')}</div>
            {:else}
              <div class="verdict warn">
                <TriangleAlert size={28} />
                {difference < 0
                  ? t('close.tooLittle', { amount: euro(-difference) })
                  : t('close.tooMuch', { amount: euro(difference) })}
              </div>
            {/if}
            <dl>
              <dt>{t('close.sales')}</dt>
              <dd class="num">{totals.salesCount}</dd>
              <dt>{t('close.cashSales')}</dt>
              <dd class="num">{euro(totals.cashCents)}</dd>
              {#if totals.otherCents}
                <dt>{till.settings.otherPayment}</dt>
                <dd class="num">{euro(totals.otherCents)}</dd>
              {/if}
              <dt>{t('close.expected')}</dt>
              <dd class="num">{euro(totals.expectedCents)}</dd>
              <dt>{t('close.counted')}</dt>
              <dd class="num">{euro(countedCents)}</dd>
            </dl>
            {#if difference !== 0}
              <label class="field">
                <span>{t('close.noteLabel')} ({t('common.optional')})</span>
                <input class="input" bind:value={note} autocomplete="off" />
              </label>
            {/if}
            <label class="field">
              <span>{t('close.handover')} ({t('common.optional')})</span>
              <input class="input" bind:value={handover} maxlength="280" placeholder={t('close.handoverHint')} autocomplete="off" />
            </label>
            <div class="row">
              <button class="btn big grow" onclick={() => (step = 'cash')}>{t('close.recount')}</button>
              <button class="btn big primary grow" disabled={busy} onclick={finish}>{t('close.finish')}</button>
            </div>
          </div>
        </div>
      {/if}
    </div>
  </div>
{:else}
  <div class="thanks">
    <img class="sticker farmer" src={farmer} alt="" width="516" height="550" />
    <h1>{t('close.thanks', { name: workerName })}</h1>
    <p class="muted">{t('close.thanksBody')}</p>
    <button class="btn big primary" data-autofocus onclick={refresh}>{t('close.done')}</button>
  </div>
{/if}

<style>
  .wizard { display: flex; flex-direction: column; height: 100%; }
  .stage { flex: 1; min-height: 0; }

  .result { display: flex; align-items: center; justify-content: center; height: 100%; padding: 16px; overflow-y: auto; }
  .summary { position: relative; display: flex; flex-direction: column; gap: 16px; width: 100%; max-width: 460px; padding: 24px; --surface: var(--paper); }
  .verdict { display: flex; align-items: center; gap: 10px; font-size: 1.5rem; font-weight: 680; }
  .verdict.warn { color: var(--warn); }
  dl { display: grid; grid-template-columns: 1fr auto; gap: 6px 16px; margin: 0; }
  dt { color: var(--muted); }
  dd { margin: 0; text-align: right; }

  /* The stamp comes down hard, and the slip gives way under it. */
  .perfect {
    position: absolute;
    top: 12px;
    right: 10px;
    z-index: 1;
    padding: 2px 10px;
    border: 4px solid currentColor;
    border-radius: 8px;
    box-shadow: inset 0 0 0 2px var(--paper), inset 0 0 0 4px currentColor;
    color: #c8202c;
    font-family: var(--display);
    font-size: 2rem;
    font-stretch: 112%;
    font-weight: 900;
    letter-spacing: 0.04em;
    line-height: 1.15;
    transform: rotate(-11deg);
    pointer-events: none;
    -webkit-mask-image: var(--worn-ink);
    mask-image: var(--worn-ink);
    --worn-ink: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='220' height='90'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='.55' numOctaves='3' seed='8'/%3E%3CfeColorMatrix values='0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 -2.6 2.2'/%3E%3C/filter%3E%3Crect width='220' height='90' filter='url(%23n)'/%3E%3C/svg%3E");
    animation: perfect-down 340ms cubic-bezier(0.3, 0, 0.2, 1) 260ms backwards;
  }
  .summary.slammed { animation: slip-bump 260ms ease-out 400ms; }
  @keyframes perfect-down {
    0% { opacity: 0; transform: rotate(-20deg) scale(3.4); }
    45% { opacity: 1; }
    62% { transform: rotate(-9deg) scale(0.9); }
    100% { transform: rotate(-11deg) scale(1); }
  }
  @keyframes slip-bump {
    0% { transform: none; }
    30% { transform: translateY(5px) scale(0.992); }
    100% { transform: none; }
  }

  .thanks { display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 12px; height: 100%; padding: 16px; text-align: center; }
  .sticker { width: auto; height: min(42vh, 340px); transform: rotate(-2deg); }
</style>
