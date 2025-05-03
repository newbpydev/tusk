// Package shared contains UI components that can be used across different parts of the application
package shared

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ModalMsg is sent when a modal action occurs
type ModalMsg struct {
	Type    string
	Payload interface{}
}

// ModalCloseMsg is sent when the modal is closed
type ModalCloseMsg struct{}

// ModalModel represents a modal window that can be displayed over the main UI
type ModalModel struct {
	Content      tea.Model     // Content model to display inside the modal
	Width        int           // Width of the modal
	Height       int           // Height of the modal
	BorderStyle  lipgloss.Style // Border style for the modal
	ContentStyle lipgloss.Style // Style for content inside the modal
	DimStyle     lipgloss.Style // Style for dimming the background
	Visible      bool           // Whether the modal is visible
}

// NewModal creates a new modal with default styling
func NewModal(content tea.Model, width, height int) ModalModel {
	return ModalModel{
		Content: content,
		Width:   width,
		Height:  height,
		BorderStyle: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4B9CD3")).
			Padding(0, 1),
		ContentStyle: lipgloss.NewStyle().
			Padding(1, 2),
		DimStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("#000000")).
			Foreground(lipgloss.Color("#AAAAAA")).
			Italic(true),
		Visible: false,
	}
}

// Show makes the modal visible
func (m *ModalModel) Show() {
	m.Visible = true
}

// Hide hides the modal
func (m *ModalModel) Hide() {
	m.Visible = false
}

// Init initializes the modal
func (m ModalModel) Init() tea.Cmd {
	return nil
}

// Update handles events for the modal
func (m ModalModel) Update(msg tea.Msg) (ModalModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
			m.Visible = false
			return m, func() tea.Msg { return ModalCloseMsg{} }
		}
	}

	// Only update content when modal is visible
	if m.Visible {
		var cmd tea.Cmd
		m.Content, cmd = m.Content.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the modal or returns an empty string if not visible
func (m ModalModel) View(baseView string, screenWidth, screenHeight int) string {
	if !m.Visible {
		return baseView
	}

	// Calculate position to center the modal
	contentWidth := m.Width
	contentHeight := m.Height
	
	// Ensure modal fits within screen bounds
	if contentWidth > screenWidth {
		contentWidth = screenWidth - 4
	}
	if contentHeight > screenHeight {
		contentHeight = screenHeight - 4
	}

	// Render content inside modal with proper styling
	contentStr := m.Content.View()
	
	// Apply content styling
	contentStr = m.ContentStyle.Render(contentStr)
	
	// Apply border styling
	modalBox := m.BorderStyle.Width(contentWidth).Render(contentStr)
	
	// Calculate position to center modal
	modalWidth := lipgloss.Width(modalBox)
	modalHeight := strings.Count(modalBox, "\n") + 1
	
	left := (screenWidth - modalWidth) / 2
	top := (screenHeight - modalHeight) / 2
	
	// Create the dimmed background by overlaying on the base view
	lines := strings.Split(baseView, "\n")
	dimmedLines := make([]string, len(lines))
	
	for i, line := range lines {
		dimmedLines[i] = m.DimStyle.Render(line)
	}
	
	dimmedView := strings.Join(dimmedLines, "\n")
	
	// Place the modal over the dimmed background
	return placeModalOverBackground(modalBox, dimmedView, left, top)
}

// placeModalOverBackground positions the modal over the background at the specified coordinates
func placeModalOverBackground(modal, background string, left, top int) string {
	// Split both into lines
	modalLines := strings.Split(modal, "\n")
	backgroundLines := strings.Split(background, "\n")
	
	// Make sure we have enough background lines
	for len(backgroundLines) < top+len(modalLines) {
		backgroundLines = append(backgroundLines, "")
	}
	
	// Insert the modal at the specified position
	for i, modalLine := range modalLines {
		pos := top + i
		if pos >= 0 && pos < len(backgroundLines) {
			// Get background line and make sure it's wide enough
			bgLine := backgroundLines[pos]
			for len([]rune(bgLine)) < left {
				bgLine += " "
			}
			
			// Convert to runes to handle multibyte characters properly
			bgRunes := []rune(bgLine)
			modalRunes := []rune(modalLine)
			
			// Insert modal into background
			result := make([]rune, 0, len(bgRunes)+len(modalRunes))
			
			// Add background up to left position
			result = append(result, bgRunes[:minInt(left, len(bgRunes))]...)
			
			// Add modal line
			result = append(result, modalRunes...)
			
			// Add any remaining background after the modal
			if left+len(modalRunes) < len(bgRunes) {
				result = append(result, bgRunes[left+len(modalRunes):]...)
			}
			
			backgroundLines[pos] = string(result)
		}
	}
	
	return strings.Join(backgroundLines, "\n")
}

// minInt returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
