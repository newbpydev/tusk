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
	Width     int
	ViewMode  string
	HelpStyle lipgloss.Style
}

// RenderFooter renders the help text footer for the current view mode
func RenderFooter(props FooterProps) string {
	var help string

	switch props.ViewMode {
	case "list":
		help = "j/k: navigate • h/l or ←/→: switch panels • enter: view details • c: toggle completion • n: new task • 1/2/3: toggle columns • q: quit"
	case "detail":
		help = "esc: back • h/l or ←/→: switch panels • e: edit • c: toggle completion • d: delete • n: new task"
	case "edit":
		help = "esc: cancel • enter: save changes"
	case "create":
		help = "tab: next field • enter: submit • esc: cancel • space: cycle priority"
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
