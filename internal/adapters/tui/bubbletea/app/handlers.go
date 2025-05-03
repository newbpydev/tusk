package app

import (
	"fmt"
	"math"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/hooks"
	"github.com/newbpydev/tusk/internal/core/task"
)

// handleKeyPress delegates keyboard input based on current view mode and active panel
func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global key handlers that work in any mode
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	}

	switch m.viewMode {
	case "list":
		// In list view, delegate to panel-specific handlers based on active panel
		switch m.activePanel {
		case 0: // Task list panel
			return m.handleTaskListPanelKeys(msg)
		case 1: // Task details panel
			return m.handleTaskDetailsPanelKeys(msg)
		case 2: // Timeline panel
			return m.handleTimelinePanelKeys(msg)
		default:
			return m, nil
		}
	case "detail":
		return m.handleDetailViewKeys(msg)
	case "create", "edit":
		// Both create and edit use the same form handling,
		// with behavior differences handled inside the form functions
		return m.handleFormKeys(msg)
	default:
		return m, nil
	}
}

// handleTaskListPanelKeys processes keyboard input when the task list panel is active
func (m *Model) handleTaskListPanelKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Initialize sections if needed
	if m.collapsibleManager == nil {
		m.initCollapsibleSections()
	}

	switch msg.String() {
	case "j", "down":
		// Handle down navigation through tasks and section headers
		if m.collapsibleManager.GetItemCount() > 0 {
			m.navigateDown()
			return m, nil
		}
		return m, nil

	case "k", "up":
		// Handle up navigation through tasks and section headers
		if m.collapsibleManager.GetItemCount() > 0 {
			m.navigateUp()
			return m, nil
		}
		return m, nil

	case "g":
		// Jump to top
		m.navigateToTop()
		return m, nil

	case "G":
		// Jump to bottom
		m.navigateToBottom()
		return m, nil

	case "tab", "right", "l":
		// Move to next panel if available
		prevPanel := m.activePanel 
		if m.showTaskDetails {
			m.activePanel = 1
		} else if m.showTimeline {
			m.activePanel = 2
		}
		
		// If we changed panels, reset certain state to ensure proper task selection
		if prevPanel != m.activePanel {
			m.resetPanelState(prevPanel, m.activePanel)
		}
		return m, nil

	case "enter", "d":
		// If on a section header, toggle expansion
		if m.cursorOnHeader {
			return m, m.toggleSection()
		}
		// If on a task, show details (if available)
		if m.showTaskDetails && !m.cursorOnHeader {
			m.activePanel = 1
		}
		return m, nil

	case " ":
		// Toggle task completion status
		if !m.cursorOnHeader && m.cursor < len(m.tasks) {
			return m, m.toggleTaskCompletion()
		}
		return m, nil

	case "n":
		// Create new task
		m.resetForm()
		m.viewMode = "create"
		m.formPriority = string(task.PriorityLow) // Set default priority
		return m, nil

	case "e":
		// Edit task
		if !m.cursorOnHeader && m.cursor < len(m.tasks) {
			m.viewMode = "edit"
			// Load current task into form
			m.loadTaskIntoForm(m.tasks[m.cursor])
			return m, nil
		}
		return m, nil

	case "r":
		// Refresh tasks
		m.setLoadingStatus("Refreshing tasks...")

		// Debug date comparison functions with a manufactured test
		now := time.Now()
		testDate1 := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)  // Today at midnight
		testDate2 := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, time.UTC) // Today at noon

		sameDay := isSameDay(testDate1, testDate2)
		beforeDay := isBeforeDay(testDate1, testDate2)
		afterDay := isAfterDay(testDate1, testDate2)

		m.setStatusMessage(
			fmt.Sprintf("Date test: same day=%v, before=%v, after=%v",
				sameDay, beforeDay, afterDay),
			"info", 2*time.Second)

		return m, m.refreshTasks()
	}

	// Handle panel visibility toggles
	return m.handlePanelVisibilityKeys(msg)
}

