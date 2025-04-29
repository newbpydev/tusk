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

// RenderTaskList renders the task list panel with a fixed header and scrollable content
func RenderTaskList(props TaskListProps) string {
	headerContent := ""

	if props.Error != nil {
		headerContent += fmt.Sprintf("Error: %v\n\n", props.Error)
	}

	if props.SuccessMsg != "" {
		headerContent += props.Styles.Done.Render(fmt.Sprintf("✓ %s\n\n", props.SuccessMsg))
		if props.ClearSuccess != nil {
			defer props.ClearSuccess()
		}
	}

	var scrollableContent strings.Builder
	if len(props.Tasks) == 0 {
		scrollableContent.WriteString("No tasks found.\n\nPress 'n' to create a new task.\n")
	} else {
		for i, t := range props.Tasks {
			renderTaskLine(&scrollableContent, t, i, props.Cursor, props.Styles)
		}
	}

	return shared.RenderScrollablePanel(shared.ScrollablePanelProps{
		Title:             "Tasks",
		HeaderContent:     headerContent,
		ScrollableContent: scrollableContent.String(),
		EmptyMessage:      "No tasks available",
		Width:             props.Width,
		Height:            props.Height,
		Offset:            props.Offset,
		CursorPosition:    props.Cursor,
		Styles:            props.Styles,
		IsActive:          props.IsActive,
		BorderColor:       shared.ColorBorder,
	})
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
