package storage

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"testing"

	assets "github.com/newbpydev/tusk/db"
	"modernc.org/sqlite"
)

func memoryRepository(t *testing.T) (*Repository, func(bool) *sql.DB) {
	t.Helper()
	name := "tusk-" + rand.Text()
	pool := func(reader bool) *sql.DB {
		params := "&_txlock=immediate"
		if reader {
			params = "&_txlock=deferred&_pragma=query_only(1)"
		}
		base, err := sqlite.NewConnector(fmt.Sprintf("file:%s?mode=memory&cache=shared&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)%s", name, params))
		if err != nil {
			t.Fatal(err)
		}
		db := sql.OpenDB(&connector{Connector: base})
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(1)
		if reader {
			db.SetMaxOpenConns(4)
			db.SetMaxIdleConns(4)
		}
		return db
	}
	r := &Repository{writer: pool(false), reader: pool(true)}
	t.Cleanup(func() {
		if err := r.Close(); err != nil {
			t.Error(err)
		}
	})
	inv, err := loadMigrations(assets.Migrations())
	if err != nil {
		t.Fatal(err)
	}
	if err := migrate(context.Background(), r.writer, inv); err != nil {
		t.Fatal(err)
	}
	return r, pool
}

func TestOpen_MemoryLifetimeAndIsolation(t *testing.T) {
	a, pool := memoryRepository(t)
	b, _ := memoryRepository(t)
	if _, err := a.writer.Exec("INSERT INTO tasks(id,title,description,status,priority,progress,tags,created_at,updated_at) VALUES('a','task','','todo',1,0,'[]','x','x')"); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := b.reader.QueryRow("SELECT count(*) FROM tasks").Scan(&count); err != nil || count != 0 {
		t.Fatal("memory fixtures shared data")
	}
	if err := a.reader.Close(); err != nil {
		t.Fatal(err)
	}
	a.reader = pool(true)
	if err := a.reader.QueryRow("SELECT count(*) FROM tasks").Scan(&count); err != nil || count != 1 {
		t.Fatalf("anchor lost data: %d %v", count, err)
	}
	var journal string
	if err := a.writer.QueryRow("PRAGMA journal_mode").Scan(&journal); err != nil || journal != "memory" {
		t.Fatalf("journal=%s %v", journal, err)
	}
	if err := a.Close(); err != nil {
		t.Fatal(err)
	}
	reopened := pool(false)
	defer reopened.Close()
	if err := reopened.QueryRow("SELECT count(*) FROM tasks").Scan(&count); err == nil {
		t.Fatal("memory survived all handles closing")
	}
}
