<script>
  import { onMount } from 'svelte';

  // A rope hung along the top of the screen in loops, with a lit pumpkin under each loop. Painted once on a canvas,
  // and only the light inside the pumpkins flickers.
  const HEIGHT = 74;
  const SAG = 10;
  const LOOP = 250;

  let canvas = $state();
  let lights = $state([]);

  const curve = (x, width, loops) => {
    const u = ((x / width) * loops) % 1;
    return 6 + SAG * (1 - (2 * u - 1) ** 2);
  };

  function pumpkin(ctx, x, y, r) {
    ctx.save();
    ctx.translate(x, y);
    ctx.strokeStyle = '#17120c';
    ctx.lineWidth = 2;

    ctx.fillStyle = '#4c7a34';
    ctx.beginPath();
    ctx.moveTo(-3, -r * 0.82);
    ctx.quadraticCurveTo(-2, -r * 1.2, 3, -r * 1.3);
    ctx.lineTo(6, -r * 1.12);
    ctx.quadraticCurveTo(3, -r * 1, 3, -r * 0.82);
    ctx.closePath();
    ctx.fill();
    ctx.stroke();

    for (const [offset, wide, tone] of [[-0.62, 0.52, '#e47612'], [0.62, 0.52, '#e47612'], [-0.28, 0.5, '#f28a1e'], [0.28, 0.5, '#f28a1e'], [0, 0.42, '#f79a2c']]) {
      ctx.fillStyle = tone;
      ctx.beginPath();
      ctx.ellipse(r * offset, 0, r * wide, r * 0.86, 0, 0, Math.PI * 2);
      ctx.fill();
      ctx.stroke();
    }

    // The face, lit from inside.
    ctx.fillStyle = '#ffe27a';
    for (const side of [-1, 1]) {
      ctx.beginPath();
      ctx.moveTo(side * r * 0.5, -r * 0.14);
      ctx.lineTo(side * r * 0.18, -r * 0.14);
      ctx.lineTo(side * r * 0.32, -r * 0.46);
      ctx.closePath();
      ctx.fill();
    }
    ctx.beginPath();
    ctx.moveTo(-r * 0.62, r * 0.14);
    ctx.quadraticCurveTo(0, r * 0.36, r * 0.62, r * 0.14);
    ctx.quadraticCurveTo(r * 0.5, r * 0.6, r * 0.22, r * 0.56);
    ctx.lineTo(r * 0.14, r * 0.38);
    ctx.lineTo(0, r * 0.58);
    ctx.lineTo(-r * 0.14, r * 0.38);
    ctx.lineTo(-r * 0.22, r * 0.56);
    ctx.quadraticCurveTo(-r * 0.5, r * 0.6, -r * 0.62, r * 0.14);
    ctx.fill();
    ctx.restore();
  }

  function paint() {
    const width = window.innerWidth;
    const ratio = Math.min(window.devicePixelRatio || 1, 2);
    canvas.width = width * ratio;
    canvas.height = HEIGHT * ratio;
    const ctx = canvas.getContext('2d');
    ctx.scale(ratio, ratio);

    const loops = Math.max(2, Math.round(width / LOOP));
    ctx.strokeStyle = '#8c7a62';
    ctx.lineWidth = 2.5;
    ctx.beginPath();
    for (let x = 0; x <= width; x += 3) {
      const y = curve(x, width, loops);
      if (x === 0) ctx.moveTo(x, y);
      else ctx.lineTo(x, y);
    }
    ctx.stroke();

    const row = [];
    for (let i = 0; i < loops; i++) {
      const x = ((i + 0.5) / loops) * width;
      const size = i % 2 ? 12 : 14;
      const y = curve(x, width, loops) + size + 6;
      ctx.strokeStyle = '#8c7a62';
      ctx.lineWidth = 2;
      ctx.beginPath();
      ctx.moveTo(x, curve(x, width, loops));
      ctx.lineTo(x, y - size);
      ctx.stroke();
      pumpkin(ctx, x, y, size);
      row.push({ x, y, delay: (i * 730) % 2600, length: 1500 + ((i * 410) % 1500) });
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

<div class="string" aria-hidden="true">
  {#each lights as light}
    <i style:left="{light.x}px" style:top="{light.y}px" style:animation-delay="-{light.delay}ms" style:animation-duration="{light.length}ms"></i>
  {/each}
  <canvas bind:this={canvas}></canvas>
</div>

<style>
  .string { position: absolute; inset: 0 0 auto 0; height: 74px; }
  canvas { position: absolute; inset: 0; width: 100%; height: 100%; }
  i {
    position: absolute;
    width: 48px;
    height: 48px;
    margin: -24px;
    border-radius: 50%;
    background: radial-gradient(closest-side, rgb(255 170 60 / 0.5), transparent);
    animation: glimmer ease-in-out infinite;
  }
  @keyframes glimmer {
    0%, 100% { opacity: 1; transform: scale(1); }
    22% { opacity: 0.55; transform: scale(0.92); }
    40% { opacity: 0.95; }
    63% { opacity: 0.6; transform: scale(1.05); }
  }
</style>
