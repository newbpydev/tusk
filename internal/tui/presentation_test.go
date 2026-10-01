package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/newbpydev/tusk/internal/core"
)

func TestWorkspace_PersistentControlsAndQuickViews(t *testing.T) {
	now := testOptions().Now()
	yesterday := now.Add(-24 * time.Hour)
	tomorrow := now.Add(24 * time.Hour)
	done := fixtureNode("done", "Finished", core.PriorityLow, nil)
	done.Task.Status = core.StatusDone
	m := loadedModel(fixtureNode("today", "Today", core.PriorityMedium, &now), fixtureNode("overdue", "Overdue", core.PriorityHigh, &yesterday), fixtureNode("later", "Later", core.PriorityLow, &tomorrow), done)
	for _, hint := range []string{"Search tasks", "1 All", "2 Today", "3 Done", "a New task"} {
		if !strings.Contains(ansi.Strip(m.View()), hint) {
			t.Errorf("missing persistent control %q", hint)
		}
	}
	if !strings.Contains(strings.Join(helpLines(), "\n"), "1 / 2 / 3") {
		t.Error("quick views missing from keyboard help")
	}
	press(m, "2")
	if len(m.rows) != 2 {
		t.Fatal("Today must show due/overdue open tasks", len(m.rows))
	}
	press(m, "3")
	if len(m.rows) != 1 || m.selectedTask().ID != "done" {
		t.Fatal("Done view")
	}
	press(m, "1")
	if len(m.rows) != 4 {
		t.Fatal("All view")
	}
	press(m, "end")
	for _, hint := range []string{"Search tasks", "1 All", "2 Today", "3 Done"} {
		if !strings.Contains(strings.Join(strings.Split(ansi.Strip(m.View()), "\n")[:5], "\n"), hint) {
			t.Fatal("toolbar scrolled away", hint)
		}
	}
	press(m, "a")
	press(m, "123")
	if m.form.draft.fields[fieldTitle] != "123" {
		t.Fatal("tabs stole form text")
	}
}

func TestPresentation_ModalInsetsAndFieldAlignment(t *testing.T) {
	m := loadedModel()
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	press(m, "a")
	press(m, "A title")
	press(m, "tab")
	press(m, "Notes text")
	l := measure(m.width, m.height)
	rows := strings.Split(ansi.Strip(m.View()), "\n")
	if strings.TrimSpace(cellSlice(rows[l.modal.y+1], l.modal.x+1, l.modal.x+l.modal.width-1)) != "" {
		t.Fatal("modal needs top padding")
	}
	if strings.TrimSpace(cellSlice(rows[l.modal.y+l.modal.height-2], l.modal.x+1, l.modal.x+l.modal.width-1)) != "" {
		t.Fatal("modal needs bottom padding")
	}
	for _, value := range []string{"│ A title", "│ Notes text", "[ Save task"} {
		if !strings.Contains(ansi.Strip(m.View()), value) {
			t.Error("missing styled field/control", value)
		}
	}
	press(m, "esc")
	press(m, "tab")
	press(m, "enter")
	press(m, "f")
	formKey(m, tea.KeySpace)
	if !strings.Contains(ansi.Strip(m.View()), "[✓] Todo") {
		t.Fatal("checkbox does not show a clear checked state")
	}
	filter := m.filterLines(66)
	for i, line := range filter {
		if strings.TrimSpace(ansi.Strip(line)) == "Priority" && (i == 0 || strings.TrimSpace(filter[i-1]) != "") {
			t.Error("checkbox groups need a blank row")
		}
	}
}

func TestWorkspace_TodayRollsOverAtLocalMidnight(t *testing.T) {
	now := testOptions().Now()
	next := now.AddDate(0, 0, 1)
	m := loadedModel(fixtureNode("next", "Next day", core.PriorityLow, &next))
	press(m, "2")
	if len(m.rows) != 0 {
		t.Fatal("tomorrow included too early")
	}
	m.Update(tickMsg{token: m.tickToken, now: next})
	if len(m.rows) != 1 {
		t.Fatal("Today did not advance at local midnight")
	}
}

func TestPresentation_DetailsPrioritizeTaskContent(t *testing.T) {
	m := loadedModel(fixtureNode("task-id", "Readable task", core.PriorityHigh, nil))
	text := ansi.Strip(strings.Join(m.detailLines(70), "\n"))
	if strings.Index(text, "Notes") > strings.Index(text, "ID ") {
		t.Fatal("technical metadata precedes task content")
	}
	for _, value := range []string{"○ Todo", "High", "Progress"} {
		if !strings.Contains(text, value) {
			t.Error("missing readable task detail", value)
		}
	}
}
