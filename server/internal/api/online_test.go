package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"testing/fstest"

	"amkpos/internal/store"
)

// webPage stands in for the OBK's online dashboard.
type webPage struct {
	*httptest.Server
	mu      sync.Mutex
	bodies  []string
	auth    []string
	refused int
}

func newWebPage(t *testing.T) *webPage {
	t.Helper()
	page := &webPage{}
	page.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		page.mu.Lock()
		defer page.mu.Unlock()
		if page.refused != 0 {
			w.WriteHeader(page.refused)
			return
		}
		page.bodies = append(page.bodies, string(body))
		page.auth = append(page.auth, r.Header.Get("Authorization"))
		if r.URL.Path != "/api/push" {
			t.Errorf("sent to %s", r.URL.Path)
		}
	}))
	t.Cleanup(page.Close)
	return page
}

func (p *webPage) received() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.bodies)
}

func newOnlineServer(t *testing.T) (*Server, http.Handler, *http.Cookie) {
	t.Helper()
	s, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	server := New(s, fstest.MapFS{"index.html": {Data: []byte("<html></html>")}})
	h := server.Handler()
	return server, h, loginAs(t, h)
}

func onlineBody(address, key string) string {
	body, _ := json.Marshal(map[string]string{"url": address, "key": key})
	return string(body)
}

func TestSavingTheLineSendsAndTheKeyNeverComesBack(t *testing.T) {
	page := newWebPage(t)
	_, h, cookie := newOnlineServer(t)

	w := call(h, "POST", "/api/admin/online", local, onlineBody(page.URL+"/", "geheimer-schluessel"), cookie)
	want(t, w, 200)
	if strings.Contains(w.Body.String(), "geheimer-schluessel") {
		t.Errorf("the key came back: %s", w.Body.String())
	}
	if page.received() != 1 || page.auth[0] != "Bearer geheimer-schluessel" {
		t.Fatalf("received %d, auth %v", page.received(), page.auth)
	}

	w = call(h, "GET", "/api/admin/online", local, "", cookie)
	want(t, w, 200)
	var status struct {
		URL      string `json:"url"`
		HasKey   bool   `json:"hasKey"`
		LastPush int64  `json:"lastPush"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &status); err != nil {
		t.Fatal(err)
	}
	if status.URL != page.URL || !status.HasKey || status.LastPush == 0 {
		t.Errorf("status = %+v", status)
	}
	if strings.Contains(w.Body.String(), "geheimer-schluessel") {
		t.Errorf("the key came back: %s", w.Body.String())
	}
}

func TestNothingNewIsNotSentAgain(t *testing.T) {
	page := newWebPage(t)
	server, h, cookie := newOnlineServer(t)
	want(t, call(h, "POST", "/api/admin/online", local, onlineBody(page.URL, "k"), cookie), 200)

	server.pushOnline(context.Background(), false)
	if page.received() != 1 {
		t.Fatalf("sent %d times with nothing new", page.received())
	}

	want(t, call(h, "POST", "/api/shifts", local, `{"room":"214","name":"Jonas","floatCents":1000}`), 200)
	server.pushOnline(context.Background(), false)
	if page.received() != 2 {
		t.Fatalf("a new shift was not sent, %d", page.received())
	}
}

func TestProblemsReachTheOBK(t *testing.T) {
	page := newWebPage(t)
	page.refused = http.StatusUnauthorized
	_, h, cookie := newOnlineServer(t)

	w := call(h, "POST", "/api/admin/online", local, onlineBody(page.URL, "falsch"), cookie)
	want(t, w, 200)
	if !strings.Contains(w.Body.String(), `"lastError":"wrong_key"`) {
		t.Errorf("wrong key: %s", w.Body.String())
	}

	page.Close()
	w = call(h, "POST", "/api/admin/online/push", local, "", cookie)
	want(t, w, 200)
	if !strings.Contains(w.Body.String(), `"lastError":"offline"`) {
		t.Errorf("no internet: %s", w.Body.String())
	}
}

func TestTheAddressMustBeSafe(t *testing.T) {
	_, h, cookie := newOnlineServer(t)
	for _, address := range []string{"http://example.com", "ftp://example.com", "example.com", "https://user:pw@example.com"} {
		want(t, call(h, "POST", "/api/admin/online", local, onlineBody(address, "k"), cookie), 400)
	}
	want(t, call(h, "POST", "/api/admin/online", local, onlineBody("https://bar.example.workers.dev", "k"), cookie), 200)
}

func TestAnEmptyAddressSwitchesItOffAndForgetsTheKey(t *testing.T) {
	page := newWebPage(t)
	_, h, cookie := newOnlineServer(t)
	want(t, call(h, "POST", "/api/admin/online", local, onlineBody(page.URL, "k"), cookie), 200)

	w := call(h, "POST", "/api/admin/online", local, onlineBody("", ""), cookie)
	want(t, w, 200)
	if !strings.Contains(w.Body.String(), `"hasKey":false`) {
		t.Errorf("status: %s", w.Body.String())
	}
	if page.received() != 1 {
		t.Errorf("still sending: %d", page.received())
	}
}

func TestOnlyTheOBKReachesTheLine(t *testing.T) {
	_, h, cookie := newOnlineServer(t)
	want(t, call(h, "GET", "/api/admin/online", local, ""), 401)
	want(t, call(h, "POST", "/api/admin/online", remote, onlineBody("", "")), 401)
	want(t, call(h, "POST", "/api/admin/online/push", local, ""), 401)
	want(t, call(h, "GET", "/api/admin/online", remote, "", cookie), 200)
}

func TestStartingAShiftBringsTheLastMessage(t *testing.T) {
	_, h, _ := newOnlineServer(t)
	want(t, call(h, "POST", "/api/shifts", local, `{"room":"214","name":"Jonas","floatCents":0}`), 200)
	want(t, call(h, "POST", "/api/shift/close", local, `{"countedCents":0,"note":"","handover":"Kühlschrank tropft"}`), 200)

	w := call(h, "POST", "/api/shifts", local, `{"room":"108","name":"Tim","floatCents":0}`)
	want(t, w, 200)
	var opened struct {
		WorkerName string `json:"workerName"`
		Message    *struct {
			Text string `json:"text"`
			From string `json:"from"`
		} `json:"message"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &opened); err != nil {
		t.Fatal(err)
	}
	if opened.WorkerName != "Tim" || opened.Message == nil || opened.Message.Text != "Kühlschrank tropft" || opened.Message.From != "Jonas" {
		t.Errorf("opened = %s", w.Body.String())
	}
}
