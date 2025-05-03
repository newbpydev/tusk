// Package commands contains tea.Cmd factories for side effects
// in the TUI application. This follows the command pattern to
// separate side effects from the application's core logic.
package commands

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/messages"
	"github.com/newbpydev/tusk/internal/core/task"
)

// TaskService defines the operations required for task commands
type TaskService interface {
	List(ctx context.Context, userID int32) ([]task.Task, error)
	Create(ctx context.Context, t task.Task) (task.Task, error)
	Update(ctx context.Context, t task.Task) (task.Task, error)
	Delete(ctx context.Context, taskID int32) error
}

// FetchTasks creates a command to fetch tasks from the service
func FetchTasks(ctx context.Context, taskSvc TaskService, userID int32) tea.Cmd {
	return func() tea.Msg {
		tasks, err := taskSvc.List(ctx, userID)
		if err != nil {
			return messages.ErrorMsg(fmt.Errorf("failed to fetch tasks: %v", err))
		}

		return messages.TasksRefreshedMsg{Tasks: tasks}
	}
}

// ToggleTaskCompletion toggles the completion status of a task
func ToggleTaskCompletion(
	ctx context.Context,
	taskSvc TaskService,
	taskID int32,
	currentTask task.Task,
) tea.Cmd {
	// Determine new status
	var newStatus task.Status
	if currentTask.Status != task.StatusDone {
		newStatus = task.StatusDone
	} else {
		newStatus = task.StatusTodo
	}

	// Create updated task with toggled status
	updatedTask := currentTask
	updatedTask.Status = newStatus
	updatedTask.IsCompleted = (newStatus == task.StatusDone)

	// Return command that will perform the actual update
	return func() tea.Msg {
		// Perform update in the database
		result, err := taskSvc.Update(ctx, updatedTask)
		if err != nil {
			return messages.StatusUpdateErrorMsg{
				Err:       err,
				TaskTitle: currentTask.Title,
			}
		}
		
		message := fmt.Sprintf("Task '%s' marked as %s", currentTask.Title, newStatus)
		return messages.StatusUpdateSuccessMsg{
			Task:    result,
			Message: message,
		}
	}
}

// CreateTask creates a new task
func CreateTask(
	ctx context.Context,
	taskSvc TaskService,
	userID int64,
	title string,
	description *string,
	priority task.Priority,
	dueDate *time.Time,
	isCompleted bool,
	parentID *int32,
) tea.Cmd {
	// Create task with the provided information
	// Convert userID from int64 to int32 for the task struct
	userID32 := int32(userID)
	newTask := task.Task{
		UserID:      userID32,
		Title:       title,
		Description: description,
		Priority:    priority,
		DueDate:     dueDate,
		Status:      task.StatusTodo,
		IsCompleted: isCompleted,
		ParentID:    parentID,
	}
	
	// If marked as completed, set status accordingly
	if isCompleted {
		newTask.Status = task.StatusDone
	}

	return func() tea.Msg {
		// Create task in the database
		result, err := taskSvc.Create(ctx, newTask)
		if err != nil {
			return messages.ErrorMsg(fmt.Errorf("failed to create task: %v", err))
		}
		
		message := fmt.Sprintf("Task '%s' created", title)
		return messages.StatusUpdateSuccessMsg{
			Task:    result,
			Message: message,
		}
	}
}

// UpdateTask creates a command to update an existing task
func UpdateTask(
	ctx context.Context,
	taskSvc TaskService,
	taskID int32,
	title string,
	description *string,
	priority task.Priority,
	dueDate *time.Time,
	isCompleted bool,
	parentID *int32,
) tea.Cmd {
	// Create task with updated information
	updatedTask := task.Task{
		ID:          taskID,
		Title:       title,
		Description: description,
		Priority:    priority,
		DueDate:     dueDate,
		Status:      task.StatusTodo,
		IsCompleted: isCompleted,
		ParentID:    parentID,
	}
	
	// If marked as completed, set status accordingly
	if isCompleted {
		updatedTask.Status = task.StatusDone
	}

	return func() tea.Msg {
		// Update task in the database
		result, err := taskSvc.Update(ctx, updatedTask)
		if err != nil {
			return messages.ErrorMsg(fmt.Errorf("failed to update task: %v", err))
		}
		
		message := fmt.Sprintf("Task '%s' updated", title)
		return messages.StatusUpdateSuccessMsg{
			Task:    result,
			Message: message,
		}
	}
}

// DeleteTask creates a command to delete a task
func DeleteTask(
	ctx context.Context,
	taskSvc TaskService,
	taskID int32,
	taskTitle string,
) tea.Cmd {
	return func() tea.Msg {
		err := taskSvc.Delete(ctx, taskID)
		if err != nil {
			return messages.ErrorMsg(fmt.Errorf("failed to delete task: %v", err))
		}
		return nil
	}
}
