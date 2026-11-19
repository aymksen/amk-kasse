import { get } from './api.js';
import { errorText, setLanguage } from './i18n.svelte.js';

export const till = $state({
  loaded: false,
  shift: null,
  categories: [],
  products: [],
  expectedFloatCents: 0,
  settings: { barName: 'AMK Bar', language: 'de', otherPayment: 'PayPal' },
  demo: false,
  cart: [],
  message: null, // what the last shift left for this one, until it has been read
});

let languagePicked = false;

export async function refresh() {
  const state = await get('/api/state');
  Object.assign(till, state, { loaded: true });
  if (!languagePicked) setLanguage(state.settings.language);
}

/** A worker's own language choice lasts until the program window is reloaded. */
export function pickLanguage(language) {
  languagePicked = true;
  setLanguage(language);
}

// Toasts ----------------------------------------------------------------------

export const toasts = $state([]);
let nextToast = 1;

export function toast(text, kind = 'info', ms = 4000) {
  const id = nextToast++;
  toasts.push({ id, text, kind });
  setTimeout(() => {
    const index = toasts.findIndex((item) => item.id === id);
    if (index >= 0) toasts.splice(index, 1);
  }, ms);
}

export const toastError = (err) => toast(errorText(err), 'bad', 6000);

// Cart ------------------------------------------------------------------------

export function addToCart(product, qty = 1) {
  const line = till.cart.find((l) => l.product.id === product.id);
  if (line) line.qty += qty;
  else till.cart.push({ product, qty });
}

export function changeLine(line, delta) {
  line.qty += delta;
  if (line.qty <= 0) till.cart.splice(till.cart.indexOf(line), 1);
}

export function cartTotal() {
  return till.cart.reduce((sum, l) => sum + l.qty * l.product.priceCents, 0);
}

/** Matches a scanned code to a product. A crate barcode counts as a full crate. */
export function findByCode(code) {
  for (const product of till.products) {
    if (product.barcodes.includes(code)) return { product, qty: 1, crate: false };
    if (product.crateBarcode === code) return { product, qty: product.crateSize, crate: true };
  }
  return null;
}
