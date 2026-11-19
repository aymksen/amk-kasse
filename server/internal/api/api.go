// Package api is the JSON API and static file server for the web UI.
package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"amkpos/internal/store"
)

const (
	KeyPIN          = "admin_pin_hash"
	keyBarName      = "bar_name"
	keyLanguage     = "language"
	keyOtherPayment = "other_payment_label"
	keyBackupDir    = "backup_extra_dir"
)

// Server carries the dependencies of every handler.
type Server struct {
	store  *store.Store
	auth   *auth
	ui     fs.FS
	online *online
	Demo   bool
}

func New(s *store.Store, ui fs.FS) *Server {
	return &Server{store: s, auth: newAuth(s.Now), ui: ui, online: newOnline()}
}

func (a *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	till := func(pattern string, h http.HandlerFunc) { mux.Handle(pattern, a.tillOnly(h)) }
	admin := func(pattern string, h http.HandlerFunc) { mux.Handle(pattern, a.adminOnly(h)) }

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, 200, ok) })

	till("GET /api/state", a.state)
	till("GET /api/workers/{room}", a.workerNames)
	till("GET /api/stock/check", a.stockCheck)
	till("POST /api/shifts", a.openShift)
	till("GET /api/shift/totals", a.shiftTotals)
	till("POST /api/shift/close", a.closeShift)
	till("GET /api/shift/sales", a.shiftSales)
	till("POST /api/sales", a.createSale)
	till("POST /api/sales/{id}/void", a.voidSale)
	till("POST /api/stocktakes", a.saveStocktake)
	till("GET /api/rentals", a.rentalBoard)
	till("POST /api/rentals", a.startRental)
	till("GET /api/rentals/{id}/quote", a.rentalQuote)
	till("POST /api/rentals/{id}/return", a.returnRental)
	till("GET /api/images/{id}", a.image)
	till("GET /api/recap", a.recap)
	till("GET /api/games", a.gameBoards)
	till("POST /api/games/{game}/scores", a.saveGameScore)

	mux.HandleFunc("GET /api/admin/session", a.session)
	mux.HandleFunc("POST /api/admin/setup", a.setup)
	mux.HandleFunc("POST /api/admin/login", a.login)
	mux.HandleFunc("POST /api/admin/logout", a.logout)
	admin("POST /api/admin/pin", a.changePIN)
	admin("GET /api/admin/dashboard", a.dashboard)
	admin("GET /api/admin/products", a.products)
	admin("POST /api/admin/products", a.saveProduct)
	admin("POST /api/admin/products/{id}/image", a.saveImage)
	admin("DELETE /api/admin/products/{id}/image", a.deleteImage)
	admin("GET /api/admin/categories", a.categories)
	admin("POST /api/admin/categories", a.saveCategory)
	admin("GET /api/admin/restocks", a.restocks)
	admin("POST /api/admin/restocks", a.saveRestock)
	admin("GET /api/admin/refills", a.refills)
	admin("POST /api/admin/refills", a.saveRefill)
	admin("GET /api/admin/stocktakes", a.stocktakes)
	admin("POST /api/admin/stocktakes", a.saveStorageCount)
	admin("GET /api/admin/stocktakes/{id}", a.stocktakeLines)
	admin("GET /api/admin/shifts", a.listShifts)
	admin("GET /api/admin/shifts/{id}", a.shiftDetail)
	admin("GET /api/admin/cash", a.cash)
	admin("POST /api/admin/cash", a.addCashMove)
	admin("GET /api/admin/sales", a.listSales)
	admin("GET /api/admin/rental-items", a.rentalItems)
	admin("POST /api/admin/rental-items", a.saveRentalItem)
	admin("GET /api/admin/rental-history", a.rentalHistory)
	admin("GET /api/admin/report", a.report)
	admin("GET /api/admin/export/{kind}", a.export)
	admin("GET /api/admin/settings", a.getSettings)
	admin("POST /api/admin/settings", a.saveSettings)
	admin("GET /api/admin/backups", a.backups)
	admin("POST /api/admin/backups", a.backupNow)
	admin("GET /api/admin/online", a.onlineStatus)
	admin("POST /api/admin/online", a.saveOnline)
	admin("POST /api/admin/online/push", a.pushOnlineNow)

	mux.Handle("/", a.static())
	return a.guard(mux)
}

