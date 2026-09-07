// Package task_item provides a reusable component for rendering task items
// in the TUI application.
package task_item

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/tusk/internal/core/task"
)

// TaskItemOptions configures how a task item should be rendered
type TaskItemOptions struct {
	// IsSelected indicates if this task item is currently selected
	IsSelected bool
	
	// ShowDetails determines if additional task details should be shown
	ShowDetails bool
	
	// Width is the available width for rendering
	Width int
	
	// IncludeCheckbox determines if the completion checkbox should be shown
	IncludeCheckbox bool
}

// DefaultTaskItemOptions returns the default options for rendering a task item
func DefaultTaskItemOptions() TaskItemOptions {
	return TaskItemOptions{
		IsSelected:     false,
		ShowDetails:    false,
		Width:          80,
		IncludeCheckbox: true,
	}
}

// RenderTaskItem creates a string representation of a task item for display
// This is a pure function with no side effects, making it easy to test and reuse
func RenderTaskItem(t task.Task, _ interface{}, opts TaskItemOptions) string {
	// Define styles using lipgloss directly
	selectedStyle := lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("#1E88E5")).Foreground(lipgloss.Color("#FFFFFF"))
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	checkboxStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#747474"))
	
	// Base style determined by selection state
	var baseStyle lipgloss.Style
	if opts.IsSelected {
		baseStyle = selectedStyle
	} else {
		baseStyle = normalStyle
	}
	
	// Determine checkbox display
	var checkbox string
	if opts.IncludeCheckbox {
		if t.Status == task.StatusDone {
			checkbox = checkboxStyle.Render("[x] ")
		} else {
			checkbox = checkboxStyle.Render("[ ] ")
		}
	}
	
	// Format the title with priority indicator
	priorityIndicator := getPriorityIndicator(t.Priority)
	title := fmt.Sprintf("%s%s", priorityIndicator, t.Title)
	
	// Format date if exists
	dateInfo := ""
	if t.DueDate != nil {
		dateStyle := getDateStyle(*t.DueDate)
		dateInfo = dateStyle(fmt.Sprintf(" (%s)", formatDate(*t.DueDate)))
	}
	
	// Basic one-line representation
	basicView := baseStyle.Render(fmt.Sprintf("%s%s%s", checkbox, title, dateInfo))
	
	// If no details requested, return just the basic view
	if !opts.ShowDetails {
		return basicView
	}
	
	// Add description for detailed view if available
	dimmedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#747474"))
	detailedView := basicView
	
	if t.Description != nil && *t.Description != "" {
		// Indent and wrap description to fit the available width
		descLines := strings.Split(*t.Description, "\n")
		wrappedDesc := ""
		
		for _, line := range descLines {
			// Simple wrapping - in a real implementation, you'd use a proper word-wrapping algorithm
			if len(line) > opts.Width-8 { // 8 for indentation and some buffer
				wrappedDesc += fmt.Sprintf("\n    %s", dimmedStyle.Render(line[:opts.Width-8]+"..."))
			} else {
				wrappedDesc += fmt.Sprintf("\n    %s", dimmedStyle.Render(line))
			}
		}
		
		detailedView += wrappedDesc
		// Add project info if this task belongs to a project
		if t.ParentID != nil {
			projectInfo := fmt.Sprintf("\n    %s", dimmedStyle.Render("Project: "+fmt.Sprintf("%d", *t.ParentID)))
			detailedView += projectInfo
		}
		return detailedView
	}
	
	// Add project info if this task belongs to a project
	if t.ParentID != nil {
		projectInfo := fmt.Sprintf("\n    %s", dimmedStyle.Render("Project: "+fmt.Sprintf("%d", *t.ParentID)))
		detailedView += projectInfo
	}
	
	return baseStyle.Render(detailedView)
}

// getPriorityIndicator returns a styled indicator for the task priority
func getPriorityIndicator(p task.Priority) string {
	highStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true)
	mediumStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB300")).Bold(true)
	lowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#747474"))
	
	switch p {
	case task.PriorityHigh:
		return highStyle.Render("! ")
	case task.PriorityMedium:
		return mediumStyle.Render("~ ")
	default:
		return lowStyle.Render("  ")
	}
}

// getDateStyle returns the appropriate style function for a date
func getDateStyle(date time.Time) func(string) string {
	now := time.Now()
	
	// Define styles
	overdueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	todayStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00E676"))
	upcomingStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#1E88E5"))
	
	// Convert to start of day for comparison
	dateStartOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	nowStartOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	
	// Create wrapper functions with the right signature
	overdueRender := func(s string) string { return overdueStyle.Render(s) }
	todayRender := func(s string) string { return todayStyle.Render(s) }
	upcomingRender := func(s string) string { return upcomingStyle.Render(s) }
	
	// Past due dates
	if dateStartOfDay.Before(nowStartOfDay) {
		return overdueRender
	}
	
	// Today's dates
	if dateStartOfDay.Equal(nowStartOfDay) {
		return todayRender
	}
	
	// Future dates
	return upcomingRender
}

// formatDate formats a time.Time into a user-friendly string
func formatDate(t time.Time) string {
	now := time.Now()
	
	// Convert to start of day for comparison
	tStartOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	nowStartOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	
	// Today
	if tStartOfDay.Equal(nowStartOfDay) {
		return "Today " + t.Format("15:04")
	}
	
	// Tomorrow
	tomorrowStartOfDay := nowStartOfDay.AddDate(0, 0, 1)
	if tStartOfDay.Equal(tomorrowStartOfDay) {
		return "Tomorrow " + t.Format("15:04")
	}
	
	// Yesterday
	yesterdayStartOfDay := nowStartOfDay.AddDate(0, 0, -1)
	if tStartOfDay.Equal(yesterdayStartOfDay) {
		return "Yesterday " + t.Format("15:04")
	}
	
	// This week (within next 7 days)
	if tStartOfDay.After(nowStartOfDay) && tStartOfDay.Before(nowStartOfDay.AddDate(0, 0, 7)) {
		return t.Format("Mon Jan 2")
	}
	
	// Default format for other dates
	return t.Format("Jan 2, 2006")
}
