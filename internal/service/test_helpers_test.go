package service

import (
	"context"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"github.com/newbpydev/tusk/internal/storage"
	"path/filepath"
	"testing"
	"time"
)

func fixedTime() time.Time { return time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC) }
func task(id, parent string, progress int, status core.Status) core.Task {
	now := fixedTime()
	v, _ := core.NewTask(core.NewTaskParams{ID: id, Title: id, Now: now})
	v.Status = status
	v.Progress = progress
	if parent != "" {
		v.ParentID = &parent
	}
	if status == core.StatusDone {
		v.Progress = 100
		v.CompletedAt = &now
	}
	return *v
}
func repository(t *testing.T, tasks ...core.Task) *storage.Repository {
	t.Helper()
	r, e := storage.Open(context.Background(), storage.Options{Path: filepath.Join(t.TempDir(), "tasks.db")})
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() {
		if e := r.Close(); e != nil {
			t.Error(e)
		}
	})
	e = r.WithWrite(context.Background(), func(ctx context.Context, w ports.TaskWriter) error {
		for _, v := range tasks {
			if e := w.Create(ctx, &v); e != nil {
				return e
			}
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func mustGet(t *testing.T, r ports.TaskReader, id string) *core.Task {
	t.Helper()
	v, e := r.GetByID(context.Background(), id)
	if e != nil {
		t.Fatal(e)
	}
	return v
}
