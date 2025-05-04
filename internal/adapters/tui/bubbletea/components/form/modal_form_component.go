// Package form provides form handling components for the TUI application.
package form

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/shared"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/styles"
	"github.com/newbpydev/tusk/internal/core/task"
	"time"
)

// ModalFormMsg is sent when a modal form action occurs
type ModalFormMsg struct {
	Type    string
	Payload interface{}
}

// ModalFormCloseMsg is sent when the modal form is closed without submitting
type ModalFormCloseMsg struct{}

// ModalFormSubmitMsg is sent when the form is submitted successfully
type ModalFormSubmitMsg struct {
	Task task.Task
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
	DueDate      string
	ParentID     *int32
	TaskID       *int32
	
	// Form state
	FocusedField string
	IsEdit       bool
	FieldIDs     []string             // Ordered list of field IDs for navigation
	FormFields   map[string]FormField // Map of field objects by ID
	Errors       map[string]string    // Legacy error map (for backward compatibility)
	
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
		DueDate:      "",
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
	m.FormFields["dueDate"] = NewFormField("dueDate", "Due Date (YYYY-MM-DD)", &m.DueDate, false)
	
	// Set special validation for date field - need to create a new field to assign function
	datefield := m.FormFields["dueDate"]
	datefield.Validate = func() string {
		if m.DueDate == "" {
			return ""
		}
		_, err := time.Parse("2006-01-02", m.DueDate)
		if err != nil {
			return "Invalid date format. Please use YYYY-MM-DD."
		}
		return ""
	}
	m.FormFields["dueDate"] = datefield
	
	// Priority field is special - it's a cycle field
	priorityField := NewFormField("priority", "Priority", &m.Priority, false)
	priorityField.InputType = "cycle"
	// Create a cycle field with full form width to match other inputs
	priorityField.CycleField = CreatePriorityCycleField(50) // Use the standard form width (matches the one in View method)
	// Set the initial value
	priorityField.CycleField.SetValue(m.Priority)
	m.FormFields["priority"] = priorityField
	
	return m
}

// Init initializes the modal form model
func (m *ModalFormModel) Init() tea.Cmd {
	return nil
}

// Update handles events for the modal form
func (m *ModalFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return ModalFormCloseMsg{} }
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
			// Allow cycling through priorities with space key
			if m.FocusedField == "priority" {
				// Get the cycle field and advance it
				if field, ok := m.FormFields["priority"]; ok && field.CycleField != nil {
					field.CycleField.Next()
					// Update the model's value
					m.UpdateField("priority", field.CycleField.CurrentValue())
					m.FormFields["priority"] = field // Update the field in the map
				}
				return m, nil
			}
			return m, nil
			
		case "enter":
			if m.FocusedField == "save" {
				if m.Validate() {
					newTask := m.CreateTask()
					return m, func() tea.Msg {
						return ModalFormSubmitMsg{Task: newTask}
					}
				}
				// If validation fails, remain on save button
				return m, nil
			} else if m.FocusedField == "cancel" {
				return m, func() tea.Msg { return ModalFormCloseMsg{} }
			} else {
				m.NextField()
				return m, nil
			}
		}

		// Handle field input
		if m.FocusedField == "title" || m.FocusedField == "description" || 
		   m.FocusedField == "dueDate" || m.FocusedField == "priority" {
			
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
	
	if t.DueDate != nil {
		m.DueDate = t.DueDate.Format("2006-01-02")
	} else {
		m.DueDate = ""
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
		m.DueDate = value
	case "priority":
		m.Priority = value
	}
}

// Validate checks if the form data is valid using the FormField objects
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
	
	// Render input fields using FormField objects
	// Only render fields that have a field object (skip buttons and priority which is handled separately)
	for _, fieldID := range m.FieldIDs {
		// Skip rendering button fields and priority - they're handled separately
		if fieldID == "save" || fieldID == "cancel" || fieldID == "priority" {
			continue
		}
		
		// Get field data
		if field, ok := m.FormFields[fieldID]; ok {
			// Render the field
			s += m.renderFormField(field, formWidth)
		}
	}
	
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

// CreateTask creates a task from the form data
func (m *ModalFormModel) CreateTask() task.Task {
	var dueDate *time.Time

	// Convert date string to time.Time if provided
	if m.DueDate != "" {
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
