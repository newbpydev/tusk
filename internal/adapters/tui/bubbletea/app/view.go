// Package app contains the main TUI application implementation using bubbletea.  
package app

import (
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/shared"
)

// View implements tea.Model View, displaying the appropriate view based on the current mode.
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
	
	// Render the appropriate view based on the current view mode

	switch m.viewMode {
	case "create", "edit":
		// Both create and edit use the same form rendering
		return m.renderFormView(sharedStyles)
	
	default:
		// The default view is the multi-panel view for list mode and others

		return m.renderMultiPanelView(sharedStyles)
	}
}
