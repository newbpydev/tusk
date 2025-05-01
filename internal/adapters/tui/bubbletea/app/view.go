package app

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/layout"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/panels"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/shared"
	"github.com/newbpydev/tusk/internal/core/task"
	// NOTE: Need to potentially add more imports later (e.g., for max/min from utils.go).
	// Need styles, hooks as well.
)

// View implements tea.Model View, composing header, panels, and footer.
func (m *Model) View() string {
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

	// Initialize collapsible sections if needed
	if m.collapsibleManager == nil {
		// Call to initCollapsibleSections will be in sections.go
		m.initCollapsibleSections()
	}

	// Render header
	header := layout.RenderHeader(layout.HeaderProps{
		Width:         m.width,
		CurrentTime:   m.currentTime,
		StatusMessage: m.statusMessage,
		StatusType:    m.statusType,
		IsLoading:     m.isLoading,
	})

	if m.viewMode == "create" {
		// Rendering the create form might move to form.go later
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
	// Handle edit view similarly if its rendering logic is complex enough
	// if m.viewMode == "edit" { ... }

	// Default multi-panel view
	var visiblePanelCount int
	const headerHeight = 5 // These constants might become configurable or moved
	const headerGap = 0
	const footerHeight = 1
	const footerGap = 0
	const totalOffset = headerHeight + headerGap + footerHeight + footerGap
	panelHeight := m.height - totalOffset
	if m.showTaskList {
		visiblePanelCount++
	}
	if m.showTaskDetails {
		visiblePanelCount++
	}
	if m.showTimeline {
		visiblePanelCount++
	}
	availableWidth := m.width
	// Call to max will be in utils.go
	columnWidth := availableWidth / max(1, visiblePanelCount)
	var columns []string

	// Task List Panel
	if m.showTaskList {
		contentWidth := columnWidth - 2
		list := panels.RenderTaskList(panels.TaskListProps{
			Tasks:          m.tasks,
			TodoTasks:      m.todoTasks,
			ProjectTasks:   m.projectTasks,
			CompletedTasks: m.completedTasks,
			Cursor:         m.cursor,
			VisualCursor:   m.visualCursor,
			Offset:         m.taskListOffset,
			Width:          contentWidth,
			Height:         panelHeight - 2,
			Styles:         sharedStyles,
			IsActive:       m.activePanel == 0,
			Error:          m.err,
			SuccessMsg:     m.successMsg,
			ClearSuccess:   func() { m.successMsg = "" },
			CursorOnHeader: m.cursorOnHeader,
			CollapsibleMgr: m.collapsibleManager,
		})
		wrapped := shared.RenderPanel(shared.PanelProps{
			Content:     list,
			Width:       columnWidth,
			Height:      panelHeight,
			IsActive:    m.activePanel == 0,
			BorderColor: shared.ColorBorder, // Consider making colors part of styles
		})
		columns = append(columns, wrapped)
	}

	// Task Details Panel
	if m.showTaskDetails {
		contentWidth := columnWidth - 2

		// Get the appropriate task to display based on cursor position
		var selectedTask *task.Task
		if !m.cursorOnHeader && m.cursor < len(m.tasks) && m.cursor >= 0 {
			// Find which section the selected task belongs to
			// and get the correct task from the categorized lists
			taskID := m.tasks[m.cursor].ID

			// First check todoTasks
			for i, t := range m.todoTasks {
				if t.ID == taskID {
					selectedTask = &m.todoTasks[i]
					break
				}
			}

			// If not found, check projectTasks
			if selectedTask == nil {
				for i, t := range m.projectTasks {
					if t.ID == taskID {
						selectedTask = &m.projectTasks[i]
						break
					}
				}
			}

			// If still not found, check completedTasks
			if selectedTask == nil {
				for i, t := range m.completedTasks {
					if t.ID == taskID {
						selectedTask = &m.completedTasks[i]
						break
					}
				}
			}
		}

		details := panels.RenderTaskDetails(panels.TaskDetailsProps{
			Tasks:          m.tasks,
			Cursor:         m.cursor,
			SelectedTask:   selectedTask, // Pass the selected task separately
			Offset:         m.taskDetailsOffset,
			Width:          contentWidth,
			Height:         panelHeight - 2,
			Styles:         sharedStyles,
			IsActive:       m.activePanel == 1,
			CursorOnHeader: m.cursorOnHeader,
		})
		wrapped := shared.RenderPanel(shared.PanelProps{
			Content:     details,
			Width:       columnWidth,
			Height:      panelHeight,
			IsActive:    m.activePanel == 1,
			BorderColor: shared.ColorBorder,
		})
		columns = append(columns, wrapped)
	}

	// Timeline Panel
	if m.showTimeline {
		contentWidth := columnWidth - 2
		timeline := panels.RenderTimeline(panels.TimelineProps{
			Tasks:    m.tasks, // Timeline might benefit from categorized tasks (overdue, today etc.)
			Offset:   m.timelineOffset,
			Width:    contentWidth,
			Height:   panelHeight - 2,
			Styles:   sharedStyles,
			IsActive: m.activePanel == 2,
		})
		wrapped := shared.RenderPanel(shared.PanelProps{
			Content:     timeline,
			Width:       columnWidth,
			Height:      panelHeight,
			IsActive:    m.activePanel == 2,
			BorderColor: shared.ColorBorder,
		})
		columns = append(columns, wrapped)
	}

	content := lipgloss.JoinHorizontal(lipgloss.Top, columns...)

	// Render footer
	footer := layout.RenderFooter(layout.FooterProps{
		Width:          m.width,
		ViewMode:       m.viewMode,
		HelpStyle:      m.styles.Help,
		CursorOnHeader: m.cursorOnHeader, // Footer might only need a subset of state
	})

	sections := []string{header, content, footer}
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
