package storage

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	assets "github.com/newbpydev/tusk/db"
	"modernc.org/sqlite"
)

func migrationFixture(t *testing.T) (*sql.DB, []migration) {
	t.Helper()
	db := compatibilityDB(t, filepath.Join(t.TempDir(), "migration.db"), false)
	inv, err := loadMigrations(assets.Migrations())
	if err != nil {
		t.Fatal(err)
	}
	return db, inv
}

func TestMigrate_EmptyDatabase(t *testing.T) {
	db, inv := migrationFixture(t)
	ctx := context.Background()
	if err := migrate(ctx, db, inv); err != nil {
		t.Fatal(err)
	}
	var app, count int
	if err := db.QueryRow("PRAGMA application_id").Scan(&app); err != nil || app != 0x5455534B {
		t.Fatalf("app=%d err=%v", app, err)
	}
	if err := db.QueryRow("SELECT count(*) FROM schema_migrations").Scan(&count); err != nil || count != 1 {
		t.Fatalf("ledger=%d err=%v", count, err)
	}
	var before string
	if err := db.QueryRow("SELECT applied_at FROM schema_migrations").Scan(&before); err != nil {
		t.Fatal(err)
	}
	if err := migrate(ctx, db, inv); err != nil {
		t.Fatal(err)
	}
	var after string
	if err := db.QueryRow("SELECT applied_at FROM schema_migrations").Scan(&after); err != nil || before != after {
		t.Fatalf("migration replay: %s %s %v", before, after, err)
	}
}

func TestMigrationInventory_Invalid(t *testing.T) {
	for _, fixture := range []fstest.MapFS{
		{}, {"migrations/002_gap.sql": {Data: []byte("SELECT 1;")}},
		{"migrations/001_a.sql": {Data: []byte("SELECT 1;")}, "migrations/001_b.sql": {Data: []byte("SELECT 1;")}},
		{"migrations/bad.sql": {Data: []byte("SELECT 1;")}},
		{"migrations/001_empty.sql": {}},
		{"migrations/001_transaction.sql": {Data: []byte("BEGIN; SELECT 1; COMMIT;")}},
		{"migrations/001_attach.sql": {Data: []byte("ATTACH ':memory:' AS x;")}},
		{"migrations/001_vacuum.sql": {Data: []byte("VACUUM;")}},
		{"migrations/001_journal.sql": {Data: []byte("PRAGMA journal_mode=WAL;")}},
	} {
		if _, err := loadMigrations(fixture); err == nil {
			t.Fatalf("accepted invalid inventory: %v", fixture)
		}
	}
}

func TestMigrate_FailurePreservesPreviousVersion(t *testing.T) {
	db, inv := migrationFixture(t)
	ctx := context.Background()
	if err := migrate(ctx, db, inv); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("INSERT INTO tasks(id,title,description,status,priority,progress,tags,created_at,updated_at) VALUES('sentinel','original','','todo',2,0,'[]','2026-09-08T00:00:00.000000000Z','2026-09-08T00:00:00.000000000Z')"); err != nil {
		t.Fatal(err)
	}
	copyFS := fstest.MapFS{}
	for _, m := range inv {
		copyFS["migrations/"+m.name] = &fstest.MapFile{Data: []byte(m.sql)}
	}
	copyFS["migrations/002_fail.sql"] = &fstest.MapFile{Data: []byte("CREATE TABLE should_rollback (value TEXT); UPDATE tasks SET title='changed'; INSERT INTO missing VALUES (1);")}
	pending, err := loadMigrations(copyFS)
	if err != nil {
		t.Fatal(err)
	}
	if err := migrate(ctx, db, pending); err == nil {
		t.Fatal("failed migration accepted")
	}
	if n, err := inspectSchema(ctx, db, inv); err != nil || n != 1 {
		t.Fatalf("previous schema not restored: %d %v", n, err)
	}
	var title string
	if err := db.QueryRow("SELECT title FROM tasks WHERE id='sentinel'").Scan(&title); err != nil || title != "original" {
		t.Fatalf("partial data survived: %s %v", title, err)
	}
}

func TestMigrate_RefusalPreservesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "foreign.db")
	db := compatibilityDB(t, path, false)
	if _, err := db.Exec("PRAGMA journal_mode=WAL; CREATE TABLE private_data(value TEXT); INSERT INTO private_data VALUES('preserve')"); err != nil {
		t.Fatal(err)
	}
	before := make(map[string]string)
	for _, suffix := range []string{"", "-wal"} {
		b, err := os.ReadFile(path + suffix)
		if err != nil {
			t.Fatal(err)
		}
		before[suffix] = string(b)
	}
	_, inv := migrationFixture(t)
	if err := migrate(context.Background(), db, inv); !errors.Is(err, errIncompatibleSchema) {
		t.Fatal(err)
	}
	for suffix, want := range before {
		b, err := os.ReadFile(path + suffix)
		if err != nil || string(b) != want {
			t.Fatalf("refusal changed %s: %v", suffix, err)
		}
	}
}

func TestMigrate_RefusesForeignOrNewer(t *testing.T) {
	for _, setup := range []string{"PRAGMA application_id=42", "CREATE TABLE foreign_data (value TEXT)", "CREATE TABLE sqliteXsecret (value TEXT)"} {
		db, inv := migrationFixture(t)
		if _, err := db.Exec(setup); err != nil {
			t.Fatal(err)
		}
		if err := migrate(context.Background(), db, inv); !errors.Is(err, errIncompatibleSchema) {
			t.Fatalf("foreign result=%v", err)
		}
	}
	db, inv := migrationFixture(t)
	if err := migrate(context.Background(), db, inv); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("UPDATE schema_migrations SET version=2"); err != nil {
		t.Fatal(err)
	}
	if err := migrate(context.Background(), db, inv); !errors.Is(err, errIncompatibleSchema) {
		t.Fatalf("newer result=%v", err)
	}
}