var ok = map[string]bool{"ok": true}

// Access rules ---------------------------------------------------------------

// guard rejects cross-site writes and caps request size on every route.
func (a *Server) guard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			if origin := r.Header.Get("Origin"); origin != "" {
				if u, err := url.Parse(origin); err != nil || u.Host != r.Host {
					writeJSON(w, http.StatusForbidden, errBody("forbidden"))
					return
				}
			}
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		next.ServeHTTP(w, r)
	})
}

// isLocal also checks the Host header: a website that points its own domain at
// 127.0.0.1 could otherwise reach the till through the bar PC's browser.
func isLocal(r *http.Request) bool {
	addr, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(addr)
	if ip == nil || !ip.IsLoopback() {
		return false
	}
	name := r.Host
	if host, _, err := net.SplitHostPort(r.Host); err == nil {
		name = host
	}
	name = strings.Trim(name, "[]")
	parsed := net.ParseIP(name)
	return name == "localhost" || (parsed != nil && parsed.IsLoopback())
}

func (a *Server) hasSession(r *http.Request) bool {
	c, err := r.Cookie(sessionCookie)
	return err == nil && a.auth.valid(c.Value)
}

// tillOnly lets the bar PC itself ring up sales without a login. Any other
// device on the network needs an admin session.
func (a *Server) tillOnly(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isLocal(r) && !a.hasSession(r) {
			writeJSON(w, http.StatusUnauthorized, errBody("unauthorized"))
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		next(w, r)
	})
}

func (a *Server) adminOnly(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.hasSession(r) {
			writeJSON(w, http.StatusUnauthorized, errBody("unauthorized"))
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		next(w, r)
	})
}

// Helpers ---------------------------------------------------------------------

func errBody(code string) map[string]string { return map[string]string{"error": code} }

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write response: %v", err)
	}
}

// userErrors are mistakes the worker can act on; anything else is logged as a bug.
var userErrors = map[error]int{
	store.ErrNoOpenShift:    http.StatusConflict,
	store.ErrShiftOpen:      http.StatusConflict,
	store.ErrEmptySale:      http.StatusBadRequest,
	store.ErrNotFound:       http.StatusNotFound,
	store.ErrAlreadyVoided:  http.StatusConflict,
	store.ErrNotVoidable:    http.StatusConflict,
	store.ErrReasonRequired: http.StatusBadRequest,
	store.ErrItemRentedOut:  http.StatusConflict,
	store.ErrInvalid:        http.StatusBadRequest,
	store.ErrDuplicateCode:  http.StatusConflict,
}

func fail(w http.ResponseWriter, err error) {
	for known, status := range userErrors {
		if errors.Is(err, known) {
			writeJSON(w, status, errBody(known.Error()))
			return
		}
	}
	log.Printf("internal error: %v", err)
	writeJSON(w, http.StatusInternalServerError, errBody("internal"))
}

// reply sends v, or the error if there is one.
func reply(w http.ResponseWriter, v any, err error) {
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

func decode[T any](w http.ResponseWriter, r *http.Request) (T, bool) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody(store.ErrInvalid.Error()))
		return v, false
	}
	return v, true
}

func pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody(store.ErrInvalid.Error()))
		return 0, false
	}
	return id, true
}

func dateRange(w http.ResponseWriter, r *http.Request) (string, string, bool) {
	from, to := r.URL.Query().Get("from"), r.URL.Query().Get("to")
	for _, d := range []string{from, to} {
		if _, err := time.Parse("2006-01-02", d); err != nil {
			writeJSON(w, http.StatusBadRequest, errBody(store.ErrInvalid.Error()))
			return "", "", false
		}
	}
	return from, to, true
}

// Till ------------------------------------------------------------------------

