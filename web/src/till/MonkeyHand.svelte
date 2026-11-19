<script>
  import { MIRROR, crop, flip } from './monkey.js';

  // A hand of the bar's monkey, like the waving one in the logo: open, or closed around something he holds.
  // Colours come from the monkey it belongs to, through --skin, --line, --fur and --crease.
  let { side = 'left', pose = 'open' } = $props();

  // Fingers and palm are drawn twice: first thick in the line colour, then thinner in the skin colour on top. That
  // leaves one clean outline around the whole hand and dark gaps between the fingers.
  const OPEN = {
    box: [-8, 262, 200, 190],
    palm: 'M40 360C44 340 128 336 132 360C138 390 134 420 112 432C94 440 70 440 56 430C40 416 36 390 40 360Z',
    fingers: 'M46 372L22 322M68 352L58 286M92 350L98 280M114 358L132 294',
    thumb: 'M124 402L160 376',
    creases: 'M60 384q24 16 52 -6M116 382q-18 14 -16 38',
  };
  const FIST = {
    box: [14, 332, 152, 124],
    palm: 'M46 372C46 352 122 352 122 372C126 396 122 424 104 434C90 440 74 440 62 434C46 424 42 396 46 372Z',
    fingers: 'M54 364H114M50 388H118M52 412H116',
    thumb: 'M128 412L82 438',
    creases: '',
  };

  const hand = $derived(pose === 'fist' ? FIST : OPEN);
  const box = $derived((side === 'right' ? flip : crop)(...hand.box));
</script>

<svg {...box}>
  <g transform={side === 'right' ? MIRROR : undefined}>
    <path class="outline palm" d={hand.palm} />
    <path class="outline fingers" d={hand.fingers} />
    <path class="outline thumb" d={hand.thumb} />
    <path class="flesh" d={hand.palm} />
    <path class="flesh fingers" d={hand.fingers} />
    <path class="flesh thumb" d={hand.thumb} />
    {#if hand.creases}<path class="crease" d={hand.creases} />{/if}
    <path class="cuff" d="M48 444A36 36 0 0 0 120 444L118 418L106 428L98 410L86 424L76 408L66 426L56 414L50 430Z" />
  </g>
</svg>

<style>
  svg { position: absolute; display: block; overflow: visible; }
  path { fill: none; stroke-linecap: round; stroke-linejoin: round; }

  .outline { stroke: var(--line); }
  .outline.palm { stroke-width: 14; }
  .outline.fingers { stroke-width: 40; }
  .outline.thumb { stroke-width: 42; }

  .flesh { fill: var(--skin); }
  .flesh.fingers { fill: none; stroke: var(--skin); stroke-width: 26; }
  .flesh.thumb { fill: none; stroke: var(--skin); stroke-width: 28; }

  .crease { stroke: var(--crease); stroke-width: 5; }
  .cuff { fill: var(--fur); }
</style>
