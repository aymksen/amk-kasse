package store

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestTheNextShiftGetsTheMessage(t *testing.T) {
	s := newStore(t)
	base := time.Date(2026, 11, 6, 20, 0, 0, 0, time.Local)
	at(s, base)
	open(t, s, 0)
	if got, err := s.LastMessage(ctx); err != nil || got != nil {
		t.Fatalf("message before any shift ended = %+v, %v", got, err)
	}

	at(s, base.Add(4*time.Hour))
	if _, err := s.CloseShift(ctx, 0, "", "  Kühlschrank tropft  "); err != nil {
		t.Fatal(err)
	}
	got, err := s.LastMessage(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.Text != "Kühlschrank tropft" || got.From != "Jonas" || got.At != base.Add(4*time.Hour).UnixMilli() {
		t.Fatalf("message = %+v", got)
	}

	// Whoever comes after that shift has seen it, so an empty message ends it.
	at(s, base.Add(24*time.Hour))
	open(t, s, 0)
	if _, err := s.CloseShift(ctx, 0, "", ""); err != nil {
		t.Fatal(err)
	}
	if got, err := s.LastMessage(ctx); err != nil || got != nil {
		t.Fatalf("old message came back: %+v, %v", got, err)
	}
}

func TestAMessageStaysShort(t *testing.T) {
	s := newStore(t)
	open(t, s, 0)
	closed, err := s.CloseShift(ctx, 0, "", strings.Repeat("ü", maxMessage+50))
	if err != nil {
		t.Fatal(err)
	}
	if got := len([]rune(closed.Handover)); got != maxMessage {
		t.Errorf("message has %d characters, want %d", got, maxMessage)
	}
}

func TestOnlineSnapshotNamesNobody(t *testing.T) {
	s := newStore(t)
	base := time.Date(2026, 11, 6, 20, 0, 0, 0, time.Local)
	at(s, base)
	if _, err := s.OpenShift(ctx, "777", "Heimlich", 1000); err != nil {
		t.Fatal(err)
	}
	sell(t, s, "Hansa Pils", 2)
	items, err := s.RentalItems(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.StartRental(ctx, items[0].ID, "Verborgen", "888", 500, ""); err != nil {
		t.Fatal(err)
	}

	snap, err := s.OnlineSnapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	body, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"Heimlich", "Verborgen", "777", "888"} {
		if strings.Contains(string(body), secret) {
			t.Errorf("%q went online: %s", secret, body)
		}
	}
	if snap.Open == nil || snap.Open.SalesCount != 1 {
		t.Errorf("running shift = %+v", snap.Open)
	}
	if snap.TodayCents == 0 || snap.Month.RevenueCents != snap.TodayCents || len(snap.Month.Products) != 1 {
		t.Errorf("today = %d, month = %+v", snap.TodayCents, snap.Month)
	}
	if snap.RentalsOut != 1 {
		t.Errorf("rentals out = %d", snap.RentalsOut)
	}
}

func TestOnlineLastMonthIsTheWholeMonthBefore(t *testing.T) {
	s := newStore(t)
	at(s, time.Date(2026, 1, 12, 20, 0, 0, 0, time.Local))
	open(t, s, 0)
	snap, err := s.OnlineSnapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if snap.Month.From != "2026-01-01" || snap.LastMonth.From != "2025-12-01" || snap.LastMonth.To != "2025-12-31" {
		t.Errorf("month %s, last month %s to %s", snap.Month.From, snap.LastMonth.From, snap.LastMonth.To)
	}
}
