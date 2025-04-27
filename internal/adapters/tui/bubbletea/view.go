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

	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/tusk/internal/core/task"
)

// View renders the current state of the model as a string.
func (m *Model) View() string {
	if m.viewMode == "create" {
		// Special case for create form
		return m.renderCreateForm()
	}

	// Three-column layout
	var columns []string

	// Column 1: Task list
	if m.showTaskList {
		columns = append(columns, m.renderTaskList())
	}

	// Column 2: Task details
	if m.showTaskDetails {
		columns = append(columns, m.renderTaskDetails())
	}

	// Column 3: Timeline view
	if m.showTimeline {
		columns = append(columns, m.renderTimelineView())
	}

	// Join columns with dividers
	columnWidth := m.width / max(1, len(columns))

	for i := range columns {
		// Set fixed width for each column
		columns[i] = lipgloss.NewStyle().
			Width(columnWidth).
			MaxWidth(columnWidth).
			Render(columns[i])
	}

	content := lipgloss.JoinHorizontal(lipgloss.Top, columns...)

	// Calculate available height for content, leaving space for help text
	contentHeight := m.height - 2 // Reserve 2 rows for help text with padding

	// Ensure content fits available height
	contentStyle := lipgloss.NewStyle().
		MaxHeight(contentHeight).
		Height(contentHeight)

	content = contentStyle.Render(content)

	// Add help footer fixed at the bottom with proper styling
	helpText := m.renderHelpText()

	// Create a prominent footer style using the existing style variables
	footerStyle := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Bold(true).
		Background(lipgloss.Color("#333333")). // Darker background for visibility
		Foreground(lipgloss.Color("#FFFFFF")). // White text for contrast
		Padding(0, 1).
		MarginTop(1)

	styledHelp := footerStyle.Render(helpText)

	// Position content and help text
	return lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		styledHelp,
	)
}

// renderTaskList renders the first column with the list of tasks
func (m *Model) renderTaskList() string {
	s := m.styles.Title.Render("Tasks") + "\n\n"

	// Display error message if exists
	if m.err != nil {
		s += fmt.Sprintf("Error: %v\n\n", m.err)
	}

	// Display success message if exists
	if m.successMsg != "" {
		s += m.styles.Done.Render(fmt.Sprintf("✓ %s\n\n", m.successMsg))
		// Clear the success message after it's been displayed once
		// This is a deferred operation so it gets cleared on the next update
		defer func() { m.successMsg = "" }()
	}

	if len(m.tasks) == 0 {
		s += "No tasks found.\n\n"
		s += "Press 'n' to create a new task.\n"
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
			taskLine := fmt.Sprintf("%s %s (%s)",
				statusStyle.Render(statusSymbol),
				t.Title,
				priorityStyle.Render(priority))

			if i == m.cursor {
				s += m.styles.SelectedItem.Render(taskLine) + "\n"
			} else {
				s += taskLine + "\n"
			}
		}
	}

	return s
}

