package store

import (
	"context"
	"time"
)

// What leaves the bar PC for the OBK's web page is listed here, field by field, so a new field somewhere else never goes
// online by accident. No worker or renter is named and no room number is sent. Messages and notes are sent as typed.
type OnlineSnapshot struct {
	At              int64          `json:"at"`
	BarName         string         `json:"barName"`
	Language        string         `json:"language"`
	Open            *OnlineShift   `json:"open"` // the shift running right now
	TodayCents      int64          `json:"todayCents"`
	TodaySales      int64          `json:"todaySales"`
	WeekCents       int64          `json:"weekCents"`
	CashInBoxCents  int64          `json:"cashInBoxCents"`
	StockValueCents int64          `json:"stockValueCents"`
	MissingCents30  int64          `json:"missingCents30"`
	RentalsOut      int64          `json:"rentalsOut"`
	ToRefill        int64          `json:"toRefill"`
	Series          []DayRevenue   `json:"series"`
	Month           OnlinePeriod   `json:"month"`
	LastMonth       OnlinePeriod   `json:"lastMonth"`
	LowStock        []OnlineLow    `json:"lowStock"`
	Shifts          []OnlineShift  `json:"shifts"`
	Message         *OnlineMessage `json:"message"`
}

type OnlineShift struct {
	OpenedAt        int64  `json:"openedAt"`
	ClosedAt        *int64 `json:"closedAt"`
	SalesCount      int64  `json:"salesCount"`
	TotalCents      int64  `json:"totalCents"`
	DifferenceCents *int64 `json:"differenceCents"` // nil while the shift runs
	Note            string `json:"note"`
}

type OnlineMessage struct {
	Text string `json:"text"`
	At   int64  `json:"at"`
}

type OnlineProduct struct {
	Name         string `json:"name"`
	Qty          int64  `json:"qty"`
	RevenueCents int64  `json:"revenueCents"`
}

type OnlinePeriod struct {
	From         string          `json:"from"`
	To           string          `json:"to"`
	RevenueCents int64           `json:"revenueCents"`
	CostCents    int64           `json:"costCents"`
	SalesCount   int64           `json:"salesCount"`
	Products     []OnlineProduct `json:"products"`
}

type OnlineLow struct {
	Name      string `json:"name"`
	Stock     int64  `json:"stock"`
	MinStock  int64  `json:"minStock"`
	CrateSize int64  `json:"crateSize"`
}

const (
	onlineTopProducts = 10
	onlineShifts      = 10
)

func onlineShift(row ShiftRow) OnlineShift {
	sh := OnlineShift{OpenedAt: row.OpenedAt, ClosedAt: row.ClosedAt, SalesCount: row.SalesCount, TotalCents: row.TotalCents, Note: row.Note}
	if row.ClosedAt != nil && row.CountedCents != nil && row.ExpectedCents != nil {
		difference := *row.CountedCents - *row.ExpectedCents
		sh.DifferenceCents = &difference
	}
	return sh
}

func (s *Store) onlinePeriod(ctx context.Context, from, to string) (OnlinePeriod, error) {
	report, err := s.Report(ctx, from, to)
	if err != nil {
		return OnlinePeriod{}, err
	}
	period := OnlinePeriod{From: from, To: to, RevenueCents: report.RevenueCents, CostCents: report.CostCents,
		SalesCount: report.SalesCount, Products: []OnlineProduct{}}
	for _, p := range report.Products {
		if len(period.Products) == onlineTopProducts {
			break
		}
		period.Products = append(period.Products, OnlineProduct{p.Name, p.Qty, p.RevenueCents})
	}
	return period, nil
}

// OnlineSnapshot gathers the numbers of the OBK's overview for sending.
func (s *Store) OnlineSnapshot(ctx context.Context) (*OnlineSnapshot, error) {
	d, err := s.Dashboard(ctx)
	if err != nil {
		return nil, err
	}
	now := s.now()
	today := businessDay(now)
	monthStart, err := time.ParseInLocation("2006-01-02", today[:8]+"01", time.Local)
	if err != nil {
		return nil, err
	}

	snap := &OnlineSnapshot{
		At:              now.UnixMilli(),
		TodayCents:      d.Today,
		TodaySales:      d.TodaySales,
		WeekCents:       d.Week,
		CashInBoxCents:  d.CashInBox,
		StockValueCents: d.StockValue,
		MissingCents30:  d.MissingCents30,
		RentalsOut:      d.RentalsOut,
		ToRefill:        d.ToRefill,
		Series:          d.Series,
		LowStock:        []OnlineLow{},
		Shifts:          []OnlineShift{},
	}
	if snap.Month, err = s.onlinePeriod(ctx, today[:8]+"01", today); err != nil {
		return nil, err
	}
	lastFrom := monthStart.AddDate(0, -1, 0).Format("2006-01-02")
	lastTo := monthStart.AddDate(0, 0, -1).Format("2006-01-02")
	if snap.LastMonth, err = s.onlinePeriod(ctx, lastFrom, lastTo); err != nil {
		return nil, err
	}
	for _, l := range d.LowStock {
		snap.LowStock = append(snap.LowStock, OnlineLow{l.Name, l.Stock, l.MinStock, l.CrateSize})
	}
	shifts, err := s.ListShifts(ctx, onlineShifts)
	if err != nil {
		return nil, err
	}
	for _, row := range shifts {
		if row.ClosedAt == nil {
			open := onlineShift(row)
			snap.Open = &open
		}
		snap.Shifts = append(snap.Shifts, onlineShift(row))
	}
	if d.Message != nil {
		snap.Message = &OnlineMessage{d.Message.Text, d.Message.At}
	}
	return snap, nil
}
