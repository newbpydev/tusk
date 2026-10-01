package tui

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

type sessionService struct{ ports.TaskService }

func (sessionService) GetTaskTree(context.Context, string) ([]*core.TaskNode, error) { return nil, nil }

func TestSession_CancelDrainsBeforeClose(t *testing.T) {
	var closed atomic.Int32
	s := NewSession(context.Background(), func(context.Context) (ports.TaskService, func() error, error) {
		return sessionService{}, func() error { closed.Add(1); return nil }, nil
	})
	started, release, finished := make(chan struct{}), make(chan struct{}), make(chan struct{})
	go func() {
		defer close(finished)
		_, err := s.Call(context.Background(), false, func(ctx context.Context, _ ports.TaskService) (any, error) {
			close(started)
			<-ctx.Done()
			if closed.Load() != 0 {
				t.Error("close while active")
			}
			<-release
			return nil, ctx.Err()
		})
		if !errors.Is(err, context.Canceled) {
			t.Error(err)
		}
	}()
	<-started
	cleanup := make(chan struct{})
	go func() { s.Close(nil); close(cleanup) }()
	select {
	case <-cleanup:
		t.Fatal("did not drain")
	default:
	}
	close(release)
	<-finished
	<-cleanup
	if closed.Load() != 1 {
		t.Fatal(closed.Load())
	}
	_, err := s.Call(context.Background(), false, func(context.Context, ports.TaskService) (any, error) { t.Fatal("late service access"); return nil, nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	s.Close(nil)
	if closed.Load() != 1 {
		t.Fatal("double close")
	}
}

func TestSession_CommitBeforeDroppedMessage(t *testing.T) {
	s := NewSession(context.Background(), func(context.Context) (ports.TaskService, func() error, error) {
		return sessionService{}, func() error { return nil }, nil
	})
	_, err := s.Call(context.Background(), true, func(context.Context, ports.TaskService) (any, error) { return &core.Task{ID: "saved"}, nil })
	if err != nil {
		t.Fatal(err)
	}
	unknown := ports.NewTransactionError("private", ports.ErrBusy)
	_, err = s.Call(context.Background(), true, func(context.Context, ports.TaskService) (any, error) {
		return nil, errors.Join(ports.ErrConflict, &unknown)
	})
	if !IsUnknown(err) {
		t.Fatal(err)
	}
	result, _ := s.Close(nil)
	if !result.HadCommittedChanges || !result.OutcomeUnknown {
		t.Fatalf("lost receipt %+v", result)
	}
}

func TestSession_Reopen(t *testing.T) {
	for _, failClose := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "close failure"}[failClose], func(t *testing.T) {
			opened, closed := 0, 0
			s := NewSession(context.Background(), func(context.Context) (ports.TaskService, func() error, error) {
				if opened != closed {
					t.Error("reopened before close")
				}
				opened++
				return sessionService{}, func() error {
					closed++
					if failClose {
						return ports.ErrStorage
					}
					return nil
				}, nil
			})
			_, err := s.Load(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			_, err = s.Reopen(context.Background())
			if failClose && (err == nil || opened != 1) {
				t.Fatalf("%d %v", opened, err)
			}
			if failClose {
				if _, err = s.Reopen(context.Background()); err == nil || opened != 1 {
					t.Fatal("reopened after failed retirement")
				}
			}
			if !failClose && (err != nil || opened != 2) {
				t.Fatal(err)
			}
			s.Close(nil)
			if closed != opened {
				t.Fatal("owner leaked")
			}
		})
	}
}

func TestSession_LateCommandDoesNotOpen(t *testing.T) {
	s := NewSession(context.Background(), func(context.Context) (ports.TaskService, func() error, error) {
		t.Fatal("late open")
		return nil, nil, nil
	})
	s.Close(nil)
	if _, err := s.Load(context.Background()); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestSession_AdmissionAndCleanupProgress(t *testing.T) {
	for range 25 {
		s := NewSession(context.Background(), func(context.Context) (ports.TaskService, func() error, error) {
			return sessionService{}, func() error { return nil }, nil
		})
		s.cleanupDelay = 0
		started, release, finished, progress := make(chan struct{}), make(chan struct{}), make(chan struct{}), make(chan struct{})
		go func() {
			defer close(finished)
			_, _ = s.Call(context.Background(), false, func(context.Context, ports.TaskService) (any, error) { close(started); <-release; return nil, nil })
		}()
		<-started
		if _, err := s.Load(context.Background()); !errors.Is(err, ports.ErrTransactionInUse) {
			t.Fatal(err)
		}
		cleanup := make(chan struct{})
		go func() { defer close(cleanup); s.Close(func() { close(progress) }) }()
		<-progress
		if _, err := s.Load(context.Background()); !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
		close(release)
		<-finished
		<-cleanup
	}
}

func TestSession_UnknownReadAndInvalidReceipt(t *testing.T) {
	s := NewSession(context.Background(), func(context.Context) (ports.TaskService, func() error, error) {
		return sessionService{}, func() error { return ports.NewTransactionError("close", ports.ErrStorage) }, nil
	})
	_, err := s.Call(context.Background(), true, func(context.Context, ports.TaskService) (any, error) { return nil, nil })
	if !IsUnknown(err) {
		t.Fatal("nil receipt accepted")
	}
	if _, err = s.Load(context.Background()); !errors.Is(err, ports.ErrReadOnly) {
		t.Fatal("unknown owner reused")
	}
	r, err := s.Close(nil)
	if !IsUnknown(err) || !r.OutcomeUnknown || r.HadCommittedChanges {
		t.Fatalf("lost outcome: %+v %v", r, err)
	}
}

func TestSession_FactoryAndPanic(t *testing.T) {
	for _, tc := range []struct {
		name string
		open func(context.Context) (ports.TaskService, func() error, error)
	}{
		{"error", func(context.Context) (ports.TaskService, func() error, error) { return nil, nil, ports.ErrStorage }},
		{"nil", func(context.Context) (ports.TaskService, func() error, error) { return nil, nil, nil }},
		{"nil service", func(context.Context) (ports.TaskService, func() error, error) {
			return nil, func() error { return nil }, nil
		}},
		{"panic", func(context.Context) (ports.TaskService, func() error, error) { panic("PRIVATE") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := NewSession(context.Background(), tc.open)
			if _, err := s.Load(context.Background()); err == nil {
				t.Fatal("accepted invalid factory")
			}
			s.Close(nil)
		})
	}
	s := NewSession(context.Background(), func(context.Context) (ports.TaskService, func() error, error) {
		return sessionService{}, func() error { panic("PRIVATE") }, nil
	})
	_, err := s.Call(context.Background(), true, func(ctx context.Context, _ ports.TaskService) (any, error) {
		deadline, ok := ctx.Deadline()
		if !ok || time.Until(deadline) > 10*time.Second {
			t.Error("no bounded deadline")
		}
		panic("PRIVATE")
	})
	if !IsUnknown(err) {
		t.Fatal("panic lost unknown mutation")
	}
	_, err = s.Close(nil)
	if err == nil {
		t.Fatal("close panic lost")
	}
}
