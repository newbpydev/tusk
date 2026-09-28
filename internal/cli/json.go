package cli

import (
	"encoding/json"
	"slices"
	"time"

	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

type taskJSON struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Status      core.Status   `json:"status"`
	Priority    core.Priority `json:"priority"`
	ParentID    *string       `json:"parent_id"`
	Progress    int           `json:"progress"`
	Tags        []string      `json:"tags"`
	DueDate     *time.Time    `json:"due_date"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	CompletedAt *time.Time    `json:"completed_at"`
}
type nodeJSON struct {
	Task     taskJSON   `json:"task"`
	Children []nodeJSON `json:"children"`
	Depth    int        `json:"depth"`
}
type deleteJSON struct {
	ID      string   `json:"id"`
	IDs     []string `json:"deleted_ids"`
	Count   int      `json:"deleted_count"`
	Deleted bool     `json:"deleted"`
}
type statsJSON struct {
	Total              int                 `json:"total"`
	ByStatus           map[core.Status]int `json:"by_status"`
	Done               int                 `json:"done"`
	CompletionPercent  int                 `json:"completion_percent"`
	Overdue            int                 `json:"overdue"`
	CompletedLast7Days int                 `json:"completed_last_7_days"`
}
type eventJSON struct {
	Sequence   int64           `json:"sequence"`
	TaskID     string          `json:"task_id"`
	Kind       ports.EventKind `json:"kind"`
	Fields     []string        `json:"changed_fields"`
	OccurredAt time.Time       `json:"occurred_at"`
}

func utcPointer(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	v := t.UTC()
	return &v
}
func taskDTO(t core.Task) taskJSON {
	tags := make([]string, len(t.Tags))
	for n, tag := range t.Tags {
		tags[n] = string(tag)
	}
	return taskJSON{t.ID, t.Title, t.Description, t.Status, t.Priority, t.ParentID, t.Progress, tags, utcPointer(t.DueDate), t.CreatedAt.UTC(), t.UpdatedAt.UTC(), utcPointer(t.CompletedAt)}
}
func treeDTO(nodes []*core.TaskNode) ([]nodeJSON, error) {
	out := make([]nodeJSON, len(nodes))
	for n, node := range nodes {
		if node == nil {
			return nil, ports.ErrInvalidRecord
		}
		children, err := treeDTO(node.Children)
		if err != nil {
			return nil, err
		}
		out[n] = nodeJSON{taskDTO(node.Task), children, node.Depth}
	}
	return out, nil
}
func sortedStrings(values []string) []string {
	out := append([]string{}, values...)
	slices.Sort(out)
	return out
}
func jsonDTO(value any) (any, error) {
	switch v := value.(type) {
	case *core.Task:
		if v == nil {
			return nil, ports.ErrInvalidRecord
		}
		return taskDTO(*v), nil
	case []core.Task:
		out := make([]taskJSON, len(v))
		for n, t := range v {
			out[n] = taskDTO(t)
		}
		return out, nil
	case []*core.TaskNode:
		return treeDTO(v)
	case ports.DeleteResult:
		return deleteJSON{v.ID, sortedStrings(v.DeletedIDs), v.DeletedCount, v.Deleted}, nil
	case ports.TaskStats:
		counts := map[core.Status]int{}
		for _, status := range []core.Status{core.StatusTodo, core.StatusInProgress, core.StatusBlocked, core.StatusDone} {
			counts[status] = v.ByStatus[status]
		}
		return statsJSON{v.Total, counts, v.Done, v.CompletionPercent, v.Overdue, v.CompletedLast7Days}, nil
	case []ports.TaskEvent:
		out := make([]eventJSON, len(v))
		for n, e := range v {
			out[n] = eventJSON{e.Sequence, e.TaskID, e.Kind, sortedStrings(e.ChangedFields), e.OccurredAt.UTC()}
		}
		return out, nil
	default:
		return nil, ports.ErrInvalidRecord
	}
}
func encodeJSON(value any) ([]byte, error) {
	dto, err := jsonDTO(value)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(dto)
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
