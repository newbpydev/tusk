package tui

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/service/dateparse"
)

func TestFilter_ApplyClearAndLiteralRouting(t *testing.T) {
	a := fixtureNode("a", "Alpha", core.PriorityHigh, nil)
	a.Task.Tags = []core.Tag{"work", "design"}
	b := fixtureNode("b", "Beta", core.PriorityHigh, nil)
	b.Task.Tags = []core.Tag{"work"}
	b.Task.Status = core.StatusDone
	m := loadedModel(a, b)
	press(m, "f")
	press(m, " ")
	press(m, "right")
	press(m, "right")
	press(m, "right")
	press(m, " ") // todo or done
	press(m, "tab")
	press(m, "right")
	press(m, "right")
	press(m, " ") // high
	press(m, "tab")
	press(m, "work, design")
	press(m, "tab")
	press(m, "tab")
	press(m, "enter")
	if m.filters != nil || len(m.rows) != 1 || m.selectedTask().ID != "a" {
		t.Fatalf("filters not applied: %s", m.View())
	}
	press(m, "f")
	m.filters.field = 2
	cmd := press(m, "q")
	if cmd != nil || m.filters == nil || !strings.HasSuffix(m.filters.tags, "q") {
		t.Fatal("text routed to browse quit")
	}
	press(m, "esc")
	if len(m.rows) != 1 {
		t.Fatal("cancel applied draft")
	}
	press(m, "f")
	m.filters.field = 5
	press(m, "enter")
	if m.filters != nil || len(m.rows) != 2 || m.filter.HasPredicates() {
		t.Fatal("clear did not include done")
	}
	press(m, "f")
	press(m, " ")
	press(m, "ctrl+s")
	press(m, "esc")
	if len(m.rows) != 2 || m.filter.HasPredicates() {
		t.Fatal("browse Esc failed to clear")
	}
}

func TestFilter_EmptyApplyKeepsHeaderTruthful(t *testing.T) {
	m := loadedModel(fixtureNode("a", "Alpha", core.PriorityMedium, nil))
	press(m, "f")
	press(m, "ctrl+s")
	if m.filters != nil {
		t.Fatal("empty apply left the dialog open")
	}
	if m.workspaceTab != 0 {
		t.Fatalf("empty apply entered custom mode: workspaceTab=%d", m.workspaceTab)
	}
	if strings.Contains(m.View(), "Filters active") {
		t.Fatal("header reports filters with none active")
	}
	press(m, "f")
	press(m, " ")
	press(m, "ctrl+s")
	if m.workspaceTab != -1 || !strings.Contains(m.View(), "Filters active") {
		t.Fatal("predicate filter lost custom mode")
	}
}

func TestFilter_DayIsResolvedOnlyOnApply(t *testing.T) {
	m := loadedModel()
	zone, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	m.options.Location = zone
	m.now = time.Date(2026, 3, 7, 23, 59, 0, 0, zone)
	m.options.Now = func() time.Time { return time.Date(2026, 3, 7, 23, 59, 0, 0, zone) }
	refs := []time.Time{}
	m.options.DayBounds = func(s string, now time.Time, z *time.Location) (time.Time, time.Time, error) {
		refs = append(refs, now)
		return dateparse.DayBounds(s, now, z)
	}
	press(m, "f")
	m.filters.field = 3
	press(m, "not-a-day")
	press(m, "ctrl+s")
	if m.filters == nil || m.filters.err == "" || m.filters.due != "not-a-day" || m.dueStart != nil {
		t.Fatal("invalid expression discarded")
	}
	m.filters.due = "tomorrow"
	press(m, "ctrl+s")
	if m.filters != nil || m.dueEnd.Sub(*m.dueStart) != 23*time.Hour || m.dueLabel != "2026-03-08" || !strings.Contains(m.View(), "2026-03-08") {
		t.Fatal("resolved date missing or DST wrong")
	}
	start := *m.dueStart
	m.options.Now = func() time.Time { return time.Date(2026, 3, 9, 12, 0, 0, 0, zone) }
	m.Update(m.requestRefresh()())
	if *m.dueStart != start || len(refs) != 2 {
		t.Fatal("relative filter drifted during refresh")
	}
	press(m, "f")
	m.filters.due = ""
	press(m, "ctrl+s")
	if m.dueStart != nil || m.dueLabel != "" {
		t.Fatal("due clear failed")
	}
}

func TestFilter_RefreshAndResizePreserveDraft(t *testing.T) {
	m := loadedModel()
	press(m, "f")
	m.filters.field = 2
	press(m, "literal text")
	before := snapshot(m.filters)
	m.options.Load = func(context.Context) ([]*core.TaskNode, error) {
		return []*core.TaskNode{fixtureNode("a", "A", 1, nil)}, nil
	}
	m.Update(m.requestRefresh()())
	if snapshot(m.filters) != before {
		t.Fatal("refresh replaced draft")
	}
	for _, size := range [][2]int{{80, 24}, {120, 40}, {200, 60}, {79, 23}} {
		m.width, m.height = size[0], size[1]
		m.prepareFrame()
		if snapshot(m.filters) != before {
			t.Fatal("resize changed draft")
		}
	}
}

func TestFilter_FieldNavigationErrorsAndWideInput(t *testing.T) {
	m := loadedModel()
	press(m, "f")
	press(m, "up")
	if m.filters.field != 5 {
		t.Fatal("reverse focus did not wrap")
	}
	press(m, "q")
	press(m, "down")
	press(m, "left")
	press(m, " ")
	if !m.filters.statuses[3] {
		t.Fatal("choice did not wrap to done")
	}
	press(m, "tab")
	press(m, "enter")
	if m.filters.field != 2 {
		t.Fatal("Enter did not advance")
	}
	press(m, "invalid tag")
	press(m, "ctrl+s")
	if m.filters == nil || !strings.HasPrefix(m.filters.err, "Tags:") || m.filters.tags != "invalid tag" {
		t.Fatal("tag error discarded raw draft")
	}
	for range len("invalid tag") {
		press(m, "backspace")
	}
	press(m, "backspace")
	m.Update(tea.KeyMsg{Type: tea.KeySpace})
	press(m, "#Work")
	press(m, "tab")
	press(m, "tomorrow")
	press(m, "ctrl+s")
	if m.filters == nil || m.filters.err != "Due date is unavailable." {
		t.Fatal("missing calendar dependency panicked or applied")
	}
	press(m, "esc")
	press(m, "f")
	m.filters.field = 2
	text := strings.Repeat("界é", 80) + "THE END"
	press(m, text)
	frame := m.View()
	if !strings.Contains(frame, "THE END▎") || m.filters.tags != text {
		t.Fatal("long input hid cursor or rewrote data")
	}
}

func TestFilter_PlainChoiceHasVisibleFocus(t *testing.T) {
	m := loadedModel()
	press(m, "f")
	press(m, "right")
	if !strings.Contains(m.View(), ">[ ] In progress") {
		t.Fatal("plain profile hides selected checkbox focus")
	}
	press(m, " ")
	if !strings.Contains(m.View(), ">[✓] In progress") {
		t.Fatal("checked focus hidden")
	}
}
