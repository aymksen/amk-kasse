<script>
  import { onMount } from 'svelte';
  import { bauble, bough, seeded } from './pine.js';

  // A fir garland hung along the top of the screen in loops, with little lights that twinkle one by one.
  const HEIGHT = 64;
  const SAG = 17;
  const LOOP = 270;
  const BULBS = ['#ffd27a', '#ff5247', '#ffd27a', '#4cd67c', '#ffd27a', '#55b6ff'];

  let canvas = $state();
  let lights = $state([]);

  const curve = (x, width, loops) => {
    const u = ((x / width) * loops) % 1;
    return 8 + SAG * (1 - (2 * u - 1) ** 2);
  };

  function paint() {
    const width = window.innerWidth;
    const ratio = Math.min(window.devicePixelRatio || 1, 2);
    canvas.width = width * ratio;
    canvas.height = HEIGHT * ratio;
    const ctx = canvas.getContext('2d');
    ctx.scale(ratio, ratio);

    const random = seeded(4711);
    const loops = Math.max(2, Math.round(width / LOOP));
    const points = [];
    for (let x = -6; x <= width + 6; x += 2.2) {
      const here = curve(Math.max(x, 0), width, loops);
      const next = curve(Math.max(x + 4, 0), width, loops);
      points.push({ x, y: here, angle: Math.atan2(next - here, 4) + (random() < 0.5 ? Math.PI : 0) });
    }
    bough(ctx, points, { thick: 18, density: 4, length: 12, random });

    // Under each loop a ball, in turns red and gold.
    for (let i = 0; i < loops; i++) {
      const x = ((i + 0.5) / loops) * width;
      if (i % 2) bauble(ctx, x, 48, 8, '#f2c14e', '#7c5208');
      else bauble(ctx, x, 48, 8, '#e23a3a', '#6d0f16');
    }

    const row = [];
    const count = Math.round(width / 38);
    for (let i = 0; i < count; i++) {
      const x = ((i + 0.5) / count) * width + (random() - 0.5) * 16;
      row.push({
        x,
        y: curve(x, width, loops) + (random() - 0.5) * 12,
        color: BULBS[i % BULBS.length],
        delay: Math.round(random() * 5000),
        length: 2400 + Math.round(random() * 2800),
      });
    }
    lights = row;
  }

  onMount(() => {
    paint();
    let timer;
    const resized = () => {
      clearTimeout(timer);
      timer = setTimeout(paint, 200);
    };
    window.addEventListener('resize', resized);
    return () => {
      clearTimeout(timer);
      window.removeEventListener('resize', resized);
    };
  });
</script>

<div class="garland" aria-hidden="true">
  <canvas bind:this={canvas}></canvas>
  {#each lights as light}
    <i style:left="{light.x}px" style:top="{light.y}px" style:--bulb={light.color} style:animation-delay="-{light.delay}ms" style:animation-duration="{light.length}ms"></i>
  {/each}
</div>

<style>
  .garland { position: absolute; inset: 0 0 auto 0; height: 64px; }
  canvas { position: absolute; inset: 0; width: 100%; height: 100%; }
  i {
    position: absolute;
    width: 7px;
    height: 7px;
    margin: -3.5px;
    border-radius: 50%;
    background: var(--bulb);
    box-shadow: 0 0 5px 2px var(--bulb), 0 0 18px 7px color-mix(in srgb, var(--bulb) 50%, transparent);
    animation: twinkle ease-in-out infinite;
  }
  @keyframes twinkle {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.3; }
  }
</style>
