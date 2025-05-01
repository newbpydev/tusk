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

// TaskDetailsProps contains all properties needed to render the task details panel
type TaskDetailsProps struct {
	Tasks          []task.Task
	Cursor         int
	SelectedTask   *task.Task // Direct reference to the selected task
	Offset         int
	Width          int
	Height         int
	Styles         *shared.Styles
	IsActive       bool
	CursorOnHeader bool // whether selection is on a section header
}

// RenderTaskDetails renders the task details panel with a fixed header and scrollable content
func RenderTaskDetails(props TaskDetailsProps) string {
	var scrollableContent strings.Builder

	// If cursor is on a section header or out of valid range, show placeholder
	if props.CursorOnHeader || props.SelectedTask == nil && (props.Cursor < 0 || props.Cursor >= len(props.Tasks)) {
		scrollableContent.WriteString(props.Styles.Help.Render("Select a task to view details"))
		return shared.RenderScrollablePanel(shared.ScrollablePanelProps{
			Title:             "Task Details",
			ScrollableContent: scrollableContent.String(),
			EmptyMessage:      "No task selected",
			Width:             props.Width,
			Height:            props.Height,
			Offset:            props.Offset,
			Styles:            props.Styles,
			IsActive:          props.IsActive,
			BorderColor:       shared.ColorBorder,
			// Use cursor position to keep details in view
			CursorPosition: props.Cursor,
		})
	}

	if len(props.Tasks) == 0 && props.SelectedTask == nil {
		scrollableContent.WriteString("No tasks yet. Press 'n' to create your first task.\n\n")
		scrollableContent.WriteString(props.Styles.Help.Render("Tip: You can organize tasks with priorities and due dates!"))
	} else {
		// Use SelectedTask if provided, otherwise use task at cursor position
		var t task.Task
		if props.SelectedTask != nil {
			t = *props.SelectedTask
		} else if props.Cursor < len(props.Tasks) {
			t = props.Tasks[props.Cursor]
		}

		// Add more detailed task information with formatting to make it more scrollable
		scrollableContent.WriteString(props.Styles.Title.Render("Title: ") + t.Title + "\n\n")

		// Status with appropriate styling
		statusLabel := props.Styles.Title.Render("Status: ")
		var statusStyle = props.Styles.Todo
		switch t.Status {
		case task.StatusDone:
			statusStyle = props.Styles.Done
		case task.StatusInProgress:
			statusStyle = props.Styles.InProgress
		}
		scrollableContent.WriteString(statusLabel + statusStyle.Render(string(t.Status)) + "\n\n")

		// Priority with appropriate styling
		priorityLabel := props.Styles.Title.Render("Priority: ")
		var priorityStyle = props.Styles.LowPriority
		switch t.Priority {
		case task.PriorityHigh:
			priorityStyle = props.Styles.HighPriority
		case task.PriorityMedium:
			priorityStyle = props.Styles.MediumPriority
		}
		scrollableContent.WriteString(priorityLabel + priorityStyle.Render(string(t.Priority)) + "\n\n")

		// Due date if available
		if t.DueDate != nil {
			dueLabel := props.Styles.Title.Render("Due Date: ")
			dueDate := t.DueDate.Format("2006-01-02")

			// Show if overdue
			if t.DueDate.Before(time.Now()) && t.Status != task.StatusDone {
				dueDate = props.Styles.HighPriority.Render(dueDate + " (Overdue)")
			}

			scrollableContent.WriteString(dueLabel + dueDate + "\n\n")
		}

		// Description
		descriptionLabel := props.Styles.Title.Render("Description: ")
		if t.Description != nil && *t.Description != "" {
			// Format description with word wrapping to fit panel
			description := *t.Description
			scrollableContent.WriteString(descriptionLabel + "\n" + description + "\n\n")
		} else {
			scrollableContent.WriteString(descriptionLabel + "No description provided\n\n")
		}

		// Created/Updated timestamps
		if !t.CreatedAt.IsZero() {
			scrollableContent.WriteString(props.Styles.Title.Render("Created: ") + t.CreatedAt.Format("2006-01-02 15:04") + "\n")
		}

		if !t.UpdatedAt.IsZero() {
			scrollableContent.WriteString(props.Styles.Title.Render("Updated: ") + t.UpdatedAt.Format("2006-01-02 15:04") + "\n")
		}

		// Add extra info to make the content more scrollable for testing
		scrollableContent.WriteString("\n" + props.Styles.Help.Render("Task ID: ") + fmt.Sprintf("%d", t.ID) + "\n")

		// Add help text at the bottom
		scrollableContent.WriteString("\n" + props.Styles.Help.Render("Press 'e' to edit task") + "\n")
		scrollableContent.WriteString(props.Styles.Help.Render("Press 'c' to toggle completion") + "\n")
		scrollableContent.WriteString(props.Styles.Help.Render("Press 'd' to delete task") + "\n")
	}

	return shared.RenderScrollablePanel(shared.ScrollablePanelProps{
		Title:             "Task Details",
		ScrollableContent: scrollableContent.String(),
		EmptyMessage:      "No tasks available",
		Width:             props.Width,
		Height:            props.Height,
		Offset:            props.Offset,
		Styles:            props.Styles,
		IsActive:          props.IsActive,
		BorderColor:       shared.ColorBorder,
		CursorPosition:    props.Cursor,
	})
}
