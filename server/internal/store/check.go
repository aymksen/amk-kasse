package store

import (
	"context"
	"database/sql"
	"errors"
	"math"
)

// The check of the bar (version 9). Every count at the till starts from what the bar should hold, product by
// product, and the worker only corrects what is different: at the start of a shift, any time during it, and at
// the end. A gap then shows up at the count where it was found.
const openingSQL = `ALTER TABLE stocktakes ADD COLUMN opening INTEGER NOT NULL DEFAULT 0;`

// kindClosing is the count at the end of a shift. The name is older than the check: the end of a shift once
// counted only the products marked for a quick count, and the table allows only 'quick' and 'full'.
const kindClosing = "quick"

// CheckLine is one product: how the bar stood at Check.Since, what happened since, and where that leaves it.
// Start - Sold + Added + Corrected = Expected.
type CheckLine struct {
	ProductID int64 `json:"productId"`
	Start     int64 `json:"start"`     // what the bar held at Check.Since, after the count made right then
	Expected  int64 `json:"expected"`  // what the bar should hold now
	Sold      int64 `json:"sold"`      // what went over the counter since, after cancellations
	Added     int64 `json:"added"`     // what the OBK put into the bar since
	Corrected int64 `json:"corrected"` // what counts in between changed, like Zählen from the dock
}

// LastShift is who handed the bar over.
type LastShift struct {
	Room     string `json:"room"`
	Name     string `json:"name"`
	ClosedAt int64  `json:"closedAt"`
	Counted  bool   `json:"counted"` // whether they counted the bar during or at the end of their shift
}

type Check struct {
	Last  *LastShift  `json:"last"`  // before a shift starts: who worked last, nil on a new till
	Since int64       `json:"since"` // the start of the open shift, or the end of the last one; 0 on a new till
	Lines []CheckLine `json:"lines"`
}

// StockCheck is what a count at the till starts from: every product the bar counts, with what it should hold
// and what changed it since the shift began (or, before a shift, since the last one ended).
func (s *Store) StockCheck(ctx context.Context) (Check, error) {
	out := Check{Lines: []CheckLine{}}
	since := int64(math.MaxInt64) // nothing counts as added on a new till
	open, err := s.CurrentShift(ctx)
	if err != nil {
		return out, err
	}
	if open != nil {
		since = open.OpenedAt
	} else {
		var last LastShift
		var lastID int64
		err := s.db.QueryRowContext(ctx, `SELECT id, worker_room, worker_name, closed_at FROM shifts
			WHERE closed_at IS NOT NULL ORDER BY closed_at DESC LIMIT 1`).Scan(&lastID, &last.Room, &last.Name, &last.ClosedAt)
		switch {
		case errors.Is(err, sql.ErrNoRows):
		case err != nil:
			return out, err
		default:
			if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) > 0 FROM stocktakes WHERE shift_id = ? AND place = ? AND opening = 0",
				lastID, PlaceBar).Scan(&last.Counted); err != nil {
				return out, err
			}
			out.Last = &last
			since = last.ClosedAt
		}
	}
	if since != math.MaxInt64 {
		out.Since = since
	}
	// The count at the start of the shift, or at the end of the last one, carries the time of Since itself.
	// It is where the figures start, so it is not one of the changes since.
	rows, err := s.db.QueryContext(ctx, `SELECT p.id, p.bar_stock,
		COALESCE(-SUM(CASE WHEN m.kind IN ('sale', 'void') THEN m.delta END), 0),
		COALESCE(SUM(CASE WHEN m.kind IN ('refill', 'restock') THEN m.delta END), 0),
		COALESCE(SUM(CASE WHEN m.kind = 'count' THEN m.delta END), 0)
		FROM products p LEFT JOIN stock_moves m ON m.product_id = p.id AND m.place = ?
			AND (m.created_at > ? OR (m.created_at = ? AND m.kind <> 'count'))
		WHERE p.active = 1 AND p.track_stock = 1 GROUP BY p.id ORDER BY p.sort, p.name`, PlaceBar, since, since)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var l CheckLine
		if err := rows.Scan(&l.ProductID, &l.Expected, &l.Sold, &l.Added, &l.Corrected); err != nil {
			return out, err
		}
		l.Start = l.Expected + l.Sold - l.Added - l.Corrected
		out.Lines = append(out.Lines, l)
	}
	return out, rows.Err()
}
