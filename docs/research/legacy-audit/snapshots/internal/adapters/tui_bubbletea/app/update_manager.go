// Package app implements the update manager that integrates the refactored
// message handling system with the existing application.
package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/handlers"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/messages"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/update"
)

// UpdateManager handles the message routing and delegation
// to the appropriate handlers based on message type.
type UpdateManager struct {
	// The original model
	model *Model
	
	// The model adapter that implements handler interfaces
	adapter *ModelAdapter
	
	// Message handlers map by message type
	msgHandlers map[string]func(tea.Msg) (tea.Model, tea.Cmd)
	
	// Keyboard handlers map by view mode and panel
	keyHandlers map[string]map[int]func(tea.KeyMsg) (tea.Model, tea.Cmd)
	
	// Message dispatcher for the new update system
	dispatcher *update.Dispatcher
}

// NewUpdateManager creates a new update manager for the application
func NewUpdateManager(model *Model) *UpdateManager {
	adapter := NewModelAdapter(model)
	
	// Create update manager
	mgr := &UpdateManager{
		model:        model,
		adapter:      adapter,
		msgHandlers:  make(map[string]func(tea.Msg) (tea.Model, tea.Cmd)),
		keyHandlers:  make(map[string]map[int]func(tea.KeyMsg) (tea.Model, tea.Cmd)),
	}
	
	// Initialize message handlers
	mgr.registerMessageHandlers()
	
	// Initialize keyboard handlers
	mgr.registerKeyboardHandlers()
	
	return mgr
}

// registerMessageHandlers initializes the message handlers map
func (m *UpdateManager) registerMessageHandlers() {
	m.msgHandlers["tea.WindowSizeMsg"] = m.handleWindowSize
	m.msgHandlers["messages.TickMsg"] = m.handleTick
	m.msgHandlers["messages.StatusUpdateErrorMsg"] = m.handleStatusUpdateError
	m.msgHandlers["messages.StatusUpdateSuccessMsg"] = m.handleStatusUpdateSuccess
	m.msgHandlers["messages.TasksRefreshedMsg"] = m.handleTasksRefreshed
	m.msgHandlers["messages.ErrorMsg"] = m.handleError
}

// registerKeyboardHandlers initializes the keyboard handlers map
func (m *UpdateManager) registerKeyboardHandlers() {
	// Initialize map for each view mode
	m.keyHandlers["list"] = make(map[int]func(tea.KeyMsg) (tea.Model, tea.Cmd))
	m.keyHandlers["detail"] = make(map[int]func(tea.KeyMsg) (tea.Model, tea.Cmd))
	m.keyHandlers["create"] = make(map[int]func(tea.KeyMsg) (tea.Model, tea.Cmd))
	m.keyHandlers["edit"] = make(map[int]func(tea.KeyMsg) (tea.Model, tea.Cmd))
	
	// Register handlers for list view by panel
	m.keyHandlers["list"][0] = m.handleTaskListPanelKeys
	m.keyHandlers["list"][1] = m.handleTaskDetailsPanelKeys
	m.keyHandlers["list"][2] = m.handleTimelinePanelKeys
	
	// Register handlers for detail view
	m.keyHandlers["detail"][0] = m.handleDetailViewKeys
	
	// Form handlers for create and edit modes
	formHandler := m.handleFormKeys
	m.keyHandlers["create"][0] = formHandler
	m.keyHandlers["edit"][0] = formHandler
}

// HandleUpdate implements the main bubbletea update function
// delegating to the appropriate handlers based on message type
func (m *UpdateManager) HandleUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Route based on message type
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle global key shortcuts first
		if m.handleGlobalKeys(msg) {
			return m.model, tea.Quit
		}
		
		// Route to view-specific handler based on current view mode and panel
		if handlers, ok := m.keyHandlers[m.model.viewMode]; ok {
			if handler, ok := handlers[m.model.activePanel]; ok {
				return handler(msg)
			}
		}
		
		// Default handler if no specific handler found
		return m.model, nil
		
	// Type assertions for other message types
	default:
		// Get the type name for message lookup
		msgType := getMsgTypeName(msg)
		
		// Find and execute handler if registered
		if handler, ok := m.msgHandlers[msgType]; ok {
			return handler(msg)
		}
		
		// Default case if no handler found
		return m.model, nil
	}
}

// getMsgTypeName returns the type name as a string for the message
func getMsgTypeName(msg tea.Msg) string {
	// In a real implementation, use reflection to get the type name
	switch msg.(type) {
	case tea.WindowSizeMsg:
		return "tea.WindowSizeMsg"
	case messages.TickMsg:
		return "messages.TickMsg"
	case messages.StatusUpdateErrorMsg:
		return "messages.StatusUpdateErrorMsg"
	case messages.StatusUpdateSuccessMsg:
		return "messages.StatusUpdateSuccessMsg"
	case messages.TasksRefreshedMsg:
		return "messages.TasksRefreshedMsg"
	case messages.ErrorMsg:
		return "messages.ErrorMsg"
	default:
		return "unknown"
	}
}

// handleGlobalKeys processes global keyboard shortcuts
func (m *UpdateManager) handleGlobalKeys(msg tea.KeyMsg) bool {
	switch msg.String() {
	case "q", "ctrl+c":
		return true
	default:
		return false
	}
}

// Individual message handlers

// handleWindowSize processes window size change messages
func (m *UpdateManager) handleWindowSize(msg tea.Msg) (tea.Model, tea.Cmd) {
	sizeMsg := msg.(tea.WindowSizeMsg)
	m.model.width = sizeMsg.Width
	m.model.height = sizeMsg.Height
	return m.model, nil
}

