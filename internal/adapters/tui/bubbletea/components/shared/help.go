package shared

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/keymap"
)

// HelpModel represents a help bar component
type HelpModel struct {
	keyMap          *keymap.KeyMap
	help            help.Model
	width           int
	delegateKeyMaps []*keymap.KeyMap
}

// NewHelpModel creates a new help model
func NewHelpModel() HelpModel {
	h := help.New()
	h.ShowAll = false

	return HelpModel{
		keyMap: keymap.GlobalKeyMap,
		help:   h,
		width:  80,
	}
}

// SetWidth sets the width of the help bar
func (m *HelpModel) SetWidth(width int) {
	m.width = width
	m.help.Width = width
}

// SetKeyMap sets the active keymap
func (m *HelpModel) SetKeyMap(km *keymap.KeyMap) {
	m.keyMap = km
}

// AddDelegateKeyMap adds a keymap that should be included in the help display
// This is useful for when multiple keymaps are active (e.g., global + panel-specific)
func (m *HelpModel) AddDelegateKeyMap(km *keymap.KeyMap) {
	m.delegateKeyMaps = append(m.delegateKeyMaps, km)
}

// ClearDelegateKeyMaps clears all delegate keymaps
func (m *HelpModel) ClearDelegateKeyMaps() {
	m.delegateKeyMaps = nil
}

// CustomKeyMap is a helper type to satisfy the help.KeyMap interface
type CustomKeyMap struct {
	bindings []key.Binding
}

// ShortHelp returns keys for the short help view
func (c CustomKeyMap) ShortHelp() []key.Binding {
	return c.bindings
}

// FullHelp returns key sections for the full help view
func (c CustomKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{c.bindings}
}

// View returns the rendered help bar
func (m HelpModel) View() string {
	// Don't show any help if width is too narrow
	if m.width < 20 {
		return "" 
	}

	// Collect all bindings from all keymaps, prioritizing the primary keymap
	var primaryBindings, delegateBindings []key.Binding
	
	// First ensure we capture the primary (active panel) keymap bindings
	if m.keyMap != nil {
		primaryBindings = m.keyMap.ShortHelp()
	}
	
	// Then collect delegate (global) bindings
	for _, km := range m.delegateKeyMaps {
		delegateBindings = append(delegateBindings, km.ShortHelp()...)
	}
	
	// Calculate how many help items we can display
	const keyHelpAvgWidth = 12 // Average width per key binding
	maxBindings := m.width / keyHelpAvgWidth
	
	// Ensure we always show primary bindings for the current context first
	var selectedBindings []key.Binding
	primaryCount := len(primaryBindings)
	delegateCount := len(delegateBindings)
	totalCount := primaryCount + delegateCount
	
	// If we can fit all bindings, use them all
	if totalCount <= maxBindings {
		selectedBindings = append(primaryBindings, delegateBindings...)
	} else if primaryCount <= maxBindings {
		// Show all primary bindings and as many delegate bindings as we can fit
		selectedBindings = append(selectedBindings, primaryBindings...)
		remaining := maxBindings - primaryCount
		if remaining > 0 && delegateCount > 0 {
			// Add as many delegate bindings as will fit
			count := minInt(remaining, delegateCount)
			selectedBindings = append(selectedBindings, delegateBindings[:count]...)
		}
	} else {
		// Not enough space for all primary bindings, show as many as possible
		count := minInt(maxBindings, primaryCount)
		selectedBindings = primaryBindings[:count]
	}
	
	// Create a custom keymap with our selected bindings
	customKeyMap := CustomKeyMap{bindings: selectedBindings}
	
	// Configure the help view
	m.help.Width = m.width
	m.help.ShowAll = false
	
	// Get the help text
	helpText := m.help.View(customKeyMap)
	
	// Center the help text
	centeredHelpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#747474")).
		Italic(true).
		Align(lipgloss.Center).
		Width(m.width)
	
	return centeredHelpStyle.Render(helpText)
}

// minInt is a local helper that returns the minimum of two integers
// We use a different name to avoid conflicts with the panel package
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ToggleFullHelp toggles between full and simplified help
func (m *HelpModel) ToggleFullHelp() {
	m.help.ShowAll = !m.help.ShowAll
}

// FullHelpView returns a view with all available keybindings
func (m HelpModel) FullHelpView() string {
	// Collect all bindings from all keymaps
	var allBindings []key.Binding
	
	// Add primary keymap if set
	if m.keyMap != nil {
		for _, section := range m.keyMap.FullHelp() {
			allBindings = append(allBindings, section...)
		}
	}
	
	// Add delegate keymaps
	for _, km := range m.delegateKeyMaps {
		for _, section := range km.FullHelp() {
			allBindings = append(allBindings, section...)
		}
	}
	
	// Create a custom keymap that implements help.KeyMap
	customKeyMap := CustomKeyMap{bindings: allBindings}
	
	h := help.New()
	h.Width = m.width
	h.ShowAll = true
	
	helpText := h.View(customKeyMap)
	
	// Style and center the help view
	helpViewStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4B9CD3")).
		Padding(1, 2)
	
	return helpViewStyle.Render(helpText)
}
