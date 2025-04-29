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
	// Always keep the title fixed at top with consistent positioning
	headerContent := props.Styles.Title.Render("Timeline") + "\n\n"

	// Total height available for the scrollable content area
	headerLines := 2 // "Timeline" + blank line
	scrollableHeight := max(1, props.Height-headerLines)

	// The message to show when no content is available
	emptyContent := func(message string) string {
		// Pad the message to fill the scrollable height to maintain panel dimensions
		padding := ""
		msgLines := strings.Count(message, "\n") + 1
		if msgLines < scrollableHeight {
			padding = strings.Repeat("\n", scrollableHeight-msgLines)
		}
		return message + padding
	}

	// Guard against negative height which can cause array bounds errors
	if props.Height < 5 {
		return headerContent + emptyContent("Window too small")
	}

	if len(props.Tasks) == 0 {
		return headerContent + emptyContent("No tasks to display in timeline.\n\n"+
			"Create tasks with due dates to see them organized here.")
	}

	// Build the scrollable content (everything except the header)
	var scrollableContent strings.Builder

	overdue, today, upcoming := getTasksByTimeCategory(props.Tasks)

	// Overdue section
	scrollableContent.WriteString(props.Styles.HighPriority.Bold(true).Render("Overdue:") + "\n")
	if len(overdue) > 0 {
		for _, t := range overdue {
			dueDate := ""
			if t.DueDate != nil {
				dueDate = t.DueDate.Format("2006-01-02")
			}
			scrollableContent.WriteString(fmt.Sprintf("  %s (%s)\n", t.Title, dueDate))
		}
	} else {
		scrollableContent.WriteString("  No overdue tasks\n")
	}
	scrollableContent.WriteString("\n")

	// Today section
	scrollableContent.WriteString(props.Styles.MediumPriority.Bold(true).Render("Today:") + "\n")
	if len(today) > 0 {
		for _, t := range today {
			scrollableContent.WriteString(fmt.Sprintf("  %s\n", t.Title))
		}
	} else {
		scrollableContent.WriteString("  No tasks due today\n")
	}
	scrollableContent.WriteString("\n")

	// Upcoming section
	scrollableContent.WriteString(props.Styles.LowPriority.Bold(true).Render("Upcoming:") + "\n")
	if len(upcoming) > 0 {
		for _, t := range upcoming {
			dueDate := ""
			if t.DueDate != nil {
				dueDate = t.DueDate.Format("2006-01-02")
			}
			scrollableContent.WriteString(fmt.Sprintf("  %s (%s)\n", t.Title, dueDate))
		}
	} else {
		scrollableContent.WriteString("  No upcoming tasks\n")
	}

	// Get the scrollable content as text
	scrollableText := scrollableContent.String()

	// Use the shared implementation to create consistently sized scrollable content
	return headerContent + shared.CreateScrollableContent(scrollableText, props.Offset, scrollableHeight, props.Styles)
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
