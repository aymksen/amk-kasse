<script>
  import Bats from './Bats.svelte';

  // Behind the bouncing logo at Halloween: a big orange moon, bats, and now and then the monkey on a broom.
</script>

<div class="sky" aria-hidden="true">
  <div class="stars"></div>
  <div class="moon"></div>
  <Bats count={6} />
  <div class="flight">
    <svg viewBox="-10 -40 320 150">
      <path class="stick" d="M12 76L300 56" />
      <path d="M66 72L2 46L12 70L-8 78L14 84L0 108L68 80Z" />
      <path d="M170 64C150 50 112 54 88 34C96 62 134 74 174 70Z" />
      <path d="M150 66C148 40 164 26 184 26C202 26 214 42 210 62Z" />
      <path class="stick thin" d="M152 62C132 42 118 44 122 60C124 66 132 64 130 58" />
      <path class="stick thin" d="M186 64L200 92L214 90" />
      <path class="stick thin" d="M198 40L232 58" />
      <circle cx="190" cy="16" r="15" />
      <circle cx="175" cy="14" r="6" />
      <circle cx="206" cy="14" r="6" />
      <path d="M178 6C182 -8 196 -24 212 -32C208 -20 208 -6 212 6Z" />
      <path class="stick thin" d="M168 8H218" />
    </svg>
  </div>
</div>

<style>
  .sky { position: absolute; inset: 0; pointer-events: none; }
  .stars {
    position: absolute;
    inset: 0;
    opacity: 0.7;
    background:
      radial-gradient(1.5px 1.5px at 8% 18%, #fff, transparent), radial-gradient(1px 1px at 18% 58%, #fff, transparent),
      radial-gradient(1.5px 1.5px at 30% 10%, #fff, transparent), radial-gradient(1px 1px at 42% 42%, #fff, transparent),
      radial-gradient(1px 1px at 52% 82%, #fff, transparent), radial-gradient(1.5px 1.5px at 60% 22%, #fff, transparent),
      radial-gradient(1px 1px at 92% 64%, #fff, transparent), radial-gradient(1px 1px at 96% 10%, #fff, transparent),
      radial-gradient(1px 1px at 5% 88%, #fff, transparent), radial-gradient(1.5px 1.5px at 38% 92%, #fff, transparent);
  }
  .moon {
    position: absolute;
    top: 9vh;
    right: 9vw;
    width: 40vh;
    height: 40vh;
    border-radius: 50%;
    background:
      radial-gradient(circle at 62% 30%, rgb(180 90 20 / 0.34) 0 8%, transparent 9%),
      radial-gradient(circle at 34% 58%, rgb(180 90 20 / 0.3) 0 12%, transparent 13%),
      radial-gradient(circle at 70% 72%, rgb(180 90 20 / 0.28) 0 6%, transparent 7%),
      radial-gradient(circle at 36% 32%, #ffe6a6, #ffb648 60%, #e8821e);
    box-shadow: 0 0 80px 22px rgb(255 160 60 / 0.3), 0 0 220px 100px rgb(255 140 40 / 0.1);
  }
  .flight {
    position: absolute;
    top: 0;
    left: 0;
    width: clamp(260px, 30vw, 500px);
    will-change: transform;
    animation: broom 36s linear infinite;
  }
  svg { display: block; width: 100%; height: auto; overflow: visible; fill: #0d0814; filter: drop-shadow(0 0 3px rgb(255 170 90 / 0.75)); }
  .stick { fill: none; stroke: #0d0814; stroke-linecap: round; stroke-linejoin: round; stroke-width: 7; }
  .stick.thin { stroke-width: 6; }
  @keyframes broom {
    0% { transform: translate(-40vw, 56vh) rotate(-6deg); }
    50%, 100% { transform: translate(112vw, 12vh) rotate(-6deg); }
  }
</style>
