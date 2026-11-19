package store

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"
)

var ctx = context.Background()

func newStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

// at pins the store clock.
func at(s *Store, tm time.Time) { s.now = func() time.Time { return tm } }

func product(t *testing.T, s *Store, name string) Product {
	t.Helper()
	all, err := s.Products(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range all {
		if p.Name == name {
			return p
		}
	}
	t.Fatalf("no product %q", name)
	return Product{}
}

func open(t *testing.T, s *Store, float int64) *Shift {
	t.Helper()
	sh, err := s.OpenShift(ctx, "214", "Jonas", float)
	if err != nil {
		t.Fatal(err)
	}
	return sh
}

func TestSaleLowersStockAndRaisesExpectedCash(t *testing.T) {
	s := newStore(t)
	sh := open(t, s, 5000)
	beer := product(t, s, "Früh Kölsch")

	sale, err := s.CreateSale(ctx, []LineInput{{beer.ID, 3}}, PayCash)
	if err != nil {
		t.Fatal(err)
	}
	if sale.TotalCents != 3*beer.PriceCents {
		t.Errorf("total = %d", sale.TotalCents)
	}
	if got := product(t, s, "Früh Kölsch").BarStock; got != beer.BarStock-3 {
		t.Errorf("stock = %d, want %d", got, beer.BarStock-3)
	}
	totals, err := s.ShiftTotals(ctx, *sh)
	if err != nil {
		t.Fatal(err)
	}
	if want := 5000 + 3*beer.PriceCents; totals.ExpectedCents != want {
		t.Errorf("expected cash = %d, want %d", totals.ExpectedCents, want)
	}
}

func TestNonCashSaleDoesNotTouchTheBox(t *testing.T) {
	s := newStore(t)
	sh := open(t, s, 5000)
	cola := product(t, s, "Coca-Cola")
	if _, err := s.CreateSale(ctx, []LineInput{{cola.ID, 2}}, PayOther); err != nil {
		t.Fatal(err)
	}
	totals, _ := s.ShiftTotals(ctx, *sh)
	if totals.ExpectedCents != 5000 {
		t.Errorf("expected cash = %d, want 5000", totals.ExpectedCents)
	}
	if totals.OtherCents != 2*cola.PriceCents {
		t.Errorf("other = %d", totals.OtherCents)
	}
}

func TestSaleNeedsAnOpenShift(t *testing.T) {
	s := newStore(t)
	beer := product(t, s, "Früh Kölsch")
	if _, err := s.CreateSale(ctx, []LineInput{{beer.ID, 1}}, PayCash); !errors.Is(err, ErrNoOpenShift) {
		t.Errorf("err = %v", err)
	}
}

func TestOnlyOneShiftAtATime(t *testing.T) {
	s := newStore(t)
	open(t, s, 0)
	if _, err := s.OpenShift(ctx, "108", "Tim", 0); !errors.Is(err, ErrShiftOpen) {
		t.Errorf("err = %v", err)
	}
}

func TestSaleCanTakeStockBelowZero(t *testing.T) {
	s := newStore(t)
	open(t, s, 0)
	bar := product(t, s, "Schokowaffel")
	if bar.BarStock != 0 {
		t.Fatalf("seed changed: stock %d", bar.BarStock)
	}
	if _, err := s.CreateSale(ctx, []LineInput{{bar.ID, 1}}, PayCash); err != nil {
		t.Fatalf("a stock number must never block a sale: %v", err)
	}
	if got := product(t, s, "Schokowaffel").BarStock; got != -1 {
		t.Errorf("stock = %d", got)
	}
}

func TestPriceChangeDoesNotRewriteHistory(t *testing.T) {
	s := newStore(t)
	open(t, s, 0)
	beer := product(t, s, "Früh Kölsch")
	sale, _ := s.CreateSale(ctx, []LineInput{{beer.ID, 1}}, PayCash)
	beer.PriceCents = 999
	if _, err := s.SaveProduct(ctx, beer); err != nil {
		t.Fatal(err)
	}
	again, _ := s.GetSale(ctx, sale.ID)
	if again.Lines[0].UnitPriceCents == 999 || again.TotalCents == 999 {
		t.Errorf("history changed: %+v", again)
	}
}

func TestVoidReversesEverythingOnce(t *testing.T) {
	s := newStore(t)
	sh := open(t, s, 1000)
	beer := product(t, s, "Früh Kölsch")
	sale, _ := s.CreateSale(ctx, []LineInput{{beer.ID, 2}}, PayCash)

	if _, err := s.VoidSale(ctx, sale.ID, " "); !errors.Is(err, ErrReasonRequired) {
		t.Errorf("empty reason: %v", err)
	}
	void, err := s.VoidSale(ctx, sale.ID, "Falsch getippt")
	if err != nil {
		t.Fatal(err)
	}
	if void.TotalCents != -sale.TotalCents {
		t.Errorf("void total = %d", void.TotalCents)
	}
	if got := product(t, s, "Früh Kölsch").BarStock; got != beer.BarStock {
		t.Errorf("stock not restored: %d", got)
	}
	totals, _ := s.ShiftTotals(ctx, *sh)
	if totals.ExpectedCents != 1000 {
		t.Errorf("expected cash = %d", totals.ExpectedCents)
	}
	if totals.SalesCount != 1 {
		t.Errorf("a reversal must not count as a sale: %d", totals.SalesCount)
	}
	if _, err := s.VoidSale(ctx, sale.ID, "nochmal"); !errors.Is(err, ErrAlreadyVoided) {
		t.Errorf("double void: %v", err)
	}
	if _, err := s.VoidSale(ctx, void.ID, "zurück"); !errors.Is(err, ErrNotVoidable) {
		t.Errorf("void of a void: %v", err)
	}
}

func TestVoidAfterCloseChangesTheNextExpectedFloat(t *testing.T) {
	s := newStore(t)
	open(t, s, 0)
	beer := product(t, s, "Früh Kölsch")
	sale, _ := s.CreateSale(ctx, []LineInput{{beer.ID, 1}}, PayCash)
	if _, err := s.CloseShift(ctx, beer.PriceCents, "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := s.VoidSale(ctx, sale.ID, "Erstattet"); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.ExpectedFloat(ctx); got != 0 {
		t.Errorf("box = %d, want 0", got)
	}
}

func TestShiftCloseAndHandover(t *testing.T) {
	s := newStore(t)
	base := time.Date(2026, 11, 6, 20, 0, 0, 0, time.Local)
	at(s, base)
	open(t, s, 5000)
	beer := product(t, s, "Früh Kölsch")
	s.CreateSale(ctx, []LineInput{{beer.ID, 2}}, PayCash)

	at(s, base.Add(4*time.Hour))
	closed, err := s.CloseShift(ctx, 5200, "fehlt was", "")
	if err != nil {
		t.Fatal(err)
	}
	if *closed.ExpectedCents != 5000+2*beer.PriceCents || *closed.CountedCents != 5200 {
		t.Errorf("closed = %+v", closed)
	}

	// The OBK takes money out between shifts.
	at(s, base.Add(14*time.Hour))
	if err := s.AddCashMove(ctx, "withdrawal", 3000, "Bank"); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.ExpectedFloat(ctx); got != 2200 {
		t.Errorf("expected float = %d, want 2200", got)
	}
	at(s, base.Add(24*time.Hour))
	next, err := s.OpenShift(ctx, "108", "Tim", 2000)
	if err != nil {
		t.Fatal(err)
	}
	if next.ExpectedFloatCents != 2200 {
		t.Errorf("handover expected = %d", next.ExpectedFloatCents)
	}
}

func TestBusinessDayRunsPastMidnight(t *testing.T) {
	late := time.Date(2026, 11, 7, 1, 30, 0, 0, time.Local)
	if got := businessDay(late); got != "2026-11-06" {
		t.Errorf("01:30 belongs to the evening before, got %s", got)
	}
	morning := time.Date(2026, 11, 7, 6, 30, 0, 0, time.Local)
	if got := businessDay(morning); got != "2026-11-07" {
		t.Errorf("06:30 starts a new day, got %s", got)
	}
}

func TestStocktakeCorrectsStockAndKeepsTheDifference(t *testing.T) {
	s := newStore(t)
	beer := product(t, s, "Früh Kölsch")
	id, err := s.SaveStocktake(ctx, "full", PlaceBar, "214", "Jonas", []CountLine{{beer.ID, beer.BarStock - 4}})
	if err != nil {
		t.Fatal(err)
	}
	if got := product(t, s, "Früh Kölsch").BarStock; got != beer.BarStock-4 {
		t.Errorf("stock = %d", got)
	}
	lines, _ := s.StocktakeLines(ctx, id)
	if len(lines) != 1 || lines[0].Expected != beer.BarStock || lines[0].Counted != beer.BarStock-4 {
		t.Errorf("lines = %+v", lines)
	}
	takes, _ := s.Stocktakes(ctx, 10)
	if takes[0].Diffs != 1 || takes[0].DiffCents != -4*beer.CostCents {
		t.Errorf("summary = %+v", takes[0])
	}
}

func TestTillNeverRevealsTheStockNumber(t *testing.T) {
	s := newStore(t)
	// Three of twelve is a quarter of what the bar should hold.
	if _, err := s.db.ExecContext(ctx, "UPDATE products SET bar_stock = 3 WHERE name = 'Durstlöscher'"); err != nil {
		t.Fatal(err)
	}
	products, err := s.TillProducts(ctx)
	if err != nil {
		t.Fatal(err)
	}
	levels := map[string]string{}
	for _, p := range products {
		levels[p.Name] = p.Level
	}
	if levels["Schokowaffel"] != "empty" || levels["Durstlöscher"] != "low" || levels["Früh Kölsch"] != "ok" {
		t.Errorf("levels = %v", levels)
	}
}

func TestRestockAddsStockAndUpdatesCost(t *testing.T) {
	s := newStore(t)
	beer := product(t, s, "Früh Kölsch")
	id, err := s.SaveRestock(ctx, "Getränkemarkt", "", PlaceStorage, []RestockLine{{ProductID: beer.ID, Qty: 40, UnitCostCents: 90}})
	if err != nil {
		t.Fatal(err)
	}
	after := product(t, s, "Früh Kölsch")
	if after.StorageStock != beer.StorageStock+40 || after.BarStock != beer.BarStock || after.CostCents != 90 {
		t.Errorf("after = %+v", after)
	}
	list, _ := s.Restocks(ctx, 5)
	if list[0].ID != id || list[0].TotalCostCents != 3600 || len(list[0].Lines) != 1 {
		t.Errorf("restock = %+v", list[0])
	}
}

func TestRentalFlow(t *testing.T) {
	s := newStore(t)
	base := time.Date(2026, 11, 6, 20, 0, 0, 0, time.Local)
	at(s, base)
	sh := open(t, s, 1000)
	// The bar lends for a deposit only. An hourly fee stays possible for anything the OBK adds.
	queue, err := s.SaveRentalItem(ctx, RentalItem{Name: "Kicker-Set", FeeCents: 50, FeePer: "hour", DepositCents: 500, Active: true})
	if err != nil {
		t.Fatal(err)
	}

	if err := s.StartRental(ctx, queue.ID, "Tim", "108", 500, ""); err != nil {
		t.Fatal(err)
	}
	if err := s.StartRental(ctx, queue.ID, "Lena", "201", 0, ""); !errors.Is(err, ErrItemRentedOut) {
		t.Errorf("double rental: %v", err)
	}
	totals, _ := s.ShiftTotals(ctx, *sh)
	if totals.ExpectedCents != 1500 {
		t.Errorf("deposit not in box: %d", totals.ExpectedCents)
	}

	board, _ := s.RentalBoard(ctx)
	var lent *Rental
	for _, item := range board {
		if item.ID == queue.ID {
			lent = item.Rental
		}
	}
	if lent == nil || lent.RenterName != "Tim" {
		t.Fatalf("board = %+v", board)
	}

	at(s, base.Add(90*time.Minute))
	quote, err := s.QuoteReturn(ctx, lent.ID)
	if err != nil {
		t.Fatal(err)
	}
	if quote.FeeCents != 100 {
		t.Errorf("90 minutes is 2 started hours: fee = %d", quote.FeeCents)
	}
	if err := s.ReturnRental(ctx, lent.ID, PayCash); err != nil {
		t.Fatal(err)
	}
	totals, _ = s.ShiftTotals(ctx, *sh)
	if totals.ExpectedCents != 1100 {
		t.Errorf("box after return = %d, want float + fee", totals.ExpectedCents)
	}
	if err := s.ReturnRental(ctx, lent.ID, PayCash); !errors.Is(err, ErrNotFound) {
		t.Errorf("double return: %v", err)
	}
	if err := s.StartRental(ctx, queue.ID, "Lena", "201", 0, "Ausweis"); err != nil {
		t.Errorf("item should be free again: %v", err)
	}
}

func TestCueIsLentForADepositOnly(t *testing.T) {
	s := newStore(t)
	base := time.Date(2026, 11, 6, 20, 0, 0, 0, time.Local)
	at(s, base)
	sh := open(t, s, 1000)
	items, _ := s.RentalItems(ctx)
	var cue RentalItem
	for _, item := range items {
		if item.Name == "Billardqueue 1" {
			cue = item
		}
	}
	if cue.FeeCents != 0 || cue.DepositCents != 500 {
		t.Fatalf("cue = %+v", cue)
	}
	if err := s.StartRental(ctx, cue.ID, "Tim", "108", cue.DepositCents, ""); err != nil {
		t.Fatal(err)
	}
	totals, _ := s.ShiftTotals(ctx, *sh)
	if totals.ExpectedCents != 1500 {
		t.Errorf("deposit not in box: %d", totals.ExpectedCents)
	}
	history, _ := s.RentalHistory(ctx, 1)
	at(s, base.Add(3*time.Hour))
	quote, err := s.QuoteReturn(ctx, history[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if quote.FeeCents != 0 || quote.DepositCents != 500 {
		t.Errorf("quote = %+v", quote)
	}
	if err := s.ReturnRental(ctx, history[0].ID, PayCash); err != nil {
		t.Fatal(err)
	}
	totals, _ = s.ShiftTotals(ctx, *sh)
	if totals.ExpectedCents != 1000 {
		t.Errorf("box after return = %d, want the float", totals.ExpectedCents)
	}
	r, err := s.Report(ctx, "2026-11-06", "2026-11-06")
	if err != nil {
		t.Fatal(err)
	}
	if r.RevenueCents != 0 {
		t.Errorf("a deposit is not revenue: %d", r.RevenueCents)
	}
}

func TestDepositReturnProductGivesCashBack(t *testing.T) {
	s := newStore(t)
	sh := open(t, s, 1000)
	back := product(t, s, "Pfand Flasche")
	if back.PriceCents != -15 {
		t.Fatalf("a returned bottle pays out 0,15 €, not %d cents", back.PriceCents)
	}
	if _, err := s.CreateSale(ctx, []LineInput{{back.ID, 4}}, PayCash); err != nil {
		t.Fatal(err)
	}
	totals, _ := s.ShiftTotals(ctx, *sh)
	if totals.ExpectedCents != 940 {
		t.Errorf("box = %d, want 940", totals.ExpectedCents)
	}
}

func TestReportCancelsVoidsAndShowsLosses(t *testing.T) {
	s := newStore(t)
	base := time.Date(2026, 11, 6, 20, 0, 0, 0, time.Local)
	at(s, base)
	open(t, s, 0)
	beer := product(t, s, "Früh Kölsch")
	cola := product(t, s, "Coca-Cola")
	keep, _ := s.CreateSale(ctx, []LineInput{{beer.ID, 2}}, PayCash)
	undo, _ := s.CreateSale(ctx, []LineInput{{cola.ID, 1}}, PayOther)
	if _, err := s.VoidSale(ctx, undo.ID, "Fehler"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SaveStocktake(ctx, "full", PlaceBar, "214", "Jonas", []CountLine{{cola.ID, cola.BarStock - 2}}); err != nil {
		t.Fatal(err)
	}
	r, err := s.Report(ctx, "2026-11-06", "2026-11-06")
	if err != nil {
		t.Fatal(err)
	}
	if r.RevenueCents != keep.TotalCents || r.OtherCents != 0 {
		t.Errorf("revenue = %d other = %d", r.RevenueCents, r.OtherCents)
	}
	if len(r.Products) != 1 || r.Products[0].Name != beer.Name || r.Products[0].Qty != 2 {
		t.Errorf("products = %+v", r.Products)
	}
	if r.CostCents != 2*beer.CostCents {
		t.Errorf("cost = %d", r.CostCents)
	}
	if len(r.Losses) != 1 || r.Losses[0].Units != -2 {
		t.Errorf("losses = %+v", r.Losses)
	}
}

func TestBackupWritesACopyAndPrunes(t *testing.T) {
	s := newStore(t)
	base := time.Date(2026, 11, 6, 20, 0, 0, 0, time.Local)
	for i := 0; i < keepBackups+3; i++ {
		at(s, base.Add(time.Duration(i)*time.Minute))
		if err := s.Backup(ctx, ""); err != nil {
			t.Fatal(err)
		}
	}
	list, err := s.Backups()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != keepBackups {
		t.Errorf("kept %d backups", len(list))
	}
	if !s.BackedUpToday() {
		t.Error("today's backup not found")
	}
}

func TestDatabaseSurvivesReopen(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.OpenShift(ctx, "214", "Jonas", 100); err != nil {
		t.Fatal(err)
	}
	s.Close()
	again, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer again.Close()
	sh, err := again.CurrentShift(ctx)
	if err != nil || sh == nil || sh.WorkerName != "Jonas" {
		t.Errorf("shift lost: %+v, %v", sh, err)
	}
	products, _ := again.Products(ctx)
	if len(products) != len(catalog) {
		t.Errorf("seed ran twice or not at all: %d", len(products))
	}
}

func TestDashboardAndDemoHistory(t *testing.T) {
	s := newStore(t)
	if err := s.SeedDemoHistory(ctx); err != nil {
		t.Fatal(err)
	}
	d, err := s.Dashboard(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Series) != 30 || len(d.TopProducts) == 0 || len(d.RecentShifts) == 0 {
		t.Errorf("dashboard = %+v", d)
	}
	if d.Month < 0 || d.CashInBox <= 0 {
		t.Errorf("month %d box %d", d.Month, d.CashInBox)
	}
}

func TestProductImageLifecycle(t *testing.T) {
	s := newStore(t)
	id := product(t, s, "Pfand Flasche").ID
	if v := product(t, s, "Pfand Flasche").ImageVersion; v == 0 {
		t.Fatal("the catalog picture is missing")
	}

	at(s, time.UnixMilli(1000))
	if _, err := s.SetProductImage(ctx, id, "image/webp", []byte("one")); err != nil {
		t.Fatal(err)
	}
	at(s, time.UnixMilli(2000))
	version, err := s.SetProductImage(ctx, id, "image/png", []byte("two"))
	if err != nil {
		t.Fatal(err)
	}
	if got := product(t, s, "Pfand Flasche").ImageVersion; got != version || got != 2000 {
		t.Errorf("version = %d, want %d", got, version)
	}
	mime, data, err := s.ProductImage(ctx, id)
	if err != nil || mime != "image/png" || string(data) != "two" {
		t.Errorf("image = %q %q %v", mime, data, err)
	}

	if err := s.DeleteProductImage(ctx, id); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.ProductImage(ctx, id); !errors.Is(err, ErrNotFound) {
		t.Errorf("after delete: %v", err)
	}
	if got := product(t, s, "Pfand Flasche").ImageVersion; got != 0 {
		t.Errorf("version after delete = %d", got)
	}
}

func TestProductImageLimits(t *testing.T) {
	s := newStore(t)
	id := product(t, s, "Pfand Flasche").ID
	if _, err := s.SetProductImage(ctx, id, "image/webp", make([]byte, MaxImageBytes+1)); !errors.Is(err, ErrInvalid) {
		t.Errorf("too large: %v", err)
	}
	if _, err := s.SetProductImage(ctx, id, "image/webp", nil); !errors.Is(err, ErrInvalid) {
		t.Errorf("empty: %v", err)
	}
	if _, err := s.SetProductImage(ctx, 99999, "image/webp", []byte("x")); !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown product: %v", err)
	}
}

func TestEveryCatalogPictureExists(t *testing.T) {
	for _, p := range catalog {
		if p.image == "" {
			continue
		}
		if _, err := catalogImages.ReadFile("catalog/" + p.image + ".webp"); err != nil {
			t.Errorf("%s: %v", p.name, err)
		}
	}
}

func TestNewDatabaseHasPicturesAndBarcodes(t *testing.T) {
	s := newStore(t)
	products, err := s.Products(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(products) != len(catalog) {
		t.Errorf("products = %d, catalog = %d", len(products), len(catalog))
	}
	for _, p := range products {
		if p.ImageVersion == 0 {
			t.Errorf("%s has no picture", p.Name)
		}
	}
}

func TestDeliveriesArriveInStorageAndSalesLeaveTheBar(t *testing.T) {
	s := newStore(t)
	beer := product(t, s, "Früh Kölsch")
	if _, err := s.SaveRestock(ctx, "", "", PlaceBar, []RestockLine{{ProductID: beer.ID, Qty: 5}}); err != nil {
		t.Fatal(err)
	}
	after := product(t, s, "Früh Kölsch")
	if after.BarStock != beer.BarStock+5 || after.StorageStock != beer.StorageStock {
		t.Errorf("delivery straight to the bar: bar %d, storage %d", after.BarStock, after.StorageStock)
	}
	if _, err := s.SaveRestock(ctx, "", "", "keller", []RestockLine{{ProductID: beer.ID, Qty: 5}}); !errors.Is(err, ErrInvalid) {
		t.Errorf("unknown place: %v", err)
	}
	list, _ := s.Restocks(ctx, 1)
	if list[0].Place != PlaceBar {
		t.Errorf("place = %q", list[0].Place)
	}
}

func TestRefillMovesGoodsAndKeepsTheTotal(t *testing.T) {
	s := newStore(t)
	beer := product(t, s, "Früh Kölsch")
	if err := s.SaveRefill(ctx, []RefillLine{{beer.ID, 6}}); err != nil {
		t.Fatal(err)
	}
	after := product(t, s, "Früh Kölsch")
	if after.BarStock != beer.BarStock+6 || after.StorageStock != beer.StorageStock-6 {
		t.Errorf("bar %d, storage %d", after.BarStock, after.StorageStock)
	}
	list, err := s.Refills(ctx, 5)
	if err != nil || len(list) != 1 || list[0].Name != beer.Name || list[0].Qty != 6 {
		t.Errorf("history = %+v, %v", list, err)
	}
	var moved int64
	if err := s.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(delta), 0) FROM stock_moves WHERE kind = 'refill'").Scan(&moved); err != nil || moved != 0 {
		t.Errorf("a refill must add up to nothing: %d, %v", moved, err)
	}
}

func TestRefillIsNeverBlockedByTheStorageFigure(t *testing.T) {
	s := newStore(t)
	bar := product(t, s, "Schokowaffel")
	if err := s.SaveRefill(ctx, []RefillLine{{bar.ID, 3}}); err != nil {
		t.Fatal(err)
	}
	if got := product(t, s, "Schokowaffel"); got.BarStock != 3 || got.StorageStock != -3 {
		t.Errorf("bar %d, storage %d", got.BarStock, got.StorageStock)
	}
}

func TestRefillRefusesNonsense(t *testing.T) {
	s := newStore(t)
	beer := product(t, s, "Früh Kölsch")
	pfand := product(t, s, "Pfand Flasche")
	for _, lines := range [][]RefillLine{nil, {{beer.ID, 0}}, {{beer.ID, -2}}, {{pfand.ID, 1}}} {
		if err := s.SaveRefill(ctx, lines); !errors.Is(err, ErrInvalid) {
			t.Errorf("%+v: %v", lines, err)
		}
	}
	if err := s.SaveRefill(ctx, []RefillLine{{99999, 1}}); !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown product: %v", err)
	}
}

func TestCountingOnePlaceLeavesTheOtherAlone(t *testing.T) {
	s := newStore(t)
	beer := product(t, s, "Früh Kölsch")
	id, err := s.SaveStocktake(ctx, "full", PlaceStorage, "214", "Jonas", []CountLine{{beer.ID, beer.StorageStock + 3}})
	if err != nil {
		t.Fatal(err)
	}
	after := product(t, s, "Früh Kölsch")
	if after.StorageStock != beer.StorageStock+3 || after.BarStock != beer.BarStock {
		t.Errorf("bar %d, storage %d", after.BarStock, after.StorageStock)
	}
	lines, _ := s.StocktakeLines(ctx, id)
	if len(lines) != 1 || lines[0].Expected != beer.StorageStock {
		t.Errorf("lines = %+v", lines)
	}
	takes, _ := s.Stocktakes(ctx, 1)
	if takes[0].Place != PlaceStorage {
		t.Errorf("place = %q", takes[0].Place)
	}
	if _, err := s.SaveStocktake(ctx, "full", "", "214", "Jonas", []CountLine{{beer.ID, 1}}); !errors.Is(err, ErrInvalid) {
		t.Errorf("missing place: %v", err)
	}
}

func TestRefillSuggestionWaitsUntilTheBarIsHalfEmpty(t *testing.T) {
	for _, c := range []struct {
		name                 string
		bar, storage, target int64
		want                 int64
	}{
		{"full bar", 20, 40, 20, 0},
		{"one sold", 19, 40, 20, 0},
		{"half empty", 10, 40, 20, 10},
		{"empty", 0, 40, 20, 20},
		{"storage runs out", 2, 5, 20, 5},
		{"nothing to fetch", 2, 0, 20, 0},
		{"storage below zero", 2, -3, 20, 0},
		{"no target", 0, 40, 0, 0},
	} {
		p := Product{TrackStock: true, BarStock: c.bar, StorageStock: c.storage, BarTarget: c.target}
		if got := p.refillQty(); got != c.want {
			t.Errorf("%s: refill %d, want %d", c.name, got, c.want)
		}
	}
	if got := (Product{BarTarget: 20, StorageStock: 40}).refillQty(); got != 0 {
		t.Errorf("untracked product: %d", got)
	}
}

func TestBarLevelDrivesTheKeys(t *testing.T) {
	for _, c := range []struct {
		bar, storage int64
		want         string
	}{
		{20, 0, "ok"}, {6, 0, "ok"}, {5, 100, "low"}, {1, 100, "low"}, {0, 100, "empty"}, {-1, 100, "empty"},
	} {
		p := Product{TrackStock: true, BarStock: c.bar, StorageStock: c.storage, BarTarget: 20}
		if got := p.level(); got != c.want {
			t.Errorf("bar %d: level %q, want %q", c.bar, got, c.want)
		}
	}
}

func TestStockValueAndBuyListCountBothPlaces(t *testing.T) {
	s := newStore(t)
	beer := product(t, s, "Früh Kölsch")
	before, err := s.Dashboard(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SaveRefill(ctx, []RefillLine{{beer.ID, 4}}); err != nil {
		t.Fatal(err)
	}
	after, err := s.Dashboard(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if after.StockValue != before.StockValue {
		t.Errorf("a refill changed the stock value: %d to %d", before.StockValue, after.StockValue)
	}
	for _, l := range after.LowStock {
		if l.Name == beer.Name {
			t.Errorf("%d in total is above the minimum of %d", beer.BarStock+beer.StorageStock, beer.MinStock)
		}
	}
}

// seedNamed is the catalog line for a product.
func seedNamed(t *testing.T, name string) seedProduct {
	t.Helper()
	for _, p := range catalog {
		if p.name == name {
			return p
		}
	}
	t.Fatalf("%s is not in the catalog", name)
	return seedProduct{}
}

func hasProduct(t *testing.T, s *Store, name string) bool {
	t.Helper()
	all, err := s.Products(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return slices.ContainsFunc(all, func(p Product) bool { return p.Name == name })
}

// addDepositReturn adds a key with a negative price, as the OBK could make one.
func addDepositReturn(t *testing.T, s *Store) {
	t.Helper()
	if _, err := s.db.ExecContext(ctx, `INSERT INTO products(category_id, name, price_cents, track_stock)
		SELECT id, 'Pfand zurück', -15, 0 FROM categories WHERE name = 'Pfand'`); err != nil {
		t.Fatal(err)
	}
}

// countedSince adds up the count moves of a product: what the update booked to reach the list's figure.
func countedSince(t *testing.T, s *Store, id int64) int64 {
	t.Helper()
	var sum int64
	if err := s.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(delta), 0) FROM stock_moves WHERE product_id = ? AND kind = 'count'", id).Scan(&sum); err != nil {
		t.Fatal(err)
	}
	return sum
}

// A version 1 till gets the bar's real range. A product found under an earlier name keeps its id and
// history but takes the list's name, price and count, and a barcode in the list moves to its product.
func TestUpgradeFromVersionOneTakesTheRealRange(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	old, fanta := seedNamed(t, "Beck's"), seedNamed(t, "Fanta")
	id := product(t, s, old.name).ID
	asVersionEight(t, s)
	for _, q := range []struct {
		sql  string
		args []any
	}{
		{"DELETE FROM product_images", nil},
		{"ALTER TABLE products DROP COLUMN image_version", nil},
		{"DROP TABLE product_images", nil},
		{"DROP TABLE refills", nil},
		{"ALTER TABLE stock_moves DROP COLUMN place", nil},
		{"ALTER TABLE stocktakes DROP COLUMN place", nil},
		{"ALTER TABLE restocks DROP COLUMN place", nil},
		{"ALTER TABLE products DROP COLUMN storage_stock", nil},
		{"ALTER TABLE products DROP COLUMN bar_target", nil},
		{"ALTER TABLE products RENAME COLUMN bar_stock TO stock", nil},
		{"UPDATE products SET name = ?, stock = 33, price_cents = 999 WHERE name = ?", []any{old.was[0], old.name}},
		{"DELETE FROM products WHERE name = ?", []any{fanta.name}},
		{"UPDATE products SET barcode = ? WHERE name = 'Pfand Flasche'", []any{fanta.code}},
		{"INSERT INTO products(category_id, name, price_cents, track_stock, stock) VALUES (1, 'Krombacher Pils 0,5 l', 150, 1, 20)", nil},
		{"ALTER TABLE shifts DROP COLUMN handover", nil},
		{"PRAGMA user_version = 1", nil},
	} {
		if _, err := s.db.ExecContext(ctx, q.sql, q.args...); err != nil {
			s.Close()
			t.Fatalf("%s: %v", q.sql, err)
		}
	}
	s.Close()

	s, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	got := product(t, s, old.name)
	if got.ID != id || got.PriceCents != old.price || got.BarStock != old.stock || got.StorageStock != 0 || !got.Active {
		t.Errorf("%s = %+v", old.name, got)
	}
	if got.ImageVersion == 0 {
		t.Error("renamed product has no picture")
	}
	if got := countedSince(t, s, id); got != old.stock-33 {
		t.Errorf("the new figure is booked as %d, want %d", got, old.stock-33)
	}
	if got := product(t, s, fanta.name); got.ImageVersion == 0 || !slices.Equal(got.Barcodes, []string{fanta.code}) {
		t.Errorf("%s = %+v", fanta.name, got)
	}
	if got := product(t, s, "Pfand Flasche").Barcodes; len(got) != 0 {
		t.Errorf("Pfand Flasche kept the code %v", got)
	}
	if hasProduct(t, s, "Krombacher Pils 0,5 l") {
		t.Error("a product the list does not have must be gone")
	}

	var version int
	if err := s.db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil || version != schemaVersion {
		t.Errorf("user_version = %d, %v", version, err)
	}
}

func TestUpgradeFromVersionTwoKeepsTheOldMoves(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	beer := product(t, s, "Hansa Pils")
	asVersionEight(t, s)
	for _, q := range []string{
		"DROP TABLE refills",
		"ALTER TABLE stock_moves DROP COLUMN place",
		"ALTER TABLE stocktakes DROP COLUMN place",
		"ALTER TABLE restocks DROP COLUMN place",
		"ALTER TABLE products DROP COLUMN storage_stock",
		"ALTER TABLE products DROP COLUMN bar_target",
		"ALTER TABLE products RENAME COLUMN bar_stock TO stock",
		"UPDATE products SET stock = 53 WHERE name = 'Hansa Pils'",
		"INSERT INTO stock_moves(product_id, delta, kind, created_at) VALUES (1, 5, 'restock', 1)",
		"ALTER TABLE shifts DROP COLUMN handover",
		"PRAGMA user_version = 2",
	} {
		if _, err := s.db.ExecContext(ctx, q); err != nil {
			s.Close()
			t.Fatalf("%s: %v", q, err)
		}
	}
	s.Close()

	s, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	want := seedNamed(t, beer.Name)
	got := product(t, s, beer.Name)
	if got.BarStock != want.stock || got.StorageStock != 0 || got.BarTarget != want.target {
		t.Errorf("bar %d, storage %d, target %d", got.BarStock, got.StorageStock, got.BarTarget)
	}
	if got := countedSince(t, s, beer.ID); got != want.stock-53 {
		t.Errorf("the new figure is booked as %d, want %d", got, want.stock-53)
	}
	var moves int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM stock_moves WHERE kind = 'restock' AND place = 'bar'").Scan(&moves); err != nil || moves != 1 {
		t.Errorf("old stock moves: %d, %v", moves, err)
	}
	if err := s.SaveRefill(ctx, []RefillLine{{got.ID, 1}}); err != nil {
		t.Errorf("refill after the upgrade: %v", err)
	}
	if got := product(t, s, "Pfand Flasche"); got.BarTarget != 0 {
		t.Errorf("a deposit key has no target: %d", got.BarTarget)
	}
}

func TestNewTillStartsWithTheBarRange(t *testing.T) {
	s := newStore(t)
	all, err := s.Products(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != len(catalog) {
		t.Fatalf("%d products, catalog has %d", len(all), len(catalog))
	}
	for i, p := range all {
		want := catalog[i]
		if p.Name != want.name || p.PriceCents != want.price || p.BarStock != want.stock || p.StorageStock != 0 ||
			p.Location != want.loc || p.QuickCount != want.quick || p.TrackStock != want.track || !p.Active {
			t.Errorf("%d: %+v, want %+v", i, p, want)
		}
	}

	categories, _ := s.Categories(ctx)
	names := []string{}
	for _, c := range categories {
		names = append(names, c.Name)
	}
	if got := strings.Join(names, " "); got != "Bier Alkoholfrei Snacks Sonstiges Pfand" {
		t.Errorf("categories = %s", got)
	}

	items, _ := s.RentalItems(ctx)
	if len(items) != len(lending) {
		t.Errorf("%d things to lend, the list has %d", len(items), len(lending))
	}
	byName := map[string]RentalItem{}
	for _, item := range items {
		byName[item.Name] = item
	}
	for _, l := range lending {
		got := byName[l.name]
		if got.DepositCents != l.deposit || got.Code != l.code || got.FeeCents != 0 || got.FeePer != "flat" || !got.Active {
			t.Errorf("%s = %+v", l.name, got)
		}
	}
}

// The till in the bar is on version 5 with the example range. The update hides that range, carries over
// the products the list names together with their sales, and brings the lending in line with the list.
func TestUpgradeFromVersionFiveTakesTheRealRange(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	hansa := product(t, s, "Hansa Pils")
	for _, q := range []string{
		"UPDATE products SET name = 'Hansa Pils 0,5 l', price_cents = 130, bar_stock = 10, storage_stock = 30, image_version = 7 WHERE name = 'Hansa Pils'",
		"INSERT INTO products(category_id, name, price_cents, track_stock, bar_stock) VALUES (1, 'Krombacher Pils 0,5 l', 150, 1, 20)",
		"UPDATE rental_items SET name = 'Airhockey-Set' WHERE name = 'Airhockey-Set 1'",
		"DELETE FROM rental_items WHERE name IN ('Airhockey-Set 2', 'Dartpfeile 3')",
		"UPDATE rental_items SET fee_cents = 50, fee_per = 'hour' WHERE name = 'Billardqueue 1'",
		"UPDATE rental_items SET deposit_cents = 500 WHERE name = 'Dartpfeile 1'",
		"INSERT INTO rental_items(name, code, deposit_cents) VALUES ('Billardqueue 2', '90002', 500)",
		"INSERT INTO rental_items(name, code, deposit_cents) VALUES ('Tischtennis-Set', '90021', 300)",
	} {
		if _, err := s.db.ExecContext(ctx, q); err != nil {
			s.Close()
			t.Fatalf("%s: %v", q, err)
		}
	}
	addDepositReturn(t, s)
	open(t, s, 0)
	sale := sell(t, s, "Hansa Pils 0,5 l", 2)
	oldSale := sell(t, s, "Krombacher Pils 0,5 l", 1)
	var queue, airhockey int64
	if err := s.db.QueryRowContext(ctx, "SELECT id FROM rental_items WHERE name = 'Billardqueue 2'").Scan(&queue); err != nil {
		t.Fatal(err)
	}
	if err := s.db.QueryRowContext(ctx, "SELECT id FROM rental_items WHERE name = 'Airhockey-Set'").Scan(&airhockey); err != nil {
		t.Fatal(err)
	}
	if err := s.StartRental(ctx, queue, "Jonas", "214", 500, ""); err != nil {
		t.Fatal(err)
	}
	asVersionEight(t, s)
	if _, err := s.db.ExecContext(ctx, "PRAGMA user_version = 5"); err != nil {
		t.Fatal(err)
	}
	s.Close()

	s, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	got := product(t, s, "Hansa Pils")
	if got.ID != hansa.ID || got.PriceCents != 100 || got.BarStock != 41 || got.StorageStock != 0 || !got.Active {
		t.Errorf("Hansa Pils = %+v", got)
	}
	if got.ImageVersion == 7 || got.ImageVersion == 0 {
		t.Errorf("picture version %d: browsers would keep the old picture", got.ImageVersion)
	}
	if got := countedSince(t, s, hansa.ID); got != 41-38 {
		t.Errorf("the new figure is booked as %d, want 3", got)
	}
	var sold int64
	if err := s.db.QueryRowContext(ctx, "SELECT product_id FROM sale_lines WHERE sale_id = ?", sale.ID).Scan(&sold); err != nil || sold != hansa.ID {
		t.Errorf("the old sale points to %d, %v", sold, err)
	}
	for _, name := range []string{"Krombacher Pils 0,5 l", "Pfand zurück"} {
		if hasProduct(t, s, name) {
			t.Errorf("%s is not on the list and must be gone", name)
		}
	}
	// The old product's sale stays, with its name and price, and points to no product any more.
	var line struct {
		product *int64
		name    string
		price   int64
	}
	if err := s.db.QueryRowContext(ctx, "SELECT product_id, name, unit_price_cents FROM sale_lines WHERE sale_id = ?", oldSale.ID).
		Scan(&line.product, &line.name, &line.price); err != nil || line.product != nil || line.name != "Krombacher Pils 0,5 l" || line.price != 150 {
		t.Errorf("the old sale = %+v, %v", line, err)
	}
	var moves int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM stock_moves WHERE product_id NOT IN (SELECT id FROM products)").Scan(&moves); err != nil || moves != 0 {
		t.Errorf("%d stock moves of deleted products, %v", moves, err)
	}
	all, _ := s.Products(ctx)
	active := 0
	for _, p := range all {
		if p.Active {
			active++
		}
	}
	if active != len(catalog) || len(all) != len(catalog) {
		t.Errorf("%d products, %d active: a renamed product must not come back as a second one", len(all), active)
	}

	items, _ := s.RentalItems(ctx)
	byName := map[string]RentalItem{}
	for _, item := range items {
		byName[item.Name] = item
	}
	if got := byName["Airhockey-Set 1"]; got.ID != airhockey || got.DepositCents != 1000 || !got.Active {
		t.Errorf("airhockey 1 = %+v", got)
	}
	if got := byName["Airhockey-Set 2"]; got.Code != "90032" || got.DepositCents != 1000 || !got.Active {
		t.Errorf("airhockey 2 = %+v", got)
	}
	if got := byName["Billardqueue 1"]; got.FeeCents != 0 || got.FeePer != "flat" || got.DepositCents != 500 {
		t.Errorf("cue = %+v", got)
	}
	if got := byName["Dartpfeile 1"]; got.DepositCents != 100 {
		t.Errorf("darts 1 = %+v", got)
	}
	if got := byName["Dartpfeile 3"]; got.Code != "90013" || !got.Active {
		t.Errorf("darts 3 = %+v", got)
	}
	if queue, found := byName["Billardqueue 2"]; !found || queue.Active {
		t.Errorf("a cue the list does not have but that was lent stays hidden: %+v", queue)
	}
	if _, found := byName["Tischtennis-Set"]; found {
		t.Error("a retired item that was never lent must be gone")
	}
	board, _ := s.RentalBoard(ctx)
	lent, gone := false, true
	for _, b := range board {
		lent = lent || (b.Name == "Billardqueue 2" && b.Rental != nil)
		gone = gone && b.Name != "Tischtennis-Set"
	}
	if !lent || !gone {
		t.Errorf("a hidden cue that is still out must stay on the board until it is back (%v, %v)", lent, gone)
	}
}
