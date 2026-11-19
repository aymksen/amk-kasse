<script>
  import { onMount } from 'svelte';
  import ImagePlus from 'lucide-svelte/icons/image-plus';
  import Plus from 'lucide-svelte/icons/plus';
  import ScanBarcode from 'lucide-svelte/icons/scan-barcode';
  import X from 'lucide-svelte/icons/x';
  import MoneyInput from './MoneyInput.svelte';
  import Modal from '../components/Modal.svelte';
  import { post, remove, upload } from '../lib/api.js';
  import { t } from '../lib/i18n.svelte.js';
  import { pictureUrl, shrinkPicture } from '../lib/picture.js';
  import { listenForScans } from '../lib/scanner.js';
  import { toast, toastError } from '../lib/state.svelte.js';

  let { product, categories, onclose, onsaved } = $props();

  // svelte-ignore state_referenced_locally
  let form = $state({ ...product, barcodes: [...(product.barcodes ?? [])] });
  // A product can have several codes: one key such as Chips sells every kind, and each bag has its own code.
  let newCode = $state('');
  let codeInput = $state();

  function addCode(code = newCode) {
    code = code.trim();
    if (!code) return;
    if (form.barcodes.includes(code)) toast(t('prod.codeTwice'));
    else form.barcodes.push(code);
    newCode = '';
  }

  // A scan while no field has the focus lands in the list too, so the OBK can scan bag after bag.
  onMount(() => listenForScans({ onScan: (code) => addCode(code) }));
  let busy = $state(false);
  // A new picture waits here until Save, so Cancel really cancels it.
  let picture = $state(null);
  let pictureRemoved = $state(false);
  // svelte-ignore state_referenced_locally
  let preview = $state(pictureUrl(product));

  async function usePicture(file) {
    if (!file) return;
    try {
      picture = await shrinkPicture(file);
      pictureRemoved = false;
      preview = URL.createObjectURL(picture);
    } catch {
      toast(t('prod.imageBad'), 'bad');
    }
  }

  function dropPicture() {
    picture = null;
    pictureRemoved = true;
    preview = null;
  }

  function pasted(e) {
    const file = e.clipboardData?.files[0];
    if (!file) return;
    e.preventDefault();
    usePicture(file);
  }

  async function submit(e) {
    e.preventDefault();
    if (busy) return;
    addCode(); // a code typed but not added yet counts too
    busy = true;
    try {
      const saved = await post('/api/admin/products', {
        ...form,
        crateSize: Number(form.crateSize) || 1,
        barTarget: Number(form.barTarget) || 0,
        minStock: Number(form.minStock) || 0,
        categoryId: Number(form.categoryId),
      });
      form.id = saved.id; // a failed upload must not create the product twice on the next try
      if (picture) await upload(`/api/admin/products/${saved.id}/image`, picture);
      else if (pictureRemoved) await remove(`/api/admin/products/${saved.id}/image`);
      toast(t('common.saved'), 'good');
      onsaved();
    } catch (err) {
      toastError(err);
    } finally {
      busy = false;
    }
  }
</script>

