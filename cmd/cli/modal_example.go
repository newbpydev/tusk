package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/shared"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/types"
)

// Simple application to demonstrate the modal functionality
// This can be run with: go run cmd/cli/modal_example.go

// Initialize a basic model
type model struct {
	showModal bool
	modal     shared.ModalModel
	width     int
	height    int
}

// Initialize the model
func initialModel() model {
	return model{}
}

// Initialize the model with tea.Init
func (m model) Init() tea.Cmd {
	// Show a sample modal after initialization
	return nil
}

// Update the model based on messages
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "c":
			// Show modal in content area mode
			sampleModal := shared.NewSampleModal("Sample Modal", "This is a reusable modal component that can be used throughout the application. Press ESC or click Cancel to close.")
			m.modal = shared.NewModal(sampleModal, 50, 10, types.ContentArea)
			m.modal.Show()
			m.showModal = true
			return m, nil
		case "f":
			// Show modal in fullscreen mode 
			sampleModal := shared.NewSampleModal("Fullscreen Modal", "This is a modal covering the entire screen. Press ESC or click Cancel to close.")
			m.modal = shared.NewModal(sampleModal, 50, 10, types.FullScreen)
			m.modal.Show()
			m.showModal = true
			return m, nil
		case "esc":
			if m.showModal {
				m.showModal = false
				return m, nil
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case shared.SampleModalMsg:
		if msg.Action == "ok" {
			m.showModal = false
		}
		return m, nil
	case shared.ModalCloseMsg, shared.HideModalMessage:
		m.showModal = false
		return m, nil
	}

	// If modal is visible, pass all other messages to it
	if m.showModal {
		var cmd tea.Cmd
		m.modal, cmd = m.modal.Update(msg)
		return m, cmd
	}

	return m, nil
}

// Render the view
func (m model) View() string {
	// Create the main view with a header, content, and footer
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#333333")).
		Width(m.width).
		Align(lipgloss.Center).
		Padding(1, 0)

	header := headerStyle.Render("TUSK\nTask Management Simplified")
	
	// Add 3 more empty lines to make header 5 lines high (matching app layout)
	header += "\n\n\n"

	// Create a sample content area
	contentStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height - 6) // Account for header (5) and footer (1)
		
	content := contentStyle.Render(fmt.Sprintf(
		"\n\n\n\n%s\n\n%s\n\n\n\nTry pressing 'c' to open a ContentArea modal or 'f' for a FullScreen modal.",
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Render("Modal Display Demo"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).Render("This example demonstrates how modals can be displayed either in the content area or full screen."),
	))

	// Create a footer
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#AAAAAA")).
		Background(lipgloss.Color("#222222")).
		Width(m.width)
		
	footer := footerStyle.Render("q: Quit  c: Content Modal  f: Fullscreen Modal  esc: Close")

	// Combine all parts
	mainView := lipgloss.JoinVertical(lipgloss.Top, header, content, footer)
	
	// If modal is visible, render it on top
	if m.showModal {
		return m.modal.View(mainView, m.width, m.height)
	}
	
	return mainView
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
