// Package modal provides examples of modal functionality in Tusk
package modal

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/shared"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/types"
)

// Simple application to demonstrate the modal functionality

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
	// Create the main view with a header, content, and footer that EXACTLY matches the app structure
	
	// Define the same background color as in the main app for consistency
	headerBgColor := lipgloss.Color("#2d3748")
	
	// --- HEADER - MUST BE EXACTLY 5 LINES ---
	// This matches the structure in RenderHeader() from header.go
	
	// Create two lines of top padding with correct background - these are crucial
	padding := lipgloss.NewStyle().
		Width(m.width).
		Height(1).
		Background(headerBgColor).
		Render("")
	
	// Logo row
	logoStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#48bb78")).
		Background(headerBgColor).
		PaddingLeft(2)
	
	timeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Background(headerBgColor).
		Bold(true).
		Align(lipgloss.Center)
	
	// First row with logo and time
	row1Left := logoStyle.Width(m.width / 4).Render("TUSK")
	row1Middle := timeStyle.Width(m.width / 2).Render("09:59:28")
	row1Right := lipgloss.NewStyle().Width(m.width / 4).Background(headerBgColor).Render("")
	row1 := lipgloss.JoinHorizontal(lipgloss.Top, row1Left, row1Middle, row1Right)
	
	// Second row with tagline and date
	taglineStyle := lipgloss.NewStyle().
		Italic(true).
		Foreground(lipgloss.Color("#a0aec0")).
		Background(headerBgColor).
		PaddingLeft(2)
	
	dateStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#a0aec0")).
		Background(headerBgColor).
		Align(lipgloss.Center)
	
	row2Left := taglineStyle.Width(m.width / 4).Render("Task Management Simplified")
	row2Middle := dateStyle.Width(m.width / 2).Render("Sunday, May 4, 2025")
	row2Right := lipgloss.NewStyle().Width(m.width / 4).Background(headerBgColor).Render("")
	row2 := lipgloss.JoinHorizontal(lipgloss.Top, row2Left, row2Middle, row2Right)
	
	// Build the exact 5-line header that matches the app
	header := lipgloss.JoinVertical(
		lipgloss.Left,
		padding,    // First padding line
		padding,    // Second padding line
		row1,       // Logo and time row
		row2,       // Tagline and date row
		padding,    // Bottom padding line
	)
	
	// --- CONTENT AREA - BETWEEN HEADER AND FOOTER ---
	// Calculate content height to exactly match the app layout
	contentHeight := m.height - 5 - 1 // header (5) and footer (1)
	
	// Create a sample content area
	contentStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(contentHeight)
		
	content := contentStyle.Render(fmt.Sprintf(
		"\n\n\n\n%s\n\n%s\n\n\n\nTry pressing 'c' to open a ContentArea modal or 'f' for a FullScreen modal.",
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Render("Modal Display Demo"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).Render("This example demonstrates how modals can be displayed either in the content area or full screen."),
	))
	
	// --- FOOTER - EXACTLY 1 LINE ---
	// Create a footer that matches the app's help footer exactly
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#AAAAAA")).
		Background(lipgloss.Color("#222222")).
		Width(m.width).
		Height(1).
		Padding(0, 1)
		
	footer := footerStyle.Render("q: Quit  c: Content Modal  f: Fullscreen Modal  esc: Close")
	
	// Combine all parts with exactly the same structure as the main app
	mainView := lipgloss.JoinVertical(lipgloss.Top, header, content, footer)
	
	// If modal is visible, render it on top
	if m.showModal {
		return m.modal.View(mainView, m.width, m.height)
	}
	
	return mainView
}

// RunModalExample runs the modal example
func RunModalExample() {
	// Initialize the program with alt screen mode for fullscreen TUI
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
	)

	fmt.Println("Starting modal example...")
	fmt.Println("Press 'm' to show the modal, ESC to close it")

	// Run the program without database dependencies
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
