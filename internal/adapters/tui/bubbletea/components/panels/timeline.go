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

// RenderTimeline renders the third column with time-based task categories
func RenderTimeline(props TimelineProps) string {
	// Build full content first
	var fullContent strings.Builder

	fullContent.WriteString(props.Styles.Title.Render("Timeline") + "\n\n")

	if len(props.Tasks) == 0 {
		fullContent.WriteString("No tasks to display in timeline.\n\n")
		fullContent.WriteString("Create tasks with due dates to see them organized here.")
		return fullContent.String() // No need for scrolling with minimal content
	}

	overdue, today, upcoming := getTasksByTimeCategory(props.Tasks)

	// Overdue section
	fullContent.WriteString(props.Styles.HighPriority.Bold(true).Render("Overdue:") + "\n")
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
	fullContent.WriteString(props.Styles.MediumPriority.Bold(true).Render("Today:") + "\n")
	if len(today) > 0 {
		for _, t := range today {
			fullContent.WriteString(fmt.Sprintf("  %s\n", t.Title))
		}
	} else {
		fullContent.WriteString("  No tasks due today\n")
	}
	fullContent.WriteString("\n")

	// Upcoming section
	fullContent.WriteString(props.Styles.LowPriority.Bold(true).Render("Upcoming:") + "\n")
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

	// Use less aggressive height reduction to show more content
	const headerOffset = 3 // Reduced from previous value
	viewableHeight := props.Height - headerOffset
	viewableHeight = max(5, viewableHeight) // Ensure minimum reasonable height

	// Check if the content fits without scrolling
	contentLines := strings.Split(fullContent.String(), "\n")
	if len(contentLines) <= viewableHeight {
		// Content fits without scrolling, return as is
		return fullContent.String()
	}

	// Apply scrolling logic to the content
	return createScrollableContent(fullContent.String(), props.Offset, viewableHeight, props.Styles)
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
