package store

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// seedProduct is one line of the range. was lists names the product had in earlier versions, so an
// update finds it again and its sales stay attached.
type seedProduct struct {
	cat, name, loc string
	price          int64
	crate, target  int64 // a crate holds crate, and the bar should hold target
	stock          int64 // the last count, all of it at the bar
	quick, track   bool
	image, code    string
	was            []string
}

//go:embed catalog/*.webp
var catalogImages embed.FS

const (
	bigFridge   = "Großer Kühlschrank"
	smallFridge = "Kleiner Kühlschrank"
	shelf       = "Regal"
	freezer     = "Gefrierfach"
	cashBox     = "Kasse"
)

// catalog is the range of the real bar, as the OBK's own list showed it on 13 November 2026: names,
// prices, where each product stands and the last count. Purchase prices were not on the list and
// start at 0. The order is the list's, which the workers know.
var catalog = []seedProduct{
	{"Bier", "Franziskaner Weißbier", bigFridge, 140, 20, 20, 34, true, true, "franziskaner", "", nil},
	{"Bier", "Bosch Lager Hell", bigFridge, 120, 20, 20, 45, true, true, "bosch-lager-hell", "", nil},
	{"Bier", "Bitburger", bigFridge, 140, 20, 20, 13, true, true, "bitburger", "", nil},
	{"Bier", "Bitburger Alkoholfrei", bigFridge, 140, 20, 20, 37, true, true, "bitburger-alkoholfrei", "", nil},
	{"Bier", "Pinkus Spezial", bigFridge, 140, 20, 20, 23, true, true, "pinkus-spezial", "", nil},
	{"Bier", "Beck's", bigFridge, 120, 20, 20, 58, true, true, "becks-flasche", "", []string{"Beck's 0,5 l"}},
	{"Bier", "Früh Kölsch", bigFridge, 120, 20, 20, 36, true, true, "frueh-koelsch-flasche", "", []string{"Früh Kölsch 0,5 l"}},
	{"Bier", "Veltins", bigFridge, 140, 20, 20, 33, true, true, "veltins", "", nil},
	{"Bier", "Hansa Pils", bigFridge, 100, 20, 20, 41, true, true, "hansa-pils", "4001982198019", []string{"Hansa Pils 0,5 l"}},
	{"Bier", "Bolten Alt", bigFridge, 140, 20, 20, 38, true, true, "bolten-alt", "", nil},
	{"Bier", "Traugott Simon", bigFridge, 60, 20, 20, 25, true, true, "traugott-simon", "", nil},
	{"Bier", "Gösser Radler", bigFridge, 150, 20, 20, 21, true, true, "goesser-radler", "", nil},
	{"Bier", "Tyskie", bigFridge, 140, 20, 20, 50, true, true, "tyskie", "", nil},
	{"Alkoholfrei", "Vitamalz", bigFridge, 130, 24, 24, 28, true, true, "vitamalz", "", nil},
	{"Alkoholfrei", "Paulaner Spezi", bigFridge, 130, 20, 20, 33, true, true, "paulaner-spezi", "", nil},
	{"Alkoholfrei", "Durstlöscher", bigFridge, 100, 12, 12, 6, true, true, "durstloescher", "", nil},
	{"Alkoholfrei", "Club-Mate", bigFridge, 110, 20, 20, 22, true, true, "club-mate", "4029764001807", []string{"Club-Mate 0,5 l", "Mate 0,5 l"}},
	{"Alkoholfrei", "Coca-Cola", bigFridge, 170, 12, 12, 22, true, true, "coca-cola", "5449000017888", []string{"Coca-Cola 1 l"}},
	{"Alkoholfrei", "Coca-Cola Zero", bigFridge, 170, 12, 12, 12, true, true, "coca-cola-zero", "", nil},
	{"Alkoholfrei", "Mezzo Mix", bigFridge, 170, 12, 12, 16, true, true, "mezzo-mix", "5449000017956", []string{"Mezzo Mix 1 l"}},
	{"Alkoholfrei", "Fanta", bigFridge, 170, 12, 12, 25, true, true, "fanta", "5449000017918", []string{"Fanta 1 l"}},
	{"Alkoholfrei", "Sprite", bigFridge, 170, 12, 12, 15, true, true, "sprite", "5449000017932", []string{"Sprite 1 l"}},
	{"Alkoholfrei", "Wasser Classic (blau)", bigFridge, 100, 12, 12, 21, true, true, "wasser-classic", "", nil},
	{"Alkoholfrei", "Wasser Medium (grün)", bigFridge, 100, 12, 12, 19, true, true, "wasser-medium", "", nil},
	{"Alkoholfrei", "Fassbrause Zitrone", smallFridge, 130, 24, 24, 15, false, true, "fassbrause-zitrone-krombacher", "", nil},
	{"Alkoholfrei", "Fassbrause Rhabarber", smallFridge, 130, 24, 24, 19, false, true, "fassbrause-rhabarber", "", nil},
	{"Alkoholfrei", "Fassbrause Holunder", smallFridge, 130, 24, 24, 34, false, true, "fassbrause-holunder", "", nil},
	{"Alkoholfrei", "Fassbrause Maracuja", smallFridge, 130, 24, 24, 21, false, true, "fassbrause-maracuja", "", nil},
	{"Snacks", "Schokowaffel", shelf, 100, 1, 8, 0, false, true, "schokowaffel", "", nil},
	{"Snacks", "Schokoriegel", shelf, 50, 1, 8, 8, false, true, "schokoriegel-mix", "", []string{"Andere Schokoriegel"}},
	{"Snacks", "Milka", shelf, 80, 1, 8, 8, false, true, "milka-alpenmilch", "", []string{"Milka Alpenmilch"}},
	{"Snacks", "Haribo", shelf, 180, 1, 8, 0, false, true, "haribo-goldbaeren", "", []string{"Haribo Goldbären", "Gummibärchen"}},
	{"Snacks", "Chips", shelf, 250, 1, 8, 8, false, true, "chipsfrisch", "", []string{"Chipsfrisch ungarisch"}},
	{"Snacks", "Salzstangen", shelf, 100, 1, 4, 2, false, true, "salzstangen", "", nil},
	{"Snacks", "Pizza", freezer, 300, 1, 4, 4, false, true, "pizza", "", nil},
	{"Sonstiges", "Eis", freezer, 100, 1, 8, 7, false, true, "eis", "", nil},
	{"Sonstiges", "Weißwein", smallFridge, 500, 1, 2, 0, false, true, "weisswein", "", nil},
	{"Sonstiges", "Rotwein", smallFridge, 500, 1, 2, 2, false, true, "rotwein", "", nil},
	{"Sonstiges", "Waschmarke", cashBox, 50, 1, 40, 67, true, true, "waschmarke", "", nil},
	{"Pfand", "Pfand Bierflasche", "", -8, 1, 0, 0, false, false, "pfand-bierflasche", "", nil},
	{"Pfand", "Pfand Flasche", "", -15, 1, 0, 0, false, false, "pfand-flasche", "", nil},
	{"Pfand", "Pfand Kiste", "", 150, 1, 0, 0, false, false, "pfand-kiste", "", nil},
}

