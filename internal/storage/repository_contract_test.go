package storage

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

func taskIDs(tasks []core.Task) []string {
	ids := make([]string, 0, len(tasks))
	for _, task := range tasks {
		ids = append(ids, task.ID)
	}
	return ids
}

func TestRepository_FilterMatchesCore(t *testing.T) {
	r, _ := memoryRepository(t)
	ctx := context.Background()
	corpus := []core.Task{}
	for i := 0; i < 20; i++ {
		task := taskFixture(fmt.Sprintf("task-%02d", i))
		task.Title = "Ä %_"
		task.Priority = core.Priority(i%4 + 1)
		task.CreatedAt = task.CreatedAt.Add(time.Duration(i%3) * time.Second)
		if i%2 == 1 {
			task.Tags = []core.Tag{"z"}
		}
		if i%3 == 0 {
			task.Status = core.StatusDone
			task.Progress = 100
			task.CompletedAt = &task.UpdatedAt
		}
		if i%4 == 0 {
			due := task.CreatedAt.Add(time.Duration(i) * time.Hour)
			task.DueDate = &due
		}
		if i > 0 && i%2 == 0 {
			parent := "task-00"
			task.ParentID = &parent
		}
		createFixture(t, r, task)
		corpus = append(corpus, *task)
	}
	parent := "task-00"
	before := taskFixture("x").CreatedAt.Add(12 * time.Hour)
	after := taskFixture("x").CreatedAt
	for _, f := range []core.TaskFilter{
		{}, {Statuses: []core.Status{"invalid", core.StatusDone}}, {Priorities: []core.Priority{0, core.PriorityUrgent}}, {Tags: []core.Tag{"a", "z"}}, {RootOnly: true}, {ParentID: &parent}, {RootOnly: true, ParentID: &parent}, {DueBefore: &before}, {DueAfter: &after}, {SearchTerm: "ä"}, {SearchTerm: "%_"}, {SearchTerm: "missing"}, {Statuses: []core.Status{"invalid"}},
	} {
		want := core.FilterTasks(corpus, f)
		core.SortTasks(want, []core.SortOrder{{Field: core.SortByPriority, Direction: core.SortDesc}, {Field: core.SortByDueDate, Direction: core.SortAsc}, {Field: core.SortByCreatedAt, Direction: core.SortAsc}})
		got, err := r.List(ctx, f)
		if err != nil || got == nil || !reflect.DeepEqual(taskIDs(got), taskIDs(want)) {
			t.Fatalf("filter %+v: %v want %v err=%v", f, taskIDs(got), taskIDs(want), err)
		}
	}
}

func TestRepository_TreeReadContract(t *testing.T) {
	r, _ := memoryRepository(t)
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		task := taskFixture(fmt.Sprintf("n%d", i))
		if i > 0 {
			parent := fmt.Sprintf("n%d", i-1)
			task.ParentID = &parent
		}
		createFixture(t, r, task)
	}
	for _, tc := range []struct {
		id       string
		sub, anc int
	}{{"n0", 10, 0}, {"n4", 6, 4}, {"n9", 1, 9}} {
		sub, err := r.GetSubtree(ctx, tc.id)
		if err != nil || len(sub) != tc.sub || sub[0].ID != tc.id {
			t.Fatalf("subtree: %v %v", sub, err)
		}
		anc, err := r.GetAncestors(ctx, tc.id)
		if err != nil || len(anc) != tc.anc || anc == nil {
			t.Fatalf("ancestors: %v %v", anc, err)
		}
	}
	children, err := r.ListChildren(ctx, "n9")
	if err != nil || children == nil || len(children) != 0 {
		t.Fatal("leaf not empty")
	}
	for _, read := range []func(context.Context, string) ([]core.Task, error){r.GetSubtree, r.GetAncestors, r.ListChildren} {
		if tasks, err := read(ctx, "missing"); tasks != nil || !errors.Is(err, core.ErrTaskNotFound) {
			t.Fatalf("missing=%v %v", tasks, err)
		}
	}
}

