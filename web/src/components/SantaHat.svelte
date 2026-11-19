<!-- Sits on the monkey's head, on the logo without the banana (logo-bare.png). The numbers are shares of the
     logo picture, so it fits the small logo in the dock as well as the big one. The brim stays above the eyes.
     The parent must be positioned. -->
<svg class="hat" viewBox="0 0 176 170" aria-hidden="true">
  <path class="line" d="M28 128C32 82 50 30 104 16C138 8 160 30 164 74C150 60 136 58 128 66C136 86 142 108 146 128Z" fill="#d42a2a" />
  <path d="M104 16C138 8 160 30 164 74C150 60 136 58 128 66C136 86 142 108 146 128H110C112 92 112 50 104 16Z" fill="#a51a22" />
  <path d="M60 58C72 40 86 30 100 26" fill="none" stroke="#f26a5e" stroke-linecap="round" stroke-width="7" />
  <path class="stroke" d="M28 128C32 82 50 30 104 16C138 8 160 30 164 74C150 60 136 58 128 66C136 86 142 108 146 128Z" fill="none" />
  <path class="fur" d="M12 124C8 134 12 150 28 152C40 158 52 152 62 154C74 160 86 152 96 155C108 160 120 152 130 154C146 156 158 148 160 136C164 124 152 112 140 114C128 108 116 116 106 113C94 108 82 116 72 113C60 108 48 116 38 113C24 110 14 114 12 124Z" />
  <path d="M26 142c12 5 22 2 32 3M74 146c10 3 20 1 30 2M116 145c12 2 22 0 30-5" fill="none" stroke="#c9d6e4" stroke-linecap="round" stroke-width="5" />
  <circle class="fur" cx="162" cy="82" r="13" />
  <path d="M155 88c5 4 11 3 15-1" fill="none" stroke="#c9d6e4" stroke-linecap="round" stroke-width="4" />
</svg>

<style>
  .hat {
    position: absolute;
    top: 12.5%;
    left: 42%;
    width: 23.5%;
    height: auto;
    overflow: visible;
    transform: rotate(-6deg);
    pointer-events: none;
  }
  .stroke, .fur { stroke: #17120c; stroke-linejoin: round; stroke-width: 5; }
  .fur { fill: #fff; }
</style>
