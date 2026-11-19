package store

import (
	"context"
	"sort"
	"strings"
	"time"
)

// A recap counts pieces, names and rooms, never money or stock, because it is shown at the bar, also before anyone
// has started a shift.
//
// Deposits and everything under Sonstiges (the Waschmarken) are not what anyone means by the best seller,
// so they stay out, and snacks stay out of the drinks as well. A reversed sale counts on the night and for
// the worker of the sale it undoes, not for whoever happened to reverse it.
const recapLines = `SELECT COALESCE(o.day, s.day), COALESCE(o.shift_id, s.shift_id), sh.worker_room, sh.worker_name,
		l.name, l.category, COALESCE(l.product_id, 0), COALESCE(p.image_version, 0), SUM(l.qty)
	FROM sale_lines l
	JOIN sales s ON s.id = l.sale_id
	LEFT JOIN sales o ON o.id = s.voids_sale_id
	LEFT JOIN products p ON p.id = l.product_id
	JOIN shifts sh ON sh.id = COALESCE(o.shift_id, s.shift_id)
	WHERE COALESCE(o.day, s.day) >= ? AND l.unit_price_cents > 0 AND l.category NOT IN ('Pfand', 'Sonstiges')
	GROUP BY 1, 2, 3, 4, 5, 6, 7, 8`

const recapTop = 5

type RecapProduct struct {
	ProductID    int64  `json:"id"`
	Name         string `json:"name"`
	Category     string `json:"-"`
	ImageVersion int64  `json:"imageVersion"`
	Qty          int64  `json:"qty"`
}

type RecapWorker struct {
	Name   string `json:"name"`
	Room   string `json:"room"`
	Qty    int64  `json:"qty"`
	Shifts int64  `json:"shifts"`
}

// RecapNight is the business day with the most pieces sold.
type RecapNight struct {
	Day     string   `json:"day"`
	Qty     int64    `json:"qty"`
	Orders  int64    `json:"orders"`
	Workers []string `json:"workers"`
}

// RecapPeriod is one stretch of time. The lists are sorted with the best first.
type RecapPeriod struct {
	From     string         `json:"from"`
	To       string         `json:"to"`
	Qty      int64          `json:"qty"`
	Orders   int64          `json:"orders"`
	Nights   int64          `json:"nights"`
	Drinks   []RecapProduct `json:"drinks"`
	Snacks   []RecapProduct `json:"snacks"`
	Products []RecapProduct `json:"products"`
	Workers  []RecapWorker  `json:"workers"`
	Busiest  *RecapNight    `json:"busiest"`
	LastAt   int64          `json:"lastAt"` // the moment of the latest order of any night
}

type Term struct {
	Season string `json:"season"` // summer or winter
	Year   int    `json:"year"`   // the year it starts in
}

type Recap struct {
	Term         Term        `json:"term"`
	LastTerm     Term        `json:"lastTerm"`
	Week         RecapPeriod `json:"week"`
	Month        RecapPeriod `json:"month"`
	LastMonth    RecapPeriod `json:"lastMonth"`
	Semester     RecapPeriod `json:"semester"`
	LastSemester RecapPeriod `json:"lastSemester"`
}

// termOf returns the German semester around a day: summer is April to September, winter October to March.
func termOf(day time.Time) (term Term, from, to time.Time) {
	y, m := day.Year(), day.Month()
	switch {
	case m >= time.April && m <= time.September:
		return Term{"summer", y}, calendarDay(y, time.April, 1), calendarDay(y, time.September, 30)
	case m >= time.October:
		return Term{"winter", y}, calendarDay(y, time.October, 1), calendarDay(y+1, time.March, 31)
	default:
		return Term{"winter", y - 1}, calendarDay(y-1, time.October, 1), calendarDay(y, time.March, 31)
	}
}

func calendarDay(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 12, 0, 0, 0, time.Local)
}

func isoDay(t time.Time) string { return t.Format("2006-01-02") }

type recapRow struct {
	day     string
	shift   int64
	room    string
	worker  string
	key     string
	product RecapProduct
	qty     int64
}

type recapOrder struct {
	day string
	at  time.Time
}

func (s *Store) Recap(ctx context.Context) (*Recap, error) {
	today := businessDay(s.now())
	day, err := time.ParseInLocation("2006-01-02", today, time.Local)
	if err != nil {
		return nil, err
	}
	term, termFrom, _ := termOf(day)
	lastTerm, lastFrom, lastTo := termOf(termFrom.AddDate(0, 0, -1))
	first := isoDay(lastFrom) // the last semester always reaches back further than last month
	month := calendarDay(day.Year(), day.Month(), 1)

	rows, err := s.recapRows(ctx, first)
	if err != nil {
		return nil, err
	}
	orders, err := s.recapOrders(ctx, first)
	if err != nil {
		return nil, err
	}
	return &Recap{
		Term:         term,
		LastTerm:     lastTerm,
		Week:         summarise(rows, orders, isoDay(day.AddDate(0, 0, -6)), today),
		Month:        summarise(rows, orders, isoDay(month), today),
		LastMonth:    summarise(rows, orders, isoDay(month.AddDate(0, -1, 0)), isoDay(month.AddDate(0, 0, -1))),
		Semester:     summarise(rows, orders, isoDay(termFrom), today),
		LastSemester: summarise(rows, orders, first, isoDay(lastTo)),
	}, nil
}

