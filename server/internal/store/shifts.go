package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

type Shift struct {
	ID                 int64  `json:"id"`
	WorkerRoom         string `json:"workerRoom"`
	WorkerName         string `json:"workerName"`
	OpenedAt           int64  `json:"openedAt"`
	ClosedAt           *int64 `json:"closedAt"`
	Day                string `json:"day"`
	FloatCents         int64  `json:"floatCents"`
	ExpectedFloatCents int64  `json:"expectedFloatCents"`
	ExpectedCents      *int64 `json:"expectedCents"`
	CountedCents       *int64 `json:"countedCents"`
	Note               string `json:"note"`     // about a cash difference, for the OBK
	Handover           string `json:"handover"` // the message for the next shift
}

// ShiftMessage is what a worker left for whoever comes next.
type ShiftMessage struct {
	Text string `json:"text"`
	From string `json:"from"`
	At   int64  `json:"at"`
}

// ShiftTotals is what a shift has taken so far.
type ShiftTotals struct {
	SalesCount    int64 `json:"salesCount"`
	CashCents     int64 `json:"cashCents"`
	OtherCents    int64 `json:"otherCents"`
	CashMoveCents int64 `json:"cashMoveCents"`
	ExpectedCents int64 `json:"expectedCents"`
}

const shiftColumns = `id, worker_room, worker_name, opened_at, closed_at, day, float_cents,
	expected_float_cents, expected_cents, counted_cents, note, handover`

// handoverSQL is schema version 5.
const handoverSQL = `ALTER TABLE shifts ADD COLUMN handover TEXT NOT NULL DEFAULT ''`

// A message is a line for the next shift, not a letter.
const maxMessage = 280

func scanShift(r scanner) (Shift, error) {
	var sh Shift
	err := r.Scan(&sh.ID, &sh.WorkerRoom, &sh.WorkerName, &sh.OpenedAt, &sh.ClosedAt, &sh.Day, &sh.FloatCents,
		&sh.ExpectedFloatCents, &sh.ExpectedCents, &sh.CountedCents, &sh.Note, &sh.Handover)
	return sh, err
}

// CurrentShift returns the open shift, or nil if the bar is closed.
func (s *Store) CurrentShift(ctx context.Context) (*Shift, error) {
	sh, err := scanShift(s.db.QueryRowContext(ctx,
		"SELECT "+shiftColumns+" FROM shifts WHERE closed_at IS NULL"))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sh, nil
}

func currentShiftID(ctx context.Context, tx *sql.Tx) (int64, error) {
	var id int64
	err := tx.QueryRowContext(ctx, "SELECT id FROM shifts WHERE closed_at IS NULL").Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNoOpenShift
	}
	return id, err
}

// WorkerNames lists the names used for a room, the latest first. They are only suggestions, because a room changes
// hands: whoever lives there now may be somebody else.
func (s *Store) WorkerNames(ctx context.Context, room string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT worker_name FROM shifts WHERE worker_room = ? ORDER BY id DESC", strings.TrimSpace(room))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	names := []string{}
	seen := map[string]bool{}
	for rows.Next() && len(names) < maxSuggestedNames {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		name = strings.TrimSpace(name)
		if key := strings.ToLower(name); !seen[key] {
			seen[key] = true
			names = append(names, name)
		}
	}
	return names, rows.Err()
}

const maxSuggestedNames = 3

// boxBalance computes the cash that should be in the box while no shift is open:
// the last counted amount, plus anything moved or refunded since then.
func boxBalance(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}) (int64, error) {
	var counted, closedAt int64
	err := q.QueryRowContext(ctx, `SELECT COALESCE(counted_cents, 0), closed_at FROM shifts
		WHERE closed_at IS NOT NULL ORDER BY closed_at DESC, id DESC LIMIT 1`).Scan(&counted, &closedAt)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	var moves, voids int64
	if err := q.QueryRowContext(ctx,
		"SELECT COALESCE(SUM(amount_cents), 0) FROM cash_moves WHERE shift_id IS NULL AND created_at >= ?",
		closedAt).Scan(&moves); err != nil {
		return 0, err
	}
	if err := q.QueryRowContext(ctx,
		"SELECT COALESCE(SUM(total_cents), 0) FROM sales WHERE shift_id IS NULL AND payment = 'cash' AND created_at >= ?",
		closedAt).Scan(&voids); err != nil {
		return 0, err
	}
	return counted + moves + voids, nil
}

// ExpectedFloat is what the next worker should find in the box.
func (s *Store) ExpectedFloat(ctx context.Context) (int64, error) { return boxBalance(ctx, s.db) }

// CashInBox is the expected cash right now, with or without an open shift.
func (s *Store) CashInBox(ctx context.Context) (int64, error) {
	sh, err := s.CurrentShift(ctx)
	if err != nil {
		return 0, err
	}
	if sh == nil {
		return boxBalance(ctx, s.db)
	}
	t, err := s.ShiftTotals(ctx, *sh)
	return t.ExpectedCents, err
}

