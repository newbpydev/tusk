// Package app implements the main TUI application using the bubbletea framework.
// The app follows the Model-View-Update architecture pattern.
package app

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/messages"
)

// Update implements tea.Model Update, handling all message types.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Delegate all keyboard handling to the handler functions in handlers.go
		newModel, cmd := m.handleKeyPress(msg)
		return newModel, cmd

	case tea.WindowSizeMsg:
		// Update window dimensions for layout calculations
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case messages.TickMsg:
		// Update current time and check for status message expiry
		m.currentTime = time.Time(msg)
		if (!m.statusExpiry.IsZero()) && time.Now().After(m.statusExpiry) {
			m.statusMessage = ""
			m.statusType = ""
			m.statusExpiry = time.Time{}
		}
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return messages.TickMsg(t)
		})

	case messages.StatusUpdateErrorMsg:
		// Handle task update error
		m.err = msg.Err
		m.setErrorStatus(fmt.Sprintf("Error updating task '%s': %v", msg.TaskTitle, msg.Err))
		return m, m.refreshTasks()

	case messages.StatusUpdateSuccessMsg:
		// Handle successful task update
		m.setSuccessStatus(msg.Message)

		// Keep track of the updated task ID
		updatedTaskID := msg.Task.ID

		// Just update the task data in the main list without changing cursor position
		for i := range m.tasks {
			if m.tasks[i].ID == updatedTaskID {
				// Update the task with server data
				m.tasks[i] = msg.Task
				break
			}
		}

		// To ensure consistency, preserve the current cursor positions
		originalCursor := m.cursor
		originalVisualCursor := m.visualCursor
		originalCursorOnHeader := m.cursorOnHeader

		// Re-categorize tasks with updated data
		m.categorizeTasks(m.tasks)

		// Restore cursor positions
		m.cursor = originalCursor
		m.visualCursor = originalVisualCursor
		m.cursorOnHeader = originalCursorOnHeader

		// Refresh the visual cursor from task cursor to ensure consistency
		// This is important for cases where the task moves between sections
		m.updateVisualCursorFromTaskCursor()

		return m, nil

	case messages.TasksRefreshedMsg:
		// Handle refreshed task list
		m.tasks = msg.Tasks
		if m.cursor >= len(m.tasks) {
			m.cursor = max(0, len(m.tasks)-1)
		}
		m.clearLoadingStatus()
		m.initCollapsibleSections()
		return m, nil

	case messages.ErrorMsg:
		// Handle general error
		m.err = error(msg)
		m.setErrorStatus(fmt.Sprintf("Error: %v", error(msg)))
		return m, nil

	default:
		return m, nil
	}
}
