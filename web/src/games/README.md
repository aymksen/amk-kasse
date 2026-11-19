# Spiele (mini-games)

Seven quick taps on the AMK Bar logo in the till's dock while a shift is open open the games window
(`Arcade.svelte`). The window grows out of the logo and lies over the dimmed till. It shows the game's picture,
name and blurb, the player's best and place, the record, a big Play button, and the leaderboard: place, name,
room, date and score. Play starts the game inside the same window, under its title bar. The X in the title
bar's top-right corner (or Escape) leaves the game for the leaderboard, and on the leaderboard closes the window.
The player is always whoever is on shift. The server takes room and name from the open shift.

There is one game for now, Kistenflug (`flyer`), and the window shows only the first game in `index.js`. A
second game needs a way to pick it in the window, its id in `index.js` and in `GameIDs` in
`server/internal/store/games.go`. A game's id never changes once scores exist.

Each game lives in its own folder and touches nothing outside it:

```
src/games/
  index.js           the list of games (id, strings, picture, lazy import)
  Arcade.svelte      the games window, the leaderboard and the move into a game (not a game's business)
  Sandbox.svelte     the test page, games.html
  flyer/             Kistenflug: the bar's helicopter between crate stacks (tap to rise)
```

## What a game folder holds

- `meta.js`: `export default { title: [de, en], blurb: [de, en], unit: [de, en] }`.
  - `title` is the game's name on its card.
  - `blurb` is one short line under it.
  - `unit` is what the score counts, shown in the board's score column header ("Kisten", "Punkte" …).
- `Game.svelte`: the game itself, see the contract below.
- The game's picture in the window is a still image set as `picture` in `index.js` (Kistenflug uses the
  helicopter sticker from `src/assets`). The window shows no moving card, so a game needs no `Cover.svelte`.
- Anything else the game needs: more components, `strings.js`, art the game draws or that the session made itself.

## Game.svelte contract

```js
let {
  player,   // { name, room } of the worker on shift, for the game's own screens
  best,     // this player's best score in this game so far, 0 if none. Kept current by the arcade.
  record,   // the top score on this game's board, 0 if none. Kept current by the arcade.
  onfinish, // (score) => Promise<Result>. Call it once when a run ends.
  onquit,   // () => void. Back to the games screen.
} = $props();

// Result: { saved, newBest, best, place, record }
//   saved   false for a score of 0 (not stored) or when saving failed
//   newBest this run beat the player's own best
//   best    the player's best after this run
//   place   the player's place on the board after this run, 1-based. It can be above 10. null when not saved.
//   record  the top score on the board after this run
// The promise rejects when the till cannot be reached. Show a calm "nicht gespeichert" line and carry on.
```

The game owns its own flow: a start screen (tap or key to start), the run, and a result screen that shows the score,
then the place once `onfinish` answers. The result screen offers "Nochmal" (play again) and "Zurück" (calls `onquit`).
The arcade puts the game in the window's body, under a 60 px title bar whose top-right corner holds the X, and
grows it out of the Play button, so the game's start screen should fade or slide in on mount rather than pop. The
title bar already shows who is playing and a way out, so the game does not need its own.

## Rules every game keeps

1. **Fills its box.** The parent is the window's body. Use 100% width and height, follow resizes (`ResizeObserver`),
   and draw sharp on `devicePixelRatio`. The bar's screen may stand upright (portrait, about 1080×1920) or lie
   (1920×1080, also 1366×768). The box is then about 1030×1640 upright, 1560×880 at full HD and 1330×680 at
   1366×768. Both shapes must play well, not just fit.
2. **Input.** Mouse click and touch (`pointerdown`) and the keyboard: Space, arrow keys, WASD. Call `preventDefault`
   on the keys the game uses, so a focused button underneath never gets them.
   - **Never use Enter or digits.** The barcode scanner types digits and Enter, and a scan closes the games screen and goes into the sale.
   - Escape belongs to the arcade (it leaves the game for the leaderboard).
3. **Same speed on every screen.** Use a fixed time step or a clamped `dt`, so 60 Hz and 144 Hz play the same.
   After a pause or a hidden tab, do not jump ahead.
4. **Pause when the window is hidden** (`visibilitychange`). Stop every `requestAnimationFrame`, timer and listener when the component is destroyed.
5. **No sound at all.** The OBK dislikes sound on the till.
6. **No network and no downloads.**
   - Art is drawn in code (canvas or SVG), or is a file the session made itself in the game's folder.
   - The shared pictures in `src/assets` (logo, stickers, tin sign) may be imported, read-only.
   - No product photos.
   - No new npm packages.
   - Fonts: only the ones the app already loads (see `src/main.js`) or system fonts.
7. **Original.** Bar-themed, our own idea. No names, characters, colours or look taken from an existing game (no green pipes, no little bird, no Nokia-style screen). The genre mechanics are fine.
8. **Strings** in German first, with English, picked with `getLanguage()` from `../../lib/i18n.svelte.js`.
9. **Calm mode:** when `calmNow()` from `../../lib/moments.svelte.js` is true, leave out screen shake and flashes.
10. **Old Chrome.** The till runs Chrome 109. This rules out the View Transitions API and CSS nesting. Keep 60 fps on a Ryzen 5 with integrated graphics. Pre-render sprites to offscreen canvases. Avoid per-frame `shadowBlur` and `filter`.

## Trying a game

Each game has its own dev server, started with `preview_start` by name from `.claude/launch.json`: `games-flyer` on
port 5181 (`games-arcade` on 5183 is the main session's). By hand, from `web/`:
`npx vite --port 5181 --strictPort`. Each port keeps its own cache (`node_modules/.vite-<port>`).

- `http://localhost:5181/games.html?game=flyer` opens the game alone, with a test player and a pretend leaderboard
  kept in the browser (`&fail` makes saving fail, `&best=12&record=40` starts with those numbers).
- `http://localhost:5181/games.html` shows the real games window with the pretend leaderboard, growing out of
  where the logo sits in the till.

Never run `npm run build` from a game session: it empties the shared build folder the till is made from.
