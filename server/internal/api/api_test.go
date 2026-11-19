package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"amkpos/internal/store"
)

const (
	local  = "127.0.0.1:50000"
	remote = "192.168.1.50:50000"
)

func newServer(t *testing.T) http.Handler {
	t.Helper()
	s, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	ui := fstest.MapFS{"index.html": {Data: []byte("<html>till</html>")}, "assets/a.js": {Data: []byte("1")}}
	return New(s, ui).Handler()
}

func call(h http.Handler, method, url, addr, body string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, url, strings.NewReader(body))
	r.RemoteAddr = addr
	r.Host = "127.0.0.1:8417"
	for _, c := range cookies {
		r.AddCookie(c)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func want(t *testing.T, w *httptest.ResponseRecorder, status int) {
	t.Helper()
	if w.Code != status {
		t.Fatalf("status = %d, want %d: %s", w.Code, status, w.Body.String())
	}
}

// loginAs creates the PIN from the bar PC and returns the session cookie.
func loginAs(t *testing.T, h http.Handler) *http.Cookie {
	t.Helper()
	w := call(h, "POST", "/api/admin/setup", local, `{"pin":"4711"}`)
	want(t, w, 200)
	return w.Result().Cookies()[0]
}

func TestBarPCCanSellWithoutLogin(t *testing.T) {
	h := newServer(t)
	want(t, call(h, "GET", "/api/state", local, ""), 200)
	want(t, call(h, "POST", "/api/shifts", local, `{"room":"214","name":"Jonas","floatCents":5000}`), 200)
	want(t, call(h, "POST", "/api/sales", local, `{"lines":[{"productId":1,"qty":2}],"payment":"cash"}`), 200)
}

type stockCheck struct {
	Last *struct {
		Name    string `json:"name"`
		Counted bool   `json:"counted"`
	} `json:"last"`
	Lines []struct {
		ProductID int64 `json:"productId"`
		Expected  int64 `json:"expected"`
	} `json:"lines"`
}

func readCheck(t *testing.T, h http.Handler) stockCheck {
	t.Helper()
	w := call(h, "GET", "/api/stock/check", local, "")
	want(t, w, 200)
	var check stockCheck
	if err := json.Unmarshal(w.Body.Bytes(), &check); err != nil || len(check.Lines) == 0 {
		t.Fatalf("check = %s, %v", w.Body, err)
	}
	return check
}

// The start and the end of a shift show the bar's figures, and the worker's check is saved with the shift.
func TestTheShiftStartsAndEndsWithTheCheckOfTheBar(t *testing.T) {
	h := newServer(t)
	first := readCheck(t, h).Lines[0]
	body := fmt.Sprintf(`{"room":"214","name":"Jonas","floatCents":0,"count":[{"productId":%d,"counted":%d}]}`, first.ProductID, first.Expected+1)
	want(t, call(h, "POST", "/api/shifts", local, body), 200)

	during := readCheck(t, h)
	if during.Last != nil || during.Lines[0].Expected != first.Expected+1 {
		t.Errorf("during the shift: last %+v, first line %+v", during.Last, during.Lines[0])
	}
	body = fmt.Sprintf(`{"countedCents":0,"count":[{"productId":%d,"counted":%d}]}`, first.ProductID, first.Expected)
	want(t, call(h, "POST", "/api/shift/close", local, body), 200)
	if after := readCheck(t, h); after.Last == nil || after.Last.Name != "Jonas" || !after.Last.Counted || after.Lines[0].Expected != first.Expected {
		t.Errorf("after the shift: last %+v, first line %+v", after.Last, after.Lines[0])
	}
	want(t, call(h, "GET", "/api/stock/check", remote, ""), 401)
}

func TestOtherDevicesCannotUseTheTillWithoutASession(t *testing.T) {
	h := newServer(t)
	want(t, call(h, "GET", "/api/state", remote, ""), 401)
	want(t, call(h, "POST", "/api/sales", remote, `{"lines":[{"productId":1,"qty":1}],"payment":"cash"}`), 401)
	want(t, call(h, "POST", "/api/shifts", remote, `{"room":"1","name":"x","floatCents":0}`), 401)
}

func TestAdminPagesNeedASessionEvenOnTheBarPC(t *testing.T) {
	h := newServer(t)
	want(t, call(h, "GET", "/api/admin/dashboard", local, ""), 401)
	cookie := loginAs(t, h)
	want(t, call(h, "GET", "/api/admin/dashboard", local, "", cookie), 200)
}

func TestFreshInstallCanOnlyBeClaimedFromTheBarPC(t *testing.T) {
	h := newServer(t)
	want(t, call(h, "POST", "/api/admin/setup", remote, `{"pin":"0000"}`), 403)
	loginAs(t, h)
	want(t, call(h, "POST", "/api/admin/setup", local, `{"pin":"1234"}`), 403)
}

func TestOnlyTheOBKReachesTheStoreroom(t *testing.T) {
	h := newServer(t)
	cookie := loginAs(t, h)
	refill := `{"lines":[{"productId":1,"qty":2}]}`
	count := `{"lines":[{"productId":1,"counted":5}]}`
	for _, url := range []string{"/api/admin/refills", "/api/admin/stocktakes"} {
		want(t, call(h, "GET", url, local, ""), 401) // the bar PC needs the PIN here like anyone else
	}
	want(t, call(h, "POST", "/api/admin/refills", local, refill), 401)
	want(t, call(h, "POST", "/api/admin/stocktakes", local, count), 401)
	want(t, call(h, "POST", "/api/refills", local, refill), 404)

	want(t, call(h, "POST", "/api/admin/refills", local, refill, cookie), 200)
	want(t, call(h, "POST", "/api/admin/stocktakes", local, count, cookie), 200)
	w := call(h, "GET", "/api/admin/stocktakes", local, "", cookie)
	if !strings.Contains(w.Body.String(), `"place":"storage"`) || !strings.Contains(w.Body.String(), `"byName":"OBK"`) {
		t.Errorf("counts = %s", w.Body)
	}
}

func TestTheTillCanOnlyCountTheBar(t *testing.T) {
	h := newServer(t)
	cookie := loginAs(t, h)
	body := `{"kind":"full","place":"storage","room":"214","name":"Jonas","lines":[{"productId":1,"counted":7}]}`
	want(t, call(h, "POST", "/api/stocktakes", local, body), 200)
	w := call(h, "GET", "/api/admin/stocktakes", local, "", cookie)
	if !strings.Contains(w.Body.String(), `"place":"bar"`) {
		t.Errorf("counts = %s", w.Body)
	}
}

func TestDashboardToldTheOBKWhatToRefill(t *testing.T) {
	h := newServer(t)
	cookie := loginAs(t, h)
	w := call(h, "GET", "/api/admin/dashboard", local, "", cookie)
	want(t, w, 200)
	if !strings.Contains(w.Body.String(), `"toRefill":`) {
		t.Errorf("dashboard = %s", w.Body)
	}
}

func TestDeliveryNeedsAPlace(t *testing.T) {
	h := newServer(t)
	cookie := loginAs(t, h)
	line := `"lines":[{"productId":1,"qty":6,"unitCostCents":90}]`
	want(t, call(h, "POST", "/api/admin/restocks", local, `{"place":"keller",`+line+`}`, cookie), 400)
	want(t, call(h, "POST", "/api/admin/restocks", local, `{"place":"storage",`+line+`}`, cookie), 200)
}

func TestTheSurprisesHaveNoOffSwitch(t *testing.T) {
	h := newServer(t)
	cookie := loginAs(t, h)
	// An older admin page may still send the switch; the till ignores it.
	want(t, call(h, "POST", "/api/admin/settings", local, `{"barName":"AMK Bar","language":"de","fun":"off"}`, cookie), 200)
	w := call(h, "GET", "/api/state", local, "")
	want(t, w, 200)
	var state struct {
		Settings map[string]string `json:"settings"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if got, found := state.Settings["fun"]; found {
		t.Fatalf("the till still sends a surprise switch: %q", got)
	}
}

func TestWrongPINsLockTheLogin(t *testing.T) {
	h := newServer(t)
	loginAs(t, h)
	for i := 0; i < maxFailures; i++ {
		want(t, call(h, "POST", "/api/admin/login", remote, `{"pin":"9999"}`), 401)
	}
	want(t, call(h, "POST", "/api/admin/login", remote, `{"pin":"4711"}`), 429)
}

func TestSessionWorksFromAnotherDevice(t *testing.T) {
	h := newServer(t)
	loginAs(t, h)
	w := call(h, "POST", "/api/admin/login", remote, `{"pin":"4711"}`)
	want(t, w, 200)
	cookie := w.Result().Cookies()[0]
	want(t, call(h, "GET", "/api/admin/products", remote, "", cookie), 200)
	want(t, call(h, "POST", "/api/admin/logout", remote, "", cookie), 200)
	want(t, call(h, "GET", "/api/admin/products", remote, "", cookie), 401)
}

func TestSessionExpiresWhenIdle(t *testing.T) {
	now := time.Now()
	a := newAuth(func() time.Time { return now })
	token := a.startSession()
	now = now.Add(sessionIdle - time.Minute)
	if !a.valid(token) {
		t.Fatal("session died too early")
	}
	now = now.Add(sessionIdle - time.Minute)
	if !a.valid(token) {
		t.Fatal("activity should extend the session")
	}
	now = now.Add(sessionIdle + time.Minute)
	if a.valid(token) {
		t.Fatal("idle session should expire")
	}
}

func TestCrossSiteWritesAreRejected(t *testing.T) {
	h := newServer(t)
	r := httptest.NewRequest("POST", "http://127.0.0.1:8417/api/shifts", strings.NewReader(`{"room":"1","name":"x"}`))
	r.RemoteAddr = local
	r.Header.Set("Origin", "http://evil.example")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	want(t, w, 403)
}

func TestForeignHostNameIsNotTheBarPC(t *testing.T) {
	h := newServer(t)
	r := httptest.NewRequest("GET", "/api/state", nil)
	r.RemoteAddr = local
	r.Host = "evil.example:8417"
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	want(t, w, 401)
}

func TestErrorsReachTheClientAsCodes(t *testing.T) {
	h := newServer(t)
	w := call(h, "POST", "/api/sales", local, `{"lines":[{"productId":1,"qty":1}],"payment":"cash"}`)
	want(t, w, 409)
	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["error"] != "no_open_shift" {
		t.Errorf("body = %v", body)
	}
}

func TestStaticFilesAndClientRoutes(t *testing.T) {
	h := newServer(t)
	w := call(h, "GET", "/admin/products", remote, "")
	want(t, w, 200)
	if !strings.Contains(w.Body.String(), "till") {
		t.Errorf("client route should serve index.html: %q", w.Body.String())
	}
	w = call(h, "GET", "/assets/a.js", remote, "")
	want(t, w, 200)
	if !strings.Contains(w.Header().Get("Cache-Control"), "immutable") {
		t.Errorf("assets should cache: %v", w.Header())
	}
	want(t, call(h, "GET", "/api/nope", remote, ""), 404)
}

func TestExportIsExcelReady(t *testing.T) {
	h := newServer(t)
	cookie := loginAs(t, h)
	call(h, "POST", "/api/shifts", local, `{"room":"214","name":"Jonas","floatCents":0}`)
	call(h, "POST", "/api/sales", local, `{"lines":[{"productId":1,"qty":2}],"payment":"cash"}`)
	day := time.Now().Add(-6 * time.Hour).Format("2006-01-02")

	w := call(h, "GET", "/api/admin/export/sales.csv?from="+day+"&to="+day, local, "", cookie)
	want(t, w, 200)
	body := w.Body.String()
	if !strings.HasPrefix(body, "\xEF\xBB\xBF") {
		t.Error("missing BOM")
	}
	if !strings.Contains(body, "Franziskaner Weißbier;Bier;2;1,40;2,80;Bar") {
		t.Errorf("row missing: %s", body)
	}
	want(t, call(h, "GET", "/api/admin/export/unknown.csv", local, "", cookie), 404)
}

func TestMoneyAndFormulaSafety(t *testing.T) {
	if got := money(-15); got != "-0,15" {
		t.Errorf("money = %s", got)
	}
	if got := money(123450); got != "1234,50" {
		t.Errorf("money = %s", got)
	}
	if got := safeText("=SUM(A1)"); got != "'=SUM(A1)" {
		t.Errorf("safeText = %s", got)
	}
}

// A tiny but real WebP header, enough for the type check.
const webp = "RIFF\x24\x00\x00\x00WEBPVP8 \x18\x00\x00\x00"

func firstProduct(t *testing.T, h http.Handler, cookie *http.Cookie) int64 {
	t.Helper()
	w := call(h, "GET", "/api/admin/products", local, "", cookie)
	want(t, w, 200)
	var products []store.Product
	if err := json.Unmarshal(w.Body.Bytes(), &products); err != nil {
		t.Fatal(err)
	}
	return products[0].ID
}

func TestPictureUploadNeedsTheAdminLogin(t *testing.T) {
	h := newServer(t)
	cookie := loginAs(t, h)
	url := "/api/admin/products/" + strconv.FormatInt(firstProduct(t, h, cookie), 10) + "/image"
	want(t, call(h, "POST", url, local, webp), 401)
	want(t, call(h, "POST", url, remote, webp), 401)
	want(t, call(h, "DELETE", url, remote, ""), 401)
}

func TestPictureUploadRefusesOtherFiles(t *testing.T) {
	h := newServer(t)
	cookie := loginAs(t, h)
	url := "/api/admin/products/" + strconv.FormatInt(firstProduct(t, h, cookie), 10) + "/image"
	for _, body := range []string{`<svg onload="alert(1)"></svg>`, "<html><script>1</script></html>", "plain text", ""} {
		want(t, call(h, "POST", url, local, body, cookie), 400)
	}
}

func TestPictureIsServedAndCached(t *testing.T) {
	h := newServer(t)
	cookie := loginAs(t, h)
	id := strconv.FormatInt(firstProduct(t, h, cookie), 10)
	want(t, call(h, "POST", "/api/admin/products/"+id+"/image", local, webp, cookie), 200)

	w := call(h, "GET", "/api/images/"+id+"?v=1", local, "")
	want(t, w, 200)
	if w.Header().Get("Content-Type") != "image/webp" || w.Body.String() != webp {
		t.Errorf("type %q, body %q", w.Header().Get("Content-Type"), w.Body.String())
	}
	if !strings.Contains(w.Header().Get("Cache-Control"), "max-age=31536000") {
		t.Errorf("cache = %q", w.Header().Get("Cache-Control"))
	}
	want(t, call(h, "GET", "/api/images/"+id, remote, ""), 401)

	want(t, call(h, "DELETE", "/api/admin/products/"+id+"/image", local, "", cookie), 200)
	want(t, call(h, "GET", "/api/images/"+id, local, ""), 404)
}

// The recap names the people behind the bar, so it stays on the bar PC and with a logged-in OBK.
func TestRecapNeedsTheBarPCOrASession(t *testing.T) {
	h := newServer(t)
	want(t, call(h, "GET", "/api/recap", remote, ""), 401)
	want(t, call(h, "GET", "/api/recap", local, ""), 200)
	cookie := loginAs(t, h)
	want(t, call(h, "GET", "/api/recap", remote, "", cookie), 200)
}

// A game score goes to whoever is on shift, never to a name the request brings along.
func TestGameScoresBelongToTheShift(t *testing.T) {
	h := newServer(t)
	want(t, call(h, "POST", "/api/games/flyer/scores", remote, `{"score":5}`), 401)
	want(t, call(h, "POST", "/api/games/flyer/scores", local, `{"score":5}`), 409)
	want(t, call(h, "POST", "/api/shifts", local, `{"room":"214","name":"Jonas","floatCents":0}`), 200)
	w := call(h, "GET", "/api/games", local, "")
	want(t, w, 200)
	if body := strings.TrimSpace(w.Body.String()); body != `{"flyer":{"top":[],"mine":null}}` {
		t.Errorf("empty boards = %q", body)
	}
	want(t, call(h, "POST", "/api/games/snake/scores", local, `{"score":5}`), 400)
	w = call(h, "POST", "/api/games/flyer/scores", local, `{"score":5,"name":"Someone","room":"1"}`)
	want(t, w, 200)
	if body := w.Body.String(); !strings.Contains(body, `"place":1`) || !strings.Contains(body, `"name":"Jonas","room":"214","score":5`) {
		t.Errorf("result = %s", body)
	}
}
