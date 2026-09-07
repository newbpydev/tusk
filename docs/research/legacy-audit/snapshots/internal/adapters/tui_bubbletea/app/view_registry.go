// Package app contains the main TUI application implementation using bubbletea.
package app

import (
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/shared"
)

// ViewRenderer is a function type that renders a specific view's content
type ViewRenderer func(m *Model, styles *shared.Styles) string

// ViewRegistry stores all available views that can be rendered in the application
type ViewRegistry map[string]ViewRenderer

// RegisterViews initializes and returns the view registry with all available views
func RegisterViews() ViewRegistry {
	registry := ViewRegistry{
		// Main multi-panel view for task management
		"list": func(m *Model, styles *shared.Styles) string {
			return m.renderMultiPanelView(styles)
		},
		
		// Task detail view
		"detail": func(m *Model, styles *shared.Styles) string {
			return m.renderMultiPanelView(styles)
		},
		
		// Form views
		"create": func(m *Model, styles *shared.Styles) string {
			return m.renderFormView(styles)
		},
		
		"edit": func(m *Model, styles *shared.Styles) string {
			return m.renderFormView(styles)
		},
		
		// Default view - falls back to multi-panel 
		"default": func(m *Model, styles *shared.Styles) string {
			return m.renderMultiPanelView(styles)
		},
	}
	
	return registry
}

// RenderMainView renders the main application view using the appropriate renderer for the current view mode
func (m *Model) RenderMainView(styles *shared.Styles) string {
	// Initialize view registry if not done already
	if m.viewRegistry == nil {
		m.viewRegistry = RegisterViews()
	}
	
	// Get the renderer for the current view mode
	var renderer ViewRenderer
	
	if r, ok := m.viewRegistry[m.viewMode]; ok {
		renderer = r
	} else {
		// Use default renderer if the view mode isn't recognized
		renderer = m.viewRegistry["default"]
	}
	
	// Initialize collapsible sections if needed
	if m.collapsibleManager == nil {
		m.initCollapsibleSections()
	}
	
	// Use the renderer to generate the view content
	return renderer(m, styles)
}
