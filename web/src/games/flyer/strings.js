import { getLanguage } from '../../lib/i18n.svelte.js';

// Every word the flyer shows, [German, English].
const STRINGS = {
  title: ['Kistenflug', 'Crate Flight'],
  tagline: ['Mit dem Bar-Heli zwischen den Kistenstapeln durch.', 'Fly the bar’s chopper between the crate stacks.'],
  best: ['Deine Bestleistung', 'Your best'],
  record: ['Rekord', 'Record'],
  none: ['–', '–'],
  start: ['Tippen oder Leertaste', 'Tap or press Space'],
  startSub: ['zum Abheben', 'to take off'],
  how: ['Jedes Tippen hebt den Heli an. Nicht an die Kisten stoßen!', 'Every tap lifts the chopper. Don’t touch the crates!'],
  ready: ['Bereit?', 'Ready?'],
  back: ['Zurück', 'Back'],
  again: ['Nochmal', 'Again'],
  crashed: ['Bruchlandung!', 'Crash landing!'],
  unit: ['Kisten', 'crates'],
  unitOne: ['Kiste', 'crate'],
  saving: ['Wird gespeichert …', 'Saving …'],
  place: ['Platz {place} auf der Bestenliste', 'Place {place} on the board'],
  notSaved: ['Nicht gespeichert – die Kasse ist gerade nicht erreichbar.', 'Not saved – the till can’t be reached right now.'],
  notSavedPlain: ['Nicht gespeichert.', 'Not saved.'],
  zero: ['Ab einer Kiste kommst du auf die Bestenliste.', 'Pass one crate stack to get on the board.'],
  newRecord: ['Neuer Rekord!', 'New record!'],
  newBest: ['Neue Bestleistung!', 'New personal best!'],
  standing: ['Bestleistung {best} · Rekord {record}', 'Best {best} · Record {record}'],
  againHint: ['Leertaste: nochmal', 'Space: again'],
  paused: ['Pause', 'Paused'],
  resume: ['Tippen zum Weiterfliegen', 'Tap to fly on'],
  toBest: ['Bestleistung {n}', 'Best {n}'],
  toRecord: ['Rekord {n}', 'Record {n}'],
  beatBest: ['Neue Bestleistung!', 'New personal best!'],
  beatRecord: ['Neuer Rekord!', 'New record!'],
};

/** The text for `key` in the till's language, with {name} style blanks filled in. */
export function say(key, vars) {
  let text = STRINGS[key][getLanguage() === 'en' ? 1 : 0];
  if (vars) for (const name in vars) text = text.split(`{${name}}`).join(String(vars[name]));
  return text;
}
