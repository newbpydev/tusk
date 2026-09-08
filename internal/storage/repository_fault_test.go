package storage

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"

	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	generated "github.com/newbpydev/tusk/internal/storage/sqlc"
)

type queryFaultConnector struct {
	driver.Connector
	needle  string
	iterate bool
}

func (f queryFaultConnector) Connect(ctx context.Context) (driver.Conn, error) {
	c, err := f.Connector.Connect(ctx)
	if err != nil {
		return nil, err
	}
	return &queryFaultConn{sqliteConn: c.(sqliteConn), needle: f.needle, iterate: f.iterate}, nil
}

type queryFaultConn struct {
	sqliteConn
	needle  string
	iterate bool
}

func (c *queryFaultConn) ExecContext(ctx context.Context, q string, args []driver.NamedValue) (driver.Result, error) {
	if strings.Contains(q, "-- name: "+c.needle+" ") {
		return nil, errors.New("PRIVATE-NOTES injected statement failure")
	}
	return c.sqliteConn.ExecContext(ctx, q, args)
}
func (c *queryFaultConn) QueryContext(ctx context.Context, q string, args []driver.NamedValue) (driver.Rows, error) {
	if strings.Contains(q, "-- name: "+c.needle+" ") {
		if !c.iterate {
			return nil, errors.New("PRIVATE-NOTES injected query failure")
		}
		rows, err := c.sqliteConn.QueryContext(ctx, q, args)
		if err != nil {
			return nil, err
		}
		return &faultRows{Rows: rows}, nil
	}
	return c.sqliteConn.QueryContext(ctx, q, args)
}

type faultRows struct {
	driver.Rows
	seen bool
}

func (r *faultRows) Next(values []driver.Value) error {
	if r.seen {
		return errors.New("PRIVATE-NOTES iteration failure")
	}
	r.seen = true
	return r.Rows.Next(values)
}

func TestRepository_QueryFailuresNeverReturnPartialValues(t *testing.T) {
	for _, needle := range []string{"GetTask", "ListCandidates", "GetAncestors", "GetSubtree", "ListEvents", "ListChildren", "UpdateTask", "DeleteTask", "AppendEvent"} {
		t.Run(needle, func(t *testing.T) {
			r, path := diskRepository(t)
			createFixture(t, r, taskFixture("task"))
			base, err := newConnector(path, false)
			if err != nil {
				t.Fatal(err)
			}
			r.writer.Close()
			r.writer = sql.OpenDB(queryFaultConnector{Connector: base, needle: needle})
			r.writer.SetMaxOpenConns(1)
			err = r.WithWrite(context.Background(), func(c context.Context, w ports.TaskWriter) error {
				switch needle {
				case "GetTask":
					_, err := w.GetByID(c, "task")
					return err
				case "ListCandidates":
					rows, err := w.List(c, core.TaskFilter{})
					if rows != nil {
						t.Error("partial list")
					}
					return err
				case "GetAncestors":
					_, err := w.GetAncestors(c, "task")
					return err
				case "GetSubtree":
					_, err := w.GetSubtree(c, "task")
					return err
				case "ListChildren":
					_, err := w.ListChildren(c, "task")
					return err
				case "ListEvents":
					_, err := w.ListEvents(c, "task")
					return err
				case "UpdateTask":
					return w.Update(c, taskFixture("task"))
				case "DeleteTask":
					_, err := w.Delete(c, "task", true)
					return err
				default:
					_, err := w.AppendEvent(c, ports.TaskEvent{TaskID: "task", Kind: ports.EventCreate, ChangedFields: []string{"title"}, OccurredAt: taskFixture("task").CreatedAt})
					return err
				}
			})
			if !errors.Is(err, ports.ErrStorage) || strings.Contains(err.Error(), "PRIVATE-NOTES") {
				t.Fatalf("query failure=%v", err)
			}
		})
	}
	r, path := diskRepository(t)
	createFixture(t, r, taskFixture("a"))
	createFixture(t, r, taskFixture("b"))
	r.reader.Close()
	base, err := newConnector(path, true)
	if err != nil {
		t.Fatal(err)
	}
	r.reader = sql.OpenDB(queryFaultConnector{base, "ListCandidates", true})
	if rows, err := r.List(context.Background(), core.TaskFilter{}); rows != nil || !errors.Is(err, ports.ErrStorage) {
		t.Fatalf("iteration exposed partial rows: %v %v", rows, err)
	}
}

func TestHandle_CancellationReadOnlyAndLatchedFailure(t *testing.T) {
	r, _ := memoryRepository(t)
	ctx := context.Background()
	h := newTaskHandle(ctx, generated.New(r.reader), false)
	if err := h.Create(ctx, taskFixture("a")); !errors.Is(err, ports.ErrReadOnly) {
		t.Fatal(err)
	}
	dead, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := h.List(dead, core.TaskFilter{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	h.ctx = dead
	if _, err := h.List(ctx, core.TaskFilter{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	err := r.WithWrite(ctx, func(c context.Context, w ports.TaskWriter) error {
		_ = w.Create(c, nil)
		if _, err := w.GetByID(c, "missing"); !errors.Is(err, ports.ErrInvalidRecord) {
			t.Errorf("latched failure changed: %v", err)
		}
		return nil
	})
	if !errors.Is(err, ports.ErrInvalidRecord) {
		t.Fatal(err)
	}
}
