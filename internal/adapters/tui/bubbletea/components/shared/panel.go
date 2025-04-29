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

// RenderPanel wraps content with a panel border
func RenderPanel(props PanelProps) string {
	const borderWidth = 1     // Width of border on each side
	const paddingWidth = 0    // Reduced padding to 0 (was 1)
	const totalFrameWidth = 2 // Total extra width: (borderWidth) * 2 (removed padding)

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

	// Create frame style - either with visible border or with spacing
	var frameStyle lipgloss.Style
	if props.IsActive {
		// Active panel - visible borders with custom or default color
		borderColor := props.BorderColor
		if borderColor == "" {
			borderColor = ColorBorder
		}

		frameStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(borderColor)).
			BorderLeft(true).
			BorderRight(true).
			BorderTop(true).
			BorderBottom(true).
			Padding(0, 0, 0, 0).     // Removed all padding
			Width(props.Width - 2).  // Account for left and right border
			Height(props.Height - 2) // Account for top and bottom border
	} else {
		// Inactive panel - invisible placeholder borders
		frameStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.HiddenBorder()).
			BorderLeft(true).
			BorderRight(true).
			BorderTop(true).
			BorderBottom(true).
			Padding(0, 0, 0, 0).     // Removed all padding
			Width(props.Width - 2).  // Account for left and right border
			Height(props.Height - 2) // Account for top and bottom border
	}

	// Apply frame and return
	return frameStyle.Render(styledContent)
}

// max returns the larger of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// CreateScrollableContent creates a scrollable view of given content
func CreateScrollableContent(content string, offset int, maxHeight int, styles *Styles) string {
	// Guard against negative height which causes panic
	if maxHeight <= 0 {
		return "Content not available (window too small)"
	}

	lines := strings.Split(content, "\n")

	// Calculate actual content height
	contentHeight := len(lines)

	// Determine if scrolling is needed
	needsScrolling := contentHeight > maxHeight

	// If scrolling not needed, just return the whole content padded to maxHeight
	if !needsScrolling {
		// Pad content to maxHeight to maintain consistent height
		if len(lines) < maxHeight {
			paddingLines := maxHeight - len(lines)
			padding := strings.Repeat("\n", paddingLines)
			return content + padding
		}
		return content
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
