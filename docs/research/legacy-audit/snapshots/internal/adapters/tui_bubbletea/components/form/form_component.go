// Package form provides form handling components for the TUI application.
// This package follows Go's idiomatic patterns for component design.
package form

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/styles"
	"github.com/newbpydev/tusk/internal/core/task"
)

// FormModel represents a form component with fields and state
type FormModel struct {
	// Form data
	Title        string
	Description  string
	Priority     string
	DueDate      string
	IsCompleted  bool
	ParentID     *int32
	TaskID       *int32
	
	// Form state
	FocusedField string
	IsEdit       bool
	Fields       []string
	Errors       map[string]string
	
	// Component dependencies
	styles       *styles.Styles
}

// NewFormModel creates a new form component
func NewFormModel(styles *styles.Styles) *FormModel {
	fields := []string{
		"title",
		"description",
		"dueDate",
		"priority",
		"isCompleted",
		"save",
		"cancel",
	}
	
	return &FormModel{
		Priority:     string(task.PriorityLow),
		FocusedField: "title",
		Fields:       fields,
		Errors:       make(map[string]string),
		styles:       styles,
	}
}

// Reset clears the form fields
func (m *FormModel) Reset() {
	m.Title = ""
	m.Description = ""
	m.Priority = string(task.PriorityLow)
	m.DueDate = ""
	m.IsCompleted = false
	m.ParentID = nil
	m.TaskID = nil
	m.FocusedField = "title"
	m.Errors = make(map[string]string)
}

// LoadTask loads a task's data into the form
func (m *FormModel) LoadTask(t task.Task) {
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
	
	m.IsCompleted = t.IsCompleted
	m.ParentID = t.ParentID
	m.TaskID = &t.ID
	m.FocusedField = "title"
	m.IsEdit = true
}

// NextField moves focus to the next field
func (m *FormModel) NextField() {
	currentIdx := -1
	for i, field := range m.Fields {
		if field == m.FocusedField {
			currentIdx = i
			break
		}
	}
	
	if currentIdx != -1 && currentIdx < len(m.Fields)-1 {
		m.FocusedField = m.Fields[currentIdx+1]
	} else {
		// Wrap around to the first field
		m.FocusedField = m.Fields[0]
	}
}

// PreviousField moves focus to the previous field
func (m *FormModel) PreviousField() {
	currentIdx := -1
	for i, field := range m.Fields {
		if field == m.FocusedField {
			currentIdx = i
			break
		}
	}
	
	if currentIdx > 0 {
		m.FocusedField = m.Fields[currentIdx-1]
	} else {
		// Wrap around to the last field
		m.FocusedField = m.Fields[len(m.Fields)-1]
	}
}

// ToggleField toggles a boolean field value
func (m *FormModel) ToggleField(field string) {
	if field == "isCompleted" {
		m.IsCompleted = !m.IsCompleted
	}
}

// UpdateField updates a field value
func (m *FormModel) UpdateField(field, value string) {
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

// Validate checks if the form data is valid
func (m *FormModel) Validate() bool {
	m.Errors = make(map[string]string)
	valid := true
	
	// Title is required
	if m.Title == "" {
		m.Errors["title"] = "Title is required"
		valid = false
	}
	
	// Validate date format if provided
	if m.DueDate != "" {
		_, err := time.Parse("2006-01-02", m.DueDate)
		if err != nil {
			m.Errors["dueDate"] = "Invalid date format (use YYYY-MM-DD)"
			valid = false
		}
	}
	
	return valid
}

// View renders the form
func (m *FormModel) View() string {
	var s string
	
	// Form title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#1E88E5"))
	if m.IsEdit {
		s += titleStyle.Render("Edit Task\n\n")
	} else {
		s += titleStyle.Render("Create New Task\n\n")
	}
	
	// Title field
	s += m.renderField("title", "Title", m.Title)
	
	// Description field
	s += m.renderField("description", "Description", m.Description)
	
	// Due date field
	s += m.renderField("dueDate", "Due Date (YYYY-MM-DD)", m.DueDate)
	
	// Priority field
	priorityField := fmt.Sprintf("1 - Low | 2 - Medium | 3 - High (current: %s)", m.Priority)
	s += m.renderField("priority", "Priority", priorityField)
	
	// Is completed checkbox
	checkboxValue := "[ ]"
	if m.IsCompleted {
		checkboxValue = "[x]"
	}
	s += m.renderField("isCompleted", "Completed", checkboxValue)
	
	// Action buttons
	s += "\n"
	s += m.renderButton("save", "Save")
	s += " "
	s += m.renderButton("cancel", "Cancel")
	
	return s
}

// renderField formats a form field with label and value
func (m *FormModel) renderField(id, label, value string) string {
	// Apply styles based on focus state
	dimmedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#747474"))
	focusedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#1E88E5"))
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	
	labelText := dimmedStyle.Render(label)
	var valueText, errorText string
	
	// Format value based on focus
	if m.FocusedField == id {
		valueText = focusedStyle.Render(value)
	} else {
		valueText = normalStyle.Render(value)
	}
	
	// Add error message if present
	if err, ok := m.Errors[id]; ok {
		errorText = errorStyle.Render(" " + err)
	}
	
	// Format the full field line
	field := fmt.Sprintf("%s: %s%s\n", labelText, valueText, errorText)
	
	// Apply highlight to the entire line if focused
	if m.FocusedField == id {
		return focusedStyle.Underline(true).Render(field)
	}
	return field
}

// renderButton formats a button
func (m *FormModel) renderButton(id, label string) string {
	focusedButtonStyle := lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("#1E88E5")).Foreground(lipgloss.Color("#FFFFFF")).Padding(0, 1)
	normalButtonStyle := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1)
	
	if m.FocusedField == id {
		return focusedButtonStyle.Render(label)
	}
	return normalButtonStyle.Render(label)
}

// CreateTask creates a task from the form data
func (m *FormModel) CreateTask(userID int32) task.Task {
	var dueDate *time.Time
	
	// Parse due date if provided
	if m.DueDate != "" {
		parsed, err := time.Parse("2006-01-02", m.DueDate)
		if err == nil {
			dueDate = &parsed
		}
	}
	
	// Create a pointer to the description string
	descPtr := m.Description
	
	// Create task with the provided information
	t := task.Task{
		UserID:      userID,
		Title:       m.Title,
		Description: &descPtr,
		Priority:    task.Priority(m.Priority),
		DueDate:     dueDate,
		Status:      task.StatusTodo,
		IsCompleted: m.IsCompleted,
		ParentID:    m.ParentID,
	}
	
	// If marked as completed, set status accordingly
	if m.IsCompleted {
		t.Status = task.StatusDone
	}

	// For edits, include the task ID
	if m.TaskID != nil {
		t.ID = *m.TaskID
	}
	
	return t
}