// handleTaskDetailsPanelKeys processes keyboard input when the task details panel is active
func (m *Model) handleTaskDetailsPanelKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab", "right", "l":
		// Move to next panel if available
		if m.showTimeline {
			m.activePanel = 2
		}
		return m, nil

	case "shift+tab", "left", "h", "esc":
		// Move to previous panel
		if m.showTaskList {
			m.activePanel = 0
		}
		return m, nil

	case "j", "down":
		// Scroll down in task details
		if m.taskDetailsOffset < 100 { // Arbitrary limit that could be calculated
			m.taskDetailsOffset++
		}
		return m, nil

	case "k", "up":
		// Scroll up in task details
		if m.taskDetailsOffset > 0 {
			m.taskDetailsOffset--
		}
		return m, nil

	case "e":
		// Edit current task
		if !m.cursorOnHeader && m.cursor < len(m.tasks) {
			m.viewMode = "edit"
			// Load current task into form
			m.loadTaskIntoForm(m.tasks[m.cursor])
			return m, nil
		}
		return m, nil

	case "r":
		// Refresh tasks
		m.setLoadingStatus("Refreshing tasks...")
		return m, m.refreshTasks()
	}

	// Handle panel visibility toggles
	return m.handlePanelVisibilityKeys(msg)
}

