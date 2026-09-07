package layout

import (
	"time"

	"github.com/charmbracelet/lipgloss"
)

// MainLayoutProps contains all the properties needed to render the main application layout.
type MainLayoutProps struct {
	// Header properties
	Width         int
	Height        int // Total available height of the screen
	CurrentTime   time.Time
	StatusMessage string
	StatusType    string
	IsLoading     bool

	// Main content
	Content string
}

// RenderMainLayout creates a consistent layout with header and content.
// This serves as the main container for all screens in the application.
func RenderMainLayout(props MainLayoutProps) string {
	// Define layout constants
	const headerHeight = 5
	// Reserve space for help footer which will be rendered separately
	const helpFooterHeight = 1
	
	// Render the header
	header := RenderHeader(HeaderProps{
		Width:         props.Width,
		CurrentTime:   props.CurrentTime,
		StatusMessage: props.StatusMessage,
		StatusType:    props.StatusType,
		IsLoading:     props.IsLoading,
	})

	// Calculate content height to fill available space between header and help footer
	contentHeight := props.Height - headerHeight - helpFooterHeight
	
	// Create a style for the content to ensure it takes the full available height
	contentStyle := lipgloss.NewStyle().
		Width(props.Width).
		Height(contentHeight)
	
	// Style the content to ensure it takes the full available height
	content := contentStyle.Render(props.Content)

	// Combine header and content with proper vertical alignment
	layout := lipgloss.JoinVertical(lipgloss.Left, header, content)
	
	// Create a full-screen container that ensures the layout takes up the entire screen
	fullScreenStyle := lipgloss.NewStyle().
		Width(props.Width).
		Height(props.Height)
	
	return fullScreenStyle.Render(layout)
}
