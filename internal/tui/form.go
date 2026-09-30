package tui

import (
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/terminaltext"
)

const editorByteLimit = 64 * 1024

type formPrompt uint8

const (
	promptNone formPrompt = iota
	promptDiscard
	promptReplace
	promptConflict
)

type parentChoice struct{ id, title string }
type parentPicker struct {
	choices  []parentChoice
	query    string
	selected int
}
type taskForm struct {
	draft                                *taskDraft
	inputs                               [fieldCount]textinput.Model
	notes                                textarea.Model
	readonly                             [fieldCount]bool
	previews                             [fieldCount]string
	field, offset                        int
	origin                               panelFocus
	err                                  string
	prompt                               formPrompt
	confirm, saving, conflict, reloading bool
	picker                               *parentPicker
	calendar                             *datePicker
}

func (f *taskForm) count() int {
	if f.draft.base == nil {
		return fieldStatus
	}
	return fieldCount
}

func editableText(value string, notes bool) bool {
	if !utf8.ValidString(value) || len(value) > editorByteLimit || strings.Count(value, "\n") >= 10000 {
		return false
	}
	if notes {
		return terminaltext.Multiline(value) == value
	}
	return terminaltext.Scalar(value) == value
}

func (m *Model) beginForm(edit bool) {
	if !m.canWrite() {
		return
	}
	var base *core.Task
	children := false
	if edit {
		base = m.selectedTask()
		if base == nil {
			return
		}
		children = len(m.rows[m.selected].node.Children) > 0
	}
	m.formSequence++
	f := &taskForm{draft: newTaskDraft(base, children, m.options.Location), origin: m.focus}
	f.draft.id = m.formSequence
	f.notes = m.notes
	f.notes.CharLimit = 0
	f.notes.MaxHeight = 0
	f.notes.ShowLineNumbers = false
	f.notes.Prompt = ""
	f.notes.SetHeight(3)
	for i := range f.inputs {
		input := textinput.New()
		input.Prompt = ""
		input.CharLimit = 0
		input.Cursor.SetMode(cursor.CursorStatic)
		input.KeyMap.Paste.SetEnabled(false)
		plain := m.renderer.NewStyle()
		input.TextStyle = plain
		input.PromptStyle = plain
		input.PlaceholderStyle = plain
		input.Cursor.Style = plain
		input.Cursor.TextStyle = plain
		f.inputs[i] = input
		m.loadFormField(f, i)
	}
	f.readonly[fieldProgress] = f.draft.progressLocked
	m.form = f
	m.focusForm(fieldTitle)
}

func (m *Model) loadFormField(f *taskForm, i int) {
	value := f.draft.fields[i]
	// Only one display row is shown for unsupported text. Prepare it once;
	// the complete raw value remains in the draft for omission/replacement.
	f.previews[i] = titleCells(terminaltext.Scalar(value), modalMaxWidth)
	f.readonly[i] = !editableText(value, i == fieldNotes)
	if f.readonly[i] {
		return
	}
	if i == fieldNotes {
		f.notes.SetValue(value)
		f.readonly[i] = f.notes.Value() != value
	} else {
		f.inputs[i].SetValue(value)
		f.readonly[i] = f.inputs[i].Value() != value
	}
}

func (m *Model) focusForm(field int) {
	f := m.form
	f.field = (field + f.count() + 2) % (f.count() + 2)
	f.notes.Blur()
	for i := range f.inputs {
		f.inputs[i].Blur()
	}
	if f.field < f.count() && !f.readonly[f.field] {
		if f.field == fieldNotes {
			f.notes.Focus()
		} else {
			f.inputs[f.field].Focus()
		}
	}
}

func (m *Model) closeForm() { m.focus = m.form.origin; m.form = nil }
func (m *Model) cancelForm() {
	if m.form.draft.dirty() {
		m.form.prompt = promptDiscard
		m.form.confirm = false
	} else {
		m.closeForm()
	}
}

