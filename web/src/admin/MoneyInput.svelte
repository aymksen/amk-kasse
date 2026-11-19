<script>
  // Typed as euros with a comma, stored as cents.
  let { cents = $bindable(0), id = undefined, ...rest } = $props();

  const show = (value) => (value / 100).toFixed(2).replace('.', ',');
  let text = $state(show(cents));
  let lastCents = cents;

  $effect(() => {
    if (cents !== lastCents) {
      lastCents = cents;
      text = show(cents);
    }
  });

  function input() {
    const parsed = Math.round(Number(text.replace(/\s|€/g, '').replace(',', '.')) * 100);
    if (Number.isFinite(parsed)) {
      lastCents = parsed;
      cents = parsed;
    }
  }
</script>

<input {id} class="input num" inputmode="decimal" autocomplete="off" bind:value={text} oninput={input} onblur={() => (text = show(cents))} {...rest} />
