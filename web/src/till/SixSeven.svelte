<script>
  import { onMount } from 'svelte';
  import { moments } from '../lib/moments.svelte.js';
  import MonkeyHand from './MonkeyHand.svelte';
  import { ARM_SHEEN, BELLY, CHEST_TUFT, EAR_FOLD, FACE, FOREARM, FOREARM_SHEEN, MIRROR, NOSE, TORSO, UPPER_ARM, crop, flip } from './monkey.js';

  // The till's Easter egg for a total of exactly 6,70 €: the bar's monkey does the "six seven" hand gesture.
  // It only looks. Every tap goes through it to the till underneath, and Sell.svelte ends it as soon as the total changes.
  const LENGTH = 2000;

  let { onend, line } = $props();

  // Digits falling behind the monkey. Worked out once, so a second run looks the same as the first.
  const rain = Array.from({ length: 16 }, (_, i) => ({
    digit: i % 2 ? '7' : '6',
    left: (i * 37) % 100,
    delay: 80 + ((i * 193) % 900),
    size: 34 + ((i * 53) % 74),
    spin: ((i * 71) % 90) - 45,
    tone: ['#ffd84a', '#f4ead2', '#ff8fa3', '#7fe0a8'][i % 4],
  }));

  // Painted once, small, and stretched over the screen. As a CSS gradient the rays were the biggest layer of all.
  function paintRays(canvas) {
    const ctx = canvas.getContext('2d');
    ctx.translate(256, 256);
    ctx.fillStyle = 'rgb(255 216 74 / 0.3)';
    for (let i = 0; i < 20; i++) {
      ctx.rotate(Math.PI / 10);
      ctx.beginPath();
      ctx.moveTo(0, 0);
      ctx.arc(0, 0, 256, 0, Math.PI / 20);
      ctx.fill();
    }
  }

  onMount(() => {
    const timer = setTimeout(onend, LENGTH);
    return () => clearTimeout(timer);
  });
</script>

