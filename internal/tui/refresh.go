package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/core"
	"time"
)

type tickMsg struct {
	token uint64
	now   time.Time
	err   error
}

func (m *Model) nextTick() tea.Cmd {
	m.tickToken++
	token, ctx, wait, now := m.tickToken, m.options.Context, m.options.Wait, m.options.Now
	return safeCommand(func() tea.Msg {
		err := wait(ctx, 2*time.Second)
		if err != nil {
			return tickMsg{token: token, err: err}
		}
		return tickMsg{token: token, now: now()}
	})
}

func (m *Model) requestRefresh() tea.Cmd {
	if m.recoveryNeeded {
		return nil
	}
	if m.busy {
		m.refreshPending = true
		return nil
	}
	m.busy = true
	m.operation++
	m.startRead()
	if m.state == loadFailed {
		m.state = loading
	}
	return m.Init()
}

func (m *Model) acceptForest(msg forestMsg) tea.Cmd {
	if msg.owner != m.owner || msg.operation != m.operation || !m.busy {
		return nil
	}
	// Resource completion and uncertainty precede display freshness checks.
	if cmd, handled := m.completeRead(msg.err); handled {
		return cmd
	}
	var timer tea.Cmd
	if msg.generation == m.generation {
		m.now = msg.now
		if msg.err != nil {
			if m.awaitingRead {
				m.notice = m.writeNotice() + "; refresh failed"
			}
			if m.form != nil && m.form.reloading {
				m.form.reloading = false
				m.form.err = "Reload failed. Draft retained; Ctrl+R retries."
			}
			if m.state == loaded {
				m.stale = true
			} else {
				m.state = loadFailed
			}
		} else {
			m.state = loaded
			m.stale = false
			m.forest = msg.forest
			if m.awaitingRead {
				m.notice = m.writeNotice()
			}
			m.awaitingRead = false
			m.reloadForm()
			m.historyRefresh = true
			m.pruneCollapsed()
			if !m.timerStarted {
				m.timerStarted = true
				timer = m.nextTick()
			}
		}
	}
	if cmd := m.dispatchPreview(); cmd != nil {
		return tea.Batch(timer, cmd)
	}
	if m.refreshPending {
		m.refreshPending = false
		return tea.Batch(timer, m.requestRefresh())
	}
	return timer
}

func (m *Model) pruneCollapsed() {
	present := make(map[string]bool)
	var walk func([]*core.TaskNode)
	walk = func(nodes []*core.TaskNode) {
		for _, node := range nodes {
			present[node.Task.ID] = true
			walk(node.Children)
		}
	}
	walk(m.forest)
	for id := range m.collapsed {
		if !present[id] {
			delete(m.collapsed, id)
		}
	}
}
