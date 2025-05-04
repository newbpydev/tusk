package messages

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/types"
)

// ShowModalMsg is sent when a modal needs to be shown
type ShowModalMsg struct {
	Content     tea.Model
	Width       int
	Height      int
	DisplayMode types.ModalDisplayMode // How the modal should be displayed
}

// NewShowModalMsg creates a new modal message with the specified content and dimensions
func NewShowModalMsg(content tea.Model, width, height int, displayMode types.ModalDisplayMode) ShowModalMsg {
	return ShowModalMsg{
		Content:     content,
		Width:       width,
		Height:      height,
		DisplayMode: displayMode,
	}
}

// HideModalMsg is sent when a modal needs to be hidden
type HideModalMsg struct{}
