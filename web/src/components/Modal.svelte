<script>
  import X from 'lucide-svelte/icons/x';
  import { t } from '../lib/i18n.svelte.js';

  let { title, onclose, wide = false, children } = $props();

  function focusFirst(node) {
    node.querySelector('[data-autofocus], input, select, textarea')?.focus();
  }

  function onkeydown(e) {
    if (e.key === 'Escape') onclose?.();
  }
</script>

<svelte:window {onkeydown} />

<div
  class="backdrop"
  role="presentation"
  onclick={(e) => {
    if (e.target === e.currentTarget) onclose?.();
  }}
>
  <div class="modal" class:wide role="dialog" aria-modal="true" aria-label={title} use:focusFirst>
    <header>
      <h2>{title}</h2>
      {#if onclose}
        <button class="icon-btn" onclick={onclose} aria-label={t('common.close')}><X size={22} /></button>
      {/if}
    </header>
    <div class="body">{@render children()}</div>
  </div>
</div>

<style>
  .backdrop {
    position: fixed;
    inset: 0;
    z-index: 20;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 16px;
    background: var(--backdrop);
  }
  .modal {
    display: flex;
    flex-direction: column;
    width: 100%;
    max-width: 440px;
    max-height: 100%;
    border: var(--modal-edge);
    border-radius: var(--radius-lg);
    background: var(--surface);
    box-shadow: var(--modal-shadow);
  }
  :global(.ledger) .modal { border-color: var(--text); }
  .modal.wide { max-width: 720px; }
  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    min-height: 60px;
    padding: 8px 8px 8px 20px;
    border-bottom: var(--hair) solid var(--line);
  }
  .body { padding: 20px; overflow-y: auto; }
</style>
