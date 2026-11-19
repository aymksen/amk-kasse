<script>
  let { label, value, note = '', tone = '' } = $props();
</script>

<div class="tile">
  <span class="muted small">{label}</span>
  <strong class="value num {tone}">{value}</strong>
  {#if note}<span class="muted small">{note}</span>{/if}
</div>

<style>
  .tile { display: flex; flex-direction: column; gap: 2px; min-width: 0; padding: 12px 16px; background: var(--surface); }
  .value { font-size: 1.375rem; font-weight: 600; }
</style>
