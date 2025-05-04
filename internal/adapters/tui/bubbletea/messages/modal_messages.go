package messages

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/types"
)

// HelpProvider is an interface for components that can provide context-specific help
type HelpProvider interface {
	HelpContext() string
}

// ShowModalMsg is sent when a modal needs to be shown
type ShowModalMsg struct {
	Content     tea.Model
	Width       int
	Height      int
	DisplayMode types.ModalDisplayMode // How the modal should be displayed
	// Help text will be populated automatically if Content implements HelpProvider
	HelpText    string
}

// NewShowModalMsg creates a new modal message with the specified content and dimensions
func NewShowModalMsg(content tea.Model, width, height int, displayMode types.ModalDisplayMode) ShowModalMsg {
	msg := ShowModalMsg{
		Content:     content,
		Width:       width,
		Height:      height,
		DisplayMode: displayMode,
		HelpText:    "", // Default empty help text
	}
	
	// Check if the content provides help context
	if helpProvider, ok := content.(HelpProvider); ok {
		msg.HelpText = helpProvider.HelpContext()
	}
	
	return msg
}

// HideModalMsg is sent when a modal needs to be hidden
type HideModalMsg struct{}
