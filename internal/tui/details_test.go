package tui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

// Deliver immediately-completing UI commands; leave the recurring clock chain
// unconsumed. Production Bubble Tea runs these commands independently.
func deliverUI(m *Model, cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	switch msg := cmd().(type) {
	case tea.BatchMsg:
		for _, c := range msg {
			deliverUI(m, c)
		}
	case tickMsg:
	default:
		_, next := m.Update(msg)
		deliverUI(m, next)
	}
}

func TestDetails_MetadataAndIndependentTimeline(t *testing.T) {
	n := fixtureNode("0199abcdef-0123-7123-8123-123456789abc", "Complete title with 界 é", core.PriorityHigh, nil)
	n.Task.Progress = 42
	n.Task.Tags = []core.Tag{"design", "work"}
	n.Task.UpdatedAt = n.Task.CreatedAt.Add(time.Hour)
	o := testOptions()
	o.Location = time.FixedZone("TEST", -3*3600)
	o.Load = func(context.Context) ([]*core.TaskNode, error) { return []*core.TaskNode{n}, nil }
	o.History = func(context.Context, string) ([]ports.TaskEvent, error) {
		return []ports.TaskEvent{{Sequence: 9007199254740993, Kind: ports.EventMetadata, OccurredAt: n.Task.UpdatedAt, ChangedFields: []string{"priority", "title"}}, {Sequence: 1, Kind: ports.EventCreate, OccurredAt: n.Task.CreatedAt, ChangedFields: []string{"title"}}}, nil
	}
	m := sizedModel(o)
	deliverUI(m, m.Init())
	all := strings.Join(m.detailLines(44), "\n")
	for _, want := range []string{n.Task.ID, "Complete title", "42%", "Root task", "design", "work", "Due", "Not set", "Created", "Updated", "Completed", "TEST (-03:00)", "No notes", "Activity", "9007199254740993", "priority, title"} {
		if !strings.Contains(strings.ReplaceAll(all, "\n  ", ""), want) {
			t.Errorf("missing %q in details", want)
		}
	}
	if strings.Index(all, "#1 ") > strings.Index(all, "#9007199254740993 ") {
		t.Fatal("history sequence is not ascending int64")
	}
	if n.Task.CreatedAt.Location() != time.UTC {
		t.Fatal("presentation modified stored time")
	}
	before := snapshot(m)
	m.View()
	if snapshot(m) != before {
		t.Fatal("details View mutated state")
	}
}

func TestMarkdown_LatestPendingRenderAndFallback(t *testing.T) {
	n := fixtureNode("one", "One", 1, nil)
	m := loadedModel(n)
	inputs := []string{}
	m.options.Render = func(_ context.Context, s string, _ int, _ termenv.Profile) (string, error) {
		inputs = append(inputs, s)
		if s == "failure" {
			return "", errors.New("private renderer path")
		}
		return "rendered " + s, nil
	}
	n.Task.Description = "first"
	_, first := m.Update(struct{}{})
	if first == nil {
		t.Fatal("no render command")
	}
	n.Task.Description = "second"
	if _, cmd := m.Update(struct{}{}); cmd != nil {
		t.Fatal("parallel render admitted")
	}
	n.Task.Description = "latest"
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	_, next := m.Update(first())
	if next == nil {
		t.Fatal("latest pending render lost")
	}
	deliverUI(m, next)
	if len(inputs) != 2 || inputs[0] != "first" || inputs[1] != "latest" || !strings.Contains(strings.Join(m.detailLines(70), "\n"), "rendered latest") {
		t.Fatalf("wrong immutable render inputs: %v", inputs)
	}
	n.Task.Description = "failure"
	_, cmd := m.Update(struct{}{})
	deliverUI(m, cmd)
	all := strings.Join(m.detailLines(70), "\n")
	if !strings.Contains(all, "Plain text") || !strings.Contains(all, "failure") || strings.Contains(all, "private renderer") {
		t.Fatal("unsafe or absent plain fallback")
	}
}

func TestDetails_LateHistoryForOtherTask(t *testing.T) {
	a, b := fixtureNode("a", "A", 2, nil), fixtureNode("b", "B", 1, nil)
	m := loadedModel(a, b)
	m.options.History = func(_ context.Context, id string) ([]ports.TaskEvent, error) {
		return []ports.TaskEvent{{Sequence: 1, TaskID: id, Kind: ports.EventCreate, ChangedFields: []string{id}}}, nil
	}
	m.historyRefresh = true
	_, first := m.Update(struct{}{})
	press(m, "down")
	if strings.Contains(strings.Join(m.detailLines(44), "\n"), "Changed a") {
		t.Fatal("prior history retained on new task")
	}
	_, next := m.Update(first())
	if next == nil {
		t.Fatal("latest history request lost")
	}
	deliverUI(m, next)
	if m.selectedTask().ID != "b" || len(m.history.events) != 1 || m.history.events[0].TaskID != "b" {
		t.Fatal("stale history entered other task")
	}
}

