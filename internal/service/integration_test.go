package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestService_DeepLifecycleAndUnrelatedBranch(t *testing.T) {
	rows := []core.Task{}
	parent := ""
	for i := 0; i < 10; i++ {
		id := string(rune('a' + i))
		rows = append(rows, task(id, parent, 100, core.StatusDone))
		parent = id
	}
	rows = append(rows, task("unrelated", "", 40, core.StatusTodo))
	r := repository(t, rows...)
	s := serviceFor(t, r, true)
	v, e := s.ReopenTask(context.Background(), ports.ReopenTaskCommand{ID: "c", Status: core.StatusTodo})
	if e != nil || v.Progress != 100 {
		t.Fatalf("%v %v", v, e)
	}
	for _, id := range []string{"a", "b"} {
		v := mustGet(t, r, id)
		if v.Status != core.StatusInProgress || v.Progress != 100 {
			t.Fatalf("%+v", v)
		}
	}
	if mustGet(t, r, "j").Status != core.StatusDone {
		t.Fatal("descendant changed")
	}
	if _, e = s.CompleteTask(context.Background(), ports.TaskCommand{ID: "c"}); e != nil {
		t.Fatal(e)
	}
	if mustGet(t, r, "a").Status != core.StatusDone {
		t.Fatal("auto completion did not cascade")
	}
	if got := mustGet(t, r, "unrelated"); !reflect.DeepEqual(*got, rows[len(rows)-1]) {
		t.Fatal("unrelated mutation")
	}
	tree, e := s.GetTaskTree(context.Background(), "")
	if e != nil || len(tree) != 2 {
		t.Fatalf("%v %v", tree, e)
	}
	if _, e = s.CreateTask(context.Background(), ports.CreateTaskCommand{Title: "too deep", ParentID: ptr("j")}); !errors.Is(e, core.ErrMaxDepthExceeded) {
		t.Fatal(e)
	}
	if v, e = s.CreateTask(context.Background(), ports.CreateTaskCommand{Title: "depth ten", ParentID: ptr("i")}); e != nil || *v.ParentID != "i" {
		t.Fatalf("depth ten creation: %v %v", v, e)
	}
}

func TestService_CompoundMoveCompletionAndOpenPrecedence(t *testing.T) {
	for _, status := range []core.Status{core.StatusDone, core.StatusTodo, core.StatusInProgress} {
		r := repository(t, task("old", "", 100, core.StatusDone), task("new", "", 100, core.StatusDone), task("p", "old", 100, core.StatusDone), task("a", "p", 100, core.StatusDone))
		s := serviceFor(t, r, true)
		v, e := s.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: "p", ParentID: ptr("new"), Status: &status})
		if e != nil || v.Status != status || v.Progress != 100 || *v.ParentID != "new" {
			t.Fatalf("%s: %v %v", status, v, e)
		}
		want := core.StatusInProgress
		if status == core.StatusDone {
			want = core.StatusDone
		}
		if mustGet(t, r, "new").Status != want || mustGet(t, r, "old").Status != core.StatusDone || mustGet(t, r, "a").Status != core.StatusDone {
			t.Fatal("compound move changed wrong scope or lost explicit open")
		}
	}
}

