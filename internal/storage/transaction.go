package storage

import (
	"context"
	"database/sql"
	"errors"
	"sync"

	"github.com/newbpydev/tusk/internal/ports"
	generated "github.com/newbpydev/tusk/internal/storage/sqlc"
)

type transactionKey struct{ repository *Repository }
type readHandle struct{ ports.TaskReader }

type taskHandle struct {
	q            *generated.Queries
	ctx          context.Context
	write        bool
	mu           sync.Mutex
	cond         *sync.Cond
	busy, closed bool
	failure      error
}

func newTaskHandle(ctx context.Context, q *generated.Queries, write bool) *taskHandle {
	h := &taskHandle{ctx: ctx, q: q, write: write}
	h.cond = sync.NewCond(&h.mu)
	return h
}

func (h *taskHandle) finish() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.closed = true
	for h.busy {
		h.cond.Wait()
	}
	return h.failure
}

// handleCall centralizes lifetime, concurrent-use rejection and failure latching.
// Callback cancellation and an operation-specific cancellation both stop SQL.
func handleCall[T any](h *taskHandle, ctx context.Context, write bool, fn func(context.Context) (T, error)) (value T, err error) {
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return value, ports.ErrTransactionClosed
	}
	if h.busy {
		if h.write && h.failure == nil {
			h.failure = ports.ErrTransactionInUse
		}
		h.mu.Unlock()
		return value, ports.ErrTransactionInUse
	}
	if h.failure != nil {
		err = h.failure
		h.mu.Unlock()
		return value, err
	}
	h.busy = true
	h.mu.Unlock()
	defer func() {
		h.mu.Lock()
		if err != nil && h.write && h.failure == nil {
			h.failure = err
		}
		h.busy = false
		h.cond.Broadcast()
		h.mu.Unlock()
	}()
	if write && !h.write {
		return value, ports.ErrReadOnly
	}
	if err := h.ctx.Err(); err != nil {
		return value, err
	}
	if err := ctx.Err(); err != nil {
		return value, err
	}
	operation, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(h.ctx, cancel)
	defer func() { stop(); cancel() }()
	return fn(operation)
}

func (r *Repository) WithRead(ctx context.Context, callback func(context.Context, ports.TaskReader) error) error {
	if callback == nil {
		return ports.ErrInvalidCallback
	}
	return r.transact(ctx, false, func(c context.Context, h *taskHandle) error { return callback(c, readHandle{h}) })
}
func (r *Repository) WithWrite(ctx context.Context, callback func(context.Context, ports.TaskWriter) error) error {
	if callback == nil {
		return ports.ErrInvalidCallback
	}
	return r.transact(ctx, true, func(c context.Context, h *taskHandle) error { return callback(c, h) })
}

func rollbackTransaction(tx *sql.Tx, conn *sql.Conn, cause error) error {
	if err := tx.Rollback(); err != nil {
		discardConnection(conn)
		if !errors.Is(err, sql.ErrTxDone) {
			return ports.NewTransactionError("rollback", errors.Join(storageCause(cause), storageCause(err)))
		}
	}
	return cause
}

func (r *Repository) transact(ctx context.Context, write bool, callback func(context.Context, *taskHandle) error) error {
	if err := r.admit(ctx); err != nil {
		return err
	}
	defer r.active.Done()
	pool := r.reader
	if write {
		pool = r.writer
	}
	conn, err := pool.Conn(ctx)
	if err != nil {
		return storageCause(err)
	}
	defer conn.Close()
	tx, err := conn.BeginTx(ctx, &sql.TxOptions{ReadOnly: !write})
	if err != nil {
		return storageCause(err)
	}
	callbackContext := context.WithValue(ctx, transactionKey{r}, true)
	h := newTaskHandle(callbackContext, generated.New(tx), write)
	completed := false
	defer func() {
		h.finish()
		if !completed {
			_ = rollbackTransaction(tx, conn, nil)
		}
	}()
	err = callback(callbackContext, h)
	failure := h.finish()
	if err == nil {
		err = failure
	}
	if err == nil {
		err = ctx.Err()
	}
	completed = true
	if err != nil || !write {
		return rollbackTransaction(tx, conn, err)
	}
	if err := tx.Commit(); err != nil {
		discardConnection(conn)
		return ports.NewTransactionError("commit", storageCause(err))
	}
	return nil // A known commit remains success even if cancellation follows it.
}