// handleTimelinePanelKeys processes keyboard input when the timeline panel is active
func (m *Model) handleTimelinePanelKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Make sure the timeline collapsible manager is initialized
	if m.timelineCollapsibleMgr == nil {
		m.initTimelineCollapsibleSections()
	}

	// Initialize the timeline cursor state if needed
	if m.timelineCursor == 0 && m.timelineCollapsibleMgr.GetItemCount() > 0 {
		// Start with cursor on the first section header
		m.timelineCursor = 0
		m.timelineCursorOnHeader = m.timelineCollapsibleMgr.IsSectionHeader(0)
	}

	switch msg.String() {
	case "shift+tab", "left", "h":
		// Move to previous panel if available
		prevPanel := m.activePanel
		// When in timeline panel (panel 2), first check if task details panel is available
		if m.activePanel == 2 && m.showTaskDetails {
			// Go to task details panel
			m.activePanel = 1
		} else if m.showTaskList {
			// Otherwise go to task list panel if it's visible
			m.activePanel = 0
		} else {
			// If no panels are available to switch to, just do nothing
		}
		
		// If we changed panels, reset certain state to ensure proper task selection
		if prevPanel != m.activePanel {
			m.resetPanelState(prevPanel, m.activePanel)
		}
		return m, nil

	case "j", "down":
		// Navigate down through the timeline sections and items
		if m.timelineCollapsibleMgr.GetItemCount() > 0 {
			// Store the previous cursor state to check if selection changed
			prevCursor := m.timelineCursor
			prevOnHeader := m.timelineCursorOnHeader

			// Navigate down in the timeline sections
			m.timelineCursor = m.timelineCollapsibleMgr.GetNextCursorPosition(m.timelineCursor, 1)
			m.timelineCursorOnHeader = m.timelineCollapsibleMgr.IsSectionHeader(m.timelineCursor)

			// If selection changed and showing task details, reset the details scroll offset
			if (m.timelineCursor != prevCursor || m.timelineCursorOnHeader != prevOnHeader) && m.showTaskDetails {
				m.taskDetailsOffset = 0
			}

			// Calculate visible height in terms of logical items, not raw lines
			// Typical item height is ~2 lines, so divide available height by 2 for better estimation
			visibleHeight := (m.height - 10) / 2 // Height in terms of logical items
			
			// Calculate position of cursor relative to visible region and adjust if needed
			// The key is to account for varying item heights (headers vs tasks with/without descriptions)
			
			// For downward navigation, we need to keep items fully visible
			// The key adjustments are:
			// 1. Start scrolling with just a 1-item offset to be more responsive
			// 2. Account for scroll indicators (▲ and ▼) that take up extra space
			// 3. Consider the height of the current item to ensure it's fully visible
			
			// First, determine the exact height of the current item
			cursorItemHeight := 1 // Default for section headers
			
			if !m.timelineCursorOnHeader {
				// Tasks have variable heights depending on presence of description
				taskID := m.getTimelineTaskID()
				if taskID > 0 {
					// Find the actual task to check its description
					hasDescription := false
					for _, tasks := range [][]task.Task{m.overdueTasks, m.todayTasks, m.upcomingTasks} {
						for _, t := range tasks {
							if t.ID == taskID && t.Description != nil && *t.Description != "" {
								hasDescription = true
								break
							}
						}
					}
					
					if hasDescription {
						cursorItemHeight = 3 // Title + description + separator
					} else {
						cursorItemHeight = 2 // Title + separator
					}
				}
			}
			
			// Calculate actual visible height accounting for borders, padding and scroll indicators
			// Subtract 2 for possible scroll indicators (▲ and ▼)
			actualHeight := m.height - 12 // 10 for borders/padding + 2 for scroll indicators
			
			// Convert to logical items (divide by 2 as average item height)
			visibleHeight = actualHeight / 2
			
			// Use a smaller offset increment for more responsive scrolling
			// This means the cursor will follow more closely with each keypress
			offsetIncrement := 1
			
			// ROBUST CURSOR TRACKING: Always ensure cursor stays in viewport
			// Calculate exact cursor position relative to viewport
			relativePos := m.timelineCursor - m.timelineOffset
			
			// CRITICAL: Always adjust offset to keep cursor visible when scrolling down
			// Maintain cursor in the middle half of the viewport when possible
			if relativePos > (visibleHeight * 2/3) {
				// For smoother scrolling, move in single-line increments
				m.timelineOffset += offsetIncrement
				
				// Special handling for taller items (with descriptions)
				if cursorItemHeight >= 3 {
					// Ensure we can see the full item including description
					// The key is to make sure we scroll enough to see the entire item
					visibleBottom := m.timelineOffset + visibleHeight
					itemBottom := m.timelineCursor + cursorItemHeight - 1
					
					// If any part of the item is beyond visible area, scroll more
					if itemBottom > visibleBottom {
						needExtraScroll := itemBottom - visibleBottom
						m.timelineOffset += needExtraScroll
					}
				}
			}
			
			// Safety guard: never scroll past the last item
			// Use explicit assignment since maxOffset might be already defined elsewhere
			maxOffsetValue := max(0, m.timelineCollapsibleMgr.GetItemCount() - visibleHeight)
			m.timelineOffset = min(m.timelineOffset, maxOffsetValue)
			
			// Bounds check to prevent scrolling past the end
			maxItems := m.timelineCollapsibleMgr.GetItemCount()
			// Leave room for all remaining items
			maxOffset := max(0, maxItems - visibleHeight)
			m.timelineOffset = min(m.timelineOffset, maxOffset)
		} else {
			// Fall back to just scrolling if no collapsible sections
			const maxTimelineScroll = 500
			if m.timelineOffset < maxTimelineScroll {
				m.timelineOffset++
			}
		}
		return m, nil

	case "k", "up":
		// Navigate up through the timeline sections and items
		if m.timelineCollapsibleMgr.GetItemCount() > 0 {
			if m.timelineCursor > 0 {
				// Store the previous cursor state to check if selection changed
				prevCursor := m.timelineCursor
				prevOnHeader := m.timelineCursorOnHeader

				m.timelineCursor = m.timelineCollapsibleMgr.GetNextCursorPosition(m.timelineCursor, -1)
				m.timelineCursorOnHeader = m.timelineCollapsibleMgr.IsSectionHeader(m.timelineCursor)

				// If selection changed and showing task details, reset the details scroll offset
				if (m.timelineCursor != prevCursor || m.timelineCursorOnHeader != prevOnHeader) && m.showTaskDetails {
					m.taskDetailsOffset = 0
				}

				// ROBUST CURSOR TRACKING FOR UPWARD NAVIGATION
				// This matches the downward approach for consistency and complete visibility
				
				// First, determine the exact height of the current item
				cursorItemHeight := 1 // Default for section headers
				
				if !m.timelineCursorOnHeader {
					// Tasks have variable heights depending on description
					taskID := m.getTimelineTaskID()
					if taskID > 0 {
						hasDescription := false
						for _, tasks := range [][]task.Task{m.overdueTasks, m.todayTasks, m.upcomingTasks} {
							for _, t := range tasks {
								if t.ID == taskID && t.Description != nil && *t.Description != "" {
									hasDescription = true
									break
								}
							}
						}
						
						if hasDescription {
							cursorItemHeight = 3 // Title + description + separator
						} else {
							cursorItemHeight = 2 // Title + separator
						}
					}
				}
				
				// Calculate actual visible height accounting for borders, padding and scroll indicators
				// Subtract 2 for possible scroll indicators (▲ and ▼)
				actualHeight := m.height - 12 // 10 for borders/padding + 2 for scroll indicators
				
				// Convert to logical items (divide by approximate average item height)
				visibleHeight := actualHeight / 2
				
				// Calculate cursor position relative to viewport
				relativePos := m.timelineCursor - m.timelineOffset
				
				// CRITICAL: Always ensure cursor stays in view when navigating up
				// The top of the viewport requires special attention
				if relativePos < visibleHeight/3 {
					// Calculate how much to scroll to keep cursor in optimal position
					// Use smaller increments for smoother scrolling
					m.timelineOffset = max(0, m.timelineOffset - 1)
					
					// Special case: ensure full visibility for taller items (with descriptions)
					if cursorItemHeight >= 3 {
						// Check if the entire item is visible at the top of viewport
						itemTop := m.timelineCursor
						itemBottom := itemTop + cursorItemHeight - 1
						visibleTop := m.timelineOffset
						visibleBottom := visibleTop + visibleHeight - 1
						
						// If any part of the item would be cut off, scroll to show it completely
						if itemTop < visibleTop || itemBottom > visibleBottom {
							// Position item toward the top with some context
							m.timelineOffset = max(0, itemTop - 1)
						}
					}
				}
				
				// Safety guard: never scroll before the first item
				m.timelineOffset = max(0, m.timelineOffset)
			}
		} else {
			// Fall back to just scrolling if no collapsible sections
			if m.timelineOffset > 0 {
				m.timelineOffset--
			}
		}
		return m, nil

	case "g":
		// Jump to top of timeline
		m.timelineOffset = 0
		m.timelineCursor = 0
		m.timelineCursorOnHeader = m.timelineCollapsibleMgr.IsSectionHeader(0)

		// Reset task details offset if task details panel is visible
		if m.showTaskDetails {
			m.taskDetailsOffset = 0
		}

		return m, nil

	case "G":
		// Jump to bottom of timeline
		if m.timelineCollapsibleMgr.GetItemCount() > 0 {
			lastIndex := m.timelineCollapsibleMgr.GetItemCount() - 1
			m.timelineCursor = lastIndex
			m.timelineCursorOnHeader = m.timelineCollapsibleMgr.IsSectionHeader(lastIndex)

			// Ensure the cursor is visible
			visibleHeight := m.height - 8
			m.timelineOffset = int(math.Max(0, float64(lastIndex-visibleHeight)))
		} else {
			// Fall back to approximate scrolling
			m.timelineOffset = 500 // Large value that should be near the bottom
		}

		// Reset task details offset if task details panel is visible
		if m.showTaskDetails {
			m.taskDetailsOffset = 0
		}

		return m, nil

	case "enter", "space":
		// If on a section header, toggle expansion
		if m.timelineCursorOnHeader {
			section := m.timelineCollapsibleMgr.GetSectionAtIndex(m.timelineCursor)
			if section != nil {
				return m, m.toggleTimelineSection(section.Type)
			}
		} else {
			// If on a task, toggle its completion status
			return m, m.toggleTimelineTaskCompletion()
		}
		return m, nil

	case "tab", "right", "l":
		// Show task details if a task is selected
		if !m.timelineCursorOnHeader {
			// Find the task by timeline index
			taskIndex := m.getTimelineTaskIndex()
			if taskIndex >= 0 {
				m.cursor = taskIndex // Set the main cursor to this task
				m.activePanel = 1    // Switch to task details panel
				return m, nil
			}
		}
		return m, nil

	case "c":
		// Toggle task completion status when 'c' is pressed (similar to Space)
		if !m.timelineCursorOnHeader {
			return m, m.toggleTimelineTaskCompletion()
		}
		return m, nil

	case "r":
		// Refresh tasks
		m.setLoadingStatus("Refreshing tasks...")
		return m, m.refreshTasks()
	}

	// Handle panel visibility toggles
	return m.handlePanelVisibilityKeys(msg)
}

