// Package handlers contains keyboard input handlers for the TUI application.
// These handlers are grouped by functional area to improve code organization.
package handlers

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/hooks"
	"github.com/newbpydev/tusk/internal/core/task"
)

// AppModel defines the interface that the TUI Model must implement
// to work with these handlers.
type AppModel interface {
	// Navigation methods
	NavigateDown()
	NavigateUp()
	NavigateToTop()
	NavigateToBottom()
	
	// Task operations
	ToggleTaskCompletion() tea.Cmd
	RefreshTasks() tea.Cmd
	ResetForm()
	LoadTaskIntoForm(task task.Task)
	
	// Status updates
	SetLoadingStatus(msg string)
	SetStatusMessage(msg, statusType string, duration time.Duration)
	
	// Section management
	ToggleSection() tea.Cmd
	InitCollapsibleSections()
	
	// Getters and state accessors
	GetCollapsibleManager() *hooks.CollapsibleManager
	GetCursor() int
	GetCursorOnHeader() bool
	GetTasks() []task.Task
	
	// Setters and state modifiers
	SetActivePanel(panel int)
	SetViewMode(mode string)
	SetFormPriority(priority string)
}

// HandleTaskListPanelKeys processes keyboard input when the task list panel is active
func HandleTaskListPanelKeys(m AppModel, msg tea.KeyMsg) (tea.Cmd, bool) {
	// Initialize sections if needed
	if m.GetCollapsibleManager() == nil {
		m.InitCollapsibleSections()
	}

	switch msg.String() {
	case "j", "down":
		// Handle down navigation through tasks and section headers
		if m.GetCollapsibleManager().GetItemCount() > 0 {
			m.NavigateDown()
			return nil, true
		}
		return nil, true

	case "k", "up":
		// Handle up navigation through tasks and section headers
		if m.GetCollapsibleManager().GetItemCount() > 0 {
			m.NavigateUp()
			return nil, true
		}
		return nil, true

	case "g":
		// Jump to top
		m.NavigateToTop()
		return nil, true

	case "G":
		// Jump to bottom
		m.NavigateToBottom()
		return nil, true

	case "tab", "right", "l":
		// Move to next panel if available
		// TODO: This references the model's showTaskDetails directly
		// We should add a method to the AppModel interface to handle this
		m.SetActivePanel(1)
		return nil, true

	case "enter", "d":
		// If on a section header, toggle expansion
		if m.GetCursorOnHeader() {
			return m.ToggleSection(), true
		}
		// If on a task, show details (if available)
		// TODO: This references the model's showTaskDetails directly
		// We should add a method to the AppModel interface to handle this
		m.SetActivePanel(1)
		return nil, true

	case " ":
		// Toggle task completion status
		if !m.GetCursorOnHeader() && m.GetCursor() < len(m.GetTasks()) {
			return m.ToggleTaskCompletion(), true
		}
		return nil, true

	case "n":
		// Create new task
		m.ResetForm()
		m.SetViewMode("create")
		m.SetFormPriority(string(task.PriorityLow)) // Set default priority
		return nil, true

	case "e":
		// Edit task
		if !m.GetCursorOnHeader() && m.GetCursor() < len(m.GetTasks()) {
			m.SetViewMode("edit")
			// Load current task into form
			m.LoadTaskIntoForm(m.GetTasks()[m.GetCursor()])
			return nil, true
		}
		return nil, true

	case "r":
		// Refresh tasks
		m.SetLoadingStatus("Refreshing tasks...")
		return m.RefreshTasks(), true
	}

	return nil, false
}
