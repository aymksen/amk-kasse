package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

type RentalItem struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Code         string `json:"code"`
	FeeCents     int64  `json:"feeCents"`
	FeePer       string `json:"feePer"` // flat or hour
	DepositCents int64  `json:"depositCents"`
	Active       bool   `json:"active"`
}

type Rental struct {
	ID           int64  `json:"id"`
	ItemID       int64  `json:"itemId"`
	ItemName     string `json:"itemName"`
	RenterName   string `json:"renterName"`
	RenterRoom   string `json:"renterRoom"`
	DepositCents int64  `json:"depositCents"`
	DepositNote  string `json:"depositNote"`
	StartedAt    int64  `json:"startedAt"`
	ReturnedAt   *int64 `json:"returnedAt"`
	FeeCents     int64  `json:"feeCents"`
}

// BoardItem is one tile on the rental board. Rental is set while the item is out.
type BoardItem struct {
	RentalItem
	Rental *Rental `json:"rental"`
}

func (s *Store) RentalItems(ctx context.Context) ([]RentalItem, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, name, COALESCE(code, ''), fee_cents, fee_per, deposit_cents, active FROM rental_items ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []RentalItem{}
	for rows.Next() {
		var r RentalItem
		if err := rows.Scan(&r.ID, &r.Name, &r.Code, &r.FeeCents, &r.FeePer, &r.DepositCents, &r.Active); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) SaveRentalItem(ctx context.Context, r RentalItem) (RentalItem, error) {
	r.Name = strings.TrimSpace(r.Name)
	if r.Name == "" || r.FeeCents < 0 || r.DepositCents < 0 || (r.FeePer != "flat" && r.FeePer != "hour") {
		return r, ErrInvalid
	}
	var err error
	if r.ID == 0 {
		var res sql.Result
		res, err = s.db.ExecContext(ctx,
			"INSERT INTO rental_items(name, code, fee_cents, fee_per, deposit_cents, active) VALUES (?, ?, ?, ?, ?, ?)",
			r.Name, nullable(r.Code), r.FeeCents, r.FeePer, r.DepositCents, boolInt(r.Active))
		if err == nil {
			r.ID, err = res.LastInsertId()
		}
	} else {
		_, err = s.db.ExecContext(ctx,
			"UPDATE rental_items SET name = ?, code = ?, fee_cents = ?, fee_per = ?, deposit_cents = ?, active = ? WHERE id = ?",
			r.Name, nullable(r.Code), r.FeeCents, r.FeePer, r.DepositCents, boolInt(r.Active), r.ID)
	}
	if isUniqueViolation(err) {
		return r, ErrDuplicateCode
	}
	return r, err
}

const rentalColumns = `r.id, r.item_id, i.name, r.renter_name, r.renter_room, r.deposit_cents,
	r.deposit_note, r.started_at, r.returned_at, r.fee_cents`

func scanRental(row scanner) (Rental, error) {
	var r Rental
	err := row.Scan(&r.ID, &r.ItemID, &r.ItemName, &r.RenterName, &r.RenterRoom, &r.DepositCents,
		&r.DepositNote, &r.StartedAt, &r.ReturnedAt, &r.FeeCents)
	return r, err
}

// RentalBoard lists active items and who currently has them.
func (s *Store) RentalBoard(ctx context.Context) ([]BoardItem, error) {
	items, err := s.RentalItems(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, "SELECT "+rentalColumns+
		" FROM rentals r JOIN rental_items i ON i.id = r.item_id WHERE r.returned_at IS NULL")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int64]*Rental{}
	for rows.Next() {
		r, err := scanRental(rows)
		if err != nil {
			return nil, err
		}
		out[r.ItemID] = &r
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	board := make([]BoardItem, 0, len(items))
	for _, it := range items {
		// A retired item stays visible until it has been brought back.
		if it.Active || out[it.ID] != nil {
			board = append(board, BoardItem{RentalItem: it, Rental: out[it.ID]})
		}
	}
	return board, nil
}

// RentalHistory lists past and current rentals, newest first.
func (s *Store) RentalHistory(ctx context.Context, limit int) ([]Rental, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+rentalColumns+
		" FROM rentals r JOIN rental_items i ON i.id = r.item_id ORDER BY r.id DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Rental{}
	for rows.Next() {
		r, err := scanRental(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// StartRental hands out an item. A cash deposit goes into the box until the return.
func (s *Store) StartRental(ctx context.Context, itemID int64, name, room string, depositCents int64, depositNote string) error {
	name, room = strings.TrimSpace(name), strings.TrimSpace(room)
	if name == "" || room == "" || depositCents < 0 {
		return ErrInvalid
	}
	return s.inTx(ctx, func(tx *sql.Tx) error {
		shiftID, err := currentShiftID(ctx, tx)
		if err != nil {
			return err
		}
		var itemName string
		if err := tx.QueryRowContext(ctx, "SELECT name FROM rental_items WHERE id = ?", itemID).Scan(&itemName); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrNotFound
			}
			return err
		}
		at := s.now().UnixMilli()
		_, err = tx.ExecContext(ctx, `INSERT INTO rentals(item_id, shift_id, renter_name, renter_room, deposit_cents, deposit_note, started_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)`, itemID, shiftID, name, room, depositCents, strings.TrimSpace(depositNote), at)
		if isUniqueViolation(err) {
			return ErrItemRentedOut
		}
		if err != nil {
			return err
		}
		if depositCents > 0 {
			_, err = tx.ExecContext(ctx,
				"INSERT INTO cash_moves(shift_id, kind, amount_cents, note, created_at) VALUES (?, 'rental_deposit', ?, ?, ?)",
				shiftID, depositCents, itemName+" – "+name, at)
		}
		return err
	})
}

// RentalQuote is what the worker sees before confirming a return.
type RentalQuote struct {
	FeeCents     int64 `json:"feeCents"`
	DepositCents int64 `json:"depositCents"`
	Minutes      int64 `json:"minutes"`
}

func hourlyFee(feeCents, startedAt, endedAt int64) int64 {
	const hour = 60 * 60 * 1000
	hours := (endedAt - startedAt + hour - 1) / hour
	if hours < 1 {
		hours = 1
	}
	return hours * feeCents
}

func (s *Store) quote(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, rentalID int64) (RentalQuote, string, error) {
	var q2 RentalQuote
	var fee, startedAt int64
	var per, itemName string
	var returned *int64
	err := q.QueryRowContext(ctx, `SELECT i.fee_cents, i.fee_per, i.name, r.started_at, r.deposit_cents, r.returned_at
		FROM rentals r JOIN rental_items i ON i.id = r.item_id WHERE r.id = ?`, rentalID).
		Scan(&fee, &per, &itemName, &startedAt, &q2.DepositCents, &returned)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && returned != nil) {
		return q2, "", ErrNotFound
	}
	if err != nil {
		return q2, "", err
	}
	at := s.now().UnixMilli()
	q2.Minutes = (at - startedAt) / 60000
	q2.FeeCents = fee
	if per == "hour" {
		q2.FeeCents = hourlyFee(fee, startedAt, at)
	}
	return q2, itemName, nil
}

func (s *Store) QuoteReturn(ctx context.Context, rentalID int64) (RentalQuote, error) {
	q, _, err := s.quote(ctx, s.db, rentalID)
	return q, err
}

// ReturnRental ends a rental: the fee becomes a sale and the deposit goes back out.
func (s *Store) ReturnRental(ctx context.Context, rentalID int64, payment string) error {
	if payment != PayCash && payment != PayOther {
		return ErrInvalid
	}
	return s.inTx(ctx, func(tx *sql.Tx) error {
		shiftID, err := currentShiftID(ctx, tx)
		if err != nil {
			return err
		}
		q, itemName, err := s.quote(ctx, tx, rentalID)
		if err != nil {
			return err
		}
		at := s.now().UnixMilli()
		if _, err := tx.ExecContext(ctx, "UPDATE rentals SET returned_at = ?, return_shift_id = ?, fee_cents = ? WHERE id = ?",
			at, shiftID, q.FeeCents, rentalID); err != nil {
			return err
		}
		if q.FeeCents > 0 {
			line := SaleLine{Name: itemName, Category: "Verleih", Qty: 1, UnitPriceCents: q.FeeCents}
			if _, err := s.insertSale(ctx, tx, shiftID, payment, []SaleLine{line}, nil, ""); err != nil {
				return err
			}
		}
		if q.DepositCents > 0 {
			_, err = tx.ExecContext(ctx,
				"INSERT INTO cash_moves(shift_id, kind, amount_cents, note, created_at) VALUES (?, 'rental_refund', ?, ?, ?)",
				shiftID, -q.DepositCents, itemName, at)
		}
		return err
	})
}
