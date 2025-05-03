package app

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/core/task"
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

// handleDateField handles date input using the interactive date input component.
// It processes keyboard events for date selection and manipulation.
func (m *Model) handleDateField(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Focus the due date field in the date input handler
	m.dateInputHandler.Focus("dueDate")
	
	// Handle the key message with the date input handler
	cmd := m.dateInputHandler.HandleKey(msg, "dueDate")
	
	// Update the form's string representation of the date for backward compatibility
	// This ensures existing code continues to work while we transition to the new system
	dateInput := m.dateInputHandler.GetInput("dueDate")
	if dateInput.HasValue {
		m.formDueDate = dateInput.DateString()
	} else {
		m.formDueDate = ""
	}
	
	return m, cmd
}

// handleFormKeys processes keyboard input for both create and edit forms
func (m *Model) handleFormKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// First process any field-specific input
	newModel, cmd := m.handleFormInput(msg)
	if cmd != nil {
		return newModel, cmd
	}
	
	// Then process form navigation and submission
	return m.handleFormNavigationAndSubmit(msg)
}
// handleFormInput processes field-specific input based on the active field
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
		return m.handleDateField(msg)
	// case 4: // Submit button - No direct input handling needed here
	}
	return m, nil
}

// handleFormNavigationAndSubmit processes form navigation and submission actions
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
				m.setErrorStatus("Title is required")
				return m, nil
			}
			// Determine if creating or updating based on viewMode or presence of an ID
			if m.viewMode == "create" { 
				return m, m.createNewTask()
			} else {
				return m, m.updateCurrentTask()
			}
		} else {
			// Move to next field on Enter if not on submit
			m.activeField = (m.activeField + 1) % 5
			return m, nil
		}
	}
	return m, nil // Pass through unhandled keys
}

// resetForm clears all form fields and resets form state
func (m *Model) resetForm() {
	m.formTitle = ""
	m.formDescription = ""
	m.formPriority = string(task.PriorityLow) // Default to low priority
	m.formDueDate = ""
	m.formStatus = ""
	m.activeField = 0
	m.err = nil // Clear any previous form errors
	
	// Reset date input handler if it exists
	if m.dateInputHandler != nil {
		m.dateInputHandler.ResetAllInputs()  
	}
}

// loadTaskIntoForm loads a task's data into the form fields for editing
func (m *Model) loadTaskIntoForm(t task.Task) {
	m.formTitle = t.Title
	
	// Handle potential nil pointer for Description
	if t.Description != nil {
		m.formDescription = *t.Description
	} else {
		m.formDescription = ""
	}
	
	m.formPriority = string(t.Priority)
	
	// Format the date as YYYY-MM-DD if it exists
	if t.DueDate != nil && !t.DueDate.IsZero() {
		// Set the date in both the form field (for backward compatibility)
		// and in the interactive date input component
		m.formDueDate = t.DueDate.Format("2006-01-02")
		m.dateInputHandler.SetValue("dueDate", *t.DueDate)
	} else {
		m.formDueDate = ""
		// Reset the date input component
		m.dateInputHandler.GetInput("dueDate").Reset()
	}
	
	m.formStatus = string(t.Status)
	m.activeField = 0
}

// parseFormData creates a task from the form data
func (m *Model) parseFormData() task.Task {
	// Create a new task with the form data
	description := m.formDescription // Create a local copy for the pointer
	
	t := task.Task{
		Title:       m.formTitle,
		Description: &description,
		Priority:    task.Priority(m.formPriority),
		Status:      task.Status(m.formStatus),
	}
	
	// Get the due date from the date input handler
	dueDate := m.dateInputHandler.GetValue("dueDate")
	if dueDate != nil {
		t.DueDate = dueDate
	} else if m.formDueDate != "" {
		// Fallback to parsing from string (backward compatibility)
		parsedDate, err := parseDate(m.formDueDate)
		if err == nil {
			t.DueDate = &parsedDate
		}
	}
	
	return t
}

// updateCurrentTask updates the current task with form data
func (m *Model) updateCurrentTask() tea.Cmd {
	// First, ensure we have a valid cursor position
	if m.cursor < 0 || m.cursor >= len(m.tasks) || m.cursorOnHeader {
		m.setErrorStatus("No task selected for update")
		return nil
	}
	
	// Get the current task (we'll just log the ID for now but the API doesn't need it)
	_ = m.tasks[m.cursor].ID
	// Add debug info with timestamp to confirm time package is used
	m.setStatusMessage(fmt.Sprintf("Updating task created at %s", time.Now().Format(time.RFC3339)), "info", 5*time.Second)
	
	// Create updated task data from form
	updatedTask := m.parseFormData()

	// Reset form and return to list view
	m.resetForm()
	m.viewMode = "list"
	
	// Update the task in the database
	m.setLoadingStatus("Updating task...")
	return func() tea.Msg {
		// Extract the necessary fields from the updatedTask
		title := updatedTask.Title
		description := ""
		if updatedTask.Description != nil {
			description = *updatedTask.Description
		}
		priority := updatedTask.Priority
		// Pass empty tags for now (or extract from task if needed)
		var tags []string

		// Call the service with individual parameters - the taskID param may vary based on service implementation
		_, err := m.taskSvc.Update(m.ctx, m.userID, title, description, updatedTask.DueDate, priority, tags)
		if err != nil {
			return fmt.Errorf("failed to update task: %w", err)
		}
		
		return m.refreshTasks()
	}
}