// renderTaskDetails renders the second column with details of the currently selected task
func (m *Model) renderTaskDetails() string {
	if len(m.tasks) == 0 {
		s := m.styles.Title.Render("Task Details") + "\n\n"
		s += "No tasks yet. Press 'n' to create your first task.\n\n"
		s += m.styles.Help.Render("Tip: You can organize tasks with priorities and due dates!")
		return s
	}

	if len(m.tasks) == 0 || m.cursor >= len(m.tasks) {
		return m.styles.Title.Render("Task Details") + "\n\nNo task selected"
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

	// Subtasks section
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

	// Progress
	if len(t.SubTasks) > 0 {
		s += "\n" + m.styles.Title.Render(fmt.Sprintf("Progress: %d%% (%d/%d tasks completed)\n",
			int(t.Progress*100), t.CompletedCount, t.TotalCount))
	}

	return s
}

// renderTimelineView renders the third column with time-based task categories
func (m *Model) renderTimelineView() string {
	s := m.styles.Title.Render("Timeline") + "\n\n"

	if len(m.tasks) == 0 {
		s += "No tasks to display in timeline.\n\n"
		s += "Create tasks with due dates to see them organized here."
		return s
	}

	overdue, today, upcoming := m.getTasksByTimeCategory()

	// Overdue section
	s += m.styles.HighPriority.Bold(true).Render("Overdue:") + "\n"
	if len(overdue) > 0 {
		for _, t := range overdue {
			dueDate := ""
			if t.DueDate != nil {
				dueDate = t.DueDate.Format("2006-01-02")
			}
			s += fmt.Sprintf("  %s (%s)\n", t.Title, dueDate)
		}
	} else {
		s += "  No overdue tasks\n"
	}
	s += "\n"

	// Today section
	s += m.styles.MediumPriority.Bold(true).Render("Today:") + "\n"
	if len(today) > 0 {
		for _, t := range today {
			s += fmt.Sprintf("  %s\n", t.Title)
		}
	} else {
		s += "  No tasks due today\n"
	}
	s += "\n"

	// Upcoming section
	s += m.styles.LowPriority.Bold(true).Render("Upcoming:") + "\n"
	if len(upcoming) > 0 {
		for _, t := range upcoming {
			dueDate := ""
			if t.DueDate != nil {
				dueDate = t.DueDate.Format("2006-01-02")
			}
			s += fmt.Sprintf("  %s (%s)\n", t.Title, dueDate)
		}
	} else {
		s += "  No upcoming tasks\n"
	}

	return s
}

// renderCreateForm renders the task creation form
func (m *Model) renderCreateForm() string {
	s := m.styles.Title.Render("Create New Task") + "\n\n"

	if m.err != nil {
		s += m.styles.HighPriority.Render(fmt.Sprintf("Error: %v\n\n", m.err))
	}

	// Form fields
	formFields := []struct {
		label    string
		value    string
		active   bool
		required bool
	}{
		{"Title", m.formTitle, m.activeField == 0, true},
		{"Description", m.formDescription, m.activeField == 1, false},
		{"Priority", m.formPriority, m.activeField == 2, false},
		{"Due Date (YYYY-MM-DD)", m.formDueDate, m.activeField == 3, false},
	}

	// Render each field
	for i, field := range formFields {
		// Field label with required indicator
		fieldLabel := field.label
		if field.required {
			fieldLabel += " *"
		}

		if field.active {
			s += m.styles.SelectedItem.Render(fieldLabel) + ": "
		} else {
			s += m.styles.Title.Render(fieldLabel) + ": "
		}

		// Field value
		if field.active {
			s += m.styles.SelectedItem.Render(field.value + "█") // Add cursor
		} else {
			s += field.value
		}

		// Special handling for priority field
		if i == 2 {
			var priorityStyle lipgloss.Style
			switch m.formPriority {
			case string(task.PriorityHigh):
				priorityStyle = m.styles.HighPriority
			case string(task.PriorityMedium):
				priorityStyle = m.styles.MediumPriority
			default:
				priorityStyle = m.styles.LowPriority
			}

			s += " (" + priorityStyle.Render(m.formPriority) + ")" +
				" - Press Space to cycle"
		}

		s += "\n\n"
	}

	// Submit button
	if m.activeField == 4 {
		s += m.styles.SelectedItem.Render("[Save Task]")
	} else {
		s += "[Save Task]"
	}

	s += "\n\n" + m.styles.Help.Render("Tab: next field • Enter: submit • Esc: cancel")

	return s
}

// renderHelpText renders the help text footer for the current view mode
func (m *Model) renderHelpText() string {
	var help string

	switch m.viewMode {
	case "list":
		help = "j/k: navigate • enter: view details • c: toggle completion • n: new task • 1/2/3: toggle columns • q: quit"
	case "detail":
		help = "esc: back • e: edit • c: toggle completion • d: delete • n: new task"
	case "edit":
		help = "esc: cancel • enter: save changes"
	case "create":
		help = "tab: next field • enter: submit • esc: cancel • space: cycle priority"
	}

	return m.styles.Help.Render(help)
}

// renderEditView renders the edit view of the selected task
func (m *Model) renderEditView() string {
	// Basic placeholder for edit view - would require text input
	s := m.styles.Title.Render("Edit Task") + "\n\n"
	s += "Edit mode not fully implemented\n\n"
	s += m.styles.Help.Render("esc: cancel • enter: save changes")
	return s
}

// max returns the larger of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
