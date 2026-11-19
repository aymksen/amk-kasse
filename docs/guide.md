# AMK Kasse: the full guide

The till for the bar at the Albertus Magnus Wohnheim in Münster. It sells drinks and snacks, lends cues and darts for a deposit, counts stock, and gives the OBK a place to restock and check the money.

It runs offline on one PC. There is nothing to install: one program, one data folder.

## Set it up on the bar PC

1. Copy the `dist` folder to a normal folder, for example `C:\AMK-Kasse`. Not `C:\Program Files`, the program has to write next to itself.
2. Double-click `AMK-Kasse.exe`. The till opens full screen in Edge.
3. To start it with Windows: press `Win + R`, type `shell:startup`, press Enter, and put a shortcut to `AMK-Kasse.exe` in that folder.
4. Click **OBK** and choose a PIN. Only the bar PC can set the first PIN.
5. Under **Artikel**, check the products. The range, the prices, the spots and the stock figures come from the list in the old till. All of the stock is booked at the Bar and none in the Lager, so count the Lager once. The purchase prices are still 0: fill them in for the reports. Pfand Bierflasche (−0,08 €) and Pfand Flasche (−0,15 €) pay the deposit back when empties come in; Pfand Kiste charges 1,50 €. The lending in Verleih also follows the old list: one cue, three sets of darts and two airhockey sets, each for a deposit and no fee.

Use `AMK-Kasse-32bit.exe` instead if the PC runs 32-bit Windows. Windows 10 or newer is required.

To look around first, double-click `Demo starten.cmd`. It opens a window with invented sales and keeps its data apart from the real till. Everything in it is reset the next time it starts.

Press `Alt + F4` to close the full-screen window. The till keeps running in the background, and a double-click on `AMK-Kasse.exe` brings the window back. Edge never offers to translate the till, and F1, F3 and F7 do nothing, so its help page, search bar and caret-browsing question cannot pop up over a sale.

## Barcode scanner

We recommend the Inateck BCST-33: a wired USB laser scanner with a stand, around 30 €. Plug it in and it works like a keyboard, with no driver. It must end each scan with Enter, which is the usual default.

It reads the barcodes on bottles and crates, but not QR codes or phone screens.

- **Teach a product:** OBK → Artikel → open the product and scan. A product can have several codes: Chips is one key, and every kind of chips has its own code, so scan one bag after another and each lands in the list under Barcodes. A code typed by hand goes in with Enter or Hinzufügen, and the cross on a code takes it out. The crate has its own field, so a scanned crate counts as a full crate. A code belongs to one product only, and never to a bottle and a crate at once.
- **Verleih stickers:** print a Code 128 barcode with the digits from OBK → Verleih, for example `90001`. Without a sticker, just tap the item on the screen.

## Product pictures

OBK → Artikel → open the product. Choose a picture, drag one onto the white square, or copy a picture from a website and press `Ctrl + V`. A white background looks best. The picture is shrunk to about 10 KB before it is saved, so the size of the original does not matter, and it is part of every backup.

The picture also shows small beside the product's line on the receipt, and on the rows of the stock count. A product without a picture simply shows its name on the key, and an empty white square on the receipt.

## Styles

The palette button next to the language button picks the look. It is on the start screen and in the bar at the bottom of the till.

- **Basic:** the normal look. It comes as **Dunkel** or **Hell**, a switch that appears under the list when Basic is selected. Hell is the same till in a light casing, with a calculator display.
- **Pop:** flat bright colour with thick black outlines.
- **Glass:** frosted panels over a background that fades from black or white into a colour, with glowing buttons.
- **Y2K:** the computers and consoles of around 2004. Chrome bars, an icy pinstriped wallpaper, flat white product cards with a coloured price pill, a backlit blue LCD for the total, fat gel scrollbars and one glossy blue main button. This one is hidden: nobody is told about it. Five quick taps on the helicopter in Verleih switch to it straight away and put it in the menu. It lasts until the shift ends. Then the till goes back to Basic Dunkel and Y2K is hidden again, until someone finds the helicopter.

A style only changes colours and finish. Every button stays in the same place, so nobody has to learn anything new. In the OBK area the same two buttons are at the bottom of the side bar. The OBK area follows the same choice, and it is saved on that PC.

The till also works on a screen turned upright. The sales screen keeps its layout, with the total and the receipt on the right, and the product keys get taller to use the height. The counts are one long list either way, so they simply get more rows. This follows the shape of the window, so there is nothing to switch on. Below 900 pixels of width the bottom bar takes two rows.

