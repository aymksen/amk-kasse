<script>
  import Garland from './Garland.svelte';
  import Snow from './Snow.svelte';
</script>

<!-- Over the till but under dialogs, and nothing here can be tapped: every touch goes through to the keys. -->
<div class="christmas" aria-hidden="true">
  <Snow />
  <Garland />
</div>

<style>
  .christmas {
    position: fixed;
    inset: 0;
    z-index: 15;
    overflow: hidden;
    pointer-events: none;
  }

  /* Room for the garland, so it does not hang in front of the category tabs. */
  :global([data-season='christmas'] .till) { padding-top: 38px; }
  :global([data-season='christmas'] .start) { padding-top: 100px; }
  :global([data-season='christmas'] .start .top) { top: 36px; }

  /* A flake needs an edge to show on a light background. */
  :global(:root[data-theme='light']), :global(:root[data-theme='pop']), :global(:root[data-theme='y2k']), :global(:root[data-mode='light']) {
    --snow-edge: rgb(84 120 170 / 0.55);
  }

  /* A little snow on the display, baked into one picture. */
  :global([data-season='christmas'] .register > .led) { position: relative; }
  :global([data-season='christmas'] .register > .led::before) {
    position: absolute;
    top: -9px;
    right: -3px;
    left: -3px;
    height: 20px;
    background: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 300 20' preserveAspectRatio='none'%3E%3Cpath d='M0 10C8 2 24 2 36 7C44 1 62 0 74 6C90 0 112 2 122 8C136 2 156 1 170 8C184 2 204 2 216 7C230 1 250 1 262 7C274 3 292 3 300 10V13C292 15 288 11 280 13C276 20 264 20 260 13C250 11 242 13 234 13C224 13 216 11 206 13C198 18 188 18 182 13C170 11 160 13 150 13C140 13 132 11 122 13C116 18 106 18 100 13C90 11 80 13 70 13C60 11 50 13 42 13C38 19 26 19 22 13C14 11 8 15 0 13Z' fill='%23fff'/%3E%3C/svg%3E") center / 100% 100% no-repeat;
    content: '';
    filter: drop-shadow(0 1px 1px rgb(40 70 110 / 0.5));
    pointer-events: none;
  }
</style>