func (a *Server) state(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shift, err := a.store.CurrentShift(ctx)
	if err != nil {
		fail(w, err)
		return
	}
	categories, err := a.store.Categories(ctx)
	if err != nil {
		fail(w, err)
		return
	}
	active := make([]store.Category, 0, len(categories))
	for _, c := range categories {
		if c.Active {
			active = append(active, c)
		}
	}
	products, err := a.store.TillProducts(ctx)
	if err != nil {
		fail(w, err)
		return
	}
	float, err := a.store.ExpectedFloat(ctx)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"shift":              shift,
		"categories":         active,
		"products":           products,
		"expectedFloatCents": float,
		"settings":           a.publicSettings(ctx),
		"demo":               a.Demo,
	})
}

func (a *Server) publicSettings(ctx context.Context) map[string]string {
	get := func(key, fallback string) string {
		v, err := a.store.Setting(ctx, key, fallback)
		if err != nil {
			log.Printf("read setting %s: %v", key, err)
			return fallback
		}
		return v
	}
	return map[string]string{
		"barName":      get(keyBarName, "AMK Bar"),
		"language":     get(keyLanguage, "de"),
		"otherPayment": get(keyOtherPayment, "PayPal"),
	}
}

func (a *Server) recap(w http.ResponseWriter, r *http.Request) {
	recap, err := a.store.Recap(r.Context())
	reply(w, recap, err)
}

func (a *Server) gameBoards(w http.ResponseWriter, r *http.Request) {
	boards, err := a.store.GameBoards(r.Context())
	reply(w, boards, err)
}

func (a *Server) saveGameScore(w http.ResponseWriter, r *http.Request) {
	in, valid := decode[struct {
		Score int64 `json:"score"`
	}](w, r)
	if !valid {
		return
	}
	res, err := a.store.SaveGameScore(r.Context(), r.PathValue("game"), in.Score)
	reply(w, res, err)
}

func (a *Server) workerNames(w http.ResponseWriter, r *http.Request) {
	names, err := a.store.WorkerNames(r.Context(), r.PathValue("room"))
	reply(w, map[string][]string{"names": names}, err)
}

func (a *Server) openShift(w http.ResponseWriter, r *http.Request) {
	in, valid := decode[struct {
		Room       string            `json:"room"`
		Name       string            `json:"name"`
		FloatCents int64             `json:"floatCents"`
		Count      []store.CountLine `json:"count"` // the check of the bar, saved with the shift
	}](w, r)
	if !valid {
		return
	}
	sh, err := a.store.OpenShift(r.Context(), in.Room, in.Name, in.FloatCents, in.Count...)
	if err != nil {
		fail(w, err)
		return
	}
	a.online.nudge()
	// The worker sees this once, right now, so it is not part of the state the till reads all the time.
	message, err := a.store.LastMessage(r.Context())
	reply(w, struct {
		*store.Shift
		Message *store.ShiftMessage `json:"message"`
	}{sh, message}, err)
}

// stockCheck is what every count at the till starts from: at the start of a shift, during it and at its end.
// The worker sees what the bar should hold and only corrects what is different. The cash count stays blind.
func (a *Server) stockCheck(w http.ResponseWriter, r *http.Request) {
	check, err := a.store.StockCheck(r.Context())
	reply(w, check, err)
}

func (a *Server) shiftTotals(w http.ResponseWriter, r *http.Request) {
	sh, err := a.store.CurrentShift(r.Context())
	if err != nil {
		fail(w, err)
		return
	}
	if sh == nil {
		fail(w, store.ErrNoOpenShift)
		return
	}
	t, err := a.store.ShiftTotals(r.Context(), *sh)
	reply(w, t, err)
}

func (a *Server) closeShift(w http.ResponseWriter, r *http.Request) {
	in, valid := decode[struct {
		CountedCents int64             `json:"countedCents"`
		Note         string            `json:"note"`
		Handover     string            `json:"handover"`
		Count        []store.CountLine `json:"count"` // the check of the bar, saved with the end of the shift
	}](w, r)
	if !valid {
		return
	}
	sh, err := a.store.CloseShift(r.Context(), in.CountedCents, in.Note, in.Handover, in.Count...)
	if err != nil {
		fail(w, err)
		return
	}
	go a.runBackup() // the worker should not wait for a copy of the database
	a.online.nudge()
	writeJSON(w, http.StatusOK, sh)
}