The glass in the Glass style is painted, not a live blur. Real blur has to recalculate everything behind a panel, which a weak PC cannot afford. The backgrounds are already soft, so it looks nearly the same and runs about as fast as Basic.

Glass adds a few finishing touches: a plasma picture on the start screen, a sheen on the keys, a glow on the main button of each screen, and a receipt that looks like a small code editor.

With Glass selected, the style menu also shows **Farben**: 16 circles with two colours each. The black or white half picks a dark or light look, and the coloured half picks the background, the tint of the panels and the buttons. There are eight colours: violet, blue, teal, green, pink, red, orange and grey. The choice is saved on that PC.

## Starting a shift

Workers type their room number and their name. If the room has been used before, the till offers the last names used for it, under the fields. Tap one, or ignore it and type: rooms change hands, so nothing is filled in by itself. Lending something in Verleih offers the names the same way.

Then comes the step **Bestand**, "Stimmt der Bestand an der Bar?": every product the bar counts, in the order of the spots behind the bar, with what should be there already filled in. That is what the last worker left at the end of their shift, plus anything the OBK carried over from the Lager since. The top says who closed last and when, or that they did not count. The worker walks past each spot and only changes what is wrong:

- **−** and **+** next to the figure, or the arrow keys: up and down move through the list, left and right change the figure.
- Tapping the figure opens a keypad, for when more than a bottle or two is off.
- Beside each figure is where it came from: **Übergabe 9** is what the last shift left. A changed figure also shows what it was and the difference, like "war 12, −2", and its row gets a light tint so the changes stand out. Tapping "war 12" puts it back.
- A scanned bottle jumps to its row.

When everything is right, the button just says **Alles stimmt**. Otherwise it says how many figures changed. The check is saved as a count named Schichtbeginn and corrects the bar figures, so a gap shows up in the history at the shift where it was found, not hours later.

The same screen is the only way the till counts stock: here, from **Zählen** in the dock at any time of the shift, and at its end. During and at the end of a shift every row reads like "Beginn 12 → 9", both on the display font, the earlier one dimmer: what the bar had when the shift started, and what it should hold now, with what was sold taken off and what the OBK filled in added. Under the name it says how many were sold, filled in, or corrected by an earlier Zählen; those rows get a fainter tint than a changed one, so what moved during the shift is easy to find. Zählen saves with **Alles stimmt** or "2 Änderungen speichern" and goes back to selling. Only the money is counted blind.

## Ending a shift

**Schicht beenden** walks through the steps. A bar on top shows them all, ticks off the ones done and keeps a way back to selling; the dock steps aside meanwhile. The button that goes on sits at the bottom right of every step.

- **Noch verliehen.** If something is still out, this comes first: what it is, who has it, their room, since when, and the deposit. Anything that is back can be taken back right there, before the cash is counted, because a deposit handed back afterwards would make the count wrong. Nothing has to be back. Items can stay out overnight.
- **Bestand** is the check from the start of a shift: every product with what should be there now, and **Alles stimmt** when it is right. It is saved together with the end of the shift, as a count named Schichtende.
- **Geld zählen** has two ways. The quick one is picked from the start: the keypad beside the total (under the list on an upright screen) types the total and nothing else, the last two digits being the cents. The careful one is note by note: tap a row, type how many on the keyboard, and Enter or Tab goes on to the next, 50 € to 20 € to 10 € and down to the cents; Shift+Tab goes back, and after the last coin comes Weiter. − and + work on every row. Whatever was touched last counts: a tap on the keypad goes back to the total, starting a new amount. Until something is counted, the button reads **Kasse ist leer** instead of Weiter, so nobody skips the count by mistake. Nothing on this step says what should be in the box.
- **Ergebnis** shows whether the box is right, then **Schicht beenden** closes the shift. When the count is right to the cent and there was at least one sale, the slip gets a red PERFEKT stamp. Below it is a field for a short message to the next shift, like "Kühlschrank tropft", up to 280 characters.

The next worker sees that message once, right after starting. The OBK sees it at the top of the Übersicht and under Schichten.

## Bestenliste

The cup button opens the leaderboard: who sold the most this month or last month, with their room numbers, the top drinks and snacks, the busiest night and the latest order. It is on the start screen, so it works before anyone has started a shift, in the bottom bar of the till, and under OBK → Rückblick. Any key or scan closes it, and a scan still lands in the cart.

