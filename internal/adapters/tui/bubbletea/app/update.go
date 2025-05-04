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
	// If modal is visible, we need to handle some messages specially
	if m.showModal {
		// Handle Escape key globally to close modal
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "esc" {
			m.showModal = false
			return m, nil
		}
		
		// Handle special message types
		switch msgType := msg.(type) {
		// Messages that should close the modal
		case messages.HideModalMsg, shared.HideModalMessage, shared.ModalCloseMsg:
			m.showModal = false
			return m, nil
			
		// Time tick messages - allow these to update the main app first, then the modal
		case messages.TickMsg:
			// Update current time in the main app
			m.currentTime = time.Now() 
			// Check status expiry
			if (!m.statusExpiry.IsZero()) && time.Now().After(m.statusExpiry) {
				m.statusMessage = ""
				m.statusType = ""
				m.statusExpiry = time.Time{}
			}
			
			// Also pass the tick to the modal, but discard any commands it returns
			m.modal, _ = m.modal.Update(msg)
			
			// Keep ticking with our own timer command
			return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
				return messages.TickMsg(t)
			})
			
		// Messages that should be handled by the main update method
		case messages.ShowModalMsg, tea.WindowSizeMsg:
			// These fall through to the main switch statement
			_ = msgType // Avoid unused variable warning
			
		// All other messages go to the modal
		default:
			_ = msgType // Avoid unused variable warning
			// Pass message to modal
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
		// Show a modal with the provided content - use the display mode from the message
		m.modal = shared.NewModal(msg.Content, msg.Width, msg.Height, msg.DisplayMode)
		m.modal.Show()
		m.showModal = true
		
		// The help footer will remain visible because of how we've updated the modal implementation
		// The modal.View method now correctly preserves the header and footer
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
