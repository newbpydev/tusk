package tui

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/core"
)

func formKey(m *Model, key tea.KeyType) tea.Cmd {
	_, cmd := m.Update(tea.KeyMsg{Type: key})
	return cmd
}

func TestForm_ReadonlyPreviewPreparationIsBounded(t *testing.T) {
	n := fixtureNode("id", "Read-only preview", 2, nil)
	raw := "prefix\t界 " + strings.Repeat("large note\n", 100000)
	n.Task.Description = raw
	m := loadedModel(n)
	press(m, "e")
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	_, lines, _, _ := m.formContent(68, 16)
	runtime.ReadMemStats(&after)
	if after.TotalAlloc-before.TotalAlloc > 128*1024 {
		t.Fatalf("one-line read-only preview allocated %d bytes", after.TotalAlloc-before.TotalAlloc)
	}
	if !strings.Contains(strings.Join(lines, "\n"), `prefix\t界`) || m.form.draft.fields[fieldNotes] != raw {
		t.Fatal("preview or raw text lost")
	}
}

func TestForm_FocusLiteralTextAndDiscard(t *testing.T) {
	m := loadedModel()
	press(m, "a")
	if m.form == nil {
		t.Fatal("create did not open")
	}
	press(m, "q d?x ")
	if m.form.draft.fields[fieldTitle] != "q d?x " {
		t.Fatal("browse keys consumed text")
	}
	press(m, "tab")
	press(m, "First")
	press(m, "enter")
	press(m, "第二")
	if m.form.draft.fields[fieldNotes] != "First\n第二" {
		t.Fatalf("notes %q", m.form.draft.fields[fieldNotes])
	}
	formKey(m, tea.KeyShiftTab)
	if m.form.field != fieldTitle {
		t.Fatal("backward focus")
	}
	formKey(m, tea.KeyShiftTab)
	if m.form.field != 7 {
		t.Fatal("focus did not wrap to cancel")
	}
	press(m, "esc")
	if m.form.prompt != promptDiscard || m.form.confirm {
		t.Fatal("dirty cancel must default keep")
	}
	press(m, "enter")
	if m.form == nil || m.form.prompt != promptNone {
		t.Fatal("default discarded")
	}
	press(m, "esc")
	press(m, "tab")
	press(m, "enter")
	if m.form != nil {
		t.Fatal("explicit discard")
	}
	press(m, "a")
	press(m, "esc")
	if m.form != nil {
		t.Fatal("clean close")
	}
}

func TestForm_AtomicPasteAndReadonlyReplacement(t *testing.T) {
	m := loadedModel()
	press(m, "a")
	press(m, "Title")
	press(m, "tab")
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a\n世界"), Paste: true})
	original := m.form.draft.fields[fieldNotes]
	for _, value := range []string{"bad\tvalue", "bad\x1bvalue", "bad\u202evalue", "bad\ufffdvalue", strings.Repeat("x", editorByteLimit), strings.Repeat("\n", 10000)} {
		_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(value), Paste: true})
		if m.form.draft.fields[fieldNotes] != original || m.form.err == "" {
			t.Fatal("paste changed or silently truncated draft")
		}
	}
	for _, raw := range []string{"raw\tnotes", "\x1b[31mnotes", strings.Repeat("a", editorByteLimit+1), strings.Repeat("\n", 10000)} {
		n := fixtureNode("id", "Title", 2, nil)
		n.Task.Description = raw
		m = loadedModel(n)
		press(m, "e")
		if !m.form.readonly[fieldNotes] || m.form.draft.fields[fieldNotes] != raw {
			t.Fatal("raw overwritten")
		}
		press(m, "tab")
		formKey(m, tea.KeyCtrlU)
		press(m, "x")
		if m.form.draft.fields[fieldNotes] != raw {
			t.Fatal("readonly changed")
		}
		formKey(m, tea.KeyCtrlE)
		if m.form.prompt != promptReplace || m.form.confirm {
			t.Fatal("replace default")
		}
		press(m, "enter")
		if m.form.draft.fields[fieldNotes] != raw {
			t.Fatal("default replaced")
		}
		formKey(m, tea.KeyCtrlE)
		press(m, "tab")
		press(m, "enter")
		if m.form.readonly[fieldNotes] || m.form.draft.fields[fieldNotes] != "" || m.form.draft.original[fieldNotes] != raw {
			t.Fatal("replacement lost original")
		}
	}
}