It shows names, rooms and pieces, never money or stock. On the bar PC anyone can open it. On any other device it needs the OBK PIN.

## Spiele

The games are hidden. During a shift, seven quick taps on the AMK Bar logo bottom-left open them; from the third tap the logo gives a small wobble, and a single tap shows nothing. They open in a window over the till that grows out of the logo. For now there is one game: Kistenflug, the bar's helicopter flying between stacks of beer crates, one tap to rise.

The window shows the leaderboard (place, name, room, date and score in crates passed), your own best and place, the record, and a big Spielen button. Spielen starts the game inside the same window. A run counts for whoever is on shift, and each worker shows once, with their best run. The top ten are shown, and your own place even below that.

The cross in the top-right corner of the window (or Escape) leaves the game for the leaderboard, and on the leaderboard closes the window. A tap beside the window closes it too, but not while a game runs. A scan closes the games and lands in the cart as usual. After four minutes without a touch the till comes back by itself. The games have no sound.

The scores stay on the till (schema version 7, table `game_scores`). They are not money, stock or sales, and they never go to the online dashboard.

## Stock: Lager, Bar and Gesamt

Every product has two figures, and OBK → Bestand shows both with their total.

- **Bar:** everything at the bar, including the small stock and the crates there. Sales take from here, and the keys only warn about this figure.
- **Lager:** the storeroom, which only the OBK uses. Workers never see or count it. Deliveries in OBK → Wareneingang land here, or straight at the bar if you pick that.
- **Auffüllen:** when the OBK carries goods from the Lager to the bar, they book it under OBK → Bestand. The total stays the same. The Übersicht says how many products need it.
- **Zählen:** workers check the bar at the till, at the start and end of every shift and from the dock whenever they like, with crates and single bottles together as one figure per product. The OBK counts the Lager with the Lager zählen button under Bestand.

Two numbers on each product drive the suggestions. **Soll an der Bar** is how many belong there once it is refilled, small stock and crates included; once the bar is down to half of it, the product shows under Auffüllen with the amount to fetch. **Mindestbestand** is for the bar and Lager together; at or below it, the product shows under Nachkaufen.

The Lager figure can go below zero if more was fetched than the system expected. The goods were clearly there, so nothing is refused: it shows in red until the next Lager count corrects it.

After updating from a version without a Lager, each product starts with up to one crate at the bar and the rest in the Lager. That split is only a guess, so count both places once.

The update to the real range (version 6) replaces the example range of the first versions. Products the list names are kept with their sales, also under an earlier name (`Hansa Pils 0,5 l` becomes `Hansa Pils`), and take the list's name, price, spot and figure. The change in stock is booked as a count, so it shows in the history. Every other product was hidden at first. Lending items the list does not have are hidden the same way, but one that is still lent out stays on the board until it comes back.

The update to version 9 then deletes the hidden products for good, so only the range with pictures is left. Their old sales stay in the reports with the name and price they were sold at; only the link to the product goes. Their stock history goes with them, and so do hidden lending items that were never lent. Barcodes move to their own table in the same update, since a product can now have several, and the till gets the check at the start of a shift. A product hidden later with the aktiv box stays hidden, not deleted.

## Online dashboard

The OBK can see the numbers from home on a small web page of your own, on a phone or a laptop, with a password. The till sends the page a summary whenever something changed, at most every five minutes, and once an hour otherwise. It only sends and never receives, so nothing on the internet can change anything at the bar, and without internet the till simply carries on. The page then shows the time of the last update.

Setting it up takes one free Cloudflare account and ten minutes on any PC with Node.js. The steps are in [dashboard/README.md](../dashboard/README.md).

What goes online: revenue, stock figures, quantities sold, cash differences, the handover messages and the shift notes, as typed. What stays on the bar PC: the names and rooms of workers and renters. The list is `OnlineSnapshot` in `server/internal/store/online.go`, field by field, so nothing new goes online by accident.

## Backups

The till saves a backup after every shift and once a day, in `data\backups`. It keeps the latest 30.

A USB stick protects against a dead PC. Plug one in and enter a folder on it under OBK → Einstellungen → Zweiter Speicherort.

To restore a backup:

1. End `AMK-Kasse` in Task Manager.
2. Copy the backup to `data\amk-bar.db`, replacing the file there.
3. Delete `amk-bar.db-wal` and `amk-bar.db-shm` if they exist.
4. Start the program again.

