package tui

import (
	"context"
	"testing"
	"time"

	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/service/dateparse"
)

func TestForm_RelativeDueEditIsNotStaleNoop(t *testing.T) {
	zone, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Fatal(err)
	}
	before := time.Date(2026, 9, 30, 23, 59, 59, 900000000, zone)
	current := before.Add(200 * time.Millisecond)
	for _, expression := range []string{"today", "tomorrow"} {
		t.Run(expression, func(t *testing.T) {
			oldDue, err := dateparse.ParseDue(expression, before, zone)
			if err != nil {
				t.Fatal(err)
			}
			node := fixtureNode("existing", "Existing task", core.PriorityMedium, &oldDue)
			m := loadedModel(node)
			// The cached display clock still precedes midnight; no timer or
			// forest response has arrived before the user's explicit Save.
			m.now = before
			m.options.Location = zone
			m.options.Now = func() time.Time { return current }
			m.options.ParseDue = dateparse.ParseDue
			var requests []mutationRequest
			m.options.Mutate = func(_ context.Context, request mutationRequest) (*core.Task, error) {
				requests = append(requests, request)
				task := node.Task.Clone()
				return &task, nil
			}
			press(m, "e")
			setFormField(m, fieldDue, expression)
			deliverUI(m, press(m, "ctrl+s"))
			if len(requests) != 1 || requests[0].update.Due == nil || *requests[0].update.Due != expression {
				t.Fatalf("relative edit discarded as stale no-op: requests=%+v notice=%q", requests, m.notice)
			}
			if !requests[0].update.Base.DueDate.Equal(oldDue) {
				t.Fatal("submission clock changed detached Base")
			}
		})
	}
}

func TestFilter_RelativeDayUsesApplyClock(t *testing.T) {
	zone, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	// Cross into the DST transition day between the last display tick and Apply.
	before := time.Date(2026, 3, 7, 23, 59, 59, 900000000, zone)
	current := before.Add(200 * time.Millisecond)
	start, end, err := dateparse.DayBounds("today", current, zone)
	if err != nil {
		t.Fatal(err)
	}
	oldDue, newDue := before, current
	m := loadedModel(fixtureNode("old", "Previous day", core.PriorityMedium, &oldDue), fixtureNode("new", "Current day", core.PriorityMedium, &newDue))
	m.now = before
	m.options.Location = zone
	m.options.Now = func() time.Time { return current }
	m.options.DayBounds = dateparse.DayBounds
	press(m, "f")
	m.filters.due = "today"
	press(m, "ctrl+s")
	if m.dueStart == nil || m.dueEnd == nil || !m.dueStart.Equal(start) || !m.dueEnd.Equal(end) {
		t.Fatalf("applied bounds=%v..%v want %v..%v", m.dueStart, m.dueEnd, start, end)
	}
	if len(m.rows) != 1 || m.selectedTask().ID != "new" || m.rows[0].group != "Today" {
		t.Fatalf("current-day projection uses stale display clock: %+v", m.rows)
	}
}
