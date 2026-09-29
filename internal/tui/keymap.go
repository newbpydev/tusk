package tui

import "github.com/charmbracelet/bubbles/key"

type panelFocus uint8

const (
	listFocus panelFocus = iota
	detailsFocus
)

func browseHints() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("Tab", "focus")),
		key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	}
}

func helpLines() []string {
	return []string{
		"Move around", "Tab / Shift+Tab   Switch list and details",
		"↑ ↓ / j k         Move or scroll", "g / G, Home / End First / last",
		"PgUp / PgDn       Move one page", "← → / h l         Collapse / expand",
		"", "Work with tasks", "a  Create    e  Edit    d  Delete",
		"Space / x         Complete / reopen", "/  Search    f  Filter    r  Refresh",
		"", "Forms", "Tab / Shift+Tab   Next / previous field",
		"Ctrl+S            Save", "Enter in notes    New line",
		"Esc               Cancel / clear filters", "Ctrl+C            Exit safely",
	}
}
