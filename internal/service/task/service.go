package task

import (
	"context"
	"time"

	"github.com/newbpydev/tusk/internal/core/task"
	"github.com/newbpydev/tusk/internal/ports/output"
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

	// Search and filtering methods

	// SearchByTitle searches for tasks with titles matching the given pattern.
	SearchByTitle(ctx context.Context, userID int64, titlePattern string) ([]task.Task, error)

	// SearchByTag searches for tasks that have the specified tag.
	SearchByTag(ctx context.Context, userID int64, tag string) ([]task.Task, error)

	// ListByStatus retrieves all tasks for a user with the given status.
	ListByStatus(ctx context.Context, userID int64, status task.Status) ([]task.Task, error)

	// ListByPriority retrieves all tasks for a user with the given priority.
	ListByPriority(ctx context.Context, userID int64, priority task.Priority) ([]task.Task, error)

	// Due date related methods

	// ListTasksDueToday retrieves all incomplete tasks due on the current day.
	ListTasksDueToday(ctx context.Context, userID int64) ([]task.Task, error)

	// ListTasksDueSoon retrieves all incomplete tasks due within the next 7 days.
	ListTasksDueSoon(ctx context.Context, userID int64) ([]task.Task, error)

	// ListOverdueTasks retrieves all incomplete tasks that are past their due date.
	ListOverdueTasks(ctx context.Context, userID int64) ([]task.Task, error)

	// Statistics and metrics

	// GetTaskCountsByStatus retrieves counts of tasks grouped by status.
	GetTaskCountsByStatus(ctx context.Context, userID int64) (output.TaskStatusCounts, error)

	// GetTaskCountsByPriority retrieves counts of incomplete tasks grouped by priority.
	GetTaskCountsByPriority(ctx context.Context, userID int64) (output.TaskPriorityCounts, error)

	// GetRecentlyCompletedTasks retrieves recently completed tasks, limited by count.
	GetRecentlyCompletedTasks(ctx context.Context, userID int64, limit int) ([]task.Task, error)

	// Batch operations

	// BulkUpdateStatus updates the status of multiple tasks at once.
	BulkUpdateStatus(ctx context.Context, taskIDs []int32, status task.Status) error

	// Tag operations

	// GetAllTags retrieves all unique tags used by a user.
	GetAllTags(ctx context.Context, userID int64) ([]string, error)
}