func (a *Server) shiftSales(w http.ResponseWriter, r *http.Request) {
	sh, err := a.store.CurrentShift(r.Context())
	if err != nil {
		fail(w, err)
		return
	}
	if sh == nil {
		fail(w, store.ErrNoOpenShift)
		return
	}
	sales, err := a.store.ShiftSales(r.Context(), sh.ID)
	reply(w, sales, err)
}

func (a *Server) createSale(w http.ResponseWriter, r *http.Request) {
	in, valid := decode[struct {
		Lines   []store.LineInput `json:"lines"`
		Payment string            `json:"payment"`
	}](w, r)
	if !valid {
		return
	}
	sale, err := a.store.CreateSale(r.Context(), in.Lines, in.Payment)
	reply(w, sale, err)
}

func (a *Server) voidSale(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(w, r)
	if !valid {
		return
	}
	in, valid := decode[struct {
		Reason string `json:"reason"`
	}](w, r)
	if !valid {
		return
	}
	sale, err := a.store.VoidSale(r.Context(), id, in.Reason)
	reply(w, sale, err)
}

func (a *Server) saveStocktake(w http.ResponseWriter, r *http.Request) {
	in, valid := decode[struct {
		Kind  string            `json:"kind"`
		Room  string            `json:"room"`
		Name  string            `json:"name"`
		Lines []store.CountLine `json:"lines"`
	}](w, r)
	if !valid {
		return
	}
	id, err := a.store.SaveStocktake(r.Context(), in.Kind, store.PlaceBar, in.Room, in.Name, in.Lines)
	reply(w, map[string]int64{"id": id}, err)
}

// saveStorageCount is the OBK counting the storeroom. The till can only count the bar.
func (a *Server) saveStorageCount(w http.ResponseWriter, r *http.Request) {
	in, valid := decode[struct {
		Lines []store.CountLine `json:"lines"`
	}](w, r)
	if !valid {
		return
	}
	id, err := a.store.SaveStocktake(r.Context(), "full", store.PlaceStorage, "", "OBK", in.Lines)
	reply(w, map[string]int64{"id": id}, err)
}

func (a *Server) saveRefill(w http.ResponseWriter, r *http.Request) {
	in, valid := decode[struct {
		Lines []store.RefillLine `json:"lines"`
	}](w, r)
	if !valid {
		return
	}
	reply(w, ok, a.store.SaveRefill(r.Context(), in.Lines))
}

func (a *Server) rentalBoard(w http.ResponseWriter, r *http.Request) {
	board, err := a.store.RentalBoard(r.Context())
	reply(w, board, err)
}

func (a *Server) startRental(w http.ResponseWriter, r *http.Request) {
	in, valid := decode[struct {
		ItemID       int64  `json:"itemId"`
		Name         string `json:"name"`
		Room         string `json:"room"`
		DepositCents int64  `json:"depositCents"`
		DepositNote  string `json:"depositNote"`
	}](w, r)
	if !valid {
		return
	}
	reply(w, ok, a.store.StartRental(r.Context(), in.ItemID, in.Name, in.Room, in.DepositCents, in.DepositNote))
}

func (a *Server) rentalQuote(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(w, r)
	if !valid {
		return
	}
	q, err := a.store.QuoteReturn(r.Context(), id)
	reply(w, q, err)
}

func (a *Server) returnRental(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(w, r)
	if !valid {
		return
	}
	in, valid := decode[struct {
		Payment string `json:"payment"`
	}](w, r)
	if !valid {
		return
	}
	reply(w, ok, a.store.ReturnRental(r.Context(), id, in.Payment))
}

// Admin login -----------------------------------------------------------------

func (a *Server) pinHash(ctx context.Context) (string, error) {
	return a.store.Setting(ctx, KeyPIN, "")
}

