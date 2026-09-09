package ports

import (
	"context"
	"github.com/newbpydev/tusk/internal/core"
)

// TaskService is the synchronous application boundary shared by all adapters.
// Results are detached and published only after transaction success. An error
// returns no partial result; unknown outcomes require caller-owned readback.
type TaskService interface {
	CreateTask(context.Context, CreateTaskCommand) (*core.Task, error)
	UpdateTask(context.Context, UpdateTaskCommand) (*core.Task, error)
	CompleteTask(context.Context, TaskCommand) (*core.Task, error)
	ReopenTask(context.Context, ReopenTaskCommand) (*core.Task, error)
	PreviewDeleteTask(context.Context, string) (DeletePreview, error)
	DeleteTask(context.Context, DeleteTaskCommand) (DeleteResult, error)
	GetTask(context.Context, string) (*core.Task, error)
	ListTasks(context.Context, TaskQuery) ([]core.Task, error)
	GetTaskTree(context.Context, string) ([]*core.TaskNode, error)
	GetStats(context.Context) (TaskStats, error)
	GetTaskHistory(context.Context, string) ([]TaskEvent, error)
}

type CreateTaskCommand struct {
	Title, Description string
	Priority           core.Priority
	Tags               []string
	ParentID, Due      *string
}

// Nil patch fields are omitted. Nullable fields have explicit clear intent.
// Base optionally compares editable values and incarnation against current data.
type UpdateTaskCommand struct {
	ID                    string
	Title, Description    *string
	Priority              *core.Priority
	Status                *core.Status
	Progress              *int
	Tags                  *[]string
	Due, ParentID         *string
	ClearDue, ClearParent bool
	Base                  *core.Task
}
type TaskCommand struct {
	ID   string
	Base *core.Task
}
type ReopenTaskCommand struct {
	ID     string
	Status core.Status
	Base   *core.Task
}

// Expected is local consent, not an authorization token. Force bypasses consent
// only; Recursive is still required when the authoritative task has children.
type DeleteTaskCommand struct {
	ID               string
	Recursive, Force bool
	Expected         *DeletePreview
}
type DeletePreview struct {
	Target core.Task
	IDs    []string
}
type DeleteResult struct {
	ID           string
	DeletedIDs   []string
	DeletedCount int
	Deleted      bool
}
type TaskQuery struct {
	Filter core.TaskFilter
	All    bool
	Due    *string
}
type TaskStats struct {
	Total, Done, CompletionPercent, Overdue, CompletedLast7Days int
	ByStatus                                                    map[core.Status]int
}
