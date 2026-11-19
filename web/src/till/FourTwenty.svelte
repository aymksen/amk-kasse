<script>
  import { onMount } from 'svelte';
  import MonkeyHand from './MonkeyHand.svelte';
  import { LEAF_PATH, LEAF_TRANSFORM, LEAF_VEINS, LEAF_VIEWBOX } from './cannabisLeaf.js';
  import { ARM_SHEEN, BELLY, CHEST_TUFT, EAR_FOLD, FACE, FOREARM, FOREARM_SHEEN, MIRROR, NOSE, TORSO, UPPER_ARM, crop, flip } from './monkey.js';

  // The till's joke for a total of 4,20 €, 42,00 € and 420,00 €, after the old 420 meme videos. The monkey grooves
  // in his beanie with a joint, pixel sunglasses slide down onto his face, the screen jumps and 420 slams in. On the
  // next big beat CHILL MAN follows, and he blows smoke rings while leaves rain down. Eight beats, 3,5 seconds in all.
  // The leaf is a free drawing from Wikimedia Commons (cannabisLeaf.js).
  // It only looks: every tap goes through it, and Sell.svelte ends it as soon as the total changes or someone pays.
  // Only transforms and opacity move, every moving part is its own small layer, and the two big pictures (the rays
  // and the haze) are small canvases painted once and stretched.
  // Every time in here is written for a run of 4,2 seconds and multiplied by PACE, in the script and through --pace in
  // the styles, so the whole thing lasts 3,5 seconds and keeps its rhythm.
  const PACE = 5 / 6;
  const LENGTH = 4200 * PACE;

  let { onend, line } = $props();

  // A number between 0 and 1 that is always the same for the same i and k, so every run looks the same.
  function hash(i, k) {
    const s = Math.sin(i * 12.9898 + k * 78.233) * 43758.5453;
    return s - Math.floor(s);
  }

  const TONES = ['#3fae5f', '#2a8a4a', '#6fcf6c', '#1f7a43', '#58c06a'];

  // Three depths of falling leaves: far ones small, dim and slow behind him, near ones big and fast in front of all.
  function rain(count, salt, [small, big], [quick, slow], latest) {
    return Array.from({ length: count }, (_, i) => {
      const r = (k) => hash(i + salt, k);
      return {
        left: ((i + r(1) * 0.9) / count) * 100,
        size: small + r(2) * (big - small),
        time: (quick + r(3) * (slow - quick)) * PACE,
        delay: r(4) * latest * PACE,
        spin: Math.round(r(5) * 360),
        sway: Math.round((r(6) - 0.5) * 18),
        flip: Math.round(50 + r(7) * 110),
        tone: TONES[Math.floor(r(8) * TONES.length)],
      };
    });
  }
  const far = rain(12, 0, [3, 5], [3000, 3600], 2400);
  const mid = rain(10, 40, [5.5, 8.5], [2200, 2700], 2600);
  const near = rain(6, 80, [11, 16], [1500, 1900], 2800);

  // When the glasses land, leaves fly out from behind his head.
  const burst = Array.from({ length: 12 }, (_, i) => {
    const angle = (i / 12) * Math.PI * 2 + 0.3;
    const reach = 34 + ((i * 7) % 18);
    return {
      x: Math.round(Math.cos(angle) * reach),
      y: Math.round(Math.sin(angle) * reach),
      spin: (i % 2 ? 1 : -1) * (160 + i * 20),
      size: 7 + ((i * 13) % 6),
      tone: TONES[i % TONES.length],
    };
  });

  // Smoke: a thin trail off the tip of the joint all along, a big cloud after the first drag, rings after that, and a
  // smaller cloud after the second drag. Numbers are in units of the 800 × 800 drawing.
  const wisps = [18, -10, 26, 4, -16, 22, 8, -6].map((drift, i) => ({ delay: (250 + i * 420) * PACE, drift }));
  const clouds = [
    { delay: 1650, size: 120, x: -150, y: -170 },
    { delay: 1710, size: 160, x: -60, y: -240 },
    { delay: 1770, size: 190, x: -230, y: -120 },
    { delay: 1830, size: 150, x: 20, y: -300 },
    { delay: 1890, size: 210, x: -120, y: -330 },
    { delay: 1960, size: 170, x: -270, y: -260 },
    { delay: 3380, size: 110, x: -120, y: -200 },
    { delay: 3440, size: 140, x: -30, y: -260 },
    { delay: 3500, size: 120, x: -200, y: -150 },
  ];
  const rings = [2350, 2700, 3050];

  // Pixel sunglasses, 26 × 6 pixels of 10 units: B black, W shine, a dot nothing. A bar across the top, then two
  // lenses of ten pixels with the bridge between them.
  const PIXEL = 10;
  const LENS = ['BWWBBBBBBB', 'BBWWBBBBBB', 'BBBWWBBBBB', '.BBBBBBBB.', '..BBBBBB..'];
  const SHADES = ['B'.repeat(26), ...LENS.map((row, i) => `..${row}${i === 0 ? 'BB' : '..'}${row}..`)];
  const shades = SHADES.flatMap((row, y) => {
    const runs = [];
    let start = 0;
    for (let x = 1; x <= row.length; x++) {
      if (x < row.length && row[x] === row[start]) continue;
      if (row[start] !== '.') runs.push({ x: start * PIXEL, y: y * PIXEL, w: (x - start) * PIXEL, shine: row[start] === 'W' });
      start = x;
    }
    return runs;
  });

  // Painted once, small, and stretched far behind him, so turning them costs nothing.
  function paintRays(canvas) {
    const ctx = canvas.getContext('2d');
    const c = canvas.width / 2;
    const RAYS = 16;
    ctx.fillStyle = 'rgb(150 230 140 / 0.22)';
    for (let i = 0; i < RAYS; i++) {
      const a = (i / RAYS) * Math.PI * 2;
      ctx.beginPath();
      ctx.moveTo(c, c);
      ctx.arc(c, c, c, a, a + Math.PI / RAYS);
      ctx.closePath();
      ctx.fill();
    }
    // The rays fade out towards the edge, so the canvas has no corners to show.
    ctx.globalCompositeOperation = 'destination-in';
    const fade = ctx.createRadialGradient(c, c, 0, c, c, c);
    fade.addColorStop(0, 'rgb(0 0 0 / 0.2)');
    fade.addColorStop(0.3, 'rgb(0 0 0 / 1)');
    fade.addColorStop(1, 'rgb(0 0 0 / 0)');
    ctx.fillStyle = fade;
    ctx.fillRect(0, 0, canvas.width, canvas.height);
  }

  // The smoke that gathers at the bottom, painted once like the rays.
  function paintHaze(canvas) {
    const ctx = canvas.getContext('2d');
    for (let i = 0; i < 12; i++) {
      const x = 16 + i * 44;
      const y = 92 + ((i * 29) % 42);
      const r = 46 + ((i * 17) % 34);
      const cloud = ctx.createRadialGradient(x, y, 0, x, y, r);
      cloud.addColorStop(0, 'rgb(186 224 190 / 0.42)');
      cloud.addColorStop(1, 'rgb(186 224 190 / 0)');
      ctx.fillStyle = cloud;
      ctx.fillRect(x - r, y - r, r * 2, r * 2);
    }
  }

  onMount(() => {
    const timer = setTimeout(onend, LENGTH);
    return () => clearTimeout(timer);
  });
