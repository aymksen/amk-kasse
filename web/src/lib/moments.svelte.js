import { firstAdvent, seasonAt } from './seasons.js';

// The things that happen around the till by themselves: the screensaver, the seasons and the recap.
// None of it touches a sale. It can look, but money, stock and sales are never its business.

// Windows' "Animationseffekte" switch turns every animation off, so all of this stays away too.
const calm = matchMedia('(prefers-reduced-motion: reduce)');

export const moments = $state({
  shift: 0, // how far the clock is moved during a preview, in ms
  preview: null, // 'christmas', 'halloween', 'newyear', 'screensaver', or 'date' for a moved clock
  season: null,
  screensaver: false,
  busy: false, // someone is in the middle of something, such as counting, so no screensaver
});

// For trying a date without waiting for it: /?at=2026-12-31T23:59:50, and &saver to go straight to the screensaver.
const address = new URLSearchParams(location.search);
if (address.has('at')) {
  moments.shift = new Date(address.get('at')) - Date.now();
  moments.preview = 'date';
}
if (address.has('saver')) {
  moments.preview = 'screensaver';
  moments.screensaver = true;
}

/** The time, moved along with a preview. Sales always use the server's clock, so they never see this one. */
export const clock = () => new Date(Date.now() + moments.shift);

export const calmNow = () => calm.matches;

/** At these times the monkey wears a hat, on the logo without the banana. */
export const wearsHat = () => moments.season?.kind === 'christmas' || moments.season?.kind === 'halloween';

/** Whether surprises are allowed right now: always, unless Windows has its animation effects off on this PC. */
export const allowed = () => !calm.matches;

export function updateSeason() {
  moments.season = allowed() ? seasonAt(clock()) : null;
}

/** Shows the OBK what they would otherwise only see once a year. Nothing is saved and a reload ends it. */
export function startPreview(kind) {
  moments.preview = kind;
  moments.shift = 0;
  if (kind === 'christmas') {
    const year = new Date().getFullYear();
    const thirdAdvent = firstAdvent(year);
    thirdAdvent.setDate(thirdAdvent.getDate() + 15);
    thirdAdvent.setHours(21, 30);
    moments.shift = thirdAdvent - Date.now();
  } else if (kind === 'halloween') {
    moments.shift = new Date(new Date().getFullYear(), 9, 31, 21, 30) - Date.now();
  } else if (kind === 'newyear') {
    moments.shift = new Date(new Date().getFullYear(), 11, 31, 23, 59, 45) - Date.now();
  } else if (kind === 'screensaver') {
    moments.screensaver = true;
  }
  updateSeason();
}

export function endPreview() {
  moments.preview = null;
  moments.shift = 0;
  moments.screensaver = false;
  updateSeason();
}
