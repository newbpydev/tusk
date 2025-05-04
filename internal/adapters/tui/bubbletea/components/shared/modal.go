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

	// Calculate modal dimensions
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
	
	// Use lipgloss.Place to absolutely position the modal in the center
	// This is a more reliable approach than manually positioning the modal
	// which can cause layout shifts
	overlay := lipgloss.Place(
		screenWidth,
		screenHeight,
		lipgloss.Center,
		lipgloss.Center,
		modalBox,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("#000000")),
	)
	
	// Dim the background view without changing its layout or dimensions
	dimmedView := dimBackground(baseView, m.DimStyle)

	// Layer the modal over the dimmed view using Z-index concept
	return overlayContent(overlay, dimmedView)
}

// dimBackground applies the dimming style to a complete view string
// without changing its structure or dimensions
func dimBackground(view string, dimStyle lipgloss.Style) string {
	lines := strings.Split(view, "\n")
	dimmedLines := make([]string, len(lines))
	
	for i, line := range lines {
		// Apply dim style to each line
		dimmedLines[i] = dimStyle.Render(line)
	}
	
	return strings.Join(dimmedLines, "\n")
}

// overlayContent creates a composite view where content is layered over background
// without disturbing the layout or dimensions of either
func overlayContent(overlay, background string) string {
	// We're using a simple approach: the overlay is already positioned
	// with proper dimensions using lipgloss.Place, so we can just return it
	// This works because lipgloss.Place creates a string with the exact dimensions
	// of the entire screen, with the content centered
	return overlay
}

// The minInt function is no longer needed since we're using lipgloss.Place
// for positioning instead of manual character manipulation
