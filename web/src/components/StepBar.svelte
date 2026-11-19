<script>
  import ArrowLeft from 'lucide-svelte/icons/arrow-left';
  import Check from 'lucide-svelte/icons/check';
  import { t } from '../lib/i18n.svelte.js';

  // The bar on top of a screen with steps, at the start and at the end of a shift: what this is, the way
  // through it with the step the worker is on, and the way back.
  // steps: [{ id, label }]. index: the step the worker is on, -1 while it is not clear yet.
  let { title, steps, index, backLabel, onback } = $props();
</script>

<header class="bar">
  <div class="heading">
    <strong>{title}</strong>
    {#if index >= 0}<span class="muted small">{t('close.stepOf', { n: index + 1, total: steps.length })}</span>{/if}
  </div>
  <ol class="steps">
    {#each steps as item, i (item.id)}
      <li class:active={i === index} class:done={i < index}>
        <span class="dot num">{#if i < index}<Check size={16} strokeWidth={3} />{:else}{i + 1}{/if}</span>
        <span class="label">{item.label}</span>
      </li>
    {/each}
  </ol>
  <button class="btn" onclick={onback}><ArrowLeft size={18} /> {backLabel}</button>
</header>

<style>
  .bar {
    display: grid;
    grid-template-columns: auto minmax(0, 1fr) auto;
    align-items: center;
    gap: 24px;
    min-height: 68px;
    padding: 8px 12px 8px 16px;
    background: var(--surface);
    box-shadow: inset 0 -1px 0 var(--line);
  }
  .heading { display: flex; flex-direction: column; line-height: 1.2; white-space: nowrap; }
  .heading strong { font-size: 1.25rem; font-weight: 750; }
  .steps { display: flex; align-items: center; justify-content: center; gap: 10px; min-width: 0; margin: 0; padding: 0; color: var(--muted); font-weight: 650; list-style: none; }
  .steps li { display: flex; align-items: center; gap: 8px; white-space: nowrap; }
  .steps li + li::before { content: ''; width: 28px; height: 2px; border-radius: 1px; background: var(--line); }
  .dot { display: grid; flex: none; place-items: center; width: 30px; height: 30px; border: 2px solid currentColor; border-radius: 50%; font-size: 0.875rem; font-weight: 750; }
  .steps .done { color: var(--text); }
  .steps .done .dot { border-color: var(--good); color: var(--good); }
  .steps .active { color: var(--text); }
  .steps .active .dot { border-color: var(--accent); background: var(--accent); color: var(--on-accent); }
  /* Not enough room for every name: the step the worker is on keeps its name. */
  @media (max-width: 1000px) {
    .steps li:not(.active) .label { display: none; }
  }
</style>