// lending is what the bar lends, as the OBK's list has it. Nothing costs anything: the deposit goes into
// the box and comes back out.
var lending = []struct {
	name    string
	deposit int64
	code    string
	was     string
}{
	{"Billardqueue 1", 500, "90001", ""},
	{"Dartpfeile 1", 100, "90011", ""},
	{"Dartpfeile 2", 100, "90012", ""},
	{"Dartpfeile 3", 100, "90013", ""},
	{"Airhockey-Set 1", 1000, "90031", "Airhockey-Set"},
	{"Airhockey-Set 2", 1000, "90032", ""},
}

var categories = []struct {
	name, color string
}{
	{"Bier", "#e8a33d"},
	{"Alkoholfrei", "#3fb6a8"},
	{"Snacks", "#d9705a"},
	{"Sonstiges", "#6f86d6"},
	{"Pfand", "#9a9284"},
}

// seedBasics fills a new database with the categories. The range and the lending list follow in
// replaceRange and syncLending, which bring older tills to the same state.
func seedBasics(ctx context.Context, tx *sql.Tx) error {
	for i, c := range categories {
		if _, err := tx.ExecContext(ctx, "INSERT INTO categories(name, color, sort) VALUES (?, ?, ?)", c.name, c.color, i); err != nil {
			return err
		}
	}
	return nil
}

