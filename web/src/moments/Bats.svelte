<script>
  // Bats crossing the screen at different heights, some of them the other way. The wings beat from the shoulders, a few
  // strokes and then a short glide, while the bat drifts up and down on its own slower beat, so the path never has a
  // corner and no two flights look alike. Each moving part is a div: the browser cannot move an svg or anything in it smoothly.
  let { count = 4 } = $props();

  const WING = 'M29 10C24 6 19 4 13 3C10 5 5 8 0 14C4 12 7 14 9 19C12 15 15 16 18 21C21 18 25 19 29 24Z';
  const BODY = 'M32 31C29 27 27 22 27 16C27 13 28 11 29 9C28 7 28 4 28.5 2L31 5.5H33L35.5 2C36 4 36 7 35 9C36 11 37 13 37 16C37 22 35 27 32 31Z';

  const bats = Array.from({ length: count }, (_, i) => ({
    top: 10 + ((i * 37) % 58),
    size: 34 + ((i * 17) % 30),
    time: 24 + ((i * 11) % 16), // half of it across the screen, half out of sight
    wait: -((i * 7) % 19),
    rise: 5 + ((i * 3) % 6),
    drift: 2.3 + ((i * 7) % 11) / 10,
    beat: 1.2 + ((i * 3) % 5) / 10,
    back: i % 2 === 1,
  }));
</script>

{#each bats as bat}
  <div
    class="bat"
    class:back={bat.back}
    style:--top="{bat.top}vh"
    style:--size="{bat.size}px"
    style:--time="{bat.time}s"
    style:--wait="{bat.wait}s"
    style:--rise="{bat.rise}vh"
    style:--drift="{bat.drift}s"
    style:--beat="{bat.beat}s"
  >
    <div class="flyer">
      <div class="wing"><svg viewBox="0 0 64 34"><path d={WING} /></svg></div>
      <div class="wing right"><svg viewBox="0 0 64 34"><path d={WING} /></svg></div>
      <svg viewBox="0 0 64 34"><path d={BODY} /></svg>
    </div>
  </div>
{/each}

<style>
  .bat {
    --lean: 9deg;
    position: absolute;
    top: var(--top);
    left: 0;
    width: var(--size);
    will-change: translate, rotate, transform;
    animation:
      bat-across var(--time) linear var(--wait) infinite,
      bat-rise var(--drift) ease-in-out infinite alternate,
      bat-lean var(--drift) ease-in-out calc(var(--drift) * -0.5) infinite alternate;
  }
  .bat.back { --lean: -9deg; animation-direction: reverse, alternate, alternate; }

  /* The body lifts as the wings push down. */
  .flyer {
    position: relative;
    will-change: translate, transform;
    animation:
      bat-bob var(--beat) ease-in-out infinite,
      bat-jink calc(var(--drift) * 0.63) ease-in-out infinite alternate;
  }
  .wing {
    --side: 1;
    position: absolute;
    inset: 0;
    transform-origin: 45% 32%;
    will-change: transform;
    animation: bat-flap var(--beat) ease-in-out infinite;
  }
  .wing.right { --side: -1; transform-origin: 55% 32%; }
  .wing.right svg { transform: scaleX(-1); }
  svg {
    position: relative;
    display: block;
    width: 100%;
    height: auto;
    overflow: visible;
    fill: #0d0814;
    filter: drop-shadow(0 0 1.5px rgb(255 190 120 / 0.85));
  }

  @keyframes bat-across {
    from { translate: -12vw 0; }
    50%, to { translate: 112vw 0; }
  }
  @keyframes bat-rise {
    from { transform: translateY(calc(var(--rise) * -1)); }
    to { transform: translateY(var(--rise)); }
  }
  /* A quarter beat behind the rise, so the bat leans while it climbs or drops, not at the top and bottom of its path. */
  @keyframes bat-lean {
    from { rotate: calc(var(--lean) * -1); }
    to { rotate: var(--lean); }
  }
  @keyframes bat-jink {
    from { translate: 1.2vw -1.5vh; }
    to { translate: -1.2vw 1.5vh; }
  }
  /* Three strokes, then a glide with the wings held a little up. */
  @keyframes bat-flap {
    0%, 82%, 100% { transform: rotate(calc(var(--side) * 14deg)) scaleX(1); }
    10%, 34%, 58% { transform: rotate(calc(var(--side) * 38deg)) scaleX(0.84); }
    22%, 46%, 70% { transform: rotate(calc(var(--side) * -30deg)) scaleX(0.9); }
  }
  @keyframes bat-bob {
    0%, 82%, 100% { transform: translateY(0); }
    10%, 34%, 58% { transform: translateY(6%); }
    22%, 46%, 70% { transform: translateY(-8%); }
  }
</style>