func (a *Server) session(w http.ResponseWriter, r *http.Request) {
	hash, err := a.pinHash(r.Context())
	reply(w, map[string]bool{"configured": hash != "", "loggedIn": a.hasSession(r), "local": isLocal(r)}, err)
}

func (a *Server) setCookie(w http.ResponseWriter, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: value, Path: "/", MaxAge: maxAge,
		HttpOnly: true, SameSite: http.SameSiteStrictMode})
}

func validPIN(pin string) bool { return len(pin) >= 4 && len(pin) <= 32 }

// setup creates the first PIN. Only the bar PC itself may do this, so nobody on
// the dorm network can claim a fresh install.
func (a *Server) setup(w http.ResponseWriter, r *http.Request) {
	in, valid := decode[struct {
		PIN string `json:"pin"`
	}](w, r)
	if !valid {
		return
	}
	hash, err := a.pinHash(r.Context())
	if err != nil {
		fail(w, err)
		return
	}
	if hash != "" || !isLocal(r) {
		writeJSON(w, http.StatusForbidden, errBody("forbidden"))
		return
	}
	if !validPIN(in.PIN) {
		fail(w, store.ErrInvalid)
		return
	}
	hashed, err := hashPIN(in.PIN)
	if err != nil {
		fail(w, err)
		return
	}
	if err := a.store.SetSetting(r.Context(), KeyPIN, hashed); err != nil {
		fail(w, err)
		return
	}
	a.setCookie(w, a.auth.startSession(), int(sessionIdle.Seconds())*48)
	writeJSON(w, http.StatusOK, ok)
}

func (a *Server) login(w http.ResponseWriter, r *http.Request) {
	in, valid := decode[struct {
		PIN string `json:"pin"`
	}](w, r)
	if !valid {
		return
	}
	if a.auth.locked() {
		writeJSON(w, http.StatusTooManyRequests, errBody("locked"))
		return
	}
	hash, err := a.pinHash(r.Context())
	if err != nil {
		fail(w, err)
		return
	}
	if hash == "" || !checkPIN(in.PIN, hash) {
		a.auth.recordFailure()
		writeJSON(w, http.StatusUnauthorized, errBody("wrong_pin"))
		return
	}
	a.setCookie(w, a.auth.startSession(), int(sessionIdle.Seconds())*48)
	writeJSON(w, http.StatusOK, ok)
}

func (a *Server) logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookie); err == nil {
		a.auth.end(c.Value)
	}
	a.setCookie(w, "", -1)
	writeJSON(w, http.StatusOK, ok)
}

func (a *Server) changePIN(w http.ResponseWriter, r *http.Request) {
	in, valid := decode[struct {
		Current string `json:"current"`
		New     string `json:"new"`
	}](w, r)
	if !valid {
		return
	}
	hash, err := a.pinHash(r.Context())
	if err != nil {
		fail(w, err)
		return
	}
	if !checkPIN(in.Current, hash) {
		writeJSON(w, http.StatusUnauthorized, errBody("wrong_pin"))
		return
	}
	if !validPIN(in.New) {
		fail(w, store.ErrInvalid)
		return
	}
	hashed, err := hashPIN(in.New)
	if err != nil {
		fail(w, err)
		return
	}
	reply(w, ok, a.store.SetSetting(r.Context(), KeyPIN, hashed))
}

// Admin data ------------------------------------------------------------------

func (a *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	d, err := a.store.Dashboard(r.Context())
	reply(w, d, err)
}

func (a *Server) products(w http.ResponseWriter, r *http.Request) {
	p, err := a.store.Products(r.Context())
	reply(w, p, err)
}

func (a *Server) saveProduct(w http.ResponseWriter, r *http.Request) {
	in, valid := decode[store.Product](w, r)
	if !valid {
		return
	}
	p, err := a.store.SaveProduct(r.Context(), in)
	reply(w, p, err)
}

// image may be cached for a year: a new picture has a new version, and so a new URL.
func (a *Server) image(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(w, r)
	if !valid {
		return
	}
	mime, data, err := a.store.ProductImage(r.Context(), id)
	if err != nil {
		fail(w, err)
		return
	}
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Cache-Control", "private, max-age=31536000, immutable")
	if _, err := w.Write(data); err != nil {
		log.Printf("write image: %v", err)
	}
}