// updateBasics brings the categories of a till from before version 4 up to what seedBasics gives a new one.
func updateBasics(ctx context.Context, tx *sql.Tx) error {
	for _, q := range []string{
		"INSERT INTO categories(name, color, sort) SELECT 'Sonstiges', '#6f86d6', 3 WHERE NOT EXISTS (SELECT 1 FROM categories WHERE name = 'Sonstiges')",
		"UPDATE categories SET sort = 4 WHERE name = 'Pfand' AND sort = 3",
	} {
		if _, err := tx.ExecContext(ctx, q); err != nil {
			return err
		}
	}
	return nil
}

// replaceRange makes the till's range the catalog (version 6, asked for by the OBK). Everything else is
// hidden here and deleted in version 9 by dropOldRange. A product found
// under its name or an earlier one is reused, so its history stays attached, and takes the catalog's
// price, place, count and picture. A count that changes is written down as a count, so the stock
// history still adds up.
func replaceRange(ctx context.Context, tx *sql.Tx, stamp int64) error {
	for i, c := range categories {
		if _, err := tx.ExecContext(ctx, `INSERT INTO categories(name, color, sort) SELECT ?, ?, ?
			WHERE NOT EXISTS (SELECT 1 FROM categories WHERE name = ?)`, c.name, c.color, i, c.name); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, "UPDATE categories SET active = 1 WHERE name = ?", c.name); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, "UPDATE products SET active = 0"); err != nil {
		return err
	}
	for i, p := range catalog {
		id, err := findSeedProduct(ctx, tx, p)
		if err != nil {
			return err
		}
		var category int64
		if err := tx.QueryRowContext(ctx, "SELECT id FROM categories WHERE name = ? ORDER BY id LIMIT 1", p.cat).Scan(&category); err != nil {
			return err
		}
		// The code belongs to this product now. Whatever held it before is hidden.
		if p.code != "" {
			if _, err := tx.ExecContext(ctx, "UPDATE products SET barcode = NULL WHERE barcode = ? AND id <> ?", p.code, id); err != nil {
				return err
			}
		}
		target := p.target
		if !p.track {
			target = 0
		}
		if id == 0 {
			res, err := tx.ExecContext(ctx, `INSERT INTO products(category_id, name, price_cents, barcode, crate_size,
				track_stock, bar_stock, storage_stock, bar_target, min_stock, location, quick_count, sort)
				VALUES (?, ?, ?, ?, ?, ?, ?, 0, ?, ?, ?, ?, ?)`,
				category, p.name, p.price, nullable(p.code), p.crate, boolInt(p.track), p.stock, target, target, p.loc, boolInt(p.quick), i)
			if err != nil {
				return err
			}
			if id, err = res.LastInsertId(); err != nil {
				return err
			}
		} else {
			var bar, storage int64
			if err := tx.QueryRowContext(ctx, "SELECT bar_stock, storage_stock FROM products WHERE id = ?", id).Scan(&bar, &storage); err != nil {
				return err
			}
			for _, m := range []struct {
				place string
				delta int64
			}{{PlaceBar, p.stock - bar}, {PlaceStorage, -storage}} {
				if m.delta == 0 {
					continue
				}
				if _, err := tx.ExecContext(ctx, "INSERT INTO stock_moves(product_id, place, delta, kind, created_at) VALUES (?, ?, ?, 'count', ?)",
					id, m.place, m.delta, stamp); err != nil {
					return err
				}
			}
			if _, err := tx.ExecContext(ctx, `UPDATE products SET category_id = ?, name = ?, price_cents = ?,
				barcode = COALESCE(?, barcode), crate_size = ?, track_stock = ?, bar_stock = ?, storage_stock = 0,
				bar_target = ?, min_stock = ?, location = ?, quick_count = ?, sort = ?, active = 1 WHERE id = ?`,
				category, p.name, p.price, nullable(p.code), p.crate, boolInt(p.track), p.stock, target, target, p.loc,
				boolInt(p.quick), i, id); err != nil {
				return err
			}
		}
		if p.image == "" {
			if _, err := tx.ExecContext(ctx, "DELETE FROM product_images WHERE product_id = ?", id); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, "UPDATE products SET image_version = 0 WHERE id = ?", id); err != nil {
				return err
			}
			continue
		}
		data, err := catalogImages.ReadFile("catalog/" + p.image + ".webp")
		if err != nil {
			return err
		}
		// A new version even for a reused product, so no browser keeps showing the old picture.
		if err := writeImage(ctx, tx, id, "image/webp", data, stamp+int64(i)); err != nil {
			return err
		}
	}
	return nil
}

