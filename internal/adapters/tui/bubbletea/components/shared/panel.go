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
