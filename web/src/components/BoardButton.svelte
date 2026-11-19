<script>
  import Trophy from 'lucide-svelte/icons/trophy';
  import { get } from '../lib/api.js';
  import { t } from '../lib/i18n.svelte.js';
  import { toastError } from '../lib/state.svelte.js';

  // Opens the leaderboard. The code of the board only loads when someone asks for it. `labelled` shows the word next
  // to the cup, for where there is room for it.
  let { labelled = false, primary = false } = $props();

  let recap = $state(null);

  async function open() {
    try {
      recap = await get('/api/recap');
    } catch (err) {
      toastError(err);
    }
  }

  // The button sits in bars with their own layer. The board has to lie over everything, so it moves to the page itself.
  function onThePage(node) {
    document.body.append(node);
    return { destroy: () => node.remove() };
  }
</script>

<button class="btn" class:ghost={!primary} class:primary aria-label={t('board.title')} title={t('board.title')} onclick={open}>
  <Trophy size={labelled ? 18 : 20} />
  {#if labelled}{t('board.title')}{/if}
</button>

{#if recap}
  <div use:onThePage>
    {#await import('../moments/Leaderboard.svelte') then { default: Leaderboard }}
      <Leaderboard {recap} onclose={() => (recap = null)} />
    {/await}
  </div>
{/if}