<div class="six-seven" aria-hidden="true">
  <div class="dim"></div>
  <canvas class="rays" width="512" height="512" use:paintRays></canvas>
  <div class="glow"></div>
  <div class="rain">
    {#each rain as drop}
      <i style:left="{drop.left}%" style:animation-delay="{drop.delay}ms" style:font-size="{drop.size}px" style:color={drop.tone} style:--spin="{drop.spin}deg">{drop.digit}</i>
    {/each}
  </div>

  <div class="band">
    <b class="six swing">SIX</b>
    <b class="seven swing">SEVEN</b>
  </div>
  <p class="line">{line}</p>

  <!-- Every moving part is its own small SVG on its own layer, so the browser only slides finished pictures around
       instead of painting the monkey again in every frame. -->
  <div class="monkey">
    <div class="part body">
      <svg {...crop(190, 390, 420, 446)}><path class="fur" d={TORSO} /></svg>
      <div class="part belly">
        <svg {...crop(290, 530, 220, 310)}>
          <path class="skin" d={BELLY} />
          <path class="fur" d={CHEST_TUFT} />
          <path class="crease" d="M390 742q10 10 20 0" />
        </svg>
      </div>

      <div class="part upper left swing">
        <svg {...crop(106, 420, 216, 182)}><path class="fur" d={UPPER_ARM} /><path class="sheen" d={ARM_SHEEN} /></svg>
        <div class="part fore left swing">
          <svg {...crop(50, 404, 140, 200)}><path class="fur" d={FOREARM} /><path class="sheen" d={FOREARM_SHEEN} /></svg>
          <div class="part palm left swing">
            <div class="part digit"><svg {...crop(-40, 94, 224, 236)}><text x="80" y="302">6</text></svg></div>
            <MonkeyHand />
          </div>
        </div>
      </div>

      <div class="part upper right swing">
        <svg {...flip(106, 420, 216, 182)}><g transform={MIRROR}><path class="fur" d={UPPER_ARM} /><path class="sheen" d={ARM_SHEEN} /></g></svg>
        <div class="part fore right swing">
          <svg {...flip(50, 404, 140, 200)}><g transform={MIRROR}><path class="fur" d={FOREARM} /><path class="sheen" d={FOREARM_SHEEN} /></g></svg>
          <div class="part palm right swing">
            <div class="part digit"><svg {...crop(616, 94, 224, 236)}><text x="720" y="302">7</text></svg></div>
            <MonkeyHand side="right" />
          </div>
        </div>
      </div>

      <div class="part head swing">
        <svg {...crop(214, 162, 372, 268)}>
          <circle class="fur" cx="268" cy="300" r="48" />
          <circle class="fur" cx="532" cy="300" r="48" />
          <circle class="skin" cx="266" cy="302" r="28" />
          <circle class="skin" cx="534" cy="302" r="28" />
          <path class="ear-fold" d={EAR_FOLD} />
          <path class="ear-fold" transform={MIRROR} d={EAR_FOLD} />
          <ellipse class="fur" cx="400" cy="296" rx="132" ry="128" />
          <path class="skin" d={FACE} />
          <path class="brow" d="M318 224Q352 202 386 226" />
          <ellipse class="eye" cx="354" cy="264" rx="27" ry="31" />
          <ellipse class="eye" cx="446" cy="264" rx="27" ry="31" />
          <path class="nose" d={NOSE} />
          <ellipse class="nostril" cx="388" cy="324" rx="5" ry="4" />
          <ellipse class="nostril" cx="412" cy="324" rx="5" ry="4" />
          <path class="dimple" d="M320 346Q324 334 336 336M480 346Q476 334 464 336" />
          <path class="mouth" d="M332 350Q400 370 468 350Q460 404 400 406Q340 404 332 350Z" />
          <path class="teeth" d="M342 356Q400 372 458 356L454 370Q400 384 346 370Z" />
          <path class="tongue" d="M366 392Q400 372 434 392Q420 406 400 406Q380 406 366 392Z" />
        </svg>
        <div class="part brow-up swing"><svg {...crop(404, 186, 92, 36)}><path class="brow" d="M414 212Q450 182 486 214" /></svg></div>
        <div class="part pupils swing">
          <svg {...crop(336, 248, 128, 36)}>
            <circle cx="354" cy="266" r="14" />
            <circle cx="446" cy="266" r="14" />
          </svg>
        </div>
        <div class="part lid">
          <svg {...crop(408, 222, 76, 84)}>
            <ellipse class="skin" cx="446" cy="264" rx="34" ry="38" />
            <path class="wink" d="M418 268Q446 248 474 268" />
          </svg>
        </div>
        <div class="part hat swing">
          {#if moments.season?.kind === 'christmas'}
            <!-- At Christmas the banana makes way for a red hat. -->
            <svg {...crop(276, 62, 316, 180)}>
              <path fill="#d42a2a" d="M300 212C306 158 340 92 416 78C470 70 524 100 544 172C520 148 492 146 478 158C488 176 496 194 500 212Z" />
              <path fill="#a51a22" d="M416 78C470 70 524 100 544 172C520 148 492 146 478 158C488 176 496 194 500 212H448C452 160 438 112 416 78Z" />
              <rect x="284" y="192" width="232" height="44" rx="22" fill="#fff" />
              <circle cx="550" cy="186" r="24" fill="#fff" />
            </svg>
          {:else if moments.season?.kind === 'halloween'}
            <!-- At Halloween a witch's hat. -->
            <svg {...crop(262, 46, 276, 186)}>
              <path class="witch" d="M322 196C336 146 362 104 404 70C424 54 450 52 468 66C450 70 440 80 442 96C452 130 464 164 474 196Z" />
              <path class="witch" d="M268 206C290 192 340 188 400 188C460 188 510 192 532 206C510 226 460 228 400 228C340 228 290 226 268 206Z" />
              <path class="witch-band" d="M320 176C366 190 430 190 478 176L482 198C430 212 366 212 316 198Z" />
            </svg>
          {:else}
            <svg {...crop(298, 2, 202, 204)}>
              <path class="banana" d="M306 196C360 176 404 110 392 44L420 38C470 92 492 150 490 200C430 168 366 172 306 196Z" />
              <path class="banana-side" d="M436 58C466 100 482 150 482 190C470 184 460 180 450 178C462 140 456 96 436 58Z" />
              <path class="stem" d="M392 44L386 16Q398 6 412 12L420 38Z" />
            </svg>
          {/if}
        </div>
        <div class="part sparkle"><svg {...crop(514, 170, 104, 104)}><path class="star" d="M566 176Q570 218 612 222Q570 226 566 268Q562 226 520 222Q562 218 566 176Z" /></svg></div>
      </div>
    </div>
  </div>
</div>

<style>
  /* Two seconds: he pops up, swings the 6 and the 7 twice, winks, and falls out of the picture while the digits
     stay in the air for a moment, the way things do in cartoons, before they fall too. Whatever moves fast
     waits 100 ms first, so the browser has its pictures ready before the movement starts.

     The swing: the 6 is up at 300 ms, the 7 at 620 ms and the 6 again at 940 ms, before he falls. */
  .six-seven {
    /* The arms reach about 85 units past the drawing on each side, so the width limit leaves room for them. */
    --size: min(92vh, 77vw);
    --u: calc(var(--size) / 800); /* one unit of the 800 × 800 drawing */
    --fur: #15110e;
    --line: #15110e;
    --skin: #f6bb8c;
    --crease: #b5734a;
    --head: calc(100vh - var(--u) * 504); /* the middle of his head, where the rays start */
    --length: 2000ms;
    --start: 140ms;
    --gesture: 1280ms;
    --beat: 160ms;
    --sine: cubic-bezier(0.37, 0, 0.63, 1);
    --spring: cubic-bezier(0.34, 1.4, 0.64, 1);
    --drop: cubic-bezier(0.55, 0, 0.85, 0.35);
    position: fixed;
    inset: 0;
    z-index: 35;
    overflow: hidden;
    pointer-events: none;
    contain: strict;
  }

  /* One curve for everything that sways. Each part names its own two end poses and how far it trails behind. */
  .swing { animation: swing var(--gesture) var(--sine) calc(var(--start) + var(--lag, 0ms)) both; }

  .dim { position: absolute; inset: 0; background: rgb(10 7 4 / 0.66); animation: show var(--length) ease-in-out both; }
  .rays {
    position: absolute;
    top: var(--head);
    left: 50%;
    width: 140vmax;
    height: 140vmax;
    margin: -70vmax 0 0 -70vmax;
    animation: show var(--length) ease-in-out both, turn var(--length) cubic-bezier(0.15, 0.6, 0.4, 1) both;
  }
  .glow {
    position: absolute;
    top: var(--head);
    left: 50%;
    width: 120vmin;
    height: 120vmin;
    margin: -60vmin 0 0 -60vmin;
    border-radius: 50%;
    background: radial-gradient(closest-side, rgb(255 222 110 / 0.95) 0 22%, rgb(255 216 74 / 0.4) 50%, transparent);
    animation: show var(--length) ease-in-out both, beat var(--beat) var(--sine) var(--start) 8 alternate;
  }

  .rain { position: absolute; inset: 0; }
  .rain i {
    position: absolute;
    top: -130px;
    font-family: 'Archivo Variable', 'Segoe UI', sans-serif;
    font-stretch: 112%;
    font-style: normal;
    font-weight: 900;
    line-height: 1;
    -webkit-text-stroke: 3px #120d08;
    paint-order: stroke;
    will-change: transform;
    animation: fall 900ms cubic-bezier(0.45, 0.05, 0.85, 0.75) both;
  }

  .band {
    position: absolute;
    right: -2vw;
    bottom: 15vh;
    left: -2vw;
    display: grid;
    grid-template-columns: 1fr 1fr;
    align-items: center;
    justify-items: center;
    column-gap: 68vh;
    height: 21vh;
    border-block: 9px solid #1f5a3c;
    background: #f6ecd6;
    box-shadow: 0 0 0 4px #f6ecd6, 0 14px 0 rgb(0 0 0 / 0.35);
    color: #6d1d3a;
    font-family: 'Archivo Variable', 'Segoe UI', sans-serif;
    font-size: calc((50vw - 34vh) / 4.8);
    font-stretch: 118%;
    font-weight: 900;
    letter-spacing: -0.01em;
    line-height: 1;
    animation: band var(--length) linear both;
  }
  .band b { display: block; font-weight: inherit; will-change: transform; }
  .six { --a: scale(1.12); --b: scale(0.88); }
  .seven { --a: scale(0.88); --b: scale(1.12); }

  .line {
    position: absolute;
    right: 16px;
    bottom: calc(15vh - 36px);
    left: calc(50vw + 24vh);
    color: #fff5dc;
    font-family: 'IBM Plex Mono', Consolas, monospace;
    font-size: 1.0625rem;
    font-weight: 600;
    text-align: center;
    text-shadow: 0 2px 0 #120d08;
    animation: line var(--length) ease-out both;
  }

  /* These boxes have no size on purpose: a layer also covers its own box, and only the pictures should count.
     That is why pivots and distances are in units of the drawing and not in percent. */
  .monkey {
    position: absolute;
    top: calc(100vh - var(--size));
    left: calc(50vw - var(--size) / 2);
    width: 0;
    height: 0;
    transform-origin: calc(var(--u) * 400) calc(var(--u) * 400);
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
  svg { position: absolute; display: block; overflow: visible; }

  /* Upright screen: he fills the width, so the banner and the caption move above him. */
  @media (orientation: portrait) {
    .band { bottom: calc(var(--size) + 3vh); column-gap: 0; height: 15vw; font-size: 9.4vw; }
    .line { right: 16px; bottom: calc(var(--size) + 3vh + 15vw + 28px); left: 16px; }
  }

  .fur { fill: var(--fur); }
  .skin { fill: var(--skin); }
  .sheen { fill: none; stroke: #2f5a45; stroke-linecap: round; stroke-width: 6; }
  .crease { fill: none; stroke: var(--crease); stroke-linecap: round; stroke-width: 5; }
  .ear-fold { fill: none; stroke: var(--crease); stroke-linecap: round; stroke-width: 5; }
  .nose { fill: #e9a271; stroke: var(--line); stroke-linejoin: round; stroke-width: 5; }
  .dimple { fill: none; stroke: var(--line); stroke-linecap: round; stroke-width: 6; }
  .brow { fill: none; stroke: #15110e; stroke-linecap: round; stroke-width: 11; }
  .eye { fill: #fff; stroke: #15110e; stroke-width: 6; }
  .pupils { fill: #15110e; }
  .nostril { fill: #5a2c1a; }
  .mouth { fill: #4a1320; stroke: #15110e; stroke-linejoin: round; stroke-width: 6; }
  .teeth { fill: #fff; }
  .tongue { fill: #e8536b; }
  .wink { fill: none; stroke: #15110e; stroke-linecap: round; stroke-width: 9; }
  .star { fill: #fffbe8; stroke: #15110e; stroke-linejoin: round; stroke-width: 5; }
  .banana { fill: #ffd84a; stroke: #15110e; stroke-linejoin: round; stroke-width: 7; }
  .banana-side { fill: #efae24; }
  .stem { fill: #6b4a1f; stroke: #15110e; stroke-linejoin: round; stroke-width: 6; }
  .witch { fill: #2f1d47; stroke: #15110e; stroke-linejoin: round; stroke-width: 7; }
  .witch-band { fill: #ee8a1f; stroke: #15110e; stroke-linejoin: round; stroke-width: 6; }
  text {
    fill: #ffd84a;
    font-family: 'Archivo Variable', 'Segoe UI', sans-serif;
    font-size: 250px;
    font-stretch: 112%;
    font-weight: 900;
    paint-order: stroke;
    stroke: #15110e;
    stroke-linejoin: round;
    stroke-width: 18px;
    text-anchor: middle;
  }

  .body { --ox: 400; --oy: 800; animation: bounce var(--beat) var(--sine) var(--start) 8 alternate; }
  .belly { --ox: 400; --oy: 688; animation: jiggle var(--beat) var(--sine) calc(var(--start) + 40ms) 8 alternate; }
  .digit { animation: hang var(--length) linear both; }

  /* The head, the brow and the banana trail the hands a little, so the swing runs through him instead of all at once. */
  .head { --a: rotate(-7deg); --b: rotate(7deg); --mid: none; --lag: 30ms; --ox: 400; --oy: 420; }
  .hat { --a: rotate(10deg); --b: rotate(-10deg); --mid: none; --lag: 80ms; --ox: 400; --oy: 192; }
  .pupils { --a: translateX(calc(var(--u) * -11)); --b: translateX(calc(var(--u) * 11)); --mid: none; --lag: -30ms; }
  .brow-up { --a: none; --b: translateY(calc(var(--u) * -13)); --mid: none; --lag: 0ms; }

  .upper { --a: rotate(4deg); --b: rotate(-4deg); --mid: none; --lag: 0ms; }
  .upper.left { --ox: 272; --oy: 472; }
  .upper.right { --ox: 528; --oy: 472; }
  .fore.left { --a: rotate(8deg); --b: rotate(-32deg); --mid: rotate(-12deg); --ox: 150; --oy: 566; }
  .fore.right { --a: rotate(32deg); --b: rotate(-8deg); --mid: rotate(12deg); --ox: 650; --oy: 566; }
  /* Each palm turns back by what the arm turned, so the digits stay upright. */
  .palm.left { --a: rotate(-12deg); --b: rotate(36deg); --mid: rotate(12deg); --ox: 84; --oy: 438; }
  .palm.right { --a: rotate(-36deg); --b: rotate(12deg); --mid: rotate(-12deg); --ox: 716; --oy: 438; }

  .lid { --ox: 446; --oy: 226; animation: lid var(--length) ease-in-out both; }
  .sparkle { --ox: 566; --oy: 222; animation: sparkle var(--length) ease-out both; }

  @keyframes swing {
    0%, 100% { transform: var(--mid, none); }
    12.5%, 62.5% { transform: var(--a); }
    37.5%, 87.5% { transform: var(--b); }
  }
  @keyframes bounce {
    from { transform: none; }
    to { transform: translateY(calc(var(--u) * 13)) scale(1.02, 0.98); }
  }
  @keyframes jiggle {
    from { transform: none; }
    to { transform: translateY(calc(var(--u) * 8)) scale(1.05, 0.95); }
  }
  @keyframes beat {
    from { transform: none; }
    to { transform: scale(1.1); }
  }
  @keyframes show {
    0% { opacity: 0; }
    8%, 80% { opacity: 1; }
    100% { opacity: 0; }
  }
  @keyframes turn {
    from { transform: rotate(-30deg); }
    to { transform: rotate(40deg); }
  }
  @keyframes fall {
    to { transform: translateY(calc(100vh + 200px)) rotate(var(--spin)); }
  }
  @keyframes band {
    0%, 5% { transform: translateX(-104%) rotate(-2deg); animation-timing-function: var(--spring); }
    19%, 66% { transform: rotate(-2deg); animation-timing-function: var(--drop); }
    80%, 100% { transform: translateX(104%) rotate(-2deg); }
  }
  @keyframes line {
    0%, 14% { opacity: 0; transform: translateY(10px); }
    24%, 66% { opacity: 1; transform: none; }
    76%, 100% { opacity: 0; transform: none; }
  }

  /* These two cancel out on purpose. From 66 % on, the digits move against the monkey with the same timing, so
     they stay where they are while he leaves. They need to stay exact negatives of each other. */
  @keyframes monkey {
    0%, 5% { transform: translateY(calc(var(--u) * 832)) scale(0.94, 1.08); animation-timing-function: var(--spring); }
    18%, 66% { transform: none; animation-timing-function: ease-out; }
    71% { transform: translateY(calc(var(--u) * -24)); animation-timing-function: var(--drop); }
    82%, 100% { transform: translateY(calc(var(--u) * 896)); }
  }
  @keyframes hang {
    0%, 66% { transform: none; animation-timing-function: ease-out; }
    71% { transform: translateY(calc(var(--u) * 24)); animation-timing-function: var(--drop); }
    82% { transform: translateY(calc(var(--u) * -896)); }
    87% { transform: translateY(calc(var(--u) * -896)); animation-timing-function: cubic-bezier(0.5, 0, 0.9, 0.5); }
    100% { transform: translateY(calc(var(--u) * 112)); }
  }

  @keyframes lid {
    0%, 62% { opacity: 0; transform: scaleY(0.15); }
    64%, 70% { opacity: 1; transform: none; }
    73%, 100% { opacity: 0; transform: scaleY(0.15); }
  }
  @keyframes sparkle {
    0%, 62% { opacity: 0; transform: scale(0.2) rotate(-40deg); }
    66% { opacity: 1; transform: scale(1.15) rotate(10deg); }
    72%, 100% { opacity: 0; transform: scale(0.5) rotate(40deg); }
  }
</style>
