package tui

import (
	"context"
	"maps"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type searchSnapshot struct {
	query, id string
	created   time.Time
	index     int
	collapsed map[string]bool
}
type searchMsg struct {
	token uint64
	query string
	err   error
}

func (m *Model) beginSearch() {
	m.searching = true
	m.searchDraft = m.filter.SearchTerm
	m.searchBefore = searchSnapshot{query: m.filter.SearchTerm, index: m.selected, collapsed: maps.Clone(m.collapsed)}
	if t := m.selectedTask(); t != nil {
		m.searchBefore.id = t.ID
		m.searchBefore.created = t.CreatedAt
	}
}

func (m *Model) searchKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyEsc:
		m.cancelSearchTimer()
		m.searching = false
		m.filter.SearchTerm = m.searchBefore.query
		m.collapsed = maps.Clone(m.searchBefore.collapsed)
		m.pruneCollapsed()
		m.rebuildRows()
		m.selected = max(-1, min(m.searchBefore.index, len(m.rows)-1))
		for i, row := range m.rows {
			if row.node.Task.ID == m.searchBefore.id && row.node.Task.CreatedAt.Equal(m.searchBefore.created) {
				m.selected = i
				break
			}
		}
		return nil
	case tea.KeyEnter:
		m.cancelSearchTimer()
		m.searching = false
		m.filter.SearchTerm = m.searchDraft
		return nil
	case tea.KeyBackspace:
		r := []rune(m.searchDraft)
		if len(r) > 0 {
			m.searchDraft = string(r[:len(r)-1])
		}
	case tea.KeySpace:
		m.searchDraft += " "
	case tea.KeyRunes:
		m.searchDraft += string(msg.Runes)
	default:
		return nil
	}
	m.cancelSearchTimer()
	ctx, cancel := context.WithCancel(m.options.Context)
	m.searchCancel = cancel
	token, query, wait := m.searchToken, m.searchDraft, m.options.Wait
	return safeCommand(func() tea.Msg { return searchMsg{token, query, wait(ctx, 150*time.Millisecond)} })
}

func (m *Model) cancelSearchTimer() {
	if m.searchCancel != nil {
		m.searchCancel()
		m.searchCancel = nil
	}
	m.searchToken++
}
