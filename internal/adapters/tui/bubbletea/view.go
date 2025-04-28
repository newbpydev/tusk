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

package bubbletea

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/tusk/internal/core/task"
)

// renderHeader creates a header with app name, time, and status information
func (m *Model) renderHeader() string {
	headerHeight := 3

	// Create a style for the entire header
	headerStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(headerHeight).
		Padding(0, 1).
		Background(lipgloss.Color("#2d3748")).
		Foreground(lipgloss.Color("#ffffff"))

	// Left section - App logo and tagline
	logoStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#48bb78")).
		PaddingLeft(2)
	logo := logoStyle.Render("TUSK")

	taglineStyle := lipgloss.NewStyle().
		Italic(true).
		Foreground(lipgloss.Color("#a0aec0"))
	tagline := taglineStyle.Render("Task Management Simplified")

	leftSection := lipgloss.JoinVertical(
		lipgloss.Left,
		logo,
		tagline,
	)

	// Middle section - Current time and date
	timeStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Bold(true)
	dateStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("#a0aec0"))

	timeDisplay := timeStyle.Render(m.currentTime.Format("15:04:05"))
	dateDisplay := dateStyle.Render(m.currentTime.Format("Monday, January 2, 2006"))

	middleSection := lipgloss.JoinVertical(
		lipgloss.Center,
		timeDisplay,
		dateDisplay,
	)

	// Right section - Status information with fixed height and consistent alignment
	// Always use two lines to ensure consistent height regardless of status
	var statusLine1 string
	var statusLine2 string

	if m.isLoading {
		// Loading indicator
		loadingStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#90cdf4"))

		statusLine1 = loadingStyle.Render("Loading...")
		statusLine2 = m.statusMessage
	} else if m.statusMessage != "" {
		// Status message with appropriate styling
		var statusStyle lipgloss.Style
		var statusIcon string

		switch m.statusType {
		case "success":
			statusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#48bb78")).
				Bold(true)
			statusIcon = "✓"
		case "error":
			statusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#f56565")).
				Bold(true)
			statusIcon = "✗"
		case "info":
			statusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#4299e1")).
				Bold(true)
			statusIcon = "ℹ"
		default:
			statusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#a0aec0"))
			statusIcon = "→"
		}

		statusLine1 = statusStyle.Render(statusIcon + " " + m.statusMessage)
		statusLine2 = "" // Empty second line
	} else {
		// Empty status (maintain height with empty lines)
		statusLine1 = ""
		statusLine2 = ""
	}

	// Ensure status always takes exactly two lines of consistent height
	rightSection := lipgloss.JoinVertical(
		lipgloss.Right,
		statusLine1,
		statusLine2,
	)

	// Calculate widths for each section
	sectionWidth := m.width / 3

	// Style each section with appropriate width and alignment
	leftSectionStyled := lipgloss.NewStyle().
		Width(sectionWidth).
		Align(lipgloss.Left).
		Render(leftSection)

	middleSectionStyled := lipgloss.NewStyle().
		Width(sectionWidth).
		Align(lipgloss.Center).
		Render(middleSection)

	rightSectionStyled := lipgloss.NewStyle().
		Width(sectionWidth).
		Align(lipgloss.Right).
		Render(rightSection)

	// Join the sections horizontally
	headerContent := lipgloss.JoinHorizontal(
		lipgloss.Center,
		leftSectionStyled,
		middleSectionStyled,
		rightSectionStyled,
	)

	// Apply the header style to the content
	return headerStyle.Render(headerContent)
}

