// Copyright (C) 2025 Juan Antonio Gomez Pena
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package panels

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/shared"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/hooks"
	"github.com/newbpydev/tusk/internal/core/task"
)

// TimelineProps contains all properties needed to render the timeline panel
type TimelineProps struct {
	// Task category slices (replaces the single Tasks field)
	OverdueTasks    []task.Task
	TodayTasks      []task.Task
	UpcomingTasks   []task.Task
	
	// For backwards compatibility, still accept a single Tasks list
	// which will be used if the categorized slices are not provided
	Tasks           []task.Task
	
	Offset          int
	Width           int
	Height          int
	Styles          *shared.Styles
	IsActive        bool
	CollapsibleMgr  *hooks.CollapsibleManager
	CursorPosition  int  // Position for scrolling and highlighting
	CursorOnHeader  bool // Whether the cursor is on a section header
}

// RenderTimeline renders the timeline panel with a fixed header and scrollable content
func RenderTimeline(props TimelineProps) string {
	// Get the current date for comparison is now done in each helper function
	var scrollableContent strings.Builder
	var overdue, today, upcoming []task.Task
	
	// Use the dedicated category slices if they are provided, otherwise use the legacy behavior
	if len(props.OverdueTasks) > 0 || len(props.TodayTasks) > 0 || len(props.UpcomingTasks) > 0 {
		// Use the pre-categorized tasks from the model
		overdue = props.OverdueTasks
		today = props.TodayTasks
		upcoming = props.UpcomingTasks
	} else {
		// Fall back to categorizing the tasks in the component
		overdue, today, upcoming = getTasksByTimeCategory(props.Tasks)
	}
	
	// Check if we have a valid collapsible manager
	if props.CollapsibleMgr == nil {
		// Fall back to the old non-collapsible rendering if manager isn't available
		return renderLegacyTimeline(props, overdue, today, upcoming)
	}
	
	// Calculate viewport constraints for scrolling
	const scrollPadding = 3            // Number of lines to keep visible above/below selection
	viewportHeight := props.Height - 4 // Account for borders and header

	// Render the collapsible sections
	// First, get all the sections from the manager
	totalVisibleItems := props.CollapsibleMgr.GetItemCount()
	
	// Check if the Overdue section is expanded
	overdueSection := props.CollapsibleMgr.GetSection(hooks.SectionTypeOverdue)
	if overdueSection != nil {
		// Render the section header with expansion indicator
		expansionIndicator := "▼"
		if !overdueSection.IsExpanded {
			expansionIndicator = "▶"
		}
		
		// Create the section header with count
		headerText := fmt.Sprintf("%s Overdue (%d)", expansionIndicator, len(overdue))
		
		// Apply styling based on whether it's selected
		if props.CursorOnHeader && props.CursorPosition == props.CollapsibleMgr.GetSectionHeaderIndex(hooks.SectionTypeOverdue) {
			// This section header is selected
			headerText = props.Styles.SelectedItem.Render(headerText)
		} else {
			// Normal styling for section header
			headerText = props.Styles.HighPriority.Bold(true).Render(headerText)
		}
		
		scrollableContent.WriteString(headerText + "\n")
		
		// If the section is expanded, render its tasks
		if overdueSection.IsExpanded {
			renderTasksWithHighlight(&scrollableContent, overdue, props, props.Styles.HighPriority, hooks.SectionTypeOverdue)
		}
		
		scrollableContent.WriteString("\n")
	}
	
	// Check if the Today section is expanded
	todaySection := props.CollapsibleMgr.GetSection(hooks.SectionTypeToday)
	if todaySection != nil {
		// Render the section header with expansion indicator
		expansionIndicator := "▼"
		if !todaySection.IsExpanded {
			expansionIndicator = "▶"
		}
		
		// Create the section header with count
		headerText := fmt.Sprintf("%s Today (%d)", expansionIndicator, len(today))
		
		// Apply styling based on whether it's selected
		if props.CursorOnHeader && props.CursorPosition == props.CollapsibleMgr.GetSectionHeaderIndex(hooks.SectionTypeToday) {
			// This section header is selected
			headerText = props.Styles.SelectedItem.Render(headerText)
		} else {
			// Normal styling for section header
			headerText = props.Styles.MediumPriority.Bold(true).Render(headerText)
		}
		
		scrollableContent.WriteString(headerText + "\n")
		
		// If the section is expanded, render its tasks
		if todaySection.IsExpanded {
			renderTasksWithHighlight(&scrollableContent, today, props, props.Styles.MediumPriority, hooks.SectionTypeToday)
		}
		
		scrollableContent.WriteString("\n")
	}
	
	// Check if the Upcoming section is expanded
	upcomingSection := props.CollapsibleMgr.GetSection(hooks.SectionTypeUpcoming)
	if upcomingSection != nil {
		// Render the section header with expansion indicator
		expansionIndicator := "▼"
		if !upcomingSection.IsExpanded {
			expansionIndicator = "▶"
		}
		
		// Create the section header with count
		headerText := fmt.Sprintf("%s Upcoming (%d)", expansionIndicator, len(upcoming))
		
		// Apply styling based on whether it's selected
		if props.CursorOnHeader && props.CursorPosition == props.CollapsibleMgr.GetSectionHeaderIndex(hooks.SectionTypeUpcoming) {
			// This section header is selected
			headerText = props.Styles.SelectedItem.Render(headerText)
		} else {
			// Normal styling for section header
			headerText = props.Styles.LowPriority.Bold(true).Render(headerText)
		}
		
		scrollableContent.WriteString(headerText + "\n")
		
		// If the section is expanded, render its tasks
		if upcomingSection.IsExpanded {
			renderTasksWithHighlight(&scrollableContent, upcoming, props, props.Styles.LowPriority, hooks.SectionTypeUpcoming)
		}
	}

	// Calculate the optimal offset to keep selection centered
	halfViewport := (viewportHeight - scrollPadding) / 2
	targetOffset := max(0, props.CursorPosition-halfViewport)
	
	// Don't scroll past the end
	maxOffset := max(0, totalVisibleItems-viewportHeight+scrollPadding)
	targetOffset = min(targetOffset, maxOffset)
	
	return shared.RenderScrollablePanel(shared.ScrollablePanelProps{
		Title:             "Timeline",
		HeaderContent:     "",
		ScrollableContent: scrollableContent.String(),
		EmptyMessage:      "No tasks with due dates",
		Width:             props.Width,
		Height:            props.Height,
		Offset:            props.Offset,
		CursorPosition:    -1, // No cursor in simple timeline
		Styles:            props.Styles,
		IsActive:          props.IsActive,
		BorderColor:       shared.ColorBorder,
	})
}

