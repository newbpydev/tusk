package storage

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

type migrationFaultConnector struct {
	driver.Connector
	stage string
}

func (f migrationFaultConnector) Connect(ctx context.Context) (driver.Conn, error) {
	c, err := f.Connector.Connect(ctx)
	if err != nil {
		return nil, err
	}
	return &migrationFaultConn{sqliteConn: c.(sqliteConn), stage: f.stage}, nil
}

type migrationFaultConn struct {
	sqliteConn
	stage string
}

func (f *migrationFaultConn) ExecContext(ctx context.Context, q string, args []driver.NamedValue) (driver.Result, error) {
	if f.stage == "cancel-ddl" && strings.Contains(q, "CREATE TABLE") || (f.stage == "cancel-ledger" || f.stage == "cancel-rollback") && strings.HasPrefix(q, "INSERT INTO schema_migrations") {
		return nil, context.Canceled
	}
	if f.stage == "ledger" && strings.HasPrefix(q, "INSERT INTO schema_migrations") || f.stage == "app" && strings.HasPrefix(q, "PRAGMA application_id=") {
		return nil, errors.New("injected migration write failure")
	}
	return f.sqliteConn.ExecContext(ctx, q, args)
}
func (f *migrationFaultConn) BeginTx(ctx context.Context, o driver.TxOptions) (driver.Tx, error) {
	if f.stage == "begin" {
		return nil, errors.New("injected begin failure")
	}
	tx, err := f.sqliteConn.BeginTx(ctx, o)
	if err != nil {
		return nil, err
	}
	return &migrationFaultTx{Tx: tx, stage: f.stage}, nil
}

type migrationFaultTx struct {
	driver.Tx
	stage string
}

func (t *migrationFaultTx) Commit() error {
	if t.stage == "cancel-commit" {
		return context.Canceled
	}
	if t.stage == "commit-before" {
		return errors.New("injected before commit")
	}
	err := t.Tx.Commit()
	if err == nil && t.stage == "commit-after" {
		return errors.New("injected lost acknowledgment")
	}
	return err
}
func (t *migrationFaultTx) Rollback() error {
	err := t.Tx.Rollback()
	if err == nil && t.stage == "cancel-rollback" {
		return context.Canceled
	}
	if err == nil && (t.stage == "rollback" || t.stage == "cancel-rollback") {
		return errors.New("injected rollback acknowledgment")
	}
	return err
}

func TestMigrate_InjectedFailures(t *testing.T) {
	for _, stage := range []string{"ledger", "app", "begin", "commit-before", "commit-after", "rollback"} {
		t.Run(stage, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "failure.db")
			base, err := newConnector(path, false)
			if err != nil {
				t.Fatal(err)
			}
			db := sql.OpenDB(migrationFaultConnector{Connector: base, stage: stage})
			db.SetMaxOpenConns(1)
			_, inv := migrationFixture(t)
			if stage == "rollback" {
				inv[0].sql += " INSERT INTO absent VALUES(1);"
			}
			err = migrate(context.Background(), db, inv)
			if err == nil {
				t.Fatal("injected failure reported success")
			}
			if strings.HasPrefix(stage, "commit") || stage == "rollback" {
				if !errors.Is(err, errMigrationOutcome) {
					t.Fatalf("outcome not unknown: %v", err)
				}
			}
			if err := db.Close(); err != nil {
				t.Fatal(err)
			}
			readback := compatibilityDB(t, path, false)
			_, canonical := migrationFixture(t)
			count, err := inspectSchema(context.Background(), readback, canonical)
			want := 0
			if stage == "commit-after" {
				want = 1
			}
			if err != nil || count != want {
				t.Fatalf("readback=%d want=%d err=%v", count, want, err)
			}
			if err := migrate(context.Background(), readback, canonical); err != nil {
				t.Fatalf("recovery: %v", err)
			}
		})
	}
}

type raceConnector struct {
	driver.Connector
	afterIdentity func()
}

func (c raceConnector) Connect(ctx context.Context) (driver.Conn, error) {
	raw, err := c.Connector.Connect(ctx)
	if err != nil {
		return nil, err
	}
	return &raceConnection{sqliteConn: raw.(sqliteConn), afterIdentity: c.afterIdentity}, nil
}

type raceConnection struct {
	sqliteConn
	afterIdentity func()
}

func (c *raceConnection) QueryContext(ctx context.Context, q string, args []driver.NamedValue) (driver.Rows, error) {
	r, err := c.sqliteConn.QueryContext(ctx, q, args)
	if err == nil && q == "PRAGMA application_id" && c.afterIdentity != nil {
		hook := c.afterIdentity
		c.afterIdentity = nil
		return &identityRows{Rows: r, after: hook}, nil
	}
	return r, err
}

type identityRows struct {
	driver.Rows
	after func()
}

func (r *identityRows) Close() error {
	err := r.Rows.Close()
	if r.after != nil {
		hook := r.after
		r.after = nil
		hook()
	}
	return err
}

func TestMigrate_InspectionUsesOneSnapshot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "snapshot-race.db")
	other := compatibilityDB(t, path, false)
	if _, err := other.Exec("PRAGMA journal_mode=WAL"); err != nil {
		t.Fatal(err)
	}
	_, inv := migrationFixture(t)
	base, err := newConnector(path, false)
	if err != nil {
		t.Fatal(err)
	}
	db := sql.OpenDB(raceConnector{Connector: base, afterIdentity: func() {
		if err := migrate(context.Background(), other, inv); err != nil {
			t.Fatal(err)
		}
	}})
	defer db.Close()
	db.SetMaxOpenConns(1)
	if err := migrate(context.Background(), db, inv); err != nil {
		t.Fatalf("concurrent compatible initialization refused: %v", err)
	}
}

// Error categories must survive migration formatting and the public Open boundary.
func TestMigrate_PreservesCancellationCause(t *testing.T) {
	for _, stage := range []string{"cancel-ddl", "cancel-ledger", "cancel-commit", "cancel-rollback"} {
		t.Run(stage, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "cancel.db")
			base, err := newConnector(path, false)
			if err != nil {
				t.Fatal(err)
			}
			db := sql.OpenDB(migrationFaultConnector{Connector: base, stage: stage})
			defer db.Close()
			_, inv := migrationFixture(t)
			err = migrate(context.Background(), db, inv)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("migration lost cancellation: %v", err)
			}
			if !errors.Is(openCause(err), context.Canceled) {
				t.Fatalf("Open lost cancellation: %v", openCause(err))
			}
			if stage == "cancel-commit" || stage == "cancel-rollback" {
				if !errors.Is(err, errMigrationOutcome) {
					t.Fatalf("uncertainty lost: %v", err)
				}
			}
		})
	}
}
