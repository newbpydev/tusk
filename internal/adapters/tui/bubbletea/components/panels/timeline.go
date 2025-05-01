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

	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/shared"
	"github.com/newbpydev/tusk/internal/core/task"
)

// TimelineProps contains all properties needed to render the timeline panel
type TimelineProps struct {
	Tasks    []task.Task
	Offset   int
	Width    int
	Height   int
	Styles   *shared.Styles
	IsActive bool
}

// RenderTimeline renders the timeline panel with a fixed header and scrollable content
func RenderTimeline(props TimelineProps) string {
	// Get the current date for consistent comparison throughout the function
	now := time.Now()
	todayDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	
	var scrollableContent strings.Builder
	overdue, today, upcoming := getTasksByTimeCategory(props.Tasks)

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
