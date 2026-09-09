package service

import (
	"context"
	"errors"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"testing"
	"time"
)

func TestDelete_ConsentChanged(t *testing.T) {
	r := repository(t, task("p", "", 0, core.StatusTodo), task("a", "p", 0, core.StatusTodo))
	s := serviceFor(t, r, false)
	preview, e := s.PreviewDeleteTask(context.Background(), "p")
	if e != nil || len(preview.IDs) != 2 {
		t.Fatalf("%+v %v", preview, e)
	}
	for _, tc := range []struct {
		cmd  ports.DeleteTaskCommand
		want error
	}{{ports.DeleteTaskCommand{ID: "p"}, ports.ErrConfirmationRequired}, {ports.DeleteTaskCommand{ID: "p", Expected: &preview}, ports.ErrChildrenPresent}, {ports.DeleteTaskCommand{ID: "p", Force: true}, ports.ErrChildrenPresent}, {ports.DeleteTaskCommand{ID: "p", Force: true, Expected: &preview}, ports.ErrInvalidCommand}} {
		v, e := s.DeleteTask(context.Background(), tc.cmd)
		if !errors.Is(e, tc.want) || v.Deleted || v.DeletedIDs != nil {
			t.Fatalf("%+v %v", v, e)
		}
	}
	if _, e = s.CreateTask(context.Background(), ports.CreateTaskCommand{Title: "new", ParentID: ptr("p")}); e != nil {
		t.Fatal(e)
	}
	v, e := s.DeleteTask(context.Background(), ports.DeleteTaskCommand{ID: "p", Expected: &preview, Recursive: true})
	if !errors.Is(e, ports.ErrConflict) || v.Deleted {
		t.Fatalf("%+v %v", v, e)
	}
	v, e = s.DeleteTask(context.Background(), ports.DeleteTaskCommand{ID: "p", Force: true, Recursive: true})
	if e != nil || !v.Deleted || v.DeletedCount != 3 {
		t.Fatalf("%+v %v", v, e)
	}
	if _, e = r.GetByID(context.Background(), "p"); !errors.Is(e, core.ErrTaskNotFound) {
		t.Fatal(e)
	}
}
func TestDelete_LeafAndLastChild(t *testing.T) {
	r := repository(t, task("p", "", 50, core.StatusTodo), task("a", "p", 50, core.StatusTodo))
	s := serviceFor(t, r, false)
	preview, e := s.PreviewDeleteTask(context.Background(), "a")
	if e != nil {
		t.Fatal(e)
	}
	result, e := s.DeleteTask(context.Background(), ports.DeleteTaskCommand{ID: "a", Expected: &preview})
	if e != nil || result.DeletedCount != 1 || mustGet(t, r, "p").Progress != 0 {
		t.Fatalf("%+v %v", result, e)
	}
	if p, e := s.PreviewDeleteTask(context.Background(), "missing"); !errors.Is(e, core.ErrTaskNotFound) || p.IDs != nil {
		t.Fatalf("%+v %v", p, e)
	}
}
func TestDelete_MalformedAndMetadataConsent(t *testing.T) {
	r := repository(t, task("a", "", 0, core.StatusTodo))
	s := serviceFor(t, r, false)
	preview, _ := s.PreviewDeleteTask(context.Background(), "a")
	bad := preview
	bad.IDs = []string{"a", "a"}
	if _, e := s.DeleteTask(context.Background(), ports.DeleteTaskCommand{ID: "a", Expected: &bad}); !errors.Is(e, ports.ErrInvalidCommand) {
		t.Fatal(e)
	}
	if _, e := s.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: "a", Title: ptr("changed")}); e != nil {
		t.Fatal(e)
	}
	if _, e := s.DeleteTask(context.Background(), ports.DeleteTaskCommand{ID: "a", Expected: &preview}); !errors.Is(e, ports.ErrConflict) {
		t.Fatal(e)
	}
}

func TestDelete_FailuresAndInvalidInputs(t *testing.T) {
	for _, fail := range []string{"get", "subtree", "children", "delete", "delete-mismatch", "update", "event"} {
		r := repository(t, task("p", "", 50, core.StatusTodo), task("a", "p", 50, core.StatusTodo))
		s := serviceFor(t, faultRepository{TaskRepository: r, fail: fail}, false)
		v, e := s.DeleteTask(context.Background(), ports.DeleteTaskCommand{ID: "a", Force: true})
		if e == nil || v.Deleted || v.DeletedIDs != nil {
			t.Fatalf("%s: %+v %v", fail, v, e)
		}
		mustGet(t, r, "a")
	}
	r := repository(t, task("a", "", 0, core.StatusTodo))
	s := serviceFor(t, r, false)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := s.DeleteTask(ctx, ports.DeleteTaskCommand{ID: "a", Force: true}); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	if _, e := s.PreviewDeleteTask(ctx, "a"); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	if _, e := s.PreviewDeleteTask(context.Background(), ""); e == nil {
		t.Fatal("empty preview ID")
	}
	if _, e := s.DeleteTask(context.Background(), ports.DeleteTaskCommand{Force: true}); e == nil {
		t.Fatal("empty delete ID")
	}
	preview, _ := s.PreviewDeleteTask(context.Background(), "a")
	for _, ids := range [][]string{nil, {"b"}, {"b", "a"}, {" a", "a"}} {
		p := preview
		p.IDs = ids
		if _, e := s.DeleteTask(context.Background(), ports.DeleteTaskCommand{ID: "a", Expected: &p}); e == nil {
			t.Fatal("malformed membership")
		}
	}
	bad := preview
	bad.Target.ID = "x"
	if _, e := s.DeleteTask(context.Background(), ports.DeleteTaskCommand{ID: "a", Expected: &bad}); e == nil {
		t.Fatal("mismatched target")
	}
	s.options.Clock = func() time.Time { return time.Time{} }
	if _, e := s.DeleteTask(context.Background(), ports.DeleteTaskCommand{ID: "a", Force: true}); !errors.Is(e, ports.ErrInvalidReferenceTime) {
		t.Fatal(e)
	}
}
