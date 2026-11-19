<script>
  import { onDestroy, onMount } from 'svelte';
  import LedNumber from '../components/LedNumber.svelte';
  import SeasonHat from '../components/SeasonHat.svelte';
  import logoBare from '../assets/logo-bare.png';
  import logoPlain from '../assets/logo.png';
  import { get } from '../lib/api.js';
  import { time } from '../lib/format.js';
  import { t } from '../lib/i18n.svelte.js';
  import { clock, moments, wearsHat } from '../lib/moments.svelte.js';
  import { createPath, move, nextCorner } from './bounce.js';

  // The logo drifts around the dark screen like an old DVD player, printed in one colour that changes at every wall.
  // Only in the rare corner does it show its real colours. The first tap only brings the till back, so nobody sells
  // a beer by accident. A scan goes straight through, because a scanner types and this only catches taps.
  let { onwake } = $props();

  const COLORS = ['#ffb224', '#35e0a1', '#3aa0ff', '#ff5fc4', '#ff6a3d', '#b388ff', '#ffe14d', '#7fe0ff'];
  const BOARD_AFTER_MS = 90 * 1000;

  let box = $state();
  let face = $state();
  let party = $state();
  let hour = $state(time(clock()));
  let bullseye = $state(false);
  let showing = $state(null); // the leaderboard, while it is up
  let path;
  let color = 0;
  let frame;

  const logo = $derived(wearsHat() ? logoBare : logoPlain);

  function squash(x, y) {
    const wide = 1.09;
    const thin = 0.9;
    const scale = x ? `${thin}, ${wide}` : `${wide}, ${thin}`;
    const from = x && y ? 'scale(0.86)' : `scale(${scale})`;
    face.animate([{ transform: from }, { transform: 'none' }], { duration: 340, easing: 'cubic-bezier(0.2, 1.5, 0.4, 1)' });
  }

  async function corner(p) {
    bullseye = true;
    setTimeout(() => (bullseye = false), 3000);
    const { default: confetti } = await import('canvas-confetti');
    const shoot = confetti.create(party, { resize: true });
    const origin = { x: p.x > p.maxX / 2 ? 1 : 0, y: p.y > p.maxY / 2 ? 1 : 0 };
    const up = origin.y ? 1 : -1;
    const angle = origin.x ? 90 + up * 45 : 90 - up * 45;
    for (const delay of [0, 180, 360]) {
      setTimeout(() => shoot({ particleCount: 90, spread: 70, startVelocity: 58, angle, origin, ticks: 240, colors: COLORS, disableForReducedMotion: true }), delay);
    }
  }

  function createLogoPath() {
    const width = window.innerWidth;
    const maxX = width - box.offsetWidth;
    const maxY = window.innerHeight - box.offsetHeight;
    const now = performance.now();
    // The OBK should not have to wait ten minutes to see what happens in a corner.
    const preview = moments.preview === 'screensaver';
    return createPath({
      maxX,
      maxY,
      speed: Math.max(110, width * 0.08),
      x: maxX * (preview ? 0.7 : 0.3),
      y: maxY * (preview ? 0.6 : 0.2),
      cornerAt: preview ? 0 : nextCorner(now),
      now,
    });
  }

  function start() {
    path ??= createLogoPath(); // kept across the leaderboard, so the next corner stays on its schedule
    box.style.setProperty('--tint', COLORS[color]);

    let last = performance.now();
    function step(now) {
      const seconds = Math.min(now - last, 50) / 1000;
      last = now;
      const hit = move(path, seconds, now);
      if (hit.wallX || hit.wallY) {
        color = (color + 1) % COLORS.length;
        box.style.setProperty('--tint', COLORS[color]);
        squash(hit.wallX, hit.wallY);
      }
      if (hit.corner) corner(path);
      box.style.transform = `translate3d(${path.x}px, ${path.y}px, 0)`;
      frame = requestAnimationFrame(step);
    }
    frame = requestAnimationFrame(step);
  }

  function pause() {
    cancelAnimationFrame(frame);
  }

  onMount(() => {
    start();
    const clockTimer = setInterval(() => (hour = time(clock())), 10000);
    // The leaderboard of the month fills the screen for a while, then the logo takes over again.
    const boardTimer = setInterval(async () => {
      if (showing) return;
      const recap = await get('/api/recap').catch(() => null);
      if (!recap?.month.qty && !recap?.lastMonth.qty) return;
      pause();
      showing = recap;
    }, BOARD_AFTER_MS);
    return () => {
      clearInterval(clockTimer);
      clearInterval(boardTimer);
      pause();
    };
  });

  onDestroy(pause);

  function boardOver() {
    showing = null;
    start();
  }
