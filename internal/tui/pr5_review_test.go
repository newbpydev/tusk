package tui

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

func TestPR5_StaleFormSaveExplainsPause(t *testing.T) {
	m := loadedModel()
	press(m, "a")
	press(m, "Draft retained")
	m.stale = true
	formKey(m, tea.KeyCtrlS)
	if !strings.Contains(m.form.err, "Ctrl+R") || !strings.Contains(m.View(), "Ctrl+R") || m.saving {
		t.Fatalf("save pause hidden: error=%q saving=%v", m.form.err, m.saving)
	}
	if m.form.draft.fields[fieldTitle] != "Draft retained" {
		t.Fatal("draft lost")
	}
}

func TestPR5_DeleteAbandonedReadCanRefresh(t *testing.T) {
	for _, key := range []tea.KeyMsg{{Type: tea.KeyRunes, Runes: []rune("r")}, {Type: tea.KeyCtrlR}} {
		m, _ := deleteModel(true)
		m.options.Delete = func(context.Context, ports.DeleteTaskCommand) (ports.DeleteResult, error) {
			t.Fatal("replayed delete")
			return ports.DeleteResult{}, nil
		}
		deliverUI(m, press(m, "d"))
		press(m, "tab")
		formKey(m, tea.KeySpace)
		press(m, "tab")
		read := m.requestRefresh()
		press(m, "enter")
		reply := read().(forestMsg)
		reply.err = ports.ErrStorage
		m.Update(reply)
		if !m.stale || m.pendingMutation != nil || m.saving {
			t.Fatal("failed read did not abandon delete")
		}
		_, cmd := m.Update(key)
		if cmd == nil {
			t.Fatal("delete dialog cannot refresh")
		}
		if m.confirmation.preview != nil || m.confirmation.recursive || m.confirmation.field != 0 {
			t.Fatal("refresh retained old consent")
		}
		deliverUI(m, cmd)
		if m.stale || m.confirmation.preview == nil || m.confirmation.recursive {
			t.Fatal("fresh preview did not renew consent")
		}
	}
}

func TestPR5_FailedTogglePausesWritesUntilReadback(t *testing.T) {
	m := loadedModel(fixtureNode("task", "Task", core.PriorityMedium, nil))
	calls := 0
	m.options.Mutate = func(context.Context, mutationRequest) (*core.Task, error) { calls++; return nil, ports.ErrConflict }
	deliverUI(m, press(m, "x"))
	if !m.stale || m.canWrite() {
		t.Fatal("failed action still allows stale writes")
	}
	deliverUI(m, press(m, "x"))
	if calls != 1 {
		t.Fatal("stale toggle replayed")
	}
	deliverUI(m, press(m, "r"))
	if m.stale {
		t.Fatal("successful refresh left writes paused")
	}
}

func TestPR5_PageDownKeepsVisibleTaskOverlap(t *testing.T) {
	var nodes []*core.TaskNode
	for i := range 30 {
		n := fixtureNode(fmt.Sprint(i), fmt.Sprintf("Task %02d", i), core.PriorityMedium, nil)
		if i > 5 {
			n.Task.Status = core.StatusDone
		}
		nodes = append(nodes, n)
	}
	m := loadedModel(nodes...)
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	before := strings.Join(m.listLines(30, measure(m.width, m.height).bodyHeight-2), "\n")
	press(m, "pgdown")
	if !strings.Contains(before, m.selectedTask().Title) {
		t.Fatalf("page skipped visible overlap: %s", m.selectedTask().Title)
	}
	press(m, "pgup")
	if m.selected != 0 {
		t.Fatal("page up did not return to first page")
	}
}

