<script>
  import { onMount } from 'svelte';
  import History from 'lucide-svelte/icons/history';
  import Minus from 'lucide-svelte/icons/minus';
  import Plus from 'lucide-svelte/icons/plus';
  import Trash2 from 'lucide-svelte/icons/trash-2';
  import LedNumber from '../components/LedNumber.svelte';
  import ProductThumb from '../components/ProductThumb.svelte';
  import SlideTabs from '../components/SlideTabs.svelte';
  import PayDialog from './PayDialog.svelte';
  import RecentSales from './RecentSales.svelte';
  import ReceiptSlot from '../moments/ReceiptSlot.svelte';
  import SixSeven from './SixSeven.svelte';
  import Devil from './Devil.svelte';
  import FourTwenty from './FourTwenty.svelte';
  import { post } from '../lib/api.js';
  import { businessToday, day, euro } from '../lib/format.js';
  import { t } from '../lib/i18n.svelte.js';
  import { keyColors } from '../lib/keycolor.js';
  import { pictureUrl } from '../lib/picture.js';
  import { listenForScans } from '../lib/scanner.js';
  import { addToCart, cartTotal, changeLine, findByCode, refresh, till, toast, toastError } from '../lib/state.svelte.js';

  let { onrental } = $props();

  let category = $state(null);
  let paying = $state(false);
  let showRecent = $state(false);
  let paid = $state(null); // what the last guest paid, shown until the next order
  let busy = false;
  let sixSeven = $state(false); // the Easter egg for a total of exactly 6,70 €
  let devil = $state(false); // the Easter egg for 6,66 €, 66,60 € and 666,00 €
  let fourTwenty = $state(false); // the Easter egg for 4,20 €, 42,00 € and 420,00 €
  let lastTotal = cartTotal();
  // Windows' "Animationseffekte" switch turns every animation off, so this one stays away too.
  const calm = matchMedia('(prefers-reduced-motion: reduce)');
  // In cents. All prices so far end in 0 or 5 cents, so 6,66 € cannot come up yet, and the devil only shows at 66,60 €.
  const DEVIL_TOTALS = [666, 6660, 66600];
  const FOUR_TWENTY_TOTALS = [420, 4200, 42000];

  const visible = $derived(category === null ? till.products : till.products.filter((p) => p.categoryId === category));
  const keys = $derived(Object.fromEntries(till.categories.map((c) => [c.id, keyColors(c.color)])));
  const tabs = $derived([
    { value: null, label: t('common.all') },
    ...till.categories.map((c) => ({ value: c.id, label: c.name, color: keys[c.id].face })),
  ]);
  const total = $derived(cartTotal());
  // After a sale the display shows what matters next: the change to hand back.
  const display = $derived(
    paid && !till.cart.length
      ? paid.change > 0
        ? { label: t('sell.change'), cents: paid.change, tone: 'green' }
        : { label: paid.amount < 0 ? t('sell.paidOut') : t('sell.paid'), cents: Math.abs(paid.amount), tone: 'green' }
      : { label: t('common.total'), cents: total, tone: 'amber' },
  );

  // Plays when the total arrives at 6,70 € and stops the moment it leaves, so it never stands in a sale's way.
  $effect(() => {
    if (total === lastTotal) return;
    lastTotal = total;
    const funny = !calm.matches;
    sixSeven = total === 670 && funny;
    devil = funny && DEVIL_TOTALS.includes(total);
    fourTwenty = funny && FOUR_TWENTY_TOTALS.includes(total);
  });

  const inCart = (product) => till.cart.find((l) => l.product.id === product.id)?.qty ?? 0;

  function add(product, qty = 1) {
    paid = null;
    addToCart(product, qty);
  }

  function clearCart() {
    till.cart.length = 0;
  }

  async function pay(payment, tenderedCents) {
    if (busy) return;
    busy = true;
    const amount = total;
    try {
      await post('/api/sales', {
        lines: till.cart.map((l) => ({ productId: l.product.id, qty: l.qty })),
        payment,
      });
      paying = false;
      clearCart();
      paid = { amount, payment, tendered: tenderedCents, change: tenderedCents ? tenderedCents - amount : 0 };
      await refresh();
    } catch (err) {
      toastError(err);
    } finally {
      busy = false;
    }
  }

  onMount(() =>
    listenForScans({
      onScan(code) {
        if (paying || showRecent) return;
        const found = findByCode(code);
        if (found) add(found.product, found.qty);
        else if (/^9\d{4}$/.test(code)) onrental(code);
        else toast(t('sell.unknownCode', { code }), 'bad');
      },
    }),
  );
