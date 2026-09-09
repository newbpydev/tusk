package service

import (
	"context"
	"errors"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"testing"
)

type faultWriter struct {
	ports.TaskWriter
	fail    string
	corrupt bool
}

func (w faultWriter) GetByID(ctx context.Context, id string) (*core.Task, error) {
	if w.fail == "get" {
		return nil, ports.ErrStorage
	}
	v, e := w.TaskWriter.GetByID(ctx, id)
	if w.corrupt && v != nil {
		v.Title = ""
	}
	return v, e
}
func (w faultWriter) ListChildren(ctx context.Context, id string) ([]core.Task, error) {
	if w.fail == "children" {
		return nil, ports.ErrStorage
	}
	v, e := w.TaskWriter.ListChildren(ctx, id)
	if w.corrupt && len(v) > 0 {
		v = append(v, v[0])
	}
	return v, e
}
func (w faultWriter) Update(ctx context.Context, v *core.Task) error {
	if w.fail == "update" {
		return ports.ErrStorage
	}
	return w.TaskWriter.Update(ctx, v)
}
func (w faultWriter) Create(ctx context.Context, v *core.Task) error {
	if w.fail == "create" {
		return ports.ErrStorage
	}
	return w.TaskWriter.Create(ctx, v)
}
func (w faultWriter) AppendEvent(ctx context.Context, e ports.TaskEvent) (int64, error) {
	if w.fail == "event" {
		return 0, ports.ErrStorage
	}
	return w.TaskWriter.AppendEvent(ctx, e)
}
func TestChanges_FailureAtomicity(t *testing.T) {
	for _, fail := range []string{"get", "children", "update", "event", "create"} {
		t.Run(fail, func(t *testing.T) {
			r := repository(t, task("p", "", 0, core.StatusTodo), task("a", "p", 0, core.StatusTodo))
			e := r.WithWrite(context.Background(), func(ctx context.Context, w ports.TaskWriter) error {
				c := newChanges(ctx, faultWriter{TaskWriter: w, fail: fail}, fixedTime().AddDate(0, 0, 1), true)
				if fail == "create" {
					v := task("new", "", 0, core.StatusTodo)
					c.current[v.ID] = &v
					return c.flush()
				}
				a, e := c.get("a")
				if e != nil {
					return e
				}
				a.Title = "changed"
				if e = c.affectChain("p"); e != nil {
					return e
				}
				if e = c.recompute(); e != nil {
					return e
				}
				return c.flush()
			})
			if !errors.Is(e, ports.ErrStorage) {
				t.Fatalf("%s: %v", fail, e)
			}
			if mustGet(t, r, "a").Title != "a" {
				t.Fatal("partial write")
			}
			ev, e := r.ListEvents(context.Background(), "a")
			if e != nil || len(ev) != 0 {
				t.Fatalf("%v %v", ev, e)
			}
		})
	}
}
func TestChanges_GraphGuards(t *testing.T) {
	r := repository(t, task("p", "", 0, core.StatusTodo), task("a", "p", 0, core.StatusTodo))
	e := r.WithWrite(context.Background(), func(ctx context.Context, w ports.TaskWriter) error {
		c := newChanges(ctx, w, fixedTime(), false)
		_, e := c.get("missing")
		return e
	})
	if !errors.Is(e, core.ErrTaskNotFound) {
		t.Fatal(e)
	}
	e = r.WithWrite(context.Background(), func(ctx context.Context, w ports.TaskWriter) error {
		c := newChanges(ctx, faultWriter{TaskWriter: w, corrupt: true}, fixedTime(), false)
		if _, e := c.get("a"); !errors.Is(e, ports.ErrCorrupt) {
			t.Fatal(e)
		}
		if _, e := c.children("p"); !errors.Is(e, ports.ErrCorrupt) {
			t.Fatal(e)
		}
		c = newChanges(ctx, w, fixedTime(), false)
		a, _ := c.get("a")
		p, _ := c.get("p")
		p.ParentID = ptr("a")
		if e := c.affectChain("a"); !errors.Is(e, core.ErrCyclicDependency) {
			t.Fatal(e)
		}
		if _, e := c.depth("a"); !errors.Is(e, core.ErrCyclicDependency) {
			t.Fatal(e)
		}
		if e := c.recompute(); !errors.Is(e, core.ErrCyclicDependency) {
			t.Fatal(e)
		}
		a.Title = "changed"
		if e := c.flush(); !errors.Is(e, core.ErrCyclicDependency) {
			t.Fatal(e)
		}
		c = newChanges(ctx, w, fixedTime(), false)
		for i := 0; i < 11; i++ {
			id := string(rune('a' + i))
			v := task(id, "", 0, core.StatusTodo)
			if i < 10 {
				v.ParentID = ptr(string(rune('b' + i)))
			}
			c.current[id] = &v
		}
		if _, e := c.depth("a"); !errors.Is(e, core.ErrMaxDepthExceeded) {
			t.Fatal(e)
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
}
func TestChanges_CreationAndSameParent(t *testing.T) {
	r := repository(t, task("p", "", 0, core.StatusTodo))
	e := r.WithWrite(context.Background(), func(ctx context.Context, w ports.TaskWriter) error {
		c := newChanges(ctx, w, fixedTime(), false)
		v := task("new", "p", 0, core.StatusTodo)
		c.current[v.ID] = &v
		if e := c.reparent(&v, ptr("p")); e != nil {
			return e
		}
		return c.flush()
	})
	if e != nil {
		t.Fatal(e)
	}
	if e, err := r.ListEvents(context.Background(), "new"); err != nil || len(e) != 1 || e[0].Kind != ports.EventCreate {
		t.Fatalf("%v %v", e, err)
	}
}
