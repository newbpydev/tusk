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
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/core/task"
)

// Update handles user input and updates the model state.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	default:
		return m, nil
	}
}

// handleKeyPress processes keyboard input
func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.viewMode {
	case "list":
		return m.handleListViewKeys(msg)
	case "detail":
		return m.handleDetailViewKeys(msg)
	case "edit":
		return m.handleEditViewKeys(msg)
	case "create":
		return m.handleCreateFormKeys(msg)
	default:
		return m, nil
	}
}

// handleListViewKeys processes keyboard input in list view
func (m *Model) handleListViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--

			// Auto-scroll if cursor moves out of view (above viewport)
			switch m.activePanel {
			case 0: // Task list
				if m.cursor < m.taskListOffset {
					m.taskListOffset = m.cursor
				}
			case 1: // Task details - no cursor-based scrolling
			case 2: // Timeline - no cursor-based scrolling
			}
		}
		return m, nil

	case "down", "j":
		if m.cursor < len(m.tasks)-1 {
			m.cursor++

			// Auto-scroll if cursor moves out of view (below viewport)
			// Assuming ~10 visible lines in viewport after headers
			viewportHeight := 10
			switch m.activePanel {
			case 0: // Task list
				if m.cursor >= m.taskListOffset+viewportHeight {
					m.taskListOffset = m.cursor - viewportHeight + 1
				}
			case 1: // Task details - no cursor-based scrolling
			case 2: // Timeline - no cursor-based scrolling
			}
		}
		return m, nil

	case "page-up", "ctrl+b":
		// Page Up - scroll up by a page
		pageSize := 10 // Approximate lines per page

		switch m.activePanel {
		case 0: // Task list
			m.taskListOffset -= pageSize
			if m.taskListOffset < 0 {
				m.taskListOffset = 0
			}

			// Also move cursor if it would be off-screen
			if m.cursor >= m.taskListOffset+pageSize {
				m.cursor = m.taskListOffset
			}
		case 1: // Task details
			m.taskDetailsOffset -= pageSize
			if m.taskDetailsOffset < 0 {
				m.taskDetailsOffset = 0
			}
		case 2: // Timeline
			m.timelineOffset -= pageSize
			if m.timelineOffset < 0 {
				m.timelineOffset = 0
			}
		}
		return m, nil

	case "page-down", "ctrl+f":
		// Page Down - scroll down by a page
		pageSize := 10 // Approximate lines per page
		maxOffset := 0

		switch m.activePanel {
		case 0: // Task list
			// Calculate max offset based on content
			maxOffset = max(0, len(m.tasks)-pageSize)
			m.taskListOffset += pageSize
			if m.taskListOffset > maxOffset {
				m.taskListOffset = maxOffset
			}

			// Also move cursor if it would be off-screen
			if m.cursor < m.taskListOffset {
				m.cursor = m.taskListOffset
			}
		case 1: // Task details - rough estimate for max offset
			if len(m.tasks) > 0 && m.cursor < len(m.tasks) {
				// Rough estimate: 15 lines per task details
				maxOffset = 15
				if m.tasks[m.cursor].Description != nil {
					// Add more lines for longer descriptions
					maxOffset += len(*m.tasks[m.cursor].Description) / 30
				}
				maxOffset = max(0, maxOffset-pageSize)
			}
			m.taskDetailsOffset += pageSize
			if m.taskDetailsOffset > maxOffset {
				m.taskDetailsOffset = maxOffset
			}
		case 2: // Timeline - rough estimate for max offset
			// Rough estimate based on task counts
			overdue, today, upcoming := m.getTasksByTimeCategory()
			maxOffset = len(overdue) + len(today) + len(upcoming) - pageSize
			maxOffset = max(0, maxOffset)

			m.timelineOffset += pageSize
			if m.timelineOffset > maxOffset {
				m.timelineOffset = maxOffset
			}
		}
		return m, nil

	case "home", "g":
		// Scroll to top
		switch m.activePanel {
		case 0:
			m.taskListOffset = 0
			// Also move cursor to top if in task list
			m.cursor = 0
		case 1:
			m.taskDetailsOffset = 0
		case 2:
			m.timelineOffset = 0
		}
		return m, nil

	case "end", "G":
		// Scroll to bottom
		pageSize := 10 // Approximate lines per page

		switch m.activePanel {
		case 0:
			// Go to last task and make sure it's visible
			if len(m.tasks) > 0 {
				m.cursor = len(m.tasks) - 1
				m.taskListOffset = max(0, m.cursor-pageSize+1)
			}
		case 1:
			// Rough estimate for max offset in task details
			if len(m.tasks) > 0 && m.cursor < len(m.tasks) {
				maxOffset := 15 // Basic estimate
				if m.tasks[m.cursor].Description != nil {
					maxOffset += len(*m.tasks[m.cursor].Description) / 30
				}
				m.taskDetailsOffset = max(0, maxOffset-pageSize)
			}
		case 2:
			// Rough estimate for timeline
			overdue, today, upcoming := m.getTasksByTimeCategory()
			maxOffset := len(overdue) + len(today) + len(upcoming) - pageSize
			m.timelineOffset = max(0, maxOffset)
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

	case "enter":
		if len(m.tasks) > 0 {
			m.viewMode = "detail"
			return m, nil
		}

	case "c":
		// Toggle completion status
		if len(m.tasks) > 0 {
			return m, m.toggleTaskCompletion
		}

	case "r":
		// Refresh task list
		return m, m.refreshTasks

	case "s":
		// Change sort order (not implemented yet)
		return m, nil

	case "f":
		// Filter tasks (not implemented yet)
		return m, nil

	case "n":
		// Create new task
		m.viewMode = "create"
		// Initialize form with default values
		m.formStatus = string(task.StatusTodo)
		m.formPriority = string(task.PriorityLow)
		return m, nil

	case "1":
		// Toggle task list column
		m.showTaskList = !m.showTaskList
		// If hiding current active panel, move to next visible one
		if m.activePanel == 0 && !m.showTaskList {
			if m.showTaskDetails {
				m.activePanel = 1
			} else if m.showTimeline {
				m.activePanel = 2
			}
		}
		return m, nil

	case "2":
		// Toggle task details column
		m.showTaskDetails = !m.showTaskDetails
		// If hiding current active panel, move to next visible one
		if m.activePanel == 1 && !m.showTaskDetails {
			if m.showTimeline {
				m.activePanel = 2
			} else if m.showTaskList {
				m.activePanel = 0
			}
		}
		return m, nil

	case "3":
		// Toggle timeline column
		m.showTimeline = !m.showTimeline
		// If hiding current active panel, move to next visible one
		if m.activePanel == 2 && !m.showTimeline {
			if m.showTaskList {
				m.activePanel = 0
			} else if m.showTaskDetails {
				m.activePanel = 1
			}
		}
		return m, nil
	}

	return m, nil
}

