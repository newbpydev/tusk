// Package form provides form handling components for the TUI application.
package form

import (
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/utils"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/shared"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/messages"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/styles"
	"github.com/newbpydev/tusk/internal/core/task"
	"time"
)

// ModalFormMsg is sent when a modal form action occurs
type ModalFormMsg struct {
	Type    string
	Payload interface{}
}

// FormField represents a field in a form with validation and rendering capabilities
type FormField struct {
	ID          string           // Unique identifier for the field
	Label       string           // Display label
	Value       *string          // Pointer to the actual value (from parent form)
	Required    bool             // Whether the field is required
	Validate    func() string    // Custom validation function, returns error message or empty string
	ErrorMsg    string           // Current error message (empty means no error)
	InputType   string           // Type of input (text, date, select, etc.)
	Options     []string         // Available options for select-type fields
	CycleField  *shared.CycleField // For cycle-type fields
}

// NewFormField creates a new form field with the given properties
func NewFormField(id string, label string, value *string, required bool) FormField {
	return FormField{
		ID:        id,
		Label:     label,
		Value:     value,
		Required:  required,
		ErrorMsg:  "",
		InputType: "text", // Default to text input
		Options:   []string{},
		Validate:  nil,
	}
}

// HasError returns true if the field has a validation error
func (f FormField) HasError() bool {
	return f.ErrorMsg != ""
}

// Validate runs validation on the field and returns whether it's valid
func (f *FormField) ValidateField() bool {
	// Clear any existing error
	f.ErrorMsg = ""
	
	// Check if required and empty
	if f.Required && (*f.Value == "") {
		f.ErrorMsg = f.Label + " is required"
		return false
	}
	
	// Run custom validation if provided
	if f.Validate != nil {
		if err := f.Validate(); err != "" {
			f.ErrorMsg = err
			return false
		}
	}
	
	return true
}

// ModalFormModel represents a form component specifically designed for modal display
type ModalFormModel struct {
	// Form data - keeping these for backward compatibility
	Title        string
	Description  string
	Priority     string
	DueDate      string  // Kept for backward compatibility
	ParentID     *int32
	TaskID       *int32
	
	// Form state
	FocusedField string
	IsEdit       bool
	FieldIDs     []string             // Ordered list of field IDs for navigation
	FormFields   map[string]FormField // Map of field objects by ID
	Errors       map[string]string    // Legacy error map (for backward compatibility)
	
	// UI components
	DueDateField DateField          // New component for due date input
	
	// Component dependencies
	styles       *styles.Styles
	UserID       int32
}

// CreatePriorityCycleField creates a cycle field specifically for task priorities
func CreatePriorityCycleField(width int) *shared.CycleField {
	// Create options with appropriate colors
	options := []shared.CycleOption{
		{Value: string(task.PriorityLow), Label: "low", Color: "#4CAF50"}, // Green
		{Value: string(task.PriorityMedium), Label: "medium", Color: "#FF9800"}, // Orange
		{Value: string(task.PriorityHigh), Label: "high", Color: "#F44336"}, // Red
	}
	
	// Create and return the cycle field
	return shared.NewCycleField(options, width)
}

// NewModalFormModel creates a new modal form model
func NewModalFormModel(styles *styles.Styles, userID int32) *ModalFormModel {
	// Create new form model
	m := &ModalFormModel{
		// Default values
		Title:        "",
		Description:  "",
		Priority:     string(task.PriorityLow),
		DueDate:      "", // Kept for backward compatibility
		ParentID:     nil,
		TaskID:       nil,
		
		// Initial state
		FocusedField: "title",
		IsEdit:       false,
		FieldIDs:     []string{"title", "description", "dueDate", "priority", "save", "cancel"},
		FormFields:   make(map[string]FormField),
		Errors:       make(map[string]string),
		
		// Dependencies
		styles:       styles,
		UserID:       userID,
	}
	
	// Initialize form fields with pointers to the model's properties
	m.FormFields["title"] = NewFormField("title", "Title", &m.Title, true)
	m.FormFields["description"] = NewFormField("description", "Description", &m.Description, false)
	
	// Initialize the interactive date field component
	m.DueDateField = NewDateField("Due Date", "dueDate", *styles, false)
	
	// Keep a legacy dueDate field in the map for backward compatibility
	m.FormFields["dueDate"] = NewFormField("dueDate", "Due Date", &m.DueDate, false)
	
	// Priority field is special - it's a cycle field
	priorityField := NewFormField("priority", "Priority", &m.Priority, false)
	priorityField.InputType = "cycle"
	// Create a cycle field with full form width to match other inputs
	priorityField.CycleField = CreatePriorityCycleField(50) // Use the standard form width (matches the one in View method)
	// Set the initial value
	priorityField.CycleField.SetValue(m.Priority)
	m.FormFields["priority"] = priorityField
	
	// Add button fields to the form fields map so they're properly tracked
	// This is important for proper focus handling and event processing
	
	// Create save button field
	saveField := NewFormField("save", "Save", nil, false)
	saveField.InputType = "button"
	m.FormFields["save"] = saveField
	
	// Create cancel button field
	cancelField := NewFormField("cancel", "Cancel", nil, false)
	cancelField.InputType = "button"
	m.FormFields["cancel"] = cancelField
	
	return m
}

