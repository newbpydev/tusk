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
)

// SectionHeaderMessageProps contains properties for rendering a message when a section header is selected
type SectionHeaderMessageProps struct {
	SectionName string
	Width       int
	Height      int
	Styles      *shared.Styles
	Offset      int  // For scrolling the content
	IsActive    bool // Whether this panel is active (for styling)
}

// RenderSectionHeaderMessage renders a helpful message when a section header is selected
// This improves user experience by providing context about what a section is
func RenderSectionHeaderMessage(props SectionHeaderMessageProps) string {
	var sb strings.Builder

	// Create a title with the section name
	title := fmt.Sprintf("Section: %s", props.SectionName)
	sb.WriteString(props.Styles.Title.Render(title) + "\n\n")

	// Add a helpful description based on the section
	switch props.SectionName {
	case "Overdue":
		sb.WriteString("This section contains tasks that are past their due date.\n")
		sb.WriteString("These tasks require your immediate attention.\n\n")
		sb.WriteString("To see task details, select a specific task instead of the section header.\n")
		
	case "Today":
		sb.WriteString("This section contains tasks due today.\n")
		sb.WriteString("Focus on completing these tasks before the end of the day.\n\n")
		sb.WriteString("To see task details, select a specific task instead of the section header.\n")
		
	case "Upcoming":
		sb.WriteString("This section contains tasks due in the future.\n")
		sb.WriteString("Plan ahead and prepare for these upcoming tasks.\n\n")
		sb.WriteString("To see task details, select a specific task instead of the section header.\n")
		
	case "Todo":
		sb.WriteString("This section contains your active tasks with no project assignment.\n")
		sb.WriteString("These are your day-to-day tasks that need to be completed.\n\n")
		sb.WriteString("To see task details, select a specific task instead of the section header.\n")
		
	case "Projects":
		sb.WriteString("This section contains tasks associated with specific projects.\n")
		sb.WriteString("Group related tasks under projects for better organization.\n\n")
		sb.WriteString("To see task details, select a specific task instead of the section header.\n")
		
	case "Completed":
		sb.WriteString("This section contains tasks you have already completed.\n")
		sb.WriteString("Keep track of your accomplishments and review your progress.\n\n")
		sb.WriteString("To see task details, select a specific task instead of the section header.\n")
		
	default:
		sb.WriteString("This is a section header that groups related tasks together.\n\n")
		sb.WriteString("To see task details, select a specific task instead of the section header.\n")
	}

	// Add keyboard shortcuts
	sb.WriteString("\n" + props.Styles.Help.Render("Keyboard Shortcuts:") + "\n")
	sb.WriteString(props.Styles.Help.Render("- Space: ") + "Toggle section expand/collapse\n")
	sb.WriteString(props.Styles.Help.Render("- j/k: ") + "Navigate up/down\n")
	sb.WriteString(props.Styles.Help.Render("- Tab: ") + "Switch between panels\n")

	// Use the scrollable panel component to match the style of other panels
	return shared.RenderScrollablePanel(shared.ScrollablePanelProps{
		Title:              "Task Details",
		ScrollableContent:  sb.String(),
		EmptyMessage:       "No section header selected.",
		Width:              props.Width,
		Height:             props.Height,
		Offset:             props.Offset, // Use passed offset for consistent scrolling
		Styles:             props.Styles,
		IsActive:           props.IsActive, // Use passed active state for proper styling
		BorderColor:        shared.ColorBorder,
		CursorPosition:     props.Offset, // Match cursor position with offset for scrolling
	})
}
