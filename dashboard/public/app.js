// The OBK's page. It only reads: the numbers come from the till, and nothing here can change anything at the bar.
// Everything from the till goes on the page as text, never as HTML.

const REFRESH = 60 * 1000;
const QUIET = 2 * 60 * 60 * 1000; // the till says hello every hour while it runs
const ARROW_KEYS = { ArrowLeft: -1, ArrowRight: 1 };

const WORDS = {
  de: {
    password: 'Passwort', login: 'Anmelden', logout: 'Abmelden', wrongPassword: 'Das Passwort stimmt nicht.',
    locked: 'Zu viele Versuche. Bitte in {time} noch einmal.', offline: 'Keine Verbindung. Bitte noch einmal versuchen.',
    notSetUp: 'Die Seite ist noch nicht fertig eingerichtet.',
    empty: 'Noch keine Zahlen da', emptyHint: 'In der Kasse unter OBK, Einstellungen, Online-Dashboard die Adresse und den Schlüssel eintragen. Danach erscheinen die Zahlen hier.',
    asOf: 'Stand', open: 'Kasse offen', closed: 'Kasse zu',
    quiet: 'Die Kasse war seitdem nicht online. Neue Verkäufe erscheinen, sobald sie wieder läuft.',
    today: 'Heute', sales: '{n} Verkäufe', sale: '1 Verkauf', week: 'Letzte 7 Tage', box: 'In der Kasse', stockValue: 'Warenwert',
    missing: 'Kassendifferenz, 30 Tage', lent: 'Verliehen', refill: 'An der Bar auffüllen',
    message: 'Nachricht aus der letzten Schicht',
    chart: 'Umsatz der letzten 30 Tage', asTable: 'Als Tabelle', asChart: 'Als Diagramm', day: 'Tag', revenue: 'Umsatz',
    thisMonth: 'Dieser Monat', lastMonth: 'Letzter Monat', salesWord: 'Verkäufe', profit: 'Rohgewinn', product: 'Artikel', quantity: 'Menge', nothingSold: 'Noch nichts verkauft.',
    low: 'Bald nachkaufen', lowNone: 'Alles ausreichend da.', left: '{n} übrig', least: 'mindestens {n}',
    shifts: 'Letzte Schichten', shift: 'Schicht', difference: 'Differenz', running: 'läuft', exact: 'stimmt', none: 'Noch keine Schichten.',
  },
  en: {
    password: 'Password', login: 'Log in', logout: 'Log out', wrongPassword: 'That is not the password.',
    locked: 'Too many tries. Please try again in {time}.', offline: 'No connection. Please try again.',
    notSetUp: 'This page is not fully set up yet.',
    empty: 'No numbers yet', emptyHint: 'In the till, enter the address and the key under OBK, Settings, Online dashboard. The numbers then show up here.',
    asOf: 'As of', open: 'Till open', closed: 'Till closed',
    quiet: 'The till has not been online since. New sales show up as soon as it runs again.',
    today: 'Today', sales: '{n} sales', sale: '1 sale', week: 'Last 7 days', box: 'In the cash box', stockValue: 'Stock value',
    missing: 'Cash difference, 30 days', lent: 'Lent out', refill: 'To refill at the bar',
    message: 'Message from the last shift',
    chart: 'Revenue over the last 30 days', asTable: 'Show table', asChart: 'Show chart', day: 'Day', revenue: 'Revenue',
    thisMonth: 'This month', lastMonth: 'Last month', salesWord: 'Sales', profit: 'Gross profit', product: 'Product', quantity: 'Quantity', nothingSold: 'Nothing sold yet.',
    low: 'Buy soon', lowNone: 'Everything is well stocked.', left: '{n} left', least: 'at least {n}',
    shifts: 'Recent shifts', shift: 'Shift', difference: 'Difference', running: 'running', exact: 'exact', none: 'No shifts yet.',
  },
};

let language = navigator.language?.toLowerCase().startsWith('de') !== false ? 'de' : 'en';
let current = null; // the last reply from the server
let asTable = false;
let period = 'month';
let timer = null;
let quietNote = null;

const root = document.getElementById('app');

