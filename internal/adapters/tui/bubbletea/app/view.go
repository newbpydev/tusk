// Package app contains the main TUI application implementation using bubbletea.  
package app

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/layout"
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
	
	// *** NEW DIRECT LAYOUT APPROACH ***
	// The key to preventing layout shifts is to ALWAYS build the layout 
	// in the same order, with the same fixed dimensions, regardless of state
	
	// 1. HEADER - Always exactly 4 lines, fixed height
	const headerHeight = 4
	// Use the existing RenderHeader function from the layout package
	header := layout.RenderHeader(layout.HeaderProps{
		Width:         m.width,
		CurrentTime:   m.currentTime,
		StatusMessage: m.statusMessage,
		StatusType:    m.statusType,
		IsLoading:     m.isLoading,
	})
	
	// 2. MAIN CONTENT - Calculate available space between header and footer
	const footerHeight = 1
	contentHeight := m.height - headerHeight - footerHeight
	
	// 3. Prepare the main content (which will either be normal view or modal)
	var content string
	if m.showModal {
		// For modal view, render the modal in ONLY the content area
		// The modal never touches the header or footer
		modalContent := m.modal.Content.View()
		modalContent = m.modal.ContentStyle.Render(modalContent)
		modalBox := m.modal.BorderStyle.Width(m.modal.Width).Render(modalContent)
		
		// Place the modal in center of content area only
		content = lipgloss.Place(
			m.width,
			contentHeight,
			lipgloss.Center,
			lipgloss.Center,
			modalBox,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#000000")),
		)
	} else {
		// Normal view - render main content
		mainView := m.RenderMainView(sharedStyles)
		content = lipgloss.NewStyle().
			Width(m.width).
			Height(contentHeight).
			Render(mainView)
	}
	
	// 4. FOOTER - Always exactly 1 line at the bottom
	footer := lipgloss.NewStyle().
		Width(m.width).
		Height(footerHeight).
		Render(m.helpModel.View())
	
	// 5. Combine all three parts with consistent order and dimensions
	return lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		content,
		footer,
	)
}
