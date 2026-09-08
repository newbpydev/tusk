package storage

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	"modernc.org/sqlite"
)

// newConnector configures physical connections without accessing the filesystem.
// Its caller supplies an absolute literal path and owns opening and closing pools.
func newConnector(path string, reader bool) (driver.Connector, error) {
	params := url.Values{
		"_pragma": {"foreign_keys(1)", "busy_timeout(5000)", "synchronous(NORMAL)"},
		"_txlock": {"immediate"},
	}
	if reader {
		params.Add("_pragma", "query_only(1)")
		params.Set("_txlock", "deferred")
	}
	literal := filepath.ToSlash(path)
	if filepath.VolumeName(path) != "" && len(literal) > 0 && literal[0] != '/' {
		literal = "/" + literal
	}
	u := url.URL{Scheme: "file", Path: literal, RawQuery: params.Encode()}
	base, err := sqlite.NewConnector(u.String())
	if err != nil {
		return nil, err
	}
	return &connector{Connector: base}, nil
}

// sqliteConn lists the optional database/sql interfaces supplied by the pinned
// driver. Preserve them when interposing on BeginTx (especially pool validity).
type sqliteConn interface {
	driver.Conn
	driver.ConnBeginTx
	driver.ConnPrepareContext
	driver.ExecerContext
	driver.QueryerContext
	driver.Pinger
	driver.SessionResetter
	driver.Validator
}

type connector struct{ driver.Connector }

func (c *connector) Connect(ctx context.Context) (driver.Conn, error) {
	raw, err := c.Connector.Connect(ctx)
	if err != nil {
		return nil, err
	}
	conn, ok := raw.(sqliteConn)
	if !ok {
		return nil, errors.Join(errors.New("storage: incompatible SQLite connection"), raw.Close())
	}
	return &connection{sqliteConn: conn}, nil
}

type connection struct {
	sqliteConn
	poisoned bool
}

func (c *connection) IsValid() bool { return !c.poisoned && c.sqliteConn.IsValid() }

// BeginTx retries only acquisition, before any caller operation. SQLite's
// default busy handler does not promptly observe sqlite3_interrupt. Short
// waits allow cancellation without a goroutine using the connection concurrently.
func (c *connection) BeginTx(ctx context.Context, opts driver.TxOptions) (tx driver.Tx, err error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if opts.ReadOnly || ctx.Done() == nil {
		return c.sqliteConn.BeginTx(ctx, opts)
	}
	if _, err := c.ExecContext(ctx, "PRAGMA busy_timeout=25", nil); err != nil {
		c.poisoned = true
		return nil, err
	}
	defer func() {
		if _, restoreErr := c.ExecContext(context.Background(), "PRAGMA busy_timeout=5000", nil); restoreErr != nil {
			c.poisoned = true
			if tx != nil {
				err = errors.Join(err, tx.Rollback())
				tx = nil
			}
			err = errors.Join(err, fmt.Errorf("storage: restore busy timeout: %w", restoreErr))
		}
	}()
	deadline := time.Now().Add(5 * time.Second)
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		tx, err = c.sqliteConn.BeginTx(ctx, opts)
		if err == nil {
			return tx, err
		}
		tx = nil // The driver may return a typed nil transaction alongside an error.
		if contextErr := ctx.Err(); contextErr != nil {
			return nil, contextErr
		}
		var busy *sqlite.Error
		if !errors.As(err, &busy) || busy.Code() != 5 || !time.Now().Before(deadline) {
			return nil, err
		}
	}
}
