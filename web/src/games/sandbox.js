import { mount } from 'svelte';
import '@fontsource-variable/archivo/wdth.css';
import '@fontsource/ibm-plex-mono/latin-400.css';
import '@fontsource/ibm-plex-mono/latin-600.css';
import '@fontsource-variable/space-grotesk/wght.css';
import '../app.css';
import '../themes/light.css';
import '../themes/pop.css';
import '../themes/fonts.css';
import '../themes/soft.css';
import '../themes/glass.css';
import '../themes/glass-effects.css';
import '../themes/y2k.css';
import '../motion.css';
import Sandbox from './Sandbox.svelte';

// The test page for the games (games.html), with the same styles as the till.
const style = new URLSearchParams(location.search).get('style');
if (style) document.documentElement.dataset.theme = style;

mount(Sandbox, { target: document.getElementById('app') });
