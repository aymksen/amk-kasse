<script>
  import { onMount } from 'svelte';
  import MonkeyHand from './MonkeyHand.svelte';
  import { ARM_SHEEN, CHEST_TUFT, EAR_FOLD, FACE, FOREARM, FOREARM_SHEEN, MIRROR, NOSE, TORSO, UPPER_ARM, crop, flip } from './monkey.js';

  // The devil's version of the 6,70 € monkey: he comes up out of the fire, three sixes land, one of them on his belly,
  // and he sinks back. It only looks. Every tap goes through it, and Sell.svelte ends it as soon as the total changes.
  const LENGTH = 1700;

  let { onend, line } = $props();

  // A row of flame tongues, 1000 units wide. The numbers only have to look uneven, not be random.
  function tongues(count, height, step) {
    const wide = 1000 / count;
    let path = 'M0 220';
    for (let i = 0; i < count; i++) {
      const x = i * wide;
      const tall = height * (0.5 + (((i + 1) * step) % 7) / 12);
      path += `C${x + wide * 0.2} 200 ${x + wide * 0.05} ${220 - tall * 0.5} ${x + wide * 0.42} ${220 - tall}`;
      path += `C${x + wide * 0.46} ${220 - tall * 0.55} ${x + wide * 0.9} ${220 - tall * 0.45} ${x + wide} 220`;
    }
    return `${path}V240H0Z`;
  }

  const fires = [
    { path: tongues(9, 210, 5), tone: '#7c0f14', time: 310 },
    { path: tongues(13, 165, 3), tone: '#e0301e', time: 240 },
    { path: tongues(17, 110, 4), tone: '#ffb224', time: 190 },
  ];

  onMount(() => {
    const timer = setTimeout(onend, LENGTH);
    return () => clearTimeout(timer);
  });
</script>