// Init initializes the modal form model
func (m *ModalFormModel) Init() tea.Cmd {
	return nil
}

// Update handles events for the modal form
func (m *ModalFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MouseMsg:
		if msg.Type == tea.MouseLeft {
			// Handle Cancel button click
			if m.isClickOnCancelButton(msg) {
				// Log cancel button click
				utils.DebugLog("CANCEL BUTTON: Mouse click detected")
				
				// Focus the cancel button and send close message
				m.FocusedField = "cancel"
				
				// Generate modal close message (pure function call, no batching)
				return m, func() tea.Msg {
					closeMsg := messages.ModalFormCloseMsg{}
					utils.DebugLog("CANCEL BUTTON: Generated message %+v", closeMsg)
					return closeMsg
				}
			}
			
			// Handle Save button click
			if m.isClickOnSaveButton(msg) {
				// Log save button click
				utils.DebugLog("SAVE BUTTON: Mouse click detected")
				
				// Focus the save button
				m.FocusedField = "save"
				
				// Validate before submitting
				if m.Validate() {
					// Create task from form data
					newTask := m.CreateTask()
					// Generate submit message with task data
					return m, func() tea.Msg {
						submitMsg := messages.ModalFormSubmitMsg{Task: newTask}
						utils.DebugLog("SAVE BUTTON: Generated message %+v", submitMsg)
						return submitMsg
					}
				}
			}
		}
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			// Simple direct message for ESC key
			return m, func() tea.Msg {
				return messages.ModalFormCloseMsg{}
			}
		case "tab":
			m.NextField()
			return m, nil
		case "shift+tab":
			m.PreviousField()
			return m, nil
		case "up":
			m.PreviousField()
			return m, nil
		case "down":
			m.NextField()
			return m, nil
		case "left":
			// Handle left arrow for priority selection
			if m.FocusedField == "priority" {
				switch m.Priority {
				case string(task.PriorityMedium):
					m.UpdateField("priority", string(task.PriorityLow))
				case string(task.PriorityHigh):
					m.UpdateField("priority", string(task.PriorityMedium))
				}
			}
			return m, nil
		case "right":
			// Handle right arrow for priority selection
			if m.FocusedField == "priority" {
				switch m.Priority {
				case string(task.PriorityLow):
					m.UpdateField("priority", string(task.PriorityMedium))
				case string(task.PriorityMedium):
					m.UpdateField("priority", string(task.PriorityHigh))
				}
			}
			return m, nil
		case " ":
			// Handle different space key actions based on focused field
			switch m.FocusedField {
			case "dueDate":
				// Handle date field input through the DateField component
				updatedField, _ := m.DueDateField.Update(msg)
				m.DueDateField = updatedField
				
				// Update legacy string field for compatibility
				if dueDate := m.DueDateField.Value(); dueDate != nil {
					m.DueDate = dueDate.Format("2006-01-02")
				} else {
					m.DueDate = ""
				}
				return m, nil
				
			case "title", "description":
				// For standard text fields, add space to the field value
				m.UpdateField(m.FocusedField, m.GetField(m.FocusedField)+" ")
				return m, nil
				
			case "priority":
				// For priority field, cycle through options
				m.CyclePriority()
				return m, nil
				
			case "status":
				// For status field, cycle through options
				m.CycleStatus()
				return m, nil
				
			case "cancel":
				// Handle space on Cancel button - same as Enter
				return m, func() tea.Msg {
					return messages.ModalFormCloseMsg{}
				}
				
			case "save":
				// Handle space on Save button - same as Enter
				if m.Validate() {
					newTask := m.CreateTask()
					return m, func() tea.Msg {
						return messages.ModalFormSubmitMsg{Task: newTask}
					}
				}
				return m, nil // Validation failed
			}
			return m, nil
			
		case "enter":
			// Handle Save button enter press
			if m.FocusedField == "save" {
				// Validate before submitting
				if m.Validate() {
					// Create task and return submit message
					newTask := m.CreateTask()
					return m, func() tea.Msg {
						return messages.ModalFormSubmitMsg{Task: newTask}
					}
				}
				return m, nil // Validation failed
			} else if m.FocusedField == "cancel" {
				// Handle Cancel button enter press
				return m, func() tea.Msg {
					return messages.ModalFormCloseMsg{}
				}
			} else {
				m.NextField()
				return m, nil
			}
		}

		// Handle field input
		if m.FocusedField == "dueDate" {
			// Handle date field input through the specialized component
			updatedField, _ := m.DueDateField.Update(msg)
			m.DueDateField = updatedField
			
			// Update legacy string field for compatibility
			if dueDate := m.DueDateField.Value(); dueDate != nil {
				m.DueDate = dueDate.Format("2006-01-02")
			} else {
				m.DueDate = ""
			}
			return m, nil
		} else if m.FocusedField == "title" || m.FocusedField == "description" || 
		        m.FocusedField == "priority" {
			
			switch msg.String() {
			case "backspace":
				field := m.FocusedField
				value := m.GetField(field)
				if len(value) > 0 {
					m.UpdateField(field, value[:len(value)-1])
				}
			case "1", "2", "3":
				if m.FocusedField == "priority" {
					switch msg.String() {
					case "1":
						m.UpdateField("priority", string(task.PriorityLow))
					case "2":
						m.UpdateField("priority", string(task.PriorityMedium))
					case "3":
						m.UpdateField("priority", string(task.PriorityHigh))
					}
				} else {
					m.UpdateField(m.FocusedField, m.GetField(m.FocusedField)+msg.String())
				}
			default:
				// Only add character if it's a printable character (not a control character)
				if len(msg.String()) == 1 {
					m.UpdateField(m.FocusedField, m.GetField(m.FocusedField)+msg.String())
				}
			}
			return m, nil
		}
	}
	
	return m, nil
}