// handleDetailViewKeys processes keyboard input in detail view
func (m *Model) handleDetailViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.viewMode = "list"
		return m, nil

	case "e":
		m.viewMode = "edit"
		return m, nil

	case "d":
		// Delete task
		return m, m.deleteCurrentTask

	case "c":
		// Toggle completion status
		return m, m.toggleTaskCompletion

	case "n":
		// Create new task
		m.viewMode = "create"
		// Initialize form with default values
		m.formStatus = string(task.StatusTodo)
		m.formPriority = string(task.PriorityLow)
		return m, nil

	case "up", "k":
		// Scroll up in the detail view
		switch m.activePanel {
		case 0: // Task list - move cursor
			if m.cursor > 0 {
				m.cursor--
				// Auto-scroll if necessary
				if m.cursor < m.taskListOffset {
					m.taskListOffset = m.cursor
				}
			}
		case 1: // Task details - scroll up
			m.taskDetailsOffset -= 1
			if m.taskDetailsOffset < 0 {
				m.taskDetailsOffset = 0
			}
		case 2: // Timeline - scroll up
			m.timelineOffset -= 1
			if m.timelineOffset < 0 {
				m.timelineOffset = 0
			}
		}
		return m, nil

	case "down", "j":
		// Scroll down in the detail view
		switch m.activePanel {
		case 0: // Task list - move cursor
			if m.cursor < len(m.tasks)-1 {
				m.cursor++
				// Auto-scroll if necessary
				viewportHeight := 10 // Approximate visible lines
				if m.cursor >= m.taskListOffset+viewportHeight {
					m.taskListOffset = m.cursor - viewportHeight + 1
				}
			}
		case 1: // Task details - scroll down
			// Rough estimate of content length
			maxOffset := 15 // Base value
			if len(m.tasks) > 0 && m.cursor < len(m.tasks) && m.tasks[m.cursor].Description != nil {
				maxOffset += len(*m.tasks[m.cursor].Description) / 30
			}

			m.taskDetailsOffset += 1
			if m.taskDetailsOffset > maxOffset {
				m.taskDetailsOffset = maxOffset
			}
		case 2: // Timeline - scroll down
			// Rough estimate of content length
			overdue, today, upcoming := m.getTasksByTimeCategory()
			maxOffset := len(overdue) + len(today) + len(upcoming) + 10 // +10 for headers and spacing

			m.timelineOffset += 1
			if m.timelineOffset > maxOffset {
				m.timelineOffset = maxOffset
			}
		}
		return m, nil

	case "page-up", "ctrl+b":
		// Page Up - scroll up by a page
		pageSize := 10 // Approximate lines per page

		switch m.activePanel {
		case 0: // Task list
			m.taskListOffset -= pageSize
			if m.taskListOffset < 0 {
				m.taskListOffset = 0
			}

			// Also move cursor if it would be off-screen
			if m.cursor >= m.taskListOffset+pageSize {
				m.cursor = m.taskListOffset
			}
		case 1: // Task details
			m.taskDetailsOffset -= pageSize
			if m.taskDetailsOffset < 0 {
				m.taskDetailsOffset = 0
			}
		case 2: // Timeline
			m.timelineOffset -= pageSize
			if m.timelineOffset < 0 {
				m.timelineOffset = 0
			}
		}
		return m, nil

	case "page-down", "ctrl+f":
		// Page Down - scroll down by a page
		pageSize := 10 // Approximate lines per page

		switch m.activePanel {
		case 0: // Task list
			maxOffset := max(0, len(m.tasks)-pageSize)
			m.taskListOffset += pageSize
			if m.taskListOffset > maxOffset {
				m.taskListOffset = maxOffset
			}

			// Also move cursor if it would be off-screen
			if m.cursor < m.taskListOffset {
				m.cursor = m.taskListOffset
			}
		case 1: // Task details
			// Rough estimate of max offset
			maxOffset := 15 // Base value
			if len(m.tasks) > 0 && m.cursor < len(m.tasks) && m.tasks[m.cursor].Description != nil {
				maxOffset += len(*m.tasks[m.cursor].Description) / 30
			}
			maxOffset = max(0, maxOffset-pageSize)

			m.taskDetailsOffset += pageSize
			if m.taskDetailsOffset > maxOffset {
				m.taskDetailsOffset = maxOffset
			}
		case 2: // Timeline
			// Rough estimate of max offset
			overdue, today, upcoming := m.getTasksByTimeCategory()
			maxOffset := len(overdue) + len(today) + len(upcoming) - pageSize
			maxOffset = max(0, maxOffset)

			m.timelineOffset += pageSize
			if m.timelineOffset > maxOffset {
				m.timelineOffset = maxOffset
			}
		}
		return m, nil

	case "home", "g":
		// Scroll to top
		switch m.activePanel {
		case 0:
			m.taskListOffset = 0
			m.cursor = 0
		case 1:
			m.taskDetailsOffset = 0
		case 2:
			m.timelineOffset = 0
		}
		return m, nil

	case "end", "G":
		// Scroll to bottom
		pageSize := 10 // Approximate lines per page

		switch m.activePanel {
		case 0:
			if len(m.tasks) > 0 {
				m.cursor = len(m.tasks) - 1
				m.taskListOffset = max(0, m.cursor-pageSize+1)
			}
		case 1:
			// Rough estimate for max offset in task details
			maxOffset := 15 // Base value
			if len(m.tasks) > 0 && m.cursor < len(m.tasks) && m.tasks[m.cursor].Description != nil {
				maxOffset += len(*m.tasks[m.cursor].Description) / 30
			}
			m.taskDetailsOffset = max(0, maxOffset-pageSize)
		case 2:
			// Rough estimate for timeline
			overdue, today, upcoming := m.getTasksByTimeCategory()
			maxOffset := len(overdue) + len(today) + len(upcoming) - pageSize
			m.timelineOffset = max(0, maxOffset)
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

// handleCreateFormKeys processes keyboard input in create form
func (m *Model) handleCreateFormKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle high-priority keys first that should work regardless of active field
	switch msg.Type {
	case tea.KeyEsc:
		// Always return to list view, even if there are no tasks
		m.viewMode = "list"

		// Clear form fields when canceling
		m.formTitle = ""
		m.formDescription = ""
		m.formPriority = ""
		m.formDueDate = ""
		m.formStatus = ""
		m.activeField = 0

		return m, nil

	case tea.KeyTab:
		// Tab moves to next field
		m.activeField = (m.activeField + 1) % 5
		return m, nil

	case tea.KeyShiftTab:
		// Shift+Tab moves to previous field
		m.activeField = (m.activeField - 1) % 5
		if m.activeField < 0 {
			m.activeField = 4
		}
		return m, nil

	case tea.KeyEnter:
		if m.activeField == 4 { // Submit button field
			// Validate that title field is not empty before creating
			if m.formTitle == "" {
				m.err = fmt.Errorf("title is required")
				return m, nil
			}

			// Create the task
			return m, m.createNewTask
		}

		// Move to next field
		m.activeField = (m.activeField + 1) % 5
		return m, nil
	}

	// Now handle field-specific input
	switch m.activeField {
	case 0: // Title field
		return m.handleInputField(msg, &m.formTitle)
	case 1: // Description field
		return m.handleInputField(msg, &m.formDescription)
	case 2: // Priority field (cycle through options)
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
		return m, nil
	case 3: // Due date field
		return m.handleDateField(msg, &m.formDueDate)
	}

	// Default case
	return m, nil
}

