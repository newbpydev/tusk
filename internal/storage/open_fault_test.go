package storage

import (
	"context"
	"database/sql/driver"
	"errors"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
)

type openFaultConnector struct {
	driver.Connector
	stage          string
	opened, closed *atomic.Int32
}

func (f openFaultConnector) Connect(ctx context.Context) (driver.Conn, error) {
	if f.stage == "connect" {
		return nil, errors.New("injected connect failure")
	}
	c, err := f.Connector.Connect(ctx)
	if err != nil {
		return nil, err
	}
	f.opened.Add(1)
	return &openFaultConn{sqliteConn: c.(sqliteConn), stage: f.stage, closed: f.closed}, nil
}

type openFaultConn struct {
	sqliteConn
	stage  string
	closed *atomic.Int32
}

func (c *openFaultConn) Close() error {
	c.closed.Add(1)
	err := c.sqliteConn.Close()
	if c.stage == "close" {
		return errors.New("injected close failure")
	}
	return err
}
func (c *openFaultConn) BeginTx(ctx context.Context, o driver.TxOptions) (driver.Tx, error) {
	if c.stage == "begin" {
		return nil, errors.New("injected begin failure")
	}
	return c.sqliteConn.BeginTx(ctx, o)
}
func (c *openFaultConn) Ping(ctx context.Context) error {
	if c.stage == "ping" {
		return errors.New("injected ping failure")
	}
	return c.sqliteConn.Ping(ctx)
}
func (c *openFaultConn) QueryContext(ctx context.Context, q string, args []driver.NamedValue) (driver.Rows, error) {
	if c.stage == "catalog" {
		return nil, errors.New("injected catalog failure")
	}
	if q == "PRAGMA journal_mode=WAL" {
		if c.stage == "wal" {
			return nil, errors.New("injected WAL failure")
		}
		if c.stage == "journal" {
			return c.sqliteConn.QueryContext(ctx, "SELECT 'delete'", nil)
		}
	}
	return c.sqliteConn.QueryContext(ctx, q, args)
}
func (c *openFaultConn) ExecContext(ctx context.Context, q string, args []driver.NamedValue) (driver.Result, error) {
	if c.stage == "migration" && len(q) > 12 && q[:12] == "CREATE TABLE" {
		return nil, errors.New("injected migration failure")
	}
	return c.sqliteConn.ExecContext(ctx, q, args)
}

func TestOpen_FailureStagesReleaseEveryHandle(t *testing.T) {
	for _, tc := range []struct {
		role  connectionRole
		stage string
	}{
		{inspectionConnection, "connect"}, {inspectionConnection, "begin"}, {inspectionConnection, "catalog"}, {inspectionConnection, "close"},
		{writerConnection, "wal"}, {writerConnection, "journal"}, {writerConnection, "migration"}, {readerConnection, "ping"},
	} {
		t.Run(tc.stage, func(t *testing.T) {
			var opened, closed atomic.Int32
			factory := func(p string, role connectionRole) (driver.Connector, error) {
				base, err := connectorFor(p, role)
				if err != nil {
					return nil, err
				}
				stage := ""
				if role == tc.role {
					stage = tc.stage
				}
				return openFaultConnector{base, stage, &opened, &closed}, nil
			}
			path := filepath.Join(t.TempDir(), "fault.db")
			if r, err := openAt(context.Background(), path, factory, prepareFile); err == nil {
				r.Close()
				t.Fatal("injection did not fail")
			}
			if opened.Load() != closed.Load() {
				t.Fatalf("leaked handles: opened %d closed %d", opened.Load(), closed.Load())
			}
			r, err := Open(context.Background(), Options{Path: path})
			if err != nil {
				t.Fatalf("recovery: %v", err)
			}
			r.Close()
		})
	}
}

func TestOpen_CanceledAndInvalidOverrideNeverFallsBack(t *testing.T) {
	base := t.TempDir()
	if err := os.WriteFile(filepath.Join(base, "not-a-directory"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_DATA_HOME", filepath.Join(base, "fallback"))
	t.Setenv("TUSK_DB_PATH", filepath.Join(base, "not-a-directory", "database.db"))
	if _, err := Open(context.Background(), Options{}); err == nil {
		t.Fatal("invalid override accepted")
	}
	if _, err := os.Stat(filepath.Join(base, "fallback")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("fallback created")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Open(ctx, Options{Path: filepath.Join(base, "x.db")}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := openAt(ctx, "unused", connectorFor, prepareFile); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	live, stop := context.WithCancel(context.Background())
	if _, err := openAt(live, "unused", connectorFor, func(string) error { stop(); return nil }); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	r := &Repository{}
	if err := r.admit(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
