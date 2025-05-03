// Package section_header provides a reusable component for rendering section headers
// in the TUI application.
package section_header

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/hooks"
	"github.com/newbpydev/tusk/internal/core/task"
)

// SectionHeaderOptions configures how a section header should be rendered
type SectionHeaderOptions struct {
	// IsSelected indicates if this section is currently selected
	IsSelected bool
	
	// Width is the available width for rendering
	Width int
	
	// IncludeCount determines if the count of items should be shown
	IncludeCount bool
}

// DefaultSectionHeaderOptions returns the default options for rendering a section header
func DefaultSectionHeaderOptions() SectionHeaderOptions {
	return SectionHeaderOptions{
		IsSelected:   false,
		Width:        80,
		IncludeCount: true,
	}
}

// SectionData contains the data needed to render a section header
type SectionData struct {
	Type       hooks.SectionType
	Title      string
	Items      []task.Task
	IsExpanded bool
}

// RenderSectionHeader creates a string representation of a section header for display
// This is a pure function with no side effects, making it easy to test and reuse
func RenderSectionHeader(section SectionData, opts SectionHeaderOptions) string {
	// Determine icon based on whether the section is open or closed
	var icon string
	if section.IsExpanded {
		icon = "▼"
	} else {
		icon = "▶"
	}
	
	// Define header styles
	selectedHeaderStyle := lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("#1E88E5")).Foreground(lipgloss.Color("#FFFFFF"))
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	
	// Get base style based on selection
	var baseStyle lipgloss.Style
	if opts.IsSelected {
		baseStyle = selectedHeaderStyle
	} else {
		baseStyle = headerStyle
	}
	
	// Format the header with appropriate style based on section type
	var header string
	if opts.IncludeCount {
		header = fmt.Sprintf("%s %s (%d)", icon, section.Title, len(section.Items))
	} else {
		header = fmt.Sprintf("%s %s", icon, section.Title)
	}
	
	// Apply section-specific styling
	switch section.Type {
	case hooks.SectionTypeTodo:
		header = baseStyle.Render(header)
	case hooks.SectionTypeProjects:
		header = baseStyle.Render(header)
	case hooks.SectionTypeCompleted:
		header = baseStyle.Render(header)
	case hooks.SectionTypeOverdue:
		header = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF0000")).Render(header)
	case hooks.SectionTypeToday:
		header = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00E676")).Render(header)
	case hooks.SectionTypeUpcoming:
		header = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#1E88E5")).Render(header)
	default:
		header = baseStyle.Render(header)
	}
	
	return header
}

// GetSectionTypeColor returns a color string based on section type for consistent styling
func GetSectionTypeColor(sectionType hooks.SectionType) string {
	switch sectionType {
	case hooks.SectionTypeTodo:
		return "#2196F3" // Blue
	case hooks.SectionTypeProjects:
		return "#4CAF50" // Green
	case hooks.SectionTypeCompleted:
		return "#8BC34A" // Light Green
	case hooks.SectionTypeOverdue:
		return "#FF0000" // Red
	case hooks.SectionTypeToday:
		return "#00E676" // Green
	case hooks.SectionTypeUpcoming:
		return "#1E88E5" // Blue
	default:
		return "#FFFFFF" // White
	}
}
