package storage

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

type transactionFaultConnector struct {
	driver.Connector
	stage            string
	used             *atomic.Bool
	opened           *atomic.Int32
	cancel           context.CancelFunc
	started, release chan struct{}
}

func (f transactionFaultConnector) Connect(ctx context.Context) (driver.Conn, error) {
	raw, err := f.Connector.Connect(ctx)
	if err != nil {
		return nil, err
	}
	f.opened.Add(1)
	stage := ""
	if f.used.CompareAndSwap(false, true) {
		stage = f.stage
	}
	return &transactionFaultConn{sqliteConn: raw.(sqliteConn), stage: stage, cancel: f.cancel, started: f.started, release: f.release}, nil
}

type transactionFaultConn struct {
	sqliteConn
	stage            string
	cancel           context.CancelFunc
	started, release chan struct{}
}

func (c *transactionFaultConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if c.stage == "begin" {
		c.stage = ""
		return nil, errors.New("private begin SQL")
	}
	tx, err := c.sqliteConn.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &transactionBoundaryTx{Tx: tx, stage: c.stage, cancel: c.cancel}, nil
}
func (c *transactionFaultConn) ExecContext(ctx context.Context, q string, args []driver.NamedValue) (driver.Result, error) {
	if c.stage == "statement-cancel" && strings.Contains(q, "INSERT INTO tasks") {
		c.cancel()
	}
	return c.sqliteConn.ExecContext(ctx, q, args)
}
func (c *transactionFaultConn) QueryContext(ctx context.Context, q string, args []driver.NamedValue) (driver.Rows, error) {
	if c.stage == "block-read" && strings.Contains(q, "-- name: GetTask") {
		c.stage = ""
		close(c.started)
		<-c.release
	}
	return c.sqliteConn.QueryContext(ctx, q, args)
}

type transactionBoundaryTx struct {
	driver.Tx
	stage  string
	cancel context.CancelFunc
}

func (tx *transactionBoundaryTx) Commit() error {
	if tx.stage == "commit-before" {
		return errors.New("PRIVATE-NOTES before commit")
	}
	err := tx.Tx.Commit()
	if err == nil {
		if tx.stage == "commit-after" {
			return errors.New("PRIVATE-NOTES lost acknowledgment")
		}
		if tx.stage == "commit-cancel" {
			tx.cancel()
		}
	}
	return err
}
func (tx *transactionBoundaryTx) Rollback() error {
	err := tx.Tx.Rollback()
	if err == nil && tx.stage == "rollback" {
		return errors.New("PRIVATE-NOTES rollback failure")
	}
	return err
}

func installTransactionFault(t *testing.T, r *Repository, path, stage string, reader bool, cancel context.CancelFunc, started, release chan struct{}) *atomic.Int32 {
	t.Helper()
	pool := r.writer
	if reader {
		pool = r.reader
	}
	if err := pool.Close(); err != nil {
		t.Fatal(err)
	}
	base, err := newConnector(path, reader)
	if err != nil {
		t.Fatal(err)
	}
	opened := new(atomic.Int32)
	pool = sql.OpenDB(transactionFaultConnector{base, stage, new(atomic.Bool), opened, cancel, started, release})
	pool.SetMaxOpenConns(1)
	pool.SetMaxIdleConns(1)
	if reader {
		r.reader = pool
	} else {
		r.writer = pool
	}
	return opened
}

func diskRepository(t *testing.T) (*Repository, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "repository.db")
	r, err := Open(context.Background(), Options{Path: path})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { r.Close() })
	return r, path
}

