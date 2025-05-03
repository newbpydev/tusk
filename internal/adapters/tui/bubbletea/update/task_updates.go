package update

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/messages"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/services"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/state"
	"github.com/newbpydev/tusk/internal/core/task"
)

// TaskUpdateHandler processes task update messages
type TaskUpdateHandler struct {
	categorizer *services.TaskCategorizationService
}

// NewTaskUpdateHandler creates a new task update handler
func NewTaskUpdateHandler(categorizer *services.TaskCategorizationService) *TaskUpdateHandler {
	return &TaskUpdateHandler{
		categorizer: categorizer,
	}
}

// CanHandle checks if this handler can process the message
func (h *TaskUpdateHandler) CanHandle(msg tea.Msg) bool {
	_, ok1 := msg.(messages.StatusUpdateSuccessMsg)
	_, ok2 := msg.(messages.StatusUpdateErrorMsg)
	_, ok3 := msg.(messages.TasksRefreshedMsg)
	
	return ok1 || ok2 || ok3
}

// HandleMessage processes task-related messages
func (h *TaskUpdateHandler) HandleMessage(msg tea.Msg, appState *state.AppState) (*state.AppState, tea.Cmd) {
	switch typedMsg := msg.(type) {
	case messages.StatusUpdateErrorMsg:
		return h.handleStatusUpdateError(typedMsg, appState)
	case messages.StatusUpdateSuccessMsg:
		return h.handleStatusUpdateSuccess(typedMsg, appState)
	case messages.TasksRefreshedMsg:
		return h.handleTasksRefreshed(typedMsg, appState)
	default:
		return appState, nil
	}
}

// handleStatusUpdateError processes task update error messages
func (h *TaskUpdateHandler) handleStatusUpdateError(msg messages.StatusUpdateErrorMsg, appState *state.AppState) (*state.AppState, tea.Cmd) {
	// Update error state
	appState.Error = msg.Err
	
	// Set error status message
	appState = appState.SetStatusMessage(
		fmt.Sprintf("Error updating task '%s': %v", msg.TaskTitle, msg.Err),
		"error",
		5*time.Second,
	)
	
	// Refresh tasks to ensure UI is in sync with backend
	// This would be a command that fetches tasks
	return appState, nil
}

// handleStatusUpdateSuccess processes successful task update messages
func (h *TaskUpdateHandler) handleStatusUpdateSuccess(msg messages.StatusUpdateSuccessMsg, appState *state.AppState) (*state.AppState, tea.Cmd) {
	// Set success status message
	appState = appState.SetStatusMessage(msg.Message, "success", 3*time.Second)
	
	// Keep track of the updated task ID
	updatedTaskID := msg.Task.ID
	
	// Store the current cursor positions before recategorization
	originalCursor := appState.Cursor
	originalVisualCursor := appState.VisualCursor
	originalCursorOnHeader := appState.CursorOnHeader
	originalTimelineCursor := appState.TimelineCursor
	originalTimelineCursorOnHeader := appState.TimelineCursorOnHeader
	
	// Update the task in the main list
	for i := range appState.Tasks {
		if appState.Tasks[i].ID == updatedTaskID {
			appState.Tasks[i] = msg.Task
			break
		}
	}
	
	// Recategorize tasks with updated data
	todoTasks, projectTasks, completedTasks := h.categorizer.CategorizeTasks(appState.Tasks)
	appState.TodoTasks = todoTasks
	appState.ProjectTasks = projectTasks
	appState.CompletedTasks = completedTasks
	
	// Update timeline categories
	overdueTasks, todayTasks, upcomingTasks := h.categorizer.CategorizeTimelineTasks(appState.Tasks)
	appState.OverdueTasks = overdueTasks
	appState.TodayTasks = todayTasks
	appState.UpcomingTasks = upcomingTasks
	
	// TODO: Reinitialize collapsible sections
	// This would require passing the CollapsibleManager to this handler
	// or adding methods to the appState to handle this
	
	// Restore cursor positions
	appState.Cursor = originalCursor
	appState.VisualCursor = originalVisualCursor
	appState.CursorOnHeader = originalCursorOnHeader
	
	// Restore timeline cursor positions with special handling for status changes
	if appState.ActivePanel == 2 { // Timeline panel
		if msg.Task.Status == task.StatusTodo && appState.TimelineCursor != 0 {
			// Need special handling for unchecked tasks in timeline
			// This would be moved to a method in appState
		} else {
			appState.TimelineCursor = originalTimelineCursor
			appState.TimelineCursorOnHeader = originalTimelineCursorOnHeader
		}
	}
	
	// TODO: Update visual cursor based on task cursor
	// This would be a method in appState
	
	return appState, nil
}

// handleTasksRefreshed processes task list refresh messages
func (h *TaskUpdateHandler) handleTasksRefreshed(msg messages.TasksRefreshedMsg, appState *state.AppState) (*state.AppState, tea.Cmd) {
	// Update task list
	appState.Tasks = msg.Tasks
	
	// Ensure cursor is within bounds
	if appState.Cursor >= len(appState.Tasks) {
		appState.Cursor = max(0, len(appState.Tasks)-1)
	}
	
	// Clear loading status
	appState = appState.ClearLoadingStatus()
	
	// Recategorize tasks
	todoTasks, projectTasks, completedTasks := h.categorizer.CategorizeTasks(appState.Tasks)
	appState.TodoTasks = todoTasks
	appState.ProjectTasks = projectTasks
	appState.CompletedTasks = completedTasks
	
	// Update timeline categories
	overdueTasks, todayTasks, upcomingTasks := h.categorizer.CategorizeTimelineTasks(appState.Tasks)
	appState.OverdueTasks = overdueTasks
	appState.TodayTasks = todayTasks
	appState.UpcomingTasks = upcomingTasks
	
	// TODO: Reinitialize collapsible sections
	// This would require passing the CollapsibleManager to this handler
	// or adding methods to the appState to handle this
	
	return appState, nil
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