func TestService_ListOrderAndEmptyRanges(t *testing.T) {
	rows := []core.Task{}
	for _, id := range []string{"z", "b", "a", "c", "d", "e"} {
		rows = append(rows, task(id, "", 0, core.StatusTodo))
	}
	rows[0].Priority = core.PriorityLow
	rows[1].Priority = core.PriorityHigh
	rows[2].Priority = core.PriorityHigh
	rows[1].DueDate = ptr(fixedTime())
	rows[2].DueDate = ptr(fixedTime())
	rows[3].Priority = core.PriorityHigh
	rows[3].DueDate = ptr(fixedTime().Add(time.Hour))
	rows[4].Priority = core.PriorityUrgent
	rows[5].Priority = core.PriorityHigh
	s := serviceFor(t, repository(t, rows...), false)
	got, e := s.ListTasks(context.Background(), ports.TaskQuery{})
	if e != nil {
		t.Fatal(e)
	}
	ids := []string{}
	for _, v := range got {
		ids = append(ids, v.ID)
	}
	if !reflect.DeepEqual(ids, []string{"d", "a", "b", "c", "e", "z"}) {
		t.Fatal(ids)
	}
	for _, after := range []time.Time{fixedTime(), fixedTime().Add(time.Hour)} {
		v, e := s.ListTasks(context.Background(), ports.TaskQuery{Filter: core.TaskFilter{DueBefore: ptr(fixedTime()), DueAfter: &after}})
		if e != nil || v == nil || len(v) != 0 {
			t.Fatalf("empty range: %v %v", v, e)
		}
	}
}
func TestService_RelatedMoveAndMetadataNoOp(t *testing.T) {
	r := repository(t, task("g", "", 50, core.StatusTodo), task("p", "g", 50, core.StatusTodo), task("q", "p", 0, core.StatusTodo), task("a", "p", 100, core.StatusDone))
	s := serviceFor(t, r, false)
	if _, e := s.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: "a", ParentID: ptr("q")}); e != nil {
		t.Fatal(e)
	}
	for _, id := range []string{"g", "p", "q"} {
		if mustGet(t, r, id).Progress != 100 {
			t.Fatal(id)
		}
	}
	base := mustGet(t, r, "a")
	events, _ := s.GetTaskHistory(context.Background(), "a")
	if _, e := s.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: "a", ParentID: ptr("q"), Base: base}); e != nil {
		t.Fatal(e)
	}
	again, _ := s.GetTaskHistory(context.Background(), "a")
	if !reflect.DeepEqual(events, again) {
		t.Fatal("same parent emitted event")
	}
	s.options.AutoCompleteParent = true
	p := mustGet(t, r, "p")
	if _, e := s.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: "p", Title: &p.Title}); e != nil {
		t.Fatal(e)
	}
	if mustGet(t, r, "p").Status != core.StatusTodo {
		t.Fatal("metadata reconciled policy")
	}
	if _, e := s.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: "q", ClearParent: true}); e != nil {
		t.Fatal(e)
	}
	if *mustGet(t, r, "a").ParentID != "q" {
		t.Fatal("promotion changed descendants")
	}
}
func TestService_OneTimeAndExactMetadata(t *testing.T) {
	r := repository(t, task("g", "", 100, core.StatusDone), task("p", "g", 100, core.StatusDone), task("other", "", 35, core.StatusTodo))
	s := serviceFor(t, r, false)
	calls := 0
	now := fixedTime().Add(time.Hour)
	s.options.Clock = func() time.Time { calls++; return now }
	notes := "private 界\n\x1b]8;;file:///x\aquote ' %_"
	v, e := s.CreateTask(context.Background(), ports.CreateTaskCommand{Title: "private title", Description: notes, ParentID: ptr("p"), Due: ptr("today"), Tags: []string{"work"}})
	if e != nil || calls != 1 || v.Description != notes {
		t.Fatalf("%v %v %d", v, e, calls)
	}
	for _, id := range []string{"g", "p", v.ID} {
		x := mustGet(t, r, id)
		if !x.UpdatedAt.Equal(now) {
			t.Fatal("multiple clocks")
		}
		ev, _ := s.GetTaskHistory(context.Background(), id)
		for _, e := range ev {
			if !e.OccurredAt.Equal(now) || strings.Contains(fmt.Sprint(e), "private") {
				t.Fatalf("%v", e)
			}
		}
	}
	if mustGet(t, r, "other").Progress != 35 {
		t.Fatal("unrelated change")
	}
	before := v.Clone()
	got, e := s.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: v.ID, Tags: ptr([]string{"#Work", "work"}), Due: ptr(v.DueDate.Format(time.RFC3339Nano))})
	if e != nil || !got.UpdatedAt.Equal(before.UpdatedAt) {
		t.Fatalf("%v %v", got, e)
	}
	got, e = s.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: v.ID, Description: ptr(""), Tags: ptr([]string{}), ClearDue: true, ClearParent: true})
	if e != nil || got.Description != "" || len(got.Tags) != 0 || got.DueDate != nil || got.ParentID != nil {
		t.Fatalf("%+v %v", got, e)
	}
}
func TestService_IncarnationAndConsentVariants(t *testing.T) {
	for _, change := range []string{"remove", "move", "metadata", "incarnation", "derived"} {
		r := repository(t, task("p", "", 0, core.StatusTodo), task("a", "p", 0, core.StatusTodo), task("other", "", 0, core.StatusTodo))
		s := serviceFor(t, r, false)
		preview, _ := s.PreviewDeleteTask(context.Background(), "p")
		var e error
		switch change {
		case "remove":
			_, e = s.DeleteTask(context.Background(), ports.DeleteTaskCommand{ID: "a", Force: true})
		case "move":
			_, e = s.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: "a", ClearParent: true})
		case "metadata":
			_, e = s.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: "p", Description: ptr("changed")})
		case "derived":
			_, e = s.CompleteTask(context.Background(), ports.TaskCommand{ID: "a"})
			if e == nil {
				_, e = s.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: "other", Title: ptr("unrelated")})
			}
		case "incarnation":
			e = r.WithWrite(context.Background(), func(ctx context.Context, w ports.TaskWriter) error {
				v, e := w.GetByID(ctx, "p")
				if e != nil {
					return e
				}
				child, e := w.GetByID(ctx, "a")
				if e != nil {
					return e
				}
				if _, e = w.Delete(ctx, "p", true); e != nil {
					return e
				}
				v.CreatedAt = v.CreatedAt.Add(time.Hour)
				v.UpdatedAt = v.CreatedAt
				if e = w.Create(ctx, v); e != nil {
					return e
				}
				return w.Create(ctx, child)
			})
		}
		if e != nil {
			t.Fatal(e)
		}
		result, e := s.DeleteTask(context.Background(), ports.DeleteTaskCommand{ID: "p", Expected: &preview, Recursive: true})
		if change == "derived" {
			if e != nil || !result.Deleted {
				t.Fatalf("derived progress conflicted: %v", e)
			}
		} else if !errors.Is(e, ports.ErrConflict) || result.Deleted {
			t.Fatalf("%s: %v %v", change, result, e)
		}
	}
	r := repository(t, task("a", "", 0, core.StatusTodo))
	s := serviceFor(t, r, false)
	base := mustGet(t, r, "a")
	e := r.WithWrite(context.Background(), func(ctx context.Context, w ports.TaskWriter) error {
		v := base.Clone()
		if _, e := w.Delete(ctx, "a", false); e != nil {
			return e
		}
		v.CreatedAt = v.CreatedAt.Add(time.Hour)
		v.UpdatedAt = v.CreatedAt
		return w.Create(ctx, &v)
	})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: "a", Base: base}); !errors.Is(e, ports.ErrConflict) {
		t.Fatal(e)
	}
}
func TestService_FilterOracleAndDSTIntersection(t *testing.T) {
	z, e := time.LoadLocation("America/New_York")
	if e != nil {
		t.Fatal(e)
	}
	start := time.Date(2026, 3, 8, 0, 0, 0, 0, z)
	end := time.Date(2026, 3, 9, 0, 0, 0, 0, z)
	rows := []core.Task{task("p", "", 0, core.StatusTodo)}
	for i := 0; i < 8; i++ {
		v := task(string(rune('a'+i)), "p", 0, core.StatusTodo)
		if i%2 == 0 {
			v.Status = core.StatusDone
			v.Progress = 100
			v.CompletedAt = ptr(fixedTime())
		}
		v.Priority = []core.Priority{core.PriorityLow, core.PriorityMedium, core.PriorityHigh, core.PriorityUrgent}[i%4]
		v.Tags = []core.Tag{"work"}
		v.Description = "界 %_ '"
		if i%3 == 0 {
			v.DueDate = ptr(start.UTC())
		} else if i%3 == 1 {
			v.DueDate = ptr(end.UTC())
		}
		rows = append(rows, v)
	}
	r := repository(t, rows...)
	s := serviceFor(t, r, false)
	s.options.Location = z
	for _, filter := range []core.TaskFilter{{}, {Statuses: []core.Status{core.StatusDone, core.StatusTodo}}, {Priorities: []core.Priority{core.PriorityLow, core.PriorityUrgent}}, {Tags: []core.Tag{"work"}, SearchTerm: "界 %_ '"}, {RootOnly: true}, {ParentID: ptr("p")}} {
		q := ports.TaskQuery{Filter: filter}
		prepared, _ := prepareQuery(context.Background(), q)
		want := core.FilterTasks(rows, prepared.Filter)
		core.SortTasks(want, defaultOrder())
		got, e := s.ListTasks(context.Background(), q)
		if e != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("%+v %v", q, e)
		}
	}
	for _, all := range []bool{false, true} {
		got, e := s.ListTasks(context.Background(), ports.TaskQuery{All: all, Due: ptr("2026-03-08"), Filter: core.TaskFilter{ParentID: ptr("p"), SearchTerm: "界"}})
		if e != nil {
			t.Fatal(e)
		}
		want := 1
		if all {
			want = 3
		}
		if len(got) != want {
			t.Fatalf("all=%v: %v", all, got)
		}
		for _, v := range got {
			if !v.DueDate.Equal(start) {
				t.Fatalf("day boundary: %+v", v)
			}
		}
	}
}
func TestService_RetainedStatsTransitions(t *testing.T) {
	now := fixedTime()
	a := task("a", "", 100, core.StatusDone)
	a.CompletedAt = ptr(now.Add(time.Hour))
	b := task("b", "a", 100, core.StatusDone)
	c := task("c", "", 0, core.StatusTodo)
	r := repository(t, a, b, c)
	s := serviceFor(t, r, false)
	v, e := s.GetStats(context.Background())
	if e != nil || v.Total != 3 || v.CompletionPercent != 66 || v.CompletedLast7Days != 1 {
		t.Fatalf("%v %v", v, e)
	}
	if _, e = s.ReopenTask(context.Background(), ports.ReopenTaskCommand{ID: "b", Status: core.StatusTodo}); e != nil {
		t.Fatal(e)
	}
	v, e = s.GetStats(context.Background())
	if e != nil || v.Done != 0 || v.CompletedLast7Days != 0 {
		t.Fatalf("%v %v", v, e)
	}
	if _, e = s.DeleteTask(context.Background(), ports.DeleteTaskCommand{ID: "a", Force: true, Recursive: true}); e != nil {
		t.Fatal(e)
	}
	v, e = s.GetStats(context.Background())
	if e != nil || v.Total != 1 {
		t.Fatalf("%v %v", v, e)
	}
}

