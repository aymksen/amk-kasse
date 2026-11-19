<script>
  import Modal from '../components/Modal.svelte';
  import { post } from '../lib/api.js';
  import { t } from '../lib/i18n.svelte.js';
  import { toastError } from '../lib/state.svelte.js';

  let { categories, onclose, onchanged } = $props();

  // svelte-ignore state_referenced_locally
  let rows = $state(categories.map((c) => ({ ...c })));
  let name = $state('');
  let busy = $state(false);

  async function save(category) {
    try {
      await post('/api/admin/categories', category);
      onchanged();
    } catch (err) {
      toastError(err);
    }
  }

  async function add(e) {
    e.preventDefault();
    if (busy || !name.trim()) return;
    busy = true;
    try {
      const created = await post('/api/admin/categories', { name: name.trim(), color: '#9a9284', sort: rows.length, active: true });
      rows.push(created);
      name = '';
      onchanged();
    } catch (err) {
      toastError(err);
    } finally {
      busy = false;
    }
  }
</script>

<Modal title={t('cat.title')} {onclose}>
  <div class="stack">
    {#each rows as category (category.id)}
      <div class="row">
        <input class="swatch" type="color" bind:value={category.color} onchange={() => save(category)} aria-label={t('cat.color')} />
        <input class="input grow" bind:value={category.name} onchange={() => save(category)} aria-label={t('common.name')} autocomplete="off" />
        <label class="check small"><input type="checkbox" bind:checked={category.active} onchange={() => save(category)} /> {t('common.active')}</label>
      </div>
    {/each}
    <form class="row" onsubmit={add}>
      <input class="input grow" bind:value={name} placeholder={t('cat.add')} aria-label={t('cat.add')} autocomplete="off" />
      <button class="btn" disabled={busy || !name.trim()}>{t('common.add')}</button>
    </form>
  </div>
</Modal>

<style>
  .swatch { flex-shrink: 0; width: 44px; height: 44px; padding: 4px; border: var(--hair) solid var(--line); border-radius: var(--radius); background: var(--bg); }
</style>
