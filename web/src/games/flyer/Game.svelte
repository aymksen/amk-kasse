<script>
  import { onMount } from 'svelte';
  import { fade, fly } from 'svelte/transition';
  import { backOut, cubicOut } from 'svelte/easing';
  import { createWorld } from './world.js';
  import { say } from './strings.js';
  import { calmNow } from '../../lib/moments.svelte.js';

  // Kistenflug: the bar's helicopter between stacks of beer crates. One button: every tap lifts it.
  let { player, best = 0, record = 0, onfinish, onquit } = $props();

  const KEYS = new Set(['Space', 'ArrowUp', 'KeyW']);
  const letters = $derived([...say('title')]);

  let root;
  let canvas;
  let world = null;
  let alive = false;

  // intro → play → crash → over → leaving → ready → play …   paused and resume break into play.
  let phase = $state('intro');
  let score = $state(0);
  let final = $state(0);
  let shown = $state(0);
  let counted = $state(false);
  let result = $state(null);
  let countdown = $state(3);
  let flash = $state(0);
  let runBest = $state(0);
  let runRecord = $state(0);

  let lockUntil = 0;
  let countT = 0;
  let resumeT = 0;
  let run = 0;
  let celebrated = false;

  const unit = (n) => (n === 1 ? say('unitOne') : say('unit'));

  // What the result screen says once the till has answered.
  const verdict = $derived.by(() => {
    if (!result) return null;
    if (result.failed) return { line: say('notSaved'), quiet: true };
    if (!result.saved) return { line: final === 0 ? say('zero') : say('notSavedPlain'), quiet: true };
    const newRecord = final >= result.record && final > runRecord;
    return {
      line: say('place', { place: result.place }),
      stamp: newRecord ? say('newRecord') : result.newBest ? say('newBest') : null,
      record: newRecord,
      standing: say('standing', { best: result.best, record: result.record }),
    };
  });

  // The small line under the score while flying: the next thing to beat.
  const chase = $derived.by(() => {
    if (runRecord > 0 && score > runRecord) return { text: say('beatRecord'), hot: true };
    if (runBest > 0 && score > runBest) {
      return runRecord > score ? { text: `${say('beatBest')} · ${say('toRecord', { n: runRecord })}`, warm: true } : { text: say('beatBest'), warm: true };
    }
    if (runBest > 0) return { text: say('toBest', { n: runBest }) };
    if (runRecord > 0) return { text: say('toRecord', { n: runRecord }) };
    return null;
  });

  const hooks = {
    score: (n) => {
      score = n;
    },
    crash: (n) => {
      final = n;
      phase = 'crash';
      if (!calmNow()) flash++;
      save(n);
    },
    over: () => {
      phase = 'over';
      countT = 0;
      shown = 0;
      counted = final === 0;
      lockUntil = performance.now() + 900;
    },
    ready: () => {
      phase = 'ready';
    },
  };

  function save(n) {
    const mine = ++run;
    result = null;
    let pending;
    try {
      pending = Promise.resolve(onfinish?.(n));
    } catch (err) {
      pending = Promise.reject(err);
    }
    pending.then(
      (r) => {
        if (alive && mine === run) result = r ?? { saved: false };
      },
      () => {
        if (alive && mine === run) result = { failed: true };
      },
    );
  }

  function press() {
    if (phase === 'intro' || phase === 'ready') {
      runBest = best;
      runRecord = record;
      score = 0;
      result = null;
      celebrated = false;
      world.begin();
      phase = 'play';
    } else if (phase === 'play') {
      world.flap();
    } else if (phase === 'paused') {
      phase = 'resume';
      resumeT = 0;
      countdown = 3;
    }
  }

  function again() {
    if (phase !== 'over' || performance.now() < lockUntil) return;
    run++;
    result = null;
    score = 0;
    shown = 0;
    counted = false;
    celebrated = false;
    phase = 'leaving';
    world.again();
  }

  function quit() {
    if (phase === 'over' && performance.now() < lockUntil) return;
    onquit?.();
  }

  function onpointerdown(e) {
    if (e.pointerType === 'mouse' && e.button !== 0) return;
    if (e.target.closest('button')) return;
    e.preventDefault();
    press();
  }

  function onkey(e) {
    if (!KEYS.has(e.code)) return;
    e.preventDefault();
    if (e.type !== 'keydown' || e.repeat) return;
    const el = document.activeElement;
    if (e.code === 'Space' && el && el.tagName === 'BUTTON' && root.contains(el)) {
      el.click();
      return;
    }
    if (phase === 'over') again();
    else press();
  }

  // The loop: real time in, clamped, so a hitch or a hidden tab never jumps ahead.
  let raf = 0;
  let last = 0;

  function tick(now) {
    raf = requestAnimationFrame(tick);
    const dt = last ? Math.min((now - last) / 1000, 0.1) : 0;
    last = now;
    if (phase === 'over' && !counted) {
      countT += dt;
      const t = Math.min(1, countT / Math.min(1.2, 0.35 + final * 0.03));
      shown = Math.round(final * (1 - (1 - t) ** 3));
      if (t >= 1) counted = true;
    }
    if (phase === 'over' && counted && verdict && !celebrated) {
      celebrated = true;
      if (verdict.stamp) world.celebrate(verdict.record ? 46 : 16);
    }
    if (phase === 'resume') {
      resumeT += dt;
      const n = 3 - Math.floor(resumeT / 0.5);
      if (n !== countdown) countdown = n;
      if (resumeT >= 1.5) phase = 'play';
    }
    world.advance(dt, phase === 'paused' || phase === 'resume');
  }

  function startLoop() {
    if (raf || document.hidden) return;
    last = 0;
    raf = requestAnimationFrame(tick);
  }

  function stopLoop() {
    cancelAnimationFrame(raf);
    raf = 0;
  }

  function onvisibility() {
    if (document.hidden) {
      stopLoop();
      if (phase === 'play' || phase === 'resume') phase = 'paused';
    } else startLoop();
  }

  function fit() {
    const w = root.clientWidth;
    const h = root.clientHeight;
    if (!w || !h) return;
    let dpr = Math.min(window.devicePixelRatio || 1, 2);
    if (w * h * dpr * dpr > 2.4e6) dpr = Math.sqrt(2.4e6 / (w * h));
    const room = world.resize(w, h, dpr);
    root.style.setProperty('--ceiling', `${room.ceiling.toFixed(1)}px`);
    root.style.setProperty('--floor', `${room.floor.toFixed(1)}px`);
  }

  onMount(() => {
    alive = true;
    world = createWorld(canvas, hooks);
    fit();
    const watch = new ResizeObserver(fit);
    watch.observe(root);
    window.addEventListener('keydown', onkey);
    window.addEventListener('keyup', onkey);
    document.addEventListener('visibilitychange', onvisibility);
    // The neon sign and the crate labels use the bar's display font; paint again once it is there.
    const fonts = ['800 40px "Barlow Condensed"', '900 40px "Barlow Condensed"'];
    if (document.fonts && !fonts.every((f) => document.fonts.check(f))) {
      Promise.all(fonts.map((f) => document.fonts.load(f))).then(
        () => alive && world.repaint(),
        () => {},
      );
    }
    startLoop();
    return () => {
      alive = false;
      stopLoop();
      watch.disconnect();
      window.removeEventListener('keydown', onkey);
      window.removeEventListener('keyup', onkey);
      document.removeEventListener('visibilitychange', onvisibility);
    };
  });
