package app

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/messages"
	"github.com/newbpydev/tusk/internal/core/task"
)

// Update implements tea.Model Update, handling all message types.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Delegate all keyboard handling to the handler functions in handlers.go
		newModel, cmd := m.handleKeyPress(msg)
		return newModel, cmd

	case tea.WindowSizeMsg:
		// Update window dimensions for layout calculations
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case messages.TickMsg:
		// Update current time and check for status message expiry
		m.currentTime = time.Time(msg)
		if (!m.statusExpiry.IsZero()) && time.Now().After(m.statusExpiry) {
			m.statusMessage = ""
			m.statusType = ""
			m.statusExpiry = time.Time{}
		}
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return messages.TickMsg(t)
		})

	case messages.StatusUpdateErrorMsg:
		// Handle task update error
		m.err = msg.Err
		m.setErrorStatus(fmt.Sprintf("Error updating task '%s': %v", msg.TaskTitle, msg.Err))
		return m, m.refreshTasks()

	case messages.StatusUpdateSuccessMsg:
		// Handle successful task update
		m.setSuccessStatus(msg.Message)

		// Keep track of the updated task ID
		updatedTaskID := msg.Task.ID

		// Just update the task data in the main list without changing cursor position
		for i := range m.tasks {
			if m.tasks[i].ID == updatedTaskID {
				// Update the task with server data
				m.tasks[i] = msg.Task
				break
			}
		}

		// To ensure consistency, preserve the current cursor positions
		originalCursor := m.cursor
		originalVisualCursor := m.visualCursor
		originalCursorOnHeader := m.cursorOnHeader

		// Re-categorize tasks with updated data
		m.categorizeTasks(m.tasks)

		// Restore cursor positions
		m.cursor = originalCursor
		m.visualCursor = originalVisualCursor
		m.cursorOnHeader = originalCursorOnHeader

		// Refresh the visual cursor from task cursor to ensure consistency
		// This is important for cases where the task moves between sections
		m.updateVisualCursorFromTaskCursor()

		return m, nil

	case messages.TasksRefreshedMsg:
		// Handle refreshed task list
		m.tasks = msg.Tasks
		if m.cursor >= len(m.tasks) {
			m.cursor = max(0, len(m.tasks)-1)
		}
		m.clearLoadingStatus()
		m.initCollapsibleSections()
		return m, nil

	case messages.ErrorMsg:
		// Handle general error
		m.err = error(msg)
		m.setErrorStatus(fmt.Sprintf("Error: %v", error(msg)))
		return m, nil

	default:
		return m, nil
	}
}







