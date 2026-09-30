package tui

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

type recoveryService struct {
	ports.TaskService
	tree    func(context.Context, string) ([]*core.TaskNode, error)
	history func(context.Context, string) ([]ports.TaskEvent, error)
}

func (s recoveryService) GetTaskTree(ctx context.Context, id string) ([]*core.TaskNode, error) {
	return s.tree(ctx, id)
}
func (s recoveryService) GetTaskHistory(ctx context.Context, id string) ([]ports.TaskEvent, error) {
	return s.history(ctx, id)
}

func TestRecovery_SessionRetiresBeforeFreshTreeAndHistory(t *testing.T) {
	for _, failure := range []string{"none", "close", "open", "tree", "history"} {
		t.Run(failure, func(t *testing.T) {
			opens, closes, histories := 0, 0, 0
			s := NewSession(context.Background(), func(context.Context) (ports.TaskService, func() error, error) {
				if opens != closes {
					t.Fatal("owner overlap")
				}
				opens++
				if opens == 2 && failure == "open" {
					opens--
					return nil, nil, ports.ErrStorage
				}
				svc := recoveryService{tree: func(context.Context, string) ([]*core.TaskNode, error) {
					if opens == 2 && failure == "tree" {
						return nil, ports.ErrStorage
					}
					return []*core.TaskNode{fixtureNode("id", "Current", 2, nil)}, nil
				}, history: func(context.Context, string) ([]ports.TaskEvent, error) {
					histories++
					if failure == "history" {
						return nil, ports.ErrStorage
					}
					return []ports.TaskEvent{{TaskID: "id", Sequence: 1}}, nil
				}}
				return svc, func() error {
					closes++
					if failure == "close" {
						return ports.ErrStorage
					}
					return nil
				}, nil
			})
			_, err := s.Load(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			_, err = s.Call(context.Background(), false, func(context.Context, ports.TaskService) (any, error) {
				return nil, ports.NewTransactionError("read", ports.ErrBusy)
			})
			if !IsUnknown(err) {
				t.Fatal("fixture")
			}
			snapshot, err := s.Recover(context.Background(), "id")
			if failure == "none" {
				if err != nil || len(snapshot.forest) != 1 || len(snapshot.history) != 1 || histories != 1 {
					t.Fatal(snapshot, err)
				}
			} else if err == nil {
				t.Fatal("recovery failure ignored")
			}
			if failure == "close" {
				if !errors.Is(err, errRetireFailed) {
					t.Fatal("retirement not classified")
				}
				_, err = s.Recover(context.Background(), "id")
				if err == nil || opens != 1 {
					t.Fatal("reopened after failed close")
				}
			}
			result, _ := s.Close(nil)
			if !result.OutcomeUnknown || opens != closes {
				t.Fatal("lost outcome or owner", result, opens, closes)
			}
		})
	}
}

type emptyDeleteService struct{ sessionService }

func (emptyDeleteService) DeleteTask(context.Context, ports.DeleteTaskCommand) (ports.DeleteResult, error) {
	return ports.DeleteResult{}, nil
}
func TestDelete_InvalidSuccessCannotLookCommitted(t *testing.T) {
	s := NewSession(context.Background(), func(context.Context) (ports.TaskService, func() error, error) {
		return emptyDeleteService{}, func() error { return nil }, nil
	})
	_, err := s.Delete(context.Background(), ports.DeleteTaskCommand{ID: "id"})
	if !IsUnknown(err) {
		t.Fatal("invalid delete result accepted")
	}
	result, _ := s.Close(nil)
	if !result.OutcomeUnknown || result.HadCommittedChanges {
		t.Fatal(result)
	}
}

func TestRecovery_CloseFailureOnlyAllowsQuit(t *testing.T) {
	m := loadedModel()
	m.freezeWrites()
	calls := 0
	m.options.Recover = func(context.Context, string) (recoverySnapshot, error) {
		calls++
		return recoverySnapshot{}, errors.Join(errRetireFailed, ports.ErrStorage)
	}
	deliverUI(m, press(m, "r"))
	press(m, "r")
	press(m, "enter")
	if calls != 1 || !m.recovery.blocked || !m.recoveryNeeded {
		t.Fatal("reopened after retirement failure")
	}
	footer := strings.Split(m.View(), "\n")[m.height-1]
	if strings.Contains(footer, "r reload") || !strings.Contains(footer, "q quit") {
		t.Fatal("unavailable reload is still offered", footer)
	}
}

func TestRecovery_CancelDrainsBeforeClosingFreshOwner(t *testing.T) {
	var opened, closed atomic.Int32
	started, release := make(chan struct{}), make(chan struct{})
	s := NewSession(context.Background(), func(context.Context) (ports.TaskService, func() error, error) {
		opened.Add(1)
		svc := recoveryService{tree: func(context.Context, string) ([]*core.TaskNode, error) {
			return []*core.TaskNode{fixtureNode("id", "Task", 2, nil)}, nil
		}, history: func(ctx context.Context, _ string) ([]ports.TaskEvent, error) {
			close(started)
			<-ctx.Done()
			<-release
			return nil, ctx.Err()
		}}
		return svc, func() error { closed.Add(1); return nil }, nil
	})
	if _, err := s.Load(context.Background()); err != nil {
		t.Fatal(err)
	}
	s.Call(context.Background(), false, func(context.Context, ports.TaskService) (any, error) {
		return nil, ports.NewTransactionError("read", ports.ErrBusy)
	})
	recovered := make(chan error, 1)
	go func() { _, err := s.Recover(context.Background(), "id"); recovered <- err }()
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("recovery deadlocked on its own admission")
	}
	finished := make(chan Result, 1)
	go func() { result, _ := s.Close(nil); finished <- result }()
	if closed.Load() != 1 {
		t.Fatal("fresh owner closed under active history")
	}
	close(release)
	select {
	case result := <-finished:
		if !result.OutcomeUnknown {
			t.Fatal("cancellation erased uncertainty")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("shutdown did not drain")
	}
	if err := <-recovered; !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if opened.Load() != 2 || closed.Load() != 2 {
		t.Fatal("owner leak")
	}
}
