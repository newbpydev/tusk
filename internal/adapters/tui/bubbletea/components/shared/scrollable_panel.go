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

	// Calculate the header height accurately
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
	contentHeight := props.Height - headerLines - 2 // -2 for top/bottom borders

	// Ensure we have positive content height
	contentHeight = max(1, contentHeight)

	// Determine which content to render: actual content or empty message
	var scrollableSection string
	if props.ScrollableContent == "" {
		// Show empty message if no content
		if props.EmptyMessage != "" {
			scrollableSection = props.EmptyMessage
		}
	} else {
		// We have content, so make it scrollable
		scrollableSection = CreateScrollableContent(
			props.ScrollableContent,
			props.Offset,
			contentHeight,
			props.Styles,
			props.CursorPosition,
		)
	}

	// CRITICAL SECTION: Guarantee cursor visibility at all times - highest priority
	// This approach ensures the cursor is always visible during scrolling and window resize
	if props.CursorPosition >= 0 {
		// Split content into lines for accurate calculations
		contentLines := strings.Split(props.ScrollableContent, "\n")
		totalContentLines := len(contentLines)
		
		// Calculate available space for content, accounting for margins and indicators
		// Subtract 4 for panel borders & padding and potentially 2 more for scroll indicators
		reservedSpace := 4
		if props.Height < 10 { // Extra minimal height protection
			reservedSpace = 2
		}
		availableViewportHeight := max(1, props.Height - reservedSpace)
		
		// Count how many visible lines we can show (max viewport capacity)
		// This is critical for determining if scrolling is needed
		maxVisibleLines := availableViewportHeight
		
		// If content is smaller than viewport, no scrolling needed
		if totalContentLines <= maxVisibleLines {
			props.Offset = 0
		} else {
			// Calculate the valid offset range
			maxValidOffset := max(0, totalContentLines - maxVisibleLines)
			
			// Calculate current visible range
			visibleStart := props.Offset
			visibleEnd := min(visibleStart + maxVisibleLines, totalContentLines)
			
			// CRITICAL CHECK: Is cursor within visible area?
			if props.CursorPosition < visibleStart {
				// Cursor is above viewport - scroll up to show it
				// Place cursor at top with context above if possible
				cursorTopPosition := max(0, props.CursorPosition - 1)
				props.Offset = cursorTopPosition
			} else if props.CursorPosition >= visibleEnd {
				// Cursor is below viewport - scroll down to show it
				// Place cursor toward bottom with context below if possible
				cursorBottomPosition := min(maxValidOffset, 
					props.CursorPosition - maxVisibleLines + 2)
				props.Offset = cursorBottomPosition
			}
			
			// Ensure offset is always within valid range
			// This is our final safety check
			props.Offset = max(0, min(props.Offset, maxValidOffset))
		}
	}
	
	// Re-render the scrollable content with our guaranteed-valid offset
	scrollableSection = CreateScrollableContent(
		props.ScrollableContent,
		props.Offset,
		contentHeight,
		props.Styles,
		props.CursorPosition,
	)
	
	// Add the scrollable section to the full content
	fullContent.WriteString(scrollableSection)

	// Determine border color based on active state
	borderColor := props.BorderColor
	if props.IsActive {
		borderColor = "#00BFFF" // Bright blue color for active panel
	} else {
		borderColor = "#2d3748" // Dark background color that matches terminal background
	}

	// Create the panel style
	panelStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Width(props.Width).
		Height(props.Height).
		PaddingTop(1).
		PaddingBottom(1).
		PaddingLeft(2).
		PaddingRight(2).
		MarginTop(0).
		MarginBottom(0).
		MarginLeft(0).
		MarginRight(0)

	return panelStyle.Render(fullContent.String())
}
