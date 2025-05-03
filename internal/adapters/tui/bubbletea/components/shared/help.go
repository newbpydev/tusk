package shared

import (
	"strings"

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
	// Collect all bindings from all keymaps
	var allBindings []key.Binding
	
	// First add main keymap bindings
	if m.keyMap != nil {
		// Use strings explicitly to ensure import is used
		_ = strings.Split("a,b", ",")
		allBindings = append(allBindings, m.keyMap.ShortHelp()...)
	}
	
	// Then add delegate bindings
	for _, km := range m.delegateKeyMaps {
		allBindings = append(allBindings, km.ShortHelp()...)
	}
	
	// Create a custom keymap that implements help.KeyMap
	customKeyMap := CustomKeyMap{bindings: allBindings}
	
	// Create a temporary help view and render it
	h := help.New()
	h.Width = m.width
	h.ShowAll = false
	
	// Use the custom keymap for rendering
	helpText := h.View(customKeyMap)
	
	// Style the help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#747474")).
		Italic(true)
	
	return helpStyle.Render(helpText)
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
