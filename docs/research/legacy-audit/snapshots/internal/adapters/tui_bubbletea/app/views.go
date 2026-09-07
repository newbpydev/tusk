// Package app contains the main app logic and view renderers.
package app

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/layout"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/panels"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/components/shared"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/hooks"
	"github.com/newbpydev/tusk/internal/core/task"
)

// renderFormView renders the create or edit form screen using the main layout
func (m *Model) renderFormView(sharedStyles *shared.Styles) string {
	// Create form props based on current form state
	formProps := panels.CreateFormProps{
		FormTitle:       m.formTitle,
		FormDescription: m.formDescription,
		FormPriority:    m.formPriority,
		FormDueDate:     m.formDueDate, // Keep for backward compatibility
		ActiveField:     m.activeField,
		Error:           m.err,
		Styles:          sharedStyles,
		// Add the interactive date input when on the date field
		ActiveDateInput: m.dateInputHandler.GetInput("dueDate"),
	}

	// Render the form content
	formContent := panels.RenderCreateForm(formProps)
	
	// Help text is now handled by the keymap system and doesn't need to be passed in props
	
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
	})
}

// renderMultiPanelView renders the main multi-panel interface with list, details, and/or timeline
func (m *Model) renderMultiPanelView(sharedStyles *shared.Styles) string {
	// Calculate layout dimensions
	var visiblePanelCount int
	const headerHeight = 5 // These constants might become configurable
	const headerGap = 0
	// Reserve 1 line for our new help footer that will be added outside the layout
	const helpFooterHeight = 1
	const totalOffset = headerHeight + headerGap + helpFooterHeight
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

	// Get the appropriate task to display based on which panel is active and its cursor position
	var selectedTask *task.Task
	var details string

	// Check if we're viewing a section header in any panel
	isHeaderSelected := false
	selectedSectionName := ""

	// Handle different panels with different logic
	switch m.activePanel {
	case 2: // Timeline panel
		if m.timelineCursorOnHeader {
			// If on a header, record that for special message
			isHeaderSelected = true

			// Determine which section header is selected
			if m.timelineCollapsibleMgr != nil {
				section := m.timelineCollapsibleMgr.GetSectionAtIndex(m.timelineCursor)
				if section != nil {
					switch section.Type {
					case hooks.SectionTypeOverdue:
						selectedSectionName = "Overdue"
					case hooks.SectionTypeToday:
						selectedSectionName = "Today"
					case hooks.SectionTypeUpcoming:
						selectedSectionName = "Upcoming"
					}
				}
			}
		} else {
			// Not on a header, get the task from the cursor
			taskID := m.getTimelineTaskID()

			// Only proceed if we have a valid task ID
			if taskID > 0 {
				// Check each of the dedicated timeline task categories
				// First check overdue tasks
				for i, t := range m.overdueTasks {
					if t.ID == taskID {
						selectedTask = &m.overdueTasks[i]
						break
					}
				}

				// Then check today tasks if not found
				if selectedTask == nil {
					for i, t := range m.todayTasks {
						if t.ID == taskID {
							selectedTask = &m.todayTasks[i]
							break
						}
					}
				}

				// Then check upcoming tasks if not found
				if selectedTask == nil {
					for i, t := range m.upcomingTasks {
						if t.ID == taskID {
							selectedTask = &m.upcomingTasks[i]
							break
						}
					}
				}

				// If still not found, try the main task list as a fallback
				if selectedTask == nil {
					for i, t := range m.tasks {
						if t.ID == taskID {
							selectedTask = &m.tasks[i]
							break
						}
					}
				}
			}
		}

	case 0: // Task list panel
		if m.cursorOnHeader {
			// We're on a section header in the task list panel
			isHeaderSelected = true

			// Get the section name
			if m.collapsibleManager != nil {
				section := m.collapsibleManager.GetSectionAtIndex(m.visualCursor)
				if section != nil {
					switch section.Type {
					case hooks.SectionTypeTodo:
						selectedSectionName = "Todo"
					case hooks.SectionTypeProjects:
						selectedSectionName = "Projects"
					case hooks.SectionTypeCompleted:
						selectedSectionName = "Completed"
					}
				}
			}
		} else if m.cursor < len(m.tasks) && m.cursor >= 0 {
			// Not on a header, get the task from the cursor
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
	}

	if isHeaderSelected {
		// Show informative message about the selected section
		
		// Render the section header message with additional context
		details = panels.RenderSectionHeaderMessage(panels.SectionHeaderMessageProps{
			SectionName: selectedSectionName, 
			Width:       contentWidth,
			Height:      height - 2,
			Styles:      styles,
			Offset:      m.taskDetailsOffset,
			IsActive:    m.activePanel == 1,
		})
	} else {
		details = panels.RenderTaskDetails(panels.TaskDetailsProps{
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
	}

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
	
	// Initialize timeline collapsible sections if needed
	if m.timelineCollapsibleMgr == nil {
		m.initTimelineCollapsibleSections()
	}
	
	timeline := panels.RenderTimeline(panels.TimelineProps{
		// Pass the categorized task slices instead of the entire task list
		OverdueTasks:    m.overdueTasks,
		TodayTasks:      m.todayTasks,
		UpcomingTasks:   m.upcomingTasks,
		Offset:          m.timelineOffset,
		Width:           contentWidth,
		Height:          height - 2,
		Styles:          styles,
		IsActive:        m.activePanel == 2,
		CollapsibleMgr:  m.timelineCollapsibleMgr,
		CursorPosition:  m.timelineCursor,
		CursorOnHeader:  m.timelineCursorOnHeader,
	})
	
	return shared.RenderPanel(shared.PanelProps{
		Content:     timeline,
		Width:       width,
		Height:      height,
		IsActive:    m.activePanel == 2,
		BorderColor: shared.ColorBorder,
	})
}
