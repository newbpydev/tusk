package service

import (
	"context"
	"errors"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"github.com/newbpydev/tusk/internal/storage"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func twoOwners(t *testing.T, rows ...core.Task) (string, *storage.Repository, *storage.Repository) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "wal.db")
	open := func() *storage.Repository {
		r, e := storage.Open(context.Background(), storage.Options{Path: path})
		if e != nil {
			t.Fatal(e)
		}
		t.Cleanup(func() {
			if e := r.Close(); e != nil {
				t.Error(e)
			}
		})
		return r
	}
	a := open()
	e := a.WithWrite(context.Background(), func(ctx context.Context, w ports.TaskWriter) error {
		for _, v := range rows {
			if e := w.Create(ctx, &v); e != nil {
				return e
			}
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	return path, a, open()
}
func together(a, b func() error) []error {
	start := make(chan struct{})
	out := make(chan error, 2)
	var ready sync.WaitGroup
	ready.Add(2)
	for _, fn := range []func() error{a, b} {
		go func(fn func() error) { ready.Done(); <-start; out <- fn() }(fn)
	}
	ready.Wait()
	close(start)
	return []error{<-out, <-out}
}
func TestService_ConcurrentSiblingCompletion(t *testing.T) {
	_, a, b := twoOwners(t, task("g", "", 0, core.StatusTodo), task("p", "g", 0, core.StatusTodo), task("a", "p", 0, core.StatusTodo), task("b", "p", 0, core.StatusTodo))
	sa, sb := serviceFor(t, a, true), serviceFor(t, b, true)
	for _, e := range together(func() error { _, e := sa.CompleteTask(context.Background(), ports.TaskCommand{ID: "a"}); return e }, func() error { _, e := sb.CompleteTask(context.Background(), ports.TaskCommand{ID: "b"}); return e }) {
		if e != nil {
			t.Fatal(e)
		}
	}
	for _, id := range []string{"g", "p", "a", "b"} {
		v := mustGet(t, a, id)
		if v.Status != core.StatusDone || v.Progress != 100 {
			t.Fatalf("%+v", v)
		}
	}
	for _, id := range []string{"a", "b"} {
		ev, e := sa.GetTaskHistory(context.Background(), id)
		if e != nil || len(ev) != 1 || ev[0].Kind != ports.EventStatus {
			t.Fatalf("%v %v", ev, e)
		}
	}
}
func TestService_OppositeMoves(t *testing.T) {
	_, a, b := twoOwners(t, task("a", "", 0, core.StatusTodo), task("b", "", 0, core.StatusTodo))
	sa, sb := serviceFor(t, a, false), serviceFor(t, b, false)
	success, cycles := 0, 0
	for _, e := range together(func() error {
		_, e := sa.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: "a", ParentID: ptr("b")})
		return e
	}, func() error {
		_, e := sb.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: "b", ParentID: ptr("a")})
		return e
	}) {
		if e == nil {
			success++
		} else if errors.Is(e, core.ErrCyclicDependency) {
			cycles++
		} else {
			t.Fatal(e)
		}
	}
	if success != 1 || cycles != 1 {
		t.Fatalf("%d %d", success, cycles)
	}
	forest, e := sa.GetTaskTree(context.Background(), "")
	if e != nil || len(forest) != 1 {
		t.Fatalf("%v %v", forest, e)
	}
	count := 0
	for _, id := range []string{"a", "b"} {
		ev, _ := sa.GetTaskHistory(context.Background(), id)
		count += len(ev)
	}
	if count != 1 {
		t.Fatalf("events %d", count)
	}
}
func TestService_DescendantHeightRace(t *testing.T) {
	rows := []core.Task{}
	parent := ""
	for i := 0; i < 9; i++ {
		id := string(rune('a' + i))
		rows = append(rows, task(id, parent, 0, core.StatusTodo))
		parent = id
	}
	rows = append(rows, task("t", "", 0, core.StatusTodo), task("x", "", 0, core.StatusTodo))
	_, a, b := twoOwners(t, rows...)
	sa, sb := serviceFor(t, a, false), serviceFor(t, b, false)
	success, depth := 0, 0
	for _, e := range together(func() error {
		_, e := sa.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: "t", ParentID: ptr("i")})
		return e
	}, func() error {
		_, e := sb.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: "x", ParentID: ptr("t")})
		return e
	}) {
		if e == nil {
			success++
		} else if errors.Is(e, core.ErrMaxDepthExceeded) {
			depth++
		} else {
			t.Fatal(e)
		}
	}
	if success != 1 || depth != 1 {
		t.Fatalf("%d %d", success, depth)
	}
	if _, e := sa.GetTaskTree(context.Background(), ""); e != nil {
		t.Fatal(e)
	}
}
func TestService_StalePatchAndDeletePreview(t *testing.T) {
	_, a, b := twoOwners(t, task("p", "", 0, core.StatusTodo), task("a", "p", 0, core.StatusTodo))
	sa, sb := serviceFor(t, a, false), serviceFor(t, b, false)
	base, _ := sa.GetTask(context.Background(), "a")
	preview, _ := sa.PreviewDeleteTask(context.Background(), "p")
	if _, e := sb.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: "a", Description: ptr("newer")}); e != nil {
		t.Fatal(e)
	}
	if v, e := sa.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: "a", Title: ptr("title"), Base: base}); !errors.Is(e, ports.ErrConflict) || v != nil {
		t.Fatalf("%v %v", v, e)
	}
	v, e := sa.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: "a", Title: ptr("title")})
	if e != nil || v.Description != "newer" {
		t.Fatalf("%v %v", v, e)
	}
	if _, e := sb.CreateTask(context.Background(), ports.CreateTaskCommand{Title: "child", ParentID: ptr("p")}); e != nil {
		t.Fatal(e)
	}
	if v, e := sa.DeleteTask(context.Background(), ports.DeleteTaskCommand{ID: "p", Expected: &preview, Recursive: true}); !errors.Is(e, ports.ErrConflict) || v.Deleted {
		t.Fatalf("%v %v", v, e)
	}
	if _, e := sa.DeleteTask(context.Background(), ports.DeleteTaskCommand{ID: "p", Force: true}); !errors.Is(e, ports.ErrChildrenPresent) {
		t.Fatal(e)
	}
}
func TestService_WriterAdmission(t *testing.T) {
	if testing.Short() {
		t.Skip("five-second real contention budget")
	}
	_, a, b := twoOwners(t, task("a", "", 0, core.StatusTodo))
	ready, release := make(chan struct{}), make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- a.WithWrite(context.Background(), func(context.Context, ports.TaskWriter) error { close(ready); <-release; return nil })
	}()
	<-ready
	s := serviceFor(t, b, false)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if _, e := s.CompleteTask(ctx, ports.TaskCommand{ID: "a"}); !errors.Is(e, context.DeadlineExceeded) {
		close(release)
		<-done
		t.Fatal(e)
	}
	_, e := s.CompleteTask(context.Background(), ports.TaskCommand{ID: "a"})
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if !errors.Is(e, ports.ErrBusy) {
		t.Fatal(e)
	}
	if ev, _ := s.GetTaskHistory(context.Background(), "a"); len(ev) != 0 {
		t.Fatal("rejected admission created event")
	}
	if _, e := s.CompleteTask(context.Background(), ports.TaskCommand{ID: "a"}); e != nil {
		t.Fatal(e)
	}
}
