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
	// No commands to run at initialization
	return nil
}

// refreshTasks reloads the task list from the service
func (m *Model) refreshTasks() tea.Msg {
	tasks, err := m.taskSvc.List(m.ctx, m.userID)
	if err != nil {
		m.err = err
		return nil
	}

	m.tasks = tasks

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

	var err error
	if currentTask.Status == task.StatusDone {
		// Change from done to todo
		_, err = m.taskSvc.ChangeStatus(m.ctx, taskID, task.StatusTodo)
	} else {
		// Change to done
		_, err = m.taskSvc.ChangeStatus(m.ctx, taskID, task.StatusDone)
	}

	if err != nil {
		m.err = err
	}

	// Refresh tasks after toggle
	return m.refreshTasks()
}

// deleteCurrentTask deletes the currently selected task
func (m *Model) deleteCurrentTask() tea.Msg {
	if len(m.tasks) == 0 || m.cursor >= len(m.tasks) {
		return nil
	}

	taskID := int64(m.tasks[m.cursor].ID)
	if err := m.taskSvc.Delete(m.ctx, taskID); err != nil {
		m.err = err
		return nil
	}

	// Go back to list view after deletion
	m.viewMode = "list"

	// Refresh tasks after deletion
	return m.refreshTasks()
}

// createNewTask creates a new task from the form fields
func (m *Model) createNewTask() tea.Msg {
	if m.formTitle == "" {
		m.err = fmt.Errorf("title is required")
		return nil
	}

	// Parse due date if provided
	var dueDate *time.Time
	if m.formDueDate != "" {
		date, err := time.Parse("2006-01-02", m.formDueDate)
		if err == nil {
			dueDate = &date
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
		return nil
	}

	// Set success message
	m.err = nil
	m.successMsg = fmt.Sprintf("Task '%s' successfully saved to database", createdTask.Title)

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