func TestForm_ParentPickerSnapshotAndPurity(t *testing.T) {
	child := fixtureNode("child", "Child", 2, nil)
	root := fixtureNode("root", "Root", 3, nil, child)
	other := fixtureNode("other", "Other", 1, nil)
	m := loadedModel(root, other)
	m.focus = detailsFocus
	press(m, "e")
	for range 5 {
		press(m, "tab")
	}
	formKey(m, tea.KeyCtrlP)
	if m.form.picker == nil || len(m.form.picker.choices) != 2 {
		t.Fatal("self/descendants in picker")
	}
	press(m, "Other")
	press(m, "down")
	press(m, "enter")
	if m.form.draft.fields[fieldParent] != "other" {
		t.Fatal("parent not exact ID")
	}
	before := m.form.draft.base.Clone()
	root.Task.Title = "external"
	m.prepareFrame()
	if m.form.draft.base.Title != before.Title {
		t.Fatal("refresh rebased draft")
	}
	for _, size := range []tea.WindowSizeMsg{{Width: 80, Height: 24}, {Width: 120, Height: 40}, {Width: 79, Height: 23}, {Width: 200, Height: 60}} {
		m.Update(size)
		frame := m.View()
		for range 3 {
			if m.View() != frame {
				t.Fatal("impure View")
			}
		}
	}
	press(m, "esc")
	press(m, "tab")
	press(m, "enter")
	if m.focus != detailsFocus {
		t.Fatal("focus not restored")
	}
	done := fixtureNode("done", "Done", 1, nil)
	done.Task.Status = core.StatusDone
	m = loadedModel(done)
	press(m, "e")
	if !m.form.readonly[fieldProgress] {
		t.Fatal("done progress editable")
	}
}

func TestForm_SmallTerminalPreservesDraftAndSubmission(t *testing.T) {
	m := loadedModel(fixtureNode("first", "First", 2, nil), fixtureNode("second", "Second", 2, nil))
	selectTask(t, m, "second")
	press(m, "a")
	press(m, "Draft")
	before := m.form.draft.fields
	for _, size := range [][2]int{{0, 0}, {1, 1}, {79, 24}, {80, 23}, {-9, -4}, {79, 23}} {
		m.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		press(m, "q")
		press(m, "ctrl+s")
		press(m, "esc")
		press(m, "tab")
		if m.form == nil || m.form.draft.fields != before || m.form.field != fieldTitle || m.form.prompt != promptNone {
			t.Fatal("hidden form accepted input")
		}
		m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
		if m.selectedTask().ID != "second" || m.form.draft.fields != before {
			t.Fatal("resize lost selection or draft")
		}
	}
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	press(m, "restored")
	if m.form.draft.fields[fieldTitle] != "Draftrestored" {
		t.Fatal("restore lost draft")
	}
}

func TestForm_ControlsValidationAndParentSearch(t *testing.T) {
	m := loadedModel(fixtureNode("parent", "Parent", 2, nil))
	press(m, "a")
	press(m, "ctrl+s")
	if m.form.field != fieldTitle || m.form.err != "Enter a title." {
		t.Fatal("validation focus")
	}
	press(m, "Title")
	press(m, "ctrl+s")
	if m.form.err != "Saving is unavailable." {
		t.Fatal("missing writer")
	}
	press(m, "enter")
	press(m, "tab")
	press(m, "left")
	press(m, "right")
	formKey(m, tea.KeySpace)
	press(m, "up")
	press(m, "unused")
	if m.form.draft.fields[fieldPriority] != "medium" {
		t.Fatal("priority control")
	}
	press(m, "tab")
	press(m, "tomorrow")
	formKey(m, tea.KeyCtrlU)
	if m.form.draft.fields[fieldDue] != "" {
		t.Fatal("clear due")
	}
	press(m, "tab")
	press(m, "work")
	formKey(m, tea.KeyCtrlU)
	press(m, "tab")
	formKey(m, tea.KeyCtrlP)
	press(m, "Missing")
	press(m, "down")
	press(m, "up")
	press(m, "backspace")
	press(m, "esc")
	if m.form.picker != nil {
		t.Fatal("picker escaped form")
	}
	formKey(m, tea.KeyCtrlP)
	press(m, "down")
	press(m, "enter")
	formKey(m, tea.KeyCtrlU)
	if m.form.draft.fields[fieldParent] != "" {
		t.Fatal("root clear")
	}
	press(m, "tab")
	press(m, "x")
	press(m, "tab")
	press(m, "enter")
	press(m, "esc")
	if m.form == nil || m.form.prompt != promptNone {
		t.Fatal("cancel button discarded dirty draft")
	}
	formKey(m, tea.KeyCtrlR)
	press(m, "esc")
	press(m, "right")
	press(m, "enter")
	if m.form != nil {
		t.Fatal("discard")
	}
	press(m, "e")
	for range fieldStatus {
		press(m, "tab")
	}
	press(m, "right")
	if m.form.draft.fields[fieldStatus] != "in-progress" {
		t.Fatal("status control")
	}
}