// renderTasksWithHighlight renders a list of tasks with potential cursor highlighting
func renderTasksWithHighlight(sb *strings.Builder, tasks []task.Task, props TimelineProps, sectionStyle lipgloss.Style, sectionType hooks.SectionType) {
	// If there are no tasks, show a message
	if len(tasks) == 0 {
		sb.WriteString("  " + props.Styles.Help.Render("No tasks in this section\n"))
		return
	}
	
	// Get the current date for date formatting
	now := time.Now()
	todayDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	
	for i, t := range tasks {
		// Calculate the visual index of this task
		visualTaskIndex := -1
		sectionIndex := props.CollapsibleMgr.GetSectionHeaderIndex(sectionType)
		if sectionIndex >= 0 {
			visualTaskIndex = sectionIndex + 1 + i // Header index + 1 (to skip header) + task offset
		}
		
		// Prepare the status symbol
		var statusSymbol string
		switch t.Status {
		case task.StatusDone:
			statusSymbol = "[✓]"
		case task.StatusInProgress:
			statusSymbol = "[⟳]"
		default:
			statusSymbol = "[ ]"
		}
		
		// Format due date with appropriate styling based on section
		dueDate := ""
		if t.DueDate != nil {
			dueDate = t.DueDate.Format("2006-01-02")
			
			switch sectionType {
			case hooks.SectionTypeOverdue:
				// Calculate days overdue
				taskDueDate := time.Date(t.DueDate.Year(), t.DueDate.Month(), t.DueDate.Day(), 0, 0, 0, 0, t.DueDate.Location())
				daysOverdue := int(todayDate.Sub(taskDueDate).Hours() / 24)
				dueDate = fmt.Sprintf("%s (%d days overdue)", dueDate, daysOverdue)
				
			case hooks.SectionTypeToday:
				// Show remaining time in day
				endOfDay := todayDate.Add(24 * time.Hour)
				remaining := endOfDay.Sub(now)
				if remaining < 0 {
					remaining = 0
				}
				hours := int(remaining.Hours())
				minutes := int(remaining.Minutes()) % 60
				dueDate = fmt.Sprintf("Today (%dh %dm left)", hours, minutes)
				
			case hooks.SectionTypeUpcoming:
				// Show days until due
				taskDueDate := time.Date(t.DueDate.Year(), t.DueDate.Month(), t.DueDate.Day(), 0, 0, 0, 0, t.DueDate.Location())
				daysUntil := int(taskDueDate.Sub(todayDate).Hours() / 24)
				if daysUntil == 1 {
					dueDate = fmt.Sprintf("%s (Tomorrow)", dueDate)
				} else {
					dueDate = fmt.Sprintf("%s (in %d days)", dueDate, daysUntil)
				}
			}
		}
		
		// Handle status symbol highlighting
		var renderedStatusSymbol string
		var titlePart string
		
		// Check if this item is selected
		isSelected := !props.CursorOnHeader && props.CursorPosition == visualTaskIndex
		
		if isSelected {
			// When selected, only highlight the status symbol
			renderedStatusSymbol = props.Styles.SelectedItem.Render(statusSymbol)
			// Add the cursor arrow (→)
			renderedStatusSymbol = "→ " + renderedStatusSymbol
		} else {
			// Regular styling when not selected
			renderedStatusSymbol = statusSymbol
			// Add normal spacing without cursor
			renderedStatusSymbol = "  " + renderedStatusSymbol
		}
		
		// Build title part and due date
		titlePart = " " + t.Title
		if dueDate != "" {
			titlePart += fmt.Sprintf(" (%s)", sectionStyle.Render(dueDate))
		}
		
		// Combine all parts with proper indentation
		taskLine := renderedStatusSymbol + titlePart
		sb.WriteString(taskLine + "\n")
		
		// Add a short description if available
		if t.Description != nil && *t.Description != "" {
			desc := *t.Description
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}
			sb.WriteString(fmt.Sprintf("     %s\n", props.Styles.Help.Render(desc)))
		}
		
		// Add a separator between tasks except for the last one
		if i < len(tasks)-1 {
			sb.WriteString("     ---\n")
		}
	}
}

