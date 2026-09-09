package service

import (
	"context"
	"errors"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"testing"
)

func patchForTest(t *testing.T, r ports.TaskRepository, cmd ports.UpdateTaskCommand) (*core.Task, error) {
	t.Helper()
	cmd, e := prepareUpdate(context.Background(), cmd)
	if e != nil {
		return nil, e
	}
	var result *core.Task
	e = r.WithWrite(context.Background(), func(ctx context.Context, w ports.TaskWriter) error {
		c := newChanges(ctx, w, fixedTime().AddDate(0, 0, 1), false)
		v, e := c.get(cmd.ID)
		if e != nil {
			return e
		}
		if e = c.patch(v, cmd, nil); e != nil {
			return e
		}
		if e = c.recompute(); e != nil {
			return e
		}
		if e = c.flush(); e != nil {
			return e
		}
		result = v
		return nil
	})
	if e != nil {
		return nil, e
	}
	return result, nil
}
func TestPatch_SetClearNoOp(t *testing.T) {
	v := task("a", "", 10, core.StatusTodo)
	v.Description = "notes"
	v.Tags = []core.Tag{"work"}
	v.DueDate = ptr(fixedTime())
	r := repository(t, v)
	before := mustGet(t, r, "a")
	got, e := patchForTest(t, r, ports.UpdateTaskCommand{ID: "a"})
	if e != nil || !got.UpdatedAt.Equal(before.UpdatedAt) {
		t.Fatalf("%v %v", got, e)
	}
	got, e = patchForTest(t, r, ports.UpdateTaskCommand{ID: "a", Title: ptr(" changed "), Description: ptr(""), Tags: ptr([]string{}), Priority: ptr(core.PriorityHigh), ClearDue: true})
	if e != nil || got.Title != "changed" || got.Description != "" || len(got.Tags) != 0 || got.DueDate != nil || got.Priority != core.PriorityHigh {
		t.Fatalf("%+v %v", got, e)
	}
	base := before.Clone()
	if _, e = patchForTest(t, r, ports.UpdateTaskCommand{ID: "a", Base: &base}); !errors.Is(e, ports.ErrConflict) {
		t.Fatal(e)
	}
	got, e = patchForTest(t, r, ports.UpdateTaskCommand{ID: "a", Progress: ptr(99)})
	if e != nil || got.Progress != 99 || got.Title != "changed" {
		t.Fatalf("%+v %v", got, e)
	}
	if _, e = patchForTest(t, r, ports.UpdateTaskCommand{ID: "a", Progress: ptr(100)}); !errors.Is(e, core.ErrInvalidProgress) {
		t.Fatal(e)
	}
}
func TestPatch_ParentAndDoneProgress(t *testing.T) {
	r := repository(t, task("p", "", 0, core.StatusTodo), task("a", "p", 100, core.StatusDone))
	if _, e := patchForTest(t, r, ports.UpdateTaskCommand{ID: "p", Progress: ptr(0)}); !errors.Is(e, core.ErrInvalidProgress) {
		t.Fatal(e)
	}
	if _, e := patchForTest(t, r, ports.UpdateTaskCommand{ID: "a", Progress: ptr(100)}); e != nil {
		t.Fatal(e)
	}
	if _, e := patchForTest(t, r, ports.UpdateTaskCommand{ID: "a", Progress: ptr(0)}); !errors.Is(e, core.ErrInvalidProgress) {
		t.Fatal(e)
	}
}