func (m *Model) formKey(key tea.KeyMsg) tea.Cmd {
	f := m.form
	if f.saving || f.reloading || m.recoveryNeeded {
		return nil
	}
	if f.picker != nil {
		m.parentKey(key)
		return nil
	}
	if f.calendar != nil {
		m.calendarKey(key)
		return nil
	}
	if f.prompt != promptNone {
		switch key.String() {
		case "tab", "shift+tab", "left", "right":
			f.confirm = !f.confirm
		case "esc":
			f.prompt = promptNone
			f.confirm = false
		case "enter":
			prompt, confirm := f.prompt, f.confirm
			f.prompt = promptNone
			f.confirm = false
			if !confirm {
				return nil
			}
			switch prompt {
			case promptDiscard:
				m.closeForm()
			case promptReplace:
				f.draft.fields[f.field] = ""
				m.loadFormField(f, f.field)
				m.focusForm(f.field)
			case promptConflict:
				f.reloading = true
				return m.requestRefresh()
			}
		}
		return nil
	}
	switch key.String() {
	case "esc":
		m.cancelForm()
		return nil
	case "ctrl+r":
		if f.conflict {
			f.prompt = promptConflict
			f.confirm = false
			return nil
		}
		return m.requestRefresh()
	case "ctrl+s":
		return m.submitForm()
	case "tab":
		m.focusForm(f.field + 1)
		return nil
	case "shift+tab":
		m.focusForm(f.field - 1)
		return nil
	case "enter":
		if f.field == f.count() {
			return m.submitForm()
		}
		if f.field == f.count()+1 {
			m.cancelForm()
			return nil
		}
		if f.field != fieldNotes {
			m.focusForm(f.field + 1)
			return nil
		}
	}
	if f.field >= f.count() || f.conflict {
		return nil
	}
	i := f.field
	if key.String() == "ctrl+e" && f.readonly[i] && !(i == fieldProgress && f.draft.progressLocked) {
		f.prompt = promptReplace
		f.confirm = false
		return nil
	}
	if f.readonly[i] {
		return nil
	}
	if key.String() == "ctrl+p" && i == fieldParent {
		m.beginParentPicker()
		return nil
	}
	if key.String() == "ctrl+p" && i == fieldDue {
		m.beginCalendar()
		return nil
	}
	if key.String() == "ctrl+u" && (i == fieldDue || i == fieldTags || i == fieldParent) {
		f.draft.fields[i] = ""
		f.inputs[i].SetValue("")
		return nil
	}
	if i == fieldPriority || i == fieldStatus {
		values := []string{"low", "medium", "high", "urgent"}
		if i == fieldStatus {
			values = []string{"todo", "in-progress", "blocked", "done"}
		}
		index := slices.Index(values, f.draft.fields[i])
		switch key.String() {
		case "right", "down", " ":
			index = (index + 1) % len(values)
		case "left", "up":
			index = (index + len(values) - 1) % len(values)
		default:
			return nil
		}
		f.draft.fields[i] = values[index]
		return nil
	}
	incoming := ""
	if key.Type == tea.KeyRunes {
		incoming = string(key.Runes)
	} else if key.Type == tea.KeySpace {
		incoming = " "
	} else if key.Type == tea.KeyEnter {
		incoming = "\n"
	}
	old := f.draft.fields[i]
	if incoming != "" && (!editableText(incoming, i == fieldNotes) || strings.ContainsRune(incoming, utf8.RuneError) || len(old)+len(incoming) > editorByteLimit || (i == fieldNotes && strings.Count(old, "\n")+strings.Count(incoming, "\n") >= 10000)) {
		f.err = "Input rejected: unsupported text or editor capacity exceeded."
		return nil
	}
	f.err = ""
	if i == fieldNotes {
		f.notes, _ = f.notes.Update(key)
		f.draft.fields[i] = f.notes.Value()
	} else {
		f.inputs[i], _ = f.inputs[i].Update(key)
		f.draft.fields[i] = f.inputs[i].Value()
	}
	return nil
}

func (m *Model) beginParentPicker() {
	f := m.form
	p := &parentPicker{choices: []parentChoice{{title: "Root task (no parent)"}}}
	var walk func([]*core.TaskNode)
	walk = func(nodes []*core.TaskNode) {
		for _, n := range nodes {
			if f.draft.base != nil && n.Task.ID == f.draft.base.ID {
				continue
			}
			p.choices = append(p.choices, parentChoice{n.Task.ID, n.Task.Title})
			walk(n.Children)
		}
	}
	walk(m.forest)
	f.picker = p
}
func (p *parentPicker) matches() []parentChoice {
	result := []parentChoice{p.choices[0]}
	query := strings.ToLower(p.query)
	for _, v := range p.choices[1:] {
		if strings.Contains(strings.ToLower(v.id+" "+v.title), query) {
			result = append(result, v)
		}
	}
	return result
}
func (m *Model) parentKey(key tea.KeyMsg) {
	f := m.form
	p := f.picker
	switch key.String() {
	case "esc":
		f.picker = nil
		return
	case "enter":
		v := p.matches()[p.selected]
		f.draft.fields[fieldParent] = v.id
		f.inputs[fieldParent].SetValue(v.id)
		f.picker = nil
		return
	case "down", "tab":
		p.selected++
	case "up", "shift+tab":
		p.selected--
	case "backspace":
		r := []rune(p.query)
		if len(r) > 0 {
			p.query = string(r[:len(r)-1])
		}
		p.selected = 0
	case " ":
		if len(p.query) < 255 {
			p.query += " "
			p.selected = 0
		}
	default:
		if key.Type == tea.KeyRunes && editableText(string(key.Runes), false) && len(p.query)+len(string(key.Runes)) <= 255 {
			p.query += string(key.Runes)
			p.selected = 0
		}
	}
	p.selected = max(0, min(p.selected, len(p.matches())-1))
}

