package tui

import (
	"context"
	"maps"
	"strings"
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

// selectionRestore defers a selection restoration to the next rebuildRows, so
// a caller that restores view state can leave the projection to finish's
// single rebuild instead of projecting twice in one Update.
type selectionRestore struct {
	ref   taskRef
	index int
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
		// finish owns the single rebuild for this Update; hand it the
		// pre-search selection so rebuildRows restores by ID exactly as the
		// dismissal always has, without projecting the forest twice.
		m.restore = &selectionRestore{ref: taskRef{m.searchBefore.id, m.searchBefore.created.UTC()}, index: m.searchBefore.index}
		return nil
	case tea.KeyEnter:
		m.cancelSearchTimer()
		m.searching = false
		m.filter.SearchTerm = strings.TrimSpace(m.searchDraft)
		return nil
	case tea.KeyBackspace:
		r := []rune(m.searchDraft)
		if len(r) > 0 {
			m.searchDraft = string(r[:len(r)-1])
		}
	case tea.KeySpace:
		if len(m.searchDraft) >= editorByteLimit {
			m.notice = "Search input is too large."
			return nil
		}
		m.searchDraft += " "
	case tea.KeyRunes:
		if len(m.searchDraft)+len(string(msg.Runes)) > editorByteLimit || !editableText(string(msg.Runes), false) {
			m.notice = "Search input is too large or contains unsupported controls."
			return nil
		}
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