</script>

{#snippet leaves(list, depth)}
  {#each list as leaf}
    <svg
      class="leaf {depth}"
      style:left="{leaf.left}%"
      style:width="{leaf.size}vmin"
      style:height="{leaf.size}vmin"
      style:color={leaf.tone}
      style:animation-duration="{leaf.time}ms"
      style:animation-delay="{leaf.delay}ms"
      style:--spin="{leaf.spin}deg"
      style:--sway="{leaf.sway}vw"
      style:--flip="{leaf.flip}deg"><use href="#amk-leaf" /></svg>
  {/each}
{/snippet}

<div class="four-twenty" aria-hidden="true" style:--pace={PACE}>
  <svg class="sprites">
    <symbol id="amk-leaf" viewBox={LEAF_VIEWBOX}>
      <path fill="currentColor" transform={LEAF_TRANSFORM} d={LEAF_PATH} />
      <path d={LEAF_VEINS} fill="none" stroke="#ffffff" stroke-opacity="0.28" stroke-width="12" stroke-linecap="round" />
    </symbol>
  </svg>

  <div class="dim"></div>

  <div class="stage">
    <canvas class="rays" width="512" height="512" use:paintRays></canvas>
    <div class="glow"></div>
    <div class="big-leaf"><svg class="beat"><use href="#amk-leaf" /></svg></div>

    <div class="rain">{@render leaves(far, 'far')}</div>

    {#each burst as leaf}
      <svg class="burst" style:--s="{leaf.size}vmin" style:--x="{leaf.x}vmin" style:--y="{leaf.y}vmin" style:--spin="{leaf.spin}deg" style:color={leaf.tone}><use href="#amk-leaf" /></svg>
    {/each}

    <!-- These boxes have no size on purpose, as in SixSeven.svelte: only the pictures become layers. -->
    <div class="monkey">
      <div class="part body">
        <svg {...crop(190, 390, 420, 446)}><path class="fur" d={TORSO} /></svg>
        <svg {...crop(290, 530, 220, 310)}>
          <path class="skin" d={BELLY} />
          <path class="fur" d={CHEST_TUFT} />
          <path class="crease" d="M390 742q10 10 20 0" />
        </svg>

        <div class="part upper left">
          <svg {...crop(106, 420, 216, 182)}><path class="fur" d={UPPER_ARM} /><path class="sheen" d={ARM_SHEEN} /></svg>
          <div class="part fore left">
            <svg {...crop(50, 404, 140, 200)}><path class="fur" d={FOREARM} /><path class="sheen" d={FOREARM_SHEEN} /></svg>
            <div class="part palm left"><MonkeyHand /></div>
          </div>
        </div>

        <div class="part upper right">
          <svg {...flip(106, 420, 216, 182)}><g transform={MIRROR}><path class="fur" d={UPPER_ARM} /><path class="sheen" d={ARM_SHEEN} /></g></svg>
          <div class="part fore right">
            <svg {...flip(50, 404, 140, 200)}><g transform={MIRROR}><path class="fur" d={FOREARM} /><path class="sheen" d={FOREARM_SHEEN} /></g></svg>
            <div class="part palm right"><MonkeyHand side="right" /></div>
          </div>
        </div>

        <div class="part head">
          <svg {...crop(214, 162, 372, 268)}>
            <circle class="fur" cx="268" cy="300" r="48" />
            <circle class="fur" cx="532" cy="300" r="48" />
            <circle class="skin" cx="266" cy="302" r="28" />
            <circle class="skin" cx="534" cy="302" r="28" />
            <path class="ear-fold" d={EAR_FOLD} />
            <path class="ear-fold" transform={MIRROR} d={EAR_FOLD} />
            <ellipse class="fur" cx="400" cy="296" rx="132" ry="128" />
            <path class="skin" d={FACE} />
            <ellipse class="eye" cx="354" cy="268" rx="27" ry="28" />
            <ellipse class="eye" cx="446" cy="268" rx="27" ry="28" />
            <path class="vein" d="M330 274l10 -2M332 282l9 2M470 274l-10 -2M468 282l-9 2" />
            <circle class="pupil" cx="356" cy="282" r="11" />
            <circle class="pupil" cx="448" cy="282" r="11" />
            <path class="heavy-lid" d="M325 276A29 31 0 0 1 383 276Z" />
            <path class="heavy-lid" d="M417 276A29 31 0 0 1 475 276Z" />
            <path class="nose" d={NOSE} />
            <ellipse class="nostril" cx="388" cy="324" rx="5" ry="4" />
            <ellipse class="nostril" cx="412" cy="324" rx="5" ry="4" />
            <path class="dimple" d="M320 350Q324 338 336 340" />
            <path class="smile" d="M330 350Q396 396 470 348" />
          </svg>
          <svg {...crop(440, 306, 136, 80)}>
            <path class="joint" d="M446 364L534 322L546 352L452 376Z" />
            <path class="ember" d="M534 322C544 318 556 330 558 340C560 348 552 352 546 352Z" />
          </svg>
          <i class="ember-glow"></i>
          <svg {...crop(270, 96, 260, 136)}>
            <path class="beanie top" d="M296 170C320 124 358 102 400 102C442 102 480 124 504 170C460 150 340 150 296 170Z" />
            <path class="beanie mid" d="M296 170C340 150 460 150 504 170C512 182 516 194 518 204C470 182 330 182 282 204C284 194 288 182 296 170Z" />
            <path class="beanie low" d="M282 204C330 182 470 182 518 204C522 214 522 222 520 228C470 206 330 206 280 228C278 222 278 214 282 204Z" />
          </svg>
          <svg class="shades" shape-rendering="crispEdges" {...crop(270, 238, 260, 60)}>
            <g transform="translate(270 238)">
              {#each shades as px}
                <rect x={px.x} y={px.y} width={px.w} height={PIXEL} fill={px.shine ? '#ffffff' : '#0a0a0a'} />
              {/each}
            </g>
          </svg>
        </div>

        {#each wisps as wisp}
          <i class="wisp" style:--drift={wisp.drift} style:animation-delay="{wisp.delay}ms"></i>
        {/each}
        {#each clouds as cloud}
          <i class="cloud" style:--wide={cloud.size} style:--x={cloud.x} style:--y={cloud.y} style:animation-delay="{cloud.delay * PACE}ms"></i>
        {/each}
      </div>

      {#each rings as delay}
        <i class="ring" style:animation-delay="{delay * PACE}ms"></i>
      {/each}
    </div>

    <div class="rain">{@render leaves(mid, 'mid')}</div>

    <canvas class="haze" width="512" height="160" use:paintHaze></canvas>

    <div class="titles">
      <div class="title four">
        <div class="pulse">
          <svg viewBox="0 0 600 250">
            <text class="shadow" x="312" y="240" textLength="470" lengthAdjust="spacingAndGlyphs">420</text>
            <text class="face" x="300" y="228" textLength="470" lengthAdjust="spacingAndGlyphs">420</text>
          </svg>
        </div>
      </div>
      <div class="side">
        <div class="title chill">
          <div class="pulse">
            <svg viewBox="0 0 640 140">
              <text class="shadow" x="327" y="129" textLength="590" lengthAdjust="spacingAndGlyphs">CHILL MAN</text>
              <text class="face" x="320" y="122" textLength="590" lengthAdjust="spacingAndGlyphs">CHILL MAN</text>
            </svg>
          </div>
        </div>
        <p class="line">{line}</p>
      </div>
    </div>

    <div class="rain">{@render leaves(near, 'near')}</div>
    <div class="flash"></div>
  </div>
</div>

<style>
  .four-twenty {
    --size: min(96vw, 56vh);
    --u: calc(var(--size) / 800); /* one unit of the 800 × 800 drawing */
    --head: calc(100vh - var(--u) * 504); /* the middle of his head */
    --fur: #15110e;
    --line: #15110e;
    --skin: #f6bb8c;
    --crease: #b5734a;
    --length: calc(4200ms * var(--pace));
    --beat: calc(525ms * var(--pace));
    --sine: cubic-bezier(0.45, 0, 0.55, 1);
    position: fixed;
    inset: 0;
    z-index: 35;
    overflow: hidden;
    pointer-events: none;
    contain: strict;
  }
  .sprites { position: absolute; width: 0; height: 0; }

  .dim { position: absolute; inset: 0; background: rgb(5 18 10 / 0.9); animation: show var(--length) ease-in-out both; }

  /* Everything that jumps when the glasses land and when CHILL MAN hits. */
  .stage { position: absolute; inset: 0; will-change: transform, opacity; animation: shake var(--length) linear both, leave var(--length) linear both; }

  .rays {
    position: absolute;
    top: var(--head);
    left: 50%;
    width: 170vmax;
    height: 170vmax;
    margin: -85vmax 0 0 -85vmax;
    will-change: transform, opacity;
    animation: show var(--length) ease-in-out both, turn var(--length) linear both;
  }
  .glow {
    position: absolute;
    top: var(--head);
    left: 50%;
    width: 120vmin;
    height: 120vmin;
    margin: -60vmin 0 0 -60vmin;
    border-radius: 50%;
    background: radial-gradient(closest-side, rgb(116 207 114 / 0.55), rgb(42 138 74 / 0.22) 58%, transparent);
    animation: show var(--length) ease-in-out both;
  }
  .big-leaf {
    position: absolute;
    top: var(--head);
    left: 50%;
    width: 84vmin;
    height: 84vmin;
    margin: -48vmin 0 0 -42vmin;
    color: #1d6b3f;
    will-change: transform, opacity;
    animation: show var(--length) ease-in-out both, leaf-sway var(--length) var(--sine) both;
  }
  .big-leaf .beat { display: block; width: 100%; height: 100%; will-change: transform; animation: beat var(--beat) ease-out calc(1575ms * var(--pace)) 5; }

  .rain { position: absolute; inset: 0; }
  .leaf {
    --fall: calc(100vh + 36vmin);
    position: absolute;
    top: -18vmin;
    will-change: transform;
    animation-name: fall;
    animation-timing-function: linear;
    animation-fill-mode: both;
  }
  .leaf.far { opacity: 0.5; }

  .burst {
    position: absolute;
    top: var(--head);
    left: 50%;
    width: var(--s);
    height: var(--s);
    margin: calc(var(--s) / -2);
    opacity: 0;
    will-change: transform, opacity;
    animation: burst calc(1300ms * var(--pace)) cubic-bezier(0.1, 0.75, 0.3, 1) calc(1575ms * var(--pace)) both;
  }

  .monkey {
    position: absolute;
    top: calc(100vh - var(--size));
    left: calc(50vw - var(--size) / 2);
    width: 0;
    height: 0;
    animation: monkey var(--length) linear both;
  }
  .part {
    position: absolute;
    top: 0;
    left: 0;
    width: 0;
    height: 0;
    transform-origin: calc(var(--u) * var(--ox, 0)) calc(var(--u) * var(--oy, 0));
    will-change: transform;
  }
  .monkey svg { position: absolute; display: block; overflow: visible; }

  /* The groove: down on every beat, leaning left and right in turn, the head a moment later and the hands later still,
     so it all swings through. It starts half a beat in, while he is still coming up. */
  .body, .head, .upper, .fore, .palm { animation: groove calc(var(--beat) * 2) var(--sine) var(--lag, calc(262ms * var(--pace))) 3.75 both; }
  .body { --ox: 400; --oy: 800; --a: translate(calc(var(--u) * -14), calc(var(--u) * 12)) rotate(-3deg); --b: translate(calc(var(--u) * 14), calc(var(--u) * 12)) rotate(3deg); }
  .head { --ox: 400; --oy: 430; --lag: calc(322ms * var(--pace)); --a: translateY(calc(var(--u) * 14)) rotate(5deg); --b: translateY(calc(var(--u) * 14)) rotate(-5deg); }
  .upper { --a: rotate(7deg); --b: rotate(-7deg); }
  .upper.left { --ox: 272; --oy: 472; }
  .upper.right { --ox: 528; --oy: 472; }
  .fore { --lag: calc(312ms * var(--pace)); --a: rotate(12deg); --b: rotate(-12deg); }
  .fore.left { --ox: 150; --oy: 566; }
  .fore.right { --ox: 650; --oy: 566; }
  .palm { --lag: calc(362ms * var(--pace)); --a: rotate(-10deg); --b: rotate(10deg); }
  .palm.left { --ox: 84; --oy: 438; }
  .palm.right { --ox: 716; --oy: 438; }

  /* The tip of the joint lights up when he pulls on it, twice. */
  .ember-glow {
    position: absolute;
    top: calc(var(--u) * 336);
    left: calc(var(--u) * 552);
    width: calc(var(--u) * 110);
    height: calc(var(--u) * 110);
    margin: calc(var(--u) * -55);
    border-radius: 50%;
    background: radial-gradient(closest-side, rgb(255 236 150 / 0.95), rgb(255 120 40 / 0.6) 40%, rgb(255 90 31 / 0));
    will-change: transform, opacity;
    animation: ember var(--length) linear both;
  }

  /* They slide down slowly and land on the third beat, the way the meme does it. */
  .shades { will-change: transform; animation: shades var(--length) linear both; }

  .wisp, .cloud {
    position: absolute;
    border-radius: 50%;
    background: radial-gradient(closest-side, rgb(232 244 232 / 0.95), rgb(226 240 226 / 0.6) 50%, rgb(226 240 226 / 0));
    opacity: 0;
    will-change: transform, opacity;
  }
  .wisp {
    top: calc(var(--u) * 318);
    left: calc(var(--u) * 556);
    width: calc(var(--u) * 44);
    height: calc(var(--u) * 44);
    margin: calc(var(--u) * -22);
    animation: wisp calc(1500ms * var(--pace)) ease-out both;
  }
  .cloud {
    top: calc(var(--u) * 388);
    left: calc(var(--u) * 392);
    width: calc(var(--u) * var(--wide));
    height: calc(var(--u) * var(--wide));
    margin: calc(var(--u) * var(--wide) / -2);
    animation: cloud calc(1900ms * var(--pace)) cubic-bezier(0.3, 0.6, 0.5, 1) both;
  }
  .ring {
    position: absolute;
    top: calc(var(--u) * 376);
    left: calc(var(--u) * 396);
    width: calc(var(--u) * 96);
    height: calc(var(--u) * 96);
    margin: calc(var(--u) * -48);
    border: calc(var(--u) * 12) solid rgb(230 242 230 / 0.85);
    border-radius: 50%;
    box-shadow: 0 0 calc(var(--u) * 10) calc(var(--u) * 2) rgb(230 242 230 / 0.5), inset 0 0 calc(var(--u) * 10) calc(var(--u) * 2) rgb(230 242 230 / 0.5);
    opacity: 0;
    will-change: transform, opacity;
    animation: ring calc(1500ms * var(--pace)) ease-out both;
  }

  .haze {
    position: absolute;
    right: -6vw;
    bottom: -3vh;
    left: -6vw;
    width: 112vw;
    height: 30vh;
    will-change: transform, opacity;
    animation: haze var(--length) ease-in-out both, roll var(--length) linear both;
  }

  /* Upright: 420 and CHILL MAN above him. Wide: one on each side. */
  .titles { position: absolute; top: 5vh; right: 0; left: 0; display: flex; flex-direction: column; align-items: center; gap: 1.5vh; }
  .side { display: flex; flex-direction: column; align-items: center; gap: 1.2vh; width: min(70vw, 42vh); }
  .title { will-change: transform, opacity; }
  .title.four { width: min(78vw, 46vh); animation: slam-four var(--length) linear both; }
  .title.chill { width: 100%; animation: slam-chill var(--length) linear both; }
  .pulse { will-change: transform; }
  .four .pulse { animation: pulse var(--beat) ease-out calc(2100ms * var(--pace)) 4; }
  .chill .pulse { animation: pulse var(--beat) ease-out calc(2625ms * var(--pace)) 3; }
  .title svg { display: block; width: 100%; height: auto; overflow: visible; }
  text {
    font-family: Impact, Haettenschweiler, 'Archivo Variable', 'Segoe UI', sans-serif;
    font-stretch: 100%;
    font-synthesis: none;
    font-weight: 900;
    paint-order: stroke;
    stroke: #0b1a10;
    stroke-linejoin: round;
    text-anchor: middle;
  }
  .four text { font-size: 260px; stroke-width: 22px; }
  .chill text { font-size: 130px; stroke-width: 14px; }
  text.shadow { fill: #0b1a10; }
  .four text.face { fill: #3fd06a; animation: rasta calc(var(--beat) * 3) steps(1, end) calc(1575ms * var(--pace)) infinite; }
  .chill text.face { fill: #ffffff; }

  .line {
    margin: 0;
    color: #eefbe6;
    font-family: 'IBM Plex Mono', Consolas, monospace;
    font-size: 1.0625rem;
    font-weight: 600;
    text-align: center;
    text-shadow: 0 2px 0 #0b1a10;
    animation: line var(--length) ease-out both;
  }

  .flash { position: absolute; inset: 0; background: #fffbe0; opacity: 0; will-change: opacity; animation: flash var(--length) linear both; }

  @media (orientation: landscape) {
    .four-twenty { --size: min(84vh, 58vw); }
    .titles { top: 0; bottom: 0; flex-direction: row; justify-content: space-between; align-items: center; gap: 0; padding: 0 3vw 16vh; }
    .title.four, .side { width: calc(50vw - var(--size) / 2 - 4vw); }
  }

  .fur { fill: var(--fur); }
  .skin { fill: var(--skin); }
  .sheen { fill: none; stroke: #2f5a45; stroke-linecap: round; stroke-width: 6; }
  .crease { fill: none; stroke: var(--crease); stroke-linecap: round; stroke-width: 5; }
  .ear-fold { fill: none; stroke: var(--crease); stroke-linecap: round; stroke-width: 5; }
  .eye { fill: #f6c3b8; stroke: var(--line); stroke-width: 6; }
  .vein { fill: none; stroke: #d2453a; stroke-linecap: round; stroke-width: 3; }
  .pupil { fill: var(--line); }
  .heavy-lid { fill: var(--skin); stroke: var(--line); stroke-linejoin: round; stroke-width: 7; }
  .nose { fill: #e9a271; stroke: var(--line); stroke-linejoin: round; stroke-width: 5; }
  .nostril { fill: #5a2c1a; }
  .dimple { fill: none; stroke: var(--line); stroke-linecap: round; stroke-width: 6; }
  .smile { fill: none; stroke: var(--line); stroke-linecap: round; stroke-width: 8; }
  .joint { fill: #f6efdf; stroke: var(--line); stroke-linejoin: round; stroke-width: 6; }
  .ember { fill: #ff5a1f; stroke: var(--line); stroke-linejoin: round; stroke-width: 6; }
  .beanie { stroke: var(--line); stroke-linejoin: round; stroke-width: 7; }
  .beanie.top { fill: #cf3029; }
  .beanie.mid { fill: #f7c62f; }
  .beanie.low { fill: #1f8a4c; }

  /* The whole run is 4200 ms before PACE, so 1 % is 42 ms: the glasses land at 37.5 % (1575 ms), CHILL MAN at 50 % (2100 ms). */
  @keyframes show {
    0% { opacity: 0; }
    8%, 90% { opacity: 1; }
    100% { opacity: 0; }
  }
  @keyframes leave {
    0%, 90% { opacity: 1; }
    100% { opacity: 0; }
  }
  @keyframes shake {
    0%, 37.4% { transform: none; }
    37.5% { transform: scale(1.07) translate(-1.2vmin, 0.8vmin) rotate(-0.7deg); }
    39% { transform: scale(1.045) translate(1vmin, -0.8vmin) rotate(0.5deg); }
    40.5% { transform: scale(1.02) translate(-0.6vmin, 0.4vmin); }
    42% { transform: scale(1.01) translate(0.3vmin, -0.2vmin); }
    44%, 49.9% { transform: none; }
    50% { transform: scale(1.035) translate(0.8vmin, 0.5vmin) rotate(0.4deg); }
    51.5% { transform: scale(1.015) translate(-0.5vmin, -0.3vmin); }
    53.5%, 100% { transform: none; }
  }
  @keyframes flash {
    0%, 37.4% { opacity: 0; }
    37.5% { opacity: 0.7; }
    43%, 49.9% { opacity: 0; }
    50% { opacity: 0.3; }
    53%, 100% { opacity: 0; }
  }
  @keyframes turn {
    from { transform: rotate(0deg); }
    to { transform: rotate(50deg); }
  }
  @keyframes leaf-sway {
    from { transform: scale(0.82) rotate(-7deg); }
    to { transform: scale(1) rotate(7deg); }
  }
  @keyframes beat {
    from { transform: scale(1.08); }
    to { transform: none; }
  }
  @keyframes fall {
    0% { transform: translate3d(0, 0, 0) rotate(var(--spin)) rotateY(0deg); }
    25% { transform: translate3d(var(--sway), calc(var(--fall) * 0.25), 0) rotate(calc(var(--spin) + 40deg)) rotateY(var(--flip)); }
    50% { transform: translate3d(0, calc(var(--fall) * 0.5), 0) rotate(var(--spin)) rotateY(0deg); }
    75% { transform: translate3d(calc(var(--sway) * -1), calc(var(--fall) * 0.75), 0) rotate(calc(var(--spin) - 40deg)) rotateY(calc(var(--flip) * -1)); }
    100% { transform: translate3d(0, var(--fall), 0) rotate(var(--spin)) rotateY(0deg); }
  }
  @keyframes burst {
    0% { opacity: 1; transform: scale(0.2) rotate(0deg); }
    70% { opacity: 1; }
    100% { opacity: 0; transform: translate(var(--x), var(--y)) scale(1) rotate(var(--spin)); }
  }
  @keyframes monkey {
    0% { transform: translateY(calc(var(--u) * 860)); animation-timing-function: cubic-bezier(0.2, 0.9, 0.3, 1.08); }
    11%, 88% { transform: none; animation-timing-function: cubic-bezier(0.6, 0, 0.9, 0.5); }
    100% { transform: translateY(calc(var(--u) * 900)); }
  }
  @keyframes groove {
    0%, 50%, 100% { transform: none; }
    25% { transform: var(--a); }
    75% { transform: var(--b); }
  }
  @keyframes ember {
    0% { opacity: 0.3; transform: scale(0.8); }
    16% { opacity: 0.35; transform: scale(0.85); }
    24% { opacity: 0.8; transform: scale(1.1); }
    27% { opacity: 0.65; transform: scale(1); }
    33% { opacity: 1; transform: scale(1.3); }
    38% { opacity: 0.4; transform: scale(0.9); }
    66% { opacity: 0.4; transform: scale(0.9); }
    72% { opacity: 0.95; transform: scale(1.2); }
    78%, 100% { opacity: 0.4; transform: scale(0.9); }
  }
  @keyframes shades {
    0%, 16% { transform: translateY(-110vh); }
    37.5% { transform: none; animation-timing-function: ease-out; }
    39.5% { transform: translateY(calc(var(--u) * -10)); animation-timing-function: ease-in; }
    41.5%, 100% { transform: none; }
  }
  @keyframes wisp {
    0% { opacity: 0; transform: scale(0.4); }
    20% { opacity: 0.75; }
    50% { transform: translate(calc(var(--u) * var(--drift) * -0.6), calc(var(--u) * -120)) scale(2); }
    100% { opacity: 0; transform: translate(calc(var(--u) * var(--drift)), calc(var(--u) * -260)) scale(3.2); }
  }
  /* The cloud stays thick for a while before it thins out. */
  @keyframes cloud {
    0% { opacity: 0; transform: scale(0.15); }
    10% { opacity: 1; }
    55% { opacity: 0.8; transform: translate(calc(var(--u) * var(--x) * 0.7), calc(var(--u) * var(--y) * 0.7)) scale(1.5); }
    100% { opacity: 0; transform: translate(calc(var(--u) * var(--x)), calc(var(--u) * var(--y))) scale(1.9); }
  }
  @keyframes ring {
    0% { opacity: 0; transform: scale(0.35, 0.2); }
    12% { opacity: 0.95; }
    100% { opacity: 0; transform: translate(calc(var(--u) * -30), calc(var(--u) * -420)) scale(2.4, 1.3); }
  }
  @keyframes haze {
    0% { opacity: 0; }
    30% { opacity: 0.6; }
    85% { opacity: 1; }
    100% { opacity: 0; }
  }
  @keyframes roll {
    from { transform: translateX(-3vw); }
    to { transform: translateX(3vw); }
  }
  /* In big and tilted, then down in a sixth of a second. */
  @keyframes slam-four {
    0%, 37.4% { opacity: 0; transform: scale(2.8) rotate(-12deg); }
    37.5% { opacity: 1; transform: scale(2.8) rotate(-12deg); animation-timing-function: cubic-bezier(0.2, 0.9, 0.3, 1.2); }
    41%, 88% { opacity: 1; transform: rotate(-5deg); animation-timing-function: ease-in; }
    97%, 100% { opacity: 0; transform: scale(1.6) rotate(-5deg); }
  }
  @keyframes slam-chill {
    0%, 49.9% { opacity: 0; transform: scale(2.8) rotate(10deg); }
    50% { opacity: 1; transform: scale(2.8) rotate(10deg); animation-timing-function: cubic-bezier(0.2, 0.9, 0.3, 1.2); }
    53.5%, 88% { opacity: 1; transform: rotate(4deg); animation-timing-function: ease-in; }
    97%, 100% { opacity: 0; transform: scale(1.6) rotate(4deg); }
  }
  @keyframes pulse {
    from { transform: scale(1.07); }
    to { transform: none; }
  }
  @keyframes rasta {
    0% { fill: #3fd06a; }
    33.33% { fill: #ffd23f; }
    66.67% { fill: #ef4135; }
  }
  @keyframes line {
    0%, 57% { opacity: 0; transform: translateY(10px); }
    62%, 100% { opacity: 1; transform: none; }
  }
</style>
