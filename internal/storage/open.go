package storage

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"os"
	"sync"

	assets "github.com/newbpydev/tusk/db"
	"github.com/newbpydev/tusk/internal/ports"
)

const errClosedRepository = ports.ErrClosedRepository

// Options configures explicit storage access. Path overrides environment lookup.
type Options struct{ Path string }

// Repository owns its pools and all admitted operations. Close it after use.
type Repository struct {
	writer, reader *sql.DB
	mu             sync.Mutex
	closed         bool
	active         sync.WaitGroup
	closeOnce      sync.Once
	closeErr       error
}

// Open explicitly opens storage; importing this package never accesses a database.
func Open(ctx context.Context, options Options) (_ *Repository, err error) {
	defer func() {
		if err != nil {
			err = openCause(err)
		}
	}()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	path, err := resolvePath(options.Path, pathInputs{lookup: os.Getenv, home: os.UserHomeDir, cwd: os.Getwd})
	if err != nil {
		return nil, err
	}
	return openAt(ctx, path, connectorFor, prepareFile)
}

type connectorFactory func(string, connectionRole) (driver.Connector, error)

func openAt(ctx context.Context, path string, factory connectorFactory, prepare func(string) error) (_ *Repository, err error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := prepare(path); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	inventory, err := loadMigrations(assets.Migrations())
	if err != nil {
		return nil, err
	}
	inspection, err := factory(path, inspectionConnection)
	if err != nil {
		return nil, err
	}
	probe := sql.OpenDB(inspection)
	probe.SetMaxOpenConns(1)
	read, err := probe.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		probe.Close()
		return nil, err
	}
	_, inspectErr := inspectSchema(ctx, read, inventory)
	err = errors.Join(inspectErr, read.Rollback(), probe.Close())
	if err != nil {
		return nil, err
	}
	r := &Repository{}
	defer func() {
		if err != nil {
			err = errors.Join(err, r.Close())
		}
	}()
	w, err := factory(path, writerConnection)
	if err != nil {
		return nil, err
	}
	r.writer = sql.OpenDB(w)
	r.writer.SetMaxOpenConns(1)
	r.writer.SetMaxIdleConns(1)
	var journal string
	if err := r.writer.QueryRowContext(ctx, "PRAGMA journal_mode=WAL").Scan(&journal); err != nil {
		return nil, err
	}
	if journal != "wal" {
		return nil, fmt.Errorf("storage: WAL journal unavailable")
	}
	if err := migrate(ctx, r.writer, inventory); err != nil {
		return nil, err
	}
	reader, err := factory(path, readerConnection)
	if err != nil {
		return nil, err
	}
	r.reader = sql.OpenDB(reader)
	r.reader.SetMaxOpenConns(4)
	r.reader.SetMaxIdleConns(4)
	if err := r.reader.PingContext(ctx); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Repository) admit(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if ctx.Value(transactionKey{r}) != nil {
		return ports.ErrNestedTransaction
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return errClosedRepository
	}
	r.active.Add(1)
	return nil
}

// Close rejects new work, waits for admitted work, and releases both pools.
// Calling Close inside an admitted callback is prohibited.
func (r *Repository) Close() error {
	r.closeOnce.Do(func() {
		r.mu.Lock()
		r.closed = true
		r.mu.Unlock()
		r.active.Wait()
		if r.reader != nil {
			r.closeErr = r.reader.Close()
		}
		if r.writer != nil {
			r.closeErr = errors.Join(r.closeErr, r.writer.Close())
		}
		r.closeErr = storageCause(r.closeErr)
	})
	return r.closeErr
}
