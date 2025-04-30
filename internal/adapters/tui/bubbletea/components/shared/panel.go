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
	contentHeight := len(lines)

	// If scrolling not needed, just return the whole content padded to maxHeight
	if contentHeight <= maxHeight {
		if len(lines) < maxHeight {
			paddingLines := maxHeight - len(lines)
			padding := strings.Repeat("\n", paddingLines)
			return content + padding
		}
		return content
	}

	// Define a comfortable padding to keep around the cursor
	const visibilityPadding = 2

	// Calculate maximum valid offset
	maxOffset := max(0, contentHeight-maxHeight)

	// If we have a cursor position, ensure it's visible and centered if possible
	if len(cursorPosition) > 0 && cursorPosition[0] >= 0 {
		cursor := cursorPosition[0]

		// Try to center the cursor in the viewport
		halfHeight := maxHeight / 2
		targetOffset := cursor - halfHeight

		// Ensure we don't scroll past the content boundaries
		if targetOffset > maxOffset {
			targetOffset = maxOffset
		}
		if targetOffset < 0 {
			targetOffset = 0
		}

		offset = targetOffset
	}

	// Clamp offset within valid range
	offset = min(offset, maxOffset)
	offset = max(0, offset)

	// Calculate visible range
	startLine := offset
	endLine := min(startLine+maxHeight, contentHeight)

	// Show scroll indicators if needed
	needTopIndicator := offset > 0
	needBottomIndicator := offset < maxOffset

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

	// Add padding if needed to maintain consistent height
	remainingSpace := contentSpace - (endLine - startLine)
	if remainingSpace > 0 {
		result.WriteString(strings.Repeat("\n", remainingSpace))
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
