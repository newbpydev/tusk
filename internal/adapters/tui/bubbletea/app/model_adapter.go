// Package app implements the main TUI application with an adapter pattern
// to bridge between the old monolithic code and the new modular architecture.
package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/form"
	// Comment out unused imports for now
	// "github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/handlers"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/hooks"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/services"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/state"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/styles"
	"github.com/newbpydev/tusk/internal/core/task"
)

// ModelAdapter implements the handler interfaces and integrates the refactored components
type ModelAdapter struct {
	// Original model to adapt
	model *Model
	
	// Refactored services
	taskService *services.TaskService
	categorizationService *services.TaskCategorizationService
	
	// Refactored components
	formComponent *form.FormModel
	
	// Application state
	appState *state.AppState
}

// NewModelAdapter creates a new adapter that wraps the original Model
func NewModelAdapter(model *Model) *ModelAdapter {
	// Create adapter
	adapter := &ModelAdapter{
		model: model,
		taskService: services.NewTaskService(services.NewCoreTaskServiceAdapter(model.taskSvc)),
		categorizationService: services.NewTaskCategorizationService(),
		formComponent: form.NewFormModel(&styles.Styles{}), // TODO: Get styles from model
		appState: state.NewAppState(),
	}
	
	// Initialize adapter state from model
	adapter.syncStateFromModel()
	
	return adapter
}

// syncStateFromModel updates the adapter's state from the original model
func (a *ModelAdapter) syncStateFromModel() {
	// Copy view mode and related state
	a.appState.ViewMode = a.model.viewMode
	a.appState.ActivePanel = a.model.activePanel
	a.appState.CurrentTime = a.model.currentTime
	
	// Copy task data
	a.appState.Tasks = a.model.tasks
	a.appState.TodoTasks = a.model.todoTasks
	a.appState.ProjectTasks = a.model.projectTasks
	a.appState.CompletedTasks = a.model.completedTasks
	a.appState.OverdueTasks = a.model.overdueTasks
	a.appState.TodayTasks = a.model.todayTasks
	a.appState.UpcomingTasks = a.model.upcomingTasks
	
	// Copy cursor state
	a.appState.Cursor = a.model.cursor
	a.appState.VisualCursor = a.model.visualCursor
	a.appState.CursorOnHeader = a.model.cursorOnHeader
	a.appState.TaskListOffset = a.model.taskListOffset
	a.appState.TaskDetailsOffset = a.model.taskDetailsOffset
	a.appState.TimelineCursor = a.model.timelineCursor
	a.appState.TimelineCursorOnHeader = a.model.timelineCursorOnHeader
	
	// Copy status state
	a.appState.StatusMessage = a.model.statusMessage
	a.appState.StatusType = "" // Default value
	a.appState.StatusExpiry = time.Time{} // Default value
	a.appState.IsLoading = a.model.isLoading
	a.appState.LoadingMessage = "" // Default value
	a.appState.Error = a.model.err
	
	// Copy form state
	a.formComponent.Title = a.model.formTitle
	a.formComponent.Description = a.model.formDescription
	a.formComponent.Priority = a.model.formPriority
	a.formComponent.DueDate = a.model.formDueDate
	// Commented out until these fields are added to the Model struct
	// a.formComponent.IsCompleted = a.model.formIsCompleted
	// a.formComponent.ParentID = a.model.formParentID
	// a.formComponent.TaskID = a.model.formTaskID
	// Convert activeField (int) to a string field name
	switch a.model.activeField {
	case 0:
		a.formComponent.FocusedField = "Title"
	case 1:
		a.formComponent.FocusedField = "Description"
	case 2:
		a.formComponent.FocusedField = "Priority"
	case 3:
		a.formComponent.FocusedField = "DueDate"
	default:
		a.formComponent.FocusedField = "Title"
	}
	
	// Copy panel visibility
	a.appState.ShowTaskList = a.model.showTaskList
	a.appState.ShowTaskDetails = a.model.showTaskDetails
	a.appState.ShowTimeline = a.model.showTimeline
	
	// Collapsible managers are references
	a.appState.CollapsibleManager = a.model.collapsibleManager
	// timelineCollapsibleManager doesn't exist in Model struct
	a.appState.TimelineCollapsibleManager = a.model.timelineCollapsibleMgr
}

