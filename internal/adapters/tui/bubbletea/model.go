// Copyright (C) 2025 Juan Antonio Gomez Pena
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package bubbletea

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/core/task"
	taskService "github.com/newbpydev/tusk/internal/service/task"
)

// Model represents the state of the TUI application.
// It contains the context, a slice of tasks, a cursor for navigation, and an error field.
type Model struct {
	ctx      context.Context
	tasks    []task.Task
	cursor   int
	err      error
	taskSvc  taskService.Service
	userID   int64
	viewMode string // "list", "detail", "edit", "create"
	width    int
	height   int
	styles   *Styles // Reference to the current styles

	// Form fields for creating/editing tasks
	formTitle       string
	formDescription string
	formPriority    string
	formDueDate     string
	formStatus      string
	activeField     int // Current field in focus when editing

	// Column states
	showTaskList    bool
	showTaskDetails bool
	showTimeline    bool
	activePanel     int // 0 = task list, 1 = task details, 2 = timeline

	// Scrolling offset for each panel
	taskListOffset    int // Vertical scroll position for task list panel
	taskDetailsOffset int // Vertical scroll position for task details panel
	timelineOffset    int // Vertical scroll position for timeline panel

	// Header information
	currentTime   time.Time // Current time to display in header
	statusMessage string    // Status message to display in header (success, error, etc.)
	statusType    string    // Type of status message: "success", "error", "info", "loading"
	statusExpiry  time.Time // When to clear the status message
	isLoading     bool      // Whether the app is currently loading data

	// Success message
	successMsg string
}

// NewModel initializes a new Model instance with the provided context and tasks.
// It sets the cursor to 0 and the error field to nil.
func NewModel(ctx context.Context, svc taskService.Service, userID int64) *Model {
	roots, err := svc.List(ctx, userID)

	model := &Model{
		ctx:             ctx,
		tasks:           roots,
		cursor:          0,
		err:             err,
		taskSvc:         svc,
		userID:          userID,
		styles:          ActiveStyles, // Use the active styles
		showTaskList:    true,
		showTaskDetails: true,
		showTimeline:    true,
		activePanel:     0, // Start with focus on the task list panel
		// Default to list view even if there are no tasks
		viewMode: "list",
	}

	return model
}