func (s *Store) recapRows(ctx context.Context, from string) ([]recapRow, error) {
	rows, err := s.db.QueryContext(ctx, recapLines, from)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []recapRow{}
	for rows.Next() {
		var r recapRow
		if err := rows.Scan(&r.day, &r.shift, &r.room, &r.worker, &r.product.Name, &r.product.Category,
			&r.product.ProductID, &r.product.ImageVersion, &r.qty); err != nil {
			return nil, err
		}
		// People type their own name at the start of a shift, so Jonas and jonas from one room are one person.
		r.key = r.room + "|" + strings.ToLower(strings.TrimSpace(r.worker))
		out = append(out, r)
	}
	return out, rows.Err()
}

// recapOrders lists the sales that still stand. A pure deposit refund is not an order.
func (s *Store) recapOrders(ctx context.Context, from string) ([]recapOrder, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT s.day, s.created_at FROM sales s
		WHERE s.day >= ? AND s.voids_sale_id IS NULL AND s.total_cents > 0
		AND NOT EXISTS (SELECT 1 FROM sales v WHERE v.voids_sale_id = s.id)`, from)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []recapOrder{}
	for rows.Next() {
		var o recapOrder
		var ms int64
		if err := rows.Scan(&o.day, &ms); err != nil {
			return nil, err
		}
		o.at = time.UnixMilli(ms).Local()
		out = append(out, o)
	}
	return out, rows.Err()
}

func summarise(rows []recapRow, orders []recapOrder, from, to string) RecapPeriod {
	p := RecapPeriod{From: from, To: to, Drinks: []RecapProduct{}, Snacks: []RecapProduct{}, Products: []RecapProduct{}, Workers: []RecapWorker{}}
	inside := func(day string) bool { return day >= from && day <= to }

	products := map[string]*RecapProduct{}
	workers := map[string]*RecapWorker{}
	shifts := map[string]map[int64]bool{}
	nights := map[string]int64{}
	crew := map[string]map[string]bool{}
	for _, r := range rows {
		if !inside(r.day) {
			continue
		}
		p.Qty += r.qty
		nights[r.day] += r.qty

		item := products[r.product.Name]
		if item == nil {
			copied := r.product
			copied.Qty = 0
			item = &copied
			products[r.product.Name] = item
		}
		item.Qty += r.qty

		worker := workers[r.key]
		if worker == nil {
			worker = &RecapWorker{Name: strings.TrimSpace(r.worker), Room: r.room}
			workers[r.key] = worker
			shifts[r.key] = map[int64]bool{}
		}
		worker.Qty += r.qty
		shifts[r.key][r.shift] = true

		if crew[r.day] == nil {
			crew[r.day] = map[string]bool{}
		}
		crew[r.day][worker.Name] = true
	}

	// A night counts when something stayed sold. A sale reversed in full leaves nothing behind.
	for _, qty := range nights {
		if qty > 0 {
			p.Nights++
		}
	}
	rankProducts(&p, products)
	rankWorkers(&p, workers, shifts)

	var busiest string
	for day, qty := range nights {
		if qty > 0 && (busiest == "" || qty > nights[busiest] || (qty == nights[busiest] && day > busiest)) {
			busiest = day
		}
	}
	if busiest != "" {
		p.Busiest = &RecapNight{Day: busiest, Qty: nights[busiest], Workers: []string{}}
		for name := range crew[busiest] {
			p.Busiest.Workers = append(p.Busiest.Workers, name)
		}
		sort.Strings(p.Busiest.Workers)
	}

	var latest time.Duration
	for _, o := range orders {
		if !inside(o.day) {
			continue
		}
		p.Orders++
		if p.Busiest != nil && o.day == p.Busiest.Day {
			p.Busiest.Orders++
		}
		// Measured from 06:00, so 01:00 is later than 23:00.
		shifted := o.at.Add(-6 * time.Hour)
		past := time.Duration(shifted.Hour())*time.Hour + time.Duration(shifted.Minute())*time.Minute + time.Duration(shifted.Second())*time.Second
		if past >= latest {
			latest, p.LastAt = past, o.at.UnixMilli()
		}
	}
	return p
}

func rankProducts(p *RecapPeriod, products map[string]*RecapProduct) {
	ranked := make([]RecapProduct, 0, len(products))
	for _, item := range products {
		if item.Qty > 0 {
			ranked = append(ranked, *item)
		}
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].Qty != ranked[j].Qty {
			return ranked[i].Qty > ranked[j].Qty
		}
		return ranked[i].Name < ranked[j].Name
	})
	for _, item := range ranked {
		if len(p.Products) < recapTop {
			p.Products = append(p.Products, item)
		}
		if item.Category == "Snacks" {
			if len(p.Snacks) < recapTop {
				p.Snacks = append(p.Snacks, item)
			}
		} else if len(p.Drinks) < recapTop {
			p.Drinks = append(p.Drinks, item)
		}
	}
}

// rankWorkers keeps everybody, so the leaderboard can list as many as fit on the screen.
func rankWorkers(p *RecapPeriod, workers map[string]*RecapWorker, shifts map[string]map[int64]bool) {
	for key, worker := range workers {
		worker.Shifts = int64(len(shifts[key]))
		if worker.Qty > 0 {
			p.Workers = append(p.Workers, *worker)
		}
	}
	sort.Slice(p.Workers, func(i, j int) bool {
		if p.Workers[i].Qty != p.Workers[j].Qty {
			return p.Workers[i].Qty > p.Workers[j].Qty
		}
		return p.Workers[i].Name < p.Workers[j].Name
	})
}
