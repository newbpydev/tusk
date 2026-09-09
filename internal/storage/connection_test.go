package storage

import (
	"context"
	"database/sql/driver"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"modernc.org/sqlite"
)

type faultConnection struct {
	sqliteConn
	exec  func(string) error
	begin func() (driver.Tx, error)
}

func (c *faultConnection) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	return nil, c.exec(query)
}
func (c *faultConnection) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return c.begin()
}
func (c *faultConnection) IsValid() bool { return true }

type faultTx struct {
	rolledBack bool
	err        error
}

func (t *faultTx) Commit() error   { return nil }
func (t *faultTx) Rollback() error { t.rolledBack = true; return t.err }

func TestConnection_AcquisitionFailures(t *testing.T) {
	failure := errors.New("injected failure")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for _, stage := range []string{"configure", "begin", "restore", "restore-after-begin-error"} {
		t.Run(stage, func(t *testing.T) {
			tx := &faultTx{err: failure}
			c := &connection{sqliteConn: &faultConnection{
				exec: func(query string) error {
					if stage == "configure" && query == "PRAGMA busy_timeout=25" || (stage == "restore" || stage == "restore-after-begin-error") && query == "PRAGMA busy_timeout=5000" {
						return failure
					}
					return nil
				},
				begin: func() (driver.Tx, error) {
					if stage == "begin" || stage == "restore-after-begin-error" {
						return nil, failure
					}
					return tx, nil
				},
			}}
			got, err := c.BeginTx(ctx, driver.TxOptions{})
			if got != nil || !errors.Is(err, failure) {
				t.Fatalf("got=%v err=%v", got, err)
			}
			if c.IsValid() != (stage == "begin") {
				t.Fatalf("unexpected pool validity for %s", stage)
			}
			if tx.rolledBack != (stage == "restore") {
				t.Fatal("incorrect rollback ownership")
			}
		})
	}
	dead, stop := context.WithCancel(context.Background())
	stop()
	c := &connection{}
	if _, err := c.BeginTx(dead, driver.TxOptions{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestConnection_CancelBetweenAttempts(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := &connection{sqliteConn: &faultConnection{
		exec: func(query string) error {
			if query == "PRAGMA busy_timeout=25" {
				cancel()
			}
			return nil
		},
		begin: func() (driver.Tx, error) { t.Fatal("begin after cancellation"); return nil, nil },
	}}
	if _, err := c.BeginTx(ctx, driver.TxOptions{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

type incompatibleConnection struct {
	driver.Conn
	closed bool
}

func (c *incompatibleConnection) Close() error { c.closed = true; return nil }

type fixtureConnector struct {
	driver.Connector
	raw driver.Conn
	err error
}

func (c fixtureConnector) Connect(context.Context) (driver.Conn, error) { return c.raw, c.err }

func TestConnector_RejectsIncompatibleAndFailedConnections(t *testing.T) {
	raw := &incompatibleConnection{}
	c := &connector{Connector: fixtureConnector{raw: raw}}
	if _, err := c.Connect(context.Background()); err == nil || !raw.closed {
		t.Fatal("incompatible connection not closed")
	}
	failure := errors.New("connect failed")
	c.Connector = fixtureConnector{err: failure}
	if _, err := c.Connect(context.Background()); !errors.Is(err, failure) {
		t.Fatal(err)
	}
}

func TestWriterConnection_WaitBudget(t *testing.T) {
	path := filepath.Join(t.TempDir(), "budget.db")
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	start := time.Now()
	if tx, err := b.BeginTx(ctx, nil); err == nil {
		tx.Rollback()
		t.Fatal("writer entered locked database")
	}
	if elapsed := time.Since(start); elapsed < 5*time.Second || elapsed > 7*time.Second {
		t.Fatalf("busy budget=%s", elapsed)
	}
	var timeout int
	if err := b.QueryRow("PRAGMA busy_timeout").Scan(&timeout); err != nil || timeout != 5000 {
		t.Fatalf("restoration: %d %v", timeout, err)
	}
}

func TestReaderConnection_DeferredReplacement(t *testing.T) {
	path := filepath.Join(t.TempDir(), "snapshot.db")
	w := compatibilityDB(t, path, false)
	r := compatibilityDB(t, path, true)
	var mode string
	if err := w.QueryRow("PRAGMA journal_mode=WAL").Scan(&mode); err != nil || mode != "wal" {
		t.Fatalf("journal=%s err=%v", mode, err)
	}
	if _, err := w.Exec("CREATE TABLE probe (value INTEGER)"); err != nil {
		t.Fatal(err)
	}
	held, err := w.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer held.Rollback()
	r.SetMaxIdleConns(0)
	for i := 0; i < 2; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		tx, err := r.BeginTx(ctx, nil) // No ReadOnly override: the DSN itself must be deferred.
		if err != nil {
			cancel()
			t.Fatal(err)
		}
		var count int
		if err := tx.QueryRow("SELECT count(*) FROM probe").Scan(&count); err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec("INSERT INTO probe VALUES (1)"); err == nil {
			t.Fatal("reader write accepted")
		}
		if err := tx.Rollback(); err != nil {
			t.Fatal(err)
		}
		cancel()
	}
}

func TestWriterConnection_ReleaseDuringAcquisition(t *testing.T) {
	path := filepath.Join(t.TempDir(), "release.db")
	a := compatibilityDB(t, path, false)
	b := compatibilityDB(t, path, false)
	b.SetMaxIdleConns(0)
	for i := 0; i < 2; i++ {
		held, err := a.Begin()
		if err != nil {
			t.Fatal(err)
		}
		done := make(chan error, 1)
		timer := time.AfterFunc(100*time.Millisecond, func() { done <- held.Rollback() })
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		tx, err := b.BeginTx(ctx, nil)
		if err != nil {
			timer.Stop()
			held.Rollback()
			cancel()
			t.Fatal(err)
		}
		if err := <-done; err != nil {
			t.Fatal(err)
		}
		if err := tx.Rollback(); err != nil {
			t.Fatal(err)
		}
		cancel()
	}
}

func TestWriterConnection_ExtendedBusyAcquisition(t *testing.T) {
	path := filepath.Join(t.TempDir(), "extended-busy.db")
	a := compatibilityDB(t, path, false)
	b := compatibilityDB(t, path, false)
	for _, query := range []string{"PRAGMA journal_mode=WAL", "CREATE TABLE probe(value INTEGER)", "BEGIN"} {
		if _, err := a.Exec(query); err != nil {
			t.Fatal(err)
		}
	}
	defer a.Exec("ROLLBACK")
	var count int
	if err := a.QueryRow("SELECT count(*) FROM probe").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if _, err := b.Exec("INSERT INTO probe VALUES (1)"); err != nil {
		t.Fatal(err)
	}
	_, busy := a.Exec("INSERT INTO probe VALUES (2)")
	var concrete *sqlite.Error
	if !errors.As(busy, &concrete) || concrete.Code() != 517 {
		t.Fatalf("expected real SQLITE_BUSY_SNAPSHOT: %v", busy)
	}
	// Replay only acquisition, before a transaction or callback has been admitted.
	attempts := 0
	tx := &faultTx{}
	c := &connection{sqliteConn: &faultConnection{
		exec: func(string) error { return nil },
		begin: func() (driver.Tx, error) {
			attempts++
			if attempts == 1 {
				return nil, busy
			}
			return tx, nil
		},
	}}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	got, err := c.BeginTx(ctx, driver.TxOptions{})
	if err != nil || got != tx || attempts != 2 {
		t.Fatalf("extended busy aborted acquisition: tx=%v attempts=%d err=%v", got, attempts, err)
	}
}
