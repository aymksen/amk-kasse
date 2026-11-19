<script>
  import { onMount } from 'svelte';
  import StockCheck from './StockCheck.svelte';
  import { get, post } from '../lib/api.js';
  import { t } from '../lib/i18n.svelte.js';
  import { refresh, till, toast, toastError } from '../lib/state.svelte.js';

  // Zählen in the dock: the check of the bar at any time of the shift, the same as at its start and end.
  let { ondone } = $props();

  let data = $state(null);
  let busy = $state(false);

  // Fetched fresh, so what was sold a minute ago is already taken off.
  onMount(async () => {
    try {
      data = await get('/api/stock/check');
    } catch (err) {
      toastError(err);
      ondone();
    }
  });

  async function save(lines, changes) {
    if (!lines.length) return ondone();
    busy = true;
    try {
      await post('/api/stocktakes', { kind: 'full', room: till.shift.workerRoom, name: till.shift.workerName, lines });
      await refresh();
      toast(changes ? t('check.savedChanges', { n: changes }) : t('check.saved'), 'good');
      ondone();
    } catch (err) {
      toastError(err);
    } finally {
      busy = false;
    }
  }
</script>

{#if data}
  <StockCheck when="during" {data} {busy} ondone={save} />
{/if}
