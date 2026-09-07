// Package app contains the main TUI application implementation using bubbletea.  
package app

import (
	"fmt"

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
	// Create a context ID from view mode and active panel to detect context changes
	contextID := m.viewMode + "-" + fmt.Sprintf("%d", m.activePanel)
	m.helpModel.SetKeyMap(m.activeKeyMap, contextID)
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
	
	// Simple layout approach - mainView contains everything except help footer
	// No fancy calculations, just static dimensions to prevent layout shifts
	
	// Step 1: Create a content container with explicit height - this is crucial
	// We use height-1 to reserve space for the footer and prevent jumps
	contentHeight := m.height - 1 // Hard-coded 1 line for help footer
	
	// Style the main content with fixed dimensions
	mainContainer := lipgloss.NewStyle().
		Width(m.width).       // Full width
		Height(contentHeight) // Fixed height
	
	// Render the main content
	mainViewStyled := mainContainer.Render(mainView)
	
	// Step 2: Get the help footer - this new implementation won't accumulate text
	helpText := m.helpModel.View()
	
	// Step 3: Simply stack the two components vertically
	// By using fixed heights, we prevent layout shifts
	mainViewWithHelp := lipgloss.JoinVertical(
		lipgloss.Top, // Top alignment for stability
		mainViewStyled,
		helpText,
	)
	
	// Return the complete view
	return mainViewWithHelp
}