const say = (key, vars = {}) => WORDS[language][key].replace(/\{(\w+)\}/g, (_, name) => vars[name]);
const locale = () => (language === 'de' ? 'de-DE' : 'en-GB');
const euroFormat = new Intl.NumberFormat('de-DE', { style: 'currency', currency: 'EUR' });
const euro = (cents) => euroFormat.format((cents ?? 0) / 100);
const signed = (cents) => (cents > 0 ? '+' : '') + euro(cents);
const moment = (ms) => new Intl.DateTimeFormat(locale(), { weekday: 'short', day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' }).format(ms);
const dayName = (iso) => new Intl.DateTimeFormat(locale(), { weekday: 'short', day: '2-digit', month: '2-digit' }).format(new Date(`${iso}T12:00:00`));
const monthName = (iso) => new Intl.DateTimeFormat(locale(), { month: 'long', year: 'numeric' }).format(new Date(`${iso}T12:00:00`));
const count = (n) => (n === 1 ? say('sale') : say('sales', { n }));

function fill(node, props, children) {
  for (const [name, value] of Object.entries(props)) {
    if (name.startsWith('on')) node.addEventListener(name.slice(2), value);
    else if (value !== false && value !== null) node.setAttribute(name, value === true ? '' : value);
  }
  for (const child of children.flat()) if (child !== null && child !== false) node.append(child);
  return node;
}
const h = (tag, props = {}, ...children) => fill(document.createElement(tag), props, children);
const svg = (tag, props = {}, ...children) => fill(document.createElementNS('http://www.w3.org/2000/svg', tag), props, children);

async function call(path, options) {
  const res = await fetch(path, options);
  const data = await res.json().catch(() => ({}));
  return { status: res.status, data };
}

/* ---------------------------------------- Login ---------------------------------------- */

function showLogin() {
  stopRefresh();
  document.documentElement.lang = language;
  const problem = h('p', { class: 'problem', role: 'alert' });
  const input = h('input', { id: 'password', type: 'password', autocomplete: 'current-password', required: true });
  const button = h('button', { class: 'primary', type: 'submit' }, say('login'));

  async function submit(event) {
    event.preventDefault();
    button.disabled = true;
    problem.textContent = '';
    try {
      const { status, data } = await call('/api/login', { method: 'POST', body: JSON.stringify({ password: input.value }) });
      if (status === 200) return void (await load());
      if (status === 429) {
        const minutes = Math.ceil(data.wait / 60);
        problem.textContent = say('locked', { time: `${minutes} min` });
      } else {
        problem.textContent = say(status === 503 ? 'notSetUp' : 'wrongPassword');
      }
    } catch {
      problem.textContent = say('offline');
    }
    button.disabled = false;
    input.select();
  }

  root.replaceChildren(
    h('main', { class: 'door' },
      h('form', { class: 'door-card', onsubmit: submit },
        h('img', { src: 'logo-small.png', alt: 'AMK Bar', width: 96, height: 94 }),
        h('label', { for: 'password' }, say('password')),
        input,
        problem,
        button)),
  );
  input.focus();
}

async function logout() {
  await call('/api/logout', { method: 'POST' }).catch(() => {});
  current = null;
  showLogin();
}

/* ---------------------------------------- Chart ---------------------------------------- */

// Rounds up to 1, 2 or 5 times a power of ten, so the axis shows clean numbers.
function niceMax(cents) {
  const magnitude = 10 ** Math.floor(Math.log10(Math.max(cents, 100)));
  return [1, 2, 5, 10].map((m) => m * magnitude).find((v) => v >= cents);
}

function barChart(series) {
  const HEIGHT = 220;
  const LEFT = 48;
  const RIGHT = 8;
  const TOP = 8;
  const BOTTOM = 26;
  const BAR_MAX = 24;
  const RADIUS = 4;
  const baseline = HEIGHT - BOTTOM;
  const max = niceMax(Math.max(...series.map((row) => row.totalCents), 1));
  const y = (cents) => baseline - (Math.max(cents, 0) / max) * (baseline - TOP);

  const tip = h('div', { class: 'tip', hidden: true });
  const plot = h('div', { class: 'plot' });
  let bars = [];
  let slot = 1;
  let barWidth = 2;
  let picked = null;

  const barX = (i) => LEFT + i * slot + (slot - barWidth) / 2;

  function pick(index) {
    if (index === picked) return;
    bars[picked]?.classList.remove('active');
    picked = index;
    tip.hidden = index === null;
    if (index === null) return;
    bars[index]?.classList.add('active');
    const row = series[index];
    tip.replaceChildren(h('span', {}, dayName(row.day)), h('strong', {}, euro(row.totalCents)));
    const middle = barX(index) + barWidth / 2;
    tip.style.left = `${Math.max(Math.min(middle, plot.clientWidth - 60), 60)}px`;
    tip.style.top = `${Math.max(Math.min(y(row.totalCents), baseline - 56), 56)}px`;
  }

  function draw() {
    const width = plot.clientWidth;
    if (!width) return;
    slot = (width - LEFT - RIGHT) / series.length;
    barWidth = Math.max(Math.min(slot - 2, BAR_MAX), 2);

    const art = svg('svg', { width, height: HEIGHT, role: 'img', 'aria-label': say('chart') });
    for (const tick of [0, max / 2, max]) {
      art.append(
        svg('line', { x1: LEFT, x2: width - RIGHT, y1: y(tick), y2: y(tick), class: 'grid' }),
        svg('text', { x: LEFT - 8, y: y(tick) + 4, 'text-anchor': 'end', class: 'axis' }, `${Math.round(tick / 100)} €`),
      );
    }
    bars = series.map((row, i) => {
      let bar = null;
      if (row.totalCents > 0) {
        const x = barX(i);
        const top = y(row.totalCents);
        const r = Math.min(RADIUS, barWidth / 2, baseline - top);
        // Rounded at the data end, square at the baseline.
        bar = svg('path', { class: 'bar', d: `M${x} ${baseline}V${top + r}Q${x} ${top} ${x + r} ${top}H${x + barWidth - r}Q${x + barWidth} ${top} ${x + barWidth} ${top + r}V${baseline}Z` });
        art.append(bar);
      }
      if (i % 7 === (series.length - 1) % 7) {
        art.append(svg('text', { x: barX(i) + barWidth / 2, y: HEIGHT - 8, 'text-anchor': 'middle', class: 'axis' }, `${row.day.slice(8)}.${row.day.slice(5, 7)}.`));
      }
      return bar;
    });

    const follow = (event) => {
      const box = art.getBoundingClientRect();
      const index = Math.floor((event.clientX - box.left - LEFT) / slot);
      pick(index >= 0 && index < series.length ? index : null);
    };
    art.addEventListener('pointermove', follow);
    art.addEventListener('pointerdown', follow);
    art.addEventListener('pointerleave', () => pick(null));

    picked = null;
    tip.hidden = true;
    plot.replaceChildren(art, tip);
  }

  new ResizeObserver(draw).observe(plot);
  plot.tabIndex = 0;
  plot.setAttribute('aria-label', say('chart'));
  plot.addEventListener('keydown', (event) => {
    const step = ARROW_KEYS[event.key];
    if (!step) return;
    event.preventDefault();
    pick(Math.min(Math.max((picked ?? series.length) + step, 0), series.length - 1));
  });
  plot.addEventListener('blur', () => pick(null));
  return plot;
}

function dayTable(series) {
  return h('div', { class: 'scroll' },
    h('table', {},
      h('thead', {}, h('tr', {}, h('th', {}, say('day')), h('th', { class: 'right' }, say('revenue')))),
      h('tbody', {}, [...series].reverse().map((row) => h('tr', {}, h('td', {}, dayName(row.day)), h('td', { class: 'right' }, euro(row.totalCents)))))));
}

/* ---------------------------------------- Sections ---------------------------------------- */

const line = (label, value, tone) => h('div', { class: 'line' }, h('dt', {}, label), h('dd', { class: tone ?? false }, value));

function todaySection(snap) {
  return h('section', { class: 'today' },
    h('p', { class: 'label' }, say('today')),
    h('p', { class: 'hero' }, euro(snap.todayCents)),
    h('p', { class: 'muted' }, count(snap.todaySales)),
    h('dl', {},
      line(say('week'), euro(snap.weekCents)),
      line(say('box'), euro(snap.cashInBoxCents)),
      line(say('stockValue'), euro(snap.stockValueCents)),
      line(say('missing'), signed(snap.missingCents30), snap.missingCents30 < 0 ? 'bad' : null),
      line(say('lent'), String(snap.rentalsOut)),
      line(say('refill'), String(snap.toRefill))));
}

function messageSection(snap) {
  if (!snap.message) return null;
  return h('section', { class: 'message' },
    h('h2', {}, say('message')),
    h('p', { class: 'said' }, snap.message.text),
    h('p', { class: 'muted small' }, moment(snap.message.at)));
}

function chartSection(snap) {
  const toggle = h('button', { class: 'quiet', onclick: () => { asTable = !asTable; render(); } }, say(asTable ? 'asChart' : 'asTable'));
  return h('section', { class: 'chart' },
    h('div', { class: 'heading' }, h('h2', {}, say('chart')), toggle),
    asTable ? dayTable(snap.series) : barChart(snap.series));
}

function monthSection(snap) {
  const shown = period === 'month' ? snap.month : snap.lastMonth;
  const tab = (id, label) => h('button', { class: period === id ? 'on' : false, 'aria-pressed': String(period === id), onclick: () => { period = id; render(); } }, label);
  return h('section', { class: 'month' },
    h('div', { class: 'heading' }, h('h2', {}, monthName(shown.from)),
      h('div', { class: 'tabs' }, tab('month', say('thisMonth')), tab('last', say('lastMonth')))),
    h('dl', {},
      line(say('revenue'), euro(shown.revenueCents)),
      line(say('profit'), euro(shown.revenueCents - shown.costCents)),
      line(say('salesWord'), String(shown.salesCount))),
    shown.products.length
      ? h('table', {},
          h('thead', {}, h('tr', {}, h('th', {}, say('product')), h('th', { class: 'right' }, say('quantity')), h('th', { class: 'right' }, say('revenue')))),
          h('tbody', {}, shown.products.map((p) => h('tr', {}, h('td', {}, p.name), h('td', { class: 'right' }, String(p.qty)), h('td', { class: 'right' }, euro(p.revenueCents))))))
      : h('p', { class: 'muted' }, say('nothingSold')));
}

function lowSection(snap) {
  return h('section', { class: 'low' },
    h('h2', {}, say('low')),
    snap.lowStock.length
      ? h('ul', {}, snap.lowStock.map((item) => h('li', {}, h('span', {}, item.name), h('span', { class: 'muted' }, `${say('left', { n: item.stock })}, ${say('least', { n: item.minStock })}`))))
      : h('p', { class: 'muted' }, say('lowNone')));
}

function differenceCell(shift) {
  if (shift.closedAt === null) return h('td', { class: 'right muted' }, say('running'));
  if (shift.differenceCents === null) return h('td', { class: 'right muted' }, '–');
  if (shift.differenceCents === 0) return h('td', { class: 'right muted' }, say('exact'));
  return h('td', { class: `right ${shift.differenceCents < 0 ? 'bad' : ''}` }, signed(shift.differenceCents));
}

function shiftsSection(snap) {
  if (!snap.shifts.length) return h('section', { class: 'shifts' }, h('h2', {}, say('shifts')), h('p', { class: 'muted' }, say('none')));
  return h('section', { class: 'shifts' },
    h('h2', {}, say('shifts')),
    h('div', { class: 'wide' },
      h('table', {},
        h('thead', {}, h('tr', {}, h('th', {}, say('shift')), h('th', { class: 'right' }, say('revenue')), h('th', { class: 'right' }, say('difference')))),
        h('tbody', {}, snap.shifts.map((shift) => h('tr', {},
          h('td', {}, moment(shift.openedAt), h('span', { class: 'muted small under' }, [count(shift.salesCount), shift.note].filter(Boolean).join(' · '))),
          h('td', { class: 'right' }, euro(shift.totalCents)),
          differenceCell(shift)))))));
}

/* ---------------------------------------- Page ---------------------------------------- */

function render() {
  const { receivedAt, snapshot: snap } = current;
  if (!snap) {
    root.replaceChildren(h('main', { class: 'door' }, h('div', { class: 'door-card' },
      h('img', { src: 'logo-small.png', alt: 'AMK Bar', width: 96, height: 94 }),
      h('h1', {}, say('empty')),
      h('p', { class: 'muted' }, say('emptyHint')),
      h('button', { onclick: logout }, say('logout')))));
    return;
  }

  language = snap.language === 'en' ? 'en' : 'de';
  document.documentElement.lang = language;
  document.title = snap.barName || 'AMK Bar';
  quietNote = h('p', { class: 'quiet-note', hidden: Date.now() - receivedAt < QUIET }, say('quiet'));

  root.replaceChildren(
    h('header', {},
      h('img', { src: 'logo-small.png', alt: '', width: 40, height: 39 }),
      h('div', { class: 'title' },
        h('h1', {}, snap.barName || 'AMK Bar'),
        h('p', { class: 'muted small' }, `${say('asOf')}: ${moment(receivedAt)} · ${say(snap.open ? 'open' : 'closed')}`)),
      h('button', { class: 'quiet', onclick: logout }, say('logout'))),
    quietNote,
    h('main', { class: 'page' },
      todaySection(snap),
      messageSection(snap),
      chartSection(snap),
      monthSection(snap),
      lowSection(snap),
      shiftsSection(snap)),
  );
}

function stopRefresh() {
  clearInterval(timer);
  timer = null;
}

async function load() {
  let reply;
  try {
    reply = await call('/api/data');
  } catch {
    if (!current) root.replaceChildren(h('p', { class: 'plain' }, say('offline')));
    return; // the numbers on screen stay, and the next round tries again
  }
  if (reply.status === 401) return showLogin();
  if (reply.status !== 200) return root.replaceChildren(h('p', { class: 'plain' }, say(reply.status === 503 ? 'notSetUp' : 'offline')));

  // Only redrawn when something came in, so a button in use keeps its place.
  const fresh = reply.data.receivedAt !== current?.receivedAt;
  current = reply.data;
  if (fresh) render();
  if (quietNote) quietNote.hidden = Date.now() - current.receivedAt < QUIET;
  timer ??= setInterval(() => { if (!document.hidden) load(); }, REFRESH);
}

document.addEventListener('visibilitychange', () => { if (!document.hidden && timer) load(); });
load();
