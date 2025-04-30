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
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/hooks"
	"github.com/newbpydev/tusk/internal/core/task"
)

// TaskListProps contains all properties needed to render the task list panel
type TaskListProps struct {
	Tasks          []task.Task
	Cursor         int
	VisualCursor   int // Cursor position in the collapsible section view
	Offset         int
	Width          int
	Height         int
	Styles         *shared.Styles
	IsActive       bool
	Error          error
	SuccessMsg     string
	ClearSuccess   func()
	CursorOnHeader bool // Whether cursor is on a section header
	CollapsibleMgr *hooks.CollapsibleManager
}

// RenderTaskList renders the task list panel with a fixed header and scrollable content
func RenderTaskList(props TaskListProps) string {
	// Keep track of total visible height for calculations
	const scrollPadding = 3            // Number of lines to keep visible above/below selection
	viewportHeight := props.Height - 4 // Account for borders and header

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
		// Check if collapsible manager is available - if not, fall back to flat list
		if props.CollapsibleMgr == nil {
			renderFlatTaskList(&scrollableContent, props)
		} else {
			// Render tasks in collapsible sections
			renderCollapsibleTaskList(&scrollableContent, props)
		}
	}

	// Calculate the optimal offset to keep selection centered
	totalVisibleItems := props.CollapsibleMgr.GetItemCount()
	halfViewport := (viewportHeight - scrollPadding) / 2

	// Adjust offset to center the selected item
	targetOffset := max(0, props.VisualCursor-halfViewport)

	// Don't scroll past the end
	maxOffset := max(0, totalVisibleItems-viewportHeight+scrollPadding)
	targetOffset = min(targetOffset, maxOffset)

	return shared.RenderScrollablePanel(shared.ScrollablePanelProps{
		Title:             "Tasks",
		HeaderContent:     headerContent,
		ScrollableContent: scrollableContent.String(),
		EmptyMessage:      "No tasks available",
		Width:             props.Width,
		Height:            props.Height,
		Offset:            targetOffset,
		CursorPosition:    props.VisualCursor,
		Styles:            props.Styles,
		IsActive:          props.IsActive,
		BorderColor:       shared.ColorBorder,
	})
}

// renderFlatTaskList renders tasks in a traditional flat list (used as fallback)
func renderFlatTaskList(builder *strings.Builder, props TaskListProps) {
	for i, t := range props.Tasks {
		renderTaskLine(builder, t, i, props.Cursor, props.Styles)
	}
}

// renderCollapsibleTaskList renders tasks organized into collapsible sections
func renderCollapsibleTaskList(builder *strings.Builder, props TaskListProps) {
	// First organize tasks by completion status
	var todoTasks, completedTasks []task.Task

	// Separate tasks by status
	for _, t := range props.Tasks {
		if t.Status == task.StatusDone {
			completedTasks = append(completedTasks, t)
		} else {
			todoTasks = append(todoTasks, t)
		}
	}

	// Clear existing sections
	props.CollapsibleMgr.ClearSections()

	// Add our sections
	// Todo tasks section (expanded by default)
	props.CollapsibleMgr.AddSection(hooks.SectionTypeTodo, "Todo", len(todoTasks), 0)

	// Projects section - currently empty as placeholder
	props.CollapsibleMgr.AddSection(hooks.SectionTypeProjects, "Projects", 0, len(todoTasks))

	// Completed tasks section
	props.CollapsibleMgr.AddSection(hooks.SectionTypeCompleted, "Completed", len(completedTasks), len(todoTasks))

	// Now render the sections and their contents
	var visibleIndex int = 0

	// Todo section
	visibleIndex = renderSection(builder, props, hooks.SectionTypeTodo, todoTasks, visibleIndex)

	// Projects section (placeholder for now)
	visibleIndex = renderSection(builder, props, hooks.SectionTypeProjects, nil, visibleIndex)

	// Completed tasks section
	renderSection(builder, props, hooks.SectionTypeCompleted, completedTasks, visibleIndex)
}

// renderSection renders a collapsible section and its tasks if expanded
func renderSection(builder *strings.Builder, props TaskListProps, sectionType hooks.SectionType, sectionTasks []task.Task, visibleIndex int) int {
	// Get section settings
	var isExpanded bool
	var sectionTitle string

	for _, section := range props.CollapsibleMgr.Sections {
		if section.Type == sectionType {
			isExpanded = section.IsExpanded
			sectionTitle = section.Title
			break
		}
	}

	// Build section header with collapse/expand indicator
	headerSymbol := "▼ " // Expanded
	if !isExpanded {
		headerSymbol = "► " // Collapsed
	}

	// Format the section header
	headerText := fmt.Sprintf("%s%s (%d)", headerSymbol, sectionTitle, len(sectionTasks))

	// Check if this section header is selected
	if props.CursorOnHeader && props.VisualCursor == visibleIndex {
		builder.WriteString("→ " + props.Styles.SelectedItem.Render(headerText) + "\n")
	} else {
		// Regular section header - use a distinctive style to show it's clickable
		builder.WriteString("  " + props.Styles.Title.Render(headerText) + "\n")
	}
	visibleIndex++

	// If section is expanded, render its tasks
	if isExpanded && len(sectionTasks) > 0 {
		for i, t := range sectionTasks {
			// Calculate the visible index of this task
			taskVisibleIndex := visibleIndex + i

			// Determine if this task is selected
			isSelected := !props.CursorOnHeader && props.VisualCursor == taskVisibleIndex

			// Render with additional indentation for tree-like appearance
			renderTaskLineWithIndent(builder, t, i, isSelected, props.Styles)
		}
		visibleIndex += len(sectionTasks)
	}

	// Only add spacing after non-final sections
	if sectionType != hooks.SectionTypeCompleted {
		builder.WriteString("\n")
	}

	return visibleIndex
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

// renderTaskLineWithIndent renders a task line with indentation for the tree view
func renderTaskLineWithIndent(builder *strings.Builder, t task.Task, index int, isSelected bool, styles *shared.Styles) {
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

	if isSelected {
		// Add cursor indicator, indentation and highlight
		builder.WriteString("→   " + styles.SelectedItem.Render(taskLine) + "\n")
	} else {
		// Add indentation only
		builder.WriteString("    " + taskLine + "\n")
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
