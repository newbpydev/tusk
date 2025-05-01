package app

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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

	// Determine new status
	var newStatus task.Status
	if curr.Status != task.StatusDone {
		newStatus = task.StatusDone
	} else {
		newStatus = task.StatusTodo
	}

	// --- Optimistic Update ---
	// Update local task state immediately
	m.tasks[m.cursor].Status = newStatus
	m.tasks[m.cursor].IsCompleted = (newStatus == task.StatusDone)

	// Re-categorize tasks locally
	m.categorizeTasks(m.tasks)

	// Keep track of the task ID for re-selection
	targetTaskID := toggledID

	// Find the task in the recategorized lists and update cursor
	foundTask := false
	for i, t := range m.tasks {
		if t.ID == targetTaskID {
			m.cursor = i
			foundTask = true
			break
		}
	}

	// If for some reason the task isn't found (shouldn't happen), just keep the current cursor
	if foundTask {
		// Update visual cursor and task cursor based on the potentially new position
		m.updateVisualCursorFromTaskCursor()
		m.updateTaskCursorFromVisualCursor()
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
	todayDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := todayDate.AddDate(0, 0, 1)
	for _, t := range m.tasks {
		// Filter out tasks without due dates or completed tasks? Depends on requirements.
		if t.DueDate == nil || t.Status == task.StatusDone {
			continue
		}
		dueDate := *t.DueDate
		if dueDate.Before(todayDate) {
			overdue = append(overdue, t)
		} else if dueDate.Before(tomorrow) { // Tasks due today
			todayTasks = append(todayTasks, t)
		} else { // Tasks due tomorrow or later
			upcoming = append(upcoming, t)
		}
	}
	// Sorting logic could be added here if needed (e.g., sort each category by due date)
	return overdue, todayTasks, upcoming
}
