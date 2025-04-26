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
	"fmt"

	"github.com/newbpydev/tusk/internal/core/task"
)

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
		return "Unknown view mode"
	}
}

// renderListView renders the list of tasks
func (m Model) renderListView() string {
	s := m.styles.Title.Render("Tusk: Task Manager") + "\n\n"

	if m.err != nil {
		s += fmt.Sprintf("Error: %v\n\n", m.err)
	}

	if len(m.tasks) == 0 {
		s += "No tasks found. Add tasks using the CLI first.\n"
	} else {
		for i, t := range m.tasks {
			statusSymbol := "[ ]"
			var statusStyle = m.styles.Todo

			switch t.Status {
			case task.StatusDone:
				statusSymbol = "[✓]"
				statusStyle = m.styles.Done
			case task.StatusInProgress:
				statusSymbol = "[⟳]"
				statusStyle = m.styles.InProgress
			}

			var priorityStyle = m.styles.LowPriority
			switch t.Priority {
			case task.PriorityHigh:
				priorityStyle = m.styles.HighPriority
			case task.PriorityMedium:
				priorityStyle = m.styles.MediumPriority
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
				s += m.styles.SelectedItem.Render(taskLine) + "\n"
			} else {
				s += taskLine + "\n"
			}
		}
	}

	s += "\n" + m.styles.Help.Render("j/k: navigate • enter: view details • c: toggle completion • q: quit")

	return s
}

// renderDetailView renders the detailed view of the selected task
func (m Model) renderDetailView() string {
	if len(m.tasks) == 0 || m.cursor >= len(m.tasks) {
		return "No task selected"
	}

	t := m.tasks[m.cursor]

	s := m.styles.Title.Render("Task Details") + "\n\n"

	// Task title
	s += m.styles.Title.Render("Title: ") + t.Title + "\n\n"

	// Task description
	s += m.styles.Title.Render("Description: ") + "\n"
	if t.Description != nil && *t.Description != "" {
		s += *t.Description + "\n\n"
	} else {
		s += "No description\n\n"
	}

	// Status
	s += m.styles.Title.Render("Status: ")
	switch t.Status {
	case task.StatusDone:
		s += m.styles.Done.Render(string(t.Status))
	case task.StatusInProgress:
		s += m.styles.InProgress.Render(string(t.Status))
	default:
		s += m.styles.Todo.Render(string(t.Status))
	}
	s += "\n\n"

	// Priority
	s += m.styles.Title.Render("Priority: ")
	switch t.Priority {
	case task.PriorityHigh:
		s += m.styles.HighPriority.Render(string(t.Priority))
	case task.PriorityMedium:
		s += m.styles.MediumPriority.Render(string(t.Priority))
	default:
		s += m.styles.LowPriority.Render(string(t.Priority))
	}
	s += "\n\n"

	// Due date
	s += m.styles.Title.Render("Due date: ")
	if t.DueDate != nil {
		s += t.DueDate.Format("2006-01-02")
	} else {
		s += "No due date"
	}
	s += "\n\n"

	// Tags
	s += m.styles.Title.Render("Tags: ")
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
	s += m.styles.Title.Render("Subtasks:") + "\n"
	if len(t.SubTasks) > 0 {
		for _, st := range t.SubTasks {
			statusSymbol := "[ ]"
			var statusStyle = m.styles.Todo

			switch st.Status {
			case task.StatusDone:
				statusSymbol = "[✓]"
				statusStyle = m.styles.Done
			case task.StatusInProgress:
				statusSymbol = "[⟳]"
				statusStyle = m.styles.InProgress
			}

			s += fmt.Sprintf("  %s %s\n", statusStyle.Render(statusSymbol), st.Title)
		}
	} else {
		s += "  No subtasks\n"
	}
	s += "\n"

	// Progress
	if len(t.SubTasks) > 0 {
		s += m.styles.Title.Render(fmt.Sprintf("Progress: %d%% (%d/%d tasks completed)\n",
			int(t.Progress*100), t.CompletedCount, t.TotalCount))
	}

	s += "\n" + m.styles.Help.Render("esc: back • e: edit • c: toggle completion • d: delete")

	return s
}

// renderEditView renders the edit view of the selected task
func (m Model) renderEditView() string {
	// Basic placeholder for edit view - would require text input
	s := m.styles.Title.Render("Edit Task") + "\n\n"
	s += "Edit mode not fully implemented\n\n"
	s += m.styles.Help.Render("esc: cancel • enter: save changes")
	return s
}