// handleInputField handles text input in a string field
func (m *Model) handleInputField(msg tea.KeyMsg, field *string) (tea.Model, tea.Cmd) {
	// Check specifically for the key type, not the string representation
	switch msg.Type {
	case tea.KeyRunes:
		// Append typed characters to the field
		*field += string(msg.Runes)
		return m, nil
	case tea.KeyBackspace, tea.KeyDelete:
		// Remove the last character if the field is not empty
		if len(*field) > 0 {
			*field = (*field)[:len(*field)-1]
		}
		return m, nil
	case tea.KeyEsc:
		// ESC key is also handled here for better responsiveness
		m.viewMode = "list"

		// Clear form fields when canceling
		m.formTitle = ""
		m.formDescription = ""
		m.formPriority = ""
		m.formDueDate = ""
		m.formStatus = ""
		m.activeField = 0

		return m, nil
	case tea.KeyEnter:
		// Move to the next field or submit if on the last field
		if m.activeField == 4 { // Submit button field
			// Validate that title field is not empty
			if m.formTitle == "" {
				m.err = fmt.Errorf("title is required")
				return m, nil
			}
			// Create the task
			return m, m.createNewTask
		}

		// Move to next field
		m.activeField = (m.activeField + 1) % 5
		return m, nil
	case tea.KeyTab:
		// Tab moves to next field
		m.activeField = (m.activeField + 1) % 5
		return m, nil
	case tea.KeyShiftTab:
		// Shift+Tab moves to previous field
		m.activeField = (m.activeField - 1) % 5
		if m.activeField < 0 {
			m.activeField = 4
		}
		return m, nil
	}
	return m, nil
}