// OpenShift starts a shift. count is the worker's check of the bar right before, if there was one. It is
// saved with the shift in one step, as the shift's first count, so a shift never starts half.
func (s *Store) OpenShift(ctx context.Context, room, name string, floatCents int64, count ...CountLine) (*Shift, error) {
	room, name = strings.TrimSpace(room), strings.TrimSpace(name)
	if room == "" || name == "" || floatCents < 0 {
		return nil, ErrInvalid
	}
	var id int64
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		expected, err := boxBalance(ctx, tx)
		if err != nil {
			return err
		}
		t := s.now()
		res, err := tx.ExecContext(ctx, `INSERT INTO shifts(worker_room, worker_name, opened_at, day, float_cents, expected_float_cents)
			VALUES (?, ?, ?, ?, ?, ?)`, room, name, t.UnixMilli(), businessDay(t), floatCents, expected)
		if isUniqueViolation(err) {
			return ErrShiftOpen
		}
		if err != nil {
			return err
		}
		if id, err = res.LastInsertId(); err != nil {
			return err
		}
		if len(count) > 0 {
			_, err = saveCount(ctx, tx, id, "full", PlaceBar, room, name, true, count, t.UnixMilli())
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return s.GetShift(ctx, id)
}

func (s *Store) GetShift(ctx context.Context, id int64) (*Shift, error) {
	sh, err := scanShift(s.db.QueryRowContext(ctx, "SELECT "+shiftColumns+" FROM shifts WHERE id = ?", id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &sh, nil
}

// ShiftTotals sums a shift's sales and cash moves. Non-cash sales never touch the box.
func (s *Store) ShiftTotals(ctx context.Context, sh Shift) (ShiftTotals, error) {
	return shiftTotals(ctx, s.db, sh)
}

func shiftTotals(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, sh Shift) (ShiftTotals, error) {
	var t ShiftTotals
	err := q.QueryRowContext(ctx, `SELECT
		COALESCE(SUM(CASE WHEN voids_sale_id IS NULL THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN payment = 'cash' THEN total_cents ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN payment = 'other' THEN total_cents ELSE 0 END), 0)
		FROM sales WHERE shift_id = ?`, sh.ID).Scan(&t.SalesCount, &t.CashCents, &t.OtherCents)
	if err != nil {
		return t, err
	}
	err = q.QueryRowContext(ctx, "SELECT COALESCE(SUM(amount_cents), 0) FROM cash_moves WHERE shift_id = ?", sh.ID).
		Scan(&t.CashMoveCents)
	t.ExpectedCents = sh.FloatCents + t.CashCents + t.CashMoveCents
	return t, err
}

// LastMessage is what the shift that ended last left behind, if anything. Only that shift counts: whoever worked since
// has already seen it, and an old message on the screen would only be ignored.
func (s *Store) LastMessage(ctx context.Context) (*ShiftMessage, error) {
	var m ShiftMessage
	err := s.db.QueryRowContext(ctx, `SELECT handover, worker_name, closed_at FROM shifts
		WHERE closed_at IS NOT NULL ORDER BY closed_at DESC, id DESC LIMIT 1`).Scan(&m.Text, &m.From, &m.At)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && m.Text == "") {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func shortMessage(text string) string {
	runes := []rune(strings.TrimSpace(text))
	if len(runes) > maxMessage {
		runes = runes[:maxMessage]
	}
	return strings.TrimSpace(string(runes))
}

// CloseShift records the counted cash, keeps the message for the next shift and ends this one. count is the
// check of the bar at the end, saved in the same step, so a shift never ends half counted.
func (s *Store) CloseShift(ctx context.Context, countedCents int64, note, handover string, count ...CountLine) (*Shift, error) {
	if countedCents < 0 {
		return nil, ErrInvalid
	}
	sh, err := s.CurrentShift(ctx)
	if err != nil {
		return nil, err
	}
	if sh == nil {
		return nil, ErrNoOpenShift
	}
	totals, err := s.ShiftTotals(ctx, *sh)
	if err != nil {
		return nil, err
	}
	at := s.now().UnixMilli()
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		if len(count) > 0 {
			if _, err := saveCount(ctx, tx, sh.ID, kindClosing, PlaceBar, sh.WorkerRoom, sh.WorkerName, false, count, at); err != nil {
				return err
			}
		}
		_, err := tx.ExecContext(ctx,
			"UPDATE shifts SET closed_at = ?, expected_cents = ?, counted_cents = ?, note = ?, handover = ? WHERE id = ?",
			at, totals.ExpectedCents, countedCents, strings.TrimSpace(note), shortMessage(handover), sh.ID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return s.GetShift(ctx, sh.ID)
}

// CashMove is money entering or leaving the box outside a sale.
type CashMove struct {
	ID          int64  `json:"id"`
	ShiftID     *int64 `json:"shiftId"`
	Kind        string `json:"kind"`
	AmountCents int64  `json:"amountCents"`
	Note        string `json:"note"`
	CreatedAt   int64  `json:"createdAt"`
}

// AddCashMove books an OBK withdrawal or top-up. amountCents is always positive here.
func (s *Store) AddCashMove(ctx context.Context, kind string, amountCents int64, note string) error {
	if amountCents <= 0 || (kind != "withdrawal" && kind != "topup") {
		return ErrInvalid
	}
	if kind == "withdrawal" {
		amountCents = -amountCents
	}
	sh, err := s.CurrentShift(ctx)
	if err != nil {
		return err
	}
	var shiftID any
	if sh != nil {
		shiftID = sh.ID
	}
	_, err = s.db.ExecContext(ctx,
		"INSERT INTO cash_moves(shift_id, kind, amount_cents, note, created_at) VALUES (?, ?, ?, ?, ?)",
		shiftID, kind, amountCents, strings.TrimSpace(note), s.now().UnixMilli())
	return err
}

func (s *Store) CashMoves(ctx context.Context, limit int) ([]CashMove, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, shift_id, kind, amount_cents, note, created_at FROM cash_moves ORDER BY id DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []CashMove{}
	for rows.Next() {
		var m CashMove
		if err := rows.Scan(&m.ID, &m.ShiftID, &m.Kind, &m.AmountCents, &m.Note, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
