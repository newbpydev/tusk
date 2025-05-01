package app

import (
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/hooks"
	"github.com/newbpydev/tusk/internal/core/task"
)

// initCollapsibleSections initializes or resets the sections in the task list view.
func (m *Model) initCollapsibleSections() {
	if m.collapsibleManager == nil {
		m.collapsibleManager = hooks.NewCollapsibleManager()
	}

	// Recalculate task counts for categorization
	m.categorizeTasks(m.tasks) // This ensures counts are based on the latest task data

	// Update collapsible sections with latest counts
	m.collapsibleManager.ClearSections()
	m.collapsibleManager.AddSection(hooks.SectionTypeTodo, "Todo", len(m.todoTasks), 0)
	// Projects section might need different logic if it represents nested tasks/folders
	m.collapsibleManager.AddSection(hooks.SectionTypeProjects, "Projects", len(m.projectTasks), len(m.todoTasks))
	m.collapsibleManager.AddSection(hooks.SectionTypeCompleted, "Completed", len(m.completedTasks), len(m.todoTasks)+len(m.projectTasks))

	// Reset visual cursor based on the current task cursor, accounting for sections
	m.updateVisualCursorFromTaskCursor()
}

// categorizeTasks separates the main task list into Todo, Projects, and Completed slices.
// This is used by initCollapsibleSections and potentially the View logic.
func (m *Model) categorizeTasks(tasks []task.Task) {
	// Clear existing categorized slices
	m.todoTasks = m.todoTasks[:0]
	m.projectTasks = m.projectTasks[:0]
	m.completedTasks = m.completedTasks[:0]

	// Iterate through the main tasks list and append to appropriate slices
	for _, t := range tasks {
		if t.Status == task.StatusDone {
			m.completedTasks = append(m.completedTasks, t)
		} else if t.ParentID != nil {
			// Assuming tasks with a ParentID belong to the "Projects" category for now
			// This might need refinement based on how projects are structured
			m.projectTasks = append(m.projectTasks, t)
		} else {
			// Tasks that are not Done and have no ParentID go to Todo
			m.todoTasks = append(m.todoTasks, t)
		}
	}
	// Note: This function only categorizes; it doesn't update the collapsible manager counts directly.
	// initCollapsibleSections is responsible for using these categorized slices to update the manager.
}

// updateVisualCursorFromTaskCursor translates the internal task index (m.cursor)
// into the visible cursor position (m.visualCursor) considering collapsed sections.
func (m *Model) updateVisualCursorFromTaskCursor() {
	if m.collapsibleManager == nil {
		m.visualCursor = m.cursor // Fallback if manager isn't initialized
		m.cursorOnHeader = false
		return
	}

	// If the task list is empty, reset cursors
	if len(m.tasks) == 0 {
		m.cursor = 0
		m.visualCursor = 0
		m.cursorOnHeader = false // Or true if the first item is always a header?
		return
	}

	// Ensure cursor is within bounds of the actual task list
	// Call to max/min would be from utils.go
	m.cursor = min(max(0, m.cursor), len(m.tasks)-1)

	// Find the visual index corresponding to the actual task index
	m.visualCursor = m.collapsibleManager.GetVisibleIndexFromTaskIndex(m.cursor)

	// If the task wasn't found (e.g., it's in a collapsed section),
	// try to position the cursor reasonably, perhaps on the section header.
	if m.visualCursor == -1 {
		// Find the section the task *would* be in and get the header index
		// This requires additional logic in CollapsibleManager or here.
		// For now, reset to the top.
		m.visualCursor = 0
	}

	// After finding the visual cursor, check if it landed on a header.
	// This check needs to be robust.
	m.cursorOnHeader = m.collapsibleManager.IsSectionHeader(m.visualCursor)

	// If the visual cursor landed on a header *unintentionally* (because the target task
	// was hidden), we might want to adjust it, e.g., move to the next visible task.
	// This requires careful handling. For now, we accept landing on the header.
}

// updateTaskCursorFromVisualCursor translates the visual cursor position (m.visualCursor)
// back to the internal task index (m.cursor), handling section headers.
func (m *Model) updateTaskCursorFromVisualCursor() {
	if m.collapsibleManager == nil {
		m.cursor = m.visualCursor // Fallback
		m.cursorOnHeader = false
		return
	}

	totalVisibleItems := m.collapsibleManager.GetItemCount()
	if totalVisibleItems == 0 {
		m.cursor = 0
		m.visualCursor = 0
		m.cursorOnHeader = false
		return
	}

	// Ensure visual cursor is within bounds
	// Call to max/min would be from utils.go
	m.visualCursor = min(max(0, m.visualCursor), totalVisibleItems-1)

	// Check if the visual cursor is pointing to a section header
	if m.collapsibleManager.IsSectionHeader(m.visualCursor) {
		m.cursorOnHeader = true
		// Keep m.cursor potentially pointing to the last selected *task*? Or reset it?
		// Resetting might be safer if actions depend on a valid task index.
		// m.cursor = -1 // Or some invalid index
	} else {
		// If not on a header, get the corresponding actual task index
		taskIndex := m.collapsibleManager.GetActualTaskIndex(m.visualCursor)
		if taskIndex != -1 && taskIndex < len(m.tasks) {
			m.cursor = taskIndex
			m.cursorOnHeader = false
		} else {
			// This case (visual cursor not header, but invalid task index) shouldn't
			// happen with correct CollapsibleManager logic, but handle defensively.
			m.cursorOnHeader = true // Treat as if on a header or invalid state
			// Reset m.cursor?
		}
	}
}
