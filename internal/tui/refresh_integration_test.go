package tui

import (
	"context"
	"crypto/rand"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/tusk/internal/ports"
	"github.com/newbpydev/tusk/internal/service"
	"github.com/newbpydev/tusk/internal/storage"
)

func TestRefresh_TwoDiskOwnersPreserveSelectionAndSearch(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "tasks.db")
	factory := func(ctx context.Context) (ports.TaskService, func() error, error) {
		repo, err := storage.Open(ctx, storage.Options{Path: path})
		if err != nil {
			return nil, nil, err
		}
		svc, err := service.NewTaskService(repo, service.Options{Clock: time.Now, NewID: func(now time.Time) (string, error) { return service.NewUUIDv7(now, rand.Reader) }, Location: time.UTC})
		if err != nil {
			repo.Close()
			return nil, nil, err
		}
		return svc, repo.Close, nil
	}
	external, closeExternal, err := factory(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer closeExternal()
	parent, err := external.CreateTask(ctx, ports.CreateTaskCommand{Title: "Parent"})
	if err != nil {
		t.Fatal(err)
	}
	child, err := external.CreateTask(ctx, ports.CreateTaskCommand{Title: "Needle child", ParentID: &parent.ID})
	if err != nil {
		t.Fatal(err)
	}
	session := NewSession(ctx, factory)
	defer session.Close(nil)
	o := testOptions()
	o.Load = session.Load
	m := New(o)
	m.Update(m.Init()())
	press(m, "/")
	press(m, "needle")
	press(m, "enter")
	press(m, "end")
	if m.selectedTask().ID != child.ID || len(m.rows) != 2 {
		t.Fatal("real forest projection lost path")
	}
	title := "Needle changed externally"
	_, err = external.UpdateTask(ctx, ports.UpdateTaskCommand{ID: child.ID, Title: &title, ClearParent: true, Base: child})
	if err != nil {
		t.Fatal(err)
	}
	m.Update(m.requestRefresh()())
	if len(m.rows) != 1 || m.selectedTask().ID != child.ID || m.selectedTask().ParentID != nil || !strings.Contains(m.View(), title) {
		t.Fatal("external move replaced selection or left stale hierarchy")
	}
	preview, err := external.PreviewDeleteTask(ctx, child.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = external.DeleteTask(ctx, ports.DeleteTaskCommand{ID: child.ID, Expected: &preview}); err != nil {
		t.Fatal(err)
	}
	m.Update(m.requestRefresh()())
	if m.selectedTask() != nil || !strings.Contains(m.View(), "No matching tasks") {
		t.Fatal("deleted row survived refresh")
	}
	press(m, "esc")
	if m.selectedTask().ID != parent.ID {
		t.Fatal("clear failed to restore surviving forest")
	}
}