// NOTICE: This was moved to handlers.go. 
// This function is kept here temporarily for legacy compatibility.
// All new code should call the version in handlers.go.
// TODO: Remove this function when the refactoring is complete.
func handleLegacyListViewKeys(m *Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Initialize sections if needed
	if m.collapsibleManager == nil {
		// Call to initCollapsibleSections will be moved to sections.go
		m.initCollapsibleSections()
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "up", "k":
		switch m.activePanel {
		case 0: // Task list panel
			if m.collapsibleManager != nil {
				prevVisual := m.visualCursor

				// Move up in the visual list (includes section headers)
				m.visualCursor = m.collapsibleManager.GetNextCursorPosition(m.visualCursor, -1)

				// Auto-scroll if cursor moves out of view
				if m.visualCursor < m.taskListOffset {
					m.taskListOffset = m.visualCursor
				}

				// Update the task cursor based on the new visual position
				// Call to updateTaskCursorFromVisualCursor will be moved to sections.go
				m.updateTaskCursorFromVisualCursor()

				// Reset detail & timeline scroll offsets if selection changed
				if prevVisual != m.visualCursor {
					m.taskDetailsOffset = 0
					m.timelineOffset = 0
				}
			} else {
				// Legacy behavior without collapsible sections
				if m.cursor > 0 {
					prevCursor := m.cursor
					m.cursor--
					if m.cursor < m.taskListOffset {
						m.taskListOffset = m.cursor
					}
					if prevCursor != m.cursor {
						m.taskDetailsOffset = 0
						m.timelineOffset = 0
					}
				}
			}
		case 1: // Task Details panel - instant scroll
			viewportHeight := 10
			// Call to max will be moved to utils.go
			m.taskDetailsOffset = max(0, m.taskDetailsOffset-viewportHeight)
		case 2: // Timeline panel - instant scroll
			viewportHeight := 10
			// Call to max will be moved to utils.go
			m.timelineOffset = max(0, m.timelineOffset-viewportHeight)
		}
		return m, nil

	case "down", "j":
		switch m.activePanel {
		case 0: // Task list panel
			if m.collapsibleManager != nil {
				prevVisual := m.visualCursor

				// Move down in the visual list (includes section headers)
				m.visualCursor = m.collapsibleManager.GetNextCursorPosition(m.visualCursor, 1)

				// Auto-scroll if cursor moves out of view
				viewportHeight := 10 // Approximate visible lines
				if m.visualCursor >= m.taskListOffset+viewportHeight {
					m.taskListOffset = m.visualCursor - viewportHeight + 1
				}

				// Update the task cursor based on the new visual position
				// Call to updateTaskCursorFromVisualCursor will be moved to sections.go
				m.updateTaskCursorFromVisualCursor()

				// Reset detail & timeline scroll offsets if selection changed
				if prevVisual != m.visualCursor {
					m.taskDetailsOffset = 0
					m.timelineOffset = 0
				}
			} else {
				// Legacy behavior without collapsible sections
				if m.cursor < len(m.tasks)-1 {
					prevCursor := m.cursor
					m.cursor++
					viewportHeight := 10
					if m.cursor >= m.taskListOffset+viewportHeight {
						m.taskListOffset = m.cursor - viewportHeight + 1
					}
					if prevCursor != m.cursor {
						m.taskDetailsOffset = 0
						m.timelineOffset = 0
					}
				}
			}
		case 1: // Task Details panel - instant scroll
			viewportHeight := 10
			maxOffset := 15
			if len(m.tasks) > 0 && m.cursor < len(m.tasks) && m.tasks[m.cursor].Description != nil {
				maxOffset += len(*m.tasks[m.cursor].Description) / 30
			}
			// Call to min will be moved to utils.go
			m.taskDetailsOffset = min(m.taskDetailsOffset+viewportHeight, maxOffset)
		case 2: // Timeline panel - instant scroll
			viewportHeight := 10
			// Call to getTasksByTimeCategory will be moved to tasks.go
			overdue, today, upcoming := m.getTasksByTimeCategory()
			maxOffset := len(overdue) + len(today) + len(upcoming) + 15
			// Call to min will be moved to utils.go
			m.timelineOffset = min(m.timelineOffset+viewportHeight, maxOffset)
		}
		return m, nil

	case "page-up", "ctrl+b":
		pageSize := 20 // Larger page size for details and timeline
		switch m.activePanel {
		case 0:
			if m.collapsibleManager != nil {
				// Move up by a page in the visual list
				prevVisual := m.visualCursor

				// Call to max will be moved to utils.go
				m.visualCursor = max(0, m.visualCursor-pageSize)
				m.taskListOffset = max(0, m.taskListOffset-pageSize)

				// Update the task cursor based on the new visual position
				// Call to updateTaskCursorFromVisualCursor will be moved to sections.go
				m.updateTaskCursorFromVisualCursor()

				// Reset detail & timeline scroll offsets if selection changed
				if prevVisual != m.visualCursor {
					m.taskDetailsOffset = 0
					m.timelineOffset = 0
				}
			} else {
				// Legacy behavior
				prev := m.cursor
				m.taskListOffset -= pageSize
				if m.taskListOffset < 0 {
					m.taskListOffset = 0
				}
				if m.cursor >= m.taskListOffset+pageSize {
					m.cursor = m.taskListOffset
				}
				if prev != m.cursor {
					m.taskDetailsOffset = 0
					m.timelineOffset = 0
				}
			}
		case 1:
			// Call to max will be moved to utils.go
			m.taskDetailsOffset = max(0, m.taskDetailsOffset-pageSize)
		case 2:
			// Call to max will be moved to utils.go
			m.timelineOffset = max(0, m.timelineOffset-pageSize)
		}
		return m, nil

	case "page-down", "ctrl+f":
		pageSize := 20 // Larger page size for details and timeline
		var maxOffset int
		switch m.activePanel {
		case 0:
			if m.collapsibleManager != nil {
				// Move down by a page in the visual list
				prevVisual := m.visualCursor
				totalItems := m.collapsibleManager.GetItemCount()

				// Calculate maximum offset and cursor positions
				// Call to max will be moved to utils.go
				maxOffset = max(0, totalItems-pageSize)
				m.taskListOffset += pageSize
				if m.taskListOffset > maxOffset {
					m.taskListOffset = maxOffset
				}

				// Call to min will be moved to utils.go
				m.visualCursor = min(totalItems-1, m.visualCursor+pageSize)

				// Update the task cursor based on the new visual position
				// Call to updateTaskCursorFromVisualCursor will be moved to sections.go
				m.updateTaskCursorFromVisualCursor()

				// Reset detail & timeline scroll offsets if selection changed
				if prevVisual != m.visualCursor {
					m.taskDetailsOffset = 0
					m.timelineOffset = 0
				}
			} else {
				// Legacy behavior
				// Call to max will be moved to utils.go
				maxOffset = max(0, len(m.tasks)-pageSize)
				m.taskListOffset += pageSize
				if m.taskListOffset > maxOffset {
					m.taskListOffset = maxOffset
				}
				if m.cursor < m.taskListOffset {
					m.cursor = m.taskListOffset
				}
			}
		case 1:
			maxOffset = 15
			if len(m.tasks) > 0 && m.cursor < len(m.tasks) && m.tasks[m.cursor].Description != nil {
				maxOffset += len(*m.tasks[m.cursor].Description) / 30
			}
			// Call to min will be moved to utils.go
			m.taskDetailsOffset = min(m.taskDetailsOffset+pageSize, maxOffset)
		case 2:
			// Call to getTasksByTimeCategory will be moved to tasks.go
			overdue, today, upcoming := m.getTasksByTimeCategory()
			maxOffset = len(overdue) + len(today) + len(upcoming) + 15
			// Call to min will be moved to utils.go
			m.timelineOffset = min(m.timelineOffset+pageSize, maxOffset)
		}
		return m, nil

	case "home", "g":
		switch m.activePanel {
		case 0:
			if m.collapsibleManager != nil {
				// Go to the first item (should be the first section header)
				prevVisual := m.visualCursor
				m.visualCursor = 0
				m.taskListOffset = 0

				// Update the task cursor
				// Call to updateTaskCursorFromVisualCursor will be moved to sections.go
				m.updateTaskCursorFromVisualCursor()

				// Reset scroll positions if cursor changed
				if prevVisual != m.visualCursor {
					m.taskDetailsOffset = 0
					m.timelineOffset = 0
				}
			} else {
				// Legacy behavior
				m.taskListOffset = 0
				m.cursor = 0
			}
		case 1:
			m.taskDetailsOffset = 0
		case 2:
			m.timelineOffset = 0
		}
		return m, nil

	case "end", "G":
		pageSize := 10
		switch m.activePanel {
		case 0:
			if m.collapsibleManager != nil {
				// Go to the last item in the visual list
				prevVisual := m.visualCursor
				totalItems := m.collapsibleManager.GetItemCount()

				if totalItems > 0 {
					m.visualCursor = totalItems - 1
					// Call to max will be moved to utils.go
					m.taskListOffset = max(0, m.visualCursor-pageSize+1)

					// Update the task cursor
					// Call to updateTaskCursorFromVisualCursor will be moved to sections.go
					m.updateTaskCursorFromVisualCursor()

					// Reset scroll positions if cursor changed
					if prevVisual != m.visualCursor {
						m.taskDetailsOffset = 0
						m.timelineOffset = 0
					}
				}
			} else {
				// Legacy behavior
				if len(m.tasks) > 0 {
					m.cursor = len(m.tasks) - 1
					// Call to max will be moved to utils.go
					m.taskListOffset = max(0, m.cursor-pageSize+1)
				}
			}
		case 1:
			maxOff := 15
			if len(m.tasks) > 0 && m.cursor < len(m.tasks) && m.tasks[m.cursor].Description != nil {
				maxOff += len(*m.tasks[m.cursor].Description) / 30
			}
			// Call to max will be moved to utils.go
			m.taskDetailsOffset = max(0, maxOff-pageSize)
		case 2:
			// Call to getTasksByTimeCategory will be moved to tasks.go
			overdue, today, upcoming := m.getTasksByTimeCategory()
			maxOff := len(overdue) + len(today) + len(upcoming) + 15
			// Call to max will be moved to utils.go
			m.timelineOffset = max(0, maxOff-pageSize)
		}
		return m, nil

	case "enter", " ":
		// Only respond when task list panel is active
		if m.activePanel == 0 {
			// Check if we're on a section header
			if m.cursorOnHeader && m.collapsibleManager != nil {
				// Toggle the section's expanded state
				section := m.collapsibleManager.GetSectionAtIndex(m.visualCursor)
				if section != nil {
					m.collapsibleManager.ToggleSection(section.Type)
					// Don't reset cursor state after toggling
					return m, nil
				}
				return m, nil
			} else if !m.cursorOnHeader && len(m.tasks) > 0 && m.cursor < len(m.tasks) {
				// We're on a task (not a header), go to detail view
				m.viewMode = "detail"
				return m, nil
			}
		}
		return m, nil

	case "c":
		// Only toggle completion status when on a task (not section header)
		if m.activePanel == 0 && !m.cursorOnHeader && len(m.tasks) > 0 && m.cursor < len(m.tasks) {
			// Call to toggleTaskCompletion will be moved to tasks.go
			return m, m.toggleTaskCompletion()
		}
		return m, nil

	// Keep the rest of the key bindings unchanged
	case "r":
		// Call to refreshTasks will be moved to tasks.go
		return m, m.refreshTasks()

	case "n":
		m.viewMode = "create"
		m.formStatus = string(task.StatusTodo)
		m.formPriority = string(task.PriorityLow)
		return m, nil

	case "1":
		m.showTaskList = !m.showTaskList
		if m.activePanel == 0 && !m.showTaskList {
			if m.showTaskDetails {
				m.activePanel = 1
			} else if m.showTimeline {
				m.activePanel = 2
			}
		}
		return m, nil

	case "2":
		m.showTaskDetails = !m.showTaskDetails
		if m.activePanel == 1 && !m.showTaskDetails {
			if m.showTimeline {
				m.activePanel = 2
			} else if m.showTaskList {
				m.activePanel = 0
			}
		}
		return m, nil

	case "3":
		m.showTimeline = !m.showTimeline
		if m.activePanel == 2 && !m.showTimeline {
			if m.showTaskList {
				m.activePanel = 0
			} else if m.showTaskDetails {
				m.activePanel = 1
			}
		}
		return m, nil

	case "right", "l":
		// Move focus to the next visible panel
		visiblePanels := []bool{m.showTaskList, m.showTaskDetails, m.showTimeline}
		originalPanel := m.activePanel

		// Find next visible panel
		for i := 0; i < 3; i++ {
			m.activePanel = (m.activePanel + 1) % 3
			if visiblePanels[m.activePanel] {
				break
			}
		}

		// If no other panels are visible, keep the original panel
		if !visiblePanels[m.activePanel] {
			m.activePanel = originalPanel
		}
		return m, nil

	case "left", "h":
		// Move focus to the previous visible panel
		visiblePanels := []bool{m.showTaskList, m.showTaskDetails, m.showTimeline}
		originalPanel := m.activePanel

		// Find previous visible panel
		for i := 0; i < 3; i++ {
			m.activePanel = (m.activePanel + 2) % 3 // +2 is equivalent to -1 in modulo 3
			if visiblePanels[m.activePanel] {
				break
			}
		}

		// If no other panels are visible, keep the original panel
		if !visiblePanels[m.activePanel] {
			m.activePanel = originalPanel
		}
		return m, nil
	}
	return m, nil
}