// syncModelFromState updates the original model from the adapter's state
func (a *ModelAdapter) syncModelFromState() {
	// Copy view mode and related state
	a.model.viewMode = a.appState.ViewMode
	a.model.activePanel = a.appState.ActivePanel
	a.model.currentTime = a.appState.CurrentTime
	
	// Copy task data
	a.model.tasks = a.appState.Tasks
	a.model.todoTasks = a.appState.TodoTasks
	a.model.projectTasks = a.appState.ProjectTasks
	a.model.completedTasks = a.appState.CompletedTasks
	a.model.overdueTasks = a.appState.OverdueTasks
	a.model.todayTasks = a.appState.TodayTasks
	a.model.upcomingTasks = a.appState.UpcomingTasks
	
	// Copy cursor state
	a.model.cursor = a.appState.Cursor
	a.model.visualCursor = a.appState.VisualCursor
	a.model.cursorOnHeader = a.appState.CursorOnHeader
	a.model.taskListOffset = a.appState.TaskListOffset
	a.model.taskDetailsOffset = a.appState.TaskDetailsOffset
	a.model.timelineCursor = a.appState.TimelineCursor
	a.model.timelineCursorOnHeader = a.appState.TimelineCursorOnHeader
	
	// Copy status state
	a.model.statusMessage = a.appState.StatusMessage
	// Commenting out fields that don't exist in the Model struct
	// a.model.statusType = a.appState.StatusType
	// a.model.statusExpiry = a.appState.StatusExpiry
	a.model.isLoading = a.appState.IsLoading
	// a.model.loadingMessage = a.appState.LoadingMessage
	a.model.err = a.appState.Error
	
	// Copy form state from form component
	a.model.formTitle = a.formComponent.Title
	a.model.formDescription = a.formComponent.Description
	a.model.formPriority = a.formComponent.Priority
	a.model.formDueDate = a.formComponent.DueDate
	// These fields don't exist in the Model struct - commenting out
	// a.model.formIsCompleted = a.formComponent.IsCompleted
	// a.model.formParentID = a.formComponent.ParentID
	// a.model.formTaskID = a.formComponent.TaskID
	// Convert string field name to integer index
	switch a.formComponent.FocusedField {
	case "Title":
		a.model.activeField = 0
	case "Description":
		a.model.activeField = 1
	case "Priority":
		a.model.activeField = 2
	case "DueDate":
		a.model.activeField = 3
	default:
		a.model.activeField = 0
	}
	
	// Copy panel visibility
	a.model.showTaskList = a.appState.ShowTaskList
	a.model.showTaskDetails = a.appState.ShowTaskDetails
	a.model.showTimeline = a.appState.ShowTimeline
	
	// Collapsible managers are references
	a.model.collapsibleManager = a.appState.CollapsibleManager
	// timelineCollapsibleManager doesn't exist in Model struct
	a.model.timelineCollapsibleMgr = a.appState.TimelineCollapsibleManager
}

// Implement handlers.AppModel interface methods

// NavigateDown moves the cursor down
func (a *ModelAdapter) NavigateDown() {
	a.model.navigateDown()
	a.syncStateFromModel()
}

// NavigateUp moves the cursor up
func (a *ModelAdapter) NavigateUp() {
	a.model.navigateUp()
	a.syncStateFromModel()
}

// NavigateToTop moves cursor to top of the list
func (a *ModelAdapter) NavigateToTop() {
	a.model.navigateToTop()
	a.syncStateFromModel()
}

// NavigateToBottom moves cursor to bottom of the list
func (a *ModelAdapter) NavigateToBottom() {
	a.model.navigateToBottom()
	a.syncStateFromModel()
}

// ToggleTaskCompletion toggles the completion status of the current task
func (a *ModelAdapter) ToggleTaskCompletion() tea.Cmd {
	// Use the refactored command pattern instead of calling model directly
	if len(a.model.tasks) == 0 || a.model.cursor >= len(a.model.tasks) || a.model.cursorOnHeader {
		return nil // Cannot toggle status if no task is selected or cursor is on header
	}
	
	currentTask := a.model.tasks[a.model.cursor]
	return a.taskService.ToggleTaskCompletion(
		a.model.ctx,
		currentTask.ID,
		currentTask,
	)
}

// RefreshTasks initiates task refresh
func (a *ModelAdapter) RefreshTasks() tea.Cmd {
	return a.taskService.RefreshTasks(
		a.model.ctx,
		a.model.userID,
		a.SetLoadingStatus,
	)
}

// ResetForm clears the form
func (a *ModelAdapter) ResetForm() {
	a.formComponent.Reset()
	a.syncModelFromState()
}

// LoadTaskIntoForm loads a task's data into the form
func (a *ModelAdapter) LoadTaskIntoForm(t task.Task) {
	a.formComponent.LoadTask(t)
	a.syncModelFromState()
}

// SetLoadingStatus updates the loading status
func (a *ModelAdapter) SetLoadingStatus(msg string) {
	a.model.setLoadingStatus(msg)
	a.syncStateFromModel()
}

// SetStatusMessage updates the status message
func (a *ModelAdapter) SetStatusMessage(msg, statusType string, duration time.Duration) {
	a.model.setStatusMessage(msg, statusType, duration)
	a.syncStateFromModel()
}

// ToggleSection expands or collapses the section at the cursor
func (a *ModelAdapter) ToggleSection() tea.Cmd {
	return a.model.toggleSection()
}

// InitCollapsibleSections initializes section data
func (a *ModelAdapter) InitCollapsibleSections() {
	a.model.initCollapsibleSections()
	a.syncStateFromModel()
}

// GetCollapsibleManager returns the collapsible manager
func (a *ModelAdapter) GetCollapsibleManager() *hooks.CollapsibleManager {
	return a.model.collapsibleManager
}

