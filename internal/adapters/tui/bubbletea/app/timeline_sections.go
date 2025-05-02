package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/hooks"
	"github.com/newbpydev/tusk/internal/core/task"
)

// initTimelineCollapsibleSections initializes or resets the sections in the timeline view.
func (m *Model) initTimelineCollapsibleSections() {
	if m.timelineCollapsibleMgr == nil {
		m.timelineCollapsibleMgr = hooks.NewCollapsibleManager()
	}

	// Get the task counts for each timeline section
	overdue, today, upcoming := getTasksByTimeCategory(m.tasks)

	// Update collapsible sections with latest counts
	m.timelineCollapsibleMgr.ClearSections()
	m.timelineCollapsibleMgr.AddSection(hooks.SectionTypeOverdue, "Overdue", len(overdue), 0)
	m.timelineCollapsibleMgr.AddSection(hooks.SectionTypeToday, "Today", len(today), len(overdue))
	m.timelineCollapsibleMgr.AddSection(hooks.SectionTypeUpcoming, "Upcoming", len(upcoming), len(overdue)+len(today))
}

// getTasksByTimeCategory categorizes tasks into overdue, due today, and upcoming
func getTasksByTimeCategory(tasks []task.Task) (overdue, today, upcoming []task.Task) {
	// Get the current date for consistent comparison
	now := time.Now()
	todayDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	for _, t := range tasks {
		// Skip tasks without due dates
		if t.DueDate == nil {
			continue
		}
		
		// Extract just the date part for comparison
		taskDueDate := time.Date(t.DueDate.Year(), t.DueDate.Month(), t.DueDate.Day(), 0, 0, 0, 0, t.DueDate.Location())
		
		// Categorize based on due date
		if taskDueDate.Before(todayDate) {
			overdue = append(overdue, t)
		} else if taskDueDate.Equal(todayDate) {
			today = append(today, t)
		} else {
			upcoming = append(upcoming, t)
		}
	}

	return overdue, today, upcoming
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
				m.setStatusMessage("Collapsing " + sectionName + " section", "info", 1*time.Second)
			} else {
				m.setStatusMessage("Expanding " + sectionName + " section", "info", 1*time.Second)
			}
		}
	}
	return nil
}
