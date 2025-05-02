package app

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/hooks"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/messages"
	"github.com/newbpydev/tusk/internal/core/task"
)

// refreshTasks initiates a fetch for the latest tasks.
func (m *Model) refreshTasks() tea.Cmd {
	// Call to setLoadingStatus will be in status.go
	m.setLoadingStatus("Loading tasks...")
	return func() tea.Msg {
		tasks, err := m.taskSvc.List(m.ctx, m.userID)
		if err != nil {
			return messages.ErrorMsg(fmt.Errorf("failed to refresh tasks: %v", err))
		}

		// Debug check for Task 9 and Task 12
		var found9, found12 bool
		var date9, date12 string
		now := time.Now()

		for _, t := range tasks {
			if t.Title == "Task 9" && t.DueDate != nil {
				found9 = true
				date9 = t.DueDate.Format("2006-01-02 15:04:05")

				// Check date classification
				if isSameDay(*t.DueDate, now) {
					date9 += " (Today)"
				} else if isBeforeDay(*t.DueDate, now) {
					date9 += " (Before)"
				} else {
					date9 += " (After)"
				}
			}
			if t.Title == "Task 12" && t.DueDate != nil {
				found12 = true
				date12 = t.DueDate.Format("2006-01-02 15:04:05")

				// Check date classification
				if isSameDay(*t.DueDate, now) {
					date12 += " (Today)"
				} else if isBeforeDay(*t.DueDate, now) {
					date12 += " (Before)"
				} else {
					date12 += " (After)"
				}
			}
		}

		if found9 || found12 {
			m.setStatusMessage(
				fmt.Sprintf("Found in refresh: Task 9=%v [%s], Task 12=%v [%s]",
					found9, date9, found12, date12),
				"info", 2*time.Second)
		}

		// Call to categorizeTasks will be in sections.go
		m.categorizeTasks(tasks)
		return messages.TasksRefreshedMsg{Tasks: tasks}
	}
}