// GetField returns the value of a field
func (m *ModalFormModel) GetField(field string) string {
	switch field {
	case "title":
		return m.Title
	case "description":
		return m.Description
	case "dueDate":
		return m.DueDate
	case "priority":
		return m.Priority
	default:
		return ""
	}
}

// Reset clears the form fields
func (m *ModalFormModel) Reset() {
	m.Title = ""
	m.Description = ""
	m.Priority = string(task.PriorityLow)
	m.DueDate = ""
	m.ParentID = nil
	m.TaskID = nil
	m.FocusedField = "title"
	m.Errors = make(map[string]string)
}

// LoadTask loads a task's data into the form
func (m *ModalFormModel) LoadTask(t task.Task) {
	m.Title = t.Title
	// Handle nil Description pointer
	if t.Description != nil {
		m.Description = *t.Description
	} else {
		m.Description = ""
	}
	m.Priority = string(t.Priority)
	
	// Set both the legacy string field and our new component
	if t.DueDate != nil {
		// Set the legacy string field for backward compatibility
		m.DueDate = t.DueDate.Format("2006-01-02")
		
		// Set the value in our new date field component
		m.DueDateField.SetValue(t.DueDate)
	} else {
		m.DueDate = ""
		m.DueDateField.SetValue(nil)
	}
	
	m.ParentID = t.ParentID
	m.TaskID = &t.ID
	m.FocusedField = "title"
	m.IsEdit = true
}

