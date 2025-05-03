// This file contains functions for managing the state of different panels
// and ensuring proper synchronization between panels
package app

import (
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/hooks"
)

// resetPanelState handles panel state synchronization when switching between panels
// This ensures that task details are properly displayed regardless of which panel is active
func (m *Model) resetPanelState(fromPanel, toPanel int) {
	// Special handling for timeline panel to task details panel
	if fromPanel == 2 && toPanel == 1 { // Timeline to task details
		// If we have a valid task selected in the timeline (not a header)
		if !m.timelineCursorOnHeader {
			taskID := m.getTimelineTaskID()
			if taskID > 0 {
				// Find the task in the main task list and update the main cursor
				for i, t := range m.tasks {
					if t.ID == taskID {
						m.cursor = i
						m.cursorOnHeader = false
						break
					}
				}
			}
		}
	} else if fromPanel == 0 && toPanel == 2 { // Task list to timeline
		// If we're coming from task list panel and had a task selected (not a header)
		if !m.cursorOnHeader && m.cursor < len(m.tasks) && m.cursor >= 0 {
			taskID := m.tasks[m.cursor].ID
			
			// Try to find and select this task in the timeline
			m.selectTaskInTimeline(taskID)
		}
	}
	
	// Reset scroll offset when switching panels
	m.taskDetailsOffset = 0
}

// selectTaskInTimeline tries to find a task with the given ID in the timeline
// and position the timeline cursor on it
func (m *Model) selectTaskInTimeline(taskID int32) {
	if taskID <= 0 || m.timelineCollapsibleMgr == nil {
		return
	}
	
	// Get the tasks categorized into timeline sections
	overdue, today, upcoming := m.getTimelineFilteredTasks()
	
	// Get section positions
	overdueHeaderIndex := m.timelineCollapsibleMgr.GetSectionHeaderIndex(hooks.SectionTypeOverdue)
	todayHeaderIndex := m.timelineCollapsibleMgr.GetSectionHeaderIndex(hooks.SectionTypeToday)
	upcomingHeaderIndex := m.timelineCollapsibleMgr.GetSectionHeaderIndex(hooks.SectionTypeUpcoming)
	
	// Check if sections are expanded
	overdueExpanded := false
	todayExpanded := false
	upcomingExpanded := false
	
	if section := m.timelineCollapsibleMgr.GetSection(hooks.SectionTypeOverdue); section != nil {
		overdueExpanded = section.IsExpanded
	}
	if section := m.timelineCollapsibleMgr.GetSection(hooks.SectionTypeToday); section != nil {
		todayExpanded = section.IsExpanded
	}
	if section := m.timelineCollapsibleMgr.GetSection(hooks.SectionTypeUpcoming); section != nil {
		upcomingExpanded = section.IsExpanded
	}
	
	// Search for the task in each section and set cursor position if found
	
	// First check overdue tasks
	if overdueExpanded {
		for i, task := range overdue {
			if task.ID == taskID {
				m.timelineCursor = overdueHeaderIndex + 1 + i // +1 to skip header
				m.timelineCursorOnHeader = false
				return
			}
		}
	}
	
	// Then check today tasks
	if todayExpanded {
		for i, task := range today {
			if task.ID == taskID {
				m.timelineCursor = todayHeaderIndex + 1 + i // +1 to skip header
				m.timelineCursorOnHeader = false
				return
			}
		}
	}
	
	// Finally check upcoming tasks
	if upcomingExpanded {
		for i, task := range upcoming {
			if task.ID == taskID {
				m.timelineCursor = upcomingHeaderIndex + 1 + i // +1 to skip header
				m.timelineCursorOnHeader = false
				return
			}
		}
	}
	
	// If task not found or section not expanded, position on appropriate section header
	// and make sure the section is expanded
	
	// Determine which section the task is in
	for _, task := range overdue {
		if task.ID == taskID {
			m.timelineCursor = overdueHeaderIndex
			m.timelineCursorOnHeader = true
			// Expand the section if it's collapsed
			if !overdueExpanded {
				m.toggleTimelineSection(hooks.SectionTypeOverdue)
			}
			return
		}
	}
	
	for _, task := range today {
		if task.ID == taskID {
			m.timelineCursor = todayHeaderIndex
			m.timelineCursorOnHeader = true
			// Expand the section if it's collapsed
			if !todayExpanded {
				m.toggleTimelineSection(hooks.SectionTypeToday)
			}
			return
		}
	}
	
	for _, task := range upcoming {
		if task.ID == taskID {
			m.timelineCursor = upcomingHeaderIndex
			m.timelineCursorOnHeader = true
			// Expand the section if it's collapsed
			if !upcomingExpanded {
				m.toggleTimelineSection(hooks.SectionTypeUpcoming)
			}
			return
		}
	}
}