// toggleTaskCompletion changes the status of the selected task between Todo and Done.
func (m *Model) toggleTaskCompletion() tea.Cmd {
	if len(m.tasks) == 0 || m.cursor >= len(m.tasks) || m.cursorOnHeader {
		return nil // Cannot toggle status if no task is selected or cursor is on header
	}

	// Get current task and its ID
	curr := m.tasks[m.cursor]
	toggledID := curr.ID

	// Determine the current section for this task
	var currentSectionType hooks.SectionType
	if curr.Status == task.StatusDone {
		currentSectionType = hooks.SectionTypeCompleted
	} else {
		if curr.ParentID != nil {
			currentSectionType = hooks.SectionTypeProjects
		} else {
			currentSectionType = hooks.SectionTypeTodo
		}
	}

	// Store the current cursor positions
	originalVisualCursor := m.visualCursor

	// IMPORTANT: First get a copy of the tasks in the current section BEFORE changing anything
	// This ensures we properly identify the position and next task
	var tasksInCurrentSection []task.Task
	switch currentSectionType {
	case hooks.SectionTypeTodo:
		tasksInCurrentSection = make([]task.Task, len(m.todoTasks))
		copy(tasksInCurrentSection, m.todoTasks)
	case hooks.SectionTypeProjects:
		tasksInCurrentSection = make([]task.Task, len(m.projectTasks))
		copy(tasksInCurrentSection, m.projectTasks)
	case hooks.SectionTypeCompleted:
		tasksInCurrentSection = make([]task.Task, len(m.completedTasks))
		copy(tasksInCurrentSection, m.completedTasks)
	}

	// Find the position of the current task in its section
	currentPositionInSection := -1
	for i, t := range tasksInCurrentSection {
		if t.ID == toggledID {
			currentPositionInSection = i
			break
		}
	}

	// Find the next task ID to select (if any)
	var nextTaskID int32 = -1
	if currentPositionInSection != -1 {
		if currentPositionInSection+1 < len(tasksInCurrentSection) {
			// There's a next task in this section
			nextTaskID = tasksInCurrentSection[currentPositionInSection+1].ID
		} else if len(tasksInCurrentSection) > 1 {
			// We're at the last task, but there are other tasks in the section
			// In this case, after the task is removed, the previous task becomes the last
			nextTaskID = tasksInCurrentSection[currentPositionInSection-1].ID
		}
	}

	// Determine new status
	var newStatus task.Status
	if curr.Status != task.StatusDone {
		newStatus = task.StatusDone
	} else {
		newStatus = task.StatusTodo
	}

	// --- Start Optimistic Update ---
	// Create updated task
	updatedTask := curr
	updatedTask.Status = newStatus
	updatedTask.IsCompleted = (newStatus == task.StatusDone)

	// Update task in the main list
	for i, t := range m.tasks {
		if t.ID == toggledID {
			m.tasks[i] = updatedTask
			break
		}
	}

	// Re-categorize tasks with the updated data
	m.categorizeTasks(m.tasks)

	// Also update timeline categories to ensure the task appears correctly in the timeline
	// This is critical when toggling task completion, especially for tasks with due dates
	m.overdueTasks, m.todayTasks, m.upcomingTasks = m.categorizeTimelineTasks(m.tasks)

	// Now handle cursor positioning after task recategorization
	if nextTaskID != -1 {
		// If we identified a next task, find and select it
		found := false
		for i, t := range m.tasks {
			if t.ID == nextTaskID {
				m.cursor = i
				m.cursorOnHeader = false
				found = true
				break
			}
		}

		if found {
			// Update visual cursor position
			m.updateVisualCursorFromTaskCursor()
		} else {
			// If next task wasn't found, fall back to selecting the section header
			m.selectSectionHeader(currentSectionType)
		}
	} else {
		// No next task identified, select the section header
		m.selectSectionHeader(currentSectionType)
	}

	// Ensure cursor is visible in the viewport
	if m.visualCursor != originalVisualCursor {
		viewportHeight := 10 // Approximate visible lines

		// Adjust scroll if cursor moved above visible area
		if m.visualCursor < m.taskListOffset {
			m.taskListOffset = m.visualCursor
		}

		// Adjust scroll if cursor moved below visible area
		if m.visualCursor >= m.taskListOffset+viewportHeight {
			m.taskListOffset = m.visualCursor - viewportHeight + 1
		}

		// Reset detail panel offset when selection changes
		m.taskDetailsOffset = 0
	}
	// --- End Optimistic Update ---

	// Call server to update
	return func() tea.Msg {
		updatedTask, err := m.taskSvc.ChangeStatus(m.ctx, int64(toggledID), newStatus)
		if err != nil {
			return messages.StatusUpdateErrorMsg{TaskIndex: m.cursor, TaskTitle: curr.Title, Err: err}
		}

		return messages.StatusUpdateSuccessMsg{
			Task:    updatedTask,
			Message: "Task status updated successfully",
		}
	}
}

// selectSectionHeader helps position the cursor on a specific section header
func (m *Model) selectSectionHeader(sectionType hooks.SectionType) {
	for i, section := range m.collapsibleManager.Sections {
		if section.Type == sectionType {
			// Calculate the visual index of the section header
			var headerIndex int = 0
			for j := 0; j < i; j++ {
				headerIndex++ // Count the header
				if m.collapsibleManager.Sections[j].IsExpanded {
					// Add items in expanded sections
					// Get the item count directly from the section, which will work for any section type
					headerIndex += m.collapsibleManager.Sections[j].ItemCount
				}
			}

			// Set cursor to the section header
			m.visualCursor = headerIndex
			m.cursorOnHeader = true

			// Update the task cursor from this visual position
			m.updateTaskCursorFromVisualCursor()
			break
		}
	}
}

