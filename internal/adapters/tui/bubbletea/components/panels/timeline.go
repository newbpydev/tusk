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
	var scrollableContent strings.Builder
	overdue, today, upcoming := getTasksByTimeCategory(props.Tasks)

	// Create a visually rich timeline that's worth scrolling through
	scrollableContent.WriteString(props.Styles.HighPriority.Bold(true).Render("Overdue:") + "\n")
	if len(overdue) > 0 {
		for i, t := range overdue {
			dueDate := ""
			if t.DueDate != nil {
				dueDate = t.DueDate.Format("2006-01-02")
				daysSince := int(time.Since(*t.DueDate).Hours() / 24)
				dueDate = fmt.Sprintf("%s (%d days overdue)", dueDate, daysSince)
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

			line := fmt.Sprintf("  %s %s\n", statusSymbol, t.Title)
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
				daysUntil := int(t.DueDate.Sub(time.Now()).Hours() / 24)
				if daysUntil == 1 {
					dueDate = fmt.Sprintf("%s (Tomorrow)", dueDate)
				} else {
					dueDate = fmt.Sprintf("%s (in %d days)", dueDate, daysUntil)
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
	scrollableContent.WriteString("\n" + props.Styles.Title.Render("Summary:") + "\n")
	scrollableContent.WriteString(fmt.Sprintf("  Overdue: %d tasks\n", len(overdue)))
	scrollableContent.WriteString(fmt.Sprintf("  Today: %d tasks\n", len(today)))
	scrollableContent.WriteString(fmt.Sprintf("  Upcoming: %d tasks\n", len(upcoming)))
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
func getTasksByTimeCategory(tasks []task.Task) ([]task.Task, []task.Task, []task.Task) {
	var overdue, todayTasks, upcoming []task.Task

	now := time.Now()
	todayDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := todayDate.AddDate(0, 0, 1)

	for _, t := range tasks {
		if t.DueDate == nil {
			continue
		}

		dueDate := *t.DueDate
		if dueDate.Before(todayDate) {
			overdue = append(overdue, t)
		} else if dueDate.Before(tomorrow) {
			todayTasks = append(todayTasks, t)
		} else {
			upcoming = append(upcoming, t)
		}
	}

	return overdue, todayTasks, upcoming
}
