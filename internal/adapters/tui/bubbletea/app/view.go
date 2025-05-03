// Package app contains the main TUI application implementation using bubbletea.  
package app

import (
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/shared"
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
	
	// Render the main view first
	mainView := m.RenderMainView(sharedStyles)
	
	// If a modal is active, render it on top of the main view
	if m.showModal {
		return m.modal.View(mainView, m.width, m.height)
	}
	
	// Otherwise just return the main view
	return mainView
}
