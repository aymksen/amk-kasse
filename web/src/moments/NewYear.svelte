<script>
  import { onMount, untrack } from 'svelte';
  import confetti from 'canvas-confetti';
  import { Fireworks } from 'fireworks-js';
  import LedNumber from '../components/LedNumber.svelte';
  import { t } from '../lib/i18n.svelte.js';
  import { clock, moments } from '../lib/moments.svelte.js';
  import { SHOW_MS, seasonAt } from '../lib/seasons.js';

  // Midnight. The last minute counts down on a big display, the last ten seconds fill the screen, and at zero the
  // sky goes off. It all lies over the till without catching anything: taps and scans go on working underneath.
  const FADE_MS = 3000;

  let season = $state(seasonAt(clock()));
  let sky = $state();
  let paper = $state();
  let fading = $state(false);

  const phase = $derived(season?.kind === 'newyear' ? season.phase : null);
  const seconds = $derived(season ? Math.ceil(season.left / 1000) : 0);
  const finale = $derived(phase === 'final' && seconds <= 10);
  const glow = $derived(phase === 'show' ? t('moments.happyShort') : t('moments.until', { year: season?.year }));

  // Colours of the sky at midnight, gold first.
  const COLORS = ['#ffd24a', '#ffffff', '#ff5247', '#55b6ff', '#4cd67c', '#ff5fc4'];

  function startShow() {
    const fireworks = new Fireworks(sky, {
      autoresize: true,
      opacity: 0.45,
      acceleration: 1.03,
      friction: 0.97,
      gravity: 1.4,
      particles: 110,
      traceLength: 3,
      traceSpeed: 9,
      explosion: 8,
      intensity: 130, // the opening volley, calmed down below
      flickering: 50,
      lineStyle: 'round',
      hue: { min: 0, max: 360 },
      delay: { min: 22, max: 42 },
      rocketsPoint: { min: 15, max: 85 },
      lineWidth: { explosion: { min: 1, max: 3.4 }, trace: { min: 1, max: 2 } },
      brightness: { min: 55, max: 85 },
      decay: { min: 0.012, max: 0.022 },
      mouse: { click: false, move: false, max: 1 },
      sound: { enabled: false },
    });
    // Not launch(): that one fires its rockets and then stops the whole show.
    fireworks.start();

    const shoot = confetti.create(paper, { resize: true });
    const cannons = (spread) => {
      shoot({ particleCount: 70, angle: 62, spread, startVelocity: 62, origin: { x: 0, y: 0.9 }, colors: COLORS, ticks: 320, gravity: 0.8 });
      shoot({ particleCount: 70, angle: 118, spread, startVelocity: 62, origin: { x: 1, y: 0.9 }, colors: COLORS, ticks: 320, gravity: 0.8 });
    };
    const timers = [0, 450, 900, 1500, 2600, 4200].map((delay) => setTimeout(() => cannons(58), delay));
    timers.push(setTimeout(() => fireworks.updateOptions({ intensity: 42 }), 3000));
    const left = SHOW_MS + (season?.left ?? 0); // joining late, for example after a restart at 00:00:20
    timers.push(setTimeout(() => (fading = true), Math.max(0, left - FADE_MS)));

    return () => {
      timers.forEach(clearTimeout);
      fireworks.stop(true);
      shoot.reset();
    };
  }

  $effect(() => {
    // Untracked, or the clock ticking inside it would start the show again ten times a second.
    if (phase === 'show' && sky) return untrack(startShow);
  });

  onMount(() => {
    const timer = setInterval(() => (season = seasonAt(clock())), 100);
    return () => clearInterval(timer);
  });
</script>

<div class="newyear" class:finale class:show={phase === 'show'} class:fading aria-hidden="true">
  {#if phase === 'final' || phase === 'show'}
    <div class="dim"></div>
  {/if}

  {#if phase === 'final' && !finale}
    <div class="panel">
      <span>{t('moments.until', { year: season.year })}</span>
      <LedNumber text={String(seconds)} digits={2} size={92} plain />
    </div>
  {:else if finale}
    {#key seconds}
      <div class="count"><LedNumber text={String(seconds)} digits={seconds === 10 ? 2 : 1} size={Math.round(Math.min(window.innerHeight * 0.56, window.innerWidth * 0.4))} plain /></div>
    {/key}
  {/if}

  {#if phase === 'show'}
    <div class="sky" bind:this={sky}></div>
    <canvas class="paper" bind:this={paper}></canvas>
    <div class="wish">
      <LedNumber text={String(season.year)} digits={4} size={Math.round(Math.min(window.innerHeight * 0.3, window.innerWidth * 0.16))} plain />
      <p>{glow}</p>
    </div>
  {/if}
</div>

<style>
  .newyear {
    --led-bg: transparent;
    --led-amber: #ffd24a;
    --led-amber-off: rgb(255 210 74 / 0.06);
    position: fixed;
    inset: 0;
    z-index: 60;
    overflow: hidden;
    color: #fff6e5;
    pointer-events: none;
  }
  .dim { position: absolute; inset: 0; background: #05040a; opacity: 0; transition: opacity 900ms ease-out; }
  .finale .dim { opacity: 0.88; }
  .show .dim { opacity: 0.86; }
  .fading .dim, .fading .wish, .fading .sky, .fading .paper { opacity: 0; transition: opacity 3000ms ease-in; }

  .panel {
    position: absolute;
    top: 14px;
    left: 50%;
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 8px 22px;
    border-radius: var(--radius-lg);
    background: rgb(6 5 12 / 0.9);
    box-shadow: 0 0 0 1px rgb(255 210 74 / 0.4), 0 10px 40px rgb(0 0 0 / 0.5);
    font-size: 1.125rem;
    font-weight: 700;
    transform: translateX(-50%);
    animation: drop 500ms cubic-bezier(0.2, 1.3, 0.4, 1);
    white-space: nowrap;
  }

  .count {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    filter: drop-shadow(0 0 30px rgb(255 210 74 / 0.55));
    animation: tick 900ms cubic-bezier(0.15, 0.9, 0.25, 1) both;
  }

  .sky, .paper { position: absolute; inset: 0; width: 100%; height: 100%; }
  .wish {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 10px;
    filter: drop-shadow(0 0 26px rgb(255 210 74 / 0.6));
    animation: arrive 900ms cubic-bezier(0.2, 1.4, 0.4, 1);
  }
  .wish p {
    font-family: var(--display);
    font-size: clamp(1.6rem, 5vw, 3.75rem);
    font-stretch: 125%;
    font-weight: 900;
    letter-spacing: -0.015em;
    text-align: center;
  }
  @keyframes drop {
    from { opacity: 0; transform: translate(-50%, -60px); }
  }
  @keyframes tick {
    0% { opacity: 0; transform: scale(1.5); }
    30% { opacity: 1; transform: scale(1); }
    85% { opacity: 1; }
    100% { opacity: 0.6; transform: scale(0.96); }
  }
  @keyframes arrive {
    from { opacity: 0; transform: scale(0.5); }
  }
</style>
