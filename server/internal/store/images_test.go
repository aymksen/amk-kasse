package store

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"
)

func TestUpgradeToVersionEightSwapsOnlyTheOldPicture(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	old := []byte("the picture the catalog used to have")
	custom := []byte("a photo someone uploaded in OBK")
	sum := sha256.Sum256(old)
	keep := pictureFixes
	t.Cleanup(func() { pictureFixes = keep })
	pictureFixes = []pictureFix{
		{"Bolten Alt", "bolten-alt", hex.EncodeToString(sum[:])},
		{"Veltins", "veltins", hex.EncodeToString(sum[:])},
	}

	at(s, time.Date(2026, 1, 1, 12, 0, 0, 0, time.Local))
	bolten, veltins := product(t, s, "Bolten Alt").ID, product(t, s, "Veltins").ID
	if _, err := s.SetProductImage(ctx, bolten, "image/webp", old); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetProductImage(ctx, veltins, "image/webp", custom); err != nil {
		t.Fatal(err)
	}
	boltenBefore, veltinsBefore := product(t, s, "Bolten Alt").ImageVersion, product(t, s, "Veltins").ImageVersion
	asVersionEight(t, s)
	if _, err := s.db.ExecContext(ctx, "PRAGMA user_version = 7"); err != nil {
		t.Fatal(err)
	}
	s.Close()

	s, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	var version int
	if err := s.db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil || version != schemaVersion {
		t.Fatalf("version %d, %v", version, err)
	}
	want, _ := catalogImages.ReadFile("catalog/bolten-alt.webp")
	if _, data, err := s.ProductImage(ctx, bolten); err != nil || !bytes.Equal(data, want) {
		t.Errorf("Bolten Alt still has the old picture (%d bytes, %v)", len(data), err)
	}
	if v := product(t, s, "Bolten Alt").ImageVersion; v == boltenBefore {
		t.Error("the new picture needs a new version, or browsers keep the old one")
	}
	if _, data, err := s.ProductImage(ctx, veltins); err != nil || !bytes.Equal(data, custom) {
		t.Errorf("an uploaded picture was replaced (%d bytes, %v)", len(data), err)
	}
	if v := product(t, s, "Veltins").ImageVersion; v != veltinsBefore {
		t.Errorf("the uploaded picture's version changed: %d → %d", veltinsBefore, v)
	}
}

func TestPictureFixesPointAtTheCatalog(t *testing.T) {
	for _, f := range pictureFixes {
		data, err := catalogImages.ReadFile("catalog/" + f.image + ".webp")
		if err != nil {
			t.Errorf("%s: %v", f.name, err)
			continue
		}
		if sum := sha256.Sum256(data); hex.EncodeToString(sum[:]) == f.old {
			t.Errorf("%s: the catalog still has the old picture", f.name)
		}
		found := false
		for _, p := range catalog {
			found = found || p.name == f.name && p.image == f.image
		}
		if !found {
			t.Errorf("%s with %s is not in the catalog", f.name, f.image)
		}
	}
}