// View renders the current state of the model as a string.
func (m *Model) View() string {
	// Render the header first
	header := m.renderHeader()

	// For the create form view, just show header and form
	if m.viewMode == "create" {
		createForm := m.renderCreateForm()
		return lipgloss.JoinVertical(lipgloss.Left, header, "\n", createForm)
	}

	// Three-column layout for main content
	var columns []string
	var visiblePanelCount int

	// Count visible panels
	if m.showTaskList {
		visiblePanelCount++
	}
	if m.showTaskDetails {
		visiblePanelCount++
	}
	if m.showTimeline {
		visiblePanelCount++
	}

	// Calculate column width - account for visible panels
	availableWidth := m.width
	columnWidth := availableWidth / max(1, visiblePanelCount)

	// Constants for border and padding
	const borderWidth = 1     // Width of border on each side
	const paddingWidth = 1    // Width of padding on each side
	const totalFrameWidth = 4 // Total extra width: (borderWidth + paddingWidth) * 2

	// Content width is column width minus frame elements for consistency
	contentWidth := columnWidth - totalFrameWidth

	// Prepare content for each panel
	var taskListContent, taskDetailsContent, timelineContent string

	// Get content for each panel
	if m.showTaskList {
		taskListContent = m.renderTaskList()
	}

	if m.showTaskDetails {
		taskDetailsContent = m.renderTaskDetails()
	}

	if m.showTimeline {
		timelineContent = m.renderTimelineView()
	}

	// Always create panels with consistent dimensions, with or without borders
	if m.showTaskList {
		// Create base style for content
		contentStyle := lipgloss.NewStyle().
			Width(contentWidth).
			MaxWidth(contentWidth)

		// Apply style to content
		taskListContent = contentStyle.Render(taskListContent)

		// Create frame style - either with visible border or with spacing
		var frameStyle lipgloss.Style
		if m.activePanel == 0 {
			// Active panel - visible borders
			frameStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(colorBorder)).
				BorderLeft(true).
				BorderRight(true).
				BorderTop(true).
				BorderBottom(true).
				Padding(paddingWidth).
				Width(columnWidth - 2) // Account for left and right border
		} else {
			// Inactive panel - invisible placeholder borders
			frameStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.HiddenBorder()).
				BorderLeft(true).
				BorderRight(true).
				BorderTop(true).
				BorderBottom(true).
				Padding(paddingWidth).
				Width(columnWidth - 2)
		}

		// Apply frame and add to columns
		columns = append(columns, frameStyle.Render(taskListContent))
	}

	if m.showTaskDetails {
		// Create base style for content
		contentStyle := lipgloss.NewStyle().
			Width(contentWidth).
			MaxWidth(contentWidth)

		// Apply style to content
		taskDetailsContent = contentStyle.Render(taskDetailsContent)

		// Create frame style - either with visible border or with spacing
		var frameStyle lipgloss.Style
		if m.activePanel == 1 {
			// Active panel - visible borders
			frameStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(colorBorder)).
				BorderLeft(true).
				BorderRight(true).
				BorderTop(true).
				BorderBottom(true).
				Padding(paddingWidth).
				Width(columnWidth - 2) // Account for left and right border
		} else {
			// Inactive panel - invisible placeholder borders
			frameStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.HiddenBorder()).
				BorderLeft(true).
				BorderRight(true).
				BorderTop(true).
				BorderBottom(true).
				Padding(paddingWidth).
				Width(columnWidth - 2)
		}

		// Apply frame and add to columns
		columns = append(columns, frameStyle.Render(taskDetailsContent))
	}

	if m.showTimeline {
		// Create base style for content
		contentStyle := lipgloss.NewStyle().
			Width(contentWidth).
			MaxWidth(contentWidth)

		// Apply style to content
		timelineContent = contentStyle.Render(timelineContent)

		// Create frame style - either with visible border or with spacing
		var frameStyle lipgloss.Style
		if m.activePanel == 2 {
			// Active panel - visible borders
			frameStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(colorBorder)).
				BorderLeft(true).
				BorderRight(true).
				BorderTop(true).
				BorderBottom(true).
				Padding(paddingWidth).
				Width(columnWidth - 2) // Account for left and right border
		} else {
			// Inactive panel - invisible placeholder borders
			frameStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.HiddenBorder()).
				BorderLeft(true).
				BorderRight(true).
				BorderTop(true).
				BorderBottom(true).
				Padding(paddingWidth).
				Width(columnWidth - 2)
		}

		// Apply frame and add to columns
		columns = append(columns, frameStyle.Render(timelineContent))
	}

	// Join columns horizontally with consistent spacing
	content := lipgloss.JoinHorizontal(lipgloss.Top, columns...)

	// Calculate available height for content, leaving space for help text
	// Adjust to account for header and a gap
	contentHeight := m.height - 2 - 4 // Reserve 2 rows for help text with padding and 4 for header with gap

	// Ensure content fits available height
	contentStyle := lipgloss.NewStyle().
		MaxHeight(contentHeight).
		Height(contentHeight)

	content = contentStyle.Render(content)

	// Add help footer fixed at the bottom with proper styling
	helpText := m.renderHelpText()

	// Create a prominent footer style using the existing style variables
	footerStyle := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Bold(true).
		Background(lipgloss.Color("#333333")). // Darker background for visibility
		Foreground(lipgloss.Color("#FFFFFF")). // White text for contrast
		Padding(0, 1).
		MarginTop(1)

	styledHelp := footerStyle.Render(helpText)

	// Position header, content and help text
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"", // Gap after header
		content,
		styledHelp,
	)
}

