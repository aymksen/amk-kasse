// A barcode scanner behaves like a very fast keyboard that ends with Enter.
// Keys are held back for a moment so a scan can be told apart from a person
// typing on the number row, and a scan never lands in a keypad by accident.
const SCAN_GAP_MS = 50;
const MIN_LENGTH = 4;

/**
 * @param {{ onScan: (code: string) => void, onKey?: (key: string) => void }} handlers
 *   onKey receives single characters plus 'Enter' and 'Backspace' typed by hand.
 * @returns {() => void} stops listening
 */
export function listenForScans({ onScan, onKey }) {
  let buffer = '';
  let timer;

  const flush = () => {
    const typed = buffer;
    buffer = '';
    for (const char of typed) onKey?.(char);
  };

  function handle(e) {
    // Real input fields keep working, so a scan can fill the barcode field in the admin area.
    if (e.target instanceof Element && e.target.matches('input, textarea, select')) return;
    if (e.ctrlKey || e.altKey || e.metaKey) return;

    if (e.key === 'Enter') {
      clearTimeout(timer);
      if (buffer.length >= MIN_LENGTH) {
        e.preventDefault(); // otherwise Enter would also press the focused button
        const code = buffer;
        buffer = '';
        onScan(code);
      } else if (onKey) {
        flush();
        e.preventDefault();
        onKey('Enter');
      }
      return;
    }
    if (e.key === 'Backspace') {
      clearTimeout(timer);
      flush();
      onKey?.('Backspace');
      return;
    }
    if (e.key.length !== 1) return;
    buffer += e.key;
    clearTimeout(timer);
    timer = setTimeout(flush, SCAN_GAP_MS);
  }

  window.addEventListener('keydown', handle, true);
  return () => {
    clearTimeout(timer);
    window.removeEventListener('keydown', handle, true);
  };
}