// pictureTypes excludes SVG on purpose: it can carry scripts, the others cannot.
var pictureTypes = map[string]bool{"image/webp": true, "image/jpeg": true, "image/png": true}

func (a *Server) saveImage(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(w, r)
	if !valid {
		return
	}
	data, err := io.ReadAll(r.Body)
	mime := http.DetectContentType(data)
	if err != nil || !pictureTypes[mime] {
		fail(w, store.ErrInvalid)
		return
	}
	version, err := a.store.SetProductImage(r.Context(), id, mime, data)
	reply(w, map[string]int64{"imageVersion": version}, err)
}

func (a *Server) deleteImage(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(w, r)
	if !valid {
		return
	}
	reply(w, ok, a.store.DeleteProductImage(r.Context(), id))
}

func (a *Server) categories(w http.ResponseWriter, r *http.Request) {
	c, err := a.store.Categories(r.Context())
	reply(w, c, err)
}

func (a *Server) saveCategory(w http.ResponseWriter, r *http.Request) {
	in, valid := decode[store.Category](w, r)
	if !valid {
		return
	}
	c, err := a.store.SaveCategory(r.Context(), in)
	reply(w, c, err)
}

func (a *Server) restocks(w http.ResponseWriter, r *http.Request) {
	list, err := a.store.Restocks(r.Context(), 30)
	reply(w, list, err)
}

func (a *Server) saveRestock(w http.ResponseWriter, r *http.Request) {
	in, valid := decode[struct {
		Supplier string              `json:"supplier"`
		Note     string              `json:"note"`
		Place    string              `json:"place"`
		Lines    []store.RestockLine `json:"lines"`
	}](w, r)
	if !valid {
		return
	}
	id, err := a.store.SaveRestock(r.Context(), in.Supplier, in.Note, in.Place, in.Lines)
	reply(w, map[string]int64{"id": id}, err)
}

func (a *Server) refills(w http.ResponseWriter, r *http.Request) {
	list, err := a.store.Refills(r.Context(), 30)
	reply(w, list, err)
}

func (a *Server) stocktakes(w http.ResponseWriter, r *http.Request) {
	list, err := a.store.Stocktakes(r.Context(), 50)
	reply(w, list, err)
}

func (a *Server) stocktakeLines(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(w, r)
	if !valid {
		return
	}
	lines, err := a.store.StocktakeLines(r.Context(), id)
	reply(w, lines, err)
}

func (a *Server) listShifts(w http.ResponseWriter, r *http.Request) {
	list, err := a.store.ListShifts(r.Context(), 100)
	reply(w, list, err)
}

func (a *Server) shiftDetail(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(w, r)
	if !valid {
		return
	}
	sh, err := a.store.GetShift(r.Context(), id)
	if err != nil {
		fail(w, err)
		return
	}
	sales, err := a.store.ShiftSales(r.Context(), id)
	if err != nil {
		fail(w, err)
		return
	}
	t, err := a.store.ShiftTotals(r.Context(), *sh)
	reply(w, map[string]any{"shift": sh, "sales": sales, "totals": t}, err)
}

func (a *Server) cash(w http.ResponseWriter, r *http.Request) {
	box, err := a.store.CashInBox(r.Context())
	if err != nil {
		fail(w, err)
		return
	}
	moves, err := a.store.CashMoves(r.Context(), 100)
	reply(w, map[string]any{"boxCents": box, "moves": moves}, err)
}

func (a *Server) addCashMove(w http.ResponseWriter, r *http.Request) {
	in, valid := decode[struct {
		Kind        string `json:"kind"`
		AmountCents int64  `json:"amountCents"`
		Note        string `json:"note"`
	}](w, r)
	if !valid {
		return
	}
	reply(w, ok, a.store.AddCashMove(r.Context(), in.Kind, in.AmountCents, in.Note))
}