func TestRepository_TraversalRejectsCorruption(t *testing.T) {
	for _, kind := range []string{"self", "cycle", "orphan", "depth"} {
		r, _ := memoryRepository(t)
		ctx := context.Background()
		parent := taskFixture("a")
		child := taskFixture("b")
		child.ParentID = &parent.ID
		createFixture(t, r, parent)
		createFixture(t, r, child)
		want := error(core.ErrCyclicDependency)
		switch kind {
		case "self":
			if _, err := r.writer.Exec("PRAGMA ignore_check_constraints=ON; UPDATE tasks SET parent_id='a' WHERE id='a'"); err != nil {
				t.Fatal(err)
			}
		case "cycle":
			if _, err := r.writer.Exec("UPDATE tasks SET parent_id='b' WHERE id='a'"); err != nil {
				t.Fatal(err)
			}
		case "orphan":
			want = core.ErrTaskNotFound
			if _, err := r.writer.Exec("PRAGMA foreign_keys=OFF; UPDATE tasks SET parent_id='missing' WHERE id='a'"); err != nil {
				t.Fatal(err)
			}
		case "depth":
			want = core.ErrMaxDepthExceeded
			previous := "b"
			for i := 2; i < 11; i++ {
				task := taskFixture(fmt.Sprintf("n%d", i))
				p := previous
				task.ParentID = &p
				createFixture(t, r, task)
				previous = task.ID
			}
		}
		bounded, cancel := context.WithTimeout(ctx, time.Second)
		if tasks, err := r.GetSubtree(bounded, "a"); tasks != nil || !errors.Is(err, ports.ErrCorrupt) || !errors.Is(err, want) {
			t.Fatalf("%s: rows=%v err=%v", kind, tasks, err)
		}
		cancel()
	}
}