// findSeedProduct returns the product with the catalog name or one of its earlier names, or 0. The
// catalog name wins over an earlier one, and the newest product over older ones.
func findSeedProduct(ctx context.Context, tx *sql.Tx, p seedProduct) (int64, error) {
	for _, name := range append([]string{p.name}, p.was...) {
		var id int64
		err := tx.QueryRowContext(ctx, "SELECT id FROM products WHERE name = ? ORDER BY id DESC LIMIT 1", name).Scan(&id)
		if err == nil {
			return id, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return 0, err
		}
	}
	return 0, nil
}

// syncLending makes the lending list the bar's (version 6). An item found by name keeps its history and
// takes the new deposit. Anything else is retired, and an item that is out stays visible until it is back.
func syncLending(ctx context.Context, tx *sql.Tx) error {
	keep := make([]any, 0, len(lending))
	for _, l := range lending {
		if l.was != "" {
			if _, err := tx.ExecContext(ctx, `UPDATE rental_items SET name = ? WHERE name = ?
				AND NOT EXISTS (SELECT 1 FROM rental_items WHERE name = ?)`, l.name, l.was, l.name); err != nil {
				return err
			}
		}
		res, err := tx.ExecContext(ctx, "UPDATE rental_items SET deposit_cents = ?, fee_cents = 0, fee_per = 'flat', active = 1 WHERE name = ?",
			l.deposit, l.name)
		if err != nil {
			return err
		}
		n, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if n == 0 {
			var used int
			if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM rental_items WHERE code = ?", l.code).Scan(&used); err != nil {
				return err
			}
			code := any(l.code)
			if used > 0 {
				code = nil
			}
			if _, err := tx.ExecContext(ctx, "INSERT INTO rental_items(name, code, deposit_cents) VALUES (?, ?, ?)", l.name, code, l.deposit); err != nil {
				return err
			}
		}
		keep = append(keep, l.name)
	}
	_, err := tx.ExecContext(ctx, "UPDATE rental_items SET active = 0 WHERE name NOT IN (?"+strings.Repeat(", ?", len(keep)-1)+")", keep...)
	return err
}

// dropOldRange (version 9) deletes what replaceRange only hid, as the OBK asked: the products that are not
// in the bar any more, and the lending items that are retired and were never lent. Sold lines keep their
// name, category and price and only stop pointing to a product, so the old sales and reports still add
// up. The stock moves, counts, deliveries and refills of those products go with them.
func dropOldRange(ctx context.Context, tx *sql.Tx) error {
	const old = "(SELECT id FROM products WHERE active = 0)"
	for _, q := range []string{
		"UPDATE sale_lines SET product_id = NULL WHERE product_id IN " + old,
		"DELETE FROM stock_moves WHERE product_id IN " + old,
		"DELETE FROM stocktake_lines WHERE product_id IN " + old,
		"DELETE FROM restock_lines WHERE product_id IN " + old,
		"DELETE FROM refills WHERE product_id IN " + old,
		"DELETE FROM product_images WHERE product_id IN " + old,
		"DELETE FROM products WHERE active = 0",
		// A count or delivery of old products alone is empty now. Neither can be saved empty.
		"DELETE FROM stocktakes WHERE NOT EXISTS (SELECT 1 FROM stocktake_lines l WHERE l.stocktake_id = stocktakes.id)",
		"DELETE FROM restocks WHERE NOT EXISTS (SELECT 1 FROM restock_lines l WHERE l.restock_id = restocks.id)",
		"DELETE FROM rental_items WHERE active = 0 AND NOT EXISTS (SELECT 1 FROM rentals r WHERE r.item_id = rental_items.id)",
	} {
		if _, err := tx.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("%s: %w", q, err)
		}
	}
	return nil
}

