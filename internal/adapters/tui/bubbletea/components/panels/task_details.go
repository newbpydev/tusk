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

	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/shared"
	"github.com/newbpydev/tusk/internal/core/task"
)

// TaskDetailsProps contains all properties needed to render the task details panel
type TaskDetailsProps struct {
	Tasks    []task.Task
	Cursor   int
	Offset   int
	Width    int
	Height   int
	Styles   *shared.Styles
	IsActive bool
}

// RenderTaskDetails renders the second column with details of the currently selected task
func RenderTaskDetails(props TaskDetailsProps) string {
	// Always keep the title fixed at top with consistent positioning
	headerContent := props.Styles.Title.Render("Task Details") + "\n\n"

	// Total height available for the scrollable content area
	headerLines := 2 // "Task Details" + blank line
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

	// Handle empty task list
	if len(props.Tasks) == 0 {
		return headerContent + emptyContent("No tasks yet. Press 'n' to create your first task.\n\n"+
			props.Styles.Help.Render("Tip: You can organize tasks with priorities and due dates!"))
	}

	// Handle invalid cursor
	if props.Cursor >= len(props.Tasks) {
		return headerContent + emptyContent("No task selected")
	}

	// Build the scrollable content (everything except the header)
	var scrollableContent strings.Builder
	t := props.Tasks[props.Cursor]

	// Task title
	scrollableContent.WriteString(props.Styles.Title.Render("Title: ") + t.Title + "\n\n")

	// Task description
	scrollableContent.WriteString(props.Styles.Title.Render("Description: ") + "\n")
	if t.Description != nil && *t.Description != "" {
		scrollableContent.WriteString(*t.Description + "\n\n")
	} else {
		scrollableContent.WriteString("No description\n\n")
	}

	// Status
	scrollableContent.WriteString(props.Styles.Title.Render("Status: "))
	switch t.Status {
	case task.StatusDone:
		scrollableContent.WriteString(props.Styles.Done.Render(string(t.Status)))
	case task.StatusInProgress:
		scrollableContent.WriteString(props.Styles.InProgress.Render(string(t.Status)))
	default:
		scrollableContent.WriteString(props.Styles.Todo.Render(string(t.Status)))
	}
	scrollableContent.WriteString("\n\n")

	// Priority
	scrollableContent.WriteString(props.Styles.Title.Render("Priority: "))
	switch t.Priority {
	case task.PriorityHigh:
		scrollableContent.WriteString(props.Styles.HighPriority.Render(string(t.Priority)))
	case task.PriorityMedium:
		scrollableContent.WriteString(props.Styles.MediumPriority.Render(string(t.Priority)))
	default:
		scrollableContent.WriteString(props.Styles.LowPriority.Render(string(t.Priority)))
	}
	scrollableContent.WriteString("\n\n")

	// Due date
	scrollableContent.WriteString(props.Styles.Title.Render("Due date: "))
	if t.DueDate != nil {
		scrollableContent.WriteString(t.DueDate.Format("2006-01-02"))
	} else {
		scrollableContent.WriteString("No due date")
	}
	scrollableContent.WriteString("\n\n")

	// Subtasks section
	scrollableContent.WriteString(props.Styles.Title.Render("Subtasks:") + "\n")
	if len(t.SubTasks) > 0 {
		for _, st := range t.SubTasks {
			statusSymbol := "[ ]"
			var statusStyle = props.Styles.Todo

			switch st.Status {
			case task.StatusDone:
				statusSymbol = "[✓]"
				statusStyle = props.Styles.Done
			case task.StatusInProgress:
				statusSymbol = "[⟳]"
				statusStyle = props.Styles.InProgress
			}

			scrollableContent.WriteString(fmt.Sprintf("  %s %s\n", statusStyle.Render(statusSymbol), st.Title))
		}
	} else {
		scrollableContent.WriteString("  No subtasks\n")
	}

	// Progress
	if len(t.SubTasks) > 0 {
		scrollableContent.WriteString("\n" + props.Styles.Title.Render(fmt.Sprintf("Progress: %d%% (%d/%d tasks completed)\n",
			int(t.Progress*100), t.CompletedCount, t.TotalCount)))
	}

	// Get the scrollable content with proper height management
	scrollableText := scrollableContent.String()

	// Use the shared implementation to create consistently sized scrollable content
	return headerContent + shared.CreateScrollableContent(scrollableText, props.Offset, scrollableHeight, props.Styles)
}
