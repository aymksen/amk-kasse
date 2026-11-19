<script module>
  // Seven-segment digits as SVG paths, one path per digit, so a new number redraws almost nothing.
  const W = 50;
  const H = 96;
  const T = 11;
  const GAP = 1.4;
  const CELL = W + 15;
  const SLANT = 10;

  function bar(x1, y1, x2, y2) {
    const h = T / 2;
    if (y1 === y2) return `M${x1} ${y1}L${x1 + h} ${y1 - h}H${x2 - h}L${x2} ${y2}L${x2 - h} ${y2 + h}H${x1 + h}Z`;
    return `M${x1} ${y1}L${x1 + h} ${y1 + h}V${y2 - h}L${x1} ${y2}L${x1 - h} ${y2 - h}V${y1 + h}Z`;
  }

  const left = T / 2;
  const right = W - T / 2;
  const top = T / 2;
  const middle = H / 2;
  const bottom = H - T / 2;

  const SEGMENTS = {
    a: bar(left + GAP, top, right - GAP, top),
    b: bar(right, top + GAP, right, middle - GAP),
    c: bar(right, middle + GAP, right, bottom - GAP),
    d: bar(left + GAP, bottom, right - GAP, bottom),
    e: bar(left, middle + GAP, left, bottom - GAP),
    f: bar(left, top + GAP, left, middle - GAP),
    g: bar(left + GAP, middle, right - GAP, middle),
  };

  const ALL = Object.values(SEGMENTS).join('');
  const GLYPHS = {};
  for (const [char, segments] of Object.entries({
    0: 'abcdef', 1: 'bc', 2: 'abdeg', 3: 'abcdg', 4: 'bcfg', 5: 'acdfg', 6: 'acdefg', 7: 'abc', 8: 'abcdefg', 9: 'abcdfg', '-': 'g', ' ': '',
  })) {
    GLYPHS[char] = [...segments].map((s) => SEGMENTS[s]).join('');
  }

  function toCells(text, digits) {
    const cells = [];
    for (const char of text) {
      if (char === ',' && cells.length) cells[cells.length - 1].dot = true;
      else if (char === ':' && cells.length) cells[cells.length - 1].colon = true;
      else if (char in GLYPHS) cells.push({ path: GLYPHS[char], dot: false });
    }
    while (cells.length < digits) cells.unshift({ path: '', dot: false });
    return cells;
  }
</script>

<script>
  // Pass `cents` for money or `text` for a plain count. `size` is the digit height in pixels.
  let { cents = null, text = null, digits = 6, size = 52, label = '', unit = '', tone = 'amber', plain = false, dance = false } = $props();

  const shown = $derived(cents !== null ? (cents / 100).toFixed(2).replace('.', ',') : String(text ?? ''));
  const cells = $derived(toCells(shown, digits));
  const width = $derived(cells.length * CELL + SLANT);
</script>

<div class="led {tone}" class:plain class:dance>
  {#if label || unit}
    <div class="caption">
      <span>{label}</span>
      <span>{unit}</span>
    </div>
  {/if}
  <div class="face">
    <svg viewBox="0 0 {width} {H}" width={(width / H) * size} height={size} role="img" aria-label={`${label} ${shown} ${unit}`.trim()}>
      <g transform="translate({SLANT} 0) skewX(-5.5)">
        {#each cells as cell, i}
          <g transform="translate({i * CELL} 0)">
            <path class="ghost" d={ALL} />
            <rect class="ghost" x={W + 3} y={H - T} width={T - 1} height={T} />
            {#if cell.path}<path class="lit" d={cell.path} />{/if}
            {#if cell.dot}<rect class="lit" x={W + 3} y={H - T} width={T - 1} height={T} />{/if}
            {#if cell.colon}
              <rect class="lit" x={W + 3} y={28} width={T - 1} height={T} />
              <rect class="lit" x={W + 3} y={58} width={T - 1} height={T} />
            {/if}
          </g>
        {/each}
      </g>
    </svg>
    {#if dance}
      <!-- Anything that moves inside an SVG redraws the page in every frame, so these two digits move as boxes around their own SVG. -->
      {#each [-3, -2] as place}
        <div class="hop">
          <svg viewBox="0 0 {width} {H}" aria-hidden="true">
            <g transform="translate({SLANT} 0) skewX(-5.5)">
              <path class="lit" transform="translate({(cells.length + place) * CELL} 0)" d={cells.at(place).path} />
            </g>
          </svg>
        </div>
      {/each}
    {/if}
  </div>
</div>

<style>
  .led {
    --glow: var(--led-amber);
    --ghost: var(--led-amber-off);
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 4px;
    min-width: 0;
    padding: 8px 12px 10px;
    border-radius: var(--radius);
    background: var(--led-bg);
    box-shadow: var(--dent), var(--dent-edge);
    color: var(--glow);
  }
  .led.green { --glow: var(--led-green); --ghost: var(--led-green-off); }
  .led.red { --glow: var(--led-red); --ghost: var(--led-red-off); }
  /* A figure from earlier, beside the one that counts now: the same display, turned down. */
  .led.dim { --glow: color-mix(in srgb, var(--led-amber) 58%, transparent); }
  .led.plain { padding: 0; background: none; box-shadow: none; }
  .caption {
    display: flex;
    justify-content: space-between;
    gap: 12px;
    width: 100%;
    font-family: var(--display);
    font-size: 0.6875rem;
    font-weight: var(--display-weight);
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }
  .face { position: relative; max-width: 100%; }
  svg { display: block; max-width: 100%; }
  path.ghost, rect.ghost { fill: var(--ghost); }
  path.lit, rect.lit { fill: var(--glow); stroke: var(--glow); stroke-opacity: 0.22; stroke-width: 3; }

  /* 6,70 € on the display: the 6 and the 7 hop out of their places in turns, the way the hands go in the meme. */
  .dance g:nth-last-child(3) path.lit, .dance g:nth-last-child(2) path.lit { visibility: hidden; }
  .hop { position: absolute; inset: 0; will-change: transform; animation: hop 1280ms cubic-bezier(0.37, 0, 0.63, 1) 140ms; }
  .hop svg { width: 100%; height: 100%; }
  .hop:nth-of-type(1) { --a: -21%; --b: 21%; }
  .hop:nth-of-type(2) { --a: 21%; --b: -21%; }
  @keyframes hop {
    12.5%, 62.5% { transform: translateY(var(--a)); }
    37.5%, 87.5% { transform: translateY(var(--b)); }
  }
</style>