// SeedDemoHistory adds five weeks of invented shifts, so the demo mode has charts
// to show. It runs only in demo mode, which uses its own database.
func (s *Store) SeedDemoHistory(ctx context.Context) error {
	rng := rand.New(rand.NewSource(7))
	names := []struct{ room, name string }{
		{"214", "Jonas"}, {"108", "Mehmet"}, {"322", "Lukas"}, {"117", "Tim"}, {"305", "David"}, {"226", "Felix"},
	}
	products, err := s.Products(ctx)
	if err != nil {
		return err
	}
	sellable := []Product{}
	for _, p := range products {
		if p.TrackStock && p.BarStock > 0 {
			sellable = append(sellable, p)
		}
	}
	realNow := s.now
	defer func() { s.now = realNow }()
	base := realNow()
	var box int64 = 10000

	for daysAgo := 35; daysAgo >= 1; daysAgo-- {
		day := base.AddDate(0, 0, -daysAgo)
		if wd := day.Weekday(); wd != time.Tuesday && wd != time.Thursday && wd != time.Friday && wd != time.Saturday {
			continue
		}
		start := time.Date(day.Year(), day.Month(), day.Day(), 20, rng.Intn(30), 0, 0, time.Local)
		s.now = func() time.Time { return start }
		worker := names[rng.Intn(len(names))]
		if _, err := s.OpenShift(ctx, worker.room, worker.name, box); err != nil {
			return err
		}
		sales := 18 + rng.Intn(40)
		for i := 0; i < sales; i++ {
			at := start.Add(time.Duration(i) * 5 * time.Minute)
			s.now = func() time.Time { return at }
			lines := []LineInput{{sellable[rng.Intn(len(sellable))].ID, int64(1 + rng.Intn(3))}}
			if rng.Intn(3) == 0 {
				lines = append(lines, LineInput{sellable[rng.Intn(len(sellable))].ID, 1})
			}
			payment := PayCash
			if rng.Intn(5) == 0 {
				payment = PayOther
			}
			if _, err := s.CreateSale(ctx, dedupe(lines), payment); err != nil {
				return err
			}
		}
		end := start.Add(4 * time.Hour)
		s.now = func() time.Time { return end }
		sh, err := s.CurrentShift(ctx)
		if err != nil {
			return err
		}
		totals, err := s.ShiftTotals(ctx, *sh)
		if err != nil {
			return err
		}
		counted := totals.ExpectedCents
		if rng.Intn(4) == 0 {
			counted -= int64(10 * (1 + rng.Intn(30)))
		}
		if _, err := s.CloseShift(ctx, counted, "", ""); err != nil {
			return err
		}
		box = counted
		if box > 25000 {
			if err := s.AddCashMove(ctx, "withdrawal", box-10000, "Zur Bank"); err != nil {
				return err
			}
			box = 10000
		}
	}
	s.now = realNow
	// So the demo shows what a message for the next shift looks like.
	if _, err := s.db.ExecContext(ctx, "UPDATE shifts SET handover = ? WHERE id = (SELECT MAX(id) FROM shifts)",
		"Der Kühlschrank tropft hinten links, ich habe ein Tuch hingelegt."); err != nil {
		return err
	}
	// Five weeks of sales with no deliveries emptied everything. Set a believable evening instead.
	for _, p := range sellable {
		bar := p.BarTarget * int64(20+rng.Intn(81)) / 100
		storage := p.CrateSize * int64(rng.Intn(4))
		if _, err := s.db.ExecContext(ctx, "UPDATE products SET bar_stock = ?, storage_stock = ? WHERE id = ?", bar, storage, p.ID); err != nil {
			return err
		}
	}
	return nil
}

func dedupe(lines []LineInput) []LineInput {
	seen := map[int64]int{}
	out := lines[:0]
	for _, l := range lines {
		if i, ok := seen[l.ProductID]; ok {
			out[i].Qty += l.Qty
			continue
		}
		seen[l.ProductID] = len(out)
		out = append(out, l)
	}
	return out
}
