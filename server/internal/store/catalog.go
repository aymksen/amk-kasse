package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

type Category struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Color  string `json:"color"`
	Sort   int64  `json:"sort"`
	Active bool   `json:"active"`
}

type Product struct {
	ID           int64    `json:"id"`
	CategoryID   int64    `json:"categoryId"`
	Name         string   `json:"name"`
	PriceCents   int64    `json:"priceCents"`
	CostCents    int64    `json:"costCents"`
	Barcodes     []string `json:"barcodes"` // every code that sells one, such as each kind of chips
	CrateBarcode string   `json:"crateBarcode"`
	CrateSize    int64    `json:"crateSize"`
	TrackStock   bool     `json:"trackStock"`
	BarStock     int64    `json:"barStock"`
	StorageStock int64    `json:"storageStock"`
	BarTarget    int64    `json:"barTarget"` // how many the bar should hold
	RefillQty    int64    `json:"refillQty"` // what to fetch from the storeroom now
	MinStock     int64    `json:"minStock"`  // of both places together, below it the OBK buys more
	Location     string   `json:"location"`  // the spot at the bar
	QuickCount   bool     `json:"quickCount"`
	Sort         int64    `json:"sort"`
	Active       bool     `json:"active"`
	ImageVersion int64    `json:"imageVersion"`
}

// TillProduct is what workers see. It holds a level instead of the stock number, because the counts
// during and at the end of a shift are blind. Only the check at the start of a shift shows the figures,
// through OpeningStock.
type TillProduct struct {
	ID           int64    `json:"id"`
	CategoryID   int64    `json:"categoryId"`
	Name         string   `json:"name"`
	PriceCents   int64    `json:"priceCents"`
	Barcodes     []string `json:"barcodes"`
	CrateBarcode string   `json:"crateBarcode"`
	CrateSize    int64    `json:"crateSize"`
	TrackStock   bool     `json:"trackStock"`
	Level        string   `json:"level"` // ok, low or empty
	Location     string   `json:"location"`
	QuickCount   bool     `json:"quickCount"`
	ImageVersion int64    `json:"imageVersion"`
}

// level tells the worker how the bar stands. A quarter of the target left is the warning,
// so the keys stay quiet for most of the evening.
func (p Product) level() string {
	switch {
	case !p.TrackStock:
		return "ok"
	case p.BarStock <= 0:
		return "empty"
	case p.BarStock*4 <= p.BarTarget:
		return "low"
	}
	return "ok"
}

// refillQty stays 0 until the bar is half empty, so nobody walks for a single bottle.
func (p Product) refillQty() int64 {
	if !p.TrackStock || p.StorageStock <= 0 || p.BarStock*2 > p.BarTarget {
		return 0
	}
	return min(p.BarTarget-p.BarStock, p.StorageStock)
}

func (p Product) stockAt(place string) int64 {
	if place == PlaceStorage {
		return p.StorageStock
	}
	return p.BarStock
}

const productColumns = `id, category_id, name, price_cents, cost_cents,
	COALESCE(crate_barcode, ''), crate_size, track_stock, bar_stock, storage_stock, bar_target, min_stock, location, quick_count, sort, active, image_version`

type scanner interface{ Scan(...any) error }

func scanProduct(r scanner) (Product, error) {
	p := Product{Barcodes: []string{}}
	err := r.Scan(&p.ID, &p.CategoryID, &p.Name, &p.PriceCents, &p.CostCents, &p.CrateBarcode,
		&p.CrateSize, &p.TrackStock, &p.BarStock, &p.StorageStock, &p.BarTarget, &p.MinStock, &p.Location, &p.QuickCount,
		&p.Sort, &p.Active, &p.ImageVersion)
	p.RefillQty = p.refillQty()
	return p, err
}

func (s *Store) Categories(ctx context.Context) ([]Category, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, name, color, sort, active FROM categories ORDER BY sort, name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Category{}
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Color, &c.Sort, &c.Active); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) SaveCategory(ctx context.Context, c Category) (Category, error) {
	c.Name = strings.TrimSpace(c.Name)
	if c.Name == "" {
		return c, ErrInvalid
	}
	if c.Color == "" {
		c.Color = "#8b93a3"
	}
	if c.ID == 0 {
		res, err := s.db.ExecContext(ctx,
			"INSERT INTO categories(name, color, sort, active) VALUES (?, ?, ?, ?)",
			c.Name, c.Color, c.Sort, boolInt(c.Active))
		if err != nil {
			return c, err
		}
		c.ID, err = res.LastInsertId()
		return c, err
	}
	_, err := s.db.ExecContext(ctx,
		"UPDATE categories SET name = ?, color = ?, sort = ?, active = ? WHERE id = ?",
		c.Name, c.Color, c.Sort, boolInt(c.Active), c.ID)
	return c, err
}

