// Copyright (C) 2025 Juan Antonio Gomez Pena
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package bubbletea

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/layout"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/panels"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/shared"
)

// View renders the current state of the model as a string.
func (m *Model) View() string {
	// Convert bubbletea.Styles to shared.Styles
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

	// Render the header using our header component
	header := layout.RenderHeader(layout.HeaderProps{
		Width:         m.width,
		CurrentTime:   m.currentTime,
		StatusMessage: m.statusMessage,
		StatusType:    m.statusType,
		IsLoading:     m.isLoading,
	})

	// For the create form view, just show header and form
	if m.viewMode == "create" {
		createForm := panels.RenderCreateForm(panels.CreateFormProps{
			FormTitle:       m.formTitle,
			FormDescription: m.formDescription,
			FormPriority:    m.formPriority,
			FormDueDate:     m.formDueDate,
			ActiveField:     m.activeField,
			Error:           m.err,
			Styles:          sharedStyles,
		})

		return lipgloss.JoinVertical(lipgloss.Left, header, createForm)
	}

	// Prepare for multi-panel layout
	var visiblePanelCount int

	// Use minimal spacing between components to maximize content area
	const headerHeight = 5 // Actual header component height (reduced from 3)
	const headerGap = 0    // Minimal gap between header and content
	const footerHeight = 1 // Footer height
	const footerGap = 0    // Minimal gap between content and footer

	// Total height reduction for content area
	const totalHeightOffset = headerHeight + headerGap + footerHeight + footerGap

	// Ensure consistent panel height across all panels
	panelHeight := m.height - totalHeightOffset

	// Count visible panels
	if m.showTaskList {
		visiblePanelCount++
	}
	if m.showTaskDetails {
		visiblePanelCount++
	}
	if m.showTimeline {
		visiblePanelCount++
	}

	// Calculate column width - account for visible panels
	availableWidth := m.width
	columnWidth := availableWidth / max(1, visiblePanelCount)

	// Prepare content for each panel
	var columns []string

	// Task List Panel
	if m.showTaskList {
		// Calculate the internal content width (accounting for borders only)
		contentWidth := columnWidth - 2 // Subtract borders only (no padding)

		taskListContent := panels.RenderTaskList(panels.TaskListProps{
			Tasks:        m.tasks,
			Cursor:       m.cursor,
			Offset:       m.taskListOffset,
			Width:        contentWidth,
			Height:       panelHeight - 2, // Account for borders
			Styles:       sharedStyles,
			IsActive:     m.activePanel == 0,
			Error:        m.err,
			SuccessMsg:   m.successMsg,
			ClearSuccess: func() { m.successMsg = "" },
		})

		// Wrap with panel styling - ensure full height borders
		wrappedPanel := shared.RenderPanel(shared.PanelProps{
			Content:     taskListContent,
			Width:       columnWidth,
			Height:      panelHeight,
			IsActive:    m.activePanel == 0,
			BorderColor: shared.ColorBorder,
		})

		columns = append(columns, wrappedPanel)
	}

	// Task Details Panel
	if m.showTaskDetails {
		// Calculate the internal content width (accounting for borders only)
		contentWidth := columnWidth - 2 // Subtract borders only (no padding)

		taskDetailsContent := panels.RenderTaskDetails(panels.TaskDetailsProps{
			Tasks:    m.tasks,
			Cursor:   m.cursor,
			Offset:   m.taskDetailsOffset,
			Width:    contentWidth,
			Height:   panelHeight - 2, // Account for borders
			Styles:   sharedStyles,
			IsActive: m.activePanel == 1,
		})

		// Wrap with panel styling - ensure full height borders
		wrappedPanel := shared.RenderPanel(shared.PanelProps{
			Content:     taskDetailsContent,
			Width:       columnWidth,
			Height:      panelHeight,
			IsActive:    m.activePanel == 1,
			BorderColor: shared.ColorBorder,
		})

		columns = append(columns, wrappedPanel)
	}

	// Timeline Panel
	if m.showTimeline {
		// Calculate the internal content width (accounting for borders only)
		contentWidth := columnWidth - 2 // Subtract borders only (no padding)

		timelineContent := panels.RenderTimeline(panels.TimelineProps{
			Tasks:    m.tasks,
			Offset:   m.timelineOffset,
			Width:    contentWidth,
			Height:   panelHeight - 2, // Account for borders
			Styles:   sharedStyles,
			IsActive: m.activePanel == 2,
		})

		// Wrap with panel styling - ensure full height borders
		wrappedPanel := shared.RenderPanel(shared.PanelProps{
			Content:     timelineContent,
			Width:       columnWidth,
			Height:      panelHeight,
			IsActive:    m.activePanel == 2,
			BorderColor: shared.ColorBorder,
		})

		columns = append(columns, wrappedPanel)
	}

	// Join panels horizontally with consistent spacing
	content := lipgloss.JoinHorizontal(lipgloss.Top, columns...)

	// Add help footer fixed at the bottom with proper styling
	footer := layout.RenderFooter(layout.FooterProps{
		Width:     m.width,
		ViewMode:  m.viewMode,
		HelpStyle: m.styles.Help,
	})

	// Build the final layout with minimal spacing to maximize content area
	var sections []string
	sections = append(sections, header)
	if headerGap > 0 {
		sections = append(sections, strings.Repeat(" ", m.width)) // Optional header gap
	}
	sections = append(sections, content)
	if footerGap > 0 {
		sections = append(sections, strings.Repeat(" ", m.width)) // Optional footer gap
	}
	sections = append(sections, footer)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the larger of two integers
func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}