// handleDateField handles date input with basic validation
func (m *Model) handleDateField(msg tea.KeyMsg, field *string) (tea.Model, tea.Cmd) {
	// Check specifically for the key type, not the string representation
	switch msg.Type {
	case tea.KeyRunes:
		// Only allow digits and hyphens for dates
		for _, r := range msg.Runes {
			if (r >= '0' && r <= '9') || r == '-' {
				*field += string(r)
			}
		}
		return m, nil
	case tea.KeyBackspace, tea.KeyDelete:
		// Remove the last character if the field is not empty
		if len(*field) > 0 {
			*field = (*field)[:len(*field)-1]
		}
		return m, nil
	case tea.KeyEsc:
		// ESC key is also handled here for better responsiveness
		m.viewMode = "list"

		// Clear form fields when canceling
		m.formTitle = ""
		m.formDescription = ""
		m.formPriority = ""
		m.formDueDate = ""
		m.formStatus = ""
		m.activeField = 0

		return m, nil
	case tea.KeyEnter:
		// Move to the next field or submit if on the last field
		if m.activeField == 4 { // Submit button field
			// Validate that title field is not empty
			if m.formTitle == "" {
				m.err = fmt.Errorf("title is required")
				return m, nil
			}
			// Create the task
			return m, m.createNewTask
		}

		// Move to next field
		m.activeField = (m.activeField + 1) % 5
		return m, nil
	case tea.KeyTab:
		// Tab moves to next field
		m.activeField = (m.activeField + 1) % 5
		return m, nil
	case tea.KeyShiftTab:
		// Shift+Tab moves to previous field
		m.activeField = (m.activeField - 1) % 5
		if m.activeField < 0 {
			m.activeField = 4
		}
		return m, nil
	}
	return m, nil
}

// handleEditViewKeys processes keyboard input in edit view
func (m *Model) handleEditViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = "detail"
		return m, nil

	case "enter":
		// Save edits (not fully implemented)
		m.viewMode = "detail"
		return m, nil
	}

	return m, nil
}
