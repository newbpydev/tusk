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

// TaskListProps contains all properties needed to render the task list panel
type TaskListProps struct {
	Tasks        []task.Task
	Cursor       int
	Offset       int
	Width        int
	Height       int
	Styles       *shared.Styles
	IsActive     bool
	Error        error
	SuccessMsg   string
	ClearSuccess func()
}

// RenderTaskList renders the first column with the list of tasks
func RenderTaskList(props TaskListProps) string {
	// Build the header content
	var headerContent strings.Builder
	headerContent.WriteString(props.Styles.Title.Render("Tasks") + "\n\n")

	// Display error message if exists
	if props.Error != nil {
		headerContent.WriteString(fmt.Sprintf("Error: %v\n\n", props.Error))
	}

	// Display success message if exists
	if props.SuccessMsg != "" {
		headerContent.WriteString(props.Styles.Done.Render(fmt.Sprintf("✓ %s\n\n", props.SuccessMsg)))
		// Clear the success message after it's been displayed once
		if props.ClearSuccess != nil {
			defer props.ClearSuccess()
		}
	}

	if len(props.Tasks) == 0 {
		headerContent.WriteString("No tasks found.\n\n")
		headerContent.WriteString("Press 'n' to create a new task.\n")
		// Add padding to maintain consistent layout
		headerContent.WriteString("\n\n") // Top arrow space
		headerContent.WriteString("\n")   // Bottom arrow space
		headerContent.WriteString("\n")   // Position indicator space
		return headerContent.String()
	}

	// Use as much space as possible for content
	// Only reserve minimal space for headers and indicators
	const headerFooterOffset = 2 // Reduced from previous values

	// frameHeight is total panel height
	frameHeight := props.Height - headerFooterOffset

	// Count header content lines
	headerStr := headerContent.String()
	headerLines := len(strings.Split(strings.TrimSuffix(headerStr, "\n"), "\n"))

	// Reserve space for navigation indicators - always maintain consistent layout
	const arrowLinesCount = 2   // One for top arrow, one for bottom
	const positionLineCount = 1 // One for position indicator

	// Compute available height for task list, accounting for fixed layout elements
	viewableHeight := frameHeight - headerLines - arrowLinesCount - positionLineCount
	viewableHeight = max(3, viewableHeight) // Ensure minimum reasonable height

	// Make sure cursor is in valid range
	cursor := min(max(0, props.Cursor), len(props.Tasks)-1)

	// Auto-adjust offset to ensure cursor is always visible
	offset := props.Offset

	// If cursor is below visible area, adjust offset to show cursor at bottom of view
	if cursor >= offset+viewableHeight {
		offset = cursor - viewableHeight + 1
	}

	// If cursor is above visible area, adjust offset to show cursor at top of view
	if cursor < offset {
		offset = cursor
	}

	// Clamp the offset to valid values
	maxOffset := max(0, len(props.Tasks)-viewableHeight)
	offset = min(offset, maxOffset)
	offset = max(0, offset)

	// Build list content with tasks
	var tasksContent strings.Builder

	// Always reserve space for up indicator - show it only if needed
	if offset > 0 {
		tasksContent.WriteString(props.Styles.Help.Render("↑ More tasks above ↑") + "\n")
	} else {
		tasksContent.WriteString("\n") // Empty line to maintain spacing
	}

	// Check if we can display all tasks
	if len(props.Tasks) <= viewableHeight {
		// If we have enough space, display all tasks
		for i, t := range props.Tasks {
			renderTaskLine(&tasksContent, t, i, cursor, props.Styles)
		}

		// Add padding if necessary to maintain consistent height
		for i := len(props.Tasks); i < viewableHeight; i++ {
			tasksContent.WriteString("\n")
		}
	} else {
		// Otherwise, display tasks with scrolling
		visibleStartIdx := offset
		visibleEndIdx := min(offset+viewableHeight, len(props.Tasks))

		// Render the currently visible tasks
		for i := visibleStartIdx; i < visibleEndIdx; i++ {
			renderTaskLine(&tasksContent, props.Tasks[i], i, cursor, props.Styles)
		}
	}

	// Always reserve space for down indicator - show it only if needed
	if offset+viewableHeight < len(props.Tasks) {
		tasksContent.WriteString(props.Styles.Help.Render("↓ More tasks below ↓") + "\n")
	} else {
		tasksContent.WriteString("\n") // Empty line to maintain spacing
	}

	// Always display position indicator but customize it based on list state
	position := fmt.Sprintf("[%d/%d]", cursor+1, len(props.Tasks))
	tasksContent.WriteString(props.Styles.Help.Render(position))

	// Combine header and task list
	return headerContent.String() + tasksContent.String()
}

// renderTaskLine renders a single task line
func renderTaskLine(builder *strings.Builder, t task.Task, index int, cursor int, styles *shared.Styles) {
	statusSymbol := "[ ]"
	var statusStyle = styles.Todo

	switch t.Status {
	case task.StatusDone:
		statusSymbol = "[✓]"
		statusStyle = styles.Done
	case task.StatusInProgress:
		statusSymbol = "[⟳]"
		statusStyle = styles.InProgress
	}

	var priorityStyle = styles.LowPriority
	switch t.Priority {
	case task.PriorityHigh:
		priorityStyle = styles.HighPriority
	case task.PriorityMedium:
		priorityStyle = styles.MediumPriority
	}

	priority := string(t.Priority)
	taskLine := fmt.Sprintf("%s %s (%s)",
		statusStyle.Render(statusSymbol),
		t.Title,
		priorityStyle.Render(priority))

	if index == cursor {
		// Add cursor indicator and highlight
		builder.WriteString("→ " + styles.SelectedItem.Render(taskLine) + "\n")
	} else {
		builder.WriteString("  " + taskLine + "\n")
	}
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the larger of two integers
func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}