// renderLegacyTimeline renders the timeline panel without using collapsible sections
// This is the original timeline implementation before the refactoring
func renderLegacyTimeline(props TimelineProps, overdue, today, upcoming []task.Task) string {
	// Get the current date for consistent comparison throughout the function
	now := time.Now()
	todayDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	
	var scrollableContent strings.Builder
	
	// Create a visually rich timeline that's worth scrolling through
	scrollableContent.WriteString(props.Styles.HighPriority.Bold(true).Render("Overdue:") + "\n")
	if len(overdue) > 0 {
		for i, t := range overdue {
			dueDate := ""
			if t.DueDate != nil {
				dueDate = t.DueDate.Format("2006-01-02")
				
				// For overdue tasks, calculate days since due
				// CRITICAL: Only tasks with dates STRICTLY BEFORE today should be here
				taskDueDate := time.Date(t.DueDate.Year(), t.DueDate.Month(), t.DueDate.Day(), 0, 0, 0, 0, t.DueDate.Location())
				
				// Double-check that this task is truly overdue
				if taskDueDate.Before(todayDate) {
					daysOverdue := int(todayDate.Sub(taskDueDate).Hours() / 24)
					dueDate = fmt.Sprintf("%s (%d days overdue)", dueDate, daysOverdue)
				} else {
					// This should never happen if categorization is working correctly
					// But as a safeguard, if a today task gets into overdue section, fix it
					dueDate = fmt.Sprintf("%s (Today)", dueDate)
				}
			}

			// Add status indicator
			var statusSymbol string
			switch t.Status {
			case task.StatusDone:
				statusSymbol = "[✓]"
			case task.StatusInProgress:
				statusSymbol = "[⟳]"
			default:
				statusSymbol = "[ ]"
			}

			line := fmt.Sprintf("  %s %s (%s)\n", statusSymbol, t.Title, props.Styles.HighPriority.Render(dueDate))
			scrollableContent.WriteString(line)

			// Add a short description if available
			if t.Description != nil && *t.Description != "" {
				desc := *t.Description
				if len(desc) > 50 {
					desc = desc[:47] + "..."
				}
				scrollableContent.WriteString(fmt.Sprintf("     %s\n", props.Styles.Help.Render(desc)))
			}

			// Add a separator between tasks except for the last one
			if i < len(overdue)-1 {
				scrollableContent.WriteString("     ---\n")
			}
		}
	} else {
		scrollableContent.WriteString("  " + props.Styles.Help.Render("No overdue tasks\n"))
	}
	scrollableContent.WriteString("\n")

	scrollableContent.WriteString(props.Styles.MediumPriority.Bold(true).Render("Today:") + "\n")
	if len(today) > 0 {
		for i, t := range today {
			// Add status indicator
			var statusSymbol string
			switch t.Status {
			case task.StatusDone:
				statusSymbol = "[✓]"
			case task.StatusInProgress:
				statusSymbol = "[⟳]"
			default:
				statusSymbol = "[ ]"
			}

			// For today's tasks, show the remaining time until the end of the day
			dueDateStr := ""
			if t.DueDate != nil {
				// Compute the end of day based on current time
				now := time.Now()
				todayDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
				endOfDay := todayDate.Add(24 * time.Hour)
				remaining := endOfDay.Sub(now)
				if remaining < 0 {
					remaining = 0
				}
				hours := int(remaining.Hours())
				minutes := int(remaining.Minutes()) % 60
				dueDateStr = fmt.Sprintf("Today (%dh %dm left)", hours, minutes)
			}

			line := fmt.Sprintf("  %s %s %s\n", statusSymbol, t.Title, props.Styles.MediumPriority.Render(dueDateStr))
			scrollableContent.WriteString(line)

			// Add a short description if available
			if t.Description != nil && *t.Description != "" {
				desc := *t.Description
				if len(desc) > 50 {
					desc = desc[:47] + "..."
				}
				scrollableContent.WriteString(fmt.Sprintf("     %s\n", props.Styles.Help.Render(desc)))
			}

			// Add a separator between tasks except for the last one
			if i < len(today)-1 {
				scrollableContent.WriteString("     ---\n")
			}
		}
	} else {
		scrollableContent.WriteString("  " + props.Styles.Help.Render("No tasks due today\n"))
	}
	scrollableContent.WriteString("\n")

	scrollableContent.WriteString(props.Styles.LowPriority.Bold(true).Render("Upcoming:") + "\n")
	if len(upcoming) > 0 {
		for i, t := range upcoming {
			dueDate := ""
			if t.DueDate != nil {
				dueDate = t.DueDate.Format("2006-01-02")
				// Get the date components only for reliable day comparison
				now := time.Now()
				todayDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
				tomorrowDate := todayDate.AddDate(0, 0, 1)
				taskDate := time.Date(t.DueDate.Year(), t.DueDate.Month(), t.DueDate.Day(), 0, 0, 0, 0, t.DueDate.Location())
				
				if taskDate.Equal(tomorrowDate) {
					dueDate = fmt.Sprintf("%s (Tomorrow)", dueDate)
				} else {
					// Calculate days difference using manual counting for accuracy
					
					// Add a failsafe: manually calculate day difference to ensure accuracy
					days := 0
					testDate := todayDate
					for !testDate.After(taskDate) {
						if testDate.Equal(taskDate) {
							break
						}
						testDate = testDate.AddDate(0, 0, 1)
						days++
					}
					
					dueDate = fmt.Sprintf("%s (in %d days)", dueDate, days)
				}
			}

			// Add status indicator
			var statusSymbol string
			switch t.Status {
			case task.StatusDone:
				statusSymbol = "[✓]"
			case task.StatusInProgress:
				statusSymbol = "[⟳]"
			default:
				statusSymbol = "[ ]"
			}

			line := fmt.Sprintf("  %s %s (%s)\n", statusSymbol, t.Title, props.Styles.LowPriority.Render(dueDate))
			scrollableContent.WriteString(line)

			// Add a short description if available
			if t.Description != nil && *t.Description != "" {
				desc := *t.Description
				if len(desc) > 50 {
					desc = desc[:47] + "..."
				}
				scrollableContent.WriteString(fmt.Sprintf("     %s\n", props.Styles.Help.Render(desc)))
			}

			// Add a separator between tasks except for the last one
			if i < len(upcoming)-1 {
				scrollableContent.WriteString("     ---\n")
			}
		}
	} else {
		scrollableContent.WriteString("  " + props.Styles.Help.Render("No upcoming tasks\n"))
	}

	// Add summary at the bottom
	completedCount := countCompletedTasks(props.Tasks)
	activeCount := len(props.Tasks) - completedCount

	scrollableContent.WriteString("\n" + props.Styles.Title.Render("Summary:") + "\n")
	scrollableContent.WriteString(fmt.Sprintf("  Overdue: %d tasks\n", len(overdue)))
	scrollableContent.WriteString(fmt.Sprintf("  Today: %d tasks\n", len(today)))
	scrollableContent.WriteString(fmt.Sprintf("  Upcoming: %d tasks\n", len(upcoming)))
	scrollableContent.WriteString(fmt.Sprintf("  Completed: %d tasks\n", completedCount))
	scrollableContent.WriteString(fmt.Sprintf("  Active: %d tasks\n", activeCount))
	scrollableContent.WriteString(fmt.Sprintf("  Total: %d tasks\n", len(props.Tasks)))

	// Calculate the current line based on offset
	currentLine := props.Offset

	return shared.RenderScrollablePanel(shared.ScrollablePanelProps{
		Title:             "Timeline",
		ScrollableContent: scrollableContent.String(),
		EmptyMessage:      "No tasks to display in timeline.",
		Width:             props.Width,
		Height:            props.Height,
		Offset:            props.Offset,
		Styles:            props.Styles,
		IsActive:          props.IsActive,
		BorderColor:       shared.ColorBorder,
		// Use current line as cursor position to maintain viewport
		CursorPosition: currentLine,
	})
}