// NOTICE: This was moved to handlers.go. 
// This function is kept here temporarily for legacy compatibility.
// All new code should call the version in handlers.go.
// TODO: Remove this function when the refactoring is complete.
func handleLegacyDetailViewKeys(m *Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.viewMode = "list"
		return m, nil

	case "e":
		if m.activePanel == 0 {
			m.viewMode = "edit"
			// Pre-fill form fields? Need to decide where this logic lives.
			// Maybe form.go should handle entering edit mode.
		}
		return m, nil

	case "d":
		if m.activePanel == 0 {
			// Call to deleteCurrentTask will be moved to tasks.go
			return m, m.deleteCurrentTask()
		}
		return m, nil

	case "c":
		if m.activePanel == 0 {
			// Call to toggleTaskCompletion will be moved to tasks.go
			return m, m.toggleTaskCompletion()
		}
		return m, nil

	case "n":
		m.viewMode = "create"
		m.formStatus = string(task.StatusTodo)
		m.formPriority = string(task.PriorityLow)
		return m, nil

	// Reuse list scrolling logic
	case "up", "k", "down", "j", "page-up", "ctrl+b", "page-down", "ctrl+f", "home", "g", "end", "G", "right", "l", "left", "h":
		return m.handleListViewKeys(msg)
	}
	return m, nil
}

