package storage

import (
	"database/sql"
	"encoding/json"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	generated "github.com/newbpydev/tusk/internal/storage/sqlc"
)

func validText(s string) bool    { return utf8.ValidString(s) && !strings.ContainsRune(s, 0) }
func validID(s string) bool      { return s != "" && s == strings.TrimSpace(s) && validText(s) }
func validTime(t time.Time) bool { year := t.UTC().Year(); return year >= 1 && year <= 9999 }

func validateTask(t *core.Task) error {
	if t == nil {
		return ports.ErrInvalidRecord
	}
	if !validID(t.ID) {
		return core.ErrInvalidTaskID
	}
	if t.Title == "" {
		return core.ErrEmptyTitle
	}
	if !validText(t.Title) || t.Title != strings.TrimSpace(t.Title) || !validText(t.Description) {
		return ports.ErrInvalidRecord
	}
	if utf8.RuneCountInString(t.Title) > 255 {
		return core.ErrTitleTooLong
	}
	if !t.Status.IsValid() {
		return core.ErrInvalidStatus
	}
	if !t.Priority.IsValid() {
		return core.ErrInvalidPriority
	}
	if t.Progress < 0 || t.Progress > 100 {
		return core.ErrInvalidProgress
	}
	if t.ParentID != nil {
		if !validID(*t.ParentID) {
			return core.ErrInvalidTaskID
		}
		if *t.ParentID == t.ID {
			return core.ErrSelfParenting
		}
	}
	for i, tag := range t.Tags {
		normalized, err := core.NormalizeTag(string(tag))
		if err != nil || normalized != tag || i > 0 && t.Tags[i-1] >= tag {
			return core.ErrInvalidTag
		}
	}
	if t.CreatedAt.IsZero() || t.UpdatedAt.IsZero() || !validTime(t.CreatedAt) || !validTime(t.UpdatedAt) {
		return ports.ErrInvalidRecord
	}
	for _, date := range []*time.Time{t.DueDate, t.CompletedAt} {
		if date != nil && !validTime(*date) {
			return ports.ErrInvalidRecord
		}
	}
	if t.Status == core.StatusDone {
		if t.Progress != 100 || t.CompletedAt == nil {
			return ports.ErrInvalidRecord
		}
	} else if t.CompletedAt != nil {
		return ports.ErrInvalidRecord
	}
	return nil
}

