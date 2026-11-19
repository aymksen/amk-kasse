# Online dashboard

The OBK's web page. It shows today, the last 30 days, this month and last month, what is running low, and the recent shifts, on a phone or a laptop, from anywhere.

## Set it up

You need a free [Cloudflare](https://dash.cloudflare.com/sign-up) account and, once, a PC with [Node.js](https://nodejs.org) (the LTS version). It does not have to be the bar PC.

1. Double-click `einrichten.cmd` in this folder.
2. The first time, a browser opens. Log in to Cloudflare and allow the access.
3. Choose a password for the web page, at least 8 characters. Everyone in the OBK uses this one.
4. If Cloudflare asks for a name for your workers.dev address, pick one. It becomes part of the address.
5. At the end you get an **address** and a **key**. Open the till, go to OBK → Einstellungen → Online-Dashboard, paste both, and press Speichern.

The till says "Zuletzt gesendet" with the time if it worked. Open the address and log in with the password.

Running `einrichten.cmd` again is fine, for example to change the password. The page keeps its numbers, everyone is logged out, and you get a new key to paste into the till.

## How it works

The till sends a summary to `/api/push` with the key. The page keeps only the latest one. It does not build up a history, and the OBK reads it with the password. Nothing goes back from the page to the till.

- A login lasts 30 days. After four wrong passwords the next try locks the login for a minute, then two, up to an hour. Someone already logged in stays logged in.
- The key is only for the till. It stays on the bar PC and the OBK area never shows it again.
- The free plan is far more than enough: the till sends at most one update every five minutes.
- The page speaks the language set in the till.

What is sent is listed in `OnlineSnapshot` in `server/internal/store/online.go`. Names and rooms of workers and renters stay on the bar PC. Handover messages and shift notes go online as typed, so anything personal in them does too.

To switch it off, empty the address in the till and press Speichern. The page keeps showing the last numbers until you delete the Worker `amk-bar` in the Cloudflare dashboard.

## Try it on your own PC

Put test values in a file called `.dev.vars`, then run `npm run dev`:

```
SYNC_KEY=any-test-value
PASSWORD=any-test-value
```

In the till, the address is then `http://localhost:8787`. Plain http only works for localhost.
