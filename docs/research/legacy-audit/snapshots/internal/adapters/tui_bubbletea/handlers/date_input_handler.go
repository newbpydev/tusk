// Package handlers contains handlers for various UI interactions
package handlers

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/input"
)

// DateInputUpdate is a message indicating that a date input was updated
type DateInputUpdate struct {
	Field string
	Value time.Time
}

// DateInputHandler manages interactions with date input fields
type DateInputHandler struct {
	// inputs maps field names to their date input components
	inputs map[string]*input.DateInput
}

// NewDateInputHandler creates a new date input handler
func NewDateInputHandler() *DateInputHandler {
	return &DateInputHandler{
		inputs: make(map[string]*input.DateInput),
	}
}

// RegisterInput registers a date input field with the handler
func (h *DateInputHandler) RegisterInput(field string, label string) {
	h.inputs[field] = input.NewDateInput(label)
}

// GetInput returns a date input component by field name
func (h *DateInputHandler) GetInput(field string) *input.DateInput {
	return h.inputs[field]
}

// GetValue returns the value of a date input as time.Time
func (h *DateInputHandler) GetValue(field string) *time.Time {
	input, exists := h.inputs[field]
	if !exists || !input.HasValue {
		return nil
	}
	value := input.Value
	return &value
}

// GetValueString returns the string value of a date input
func (h *DateInputHandler) GetValueString(field string) string {
	input, exists := h.inputs[field]
	if !exists || !input.HasValue {
		return ""
	}
	return input.StringValue()
}

// SetValue sets the value of a date input
func (h *DateInputHandler) SetValue(field string, value time.Time) {
	input, exists := h.inputs[field]
	if !exists {
		return
	}
	input.SetValue(value)
}

// SetValueFromString sets the value of a date input from a string
func (h *DateInputHandler) SetValueFromString(field string, value string) error {
	input, exists := h.inputs[field]
	if !exists {
		return fmt.Errorf("date input field %s not found", field)
	}
	return input.SetValueFromString(value)
}

// Focus sets focus on a date input
func (h *DateInputHandler) Focus(field string) {
	for name, input := range h.inputs {
		input.Focused = (name == field)
		// Reset the editing mode when switching focus
		if name != field && input.HasValue {
			// Reset to view mode
			input.Mode = 1 // DateModeView is 1
		}
	}
}

// HandleKey processes a key message for the currently focused date input
func (h *DateInputHandler) HandleKey(msg tea.KeyMsg, focusedField string) tea.Cmd {
	input, exists := h.inputs[focusedField]
	if !exists || !input.Focused {
		return nil
	}

	// Handle input
	input.HandleInput(msg)

	// Return a command with the updated value
	if input.HasValue {
		return func() tea.Msg {
			return DateInputUpdate{
				Field: focusedField,
				Value: input.Value,
			}
		}
	}
	return nil
}

// ResetAllInputs resets all date inputs
func (h *DateInputHandler) ResetAllInputs() {
	for _, input := range h.inputs {
		input.Reset()
	}
}

// Validate validates all date inputs
func (h *DateInputHandler) Validate() map[string]string {
	errors := make(map[string]string)
	
	for _, input := range h.inputs {
		if input.HasValue {
			// Ensure the date is valid
			// Currently time.Time guarantees validity, but we could add more business rules here
		}
	}
	
	return errors
}
