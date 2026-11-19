package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

type CountLine struct {
	ProductID int64 `json:"productId"`
	Counted   int64 `json:"counted"`
}

type Stocktake struct {
	ID        int64  `json:"id"`
	ShiftID   *int64 `json:"shiftId"`
	Kind      string `json:"kind"`
	Place     string `json:"place"`
	ByRoom    string `json:"byRoom"`
	ByName    string `json:"byName"`
	CreatedAt int64  `json:"createdAt"`
	Opening   bool   `json:"opening"` // the check at the start of a shift
	Lines     int64  `json:"lines"`
	Diffs     int64  `json:"diffs"`     // lines where count and expectation disagree
	DiffCents int64  `json:"diffCents"` // value of the difference at cost
}

type StocktakeLine struct {
	ProductID     int64  `json:"productId"`
	Name          string `json:"name"`
	Expected      int64  `json:"expected"`
	Counted       int64  `json:"counted"`
	UnitCostCents int64  `json:"unitCostCents"`
}

// SaveStocktake stores a count of one place and corrects it to what was physically found.
// Only products the worker touched are sent, so a partial count is valid.
func (s *Store) SaveStocktake(ctx context.Context, kind, place, room, name string, lines []CountLine) (int64, error) {
	if _, err := placeColumn(place); err != nil {
		return 0, err
	}
	if len(lines) == 0 || (kind != "quick" && kind != "full") {
		return 0, ErrInvalid
	}
	var id int64
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		var shiftID any
		if sid, err := currentShiftID(ctx, tx); err == nil {
			shiftID = sid
		} else if !errors.Is(err, ErrNoOpenShift) {
			return err
		}
		var err error
		id, err = saveCount(ctx, tx, shiftID, kind, place, room, name, false, lines, s.now().UnixMilli())
		return err
	})
	return id, err
}

// saveCount writes a count and corrects the stock to it, inside the caller's transaction.
func saveCount(ctx context.Context, tx *sql.Tx, shiftID any, kind, place, room, name string, opening bool, lines []CountLine, at int64) (int64, error) {
	if len(lines) == 0 {
		return 0, ErrInvalid
	}
	res, err := tx.ExecContext(ctx,
		"INSERT INTO stocktakes(shift_id, kind, place, by_room, by_name, created_at, opening) VALUES (?, ?, ?, ?, ?, ?, ?)",
		shiftID, kind, place, strings.TrimSpace(room), strings.TrimSpace(name), at, boolInt(opening))
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	for _, l := range lines {
		if l.Counted < 0 {
			return 0, ErrInvalid
		}
		p, err := getProduct(ctx, tx, l.ProductID)
		if err != nil {
			return 0, err
		}
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO stocktake_lines(stocktake_id, product_id, expected, counted, unit_cost_cents) VALUES (?, ?, ?, ?, ?)",
			id, p.ID, p.stockAt(place), l.Counted, p.CostCents); err != nil {
			return 0, err
		}
		if err := applyStock(ctx, tx, p.ID, place, l.Counted-p.stockAt(place), "count", id, at); err != nil {
			return 0, err
		}
	}
	return id, nil
}

