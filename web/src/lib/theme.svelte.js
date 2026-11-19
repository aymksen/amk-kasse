// The look of the program. The choice is kept per PC, in the browser profile next to the data,
// so a laptop on the OBK network keeps its own. index.html reads it too, before the first paint.
const KEY = 'amk-style';
const GLASS_KEY = 'amk-glass';
const BASIC_KEY = 'amk-basic';
const UNLOCKED_KEY = 'amk-unlocked';

// What the style menu lists. Basic is one entry there, and it comes dark or light. A hidden style stays out of the
// menu until someone finds it: five taps on the helicopter in Verleih put the next one on, and in the menu.
// It only lasts for the shift: when no shift is open, the normal look is back and the style is hidden again.
export const themes = [
  { id: 'basic', label: 'style.basic', colors: ['linear-gradient(135deg, #12100e 50%, #ece6d9 50%)', '#ffb224'] },
  { id: 'glass', label: 'style.glass', colors: ['linear-gradient(135deg, #000000, #6633ee)', '#35e0a1'] },
  { id: 'pop', label: 'style.pop', colors: ['#ffd23f', '#000000'] },
  { id: 'y2k', label: 'style.y2k', hidden: true, colors: ['linear-gradient(135deg, #1557ff, #ff5fc4)', '#ffffff'] },
];

export const basicLooks = [
  { id: 'machine', label: 'style.dark' },
  { id: 'light', label: 'style.light' },
];

// Glass comes dark or light, in these colours. Each pair is a finished set in glass.css.
export const glassModes = [
  { id: 'dark', label: 'style.dark', base: '#000000' },
  { id: 'light', label: 'style.light', base: '#ffffff' },
];
export const glassTints = [
  { id: 'violet', label: 'style.tintViolet', color: '#6633ee' },
  { id: 'blue', label: 'style.tintBlue', color: '#1557ff' },
  { id: 'teal', label: 'style.tintTeal', color: '#08a6b8' },
  { id: 'green', label: 'style.tintGreen', color: '#16a34a' },
  { id: 'pink', label: 'style.tintPink', color: '#e0359a' },
  { id: 'red', label: 'style.tintRed', color: '#dc2f2f' },
  { id: 'orange', label: 'style.tintOrange', color: '#ea7a12' },
  { id: 'grey', label: 'style.tintGrey', color: '#8b93a7' },
];

const isBasic = (id) => basicLooks.some((item) => item.id === id);

function savedStyle() {
  try {
    const id = localStorage.getItem(KEY);
    if (isBasic(id)) return id;
    const style = themes.find((item) => item.id === id);
    return style && (!style.hidden || savedUnlocked().includes(id)) ? id : 'machine';
  } catch {
    return 'machine'; // Storage can be blocked. The default look still works.
  }
}

function savedBasic() {
  try {
    return localStorage.getItem(BASIC_KEY) === 'light' ? 'light' : 'machine';
  } catch {
    return 'machine'; // Same as above.
  }
}

function savedUnlocked() {
  try {
    return (localStorage.getItem(UNLOCKED_KEY) ?? '').split(',').filter(Boolean);
  } catch {
    return []; // Same as above.
  }
}

function savedGlass() {
  try {
    const [mode, tint] = (localStorage.getItem(GLASS_KEY) ?? '').split('-');
    if (glassModes.some((item) => item.id === mode) && glassTints.some((item) => item.id === tint)) return { mode, tint };
  } catch {
    // Same as above.
  }
  return { mode: 'dark', tint: 'violet' };
}

// `id` is the look on screen: for Basic that is `machine` or `light`, and `basic` remembers which of the two it was.
export const theme = $state({ id: savedStyle(), basic: savedBasic(), unlocked: savedUnlocked(), ...savedGlass() });

// The entry that is ticked in the style menu.
export const themeEntry = () => (isBasic(theme.id) ? 'basic' : theme.id);

// The menu without the styles still hidden.
export const visibleThemes = () => themes.filter((item) => !item.hidden || theme.unlocked.includes(item.id));

function apply() {
  const root = document.documentElement;
  if (theme.id === 'machine') delete root.dataset.theme;
  else root.dataset.theme = theme.id;
  if (theme.id === 'glass') {
    root.dataset.mode = theme.mode;
    root.dataset.tint = theme.tint;
  } else {
    delete root.dataset.mode;
    delete root.dataset.tint;
  }
}

function keep(key, value) {
  try {
    localStorage.setItem(key, value);
  } catch {
    // Without storage the choice lasts until the window is reloaded.
  }
}

export function pickTheme(id) {
  theme.id = id === 'basic' ? theme.basic : id;
  if (isBasic(theme.id)) {
    theme.basic = theme.id;
    keep(BASIC_KEY, theme.id);
  }
  apply();
  keep(KEY, theme.id);
}

export function pickGlass(mode, tint) {
  theme.mode = mode;
  theme.tint = tint;
  apply();
  keep(GLASS_KEY, `${mode}-${tint}`);
}

apply();
keep(KEY, theme.id);

/** Puts on the next hidden style and adds it to the menu. Once all are found, the taps change to one of them. */
export function revealHiddenStyle() {
  const hidden = themes.filter((item) => item.hidden);
  const next = hidden.find((item) => !theme.unlocked.includes(item.id)) ?? hidden.find((item) => item.id !== themeEntry());
  if (!next) return false;
  if (!theme.unlocked.includes(next.id)) {
    theme.unlocked.push(next.id);
    keep(UNLOCKED_KEY, theme.unlocked.join(','));
  }
  pickTheme(next.id);
  return true;
}

/** No shift is open: a hidden style makes way for the normal look and leaves the menu, until the helicopter finds it again. */
export function leaveHiddenStyle() {
  if (themes.find((item) => item.id === themeEntry())?.hidden) pickTheme('machine');
  theme.unlocked = [];
  keep(UNLOCKED_KEY, '');
}
