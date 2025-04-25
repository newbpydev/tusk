package task

import (
	"context"
	"time"

	"github.com/newbpydev/tusk/internal/core/task"
)

// Service is the interface that defines the methods for managing tasks.
// It includes methods for creating, showing, listing, reordering, updating, deleting,
// completing, changing status, and changing priority of tasks.
// Each method takes a context and relevant parameters, and returns the task or an error.
type Service interface {
	Create(ctx context.Context, userID int64, parentID *int64, title, description string,
		dueDate *time.Time, priority task.Priority, tags []string) (task.Task, error)
	Show(ctx context.Context, taskID int64) (task.Task, error)
	List(ctx context.Context, userID int64) ([]task.Task, error)
	Reorder(ctx context.Context, taskID int64, newOrder int) error
	Update(ctx context.Context, taskID int64, title, description string,
		dueDate *time.Time, priority task.Priority, tags []string) (task.Task, error)
	Delete(ctx context.Context, taskID int64) error
	Complete(ctx context.Context, taskID int64) (task.Task, error)
	ChangeStatus(ctx context.Context, taskID int64, status task.Status) (task.Task, error)
	ChangePriority(ctx context.Context, taskID int64, priority task.Priority) (task.Task, error)
}
