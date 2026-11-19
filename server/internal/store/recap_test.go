package store

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func moment(month time.Month, day, hour int) time.Time {
	return time.Date(2026, month, day, hour, 0, 0, 0, time.Local)
}

func openAs(t *testing.T, s *Store, room, name string) {
	t.Helper()
	if _, err := s.OpenShift(ctx, room, name, 0); err != nil {
		t.Fatal(err)
	}
}

func sell(t *testing.T, s *Store, name string, qty int64) Sale {
	t.Helper()
	sale, err := s.CreateSale(ctx, []LineInput{{product(t, s, name).ID, qty}}, PayCash)
	if err != nil {
		t.Fatal(err)
	}
	return *sale
}

func closeShift(t *testing.T, s *Store) {
	t.Helper()
	if _, err := s.CloseShift(ctx, 0, "", ""); err != nil {
		// The count does not matter here, a closed shift with a difference is still closed.
		t.Fatal(err)
	}
}

func TestRecapFindsTheBestOfEachPeriod(t *testing.T) {
	s := newStore(t)

	at(s, moment(time.October, 3, 21))
	openAs(t, s, "214", "Jonas")
	sell(t, s, "Hansa Pils", 30)
	closeShift(t, s)

	at(s, moment(time.November, 5, 21))
	openAs(t, s, "108", "Mehmet")
	sell(t, s, "Coca-Cola", 4)
	sell(t, s, "Milka", 9)
	closeShift(t, s)

	at(s, moment(time.November, 7, 22))
	r, err := s.Recap(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if r.Semester.From != "2026-10-01" || r.Semester.To != "2026-11-07" {
		t.Errorf("semester runs %s to %s", r.Semester.From, r.Semester.To)
	}
	if got := r.Semester.Drinks[0].Name; got != "Hansa Pils" {
		t.Errorf("top drink of the semester = %q", got)
	}
	if got := r.Semester.Workers[0]; got.Name != "Jonas" || got.Qty != 30 || got.Shifts != 1 {
		t.Errorf("top worker of the semester = %+v", got)
	}
	if got := r.Month.Products[0].Name; got != "Milka" {
		t.Errorf("best seller of the month = %q", got)
	}
	if got := r.Month.Drinks[0].Name; got != "Coca-Cola" {
		t.Errorf("snacks must not be drinks, top drink = %q", got)
	}
	if got := r.Month.Workers[0].Name; got != "Mehmet" {
		t.Errorf("top worker of the month = %q", got)
	}
	if got := r.Week.Qty; got != 13 {
		t.Errorf("pieces in the last 7 days = %d, want 13", got)
	}
	if got := r.Semester.Nights; got != 2 {
		t.Errorf("nights = %d, want 2", got)
	}
	if r.Semester.Busiest == nil || r.Semester.Busiest.Day != "2026-10-03" || r.Semester.Busiest.Qty != 30 {
		t.Errorf("busiest night = %+v", r.Semester.Busiest)
	}
}

func TestRecapBusiestNightRunsPastMidnight(t *testing.T) {
	s := newStore(t)
	at(s, moment(time.November, 6, 23))
	openAs(t, s, "214", "Jonas")
	sell(t, s, "Hansa Pils", 2)
	at(s, moment(time.November, 7, 2))
	sell(t, s, "Hansa Pils", 3)
	closeShift(t, s)

	at(s, moment(time.November, 7, 12))
	r, err := s.Recap(ctx)
	if err != nil {
		t.Fatal(err)
	}
	night := r.Week.Busiest
	if night == nil || night.Day != "2026-11-06" || night.Qty != 5 || night.Orders != 2 {
		t.Fatalf("the sale at 02:00 belongs to the evening before: %+v", night)
	}
	if want := moment(time.November, 7, 2).UnixMilli(); r.Week.LastAt != want {
		t.Errorf("latest order = %v, want 02:00", time.UnixMilli(r.Week.LastAt))
	}
}

func TestRecapReversedSaleCountsAgainstItsOwnWorker(t *testing.T) {
	s := newStore(t)
	at(s, moment(time.November, 5, 21))
	openAs(t, s, "214", "Jonas")
	sale := sell(t, s, "Hansa Pils", 10)
	sell(t, s, "Hansa Pils", 2)
	closeShift(t, s)

	at(s, moment(time.November, 6, 21))
	openAs(t, s, "108", "Mehmet")
	sell(t, s, "Hansa Pils", 3)
	if _, err := s.VoidSale(ctx, sale.ID, "Falsch getippt"); err != nil {
		t.Fatal(err)
	}
	closeShift(t, s)

	at(s, moment(time.November, 7, 12))
	r, err := s.Recap(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got := r.Week.Workers[0]; got.Name != "Mehmet" || got.Qty != 3 {
		t.Errorf("top worker = %+v, want Mehmet with 3", got)
	}
	if got := r.Week.Workers[1]; got.Name != "Jonas" || got.Qty != 2 {
		t.Errorf("second worker = %+v, want Jonas with 2", got)
	}
	if got := r.Week.Orders; got != 2 {
		t.Errorf("orders = %d, the reversed one is gone", got)
	}
	if got := r.Week.Busiest; got == nil || got.Day != "2026-11-06" {
		t.Errorf("busiest night = %+v", got)
	}
}

func TestRecapLeavesDepositsAndWaschmarkenOut(t *testing.T) {
	s := newStore(t)
	at(s, moment(time.November, 5, 21))
	openAs(t, s, "214", "Jonas")
	sell(t, s, "Pfand Flasche", 40)
	sell(t, s, "Waschmarke", 40)
	addDepositReturn(t, s)
	sell(t, s, "Pfand zurück", 5)
	sell(t, s, "Schokowaffel", 1)
	closeShift(t, s)

	at(s, moment(time.November, 7, 12))
	r, err := s.Recap(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got := r.Week.Products; len(got) != 1 || got[0].Name != "Schokowaffel" {
		t.Errorf("products = %+v, want Schokowaffel alone", got)
	}
	if got := r.Week.Qty; got != 1 {
		t.Errorf("pieces = %d, want 1", got)
	}
	if got := len(r.Week.Drinks); got != 0 {
		t.Errorf("drinks = %d, want none", got)
	}
}

func TestRecapCountsOnePersonOnce(t *testing.T) {
	s := newStore(t)
	at(s, moment(time.November, 5, 21))
	openAs(t, s, "214", "Jonas")
	sell(t, s, "Schokowaffel", 1)
	closeShift(t, s)
	at(s, moment(time.November, 6, 21))
	openAs(t, s, "214", " jonas ")
	sell(t, s, "Schokowaffel", 1)
	closeShift(t, s)

	at(s, moment(time.November, 7, 12))
	r, err := s.Recap(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got := r.Week.Workers; len(got) != 1 || got[0].Qty != 2 || got[0].Shifts != 2 {
		t.Errorf("workers = %+v, want one person with 2 pieces in 2 shifts", got)
	}
}

func TestRecapFollowsTheGermanSemesters(t *testing.T) {
	cases := []struct {
		now      time.Time
		term     Term
		from, to string
		lastFrom string
	}{
		{time.Date(2026, time.November, 7, 12, 0, 0, 0, time.Local), Term{"winter", 2026}, "2026-10-01", "2026-11-07", "2026-04-01"},
		{time.Date(2027, time.February, 7, 12, 0, 0, 0, time.Local), Term{"winter", 2026}, "2026-10-01", "2027-02-07", "2026-04-01"},
		{time.Date(2027, time.May, 7, 12, 0, 0, 0, time.Local), Term{"summer", 2027}, "2027-04-01", "2027-05-07", "2026-10-01"},
		// 02:00 on the first of April is still the last night of the winter.
		{time.Date(2027, time.April, 1, 2, 0, 0, 0, time.Local), Term{"winter", 2026}, "2026-10-01", "2027-03-31", "2026-04-01"},
	}
	for _, c := range cases {
		s := newStore(t)
		at(s, c.now)
		r, err := s.Recap(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if r.Term != c.term || r.Semester.From != c.from || r.Semester.To != c.to || r.LastSemester.From != c.lastFrom {
			t.Errorf("%v: %+v from %s to %s, last from %s", c.now, r.Term, r.Semester.From, r.Semester.To, r.LastSemester.From)
		}
	}
}

func TestRecapLastMonthIsTheWholeMonthBefore(t *testing.T) {
	s := newStore(t)

	at(s, time.Date(2026, time.September, 30, 21, 0, 0, 0, time.Local))
	openAs(t, s, "301", "Too Early")
	sell(t, s, "Hansa Pils", 50)
	closeShift(t, s)

	at(s, moment(time.October, 1, 21))
	openAs(t, s, "214", "Jonas")
	sell(t, s, "Hansa Pils", 8)
	sell(t, s, "Milka", 3)
	closeShift(t, s)

	// The night of 31 October runs past midnight and still belongs to October.
	at(s, moment(time.October, 31, 22))
	openAs(t, s, "108", "Mehmet")
	at(s, moment(time.November, 1, 2))
	sell(t, s, "Coca-Cola", 5)
	closeShift(t, s)

	at(s, moment(time.November, 5, 21))
	openAs(t, s, "214", "Jonas")
	sell(t, s, "Coca-Cola", 2)
	closeShift(t, s)

	at(s, moment(time.November, 7, 12))
	r, err := s.Recap(ctx)
	if err != nil {
		t.Fatal(err)
	}
	last := r.LastMonth
	if last.From != "2026-10-01" || last.To != "2026-10-31" {
		t.Fatalf("last month runs from %s to %s", last.From, last.To)
	}
	if last.Qty != 16 {
		t.Errorf("last month sold %d, want 16", last.Qty)
	}
	if len(last.Workers) != 2 || last.Workers[0].Name != "Jonas" || last.Workers[0].Room != "214" {
		t.Errorf("workers of last month: %+v", last.Workers)
	}
	if len(last.Snacks) != 1 || last.Snacks[0].Name != "Milka" {
		t.Errorf("snacks of last month: %+v", last.Snacks)
	}
	if len(last.Drinks) != 2 || last.Drinks[0].Name != "Hansa Pils" {
		t.Errorf("drinks of last month: %+v", last.Drinks)
	}
	if r.Month.Qty != 2 || r.Month.From != "2026-11-01" {
		t.Errorf("this month: %d from %s", r.Month.Qty, r.Month.From)
	}
}

func TestRecapLastMonthCrossesTheYear(t *testing.T) {
	s := newStore(t)
	at(s, time.Date(2027, time.January, 12, 21, 0, 0, 0, time.Local))
	r, err := s.Recap(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if r.LastMonth.From != "2026-12-01" || r.LastMonth.To != "2026-12-31" {
		t.Errorf("last month runs from %s to %s", r.LastMonth.From, r.LastMonth.To)
	}
}

func TestRecapStaysFreeOfMoney(t *testing.T) {
	s := newStore(t)
	at(s, moment(time.November, 5, 21))
	openAs(t, s, "9214", "Jonas")
	sell(t, s, "Hansa Pils", 7)
	closeShift(t, s)

	at(s, moment(time.November, 7, 12))
	r, err := s.Recap(ctx)
	if err != nil {
		t.Fatal(err)
	}
	body, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	for _, leak := range []string{"Cents", "cents", "stock"} {
		if strings.Contains(string(body), leak) {
			t.Errorf("the recap is shown at the bar and must not contain %q: %s", leak, body)
		}
	}
	if len(r.Month.Workers) != 1 || r.Month.Workers[0].Room != "9214" {
		t.Errorf("the leaderboard shows rooms: %+v", r.Month.Workers)
	}
}

func TestWorkerNamesSuggestsWhoUsedARoom(t *testing.T) {
	s := newStore(t)
	for _, name := range []string{"Jonas", "jonas ", "Lena", "Mehmet", "Sara", "Mehmet"} {
		at(s, moment(time.November, 2, 21))
		openAs(t, s, "214", name)
		closeShift(t, s)
	}
	openAs(t, s, "108", "Somebody Else")
	closeShift(t, s)

	names, err := s.WorkerNames(ctx, " 214 ")
	if err != nil {
		t.Fatal(err)
	}
	if want := "Mehmet,Sara,Lena"; strings.Join(names, ",") != want {
		t.Errorf("got %v, want the latest three names without repeats: %s", names, want)
	}

	names, err = s.WorkerNames(ctx, "999")
	if err != nil {
		t.Fatal(err)
	}
	if names == nil || len(names) != 0 {
		t.Errorf("an unknown room gets an empty list, got %#v", names)
	}
}
