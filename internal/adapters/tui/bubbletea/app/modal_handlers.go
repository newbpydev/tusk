package app

import (
	"errors"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/form"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/shared"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/types"
	"github.com/newbpydev/tusk/internal/core/task"
)

// ShowCreateTaskModal displays the modal form for creating a new task
func (m *Model) ShowCreateTaskModal() tea.Cmd {
	// Create a new modal form model
	modalForm := form.NewModalFormModel(m.styles, int32(m.userID))
	
	// Default to empty form for creating a new task
	modalForm.Reset()
	
	// Create modal with the form component
	m.modal = shared.NewModal(modalForm, 60, 25, types.ContentArea)
	
	// Show the modal
	m.modal.Show()
	m.showModal = true
	
	return nil
}

// ShowEditTaskModal displays the modal form for editing an existing task
func (m *Model) ShowEditTaskModal(t task.Task) tea.Cmd {
	// Create a new modal form model
	modalForm := form.NewModalFormModel(m.styles, int32(m.userID))
	
	// Load the task data into the form
	modalForm.LoadTask(t)
	
	// Create modal with the form component
	m.modal = shared.NewModal(modalForm, 60, 25, types.ContentArea)
	
	// Show the modal
	m.modal.Show()
	m.showModal = true
	
	return nil
}

// HandleModalFormClose handles closing the modal form
func (m *Model) HandleModalFormClose() (tea.Model, tea.Cmd) {
	m.modal.Hide()
	m.showModal = false
	return m, nil
}

// HandleModalFormSubmit handles task creation/editing from modal form
func (m *Model) HandleModalFormSubmit(taskData task.Task) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	// Handle task creation or updating
	if taskData.ID != 0 {
		// Editing an existing task
		cmd = m.handleTaskUpdate(taskData)
	} else {
		// Creating a new task
		cmd = m.handleTaskCreate(taskData)
	}
	
	// Hide the modal after successful submission
	m.modal.Hide()
	m.showModal = false
	
	return m, cmd
}

// ErrorMsg is sent when an error occurs
type ErrorMsg struct {
	Err error
}

func (e ErrorMsg) Error() string { return e.Err.Error() }

// TaskCreatedMsg is sent when a new task is created
type TaskCreatedMsg struct {
	Task task.Task
}

// TaskUpdatedMsg is sent when a task is updated
type TaskUpdatedMsg struct {
	Task task.Task
}

// handleTaskCreate adds a new task to the system
func (m *Model) handleTaskCreate(t task.Task) tea.Cmd {
	return func() tea.Msg {
		// Extract task properties to match the service interface
		var parentID *int64
		if t.ParentID != nil {
			pid := int64(*t.ParentID)
			parentID = &pid
		}
		
		// Create the task
		newTask, err := m.taskSvc.Create(
			m.ctx,
			m.userID,
			parentID,
			t.Title,
			// Handle description
			func() string {
				if t.Description != nil {
					return *t.Description
				}
				return ""
			}(),
			t.DueDate,
			t.Priority,
			[]string{}, // Tags, which appear to be empty
		)
		
		if err != nil {
			return ErrorMsg{err}
		}
		
		// Show success message
		m.statusMessage = "Task created successfully"
		m.statusType = "success"
		
		// Refresh task list
		return TaskCreatedMsg{Task: newTask}
	}
}

// handleTaskUpdate updates an existing task
func (m *Model) handleTaskUpdate(t task.Task) tea.Cmd {
	return func() tea.Msg {
		// Check if task ID is valid
		if t.ID == 0 {
			return ErrorMsg{errors.New("invalid task ID")}
		}
		
		// Update the task
		updatedTask, err := m.taskSvc.Update(
			m.ctx,
			int64(t.ID),
			t.Title,
			// Handle description
			func() string {
				if t.Description != nil {
					return *t.Description
				}
				return ""
			}(),
			t.DueDate,
			t.Priority,
			[]string{}, // Tags, which appear to be empty
		)
		
		if err != nil {
			return ErrorMsg{err}
		}
		
		// Show success message
		m.statusMessage = "Task updated successfully"
		m.statusType = "success"
		
		// Refresh task list
		return TaskUpdatedMsg{Task: updatedTask}
	}
}
