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
	"github.com/charmbracelet/lipgloss"
)

// FooterProps contains all the data needed to render the footer component
type FooterProps struct {
	Width          int
	ViewMode       string
	HelpStyle      lipgloss.Style
	CursorOnHeader bool   // Whether cursor is on a section header
	CustomHelpText string // Optional custom help text to override defaults
	ActivePanel    int    // The currently active panel (0: task list, 1: task details, 2: timeline)
}

// RenderFooter renders the help text footer for the current view mode
func RenderFooter(props FooterProps) string {
	var help string

	// Use custom help text if provided, otherwise use default based on view mode
	if props.CustomHelpText != "" {
		help = props.CustomHelpText
	} else {
		// Default help text based on view mode and active panel
		switch props.ViewMode {
		case "list":
			// Different help texts based on which panel is active
			switch props.ActivePanel {
			case 0: // Task list panel
				if props.CursorOnHeader {
					// When cursor is on a section header
					help = "j/k: navigate • tab/l/→: next panel • enter: expand/collapse • n: new task • 1/2/3: toggle panels • q: quit"
				} else {
					// Normal task item actions
					help = "j/k: navigate • tab/l/→: next panel • enter: view details • space: toggle completion • e: edit • n: new task • q: quit"
				}
			case 1: // Task details panel
				help = "j/k: scroll • tab/l/→: next panel • shift+tab/h/←/esc: prev panel • e: edit • r: refresh • 1/2/3: toggle panels • q: quit"
			case 2: // Timeline panel
				help = "j/k: scroll • shift+tab/h/←/esc: prev panel • r: refresh • 1/2/3: toggle panels • q: quit"
			}
		case "detail":
			help = "esc: back • e: edit • r: refresh • q: quit"
		case "edit":
			help = "tab: next field • enter: save changes • esc: cancel"
		case "create":
			help = "tab: next field • enter: submit • esc: cancel • space: cycle priority"
		default:
			help = "q: quit • ?: help"
		}
	}

	// Create a prominent footer style
	footerStyle := lipgloss.NewStyle().
		Width(props.Width).
		Align(lipgloss.Center).
		Bold(true).
		Background(lipgloss.Color("#333333")). // Darker background for visibility
		Foreground(lipgloss.Color("#FFFFFF")). // White text for contrast
		Padding(0, 1).
		MarginTop(1)

	return footerStyle.Render(props.HelpStyle.Render(help))
}