// renderTaskList renders the first column with the list of tasks
func (m *Model) renderTaskList() string {
	// Build the header content
	var headerContent strings.Builder
	headerContent.WriteString(m.styles.Title.Render("Tasks") + "\n\n")

	// Display error message if exists
	if m.err != nil {
		headerContent.WriteString(fmt.Sprintf("Error: %v\n\n", m.err))
	}

	// Display success message if exists
	if m.successMsg != "" {
		headerContent.WriteString(m.styles.Done.Render(fmt.Sprintf("✓ %s\n\n", m.successMsg)))
		// Clear the success message after it's been displayed once
		defer func() { m.successMsg = "" }()
	}

	if len(m.tasks) == 0 {
		headerContent.WriteString("No tasks found.\n\n")
		headerContent.WriteString("Press 'n' to create a new task.\n")
		return headerContent.String()
	}

	// Estimate the height of the viewport for tasks (after headers)
	// This determines how many tasks we can show at once
	viewableHeight := m.height - 14         // Adjust for header, borders, scroll indicators
	viewableHeight = max(5, viewableHeight) // Ensure minimum reasonable height

	// Make sure the cursor is visible
	if m.cursor < m.taskListOffset {
		// Cursor has moved above visible area
		m.taskListOffset = m.cursor
	} else if m.cursor >= m.taskListOffset+viewableHeight {
		// Cursor has moved below visible area
		m.taskListOffset = m.cursor - viewableHeight + 1
	}

	// Clamp taskListOffset to ensure it's never negative or too far down
	maxOffset := max(0, len(m.tasks)-viewableHeight)
	m.taskListOffset = min(m.taskListOffset, maxOffset)
	m.taskListOffset = max(0, m.taskListOffset)

	// Build list content with tasks
	var tasksContent strings.Builder

	// Add up-scroll indicator if needed
	if m.taskListOffset > 0 {
		tasksContent.WriteString(m.styles.Help.Render("↑ More tasks above ↑") + "\n")
	}

	// Build the visible task items
	visibleStartIdx := m.taskListOffset
	visibleEndIdx := min(m.taskListOffset+viewableHeight, len(m.tasks))

	// Render the currently visible tasks
	for i := visibleStartIdx; i < visibleEndIdx; i++ {
		t := m.tasks[i]
		statusSymbol := "[ ]"
		var statusStyle = m.styles.Todo

		switch t.Status {
		case task.StatusDone:
			statusSymbol = "[✓]"
			statusStyle = m.styles.Done
		case task.StatusInProgress:
			statusSymbol = "[⟳]"
			statusStyle = m.styles.InProgress
		}

		var priorityStyle = m.styles.LowPriority
		switch t.Priority {
		case task.PriorityHigh:
			priorityStyle = m.styles.HighPriority
		case task.PriorityMedium:
			priorityStyle = m.styles.MediumPriority
		}

		priority := string(t.Priority)
		taskLine := fmt.Sprintf("%s %s (%s)",
			statusStyle.Render(statusSymbol),
			t.Title,
			priorityStyle.Render(priority))

		if i == m.cursor {
			// Add cursor indicator and highlight
			tasksContent.WriteString("→ " + m.styles.SelectedItem.Render(taskLine) + "\n")
		} else {
			tasksContent.WriteString("  " + taskLine + "\n")
		}
	}

	// Add down-scroll indicator if needed
	if visibleEndIdx < len(m.tasks) {
		tasksContent.WriteString(m.styles.Help.Render("↓ More tasks below ↓") + "\n")
	}

	// Position indicator
	if len(m.tasks) > viewableHeight {
		position := fmt.Sprintf("[%d/%d]", m.cursor+1, len(m.tasks))
		tasksContent.WriteString(m.styles.Help.Render(position))
	}

	// Combine header and task list
	return headerContent.String() + tasksContent.String()
}

