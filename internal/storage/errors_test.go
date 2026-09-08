package storage

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"modernc.org/sqlite"
)

func TestStorageErrors_RedactData(t *testing.T) {
	secret := filepath.Join(t.TempDir(), "PRIVATE-NOTES")
	if err := os.WriteFile(secret, nil, 0600); err != nil {
		t.Fatal(err)
	}
	_, err := Open(context.Background(), Options{Path: filepath.Join(secret, "database.db")})
	if err == nil || strings.Contains(err.Error(), "PRIVATE-NOTES") {
		t.Fatalf("private path exposed: %v", err)
	}
	r, path := diskRepository(t)
	r.writer.Close()
	base, err := newConnector(path, false)
	if err != nil {
		t.Fatal(err)
	}
	r.writer = sql.OpenDB(openFaultConnector{base, "close", new(atomic.Int32), new(atomic.Int32)})
	if err := r.writer.Ping(); err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); !errors.Is(err, ports.ErrStorage) {
		t.Fatalf("close category leaked: %v", err)
	}
}

func TestStorageErrors_CategoriesAndNoDriverUnwrap(t *testing.T) {
	for _, err := range []error{context.Canceled, context.DeadlineExceeded, core.ErrTaskNotFound, core.ErrDuplicateTaskID, ports.ErrBusy, ports.ErrCorrupt, ports.ErrReadOnly, ports.ErrStorage} {
		if !errors.Is(storageCause(err), err) {
			t.Fatal(err)
		}
	}
	if storageCause(nil) != nil {
		t.Fatal("nil mapped to failure")
	}
	path := filepath.Join(t.TempDir(), "errors.db")
	a := compatibilityDB(t, path, false)
	b := compatibilityDB(t, path, false)
	if _, err := a.Exec("CREATE TABLE probe(value TEXT)"); err != nil {
		t.Fatal(err)
	}
	if _, err := b.Exec("PRAGMA busy_timeout=1"); err != nil {
		t.Fatal(err)
	}
	held, err := a.Begin()
	if err != nil {
		t.Fatal(err)
	}
	_, busy := b.Exec("INSERT INTO probe VALUES ('PRIVATE-NOTES')")
	held.Rollback()
	if !errors.Is(storageCause(busy), ports.ErrBusy) {
		t.Fatalf("busy=%v", busy)
	}
	reader := compatibilityDB(t, path, true)
	_, readonly := reader.Exec("INSERT INTO probe VALUES('PRIVATE-NOTES')")
	if !errors.Is(storageCause(readonly), ports.ErrReadOnly) {
		t.Fatal(readonly)
	}
	for _, raw := range []error{busy, readonly, errors.New("PRIVATE-NOTES SQL failed")} {
		mapped := storageCause(raw)
		var concrete *sqlite.Error
		if errors.As(mapped, &concrete) || strings.Contains(mapped.Error(), "PRIVATE-NOTES") {
			t.Fatalf("driver data escaped: %v", mapped)
		}
	}
}