// Products returns every product, including inactive ones, for the admin area.
func (s *Store) Products(ctx context.Context) ([]Product, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+productColumns+" FROM products ORDER BY sort, name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Product{}
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()
	codes, err := barcodesOf(ctx, s.db)
	if err != nil {
		return nil, err
	}
	for i := range out {
		if c := codes[out[i].ID]; c != nil {
			out[i].Barcodes = c
		}
	}
	return out, nil
}

func (s *Store) TillProducts(ctx context.Context) ([]TillProduct, error) {
	all, err := s.Products(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]TillProduct, 0, len(all))
	for _, p := range all {
		if !p.Active {
			continue
		}
		out = append(out, TillProduct{
			ID: p.ID, CategoryID: p.CategoryID, Name: p.Name, PriceCents: p.PriceCents,
			Barcodes: p.Barcodes, CrateBarcode: p.CrateBarcode, CrateSize: p.CrateSize,
			TrackStock: p.TrackStock, Level: p.level(), Location: p.Location, QuickCount: p.QuickCount,
			ImageVersion: p.ImageVersion,
		})
	}
	return out, nil
}

// SaveProduct creates or updates a product. Stock is deliberately not writable
// here: it only changes through sales, restocks, refills and counts, so it stays explainable.
func (s *Store) SaveProduct(ctx context.Context, p Product) (Product, error) {
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" || p.CategoryID == 0 || p.CrateSize < 1 || p.CostCents < 0 || p.MinStock < 0 || p.BarTarget < 0 {
		return p, ErrInvalid
	}
	p.Barcodes = cleanCodes(p.Barcodes)
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		// The product's own codes go first, so a code can move from the bottle to the crate in one save.
		if _, err := tx.ExecContext(ctx, "DELETE FROM product_barcodes WHERE product_id = ?", p.ID); err != nil {
			return err
		}
		if taken, err := crateCodeTaken(ctx, tx, p.CrateBarcode); err != nil {
			return err
		} else if taken {
			return ErrDuplicateCode
		}
		var err error
		if p.ID == 0 {
			var res sql.Result
			res, err = tx.ExecContext(ctx, `INSERT INTO products(category_id, name, price_cents, cost_cents,
				crate_barcode, crate_size, track_stock, bar_target, min_stock, location, quick_count, sort, active)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				p.CategoryID, p.Name, p.PriceCents, p.CostCents, nullable(p.CrateBarcode),
				p.CrateSize, boolInt(p.TrackStock), p.BarTarget, p.MinStock, p.Location, boolInt(p.QuickCount), p.Sort, boolInt(p.Active))
			if err == nil {
				p.ID, err = res.LastInsertId()
			}
		} else {
			_, err = tx.ExecContext(ctx, `UPDATE products SET category_id = ?, name = ?, price_cents = ?, cost_cents = ?,
				crate_barcode = ?, crate_size = ?, track_stock = ?, bar_target = ?, min_stock = ?, location = ?,
				quick_count = ?, sort = ?, active = ? WHERE id = ?`,
				p.CategoryID, p.Name, p.PriceCents, p.CostCents, nullable(p.CrateBarcode),
				p.CrateSize, boolInt(p.TrackStock), p.BarTarget, p.MinStock, p.Location, boolInt(p.QuickCount), p.Sort, boolInt(p.Active), p.ID)
		}
		if isUniqueViolation(err) {
			return ErrDuplicateCode
		}
		if err != nil {
			return err
		}
		return setBarcodes(ctx, tx, p.ID, p.Barcodes)
	})
	return p, err
}

func getProduct(ctx context.Context, tx *sql.Tx, id int64) (Product, error) {
	p, err := scanProduct(tx.QueryRowContext(ctx, "SELECT "+productColumns+" FROM products WHERE id = ?", id))
	if errors.Is(err, sql.ErrNoRows) {
		return p, ErrNotFound
	}
	return p, err
}

// applyStock moves stock and logs why, in the same transaction.
func applyStock(ctx context.Context, tx *sql.Tx, productID int64, place string, delta int64, kind string, refID, at int64) error {
	if delta == 0 {
		return nil
	}
	column, err := placeColumn(place)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "UPDATE products SET "+column+" = "+column+" + ? WHERE id = ?", delta, productID); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx,
		"INSERT INTO stock_moves(product_id, place, delta, kind, ref_id, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		productID, place, delta, kind, refID, at)
	return err
}