// NextField moves focus to the next field
func (m *ModalFormModel) NextField() {
	currentIdx := -1
	for i, field := range m.FieldIDs {
		if field == m.FocusedField {
			currentIdx = i
			break
		}
	}
	
	// Move to the next field, wrapping around if needed
	if currentIdx >= 0 {
		nextIdx := (currentIdx + 1) % len(m.FieldIDs)
		m.FocusedField = m.FieldIDs[nextIdx]
	} else {
		// If no field is focused, focus the first one
		m.FocusedField = m.FieldIDs[0]
	}
	
	// Skip the completed field as it's been removed
	if m.FocusedField == "completed" {
		m.NextField()
	}
}

// PreviousField moves focus to the previous field
func (m *ModalFormModel) PreviousField() {
	currentIdx := -1
	for i, field := range m.FieldIDs {
		if field == m.FocusedField {
			currentIdx = i
			break
		}
	}
	
	// Move to the previous field, wrapping around if needed
	if currentIdx > 0 {
		m.FocusedField = m.FieldIDs[currentIdx-1]
	} else {
		// Wrap around to the last field
		m.FocusedField = m.FieldIDs[len(m.FieldIDs)-1]
	}
	
	// Skip the completed field as it's been removed
	if m.FocusedField == "completed" {
		m.PreviousField()
	}
}

// ToggleField toggles a boolean field value
func (m *ModalFormModel) ToggleField(field string) {
	// No boolean fields after removing IsCompleted
}

// UpdateField updates a field value
func (m *ModalFormModel) UpdateField(field, value string) {
	switch field {
	case "title":
		m.Title = value
	case "description":
		m.Description = value
	case "dueDate":
		// Update legacy string field
		m.DueDate = value
		
		// Also update the field component if it's a valid date
		if value != "" {
			// Try to parse the date
			if parsedDate, err := time.Parse("2006-01-02", value); err == nil {
				m.DueDateField.SetValue(&parsedDate)
			}
		} else {
			// Empty string, clear the field
			m.DueDateField.SetValue(nil)
		}
	case "priority":
		m.Priority = value
	}
}

// Validate checks all form fields and returns whether the form is valid
func (m *ModalFormModel) Validate() bool {
	valid := true
	m.Errors = make(map[string]string)
	
	// Validate each field that has a validation rule
	for id, field := range m.FormFields {
		// Skip button fields
		if id == "save" || id == "cancel" {
			continue
		}
		
		// Make a copy of the field so we can modify it
		fieldCopy := field
		
		// Run validation
		if !fieldCopy.ValidateField() {
			// If validation failed, store the error message
			m.Errors[id] = fieldCopy.ErrorMsg
			valid = false
			
			// Update the field with the error
			m.FormFields[id] = fieldCopy
		}
	}
	
	// Validate the date field component
	if !m.DueDateField.Validate() {
		// Store the error in our error map
		m.Errors["dueDate"] = m.DueDateField.Error
		valid = false
	}
	
	return valid
}

