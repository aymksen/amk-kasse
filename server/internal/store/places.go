package store

import (
	"context"
	"database/sql"
)

// Goods sit in one of two places. The bar is everything behind it, the small stock and the
// crates included: sales take from there and workers count it. The storeroom is the OBK's
// alone: deliveries arrive there and a refill carries them across.
const (
	PlaceBar     = "bar"
	PlaceStorage = "storage"
)

// placesSQL is schema version 3. The split is a first guess for the OBK to correct by
// counting: one crate at the bar, the rest in the storeroom.
const placesSQL = `
ALTER TABLE products RENAME COLUMN stock TO bar_stock;
ALTER TABLE products ADD COLUMN storage_stock INTEGER NOT NULL DEFAULT 0;
ALTER TABLE products ADD COLUMN bar_target INTEGER NOT NULL DEFAULT 0;
UPDATE products SET bar_target = CASE WHEN crate_size > 1 THEN crate_size ELSE MAX(min_stock, 1) END
  WHERE track_stock = 1;
UPDATE products SET storage_stock = MAX(bar_stock - bar_target, 0), bar_stock = MIN(bar_stock, bar_target)
  WHERE track_stock = 1;

CREATE TABLE stock_moves_v3 (
  id         INTEGER PRIMARY KEY,
  product_id INTEGER NOT NULL REFERENCES products(id),
  place      TEXT NOT NULL DEFAULT 'bar' CHECK (place IN ('bar', 'storage')),
  delta      INTEGER NOT NULL,
  kind       TEXT NOT NULL CHECK (kind IN ('sale', 'void', 'restock', 'count', 'refill')),
  ref_id     INTEGER,
  created_at INTEGER NOT NULL
);
INSERT INTO stock_moves_v3(id, product_id, delta, kind, ref_id, created_at)
  SELECT id, product_id, delta, kind, ref_id, created_at FROM stock_moves;
DROP TABLE stock_moves;
ALTER TABLE stock_moves_v3 RENAME TO stock_moves;
CREATE INDEX stock_moves_product ON stock_moves(product_id);

ALTER TABLE stocktakes ADD COLUMN place TEXT NOT NULL DEFAULT 'bar';
ALTER TABLE restocks ADD COLUMN place TEXT NOT NULL DEFAULT 'storage';

CREATE TABLE refills (
  id         INTEGER PRIMARY KEY,
  product_id INTEGER NOT NULL REFERENCES products(id),
  qty        INTEGER NOT NULL CHECK (qty > 0),
  created_at INTEGER NOT NULL
);
`

func placeColumn(place string) (string, error) {
	switch place {
	case PlaceBar:
		return "bar_stock", nil
	case PlaceStorage:
		return "storage_stock", nil
	}
	return "", ErrInvalid
}

type RefillLine struct {
	ProductID int64 `json:"productId"`
	Qty       int64 `json:"qty"`
}

type Refill struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Qty       int64  `json:"qty"`
	CreatedAt int64  `json:"createdAt"`
}

// SaveRefill books goods the OBK carried from the storeroom to the bar. Like a sale, it is
// never blocked by a number: if the goods are in someone's hands, the storeroom figure was
// wrong, and the next count corrects it.
func (s *Store) SaveRefill(ctx context.Context, lines []RefillLine) error {
	if len(lines) == 0 {
		return ErrInvalid
	}
	return s.inTx(ctx, func(tx *sql.Tx) error {
		at := s.now().UnixMilli()
		for _, l := range lines {
			p, err := getProduct(ctx, tx, l.ProductID)
			if err != nil {
				return err
			}
			if l.Qty <= 0 || !p.TrackStock {
				return ErrInvalid
			}
			res, err := tx.ExecContext(ctx, "INSERT INTO refills(product_id, qty, created_at) VALUES (?, ?, ?)", p.ID, l.Qty, at)
			if err != nil {
				return err
			}
			id, err := res.LastInsertId()
			if err != nil {
				return err
			}
			if err := applyStock(ctx, tx, p.ID, PlaceStorage, -l.Qty, "refill", id, at); err != nil {
				return err
			}
			if err := applyStock(ctx, tx, p.ID, PlaceBar, l.Qty, "refill", id, at); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) Refills(ctx context.Context, limit int) ([]Refill, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT r.id, p.name, r.qty, r.created_at
		FROM refills r JOIN products p ON p.id = r.product_id ORDER BY r.id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Refill{}
	for rows.Next() {
		var r Refill
		if err := rows.Scan(&r.ID, &r.Name, &r.Qty, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
