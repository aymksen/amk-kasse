package store

import (
	"errors"
	"testing"
	"time"
)

func checkLine(t *testing.T, c Check, id int64) CheckLine {
	t.Helper()
	for _, l := range c.Lines {
		if l.ProductID == id {
			return l
		}
	}
	t.Fatalf("product %d is not in the check", id)
	return CheckLine{}
}

func checkNow(t *testing.T, s *Store) Check {
	t.Helper()
	c, err := s.StockCheck(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// confirmAll is a worker who finds the bar as expected, except for the changes given.
func confirmAll(c Check, changes map[int64]int64) []CountLine {
	count := make([]CountLine, 0, len(c.Lines))
	for _, l := range c.Lines {
		count = append(count, CountLine{l.ProductID, max(l.Expected, 0) + changes[l.ProductID]})
	}
	return count
}

func TestTheShiftStartsWithACheckOfTheBar(t *testing.T) {
	s := newStore(t)
	base := time.Date(2026, 11, 14, 20, 0, 0, 0, time.Local)
	at(s, base)
	hansa := product(t, s, "Hansa Pils")

	// A new till: nobody worked before, so nothing was added either.
	first := checkNow(t, s)
	tracked := 0
	all, _ := s.Products(ctx)
	for _, p := range all {
		if p.Active && p.TrackStock {
			tracked++
		}
	}
	if first.Last != nil || first.Since != 0 || len(first.Lines) != tracked {
		t.Errorf("new till: last %+v, since %d, %d lines for %d counted products", first.Last, first.Since, len(first.Lines), tracked)
	}
	if l := checkLine(t, first, hansa.ID); l.Expected != hansa.BarStock || l.Added != 0 || l.Sold != 0 {
		t.Errorf("hansa on a new till = %+v", l)
	}

	// Jonas sells two, counts one more missing at the end, and the OBK fills five in later.
	open(t, s, 0)
	sell(t, s, "Hansa Pils", 2)
	at(s, base.Add(4*time.Hour))
	end := checkNow(t, s)
	if l := checkLine(t, end, hansa.ID); end.Last != nil || l.Start != hansa.BarStock || l.Expected != hansa.BarStock-2 || l.Sold != 2 {
		t.Errorf("hansa at the end of Jonas's shift = %+v, last %+v", l, end.Last)
	}
	if _, err := s.CloseShift(ctx, 0, "", "", confirmAll(end, map[int64]int64{hansa.ID: -1})...); err != nil {
		t.Fatal(err)
	}
	takes, _ := s.Stocktakes(ctx, 1)
	if got := takes[0]; got.Opening || got.Kind != kindClosing || got.ByName != "Jonas" || got.Diffs != 1 || got.Lines != int64(len(end.Lines)) {
		t.Errorf("the count at the end = %+v", got)
	}
	at(s, base.Add(20*time.Hour))
	if err := s.SaveRefill(ctx, []RefillLine{{hansa.ID, 5}}); err != nil {
		t.Fatal(err)
	}

	check := checkNow(t, s)
	if l := check.Last; l == nil || l.Name != "Jonas" || l.Room != "214" || !l.Counted || l.ClosedAt != base.Add(4*time.Hour).UnixMilli() {
		t.Errorf("last shift = %+v", l)
	}
	if l := checkLine(t, check, hansa.ID); l.Start != hansa.BarStock-3 || l.Expected != hansa.BarStock-3+5 || l.Added != 5 || l.Sold != 0 {
		t.Errorf("hansa after the refill = %+v", l)
	}

	// Mehmet finds one more missing and confirms the rest as it is.
	at(s, base.Add(24*time.Hour))
	count := confirmAll(check, map[int64]int64{hansa.ID: -1})
	sh, err := s.OpenShift(ctx, "108", "Mehmet", 0, count...)
	if err != nil {
		t.Fatal(err)
	}
	if got := product(t, s, "Hansa Pils").BarStock; got != hansa.BarStock-3+5-1 {
		t.Errorf("hansa after the check = %d", got)
	}
	takes, _ = s.Stocktakes(ctx, 1)
	if got := takes[0]; !got.Opening || got.ShiftID == nil || *got.ShiftID != sh.ID || got.ByName != "Mehmet" ||
		got.Lines != int64(len(count)) || got.Diffs != 1 || got.Place != PlaceBar {
		t.Errorf("the check = %+v", got)
	}

	// During his shift the check counts from its start: the refill before it is part of the figure, not news.
	// A count from the dock in between shows as a correction.
	at(s, base.Add(25*time.Hour))
	sell(t, s, "Hansa Pils", 1)
	if err := s.SaveRefill(ctx, []RefillLine{{hansa.ID, 2}}); err != nil {
		t.Fatal(err)
	}
	startOfShift := hansa.BarStock - 3 + 5 - 1
	if _, err := s.SaveStocktake(ctx, "full", PlaceBar, "108", "Mehmet", []CountLine{{hansa.ID, startOfShift - 1 + 2 + 3}}); err != nil {
		t.Fatal(err)
	}
	during := checkNow(t, s)
	if l := checkLine(t, during, hansa.ID); during.Last != nil || during.Since != sh.OpenedAt || l.Start != startOfShift ||
		l.Expected != startOfShift-1+2+3 || l.Sold != 1 || l.Added != 2 || l.Corrected != 3 {
		t.Errorf("hansa during Mehmet's shift = %+v, since %d", l, during.Since)
	}

	// A check that cannot be saved leaves no shift behind, and does not end one either.
	if _, err := s.CloseShift(ctx, 0, "", "", CountLine{hansa.ID, -1}); !errors.Is(err, ErrInvalid) {
		t.Errorf("negative count at the end: %v", err)
	}
	if open, err := s.CurrentShift(ctx); err != nil || open == nil {
		t.Errorf("a failed count ended the shift: %+v, %v", open, err)
	}
	closeShift(t, s)
	if _, err := s.OpenShift(ctx, "322", "Lukas", 0, CountLine{hansa.ID, -1}); !errors.Is(err, ErrInvalid) {
		t.Errorf("negative count: %v", err)
	}
	if _, err := s.OpenShift(ctx, "322", "Lukas", 0, CountLine{999999, 1}); !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown product: %v", err)
	}
	if open, err := s.CurrentShift(ctx); err != nil || open != nil {
		t.Errorf("a failed check opened a shift: %+v, %v", open, err)
	}
}

// The start check is no count of the shift itself: a shift that only checked at the start did not count.
func TestAStartCheckIsNoClosingCount(t *testing.T) {
	s := newStore(t)
	check := checkNow(t, s)
	if _, err := s.OpenShift(ctx, "214", "Jonas", 0, CountLine{check.Lines[0].ProductID, max(check.Lines[0].Expected, 0)}); err != nil {
		t.Fatal(err)
	}
	closeShift(t, s)
	if next := checkNow(t, s); next.Last == nil || next.Last.Counted {
		t.Errorf("last = %+v, want a shift without a closing count", next.Last)
	}
}
