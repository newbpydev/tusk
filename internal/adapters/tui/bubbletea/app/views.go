// Package app contains the main app logic and view renderers.
package app

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/layout"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/panels"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/shared"
	"github.com/newbpydev/tusk/internal/core/task"
)

// renderFormView renders the create or edit form screen using the main layout
func (m *Model) renderFormView(sharedStyles *shared.Styles) string {
	// Create form props based on current form state
	formProps := panels.CreateFormProps{
		FormTitle:       m.formTitle,
		FormDescription: m.formDescription,
		FormPriority:    m.formPriority,
		FormDueDate:     m.formDueDate,
		ActiveField:     m.activeField,
		Error:           m.err,
		Styles:          sharedStyles,
	}

	// Render the form content
	formContent := panels.RenderCreateForm(formProps)
	
	// Define custom help text for the form
	var helpText string
	if m.viewMode == "create" {
		helpText = "tab: next field • enter: create task • esc: cancel • space: cycle priority"
	} else { // edit mode
		helpText = "tab: next field • enter: update task • esc: cancel • space: cycle priority"
	}
	
	// Use the main layout for consistent UI
	return layout.RenderMainLayout(layout.MainLayoutProps{
		// Header properties
		Width:         m.width,
		Height:        m.height,
		CurrentTime:   m.currentTime,
		StatusMessage: m.statusMessage,
		StatusType:    m.statusType,
		IsLoading:     m.isLoading,
		
		// Main content
		Content:       formContent,
		
		// Footer properties
		ViewMode:      m.viewMode,
		HelpStyle:     m.styles.Help,
		HelpText:      helpText,
	})
}

// renderMultiPanelView renders the main multi-panel interface with list, details, and/or timeline
func (m *Model) renderMultiPanelView(sharedStyles *shared.Styles) string {
	// Calculate layout dimensions
	var visiblePanelCount int
	const headerHeight = 5 // These constants might become configurable
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
	columnWidth := availableWidth / max(1, visiblePanelCount)
	
	var columns []string

	// Task List Panel
	if m.showTaskList {
		columns = append(columns, m.renderTaskListPanel(sharedStyles, columnWidth, panelHeight))
	}

	// Task Details Panel
	if m.showTaskDetails {
		columns = append(columns, m.renderTaskDetailsPanel(sharedStyles, columnWidth, panelHeight))
	}

	// Timeline Panel
	if m.showTimeline {
		columns = append(columns, m.renderTimelinePanel(sharedStyles, columnWidth, panelHeight))
	}

	// Join panels horizontally
	panelsContent := lipgloss.JoinHorizontal(lipgloss.Top, columns...)
	
	// Use the main layout for consistent UI
	return layout.RenderMainLayout(layout.MainLayoutProps{
		// Header properties
		Width:          m.width,
		Height:         m.height,
		CurrentTime:    m.currentTime,
		StatusMessage:  m.statusMessage,
		StatusType:     m.statusType,
		IsLoading:      m.isLoading,
		
		// Main content is the combined panels
		Content:        panelsContent,
		
		// Footer properties
		ViewMode:       m.viewMode,
		HelpStyle:      m.styles.Help,
		CursorOnHeader: m.cursorOnHeader,
		ActivePanel:    m.activePanel,
	})
}

// renderTaskListPanel renders the task list panel
func (m *Model) renderTaskListPanel(styles *shared.Styles, width, height int) string {
	contentWidth := width - 2
	
	list := panels.RenderTaskList(panels.TaskListProps{
		Tasks:          m.tasks,
		TodoTasks:      m.todoTasks,
		ProjectTasks:   m.projectTasks,
		CompletedTasks: m.completedTasks,
		Cursor:         m.cursor,
		VisualCursor:   m.visualCursor,
		Offset:         m.taskListOffset,
		Width:          contentWidth,
		Height:         height - 2,
		Styles:         styles,
		IsActive:       m.activePanel == 0,
		Error:          m.err,
		SuccessMsg:     m.successMsg,
		ClearSuccess:   func() { m.successMsg = "" },
		CursorOnHeader: m.cursorOnHeader,
		CollapsibleMgr: m.collapsibleManager,
	})
	
	return shared.RenderPanel(shared.PanelProps{
		Content:     list,
		Width:       width,
		Height:      height,
		IsActive:    m.activePanel == 0,
		BorderColor: shared.ColorBorder,
	})
}

// renderTaskDetailsPanel renders the task details panel
func (m *Model) renderTaskDetailsPanel(styles *shared.Styles, width, height int) string {
	contentWidth := width - 2

	// Get the appropriate task to display based on cursor position
	var selectedTask *task.Task
	if !m.cursorOnHeader && m.cursor < len(m.tasks) && m.cursor >= 0 {
		// Find which section the selected task belongs to and get the correct task
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
		SelectedTask:   selectedTask,
		Offset:         m.taskDetailsOffset,
		Width:          contentWidth,
		Height:         height - 2,
		Styles:         styles,
		IsActive:       m.activePanel == 1,
		CursorOnHeader: m.cursorOnHeader,
	})
	
	return shared.RenderPanel(shared.PanelProps{
		Content:     details,
		Width:       width,
		Height:      height,
		IsActive:    m.activePanel == 1,
		BorderColor: shared.ColorBorder,
	})
}

// renderTimelinePanel renders the timeline panel
func (m *Model) renderTimelinePanel(styles *shared.Styles, width, height int) string {
	contentWidth := width - 2
	
	timeline := panels.RenderTimeline(panels.TimelineProps{
		Tasks:    m.tasks,
		Offset:   m.timelineOffset,
		Width:    contentWidth,
		Height:   height - 2,
		Styles:   styles,
		IsActive: m.activePanel == 2,
	})
	
	return shared.RenderPanel(shared.PanelProps{
		Content:     timeline,
		Width:       width,
		Height:      height,
		IsActive:    m.activePanel == 2,
		BorderColor: shared.ColorBorder,
	})
}
