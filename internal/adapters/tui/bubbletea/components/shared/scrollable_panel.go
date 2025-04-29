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

// ScrollablePanelProps contains properties for rendering a panel with a fixed header
// and scrollable content area
type ScrollablePanelProps struct {
	// Basic panel properties
	Width       int
	Height      int
	IsActive    bool
	BorderColor string
	Styles      *Styles

	// Content properties
	Title             string // The panel's title that will appear in the header
	HeaderContent     string // Additional content to show in the fixed header area
	ScrollableContent string // The main content that will be scrollable
	EmptyMessage      string // Message to show when ScrollableContent is empty
	Offset            int    // Current scroll offset for the content
	CursorPosition    int    // Position of the cursor/selected item (to ensure visibility)
}

// RenderScrollablePanel ensures borders are properly set and height takes up available space
func RenderScrollablePanel(props ScrollablePanelProps) string {
	// Build the complete content with header + scrollable area
	var fullContent strings.Builder

	// Add the title with styling
	if props.Title != "" {
		fullContent.WriteString(props.Styles.Title.Render(props.Title) + "\n\n")
	}

	// Add any additional header content if provided
	if props.HeaderContent != "" {
		fullContent.WriteString(props.HeaderContent + "\n\n")
	}

	// Calculate the header height
	headerLines := 0
	if props.Title != "" {
		headerLines += 2 // Title + blank line
	}
	if props.HeaderContent != "" {
		// Count how many lines the header content takes up
		headerContentLines := strings.Count(props.HeaderContent, "\n") + 1
		headerLines += headerContentLines + 1 // +1 for the blank line after header
	}

	// Calculate the available height for the scrollable content
	contentHeight := props.Height - headerLines

	// Ensure we have positive content height
	contentHeight = max(1, contentHeight)

	// Determine which content to render: actual content or empty message
	var scrollableSection string
	if props.ScrollableContent == "" {
		// Show empty message if no content
		if props.EmptyMessage != "" {
			// Pad the message to fill the available height
			msgLines := strings.Count(props.EmptyMessage, "\n") + 1
			if msgLines < contentHeight {
				padding := strings.Repeat("\n", contentHeight-msgLines)
				scrollableSection = props.EmptyMessage + padding
			} else {
				scrollableSection = props.EmptyMessage
			}
		} else {
			// If no empty message is provided, just add padding
			scrollableSection = strings.Repeat("\n", contentHeight)
		}
	} else {
		// We have content, so make it scrollable
		scrollableSection = CreateScrollableContent(
			props.ScrollableContent,
			props.Offset,
			contentHeight,
			props.Styles,
			props.CursorPosition, // Pass cursor position to ensure visibility
		)
	}

	// Add the scrollable section to the full content
	fullContent.WriteString(scrollableSection)

	// Update border colors to make inactive panels invisible
	// Determine border color based on active state
	borderColor := props.BorderColor
	if props.IsActive {
		borderColor = "#00BFFF" // Bright blue color for active panel
	} else {
		borderColor = "#2d3748" // Dark background color that matches terminal background
	}

	// Render the complete panel with borders
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Width(props.Width).
		Height(props.Height).
		Render(fullContent.String())
}
