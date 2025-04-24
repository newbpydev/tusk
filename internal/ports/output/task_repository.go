package output

import (
	"context"

	"github.com/newbpydev/tusk/internal/core/task"
)

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
}
