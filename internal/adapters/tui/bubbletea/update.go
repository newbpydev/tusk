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

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/core/task"
)

// Update handles user input and updates the model state.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.viewMode {
	case "list":
		return m.handleListViewKeys(msg)
	case "detail":
		return m.handleDetailViewKeys(msg)
	case "edit":
		return m.handleEditViewKeys(msg)
	case "create":
		return m.handleCreateFormKeys(msg)
	default:
		return m, nil
	}
}

// handleListViewKeys processes keyboard input in list view
func (m *Model) handleListViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

	case "n":
		// Create new task
		m.viewMode = "create"
		// Initialize form with default values
		m.formStatus = string(task.StatusTodo)
		m.formPriority = string(task.PriorityLow)
		return m, nil

	case "1":
		// Toggle task list column
		m.showTaskList = !m.showTaskList
		return m, nil

	case "2":
		// Toggle task details column
		m.showTaskDetails = !m.showTaskDetails
		return m, nil

	case "3":
		// Toggle timeline column
		m.showTimeline = !m.showTimeline
		return m, nil
	}

	return m, nil
}

// handleDetailViewKeys processes keyboard input in detail view
func (m *Model) handleDetailViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

	case "n":
		// Create new task
		m.viewMode = "create"
		// Initialize form with default values
		m.formStatus = string(task.StatusTodo)
		m.formPriority = string(task.PriorityLow)
		return m, nil
	}

	return m, nil
}

// handleCreateFormKeys processes keyboard input in create form
func (m *Model) handleCreateFormKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle high-priority keys first that should work regardless of active field
	switch msg.Type {
	case tea.KeyEsc:
		// Always return to list view, even if there are no tasks
		m.viewMode = "list"

		// Clear form fields when canceling
		m.formTitle = ""
		m.formDescription = ""
		m.formPriority = ""
		m.formDueDate = ""
		m.formStatus = ""
		m.activeField = 0

		return m, nil

	case tea.KeyTab:
		// Tab moves to next field
		m.activeField = (m.activeField + 1) % 5
		return m, nil

	case tea.KeyShiftTab:
		// Shift+Tab moves to previous field
		m.activeField = (m.activeField - 1) % 5
		if m.activeField < 0 {
			m.activeField = 4
		}
		return m, nil

	case tea.KeyEnter:
		if m.activeField == 4 { // Submit button field
			// Validate that title field is not empty before creating
			if m.formTitle == "" {
				m.err = fmt.Errorf("title is required")
				return m, nil
			}

			// Create the task
			return m, m.createNewTask
		}

		// Move to next field
		m.activeField = (m.activeField + 1) % 5
		return m, nil
	}

	// Now handle field-specific input
	switch m.activeField {
	case 0: // Title field
		return m.handleInputField(msg, &m.formTitle)
	case 1: // Description field
		return m.handleInputField(msg, &m.formDescription)
	case 2: // Priority field (cycle through options)
		if msg.String() == " " {
			switch m.formPriority {
			case string(task.PriorityLow):
				m.formPriority = string(task.PriorityMedium)
			case string(task.PriorityMedium):
				m.formPriority = string(task.PriorityHigh)
			default:
				m.formPriority = string(task.PriorityLow)
			}
		}
		return m, nil
	case 3: // Due date field
		return m.handleDateField(msg, &m.formDueDate)
	}

	// Default case
	return m, nil
}

// handleInputField handles text input in a string field
func (m *Model) handleInputField(msg tea.KeyMsg, field *string) (tea.Model, tea.Cmd) {
	// Check specifically for the key type, not the string representation
	switch msg.Type {
	case tea.KeyRunes:
		// Append typed characters to the field
		*field += string(msg.Runes)
		return m, nil
	case tea.KeyBackspace, tea.KeyDelete:
		// Remove the last character if the field is not empty
		if len(*field) > 0 {
			*field = (*field)[:len(*field)-1]
		}
		return m, nil
	case tea.KeyEsc:
		// ESC key is also handled here for better responsiveness
		m.viewMode = "list"

		// Clear form fields when canceling
		m.formTitle = ""
		m.formDescription = ""
		m.formPriority = ""
		m.formDueDate = ""
		m.formStatus = ""
		m.activeField = 0

		return m, nil
	case tea.KeyEnter:
		// Move to the next field or submit if on the last field
		if m.activeField == 4 { // Submit button field
			// Validate that title field is not empty
			if m.formTitle == "" {
				m.err = fmt.Errorf("title is required")
				return m, nil
			}
			// Create the task
			return m, m.createNewTask
		}

		// Move to next field
		m.activeField = (m.activeField + 1) % 5
		return m, nil
	case tea.KeyTab:
		// Tab moves to next field
		m.activeField = (m.activeField + 1) % 5
		return m, nil
	case tea.KeyShiftTab:
		// Shift+Tab moves to previous field
		m.activeField = (m.activeField - 1) % 5
		if m.activeField < 0 {
			m.activeField = 4
		}
		return m, nil
	}
	return m, nil
}

// handleDateField handles date input with basic validation
func (m *Model) handleDateField(msg tea.KeyMsg, field *string) (tea.Model, tea.Cmd) {
	// Check specifically for the key type, not the string representation
	switch msg.Type {
	case tea.KeyRunes:
		// Only allow digits and hyphens for dates
		for _, r := range msg.Runes {
			if (r >= '0' && r <= '9') || r == '-' {
				*field += string(r)
			}
		}
		return m, nil
	case tea.KeyBackspace, tea.KeyDelete:
		// Remove the last character if the field is not empty
		if len(*field) > 0 {
			*field = (*field)[:len(*field)-1]
		}
		return m, nil
	case tea.KeyEsc:
		// ESC key is also handled here for better responsiveness
		m.viewMode = "list"

		// Clear form fields when canceling
		m.formTitle = ""
		m.formDescription = ""
		m.formPriority = ""
		m.formDueDate = ""
		m.formStatus = ""
		m.activeField = 0

		return m, nil
	case tea.KeyEnter:
		// Move to the next field or submit if on the last field
		if m.activeField == 4 { // Submit button field
			// Validate that title field is not empty
			if m.formTitle == "" {
				m.err = fmt.Errorf("title is required")
				return m, nil
			}
			// Create the task
			return m, m.createNewTask
		}

		// Move to next field
		m.activeField = (m.activeField + 1) % 5
		return m, nil
	case tea.KeyTab:
		// Tab moves to next field
		m.activeField = (m.activeField + 1) % 5
		return m, nil
	case tea.KeyShiftTab:
		// Shift+Tab moves to previous field
		m.activeField = (m.activeField - 1) % 5
		if m.activeField < 0 {
			m.activeField = 4
		}
		return m, nil
	}
	return m, nil
}

// handleEditViewKeys processes keyboard input in edit view
func (m *Model) handleEditViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
