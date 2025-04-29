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
	// Build full content first
	var fullContent strings.Builder

	if len(props.Tasks) == 0 {
		fullContent.WriteString(props.Styles.Title.Render("Task Details") + "\n\n")
		fullContent.WriteString("No tasks yet. Press 'n' to create your first task.\n\n")
		fullContent.WriteString(props.Styles.Help.Render("Tip: You can organize tasks with priorities and due dates!"))
		return fullContent.String()
	}

	if len(props.Tasks) == 0 || props.Cursor >= len(props.Tasks) {
		fullContent.WriteString(props.Styles.Title.Render("Task Details") + "\n\nNo task selected")
		return fullContent.String()
	}

	t := props.Tasks[props.Cursor]
	fullContent.WriteString(props.Styles.Title.Render("Task Details") + "\n\n")

	// Task title
	fullContent.WriteString(props.Styles.Title.Render("Title: ") + t.Title + "\n\n")

	// Task description
	fullContent.WriteString(props.Styles.Title.Render("Description: ") + "\n")
	if t.Description != nil && *t.Description != "" {
		fullContent.WriteString(*t.Description + "\n\n")
	} else {
		fullContent.WriteString("No description\n\n")
	}

	// Status
	fullContent.WriteString(props.Styles.Title.Render("Status: "))
	switch t.Status {
	case task.StatusDone:
		fullContent.WriteString(props.Styles.Done.Render(string(t.Status)))
	case task.StatusInProgress:
		fullContent.WriteString(props.Styles.InProgress.Render(string(t.Status)))
	default:
		fullContent.WriteString(props.Styles.Todo.Render(string(t.Status)))
	}
	fullContent.WriteString("\n\n")

	// Priority
	fullContent.WriteString(props.Styles.Title.Render("Priority: "))
	switch t.Priority {
	case task.PriorityHigh:
		fullContent.WriteString(props.Styles.HighPriority.Render(string(t.Priority)))
	case task.PriorityMedium:
		fullContent.WriteString(props.Styles.MediumPriority.Render(string(t.Priority)))
	default:
		fullContent.WriteString(props.Styles.LowPriority.Render(string(t.Priority)))
	}
	fullContent.WriteString("\n\n")

	// Due date
	fullContent.WriteString(props.Styles.Title.Render("Due date: "))
	if t.DueDate != nil {
		fullContent.WriteString(t.DueDate.Format("2006-01-02"))
	} else {
		fullContent.WriteString("No due date")
	}
	fullContent.WriteString("\n\n")

	// Subtasks section
	fullContent.WriteString(props.Styles.Title.Render("Subtasks:") + "\n")
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

			fullContent.WriteString(fmt.Sprintf("  %s %s\n", statusStyle.Render(statusSymbol), st.Title))
		}
	} else {
		fullContent.WriteString("  No subtasks\n")
	}

	// Progress
	if len(t.SubTasks) > 0 {
		fullContent.WriteString("\n" + props.Styles.Title.Render(fmt.Sprintf("Progress: %d%% (%d/%d tasks completed)\n",
			int(t.Progress*100), t.CompletedCount, t.TotalCount)))
	}

	// Use minimal height reduction to maximize content display
	const headerOffset = 2 // Reduced from previous value of 3
	viewableHeight := props.Height - headerOffset
	viewableHeight = max(5, viewableHeight) // Ensure minimum reasonable height

	// Check if content fits without scrolling
	lines := strings.Split(fullContent.String(), "\n")
	if len(lines) <= viewableHeight {
		// If content fits without scrolling, display it all
		return fullContent.String()
	}

	// Apply scrolling logic to the content
	return createScrollableContent(fullContent.String(), props.Offset, viewableHeight, props.Styles)
}

// createScrollableContent creates a scrollable view of given content
func createScrollableContent(content string, offset int, maxHeight int, styles *shared.Styles) string {
	lines := strings.Split(content, "\n")

	// Calculate actual content height
	contentHeight := len(lines)

	// Determine if scrolling is needed
	needsScrolling := contentHeight > maxHeight

	// If scrolling not needed, just return the whole content
	if !needsScrolling {
		return content
	}

	// Clamp offset within valid range
	maxOffset := max(0, contentHeight-maxHeight)
	offset = min(offset, maxOffset)
	offset = max(0, offset)

	// Apply offset and limit number of lines to maxHeight
	startLine := min(offset, len(lines))
	endLine := min(startLine+maxHeight, len(lines))

	// Try to show more content if we have space
	if endLine-startLine < maxHeight && endLine < len(lines) {
		additionalLines := min(maxHeight-(endLine-startLine), len(lines)-endLine)
		endLine += additionalLines
	}

	visibleLines := lines[startLine:endLine]

	// Add scroll indicators if needed
	var scrollIndicator strings.Builder

	if offset > 0 && offset < maxOffset {
		// Both up and down scroll are available
		scrollIndicator.WriteString("▲\n")
		scrollIndicator.WriteString(strings.Join(visibleLines, "\n"))
		scrollIndicator.WriteString("\n▼")
	} else if offset > 0 {
		// Only up scroll available
		scrollIndicator.WriteString("▲\n")
		scrollIndicator.WriteString(strings.Join(visibleLines, "\n"))
	} else if offset < maxOffset {
		// Only down scroll available
		scrollIndicator.WriteString(strings.Join(visibleLines, "\n"))
		scrollIndicator.WriteString("\n▼")
	} else {
		// No scrolling needed or at exact bounds
		return strings.Join(visibleLines, "\n")
	}

	return scrollIndicator.String()
}