// handlePanelVisibilityKeys processes keyboard input for panel visibility toggles
func (m *Model) handlePanelVisibilityKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "1":
		// Toggle task list visibility
		m.showTaskList = !m.showTaskList
		if !m.showTaskList && m.activePanel == 0 {
			if m.showTaskDetails {
				m.activePanel = 1
			} else if m.showTimeline {
				m.activePanel = 2
			}
		}
		return m, nil

	case "2":
		// Toggle task details visibility
		m.showTaskDetails = !m.showTaskDetails
		if !m.showTaskDetails && m.activePanel == 1 {
			if m.showTaskList {
				m.activePanel = 0
			} else if m.showTimeline {
				m.activePanel = 2
			}
		}
		return m, nil

	case "3":
		// Toggle timeline visibility
		m.showTimeline = !m.showTimeline
		if !m.showTimeline && m.activePanel == 2 {
			if m.showTaskDetails {
				m.activePanel = 1
			} else if m.showTaskList {
				m.activePanel = 0
			}
		}
		return m, nil
	}

	return m, nil
}

// handleDetailViewKeys processes keyboard input in detail view (legacy function, kept for compatibility)
func (m *Model) handleDetailViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "esc", "tab", "shift+tab", "h", "left":
		// Return to task list panel
		m.activePanel = 0
		return m, nil

	case "j", "down":
		// Scroll down in task details
		if m.taskDetailsOffset < 100 { // Arbitrary limit that could be calculated
			m.taskDetailsOffset++
		}
		return m, nil

	case "k", "up":
		// Scroll up in task details
		if m.taskDetailsOffset > 0 {
			m.taskDetailsOffset--
		}
		return m, nil

	case "e":
		// Edit current task
		if !m.cursorOnHeader && m.cursor < len(m.tasks) {
			m.viewMode = "edit"
			// Load current task into form
			m.loadTaskIntoForm(m.tasks[m.cursor])
			return m, nil
		}
		return m, nil

	case "r":
		// Refresh tasks
		m.setLoadingStatus("Refreshing tasks...")
		return m, m.refreshTasks()
	}

	return m, nil
}