func (m *Model) formContent(width, area int) (string, []string, int, string) {
	f := m.form
	if f.calendar != nil {
		return "Choose due date", m.calendarLines(width), 0,
			m.actionButton("Choose date  Enter", true, true, false) + "  " + m.actionButton("Cancel  Esc", false, false, false)
	}
	title := "Create task"
	if f.draft.base != nil {
		title = "Edit task"
	}
	if f.field < f.count() {
		title += fmt.Sprintf(" · %d/%d", f.field+1, f.count())
	}
	if f.prompt != promptNone {
		message, keep, accept := "Discard your changes?", "Keep editing", "Discard"
		if f.prompt == promptReplace {
			message = "Replace this field? Its stored text will be removed only on save."
			keep = "Cancel"
			accept = "Replace"
		}
		if f.prompt == promptConflict {
			message = "Reload the current task and discard this draft?"
			keep = "Keep draft"
			accept = "Reload task"
		}
		controls := m.actionButton(keep, !f.confirm, false, false) + "  " + m.actionButton(accept, f.confirm, false, false)
		return title, []string{"", message, "", "Tab chooses · Enter confirms"}, 0, controls
	}
	if p := f.picker; p != nil {
		lines := []string{"Search: " + terminaltext.Scalar(p.query) + "▎", ""}
		for i, v := range p.matches() {
			prefix := "  "
			if i == p.selected {
				prefix = "> "
			}
			lines = append(lines, prefix+terminaltext.Scalar(v.title), "    "+terminaltext.Scalar(v.id))
		}
		offset := max(0, 2+p.selected*2-area+2)
		return "Choose parent", lines, offset, "↑↓ choose · Enter select · Esc cancel"
	}
	labels := []string{"Title", "Notes · Enter adds a line", "Priority · ←→ choose", "Due date · e.g. 2026-10-15 or tomorrow", "Tags · comma-separated", "Parent · Ctrl+P choose · Ctrl+U root", "Status · ←→ choose", "Progress · 0–99 for open leaves"}
	lines := []string{}
	start, end := 0, 0
	for i := 0; i < f.count(); i++ {
		if i == f.field {
			start = len(lines)
		}
		prefix := "  "
		if i == f.field {
			prefix = "> "
		}
		label := prefix + labels[i]
		if f.readonly[i] {
			label += " · read-only"
		}
		color := mutedColor
		if i == f.field {
			color = accentColor
		}
		lines = append(lines, m.paint(label, color, i == f.field))
		rule := m.paint("  │ ", color, false)
		switch {
		case f.readonly[i]:
			preview := titleCells(f.previews[i], width-4)
			lines = append(lines, rule+preview)
			if i == fieldProgress {
				lines = append(lines, "  Derived from status or subtasks.")
			} else {
				lines = append(lines, "  Preserved exactly · Ctrl+E replace")
			}
		case i == fieldNotes:
			f.notes.SetWidth(max(1, width-4))
			for _, line := range strings.Split(f.notes.View(), "\n") {
				lines = append(lines, rule+line)
			}
		case i == fieldPriority || i == fieldStatus:
			value := f.draft.fields[i]
			if i == fieldPriority {
				priority, _ := core.ParsePriority(value)
				value = priorityLabel(priority)
			} else {
				_, value = statusText(core.Status(value))
			}
			lines = append(lines, rule+"‹ "+value+" ›")
		default:
			f.inputs[i].Width = max(1, width-4)
			lines = append(lines, rule+f.inputs[i].View())
		}
		if i == fieldDue && !f.readonly[i] {
			lines = append(lines, m.paint("  Ctrl+P calendar · Ctrl+U clear · "+m.options.Location.String(), mutedColor, false))
		}
		if i == f.field {
			end = len(lines)
		}
		lines = append(lines, "")
	}
	if f.field < f.count() {
		if start < f.offset {
			f.offset = start
		}
		if end > f.offset+area {
			f.offset = end - area
		}
	}
	f.offset = max(0, min(f.offset, max(0, len(lines)-area)))
	controls := m.actionButton("Save task  Ctrl+S", f.field == f.count(), true, false) + "  " + m.actionButton("Cancel  Esc", f.field == f.count()+1, false, false)
	if f.saving {
		controls = "Saving… · Ctrl+C exits safely"
	} else if f.reloading {
		controls = "Reloading task…"
	} else if f.conflict {
		controls = "Draft retained · Ctrl+R reload · Esc discard"
	}
	return title, lines, f.offset, controls
}
