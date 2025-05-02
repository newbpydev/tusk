package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/hooks"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/messages"
	"github.com/newbpydev/tusk/internal/core/task"
)

// getTimelineTaskID finds the task ID of the task at the current timeline cursor position.
// If the cursor is on a section header or an invalid position, returns 0 (invalid ID).
func (m *Model) getTimelineTaskID() int32 {
	// If we're on a header or invalid position, return 0 (invalid ID)
	if m.timelineCursorOnHeader || m.timelineCursor < 0 {
		return int32(0) // No task selected
	}
	
	// Get overdue, today, and upcoming tasks
	overdue, today, upcoming := m.getTimelineFilteredTasks()
	
	// Find the section our cursor is in and the task it points to
	var currentIndex int = 0
	for _, section := range m.timelineCollapsibleMgr.Sections {
		// Skip the section header
		currentIndex++
		
		// If section is expanded
		if section.IsExpanded {
			// If our cursor is within this section's items
			if m.timelineCursor >= currentIndex && m.timelineCursor < currentIndex + section.ItemCount {
				// Calculate relative index within this section
				relativeIndex := m.timelineCursor - currentIndex
				
				// Return the task ID based on the section type
				switch section.Type {
				case hooks.SectionTypeOverdue:
					if relativeIndex < len(overdue) {
						return overdue[relativeIndex].ID
					}
				case hooks.SectionTypeToday:
					if relativeIndex < len(today) {
						return today[relativeIndex].ID
					}
				case hooks.SectionTypeUpcoming:
					if relativeIndex < len(upcoming) {
						return upcoming[relativeIndex].ID
					}
				}
				break
			}
			
			// Add section items to current index
			currentIndex += section.ItemCount
		}
	}
	
	// No valid task found
	return int32(0)
}

// getTaskIndexByID returns the index of a task with the given ID in the main task list
// Returns -1 if no task with that ID is found
func (m *Model) getTaskIndexByID(taskID int32) int {
	if taskID <= 0 {
		return -1
	}
	
	for i, t := range m.tasks {
		if t.ID == taskID {
			return i
		}
	}
	
	return -1
}

// getTimelineTaskIndex finds the real task index in the main task list
// based on the current timeline cursor position.
func (m *Model) getTimelineTaskIndex() int {
	// Get the task ID first
	taskID := m.getTimelineTaskID()
	
	// Then get the index from the ID
	return m.getTaskIndexByID(taskID)
}

// getTimelineFilteredTasks returns the filtered tasks used in the timeline view
// This now returns the dedicated task slices that are maintained by the model
func (m *Model) getTimelineFilteredTasks() ([]task.Task, []task.Task, []task.Task) {
	// Return the cached categorized tasks from the model
	// These are updated whenever the task list changes via initTimelineCollapsibleSections()
	return m.overdueTasks, m.todayTasks, m.upcomingTasks
}

// toggleTimelineTaskCompletion toggles the completion status of the task selected in the timeline
func (m *Model) toggleTimelineTaskCompletion() tea.Cmd {
	// Get the task ID from the current timeline cursor position
	taskID := m.getTimelineTaskID()
	if taskID <= 0 {
		return nil // No valid task selected
	}
	
	// Find the task index by ID
	taskIndex := m.getTaskIndexByID(taskID)
	if taskIndex < 0 || taskIndex >= len(m.tasks) {
		return nil // Task not found in the main list
	}
	
	// Get the task
	curr := m.tasks[taskIndex]
	
	// Determine the new status (toggle between Done and Todo)
	var newStatus task.Status
	if curr.Status == task.StatusDone {
		newStatus = task.StatusTodo
	} else {
		newStatus = task.StatusDone
	}
	
	// Optimistically update the UI
	// Copy task data for status update
	updatedTask := curr
	updatedTask.Status = newStatus
	
	// Update in the tasks list
	m.tasks[taskIndex] = updatedTask
	
	// Re-categorize tasks to update all relevant UI elements
	m.categorizeTasks(m.tasks)
	
	// Make sure to re-initialize the timeline sections as well
	m.initTimelineCollapsibleSections() 
	
	// Reset offsets to ensure good UX
	m.taskDetailsOffset = 0
	
	// Call server to update
	return func() tea.Msg {
		updatedTask, err := m.taskSvc.ChangeStatus(m.ctx, int64(taskID), newStatus)
		if err != nil {
			return messages.StatusUpdateErrorMsg{TaskIndex: taskIndex, TaskTitle: curr.Title, Err: err}
		}
		
		return messages.StatusUpdateSuccessMsg{
			Task:    updatedTask,
			Message: "Task status updated successfully",
		}
	}
}
