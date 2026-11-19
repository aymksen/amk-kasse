import { mount } from 'svelte';
import '@fontsource-variable/archivo/wdth.css';
import '@fontsource/ibm-plex-mono/latin-400.css';
import '@fontsource/ibm-plex-mono/latin-600.css';
import '@fontsource-variable/space-grotesk/wght.css';
import './app.css';
import './themes/light.css';
import './themes/pop.css';
import './themes/fonts.css';
import './themes/soft.css';
import './themes/glass.css';
import './themes/glass-effects.css';
import './themes/y2k.css';
import './motion.css';
import App from './App.svelte';

// Keys that make the browser open something over the till that a worker cannot easily close: F1 is Edge's help
// page, F3 its search bar and F7 a question about caret browsing. The till uses none of them.
const BROWSER_KEYS = ['F1', 'F3', 'F7'];
window.addEventListener('keydown', (e) => BROWSER_KEYS.includes(e.key) && e.preventDefault(), true);

mount(App, { target: document.getElementById('app') });
