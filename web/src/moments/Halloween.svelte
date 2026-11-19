<script>
  import Bats from './Bats.svelte';
  import PumpkinString from './PumpkinString.svelte';

  // A cobweb for the corner: spokes from the corner and rings that sag between them.
  function web(size, spokes, rings) {
    const point = (r, i) => {
      const angle = (i / (spokes - 1)) * (Math.PI / 2);
      return `${(r * Math.cos(angle)).toFixed(1)} ${(r * Math.sin(angle)).toFixed(1)}`;
    };
    let path = '';
    for (let i = 0; i < spokes; i++) path += `M0 0L${point(size, i)}`;
    for (let ring = 1; ring <= rings; ring++) {
      const r = (size * ring) / (rings + 0.6);
      for (let i = 0; i < spokes - 1; i++) path += `M${point(r, i)}Q${point(r * 0.84, i + 0.5)} ${point(r, i + 1)}`;
    }
    return path;
  }
  const cobweb = web(200, 7, 6);
</script>

<!-- Over the till but under dialogs, and nothing here can be tapped: every touch goes through to the keys. -->
<div class="halloween" aria-hidden="true">
  <div class="fog"></div>
  <svg class="web" viewBox="0 0 200 200"><path d={cobweb} /></svg>
  <svg class="web right" viewBox="0 0 200 200"><path d={cobweb} /></svg>
  <Bats />
  <div class="spider">
    <svg viewBox="0 0 44 40">
      <path class="legs" d="M18 18L6 8L2 16M16 21L4 20L1 30M17 24L7 30L6 39M19 27L12 34L13 40M26 18L38 8L42 16M28 21L40 20L43 30M27 24L37 30L38 39M25 27L32 34L31 40" />
      <ellipse cx="22" cy="23" rx="8" ry="10" />
      <circle cx="22" cy="12" r="5" />
      <circle class="eye" cx="20" cy="11" r="1.1" />
      <circle class="eye" cx="24" cy="11" r="1.1" />
    </svg>
  </div>
  <PumpkinString />
</div>

<style>
  .halloween {
    --web: rgb(255 255 255 / 0.3);
    position: fixed;
    inset: 0;
    z-index: 15;
    overflow: hidden;
    pointer-events: none;
  }
  /* Cobwebs need an edge to show on a light background. */
  :global(:root[data-theme='light'] .halloween), :global(:root[data-theme='pop'] .halloween), :global(:root[data-theme='y2k'] .halloween), :global(:root[data-mode='light'] .halloween) {
    --web: rgb(60 44 80 / 0.34);
  }

  /* Room for the pumpkins, so they do not hang in front of the category tabs. */
  :global([data-season='halloween'] .till) { padding-top: 44px; }
  :global([data-season='halloween'] .start) { padding-top: 104px; }
  :global([data-season='halloween'] .start .top) { top: 44px; }

  .web { position: absolute; top: 0; left: 0; width: min(24vh, 200px); height: auto; overflow: visible; }
  .web.right { right: 0; left: auto; transform: scaleX(-1); }
  .web path { fill: none; stroke: var(--web); stroke-width: 1.1; }

  .fog {
    position: absolute;
    right: -30vw;
    bottom: 0;
    left: -30vw;
    height: 24vh;
    background:
      radial-gradient(38% 90% at 22% 100%, rgb(146 112 190 / 0.3), transparent),
      radial-gradient(30% 70% at 56% 100%, rgb(120 150 130 / 0.22), transparent),
      radial-gradient(36% 80% at 84% 100%, rgb(146 112 190 / 0.26), transparent);
    will-change: transform;
    animation: fog 34s ease-in-out infinite alternate;
  }

  /* Every half minute a spider lets itself down, has a look at the till and climbs back up. */
  .spider { position: absolute; top: 0; right: 24%; width: 34px; will-change: transform; animation: spider 30s ease-in-out infinite; }
  .spider::before { position: absolute; bottom: 80%; left: 50%; width: 1.5px; height: 70vh; margin-left: -0.75px; background: var(--web); content: ''; }
  .spider svg { display: block; width: 100%; height: auto; overflow: visible; fill: #0d0814; filter: drop-shadow(0 0 1.5px rgb(255 190 120 / 0.85)); }
  .legs { fill: none; stroke: #0d0814; stroke-linecap: round; stroke-linejoin: round; stroke-width: 2.2; }
  .eye { fill: #ff4b3a; }

  @keyframes fog {
    from { transform: translate3d(-8vw, 0, 0); }
    to { transform: translate3d(8vw, 0, 0); }
  }
  @keyframes spider {
    0%, 62% { transform: translateY(-80px); }
    72% { transform: translateY(38vh); }
    74% { transform: translateY(36vh); }
    76% { transform: translateY(38vh); }
    86% { transform: translateY(38vh); }
    96%, 100% { transform: translateY(-80px); }
  }
</style>
