// Package form provides form components for the TUI
package form

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/input"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/styles"
)

// DateField is a reusable date input field component for forms that uses the
// DateInput component for interactive date selection. It provides a proper form field
// interface with focus management, validation, and error handling.
type DateField struct {
	// DateInput is the underlying date input component that handles the actual date selection logic
	DateInput *input.DateInput
	
	// Style settings from the application theme
	Styles   styles.Styles
	
	// State management
	Focused  bool     // Whether this field currently has focus
	Required bool     // Whether this field is required for form submission
	Error    string   // Current validation error message (empty if no error)
	
	// Field metadata
	Label     string  // Display label for the field
	FieldName string  // Internal field identifier
}

// NewDateField creates a new date field component with the specified label and field name.
// It initializes the underlying DateInput component and sets up proper styling and validation requirements.
func NewDateField(label, fieldName string, styles styles.Styles, required bool) DateField {
	// Create the underlying date input with the provided label
	dateInput := input.NewDateInput(label)
	
	// Initialize with empty error and proper configuration
	return DateField{
		DateInput: dateInput,
		Styles:    styles,
		Label:     label,
		FieldName: fieldName,
		Required:  required,
		Error:     "", // Start with no error
	}
}

// Focus sets focus on this field, enabling keyboard input and visual highlighting.
// This affects both the DateField container and its inner DateInput component.
func (d *DateField) Focus() {
	d.Focused = true
	d.DateInput.Focused = true
}

// Blur removes focus from this field, disabling keyboard input and visual highlighting.
// This affects both the DateField container and its inner DateInput component.
func (d *DateField) Blur() {
	d.Focused = false
	d.DateInput.Focused = false
}

// SetValue sets the field value from a time.Time pointer.
// If nil is provided, the field will be reset to an empty state.
// Otherwise, the date value will be set in the underlying DateInput component.
func (d *DateField) SetValue(t *time.Time) {
	if t == nil {
		// Clear the field when nil is provided
		d.DateInput.Reset()
		return
	}
	// Set the date value in the underlying component
	d.DateInput.SetValue(*t)
}

// Value returns the current field value as a *time.Time.
// Returns nil if the field has no value or invalid date.
func (d *DateField) Value() *time.Time {
	// If the underlying component has no value, return nil
	if !d.DateInput.HasValue {
		return nil
	}
	value := d.DateInput.Value
	return &value
}

// StringValue returns the current date as a formatted string in YYYY-MM-DD format.
// If the field has no value, an empty string is returned.
func (d *DateField) StringValue() string {
	return d.DateInput.StringValue()
}

// Validate checks if the field is valid according to its configuration.
// It sets appropriate error messages and returns false if validation fails.
// For required fields, it checks if a value is present.
func (d *DateField) Validate() bool {
	// Check if required field has a value
	if d.Required && !d.DateInput.HasValue {
		d.Error = "This field is required"
		return false
	}
	
	// Validation passed, clear any error
	d.Error = ""
	return true
}

// Update handles events for the date field and returns an updated model and command.
// Key events are passed to the underlying DateInput component when the field is focused.
func (d *DateField) Update(msg tea.Msg) (DateField, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if d.Focused {
			switch msg.String() {
			case "left", "right", "up", "down", "enter", "backspace":
				// Handle navigation keys for date manipulation
				d.DateInput.HandleInput(msg)
				
			case "tab", "shift+tab", "esc":
				// Let these pass through to the parent component
				// They will be handled by the form for field navigation
				return *d, nil
				
			case " ":
				// Special handling for space - set to today's date if empty
				if !d.DateInput.HasValue {
					d.DateInput.SetToToday()
				}
				return *d, nil
				
			default:
				// Pass other keystrokes to the date input
				d.DateInput.HandleInput(msg)
			}
			
			// Return a copy of this field to avoid unexpected mutations
			return *d, nil
		}
	}
	// No update for this message type
	return *d, nil
}

// View renders the date field with proper styling based on focus and error states.
// It formats the field to match other form fields with a separate label and
// consistent styling for the input area.
func (d *DateField) View() string {
	// Initialize result string
	var s string
	
	// Create label with instructions in parentheses - exactly like priority field
	displayLabel := d.Label + " (press ← → to select)"
	if d.Error != "" {
		// Show error message instead of instructions
		displayLabel = d.Label + ": " + d.Error
	}
	
	// Style for label - match the modal form style exactly
	labelStyle := lipgloss.NewStyle().
		Width(50).  // Standard form width
		Align(lipgloss.Left)
	
	if d.Error != "" {
		// Error state - red text with bold
		labelStyle = labelStyle.Foreground(lipgloss.Color("#F44336")).Bold(true)
	} else {
		// Normal state - gray text
		labelStyle = labelStyle.Foreground(lipgloss.Color("#555555"))
	}
	
	// Render the label
	s += labelStyle.Render(displayLabel)
	s += "\n"
	
	// Get value from DateInput and render in styled input box
	value := d.DateInput.StringValue()
	
	// Style the input field - red border for errors, blue for focus, gray for normal
	inputFieldStyle := lipgloss.NewStyle().
		Width(50). // Standard form width 
		Border(lipgloss.RoundedBorder())
	
	// Border color based on state
	if d.Error != "" {
		// Error state - red border
		inputFieldStyle = inputFieldStyle.BorderForeground(lipgloss.Color("#F44336"))
	} else if d.Focused {
		// Focused state - blue border
		inputFieldStyle = inputFieldStyle.BorderForeground(lipgloss.Color("#2196F3")) 
	} else {
		// Normal state - gray border
		inputFieldStyle = inputFieldStyle.BorderForeground(lipgloss.Color("#CCCCCC"))
	}
	
	// Add padding and margin
	inputFieldStyle = inputFieldStyle.Padding(0, 1).MarginBottom(1)

	// Add cursor indicator for focused field
	if d.Focused {
		value += "█" // Block cursor 
	}
	
	// Render the input field with value
	s += inputFieldStyle.Render(value)
	s += "\n" // Only one newline to match other form fields
	
	return s
}

// Reset clears the field value and any error messages.
// It delegates to the underlying DateInput component's Reset method.
func (d *DateField) Reset() {
	// Clear the underlying component
	d.DateInput.Reset()
	// Also clear any error message
	d.Error = ""
}