// NOTICE: This functionality was moved to form.go. 
// This function is kept here temporarily for legacy compatibility.
// All new code should call handleFormKeys in form.go.
// TODO: Remove this function when the refactoring is complete.
func handleLegacyCreateFormKeys(m *Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.viewMode = "list"
		m.formTitle = ""
		m.formDescription = ""
		m.formPriority = ""
		m.formDueDate = ""
		m.formStatus = ""
		m.activeField = 0
		return m, nil

	case tea.KeyTab:
		m.activeField = (m.activeField + 1) % 5
		return m, nil

	case tea.KeyShiftTab:
		m.activeField = (m.activeField - 1) % 5
		if m.activeField < 0 {
			m.activeField = 4
		}
		return m, nil

	case tea.KeyEnter:
		if m.activeField == 4 { // Assuming 4 is the submit button/action
			if m.formTitle == "" {
				m.err = fmt.Errorf("title is required")
				// Call to setErrorStatus will be in status.go
				m.setErrorStatus("Title is required")
				return m, nil
			}
			// Call to createNewTask will be in tasks.go
			return m, m.createNewTask()
		}
		m.activeField = (m.activeField + 1) % 5
		return m, nil
	}

	switch m.activeField {
	case 0: // Title
		// Call to handleInputField will be in form.go
		return m.handleInputField(msg, &m.formTitle)
	case 1: // Description
		// Call to handleInputField will be in form.go
		return m.handleInputField(msg, &m.formDescription)
	case 2: // Priority
		if msg.String() == " " {
			switch m.formPriority {
			case string(task.PriorityLow):
				m.formPriority = string(task.PriorityMedium)
			case string(task.PriorityMedium):
				m.formPriority = string(task.PriorityHigh)
			default:
				m.formPriority = string(task.PriorityLow)
			}
		}
		// Allow Tab/ShiftTab/Esc/Enter to pass through for navigation
		switch msg.Type {
		case tea.KeyTab, tea.KeyShiftTab, tea.KeyEsc, tea.KeyEnter:
			// Let the outer switch handle these
		}
		return m, nil // Consume other keys
	case 3: // Due Date
		// Call to handleDateField will be in form.go
		return m.handleDateField(msg, &m.formDueDate)
		// case 4: // Submit button - handled by KeyEnter above
	}
	return m, nil
}

