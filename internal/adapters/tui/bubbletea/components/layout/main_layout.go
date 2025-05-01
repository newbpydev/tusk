package layout

import (
	"time"

	"github.com/charmbracelet/lipgloss"
)

// MainLayoutProps contains all the properties needed to render the main application layout.
type MainLayoutProps struct {
	// Header properties
	Width         int
	CurrentTime   time.Time
	StatusMessage string
	StatusType    string
	IsLoading     bool

	// Main content
	Content string
	
	// Footer properties
	ViewMode       string
	HelpStyle      lipgloss.Style
	CursorOnHeader bool
	HelpText       string // Custom help text for the current view
}

// RenderMainLayout creates a consistent layout with header, content, and footer.
// This serves as the main container for all screens in the application.
func RenderMainLayout(props MainLayoutProps) string {
	// Render the header
	header := RenderHeader(HeaderProps{
		Width:         props.Width,
		CurrentTime:   props.CurrentTime,
		StatusMessage: props.StatusMessage,
		StatusType:    props.StatusType,
		IsLoading:     props.IsLoading,
	})

	// Middle content is provided by the caller
	content := props.Content

	// Render the footer with appropriate help text
	footer := RenderFooter(FooterProps{
		Width:          props.Width,
		ViewMode:       props.ViewMode,
		HelpStyle:      props.HelpStyle,
		CursorOnHeader: props.CursorOnHeader,
		CustomHelpText: props.HelpText,
	})

	// Combine all sections
	return lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
}
