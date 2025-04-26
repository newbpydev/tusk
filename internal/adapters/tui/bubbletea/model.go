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
	viewMode string // "list", "detail", "edit"
	width    int
	height   int
	styles   *Styles // Reference to the current styles
}

// NewModel initializes a new Model instance with the provided context and tasks.
// It sets the cursor to 0 and the error field to nil.
func NewModel(ctx context.Context, svc taskService.Service, userID int64) Model {
	roots, err := svc.List(ctx, userID)
	return Model{
		ctx:      ctx,
		tasks:    roots,
		cursor:   0,
		err:      err,
		taskSvc:  svc,
		userID:   userID,
		viewMode: "list",
		styles:   ActiveStyles, // Use the active styles
	}
}

// Init initializes the bubbletea model.
func (m Model) Init() tea.Cmd {
	// No commands to run at initialization
	return nil
}

// refreshTasks reloads the task list from the service
func (m Model) refreshTasks() tea.Msg {
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
func (m Model) toggleTaskCompletion() tea.Msg {
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
func (m Model) deleteCurrentTask() tea.Msg {
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
