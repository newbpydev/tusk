package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/core"
)

func press(m *Model, k string) tea.Cmd {
	types := map[string]tea.KeyType{"up": tea.KeyUp, "down": tea.KeyDown, "left": tea.KeyLeft, "right": tea.KeyRight, "tab": tea.KeyTab, "esc": tea.KeyEsc, "enter": tea.KeyEnter, "pgdown": tea.KeyPgDown, "pgup": tea.KeyPgUp, "home": tea.KeyHome, "end": tea.KeyEnd, "backspace": tea.KeyBackspace, "ctrl+s": tea.KeyCtrlS}
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
	if key, ok := types[k]; ok {
		msg = tea.KeyMsg{Type: key}
	}
	_, cmd := m.Update(msg)
	return cmd
}

func loadedModel(nodes ...*core.TaskNode) *Model {
	o := testOptions()
	o.Load = func(context.Context) ([]*core.TaskNode, error) { return nodes, nil }
	m := New(o)
	deliverUI(m, m.Init())
	return m
}

func TestNavigation_SelectionCollapseAndPages(t *testing.T) {
	child := fixtureNode("child", "Child", 2, nil)
	root := fixtureNode("root", "Root", 3, nil, child)
	other := fixtureNode("other", "Other", 1, nil)
	m := loadedModel(other, root)
	if m.selectedTask().ID != "root" {
		t.Fatal("canonical first selection")
	}
	press(m, "down")
	if m.selectedTask().ID != "child" {
		t.Fatal("group heading selected or child skipped")
	}
	press(m, "left")
	if len(m.rows) != 3 {
		t.Fatal("leaf collapse changed tree")
	}
	m.collapse("root")
	m.prepareFrame()
	if m.selectedTask().ID != "root" || len(m.rows) != 2 {
		t.Fatal("collapsing selected descendant did not choose ancestor")
	}
	press(m, "right")
	if len(m.rows) != 3 {
		t.Fatal("expand ignored")
	}
	press(m, "G")
	if m.selectedTask().ID != "other" {
		t.Fatal("last selection")
	}
	press(m, "down")
	if m.selectedTask().ID != "other" {
		t.Fatal("selection overflow")
	}
	press(m, "pgup")
	if m.selectedTask().ID != "root" {
		t.Fatal("page clamp")
	}
	press(m, "pgdown")
	if m.selectedTask().ID != "other" {
		t.Fatal("page down")
	}
	press(m, "g")
	press(m, "up")
	if m.selectedTask().ID != "root" {
		t.Fatal("first clamp")
	}
	press(m, "tab")
	press(m, "down")
	if m.selectedTask().ID != "root" {
		t.Fatal("details scrolling changed selection")
	}
}

func TestSelection_RemovedMovedAndIncarnation(t *testing.T) {
	a, b, c := fixtureNode("a", "A", 3, nil), fixtureNode("b", "B", 2, nil), fixtureNode("c", "C", 1, nil)
	m := loadedModel(a, b, c)
	press(m, "down")
	if m.selectedTask().ID != "b" {
		t.Fatal("fixture")
	}
	b.Task.Priority = 4
	m.forest = []*core.TaskNode{c, b, a}
	m.prepareFrame()
	if m.selectedTask().ID != "b" {
		t.Fatal("reorder lost ID")
	}
	m.forest = []*core.TaskNode{a, c}
	m.prepareFrame()
	if m.selectedTask().ID != "a" {
		t.Fatal("removed selection did not use clamped prior index")
	}
	m.forest = nil
	m.prepareFrame()
	if m.selectedTask() != nil || m.selected != -1 {
		t.Fatal("empty selection retained")
	}
	for _, k := range []string{"up", "down", "left", "right", "g", "G", "pgup", "pgdown", "e", "d", "x"} {
		if press(m, k) != nil {
			t.Fatal("empty selection dispatched an action")
		}
	}
}

func TestSearch_LiteralEntryAndRestore(t *testing.T) {
	child := fixtureNode("child", "A needle", 2, nil)
	root := fixtureNode("root", "Root", 3, nil, child)
	m := loadedModel(root)
	m.collapse("root")
	m.prepareFrame()
	press(m, "/")
	press(m, "needle")
	press(m, "enter")
	if len(m.rows) != 2 || !strings.Contains(m.View(), "context") {
		t.Fatal("search lost descendant context")
	}
	press(m, "esc")
	if len(m.rows) != 1 || !m.collapsed["root"] {
		t.Fatal("clear did not restore collapse")
	}
	press(m, "/")
	if cmd := press(m, "q"); cmd == nil || !m.searching {
		t.Fatal("literal q quit search")
	}
	press(m, "esc")
	if m.filter.SearchTerm != "" || len(m.rows) != 1 {
		t.Fatal("cancel did not restore query")
	}
}

func TestSelection_IncarnationAndRemovedCollapse(t *testing.T) {
	a, b := fixtureNode("a", "A", 3, nil), fixtureNode("b", "B", 2, nil)
	m := loadedModel(a, b)
	press(m, "down")
	replacement := fixtureNode("b", "New incarnation", 4, nil)
	replacement.Task.CreatedAt = replacement.Task.CreatedAt.Add(time.Hour)
	m.options.Load = func(context.Context) ([]*core.TaskNode, error) { return []*core.TaskNode{replacement, a}, nil }
	m.Update(m.requestRefresh()())
	if m.selectedTask().ID != "a" {
		t.Fatal("same ID with another incarnation inherited selection")
	}
	m.collapsed["removed"] = true
	press(m, "/")
	m.Update(m.requestRefresh()())
	press(m, "esc")
	if m.collapsed["removed"] {
		t.Fatal("search restore resurrected removed expansion")
	}
}

