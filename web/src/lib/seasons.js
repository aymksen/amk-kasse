// What time of year it is at the bar. Only dates, nothing on screen, so it can be tested without a browser.

const MINUTE = 60 * 1000;
const HOUR = 60 * MINUTE;
const HALLOWEEN_WEEK = 24; // the day in October it starts

/** First Sunday of Advent: the fourth is the last Sunday before 25 December. */
export function firstAdvent(year) {
  const christmas = new Date(year, 11, 25);
  const back = christmas.getDay() || 7;
  return new Date(year, 11, 25 - back - 21);
}

function advent(date) {
  const year = date.getFullYear();
  const first = firstAdvent(year);
  const christmasEve = new Date(year, 11, 24);
  const sunday = Math.min(4, Math.floor((date - first) / (7 * 24 * HOUR)) + 1);
  const today = new Date(year, date.getMonth(), date.getDate());
  return { sunday, daysLeft: Math.round((christmasEve - today) / (24 * HOUR)) };
}

/**
 * The season at a moment, or null for most of the year.
 *
 * halloween runs for the last week of October and through the night of the 31st, until 06:00 on 1 November.
 * christmas runs from the first Sunday of Advent to 30 December. The Advent line in the receipt stops on the 26th.
 * newyear runs for the whole night, from 06:00 on 31 December to 06:00 on 1 January, like a business day.
 */
export function seasonAt(date) {
  const year = date.getFullYear();
  const month = date.getMonth();
  const day = date.getDate();

  if (month === 11 && day === 31 && date.getHours() >= 6) return newYear(date, year + 1);
  if (month === 0 && day === 1 && date.getHours() < 6) return newYear(date, year);
  if (month === 9 && day >= HALLOWEEN_WEEK) return { kind: 'halloween', daysLeft: 31 - day };
  if (month === 10 && day === 1 && date.getHours() < 6) return { kind: 'halloween', daysLeft: 0 };
  if (month === 11 && day === 31) return { kind: 'christmas', advent: null };
  if (date >= firstAdvent(year) && (month === 10 || (month === 11 && day <= 30))) {
    return { kind: 'christmas', advent: day <= 26 || month === 10 ? advent(date) : null };
  }
  return null;
}

// The show starts at midnight and lasts this long. The till can be used all the way through.
export const SHOW_MS = 75 * 1000;

function newYear(date, year) {
  const left = new Date(year, 0, 1) - date;
  if (left > MINUTE) return { kind: 'newyear', phase: 'eve', year, left };
  if (left > 0) return { kind: 'newyear', phase: 'final', year, left };
  if (-left < SHOW_MS) return { kind: 'newyear', phase: 'show', year, left };
  return { kind: 'newyear', phase: 'after', year, left };
}
