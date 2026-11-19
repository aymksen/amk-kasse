package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"amkpos/internal/store"
)

const (
	keyOnlineURL = "online_url"
	keyOnlineKey = "online_key"

	onlineEvery     = 5 * time.Minute  // how often the till looks for something new to send
	onlineHeartbeat = time.Hour        // sent even with nothing new, so the page can tell the PC is on
	onlineTimeout   = 20 * time.Second // the bar's internet may be bad, and nothing here may ever wait for it
)

// What can go wrong when sending. The OBK page turns these into a sentence.
const (
	onlineOffline  = "offline"
	onlineWrongKey = "wrong_key"
	onlineRefused  = "refused"
)

// online is the one-way line to the OBK's web page: the till sends a picture of its numbers and takes nothing back, so
// nothing on the internet can change anything at the bar. Without an address and key it does nothing at all.
type online struct {
	mu       sync.Mutex
	lastPush int64
	lastErr  string
	lastSent [sha256.Size]byte
	wake     chan struct{}
	client   *http.Client
}

func newOnline() *online {
	return &online{
		wake: make(chan struct{}, 1),
		client: &http.Client{
			Timeout: onlineTimeout,
			// The key only ever goes to the address the OBK entered.
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
	}
}

// nudge sends soon, without making anyone wait. A shift starting or ending is worth more than the five-minute round.
func (o *online) nudge() {
	select {
	case o.wake <- struct{}{}:
	default:
	}
}

func (o *online) status() (int64, string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.lastPush, o.lastErr
}

func (o *online) record(at int64, sent [sha256.Size]byte, problem string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if problem != o.lastErr && problem != "" {
		log.Printf("online dashboard: %s", problem) // once, not every five minutes while the internet is down
	}
	o.lastErr = problem
	if problem == "" {
		o.lastPush = at
		o.lastSent = sent
	}
}

func (o *online) forgetLastSent() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.lastSent = [sha256.Size]byte{}
	o.lastPush = 0
	o.lastErr = ""
}

// OnlineSync keeps the web page up to date for as long as the program runs.
func (a *Server) OnlineSync(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(onlineEvery):
		case <-a.online.wake:
		}
		a.pushOnline(ctx, false)
	}
}

// pushOnline sends the numbers if there is something new, or if `force` says so.
func (a *Server) pushOnline(ctx context.Context, force bool) {
	address, err := a.store.Setting(ctx, keyOnlineURL, "")
	if err != nil {
		log.Printf("online dashboard: %v", err)
		return
	}
	key, err := a.store.Setting(ctx, keyOnlineKey, "")
	if err != nil {
		log.Printf("online dashboard: %v", err)
		return
	}
	if address == "" || key == "" {
		return
	}

	snapshot, err := a.store.OnlineSnapshot(ctx)
	if err != nil {
		log.Printf("online dashboard: %v", err)
		return
	}
	settings := a.publicSettings(ctx)
	snapshot.BarName, snapshot.Language = settings["barName"], settings["language"]

	// The moment of sending is always new, so it stays out of the comparison.
	at := snapshot.At
	snapshot.At = 0
	unstamped, err := json.Marshal(snapshot)
	if err != nil {
		log.Printf("online dashboard: %v", err)
		return
	}
	fingerprint := sha256.Sum256(unstamped)
	last, _ := a.online.status()
	a.online.mu.Lock()
	unchanged := fingerprint == a.online.lastSent
	a.online.mu.Unlock()
	if !force && unchanged && time.Duration(at-last)*time.Millisecond < onlineHeartbeat {
		return
	}

	snapshot.At = at
	body, err := json.Marshal(snapshot)
	if err != nil {
		log.Printf("online dashboard: %v", err)
		return
	}
	a.online.record(at, fingerprint, a.sendOnline(ctx, address, key, body))
}

func (a *Server) sendOnline(ctx context.Context, address, key string, body []byte) string {
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, address+"/api/push", bytes.NewReader(body))
	if err != nil {
		return onlineRefused
	}
	r.Header.Set("Authorization", "Bearer "+key)
	r.Header.Set("Content-Type", "application/json")
	resp, err := a.online.client.Do(r)
	if err != nil {
		return onlineOffline
	}
	defer resp.Body.Close()
	if _, err := io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<16)); err != nil {
		return onlineOffline
	}
	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		return ""
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		return onlineWrongKey
	default:
		return onlineRefused
	}
}

// validOnlineAddress wants https, because the key travels with every request. Plain http is for trying it on this PC.
func validOnlineAddress(address string) bool {
	u, err := url.Parse(address)
	if err != nil || u.Host == "" || u.User != nil {
		return false
	}
	host := u.Hostname()
	return u.Scheme == "https" || (u.Scheme == "http" && (host == "localhost" || host == "127.0.0.1"))
}

func (a *Server) writeOnlineStatus(w http.ResponseWriter, ctx context.Context) {
	address, err := a.store.Setting(ctx, keyOnlineURL, "")
	if err != nil {
		fail(w, err)
		return
	}
	key, err := a.store.Setting(ctx, keyOnlineKey, "")
	if err != nil {
		fail(w, err)
		return
	}
	last, problem := a.online.status()
	// The key goes in and never comes back out, not even to the OBK.
	writeJSON(w, http.StatusOK, map[string]any{"url": address, "hasKey": key != "", "lastPush": last, "lastError": problem})
}

func (a *Server) onlineStatus(w http.ResponseWriter, r *http.Request) {
	a.writeOnlineStatus(w, r.Context())
}

// saveOnline keeps the old key when the field stays empty, and an empty address switches the line off.
func (a *Server) saveOnline(w http.ResponseWriter, r *http.Request) {
	in, valid := decode[struct {
		URL string `json:"url"`
		Key string `json:"key"`
	}](w, r)
	if !valid {
		return
	}
	address := strings.TrimRight(strings.TrimSpace(in.URL), "/")
	key := strings.TrimSpace(in.Key)
	if address != "" && !validOnlineAddress(address) {
		fail(w, store.ErrInvalid)
		return
	}
	ctx := r.Context()
	if err := a.store.SetSetting(ctx, keyOnlineURL, address); err != nil {
		fail(w, err)
		return
	}
	if address == "" {
		key = ""
	}
	if key != "" || address == "" {
		if err := a.store.SetSetting(ctx, keyOnlineKey, key); err != nil {
			fail(w, err)
			return
		}
	}
	a.online.forgetLastSent()
	a.pushOnline(ctx, true)
	a.writeOnlineStatus(w, ctx)
}

func (a *Server) pushOnlineNow(w http.ResponseWriter, r *http.Request) {
	a.pushOnline(r.Context(), true)
	a.writeOnlineStatus(w, r.Context())
}
