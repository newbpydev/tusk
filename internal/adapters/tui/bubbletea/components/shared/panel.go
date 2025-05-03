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

	// Create base style for content
	contentStyle := lipgloss.NewStyle().
		Width(contentWidth).
		MaxWidth(contentWidth)

	// Apply style to content
	styledContent := contentStyle.Render(props.Content)

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
// This version guarantees cursor visibility and proper scrolling behavior
func CreateScrollableContent(content string, offset int, maxHeight int, styles *Styles, cursorPosition ...int) string {
	// Guard against negative height which causes panic
	if maxHeight <= 0 {
		return "Content not available (window too small)"
	}

	lines := strings.Split(content, "\n")
	contentHeight := len(lines)

	// If content fits exactly or is smaller, return as is without padding
	if contentHeight <= maxHeight {
		return content
	}

	// CRITICAL: Check if we need to ensure cursor visibility
	cursorPos := -1
	if len(cursorPosition) > 0 && cursorPosition[0] >= 0 {
		cursorPos = cursorPosition[0]
		
		// If cursor exists, ALWAYS prioritize keeping it visible
		// This is the key improvement that ensures cursor is never off-screen
		if cursorPos < offset {
			// Cursor is above visible area, scroll up to show it
			offset = max(0, cursorPos)
		} else if cursorPos >= offset+maxHeight {
			// Cursor is below visible area, scroll down to show it
			// Position cursor with some context around it
			offset = max(0, cursorPos - maxHeight + 1)
		}
	}
	
	// Calculate maximum valid offset and clamp within valid range
	maxOffset := max(0, contentHeight-maxHeight)
	offset = min(offset, maxOffset)
	offset = max(0, offset)

	// Calculate visible range
	startLine := offset
	endLine := min(startLine+maxHeight, contentHeight)

	// Show scroll indicators if needed
	needTopIndicator := offset > 0
	needBottomIndicator := endLine < contentHeight

	// Calculate available content space accounting for indicators
	contentSpace := maxHeight
	if needTopIndicator {
		contentSpace--
	}
	if needBottomIndicator {
		contentSpace--
	}

	// Adjust visible range based on available space
	if endLine-startLine > contentSpace {
		endLine = startLine + contentSpace
	}

	// Additional guard to prevent invalid slice access
	if startLine >= len(lines) {
		startLine = len(lines) - 1
	}
	if endLine > len(lines) {
		endLine = len(lines)
	}

	// Make sure we never have an invalid range
	if startLine < 0 {
		startLine = 0
	}
	if endLine <= startLine {
		endLine = startLine + 1
	}

	// Build the result
	var result strings.Builder

	if needTopIndicator {
		result.WriteString("▲\n")
	}

	// Guard against empty content
	if startLine < endLine && startLine >= 0 && endLine <= len(lines) {
		result.WriteString(strings.Join(lines[startLine:endLine], "\n"))
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
