package tui

import (
	"context"
	"errors"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

type historyState struct {
	key     taskRef
	token   uint64
	pending bool
	state   loadState
	events  []ports.TaskEvent
}
type historyMsg struct {
	owner, operation, token uint64
	key                     taskRef
	events                  []ports.TaskEvent
	err                     error
}

func (s *Session) History(ctx context.Context, id string) ([]ports.TaskEvent, error) {
	value, err := s.Call(ctx, false, func(ctx context.Context, svc ports.TaskService) (any, error) { return svc.GetTaskHistory(ctx, id) })
	if err != nil {
		return nil, err
	}
	return value.([]ports.TaskEvent), nil
}

func (m *Model) dispatchHistory() tea.Cmd {
	if m.busy || m.recoveryNeeded || !m.history.pending {
		return nil
	}
	if m.refreshPending {
		m.refreshPending = false
		return m.requestRefresh()
	}
	m.busy = true
	m.operation++
	m.history.pending = false
	owner, operation, token, key := m.owner, m.operation, m.history.token, m.history.key
	ctx, load := m.options.Context, m.options.History
	return safeCommand(func() tea.Msg {
		events, err := load(ctx, key.id)
		return historyMsg{owner, operation, token, key, events, err}
	})
}

func (m *Model) acceptHistory(msg historyMsg) tea.Cmd {
	if msg.owner != m.owner || msg.operation != m.operation || !m.busy {
		return nil
	}
	m.busy = false
	if IsUnknown(msg.err) {
		m.recoveryNeeded = true
		m.stale = true
		m.refreshPending = false
		m.history.pending = false
		return nil
	}
	if errors.Is(msg.err, errRuntime) {
		m.exitErr = errRuntime
		return tea.Quit
	}
	if msg.key == m.history.key && msg.token == m.history.token {
		if msg.err != nil {
			m.history.state = loadFailed
			if errors.Is(msg.err, core.ErrTaskNotFound) {
				m.refreshPending = true
			}
		} else {
			m.history.state = loaded
			m.history.events = append([]ports.TaskEvent(nil), msg.events...)
			slices.SortStableFunc(m.history.events, func(a, b ports.TaskEvent) int {
				if a.Sequence < b.Sequence {
					return -1
				}
				if a.Sequence > b.Sequence {
					return 1
				}
				return 0
			})
		}
	}
	if m.refreshPending {
		m.refreshPending = false
		return m.requestRefresh()
	}
	return nil
}
