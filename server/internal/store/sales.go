package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

type LineInput struct {
	ProductID int64 `json:"productId"`
	Qty       int64 `json:"qty"`
}

type SaleLine struct {
	ProductID      *int64 `json:"productId"`
	Name           string `json:"name"`
	Category       string `json:"category"`
	Qty            int64  `json:"qty"`
	UnitPriceCents int64  `json:"unitPriceCents"`
	UnitCostCents  int64  `json:"unitCostCents"`
}

type Sale struct {
	ID          int64      `json:"id"`
	ShiftID     *int64     `json:"shiftId"`
	CreatedAt   int64      `json:"createdAt"`
	Day         string     `json:"day"`
	TotalCents  int64      `json:"totalCents"`
	Payment     string     `json:"payment"`
	VoidsSaleID *int64     `json:"voidsSaleId"`
	Reason      string     `json:"reason"`
	Voided      bool       `json:"voided"`
	Lines       []SaleLine `json:"lines"`
}

// insertSale writes a sale and its lines. Callers decide what happens to stock.
func (s *Store) insertSale(ctx context.Context, tx *sql.Tx, shiftID any, payment string, lines []SaleLine, voids any, reason string) (int64, error) {
	var total int64
	for _, l := range lines {
		total += l.Qty * l.UnitPriceCents
	}
	t := s.now()
	res, err := tx.ExecContext(ctx, `INSERT INTO sales(shift_id, created_at, day, total_cents, payment, voids_sale_id, reason)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, shiftID, t.UnixMilli(), businessDay(t), total, payment, voids, reason)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	for _, l := range lines {
		if _, err := tx.ExecContext(ctx, `INSERT INTO sale_lines(sale_id, product_id, name, category, qty, unit_price_cents, unit_cost_cents)
			VALUES (?, ?, ?, ?, ?, ?, ?)`, id, l.ProductID, l.Name, l.Category, l.Qty, l.UnitPriceCents, l.UnitCostCents); err != nil {
			return 0, err
		}
	}
	return id, nil
}

// CreateSale rings up a cart. Prices come from the database, never from the client,
// and are copied onto each line so a later price change cannot alter history.
func (s *Store) CreateSale(ctx context.Context, input []LineInput, payment string) (*Sale, error) {
	if len(input) == 0 {
		return nil, ErrEmptySale
	}
	if payment != PayCash && payment != PayOther {
		return nil, ErrInvalid
	}
	var saleID int64
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		shiftID, err := currentShiftID(ctx, tx)
		if err != nil {
			return err
		}
		lines := make([]SaleLine, 0, len(input))
		for _, in := range input {
			if in.Qty <= 0 || in.Qty > 999 {
				return ErrInvalid
			}
			p, err := getProduct(ctx, tx, in.ProductID)
			if err != nil {
				return err
			}
			var category string
			if err := tx.QueryRowContext(ctx, "SELECT name FROM categories WHERE id = ?", p.CategoryID).Scan(&category); err != nil {
				return err
			}
			id := p.ID
			lines = append(lines, SaleLine{ProductID: &id, Name: p.Name, Category: category, Qty: in.Qty,
				UnitPriceCents: p.PriceCents, UnitCostCents: p.CostCents})
		}
		saleID, err = s.insertSale(ctx, tx, shiftID, payment, lines, nil, "")
		if err != nil {
			return err
		}
		at := s.now().UnixMilli()
		for _, l := range lines {
			p, err := getProduct(ctx, tx, *l.ProductID)
			if err != nil {
				return err
			}
			if p.TrackStock {
				if err := applyStock(ctx, tx, p.ID, PlaceBar, -l.Qty, "sale", saleID, at); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.GetSale(ctx, saleID)
}

// VoidSale reverses a sale with a new negative sale and puts the stock back.
// With no shift open the reversal is booked against the box directly.
func (s *Store) VoidSale(ctx context.Context, saleID int64, reason string) (*Sale, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil, ErrReasonRequired
	}
	original, err := s.GetSale(ctx, saleID)
	if err != nil {
		return nil, err
	}
	if original.VoidsSaleID != nil {
		return nil, ErrNotVoidable
	}
	if original.Voided {
		return nil, ErrAlreadyVoided
	}
	var voidID int64
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		var shiftID any
		if id, err := currentShiftID(ctx, tx); err == nil {
			shiftID = id
		} else if !errors.Is(err, ErrNoOpenShift) {
			return err
		}
		reversed := make([]SaleLine, len(original.Lines))
		for i, l := range original.Lines {
			l.Qty = -l.Qty
			reversed[i] = l
		}
		voidID, err = s.insertSale(ctx, tx, shiftID, original.Payment, reversed, saleID, reason)
		if isUniqueViolation(err) {
			return ErrAlreadyVoided
		}
		if err != nil {
			return err
		}
		at := s.now().UnixMilli()
		for _, l := range original.Lines {
			if l.ProductID == nil {
				continue
			}
			p, err := getProduct(ctx, tx, *l.ProductID)
			if err != nil {
				return err
			}
			if p.TrackStock {
				if err := applyStock(ctx, tx, p.ID, PlaceBar, l.Qty, "void", voidID, at); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.GetSale(ctx, voidID)
}

const saleColumns = `s.id, s.shift_id, s.created_at, s.day, s.total_cents, s.payment, s.voids_sale_id, s.reason,
	EXISTS (SELECT 1 FROM sales v WHERE v.voids_sale_id = s.id)`

func scanSale(r scanner) (Sale, error) {
	var sale Sale
	err := r.Scan(&sale.ID, &sale.ShiftID, &sale.CreatedAt, &sale.Day, &sale.TotalCents, &sale.Payment,
		&sale.VoidsSaleID, &sale.Reason, &sale.Voided)
	return sale, err
}

func (s *Store) GetSale(ctx context.Context, id int64) (*Sale, error) {
	sale, err := scanSale(s.db.QueryRowContext(ctx, "SELECT "+saleColumns+" FROM sales s WHERE s.id = ?", id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := s.loadLines(ctx, []*Sale{&sale}); err != nil {
		return nil, err
	}
	return &sale, nil
}

func (s *Store) loadLines(ctx context.Context, sales []*Sale) error {
	for _, sale := range sales {
		rows, err := s.db.QueryContext(ctx, `SELECT product_id, name, category, qty, unit_price_cents, unit_cost_cents
			FROM sale_lines WHERE sale_id = ? ORDER BY id`, sale.ID)
		if err != nil {
			return err
		}
		sale.Lines = []SaleLine{}
		for rows.Next() {
			var l SaleLine
			if err := rows.Scan(&l.ProductID, &l.Name, &l.Category, &l.Qty, &l.UnitPriceCents, &l.UnitCostCents); err != nil {
				rows.Close()
				return err
			}
			sale.Lines = append(sale.Lines, l)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) querySales(ctx context.Context, where string, args ...any) ([]Sale, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+saleColumns+" FROM sales s "+where, args...)
	if err != nil {
		return nil, err
	}
	out := []Sale{}
	for rows.Next() {
		sale, err := scanSale(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		out = append(out, sale)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	ptrs := make([]*Sale, len(out))
	for i := range out {
		ptrs[i] = &out[i]
	}
	return out, s.loadLines(ctx, ptrs)
}

// ShiftSales lists a shift's sales, newest first.
func (s *Store) ShiftSales(ctx context.Context, shiftID int64) ([]Sale, error) {
	return s.querySales(ctx, "WHERE s.shift_id = ? ORDER BY s.id DESC", shiftID)
}

// SalesBetween lists sales by business day, newest first.
func (s *Store) SalesBetween(ctx context.Context, from, to string, limit int) ([]Sale, error) {
	return s.querySales(ctx, "WHERE s.day >= ? AND s.day <= ? ORDER BY s.id DESC LIMIT ?", from, to, limit)
}
