package service

import (
	"context"
	"errors"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"testing"
)

func TestRollup_ChainAndPolicy(t *testing.T) {
	for _, policy := range []bool{false, true} {
		t.Run(map[bool]string{false: "manual", true: "auto"}[policy], func(t *testing.T) {
			r := repository(t, task("root", "", 0, core.StatusTodo), task("p", "root", 0, core.StatusTodo), task("a", "p", 0, core.StatusTodo), task("b", "p", 100, core.StatusDone), task("c", "p", 0, core.StatusTodo))
			e := r.WithWrite(context.Background(), func(ctx context.Context, w ports.TaskWriter) error {
				c := newChanges(ctx, w, fixedTime().AddDate(0, 0, 1), policy)
				if e := c.affectChain("p"); e != nil {
					return e
				}
				if e := c.recompute(); e != nil {
					return e
				}
				return c.flush()
			})
			if e != nil {
				t.Fatal(e)
			}
			if v := mustGet(t, r, "p"); v.Progress != 33 || v.Status != core.StatusTodo {
				t.Fatalf("%+v", v)
			}
			if v := mustGet(t, r, "root"); v.Progress != 33 {
				t.Fatalf("%+v", v)
			}
			e = r.WithWrite(context.Background(), func(ctx context.Context, w ports.TaskWriter) error {
				c := newChanges(ctx, w, fixedTime().AddDate(0, 0, 2), policy)
				for _, id := range []string{"a", "c"} {
					v, e := c.get(id)
					if e != nil {
						return e
					}
					if e = v.TransitionTo(core.StatusDone, c.now); e != nil {
						return e
					}
					c.statusProgress[id] = true
				}
				if e := c.affectChain("p"); e != nil {
					return e
				}
				if e := c.recompute(); e != nil {
					return e
				}
				return c.flush()
			})
			if e != nil {
				t.Fatal(e)
			}
			for _, id := range []string{"root", "p"} {
				v := mustGet(t, r, id)
				if v.Progress != 100 || (v.Status == core.StatusDone) != policy {
					t.Fatalf("%+v", v)
				}
			}
		})
	}
}
func TestRollup_LastChildAndExplicitOpen(t *testing.T) {
	r := repository(t, task("root", "", 100, core.StatusDone), task("p", "root", 100, core.StatusDone), task("a", "p", 100, core.StatusDone))
	e := r.WithWrite(context.Background(), func(ctx context.Context, w ports.TaskWriter) error {
		c := newChanges(ctx, w, fixedTime().AddDate(0, 0, 1), true)
		p, e := c.get("p")
		if e != nil {
			return e
		}
		if e = p.TransitionTo(core.StatusTodo, c.now); e != nil {
			return e
		}
		c.explicitOpen["p"] = true
		if e = c.affectChain("p"); e != nil {
			return e
		}
		if e = c.recompute(); e != nil {
			return e
		}
		return c.flush()
	})
	if e != nil {
		t.Fatal(e)
	}
	if p := mustGet(t, r, "p"); p.Status != core.StatusTodo || p.Progress != 100 {
		t.Fatalf("%+v", p)
	}
	if p := mustGet(t, r, "root"); p.Status != core.StatusInProgress || p.Progress != 100 {
		t.Fatalf("%+v", p)
	}
	e = r.WithWrite(context.Background(), func(ctx context.Context, w ports.TaskWriter) error {
		c := newChanges(ctx, w, fixedTime().AddDate(0, 0, 2), false)
		a, e := c.get("a")
		if e != nil {
			return e
		}
		if e = c.reparent(a, nil); e != nil {
			return e
		}
		if e = c.recompute(); e != nil {
			return e
		}
		return c.flush()
	})
	if e != nil {
		t.Fatal(e)
	}
	if p := mustGet(t, r, "p"); p.Progress != 0 {
		t.Fatalf("last child: %+v", p)
	}
}
func TestRollup_SharedAncestorOnce(t *testing.T) {
	r := repository(t, task("g", "", 50, core.StatusTodo), task("p", "g", 100, core.StatusTodo), task("q", "g", 0, core.StatusTodo), task("a", "p", 100, core.StatusDone))
	e := r.WithWrite(context.Background(), func(ctx context.Context, w ports.TaskWriter) error {
		c := newChanges(ctx, w, fixedTime().AddDate(0, 0, 1), false)
		a, e := c.get("a")
		if e != nil {
			return e
		}
		if e = c.reparent(a, ptr("q")); e != nil {
			return e
		}
		if e = c.recompute(); e != nil {
			return e
		}
		return c.flush()
	})
	if e != nil {
		t.Fatal(e)
	}
	if v := mustGet(t, r, "g"); v.Progress != 50 {
		t.Fatalf("%+v", v)
	}
	events, e := r.ListEvents(context.Background(), "g")
	if e != nil || len(events) != 0 {
		t.Fatalf("transient shared event: %v %v", events, e)
	}
}
func TestChanges_CancelAndNoOp(t *testing.T) {
	r := repository(t, task("a", "", 0, core.StatusTodo))
	before := mustGet(t, r, "a")
	e := r.WithWrite(context.Background(), func(ctx context.Context, w ports.TaskWriter) error {
		c := newChanges(ctx, w, fixedTime().AddDate(0, 0, 1), true)
		v, e := c.get("a")
		if e != nil {
			return e
		}
		v.UpdatedAt = c.now
		if e = c.affectChain("a"); e != nil {
			return e
		}
		if e = c.recompute(); e != nil {
			return e
		}
		return c.flush()
	})
	if e != nil {
		t.Fatal(e)
	}
	if !mustGet(t, r, "a").UpdatedAt.Equal(before.UpdatedAt) {
		t.Fatal("timestamp-only write")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	c := newChanges(ctx, nil, fixedTime(), false)
	if _, e = c.get("a"); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	if e = c.affectChain("a"); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
}

func TestRollup_UnchangedLeafPreservesManualProgress(t *testing.T) {
	r := repository(t, task("a", "", 40, core.StatusTodo))
	err := r.WithWrite(context.Background(), func(ctx context.Context, w ports.TaskWriter) error {
		c := newChanges(ctx, w, fixedTime().AddDate(0, 0, 1), true)
		if e := c.affectChain("a"); e != nil {
			return e
		}
		if e := c.recompute(); e != nil {
			return e
		}
		return c.flush()
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := mustGet(t, r, "a"); got.Progress != 40 {
		t.Fatalf("untouched leaf progress = %d", got.Progress)
	}
}