// View renders the form
func (m *ModalFormModel) View() string {
	var s string
	
	// Apply modal-specific styling with softer colors and more padding
	
	// Define consistent form width to be used throughout
	formWidth := 50

	// Form title - more prominent for modal
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#1E88E5")).
		MarginBottom(1).
		Padding(0, 0, 1, 0).
		Width(formWidth).
		Align(lipgloss.Center)
		
	if m.IsEdit {
		s += titleStyle.Render("Edit Task")
	} else {
		s += titleStyle.Render("Create New Task")
	}
	
	// Add some spacing after the title instead of a divider
	s += "\n"
	
	// Add a bit of spacing after the title
	s += "\n"
	
	// Use the already defined formWidth from above
	
	// Render standard text input fields using FormField objects
	for _, fieldID := range m.FieldIDs {
		// Skip special fields that are handled separately
		if fieldID == "save" || fieldID == "cancel" || fieldID == "priority" || fieldID == "dueDate" {
			continue
		}
		
		// Get field data
		if field, ok := m.FormFields[fieldID]; ok {
			// Render the field
			s += m.renderFormField(field, formWidth)
		}
	}
	
	// Add due date field using our new component
	// Update focus state based on current field focus
	if m.FocusedField == "dueDate" {
		m.DueDateField.Focus()
	} else {
		m.DueDateField.Blur()
	}
	
	// Pass any validation errors to the date field
	if errMsg, hasError := m.Errors["dueDate"]; hasError {
		m.DueDateField.Error = errMsg
	} else {
		m.DueDateField.Error = ""
	}
	
	// Add the date field to the form
	// Position just after the text inputs, before the priority field
	s += m.DueDateField.View() + "\n"
	
	// Priority is rendered as a cycle field
	if priorityField, ok := m.FormFields["priority"]; ok && priorityField.CycleField != nil {
		// Create label with special handling for errors
		displayLabel := priorityField.Label + " (press Space to cycle)"
		if errMsg, hasError := m.Errors["priority"]; hasError {
			displayLabel = priorityField.Label + ": " + errMsg
		}
		
		// Label style based on error state
		labelStyle := lipgloss.NewStyle().
			Width(formWidth).
			Align(lipgloss.Left)
		
		if _, hasError := m.Errors["priority"]; hasError {
			labelStyle = labelStyle.Foreground(lipgloss.Color("#F44336")).Bold(true)
		} else {
			labelStyle = labelStyle.Foreground(lipgloss.Color("#555555"))
		}
		
		// Render the label
		s += labelStyle.Render(displayLabel) + "\n"
		
		// Update cycle field state
		priorityField.CycleField.SetValue(m.Priority) // Ensure current value is reflected
		priorityField.CycleField.Focused = m.FocusedField == "priority"
		
		// Check if there's an error for the priority field
		_, hasError := m.Errors["priority"]
		priorityField.CycleField.HasError = hasError
		
		// Render using the CycleField component
		s += priorityField.CycleField.View() + "\n"
	}
	
	// Action buttons in a row with enhanced styling
	buttonsStyle := lipgloss.NewStyle().
		Width(formWidth).
		Align(lipgloss.Center).
		Padding(2, 0)
	
	saveButton := m.renderButton("save", "Save")
	cancelButton := m.renderButton("cancel", "Cancel")
	
	buttons := lipgloss.JoinHorizontal(
		lipgloss.Center,
		saveButton,
		lipgloss.NewStyle().Width(4).Render(""),
		cancelButton,
	)
	
	s += buttonsStyle.Render(buttons)
	
	// Wrap the whole form in a container style
	containerStyle := lipgloss.NewStyle().
		Padding(1, 2).
		BorderForeground(lipgloss.Color("#BBDEFB"))
	
	return containerStyle.Render(s)
}

// renderFormField formats a form field using the FormField object
func (m *ModalFormModel) renderFormField(field FormField, width int) string {
	var s string
	
	// Determine if field has an error
	errMsg, hasError := m.Errors[field.ID]
	
	// Prepare the label text
	displayLabel := field.Label
	if hasError {
		// Add error message to the label
		displayLabel = field.Label + ": " + errMsg
	}
	
	// Style for label
	labelStyle := lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Left)
	
	if hasError {
		// Error state - red text with bold
		labelStyle = labelStyle.Foreground(lipgloss.Color("#F44336")).Bold(true)
	} else {
		// Normal state - gray text
		labelStyle = labelStyle.Foreground(lipgloss.Color("#555555"))
	}
	
	// Render the label
	s += labelStyle.Render(displayLabel)
	s += "\n"
	
	// Style the input field - red border for errors, blue for focus, gray for normal
	inputFieldStyle := lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.RoundedBorder())
	
	// Border color based on state
	if hasError {
		// Error state - red border
		inputFieldStyle = inputFieldStyle.BorderForeground(lipgloss.Color("#F44336"))
	} else if m.FocusedField == field.ID {
		// Focused state - blue border
		inputFieldStyle = inputFieldStyle.BorderForeground(lipgloss.Color("#2196F3"))
	} else {
		// Normal state - gray border
		inputFieldStyle = inputFieldStyle.BorderForeground(lipgloss.Color("#CCCCCC"))
	}
	
	// Add padding and margin
	inputFieldStyle = inputFieldStyle.Padding(0, 1).MarginBottom(1)
		
	// Add cursor indicator for focused field
	displayValue := ""
	if field.Value != nil {
		displayValue = *field.Value
	}
	
	if m.FocusedField == field.ID {
		displayValue += "█"
	}
	
	// Render the input field with value
	s += inputFieldStyle.Render(displayValue)
	s += "\n"
	
	// Add help text for certain field types
	if field.InputType == "date" && field.ID == "dueDate" {
		helpStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9E9E9E")).
			Italic(true).
			MarginLeft(2).
			MarginBottom(1).
			Width(width - 4)
		
		s += helpStyle.Render("Format: YYYY-MM-DD")
		s += "\n"
	}
	
	return s
}

