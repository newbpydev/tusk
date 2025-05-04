// Package shared contains UI components that can be used across different parts of the application
package shared

import (
	"github.com/charmbracelet/lipgloss"
)

// CycleOption represents an option in a cycle field
type CycleOption struct {
	Value string
	Label string
	Color string // Optional color for styling
}

// CycleField is a component for cycling through a list of options
type CycleField struct {
	Options         []CycleOption
	CurrentIndex    int
	Focused         bool
	Width           int
	ShowCursor      bool
	BorderColor     string
	FocusedBorder   string
	ErrorBorder     string
	HasError        bool
	ErrorMsg        string
	CursorCharacter string
}

// NewCycleField creates a new cycle field with default styling
func NewCycleField(options []CycleOption, width int) *CycleField {
	return &CycleField{
		Options:         options,
		CurrentIndex:    0,
		Focused:         false,
		Width:           width,
		ShowCursor:      true,
		BorderColor:     "#CCCCCC",
		FocusedBorder:   "#2196F3",
		ErrorBorder:     "#F44336",
		CursorCharacter: "█",
	}
}

// CurrentValue returns the current selected value
func (c *CycleField) CurrentValue() string {
	if len(c.Options) == 0 {
		return ""
	}
	return c.Options[c.CurrentIndex].Value
}

// CurrentOption returns the current selected option
func (c *CycleField) CurrentOption() CycleOption {
	if len(c.Options) == 0 {
		return CycleOption{}
	}
	return c.Options[c.CurrentIndex]
}

// Next cycles to the next option
func (c *CycleField) Next() {
	if len(c.Options) == 0 {
		return
	}
	c.CurrentIndex = (c.CurrentIndex + 1) % len(c.Options)
}

// Previous cycles to the previous option
func (c *CycleField) Previous() {
	if len(c.Options) == 0 {
		return
	}
	c.CurrentIndex = (c.CurrentIndex - 1 + len(c.Options)) % len(c.Options)
}

// SetValue sets the current value, finding the matching option
func (c *CycleField) SetValue(value string) {
	for i, opt := range c.Options {
		if opt.Value == value {
			c.CurrentIndex = i
			return
		}
	}
}

// View renders the cycle field
func (c *CycleField) View() string {
	if len(c.Options) == 0 {
		return ""
	}

	currentOption := c.Options[c.CurrentIndex]
	
	// Create the display text
	var displayText string
	
	// If color is specified, use it
	if currentOption.Color != "" {
		colorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(currentOption.Color)).
			Bold(true)
		displayText = colorStyle.Render(currentOption.Label)
	} else {
		displayText = currentOption.Label
	}
	
	// Add cursor if focused
	if c.Focused && c.ShowCursor {
		displayText += c.CursorCharacter
	}
	
	// Apply border styling
	borderColor := c.BorderColor
	if c.HasError {
		borderColor = c.ErrorBorder
	} else if c.Focused {
		borderColor = c.FocusedBorder
	}
	
	// Set width to fill available space
	fieldStyle := lipgloss.NewStyle().
		Width(c.Width).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(0, 1).
		AlignHorizontal(lipgloss.Left) // Align text to the left like other fields
	
	return fieldStyle.Render(displayText)
}
