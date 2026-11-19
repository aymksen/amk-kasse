const euroFormat = new Intl.NumberFormat('de-DE', { style: 'currency', currency: 'EUR' });
const timeFormat = new Intl.DateTimeFormat('de-DE', { hour: '2-digit', minute: '2-digit' });
const dateTimeFormat = new Intl.DateTimeFormat('de-DE', { day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit' });
const dayFormat = new Intl.DateTimeFormat('de-DE', { day: '2-digit', month: '2-digit', year: 'numeric' });

export const euro = (cents) => euroFormat.format((cents ?? 0) / 100);
export const time = (ms) => timeFormat.format(ms);
export const dateTime = (ms) => dateTimeFormat.format(ms);

/** Formats a YYYY-MM-DD business day. */
export const day = (iso) => dayFormat.format(new Date(iso + 'T12:00:00'));

/** Today's date as YYYY-MM-DD, shifted so 02:00 still counts as the night before. */
export function businessToday(offsetDays = 0) {
  const d = new Date(Date.now() - 6 * 3600 * 1000 + offsetDays * 86400 * 1000);
  const pad = (n) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
}

/** Applies a keypad key to a digit string. Used for counts and cent amounts. */
export function press(value, key, maxLength = 6) {
  if (key === 'del') return value.slice(0, -1);
  if (key === 'clear') return '';
  if (value.length >= maxLength) return value;
  return (value + key).replace(/^0+(?=\d)/, '');
}
