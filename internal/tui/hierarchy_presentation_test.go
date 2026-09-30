package tui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

func TestHierarchy_ConnectedSpaciousBranches(t *testing.T) {
	first := fixtureNode("first", "First", 3, nil,
		fixtureNode("grand-last", "Grand last", 1, nil),
		fixtureNode("grand-first", "Grand first", 2, nil))
	last := fixtureNode("last", "Last", 1, nil)
	m := loadedModel(fixtureNode("root", "Root", 2, nil, last, first))
	assertRows := func(want []string) {
		t.Helper()
		lines := m.listLines(60, 40)[4:] // Toolbar and group heading.
		if len(lines) != len(want) {
			t.Fatalf("got %d rows, want %d: %q", len(lines), len(want), lines)
		}
		for i, line := range lines {
			if got := strings.TrimRight(ansi.Strip(line), " "); got != want[i] {
				t.Errorf("row %d: %q, want %q", i, got, want[i])
			}
		}
	}
	assertRows([]string{
		" >▾ ○ Root", "    │ Todo · Medium",
		"    ├─ ▾ ○ First", "    │  │   Todo · High",
		"    │  ├─ ○ Grand first", "    │  │    Todo · Medium",
		"    │  └─ ○ Grand last", "    │       Todo · Low",
		"    └─ ○ Last", "         Todo · Low",
	})
	m.collapse("first")
	m.prepareFrame()
	assertRows([]string{
		" >▾ ○ Root", "    │ Todo · Medium",
		"    ├─ ▸ ○ First", "    │      Todo · High",
		"    └─ ○ Last", "         Todo · Low",
	})
	// Only visible siblings determine connectors; a filtered-out sibling must
	// not leave a dangling continuation. Ancestors still retain context labels.
	m.filter.SearchTerm = "Grand first"
	m.prepareFrame()
	assertRows([]string{
		" >▾ [context] ○ Root", "    │ Todo · Medium",
		"    └─ ▾ [context] ○ First", "       │   Todo · High",
		"       └─ ○ Grand first", "            Todo · Medium",
	})
}

func TestHierarchy_ProgressCountsUseCompleteSnapshot(t *testing.T) {
	leaf := func(id string, progress int) *core.TaskNode {
		n := fixtureNode(id, id, 1, nil)
		n.Task.Progress = progress
		if progress == 100 {
			n.Task.Status = core.StatusDone
		}
		return n
	}
	parent := fixtureNode("parent", "Parent", 1, nil, leaf("done", 100), leaf("open", 0))
	parent.Task.Progress = 50
	root := fixtureNode("root", "Root", 1, nil, parent, leaf("other", 0), leaf("last", 0))
	root.Task.Progress = 16
	for _, tc := range []struct {
		name string
		node *core.TaskNode
		want string
	}{
		{"empty leaf", leaf("leaf", 0), "(0 / 1)"},
		{"manual leaf", leaf("leaf", 42), "(0.42 / 1)"},
		{"done leaf", leaf("leaf", 100), "(1 / 1)"},
		{"parent", parent, "(1 / 2)"},
		{"grandparent", root, "(0.5 / 3)"},
		{"complete open parent", fixtureNode("root", "Root", 1, nil, leaf("a", 100), leaf("b", 100)), "(2 / 2)"},
		{"hundredths", fixtureNode("root", "Root", 1, nil, leaf("a", 99), leaf("b", 99)), "(1.98 / 2)"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := loadedModel(tc.node)
			for _, profile := range []termenv.Profile{termenv.Ascii, termenv.TrueColor} {
				m.options.Profile = profile
				for _, width := range []int{42, 70, 114} {
					lines := m.detailLines(width)
					found := false
					for _, line := range lines {
						plain := ansi.Strip(line)
						if strings.Contains(plain, "Progress") && strings.Contains(plain, tc.want) {
							found = true
						}
						if ansi.StringWidth(line) > width {
							t.Fatalf("detail line overflows %d cells: %q", width, plain)
						}
					}
					if !found {
						t.Errorf("progress missing %s at width %d", tc.want, width)
					}
				}
			}
		})
	}
	m := loadedModel(root)
	m.collapse("root")
	m.prepareFrame()
	if !strings.Contains(m.View(), "(0.5 / 3)") {
		t.Fatal("collapse changed progress denominator")
	}
	m.filter.SearchTerm = "done"
	m.prepareFrame()
	if !strings.Contains(m.View(), "(0.5 / 3)") {
		t.Fatal("search changed progress denominator")
	}
	before, frame := snapshot(m), m.View()
	for range 100 {
		if m.View() != frame || snapshot(m) != before {
			t.Fatal("progress presentation made View impure")
		}
	}
}

func TestHierarchy_DeepTreeResizeAndScroll(t *testing.T) {
	root := fixtureNode("leaf", "Deep leaf 界", 1, nil)
	for i := range 9 {
		root = fixtureNode(strings.Repeat("a", i+1), "Ancestor", 2, nil, root)
	}
	m := loadedModel(root)
	for _, width := range []int{80, 120, 200} {
		m.Update(tea.WindowSizeMsg{Width: width, Height: 24})
		press(m, "end")
		if m.selectedTask().ID != "leaf" || !strings.Contains(m.View(), "Deep") {
			t.Fatal("deep selected title lost while scrolling")
		}
		for _, line := range strings.Split(m.View(), "\n") {
			if ansi.StringWidth(line) != width {
				t.Fatalf("frame width %d, want %d", ansi.StringWidth(line), width)
			}
		}
	}
}

func TestHierarchy_DescendantCompletionRefresh(t *testing.T) {
	m, svc, _ := mutationFixture(t, false)
	ctx := context.Background()
	create := func(title string, parent *core.Task) *core.Task {
		t.Helper()
		cmd := ports.CreateTaskCommand{Title: title}
		if parent != nil {
			cmd.ParentID = &parent.ID
		}
		task, err := svc.CreateTask(ctx, cmd)
		if err != nil {
			t.Fatal(err)
		}
		return task
	}
	root := create("Grandparent", nil)
	parent := create("First child", root)
	create("Second child", root)
	create("Third child", root)
	a := create("First grandchild", parent)
	b := create("Second grandchild", parent)
	check := func(rootCount, parentCount string, percent int) {
		t.Helper()
		deliverUI(m, m.requestRefresh())
		deliverUI(m, press(m, "home"))
		if m.selectedTask().Progress != percent || !strings.Contains(m.View(), rootCount) {
			t.Fatalf("grandparent missing %s / %d%%: %s", rootCount, percent, m.View())
		}
		deliverUI(m, press(m, "down"))
		if m.selectedTask().ID != parent.ID || !strings.Contains(m.View(), parentCount) {
			t.Fatalf("parent missing %s: %s", parentCount, m.View())
		}
	}
	check("(0 / 3)", "(0 / 2)", 0)
	var err error
	a, err = svc.CompleteTask(ctx, ports.TaskCommand{ID: a.ID, Base: a})
	if err != nil {
		t.Fatal(err)
	}
	check("(0.5 / 3)", "(1 / 2)", 16)
	if _, err = svc.CompleteTask(ctx, ports.TaskCommand{ID: b.ID, Base: b}); err != nil {
		t.Fatal(err)
	}
	check("(1 / 3)", "(2 / 2)", 33)
	status := core.StatusTodo
	if _, err = svc.UpdateTask(ctx, ports.UpdateTaskCommand{ID: a.ID, Base: a, Status: &status}); err != nil {
		t.Fatal(err)
	}
	check("(0.5 / 3)", "(1 / 2)", 16)
}
