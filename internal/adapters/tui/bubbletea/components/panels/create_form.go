// Copyright (C) 2025 Juan Antonio Gomez Pena
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at any later version).
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package panels

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/input"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/shared"
	"github.com/newbpydev/tusk/internal/core/task"
)

// CreateFormProps contains all properties needed to render the task creation form
type CreateFormProps struct {
	FormTitle       string
	FormDescription string
	FormPriority    string
	FormDueDate     string // Kept for backward compatibility
	ActiveDateInput *input.DateInput // New interactive date input component
	ActiveField     int
	Error           error
	Styles          *shared.Styles
}

// RenderCreateForm renders the task creation form
func RenderCreateForm(props CreateFormProps) string {
	s := props.Styles.Title.Render("Create New Task") + "\n\n"

	if props.Error != nil {
		s += props.Styles.HighPriority.Render(fmt.Sprintf("Error: %v\n\n", props.Error))
	}

	// Prepare the due date display format
	// Always include time if available for consistency
	dueDateDisplay := props.FormDueDate
	if props.ActiveDateInput != nil && props.ActiveDateInput.HasValue {
		// Use the full date+time string from the date input for consistency
		dueDateDisplay = props.ActiveDateInput.StringValue()
	}

	// Form fields
	formFields := []struct {
		label    string
		value    string
		active   bool
		required bool
	}{
		{"Title", props.FormTitle, props.ActiveField == 0, true},
		{"Description", props.FormDescription, props.ActiveField == 1, false},
		{"Priority", props.FormPriority, props.ActiveField == 2, false},
		{"Due Date", dueDateDisplay, props.ActiveField == 3, false},
	}

	// Render each field
	for i, field := range formFields {
		// Field label with required indicator
		fieldLabel := field.label
		if field.required {
			fieldLabel += " *"
		}

		// Special handling for due date field
		if i == 3 { // Due Date field
			// Use interactive date input if available
			if props.ActiveField == 3 && props.ActiveDateInput != nil {
				// Set focus state since this field is active
				props.ActiveDateInput.Focused = true
				s += props.ActiveDateInput.View() + "\n\n"
				// Skip the rest of the rendering for this field
				continue
			}
		}

		// Style the label - use blue text color for active fields
		if field.active {
			// Create blue text style for the label without background
			blueLabel := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#0D6EFD")).Render(fieldLabel)
			s += blueLabel + ": "
		} else {
			s += props.Styles.Title.Render(fieldLabel) + ": "
		}

		// Field value with cursor when active
		if field.active {
			// Don't use background color, just add cursor
			s += field.value + "█"
		} else {
			s += field.value
		}

		// Special handling for priority field
		if i == 2 {
			switch props.FormPriority {
			case string(task.PriorityHigh):
				s += " (" + props.Styles.HighPriority.Render(props.FormPriority) + ")"
			case string(task.PriorityMedium):
				s += " (" + props.Styles.MediumPriority.Render(props.FormPriority) + ")"
			default:
				s += " (" + props.Styles.LowPriority.Render(props.FormPriority) + ")"
			}
			s += " - Press Space to cycle"
		}

		s += "\n\n"
	}

	// Submit button
	if props.ActiveField == 4 {
		s += props.Styles.SelectedItem.Render("[Save Task]")
	} else {
		s += "[Save Task]"
	}

	s += "\n\n" + props.Styles.Help.Render("Tab: next field • Enter: submit/cycle date mode • Space: set today's date • ↑↓: change date values • Esc: cancel")

	return s
}