func TestWithWrite_CommitFailureNeverReplays(t *testing.T) {
	for _, stage := range []string{"commit-before", "commit-after", "rollback"} {
		t.Run(stage, func(t *testing.T) {
			r, path := diskRepository(t)
			opened := installTransactionFault(t, r, path, stage, false, nil, nil, nil)
			calls := 0
			err := r.WithWrite(context.Background(), func(ctx context.Context, w ports.TaskWriter) error {
				calls++
				if err := w.Create(ctx, taskFixture("change")); err != nil {
					return err
				}
				_, err := w.AppendEvent(ctx, ports.TaskEvent{TaskID: "change", Kind: ports.EventCreate, ChangedFields: []string{"title"}, OccurredAt: taskFixture("x").CreatedAt})
				if stage == "rollback" {
					return errors.New("reject callback")
				}
				return err
			})
			var outcome ports.TransactionError
			if !errors.As(err, &outcome) || outcome.Outcome() != "unknown" || calls != 1 || strings.Contains(err.Error(), "PRIVATE-NOTES") {
				t.Fatalf("outcome=%v callbacks=%d", err, calls)
			}
			_, readErr := r.GetByID(context.Background(), "change")
			if stage == "commit-after" {
				if readErr != nil {
					t.Fatal(readErr)
				}
			} else if !errors.Is(readErr, core.ErrTaskNotFound) {
				t.Fatal(readErr)
			}
			if err := r.WithWrite(context.Background(), func(c context.Context, w ports.TaskWriter) error { return w.Create(c, taskFixture("recovery")) }); err != nil {
				t.Fatal(err)
			}
			if opened.Load() < 2 {
				t.Fatal("failed physical connection was reused")
			}
		})
	}
}

func TestTransaction_ReadCleanupReturnsNoPartialValue(t *testing.T) {
	r, path := diskRepository(t)
	createFixture(t, r, taskFixture("task"))
	opened := installTransactionFault(t, r, path, "rollback", true, nil, nil, nil)
	if task, err := r.GetByID(context.Background(), "task"); task != nil || err == nil {
		t.Fatalf("cleanup returned partial value: %v %v", task, err)
	}
	if _, err := r.GetByID(context.Background(), "task"); err != nil {
		t.Fatal(err)
	}
	if opened.Load() < 2 {
		t.Fatal("read cleanup did not discard connection")
	}
}

