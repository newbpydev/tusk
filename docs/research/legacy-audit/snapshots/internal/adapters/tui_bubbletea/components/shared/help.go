package shared

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/keymap"
)

// HelpModel is the most minimal possible help footer implementation
// It uses a fully static approach with a single cached render
type HelpModel struct {
	// Primary keymap
	keyMap *keymap.KeyMap
	// Global keymap
	globalKeyMap *keymap.KeyMap
	// Width for rendering
	width int
	// Cached rendered help string (not updated on tick)
	cachedHelp string
	// Last context ID (hash of active panel + view mode)
	contextKey string
	// Toggle for full help view
	showFullHelp bool
}

// NewHelpModel creates a new help model
func NewHelpModel() HelpModel {
	return HelpModel{
		keyMap:       keymap.GlobalKeyMap,
		globalKeyMap: keymap.GlobalKeyMap,
		width:        80,
		cachedHelp:   "",
		contextKey:   "",
		showFullHelp: false,
	}
}

// SetWidth sets the width of the help bar
// and forces a re-render of the help text
func (m *HelpModel) SetWidth(width int) {
	if m.width != width {
		m.width = width
		// Force regeneration of cached help
		m.cachedHelp = ""
	}
}

// SetKeyMap sets the active keymap for the current context
// and generates a new context ID for caching
func (m *HelpModel) SetKeyMap(km *keymap.KeyMap, contextID string) {
	// Only update if context actually changed
	if m.keyMap != km || m.contextKey != contextID {
		m.keyMap = km 
		m.contextKey = contextID
		// Force regeneration of cached help
		m.cachedHelp = ""
	}
}

// AddDelegateKeyMap sets the global keymap
func (m *HelpModel) AddDelegateKeyMap(km *keymap.KeyMap) {
	if m.globalKeyMap != km {
		m.globalKeyMap = km
		// Force regeneration of cached help
		m.cachedHelp = ""
	}
}

// ClearDelegateKeyMaps clears global keymap
func (m *HelpModel) ClearDelegateKeyMaps() {
	if m.globalKeyMap != nil {
		m.globalKeyMap = nil
		// Force regeneration of cached help
		m.cachedHelp = ""
	}
}

// formatKeyHelp formats a key binding as a help string
func formatKeyHelp(k key.Binding) string {
	if k.Help().Key == "" {
		return ""
	}
	
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(k.Help().Key) + 
		" " + 
		lipgloss.NewStyle().Foreground(lipgloss.Color("#aaaaaa")).Render(k.Help().Desc)
}

// View returns a STATIC help bar that won't accumulate text
// This implementation uses a cache that won't update on clock ticks
func (m HelpModel) View() string {
	// Use cached version if available
	if m.cachedHelp != "" {
		return m.cachedHelp
	}
	
	// Return empty if width is too small
	if m.width < 20 {
		m.cachedHelp = ""
		return ""
	}
	
	// CRITICAL SIMPLIFICATION: Only show 4-5 most important keys, never more
	var keys []key.Binding
	
	// First 2-3 keys from context
	if m.keyMap != nil && len(m.keyMap.ShortHelp()) > 0 {
		contextKeys := m.keyMap.ShortHelp()
		if len(contextKeys) > 3 {
			contextKeys = contextKeys[:3]
		}
		for _, k := range contextKeys {
			if k.Help().Key != "" {
				keys = append(keys, k)
			}
		}
	}

	// Always show these specific global keys
	if m.globalKeyMap != nil {
		for _, k := range m.globalKeyMap.ShortHelp() {
			// Only add essential keys like quit, help and tab
			if k.Help().Key == "q" || k.Help().Key == "ctrl+c" || 
			   k.Help().Key == "?" || k.Help().Key == "tab" {
				keys = append(keys, k)
			}
		}
	}
	
	// Generate static help text with no dynamic calculations
	helpText := ""
	
	// Format each key with a simple separator
	for i, k := range keys {
		if i > 0 {
			helpText += "  " // Static two spaces between items
		}
		helpText += formatKeyHelp(k)
	}
	
	// IMPORTANT: Use a fixed style with no dynamic calculations
	style := lipgloss.NewStyle().
		Width(m.width).       // Fixed width based on screen
		Align(lipgloss.Center). // Always center
		Foreground(lipgloss.Color("#747474")).
		Italic(true)
	
	// Cache the result to prevent re-rendering on clock ticks
	m.cachedHelp = style.Render(helpText)
	
	return m.cachedHelp 
}



// ToggleFullHelp toggles between full and simplified help
func (m *HelpModel) ToggleFullHelp() {
	m.showFullHelp = !m.showFullHelp
	// Force regeneration of cached help
	m.cachedHelp = ""
}

// FullHelpView returns a simple help screen with all commands
func (m HelpModel) FullHelpView() string {
	// Build the most basic possible help view
	helpContent := "\n Keyboard Shortcuts\n\n"
	
	// Add context-specific keys
	if m.keyMap != nil {
		helpContent += " Context Commands:\n\n"
		
		// Get all relevant context keys
		for _, section := range m.keyMap.FullHelp() {
			for _, k := range section {
				if k.Help().Key != "" {
					helpContent += "   " + 
						lipgloss.NewStyle().Bold(true).Render(k.Help().Key) + 
						" - " + 
						k.Help().Desc + "\n"
				}
			}
		}
		
		helpContent += "\n"
	}
	
	// Add global keys
	if m.globalKeyMap != nil {
		helpContent += " Global Commands:\n\n"
		
		// Get all global keys
		for _, section := range m.globalKeyMap.FullHelp() {
			for _, k := range section {
				if k.Help().Key != "" {
					helpContent += "   " + 
						lipgloss.NewStyle().Bold(true).Render(k.Help().Key) + 
						" - " + 
						k.Help().Desc + "\n"
				}
			}
		}
	}
	
	// Apply minimal styling
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4B9CD3")).
		Padding(1, 2).
		Render(helpContent)
}
