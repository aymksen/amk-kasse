<script>
  import MessageCircle from 'lucide-svelte/icons/message-circle';
  import Modal from '../components/Modal.svelte';
  import { dateTime } from '../lib/format.js';
  import { t } from '../lib/i18n.svelte.js';

  // What the last shift left for this one. It shows once, right after the shift has started.
  let { message, onclose } = $props();
</script>

<Modal title={t('hand.title', { name: message.from })} {onclose}>
  <div class="stack">
    <div class="note">
      <MessageCircle size={24} />
      <p>{message.text}</p>
    </div>
    <p class="muted small num">{dateTime(message.at)}</p>
    <button class="btn big primary block" data-autofocus onclick={onclose}>{t('hand.ok')}</button>
  </div>
</Modal>

<style>
  .note { display: flex; gap: 12px; }
  .note > :global(svg) { flex: none; margin-top: 2px; color: var(--accent-text); }
  p { margin: 0; overflow-wrap: anywhere; }
  .note p { font-size: 1.25rem; font-weight: 600; line-height: 1.35; }
</style>
