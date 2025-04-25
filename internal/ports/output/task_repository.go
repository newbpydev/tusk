package output

import (
	"context"

	"github.com/newbpydev/tusk/internal/core/task"
)

// TaskStatusCounts holds counts of tasks by status
type TaskStatusCounts struct {
	TodoCount       int `json:"todo_count"`
	InProgressCount int `json:"in_progress_count"`
	DoneCount       int `json:"done_count"`
	TotalCount      int `json:"total_count"`
}

// TaskPriorityCounts holds counts of tasks by priority
type TaskPriorityCounts struct {
	LowCount    int `json:"low_count"`
	MediumCount int `json:"medium_count"`
	HighCount   int `json:"high_count"`
}

// TaskRepository is an interface that defines the methods for interacting with tasks in the database.
// It provides methods for creating, retrieving, updating, and deleting tasks.
type TaskRepository interface {
	// Create creates a new task in the database.
	// It returns the created task or an error if the task could not be created.
	// parentID = nil means root task
	Create(ctx context.Context, t task.Task) (task.Task, error)

	// Update updates an existing task in the database.
	// It returns an error if the task could not be updated.
	Update(ctx context.Context, t task.Task) error

	// Delete deletes a task from the database.
	// It returns an error if the task could not be deleted.
	Delete(ctx context.Context, id int64) error

	// GetByID retrieves a task by its ID from the database.
	// It returns the task or an error if the task could not be found.
	GetByID(ctx context.Context, id int64) (task.Task, error)

	// ListRootTasks retrieves all root tasks for a user from the database.
	// It returns a list of tasks or an error if the tasks could not be retrieved.
	ListRootTasks(ctx context.Context, userID int64) ([]task.Task, error)

	// ListSubTasks retrieves all subtasks for a task from the database.
	// It returns a list of tasks or an error if the tasks could not be retrieved.
	ListSubTasks(ctx context.Context, parentID int64) ([]task.Task, error)

	// GetTaskTree retrieves a task and its subtasks from the database.
	// It returns the task and its subtasks or an error if the task could not be found.
	GetTaskTree(ctx context.Context, rootID int64) (task.Task, error)

	// ReorderTask reorders a task in the database.
	// It returns an error if the task could not be reordered.
	ReorderTask(ctx context.Context, taskID int64, newOrder int) error

	// Search and filtering methods

	// SearchTasksByTitle searches for tasks with titles matching the given pattern.
	// Pattern can include % for wildcard matching.
	SearchTasksByTitle(ctx context.Context, userID int64, titlePattern string) ([]task.Task, error)

	// SearchTasksByTag searches for tasks that have the specified tag.
	SearchTasksByTag(ctx context.Context, userID int64, tag string) ([]task.Task, error)

	// ListTasksByStatus retrieves all tasks for a user with the given status.
	ListTasksByStatus(ctx context.Context, userID int64, status task.Status) ([]task.Task, error)

	// ListTasksByPriority retrieves all tasks for a user with the given priority.
	ListTasksByPriority(ctx context.Context, userID int64, priority task.Priority) ([]task.Task, error)

	// Due date related methods

	// ListTasksDueToday retrieves all incomplete tasks due on the current day.
	ListTasksDueToday(ctx context.Context, userID int64) ([]task.Task, error)

	// ListTasksDueSoon retrieves all incomplete tasks due within the next 7 days.
	ListTasksDueSoon(ctx context.Context, userID int64) ([]task.Task, error)

	// ListOverdueTasks retrieves all incomplete tasks that are past their due date.
	ListOverdueTasks(ctx context.Context, userID int64) ([]task.Task, error)

	// Statistics and metrics

	// GetTaskCountsByStatus retrieves counts of tasks grouped by status.
	GetTaskCountsByStatus(ctx context.Context, userID int64) (TaskStatusCounts, error)

	// GetTaskCountsByPriority retrieves counts of incomplete tasks grouped by priority.
	GetTaskCountsByPriority(ctx context.Context, userID int64) (TaskPriorityCounts, error)

	// GetRecentlyCompletedTasks retrieves recently completed tasks, limited by count.
	GetRecentlyCompletedTasks(ctx context.Context, userID int64, limit int32) ([]task.Task, error)

	// Batch operations

	// BulkUpdateTaskStatus updates the status of multiple tasks at once.
	BulkUpdateTaskStatus(ctx context.Context, taskIDs []int32, status task.Status, isCompleted bool) error

	// Tag operations

	// GetAllTagsForUser retrieves all unique tags used by a user.
	GetAllTagsForUser(ctx context.Context, userID int64) ([]string, error)
}
