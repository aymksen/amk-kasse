// Package store holds every business rule of the till and all SQL.
package store

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

const schemaVersion = 9

// Errors the API layer turns into readable messages for the worker.
var (
	ErrNoOpenShift    = errors.New("no_open_shift")
	ErrShiftOpen      = errors.New("shift_already_open")
	ErrEmptySale      = errors.New("empty_sale")
	ErrNotFound       = errors.New("not_found")
	ErrAlreadyVoided  = errors.New("already_voided")
	ErrNotVoidable    = errors.New("not_voidable")
	ErrReasonRequired = errors.New("reason_required")
	ErrItemRentedOut  = errors.New("item_rented_out")
	ErrInvalid        = errors.New("invalid_input")
	ErrDuplicateCode  = errors.New("duplicate_code")
)

// Payment methods. "other" is the optional non-cash button, labelled in settings.
const (
	PayCash  = "cash"
	PayOther = "other"
)

type Store struct {
	db  *sql.DB
	dir string
	now func() time.Time
}

// Open opens or creates the database in dir.
func Open(dir string) (*Store, error) {
	dsn := "file:" + filepath.ToSlash(filepath.Join(dir, "amk-bar.db")) +
		"?_pragma=journal_mode(WAL)&_pragma=synchronous(FULL)&_pragma=foreign_keys(ON)&_pragma=busy_timeout(5000)&_pragma=cache_size(-2000)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// One till, one writer. A single connection keeps memory low and rules out lock errors.
	db.SetMaxOpenConns(1)

	s := &Store{db: db, dir: dir, now: time.Now}
	if err := s.migrate(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

// Now exposes the store clock so the API uses the same time source as the data.
func (s *Store) Now() time.Time { return s.now() }

func (s *Store) migrate(ctx context.Context) error {
	var version int
	if err := s.db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {
		return err
	}
	if version >= schemaVersion {
		return nil
	}
	return s.inTx(ctx, func(tx *sql.Tx) error {
		if version < 1 {
			if _, err := tx.ExecContext(ctx, schemaSQL); err != nil {
				return err
			}
			if err := seedBasics(ctx, tx); err != nil {
				return err
			}
		}
		if version < 2 {
			if _, err := tx.ExecContext(ctx, imagesSQL); err != nil {
				return err
			}
		}
		if version < 3 {
			if _, err := tx.ExecContext(ctx, placesSQL); err != nil {
				return err
			}
		}
		if version >= 1 && version < 4 {
			if err := updateBasics(ctx, tx); err != nil {
				return err
			}
		}
		if version < 5 {
			if _, err := tx.ExecContext(ctx, handoverSQL); err != nil {
				return err
			}
		}
		if version < 6 {
			if err := replaceRange(ctx, tx, s.now().UnixMilli()); err != nil {
				return err
			}
			if err := syncLending(ctx, tx); err != nil {
				return err
			}
		}
		if version < 7 {
			if _, err := tx.ExecContext(ctx, gamesSQL); err != nil {
				return err
			}
		}
		// Below 6 the range above already wrote the new pictures.
		if version >= 6 && version < 8 {
			if err := refreshPictures(ctx, tx, s.now().UnixMilli()); err != nil {
				return err
			}
		}
		if version < 9 {
			// The old range goes first, so its codes do not move into the new table.
			if err := dropOldRange(ctx, tx); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, barcodesSQL+openingSQL); err != nil {
				return err
			}
		}
		_, err := tx.ExecContext(ctx, fmt.Sprintf("PRAGMA user_version = %d", schemaVersion))
		return err
	})
}

func (s *Store) inTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

// businessDay maps a moment to the bar's day, which runs from 06:00 to 06:00.
func businessDay(t time.Time) string {
	return t.Local().Add(-6 * time.Hour).Format("2006-01-02")
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

func nullable(s string) any {
	if s = strings.TrimSpace(s); s == "" {
		return nil
	}
	return s
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Settings ------------------------------------------------------------------

func (s *Store) Setting(ctx context.Context, key, fallback string) (string, error) {
	var v string
	err := s.db.QueryRowContext(ctx, "SELECT value FROM settings WHERE key = ?", key).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return fallback, nil
	}
	return v, err
}

func (s *Store) SetSetting(ctx context.Context, key, value string) error {
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO settings(key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value",
		key, value)
	return err
}

func (s *Store) DeleteSetting(ctx context.Context, key string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM settings WHERE key = ?", key)
	return err
}