func TestForm_VisiblePositionAndTimezone(t *testing.T) {
	m := loadedModel()
	m.options.Location = time.FixedZone("UTC-03", -3*3600)
	press(m, "a")
	if !strings.Contains(m.View(), "Create task · 1/6") || !strings.Contains(m.View(), "Due date · e.g. 2026-10-15 or tomorrow") || !strings.Contains(m.View(), "Ctrl+P calendar · Ctrl+U clear · UTC-03") {
		t.Fatal("form lacks position or due timezone")
	}
	for range 5 {
		press(m, "tab")
	}
	if !strings.Contains(m.View(), "Create task · 6/6") || !strings.Contains(m.View(), "Parent · Ctrl+P") {
		t.Fatal("hidden fields not discoverable")
	}
}

func TestForm_BrowseHelpAndQuitRemainVisible(t *testing.T) {
	m := loadedModel()
	m.notice = "Saved"
	m.prepareFrame()
	footer := strings.Split(m.View(), "\n")[23]
	if !strings.Contains(footer, "? help") || !strings.Contains(footer, "q quit") {
		t.Fatal("essential navigation hidden", footer)
	}
}

func TestForm_ParentSearchAcceptsSpaces(t *testing.T) {
	m := loadedModel(fixtureNode("id", "Customer research", 2, nil))
	press(m, "a")
	for range 5 {
		press(m, "tab")
	}
	formKey(m, tea.KeyCtrlP)
	press(m, "Customer")
	formKey(m, tea.KeySpace)
	press(m, "research")
	if m.form.picker.query != "Customer research" || len(m.form.picker.matches()) != 2 {
		t.Fatal("space omitted from parent search")
	}
}

func TestForm_ParentPickerWindowsLargeMatchLists(t *testing.T) {
	var nodes []*core.TaskNode
	for i := range 40 {
		nodes = append(nodes, fixtureNode(fmt.Sprintf("id-%02d", i), fmt.Sprintf("Task %02d", i), core.PriorityMedium, nil))
	}
	m := loadedModel(nodes...)
	press(m, "a")
	for range 5 {
		press(m, "tab")
	}
	formKey(m, tea.KeyCtrlP)
	if m.form.picker == nil {
		t.Fatal("picker not opened")
	}
	press(m, "Task")
	if matches := m.form.picker.matches(); len(matches) != 41 {
		t.Fatalf("match count %d, want Root plus 40 tasks", len(matches))
	}
	for range 30 {
		press(m, "down")
	}
	frame := m.View()
	if !strings.Contains(frame, "id-29") {
		t.Fatal("selected match scrolled outside the visible window")
	}
	if strings.Contains(frame, "id-00") {
		t.Fatal("off-screen match rendered inside the modal")
	}
	press(m, "enter")
	if m.form.picker != nil || m.form.draft.fields[fieldParent] != "id-29" {
		t.Fatalf("enter chose %q", m.form.draft.fields[fieldParent])
	}
}

// A deep selection must materialize only the visible window: the picker hands
// the overlay a pre-trimmed content slice at offset 0, so allocation follows
// the modal area instead of the selection depth.
func TestForm_ParentPickerDeepSelectionBuildsOnlyVisibleWindow(t *testing.T) {
	var nodes []*core.TaskNode
	for i := range 40 {
		nodes = append(nodes, fixtureNode(fmt.Sprintf("id-%02d", i), fmt.Sprintf("Task %02d", i), core.PriorityMedium, nil))
	}
	m := loadedModel(nodes...)
	press(m, "a")
	for range 5 {
		press(m, "tab")
	}
	formKey(m, tea.KeyCtrlP)
	press(m, "Task")
	for range 40 {
		press(m, "down")
	}
	const area = 18
	_, content, offset, _ := m.formContent(74, area)
	if offset != 0 {
		t.Fatalf("picker returned scroll offset %d; the window must be pre-trimmed", offset)
	}
	if len(content) > area {
		t.Fatalf("deep selection materialized %d lines for a %d-line window", len(content), area)
	}
	joined := strings.Join(content, "\n")
	if !strings.Contains(joined, "id-39") || !strings.Contains(joined, "Task 39") {
		t.Fatal("selected match outside the built window")
	}
	if strings.Contains(joined, "id-00") {
		t.Fatal("off-window match materialized")
	}
	press(m, "enter")
	if m.form.picker != nil || m.form.draft.fields[fieldParent] != "id-39" {
		t.Fatalf("enter chose %q", m.form.draft.fields[fieldParent])
	}
}
