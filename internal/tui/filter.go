package tui

import (
	"slices"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/terminaltext"
)

type filterDraft struct {
	statuses       [4]bool
	priorities     [4]bool
	field, choice  int
	tags, due, err string
}

func filterStatuses() []core.Status {
	return []core.Status{core.StatusTodo, core.StatusInProgress, core.StatusBlocked, core.StatusDone}
}
func filterPriorities() []core.Priority {
	return []core.Priority{core.PriorityLow, core.PriorityMedium, core.PriorityHigh, core.PriorityUrgent}
}

func (m *Model) beginFilters() {
	f := &filterDraft{due: m.dueExpression}
	for i, s := range filterStatuses() {
		f.statuses[i] = slices.Contains(m.filter.Statuses, s)
	}
	for i, p := range filterPriorities() {
		f.priorities[i] = slices.Contains(m.filter.Priorities, p)
	}
	tags := make([]string, len(m.filter.Tags))
	for i, t := range m.filter.Tags {
		tags[i] = string(t)
	}
	f.tags = strings.Join(tags, ", ")
	m.filters = f
}

func (m *Model) clearFilters() {
	m.workspaceTab = 0
	m.filter = core.TaskFilter{}
	m.dueStart = nil
	m.dueEnd = nil
	m.dueExpression = ""
	m.dueLabel = ""
}

func (m *Model) chooseWorkspaceTab(tab int) {
	query := m.filter.SearchTerm
	m.clearFilters()
	m.filter.SearchTerm = query
	m.workspaceTab = tab
	if tab == 1 {
		m.dueLabel = "Today and overdue"
		m.filter.Statuses = []core.Status{core.StatusTodo, core.StatusInProgress, core.StatusBlocked}
	} else if tab == 2 {
		m.filter.Statuses = []core.Status{core.StatusDone}
	}
	m.listOffset = 0
}

func (m *Model) updateTodayBounds() {
	if m.workspaceTab != 1 {
		return
	}
	local := m.now.In(m.options.Location)
	end := time.Date(local.Year(), local.Month(), local.Day()+1, 0, 0, 0, 0, m.options.Location)
	m.dueEnd = &end
}

func (m *Model) applyFilters() {
	f := m.filters
	m.now = m.options.Now()
	next := core.TaskFilter{SearchTerm: m.filter.SearchTerm}
	for i, s := range filterStatuses() {
		if f.statuses[i] {
			next.Statuses = append(next.Statuses, s)
		}
	}
	for i, p := range filterPriorities() {
		if f.priorities[i] {
			next.Priorities = append(next.Priorities, p)
		}
	}
	if strings.TrimSpace(f.tags) != "" {
		tags, err := core.NormalizeTags(strings.Split(f.tags, ","))
		if err != nil {
			f.err = "Tags: use comma-separated words, such as work, design."
			f.field = 2
			return
		}
		next.Tags = tags
	}
	if strings.TrimSpace(f.due) != "" {
		if m.options.DayBounds == nil {
			f.err = "Due date is unavailable."
			f.field = 3
			return
		}
		start, end, err := m.options.DayBounds(f.due, m.now, m.options.Location)
		if err != nil {
			f.err = "Due: use today, tomorrow or YYYY-MM-DD."
			f.field = 3
			return
		}
		m.dueStart = &start
		m.dueEnd = &end
		m.dueLabel = start.In(m.options.Location).Format("2006-01-02")
	} else {
		m.dueStart = nil
		m.dueEnd = nil
		m.dueLabel = ""
	}
	m.dueExpression = f.due
	m.filter = next
	// Enter the custom tab only when the applied draft actually constrains
	// the view; an empty apply is equivalent to Clear all, so the header must
	// not report filters that do not exist.
	if next.HasPredicates() || m.dueStart != nil {
		m.workspaceTab = -1
	} else {
		m.workspaceTab = 0
	}
	m.filters = nil
}

func (m *Model) filterKey(msg tea.KeyMsg) {
	f := m.filters
	switch msg.Type {
	case tea.KeyEsc:
		m.filters = nil
		return
	case tea.KeyTab, tea.KeyDown:
		f.field = (f.field + 1) % 6
		f.choice = 0
		return
	case tea.KeyShiftTab, tea.KeyUp:
		f.field = (f.field + 5) % 6
		f.choice = 0
		return
	case tea.KeyCtrlS:
		m.applyFilters()
		return
	case tea.KeyEnter:
		if f.field == 4 {
			m.applyFilters()
		} else if f.field == 5 {
			m.clearFilters()
			m.filters = nil
		} else {
			f.field++
			f.choice = 0
		}
		return
	}
	if f.field < 2 {
		switch msg.String() {
		case "left":
			f.choice = (f.choice + 3) % 4
		case "right":
			f.choice = (f.choice + 1) % 4
		case " ":
			if f.field == 0 {
				f.statuses[f.choice] = !f.statuses[f.choice]
			} else {
				f.priorities[f.choice] = !f.priorities[f.choice]
			}
		}
		return
	}
	if f.field > 3 {
		return
	}
	value := &f.tags
	if f.field == 3 {
		value = &f.due
	}
	switch msg.Type {
	case tea.KeyBackspace:
		r := []rune(*value)
		if len(r) > 0 {
			*value = string(r[:len(r)-1])
		}
	case tea.KeySpace:
		if len(*value) >= editorByteLimit {
			f.err = "Input is too large."
			return
		}
		*value += " "
	case tea.KeyRunes:
		if len(*value)+len(string(msg.Runes)) > editorByteLimit || !editableText(string(msg.Runes), false) {
			f.err = "Input is too large or contains unsupported controls."
			return
		}
		*value += string(msg.Runes)
	}
}

// Render only a suffix when text exceeds the field, keeping the insertion point
// visible. The underlying raw value is never clipped or rewritten.
func inputLine(value string, active bool, width int) string {
	value = terminaltext.Scalar(value)
	if active {
		value += "▎"
	}
	if n := ansi.StringWidth(value); n > width {
		value = "…" + cellSlice(value, n-width+1, n)
	}
	return fitCells(value, width)
}

func (m *Model) filterLines(width int) []string {
	f := m.filters
	label := func(field int, s string) string {
		if f.field == field {
			return m.paint("> "+s, accentColor, true)
		}
		return "  " + s
	}
	choices := func(field int, names []string, values [4]bool) string {
		var labels []string
		for i, name := range names {
			s := m.checkbox(name, values[i], f.field == field && f.choice == i)
			labels = append(labels, s)
		}
		return "  " + strings.Join(labels, "  ")
	}
	return []string{
		"Choose any statuses and priorities; empty means all.", "",
		label(0, "Status"), choices(0, []string{"Todo", "In progress", "Blocked", "Done"}, f.statuses), "",
		label(1, "Priority"), choices(1, []string{"Low", "Medium", "High", "Urgent"}, f.priorities), "",
		label(2, "Tags · all must match"), "  │ " + inputLine(f.tags, f.field == 2, width-4), "",
		label(3, "Due day · today, tomorrow or YYYY-MM-DD"), "  │ " + inputLine(f.due, f.field == 3, width-4),
		m.paint(f.err, accentColor, false),
	}
}