func TestDetails_FatalReadDoesNotStartPendingHistory(t *testing.T) {
	m := loadedModel(fixtureNode("one", "One", 1, nil))
	m.options.History = func(context.Context, string) ([]ports.TaskEvent, error) {
		t.Fatal("history after fatal read")
		return nil, nil
	}
	m.historyRefresh = true
	_ = m.requestRefresh()
	_, cmd := m.Update(forestMsg{owner: m.owner, operation: m.operation, err: errRuntime})
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("fatal read scheduled more work before quitting")
	}
}

func TestMarkdown_IdentityWidthProfileAndPanicFallback(t *testing.T) {
	a, b := fixtureNode("a", "A", 2, nil), fixtureNode("b", "B", 1, nil)
	a.Task.Description = "private A"
	b.Task.Description = "public B"
	m := sizedModel(testOptions())
	m.forest = []*core.TaskNode{a, b}
	m.state = loaded
	m.busy = false
	m.options.Render = func(_ context.Context, s string, _ int, _ termenv.Profile) (string, error) { return s, nil }
	_, first := m.Update(struct{}{})
	press(m, "down")
	m.options.Profile = termenv.ANSI
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	_, next := m.Update(first())
	if strings.Contains(strings.Join(m.detailLines(70), "\n"), "private A") {
		t.Fatal("old task render leaked")
	}
	deliverUI(m, next)
	if !strings.Contains(strings.Join(m.detailLines(70), "\n"), "public B") {
		t.Fatal("new task render lost")
	}
	stale := notesMsg{token: m.markdown.token - 1, key: m.markdown.wanted, lines: []string{"stale"}}
	m.Update(stale)
	if strings.Contains(strings.Join(m.detailLines(70), "\n"), "stale") {
		t.Fatal("old render token applied")
	}
	m.options.Render = func(context.Context, string, int, termenv.Profile) (string, error) { panic("private panic") }
	b.Task.Description = "safe fallback"
	_, cmd := m.Update(struct{}{})
	deliverUI(m, cmd)
	if !m.markdown.plain || m.exitErr != nil || !strings.Contains(strings.Join(m.detailLines(70), "\n"), "safe fallback") {
		t.Fatal("render panic did not safely degrade")
	}
	m.forest = nil
	_, cmd = m.Update(struct{}{})
	if cmd != nil || m.selectedTask() != nil || strings.Contains(m.View(), "public B") {
		t.Fatal("empty selection retained prior details")
	}
}

func TestDetails_VisibleScrollPositionAndControls(t *testing.T) {
	n := fixtureNode("one", "One", 1, nil)
	n.Task.Description = strings.Repeat("A line of notes\n", 100)
	m := loadedModel(n)
	_, cmd := m.Update(struct{}{})
	deliverUI(m, cmd)
	press(m, "tab")
	lines := strings.Split(m.View(), "\n")
	footer := lines[len(lines)-1]
	for _, hint := range []string{"scroll", "PgUp/PgDn", "Home/End", "Tab", "help", "quit"} {
		if !strings.Contains(footer, hint) {
			t.Errorf("missing %q in %s", hint, footer)
		}
	}
	if !strings.Contains(lines[1], "1–20 /") {
		t.Fatal("scroll position missing from panel title")
	}
	press(m, "end")
	if strings.Contains(strings.Split(m.View(), "\n")[1], "1–20 /") {
		t.Fatal("scroll position did not update")
	}
}

func TestDetails_EndIntentSurvivesPendingRenderAndResize(t *testing.T) {
	n := fixtureNode("one", "One", 1, nil)
	m := loadedModel(n)
	n.Task.Description = strings.Repeat("Long readable note line.\n", 700) + "END OF NOTES"
	m.options.Render = func(_ context.Context, s string, _ int, _ termenv.Profile) (string, error) { return s, nil }
	_, first := m.Update(struct{}{})
	press(m, "tab")
	press(m, "end")
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	_, next := m.Update(first())
	deliverUI(m, next)
	if !strings.Contains(m.View(), "END OF NOTES") || m.detailsScroll < 600 {
		t.Fatal("End intent was lost while notes were being rendered")
	}
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if !strings.Contains(m.View(), "END OF NOTES") {
		t.Fatal("resize discarded same-task last-good notes")
	}
	press(m, "home")
	if m.detailsScroll != 0 {
		t.Fatal("Home did not release End intent")
	}
}
