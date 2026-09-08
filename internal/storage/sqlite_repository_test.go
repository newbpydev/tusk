package storage

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

func taskFixture(id string) *core.Task {
	return &core.Task{ID: id, Title: "title", Description: "  markdown\n", Status: core.StatusTodo, Priority: core.PriorityMedium, Tags: []core.Tag{"a", "z"}, CreatedAt: time.Date(2026, 9, 8, 12, 0, 0, 123, time.UTC), UpdatedAt: time.Date(2026, 9, 8, 12, 0, 0, 123, time.UTC)}
}
func createFixture(t *testing.T, r *Repository, task *core.Task) {
	t.Helper()
	if err := r.WithWrite(context.Background(), func(ctx context.Context, w ports.TaskWriter) error { return w.Create(ctx, task) }); err != nil {
		t.Fatal(err)
	}
}

func TestRepository_RoundTripAndDetachedValues(t *testing.T) {
	r, _ := memoryRepository(t)
	ctx := context.Background()
	task := taskFixture("root")
	due := task.CreatedAt.Add(time.Hour)
	task.DueDate = &due
	createFixture(t, r, task)
	got, err := r.GetByID(ctx, task.ID)
	if err != nil || !reflect.DeepEqual(got, task) {
		t.Fatalf("round trip=%+v err=%v", got, err)
	}
	got.Tags[0] = "changed"
	got.Title = "changed"
	*got.DueDate = got.CreatedAt
	again, err := r.GetByID(ctx, task.ID)
	if err != nil || !reflect.DeepEqual(again, task) {
		t.Fatal("returned values alias storage")
	}
	child := taskFixture("child")
	child.ParentID = &task.ID
	child.Tags = nil
	child.Description = ""
	createFixture(t, r, child)
	kids, err := r.ListChildren(ctx, task.ID)
	if err != nil || len(kids) != 1 || kids[0].Tags == nil || kids[0].Description != "" {
		t.Fatalf("children=%v err=%v", kids, err)
	}
	anc, err := r.GetAncestors(ctx, child.ID)
	if err != nil || len(anc) != 1 || anc[0].ID != task.ID {
		t.Fatalf("ancestors=%v err=%v", anc, err)
	}
	sub, err := r.GetSubtree(ctx, child.ID)
	if err != nil || len(sub) != 1 || sub[0].ParentID == nil || *sub[0].ParentID != task.ID {
		t.Fatalf("subtree=%v err=%v", sub, err)
	}
	if err := r.WithWrite(ctx, func(c context.Context, w ports.TaskWriter) error {
		_, err := w.AppendEvent(c, ports.TaskEvent{TaskID: task.ID, Kind: ports.EventMetadata, ChangedFields: []string{"title"}, OccurredAt: task.CreatedAt})
		return err
	}); err != nil {
		t.Fatal(err)
	}
	events, err := r.ListEvents(ctx, task.ID)
	if err != nil || len(events) != 1 {
		t.Fatal(err)
	}
	events[0].ChangedFields[0] = "description"
	events, err = r.ListEvents(ctx, task.ID)
	if err != nil || events[0].ChangedFields[0] != "title" {
		t.Fatal("event alias")
	}
}

func TestWithWrite_ChildAndHistoryRollback(t *testing.T) {
	r, _ := memoryRepository(t)
	task := taskFixture("root")
	createFixture(t, r, task)
	err := r.WithWrite(context.Background(), func(ctx context.Context, w ports.TaskWriter) error {
		changed := task.Clone()
		changed.Title = "must rollback"
		if err := w.Update(ctx, &changed); err != nil {
			return err
		}
		_, _ = w.AppendEvent(ctx, ports.TaskEvent{TaskID: task.ID, Kind: "invalid", ChangedFields: []string{"title"}, OccurredAt: task.CreatedAt})
		return nil // A swallowed operation error must still roll back.
	})
	if !errors.Is(err, ports.ErrInvalidRecord) {
		t.Fatalf("swallowed error=%v", err)
	}
	got, err := r.GetByID(context.Background(), task.ID)
	if err != nil || got.Title != task.Title {
		t.Fatal("partial update committed")
	}
}

func TestTransaction_CallbackLifetime(t *testing.T) {
	r, _ := memoryRepository(t)
	ctx := context.Background()
	if err := r.WithRead(ctx, nil); !errors.Is(err, ports.ErrInvalidCallback) {
		t.Fatal(err)
	}
	if err := r.WithWrite(ctx, nil); !errors.Is(err, ports.ErrInvalidCallback) {
		t.Fatal(err)
	}
	var escaped ports.TaskWriter
	if err := r.WithWrite(ctx, func(c context.Context, w ports.TaskWriter) error {
		escaped = w
		_, err := r.List(c, core.TaskFilter{})
		if !errors.Is(err, ports.ErrNestedTransaction) {
			t.Errorf("nested=%v", err)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := escaped.List(ctx, core.TaskFilter{}); !errors.Is(err, ports.ErrTransactionClosed) {
		t.Fatal(err)
	}
	func() {
		defer func() {
			if recover() != "original panic" {
				t.Error("panic not preserved")
			}
		}()
		_ = r.WithWrite(ctx, func(c context.Context, w ports.TaskWriter) error {
			if err := w.Create(c, taskFixture("panic")); err != nil {
				t.Fatal(err)
			}
			panic("original panic")
		})
	}()
	if _, err := r.GetByID(ctx, "panic"); !errors.Is(err, core.ErrTaskNotFound) {
		t.Fatalf("panic did not roll back: %v", err)
	}
	if err := r.WithRead(ctx, func(_ context.Context, reader ports.TaskReader) error {
		if _, ok := reader.(ports.TaskWriter); ok {
			t.Fatal("read callback exposes writer methods")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
