package handlers

import (
	tea "github.com/charmbracelet/bubbletea"
)

// FormModelInterface extends AppModel with form-specific operations
type FormModelInterface interface {
	AppModel
	
	// Form operations
	HandleFormSubmit() tea.Cmd
	HandleFormCancel() tea.Cmd
	UpdateFormField(field, value string)
	ToggleFormField(field string)
	
	// Form navigation
	NextFormField()
	PreviousFormField()
	
	// Form state accessors
	GetCurrentFormField() string
	GetFormField(field string) string
	IsFormFieldFocused(field string) bool
}

// HandleFormKeys processes keyboard input in form view (create/edit)
func HandleFormKeys(m FormModelInterface, msg tea.KeyMsg) (tea.Cmd, bool) {
	// Handle common form navigation keys first
	switch msg.String() {
	case "esc":
		// Cancel and return to list view
		return m.HandleFormCancel(), true
		
	case "tab", "down", "j":
		// Move to next field
		m.NextFormField()
		return nil, true
		
	case "shift+tab", "up", "k":
		// Move to previous field
		m.PreviousFormField()
		return nil, true
		
	case "enter":
		// Different behavior based on current field
		currentField := m.GetCurrentFormField()
		
		if currentField == "save" {
			// Submit form
			return m.HandleFormSubmit(), true
		} else if currentField == "cancel" {
			// Cancel form
			return m.HandleFormCancel(), true
		} else if currentField == "isPriority" || currentField == "isCompleted" {
			// Toggle boolean fields
			m.ToggleFormField(currentField)
			// Move to next field
			m.NextFormField()
			return nil, true
		} else {
			// On regular fields, just move to next field
			m.NextFormField()
			return nil, true
		}
	}
	
	// Handle field-specific input
	currentField := m.GetCurrentFormField()
	
	// Only process character input if a text field is focused
	if currentField == "title" || currentField == "description" || currentField == "dueDate" {
		// Process printable characters as field input
		// This would be handled by the text input components
		// and we only need to pass it along
		return nil, false // Let the character be processed by the text input
	}
	
	// Handle priority field selection
	if currentField == "priority" {
		switch msg.String() {
		case "1", "l", "L":
			m.UpdateFormField("priority", "low")
			return nil, true
		case "2", "m", "M":
			m.UpdateFormField("priority", "medium")
			return nil, true
		case "3", "h", "H":
			m.UpdateFormField("priority", "high")
			return nil, true
		}
	}
	
	return nil, false
}
