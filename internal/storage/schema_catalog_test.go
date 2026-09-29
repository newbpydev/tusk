package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"testing"

	assets "github.com/newbpydev/tusk/db"
)

// This is both the generated-asset check and its explicit Makefile updater.
// SQLite remains the catalog compiler; no handwritten DDL parser is involved.
func TestEmbeddedSchemaCatalog(t *testing.T) {
	inv, err := loadMigrations(assets.Migrations())
	if err != nil {
		t.Fatal(err)
	}
	entries := make([]schemaCatalogEntry, len(inv))
	var hashes []string
	for i, m := range inv {
		hashes = append(hashes, fmt.Sprintf("%x", sha256.Sum256([]byte(m.sql))))
		objects, err := evaluateSchema(context.Background(), inv[:i+1])
		if err != nil {
			t.Fatal(err)
		}
		entries[i] = schemaCatalogEntry{append([]string(nil), hashes...), objects}
	}
	want, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	want = append(want, '\n')
	path := filepath.Join("..", "..", "db", "schema_catalog.json")
	if os.Getenv("TUSK_UPDATE_SCHEMA_CATALOG") == "1" {
		tmp, err := os.CreateTemp(filepath.Dir(path), ".schema-catalog-*")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmp.Name())
		if _, err := tmp.Write(want); err != nil {
			tmp.Close()
			t.Fatal(err)
		}
		if err := errors.Join(tmp.Chmod(0644), tmp.Close()); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(tmp.Name(), path); err != nil {
			t.Fatal(err)
		}
	}
	got, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(got, want) {
		t.Fatalf("embedded schema catalog differs from SQLite; run make generate: %v", err)
	}
	for i, entry := range entries {
		objects, ok := readSchemaCatalog(want, inv[:i+1])
		if !ok || !maps.Equal(objects, entry.Objects) {
			t.Fatalf("catalog prefix %d mismatch", i+1)
		}
	}
}

func TestSchemaCatalog_RejectsMismatchAndFallsBack(t *testing.T) {
	inv, err := loadMigrations(assets.Migrations())
	if err != nil {
		t.Fatal(err)
	}
	data, err := fs.ReadFile(assets.Migrations(), "schema_catalog.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range [][]byte{nil, []byte("invalid"), []byte("[]"), []byte("[{}]"), []byte(`[{"sql_sha256":["bad"],"objects":{"table:x":"bad"}}]`)} {
		if _, ok := readSchemaCatalog(bad, inv); ok {
			t.Fatalf("accepted bad catalog: %s", bad)
		}
	}
	if _, ok := readSchemaCatalog(data, nil); ok {
		t.Fatal("accepted empty prefix")
	}
	// Matching a ledger checksum alone is insufficient: match actual SQL bytes.
	inv[0].sql += "\nCREATE TABLE extra (id INTEGER);"
	if _, ok := readSchemaCatalog(data, inv); ok {
		t.Fatal("reused catalog for changed SQL with stale checksum")
	}
	objects, err := expectedSchema(context.Background(), inv)
	if err != nil || objects["table:extra"] != "CREATE TABLE extra (id INTEGER)" {
		t.Fatalf("dynamic migration fallback: %v %v", objects, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := expectedSchema(ctx, inv); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
