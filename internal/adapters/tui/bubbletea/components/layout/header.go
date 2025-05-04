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

package layout

import (
	"time"

	"github.com/charmbracelet/lipgloss"
)

// HeaderProps contains all the data needed to render the header component
type HeaderProps struct {
	Width         int
	CurrentTime   time.Time
	StatusMessage string
	StatusType    string
	IsLoading     bool
}

// RenderHeader creates a header with app name, time, and status information
func RenderHeader(props HeaderProps) string {
	// Define the background color to use consistently
	headerBgColor := lipgloss.Color("#2d3748")

	// Calculate section widths
	logoWidth := props.Width / 4                       // 25% for logo
	timeWidth := props.Width / 2                       // 50% for time in center
	statusWidth := props.Width - logoWidth - timeWidth // Remaining ~25% for status

	// 1. Left Section - Logo
	logoStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#48bb78")).
		Width(logoWidth).
		PaddingLeft(2).
		Background(headerBgColor)

	taglineStyle := lipgloss.NewStyle().
		Italic(true).
		Foreground(lipgloss.Color("#a0aec0")).
		Width(logoWidth).
		PaddingLeft(2).
		Background(headerBgColor)

	// 2. Middle Section - Time
	timeStyle := lipgloss.NewStyle().
		Width(timeWidth).
		Align(lipgloss.Center).
		Bold(true).
		Foreground(lipgloss.Color("#ffffff")).
		Background(headerBgColor)

	dateStyle := lipgloss.NewStyle().
		Width(timeWidth).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("#a0aec0")).
		Background(headerBgColor)

	// 3. Right Section - Status
	statusContainerStyle := lipgloss.NewStyle().
		Width(statusWidth).
		Align(lipgloss.Right).
		PaddingRight(2).
		Background(headerBgColor)

	// Prepare row content
	// First row: Logo + Time + Status
	row1Left := logoStyle.Render("TUSK")
	row1Middle := timeStyle.Render(props.CurrentTime.Format("15:04:05"))

	// Status message with appropriate styling and icon for first row
	var row1Right string
	if props.IsLoading {
		loadingStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#90cdf4")).
			Background(headerBgColor)
		row1Right = statusContainerStyle.Render(loadingStyle.Render("Loading..."))
	} else if props.StatusMessage != "" {
		var msgStyle lipgloss.Style
		var statusIcon string

		switch props.StatusType {
		case "success":
			msgStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#48bb78")).
				Bold(true).
				Background(headerBgColor)
			statusIcon = "✓"
		case "error":
			msgStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#f56565")).
				Bold(true).
				Background(headerBgColor)
			statusIcon = "✗"
		case "info":
			msgStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#4299e1")).
				Bold(true).
				Background(headerBgColor)
			statusIcon = "ℹ"
		default:
			msgStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#a0aec0")).
				Background(headerBgColor)
			statusIcon = "→"
		}
		row1Right = statusContainerStyle.Render(msgStyle.Render(statusIcon + " " + props.StatusMessage))
	} else {
		row1Right = statusContainerStyle.Render("") // Empty status
	}

	// Second row: Tagline + Date + Empty
	row2Left := taglineStyle.Render("Task Management Simplified")
	row2Middle := dateStyle.Render(props.CurrentTime.Format("Monday, January 2, 2006"))
	row2Right := statusContainerStyle.Render("") // Empty space or could be used for additional status info

	// Construct main content rows
	row1 := lipgloss.JoinHorizontal(lipgloss.Top, row1Left, row1Middle, row1Right)
	row2 := lipgloss.JoinHorizontal(lipgloss.Top, row2Left, row2Middle, row2Right)
	
	// Create single-line padding with the correct background color
	padding := lipgloss.NewStyle().
		Width(props.Width).
		Height(1).
		Background(headerBgColor).
		Render("")

	// Create the final header with extra padding above the logo
	// and one line of padding below the date
	header := lipgloss.JoinVertical(
		lipgloss.Left,
		padding,     // First line of top padding
		padding,     // Second line of top padding 
		row1,        // Logo and time row
		row2,        // Tagline and date row
		padding,     // Bottom padding
	)
	
	return header
}
