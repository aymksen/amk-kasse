<script>
  import Delete from 'lucide-svelte/icons/delete';

  // onkey receives '0'-'9', '00', 'del' or 'clear'.
  let { onkey, doubleZero = false, compact = false } = $props();
  const keys = ['1', '2', '3', '4', '5', '6', '7', '8', '9'];
</script>

<div class="pad" class:compact>
  {#each keys as key}
    <button class="key keycap" onclick={() => onkey(key)}>{key}</button>
  {/each}
  <button class="key keycap side" onclick={() => onkey('clear')} aria-label="Löschen">C</button>
  <button class="key keycap" onclick={() => onkey('0')}>0</button>
  {#if doubleZero}
    <button class="key keycap" onclick={() => onkey('00')}>00</button>
  {:else}
    <button class="key keycap side" onclick={() => onkey('del')} aria-label="Zurück"><Delete size={24} /></button>
  {/if}
</div>

<style>
  .pad { display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px 8px; padding-bottom: 3px; }
  .key {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 62px;
    font-family: var(--mono);
    font-size: 1.5rem;
    font-weight: 600;
  }
  .compact .key { height: 54px; }
  .side { --face: var(--surface-2); color: var(--muted); }
</style>
