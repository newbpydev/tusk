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
	
	// Completely different approach to prevent text accumulation
	// Define key dimensions
	const helpFooterHeight = 1
	contentHeight := m.height - helpFooterHeight
	
	// Step 1: Create the main content area with strict boundaries
	contentStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(contentHeight)
	
	// Render main content with strict size control
	mainViewStyled := contentStyle.Render(mainView)
	
	// Step 2: Create help footer with clear boundaries
	// Get the help text - our HelpModel already handles width constraints
	helpText := m.helpModel.View()
	
	// Style help footer with fixed dimensions
	helpStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(helpFooterHeight).
		Foreground(lipgloss.Color("#a0aec0")).
		Italic(true)
	
	helpStyled := helpStyle.Render(helpText)
	
	// Step 3: Use direct vertical joining with top alignment to prevent shifts
	mainViewWithHelp := lipgloss.JoinVertical(lipgloss.Top, mainViewStyled, helpStyled)
	
	// Return the complete view
	return mainViewWithHelp
}