// navigateDown moves the cursor down, handling section expansion/collapse
func (m *Model) navigateDown() {
	if m.collapsibleManager.GetItemCount() == 0 {
		return
	}

	// Check if there's another item below
	if m.visualCursor < m.collapsibleManager.GetItemCount()-1 {
		m.visualCursor = m.collapsibleManager.GetNextCursorPosition(m.visualCursor, 1)

		// Check if we're on a section header
		m.cursorOnHeader = m.collapsibleManager.IsSectionHeader(m.visualCursor)

		// If it's a task, update the task cursor
		if !m.cursorOnHeader {
			taskIndex := m.collapsibleManager.GetActualTaskIndex(m.visualCursor)
			if taskIndex >= 0 {
				m.cursor = taskIndex
			}
		}
	}
}

// navigateUp moves the cursor up, handling section expansion/collapse
func (m *Model) navigateUp() {
	if m.visualCursor > 0 {
		m.visualCursor = m.collapsibleManager.GetNextCursorPosition(m.visualCursor, -1)

		// Check if we're on a section header
		m.cursorOnHeader = m.collapsibleManager.IsSectionHeader(m.visualCursor)

		// If it's a task, update the task cursor
		if !m.cursorOnHeader {
			taskIndex := m.collapsibleManager.GetActualTaskIndex(m.visualCursor)
			if taskIndex >= 0 {
				m.cursor = taskIndex
			}
		}
	}
}

// navigateToTop moves cursor to top of the list
func (m *Model) navigateToTop() {
	m.visualCursor = 0
	// Check if we're on a section header
	m.cursorOnHeader = m.collapsibleManager.IsSectionHeader(0)

	// If not on a header, get the task index
	if !m.cursorOnHeader {
		taskIndex := m.collapsibleManager.GetActualTaskIndex(0)
		if taskIndex >= 0 {
			m.cursor = taskIndex
		}
	}
}

// navigateToBottom moves cursor to bottom of the list
func (m *Model) navigateToBottom() {
	lastIndex := m.collapsibleManager.GetItemCount() - 1
	if lastIndex >= 0 {
		m.visualCursor = lastIndex
		// Check if we're on a section header
		m.cursorOnHeader = m.collapsibleManager.IsSectionHeader(lastIndex)

		// If not on a header, get the task index
		if !m.cursorOnHeader {
			taskIndex := m.collapsibleManager.GetActualTaskIndex(lastIndex)
			if taskIndex >= 0 {
				m.cursor = taskIndex
			}
		}
	}
}

// toggleSection expands or collapses the section at the current cursor position
func (m *Model) toggleSection() tea.Cmd {
	if m.cursorOnHeader {
		// Get the section at the current visual cursor position
		section := m.collapsibleManager.GetSectionAtIndex(m.visualCursor)
		if section != nil {
			// Get a user-friendly name for the section being toggled
			var sectionName string
			switch section.Type {
			case hooks.SectionTypeTodo:
				sectionName = "Todo"
			case hooks.SectionTypeProjects:
				sectionName = "Projects"
			case hooks.SectionTypeCompleted:
				sectionName = "Completed"
			}

			// Use the section name for status updates
			if section.IsExpanded {
				m.setStatusMessage("Collapsing "+sectionName+" section", "info", 1*time.Second)
			} else {
				m.setStatusMessage("Expanding "+sectionName+" section", "info", 1*time.Second)
			}

			// Toggle the section
			m.collapsibleManager.ToggleSection(section.Type)
		}
	}
	return nil
}
