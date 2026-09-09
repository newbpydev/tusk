package service

import (
	"context"
	"errors"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"reflect"
	"testing"
	"time"
)

func TestList_DefaultsParityAndDay(t *testing.T) {
	a := task("a", "", 0, core.StatusTodo)
	a.DueDate = ptr(time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC))
	a.Priority = core.PriorityHigh
	a.Tags = []core.Tag{"work"}
	a.Description = "Unicode 界 %_ '"
	b := task("b", "", 100, core.StatusDone)
	b.DueDate = ptr(time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC))
	c := task("c", "", 0, core.StatusBlocked)
	r := repository(t, a, b, c)
	s := serviceFor(t, r, false)
	for _, q := range []ports.TaskQuery{{}, {All: true}, {Filter: core.TaskFilter{Statuses: []core.Status{core.StatusDone}}}, {Filter: core.TaskFilter{Tags: []core.Tag{"work"}, SearchTerm: "界 %_ '"}}, {Due: ptr("today")}, {All: true, Due: ptr("today")}} {
		got, e := s.ListTasks(context.Background(), q)
		if e != nil {
			t.Fatal(e)
		}
		prepared, _ := prepareQuery(context.Background(), q)
		want := core.FilterTasks([]core.Task{a, b, c}, prepared.Filter)
		if q.Due != nil {
			want = []core.Task{a}
		}
		core.SortTasks(want, defaultOrder())
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%+v: got %v want %v", q, got, want)
		}
	}
	got, e := s.ListTasks(context.Background(), ports.TaskQuery{Filter: core.TaskFilter{DueBefore: a.DueDate, DueAfter: a.DueDate}})
	if e != nil || got == nil || len(got) != 0 {
		t.Fatalf("%v %v", got, e)
	}
	v, e := s.GetTask(context.Background(), " a ")
	if e != nil {
		t.Fatal(e)
	}
	v.Tags[0] = "changed"
	*v.DueDate = time.Time{}
	if mustGet(t, r, "a").Tags[0] != "work" {
		t.Fatal("alias")
	}
}
func TestTree_SelectedRootPreservesParent(t *testing.T) {
	r := repository(t, task("g", "", 0, core.StatusTodo), task("p", "g", 100, core.StatusTodo), task("a", "p", 100, core.StatusDone))
	s := serviceFor(t, r, false)
	tree, e := s.GetTaskTree(context.Background(), "p")
	if e != nil || len(tree) != 1 || tree[0].Depth != 1 || *tree[0].Task.ParentID != "g" || tree[0].Children[0].Depth != 2 {
		t.Fatalf("%+v %v", tree, e)
	}
	forest, e := s.GetTaskTree(context.Background(), "")
	if e != nil || forest[0].Children[0].Depth != 2 {
		t.Fatalf("%+v %v", forest, e)
	}
	if _, e = s.GetTaskTree(context.Background(), "missing"); !errors.Is(e, core.ErrTaskNotFound) {
		t.Fatal(e)
	}
	empty := serviceFor(t, repository(t), false)
	v, e := empty.GetTaskTree(context.Background(), "")
	if e != nil || v == nil || len(v) != 0 {
		t.Fatalf("%v %v", v, e)
	}
}
func TestStats_RetainedWindow(t *testing.T) {
	now := fixedTime()
	a := task("a", "", 100, core.StatusDone)
	a.CompletedAt = &now
	b := task("b", "", 100, core.StatusDone)
	b.CompletedAt = ptr(now.Add(-168 * time.Hour))
	c := task("c", "", 0, core.StatusTodo)
	c.DueDate = ptr(now.Add(-time.Nanosecond))
	d := task("d", "", 0, core.StatusBlocked)
	d.DueDate = &now
	r := repository(t, a, b, c, d)
	s := serviceFor(t, r, false)
	stats, e := s.GetStats(context.Background())
	if e != nil || stats.Total != 4 || stats.Done != 2 || stats.CompletionPercent != 50 || stats.CompletedLast7Days != 1 || stats.Overdue != 1 || len(stats.ByStatus) != 4 {
		t.Fatalf("%+v %v", stats, e)
	}
	stats.ByStatus[core.StatusTodo] = 99
	again, e := s.GetStats(context.Background())
	if e != nil || again.ByStatus[core.StatusTodo] != 1 {
		t.Fatalf("%+v %v", again, e)
	}
	empty := serviceFor(t, repository(t), false)
	z, e := empty.GetStats(context.Background())
	if e != nil || z.Total != 0 || z.CompletionPercent != 0 || len(z.ByStatus) != 4 {
		t.Fatalf("%+v %v", z, e)
	}
}

