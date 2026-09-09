package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"reflect"
	"testing"
	"time"
)

func TestComplete_SubtreeAndRepeat(t *testing.T) {
	for _, policy := range []bool{false, true} {
		r := repository(t, task("g", "", 0, core.StatusTodo), task("p", "g", 0, core.StatusTodo), task("a", "p", 0, core.StatusBlocked), task("b", "p", 100, core.StatusDone))
		s := serviceFor(t, r, policy)
		before := mustGet(t, r, "b")
		v, e := s.CompleteTask(context.Background(), ports.TaskCommand{ID: "p"})
		if e != nil || v.Status != core.StatusDone {
			t.Fatalf("%v %v", v, e)
		}
		if !mustGet(t, r, "b").CompletedAt.Equal(*before.CompletedAt) {
			t.Fatal("changed existing completion")
		}
		ev, _ := r.ListEvents(context.Background(), "p")
		s.options.AutoCompleteParent = !policy
		s.options.Clock = func() time.Time { return fixedTime().AddDate(0, 0, 1) }
		v, e = s.CompleteTask(context.Background(), ports.TaskCommand{ID: "p"})
		ev2, _ := r.ListEvents(context.Background(), "p")
		if e != nil || !reflect.DeepEqual(ev, ev2) {
			t.Fatalf("repeat: %v %v", v, e)
		}
	}
}
func TestReopen_ParentAtHundred(t *testing.T) {
	r := repository(t, task("g", "", 100, core.StatusDone), task("p", "g", 100, core.StatusDone), task("a", "p", 100, core.StatusDone))
	s := serviceFor(t, r, true)
	v, e := s.ReopenTask(context.Background(), ports.ReopenTaskCommand{ID: "p", Status: core.StatusTodo})
	if e != nil || v.Status != core.StatusTodo || v.Progress != 100 || mustGet(t, r, "g").Status != core.StatusInProgress || mustGet(t, r, "a").Status != core.StatusDone {
		t.Fatalf("%+v %v", v, e)
	}
	if _, e = s.ReopenTask(context.Background(), ports.ReopenTaskCommand{ID: "a", Status: core.StatusBlocked}); !errors.Is(e, core.ErrInvalidStatusTransition) {
		t.Fatal(e)
	}
	v, e = s.ReopenTask(context.Background(), ports.ReopenTaskCommand{ID: "a", Status: core.StatusInProgress})
	if e != nil || v.Progress != 0 || v.CompletedAt != nil {
		t.Fatalf("%+v %v", v, e)
	}
}
func TestMove_SharedChainsAndDepth(t *testing.T) {
	r := repository(t, task("g", "", 50, core.StatusTodo), task("p", "g", 100, core.StatusTodo), task("q", "g", 0, core.StatusTodo), task("a", "p", 100, core.StatusDone))
	s := serviceFor(t, r, false)
	v, e := s.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: "a", ParentID: ptr("q"), Status: ptr(core.StatusTodo), Due: ptr("tomorrow")})
	if e != nil || v.Progress != 0 || *v.ParentID != "q" || v.DueDate == nil || mustGet(t, r, "p").Progress != 0 || mustGet(t, r, "g").Progress != 0 {
		t.Fatalf("%+v %v", v, e)
	}
	for _, tc := range []struct {
		id, parent string
		want       error
	}{{"g", "a", core.ErrCyclicDependency}, {"q", "a", core.ErrCyclicDependency}, {"a", "a", core.ErrSelfParenting}, {"a", "missing", core.ErrTaskNotFound}} {
		if _, e = s.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: tc.id, ParentID: &tc.parent}); !errors.Is(e, tc.want) {
			t.Fatalf("%+v %v", tc, e)
		}
	}
	v, e = s.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: "a", ClearParent: true})
	if e != nil || v.ParentID != nil {
		t.Fatalf("%+v %v", v, e)
	}
	tasks := []core.Task{}
	parent := ""
	for i := 0; i < 10; i++ {
		id := string(rune('a' + i))
		tasks = append(tasks, task(id, parent, 0, core.StatusTodo))
		parent = id
	}
	tasks = append(tasks, task("root", "", 0, core.StatusTodo))
	r2 := repository(t, tasks...)
	s2 := serviceFor(t, r2, false)
	if _, e = s2.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: "root", ParentID: ptr("j")}); !errors.Is(e, core.ErrMaxDepthExceeded) {
		t.Fatal(e)
	}
	if _, e = s2.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: "root", ParentID: ptr("i")}); e != nil {
		t.Fatal(e)
	}
}
func TestMutation_PreflightAndFailures(t *testing.T) {
	r := repository(t, task("a", "", 0, core.StatusTodo))
	s := serviceFor(t, r, false)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := s.CompleteTask(ctx, ports.TaskCommand{ID: "a"}); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	if _, e := s.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: ""}); e == nil {
		t.Fatal("empty ID")
	}
	if _, e := s.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: "a", Due: ptr("wrong")}); e == nil {
		t.Fatal("date")
	}
	if _, e := s.CompleteTask(context.Background(), ports.TaskCommand{ID: "missing"}); !errors.Is(e, core.ErrTaskNotFound) {
		t.Fatal(e)
	}
	for _, fail := range []string{"get", "children", "update", "event"} {
		s := serviceFor(t, faultRepository{TaskRepository: r, fail: fail}, false)
		v, e := s.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: "a", Progress: ptr(50)})
		if e == nil || v != nil || mustGet(t, r, "a").Progress != 0 {
			t.Fatalf("%s: %+v %v", fail, v, e)
		}
	}
}

type indexedRepository struct {
	ports.TaskRepository
	at    int
	after bool
}
type indexedWriter struct {
	ports.TaskWriter
	at    int
	after bool
	calls *int
}

