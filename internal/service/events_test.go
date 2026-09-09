package service

import (
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"reflect"
	"testing"
)

func TestEvents_NetDiff(t *testing.T) {
	before := task("id", "", 0, core.StatusTodo)
	after := before.Clone()
	after.Title = "new"
	after.Description = "notes"
	after.Priority = core.PriorityHigh
	after.Tags = []core.Tag{"work"}
	after.DueDate = ptr(fixedTime())
	after.ParentID = ptr("p")
	after.Status = core.StatusDone
	after.Progress = 100
	after.CompletedAt = ptr(fixedTime())
	events := taskEvents(&before, &after, fixedTime(), true, false)
	want := []struct {
		k ports.EventKind
		f []string
	}{{ports.EventMetadata, []string{"description", "due_date", "priority", "tags", "title"}}, {ports.EventMove, []string{"parent_id"}}, {ports.EventStatus, []string{"completed_at", "progress", "status"}}}
	if len(events) != len(want) {
		t.Fatalf("%+v", events)
	}
	for i, e := range events {
		if e.Kind != want[i].k || !reflect.DeepEqual(e.ChangedFields, want[i].f) || e.Sequence != 0 || e.TaskID != "id" || !e.OccurredAt.Equal(fixedTime()) {
			t.Fatalf("%+v", e)
		}
	}
	if len(taskEvents(&before, &before, fixedTime(), false, false)) != 0 {
		t.Fatal("no-op events")
	}
	for _, manual := range []bool{false, true} {
		a := before.Clone()
		a.Progress = 10
		e := taskEvents(&before, &a, fixedTime(), false, manual)
		kind := ports.EventRollup
		if manual {
			kind = ports.EventProgress
		}
		if len(e) != 1 || e[0].Kind != kind {
			t.Fatalf("%v", e)
		}
	}
	if e := taskEvents(nil, &after, fixedTime(), false, false); len(e) != 1 || e[0].Kind != ports.EventCreate || len(e[0].ChangedFields) != 9 {
		t.Fatalf("%v", e)
	}
}
