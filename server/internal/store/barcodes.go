package store

import (
	"context"
	"database/sql"
	"strings"
)

// Several barcodes per product (version 9). One key such as Chips stands for several kinds, and every bag
// has its own code, so each of those codes has to find the key. The bottle codes moved from
// products.barcode into their own table, where a code belongs to one product only. The crate code stays
// one per product.
const barcodesSQL = `
CREATE TABLE product_barcodes (
  code       TEXT PRIMARY KEY,
  product_id INTEGER NOT NULL REFERENCES products(id)
);
CREATE INDEX product_barcodes_product ON product_barcodes(product_id);
INSERT INTO product_barcodes(code, product_id) SELECT barcode, id FROM products WHERE barcode IS NOT NULL;
DROP INDEX products_barcode;
ALTER TABLE products DROP COLUMN barcode;`

type querier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// barcodesOf returns the codes of every product, each list in the order the codes were added.
func barcodesOf(ctx context.Context, q querier) (map[int64][]string, error) {
	rows, err := q.QueryContext(ctx, "SELECT product_id, code FROM product_barcodes ORDER BY rowid")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int64][]string{}
	for rows.Next() {
		var id int64
		var code string
		if err := rows.Scan(&id, &code); err != nil {
			return nil, err
		}
		out[id] = append(out[id], code)
	}
	return out, rows.Err()
}

// cleanCodes trims the codes and drops empty ones and repeats, keeping the order.
func cleanCodes(codes []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, c := range codes {
		if c = strings.TrimSpace(c); c != "" && !seen[c] {
			seen[c] = true
			out = append(out, c)
		}
	}
	return out
}

// setBarcodes gives a product exactly these codes. A code is either one product's bottle code or one
// product's crate code, never both, so a scan always means one thing.
func setBarcodes(ctx context.Context, tx *sql.Tx, id int64, codes []string) error {
	if _, err := tx.ExecContext(ctx, "DELETE FROM product_barcodes WHERE product_id = ?", id); err != nil {
		return err
	}
	for _, code := range codes {
		var crates int
		if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM products WHERE crate_barcode = ?", code).Scan(&crates); err != nil {
			return err
		}
		if crates > 0 {
			return ErrDuplicateCode
		}
		_, err := tx.ExecContext(ctx, "INSERT INTO product_barcodes(code, product_id) VALUES (?, ?)", code, id)
		if isUniqueViolation(err) {
			return ErrDuplicateCode
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// crateCodeTaken tells whether a crate code is already some product's bottle code.
func crateCodeTaken(ctx context.Context, tx *sql.Tx, code string) (bool, error) {
	if code = strings.TrimSpace(code); code == "" {
		return false, nil
	}
	var n int
	err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM product_barcodes WHERE code = ?", code).Scan(&n)
	return n > 0, err
}