// renderTaskDetails renders the second column with details of the currently selected task
func (m *Model) renderTaskDetails() string {
	// Build full content first
	var fullContent strings.Builder

	if len(m.tasks) == 0 {
		fullContent.WriteString(m.styles.Title.Render("Task Details") + "\n\n")
		fullContent.WriteString("No tasks yet. Press 'n' to create your first task.\n\n")
		fullContent.WriteString(m.styles.Help.Render("Tip: You can organize tasks with priorities and due dates!"))
		return fullContent.String()
	}

	if len(m.tasks) == 0 || m.cursor >= len(m.tasks) {
		fullContent.WriteString(m.styles.Title.Render("Task Details") + "\n\nNo task selected")
		return fullContent.String()
	}

	t := m.tasks[m.cursor]
	fullContent.WriteString(m.styles.Title.Render("Task Details") + "\n\n")

	// Task title
	fullContent.WriteString(m.styles.Title.Render("Title: ") + t.Title + "\n\n")

	// Task description
	fullContent.WriteString(m.styles.Title.Render("Description: ") + "\n")
	if t.Description != nil && *t.Description != "" {
		fullContent.WriteString(*t.Description + "\n\n")
	} else {
		fullContent.WriteString("No description\n\n")
	}

	// Status
	fullContent.WriteString(m.styles.Title.Render("Status: "))
	switch t.Status {
	case task.StatusDone:
		fullContent.WriteString(m.styles.Done.Render(string(t.Status)))
	case task.StatusInProgress:
		fullContent.WriteString(m.styles.InProgress.Render(string(t.Status)))
	default:
		fullContent.WriteString(m.styles.Todo.Render(string(t.Status)))
	}
	fullContent.WriteString("\n\n")

	// Priority
	fullContent.WriteString(m.styles.Title.Render("Priority: "))
	switch t.Priority {
	case task.PriorityHigh:
		fullContent.WriteString(m.styles.HighPriority.Render(string(t.Priority)))
	case task.PriorityMedium:
		fullContent.WriteString(m.styles.MediumPriority.Render(string(t.Priority)))
	default:
		fullContent.WriteString(m.styles.LowPriority.Render(string(t.Priority)))
	}
	fullContent.WriteString("\n\n")

	// Due date
	fullContent.WriteString(m.styles.Title.Render("Due date: "))
	if t.DueDate != nil {
		fullContent.WriteString(t.DueDate.Format("2006-01-02"))
	} else {
		fullContent.WriteString("No due date")
	}
	fullContent.WriteString("\n\n")

	// Subtasks section
	fullContent.WriteString(m.styles.Title.Render("Subtasks:") + "\n")
	if len(t.SubTasks) > 0 {
		for _, st := range t.SubTasks {
			statusSymbol := "[ ]"
			var statusStyle = m.styles.Todo

			switch st.Status {
			case task.StatusDone:
				statusSymbol = "[✓]"
				statusStyle = m.styles.Done
			case task.StatusInProgress:
				statusSymbol = "[⟳]"
				statusStyle = m.styles.InProgress
			}

			fullContent.WriteString(fmt.Sprintf("  %s %s\n", statusStyle.Render(statusSymbol), st.Title))
		}
	} else {
		fullContent.WriteString("  No subtasks\n")
	}

	// Progress
	if len(t.SubTasks) > 0 {
		fullContent.WriteString("\n" + m.styles.Title.Render(fmt.Sprintf("Progress: %d%% (%d/%d tasks completed)\n",
			int(t.Progress*100), t.CompletedCount, t.TotalCount)))
	}

	// Estimate viewable height - subtract header and some padding from total panel height
	viewableHeight := m.height - 10         // Adjust as needed based on header size and padding
	viewableHeight = max(5, viewableHeight) // Ensure minimum reasonable height

	// Apply scrolling logic to the content
	return m.createScrollableContent(fullContent.String(), m.taskDetailsOffset, viewableHeight)
}

