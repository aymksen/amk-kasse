<script>
  import { onMount } from 'svelte';
  import { get, post } from '../lib/api.js';
  import { dateTime } from '../lib/format.js';
  import { t } from '../lib/i18n.svelte.js';
  import { toast, toastError } from '../lib/state.svelte.js';

  let status = $state(null);
  let url = $state('');
  let key = $state('');
  let busy = $state(false);

  function show(next) {
    status = next;
    url = next.url;
    key = '';
  }

  // Both can take a while on bad internet: the till waits for the answer from the web page here.
  async function run(path, body) {
    if (busy) return;
    busy = true;
    try {
      show(await post(path, body));
      if (status.lastError) toast(t(`set.online_${status.lastError}`), 'bad');
      else toast(t(status.url ? 'set.onlineSent' : 'common.saved'), 'good');
    } catch (err) {
      toastError(err);
    } finally {
      busy = false;
    }
  }

  function save(e) {
    e.preventDefault();
    run('/api/admin/online', { url, key });
  }

  onMount(async () => {
    try {
      show(await get('/api/admin/online'));
    } catch (err) {
      toastError(err);
    }
  });
</script>

{#if status}
  <form class="card pad stack" onsubmit={save}>
    <h2>{t('set.online')}</h2>
    <p class="small">{t('set.onlineHint')}</p>
    <label class="field">
      <span>{t('set.onlineUrl')}</span>
      <input class="input" type="url" inputmode="url" placeholder="https://amk-bar.….workers.dev" bind:value={url} autocomplete="off" spellcheck="false" />
    </label>
    <label class="field">
      <span>{t('set.onlineKey')}</span>
      <input class="input" type="password" bind:value={key} autocomplete="new-password" required={Boolean(url) && !status.hasKey} placeholder={status.hasKey ? '••••••••' : ''} />
      {#if status.hasKey}<span class="small">{t('set.onlineKeyKept')}</span>{/if}
    </label>
    {#if status.url}
      <p class="small state" class:bad={status.lastError} role="status">
        {#if status.lastError}
          {t(`set.online_${status.lastError}`)}
        {:else if status.lastPush}
          {t('set.onlineLast', { time: dateTime(status.lastPush) })}
        {:else}
          {t('set.onlineNever')}
        {/if}
      </p>
    {/if}
    <p class="small">{t('set.onlineSends')}</p>
    <div class="row wrap">
      <button class="btn primary" disabled={busy}>{t('common.save')}</button>
      {#if status.url && status.hasKey}
        <button class="btn" type="button" disabled={busy} onclick={() => run('/api/admin/online/push')}>{t('set.onlineSend')}</button>
      {/if}
    </div>
  </form>
{/if}

<style>
  .state { color: var(--text); font-weight: 600; }
  .state.bad { color: var(--bad); }
</style>