</script>

<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
<div class="saver" onclick={onwake}>
  {#if moments.season?.kind === 'christmas'}
    {#await import('./WinterSky.svelte') then { default: WinterSky }}
      <WinterSky />
    {/await}
  {:else if moments.season?.kind === 'halloween'}
    {#await import('./HalloweenSky.svelte') then { default: HalloweenSky }}
      <HalloweenSky />
    {/await}
  {/if}

  <div class="hour"><LedNumber text={hour} digits={4} size={34} plain /></div>

  <div class="box" bind:this={box}>
    <div class="glow"></div>
    <div class="face" bind:this={face}>
      <div class="stamp" class:real={bullseye} style="--logo: url({logo})">
        <img src={logo} alt="" width="512" height="499" />
      </div>
      <SeasonHat />
    </div>
  </div>

  <canvas class="party" bind:this={party}></canvas>
  {#if bullseye}<p class="bullseye">{t('moments.corner')}</p>{/if}
  <p class="wake">{t('moments.wake')}</p>

  {#if showing}
    {#await import('./Leaderboard.svelte') then { default: Leaderboard }}
      <Leaderboard recap={showing} auto onclose={boardOver} />
    {/await}
  {/if}
</div>

<style>
  .saver {
    position: fixed;
    inset: 0;
    z-index: 50;
    overflow: hidden;
    background: #000;
    cursor: none;
    animation: fade 1200ms ease-out;
    touch-action: manipulation;
    user-select: none;
  }
  .box {
    --size: min(17vw, 24vh);
    position: absolute;
    top: 0;
    left: 0;
    width: var(--size);
    height: var(--size);
    will-change: transform;
  }
  .glow {
    position: absolute;
    inset: -30%;
    background: radial-gradient(closest-side, var(--tint), transparent);
    opacity: 0.22;
  }
  .face { position: relative; width: 100%; height: 100%; }

  /* The logo, printed in one ink: a grey copy multiplied over the colour, cut to the outline of the logo. */
  .stamp {
    background: var(--tint);
    transition: background-color 400ms ease-out;
    -webkit-mask: var(--logo) center / 100% 100% no-repeat;
    mask: var(--logo) center / 100% 100% no-repeat;
  }
  .stamp img {
    display: block;
    width: 100%;
    height: auto;
    filter: grayscale(1) contrast(1.2) brightness(1.1);
    mix-blend-mode: multiply;
    transition: filter 400ms ease-out;
  }
  .stamp.real { background: #fff; }
  .stamp.real img { filter: none; }

  .hour {
    --led-bg: transparent;
    --led-amber: #ffb224;
    --led-amber-off: rgb(255 178 36 / 0.05);
    position: absolute;
    top: 16px;
    right: 20px;
    opacity: 0.7;
  }
  .party { position: absolute; inset: 0; width: 100%; height: 100%; pointer-events: none; }
  .bullseye {
    position: absolute;
    top: 42%;
    left: 0;
    width: 100%;
    color: #fff6e5;
    font-family: var(--display);
    font-size: clamp(2rem, 6vw, 4.5rem);
    font-stretch: 125%;
    font-weight: 900;
    letter-spacing: -0.02em;
    text-align: center;
    animation: corner-word 3000ms ease-out both;
  }
  .wake {
    position: absolute;
    bottom: 18px;
    left: 0;
    width: 100%;
    color: rgb(255 255 255 / 0.55);
    font-size: 0.9375rem;
    text-align: center;
  }
  @keyframes corner-word {
    0% { opacity: 0; transform: scale(0.7); }
    12% { opacity: 1; transform: scale(1.06); }
    20% { transform: none; }
    80% { opacity: 1; }
    100% { opacity: 0; }
  }
</style>
