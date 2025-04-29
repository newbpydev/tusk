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
	"github.com/charmbracelet/lipgloss"
)

// Color constants
const (
	ColorWhite     = "#FFFFFF"
	ColorOffWhite  = "#FAFAFA"
	ColorLightGray = "#909090"
	ColorDarkGray  = "#747474"
	ColorBlue      = "#1E88E5"
	ColorYellow    = "#FFB300"
	ColorGreen     = "#00E676"
	ColorTeal      = "#009688"
	ColorOrange    = "#FB8C00"
	ColorRed       = "#E53935"
	ColorBorder    = "#4B9CD3" // Light blue color for borders
)

// Styles encapsulates all UI styling for the TUI
type Styles struct {
	// General UI styles
	Title        lipgloss.Style
	SelectedItem lipgloss.Style
	Help         lipgloss.Style
	ActiveBorder lipgloss.Style
	Background   lipgloss.Style // New field for background color

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
	s.Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(ColorOffWhite))
	s.SelectedItem = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color(ColorBlue)).Foreground(lipgloss.Color(ColorWhite))
	s.Help = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorDarkGray)).Italic(true)
	s.ActiveBorder = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorBorder)).
		Padding(0, 1)
	s.Background = lipgloss.NewStyle().Background(lipgloss.Color(ColorOffWhite)) // Initialize background color

	// Set up task status styles
	s.Todo = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorLightGray))
	s.InProgress = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorYellow))
	s.Done = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGreen))

	// Set up priority styles
	s.LowPriority = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorTeal))
	s.MediumPriority = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorOrange))
	s.HighPriority = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorRed))

	return s
}

// DefaultStyles instance to be used throughout the application
var DefaultStylesInstance = DefaultStyles()