// handleTick processes tick messages
func (m *UpdateManager) handleTick(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Ensure we update current time using the current system time to prevent drift
	m.model.currentTime = time.Now()
	
	// Check and clear expired status messages
	if (!m.model.statusExpiry.IsZero()) && time.Now().After(m.model.statusExpiry) {
		// Clear the status message fields
		m.model.statusMessage = ""
		m.model.statusType = ""
		m.model.statusExpiry = time.Time{}
		
		// Debug log for status message expiry
		// fmt.Println("Status message expired and cleared")
	}
	
	// CRITICAL: Always return the tick command to ensure clock keeps updating
	return m.model, tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return messages.TickMsg(t)
	})
}

// handleStatusUpdateError processes task update error messages
func (m *UpdateManager) handleStatusUpdateError(msg tea.Msg) (tea.Model, tea.Cmd) {
	errorMsg := msg.(messages.StatusUpdateErrorMsg)
	m.model.err = errorMsg.Err
	m.model.setErrorStatus(
		"Error updating task '" + errorMsg.TaskTitle + "': " + errorMsg.Err.Error(),
	)
	return m.model, m.model.refreshTasks()
}

// handleStatusUpdateSuccess processes successful task update messages
func (m *UpdateManager) handleStatusUpdateSuccess(msg tea.Msg) (tea.Model, tea.Cmd) {
	updateMsg := msg.(messages.StatusUpdateSuccessMsg)
	
	// Set status message
	m.model.setSuccessStatus(updateMsg.Message)
	
	// Keep track of the updated task ID
	updatedTaskID := updateMsg.Task.ID
	
	// Update the task in the main list
	for i := range m.model.tasks {
		if m.model.tasks[i].ID == updatedTaskID {
			m.model.tasks[i] = updateMsg.Task
			break
		}
	}
	
	// Re-categorize tasks 
	m.model.categorizeTasks(m.model.tasks)
	m.model.overdueTasks, m.model.todayTasks, m.model.upcomingTasks = 
		m.model.categorizeTimelineTasks(m.model.tasks)
	
	// Re-initialize sections
	m.model.initCollapsibleSections()
	m.model.initTimelineCollapsibleSections()
	
	return m.model, nil
}

// handleTasksRefreshed processes task list refresh messages
func (m *UpdateManager) handleTasksRefreshed(msg tea.Msg) (tea.Model, tea.Cmd) {
	refreshMsg := msg.(messages.TasksRefreshedMsg)
	
	// Update task list
	m.model.tasks = refreshMsg.Tasks
	
	// Ensure cursor is within bounds
	if m.model.cursor >= len(m.model.tasks) {
		m.model.cursor = max(0, len(m.model.tasks)-1)
	}
	
	// Clear loading status
	m.model.clearLoadingStatus()
	
	// Re-initialize sections
	m.model.initCollapsibleSections()
	m.model.initTimelineCollapsibleSections()
	
	return m.model, nil
}

// handleError processes general error messages
func (m *UpdateManager) handleError(msg tea.Msg) (tea.Model, tea.Cmd) {
	errorMsg := msg.(messages.ErrorMsg)
	m.model.err = error(errorMsg)
	m.model.setErrorStatus("Error: " + error(errorMsg).Error())
	return m.model, nil
}

// handleTaskListPanelKeys processes keys in the task list panel
func (m *UpdateManager) handleTaskListPanelKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Delegate to refactored handlers
	cmd, handled := handlers.HandleTaskListPanelKeys(m.adapter, msg)
	if handled {
		return m.model, cmd
	}
	
	// Try panel visibility handlers as fallback
	cmd, handled = handlers.HandlePanelVisibilityKeys(m.adapter, msg)
	if handled {
		return m.model, cmd
	}
	
	// No handler found
	return m.model, nil
}

// handleTaskDetailsPanelKeys processes keys in the task details panel
func (m *UpdateManager) handleTaskDetailsPanelKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Delegate to refactored handlers
	cmd, handled := handlers.HandleDetailsPanelKeys(m.adapter, msg)
	if handled {
		return m.model, cmd
	}
	
	// Try panel visibility handlers as fallback
	cmd, handled = handlers.HandlePanelVisibilityKeys(m.adapter, msg)
	if handled {
		return m.model, cmd
	}
	
	// No handler found
	return m.model, nil
}

// handleTimelinePanelKeys processes keys in the timeline panel
func (m *UpdateManager) handleTimelinePanelKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Delegate to refactored handlers
	cmd, handled := handlers.HandleTimelinePanelKeys(m.adapter, msg)
	if handled {
		return m.model, cmd
	}
	
	// Try panel visibility handlers as fallback
	cmd, handled = handlers.HandlePanelVisibilityKeys(m.adapter, msg)
	if handled {
		return m.model, cmd
	}
	
	// No handler found
	return m.model, nil
}

// handleDetailViewKeys processes keys in the detail view
func (m *UpdateManager) handleDetailViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Legacy function, delegate to model directly for now
	return m.model.handleDetailViewKeys(msg)
}

// handleFormKeys processes keys in the form view
func (m *UpdateManager) handleFormKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Delegate to refactored handlers
	cmd, handled := handlers.HandleFormKeys(m.adapter, msg)
	if handled {
		return m.model, cmd
	}
	
	// No handler found
	return m.model, nil
}

// Use max from utils.go