## Forgot the PIN

Double-click `PIN zuruecksetzen.cmd`. The next visit to the OBK area asks for a new PIN. Sales and stock are not touched.

## Reaching the OBK area from a laptop

Start the program with `-lan`, then open `http://<ip-of-the-bar-pc>:8417/admin` on a device in the same network. Every device outside the bar PC needs the PIN.

The connection is not encrypted, so only do this on a network you trust. Without `-lan`, nothing outside the bar PC can reach the till.

## Program options

| Option | Effect |
|---|---|
| `-lan` | Allow other devices on the network, PIN required |
| `-port 8417` | Change the port |
| `-data <folder>` | Keep data in another folder |
| `-windowed` | Open a normal window instead of full screen |
| `-demo` | Start with invented sales in `data-demo` |
| `-no-browser` | Start without opening a window |
| `-reset-pin` | Clear the OBK PIN |

Problems are written to `data\amk-kasse.log`.

## For developers

The server is Go with SQLite (`server/`, business rules in `internal/store`). The interface is Svelte (`web/`), built into the Go program. Design notes are in `docs/specs/`.

```bash
build.cmd
```

This needs Node.js and Go. It runs the tests and writes both programs to `dist`. Use a portable Go in `.tools\go`, or one from your PATH.

For interface work, run `AMK-Kasse.exe -demo -no-browser`, then `npm run dev` in `web`.

Each game lives in its own folder under `web/src/games/`. `web/src/games/README.md` holds the rules a game keeps and how to try one alone on `games.html`, with a pretend leaderboard and no till.

The dev server sends `/api` to port 8417. If the real till already holds that port, start the demo with `-port 8418` and set `AMK_API=http://127.0.0.1:8418` before `npm run dev`.

A style is a file in `web/src/themes/` imported in `web/src/main.js`, plus a line in `web/src/lib/theme.svelte.js` and a label in `web/src/lib/i18n.svelte.js`. `hidden: true` on that line keeps it out of the menu until the helicopter unlocks it, next in line after the ones before it. The start screen hides it again, so an unlock lasts for one shift. If its keys need their own colours, add a branch in `web/src/lib/keycolor.js`. `app.css` holds the tokens with their default values, and a style only overrides them. Keep it to colour and finish, so the layout stays the same in every style.

Each style also sets its fonts as tokens: `--sans` for text, `--display` and `--display-weight` for headings, the receipt head, the labels on the display and the stamps, `--narrow` for a product name too long for its tile, and `--mono` for money and the receipt. Basic uses Barlow Condensed, Barlow Semi Condensed and JetBrains Mono, Glass uses Manrope, Pop uses Space Grotesk, and Y2K uses Unbounded for headings and Lexend for the rest. They come from `@fontsource` packages and are declared in `web/src/themes/fonts.css`, latin only, so they are built into the program and nothing loads from the internet. All are under the SIL Open Font License. The Easter eggs keep Archivo and IBM Plex Mono, so those two stay installed.

Glass and Y2K share `web/src/themes/soft.css`. Do not use `backdrop-filter` or soft shadows on the product keys, because they are what slows a weak PC down. Class names repeat across screens, so name the element in a selector (`button.tile`, `section.right`).

The finish of Glass is in `web/src/themes/glass-effects.css`, rebuilt from six 21st.dev pieces. Its pictures are in `web/src/themes/img/`. Only `transform` and `opacity` may animate there, and the plasma stays a picture because a running shader would keep the PC busy for as long as the start screen is open. The plasma pictures are grey and take the chosen colour through `background-blend-mode`.

The logo and stickers are in `web/src/assets/`. A sticker sits in empty space only, never on a product key, and it has its shadow baked into the picture. It disappears on short screens.

The animations of the till are all in `web/src/motion.css`. Each one plays once and stops, and only `transform` and `opacity` move, so a weak PC has nothing to do once the movement ends. Do not add anything that loops there. If the bar PC feels slow, turning off "Animationseffekte" in Windows turns them all off, the seasons and the screensaver included.

What happens around the till by itself is in `web/src/moments/`, with the rules in `web/src/lib/moments.svelte.js` and the dates in `web/src/lib/seasons.js`. These may loop, because they run for a few weeks a year or while nobody uses the till. None of it can be tapped or touches a sale.

