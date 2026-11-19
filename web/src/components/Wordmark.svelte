<script>
  import SeasonHat from './SeasonHat.svelte';
  import logoBare from '../assets/logo-bare-small.png';
  import logoPlain from '../assets/logo-small.png';
  import { wearsHat } from '../lib/moments.svelte.js';
</script>

<span class="mark" aria-label="AMK Bar">
  <span class="badge">
    <img src={wearsHat() ? logoBare : logoPlain} alt="" width="40" height="39" />
    <SeasonHat />
  </span>
  <span class="words"><b>AMK</b><i>Bar</i></span>
</span>

<style>
  .mark {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-family: var(--display);
    font-size: 1.25rem;
    font-stretch: 125%;
    line-height: 1;
    white-space: nowrap;
  }
  .badge { position: relative; display: block; }
  img { display: block; }
  .words { display: inline-flex; align-items: baseline; gap: 0.25em; }
  b { font-weight: 900; letter-spacing: -0.02em; }
  i { color: var(--accent-text); font-style: normal; font-weight: 500; }
</style>