func (a *Server) listSales(w http.ResponseWriter, r *http.Request) {
	from, to, valid := dateRange(w, r)
	if !valid {
		return
	}
	sales, err := a.store.SalesBetween(r.Context(), from, to, 500)
	reply(w, sales, err)
}

func (a *Server) rentalItems(w http.ResponseWriter, r *http.Request) {
	items, err := a.store.RentalItems(r.Context())
	reply(w, items, err)
}

func (a *Server) saveRentalItem(w http.ResponseWriter, r *http.Request) {
	in, valid := decode[store.RentalItem](w, r)
	if !valid {
		return
	}
	item, err := a.store.SaveRentalItem(r.Context(), in)
	reply(w, item, err)
}

func (a *Server) rentalHistory(w http.ResponseWriter, r *http.Request) {
	list, err := a.store.RentalHistory(r.Context(), 100)
	reply(w, list, err)
}

func (a *Server) report(w http.ResponseWriter, r *http.Request) {
	from, to, valid := dateRange(w, r)
	if !valid {
		return
	}
	rep, err := a.store.Report(r.Context(), from, to)
	reply(w, rep, err)
}

func (a *Server) getSettings(w http.ResponseWriter, r *http.Request) {
	s := a.publicSettings(r.Context())
	dir, err := a.store.Setting(r.Context(), keyBackupDir, "")
	s["backupDir"] = dir
	reply(w, s, err)
}

func (a *Server) saveSettings(w http.ResponseWriter, r *http.Request) {
	in, valid := decode[struct {
		BarName      string `json:"barName"`
		Language     string `json:"language"`
		OtherPayment string `json:"otherPayment"`
		BackupDir    string `json:"backupDir"`
	}](w, r)
	if !valid {
		return
	}
	if strings.TrimSpace(in.BarName) == "" || (in.Language != "de" && in.Language != "en") {
		fail(w, store.ErrInvalid)
		return
	}
	for key, value := range map[string]string{
		keyBarName:      strings.TrimSpace(in.BarName),
		keyLanguage:     in.Language,
		keyOtherPayment: strings.TrimSpace(in.OtherPayment),
		keyBackupDir:    strings.TrimSpace(in.BackupDir),
	} {
		if err := a.store.SetSetting(r.Context(), key, value); err != nil {
			fail(w, err)
			return
		}
	}
	writeJSON(w, http.StatusOK, ok)
}

func (a *Server) backups(w http.ResponseWriter, _ *http.Request) {
	list, err := a.store.Backups()
	reply(w, list, err)
}

func (a *Server) backupNow(w http.ResponseWriter, r *http.Request) {
	dir, err := a.store.Setting(r.Context(), keyBackupDir, "")
	if err != nil {
		fail(w, err)
		return
	}
	if err := a.store.Backup(r.Context(), dir); err != nil {
		log.Printf("backup: %v", err)
		writeJSON(w, http.StatusInternalServerError, errBody("backup_failed"))
		return
	}
	writeJSON(w, http.StatusOK, ok)
}

func (a *Server) runBackup() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	dir, err := a.store.Setting(ctx, keyBackupDir, "")
	if err == nil {
		err = a.store.Backup(ctx, dir)
	}
	if err != nil {
		log.Printf("backup: %v", err)
	}
}

// DailyBackup makes sure one backup exists per day, even in a week without shifts.
func (a *Server) DailyBackup(ctx context.Context) {
	for {
		if !a.store.BackedUpToday() {
			a.runBackup()
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Hour):
		}
	}
}

// Static files ----------------------------------------------------------------

func (a *Server) static() http.Handler {
	files := http.FileServerFS(a.ui)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name == "" {
			name = "index.html"
		}
		if strings.HasPrefix(r.URL.Path, "/api/") {
			writeJSON(w, http.StatusNotFound, errBody("not_found"))
			return
		}
		// Unknown paths are client-side routes such as /admin.
		if _, err := fs.Stat(a.ui, name); err != nil {
			name = "index.html"
			r = r.Clone(r.Context())
			r.URL.Path = "/"
		}
		if strings.HasPrefix(name, "assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		files.ServeHTTP(w, r)
	})
}