func (s *Store) Stocktakes(ctx context.Context, limit int) ([]Stocktake, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT t.id, t.shift_id, t.kind, t.place, t.by_room, t.by_name, t.created_at, t.opening,
		COUNT(l.product_id),
		COALESCE(SUM(CASE WHEN l.counted <> l.expected THEN 1 ELSE 0 END), 0),
		COALESCE(SUM((l.counted - l.expected) * l.unit_cost_cents), 0)
		FROM stocktakes t LEFT JOIN stocktake_lines l ON l.stocktake_id = t.id
		GROUP BY t.id ORDER BY t.id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Stocktake{}
	for rows.Next() {
		var t Stocktake
		if err := rows.Scan(&t.ID, &t.ShiftID, &t.Kind, &t.Place, &t.ByRoom, &t.ByName, &t.CreatedAt, &t.Opening, &t.Lines, &t.Diffs, &t.DiffCents); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) StocktakeLines(ctx context.Context, id int64) ([]StocktakeLine, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT l.product_id, p.name, l.expected, l.counted, l.unit_cost_cents
		FROM stocktake_lines l JOIN products p ON p.id = l.product_id
		WHERE l.stocktake_id = ? ORDER BY p.sort, p.name`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []StocktakeLine{}
	for rows.Next() {
		var l StocktakeLine
		if err := rows.Scan(&l.ProductID, &l.Name, &l.Expected, &l.Counted, &l.UnitCostCents); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

type RestockLine struct {
	ProductID     int64  `json:"productId"`
	Name          string `json:"name"`
	Qty           int64  `json:"qty"`
	UnitCostCents int64  `json:"unitCostCents"`
}

type Restock struct {
	ID             int64         `json:"id"`
	CreatedAt      int64         `json:"createdAt"`
	Place          string        `json:"place"`
	Supplier       string        `json:"supplier"`
	Note           string        `json:"note"`
	TotalCostCents int64         `json:"totalCostCents"`
	Lines          []RestockLine `json:"lines"`
}

// SaveRestock adds delivered goods to a place, normally the storeroom. A given unit cost also
// becomes the product's current cost, which keeps margins and stock value up to date.
func (s *Store) SaveRestock(ctx context.Context, supplier, note, place string, lines []RestockLine) (int64, error) {
	if _, err := placeColumn(place); err != nil {
		return 0, err
	}
	if len(lines) == 0 {
		return 0, ErrInvalid
	}
	var id int64
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		at := s.now().UnixMilli()
		var total int64
		for _, l := range lines {
			if l.Qty <= 0 || l.UnitCostCents < 0 {
				return ErrInvalid
			}
			total += l.Qty * l.UnitCostCents
		}
		res, err := tx.ExecContext(ctx,
			"INSERT INTO restocks(created_at, place, supplier, note, total_cost_cents) VALUES (?, ?, ?, ?, ?)",
			at, place, strings.TrimSpace(supplier), strings.TrimSpace(note), total)
		if err != nil {
			return err
		}
		if id, err = res.LastInsertId(); err != nil {
			return err
		}
		for _, l := range lines {
			if _, err := getProduct(ctx, tx, l.ProductID); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx,
				"INSERT INTO restock_lines(restock_id, product_id, qty, unit_cost_cents) VALUES (?, ?, ?, ?)",
				id, l.ProductID, l.Qty, l.UnitCostCents); err != nil {
				return err
			}
			if err := applyStock(ctx, tx, l.ProductID, place, l.Qty, "restock", id, at); err != nil {
				return err
			}
			if l.UnitCostCents > 0 {
				if _, err := tx.ExecContext(ctx, "UPDATE products SET cost_cents = ? WHERE id = ?", l.UnitCostCents, l.ProductID); err != nil {
					return err
				}
			}
		}
		return nil
	})
	return id, err
}

func (s *Store) Restocks(ctx context.Context, limit int) ([]Restock, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, created_at, place, supplier, note, total_cost_cents FROM restocks ORDER BY id DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	out := []Restock{}
	for rows.Next() {
		var r Restock
		if err := rows.Scan(&r.ID, &r.CreatedAt, &r.Place, &r.Supplier, &r.Note, &r.TotalCostCents); err != nil {
			rows.Close()
			return nil, err
		}
		out = append(out, r)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range out {
		lines, err := s.db.QueryContext(ctx, `SELECT l.product_id, p.name, l.qty, l.unit_cost_cents
			FROM restock_lines l JOIN products p ON p.id = l.product_id WHERE l.restock_id = ? ORDER BY l.id`, out[i].ID)
		if err != nil {
			return nil, err
		}
		out[i].Lines = []RestockLine{}
		for lines.Next() {
			var l RestockLine
			if err := lines.Scan(&l.ProductID, &l.Name, &l.Qty, &l.UnitCostCents); err != nil {
				lines.Close()
				return nil, err
			}
			out[i].Lines = append(out[i].Lines, l)
		}
		lines.Close()
		if err := lines.Err(); err != nil {
			return nil, err
		}
	}
	return out, nil
}
