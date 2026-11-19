import { t } from '../lib/i18n.svelte.js';

export const termName = (term) =>
  term.season === 'summer'
    ? t('recap.summer', { year: term.year })
    : t('recap.winter', { year: `${term.year}/${String((term.year + 1) % 100).padStart(2, '0')}` });

const longDay = new Intl.DateTimeFormat('de-DE', { weekday: 'long', day: 'numeric', month: 'long' });
const longDayEn = new Intl.DateTimeFormat('en-GB', { weekday: 'long', day: 'numeric', month: 'long' });

/** Saturday, 14 November, for a YYYY-MM-DD business day. */
export const nightName = (iso, language) => (language === 'en' ? longDayEn : longDay).format(new Date(iso + 'T12:00:00'));

/** What a period is called in the OBK tabs. */
export function periodName(recap, period) {
  if (period === 'week') return t('recap.week');
  if (period === 'month') return t('recap.month');
  if (period === 'lastMonth') return t('recap.lastMonth');
  return termName(period === 'lastSemester' ? recap.lastTerm : recap.term);
}
