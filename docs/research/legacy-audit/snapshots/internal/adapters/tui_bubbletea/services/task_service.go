// Package services contains business logic for the TUI application.
// It separates the logic from the UI components and handlers.
package services

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/tusk/internal/adapters/tui/bubbletea/messages"
	"github.com/newbpydev/tusk/internal/core/task"
)

// TaskServiceInterface defines the interface required for task operations in the TUI
type TaskServiceInterface interface {
	List(ctx context.Context, userID int64) ([]task.Task, error)
	Create(ctx context.Context, userID int64, parentID *int64, title, description string,
		dueDate *time.Time, priority task.Priority, tags []string) (task.Task, error)
	Update(ctx context.Context, taskID int64, title, description string,
		dueDate *time.Time, priority task.Priority, tags []string) (task.Task, error)
	Delete(ctx context.Context, taskID int64) error
	Complete(ctx context.Context, taskID int64) (task.Task, error)
	ChangeStatus(ctx context.Context, taskID int64, status task.Status) (task.Task, error)
}

// TaskService manages task-related operations
type TaskService struct {
	taskSvc TaskServiceInterface
}

// NewTaskService creates a new TaskService
func NewTaskService(taskSvc TaskServiceInterface) *TaskService {
	return &TaskService{
		taskSvc: taskSvc,
	}
}

// RefreshTasks initiates a fetch for the latest tasks.
func (s *TaskService) RefreshTasks(ctx context.Context, userID int64, statusCallback func(string)) tea.Cmd {
	// Call status callback to update UI
	statusCallback("Loading tasks...")
	
	return func() tea.Msg {
		tasks, err := s.taskSvc.List(ctx, userID)
		if err != nil {
			return messages.ErrorMsg(fmt.Errorf("failed to refresh tasks: %v", err))
		}
		
		return messages.TasksRefreshedMsg{Tasks: tasks}
	}
}

// ToggleTaskCompletion changes the status of a task between Todo and Done.
func (s *TaskService) ToggleTaskCompletion(
	ctx context.Context, 
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

	// Create updated task
	updatedTask := currentTask
	updatedTask.Status = newStatus
	updatedTask.IsCompleted = (newStatus == task.StatusDone)

	// Return command that will perform the actual update
	return func() tea.Msg {
		// Perform update in the database using the ChangeStatus method
		result, err := s.taskSvc.ChangeStatus(ctx, int64(taskID), newStatus)
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

// DeleteTask deletes the task with the given ID.
func (s *TaskService) DeleteTask(
	ctx context.Context,
	taskID int32,
	statusCallback func(string),
) tea.Cmd {
	// Convert int32 to int64 for the core service call
	return func() tea.Msg {
		err := s.taskSvc.Delete(ctx, int64(taskID))
		if err != nil {
			return messages.ErrorMsg(fmt.Errorf("failed to delete task: %v", err))
		}
		
		message := fmt.Sprintf("Task '%d' deleted", taskID)
		return struct {
			Message string
			Type    string
		}{
			Message: message,
			Type:    "success",
		}
	}
}

// CreateTask creates a new task with the given properties.
func (s *TaskService) CreateTask(
	ctx context.Context,
	t task.Task,
	statusCallback func(string),
) tea.Cmd {
	// Call status callback to update UI
	statusCallback("Loading tasks...")
	
	return func() tea.Msg {
		// Extract fields from task struct for the core service call
		var parentID *int64
		if t.ParentID != nil {
			pID := int64(*t.ParentID)
			parentID = &pID
		}

		// Convert Priority to core enum
		pri := task.PriorityMedium
		if t.Priority == "high" {
			pri = task.PriorityHigh
		} else if t.Priority == "low" {
			pri = task.PriorityLow
		}

		// Convert user ID from int32 to int64
		userID := int64(t.UserID)
		
		// Convert tags from []task.Tag to []string
		tagStrs := make([]string, 0, len(t.Tags))
		for _, tag := range t.Tags {
			tagStrs = append(tagStrs, tag.Name)
		}
		
		// Call the task service to create the task
		createdTask, err := s.taskSvc.Create(ctx, userID, parentID, t.Title, *t.Description, t.DueDate, pri, tagStrs)
		if err != nil {
			return messages.ErrorMsg(fmt.Errorf("failed to create task: %v", err))
		}
		
		message := fmt.Sprintf("Task '%s' created", t.Title)
		return messages.StatusUpdateSuccessMsg{
			Task:    createdTask,
			Message: message,
		}
	}
}

// UpdateTask updates an existing task with new properties.
func (s *TaskService) UpdateTask(
	ctx context.Context,
	t task.Task,
	statusCallback func(string),
) tea.Cmd {
	// Call status callback to update UI
	statusCallback("Loading tasks...")
	
	return func() tea.Msg {
		// Extract fields from task struct for the core service call
		// Convert Priority to core enum
		pri := task.PriorityMedium
		if t.Priority == "high" {
			pri = task.PriorityHigh
		} else if t.Priority == "low" {
			pri = task.PriorityLow
		}

		// Convert tags from []task.Tag to []string
		tagStrs := make([]string, 0, len(t.Tags))
		for _, tag := range t.Tags {
			tagStrs = append(tagStrs, tag.Name)
		}
		
		// Call the task service to update the task
		updatedTask, err := s.taskSvc.Update(ctx, int64(t.ID), t.Title, *t.Description, t.DueDate, pri, tagStrs)
		if err != nil {
			return messages.ErrorMsg(fmt.Errorf("failed to update task: %v", err))
		}
		
		message := fmt.Sprintf("Task '%s' updated", t.Title)
		return messages.StatusUpdateSuccessMsg{
			Task:    updatedTask,
			Message: message,
		}
	}
}
