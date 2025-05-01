package app

import (
	tea "github.com/charmbracelet/bubbletea"
	// NOTE: Need to potentially add more imports later
)

// handleInputField handles text input in a generic string field.
// It handles runes, backspace, and checks for navigation keys (Esc, Enter, Tab).
func (m *Model) handleInputField(msg tea.KeyMsg, field *string) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyRunes:
		*field += string(msg.Runes)
		return m, nil // Consume rune input
	case tea.KeyBackspace: // Use Backspace, Delete might have different behavior
		if len(*field) > 0 {
			// Correctly handle multi-byte runes if necessary, though likely okay for simple titles/desc
			*field = (*field)[:len(*field)-1]
		}
		return m, nil // Consume backspace
	case tea.KeyEsc:
		// Let the calling form handler (handleCreateFormKeys/handleEditViewKeys) handle Esc
		return m, nil
	case tea.KeyEnter:
		// Let the calling form handler handle Enter (potentially submit or move to next field)
		return m, nil
	case tea.KeyTab:
		// Let the calling form handler handle Tab (move to next field)
		return m, nil
	case tea.KeyShiftTab:
		// Let the calling form handler handle Shift+Tab (move to prev field)
		return m, nil
	}
	// Allow other key types (like arrows, etc.) to pass through if needed, though they probably don't do anything here
	return m, nil
}

// handleDateField handles date input, allowing only digits and hyphens.
// It also checks for navigation keys.
func (m *Model) handleDateField(msg tea.KeyMsg, field *string) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyRunes:
		// Basic validation: allow only digits and hyphen
		for _, r := range msg.Runes {
			if (r >= '0' && r <= '9') || r == '-' {
				*field += string(r)
			}
		}
		// TODO: Add more robust date format validation (e.g., on Enter/submit)
		return m, nil // Consume valid rune input
	case tea.KeyBackspace:
		if len(*field) > 0 {
			*field = (*field)[:len(*field)-1]
		}
		return m, nil // Consume backspace
	case tea.KeyEsc:
		// Let the calling form handler handle Esc
		return m, nil
	case tea.KeyEnter:
		// Let the calling form handler handle Enter
		return m, nil
	case tea.KeyTab:
		// Let the calling form handler handle Tab
		return m, nil
	case tea.KeyShiftTab:
		// Let the calling form handler handle Shift+Tab
		return m, nil
	}
	return m, nil
}

// Note: The handleCreateFormKeys and handleEditViewKeys functions remain in update.go for now,
// as they orchestrate the form logic (navigation, calling input handlers, submitting).
// We might move the core logic here later and have update.go just call a general "handleFormKeys".

// Example of how form key handling might be structured here eventually:
/*
func (m *Model) handleFormInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.activeField {
	case 0: // Title
		return m.handleInputField(msg, &m.formTitle)
	case 1: // Description
		return m.handleInputField(msg, &m.formDescription)
	case 2: // Priority
		// Handle priority cycling (space bar)
		if msg.String() == " " {
			switch m.formPriority {
			case string(task.PriorityLow):
				m.formPriority = string(task.PriorityMedium)
			case string(task.PriorityMedium):
				m.formPriority = string(task.PriorityHigh)
			default:
				m.formPriority = string(task.PriorityLow)
			}
			return m, nil // Consume space
		}
		// Let navigation keys pass through
		switch msg.Type {
		case tea.KeyRunes, tea.KeyBackspace:
			return m, nil // Consume other input keys
		default:
			return m, nil // Allow navigation keys
		}
	case 3: // Due Date
		return m.handleDateField(msg, &m.formDueDate)
	// case 4: // Submit button - No direct input handling needed here
	}
	return m, nil
}

func (m *Model) handleFormNavigationAndSubmit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		// Reset form state and return to list view
		m.resetForm()
		m.viewMode = "list"
		return m, nil
	case tea.KeyTab:
		m.activeField = (m.activeField + 1) % 5 // 5 fields: Title, Desc, Prio, DueDate, Submit
		return m, nil
	case tea.KeyShiftTab:
		m.activeField = (m.activeField - 1 + 5) % 5 // Wrap around correctly
		return m, nil
	case tea.KeyEnter:
		if m.activeField == 4 { // If on the (virtual) submit button
			if m.formTitle == "" {
				m.err = fmt.Errorf("title is required")
				m.setErrorStatus("Title is required") // Call would be to status.go
				return m, nil
			}
			// Determine if creating or updating based on viewMode or presence of an ID
			if m.viewMode == "create" { // Or check if a task ID is set for editing
				return m, m.createNewTask() // Call would be to tasks.go
			} else {
				// return m, m.updateCurrentTask() // Call would be to tasks.go (needs implementation)
				m.resetForm()
				m.viewMode = "list"
				return m, nil // Placeholder
			}
		} else {
			// Move to next field on Enter if not on submit
			m.activeField = (m.activeField + 1) % 5
			return m, nil
		}
	}
	return m, nil // Pass through unhandled keys
}

// resetForm clears all form fields.
func (m *Model) resetForm() {
	m.formTitle = ""
	m.formDescription = ""
	m.formPriority = "" // Or set to default like task.PriorityLow
	m.formDueDate = ""
	m.formStatus = ""
	m.activeField = 0
	m.err = nil // Clear any previous form errors
}
*/