// deleteCurrentTask deletes the currently selected task.
func (m *Model) deleteCurrentTask() tea.Cmd {
	if len(m.tasks) == 0 || m.cursor >= len(m.tasks) || m.cursorOnHeader {
		return nil // Cannot delete if no task selected or cursor is on header
	}
	taskTitle := m.tasks[m.cursor].Title
	taskID := int64(m.tasks[m.cursor].ID)
	taskIndex := m.cursor // Store index before potential list modification

	// Call to setLoadingStatus will be in status.go
	m.setLoadingStatus("Deleting task...")

	// Optimistic update: Remove the task locally first
	// m.tasks = append(m.tasks[:taskIndex], m.tasks[taskIndex+1:]...)
	// Adjust cursor if needed
	// if m.cursor >= len(m.tasks) {
	// 	 m.cursor = max(0, len(m.tasks)-1)
	// }
	// m.categorizeTasks(m.tasks)
	// m.updateVisualCursorFromTaskCursor()
	// m.updateTaskCursorFromVisualCursor()
	// --- End Optimistic Delete ---

	return func() tea.Msg {
		err := m.taskSvc.Delete(m.ctx, taskID)
		if err != nil {
			// Rollback optimistic delete? Might be simpler to just refresh.
			return messages.StatusUpdateErrorMsg{TaskIndex: taskIndex, TaskTitle: taskTitle, Err: err}
		}
		// Instead of returning TasksRefreshedMsg directly, trigger a refresh command.
		// This keeps the refresh logic centralized.
		m.viewMode = "list" // Switch back to list view after delete
		// Call to setSuccessStatus will be in status.go
		m.setSuccessStatus(fmt.Sprintf("Task '%s' deleted", taskTitle))
		return m.refreshTasks()() // Immediately invoke the refresh command func
	}
}

// createNewTask creates a new task from the form fields.
// This might move mostly to form.go, which would then return a command.
func (m *Model) createNewTask() tea.Cmd {
	// This validation might live in form.go before calling the create command
	if m.formTitle == "" {
		m.err = fmt.Errorf("title is required")
		// Call to setErrorStatus will be in status.go
		m.setErrorStatus("Title is required")
		return nil
	}
	// Call to setLoadingStatus will be in status.go
	m.setLoadingStatus("Creating new task...")

	// Prepare task data from form fields
	var dueDate *time.Time
	if m.formDueDate != "" {
		// Date parsing could be a helper in utils.go or stay in form.go
		date, err := time.Parse("2006-01-02", m.formDueDate)
		if err == nil {
			dueDate = &date
		} else {
			// Error handling might live in form.go or status.go
			m.setErrorStatus(fmt.Sprintf("Invalid date format: %v", err))
			m.err = fmt.Errorf("invalid date format: %v", err)
			// Clear loading status?
			m.clearLoadingStatus()
			return nil
		}
	}
	priority := task.PriorityLow // Default priority
	if m.formPriority == string(task.PriorityMedium) {
		priority = task.PriorityMedium
	} else if m.formPriority == string(task.PriorityHigh) {
		priority = task.PriorityHigh
	}

	// Capture form data before clearing
	title := m.formTitle
	description := m.formDescription
	// Clearing form fields should happen *after* the command function is prepared,
	// or ideally, be handled within form.go when transitioning viewMode.
	m.formTitle = ""
	m.formDescription = ""
	m.formPriority = ""
	m.formDueDate = ""
	m.formStatus = ""
	m.activeField = 0
	m.viewMode = "list" // Switch back to list view after initiating create

	return func() tea.Msg {
		// Actual creation logic
		_, err := m.taskSvc.Create(m.ctx, m.userID, nil, title, description, dueDate, priority, []string{})
		if err != nil {
			// Return error message for the Update loop to handle
			return messages.ErrorMsg(fmt.Errorf("failed to create task: %v", err))
		}
		// Trigger a refresh command instead of returning TasksRefreshedMsg directly.
		// Call to setSuccessStatus will be in status.go
		m.setSuccessStatus(fmt.Sprintf("Task '%s' created", title))
		return m.refreshTasks()() // Immediately invoke the refresh command func
	}
}

// getTasksByTimeCategory organizes tasks into overdue, today, and upcoming categories.
// This is primarily used by the timeline view and might fit better there or in utils.go.
func (m *Model) getTasksByTimeCategory() ([]task.Task, []task.Task, []task.Task) {
	var overdue, todayTasks, upcoming []task.Task
	now := time.Now()

	for _, t := range m.tasks {
		// Skip tasks without due dates or completed tasks
		if t.DueDate == nil || t.Status == task.StatusDone {
			continue
		}

		// Use the same reliable date comparison logic used elsewhere
		if isBeforeDay(*t.DueDate, now) {
			// Task is due before today = overdue
			overdue = append(overdue, t)
		} else if isSameDay(*t.DueDate, now) {
			// Task is due today = today section
			todayTasks = append(todayTasks, t)
		} else {
			// Task is due after today = upcoming
			upcoming = append(upcoming, t)
		}
	}

	return overdue, todayTasks, upcoming
}
