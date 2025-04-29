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

package shared

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// PanelProps contains all properties needed to render a panel wrapper
type PanelProps struct {
	Content     string
	Width       int
	Height      int
	IsActive    bool
	BorderColor string
}

// RenderPanel wraps content without borders to avoid double borders
func RenderPanel(props PanelProps) string {
	const borderWidth = 0     // Set border width to 0 to remove borders
	const paddingWidth = 0    // Keep padding as is
	const totalFrameWidth = 0 // No extra width for borders

	// Content width is panel width minus frame elements for consistency
	contentWidth := props.Width - totalFrameWidth
	contentHeight := props.Height - 2 // Account for top and bottom borders

	// Create base style for content
	contentStyle := lipgloss.NewStyle().
		Width(contentWidth).
		MaxWidth(contentWidth)

	// Apply style to content
	styledContent := contentStyle.Render(props.Content)

	// Ensure the content fills the panel height by adding padding if needed
	contentHeight = max(0, contentHeight) // Ensure non-negative height

	// Count lines in the content
	lines := strings.Split(styledContent, "\n")
	lineCount := len(lines)

	// If content is shorter than available space, pad with empty lines
	if lineCount < contentHeight {
		paddingLines := contentHeight - lineCount
		padding := strings.Repeat("\n", paddingLines)
		styledContent += padding
	}

	return styledContent
}

// max returns the larger of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// CreateScrollableContent creates a scrollable view of given content
func CreateScrollableContent(content string, offset int, maxHeight int, styles *Styles, cursorPosition ...int) string {
	// Guard against negative height which causes panic
	if maxHeight <= 0 {
		return "Content not available (window too small)"
	}

	lines := strings.Split(content, "\n")

	// Calculate actual content height
	contentHeight := len(lines)

	// If scrolling not needed, just return the whole content padded to maxHeight
	if contentHeight <= maxHeight {
		// Pad content to maxHeight to maintain consistent height
		if len(lines) < maxHeight {
			paddingLines := maxHeight - len(lines)
			padding := strings.Repeat("\n", paddingLines)
			return content + padding
		}
		return content
	}

	// Check if we need to adjust scroll position based on cursor
	// This ensures the selected item is always visible
	if len(cursorPosition) > 0 && cursorPosition[0] >= 0 {
		cursor := cursorPosition[0]

		// Define a comfortable padding to keep around the cursor
		const visibilityPadding = 2

		// Calculate viewport boundaries
		viewportStart := offset
		viewportEnd := offset + maxHeight - 1

		// Adjust offset if cursor would be outside visible area
		if cursor < viewportStart+visibilityPadding {
			// Cursor is above the viewport or too close to top
			offset = max(0, cursor-visibilityPadding)
		} else if cursor > viewportEnd-visibilityPadding {
			// Cursor is below the viewport or too close to bottom
			offset = min(contentHeight-maxHeight, cursor-maxHeight+visibilityPadding+1)
		}
	}

	// Clamp offset within valid range
	maxOffset := max(0, contentHeight-maxHeight)
	offset = min(offset, maxOffset)
	offset = max(0, offset)

	// Apply offset and limit number of lines to maxHeight
	startLine := min(offset, len(lines))
	endLine := min(startLine+maxHeight, len(lines))

	// Safety check: ensure we have at least one line to show
	if startLine >= endLine || startLine >= len(lines) {
		return "Error: Cannot display content (display area too small)"
	}

	// Calculate available content space
	// For scroll indicators we use 2 lines max (top and bottom)
	contentSpace := maxHeight
	needTopIndicator := offset > 0
	needBottomIndicator := offset < maxOffset

	if needTopIndicator {
		contentSpace--
	}
	if needBottomIndicator {
		contentSpace--
	}

	// Make sure we don't try to show more lines than available space
	if endLine-startLine > contentSpace {
		endLine = startLine + contentSpace
	}

	visibleLines := lines[startLine:endLine]

	// Calculate padding to maintain consistent height
	remainingSpace := contentSpace - len(visibleLines)
	padding := ""
	if remainingSpace > 0 {
		padding = strings.Repeat("\n", remainingSpace)
	}

	// Build the content with indicators and padding
	var result strings.Builder

	if needTopIndicator {
		result.WriteString("▲\n")
	}

	result.WriteString(strings.Join(visibleLines, "\n"))

	// Add padding to maintain consistent height
	if padding != "" {
		result.WriteString(padding)
	}

	if needBottomIndicator {
		result.WriteString("\n▼")
	}

	return result.String()
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