func TestMigrate_BrokenLedger(t *testing.T) {
	for _, mutation := range []string{
		"DROP TABLE schema_migrations", "DROP INDEX tasks_status_idx", "CREATE TABLE surprise (x TEXT)",
		"UPDATE schema_migrations SET filename='001_changed.sql'", "UPDATE schema_migrations SET checksum='" + strings.Repeat("a", 64) + "'",
		"DELETE FROM schema_migrations", "ALTER TABLE tasks ADD COLUMN unexpected TEXT",
	} {
		t.Run(mutation, func(t *testing.T) {
			db, inv := migrationFixture(t)
			if err := migrate(context.Background(), db, inv); err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec(mutation); err != nil {
				t.Fatal(err)
			}
			if err := migrate(context.Background(), db, inv); err == nil {
				t.Fatal("accepted schema drift")
			}
		})
	}
}

func TestMigrationInventory_StableBytes(t *testing.T) {
	db, inv := migrationFixture(t)
	if err := migrate(context.Background(), db, inv); err != nil {
		t.Fatal(err)
	}
	source, err := fs.ReadFile(assets.Migrations(), "migrations/001_initial_schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(source), "\r") {
		t.Fatal("non-LF migration")
	}
	changed, err := loadMigrations(fstest.MapFS{"migrations/001_initial_schema.sql": {Data: append(source, '\n')}})
	if err != nil {
		t.Fatal(err)
	}
	if err := migrate(context.Background(), db, changed); !errors.Is(err, errIncompatibleSchema) {
		t.Fatalf("changed bytes: %v", err)
	}
}

func TestMigrate_CorruptLedgerAndCanceledInspection(t *testing.T) {
	for _, mutation := range []string{
		"PRAGMA ignore_check_constraints=ON; UPDATE schema_migrations SET version=0",
		"UPDATE schema_migrations SET applied_at='invalid'",
		"DROP TABLE schema_migrations; CREATE TABLE schema_migrations(version TEXT,filename TEXT,checksum TEXT,applied_at TEXT); INSERT INTO schema_migrations VALUES('invalid','','','')",
	} {
		db, inv := migrationFixture(t)
		if err := migrate(context.Background(), db, inv); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(mutation); err != nil {
			t.Fatal(err)
		}
		if _, err := inspectSchema(context.Background(), db, inv); !errors.Is(err, errCorruptSchema) {
			t.Fatalf("corrupt ledger accepted: %v", err)
		}
	}
	db, inv := migrationFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := inspectSchema(ctx, db, inv); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := migrate(ctx, db, inv); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := migrate(context.Background(), db, nil); err == nil {
		t.Fatal("empty migration inventory accepted")
	}
	if err := migrate(context.Background(), db, inv); err != nil {
		t.Fatal(err)
	}
	inv[0].sql = "INVALID SQL"
	if _, err := inspectSchema(context.Background(), db, inv); !errors.Is(err, errIncompatibleSchema) {
		t.Fatal(err)
	}
}

type errorFS struct{ fs.FS }

func (errorFS) Glob(string) ([]string, error) { return nil, errors.New("injected inventory failure") }
func TestMigrationInventory_Unreadable(t *testing.T) {
	if _, err := loadMigrations(errorFS{}); err == nil {
		t.Fatal("unreadable glob accepted")
	}
	if _, err := loadMigrations(fstest.MapFS{"migrations/001_directory.sql": {Mode: fs.ModeDir}}); err == nil {
		t.Fatal("unreadable migration accepted")
	}
}

type brokenSchemaReader struct {
	*sql.DB
	mode string
}

func (q brokenSchemaReader) QueryContext(ctx context.Context, sqlText string, args ...any) (*sql.Rows, error) {
	if q.mode == "query" {
		return nil, errors.New("injected query failure")
	}
	return q.DB.QueryContext(ctx, "SELECT NULL,NULL,NULL")
}
func TestInspection_FailureReturnsNoPartialCatalog(t *testing.T) {
	db, inv := migrationFixture(t)
	for _, mode := range []string{"query", "scan"} {
		q := brokenSchemaReader{db, mode}
		if objects, err := schemaObjects(context.Background(), q); err == nil || objects != nil {
			t.Fatal("partial catalog exposed")
		}
		if _, err := inspectSchema(context.Background(), q, inv); err == nil {
			t.Fatal("broken catalog accepted")
		}
	}
}

func TestMigrate_MemoryAndInverseFixture(t *testing.T) {
	c, err := sqlite.NewConnector(":memory:?_pragma=foreign_keys(1)&_txlock=immediate")
	if err != nil {
		t.Fatal(err)
	}
	db := sql.OpenDB(c)
	defer db.Close()
	db.SetMaxOpenConns(1)
	_, inv := migrationFixture(t)
	if err := migrate(context.Background(), db, inv); err != nil {
		t.Fatal(err)
	}
	var journal string
	if err := db.QueryRow("PRAGMA journal_mode").Scan(&journal); err != nil || journal != "memory" {
		t.Fatalf("journal=%s err=%v", journal, err)
	}
	inverse, err := os.ReadFile("testdata/migrations/disposable_inverse.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(string(inverse)); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow("SELECT count(*) FROM sqlite_schema WHERE name NOT GLOB 'sqlite_*'").Scan(&count); err != nil || count != 0 {
		t.Fatalf("inverse fixture=%d err=%v", count, err)
	}
}
