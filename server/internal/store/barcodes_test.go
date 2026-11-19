package store

import (
	"errors"
	"slices"
	"testing"
)

// asVersionEight turns a new database back into what version 8 left: one bottle code per product in
// products.barcode, and counts without the start-of-shift mark. The upgrade tests start from there.
func asVersionEight(t *testing.T, s *Store) {
	t.Helper()
	for _, q := range []string{
		"ALTER TABLE products ADD COLUMN barcode TEXT",
		"UPDATE products SET barcode = (SELECT code FROM product_barcodes b WHERE b.product_id = products.id ORDER BY b.rowid LIMIT 1)",
		"CREATE UNIQUE INDEX products_barcode ON products(barcode) WHERE barcode IS NOT NULL",
		"DROP TABLE product_barcodes",
		"ALTER TABLE stocktakes DROP COLUMN opening",
	} {
		if _, err := s.db.ExecContext(ctx, q); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
}

func tillProduct(t *testing.T, s *Store, id int64) TillProduct {
	t.Helper()
	all, err := s.TillProducts(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range all {
		if p.ID == id {
			return p
		}
	}
	t.Fatalf("product %d is not at the till", id)
	return TillProduct{}
}

func TestDuplicateBarcodeIsRejected(t *testing.T) {
	s := newStore(t)
	beer := product(t, s, "Früh Kölsch")
	cola := product(t, s, "Coca-Cola")
	beer.Barcodes = []string{"4001234567890"}
	if _, err := s.SaveProduct(ctx, beer); err != nil {
		t.Fatal(err)
	}
	cola.Barcodes = append(cola.Barcodes, "4001234567890")
	if _, err := s.SaveProduct(ctx, cola); !errors.Is(err, ErrDuplicateCode) {
		t.Errorf("err = %v", err)
	}
	if got := product(t, s, "Coca-Cola").Barcodes; !slices.Equal(got, []string{seedNamed(t, "Coca-Cola").code}) {
		t.Errorf("a refused save changed the codes: %v", got)
	}
}

// One key such as Chips stands for several kinds, each with its own code on the bag.
func TestAProductTakesSeveralBarcodes(t *testing.T) {
	s := newStore(t)
	chips := product(t, s, "Chips")
	chips.Barcodes = []string{" 4003586100351", "4062139007677", "", "5053990138722", "4062139007677"}
	saved, err := s.SaveProduct(ctx, chips)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"4003586100351", "4062139007677", "5053990138722"}
	if !slices.Equal(saved.Barcodes, want) {
		t.Errorf("saved %v, want %v", saved.Barcodes, want)
	}
	if got := product(t, s, "Chips").Barcodes; !slices.Equal(got, want) {
		t.Errorf("stored %v, want %v", got, want)
	}
	if got := tillProduct(t, s, chips.ID).Barcodes; !slices.Equal(got, want) {
		t.Errorf("the till gets %v, want %v", got, want)
	}

	// A code is a bottle code or a crate code, never both.
	beer := product(t, s, "Veltins")
	beer.CrateBarcode = "5053990138722"
	if _, err := s.SaveProduct(ctx, beer); !errors.Is(err, ErrDuplicateCode) {
		t.Errorf("crate code that is chips: %v", err)
	}
	beer.CrateBarcode = "4100000000017"
	if _, err := s.SaveProduct(ctx, beer); err != nil {
		t.Fatal(err)
	}
	chips.Barcodes = []string{"4100000000017"}
	if _, err := s.SaveProduct(ctx, chips); !errors.Is(err, ErrDuplicateCode) {
		t.Errorf("bottle code that is a crate: %v", err)
	}

	// A code can move from the bottle to the crate in one save, and codes can go again.
	beer = product(t, s, "Veltins")
	beer.Barcodes, beer.CrateBarcode = []string{"4100000000024"}, ""
	if _, err := s.SaveProduct(ctx, beer); err != nil {
		t.Fatal(err)
	}
	beer = product(t, s, "Veltins")
	beer.Barcodes, beer.CrateBarcode = nil, "4100000000024"
	if _, err := s.SaveProduct(ctx, beer); err != nil {
		t.Errorf("moving the code to the crate: %v", err)
	}
	chips = product(t, s, "Chips")
	chips.Barcodes = chips.Barcodes[:1]
	if _, err := s.SaveProduct(ctx, chips); err != nil {
		t.Fatal(err)
	}
	if got := product(t, s, "Chips").Barcodes; !slices.Equal(got, want[:1]) {
		t.Errorf("after removing two: %v", got)
	}

	// A new product gets its codes too.
	ice := Product{Name: "Eis am Stiel", CategoryID: chips.CategoryID, PriceCents: 150, CrateSize: 1, Barcodes: []string{"4100000000031"}, Active: true}
	if ice, err = s.SaveProduct(ctx, ice); err != nil {
		t.Fatal(err)
	}
	if got := product(t, s, "Eis am Stiel"); got.ID != ice.ID || !slices.Equal(got.Barcodes, []string{"4100000000031"}) {
		t.Errorf("new product = %+v", got)
	}
}

// Version 9 moves each bottle code into the new table and deletes the old range, including its codes.
func TestUpgradeToVersionNineKeepsTheCodes(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	asVersionEight(t, s)
	for _, q := range []string{
		"UPDATE products SET barcode = '4003586100351' WHERE name = 'Chips'",
		"INSERT INTO products(category_id, name, price_cents, barcode, active) VALUES (3, 'Pringles Original', 250, '5053990138722', 0)",
		"PRAGMA user_version = 8",
	} {
		if _, err := s.db.ExecContext(ctx, q); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	s.Close()

	s, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if got := product(t, s, "Chips").Barcodes; !slices.Equal(got, []string{"4003586100351"}) {
		t.Errorf("chips = %v", got)
	}
	if got := product(t, s, "Hansa Pils").Barcodes; !slices.Equal(got, []string{seedNamed(t, "Hansa Pils").code}) {
		t.Errorf("hansa = %v", got)
	}
	var left int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM product_barcodes WHERE code = '5053990138722'").Scan(&left); err != nil || left != 0 {
		t.Errorf("the old product's code is still there: %d, %v", left, err)
	}
	chips := product(t, s, "Chips")
	chips.Barcodes = append(chips.Barcodes, "5053990138722")
	if _, err := s.SaveProduct(ctx, chips); err != nil {
		t.Errorf("the old product's code is free for chips: %v", err)
	}
}