<Modal title={form.id ? t('prod.edit') : t('prod.new')} wide {onclose}>
  <form class="form" onsubmit={submit} onpaste={pasted}>
    <div class="picture span2">
      <div
        class="plate"
        role="img"
        aria-label={t('prod.image')}
        ondragover={(e) => e.preventDefault()}
        ondrop={(e) => {
          e.preventDefault();
          usePicture(e.dataTransfer.files[0]);
        }}
      >
        {#if preview}<img src={preview} alt="" />{:else}<ImagePlus size={28} />{/if}
      </div>
      <div class="stack">
        <div class="row wrap">
          <label class="btn">
            <ImagePlus size={18} /> {t('prod.imagePick')}
            <input
              class="file"
              type="file"
              accept="image/*"
              onchange={(e) => {
                usePicture(e.target.files[0]);
                e.target.value = '';
              }}
            />
          </label>
          {#if preview}
            <button type="button" class="btn" onclick={dropPicture}><X size={18} /> {t('prod.imageRemove')}</button>
          {/if}
        </div>
        <p class="muted small">{t('prod.imageHint')}</p>
      </div>
    </div>
    <label class="field span2">
      <span>{t('common.name')}</span>
      <input class="input" bind:value={form.name} required autocomplete="off" />
    </label>
    <label class="field">
      <span>{t('common.category')}</span>
      <select class="input" bind:value={form.categoryId}>
        {#each categories as c}
          <option value={c.id}>{c.name}</option>
        {/each}
      </select>
    </label>
    <label class="field">
      <span>{t('prod.location')}</span>
      <input class="input" bind:value={form.location} placeholder={t('prod.locationHint')} autocomplete="off" />
    </label>
    <label class="field">
      <span>{t('prod.price')} (€)</span>
      <MoneyInput bind:cents={form.priceCents} />
    </label>
    <label class="field">
      <span>{t('prod.cost')} (€)</span>
      <MoneyInput bind:cents={form.costCents} />
    </label>
    <p class="muted small span2">{t('prod.negativeHint')}</p>
    <div class="field span2">
      <label for="new-code">{t('prod.barcodes')}</label>
      {#if form.barcodes.length}
        <ul class="codes" aria-label={t('prod.barcodes')}>
          {#each form.barcodes as code, i (code)}
            <li class="code num">
              {code}
              <button type="button" class="drop" aria-label={t('prod.codeRemove', { code })} title={t('prod.codeRemove', { code })} onclick={() => form.barcodes.splice(i, 1)}>
                <X size={16} />
              </button>
            </li>
          {/each}
        </ul>
      {/if}
      <div class="row">
        <input
          id="new-code"
          class="input num grow"
          bind:this={codeInput}
          bind:value={newCode}
          placeholder={t('prod.codeNew')}
          autocomplete="off"
          onkeydown={(e) => {
            if (e.key !== 'Enter') return;
            e.preventDefault();
            addCode();
          }}
        />
        <button type="button" class="btn" disabled={!newCode.trim()} onclick={() => { addCode(); codeInput.focus(); }}><Plus size={18} /> {t('prod.codeAdd')}</button>
      </div>
      <span class="muted small row hint"><ScanBarcode size={16} /> {t('prod.barcodesHint')}</span>
    </div>
    <label class="field">
      <span>{t('prod.crateBarcode')}</span>
      <input class="input num" bind:value={form.crateBarcode} autocomplete="off" onkeydown={(e) => e.key === 'Enter' && e.preventDefault()} />
    </label>
    <label class="field">
      <span>{t('prod.crateSize')}</span>
      <input class="input num" type="number" min="1" bind:value={form.crateSize} />
    </label>
    <div class="pair span2">
      <label class="field">
        <span>{t('prod.barTarget')}</span>
        <input class="input num" type="number" min="0" bind:value={form.barTarget} disabled={!form.trackStock} />
      </label>
      <label class="field">
        <span>{t('prod.min')}</span>
        <input class="input num" type="number" min="0" bind:value={form.minStock} disabled={!form.trackStock} />
      </label>
    </div>
    <div class="checks span2">
      <label class="check"><input type="checkbox" bind:checked={form.trackStock} /> {t('prod.track')}</label>
      <label class="check"><input type="checkbox" bind:checked={form.active} /> {t('prod.active')}</label>
    </div>
    <p class="muted small span2">{t('prod.stockHint')}</p>
    <div class="row span2 actions">
      <button type="button" class="btn big grow" onclick={onclose}>{t('common.cancel')}</button>
      <button class="btn big primary grow" disabled={busy}>{t('common.save')}</button>
    </div>
  </form>
</Modal>

<style>
  .form { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
  .span2 { grid-column: span 2; }
  .pair { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
  .codes { display: flex; flex-wrap: wrap; gap: 6px; margin: 0 0 4px; padding: 0; list-style: none; }
  .code {
    display: inline-flex;
    align-items: center;
    gap: 2px;
    padding: 2px 2px 2px 10px;
    border-radius: 6px;
    background: var(--surface-2);
    box-shadow: inset 0 0 0 1px var(--line);
    font-size: 0.9375rem;
  }
  .drop {
    display: grid;
    place-items: center;
    width: 30px;
    height: 30px;
    border: 0;
    border-radius: 5px;
    background: none;
    color: var(--muted);
    cursor: pointer;
  }
  .drop:hover { background: var(--bad); color: #fff; }
  .drop:focus-visible { outline: 2px solid var(--focus); outline-offset: 1px; }
  .hint { gap: 6px; align-items: flex-start; }
  .hint :global(svg) { flex: none; margin-top: 1px; }
  .picture { display: flex; align-items: center; gap: 16px; }
  .plate {
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    width: 104px;
    height: 104px;
    border-radius: 6px;
    background: #fff;
    box-shadow: 0 0 0 1px var(--line);
    color: var(--muted);
    overflow: hidden;
  }
  .plate img { width: 100%; height: 100%; object-fit: contain; }
  .picture .btn { position: relative; cursor: pointer; }
  .file { position: absolute; width: 1px; height: 1px; opacity: 0; }
  .checks { display: flex; flex-wrap: wrap; gap: 0 24px; }
  /* Save stays in reach on a 768 px tall screen, where this form scrolls. */
  .actions {
    position: sticky;
    bottom: -20px;
    z-index: 1;
    margin: 0 -20px -20px;
    padding: 12px 20px;
    border-top: var(--hair) solid var(--line);
    background: var(--surface);
  }
  @media (max-width: 560px) {
    .form, .pair { grid-template-columns: 1fr; }
    .span2 { grid-column: auto; }
  }
</style>
