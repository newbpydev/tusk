package shared

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/messages"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/types"
)

// SampleModalMsg is sent when a button is clicked in the sample modal
type SampleModalMsg struct {
	Action string
}

// HideModalMessage is the message type to hide the modal
type HideModalMessage struct{}

// SampleModal is an example modal that can be used as a template
type SampleModal struct {
	title       string
	description string
	width       int
	height      int
	buttonFocus int
	// Help information for the modal context
	helpText     string
}

// NewSampleModal creates a new sample modal
func NewSampleModal(title, description string) *SampleModal {
	return &SampleModal{
		title:       title,
		description: description,
		width:       50,
		height:      10,
		buttonFocus: 0,
		helpText:    "[tab] switch focus [enter] select [esc] close",
	}
}

// Init initializes the modal
func (m SampleModal) Init() tea.Cmd {
	return nil
}

// Update handles events for the modal
func (m SampleModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "right", "l":
			m.buttonFocus = (m.buttonFocus + 1) % 2
			return m, nil
		case "shift+tab", "left", "h":
			m.buttonFocus = (m.buttonFocus - 1 + 2) % 2
			return m, nil
		case "enter":
			if m.buttonFocus == 0 {
				// OK button pressed
				return m, func() tea.Msg {
					return SampleModalMsg{Action: "ok"}
				}
			} else {
				// Cancel button pressed
				return m, func() tea.Msg {
					return HideModalMessage{}
				}
			}
		}
	}
	return m, nil
}

// View renders the modal content
func (m SampleModal) View() string {
	// Styles
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).MarginBottom(1)
	descStyle := lipgloss.NewStyle().MarginBottom(1)
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Italic(true).MarginBottom(1)
	
	activeButtonStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#1E88E5")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true).
		Padding(0, 3)
	
	inactiveButtonStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#333333")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 3)
	
	// Buttons
	var okButton, cancelButton string
	if m.buttonFocus == 0 {
		okButton = activeButtonStyle.Render(" OK ")
		cancelButton = inactiveButtonStyle.Render(" Cancel ")
	} else {
		okButton = inactiveButtonStyle.Render(" OK ")
		cancelButton = activeButtonStyle.Render(" Cancel ")
	}
	
	buttonRow := lipgloss.JoinHorizontal(lipgloss.Center, okButton, "  ", cancelButton)
	
	// Modal explanation text
	modalHelp := "Use [tab] to switch buttons, [enter] to select, [esc] to close"
	
	// Content
	content := titleStyle.Render(m.title) + "\n" +
		descStyle.Render(m.description) + "\n" +
		helpStyle.Render(modalHelp) + "\n" +
		buttonRow
	
	// Center everything
	lines := strings.Split(content, "\n")
	centeredLines := make([]string, len(lines))
	
	for i, line := range lines {
		centeredLines[i] = lipgloss.PlaceHorizontal(m.width-4, lipgloss.Center, line)
	}
	
	return strings.Join(centeredLines, "\n")
}

// ShowModalWithDisplayMode creates a sample modal and returns a command with the specified display mode
func ShowModalWithDisplayMode(title, description string, displayMode types.ModalDisplayMode) tea.Cmd {
	return func() tea.Msg {
		modal := NewSampleModal(title, description)
		return messages.NewShowModalMsg(
			modal,
			modal.width,
			modal.height,
			displayMode,
		)
	}
}

// ShowSampleModal creates a sample modal command with default ContentArea display mode
func ShowSampleModal(title, description string) tea.Cmd {
	return ShowModalWithDisplayMode(title, description, types.ContentArea)
}

// ShowFullScreenModal creates a sample modal command with FullScreen display mode
func ShowFullScreenModal(title, description string) tea.Cmd {
	return ShowModalWithDisplayMode(title, description, types.FullScreen)
}

// HelpContext returns the help text for this modal
func (m SampleModal) HelpContext() string {
	return m.helpText
}

// SetHelpText updates the help text for this modal
func (m *SampleModal) SetHelpText(help string) {
	m.helpText = help
}