- **Screensaver.** After five minutes with an empty cart, and never during Zählen or Kassenschluss, the logo bounces around the screen, printed in one colour that changes at every wall. A corner is planned, not luck (`web/src/moments/bounce.js`): the logo bends a course that would end in one into a near miss, and aims for one once its time has come, every 3 to 10 minutes. Only then does it show its real colours, with confetti. A mouse knocked by less than 40 px does not end it. The first tap only brings the till back, so nobody sells a beer by accident. A scan goes straight through and adds the item. Every 90 seconds the Bestenliste of the month shows for 18 seconds.
- **Christmas.** From the first Advent to 30 December: a garland with lights, snow that builds up on the bottom bar, a Santa hat on the monkey, and a short Advent line in the empty receipt until 24 December. The hat sits on `logo-bare.png`, a copy of the logo with the banana painted out, so a change to the logo has to be made there too.
- **Halloween.** From 24 October until 06:00 on 1 November: lit pumpkins on a rope along the top, cobwebs in the corners, bats, fog, a spider that lets itself down every half minute, a witch's hat on the monkey (the same `logo-bare.png`), a count of the days in the empty receipt, and on the screensaver an orange moon with the monkey crossing on a broom. It is built like Christmas, from `web/src/moments/Halloween.svelte`.
- **New Year.** On 31 December the empty receipt counts down from 06:00. The last minute shows on a panel at the top, the last ten seconds fill the screen, and at midnight 75 seconds of fireworks and confetti run over the till. Taps and scans keep working under all of it.
- **Recap.** OBK → Rückblick shows the week, this month, last month and the semesters: most sold, top drink, who sold the most, the busiest night and the latest order. The Bestenliste is the same numbers for this and last month. Both show names, rooms and pieces, never money. The week is the last seven days, a semester runs April to September or October to March, and Pfand and Waschmarken do not count. A "drink" is everything except a category called Snacks.

To see these without waiting: OBK → Einstellungen → Vorschau. `/?at=2026-12-31T23:59:50` moves the clock of the moments for that page, and `&saver` goes straight to the screensaver. Sales always use the clock of the server, so neither reaches the books. Fireworks are [fireworks-js](https://github.com/crashmax-dev/fireworks-js) (MIT) and confetti is [canvas-confetti](https://github.com/catdad/canvas-confetti) (ISC). Both are built into the program, so all of this works offline.

The first Easter egg is in `web/src/till/SixSeven.svelte`: when a total lands on exactly 6,70 €, the bar's monkey does the "six seven" hand gesture for two seconds, without sound. It is drawn as SVG, every tap goes through it, and it ends as soon as the total changes or someone presses Bezahlen. 5× Hansa Pils plus 1× Coca-Cola gets you there. There is no switch to turn the surprises off: they belong to the bar. Only "Animationseffekte" turned off in Windows stops them, together with the seasons and the screensaver, for a PC that is too slow. Easter eggs may never change money, stock or sales.

The second is its evil twin in `web/src/till/Devil.svelte`: at 6,66 €, 66,60 € or 666,00 € the monkey comes up out of the fire with horns and a trident, and three sixes land around and on him. 3× Coca-Cola and 1× Haribo with 3× Pfand Bierflasche taken back makes 6,66 €, and 37× Haribo gets you to 66,60 €.

The third is `web/src/till/FourTwenty.svelte`: at 4,20 €, 42,00 € or 420,00 € the display turns green and the monkey grooves in a beanie with a joint for eight beats, 3,5 seconds, in the style of the old 420 meme videos. Pixel sunglasses slide down onto his face. When they land the screen jumps, 420 slams in and leaves burst out from behind him, and on the next big beat CHILL MAN follows. Rays turn behind him, leaves rain in three depths, and he blows a cloud and three smoke rings. It is silent, and Impact (a Windows font) is used for the big words. The cannabis leaf is a free drawing from Wikimedia Commons, see `docs/image-credits.md`. 2× Coca-Cola plus 1× Milka gets you there. The totals of the last two are `DEVIL_TOTALS` and `FOUR_TWENTY_TOTALS` in `web/src/till/Sell.svelte`.

All three monkeys share one body, drawn after the logo: `web/src/till/monkey.js` holds the shapes and `web/src/till/MonkeyHand.svelte` the open hand and the fist, so a change there reaches all of them.

A Glass colour is one line per mode at the end of `web/src/themes/glass.css`, plus an entry in `glassTints` in `web/src/lib/theme.svelte.js`. Check the text contrast of a new colour in both modes, on the wallpaper as well as on the panels.