func TestWithRead_StableSnapshot(t *testing.T) {
	a, path := diskRepository(t)
	b, err := Open(context.Background(), Options{Path: path})
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	createFixture(t, a, taskFixture("task"))
	err = a.WithRead(context.Background(), func(ctx context.Context, reader ports.TaskReader) error {
		first, err := reader.GetByID(ctx, "task")
		if err != nil {
			return err
		}
		if err := b.WithWrite(context.Background(), func(c context.Context, w ports.TaskWriter) error {
			changed := first.Clone()
			changed.Title = "committed"
			if err := w.Update(c, &changed); err != nil {
				return err
			}
			_, err := w.AppendEvent(c, ports.TaskEvent{TaskID: changed.ID, Kind: ports.EventMetadata, ChangedFields: []string{"title"}, OccurredAt: changed.UpdatedAt})
			return err
		}); err != nil {
			return err
		}
		second, err := reader.GetByID(ctx, "task")
		if err != nil || second.Title != first.Title {
			t.Fatal("snapshot changed")
		}
		events, err := reader.ListEvents(ctx, "task")
		if err != nil || len(events) != 0 {
			t.Fatal("snapshot event set changed")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := a.GetByID(context.Background(), "task")
	if err != nil || got.Title != "committed" {
		t.Fatal("new snapshot missed commit")
	}
}

func TestWithWrite_CancellationAtCommit(t *testing.T) {
	for _, stage := range []string{"before", "statement-cancel", "before-commit", "commit-cancel"} {
		t.Run(stage, func(t *testing.T) {
			r, path := diskRepository(t)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if stage == "before" {
				cancel()
			} else {
				installTransactionFault(t, r, path, stage, false, cancel, nil, nil)
			}
			err := r.WithWrite(ctx, func(c context.Context, w ports.TaskWriter) error {
				if err := w.Create(c, taskFixture("task")); err != nil {
					return err
				}
				if stage == "before-commit" {
					cancel()
				}
				return nil
			})
			if stage == "commit-cancel" {
				if err != nil {
					t.Fatalf("known commit lost success: %v", err)
				}
			} else if !errors.Is(err, context.Canceled) {
				t.Fatalf("cancellation=%v", err)
			}
			_, readErr := r.GetByID(context.Background(), "task")
			if stage == "commit-cancel" {
				if readErr != nil {
					t.Fatal(readErr)
				}
			} else if !errors.Is(readErr, core.ErrTaskNotFound) {
				t.Fatal("canceled write persisted")
			}
		})
	}
}

func TestTransaction_ConcurrentHandleAndCompletion(t *testing.T) {
	r, path := diskRepository(t)
	createFixture(t, r, taskFixture("task"))
	started, release := make(chan struct{}), make(chan struct{})
	installTransactionFault(t, r, path, "block-read", false, nil, started, release)
	callbackReturning := make(chan struct{})
	finished := make(chan error, 1)
	operation := make(chan error, 1)
	go func() {
		finished <- r.WithWrite(context.Background(), func(ctx context.Context, w ports.TaskWriter) error {
			go func() { _, err := w.GetByID(ctx, "task"); operation <- err }()
			<-started
			if _, err := w.GetByID(ctx, "task"); !errors.Is(err, ports.ErrTransactionInUse) {
				t.Errorf("concurrent call=%v", err)
			}
			close(callbackReturning)
			return nil
		})
	}()
	<-callbackReturning
	select {
	case err := <-finished:
		t.Fatalf("callback cleaned up active operation: %v", err)
	default:
	}
	close(release)
	if err := <-operation; err != nil {
		t.Fatal(err)
	}
	if err := <-finished; !errors.Is(err, ports.ErrTransactionInUse) {
		t.Fatalf("concurrent failure not latched: %v", err)
	}
}

func TestTransaction_RollbackFailurePreservesCause(t *testing.T) {
	for _, cause := range []error{
		context.Canceled, context.DeadlineExceeded,
		core.ErrTaskNotFound, core.ErrEmptyTitle, core.ErrTitleTooLong,
		core.ErrInvalidStatus, core.ErrInvalidPriority, core.ErrInvalidStatusTransition,
		core.ErrCyclicDependency, core.ErrSelfParenting, core.ErrMaxDepthExceeded,
		core.ErrInvalidTag, core.ErrInvalidProgress, core.ErrInvalidTaskID,
		core.ErrDuplicateTaskID, core.ErrInvalidDepth,
		ports.ErrInvalidRecord, ports.ErrCorrupt, ports.ErrIncompatibleSchema,
		ports.ErrBusy, ports.ErrStorage, ports.ErrReadOnly, ports.ErrClosedRepository,
		ports.ErrInvalidCallback, ports.ErrNestedTransaction, ports.ErrTransactionClosed,
		ports.ErrTransactionInUse, ports.ErrChildrenPresent,
	} {
		t.Run(cause.Error(), func(t *testing.T) {
			r, path := diskRepository(t)
			installTransactionFault(t, r, path, "rollback", false, nil, nil, nil)
			callbackError := fmt.Errorf("PRIVATE-NOTES: %w", cause)
			err := r.WithWrite(context.Background(), func(context.Context, ports.TaskWriter) error { return callbackError })
			var outcome ports.TransactionError
			if !errors.As(err, &outcome) || outcome.Outcome() != "unknown" || !errors.Is(err, cause) || !errors.Is(err, ports.ErrStorage) {
				t.Fatalf("rollback lost cause %v: %v", cause, err)
			}
			if strings.Contains(err.Error(), "PRIVATE-NOTES") || errors.Is(err, callbackError) {
				t.Fatalf("rollback exposed original error: %v", err)
			}
		})
	}
}
