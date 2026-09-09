package storage

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func compatibilityDB(t *testing.T, path string, reader bool) *sql.DB {
	t.Helper()
	c, err := newConnector(path, reader)
	if err != nil {
		t.Fatal(err)
	}
	db := sql.OpenDB(c)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	return db
}

func TestSQLiteCompatibility(t *testing.T) {
	db := compatibilityDB(t, filepath.Join(t.TempDir(), "runtime.db"), false)
	var version string
	if err := db.QueryRow("SELECT sqlite_version()").Scan(&version); err != nil {
		t.Fatal(err)
	}
	t.Logf("compiler=%s SQLite=%s", runtime.Version(), version)
	if version != "3.53.4" {
		t.Fatalf("SQLite=%s, want 3.53.4", version)
	}
}

func TestConnectorConstruction_NoIO(t *testing.T) {
	for _, reader := range []bool{false, true} {
		path := filepath.Join(t.TempDir(), "absent", "literal ?#%.db")
		if _, err := newConnector(path, reader); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(filepath.Dir(path)); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("construction performed I/O: %v", err)
		}
	}
}

func TestConnectionFactory_Pragmas(t *testing.T) {
	for _, reader := range []bool{false, true} {
		db := compatibilityDB(t, filepath.Join(t.TempDir(), "settings.db"), reader)
		for replacement := 0; replacement < 2; replacement++ {
			conn, err := db.Conn(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			queryOnly := 0
			if reader {
				queryOnly = 1
			}
			for pragma, want := range map[string]int{"foreign_keys": 1, "busy_timeout": 5000, "synchronous": 1, "query_only": queryOnly} {
				var got int
				if err := conn.QueryRowContext(context.Background(), "PRAGMA "+pragma).Scan(&got); err != nil {
					t.Fatal(err)
				}
				if got != want {
					t.Errorf("reader=%v replacement=%d %s=%d, want %d", reader, replacement, pragma, got, want)
				}
			}
			// No idle capacity forces the next acquisition to open a new physical connection.
			db.SetMaxIdleConns(0)
			if err := conn.Close(); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestReaderConnection_QueryOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "reader.db")
	writer := compatibilityDB(t, path, false)
	if _, err := writer.Exec("CREATE TABLE probe (value INTEGER)"); err != nil {
		t.Fatal(err)
	}
	reader := compatibilityDB(t, path, true)
	tx, err := reader.BeginTx(context.Background(), &sql.TxOptions{ReadOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	var count int
	if err := tx.QueryRow("SELECT count(*) FROM probe").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec("INSERT INTO probe VALUES (1)"); err == nil {
		t.Fatal("reader allowed a write")
	}
}

func TestAcquire_CanceledContext(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cancel.db")
	db := compatibilityDB(t, path, false)
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if conn, err := db.Conn(canceled); !errors.Is(err, context.Canceled) {
		if conn != nil {
			conn.Close()
		}
		t.Fatalf("canceled acquisition: %v", err)
	}
	conn, err := db.Conn(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	waitCtx, waitCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer waitCancel()
	if second, err := db.Conn(waitCtx); !errors.Is(err, context.DeadlineExceeded) {
		if second != nil {
			second.Close()
		}
		t.Errorf("occupied pool acquisition: %v", err)
	}
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
	if err := db.PingContext(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestWriterConnection_BeginsImmediate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lock.db")
	a := compatibilityDB(t, path, false)
	b := compatibilityDB(t, path, false)
	if _, err := a.Exec("PRAGMA journal_mode=WAL"); err != nil {
		t.Fatal(err)
	}
	// Open B before taking the lock so this probes BeginTx, not file opening.
	if err := b.Ping(); err != nil {
		t.Fatal(err)
	}
	held, err := a.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer held.Rollback()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	tx, err := b.BeginTx(ctx, nil)
	elapsed := time.Since(start)
	if tx != nil {
		tx.Rollback()
		t.Error("second writer entered while the first held the lock")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("lock cancellation: %v", err)
	}
	t.Logf("100ms lock deadline returned after %s: %v", elapsed, err)
	if elapsed >= time.Second {
		t.Errorf("lock cancellation took %s; want under 1s, before 5000ms busy timeout", elapsed)
	}
	if err := held.Rollback(); err != nil {
		t.Fatal(err)
	}
	recovered, err := b.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := recovered.Rollback(); err != nil {
		t.Fatal(err)
	}
}

func TestWriterConnection_CancelAndRestore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cancel-lock.db")
	a := compatibilityDB(t, path, false)
	b := compatibilityDB(t, path, false)
	if err := b.Ping(); err != nil {
		t.Fatal(err)
	}
	held, err := a.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer held.Rollback()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	timer := time.AfterFunc(100*time.Millisecond, cancel)
	defer timer.Stop()
	start := time.Now()
	tx, err := b.BeginTx(ctx, nil)
	if tx != nil {
		tx.Rollback()
	}
	if !errors.Is(err, context.Canceled) || time.Since(start) >= time.Second {
		t.Fatalf("cancel without deadline: %v after %s", err, time.Since(start))
	}
	var timeout int
	if err := b.QueryRow("PRAGMA busy_timeout").Scan(&timeout); err != nil {
		t.Fatal(err)
	}
	if timeout != 5000 {
		t.Fatalf("timeout after cancellation=%d", timeout)
	}
	if err := held.Rollback(); err != nil {
		t.Fatal(err)
	}
	live, stop := context.WithCancel(context.Background())
	defer stop()
	recovered, err := b.BeginTx(live, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer recovered.Rollback()
	if err := recovered.QueryRow("PRAGMA busy_timeout").Scan(&timeout); err != nil {
		t.Fatal(err)
	}
	if timeout != 5000 {
		t.Fatalf("timeout inside callback=%d", timeout)
	}
}