type readFailureRepository struct {
	ports.TaskRepository
	cleanup bool
	fail    string
}
type readFailureReader struct {
	ports.TaskReader
	fail string
}

func (r readFailureReader) GetByID(ctx context.Context, id string) (*core.Task, error) {
	if r.fail == "get" {
		return nil, ports.ErrCorrupt
	}
	return r.TaskReader.GetByID(ctx, id)
}
func (r readFailureReader) List(ctx context.Context, f core.TaskFilter) ([]core.Task, error) {
	if r.fail == "list" {
		return nil, ports.ErrCorrupt
	}
	if r.fail == "orphan" {
		return []core.Task{task("a", "missing", 0, core.StatusTodo)}, nil
	}
	return r.TaskReader.List(ctx, f)
}
func (r readFailureReader) GetSubtree(ctx context.Context, id string) ([]core.Task, error) {
	if r.fail == "subtree" {
		return nil, ports.ErrCorrupt
	}
	return r.TaskReader.GetSubtree(ctx, id)
}
func (r readFailureReader) GetAncestors(ctx context.Context, id string) ([]core.Task, error) {
	if r.fail == "ancestors" {
		return nil, ports.ErrCorrupt
	}
	return r.TaskReader.GetAncestors(ctx, id)
}
func (r readFailureReader) ListEvents(ctx context.Context, id string) ([]ports.TaskEvent, error) {
	if r.fail == "history" {
		return nil, ports.ErrCorrupt
	}
	return r.TaskReader.ListEvents(ctx, id)
}
func (r readFailureRepository) WithRead(ctx context.Context, fn func(context.Context, ports.TaskReader) error) error {
	e := r.TaskRepository.WithRead(ctx, func(ctx context.Context, reader ports.TaskReader) error {
		return fn(ctx, readFailureReader{TaskReader: reader, fail: r.fail})
	})
	if e != nil {
		return e
	}
	if r.cleanup {
		return ports.NewTransactionError("read", ports.ErrStorage)
	}
	return nil
}
func TestRead_NoPartialValues(t *testing.T) {
	r := repository(t, task("a", "", 0, core.StatusTodo))
	s := serviceFor(t, readFailureRepository{TaskRepository: r, cleanup: true}, false)
	if v, e := s.GetTask(context.Background(), "a"); e == nil || v != nil {
		t.Fatalf("%v %v", v, e)
	}
	if v, e := s.ListTasks(context.Background(), ports.TaskQuery{}); e == nil || v != nil {
		t.Fatalf("%v %v", v, e)
	}
	if v, e := s.GetTaskTree(context.Background(), ""); e == nil || v != nil {
		t.Fatalf("%v %v", v, e)
	}
	if v, e := s.GetStats(context.Background()); e == nil || v.ByStatus != nil || v.Total != 0 {
		t.Fatalf("%v %v", v, e)
	}
	if v, e := s.GetTaskHistory(context.Background(), "a"); e == nil || v != nil {
		t.Fatalf("%v %v", v, e)
	}
	if v, e := s.PreviewDeleteTask(context.Background(), "a"); e == nil || v.IDs != nil || v.Target.ID != "" {
		t.Fatalf("%v %v", v, e)
	}
}

