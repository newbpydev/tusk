package cli

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

// Append the explicit task wire schema directly to the invocation's buffer.
// A large snapshot need not also allocate a full slice of intermediate DTOs.
// Non-ASCII and escaped strings use the standard encoder's exact semantics.
func appendJSONString(dst []byte, s string) []byte {
	for i := range len(s) {
		c := s[i]
		if c < 0x20 || c >= 0x80 || c == '"' || c == '\\' || c == '<' || c == '>' || c == '&' {
			encoded, _ := json.Marshal(s) // A string cannot fail JSON encoding.
			return append(dst, encoded...)
		}
	}
	dst = append(dst, '"')
	dst = append(dst, s...)
	return append(dst, '"')
}

func appendJSONTime(dst []byte, t time.Time) ([]byte, error) {
	t = t.UTC()
	if t.Year() < 0 || t.Year() > 9999 {
		return nil, ports.ErrInvalidRecord
	}
	dst = append(dst, '"')
	dst = t.AppendFormat(dst, time.RFC3339Nano)
	return append(dst, '"'), nil
}

func appendOptionalTime(dst []byte, t *time.Time) ([]byte, error) {
	if t == nil {
		return append(dst, "null"...), nil
	}
	return appendJSONTime(dst, *t)
}

func appendTaskJSON(dst []byte, t taskJSON) ([]byte, error) {
	dst = append(dst, `{"id":`...)
	dst = appendJSONString(dst, t.ID)
	dst = append(dst, `,"title":`...)
	dst = appendJSONString(dst, t.Title)
	dst = append(dst, `,"description":`...)
	dst = appendJSONString(dst, t.Description)
	dst = append(dst, `,"status":`...)
	dst = appendJSONString(dst, string(t.Status))
	dst = append(dst, `,"priority":`...)
	dst = strconv.AppendInt(dst, int64(t.Priority), 10)
	dst = append(dst, `,"parent_id":`...)
	if t.ParentID == nil {
		dst = append(dst, "null"...)
	} else {
		dst = appendJSONString(dst, *t.ParentID)
	}
	dst = append(dst, `,"progress":`...)
	dst = strconv.AppendInt(dst, int64(t.Progress), 10)
	dst = append(dst, `,"tags":[`...)
	for i, tag := range t.Tags {
		if i > 0 {
			dst = append(dst, ',')
		}
		dst = appendJSONString(dst, string(tag))
	}
	dst = append(dst, `],"due_date":`...)
	var err error
	if dst, err = appendOptionalTime(dst, t.DueDate); err != nil {
		return nil, err
	}
	dst = append(dst, `,"created_at":`...)
	if dst, err = appendJSONTime(dst, t.CreatedAt); err != nil {
		return nil, err
	}
	dst = append(dst, `,"updated_at":`...)
	if dst, err = appendJSONTime(dst, t.UpdatedAt); err != nil {
		return nil, err
	}
	dst = append(dst, `,"completed_at":`...)
	if dst, err = appendOptionalTime(dst, t.CompletedAt); err != nil {
		return nil, err
	}
	return append(dst, '}'), nil
}

// Reserve the fixed keys, punctuation and maximum timestamp widths plus the
// unescaped text. Escaped text can grow the buffer normally; this is a capacity
// estimate only, never an output limit or an assumption about valid user text.
func taskJSONCapacity(t core.Task) int {
	n := 320 + len(t.ID) + len(t.Title) + len(t.Description) + len(t.Status)
	if t.ParentID != nil {
		n += len(*t.ParentID)
	}
	for _, tag := range t.Tags {
		n += len(tag) + 3
	}
	return n
}

func treeJSONCapacity(nodes []*core.TaskNode) (int, error) {
	n := 2
	for _, node := range nodes {
		if node == nil {
			return 0, ports.ErrInvalidRecord
		}
		children, err := treeJSONCapacity(node.Children)
		if err != nil {
			return 0, err
		}
		n += 40 + taskJSONCapacity(node.Task) + children
	}
	return n, nil
}

func appendTreeJSON(dst []byte, nodes []*core.TaskNode) ([]byte, error) {
	dst = append(dst, '[')
	for i, node := range nodes {
		if i > 0 {
			dst = append(dst, ',')
		}
		dst = append(dst, `{"task":`...)
		var err error
		if dst, err = appendTaskJSON(dst, taskDTO(node.Task)); err != nil {
			return nil, err
		}
		dst = append(dst, `,"children":`...)
		if dst, err = appendTreeJSON(dst, node.Children); err != nil {
			return nil, err
		}
		dst = append(dst, `,"depth":`...)
		dst = strconv.AppendInt(dst, int64(node.Depth), 10)
		dst = append(dst, '}')
	}
	return append(dst, ']'), nil
}