// NOTICE: This functionality was moved to form.go. 
// This function is kept here temporarily for legacy compatibility.
// All new code should call handleFormKeys in form.go.
// TODO: Remove this function when the refactoring is complete.
func handleLegacyEditViewKeys(m *Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// This function needs refinement. It should likely load task data into form fields
	// upon entering edit mode, and then handle updates similar to create form,
	// finally calling an update task command instead of create.
	switch msg.String() {
	case "esc":
		m.viewMode = "list" // Exit edit mode without saving
		// Clear form fields?
		return m, nil
	case "tab":
		m.activeField = (m.activeField + 1) % 5
		return m, nil
	case "shift+tab":
		m.activeField = (m.activeField - 1) % 5
		if m.activeField < 0 {
			m.activeField = 4
		}
		return m, nil
	case "enter":
		if m.activeField == 4 { // Assuming 4 is the submit action
			if m.formTitle == "" {
				m.err = fmt.Errorf("title is required")
				// Call to setErrorStatus will be in status.go
				m.setErrorStatus("Title is required")
				return m, nil
			}
			// TODO: Implement actual update task logic
			// return m, m.updateCurrentTask()
			m.viewMode = "list" // For now, just return to list
			return m, nil
		}
		m.activeField = (m.activeField + 1) % 5
		return m, nil
	}

	// Handle input for fields based on activeField
	switch m.activeField {
	case 0: // Title
		// Call to handleInputField will be in form.go
		return m.handleInputField(msg, &m.formTitle)
	case 1: // Description
		// Call to handleInputField will be in form.go
		return m.handleInputField(msg, &m.formDescription)
	case 2: // Priority (similar to create form)
		if msg.String() == " " {
			switch m.formPriority {
			case string(task.PriorityLow):
				m.formPriority = string(task.PriorityMedium)
			case string(task.PriorityMedium):
				m.formPriority = string(task.PriorityHigh)
			default:
				m.formPriority = string(task.PriorityLow)
			}
		}
		// Allow Tab/ShiftTab/Esc/Enter to pass through for navigation
		switch msg.Type {
		case tea.KeyTab, tea.KeyShiftTab, tea.KeyEsc, tea.KeyEnter:
			// Let the outer switch handle these
		}
		return m, nil // Consume other keys
	case 3: // Due Date
		// Call to handleDateField will be in form.go
		return m.handleDateField(msg, &m.formDueDate)
		// case 4: // Submit button - handled by KeyEnter above
	}
	return m, nil
}
