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

package bubbletea

import (
	"github.com/charmbracelet/lipgloss"
)

// Color constants
const (
	colorWhite     = "#FFFFFF"
	colorOffWhite  = "#FAFAFA"
	colorLightGray = "#909090"
	colorDarkGray  = "#747474"
	colorBlue      = "#1E88E5"
	colorYellow    = "#FFB300"
	colorGreen     = "#00E676"
	colorTeal      = "#009688"
	colorOrange    = "#FB8C00"
	colorRed       = "#E53935"
	colorBorder    = "#4B9CD3" // Light blue color for borders
)

// Styles encapsulates all UI styling for the TUI
type Styles struct {
	// General UI styles
	Title        lipgloss.Style
	SelectedItem lipgloss.Style
	Help         lipgloss.Style
	ActiveBorder lipgloss.Style

	// Task status styles
	Todo       lipgloss.Style
	InProgress lipgloss.Style
	Done       lipgloss.Style

	// Priority styles
	LowPriority    lipgloss.Style
	MediumPriority lipgloss.Style
	HighPriority   lipgloss.Style
}

// DefaultStyles returns a Styles struct with the default styling
func DefaultStyles() *Styles {
	s := new(Styles)

	// Set up general UI styles
	s.Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colorOffWhite))
	s.SelectedItem = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color(colorBlue)).Foreground(lipgloss.Color(colorWhite))
	s.Help = lipgloss.NewStyle().Foreground(lipgloss.Color(colorDarkGray)).Italic(true)
	s.ActiveBorder = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colorBorder)).
		Padding(0, 1)

	// Set up task status styles
	s.Todo = lipgloss.NewStyle().Foreground(lipgloss.Color(colorLightGray))
	s.InProgress = lipgloss.NewStyle().Foreground(lipgloss.Color(colorYellow))
	s.Done = lipgloss.NewStyle().Foreground(lipgloss.Color(colorGreen))

	// Set up priority styles
	s.LowPriority = lipgloss.NewStyle().Foreground(lipgloss.Color(colorTeal))
	s.MediumPriority = lipgloss.NewStyle().Foreground(lipgloss.Color(colorOrange))
	s.HighPriority = lipgloss.NewStyle().Foreground(lipgloss.Color(colorRed))

	return s
}

// ActiveStyles holds the current active styles for the application
var ActiveStyles = DefaultStyles()
