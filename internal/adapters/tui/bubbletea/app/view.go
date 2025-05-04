// Package app contains the main TUI application implementation using bubbletea.  
package app

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/shared"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/keymap"
)

// View implements tea.Model View, displaying the appropriate view based on the current mode.
// It leverages the reusable MainLayout component for consistent UI across all views.
func (m *Model) View() string {
	// Create shared styles for all components
	sharedStyles := &shared.Styles{
		Title:          m.styles.Title,
		SelectedItem:   m.styles.SelectedItem,
		Help:           m.styles.Help,
		ActiveBorder:   m.styles.ActiveBorder,
		Todo:           m.styles.Todo,
		InProgress:     m.styles.InProgress,
		Done:           m.styles.Done,
		LowPriority:    m.styles.LowPriority,
		MediumPriority: m.styles.MediumPriority,
		HighPriority:   m.styles.HighPriority,
	}

	// Initialize collapsible sections if needed
	if m.collapsibleManager == nil {
		m.initCollapsibleSections()
	}
	
	// Set appropriate active keymap based on current context
	switch m.activePanel {
	case 0: // Task list panel
		m.activeKeyMap = keymap.TaskListKeyMap
	case 1: // Task details panel
		m.activeKeyMap = keymap.TaskDetailsKeyMap
	case 2: // Timeline panel
		m.activeKeyMap = keymap.TimelineKeyMap
	default:
		m.activeKeyMap = keymap.GlobalKeyMap
	}
	
	// Update help model with current keymap
	m.helpModel.SetKeyMap(m.activeKeyMap)
	m.helpModel.AddDelegateKeyMap(keymap.GlobalKeyMap)
	m.helpModel.SetWidth(m.width)
	
	// Render the main view first
	mainView := m.RenderMainView(sharedStyles)
	
	// If full help is toggled, show the full help view
	if m.showFullHelp {
		centeredHelp := lipgloss.Place(
			m.width, 
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			m.helpModel.FullHelpView(),
		)
		return centeredHelp
	}
	
	// If a modal is active, render it on top of the main view
	if m.showModal {
		return m.modal.View(mainView, m.width, m.height)
	}
	
	// Style the help view as a clean footer with no background
	helpFooterStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(1).
		Foreground(lipgloss.Color("#a0aec0"))
	
	// Get the help text from our help model
	helpView := m.helpModel.View()
	
	// Style and position the help footer
	styledHelpView := helpFooterStyle.Render(helpView)
	
	// Correctly position the help view at the bottom of the screen
	mainViewWithHelp := lipgloss.JoinVertical(lipgloss.Left, mainView, styledHelpView)
	
	// Return the complete view
	return mainViewWithHelp
}
