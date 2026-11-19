<script>
  import SantaHat from './SantaHat.svelte';
  import WitchHat from './WitchHat.svelte';
  import { moments } from '../lib/moments.svelte.js';
</script>

<!-- Goes with the logo without the banana. See wearsHat in lib/moments.svelte.js. -->
{#if moments.season?.kind === 'christmas'}
  <SantaHat />
{:else if moments.season?.kind === 'halloween'}
  <WitchHat />
{/if}