// renderTimelineView renders the third column with time-based task categories
func (m *Model) renderTimelineView() string {
	// Build full content first
	var fullContent strings.Builder

	fullContent.WriteString(m.styles.Title.Render("Timeline") + "\n\n")

	if len(m.tasks) == 0 {
		fullContent.WriteString("No tasks to display in timeline.\n\n")
		fullContent.WriteString("Create tasks with due dates to see them organized here.")
		return fullContent.String() // No need for scrolling with minimal content
	}

	overdue, today, upcoming := m.getTasksByTimeCategory()

	// Overdue section
	fullContent.WriteString(m.styles.HighPriority.Bold(true).Render("Overdue:") + "\n")
	if len(overdue) > 0 {
		for _, t := range overdue {
			dueDate := ""
			if t.DueDate != nil {
				dueDate = t.DueDate.Format("2006-01-02")
			}
			fullContent.WriteString(fmt.Sprintf("  %s (%s)\n", t.Title, dueDate))
		}
	} else {
		fullContent.WriteString("  No overdue tasks\n")
	}
	fullContent.WriteString("\n")

	// Today section
	fullContent.WriteString(m.styles.MediumPriority.Bold(true).Render("Today:") + "\n")
	if len(today) > 0 {
		for _, t := range today {
			fullContent.WriteString(fmt.Sprintf("  %s\n", t.Title))
		}
	} else {
		fullContent.WriteString("  No tasks due today\n")
	}
	fullContent.WriteString("\n")

	// Upcoming section
	fullContent.WriteString(m.styles.LowPriority.Bold(true).Render("Upcoming:") + "\n")
	if len(upcoming) > 0 {
		for _, t := range upcoming {
			dueDate := ""
			if t.DueDate != nil {
				dueDate = t.DueDate.Format("2006-01-02")
			}
			fullContent.WriteString(fmt.Sprintf("  %s (%s)\n", t.Title, dueDate))
		}
	} else {
		fullContent.WriteString("  No upcoming tasks\n")
	}

	// Estimate viewable height - subtract header and some padding from total panel height
	viewableHeight := m.height - 10         // Adjust as needed based on header size and padding
	viewableHeight = max(5, viewableHeight) // Ensure minimum reasonable height

	// Apply scrolling logic to the content
	return m.createScrollableContent(fullContent.String(), m.timelineOffset, viewableHeight)
}

