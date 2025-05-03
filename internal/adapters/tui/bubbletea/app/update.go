// Package app implements the main TUI application using the bubbletea framework.
// The app follows the Model-View-Update architecture pattern.
package app

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/shared"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/messages"
	"github.com/newbpydev/tusk/internal/core/task"
)

// Update implements tea.Model Update, handling all message types.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// If modal is visible, handle ESC key globally to close it
	if m.showModal {
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "esc" {
			m.showModal = false
			return m, nil
		}
		
		// If modal is visible, pass messages to modal unless it's a special modal message
		switch msgType := msg.(type) {
		case messages.HideModalMsg:
			m.showModal = false
			return m, nil
		case shared.ModalCloseMsg:
			m.showModal = false
			return m, nil
		case messages.ShowModalMsg, tea.WindowSizeMsg:
			// These should be handled by the main update flow
			// They're special cases even when a modal is visible
		default:
			_ = msgType // Avoid unused variable warning
			// When modal is visible, pass all other messages to the modal
			var cmd tea.Cmd
			m.modal, cmd = m.modal.Update(msg)
			return m, cmd
		}
	}
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Delegate all keyboard handling to the handler functions in handlers.go
		newModel, cmd := m.handleKeyPress(msg)
		return newModel, cmd

	case tea.WindowSizeMsg:
		// Call our enhanced window resize handler to ensure cursor visibility
		// This is critical for preventing the cursor from going offscreen during resize
		m.handleWindowResize(msg)
		return m, nil
	
	case messages.ShowModalMsg:
		// Show a modal with the provided content
		m.modal = shared.NewModal(msg.Content, msg.Width, msg.Height)
		m.modal.Show()
		m.showModal = true
		return m, nil
		
	case messages.HideModalMsg:
		// Hide the modal
		m.showModal = false
		return m, nil

	case messages.TickMsg:
		// Update current time with the CURRENT system time, not the msg time
		// This ensures we always display the latest time and prevents drift
		m.currentTime = time.Now()
		if (!m.statusExpiry.IsZero()) && time.Now().After(m.statusExpiry) {
			m.statusMessage = ""
			m.statusType = ""
			m.statusExpiry = time.Time{}
		}
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			// Pass the time but force refresh on receipt
			return messages.TickMsg(t)
		})

	case messages.StatusUpdateErrorMsg:
		// Handle task update error
		m.err = msg.Err
		m.setErrorStatus(fmt.Sprintf("Error updating task '%s': %v", msg.TaskTitle, msg.Err))
		return m, m.refreshTasks()
		
	case shared.SampleModalMsg:
		// Handle sample modal action
		if msg.Action == "ok" {
			// Modal OK button was clicked - implement your action here
			m.showModal = false
			m.setSuccessStatus("Modal action completed successfully!")
			return m, nil
		}
		// For any other action, just return the model as is
		return m, nil

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
		
		// Store the current timeline cursor positions
		originalTimelineCursor := m.timelineCursor
		originalTimelineCursorOnHeader := m.timelineCursorOnHeader

		// Re-categorize tasks with updated data
		m.categorizeTasks(m.tasks)

		// Also update timeline categories to ensure proper timeline display
		// This is critical when toggling task completion status
		m.overdueTasks, m.todayTasks, m.upcomingTasks = m.categorizeTimelineTasks(m.tasks)
		
		// Fully reinitialize timeline sections to ensure proper cursor mapping
		// This fixes the bug where unchecked tasks aren't selectable in timeline
		m.initTimelineCollapsibleSections()

		// Restore cursor positions
		m.cursor = originalCursor
		m.visualCursor = originalVisualCursor
		m.cursorOnHeader = originalCursorOnHeader
		
		// If we were in the timeline and the task status changed, 
		// ensure the cursor state is properly reset for the updated sections
		if m.activePanel == 2 { // Panel 2 is the timeline
			// When a task is unchecked, we need to reset the timeline cursor
			// to ensure it can be selected in its new section
			if msg.Task.Status == task.StatusTodo && m.timelineCursor != 0 {
				// If it's a task being unchecked, find it in the timeline section
				m.resetTimelineCursorForTask(updatedTaskID)
			} else {
				// Otherwise restore original cursor position
				m.timelineCursor = originalTimelineCursor
				m.timelineCursorOnHeader = originalTimelineCursorOnHeader
			}
		}

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

		// Also initialize timeline sections to ensure timeline view is up-to-date
		m.initTimelineCollapsibleSections()
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