func TestRepository_DeleteContractAndEventRetention(t *testing.T) {
	r, _ := memoryRepository(t)
	ctx := context.Background()
	root := taskFixture("root")
	child := taskFixture("child")
	child.ParentID = &root.ID
	for _, task := range []*core.Task{root, child, taskFixture("unrelated")} {
		createFixture(t, r, task)
	}
	var highest int64
	for _, kind := range []ports.EventKind{ports.EventCreate, ports.EventMetadata, ports.EventStatus, ports.EventMove, ports.EventProgress, ports.EventRollup} {
		if err := r.WithWrite(ctx, func(c context.Context, w ports.TaskWriter) error {
			seq, err := w.AppendEvent(c, ports.TaskEvent{TaskID: child.ID, Kind: kind, ChangedFields: []string{"description", "title"}, OccurredAt: child.CreatedAt})
			if seq <= highest {
				t.Error("sequence reused")
			}
			highest = seq
			return err
		}); err != nil {
			t.Fatal(err)
		}
	}
	if err := r.WithWrite(ctx, func(c context.Context, w ports.TaskWriter) error { _, err := w.Delete(c, root.ID, false); return err }); !errors.Is(err, ports.ErrChildrenPresent) {
		t.Fatal(err)
	}
	if err := r.WithWrite(ctx, func(c context.Context, w ports.TaskWriter) error {
		ids, err := w.Delete(c, root.ID, true)
		if !reflect.DeepEqual(ids, []string{"child", "root"}) {
			t.Errorf("deleted=%v", ids)
		}
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := r.GetByID(ctx, "unrelated"); err != nil {
		t.Fatal(err)
	}
	if err := r.WithWrite(ctx, func(c context.Context, w ports.TaskWriter) error {
		seq, err := w.AppendEvent(c, ports.TaskEvent{TaskID: "unrelated", Kind: ports.EventCreate, ChangedFields: []string{"title"}, OccurredAt: root.CreatedAt})
		if seq <= highest {
			t.Error("deleted maximum sequence reused")
		}
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := r.ListEvents(ctx, child.ID); !errors.Is(err, core.ErrTaskNotFound) {
		t.Fatal(err)
	}
	if err := r.WithWrite(ctx, func(c context.Context, w ports.TaskWriter) error {
		_, err := w.Delete(c, "unrelated", false)
		return err
	}); err != nil {
		t.Fatal(err)
	}
}

func TestRepository_DomainErrors(t *testing.T) {
	r, _ := memoryRepository(t)
	ctx := context.Background()
	createFixture(t, r, taskFixture("exists"))
	for _, tc := range []struct {
		want error
		run  func(context.Context, ports.TaskWriter) error
	}{
		{core.ErrDuplicateTaskID, func(c context.Context, w ports.TaskWriter) error { return w.Create(c, taskFixture("exists")) }},
		{core.ErrTaskNotFound, func(c context.Context, w ports.TaskWriter) error { return w.Update(c, taskFixture("missing")) }},
		{core.ErrTaskNotFound, func(c context.Context, w ports.TaskWriter) error { _, err := w.Delete(c, "missing", true); return err }},
		{ports.ErrInvalidRecord, func(c context.Context, w ports.TaskWriter) error { return w.Create(c, nil) }},
		{ports.ErrInvalidRecord, func(c context.Context, w ports.TaskWriter) error { return w.Update(c, nil) }},
		{ports.ErrInvalidRecord, func(c context.Context, w ports.TaskWriter) error {
			task := taskFixture("exists")
			task.CreatedAt = task.CreatedAt.Add(time.Second)
			return w.Update(c, task)
		}},
		{core.ErrTaskNotFound, func(c context.Context, w ports.TaskWriter) error {
			task := taskFixture("missing-parent")
			p := "no-parent"
			task.ParentID = &p
			return w.Create(c, task)
		}},
		{core.ErrTaskNotFound, func(c context.Context, w ports.TaskWriter) error {
			_, err := w.AppendEvent(c, ports.TaskEvent{TaskID: "missing", Kind: ports.EventCreate, ChangedFields: []string{"title"}, OccurredAt: taskFixture("x").CreatedAt})
			return err
		}},
		{ports.ErrInvalidRecord, func(c context.Context, w ports.TaskWriter) error {
			_, err := w.AppendEvent(c, ports.TaskEvent{Sequence: 1})
			return err
		}},
	} {
		if err := r.WithWrite(ctx, tc.run); !errors.Is(err, tc.want) {
			t.Fatalf("got %v want %v", err, tc.want)
		}
	}
	if _, err := r.GetByID(ctx, ""); !errors.Is(err, core.ErrInvalidTaskID) {
		t.Fatal(err)
	}
	r.Close()
	if _, err := r.List(ctx, core.TaskFilter{}); !errors.Is(err, ports.ErrClosedRepository) {
		t.Fatal(err)
	}
}

func TestRepository_ChildrenReadOnlyRequestedRows(t *testing.T) {
	r, _ := memoryRepository(t)
	root := taskFixture("root")
	child := taskFixture("child")
	child.ParentID = &root.ID
	grandchild := taskFixture("grandchild")
	grandchild.ParentID = &child.ID
	for _, task := range []*core.Task{root, child, grandchild} {
		createFixture(t, r, task)
	}
	if _, err := r.writer.Exec("UPDATE tasks SET tags='[ ]' WHERE id='grandchild'"); err != nil {
		t.Fatal(err)
	}
	children, err := r.ListChildren(context.Background(), root.ID)
	if err != nil || len(children) != 1 || children[0].ID != child.ID {
		t.Fatalf("immediate child query inspected unrelated descendants: %v %v", children, err)
	}
	if _, err := r.GetSubtree(context.Background(), root.ID); !errors.Is(err, ports.ErrCorrupt) {
		t.Fatal("full subtree ignored corrupt descendant")
	}
}

func TestRepository_InvalidRowsReturnNoPartialCollection(t *testing.T) {
	r, _ := memoryRepository(t)
	createFixture(t, r, taskFixture("good"))
	createFixture(t, r, taskFixture("bad"))
	if _, err := r.writer.Exec("UPDATE tasks SET tags='[ ]' WHERE id='bad'"); err != nil {
		t.Fatal(err)
	}
	if rows, err := r.List(context.Background(), core.TaskFilter{}); rows != nil || !errors.Is(err, ports.ErrCorrupt) {
		t.Fatalf("partial tasks=%v err=%v", rows, err)
	}
	if _, err := r.writer.Exec("INSERT INTO task_events(task_id,kind,changed_fields,occurred_at) VALUES('good','create','[\"title\"]','2026-09-08T12:00:00.000000123Z'),('good','create','[\"private_notes\"]','2026-09-08T12:00:00.000000123Z')"); err != nil {
		t.Fatal(err)
	}
	if rows, err := r.ListEvents(context.Background(), "good"); rows != nil || !errors.Is(err, ports.ErrCorrupt) {
		t.Fatalf("partial events=%v err=%v", rows, err)
	}
}