type staticRepository struct {
	ports.TaskRepository
	rows   []core.Task
	cancel context.CancelFunc
}
type staticReader struct {
	ports.TaskReader
	rows   []core.Task
	cancel context.CancelFunc
}

func (r staticRepository) WithRead(ctx context.Context, fn func(context.Context, ports.TaskReader) error) error {
	return r.TaskRepository.WithRead(ctx, func(ctx context.Context, reader ports.TaskReader) error {
		return fn(ctx, staticReader{TaskReader: reader, rows: r.rows, cancel: r.cancel})
	})
}
func (r staticReader) List(context.Context, core.TaskFilter) ([]core.Task, error) {
	if r.cancel != nil {
		r.cancel()
	}
	return r.rows, nil
}
func TestService_TreeCorruptionAndTraversalCancellation(t *testing.T) {
	deep := []core.Task{}
	parent := ""
	for i := 0; i < 11; i++ {
		id := string(rune('a' + i))
		deep = append(deep, task(id, parent, 0, core.StatusTodo))
		parent = id
	}
	for _, rows := range [][]core.Task{{task("a", "missing", 0, core.StatusTodo)}, {task("a", "", 0, core.StatusTodo), task("a", "", 0, core.StatusTodo)}, {task("a", "b", 0, core.StatusTodo), task("b", "a", 0, core.StatusTodo)}, deep} {
		r := repository(t)
		s := serviceFor(t, staticRepository{TaskRepository: r, rows: rows}, false)
		if out, e := s.GetTaskTree(context.Background(), ""); e == nil || out != nil {
			t.Fatalf("%v %v", out, e)
		}
	}
	for _, op := range []string{"list", "tree", "stats"} {
		r := repository(t)
		ctx, cancel := context.WithCancel(context.Background())
		s := serviceFor(t, staticRepository{TaskRepository: r, rows: []core.Task{task("a", "", 0, core.StatusTodo)}, cancel: cancel}, false)
		var e error
		switch op {
		case "list":
			_, e = s.ListTasks(ctx, ports.TaskQuery{})
		case "tree":
			_, e = s.GetTaskTree(ctx, "")
		case "stats":
			_, e = s.GetStats(ctx)
		}
		if !errors.Is(e, context.Canceled) {
			t.Fatalf("%s: %v", op, e)
		}
	}
}
