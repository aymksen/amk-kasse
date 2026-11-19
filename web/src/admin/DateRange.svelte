<script>
  import { businessToday } from '../lib/format.js';
  import { t } from '../lib/i18n.svelte.js';

  let { from = $bindable(), to = $bindable(), onchange } = $props();

  const iso = (d) => `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;

  const presets = [
    { label: 'rep.from7', range: () => [businessToday(-6), businessToday()] },
    { label: 'rep.from30', range: () => [businessToday(-29), businessToday()] },
    { label: 'rep.thisMonth', range: () => [businessToday().slice(0, 8) + '01', businessToday()] },
    {
      label: 'rep.lastMonth',
      range: () => {
        const first = new Date(businessToday().slice(0, 8) + '01T12:00:00');
        return [iso(new Date(first.getFullYear(), first.getMonth() - 1, 1)), iso(new Date(first.getFullYear(), first.getMonth(), 0))];
      },
    },
  ];

  function pick(preset) {
    [from, to] = preset.range();
    onchange();
  }
</script>

<div class="row wrap">
  {#each presets as preset}
    <button class="chip" onclick={() => pick(preset)}>{t(preset.label)}</button>
  {/each}
  <label class="row small muted">{t('common.from')} <input class="input date" type="date" bind:value={from} {onchange} /></label>
  <label class="row small muted">{t('common.to')} <input class="input date" type="date" bind:value={to} {onchange} /></label>
</div>

<style>
  .date { width: 150px; }
  label { gap: 8px; }
</style>