<div class="devil" aria-hidden="true">
  <div class="dim"></div>
  <div class="heat"></div>
  <div class="aura"></div>

  <div class="sixes">
    <b class="first">6</b>
    <b class="third">6</b>
  </div>

  <div class="monkey">
    <div class="part tail">
      <svg {...crop(540, 560, 220, 270)}>
        <path class="rope" d="M560 770C640 800 730 750 722 668C718 622 678 610 664 640" />
        <path class="spade" d="M668 646L626 618L680 584L696 640Z" />
      </svg>
    </div>

    <div class="part body">
      <svg {...crop(190, 390, 420, 446)}><path class="fur" d={TORSO} /></svg>
      <svg {...crop(292, 500, 216, 264)}>
        <ellipse class="skin" cx="400" cy="640" rx="104" ry="120" />
        <path class="fur" transform="translate(0 -28)" d={CHEST_TUFT} />
      </svg>
      <div class="part mark"><svg {...crop(320, 540, 160, 190)}><text x="400" y="714">6</text></svg></div>

      <div class="part arm left">
        <svg {...crop(50, 404, 272, 200)}>
          <path class="fur" d={UPPER_ARM} />
          <path class="fur" d={FOREARM} />
          <path class="sheen" d={ARM_SHEEN} />
          <path class="sheen" d={FOREARM_SHEEN} />
        </svg>
        <div class="part fireball">
          <svg {...crop(16, 166, 120, 136)}>
            <g transform="translate(0 -96)">
              <path class="flame-out" d="M76 266C84 300 128 312 128 348C128 376 104 394 76 394C48 394 24 376 24 348C24 322 44 318 50 296C58 306 62 312 66 314C70 298 70 282 76 266Z" />
              <path class="flame-in" d="M76 312C82 330 102 338 102 356C102 372 90 382 76 382C62 382 50 372 50 356C50 340 68 332 76 312Z" />
            </g>
          </svg>
        </div>
        <MonkeyHand />
      </div>

      <div class="part arm right">
        <svg {...crop(630, 76, 164, 570)}>
          <path class="pole" d="M716 200V640" />
          <path class="pole" d="M660 124C660 190 690 206 716 206C742 206 772 190 772 124" />
          <path class="pole" d="M716 110V206" />
          <path class="tip" d="M646 136L660 92L674 136Z" />
          <path class="tip" d="M702 122L716 78L730 122Z" />
          <path class="tip" d="M758 136L772 92L786 136Z" />
        </svg>
        <svg {...flip(50, 404, 272, 200)}>
          <g transform={MIRROR}>
            <path class="fur" d={UPPER_ARM} />
            <path class="fur" d={FOREARM} />
            <path class="sheen" d={ARM_SHEEN} />
            <path class="sheen" d={FOREARM_SHEEN} />
          </g>
        </svg>
        <MonkeyHand side="right" pose="fist" />
      </div>

      <div class="part head">
        <svg {...crop(214, 80, 372, 350)}>
          <path class="horn" d="M288 216C256 182 244 132 262 86C276 132 304 156 342 172Z" />
          <path class="horn" d="M512 216C544 182 556 132 538 86C524 132 496 156 458 172Z" />
          <circle class="fur" cx="268" cy="300" r="48" />
          <circle class="fur" cx="532" cy="300" r="48" />
          <circle class="skin" cx="266" cy="302" r="28" />
          <circle class="skin" cx="534" cy="302" r="28" />
          <path class="ear-fold" d={EAR_FOLD} />
          <path class="ear-fold" transform={MIRROR} d={EAR_FOLD} />
          <ellipse class="fur" cx="400" cy="296" rx="132" ry="128" />
          <path class="skin" d={FACE} />
          <ellipse class="eye" cx="354" cy="270" rx="29" ry="24" />
          <ellipse class="eye" cx="446" cy="270" rx="29" ry="24" />
          <ellipse class="slit" cx="359" cy="270" rx="5" ry="19" />
          <ellipse class="slit" cx="441" cy="270" rx="5" ry="19" />
          <path class="brow" d="M312 226L388 258" />
          <path class="brow" d="M488 226L412 258" />
          <path class="nose" d={NOSE} />
          <ellipse class="nostril" cx="388" cy="324" rx="5" ry="4" />
          <ellipse class="nostril" cx="412" cy="324" rx="5" ry="4" />
          <path class="mouth" d="M316 340Q400 378 484 340Q476 414 400 416Q324 414 316 340Z" />
          <path class="teeth" d="M326 348L342 376L358 360L376 388L394 366H406L424 388L442 360L458 376L474 348Q400 374 326 348Z" />
        </svg>
      </div>
    </div>
  </div>

  <div class="fire">
    {#each fires as fire}
      <svg viewBox="0 0 1000 240" preserveAspectRatio="none" style:--time="{fire.time}ms"><path fill={fire.tone} d={fire.path} /></svg>
    {/each}
  </div>

  <p class="line">{line}</p>
</div>

<style>
  /* Everything leaves together, so the last moment is him going down into the flames. */
  .devil {
    --size: min(92vh, 77vw);
    --u: calc(var(--size) / 800); /* one unit of the 800 × 800 drawing */
    --fur: #2c0b0e;
    --line: #120808;
    --skin: #e5392c;
    --crease: #8f1d18;
    --length: 1700ms;
    --spring: cubic-bezier(0.34, 1.5, 0.64, 1);
    --sine: cubic-bezier(0.37, 0, 0.63, 1);
    position: fixed;
    inset: 0;
    z-index: 35;
    overflow: hidden;
    pointer-events: none;
    contain: strict;
  }
  .dim { position: absolute; inset: 0; background: rgb(22 2 3 / 0.8); animation: show var(--length) ease-in-out both; }
  .heat {
    position: absolute;
    right: -10vw;
    bottom: -30vh;
    left: -10vw;
    height: 90vh;
    border-radius: 50%;
    background: radial-gradient(closest-side, rgb(255 96 32 / 0.75), rgb(190 20 20 / 0.38) 55%, transparent);
    animation: show var(--length) ease-in-out both;
  }

  /* Keeps his dark fur apart from the dark background. */
  .aura {
    position: absolute;
    bottom: 6vh;
    left: 50%;
    width: var(--size);
    height: var(--size);
    margin-left: calc(var(--size) / -2);
    border-radius: 50%;
    background: radial-gradient(closest-side, rgb(214 50 34 / 0.6) 0 56%, rgb(170 26 26 / 0.22) 82%, transparent);
    animation: show var(--length) ease-in-out both;
  }

  .fire { position: absolute; right: -6vw; bottom: 0; left: -6vw; height: 17vh; animation: fire var(--length) linear both; }
  .fire svg { position: absolute; inset: 0; width: 100%; height: 100%; transform-origin: 50% 100%; will-change: transform; animation: flick var(--time) var(--sine) infinite alternate; }

  .sixes {
    position: absolute;
    right: 0;
    bottom: 26vh;
    left: 0;
    display: grid;
    grid-template-columns: 1fr 1fr;
    justify-items: center;
    column-gap: 68vh;
    color: #ffd84a;
    font-family: 'Archivo Variable', 'Segoe UI', sans-serif;
    font-size: min(calc((50vw - 34vh) * 0.86), 46vh);
    font-stretch: 112%;
    font-weight: 900;
    line-height: 1;
    animation: leave var(--length) linear both;
  }
  .sixes b { display: block; font-weight: inherit; -webkit-text-stroke: 8px #2c0b0e; paint-order: stroke; animation: slam 300ms var(--spring) backwards; }
  .first { animation-delay: 300ms !important; }
  .third { animation-delay: 700ms !important; }

  .line {
    position: absolute;
    bottom: 18px;
    left: 50%;
    max-width: calc(100vw - 32px);
    margin: 0;
    padding: 6px 16px;
    border-radius: 8px;
    background: #120808;
    color: #fff5dc;
    font-family: 'IBM Plex Mono', Consolas, monospace;
    font-size: 1.25rem;
    font-weight: 600;
    text-align: center;
    transform: translateX(-50%);
    white-space: nowrap;
    animation: line var(--length) ease-out both;
  }

  /* These boxes have no size on purpose, as in SixSeven.svelte: only the pictures become layers. */
  .monkey {
    position: absolute;
    top: calc(94vh - var(--size));
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

  .tail { --ox: 560; --oy: 770; animation: wag 220ms var(--sine) 8 alternate; }
  .head { --ox: 400; --oy: 420; animation: laugh 120ms var(--sine) 240ms 11 alternate; }
  .arm.left { --ox: 272; --oy: 472; animation: lift 150ms var(--sine) 240ms 9 alternate; }
  .arm.right { --ox: 528; --oy: 472; animation: pump 150ms var(--sine) 240ms 9 alternate; }
  .fireball { --ox: 76; --oy: 298; animation: flare 140ms var(--sine) 8 alternate; }
  .mark { --ox: 400; --oy: 640; animation: brand 340ms var(--spring) 500ms backwards; }

  /* Upright screen: he fills the width, so the two sixes move above him. */
  @media (orientation: portrait) {
    .sixes { bottom: calc(var(--size) + 8vh); column-gap: 0; font-size: 30vw; }
    .fire { height: 10vh; }
  }

  .fur { fill: var(--fur); }
  .skin { fill: var(--skin); }
  .sheen { fill: none; stroke: #7a1f22; stroke-linecap: round; stroke-width: 6; }
  .ear-fold { fill: none; stroke: var(--crease); stroke-linecap: round; stroke-width: 5; }
  .nose { fill: #c42a20; stroke: var(--line); stroke-linejoin: round; stroke-width: 5; }
  .rope { fill: none; stroke: #2c0b0e; stroke-linecap: round; stroke-width: 22; }
  .spade { fill: #e5392c; stroke: #120808; stroke-linejoin: round; stroke-width: 6; }
  .horn { fill: #f6e7c8; stroke: #120808; stroke-linejoin: round; stroke-width: 7; }
  .eye { fill: #ffe14d; stroke: #120808; stroke-width: 6; }
  .slit { fill: #120808; }
  .brow { fill: none; stroke: #120808; stroke-linecap: round; stroke-width: 15; }
  .nostril { fill: #5a1410; }
  .mouth { fill: #3a0a10; stroke: #120808; stroke-linejoin: round; stroke-width: 6; }
  .teeth { fill: #fff; }
  .pole { fill: none; stroke: #c98f2b; stroke-linecap: round; stroke-width: 15; }
  .tip { fill: #ffd84a; stroke: #120808; stroke-linejoin: round; stroke-width: 6; }
  .flame-out { fill: #ff7a1a; stroke: #120808; stroke-linejoin: round; stroke-width: 6; }
  .flame-in { fill: #ffe14d; }
  text {
    fill: #ffd84a;
    font-family: 'Archivo Variable', 'Segoe UI', sans-serif;
    font-size: 200px;
    font-stretch: 112%;
    font-weight: 900;
    paint-order: stroke;
    stroke: #120808;
    stroke-linejoin: round;
    stroke-width: 16px;
    text-anchor: middle;
  }

  @keyframes show {
    0% { opacity: 0; }
    9%, 84% { opacity: 1; }
    100% { opacity: 0; }
  }
  @keyframes fire {
    0% { transform: translateY(100%); animation-timing-function: var(--spring); }
    12%, 88% { transform: none; animation-timing-function: ease-in; }
    100% { transform: translateY(100%); }
  }
  @keyframes flick {
    from { transform: scale(1, 1) translateX(-1%); }
    to { transform: scale(1.03, 1.22) translateX(1%); }
  }
  @keyframes monkey {
    0%, 3% { transform: translateY(calc(var(--u) * 860)) scale(0.94, 1.08); animation-timing-function: var(--spring); }
    17%, 74% { transform: none; animation-timing-function: cubic-bezier(0.55, 0, 0.85, 0.35); }
    92%, 100% { transform: translateY(calc(var(--u) * 900)); }
  }
  @keyframes slam {
    from { opacity: 0; transform: scale(2.6) rotate(-14deg); }
  }
  @keyframes brand {
    from { opacity: 0; transform: scale(2.2); }
  }
  @keyframes leave {
    0%, 74% { opacity: 1; }
    88%, 100% { opacity: 0; }
  }
  @keyframes line {
    0%, 26% { opacity: 0; }
    34%, 76% { opacity: 1; }
    90%, 100% { opacity: 0; }
  }
  @keyframes wag {
    from { transform: rotate(-9deg); }
    to { transform: rotate(9deg); }
  }
  @keyframes laugh {
    from { transform: rotate(-4deg) translateY(calc(var(--u) * -6)); }
    to { transform: rotate(4deg) translateY(calc(var(--u) * 6)); }
  }
  @keyframes pump {
    from { transform: rotate(3deg); }
    to { transform: rotate(-5deg) translateY(calc(var(--u) * -14)); }
  }
  @keyframes lift {
    from { transform: rotate(-3deg); }
    to { transform: rotate(5deg) translateY(calc(var(--u) * -10)); }
  }
  @keyframes flare {
    from { transform: scale(1); }
    to { transform: scale(1.12, 1.24); }
  }
</style>
