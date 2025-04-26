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

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/tusk/internal/core/task"
	taskService "github.com/newbpydev/tusk/internal/service/task"
)

var (
	// Styles
	titleStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA"))
	selectedItemStyle = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("#1E88E5")).Foreground(lipgloss.Color("#FFF"))
	todoStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#909090"))
	inProgressStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB300"))
	doneStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#00E676"))

	// Priority styles
	lowStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#009688"))
	mediumStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FB8C00"))
	highStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#E53935"))

	// Help style
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#747474")).Italic(true)
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
	}
}

// Init initializes the bubbletea model.
func (m Model) Init() tea.Cmd {
	// No commands to run at initialization
	return nil
}

// Update handles user input and updates the model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	default:
		return m, nil
	}
}

// handleKeyPress processes keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.viewMode {
	case "list":
		return m.handleListViewKeys(msg)
	case "detail":
		return m.handleDetailViewKeys(msg)
	case "edit":
		return m.handleEditViewKeys(msg)
	default:
		return m, nil
	}
}

// handleListViewKeys processes keyboard input in list view
func (m Model) handleListViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case "down", "j":
		if m.cursor < len(m.tasks)-1 {
			m.cursor++
		}
		return m, nil

	case "enter":
		if len(m.tasks) > 0 {
			m.viewMode = "detail"
			return m, nil
		}

	case "c":
		// Toggle completion status
		if len(m.tasks) > 0 {
			return m, m.toggleTaskCompletion
		}

	case "r":
		// Refresh task list
		return m, m.refreshTasks

	case "s":
		// Change sort order (not implemented yet)
		return m, nil

	case "f":
		// Filter tasks (not implemented yet)
		return m, nil
	}

	return m, nil
}

// handleDetailViewKeys processes keyboard input in detail view
func (m Model) handleDetailViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.viewMode = "list"
		return m, nil

	case "e":
		m.viewMode = "edit"
		return m, nil

	case "d":
		// Delete task
		return m, m.deleteCurrentTask

	case "c":
		// Toggle completion status
		return m, m.toggleTaskCompletion
	}

	return m, nil
}

// handleEditViewKeys processes keyboard input in edit view
func (m Model) handleEditViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = "detail"
		return m, nil

	case "enter":
		// Save edits (not fully implemented)
		m.viewMode = "detail"
		return m, nil
	}

	return m, nil
}

// View renders the current state of the model as a string.
func (m Model) View() string {
	switch m.viewMode {
	case "list":
		return m.renderListView()
	case "detail":
		return m.renderDetailView()
	case "edit":
		return m.renderEditView()
	default:
		if m.err != nil {
			return fmt.Sprintf("Error: %v\n\n", m.err) + "Unknown view mode"
		}
		return "Unknown view mode"
	}
}

// renderListView renders the list of tasks
func (m Model) renderListView() string {
	s := titleStyle.Render("Tusk: Task Manager") + "\n\n"

	if m.err != nil {
		s += fmt.Sprintf("Error: %v\n\n", m.err)
	}

	if len(m.tasks) == 0 {
		s += "No tasks found. Add tasks using the CLI first.\n"
	} else {
		for i, t := range m.tasks {
			statusSymbol := "[ ]"
			var statusStyle lipgloss.Style

			switch t.Status {
			case task.StatusDone:
				statusSymbol = "[✓]"
				statusStyle = doneStyle
			case task.StatusInProgress:
				statusSymbol = "[⟳]"
				statusStyle = inProgressStyle
			default:
				statusStyle = todoStyle
			}

			var priorityStyle lipgloss.Style
			switch t.Priority {
			case task.PriorityHigh:
				priorityStyle = highStyle
			case task.PriorityMedium:
				priorityStyle = mediumStyle
			default:
				priorityStyle = lowStyle
			}

			priority := string(t.Priority)
			taskLine := fmt.Sprintf("%s %s (Priority: %s)",
				statusStyle.Render(statusSymbol),
				t.Title,
				priorityStyle.Render(priority))

			// Show completion progress for parent tasks
			if len(t.SubTasks) > 0 {
				progress := int(t.Progress * 100)
				taskLine += fmt.Sprintf(" [%d%% complete]", progress)
			}

			if i == m.cursor {
				s += selectedItemStyle.Render(taskLine) + "\n"
			} else {
				s += taskLine + "\n"
			}
		}
	}

	s += "\n" + helpStyle.Render("j/k: navigate • enter: view details • c: toggle completion • q: quit")

	return s
}

// renderDetailView renders the detailed view of the selected task
func (m Model) renderDetailView() string {
	if len(m.tasks) == 0 || m.cursor >= len(m.tasks) {
		return "No task selected"
	}

	t := m.tasks[m.cursor]

	s := titleStyle.Render("Task Details") + "\n\n"

	// Task title
	s += titleStyle.Render("Title: ") + t.Title + "\n\n"

	// Task description
	s += titleStyle.Render("Description: ") + "\n"
	if t.Description != nil && *t.Description != "" {
		s += *t.Description + "\n\n"
	} else {
		s += "No description\n\n"
	}

	// Status
	s += titleStyle.Render("Status: ")
	switch t.Status {
	case task.StatusDone:
		s += doneStyle.Render(string(t.Status))
	case task.StatusInProgress:
		s += inProgressStyle.Render(string(t.Status))
	default:
		s += todoStyle.Render(string(t.Status))
	}
	s += "\n\n"

	// Priority
	s += titleStyle.Render("Priority: ")
	switch t.Priority {
	case task.PriorityHigh:
		s += highStyle.Render(string(t.Priority))
	case task.PriorityMedium:
		s += mediumStyle.Render(string(t.Priority))
	default:
		s += lowStyle.Render(string(t.Priority))
	}
	s += "\n\n"

	// Due date
	s += titleStyle.Render("Due date: ")
	if t.DueDate != nil {
		s += t.DueDate.Format("2006-01-02")
	} else {
		s += "No due date"
	}
	s += "\n\n"

	// Tags
	s += titleStyle.Render("Tags: ")
	if len(t.Tags) > 0 {
		tags := ""
		for i, tag := range t.Tags {
			if i > 0 {
				tags += ", "
			}
			tags += tag.Name
		}
		s += tags
	} else {
		s += "No tags"
	}
	s += "\n\n"

	// Subtasks
	s += titleStyle.Render("Subtasks:") + "\n"
	if len(t.SubTasks) > 0 {
		for _, st := range t.SubTasks {
			statusSymbol := "[ ]"
			if st.Status == task.StatusDone {
				statusSymbol = "[✓]"
			} else if st.Status == task.StatusInProgress {
				statusSymbol = "[⟳]"
			}
			s += fmt.Sprintf("  %s %s\n", statusSymbol, st.Title)
		}
	} else {
		s += "  No subtasks\n"
	}
	s += "\n"

	// Progress
	if len(t.SubTasks) > 0 {
		s += titleStyle.Render(fmt.Sprintf("Progress: %d%% (%d/%d tasks completed)\n",
			int(t.Progress*100), t.CompletedCount, t.TotalCount))
	}

	s += "\n" + helpStyle.Render("esc: back • e: edit • c: toggle completion • d: delete")

	return s
}

// renderEditView renders the edit view of the selected task
func (m Model) renderEditView() string {
	// Basic placeholder for edit view - would require text input
	s := titleStyle.Render("Edit Task") + "\n\n"
	s += "Edit mode not fully implemented\n\n"
	s += helpStyle.Render("esc: cancel • enter: save changes")
	return s
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
