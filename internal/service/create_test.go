package service

import (
	"context"
	"errors"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"reflect"
	"strings"
	"testing"
	"time"
)

func serviceFor(t *testing.T, r ports.TaskRepository, policy bool) *TaskService {
	t.Helper()
	s, e := NewTaskService(r, Options{Clock: fixedTime, NewID: func(time.Time) (string, error) { return "0198ff00-0000-7000-8000-000000000001", nil }, Location: time.UTC, AutoCompleteParent: policy})
	if e != nil {
		t.Fatal(e)
	}
	return s
}

type faultRepository struct {
	ports.TaskRepository
	fail    string
	outcome bool
	cancel  func()
}

func (r faultRepository) WithWrite(ctx context.Context, fn func(context.Context, ports.TaskWriter) error) error {
	e := r.TaskRepository.WithWrite(ctx, func(ctx context.Context, w ports.TaskWriter) error {
		return fn(ctx, faultWriter{TaskWriter: w, fail: r.fail})
	})
	if e != nil {
		return e
	}
	if r.cancel != nil {
		r.cancel()
	}
	if r.outcome {
		return ports.NewTransactionError("commit", ports.ErrStorage)
	}
	return nil
}
func TestCreate_AtomicDefaultsAndParent(t *testing.T) {
	r := repository(t, task("p", "", 100, core.StatusDone))
	s := serviceFor(t, r, false)
	notes := "private\n\x1b]0;x\a"
	v, e := s.CreateTask(context.Background(), ports.CreateTaskCommand{Title: strings.Repeat("界", 255), Description: notes, Tags: []string{"#Work", "work"}, ParentID: ptr("p"), Due: ptr("today")})
	if e != nil {
		t.Fatal(e)
	}
	if v.Status != core.StatusTodo || v.Priority != core.PriorityMedium || v.Progress != 0 || v.Description != notes || v.CompletedAt != nil || !reflect.DeepEqual(v.Tags, []core.Tag{"work"}) || v.DueDate == nil {
		t.Fatalf("%+v", v)
	}
	if p := mustGet(t, r, "p"); p.Status != core.StatusInProgress || p.Progress != 0 {
		t.Fatalf("%+v", p)
	}
	ev, e := r.ListEvents(context.Background(), v.ID)
	if e != nil || len(ev) != 1 || len(ev[0].ChangedFields) != 9 {
		t.Fatalf("%v %v", ev, e)
	}
	v.Tags[0] = "changed"
	if mustGet(t, r, v.ID).Tags[0] != "work" {
		t.Fatal("aliased result")
	}
}
func TestCreate_IdentityFailureAndCollision(t *testing.T) {
	for _, kind := range []string{"bad-shape", "bad-clock", "id-error", "id-format", "bad-date", "missing-parent", "duplicate"} {
		t.Run(kind, func(t *testing.T) {
			r := repository(t)
			s := serviceFor(t, r, false)
			calls := 0
			orig := s.options.NewID
			s.options.NewID = func(now time.Time) (string, error) {
				calls++
				if kind == "id-error" {
					return "", errors.New("PRIVATE")
				}
				if kind == "id-format" {
					return "bad", nil
				}
				return orig(now)
			}
			cmd := ports.CreateTaskCommand{Title: "t"}
			switch kind {
			case "bad-shape":
				cmd.Title = ""
			case "bad-clock":
				s.options.Clock = func() time.Time { return time.Time{} }
			case "bad-date":
				cmd.Due = ptr("wrong")
			case "missing-parent":
				cmd.ParentID = ptr("missing")
			case "duplicate":
				if _, e := s.CreateTask(context.Background(), cmd); e != nil {
					t.Fatal(e)
				}
				calls = 0
			}
			v, e := s.CreateTask(context.Background(), cmd)
			if e == nil || v != nil || strings.Contains(e.Error(), "PRIVATE") || calls > 1 {
				t.Fatalf("%+v %v %d", v, e, calls)
			}
		})
	}
}
func TestCreate_FailureAndUnknownOutcome(t *testing.T) {
	for _, fail := range []string{"create", "update", "event", "unknown"} {
		t.Run(fail, func(t *testing.T) {
			r := repository(t, task("p", "", 100, core.StatusDone))
			s := serviceFor(t, faultRepository{TaskRepository: r, fail: fail, outcome: fail == "unknown"}, false)
			v, e := s.CreateTask(context.Background(), ports.CreateTaskCommand{Title: "t", ParentID: ptr("p")})
			if e == nil || v != nil {
				t.Fatalf("%v %v", v, e)
			}
			rows, err := r.List(context.Background(), core.TaskFilter{})
			if err != nil {
				t.Fatal(err)
			}
			if fail == "unknown" {
				if len(rows) != 2 {
					t.Fatal("lost acknowledged state")
				}
			} else if len(rows) != 1 || mustGet(t, r, "p").Status != core.StatusDone {
				t.Fatal("partial state")
			}
		})
	}
}

func TestCreate_CollisionWithLoadedHierarchy(t *testing.T) {
	id := "0198ff00-0000-7000-8000-000000000001"
	for _, parent := range []bool{false, true} {
		t.Run(map[bool]string{false: "existing-child", true: "existing-parent"}[parent], func(t *testing.T) {
			p := "p"
			fixtures := []core.Task{task(p, "", 0, core.StatusTodo), task(id, p, 0, core.StatusTodo)}
			if parent {
				p = id
				fixtures = []core.Task{task(id, "", 0, core.StatusTodo)}
			}
			r := repository(t, fixtures...)
			s := serviceFor(t, r, false)
			v, e := s.CreateTask(context.Background(), ports.CreateTaskCommand{Title: "replacement", ParentID: &p})
			if v != nil || !errors.Is(e, core.ErrDuplicateTaskID) {
				t.Fatalf("%+v %v", v, e)
			}
			if got := mustGet(t, r, id); got.Title != id {
				t.Fatal("collision replaced existing row")
			}
		})
	}
}