// renderButton formats a button using the shared Button component
func (m *ModalFormModel) renderButton(id, label string) string {
	buttonStyle := shared.DefaultButtonStyle()
	buttonStyle.Focused = m.FocusedField == id
	buttonStyle.Primary = id == "save"
	buttonStyle.Width = 12 // Fixed width to prevent layout shifts
	buttonStyle.FixedWidth = true
	
	return shared.Button(label, buttonStyle)
}

// isClickOnCancelButton checks if a mouse click is within the cancel button area
func (m *ModalFormModel) isClickOnCancelButton(msg tea.MouseMsg) bool {
	// Use the form dimensions to compute button positions
	const (
		formWidth    = 50   // Standard form width
		buttonWidth  = 15   // Button width including padding
		buttonHeight = 1    // Button height
		buttonGap    = 5    // Gap between buttons
	)
	
	// Calculate buttons row position
	// The button row is always at the bottom of all form fields plus a margin
	buttonsRowY := len(m.FieldIDs) + 2 // Fields + margin
	
	// Cancel button is on the left side of the button row
	cancelX1 := 5                     // Left position
	cancelX2 := cancelX1 + buttonWidth // Right position
	cancelY  := buttonsRowY            // Y position
	
	// Check if click is within cancel button bounds
	if msg.Y == cancelY && msg.X >= cancelX1 && msg.X < cancelX2 {
		return true
	}
	
	return false
}

// isClickOnSaveButton checks if a mouse click is within the save button area
func (m *ModalFormModel) isClickOnSaveButton(msg tea.MouseMsg) bool {
	// Use the form dimensions to compute button positions
	const (
		formWidth    = 50   // Standard form width
		buttonWidth  = 15   // Button width including padding
		buttonHeight = 1    // Button height
		buttonGap    = 5    // Gap between buttons
	)
	
	// Calculate buttons row position - same as cancel button
	buttonsRowY := len(m.FieldIDs) + 2 // Fields + margin
	
	// Save button is on the right side of the button row
	saveX1 := 25                    // Left position (after cancel button + gap)
	saveX2 := saveX1 + buttonWidth  // Right position
	saveY  := buttonsRowY           // Y position (same row as cancel)
	
	// Check if click is within save button bounds
	if msg.Y == saveY && msg.X >= saveX1 && msg.X < saveX2 {
		return true
	}
	
	return false
}

// CyclePriority cycles through the available priority options
func (m *ModalFormModel) CyclePriority() {
	// Get the cycle field and advance it
	if field, ok := m.FormFields["priority"]; ok && field.CycleField != nil {
		field.CycleField.Next()
		// Update the model's value
		m.UpdateField("priority", field.CycleField.CurrentValue())
		m.FormFields["priority"] = field // Update the field in the map
	}
}

// CycleStatus cycles through the available status options
func (m *ModalFormModel) CycleStatus() {
	// Get the cycle field and advance it
	if field, ok := m.FormFields["status"]; ok && field.CycleField != nil {
		field.CycleField.Next()
		// Update the model's value
		m.UpdateField("status", field.CycleField.CurrentValue())
		m.FormFields["status"] = field // Update the field in the map
	}
}

// CreateTask creates a task from the form data
func (m *ModalFormModel) CreateTask() task.Task {
	// Get due date directly from our date field component
	dueDate := m.DueDateField.Value()
	
	// For backward compatibility, also try the string field if component has no value
	if dueDate == nil && m.DueDate != "" {
		parsed, err := time.Parse("2006-01-02", m.DueDate)
		if err == nil {
			dueDate = &parsed
		}
	}

	// Convert description to pointer
	var descPtr *string
	if m.Description != "" {
		descPtr = &m.Description
	}

	// Create new task
	t := task.Task{
		Title:       m.Title,
		Description: descPtr,
		Priority:    task.Priority(m.Priority),
		DueDate:     dueDate,
		IsCompleted: false, // New tasks are not completed by default
		ParentID:    m.ParentID,
		UserID:      m.UserID,
	}

	// For edit, include the task ID
	if m.IsEdit && m.TaskID != nil {
		t.ID = *m.TaskID
	}

	return t
}
