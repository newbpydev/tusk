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
type TaskEvent struct {
	Sequence      int64
	TaskID        string
	Kind          EventKind
	ChangedFields []string
	OccurredAt    time.Time
}
