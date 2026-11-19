package store

import (
	"context"
	"time"
)

type DayRevenue struct {
	Day        string `json:"day"`
	TotalCents int64  `json:"totalCents"`
}

type ProductSales struct {
	Name         string `json:"name"`
	Category     string `json:"category"`
	Qty          int64  `json:"qty"`
	RevenueCents int64  `json:"revenueCents"`
	CostCents    int64  `json:"costCents"`
}

// LowStock is a product to buy more of. Stock counts both places together.
type LowStock struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Stock     int64  `json:"stock"`
	MinStock  int64  `json:"minStock"`
	CrateSize int64  `json:"crateSize"`
}

type Dashboard struct {
	Today          int64          `json:"todayCents"`
	Week           int64          `json:"weekCents"`
	Month          int64          `json:"monthCents"`
	TodaySales     int64          `json:"todaySales"`
	CashInBox      int64          `json:"cashInBoxCents"`
	StockValue     int64          `json:"stockValueCents"`
	OpenShift      *Shift         `json:"openShift"`
	RentalsOut     int64          `json:"rentalsOut"`
	Series         []DayRevenue   `json:"series"`
	TopProducts    []ProductSales `json:"topProducts"`
	LowStock       []LowStock     `json:"lowStock"`
	ToRefill       int64          `json:"toRefill"` // products the OBK should bring to the bar
	RecentShifts   []ShiftRow     `json:"recentShifts"`
	MissingCents30 int64          `json:"missingCents30"` // cash differences over 30 days
	Message        *ShiftMessage  `json:"message"`        // what the last shift left for the next one
}

// ShiftRow is one line of the shift list, with what it took.
type ShiftRow struct {
	Shift
	SalesCount int64 `json:"salesCount"`
	TotalCents int64 `json:"totalCents"`
}

