package ports

import (
	"context"
	"github.com/newbpydev/tusk/internal/core"
)

// TaskReader returns detached values materialized in one transaction snapshot.
type TaskReader interface {
	GetByID(context.Context, string) (*core.Task, error)
	List(context.Context, core.TaskFilter) ([]core.Task, error)
	ListChildren(context.Context, string) ([]core.Task, error)
	GetSubtree(context.Context, string) ([]core.Task, error)
	GetAncestors(context.Context, string) ([]core.Task, error)
	ListEvents(context.Context, string) ([]TaskEvent, error)
}

// TaskWriter mutates only its callback transaction. Every operation failure
// invalidates the change set, even when the callback ignores the returned error.
type TaskWriter interface {
	TaskReader
	Create(context.Context, *core.Task) error
	Update(context.Context, *core.Task) error
	Delete(context.Context, string, bool) ([]string, error)
	AppendEvent(context.Context, TaskEvent) (int64, error)
}

// TaskRepository owns callback admission and transaction lifetime. Callbacks are
// synchronous and cannot nest using their callback context. Handles cannot escape
// or be used concurrently. Do not call Close from a callback.
type TaskRepository interface {
	TaskReader
	WithRead(context.Context, func(context.Context, TaskReader) error) error
	WithWrite(context.Context, func(context.Context, TaskWriter) error) error
	Close() error
}