</script>

<div class="flyer" bind:this={root} {onpointerdown} role="application" aria-label={say('title')}>
  <canvas bind:this={canvas}></canvas>
  <div class="vignette"></div>
  <div class="curtain"></div>
  {#key flash}
    {#if flash}<div class="flash"></div>{/if}
  {/key}

  {#if phase === 'play' || phase === 'crash' || phase === 'paused' || phase === 'resume'}
    <div class="hud" in:fade={{ duration: 180 }} out:fade={{ duration: 260 }}>
      {#key score}
        <div class="score" class:bump={score > 0} class:ten={score > 0 && score % 10 === 0}>{score}</div>
      {/key}
      {#if chase}
        {#key chase.text}
          <div class="chase" class:warm={chase.warm} class:hot={chase.hot}>{chase.text}</div>
        {/key}
      {/if}
    </div>
  {/if}

  {#if phase === 'intro'}
    <div class="intro" out:fade={{ duration: 240 }}>
      <header class="head">
        <h1 class="title" aria-label={say('title')}>
          {#each letters as ch, i (i)}<span class="letter" style:animation-delay="{140 + i * 42}ms">{ch}</span>{/each}
        </h1>
        <p class="tagline">{say('tagline')}</p>
      </header>
      <footer class="foot">
        <div class="stats">
          <div class="stat">
            <span class="label">{say('best')}</span>
            <span class="value">{best > 0 ? best : say('none')}</span>
          </div>
          <div class="stat">
            <span class="label">{say('record')}</span>
            <span class="value">{record > 0 ? record : say('none')}</span>
          </div>
        </div>
        <div class="prompt">
          <span class="tap" aria-hidden="true"><span class="ring"></span><svg viewBox="0 0 24 24"><path d="M12 5l7 8h-4.5v6h-5v-6H5z" /></svg></span>
          <span class="words"><strong>{say('start')}</strong> {say('startSub')}</span>
        </div>
        <p class="how">{say('how')}</p>
      </footer>
    </div>
  {:else if phase === 'ready'}
    <div class="intro quick" in:fade={{ duration: 260, delay: 120 }} out:fade={{ duration: 200 }}>
      <footer class="foot">
        <p class="ready">{say('ready')}</p>
        <div class="prompt">
          <span class="tap" aria-hidden="true"><span class="ring"></span><svg viewBox="0 0 24 24"><path d="M12 5l7 8h-4.5v6h-5v-6H5z" /></svg></span>
          <span class="words"><strong>{say('start')}</strong> {say('startSub')}</span>
        </div>
      </footer>
    </div>
  {/if}

  {#if phase === 'paused' || phase === 'resume'}
    <div class="pause" transition:fade={{ duration: 180 }}>
      {#if phase === 'paused'}
        <p class="pause-title">{say('paused')}</p>
        <p class="pause-sub">{say('resume')}</p>
      {:else}
        {#key countdown}<p class="count">{countdown}</p>{/key}
      {/if}
    </div>
  {/if}

  {#if phase === 'over'}
    <div class="scrim" transition:fade={{ duration: 300 }}></div>
    <div class="result-wrap">
      <section class="result" in:fly={{ y: 70, duration: 520, easing: backOut }} out:fly={{ y: 50, duration: 240, easing: cubicOut }}>
        {#if counted && verdict?.stamp}
          <div class="stamp" class:record={verdict.record}>{verdict.stamp}</div>
        {/if}
        <p class="crashed">{say('crashed')}</p>
        <p class="final" class:done={counted}>{shown}</p>
        <p class="unit">{unit(final)}</p>
        <div class="verdict">
          {#if !verdict}
            <p class="saving">{say('saving')}</p>
          {:else}
            {#key verdict.line}
              <p class="line" class:quiet={verdict.quiet}>{verdict.line}</p>
            {/key}
            {#if verdict.standing}<p class="standing">{verdict.standing}</p>{/if}
          {/if}
        </div>
        <div class="actions">
          <button class="key primary" type="button" onclick={again}>{say('again')}</button>
          <button class="key" type="button" onclick={quit}>{say('back')}</button>
        </div>
        <p class="hint">{say('againHint')}</p>
      </section>
    </div>
  {/if}
</div>

<style>
  .flyer {
    --amber: #ffb224;
    --amber-lip: #a96d05;
    --cream: #fff3dc;
    --ink: #12100e;
    --soft: #c9b9a2;
    --display-face: 'Barlow Condensed', 'Barlow Semi Condensed', 'Segoe UI', sans-serif;
    --text-face: 'Barlow Semi Condensed', 'Segoe UI', system-ui, sans-serif;
    position: relative;
    width: 100%;
    height: 100%;
    overflow: hidden;
    container-type: size;
    background: var(--ink);
    color: var(--cream);
    font-family: var(--text-face);
    cursor: pointer;
    touch-action: none;
    user-select: none;
    -webkit-user-select: none;
    -webkit-tap-highlight-color: transparent;
  }

  canvas {
    position: absolute;
    inset: 0;
    display: block;
    width: 100%;
    height: 100%;
  }

  .vignette {
    position: absolute;
    inset: 0;
    pointer-events: none;
    background: radial-gradient(ellipse 75% 70% at 50% 48%, transparent 58%, rgb(6 4 2 / 0.55) 100%);
  }

  /* The game fades up out of the dark when it opens. */
  .curtain {
    position: absolute;
    inset: 0;
    pointer-events: none;
    background: #0a0806;
    opacity: 0;
    animation: lift 700ms ease-out;
  }
  @keyframes lift {
    from { opacity: 1; }
  }

  .flash {
    position: absolute;
    inset: 0;
    pointer-events: none;
    background: #fff1d0;
    opacity: 0;
    animation: flash 280ms ease-out;
  }
  @keyframes flash {
    from { opacity: 0.5; }
  }

  /* While flying */
  .hud {
    position: absolute;
    top: max(12px, 2.4cqh);
    left: 0;
    right: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;
    pointer-events: none;
  }
  .score {
    color: var(--amber);
    font-family: var(--display-face);
    font-size: clamp(56px, min(15cqw, 9.5cqh), 150px);
    font-weight: 900;
    font-variant-numeric: tabular-nums;
    line-height: 1;
    text-shadow: 0 0.05em 0 var(--amber-lip), 0 0.1em 0.3em rgb(0 0 0 / 0.55);
  }
  .score.bump { animation: bump 300ms cubic-bezier(0.2, 1.7, 0.4, 1); }
  .score.ten { animation: bump-big 520ms cubic-bezier(0.2, 1.7, 0.4, 1); }
  @keyframes bump {
    from { transform: scale(1.3); }
  }
  @keyframes bump-big {
    0% { transform: scale(1.7) rotate(-6deg); }
    60% { transform: scale(0.94) rotate(2deg); }
  }
  .chase {
    padding: 4px 12px;
    border-radius: 999px;
    background: rgb(12 9 7 / 0.55);
    color: var(--soft);
    font-size: clamp(14px, min(3.4cqw, 2.2cqh), 22px);
    font-weight: 600;
    letter-spacing: 0.02em;
    animation: rise 320ms ease-out;
  }
  .chase.warm { color: #ffd890; }
  .chase.hot { background: var(--amber); color: #241703; font-weight: 800; }

  /* Start */
  .intro {
    position: absolute;
    inset: 0;
    pointer-events: none;
  }
  .head {
    position: absolute;
    top: clamp(56px, 9cqh, 200px);
    left: 16px;
    right: 16px;
    text-align: center;
  }
  .title {
    margin: 0;
    color: var(--amber);
    font-family: var(--display-face);
    font-size: clamp(54px, min(16cqw, 11cqh), 180px);
    font-weight: 900;
    line-height: 0.92;
    letter-spacing: 0.01em;
    text-shadow: 0 0.045em 0 var(--amber-lip), 0 0.09em 0 rgb(0 0 0 / 0.35), 0 0.16em 0.4em rgb(0 0 0 / 0.5);
  }
  .letter {
    display: inline-block;
    white-space: pre;
    animation: drop 640ms cubic-bezier(0.2, 0.9, 0.25, 1.25) backwards;
  }
  @keyframes drop {
    0% { opacity: 0; transform: translateY(-0.6em) rotate(-10deg) scale(0.7); }
    55% { opacity: 1; }
  }
  .tagline {
    margin: 0.5em 0 0;
    color: var(--soft);
    font-size: clamp(16px, min(4cqw, 2.5cqh), 28px);
    font-weight: 500;
    animation: rise 500ms ease-out 620ms backwards;
  }
  .foot {
    position: absolute;
    bottom: calc(var(--floor, 60px) + clamp(14px, 3cqh, 64px));
    left: 16px;
    right: 16px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: clamp(10px, 1.8cqh, 22px);
    text-align: center;
  }
  .stats {
    display: flex;
    gap: 12px;
    animation: rise 480ms ease-out 760ms backwards;
  }
  .stat {
    display: flex;
    flex-direction: column;
    align-items: center;
    min-width: clamp(110px, 20cqw, 180px);
    padding: 8px 16px 10px;
    border-radius: 12px;
    background: rgb(18 13 9 / 0.72);
    box-shadow: inset 0 1px 0 rgb(255 255 255 / 0.06), 0 0 0 1px rgb(255 178 36 / 0.16);
  }
  .label {
    color: var(--soft);
    font-size: clamp(12px, min(2.8cqw, 1.7cqh), 17px);
    font-weight: 600;
    letter-spacing: 0.06em;
    text-transform: uppercase;
  }
  .value {
    color: var(--cream);
    font-family: var(--display-face);
    font-size: clamp(30px, min(7cqw, 4.4cqh), 56px);
    font-weight: 800;
    font-variant-numeric: tabular-nums;
    line-height: 1.05;
  }
  .prompt {
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 10px 24px 10px 12px;
    border-radius: 999px;
    background: var(--amber);
    color: #241703;
    box-shadow: 0 4px 0 var(--amber-lip), 0 10px 30px rgb(0 0 0 / 0.45);
    font-size: clamp(17px, min(4cqw, 2.6cqh), 30px);
    animation: rise 480ms ease-out 900ms backwards;
  }
  .quick .prompt { animation-delay: 160ms; }
  .prompt strong { font-weight: 800; }
  .words { white-space: nowrap; }
  .tap {
    position: relative;
    display: grid;
    place-items: center;
    width: 1.9em;
    height: 1.9em;
    border-radius: 50%;
    background: #241703;
  }
  .tap svg { width: 1.2em; height: 1.2em; fill: var(--amber); }
  .ring {
    position: absolute;
    inset: 0;
    border-radius: 50%;
    border: 3px solid #241703;
    opacity: 0;
    animation: ring 1.6s ease-out infinite;
  }
  @keyframes ring {
    0% { opacity: 0.7; transform: scale(1); }
    100% { opacity: 0; transform: scale(1.9); }
  }
  .how {
    max-width: 34em;
    margin: 0;
    color: var(--soft);
    font-size: clamp(14px, min(3.2cqw, 2cqh), 22px);
    animation: rise 480ms ease-out 1000ms backwards;
  }
  .ready {
    margin: 0;
    color: var(--amber);
    font-family: var(--display-face);
    font-size: clamp(36px, min(10cqw, 6.5cqh), 96px);
    font-weight: 900;
    line-height: 1;
    text-shadow: 0 0.05em 0 var(--amber-lip), 0 0.12em 0.3em rgb(0 0 0 / 0.5);
  }

  /* Pause */
  .pause {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 8px;
    background: rgb(8 6 4 / 0.55);
    pointer-events: none;
  }
  .pause-title {
    margin: 0;
    color: var(--amber);
    font-family: var(--display-face);
    font-size: clamp(48px, min(14cqw, 9cqh), 130px);
    font-weight: 900;
    line-height: 1;
  }
  .pause-sub {
    margin: 0;
    color: var(--cream);
    font-size: clamp(18px, min(4.2cqw, 2.6cqh), 30px);
  }
  .count {
    margin: 0;
    color: var(--cream);
    font-family: var(--display-face);
    font-size: clamp(90px, min(30cqw, 18cqh), 260px);
    font-weight: 900;
    line-height: 1;
    animation: count 500ms ease-out;
  }
  @keyframes count {
    from { opacity: 0; transform: scale(1.6); }
  }

  /* Result */
  .scrim {
    position: absolute;
    inset: 0;
    background: radial-gradient(ellipse at 50% 50%, rgb(10 7 5 / 0.45), rgb(6 4 3 / 0.75));
  }
  .result-wrap {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 16px;
    pointer-events: none;
  }
  .result {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    width: min(100%, 560px);
    padding: clamp(18px, 3.4cqh, 40px) clamp(18px, 4cqw, 40px) clamp(14px, 2.4cqh, 28px);
    border-radius: 20px;
    background: linear-gradient(180deg, #241c15, #17120e);
    box-shadow: inset 0 1px 0 rgb(255 255 255 / 0.07), 0 0 0 1px rgb(255 178 36 / 0.22), 0 24px 60px rgb(0 0 0 / 0.55);
    text-align: center;
    pointer-events: auto;
    cursor: default;
  }
  .result::before {
    content: '';
    position: absolute;
    top: 0;
    left: 24px;
    right: 24px;
    height: 4px;
    border-radius: 0 0 4px 4px;
    background: var(--amber);
  }
  .crashed {
    margin: 0;
    color: var(--soft);
    font-size: clamp(15px, min(3.4cqw, 2.2cqh), 22px);
    font-weight: 700;
    letter-spacing: 0.14em;
    text-transform: uppercase;
  }
  .final {
    margin: 4px 0 0;
    color: var(--amber);
    font-family: var(--display-face);
    font-size: clamp(76px, min(26cqw, 15cqh), 200px);
    font-weight: 900;
    font-variant-numeric: tabular-nums;
    line-height: 0.95;
    text-shadow: 0 0.04em 0 var(--amber-lip), 0 0.1em 0.3em rgb(0 0 0 / 0.5);
  }
  .final.done { animation: settle 360ms cubic-bezier(0.2, 1.6, 0.4, 1); }
  @keyframes settle {
    from { transform: scale(1.12); }
  }
  .unit {
    margin: 0;
    color: var(--soft);
    font-size: clamp(16px, min(3.8cqw, 2.4cqh), 24px);
    font-weight: 600;
  }
  .verdict {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 4px;
    min-height: 3.6em;
    margin: clamp(10px, 2cqh, 22px) 0;
    font-size: clamp(16px, min(3.8cqw, 2.4cqh), 24px);
  }
  .verdict p { margin: 0; }
  .saving { color: var(--soft); animation: breathe 1.2s ease-in-out infinite; }
  @keyframes breathe {
    50% { opacity: 0.45; }
  }
  .line {
    color: var(--cream);
    font-weight: 700;
    animation: pop 420ms cubic-bezier(0.2, 1.5, 0.4, 1);
  }
  .line.quiet { color: var(--soft); font-weight: 500; }
  .standing {
    color: var(--soft);
    font-size: 0.8em;
    animation: rise 400ms ease-out 120ms backwards;
  }
  @keyframes pop {
    from { opacity: 0; transform: scale(0.8); }
  }
  .stamp {
    position: absolute;
    top: -16px;
    right: -10px;
    padding: 6px 12px;
    border: 3px solid #63d695;
    border-radius: 8px;
    background: rgb(12 30 18 / 0.92);
    color: #7ff0ae;
    font-family: var(--display-face);
    font-size: clamp(18px, min(4.6cqw, 2.9cqh), 30px);
    font-weight: 900;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    transform: rotate(8deg);
    animation: stamp 520ms cubic-bezier(0.2, 1.2, 0.3, 1);
  }
  .stamp.record {
    border-color: var(--amber);
    background: #2a1a04;
    color: var(--amber);
    box-shadow: 0 0 0 4px rgb(255 178 36 / 0.18), 0 0 26px rgb(255 178 36 / 0.35);
  }
  @keyframes stamp {
    0% { opacity: 0; transform: rotate(-6deg) scale(2.2); }
    60% { opacity: 1; transform: rotate(10deg) scale(0.92); }
  }
  .actions {
    display: flex;
    gap: 12px;
    width: 100%;
    animation: rise 380ms ease-out 380ms backwards;
  }
  .key {
    flex: 1;
    min-height: 60px;
    margin-bottom: 4px;
    border: 0;
    border-radius: 12px;
    background: #2e2a25;
    box-shadow: inset 0 1px 0 rgb(255 255 255 / 0.08), 0 0 0 1px #050403, 0 4px 0 1px #050403;
    color: var(--cream);
    font: 700 clamp(18px, min(4.2cqw, 2.6cqh), 24px) var(--text-face);
    cursor: pointer;
  }
  .key:hover { background: #39342e; }
  .key:active { transform: translateY(3px); box-shadow: inset 0 1px 0 rgb(255 255 255 / 0.08), 0 0 0 1px #050403, 0 1px 0 1px #050403; }
  .key.primary {
    flex: 1.4;
    background: var(--amber);
    box-shadow: inset 0 1px 0 rgb(255 255 255 / 0.35), 0 0 0 1px var(--amber-lip), 0 4px 0 1px var(--amber-lip);
    color: #241703;
    font-weight: 800;
  }
  .key.primary:hover { background: #ffc14d; }
  .key.primary:active { box-shadow: inset 0 1px 0 rgb(255 255 255 / 0.35), 0 0 0 1px var(--amber-lip), 0 1px 0 1px var(--amber-lip); }
  .key:focus-visible { outline: 3px solid var(--cream); outline-offset: 3px; }
  .hint {
    margin: 10px 0 0;
    color: rgb(201 185 162 / 0.6);
    font-size: clamp(12px, min(2.8cqw, 1.7cqh), 16px);
  }

  @keyframes rise {
    from { opacity: 0; transform: translateY(14px); }
  }

  @media (prefers-reduced-motion: reduce) {
    .ring { display: none; }
  }
</style>