func TestSearch_DebounceOrderAcceptCancelAndLiteralKeys(t *testing.T) {
	m := loadedModel(fixtureNode("a", "Alpha", 2, nil), fixtureNode("b", "Beta", 1, nil))
	m.options.Wait = func(ctx context.Context, d time.Duration) error {
		if d != 150*time.Millisecond {
			t.Fatalf("delay %v", d)
		}
		return ctx.Err()
	}
	press(m, "/")
	first := press(m, "a")
	second := press(m, "b")
	old := first().(searchMsg)
	if !errors.Is(old.err, context.Canceled) {
		t.Fatal("superseded timer still live")
	}
	m.Update(second())
	if m.filter.SearchTerm != "ab" {
		t.Fatal("latest query missing")
	}
	old.err = nil
	m.Update(old)
	if m.filter.SearchTerm != "ab" {
		t.Fatal("old timer applied")
	}
	pending := press(m, "q")
	if _, ok := pending().(tea.QuitMsg); ok || !m.searching {
		t.Fatal("literal q quit")
	}
	press(m, "enter")
	m.Update(searchMsg{token: m.searchToken, query: "obsolete"})
	if m.filter.SearchTerm != "abq" {
		t.Fatal("Enter lost accepted text")
	}
	press(m, "/")
	press(m, "backspace")
	m.Update(tea.KeyMsg{Type: tea.KeySpace})
	press(m, "tab")
	press(m, "esc")
	if m.filter.SearchTerm != "abq" || m.searching {
		t.Fatal("Esc failed to restore entry query")
	}
	press(m, "esc")
	press(m, "/")
	press(m, "backspace")
	press(m, "esc")
	if m.filter.SearchTerm != "" {
		t.Fatal("backspace on empty query")
	}
}

func TestNavigation_DiscoverableBrowseActions(t *testing.T) {
	m := loadedModel()
	footer := strings.Split(m.View(), "\n")[23]
	for _, hint := range []string{"search", "filter", "refresh", "help", "quit"} {
		if !strings.Contains(footer, hint) {
			t.Errorf("missing %s in %s", hint, footer)
		}
	}
}

func TestNavigation_ViewportAndSearchCollapseIsolation(t *testing.T) {
	nodes := make([]*core.TaskNode, 1000)
	for i := range nodes {
		nodes[i] = fixtureNode(fmt.Sprintf("%04d", i), fmt.Sprintf("Task %04d", i), 1, nil)
	}
	nodes[999].Task.Description = strings.Repeat("notes\n", 100)
	m := loadedModel(nodes...)
	deliverUI(m, press(m, "end"))
	if m.selectedTask().ID != "0999" || !strings.Contains(m.View(), "Task 0999") || m.listOffset == 0 {
		t.Fatal("last selected task is outside viewport")
	}
	press(m, "tab")
	press(m, "end")
	if m.detailsScroll == 0 {
		t.Fatal("details did not scroll")
	}
	press(m, "pgup")
	press(m, "up")
	press(m, "pgdown")
	press(m, "home")
	if m.detailsScroll != 0 || m.selectedTask().ID != "0999" {
		t.Fatal("details scroll changed task selection")
	}
	press(m, "tab")
	press(m, "home")
	if !strings.Contains(m.View(), "Task 0000") {
		t.Fatal("first selection outside viewport")
	}
	root := fixtureNode("root", "Root", 2, nil, fixtureNode("child", "Child", 1, nil))
	m = loadedModel(root)
	m.collapse("root")
	m.prepareFrame()
	press(m, "/")
	press(m, "child")
	press(m, "enter")
	press(m, "left")
	press(m, "right")
	press(m, "esc")
	if !m.collapsed["root"] || len(m.rows) != 1 {
		t.Fatal("filtered expansion overwrote browse collapse")
	}
}

func TestView_NewBrowseModesArePure(t *testing.T) {
	m := loadedModel(fixtureNode("one", "One", 1, nil))
	for _, key := range []string{"/", "needle", "esc", "f", "tab", "esc", "?"} {
		press(m, key)
		before := snapshot(m)
		frame := m.View()
		for range 10 {
			if m.View() != frame {
				t.Fatal("frame changed")
			}
		}
		if snapshot(m) != before {
			t.Fatal("View changed nested state")
		}
	}
	defaults := New(Options{})
	if defaults.options.Context == nil || defaults.options.Location == nil || defaults.options.Now == nil || defaults.options.Wait == nil {
		t.Fatal("missing defaults")
	}
}

func TestNavigation_SpaciousMetadataFollowsTreeIndent(t *testing.T) {
	m := loadedModel(fixtureNode("root", "Root", 2, nil, fixtureNode("child", "Child", 1, nil)))
	lines := m.listLines(46, 36)
	found := false
	for i, line := range lines {
		if strings.Contains(line, "└ Child") {
			found = true
			if !strings.HasPrefix(lines[i+1], "      todo") {
				t.Fatalf("child metadata detached from title: %q", lines[i+1])
			}
		}
	}
	if !found {
		t.Fatal("fixture child not rendered")
	}
}