// getTasksByTimeCategory organizes tasks into overdue, today, and upcoming categories
// Only returns incomplete tasks for display in timeline
func getTasksByTimeCategory(tasks []task.Task) ([]task.Task, []task.Task, []task.Task) {
	var overdue, todayTasks, upcoming []task.Task

	// Use local time for consistency
	today := time.Now().In(time.Local)
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.Local)
	
	for _, t := range tasks {
		// Skip tasks that don't have a due date
		if t.DueDate == nil {
			continue
		}

		// Skip completed tasks
		if t.Status == task.StatusDone || t.IsCompleted {
			continue
		}

		// Normalize the task's due date using time.Local
		taskDueDate := time.Date(t.DueDate.Year(), t.DueDate.Month(), t.DueDate.Day(), 0, 0, 0, 0, time.Local)

		// Compute the difference in days between today and the task due date
		diff := todayDate.Sub(taskDueDate).Hours() / 24
		if math.Abs(diff) < 0.01 {
			// Due today (allowing a small tolerance)
			todayTasks = append(todayTasks, t)
		} else if diff > 0 {
			// Overdue: taskDueDate is strictly before today
			overdue = append(overdue, t)
		} else {
			// Upcoming: taskDueDate is after today
			upcoming = append(upcoming, t)
		}
	}
	
	return overdue, todayTasks, upcoming
}

// countCompletedTasks returns the number of completed tasks from the given task list
func countCompletedTasks(tasks []task.Task) int {
	count := 0
	for _, t := range tasks {
		if t.Status == task.StatusDone || t.IsCompleted {
			count++
		}
	}
	return count
}
