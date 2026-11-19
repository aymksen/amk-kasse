<script>
  import { onMount } from 'svelte';
  import OnlineCard from './OnlineCard.svelte';
  import { get, post } from '../lib/api.js';
  import { dateTime } from '../lib/format.js';
  import { t } from '../lib/i18n.svelte.js';
  import { calmNow, startPreview } from '../lib/moments.svelte.js';
  import { route } from '../lib/route.svelte.js';
  import { pickLanguage, refresh, toast, toastError } from '../lib/state.svelte.js';

  let form = $state(null);
  let backups = $state([]);
  let pin = $state({ current: '', next: '' });
  let busy = $state(false);

  const size = (bytes) => `${(bytes / 1048576).toFixed(1).replace('.', ',')} MB`;

  async function load() {
    try {
      [form, backups] = await Promise.all([get('/api/admin/settings'), get('/api/admin/backups')]);
    } catch (err) {
      toastError(err);
    }
  }

  async function save(e) {
    e.preventDefault();
    if (busy) return;
    busy = true;
    try {
      await post('/api/admin/settings', form);
      await refresh();
      pickLanguage(form.language);
      toast(t('common.saved'), 'good');
    } catch (err) {
      toastError(err);
    } finally {
      busy = false;
    }
  }

  async function backupNow() {
    if (busy) return;
    busy = true;
    try {
      await post('/api/admin/settings', form);
      await post('/api/admin/backups');
      backups = await get('/api/admin/backups');
      toast(t('set.backupDone'), 'good');
    } catch (err) {
      toastError(err);
    } finally {
      busy = false;
    }
  }

  async function changePIN(e) {
    e.preventDefault();
    if (busy) return;
    busy = true;
    try {
      await post('/api/admin/pin', { current: pin.current, new: pin.next });
      pin = { current: '', next: '' };
      toast(t('set.pinDone'), 'good');
    } catch (err) {
      toastError(err);
    } finally {
      busy = false;
    }
  }

  function preview(kind) {
    startPreview(kind);
    route.go('/');
  }

  onMount(load);
</script>

<div class="stack gap">
  <h1>{t('admin.settings')}</h1>
  {#if form}
    <form class="card pad stack" onsubmit={save}>
      <h2>{t('set.general')}</h2>
      <div class="pair">
        <label class="field">
          <span>{t('set.barName')}</span>
          <input class="input" bind:value={form.barName} required autocomplete="off" />
        </label>
        <label class="field">
          <span>{t('set.language')}</span>
          <select class="input" bind:value={form.language}>
            <option value="de">Deutsch</option>
            <option value="en">English</option>
          </select>
        </label>
      </div>
      <label class="field">
        <span>{t('set.other')}</span>
        <input class="input" bind:value={form.otherPayment} autocomplete="off" />
        <span class="small">{t('set.otherHint')}</span>
      </label>
      <label class="field">
        <span>{t('set.backupDir')}</span>
        <input class="input" bind:value={form.backupDir} autocomplete="off" />
        <span class="small">{t('set.backupDirHint')}</span>
      </label>
      <div><button class="btn primary" disabled={busy}>{t('common.save')}</button></div>
    </form>

    <OnlineCard />

    <section class="card pad stack">
      <h2>{t('set.preview')}</h2>
      <p class="small">{calmNow() ? t('set.previewCalm') : t('set.previewHint')}</p>
      <div class="row wrap">
        <button class="btn" disabled={calmNow()} onclick={() => preview('screensaver')}>{t('set.showSaver')}</button>
        <button class="btn" disabled={calmNow()} onclick={() => preview('halloween')}>{t('set.showHalloween')}</button>
        <button class="btn" disabled={calmNow()} onclick={() => preview('christmas')}>{t('set.showChristmas')}</button>
        <button class="btn" disabled={calmNow()} onclick={() => preview('newyear')}>{t('set.showNewYear')}</button>
      </div>
    </section>

    <section class="card">
      <div class="spread head">
        <h2>{t('set.backup')}</h2>
        <button class="btn" disabled={busy} onclick={backupNow}>{t('set.backupNow')}</button>
      </div>
      {#if backups.length}
        <div class="table-wrap">
          <table class="table">
            <tbody>
              {#each backups.slice(0, 8) as backup}
                <tr>
                  <td class="num">{dateTime(backup.time)}</td>
                  <td class="muted small">{backup.name}</td>
                  <td class="right num muted">{size(backup.size)}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {:else}
        <p class="empty">{t('set.backupNone')}</p>
      {/if}
    </section>

    <form class="card pad stack" onsubmit={changePIN}>
      <h2>{t('set.pin')}</h2>
      <div class="pair even">
        <label class="field">
          <span>{t('set.pinCurrent')}</span>
          <input class="input" type="password" bind:value={pin.current} autocomplete="current-password" required />
        </label>
        <label class="field">
          <span>{t('set.pinNew')}</span>
          <input class="input" type="password" bind:value={pin.next} autocomplete="new-password" minlength="4" required />
        </label>
      </div>
      <div><button class="btn" disabled={busy}>{t('set.pin')}</button></div>
    </form>
  {/if}
</div>

<style>
  .gap { gap: 16px; max-width: 760px; }
  .pair { display: grid; grid-template-columns: minmax(0, 1fr) 180px; gap: 12px; }
  .pair.even { grid-template-columns: 1fr 1fr; }
  .head { padding: 12px 16px; }
</style>
