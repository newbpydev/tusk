// Package shared contains UI components that can be used across different parts of the application
package shared

import (
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/utils"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/messages"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/types"
)

// ModalMsg is sent when a modal action occurs
type ModalMsg struct {
	Type    string
	Payload interface{}
}

// ModalCloseMsg is sent when the modal is closed
type ModalCloseMsg struct{}

// Import ModalDisplayMode from types package

// ModalModel represents a modal window that can be displayed over the main UI
type ModalModel struct {
	Content      tea.Model             // Content model to display inside the modal
	Width        int                   // Width of the modal
	Height       int                   // Height of the modal
	BorderStyle  lipgloss.Style        // Border style for the modal
	ContentStyle lipgloss.Style        // Style for content inside the modal
	DimStyle     lipgloss.Style        // Style for dimming the background
	Visible      bool                  // Whether the modal is visible
	DisplayMode  types.ModalDisplayMode // How the modal should be displayed
}

// NewModal creates a new modal with default styling
func NewModal(content tea.Model, width, height int, displayMode types.ModalDisplayMode) ModalModel {
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
		Visible:     false,
		DisplayMode: displayMode,
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
	// First, handle special messages directly at the modal level
	utils.DebugLog("MODAL: Processing message of type %T", msg)
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
			// Handle ESC key press
			utils.DebugLog("MODAL: ESC key pressed, closing modal")
			m.Visible = false
			return m, func() tea.Msg { return ModalCloseMsg{} }
		}
	
	// Handle these special messages even when received from outside
	case messages.ModalFormCloseMsg:
		// Just hide and forward
		utils.DebugLog("MODAL: Received ModalFormCloseMsg, forwarding up")
		m.Visible = false
		return m, func() tea.Msg { return msg }
		
	case messages.ModalFormSubmitMsg:
		// Just hide and forward
		utils.DebugLog("MODAL: Received ModalFormSubmitMsg, forwarding up")
		m.Visible = false
		return m, func() tea.Msg { return msg }
	}

	// Only process content messages when modal is visible
	if !m.Visible {
		return m, nil
	}

	// Pass the message to the content
	var cmd tea.Cmd
	m.Content, cmd = m.Content.Update(msg)
	
	// Just return the content's command - DO NOT TRANSFORM
	// Let the messages bubble up naturally
	return m, cmd
}

// View renders the modal or returns an empty string if not visible
func (m ModalModel) View(baseView string, screenWidth, screenHeight int) string {
	if !m.Visible {
		return baseView
	}

	// Determine how to display the modal based on DisplayMode
	if m.DisplayMode == types.ContentArea {
		// For content area modals, we ONLY replace the main content area
		// while preserving the header and footer exactly as they are
		
		// The header has a FIXED structure of 4 lines:
		// 1. Empty padding line with background color
		// 2. Logo + Time + Status line (contains "TUSK")
		// 3. Tagline + Date line (contains "Task Management")
		// 4. Empty padding line with background color
		
		// Split the view into lines
		lines := strings.Split(baseView, "\n")
		
		// ALWAYS use exactly 4 lines for the header - this is the fixed structure per user requirements
		const headerHeight = 4
		
		// Detect the help footer line by looking for key binding patterns
		footerStartIdx := len(lines) - 1 // Default to last line
		for i := len(lines) - 1; i >= 0; i-- {
			// Find the help line (contains key binding help)
			if strings.Contains(lines[i], "[q]") || strings.Contains(lines[i], "[esc]") || 
			   strings.Contains(lines[i], "[?") || strings.Contains(lines[i], "[enter]") {
				footerStartIdx = i
				break
			}
		}
		
		// Ensure we have enough lines for a meaningful layout
		// We always need: 5 lines header + at least 1 line content + footer
		if len(lines) >= (headerHeight + 1 + (len(lines) - footerStartIdx)) { 
			// The header is EXACTLY the first 5 lines
			headerLines := lines[:headerHeight]
			header := strings.Join(headerLines, "\n")
			
			// The footer is from footerStartIdx to end
			footer := lines[footerStartIdx:]
			footerStr := strings.Join(footer, "\n")
			
			// Calculate content area dimensions
			footerHeight := len(lines) - footerStartIdx
			contentHeight := screenHeight - headerHeight - footerHeight
			
			// Calculate modal dimensions
			modalWidth := m.Width
			modalHeight := m.Height
			
			// Ensure modal fits within the content area bounds
			if modalWidth > screenWidth {
				modalWidth = screenWidth - 4
			}
			if modalHeight > contentHeight {
				modalHeight = contentHeight - 2 // Leave some margin
			}
			
			// Render content inside modal with proper styling
			contentStr := m.Content.View()
			contentStr = m.ContentStyle.Render(contentStr)
			
			// Apply border styling
			modalBox := m.BorderStyle.Width(modalWidth).Render(contentStr)
			
			// Create a solid black background for the content area and place the modal on it
			contentArea := lipgloss.Place(
				screenWidth,
				contentHeight,
				lipgloss.Center,
				lipgloss.Center,
				modalBox,
				lipgloss.WithWhitespaceChars(" "),
				lipgloss.WithWhitespaceForeground(lipgloss.Color("#000000")),
			)
			
			// Assemble the view with the exact fixed header structure (5 lines),
			// modal content area, and footer
			return header + "\n" + contentArea + "\n" + footerStr
		}
	}
	
	// For fullscreen modals or if we couldn't properly split the view
	// Simply render the modal over the entire view
	modalWidth := m.Width
	modalHeight := m.Height
	
	// Ensure modal fits within screen bounds
	if modalWidth > screenWidth {
		modalWidth = screenWidth - 4
	}
	if modalHeight > screenHeight {
		modalHeight = screenHeight - 4
	}
	
	// Render content inside modal with proper styling
	contentStr := m.Content.View()
	contentStr = m.ContentStyle.Render(contentStr)
	
	// Apply border styling
	modalBox := m.BorderStyle.Width(modalWidth).Render(contentStr)
	
	// Use lipgloss.Place to position the modal in the center of the screen
	overlay := lipgloss.Place(
		screenWidth,
		screenHeight,
		lipgloss.Center,
		lipgloss.Center,
		modalBox,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("#000000")),
	)
	
	// Dim the background view
	dimmedView := dimBackground(baseView, m.DimStyle)
	
	// Layer the modal over the dimmed view
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
