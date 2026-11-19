<script>
  import Snow from './Snow.svelte';

  // Behind the bouncing logo at Christmas: a clear night with a moon, and now and then a sleigh crossing it.
  const deer = [0, 1, 2, 3];
</script>

<div class="sky" aria-hidden="true">
  <div class="stars"></div>
  <div class="moon"></div>
  <div class="flight">
    <svg viewBox="0 0 430 90" fill="#f7efd8" stroke="#f7efd8">
      {#each deer as n}
        <g class="deer" style:--n={n} transform="translate({150 + n * 66} {n % 2 ? 8 : 16})">
          <path stroke="none" d="M14 24C20 18 38 18 46 22L52 13C53 10 56 8 59 8L65 11L64 14L58 15L54 26C50 31 44 32 38 32L22 33C16 33 12 29 14 24Z" />
          <path fill="none" stroke-linecap="round" stroke-width="2.6" d="M48 29L57 37L63 35M44 31L53 41L59 40M21 31L11 38L5 36M25 32L17 42L11 41" />
          <path fill="none" stroke-linecap="round" stroke-width="1.6" d="M58 9L56 1M57 5L52 2M61 9L63 1M62 5L67 3M14 24L9 21" />
        </g>
      {/each}
      <path class="rein" fill="none" stroke-width="0.9" d="M112 38C150 30 190 36 214 36C250 30 270 32 280 34C320 28 340 30 346 32" />
      <g transform="translate(36 22)">
        <path stroke="none" d="M6 30C14 34 58 34 74 26C80 22 86 14 90 12L94 16C90 24 84 34 74 40C60 48 22 48 8 42Z" />
        <path fill="none" stroke-linecap="round" stroke-width="3" d="M2 54H70C80 54 84 48 88 44M18 46V54M56 46V54" />
        <path stroke="none" d="M34 30C32 14 38 4 48 2C54 -2 60 -6 66 -4C60 -2 58 2 60 6C64 12 64 22 62 30Z" />
        <path stroke="none" d="M4 30C2 18 10 12 18 12C28 12 32 20 30 30Z" />
      </g>
    </svg>
    <div class="dust"></div>
  </div>
  <Snow bank={false} thick={0.7} />
</div>

<style>
  .sky { position: absolute; inset: 0; pointer-events: none; }
  .stars {
    position: absolute;
    inset: 0;
    opacity: 0.8;
    background:
      radial-gradient(1.5px 1.5px at 12% 22%, #fff, transparent), radial-gradient(1px 1px at 22% 62%, #fff, transparent),
      radial-gradient(1.5px 1.5px at 34% 12%, #fff, transparent), radial-gradient(1px 1px at 46% 38%, #fff, transparent),
      radial-gradient(1px 1px at 55% 78%, #fff, transparent), radial-gradient(1.5px 1.5px at 64% 18%, #fff, transparent),
      radial-gradient(1px 1px at 73% 52%, #fff, transparent), radial-gradient(1.5px 1.5px at 82% 30%, #fff, transparent),
      radial-gradient(1px 1px at 91% 70%, #fff, transparent), radial-gradient(1px 1px at 96% 12%, #fff, transparent),
      radial-gradient(1px 1px at 6% 84%, #fff, transparent), radial-gradient(1.5px 1.5px at 40% 90%, #fff, transparent);
  }
  .moon {
    position: absolute;
    top: 8vh;
    left: 10vw;
    width: 24vh;
    height: 24vh;
    border-radius: 50%;
    background:
      radial-gradient(circle at 62% 30%, rgb(190 172 130 / 0.5) 0 7%, transparent 8%),
      radial-gradient(circle at 34% 58%, rgb(190 172 130 / 0.45) 0 11%, transparent 12%),
      radial-gradient(circle at 70% 70%, rgb(190 172 130 / 0.4) 0 5%, transparent 6%),
      radial-gradient(circle at 36% 32%, #fffdf0, #f2e6c2 60%, #d6c594);
    box-shadow: 0 0 70px 18px rgb(255 240 200 / 0.24), 0 0 200px 90px rgb(255 240 200 / 0.08);
  }
  .flight {
    position: absolute;
    top: 0;
    left: 0;
    width: clamp(300px, 34vw, 560px);
    will-change: transform;
    animation: flight 42s linear infinite;
  }
  svg { display: block; width: 100%; height: auto; overflow: visible; filter: drop-shadow(0 0 6px rgb(255 226 160 / 0.5)); }
  .deer { animation: gallop 1100ms ease-in-out calc(var(--n) * -180ms) infinite; }
  .dust {
    position: absolute;
    top: 62%;
    right: 92%;
    width: 46%;
    height: 12%;
    background: radial-gradient(1.5px 1.5px at 8% 40%, #ffe9a8, transparent), radial-gradient(1px 1px at 24% 70%, #fff, transparent),
      radial-gradient(2px 2px at 40% 30%, #ffe9a8, transparent), radial-gradient(1px 1px at 58% 60%, #fff, transparent),
      radial-gradient(1.5px 1.5px at 74% 40%, #ffe9a8, transparent), radial-gradient(2px 2px at 90% 70%, #fff, transparent);
    mask-image: linear-gradient(90deg, transparent, #000);
  }
  @keyframes flight {
    0% { transform: translate(-46vw, 34vh) rotate(-5deg); }
    46%, 100% { transform: translate(112vw, 3vh) rotate(-5deg); }
  }
  @keyframes gallop {
    0%, 100% { translate: 0 0; }
    50% { translate: 0 -3px; }
  }
</style>