</script>

<div class="sell">
  <section class="products" aria-label={t('nav.sell')}>
    <SlideTabs options={tabs} value={category} onpick={(id) => (category = id)} label={t('common.category')} />
    <div class="grid">
      {#each visible as product (product.id)}
        {@const key = keys[product.categoryId]}
        {@const picture = pictureUrl(product)}
        <button
          class="tile"
          class:out={product.level === 'empty'}
          class:pictured={picture}
          onclick={() => add(product)}
          style:--face={key?.face}
          style:--face-hover={key?.hover}
          style:--lip={key?.lip}
        >
          {#if picture}
            <span class="window"><img src={picture} alt="" /></span>
          {/if}
          <span class="name" class:long={product.name.length > 20}>{product.name}</span>
          <span class="meta">
            <span class="price num">{euro(product.priceCents)}</span>
            {#if product.level === 'empty'}
              <span class="tag">{t('sell.out')}</span>
            {:else if product.level === 'low'}
              <span class="tag">{t('sell.low')}</span>
            {/if}
          </span>
          {#if inCart(product)}
            <span class="count num">{inCart(product)}</span>
          {/if}
        </button>
      {/each}
    </div>
  </section>

  <aside class="register" aria-label={t('sell.pay')}>
    <LedNumber label={display.label} cents={display.cents} tone={devil ? 'red' : fourTwenty ? 'green' : display.tone} unit="EUR" digits={6} size={50} dance={sixSeven} />
    <div class="slot"></div>

    <div class="tray">
      <div class="receipt slip" class:printed={paid && !till.cart.length}>
        <header class="receipt-head">
          <strong>{till.settings.barName}</strong>
          <span class="num">{day(businessToday())} · {till.shift.workerName}</span>
        </header>
        <div class="lines">
          {#if till.cart.length}
            {#each till.cart as line (line.product.id)}
              <div class="line">
                <ProductThumb product={line.product} size={34} />
                <div class="grow">
                  <div class="line-name">{line.product.name}</div>
                  <div class="line-sum num">{euro(line.qty * line.product.priceCents)}</div>
                </div>
                <button class="step" onclick={() => changeLine(line, -1)} aria-label="−"><Minus size={16} /></button>
                <span class="qty num">{line.qty}</span>
                <button class="step" onclick={() => changeLine(line, 1)} aria-label="+"><Plus size={16} /></button>
              </div>
            {/each}
          {:else if paid}
            <div class="done" role="status">
              <span class="stamp">{paid.amount < 0 ? t('sell.paidOut') : t('sell.paid')}</span>
              <dl class="num">
                <dt>{t('common.total')}</dt>
                <dd>{euro(Math.abs(paid.amount))}</dd>
                {#if paid.change > 0}
                  <dt>{t('sell.received')}</dt>
                  <dd>{euro(paid.tendered)}</dd>
                  <dt>{t('sell.change')}</dt>
                  <dd>{euro(paid.change)}</dd>
                {/if}
              </dl>
              <p>{t('sell.next')}</p>
            </div>
          {:else}
            <p class="hint">{t('sell.empty')}</p>
            <ReceiptSlot />
          {/if}
        </div>
      </div>
    </div>

    <div class="actions">
      <div class="row">
        <button class="btn grow" onclick={() => (showRecent = true)}><History size={18} /> {t('sell.recent')}</button>
        <button class="btn clear" disabled={!till.cart.length} onclick={clearCart} aria-label={t('sell.clear')}><Trash2 size={18} /></button>
      </div>
      <button class="btn big primary block" disabled={!till.cart.length} onclick={() => { sixSeven = false; devil = false; fourTwenty = false; paying = true; }}>
        {total < 0 ? t('sell.payout') : t('sell.pay')}
      </button>
    </div>
  </aside>
</div>

{#if paying}
  <PayDialog {total} onpay={pay} onclose={() => (paying = false)} />
{/if}
{#if showRecent}
  <RecentSales onclose={() => (showRecent = false)} />
{/if}
{#if sixSeven}
  <SixSeven line={t('sell.sixSeven')} onend={() => (sixSeven = false)} />
{/if}
{#if devil}
  <Devil line={t('sell.devil')} onend={() => (devil = false)} />
{/if}
{#if fourTwenty}
  <FourTwenty line={t('sell.fourTwenty')} onend={() => (fourTwenty = false)} />
{/if}

<style>
  .sell { display: grid; grid-template-columns: minmax(0, 1fr) 312px; height: 100%; }
  .products { display: flex; flex-direction: column; gap: 14px; min-height: 0; padding: 14px; }
  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
    grid-auto-rows: 150px;
    gap: 12px 10px;
    overflow-y: auto;
    padding: 1px 1px 5px;
  }

  .tile {
    position: relative;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    padding: 12px 12px 10px;
    border: 0;
    border-radius: var(--radius);
    background: var(--face);
    box-shadow: inset 0 1px 0 rgb(255 255 255 / 0.16), 0 3px 0 var(--lip);
    color: var(--tile-ink);
    text-align: left;
  }
  .tile:hover { background: var(--face-hover); }
  .tile:active { box-shadow: inset 0 1px 0 rgb(255 255 255 / 0.16), 0 1px 0 var(--lip); transform: translateY(2px); }
  .tile.out { --face: var(--surface-2) !important; --face-hover: var(--surface-3) !important; --lip: var(--seam) !important; color: var(--muted); }
  /* A paper label under the keycap, the way old tills marked their keys. */
  .tile.pictured { justify-content: flex-start; gap: 5px; padding: 6px 8px 7px; }
  .window {
    flex: 1;
    min-height: 0;
    margin: 0 -2px;
    padding: 4px;
    border-radius: 4px;
    background: #fff;
    box-shadow: 0 0 0 1px rgb(0 0 0 / 0.28);
  }
  .window img { display: block; width: 100%; height: 100%; object-fit: contain; }
  .out .window img { opacity: 0.35; }
  .name {
    display: -webkit-box;
    flex-shrink: 0;
    padding-right: 26px;
    overflow: hidden;
    font-size: 1.0625rem;
    font-stretch: 92%;
    font-weight: 700;
    line-height: 1.2;
    text-wrap: balance;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 2;
  }
  /* One line, so every picture is the same size. Long names get narrower letters, not cut off. */
  .pictured .name { display: block; padding-right: 0; font-size: 0.9375rem; text-overflow: ellipsis; white-space: nowrap; }
  .pictured .long { font-family: var(--narrow); font-stretch: 75%; }
  .meta { display: flex; align-items: center; justify-content: space-between; flex-shrink: 0; gap: 8px; min-height: 22px; }
  .price { font-size: 0.9375rem; font-weight: 600; }
  .tile .tag { background: var(--tile-tag); color: inherit; }
  .count {
    position: absolute;
    top: 8px;
    right: 8px;
    min-width: 28px;
    padding: 3px 4px;
    border-radius: 6px;
    background: var(--ink);
    box-shadow: 0 0 0 1px var(--cream);
    color: var(--cream);
    font-size: 0.875rem;
    font-weight: 600;
    line-height: 1.2;
    text-align: center;
  }

  .register {
    display: flex;
    flex-direction: column;
    min-height: 0;
    padding: 12px 12px 8px;
    background: var(--surface);
    box-shadow: inset 1px 0 0 var(--seam), inset 2px 0 0 var(--seam-light);
  }
  .slot { flex-shrink: 0; height: 5px; margin: 10px 2px 0; border-radius: 3px; background: var(--seam); }
  .tray { flex: 1; min-height: 0; padding: 0 12px 7px; overflow: hidden; }

  .receipt {
    display: flex;
    flex-direction: column;
    max-height: 100%;
    margin-top: -2px;
    font-family: var(--mono);
    font-size: 0.875rem;
    --surface-3: var(--rule);
  }
  .receipt-head {
    display: flex;
    flex-direction: column;
    align-items: center;
    flex-shrink: 0;
    padding: 12px 10px 8px;
    border-bottom: 1px dashed var(--rule);
    text-align: center;
  }
  .receipt-head strong { font-family: var(--display); font-size: 1.0625rem; font-stretch: 118%; font-weight: var(--display-weight); letter-spacing: 0.02em; text-transform: uppercase; }
  .receipt-head span { color: var(--ink-soft); font-size: 0.75rem; }
  .lines { min-height: 0; padding: 2px 10px 6px; overflow-y: auto; }
  .line { display: flex; align-items: center; gap: 4px; padding: 7px 0; border-bottom: 1px dashed var(--rule); }
  .line:last-child { border-bottom: 0; }
  /* A small picture of what was tapped, so a glance at the slip shows the right bottles. It sits on paper, so its
     edge is the paper's rule, not the machine's. */
  .line :global(.thumb) { margin-right: 4px; box-shadow: 0 0 0 1px var(--rule); }
  .line-name { font-weight: 600; line-height: 1.2; overflow-wrap: anywhere; text-wrap: balance; }
  .line-sum { color: var(--ink-soft); font-size: 0.8125rem; }
  .qty { min-width: 24px; font-weight: 600; text-align: center; }
  .step {
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    width: 36px;
    height: 36px;
    border: 1px solid var(--rule);
    border-radius: 6px;
    background: #fffdf6;
    box-shadow: 0 2px 0 var(--rule);
    color: var(--ink);
  }
  .step:active { box-shadow: none; transform: translateY(2px); }
  .hint { padding: 16px 6px 12px; color: var(--ink-soft); text-align: center; }

  .done { position: relative; display: flex; flex-direction: column; gap: 8px; padding: 12px 2px 8px; }
  .stamp {
    align-self: center;
    padding: 2px 12px;
    border: 3px solid #b5301b;
    border-radius: 6px;
    color: #b5301b;
    font-family: var(--display);
    font-size: 1.375rem;
    font-stretch: 112%;
    font-weight: 800;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    transform: rotate(-4deg);
  }
  dl { display: grid; grid-template-columns: 1fr auto; gap: 2px 12px; margin: 0; }
  dd { margin: 0; font-weight: 600; text-align: right; }
  .done p { color: var(--ink-soft); font-size: 0.8125rem; text-align: center; }

  .actions { display: flex; flex-direction: column; gap: 10px; flex-shrink: 0; padding-top: 6px; }
  .clear { padding: 0 14px; }

  @media (max-width: 1100px) {
    .sell { grid-template-columns: minmax(0, 1fr) 296px; }
    .grid { grid-template-columns: repeat(auto-fill, minmax(140px, 1fr)); }
  }
  @media (min-width: 1500px) {
    .sell { grid-template-columns: minmax(0, 1fr) 400px; }
    .grid { grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); grid-auto-rows: 196px; }
    .name, .pictured .name { font-size: 1.125rem; }
    .price { font-size: 1.0625rem; }
  }

  /* Upright screen: the same layout as on a wide screen, with a narrower register and keys that use the height. */
  @media (orientation: portrait) {
    .sell { grid-template-columns: minmax(0, 1fr) clamp(264px, 28vw, 400px); }
    .grid { flex: 1; min-height: 0; grid-template-columns: repeat(auto-fill, minmax(130px, 1fr)); grid-auto-rows: minmax(172px, 240px); }
    /* Narrow keys, so a name gets two lines. Every key keeps room for both, which keeps the pictures the same size. */
    .pictured .name { display: -webkit-box; min-height: 2.4em; white-space: normal; }
  }
</style>