// Init initializes the bubbletea model.
func (m *Model) Init() tea.Cmd {
	// Initialize time
	m.currentTime = time.Now()

	// Start a ticker to update the time every second
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// TickMsg is a message that's sent when the ticker ticks
type TickMsg time.Time

// setStatusMessage sets a status message with a type and expiry duration
func (m *Model) setStatusMessage(msg string, msgType string, duration time.Duration) {
	m.statusMessage = msg
	m.statusType = msgType
	m.statusExpiry = time.Now().Add(duration)
}

// setSuccessStatus is a helper to set success status messages
func (m *Model) setSuccessStatus(msg string) {
	m.setStatusMessage(msg, "success", 5*time.Second)
}

// setErrorStatus is a helper to set error status messages
func (m *Model) setErrorStatus(msg string) {
	m.setStatusMessage(msg, "error", 10*time.Second)
}

// setInfoStatus is a helper to set informational status messages
func (m *Model) setInfoStatus(msg string) {
	m.setStatusMessage(msg, "info", 3*time.Second)
}

// setLoadingStatus sets the app in loading state with a message
func (m *Model) setLoadingStatus(msg string) {
	m.setStatusMessage(msg, "loading", 30*time.Second)
	m.isLoading = true
}

// clearLoadingStatus clears the loading state
func (m *Model) clearLoadingStatus() {
	m.isLoading = false
	if m.statusType == "loading" {
		m.statusMessage = ""
		m.statusType = ""
	}
}

// refreshTasks reloads the task list from the service
func (m *Model) refreshTasks() tea.Msg {
	m.setLoadingStatus("Loading tasks...")

	tasks, err := m.taskSvc.List(m.ctx, m.userID)
	if err != nil {
		m.err = err
		m.setErrorStatus(fmt.Sprintf("Error loading tasks: %v", err))
		return nil
	}

	m.clearLoadingStatus()
	m.tasks = tasks
	m.setInfoStatus("Tasks refreshed")

	// Adjust cursor if needed
	if m.cursor >= len(m.tasks) && len(m.tasks) > 0 {
		m.cursor = len(m.tasks) - 1
	}

	return nil
}

// toggleTaskCompletion toggles the completion status of the selected task
func (m *Model) toggleTaskCompletion() tea.Msg {
	if len(m.tasks) == 0 || m.cursor >= len(m.tasks) {
		return nil
	}

	currentTask := m.tasks[m.cursor]
	taskID := int64(currentTask.ID)
	taskTitle := currentTask.Title

	// Immediately update UI with new status (optimistic update)
	var newStatus task.Status
	if currentTask.Status == task.StatusDone {
		// Change from done to todo
		newStatus = task.StatusTodo
		m.tasks[m.cursor].Status = newStatus
		m.tasks[m.cursor].IsCompleted = false
		m.setSuccessStatus(fmt.Sprintf("Task '%s' marked as TODO", taskTitle))
	} else {
		// Change to done
		newStatus = task.StatusDone
		m.tasks[m.cursor].Status = newStatus
		m.tasks[m.cursor].IsCompleted = true
		m.setSuccessStatus(fmt.Sprintf("Task '%s' marked as DONE", taskTitle))
	}

	// Show subtle loading indicator without blocking the UI
	m.setInfoStatus("Saving changes...")

	// Start a background operation to update the database
	return func() tea.Msg {
		// Perform the actual status change in the background
		_, err := m.taskSvc.ChangeStatus(m.ctx, taskID, newStatus)

		if err != nil {
			// If there's an error, revert the optimistic update
			return statusUpdateErrorMsg{
				taskIndex: m.cursor,
				err:       err,
				taskTitle: taskTitle,
			}
		}

		// Refresh tasks in the background to ensure data consistency
		// but the UI is already updated so user sees the change immediately
		tasks, err := m.taskSvc.List(m.ctx, m.userID)
		if err != nil {
			return errorMsg(err)
		}

		return tasksRefreshedMsg{tasks: tasks}
	}
}

// Custom message types for handling background operations
type statusUpdateErrorMsg struct {
	taskIndex int
	taskTitle string
	err       error
}

type errorMsg error

type tasksRefreshedMsg struct {
	tasks []task.Task
}

// deleteCurrentTask deletes the currently selected task
func (m *Model) deleteCurrentTask() tea.Msg {
	if len(m.tasks) == 0 || m.cursor >= len(m.tasks) {
		return nil
	}

	taskTitle := m.tasks[m.cursor].Title
	taskID := int64(m.tasks[m.cursor].ID)

	m.setLoadingStatus("Deleting task...")

	if err := m.taskSvc.Delete(m.ctx, taskID); err != nil {
		m.err = err
		m.setErrorStatus(fmt.Sprintf("Error deleting task: %v", err))
		return nil
	}

	// Success message
	m.setSuccessStatus(fmt.Sprintf("Task '%s' deleted", taskTitle))

	// Go back to list view after deletion
	m.viewMode = "list"

	// Refresh tasks after deletion
	return m.refreshTasks()
}

// createNewTask creates a new task from the form fields
func (m *Model) createNewTask() tea.Msg {
	if m.formTitle == "" {
		m.err = fmt.Errorf("title is required")
		m.setErrorStatus("Title is required")
		return nil
	}

	m.setLoadingStatus("Creating new task...")

	// Parse due date if provided
	var dueDate *time.Time
	if m.formDueDate != "" {
		date, err := time.Parse("2006-01-02", m.formDueDate)
		if err == nil {
			dueDate = &date
		} else {
			m.setErrorStatus(fmt.Sprintf("Invalid date format: %v", err))
			m.err = fmt.Errorf("invalid date format: %v", err)
			return nil
		}
	}

	// Convert priority string to task.Priority enum
	priority := task.PriorityLow
	if m.formPriority == string(task.PriorityMedium) {
		priority = task.PriorityMedium
	} else if m.formPriority == string(task.PriorityHigh) {
		priority = task.PriorityHigh
	}

	// Prepare description
	var description string
	if m.formDescription != "" {
		description = m.formDescription
	}

	// Call service to create task
	// Pass nil for parent ID as this is a root task
	createdTask, err := m.taskSvc.Create(
		m.ctx,       // context
		m.userID,    // user ID
		nil,         // parent ID (nil for root tasks)
		m.formTitle, // title
		description, // description
		dueDate,     // due date
		priority,    // priority
		[]string{},  // tags
	)

	if err != nil {
		m.err = err
		m.setErrorStatus(fmt.Sprintf("Error creating task: %v", err))
		return nil
	}

	// Set success message
	m.err = nil
	m.setSuccessStatus(fmt.Sprintf("Task '%s' successfully created", createdTask.Title))

	// Reset form fields
	m.formTitle = ""
	m.formDescription = ""
	m.formPriority = ""
	m.formDueDate = ""
	m.formStatus = ""
	m.activeField = 0

	// Switch to list view
	m.viewMode = "list"

	// Refresh tasks
	return m.refreshTasks()
}

// getTasksByTimeCategory organizes tasks into overdue, today, and upcoming categories
func (m *Model) getTasksByTimeCategory() ([]task.Task, []task.Task, []task.Task) {
	var overdue, todayTasks, upcoming []task.Task

	now := time.Now()
	todayDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := todayDate.AddDate(0, 0, 1)

	for _, t := range m.tasks {
		if t.DueDate == nil {
			continue
		}

		dueDate := *t.DueDate
		if dueDate.Before(todayDate) {
			overdue = append(overdue, t)
		} else if dueDate.Before(tomorrow) {
			todayTasks = append(todayTasks, t)
		} else {
			upcoming = append(upcoming, t)
		}
	}

	return overdue, todayTasks, upcoming
}
