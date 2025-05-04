package app

import (
	"sort"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/hooks"
	"github.com/newbpydev/tusk/internal/core/task"
)

// initTimelineCollapsibleSections initializes the timeline section state
func (m *Model) initTimelineCollapsibleSections() {
	// Create a collapsible manager if needed
	if m.timelineCollapsibleMgr == nil {
		m.timelineCollapsibleMgr = hooks.NewCollapsibleManager()
	}

	// Clear any existing sections
	m.timelineCollapsibleMgr.ClearSections()

	// Categorize tasks for the timeline view and store in the model's dedicated slices
	m.overdueTasks, m.todayTasks, m.upcomingTasks = m.categorizeTimelineTasks(m.tasks)

	// Add sections to the timeline collapsible manager
	// Start indices are computed based on section sizes to ensure proper cursor handling
	m.timelineCollapsibleMgr.AddSection(hooks.SectionTypeOverdue, "Overdue", len(m.overdueTasks), 0)
	m.timelineCollapsibleMgr.AddSection(hooks.SectionTypeToday, "Today", len(m.todayTasks), len(m.overdueTasks))
	m.timelineCollapsibleMgr.AddSection(hooks.SectionTypeUpcoming, "Upcoming", len(m.upcomingTasks), len(m.overdueTasks)+len(m.todayTasks))

	// Make all sections expanded by default (better UX)
	// First call toggleSection for each section to ensure they're all expanded
	for _, sectionType := range []hooks.SectionType{hooks.SectionTypeOverdue, hooks.SectionTypeToday, hooks.SectionTypeUpcoming} {
		// Find if the section is already expanded
		expanded := false
		for _, section := range m.timelineCollapsibleMgr.Sections {
			if section.Type == sectionType {
				expanded = section.IsExpanded
				break
			}
		}

		// If not expanded, toggle it to expand
		if !expanded {
			m.timelineCollapsibleMgr.ToggleSection(sectionType)
		}
	}

	// Initialize the cursor to the first section header if it's not set
	if m.timelineCursor == 0 && m.timelineCursorOnHeader == false {
		// Place cursor on the first section header (Overdue)
		m.timelineCursor = 0
		m.timelineCursorOnHeader = true
	}
}

// categorizeTimelineTasks separates tasks into timeline categories (overdue, today, upcoming)
func (m *Model) categorizeTimelineTasks(tasks []task.Task) ([]task.Task, []task.Task, []task.Task) {
	overdueTasks := []task.Task{}
	todayTasks := []task.Task{}
	upcomingTasks := []task.Task{}

	// Get the current date for consistent comparison
	now := time.Now()

	// Loop through all tasks and categorize
	for _, t := range tasks {
		// Skip completed tasks for timeline view
		if t.Status == task.StatusDone || t.IsCompleted {
			continue
		}

		// Skip tasks without due dates - they should not appear in timeline at all
		if t.DueDate == nil {
			continue
		}

		// Use the utility functions from utils.go for consistent date comparison that properly handles timezones
		// These functions normalize dates to UTC to avoid timezone-related issues
		if isBeforeDay(*t.DueDate, now) {
			// Task is due before today = overdue
			overdueTasks = append(overdueTasks, t)
		} else if isSameDay(*t.DueDate, now) {
			// Task is due today = today section
			todayTasks = append(todayTasks, t)
		} else {
			// Task is due after today = upcoming
			upcomingTasks = append(upcomingTasks, t)
		}
	}

	// Sort each category by due date for consistent ordering
	sort.Slice(overdueTasks, func(i, j int) bool {
		return overdueTasks[i].DueDate.Before(*overdueTasks[j].DueDate)
	})

	sort.Slice(todayTasks, func(i, j int) bool {
		return todayTasks[i].Title < todayTasks[j].Title
	})

	sort.Slice(upcomingTasks, func(i, j int) bool {
		return upcomingTasks[i].DueDate.Before(*upcomingTasks[j].DueDate)
	})

	return overdueTasks, todayTasks, upcomingTasks
}

// toggleTimelineSection expands or collapses the section at the given index in the timeline
func (m *Model) toggleTimelineSection(sectionType hooks.SectionType) tea.Cmd {
	if m.timelineCollapsibleMgr != nil {
		// Get the section
		section := m.timelineCollapsibleMgr.GetSection(sectionType)
		if section != nil {
			// Toggle the section
			m.timelineCollapsibleMgr.ToggleSection(section.Type)

			// Use the section name for status updates
			var sectionName string
			switch section.Type {
			case hooks.SectionTypeOverdue:
				sectionName = "Overdue"
			case hooks.SectionTypeToday:
				sectionName = "Today"
			case hooks.SectionTypeUpcoming:
				sectionName = "Upcoming"
			}

			// Show status message
			if section.IsExpanded {
				// Status message removed as requested - no notification for section collapse
			} else {
				m.setStatusMessage("Expanding "+sectionName+" section", "info", 1*time.Second)
			}
		}
	}
	return nil
}
