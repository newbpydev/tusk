// Package shared contains UI components that can be used across different parts of the application
package shared

import (
	"github.com/charmbracelet/lipgloss"
)

// ButtonStyle represents styling options for buttons
type ButtonStyle struct {
	Focused      bool
	Primary      bool
	Width        int
	TextAlign    lipgloss.Position
	FixedWidth   bool
	BorderRadius int
}

// DefaultButtonStyle creates standard button styling
func DefaultButtonStyle() ButtonStyle {
	return ButtonStyle{
		Focused:      false,
		Primary:      false,
		Width:        0,
		TextAlign:    lipgloss.Center,
		FixedWidth:   false,
		BorderRadius: 1,
	}
}

// Button renders a styled button with consistent focus handling
func Button(label string, style ButtonStyle) string {
	var buttonStyle lipgloss.Style

	if style.Primary {
		// Primary button (like Save)
		if style.Focused {
			// Primary button in focused state - blue highlight
			buttonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#1E88E5")).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#1E88E5")).
				Padding(0, 2)
		} else {
			// Primary button in unfocused state - use normal (gray) color like other unfocused buttons
			buttonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#78909C")).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#CCCCCC")).
				Padding(0, 2)
		}
	} else {
		// Secondary button (like Cancel)
		if style.Focused {
			// Secondary button in focused state - use blue like primary buttons for consistency
			buttonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#1E88E5")).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#1E88E5")).
				Padding(0, 2)
		} else {
			// Secondary button in unfocused state
			buttonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#78909C")).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#CCCCCC")).
				Padding(0, 2)
		}
	}

	// Apply width constraints if specified
	if style.Width > 0 {
		// lipgloss doesn't have MinWidth, so we use Width for both cases
		buttonStyle = buttonStyle.Width(style.Width)
	}

	buttonStyle = buttonStyle.Align(style.TextAlign)

	return buttonStyle.Render(label)
}
