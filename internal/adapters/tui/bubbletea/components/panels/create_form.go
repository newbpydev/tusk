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
		{"Due Date (YYYY-MM-DD)", props.FormDueDate, props.ActiveField == 3, false},
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

		if field.active {
			s += props.Styles.SelectedItem.Render(fieldLabel) + ": "
		} else {
			s += props.Styles.Title.Render(fieldLabel) + ": "
		}

		// Field value
		if field.active {
			s += props.Styles.SelectedItem.Render(field.value + "█") // Add cursor
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