// renderCreateForm renders the task creation form
func (m *Model) renderCreateForm() string {
	s := m.styles.Title.Render("Create New Task") + "\n\n"

	if m.err != nil {
		s += m.styles.HighPriority.Render(fmt.Sprintf("Error: %v\n\n", m.err))
	}

	// Form fields
	formFields := []struct {
		label    string
		value    string
		active   bool
		required bool
	}{
		{"Title", m.formTitle, m.activeField == 0, true},
		{"Description", m.formDescription, m.activeField == 1, false},
		{"Priority", m.formPriority, m.activeField == 2, false},
		{"Due Date (YYYY-MM-DD)", m.formDueDate, m.activeField == 3, false},
	}

	// Render each field
	for i, field := range formFields {
		// Field label with required indicator
		fieldLabel := field.label
		if field.required {
			fieldLabel += " *"
		}

		if field.active {
			s += m.styles.SelectedItem.Render(fieldLabel) + ": "
		} else {
			s += m.styles.Title.Render(fieldLabel) + ": "
		}

		// Field value
		if field.active {
			s += m.styles.SelectedItem.Render(field.value + "█") // Add cursor
		} else {
			s += field.value
		}

		// Special handling for priority field
		if i == 2 {
			var priorityStyle lipgloss.Style
			switch m.formPriority {
			case string(task.PriorityHigh):
				priorityStyle = m.styles.HighPriority
			case string(task.PriorityMedium):
				priorityStyle = m.styles.MediumPriority
			default:
				priorityStyle = m.styles.LowPriority
			}

			s += " (" + priorityStyle.Render(m.formPriority) + ")" +
				" - Press Space to cycle"
		}

		s += "\n\n"
	}

	// Submit button
	if m.activeField == 4 {
		s += m.styles.SelectedItem.Render("[Save Task]")
	} else {
		s += "[Save Task]"
	}

	s += "\n\n" + m.styles.Help.Render("Tab: next field • Enter: submit • Esc: cancel")

	return s
}

// renderHelpText renders the help text footer for the current view mode
func (m *Model) renderHelpText() string {
	var help string

	switch m.viewMode {
	case "list":
		help = "j/k: navigate • h/l or ←/→: switch panels • enter: view details • c: toggle completion • n: new task • 1/2/3: toggle columns • q: quit"
	case "detail":
		help = "esc: back • h/l or ←/→: switch panels • e: edit • c: toggle completion • d: delete • n: new task"
	case "edit":
		help = "esc: cancel • enter: save changes"
	case "create":
		help = "tab: next field • enter: submit • esc: cancel • space: cycle priority"
	}

	return m.styles.Help.Render(help)
}

// renderEditView renders the edit view of the selected task
func (m *Model) renderEditView() string {
	// Basic placeholder for edit view - would require text input
	s := m.styles.Title.Render("Edit Task") + "\n\n"
	s += "Edit mode not fully implemented\n\n"
	s += m.styles.Help.Render("esc: cancel • enter: save changes")
	return s
}

// max returns the larger of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// createScrollableContent creates a scrollable view of given content
func (m *Model) createScrollableContent(content string, offset int, maxHeight int) string {
	lines := strings.Split(content, "\n")

	// Calculate actual content height
	contentHeight := len(lines)

	// Determine if scrolling is needed
	needsScrolling := contentHeight > maxHeight

	// Clamp offset within valid range
	maxOffset := max(0, contentHeight-maxHeight)
	offset = min(offset, maxOffset)
	offset = max(0, offset)

	// Apply offset and limit number of lines to maxHeight
	startLine := min(offset, len(lines))
	endLine := min(startLine+maxHeight, len(lines))
	visibleLines := lines[startLine:endLine]

	// Add scroll indicators if needed
	if needsScrolling {
		// Ensure we have room for indicators
		visibleContent := strings.Join(visibleLines, "\n")

		var scrollIndicator string
		if offset > 0 && offset < maxOffset {
			// Both up and down scroll are available
			scrollIndicator = "▲\n" + visibleContent + "\n▼"
		} else if offset > 0 {
			// Only up scroll available
			scrollIndicator = "▲\n" + visibleContent
		} else if offset < maxOffset {
			// Only down scroll available
			scrollIndicator = visibleContent + "\n▼"
		} else {
			// No scrolling needed or at exact bounds
			scrollIndicator = visibleContent
		}

		return scrollIndicator
	}

	// No scrolling needed
	return strings.Join(visibleLines, "\n")
}
