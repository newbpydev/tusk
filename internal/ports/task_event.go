package ports

import "time"

type EventKind string

const (
	EventCreate   EventKind = "create"
	EventMetadata EventKind = "metadata"
	EventStatus   EventKind = "status"
	EventMove     EventKind = "move"
	EventProgress EventKind = "progress"
	EventRollup   EventKind = "rollup"
)

// TaskEvent retains only changed field names, never previous content or values.
// AppendEvent requires Sequence == 0 and assigns the persisted sequence itself.
// ChangedFields must be nonempty, sorted, unique, and drawn from title,
// description, status, priority, parent_id, progress, tags, due_date, completed_at.
// OccurredAt must be nonzero with its UTC year in 1..9999.
type TaskEvent struct {
	Sequence      int64
	TaskID        string
	Kind          EventKind
	ChangedFields []string
	OccurredAt    time.Time
}
