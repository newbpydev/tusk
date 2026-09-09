package ports

import (
	"context"
	"errors"
	"fmt"
	"github.com/newbpydev/tusk/internal/core"
	"testing"
)

func TestServiceErrorCatalog(t *testing.T) {
	seen := map[Error]bool{}
	for _, e := range []Error{ErrInvalidCommand, ErrInvalidText, ErrInvalidDate, ErrInvalidReferenceTime, ErrIdentityGeneration, ErrConflict, ErrConfirmationRequired, ErrInvalidServiceOptions} {
		if seen[e] || e.Error() == "" || !errors.Is(fmt.Errorf("operation: %w", e), e) {
			t.Fatalf("invalid category: %v", e)
		}
		seen[e] = true
	}
}

// This consumer fixture compiles the complete port independently of service or SQL.
func serviceConsumer(ctx context.Context, s TaskService) error {
	task, err := s.CreateTask(ctx, CreateTaskCommand{Title: "task"})
	if err != nil {
		return err
	}
	_, _ = s.UpdateTask(ctx, UpdateTaskCommand{ID: task.ID, ClearDue: true, Base: task})
	_, _ = s.CompleteTask(ctx, TaskCommand{ID: task.ID, Base: task})
	_, _ = s.ReopenTask(ctx, ReopenTaskCommand{ID: task.ID, Status: core.StatusTodo})
	p, _ := s.PreviewDeleteTask(ctx, task.ID)
	_, _ = s.DeleteTask(ctx, DeleteTaskCommand{ID: task.ID, Expected: &p})
	_, _ = s.GetTask(ctx, task.ID)
	_, _ = s.ListTasks(ctx, TaskQuery{All: true})
	_, _ = s.GetTaskTree(ctx, "")
	_, _ = s.GetStats(ctx)
	_, _ = s.GetTaskHistory(ctx, task.ID)
	return nil
}
