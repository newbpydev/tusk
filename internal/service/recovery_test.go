package service

import (
	"context"
	"errors"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"github.com/newbpydev/tusk/internal/storage"
	"testing"
)

type unknownRepository struct {
	ports.TaskRepository
	before bool
	calls  *int
}

func (r unknownRepository) WithWrite(ctx context.Context, fn func(context.Context, ports.TaskWriter) error) error {
	injected := false
	e := r.TaskRepository.WithWrite(ctx, func(ctx context.Context, w ports.TaskWriter) error {
		*r.calls++
		if e := fn(ctx, w); e != nil {
			return e
		}
		if r.before {
			injected = true
			return ports.ErrConflict
		}
		return nil
	})
	if e != nil && !injected {
		return e
	}
	return ports.NewTransactionError("acknowledgment", errors.Join(ports.ErrConflict, context.Canceled))
}
func TestUnknownRepository_CallbackFailure(t *testing.T) {
	r := repository(t)
	calls := 0
	s := serviceFor(t, unknownRepository{TaskRepository: r, before: true, calls: &calls}, false)
	got, err := s.CompleteTask(context.Background(), ports.TaskCommand{ID: "missing"})
	var outcome ports.TransactionError
	if got != nil || !errors.Is(err, core.ErrTaskNotFound) || errors.As(err, &outcome) || calls != 1 {
		t.Fatalf("callback error changed to unknown outcome: %v, %v, calls=%d", got, err, calls)
	}
}

func TestService_RollbackAndUnknownOutcome(t *testing.T) {
	for _, before := range []bool{false, true} {
		path, r, _ := twoOwners(t, task("a", "", 0, core.StatusTodo))
		calls := 0
		s := serviceFor(t, unknownRepository{TaskRepository: r, before: before, calls: &calls}, false)
		v, e := s.CompleteTask(context.Background(), ports.TaskCommand{ID: "a"})
		var outcome ports.TransactionError
		if v != nil || !errors.As(e, &outcome) || !errors.Is(e, ports.ErrConflict) || !errors.Is(e, context.Canceled) || calls != 1 {
			t.Fatalf("%v %v calls=%d", v, e, calls)
		}
		if e = r.Close(); e != nil {
			t.Fatal(e)
		}
		fresh, e := storage.Open(context.Background(), storage.Options{Path: path})
		if e != nil {
			t.Fatal(e)
		}
		got := mustGet(t, fresh, "a")
		ev, e := fresh.ListEvents(context.Background(), "a")
		if e != nil {
			t.Fatal(e)
		}
		if (got.Status == core.StatusDone) == before || len(ev) != map[bool]int{true: 0, false: 1}[before] {
			t.Fatalf("%+v %v", got, ev)
		}
		if e = fresh.Close(); e != nil {
			t.Fatal(e)
		}
	}
}
func TestService_CommittedThenCanceled(t *testing.T) {
	for _, deletion := range []bool{false, true} {
		path, r, _ := twoOwners(t, task("a", "", 0, core.StatusTodo))
		ctx, cancel := context.WithCancel(context.Background())
		s := serviceFor(t, faultRepository{TaskRepository: r, cancel: cancel}, false)
		if deletion {
			v, e := s.DeleteTask(ctx, ports.DeleteTaskCommand{ID: "a", Force: true})
			if e != nil || !v.Deleted {
				t.Fatalf("%v %v", v, e)
			}
		} else {
			v, e := s.CompleteTask(ctx, ports.TaskCommand{ID: "a"})
			if e != nil || v == nil || v.Status != core.StatusDone {
				t.Fatalf("%v %v", v, e)
			}
		}
		if _, e := s.GetTask(ctx, "a"); !errors.Is(e, context.Canceled) {
			t.Fatal(e)
		}
		if err := r.Close(); err != nil {
			t.Fatal(err)
		}
		fresh, err := storage.Open(context.Background(), storage.Options{Path: path})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := fresh.Close(); err != nil {
				t.Error(err)
			}
		})
		got, err := fresh.GetByID(context.Background(), "a")
		if deletion {
			if got != nil || !errors.Is(err, core.ErrTaskNotFound) {
				t.Fatalf("deletion was not durable: %v, %v", got, err)
			}
		} else {
			if err != nil || got == nil || got.Status != core.StatusDone {
				t.Fatalf("completion was not durable: %v, %v", got, err)
			}
			events, err := fresh.ListEvents(context.Background(), "a")
			if err != nil || len(events) != 1 || events[0].Kind != ports.EventStatus {
				t.Fatalf("completion history was not durable: %v, %v", events, err)
			}
		}
	}
}

type pausingRepository struct {
	ports.TaskRepository
	ready, release chan struct{}
}
type pausingReader struct {
	ports.TaskReader
	ready, release chan struct{}
}

func (r pausingRepository) WithRead(ctx context.Context, fn func(context.Context, ports.TaskReader) error) error {
	return r.TaskRepository.WithRead(ctx, func(ctx context.Context, reader ports.TaskReader) error {
		return fn(ctx, pausingReader{TaskReader: reader, ready: r.ready, release: r.release})
	})
}
func (r pausingReader) List(ctx context.Context, f core.TaskFilter) ([]core.Task, error) {
	rows, e := r.TaskReader.List(ctx, f)
	close(r.ready)
	<-r.release
	return rows, e
}
func TestService_ReadSnapshot(t *testing.T) {
	_, a, b := twoOwners(t, task("p", "", 0, core.StatusTodo), task("a", "p", 0, core.StatusTodo))
	ready, release := make(chan struct{}), make(chan struct{})
	sa := serviceFor(t, pausingRepository{TaskRepository: a, ready: ready, release: release}, false)
	sb := serviceFor(t, b, false)
	out := make(chan ports.TaskStats, 1)
	errs := make(chan error, 1)
	go func() { v, e := sa.GetStats(context.Background()); out <- v; errs <- e }()
	<-ready
	_, e := sb.CompleteTask(context.Background(), ports.TaskCommand{ID: "p"})
	close(release)
	if e != nil {
		t.Fatal(e)
	}
	stats := <-out
	if e = <-errs; e != nil || stats.Done != 0 || stats.Total != 2 {
		t.Fatalf("%v %v", stats, e)
	}
	fresh, e := sb.GetStats(context.Background())
	if e != nil || fresh.Done != 2 {
		t.Fatalf("%v %v", fresh, e)
	}
}