func (w indexedWriter) write(fn func() error) error {
	*w.calls++
	if *w.calls == w.at && !w.after {
		return ports.ErrStorage
	}
	if e := fn(); e != nil {
		return e
	}
	if *w.calls == w.at {
		return ports.ErrStorage
	}
	return nil
}
func (w indexedWriter) Create(ctx context.Context, v *core.Task) error {
	return w.write(func() error { return w.TaskWriter.Create(ctx, v) })
}
func (w indexedWriter) Update(ctx context.Context, v *core.Task) error {
	return w.write(func() error { return w.TaskWriter.Update(ctx, v) })
}
func (w indexedWriter) AppendEvent(ctx context.Context, v ports.TaskEvent) (seq int64, e error) {
	e = w.write(func() error { seq, e = w.TaskWriter.AppendEvent(ctx, v); return e })
	return
}
func (w indexedWriter) Delete(ctx context.Context, id string, recursive bool) (ids []string, e error) {
	e = w.write(func() error { ids, e = w.TaskWriter.Delete(ctx, id, recursive); return e })
	return
}
func (r indexedRepository) WithWrite(ctx context.Context, fn func(context.Context, ports.TaskWriter) error) error {
	calls := 0
	return r.TaskRepository.WithWrite(ctx, func(ctx context.Context, w ports.TaskWriter) error {
		return fn(ctx, indexedWriter{TaskWriter: w, at: r.at, after: r.after, calls: &calls})
	})
}
func TestMutation_FailureAtEveryWrite(t *testing.T) {
	for _, operation := range []string{"complete", "reopen", "move", "delete", "create"} {
		for _, after := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/after=%v", operation, after), func(t *testing.T) {
				for at := 1; at <= 20; at++ {
					r := repository(t, task("g", "", 100, core.StatusDone), task("p", "g", 100, core.StatusDone), task("a", "p", 100, core.StatusDone), task("q", "", 0, core.StatusTodo))
					if operation == "complete" {
						e := r.WithWrite(context.Background(), func(ctx context.Context, w ports.TaskWriter) error {
							a, err := w.GetByID(ctx, "a")
							if err != nil {
								return err
							}
							a.Status = core.StatusTodo
							a.Progress = 0
							a.CompletedAt = nil
							return w.Update(ctx, a)
						})
						if e != nil {
							t.Fatal(e)
						}
					}
					before, _ := r.List(context.Background(), core.TaskFilter{})
					s := serviceFor(t, indexedRepository{TaskRepository: r, at: at, after: after}, false)
					var v *core.Task
					var d ports.DeleteResult
					var e error
					switch operation {
					case "complete":
						v, e = s.CompleteTask(context.Background(), ports.TaskCommand{ID: "p"})
					case "reopen":
						v, e = s.ReopenTask(context.Background(), ports.ReopenTaskCommand{ID: "a", Status: core.StatusTodo})
					case "move":
						v, e = s.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: "a", ParentID: ptr("q"), Status: ptr(core.StatusTodo)})
					case "delete":
						d, e = s.DeleteTask(context.Background(), ports.DeleteTaskCommand{ID: "p", Force: true, Recursive: true})
					case "create":
						v, e = s.CreateTask(context.Background(), ports.CreateTaskCommand{Title: "new", ParentID: ptr("p")})
					}
					if e == nil {
						return
					}
					if !errors.Is(e, ports.ErrStorage) || v != nil || d.Deleted || d.DeletedIDs != nil {
						t.Fatalf("at %d: %+v %+v %v", at, v, d, e)
					}
					got, _ := r.List(context.Background(), core.TaskFilter{})
					if !reflect.DeepEqual(got, before) {
						t.Fatalf("partial graph at %d", at)
					}
					for _, v := range got {
						events, e := r.ListEvents(context.Background(), v.ID)
						if e != nil || len(events) != 0 {
							t.Fatalf("partial history at %d: %v %v", at, events, e)
						}
					}
				}
				t.Fatal("write bound exceeded")
			})
		}
	}
}
func TestMutation_StatusAndReadFailures(t *testing.T) {
	for _, fail := range []string{"subtree", "empty-subtree", "duplicate-subtree"} {
		r := repository(t, task("a", "", 0, core.StatusTodo))
		s := serviceFor(t, faultRepository{TaskRepository: r, fail: fail}, false)
		if v, e := s.CompleteTask(context.Background(), ports.TaskCommand{ID: "a"}); e == nil || v != nil {
			t.Fatalf("%v %v", v, e)
		}
		if v, e := s.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: "a", ParentID: ptr("x")}); e == nil || v != nil {
			t.Fatalf("%v %v", v, e)
		}
	}
	r := repository(t, task("a", "", 100, core.StatusDone))
	s := serviceFor(t, r, false)
	if _, e := s.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: "a", Status: ptr(core.StatusBlocked)}); !errors.Is(e, core.ErrInvalidStatusTransition) {
		t.Fatal(e)
	}
	s.options.Clock = func() time.Time { return time.Time{} }
	if _, e := s.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: "a"}); !errors.Is(e, ports.ErrInvalidReferenceTime) {
		t.Fatal(e)
	}
	s = serviceFor(t, r, true)
	if _, e := s.ReopenTask(context.Background(), ports.ReopenTaskCommand{ID: "a", Status: core.StatusTodo}); e != nil {
		t.Fatal(e)
	}
	before := mustGet(t, r, "a")
	if _, e := s.ReopenTask(context.Background(), ports.ReopenTaskCommand{ID: "a", Status: core.StatusTodo}); e != nil {
		t.Fatal(e)
	}
	if !mustGet(t, r, "a").UpdatedAt.Equal(before.UpdatedAt) {
		t.Fatal("repeat reopen")
	}
}