func TestRead_PreflightAndRepositoryFailures(t *testing.T) {
	r := repository(t, task("p", "", 0, core.StatusTodo), task("a", "p", 0, core.StatusTodo))
	s := serviceFor(t, r, false)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, read := range []func(context.Context) error{func(ctx context.Context) error { _, e := s.GetTask(ctx, "a"); return e }, func(ctx context.Context) error { _, e := s.ListTasks(ctx, ports.TaskQuery{}); return e }, func(ctx context.Context) error { _, e := s.GetTaskTree(ctx, ""); return e }, func(ctx context.Context) error { _, e := s.GetStats(ctx); return e }, func(ctx context.Context) error { _, e := s.GetTaskHistory(ctx, "a"); return e }} {
		if e := read(ctx); !errors.Is(e, context.Canceled) {
			t.Fatal(e)
		}
	}
	for _, read := range []func() error{func() error { _, e := s.GetTask(context.Background(), ""); return e }, func() error { _, e := s.GetTaskTree(context.Background(), " "); return e }, func() error { _, e := s.GetTaskHistory(context.Background(), ""); return e }, func() error { _, e := s.ListTasks(context.Background(), ports.TaskQuery{Due: ptr("wrong")}); return e }, func() error {
		_, e := s.ListTasks(context.Background(), ports.TaskQuery{All: true, Filter: core.TaskFilter{Statuses: []core.Status{core.StatusTodo}}})
		return e
	}} {
		if e := read(); e == nil {
			t.Fatal("invalid read accepted")
		}
	}
	for _, kind := range []string{"get", "list", "subtree", "ancestors", "history", "orphan"} {
		s := serviceFor(t, readFailureRepository{TaskRepository: r, fail: kind}, false)
		switch kind {
		case "get":
			if _, e := s.GetTask(context.Background(), "a"); e == nil {
				t.Fatal(kind)
			}
			if _, e := s.PreviewDeleteTask(context.Background(), "a"); e == nil {
				t.Fatal(kind)
			}
		case "list":
			if _, e := s.ListTasks(context.Background(), ports.TaskQuery{}); e == nil {
				t.Fatal(kind)
			}
			if _, e := s.GetStats(context.Background()); e == nil {
				t.Fatal(kind)
			}
			if _, e := s.GetTaskTree(context.Background(), ""); e == nil {
				t.Fatal(kind)
			}
		case "subtree", "ancestors":
			if _, e := s.GetTaskTree(context.Background(), "a"); e == nil {
				t.Fatal(kind)
			}
		case "history":
			if _, e := s.GetTaskHistory(context.Background(), "a"); e == nil {
				t.Fatal(kind)
			}
		case "orphan":
			if _, e := s.GetTaskTree(context.Background(), ""); e == nil {
				t.Fatal(kind)
			}
		}
	}
	s.options.Clock = func() time.Time { return time.Time{} }
	if _, e := s.ListTasks(context.Background(), ports.TaskQuery{Due: ptr("today")}); !errors.Is(e, ports.ErrInvalidReferenceTime) {
		t.Fatal(e)
	}
	if _, e := s.GetStats(context.Background()); !errors.Is(e, ports.ErrInvalidReferenceTime) {
		t.Fatal(e)
	}
}
func TestTree_SiblingOrderAndDepth(t *testing.T) {
	r := repository(t, task("z", "", 0, core.StatusTodo), task("a", "", 0, core.StatusTodo), task("c", "z", 0, core.StatusTodo), task("b", "z", 0, core.StatusTodo))
	s := serviceFor(t, r, false)
	forest, e := s.GetTaskTree(context.Background(), "")
	if e != nil || forest[0].Task.ID != "a" || forest[1].Children[0].Task.ID != "b" {
		t.Fatalf("%v %v", forest, e)
	}
	selected, e := s.GetTaskTree(context.Background(), "c")
	if e != nil || selected[0].Depth != 1 || *selected[0].Task.ParentID != "z" {
		t.Fatalf("%v %v", selected, e)
	}
}
