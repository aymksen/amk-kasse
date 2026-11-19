<script>
  import { onMount } from 'svelte';
  import Toaster from './components/Toaster.svelte';
  import Moments from './moments/Moments.svelte';
  import Till from './till/Till.svelte';
  import StartShift from './till/StartShift.svelte';
  import { errorText, t } from './lib/i18n.svelte.js';
  import { route } from './lib/route.svelte.js';
  import { refresh, till } from './lib/state.svelte.js';

  const isAdmin = $derived(route.path.startsWith('/admin'));
  let failure = $state(null);

  async function load() {
    failure = null;
    try {
      await refresh();
    } catch (err) {
      // Another device without a session goes straight to the admin login.
      if (err.status === 401) route.go('/admin');
      else failure = err;
    }
  }

  onMount(load);
</script>

{#if isAdmin}
  {#await import('./admin/Admin.svelte') then { default: Admin }}
    <Admin />
  {/await}
{:else if failure}
  <div class="center">
    <p>{errorText(failure)}</p>
    <button class="btn big" onclick={load}>{t('common.next')}</button>
  </div>
{:else if !till.loaded}
  <div class="center muted">{t('common.loading')}</div>
{:else if till.shift}
  <Till />
  <Moments />
{:else}
  <StartShift />
  <Moments />
{/if}

<Toaster />

<style>
  .center {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 16px;
    height: 100%;
    padding: 16px;
  }
</style>