func (s *Store) revenueBetween(ctx context.Context, from, to string) (cents, count int64, err error) {
	err = s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(total_cents), 0),
		COALESCE(SUM(CASE WHEN voids_sale_id IS NULL THEN 1 ELSE 0 END), 0)
		FROM sales WHERE day >= ? AND day <= ?`, from, to).Scan(&cents, &count)
	return
}

func (s *Store) ListShifts(ctx context.Context, limit int) ([]ShiftRow, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+shiftColumns+` FROM shifts ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	out := []ShiftRow{}
	for rows.Next() {
		sh, err := scanShift(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		out = append(out, ShiftRow{Shift: sh})
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range out {
		t, err := s.ShiftTotals(ctx, out[i].Shift)
		if err != nil {
			return nil, err
		}
		out[i].SalesCount = t.SalesCount
		out[i].TotalCents = t.CashCents + t.OtherCents
	}
	return out, nil
}

func (s *Store) Dashboard(ctx context.Context) (*Dashboard, error) {
	now := s.now()
	today := businessDay(now)
	d := &Dashboard{Series: []DayRevenue{}}
	var err error

	if d.Today, d.TodaySales, err = s.revenueBetween(ctx, today, today); err != nil {
		return nil, err
	}
	if d.Week, _, err = s.revenueBetween(ctx, businessDay(now.AddDate(0, 0, -6)), today); err != nil {
		return nil, err
	}
	if d.Month, _, err = s.revenueBetween(ctx, today[:8]+"01", today); err != nil {
		return nil, err
	}
	if d.CashInBox, err = s.CashInBox(ctx); err != nil {
		return nil, err
	}
	if d.OpenShift, err = s.CurrentShift(ctx); err != nil {
		return nil, err
	}
	if err = s.db.QueryRowContext(ctx,
		"SELECT COALESCE(SUM(MAX(bar_stock + storage_stock, 0) * cost_cents), 0) FROM products WHERE active = 1 AND track_stock = 1").
		Scan(&d.StockValue); err != nil {
		return nil, err
	}
	if err = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM rentals WHERE returned_at IS NULL").Scan(&d.RentalsOut); err != nil {
		return nil, err
	}

	// A 30-day series with no gaps, so quiet days still show on the chart.
	byDay := map[string]int64{}
	first := businessDay(now.AddDate(0, 0, -29))
	rows, err := s.db.QueryContext(ctx,
		"SELECT day, SUM(total_cents) FROM sales WHERE day >= ? GROUP BY day", first)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var day string
		var total int64
		if err := rows.Scan(&day, &total); err != nil {
			rows.Close()
			return nil, err
		}
		byDay[day] = total
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := 29; i >= 0; i-- {
		day := businessDay(now.AddDate(0, 0, -i))
		d.Series = append(d.Series, DayRevenue{Day: day, TotalCents: byDay[day]})
	}

	report, err := s.Report(ctx, first, today)
	if err != nil {
		return nil, err
	}
	d.TopProducts = report.Products
	if len(d.TopProducts) > 5 {
		d.TopProducts = d.TopProducts[:5]
	}

	products, err := s.Products(ctx)
	if err != nil {
		return nil, err
	}
	for _, p := range products {
		if p.Active && p.RefillQty > 0 {
			d.ToRefill++
		}
	}
	if d.LowStock, err = s.lowStock(ctx); err != nil {
		return nil, err
	}
	if d.RecentShifts, err = s.ListShifts(ctx, 6); err != nil {
		return nil, err
	}
	if d.Message, err = s.LastMessage(ctx); err != nil {
		return nil, err
	}
	err = s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(counted_cents - expected_cents), 0) FROM shifts
		WHERE closed_at IS NOT NULL AND day >= ?`, first).Scan(&d.MissingCents30)
	return d, err
}

func (s *Store) lowStock(ctx context.Context) ([]LowStock, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, bar_stock + storage_stock, min_stock, crate_size FROM products
		WHERE active = 1 AND track_stock = 1 AND bar_stock + storage_stock <= min_stock
		ORDER BY bar_stock + storage_stock - min_stock, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []LowStock{}
	for rows.Next() {
		var l LowStock
		if err := rows.Scan(&l.ID, &l.Name, &l.Stock, &l.MinStock, &l.CrateSize); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

type StockLoss struct {
	Name      string `json:"name"`
	Units     int64  `json:"units"`
	CostCents int64  `json:"costCents"`
}

type Report struct {
	From         string         `json:"from"`
	To           string         `json:"to"`
	RevenueCents int64          `json:"revenueCents"`
	CostCents    int64          `json:"costCents"`
	CashCents    int64          `json:"cashCents"`
	OtherCents   int64          `json:"otherCents"`
	SalesCount   int64          `json:"salesCount"`
	Products     []ProductSales `json:"products"`
	Categories   []ProductSales `json:"categories"`
	Losses       []StockLoss    `json:"losses"`
}

// Report sums sale lines, so reversals cancel out the sales they undo.
func (s *Store) Report(ctx context.Context, from, to string) (*Report, error) {
	r := &Report{From: from, To: to, Products: []ProductSales{}, Categories: []ProductSales{}, Losses: []StockLoss{}}
	err := s.db.QueryRowContext(ctx, `SELECT
		COALESCE(SUM(CASE WHEN payment = 'cash' THEN total_cents ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN payment = 'other' THEN total_cents ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN voids_sale_id IS NULL THEN 1 ELSE 0 END), 0)
		FROM sales WHERE day >= ? AND day <= ?`, from, to).Scan(&r.CashCents, &r.OtherCents, &r.SalesCount)
	if err != nil {
		return nil, err
	}
	r.RevenueCents = r.CashCents + r.OtherCents

	load := func(query string, into *[]ProductSales) error {
		rows, err := s.db.QueryContext(ctx, query, from, to)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var p ProductSales
			if err := rows.Scan(&p.Name, &p.Category, &p.Qty, &p.RevenueCents, &p.CostCents); err != nil {
				return err
			}
			*into = append(*into, p)
		}
		return rows.Err()
	}
	if err := load(`SELECT l.name, l.category, SUM(l.qty), SUM(l.qty * l.unit_price_cents), SUM(l.qty * l.unit_cost_cents)
		FROM sale_lines l JOIN sales s ON s.id = l.sale_id WHERE s.day >= ? AND s.day <= ?
		GROUP BY l.name, l.category HAVING SUM(l.qty) <> 0 ORDER BY 4 DESC`, &r.Products); err != nil {
		return nil, err
	}
	if err := load(`SELECT l.category, '', SUM(l.qty), SUM(l.qty * l.unit_price_cents), SUM(l.qty * l.unit_cost_cents)
		FROM sale_lines l JOIN sales s ON s.id = l.sale_id WHERE s.day >= ? AND s.day <= ?
		GROUP BY l.category HAVING SUM(l.qty) <> 0 ORDER BY 4 DESC`, &r.Categories); err != nil {
		return nil, err
	}
	for _, p := range r.Products {
		r.CostCents += p.CostCents
	}

	fromT, err := time.ParseInLocation("2006-01-02", from, time.Local)
	if err != nil {
		return nil, ErrInvalid
	}
	toT, err := time.ParseInLocation("2006-01-02", to, time.Local)
	if err != nil {
		return nil, ErrInvalid
	}
	rows, err := s.db.QueryContext(ctx, `SELECT p.name, SUM(l.counted - l.expected), SUM((l.counted - l.expected) * l.unit_cost_cents)
		FROM stocktake_lines l JOIN stocktakes t ON t.id = l.stocktake_id JOIN products p ON p.id = l.product_id
		WHERE t.created_at >= ? AND t.created_at < ? GROUP BY p.id HAVING SUM(l.counted - l.expected) <> 0 ORDER BY 3`,
		fromT.Add(6*time.Hour).UnixMilli(), toT.AddDate(0, 0, 1).Add(6*time.Hour).UnixMilli())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var l StockLoss
		if err := rows.Scan(&l.Name, &l.Units, &l.CostCents); err != nil {
			return nil, err
		}
		r.Losses = append(r.Losses, l)
	}
	return r, rows.Err()
}
