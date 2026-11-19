<script>
  import { onMount } from 'svelte';
  import { theme } from '../lib/theme.svelte.js';

  // Snow in four depths: tiny far-off dots, soft flakes, a few real crystals close by, and a couple of big blurred
  // ones right in front of the lens. Whatever lands builds up a bank on the bottom bar. Each flake is painted once
  // onto a small canvas and from then on only copied, which is far cheaper than drawing it again in every frame.
  // `bank` is off where there is no bar to land on.
  let { bank = true, thick = 1 } = $props();

  const COLUMN = 6; // width of one column of the bank, in pixels
  const BANK_MAX = 25;

  let canvas = $state();
  let sprites = null;

  const between = (a, b) => a + Math.random() * (b - a);

  function softDot(size, core, color, edge) {
    const sprite = document.createElement('canvas');
    sprite.width = sprite.height = size;
    const ctx = sprite.getContext('2d');
    const r = size / 2;
    if (edge) {
      const rim = ctx.createRadialGradient(r, r, 0, r, r, r);
      rim.addColorStop(0.45, edge);
      rim.addColorStop(1, 'transparent');
      ctx.fillStyle = rim;
      ctx.fillRect(0, 0, size, size);
    }
    const dot = ctx.createRadialGradient(r, r, 0, r, r, r);
    dot.addColorStop(0, color);
    dot.addColorStop(core, color);
    dot.addColorStop(1, 'rgb(255 255 255 / 0)');
    ctx.fillStyle = dot;
    ctx.fillRect(0, 0, size, size);
    return sprite;
  }

  // A six-armed snow crystal.
  function crystal(size, color, edge, variant) {
    const sprite = document.createElement('canvas');
    sprite.width = sprite.height = size;
    const ctx = sprite.getContext('2d');
    const c = size / 2;
    const arm = size * 0.44;
    ctx.translate(c, c);
    ctx.lineCap = 'round';
    ctx.strokeStyle = color;
    ctx.lineWidth = size / 26;
    ctx.shadowColor = edge || color;
    ctx.shadowBlur = size / 10;
    for (let i = 0; i < 6; i++) {
      ctx.rotate(Math.PI / 3);
      ctx.beginPath();
      ctx.moveTo(0, 0);
      ctx.lineTo(0, -arm);
      const branches = variant === 0 ? [0.42, 0.68] : [0.3, 0.55, 0.78];
      branches.forEach((at, n) => {
        const reach = arm * (variant === 0 ? 0.3 - n * 0.06 : 0.26 - n * 0.05);
        ctx.moveTo(0, -arm * at);
        ctx.lineTo(-reach, -arm * at - reach);
        ctx.moveTo(0, -arm * at);
        ctx.lineTo(reach, -arm * at - reach);
      });
      ctx.stroke();
    }
    return sprite;
  }

  function paintSprites() {
    const style = getComputedStyle(document.documentElement);
    const edge = style.getPropertyValue('--snow-edge').trim();
    const color = '#ffffff';
    sprites = {
      far: softDot(16, 0.3, color, edge),
      soft: softDot(32, 0.45, color, edge),
      lens: softDot(96, 0.15, color, ''),
      crystals: [crystal(72, color, edge, 0), crystal(72, color, edge, 1)],
    };
  }

  function makeFlake(layer, width, height, anywhere) {
    const f = { layer, x: Math.random() * width, y: anywhere ? Math.random() * height : between(-60, -10), phase: Math.random() * 6.3, turn: 0, spin: 0, sprite: 0 };
    if (layer === 0) Object.assign(f, { size: between(2.5, 5), fall: between(16, 30), sway: between(4, 10), alpha: between(0.3, 0.6) });
    if (layer === 1) Object.assign(f, { size: between(6, 11), fall: between(38, 62), sway: between(10, 22), alpha: between(0.7, 0.95) });
    if (layer === 2) Object.assign(f, { size: between(14, 25), fall: between(66, 104), sway: between(16, 34), alpha: 1, spin: between(-0.7, 0.7), turn: Math.random() * 6.3, sprite: Math.round(Math.random()) });
    if (layer === 3) Object.assign(f, { size: between(46, 84), fall: between(120, 170), sway: between(20, 40), alpha: between(0.12, 0.22) });
    return f;
  }

  $effect(() => {
    // A dark style and a light one need a flake with a different edge.
    void theme.id;
    void theme.mode;
    paintSprites();
  });

  onMount(() => {
    const ctx = canvas.getContext('2d');
    const ratio = Math.min(window.devicePixelRatio || 1, 1.5);
    let width = 0;
    let height = 0;
    let ground = 0;
    let flakes = [];
    let heap = new Float32Array(0);
    let glints = [];
    let frame;

    function size() {
      width = window.innerWidth;
      height = window.innerHeight;
      canvas.width = width * ratio;
      canvas.height = height * ratio;
      const area = Math.min(1.25, Math.max(0.5, (width * height) / (1920 * 1080))) * thick;
      const counts = [100, 56, 15, 3];
      flakes = counts.flatMap((count, layer) => Array.from({ length: Math.round(count * area) }, () => makeFlake(layer, width, height, true)));

      // Start with some snow already there, so the bank does not look like it has just been switched on.
      heap = new Float32Array(Math.ceil(width / COLUMN) + 3);
      for (let i = 0; i < heap.length; i++) heap[i] = 6 + 3 * Math.sin(i * 0.09) + 2.5 * Math.sin(i * 0.031 + 1) + 1.5 * Math.sin(i * 0.27);
      glints = Array.from({ length: Math.round(width / 70) }, () => ({ x: Math.random() * width, phase: Math.random() * 6.3, speed: between(0.8, 2) }));
    }

    function findGround() {
      const dock = bank && document.querySelector('.dock');
      ground = dock ? dock.getBoundingClientRect().top : height + 90;
    }

    function land(x) {
      const column = Math.floor(x / COLUMN) + 1;
      for (let i = -3; i <= 3; i++) {
        const c = column + i;
        if (c >= 0 && c < heap.length && heap[c] < BANK_MAX) heap[c] += 0.11 * (1 - Math.abs(i) / 4);
      }
    }

    function drawBank(time) {
      ctx.setTransform(ratio, 0, 0, ratio, 0, 0);
      const fill = ctx.createLinearGradient(0, ground - BANK_MAX, 0, ground + 4);
      fill.addColorStop(0, '#ffffff');
      fill.addColorStop(1, '#c8d9ef');
      ctx.fillStyle = fill;
      ctx.beginPath();
      ctx.moveTo(-COLUMN, ground + 4);
      ctx.lineTo(-COLUMN, ground - heap[0]);
      for (let i = 1; i < heap.length - 1; i++) {
        const x = (i - 1) * COLUMN;
        ctx.quadraticCurveTo(x, ground - heap[i], x + COLUMN / 2, ground - (heap[i] + heap[i + 1]) / 2);
      }
      ctx.lineTo(width + COLUMN, ground + 4);
      ctx.closePath();
      ctx.fill();

      ctx.fillStyle = '#ffffff';
      for (const glint of glints) {
        const shine = Math.sin(time * glint.speed + glint.phase);
        if (shine < 0.8) continue;
        const column = Math.floor(glint.x / COLUMN) + 1;
        const y = ground - heap[column] + 3;
        ctx.globalAlpha = (shine - 0.8) * 5;
        ctx.fillRect(glint.x - 3, y - 0.5, 6, 1);
        ctx.fillRect(glint.x - 0.5, y - 3, 1, 6);
      }
      ctx.globalAlpha = 1;
    }

    let last = performance.now();
    let lastGround = 0;
    function step(now) {
      const seconds = Math.min(now - last, 60) / 1000;
      last = now;
      const time = now / 1000;
      if (now - lastGround > 1000) {
        lastGround = now;
        findGround();
      }
      // Wind that comes and goes, with the odd gust.
      const wind = 16 * Math.sin(time * 0.13) + 10 * Math.sin(time * 0.37 + 1) + 22 * Math.max(0, Math.sin(time * 0.05) - 0.75) * 4;

      ctx.setTransform(1, 0, 0, 1, 0, 0);
      ctx.clearRect(0, 0, canvas.width, canvas.height);

      for (const f of flakes) {
        const depth = f.fall / 60;
        f.y += f.fall * seconds;
        f.x += (wind * depth + Math.cos(time * 0.9 + f.phase) * f.sway) * seconds;
        f.turn += f.spin * seconds;
        if (f.x < -90) f.x += width + 180;
        if (f.x > width + 90) f.x -= width + 180;

        const column = Math.floor(f.x / COLUMN) + 1;
        const floor = bank && heap[column] !== undefined ? ground - heap[column] + 2 : ground;
        if (f.y > (f.layer === 0 ? ground : f.layer === 3 ? height + 90 : floor)) {
          if (bank && (f.layer === 1 || f.layer === 2)) land(f.x);
          Object.assign(f, makeFlake(f.layer, width, height, false));
          continue;
        }

        const image = f.layer === 0 ? sprites.far : f.layer === 1 ? sprites.soft : f.layer === 2 ? sprites.crystals[f.sprite] : sprites.lens;
        const r = f.size / 2;
        ctx.globalAlpha = f.alpha;
        if (f.layer === 2) {
          const cos = Math.cos(f.turn) * ratio;
          const sin = Math.sin(f.turn) * ratio;
          ctx.setTransform(cos, sin, -sin, cos, f.x * ratio, f.y * ratio);
          ctx.drawImage(image, -r, -r, f.size, f.size);
        } else {
          ctx.setTransform(ratio, 0, 0, ratio, 0, 0);
          ctx.drawImage(image, f.x - r, f.y - r, f.size, f.size);
        }
      }
      ctx.globalAlpha = 1;
      if (bank) drawBank(time);
      frame = requestAnimationFrame(step);
    }

    size();
    findGround();
    window.addEventListener('resize', size);
    frame = requestAnimationFrame(step);
    return () => {
      cancelAnimationFrame(frame);
      window.removeEventListener('resize', size);
    };
  });
</script>

<canvas class="snow" bind:this={canvas} aria-hidden="true"></canvas>

<style>
  .snow { position: absolute; inset: 0; width: 100%; height: 100%; }
</style>
