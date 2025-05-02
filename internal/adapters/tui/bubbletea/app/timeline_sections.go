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
}

// categorizeTimelineTasks separates tasks into timeline categories (overdue, today, upcoming)
func (m *Model) categorizeTimelineTasks(tasks []task.Task) ([]task.Task, []task.Task, []task.Task) {
	overdueTasks := []task.Task{}
	todayTasks := []task.Task{}
	upcomingTasks := []task.Task{}
	
	// Get the current date for consistent comparison
	now := time.Now()
	todayDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Loop through all tasks and categorize
	for _, t := range tasks {
		// Skip tasks without due dates
		if t.DueDate == nil {
			continue
		}
		
		// Skip completed tasks for timeline view
		if t.Status == task.StatusDone || t.IsCompleted {
			continue
		}
		
		// Normalize task due date
		taskDueDate := time.Date(t.DueDate.Year(), t.DueDate.Month(), t.DueDate.Day(), 0, 0, 0, 0, t.DueDate.Location())
		
		// Compare to determine category
		if taskDueDate.Before(todayDate) {
			overdueTasks = append(overdueTasks, t)
		} else if taskDueDate.Equal(todayDate) {
			todayTasks = append(todayTasks, t)
		} else {
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
				m.setStatusMessage("Collapsing " + sectionName + " section", "info", 1*time.Second)
			} else {
				m.setStatusMessage("Expanding " + sectionName + " section", "info", 1*time.Second)
			}
		}
	}
	return nil
}
