// Package task_details provides a component for rendering detailed task information
// in the TUI application.
package task_details

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/shared"
	"github.com/newbpydev/tusk/internal/core/task"
)

// Define styles for different date types
var (
	overdueStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	todayStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#00E676"))
	upcomingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1E88E5"))
)



// TaskDetailsOptions configures how task details should be rendered
type TaskDetailsOptions struct {
	// Width is the available width for rendering
	Width int
	
	// Height is the available height for rendering
	Height int
	
	// Offset is the vertical scroll position
	Offset int
}

// DefaultTaskDetailsOptions returns the default options for rendering task details
func DefaultTaskDetailsOptions() TaskDetailsOptions {
	return TaskDetailsOptions{
		Width:  80,
		Height: 24,
		Offset: 0,
	}
}

// RenderTaskDetails creates a string representation of detailed task information
// This is a pure function with no side effects, making it easy to test and reuse
func RenderTaskDetails(t task.Task, opts TaskDetailsOptions) string {
	var result strings.Builder
	var lines []string
	
	// Define styles
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#1E88E5")).MarginBottom(1)
	subtitleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#9C27B0"))
	dimmedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#747474"))
	
	// Title with priority indicator
	priorityStr := formatPriority(t.Priority)
	title := fmt.Sprintf("%s %s", priorityStr, t.Title)
	lines = append(lines, titleStyle.Render(title))
	lines = append(lines, "")
	
	// Status
	status := fmt.Sprintf("Status: %s", formatStatus(t.Status))
	lines = append(lines, status)
	
	// Due date
	if t.DueDate != nil {
		// Use the shared FormatDueDate function for consistent styling
		formattedDate, dateType := shared.FormatDueDate(t.DueDate, string(t.Status))
		
		// Apply appropriate styling based on date type
		var styledDate string
		switch dateType {
		case "overdue":
			styledDate = overdueStyle.Render(formattedDate)
		case "today":
			styledDate = todayStyle.Render(formattedDate)
		case "upcoming":
			styledDate = upcomingStyle.Render(formattedDate)
		default:
			styledDate = formattedDate
		}
		
		dueDate := fmt.Sprintf("Due Date: %s", styledDate)
		lines = append(lines, dueDate)
	}
	
	// Created/Updated dates
	if !t.CreatedAt.IsZero() {
		created := fmt.Sprintf("Created: %s", t.CreatedAt.Format("Jan 2, 2006 15:04"))
		lines = append(lines, created)
	}
	
	if !t.UpdatedAt.IsZero() {
		updated := fmt.Sprintf("Updated: %s", t.UpdatedAt.Format("Jan 2, 2006 15:04"))
		lines = append(lines, updated)
	}
	
	// Parent project information
	if t.ParentID != nil {
		project := fmt.Sprintf("Project ID: %d", *t.ParentID)
		lines = append(lines, project)
	}
	
	lines = append(lines, "")
	
	// Description with header
	lines = append(lines, subtitleStyle.Render("Description:"))
	
	if t.Description == nil || *t.Description == "" {
		lines = append(lines, dimmedStyle.Render("No description provided."))
	} else {
		// Word wrap description to fit the width
		// In a real implementation, you'd use a proper word-wrapping algorithm
		maxWidth := opts.Width - 2 // Account for some padding
		descLines := strings.Split(*t.Description, "\n")
		
		for _, line := range descLines {
			// Simple wrapping approach
			for len(line) > maxWidth {
				lines = append(lines, line[:maxWidth])
				line = line[maxWidth:]
			}
			if len(line) > 0 {
				lines = append(lines, line)
			}
		}
	}
	
	// Only show the visible portion based on height and offset
	start := max(0, opts.Offset)
	end := min(len(lines), start+opts.Height)
	
	for i := start; i < end; i++ {
		result.WriteString(lines[i])
		result.WriteString("\n")
	}
	
	// Add scroll indicator if needed
	if len(lines) > opts.Height {
		scrollPercent := float64(start) / float64(max(1, len(lines)-opts.Height))
		scrollMsg := fmt.Sprintf("Scroll: %.0f%% (↑/↓ to scroll)", scrollPercent*100)
		
		// Position at the bottom of the view
		result.WriteString(dimmedStyle.Render(scrollMsg))
	}
	
	return result.String()
}

// formatPriority returns a formatted priority string
func formatPriority(p task.Priority) string {
	highStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true)
	mediumStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB300")).Bold(true)
	lowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#747474"))
	
	var result string
	switch p {
	case task.PriorityHigh:
		result = highStyle.Render("[!]")
	case task.PriorityMedium:
		result = mediumStyle.Render("[~]")
	default:
		result = lowStyle.Render("[ ]")
	}
	return result
}

// formatStatus returns a formatted status string
func formatStatus(s task.Status) string {
	completedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#4CAF50")).Bold(true)
	todoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#1E88E5")).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	
	var result string
	switch s {
	case task.StatusDone:
		result = completedStyle.Render(string(s))
	case task.StatusTodo:
		result = todoStyle.Render(string(s))
	default:
		result = normalStyle.Render(string(s))
	}
	return result
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// RenderPlaceholder creates a placeholder message when no task is selected
func RenderPlaceholder(message string, width, height int, styles *shared.Styles) string {
	// Style for the placeholder message
	placeholderStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(lipgloss.Color("#666666")).
		Margin(1, 0, 0, 0)

	return placeholderStyle.Render(message)
}
