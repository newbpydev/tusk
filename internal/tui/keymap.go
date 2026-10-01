package tui

import "github.com/charmbracelet/bubbles/key"

type panelFocus uint8

const (
	listFocus panelFocus = iota
	detailsFocus
)

func browseHints() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "new")),
		key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
		key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
		key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "filter")),
		key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("Tab", "focus")),
	}
}

func helpLines() []string {
	return []string{
		"Move around", "Tab / Shift+Tab   Switch list and details",
		"↑ ↓ / j k         Move or scroll", "g / G, Home / End First / last",
		"PgUp / PgDn       Move one page", "← → / h l         Collapse / expand",
		"", "Work with tasks", "a  Create    e  Edit    d  Delete",
		"Space / x         Complete subtree / reopen task", "/  Search    f  Filter    r  Refresh",
		"1 / 2 / 3         All tasks / Today / Done",
		"", "Forms", "Tab / Shift+Tab   Next / previous field",
		"Ctrl+S            Save", "Enter in notes    New line",
		"Ctrl+P            Calendar / choose parent", "Ctrl+U            Clear due, tags or parent",
		"Ctrl+E            Replace read-only text", "Ctrl+R            Refresh / reload conflicted draft",
		"Esc               Cancel / clear filters", "Ctrl+C            Exit safely",
	}
}
