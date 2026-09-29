package tui

import (
	"github.com/newbpydev/tusk/internal/core"
)

func (m *Model) selectedTask() *core.Task {
	if m.selected < 0 || m.selected >= len(m.rows) {
		return nil
	}
	return &m.rows[m.selected].node.Task
}

func (m *Model) rebuildRows() {
	previous, index := m.selectedTask(), m.selected
	m.rows = project(m.forest, m.filter, m.dueStart, m.dueEnd, m.collapsed, m.now, m.options.Location)
	m.selected = -1
	if len(m.rows) == 0 {
		m.listOffset = 0
		return
	}
	m.selected = max(0, min(index, len(m.rows)-1))
	if previous != nil {
		for i, row := range m.rows {
			if row.node.Task.ID == previous.ID && row.node.Task.CreatedAt.Equal(previous.CreatedAt) {
				m.selected = i
				break
			}
		}
	}
}

func (m *Model) collapse(id string) {
	if m.filter.HasPredicates() || m.dueStart != nil {
		return
	}
	for i, row := range m.rows {
		if row.node.Task.ID != id || len(row.node.Children) == 0 {
			continue
		}
		for j := i + 1; j < len(m.rows) && m.rows[j].depth > row.depth; j++ {
			if j == m.selected {
				m.selected = i
				break
			}
		}
		m.collapsed[id] = true
		return
	}
}

func (m *Model) navigate(k string) {
	page := max(1, (measure(m.width, m.height).bodyHeight-2)/2)
	if m.focus == detailsFocus {
		switch k {
		case "j", "down", "k", "up", "pgdown", "pgup", "g", "home":
			m.detailsEnd = false
		}
		switch k {
		case "j", "down":
			m.detailsScroll++
		case "k", "up":
			m.detailsScroll--
		case "pgdown":
			m.detailsScroll += page * 2
		case "pgup":
			m.detailsScroll -= page * 2
		case "g", "home":
			m.detailsScroll = 0
		case "G", "end":
			m.detailsEnd = true
			m.detailsScroll = int(^uint(0) >> 1)
		}
		return
	}
	if len(m.rows) == 0 {
		return
	}
	old := m.selected
	switch k {
	case "j", "down":
		m.selected++
	case "k", "up":
		m.selected--
	case "g", "home":
		m.selected = 0
	case "G", "end":
		m.selected = len(m.rows) - 1
	case "pgdown":
		m.selected += page
	case "pgup":
		m.selected -= page
	case "h", "left":
		m.collapse(m.rows[m.selected].node.Task.ID)
	case "l", "right":
		if !m.filter.HasPredicates() && m.dueStart == nil {
			delete(m.collapsed, m.rows[m.selected].node.Task.ID)
		}
	}
	m.selected = max(0, min(m.selected, len(m.rows)-1))
	if old != m.selected {
		m.detailsScroll = 0
	}
}
