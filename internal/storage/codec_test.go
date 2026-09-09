package storage

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	generated "github.com/newbpydev/tusk/internal/storage/sqlc"
)

func TestCodec_RejectsInvalidInput(t *testing.T) {
	if _, err := encodeTask(nil); !errors.Is(err, ports.ErrInvalidRecord) {
		t.Fatal(err)
	}
	for _, change := range []func(*core.Task){
		func(t *core.Task) { t.ID = "" }, func(t *core.Task) { t.ID = " padded" }, func(t *core.Task) { t.Title = "" }, func(t *core.Task) { t.Title = " padded" }, func(t *core.Task) { t.Title = strings.Repeat("ü", 256) },
		func(t *core.Task) { t.Description = "secret\x00" }, func(t *core.Task) { t.Title = "\xff" }, func(t *core.Task) { t.Status = "invalid" }, func(t *core.Task) { t.Priority = 0 }, func(t *core.Task) { t.Progress = 101 },
		func(t *core.Task) { t.Tags = []core.Tag{"Z"} }, func(t *core.Task) { t.Tags = []core.Tag{"a", "a"} }, func(t *core.Task) { t.Tags = []core.Tag{"z", "a"} },
		func(t *core.Task) { t.CreatedAt = time.Time{} }, func(t *core.Task) { t.UpdatedAt = time.Time{} }, func(t *core.Task) {
			t.DueDate = &t.UpdatedAt
			t.UpdatedAt = time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC)
		},
		func(t *core.Task) { t.ParentID = &t.ID }, func(t *core.Task) { s := " "; t.ParentID = &s }, func(t *core.Task) { t.Status = core.StatusDone }, func(t *core.Task) { t.CompletedAt = &t.CreatedAt },
	} {
		task := taskFixture("task")
		change(task)
		if _, err := encodeTask(task); err == nil {
			t.Fatalf("accepted invalid task %+v", task)
		}
	}
}

func storedTask() generated.Task {
	return generated.Task{ID: "task", Title: "title", Description: "", Status: "todo", Priority: 2, Progress: 0, Tags: "[]", CreatedAt: "2026-09-08T12:00:00.000000123Z", UpdatedAt: "2026-09-08T12:00:00.000000123Z"}
}
func TestCodec_RejectsCorruptRows(t *testing.T) {
	for _, change := range []func(*generated.Task){
		func(r *generated.Task) { r.Tags = "not-json" }, func(r *generated.Task) { r.Tags = "null" }, func(r *generated.Task) { r.Tags = "[1]" }, func(r *generated.Task) { r.Tags = "[ ]" }, func(r *generated.Task) { r.Tags = "[\"z\",\"a\"]" },
		func(r *generated.Task) { r.CreatedAt = "secret" }, func(r *generated.Task) { r.UpdatedAt = "secret" }, func(r *generated.Task) { r.DueDate = sql.NullString{String: "", Valid: true} }, func(r *generated.Task) { r.CompletedAt = sql.NullString{String: "bad", Valid: true} },
		func(r *generated.Task) { r.ID = " padded" }, func(r *generated.Task) { r.Title = "" }, func(r *generated.Task) { r.Title = "\xff" }, func(r *generated.Task) { r.Description = "private\x00" }, func(r *generated.Task) { r.Status = "done" }, func(r *generated.Task) { r.Priority = 0 }, func(r *generated.Task) { r.Progress = -1 }, func(r *generated.Task) { r.ParentID = sql.NullString{String: r.ID, Valid: true} },
	} {
		r := storedTask()
		change(&r)
		if got, err := decodeTask(r); got != nil || !errors.Is(err, ports.ErrCorrupt) {
			t.Fatalf("corrupt row returned: %+v %v", got, err)
		}
	}
	task := taskFixture("utc")
	task.Status = core.StatusDone
	task.Progress = 100
	complete := task.CreatedAt.In(time.FixedZone("offset", -3*3600))
	task.CompletedAt = &complete
	if _, err := encodeTask(task); err != nil {
		t.Fatal(err)
	}
	task.Progress = 99
	if _, err := encodeTask(task); err == nil {
		t.Fatal("incomplete done accepted")
	}
	task = taskFixture("parent")
	task.Progress = 100
	task.Tags = nil
	if p, err := encodeTask(task); err != nil || p.Tags != "[]" {
		t.Fatalf("open progress 100: %v", err)
	}
	task.DueDate = &time.Time{}
	if _, err := encodeTask(task); err != nil {
		t.Fatal("valid year 0001 optional date rejected")
	}
	outOfRange := time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC)
	task.DueDate = &outOfRange
	if _, err := encodeTask(task); err == nil {
		t.Fatal("out-of-range optional date accepted")
	}
}

func TestCodec_EventValidation(t *testing.T) {
	valid := ports.TaskEvent{TaskID: "task", Kind: ports.EventCreate, ChangedFields: []string{"title"}, OccurredAt: taskFixture("x").CreatedAt}
	for _, change := range []func(*ports.TaskEvent){
		func(e *ports.TaskEvent) { e.TaskID = "" }, func(e *ports.TaskEvent) { e.Kind = "delete" }, func(e *ports.TaskEvent) { e.OccurredAt = time.Time{} }, func(e *ports.TaskEvent) { e.ChangedFields = nil }, func(e *ports.TaskEvent) { e.ChangedFields = []string{"title", "description"} }, func(e *ports.TaskEvent) { e.ChangedFields = []string{"title", "title"} }, func(e *ports.TaskEvent) { e.ChangedFields = []string{"old_notes"} },
	} {
		e := valid
		change(&e)
		if err := validateEvent(e); err == nil {
			t.Fatal("invalid event accepted")
		}
	}
	for _, kind := range []ports.EventKind{ports.EventCreate, ports.EventMetadata, ports.EventStatus, ports.EventMove, ports.EventProgress, ports.EventRollup} {
		e := valid
		e.Kind = kind
		if err := validateEvent(e); err != nil {
			t.Fatal(err)
		}
	}
	for _, fields := range []string{"bad", "null", "[ ]", "[\"private_notes\"]", "[\"title\",\"title\"]"} {
		if _, err := decodeEvent(generated.TaskEvent{Sequence: 1, TaskID: "task", Kind: "create", ChangedFields: fields, OccurredAt: valid.OccurredAt.Format(dateLayout)}); !errors.Is(err, ports.ErrCorrupt) {
			t.Fatal(err)
		}
	}
	for _, row := range []generated.TaskEvent{
		{Sequence: 0, TaskID: "task", Kind: "create", ChangedFields: "[\"title\"]", OccurredAt: valid.OccurredAt.Format(dateLayout)},
		{Sequence: 1, TaskID: "task", Kind: "create", ChangedFields: "[\"title\"]", OccurredAt: "bad"},
	} {
		if _, err := decodeEvent(row); !errors.Is(err, ports.ErrCorrupt) {
			t.Fatal(err)
		}
	}
}