func TestPR5_SearchWhitespaceAndAtomicLimits(t *testing.T) {
	for _, debounce := range []bool{false, true} {
		m := loadedModel()
		press(m, "/")
		cmd := press(m, "   ")
		if debounce {
			deliverUI(m, cmd)
		} else {
			press(m, "enter")
		}
		if m.filter.HasPredicates() {
			t.Fatal("whitespace search activates filters")
		}
	}
	m := loadedModel()
	press(m, "/")
	press(m, "ok")
	for _, raw := range []string{strings.Repeat("x", editorByteLimit), "bad\x1btext"} {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(raw), Paste: true})
		if m.searchDraft != "ok" {
			t.Fatal("search accepted oversized or unsupported paste")
		}
	}
	for _, field := range []int{2, 3} {
		m.beginFilters()
		m.filters.field = field
		m.filterKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ok")})
		for _, raw := range []string{strings.Repeat("x", editorByteLimit), "bad\x1btext"} {
			m.filterKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(raw), Paste: true})
			value := m.filters.tags
			if field == 3 {
				value = m.filters.due
			}
			if value != "ok" || m.filters.err == "" {
				t.Fatal("filter paste was not atomically refused")
			}
		}
	}
}

func TestPR5_EndOnlyFilterMatchesCollapsePresentation(t *testing.T) {
	child := fixtureNode("child", "Child", core.PriorityMedium, nil)
	root := fixtureNode("root", "Root", core.PriorityMedium, nil, child)
	end := testOptions().Now().Add(time.Hour)
	due := end
	root.Task.DueDate = &due
	child.Task.DueDate = &due
	end = end.Add(time.Hour)
	m := loadedModel(root)
	m.collapsed["root"] = true
	m.dueEnd = &end
	m.rebuildRows()
	m.prepareFrame()
	if len(m.rows) != 2 || strings.Contains(strings.Join(m.listLines(30, 20), "\n"), "▸") {
		t.Fatal("end-only filter renders collapsed marker")
	}
	delete(m.collapsed, "root")
	m.collapse("root")
	if m.collapsed["root"] {
		t.Fatal("collapse accepted during end-only filter")
	}
}

func TestPR5_ParentPickerWindowKeepsPairs(t *testing.T) {
	m := loadedModel()
	press(m, "a")
	p := &parentPicker{selected: 12}
	for i := range 20 {
		p.choices = append(p.choices, parentChoice{id: fmt.Sprintf("id%02d", i), title: fmt.Sprintf("Title%02d", i)})
	}
	m.form.picker = p
	for _, area := range []int{13, 15, 17} {
		_, lines, _, _ := m.formContent(60, area)
		if len(lines) == 0 || !strings.Contains(lines[0], "Title") {
			t.Fatalf("orphaned first ID: %q", lines)
		}
		if !strings.Contains(strings.Join(lines, "\n"), "id12") {
			t.Fatal("selected pair clipped")
		}
	}
}

func TestPR5_DetailsDoesNotRenderTemplateEditor(t *testing.T) {
	m := loadedModel(fixtureNode("one", "One", core.PriorityMedium, nil))
	m.notes.SetValue("TEMPLATE EDITOR")
	m.notes.Focus()
	if strings.Contains(strings.Join(m.detailLines(60), "\n"), "TEMPLATE EDITOR") {
		t.Fatal("details renders form template")
	}
}

func TestPR5_RecoveryZeroDueMatchesDetails(t *testing.T) {
	n := fixtureNode("one", "One", core.PriorityMedium, nil)
	zero := time.Time{}
	n.Task.DueDate = &zero
	m := loadedModel(n)
	m.freezeWrites()
	m.recovery.state = loaded
	m.recovery.target = taskIdentity(&n.Task)
	m.recovery.observed = &n.Task
	lines, _, _ := m.recoveryContent(60, 200)
	if strings.Contains(strings.Join(lines, "\n"), "Due: 0001-01-01") || !strings.Contains(strings.Join(lines, "\n"), "Due: Not set") {
		t.Fatalf("zero due presentation: %v", lines)
	}
	if strings.Contains(strings.Join(m.detailLines(60), "\n"), "Jan 1") {
		t.Fatal("zero due shown in summary")
	}
}