// GetCursor returns the current cursor position
func (a *ModelAdapter) GetCursor() int {
	return a.model.cursor
}

// GetCursorOnHeader returns whether the cursor is on a header
func (a *ModelAdapter) GetCursorOnHeader() bool {
	return a.model.cursorOnHeader
}

// GetTasks returns the task list
func (a *ModelAdapter) GetTasks() []task.Task {
	return a.model.tasks
}

// SetActivePanel sets the active panel
func (a *ModelAdapter) SetActivePanel(panel int) {
	a.model.activePanel = panel
	a.syncStateFromModel()
}

// SetViewMode sets the view mode
func (a *ModelAdapter) SetViewMode(mode string) {
	a.model.viewMode = mode
	a.syncStateFromModel()
}

// SetFormPriority sets the form priority
func (a *ModelAdapter) SetFormPriority(priority string) {
	a.formComponent.Priority = priority
	a.syncModelFromState()
}

// Next functions would implement the FormModelInterface required by form_handlers.go
// Functions like NextFormField, PreviousFormField, etc.

// HandleFormSubmit handles form submission
func (a *ModelAdapter) HandleFormSubmit() tea.Cmd {
	// Validate form
	if !a.formComponent.Validate() {
		a.syncModelFromState()
		return nil
	}
	
	// Create task from form data
	if a.formComponent.IsEdit {
		// Editing existing task
		if a.formComponent.TaskID == nil {
			return nil
		}
		
		var dueDate *time.Time
		if a.formComponent.DueDate != "" {
			parsed, err := time.Parse("2006-01-02", a.formComponent.DueDate)
			if err == nil {
				dueDate = &parsed
			}
		}
		// Convert description to pointer
		descPtr := a.formComponent.Description
		
		// Create task struct for update
		taskToUpdate := task.Task{
			ID:          *a.formComponent.TaskID,
			Title:       a.formComponent.Title,
			Description: &descPtr,
			Priority:    task.Priority(a.formComponent.Priority),
			DueDate:     dueDate,
			IsCompleted: a.formComponent.IsCompleted,
			ParentID:    a.formComponent.ParentID,
		}

		return a.taskService.UpdateTask(
			a.model.ctx,
			taskToUpdate,
			a.SetLoadingStatus,
		)
	} else {
		// Creating new task
		var dueDate *time.Time
		if a.formComponent.DueDate != "" {
			parsed, err := time.Parse("2006-01-02", a.formComponent.DueDate)
			if err == nil {
				dueDate = &parsed
			}
		}
		// Convert description to pointer
		descPtr := a.formComponent.Description
		
		// Create task struct for creation
		newTask := task.Task{
			UserID:      int32(a.model.userID), // Convert int64 to int32
			Title:       a.formComponent.Title,
			Description: &descPtr,
			Priority:    task.Priority(a.formComponent.Priority),
			DueDate:     dueDate,
			IsCompleted: a.formComponent.IsCompleted,
			ParentID:    a.formComponent.ParentID,
		}

		return a.taskService.CreateTask(
			a.model.ctx,
			newTask,
			a.SetLoadingStatus,
		)
	}
}

// HandleFormCancel cancels form editing
func (a *ModelAdapter) HandleFormCancel() tea.Cmd {
	a.model.viewMode = "list"
	a.syncStateFromModel()
	return nil
}

// Additional methods needed for implementing FormModelInterface
// UpdateFormField updates a form field
func (a *ModelAdapter) UpdateFormField(field, value string) {
	a.formComponent.UpdateField(field, value)
	a.syncModelFromState()
}

// ToggleFormField toggles a boolean form field
func (a *ModelAdapter) ToggleFormField(field string) {
	a.formComponent.ToggleField(field)
	a.syncModelFromState()
}

// NextFormField moves to the next form field
func (a *ModelAdapter) NextFormField() {
	a.formComponent.NextField()
	a.syncModelFromState()
}

// PreviousFormField moves to the previous form field
func (a *ModelAdapter) PreviousFormField() {
	a.formComponent.PreviousField()
	a.syncModelFromState()
}

// GetCurrentFormField returns the currently focused form field
func (a *ModelAdapter) GetCurrentFormField() string {
	return a.formComponent.FocusedField
}

// GetFormField returns a form field value
func (a *ModelAdapter) GetFormField(field string) string {
	switch field {
	case "title":
		return a.formComponent.Title
	case "description":
		return a.formComponent.Description
	case "dueDate":
		return a.formComponent.DueDate
	case "priority":
		return a.formComponent.Priority
	default:
		return ""
	}
}

// IsFormFieldFocused checks if a form field is focused
func (a *ModelAdapter) IsFormFieldFocused(field string) bool {
	return a.formComponent.FocusedField == field
}

// RenderView integrates components for a complete view
func (a *ModelAdapter) RenderView() string {
	// Would use the task_list, task_details components to render the view
	// based on current app state
	// This is just a placeholder
	return "ModelAdapter view rendering is not implemented yet"
}
