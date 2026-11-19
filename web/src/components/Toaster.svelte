<script>
  import { toasts } from '../lib/state.svelte.js';
</script>

<div class="toaster" role="status" aria-live="polite">
  {#each toasts as item (item.id)}
    <div class="toast {item.kind}">{item.text}</div>
  {/each}
</div>

<style>
  .toaster {
    position: fixed;
    left: 50%;
    top: 12px;
    z-index: 40;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    width: min(560px, calc(100vw - 32px));
    transform: translateX(-50%);
    pointer-events: none;
  }
  .toast {
    padding: 12px 20px;
    border: 1px solid #000;
    border-radius: var(--radius);
    background: var(--surface-3);
    font-weight: 620;
    text-align: center;
  }
  .toast.bad { background: var(--bad-bg); color: var(--toast-bad); }
  .toast.good { background: var(--good-bg); color: var(--toast-good); }
</style>
