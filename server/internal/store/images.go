package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
)

// MaxImageBytes is far above what the admin page sends, because it shrinks every picture
// before uploading. The limit only stops a file that skipped that step.
const MaxImageBytes = 256 << 10

// Pictures live in the database, so a backup covers them. They sit in their own table,
// so reading the product list never loads them. image_version 0 means no picture.
const imagesSQL = `
ALTER TABLE products ADD COLUMN image_version INTEGER NOT NULL DEFAULT 0;
CREATE TABLE product_images (
  product_id INTEGER PRIMARY KEY REFERENCES products(id),
  mime       TEXT NOT NULL,
  data       BLOB NOT NULL
);`

func (s *Store) ProductImage(ctx context.Context, id int64) (string, []byte, error) {
	var mime string
	var data []byte
	err := s.db.QueryRowContext(ctx, "SELECT mime, data FROM product_images WHERE product_id = ?", id).Scan(&mime, &data)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil, ErrNotFound
	}
	return mime, data, err
}

// SetProductImage returns the picture's version. It is the upload time and never repeats,
// because browsers keep a picture URL for a year and the version is part of that URL.
func (s *Store) SetProductImage(ctx context.Context, id int64, mime string, data []byte) (int64, error) {
	if len(data) == 0 || len(data) > MaxImageBytes {
		return 0, ErrInvalid
	}
	version := s.now().UnixMilli()
	err := s.inTx(ctx, func(tx *sql.Tx) error { return writeImage(ctx, tx, id, mime, data, version) })
	return version, err
}

func writeImage(ctx context.Context, tx *sql.Tx, id int64, mime string, data []byte, version int64) error {
	res, err := tx.ExecContext(ctx, "UPDATE products SET image_version = ? WHERE id = ?", version, id)
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); err != nil {
		return err
	} else if n == 0 {
		return ErrNotFound
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO product_images(product_id, mime, data) VALUES (?, ?, ?)
		ON CONFLICT(product_id) DO UPDATE SET mime = excluded.mime, data = excluded.data`, id, mime, data)
	return err
}

func (s *Store) DeleteProductImage(ctx context.Context, id int64) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, "DELETE FROM product_images WHERE product_id = ?", id); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, "UPDATE products SET image_version = 0 WHERE id = ?", id)
		return err
	})
}

// pictureFix is a catalog picture that changed after tills already stored the old one. old is the SHA-256 of
// the old picture as the catalog wrote it.
type pictureFix struct{ name, image, old string }

// pictureFixes (version 8): Bolten Alt now shows the bar's own photo instead of the Open Food Facts one.
var pictureFixes = []pictureFix{
	{"Bolten Alt", "bolten-alt", "5b656018fa63842ae4034aef259677a568758c22c790f4a1616aa937f7642e4e"},
}

// refreshPictures gives each fixed product the new catalog picture, but only while it still holds exactly
// the old one. A picture someone uploaded in OBK → Artikel is a different file and stays.
func refreshPictures(ctx context.Context, tx *sql.Tx, stamp int64) error {
	for i, f := range pictureFixes {
		rows, err := tx.QueryContext(ctx, `SELECT p.id, i.data FROM products p
			JOIN product_images i ON i.product_id = p.id WHERE p.name = ?`, f.name)
		if err != nil {
			return err
		}
		var stale []int64
		for rows.Next() {
			var id int64
			var data []byte
			if err := rows.Scan(&id, &data); err != nil {
				rows.Close()
				return err
			}
			if sum := sha256.Sum256(data); hex.EncodeToString(sum[:]) == f.old {
				stale = append(stale, id)
			}
		}
		if err := rows.Close(); err != nil {
			return err
		}
		if err := rows.Err(); err != nil {
			return err
		}
		if len(stale) == 0 {
			continue
		}
		data, err := catalogImages.ReadFile("catalog/" + f.image + ".webp")
		if err != nil {
			return err
		}
		// A new version, so no browser keeps showing the old picture from its cache.
		for _, id := range stale {
			if err := writeImage(ctx, tx, id, "image/webp", data, stamp+int64(i)); err != nil {
				return err
			}
		}
	}
	return nil
}