func nullDate(t *time.Time) sql.NullString {
	if t == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: t.UTC().Format(dateLayout), Valid: true}
}
func parseDate(s string) (time.Time, error) {
	// Stored timestamps are exactly UTC with nine fractional digits. The RFC
	// parser has a fast path; shape checks retain the canonical disk contract
	// without allocating a formatted copy for every field on every task.
	if len(s) != len(dateLayout) || s[19] != '.' || s[29] != 'Z' {
		return time.Time{}, ports.ErrCorrupt
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil || !validTime(t) {
		return time.Time{}, ports.ErrCorrupt
	}
	return t, nil
}
func parseNullDate(s sql.NullString) (*time.Time, error) {
	if !s.Valid {
		return nil, nil
	}
	t, err := parseDate(s.String)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func encodeTask(t *core.Task) (generated.CreateTaskParams, error) {
	if err := validateTask(t); err != nil {
		return generated.CreateTaskParams{}, err
	}
	tags := t.Tags
	if tags == nil {
		tags = []core.Tag{}
	}
	jsonTags, _ := json.Marshal(tags)
	p := generated.CreateTaskParams{ID: t.ID, Title: t.Title, Description: t.Description, Status: string(t.Status), Priority: int64(t.Priority), Progress: int64(t.Progress), Tags: string(jsonTags), CreatedAt: t.CreatedAt.UTC().Format(dateLayout), UpdatedAt: t.UpdatedAt.UTC().Format(dateLayout), DueDate: nullDate(t.DueDate), CompletedAt: nullDate(t.CompletedAt)}
	if t.ParentID != nil {
		p.ParentID = sql.NullString{String: *t.ParentID, Valid: true}
	}
	return p, nil
}

func decodeTask(row generated.Task) (*core.Task, error) {
	t := new(core.Task)
	if err := decodeTaskInto(&row, t); err != nil {
		return nil, err
	}
	return t, nil
}

func decodeTaskInto(row *generated.Task, t *core.Task) error {
	*t = core.Task{ID: row.ID, Title: row.Title, Description: row.Description, Status: core.Status(row.Status), Priority: core.Priority(row.Priority), Progress: int(row.Progress)}
	if row.ParentID.Valid {
		p := row.ParentID.String
		t.ParentID = &p
	}
	if tags, ok := canonicalTags(row.Tags); ok {
		t.Tags = tags
	} else {
		// Keep the general decoder for malformed/noncanonical input so the
		// existing corruption and domain-error classifications are preserved.
		if err := json.Unmarshal([]byte(row.Tags), &t.Tags); err != nil || t.Tags == nil {
			return ports.ErrCorrupt
		}
		encoded, _ := json.Marshal(t.Tags)
		if string(encoded) != row.Tags {
			return ports.ErrCorrupt
		}
	}
	var err error
	if t.CreatedAt, err = parseDate(row.CreatedAt); err != nil {
		return err
	}
	if t.UpdatedAt, err = parseDate(row.UpdatedAt); err != nil {
		return err
	}
	if t.DueDate, err = parseNullDate(row.DueDate); err != nil {
		return err
	}
	if t.CompletedAt, err = parseNullDate(row.CompletedAt); err != nil {
		return err
	}
	if err := validateTask(t); err != nil {
		return corruptCause(err)
	}
	return nil
}

// Canonical persisted tags contain only unescaped ASCII. Parse that common
// representation without a JSON decoder, copied strings or a round-trip encode.
// validateTask still checks tag normalization, ordering and uniqueness.
func canonicalTags(raw string) ([]core.Tag, bool) {
	if raw == "[]" {
		return []core.Tag{}, true
	}
	if len(raw) < 4 || !strings.HasPrefix(raw, `["`) || !strings.HasSuffix(raw, `"]`) {
		return nil, false
	}
	inner := raw[2 : len(raw)-2]
	tags := make([]core.Tag, 0, strings.Count(inner, `","`)+1)
	for tag := range strings.SplitSeq(inner, `","`) {
		for i := range len(tag) {
			c := tag[i]
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
				return nil, false
			}
		}
		tags = append(tags, core.Tag(tag))
	}
	return tags, true
}

func validateEvent(e ports.TaskEvent) error {
	if !validID(e.TaskID) {
		return core.ErrInvalidTaskID
	}
	switch e.Kind {
	case ports.EventCreate, ports.EventMetadata, ports.EventStatus, ports.EventMove, ports.EventProgress, ports.EventRollup:
	default:
		return ports.ErrInvalidRecord
	}
	if e.OccurredAt.IsZero() || !validTime(e.OccurredAt) || len(e.ChangedFields) == 0 || !slices.IsSorted(e.ChangedFields) {
		return ports.ErrInvalidRecord
	}
	for i, f := range e.ChangedFields {
		if i > 0 && e.ChangedFields[i-1] == f {
			return ports.ErrInvalidRecord
		}
		switch f {
		case "title", "description", "status", "priority", "parent_id", "progress", "tags", "due_date", "completed_at":
		default:
			return ports.ErrInvalidRecord
		}
	}
	return nil
}

func decodeEvent(row generated.TaskEvent) (ports.TaskEvent, error) {
	e := ports.TaskEvent{Sequence: row.Sequence, TaskID: row.TaskID, Kind: ports.EventKind(row.Kind)}
	if err := json.Unmarshal([]byte(row.ChangedFields), &e.ChangedFields); err != nil {
		return ports.TaskEvent{}, ports.ErrCorrupt
	}
	encoded, _ := json.Marshal(e.ChangedFields)
	if string(encoded) != row.ChangedFields {
		return ports.TaskEvent{}, ports.ErrCorrupt
	}
	var err error
	e.OccurredAt, err = parseDate(row.OccurredAt)
	if err != nil || e.Sequence <= 0 || validateEvent(e) != nil {
		return ports.TaskEvent{}, ports.ErrCorrupt
	}
	return e, nil
}
