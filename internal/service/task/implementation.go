package task

import (
	"context"
	"time"

	"github.com/newbpydev/tusk/internal/core/errors"
	"github.com/newbpydev/tusk/internal/core/task"
	repo "github.com/newbpydev/tusk/internal/ports/output"
)

// taskService implements the Service interface
type taskService struct {
	repo repo.TaskRepository
}

// NewTaskService creates a new instance of the task service
func NewTaskService(r repo.TaskRepository) Service {
	return &taskService{repo: r}
}

// Create creates a new task with the given parameters
func (s *taskService) Create(ctx context.Context, userID int64, parentID *int64, title, description string,
	dueDate *time.Time, priority task.Priority, tags []string) (task.Task, error) {

	// Validate input
	if userID <= 0 {
		return task.Task{}, errors.InvalidInput("user ID must be positive")
	}
	if title == "" {
		return task.Task{}, errors.InvalidInput("title is required")
	}
	if !isValidPriority(priority) {
		priority = task.PriorityMedium // Set default priority if invalid
	}

	// Convert tags to Tag objects
	var taskTags []task.Tag
	for _, tagName := range tags {
		if tagName != "" {
			taskTags = append(taskTags, task.Tag{Name: tagName})
		}
	}

	var desc *string
	if description != "" {
		desc = &description
	}

	var parentTaskID *int32
	if parentID != nil {
		id := int32(*parentID)
		parentTaskID = &id
	}

	now := time.Now()
	newTask := task.Task{
		UserID:       int32(userID),
		ParentID:     parentTaskID,
		Title:        title,
		Description:  desc,
		CreatedAt:    now,
		UpdatedAt:    now,
		DueDate:      dueDate,
		IsCompleted:  false,
		Status:       task.StatusTodo,
		Priority:     priority,
		Tags:         taskTags,
		DisplayOrder: 0, // Will be set by the repository
	}

	createdTask, err := s.repo.Create(ctx, newTask)
	if err != nil {
		return task.Task{}, err
	}

	return createdTask, nil
}

// Show retrieves a task by its ID
func (s *taskService) Show(ctx context.Context, taskID int64) (task.Task, error) {
	if taskID <= 0 {
		return task.Task{}, errors.InvalidInput("task ID must be positive")
	}

	// Get the task with its full tree
	taskWithTree, err := s.repo.GetTaskTree(ctx, taskID)
	if err != nil {
		return task.Task{}, err
	}

	return taskWithTree, nil
}

// List retrieves all tasks for a user
func (s *taskService) List(ctx context.Context, userID int64) ([]task.Task, error) {
	if userID <= 0 {
		return nil, errors.InvalidInput("user ID must be positive")
	}

	// Get all root tasks for the user
	rootTasks, err := s.repo.ListRootTasks(ctx, userID)
	if err != nil {
		return nil, err
	}

	// For each root task, populate its subtasks
	for i, rootTask := range rootTasks {
		fullTask, err := s.repo.GetTaskTree(ctx, int64(rootTask.ID))
		if err != nil {
			continue // Skip this task if we can't get its tree
		}
		rootTasks[i] = fullTask
	}

	return rootTasks, nil
}

// Reorder changes the display order of a task
func (s *taskService) Reorder(ctx context.Context, taskID int64, newOrder int) error {
	if taskID <= 0 {
		return errors.InvalidInput("task ID must be positive")
	}
	if newOrder < 0 {
		return errors.InvalidInput("order must be non-negative")
	}

	return s.repo.ReorderTask(ctx, taskID, newOrder)
}

// Update updates an existing task with the given parameters
func (s *taskService) Update(ctx context.Context, taskID int64, title, description string,
	dueDate *time.Time, priority task.Priority, tags []string) (task.Task, error) {

	if taskID <= 0 {
		return task.Task{}, errors.InvalidInput("task ID must be positive")
	}

	// Get the existing task
	existingTask, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return task.Task{}, err
	}

	// Update the task fields
	if title != "" {
		existingTask.Title = title
	}

	if description != "" {
		existingTask.Description = &description
	} else if description == "" && existingTask.Description != nil {
		// If description is empty string, remove the description
		existingTask.Description = nil
	}

	if dueDate != nil {
		existingTask.DueDate = dueDate
	}

	if isValidPriority(priority) {
		existingTask.Priority = priority
	}

	if tags != nil {
		// Convert tags to Tag objects
		var taskTags []task.Tag
		for _, tagName := range tags {
			if tagName != "" {
				taskTags = append(taskTags, task.Tag{Name: tagName})
			}
		}
		existingTask.Tags = taskTags
	}

	existingTask.UpdatedAt = time.Now()

	// Update the task in the repository
	err = s.repo.Update(ctx, existingTask)
	if err != nil {
		return task.Task{}, err
	}

	// Get the updated task with its full tree
	updatedTask, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return task.Task{}, err
	}

	return updatedTask, nil
}

// Delete removes a task
func (s *taskService) Delete(ctx context.Context, taskID int64) error {
	if taskID <= 0 {
		return errors.InvalidInput("task ID must be positive")
	}

	// Check if the task exists
	_, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, taskID)
}

// Complete marks a task as completed
func (s *taskService) Complete(ctx context.Context, taskID int64) (task.Task, error) {
	if taskID <= 0 {
		return task.Task{}, errors.InvalidInput("task ID must be positive")
	}

	// Get the existing task
	existingTask, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return task.Task{}, err
	}

	// Mark as completed and set status to done
	existingTask.IsCompleted = true
	existingTask.Status = task.StatusDone
	existingTask.UpdatedAt = time.Now()

	// Update the task in the repository
	err = s.repo.Update(ctx, existingTask)
	if err != nil {
		return task.Task{}, err
	}

	// Get the updated task
	updatedTask, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return task.Task{}, err
	}

	return updatedTask, nil
}

// ChangeStatus updates the status of a task
func (s *taskService) ChangeStatus(ctx context.Context, taskID int64, status task.Status) (task.Task, error) {
	if taskID <= 0 {
		return task.Task{}, errors.InvalidInput("task ID must be positive")
	}

	if !isValidStatus(status) {
		return task.Task{}, errors.InvalidInput("invalid status")
	}

	// Get the existing task
	existingTask, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return task.Task{}, err
	}

	// Update status and completion based on the new status
	existingTask.Status = status
	if status == task.StatusDone {
		existingTask.IsCompleted = true
	} else {
		existingTask.IsCompleted = false
	}
	existingTask.UpdatedAt = time.Now()

	// Update the task in the repository
	err = s.repo.Update(ctx, existingTask)
	if err != nil {
		return task.Task{}, err
	}

	// Get the updated task
	updatedTask, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return task.Task{}, err
	}

	return updatedTask, nil
}

// ChangePriority updates the priority of a task
func (s *taskService) ChangePriority(ctx context.Context, taskID int64, priority task.Priority) (task.Task, error) {
	if taskID <= 0 {
		return task.Task{}, errors.InvalidInput("task ID must be positive")
	}

	if !isValidPriority(priority) {
		return task.Task{}, errors.InvalidInput("invalid priority")
	}

	// Get the existing task
	existingTask, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return task.Task{}, err
	}

	// Update priority
	existingTask.Priority = priority
	existingTask.UpdatedAt = time.Now()

	// Update the task in the repository
	err = s.repo.Update(ctx, existingTask)
	if err != nil {
		return task.Task{}, err
	}

	// Get the updated task
	updatedTask, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return task.Task{}, err
	}

	return updatedTask, nil
}

// SearchByTitle searches for tasks with titles matching the given pattern
func (s *taskService) SearchByTitle(ctx context.Context, userID int64, titlePattern string) ([]task.Task, error) {
	if userID <= 0 {
		return nil, errors.InvalidInput("user ID must be positive")
	}
	if titlePattern == "" {
		return nil, errors.InvalidInput("search pattern is required")
	}

	return s.repo.SearchTasksByTitle(ctx, userID, titlePattern)
}

// SearchByTag searches for tasks that have the specified tag
func (s *taskService) SearchByTag(ctx context.Context, userID int64, tag string) ([]task.Task, error) {
	if userID <= 0 {
		return nil, errors.InvalidInput("user ID must be positive")
	}
	if tag == "" {
		return nil, errors.InvalidInput("tag is required")
	}

	return s.repo.SearchTasksByTag(ctx, userID, tag)
}

// ListByStatus retrieves all tasks for a user with the given status
func (s *taskService) ListByStatus(ctx context.Context, userID int64, status task.Status) ([]task.Task, error) {
	if userID <= 0 {
		return nil, errors.InvalidInput("user ID must be positive")
	}
	if !isValidStatus(status) {
		return nil, errors.InvalidInput("invalid status")
	}

	return s.repo.ListTasksByStatus(ctx, userID, status)
}

// ListByPriority retrieves all tasks for a user with the given priority
func (s *taskService) ListByPriority(ctx context.Context, userID int64, priority task.Priority) ([]task.Task, error) {
	if userID <= 0 {
		return nil, errors.InvalidInput("user ID must be positive")
	}
	if !isValidPriority(priority) {
		return nil, errors.InvalidInput("invalid priority")
	}

	return s.repo.ListTasksByPriority(ctx, userID, priority)
}

// ListTasksDueToday retrieves all incomplete tasks due on the current day
func (s *taskService) ListTasksDueToday(ctx context.Context, userID int64) ([]task.Task, error) {
	if userID <= 0 {
		return nil, errors.InvalidInput("user ID must be positive")
	}

	return s.repo.ListTasksDueToday(ctx, userID)
}

// ListTasksDueSoon retrieves all incomplete tasks due within the next 7 days
func (s *taskService) ListTasksDueSoon(ctx context.Context, userID int64) ([]task.Task, error) {
	if userID <= 0 {
		return nil, errors.InvalidInput("user ID must be positive")
	}

	return s.repo.ListTasksDueSoon(ctx, userID)
}

// ListOverdueTasks retrieves all incomplete tasks that are past their due date
func (s *taskService) ListOverdueTasks(ctx context.Context, userID int64) ([]task.Task, error) {
	if userID <= 0 {
		return nil, errors.InvalidInput("user ID must be positive")
	}

	return s.repo.ListOverdueTasks(ctx, userID)
}

// GetTaskCountsByStatus retrieves counts of tasks grouped by status
func (s *taskService) GetTaskCountsByStatus(ctx context.Context, userID int64) (repo.TaskStatusCounts, error) {
	if userID <= 0 {
		return repo.TaskStatusCounts{}, errors.InvalidInput("user ID must be positive")
	}

	return s.repo.GetTaskCountsByStatus(ctx, userID)
}

// GetTaskCountsByPriority retrieves counts of incomplete tasks grouped by priority
func (s *taskService) GetTaskCountsByPriority(ctx context.Context, userID int64) (repo.TaskPriorityCounts, error) {
	if userID <= 0 {
		return repo.TaskPriorityCounts{}, errors.InvalidInput("user ID must be positive")
	}

	return s.repo.GetTaskCountsByPriority(ctx, userID)
}

// GetRecentlyCompletedTasks retrieves recently completed tasks, limited by count
func (s *taskService) GetRecentlyCompletedTasks(ctx context.Context, userID int64, limit int) ([]task.Task, error) {
	if userID <= 0 {
		return nil, errors.InvalidInput("user ID must be positive")
	}
	if limit <= 0 {
		return nil, errors.InvalidInput("limit must be positive")
	}

	return s.repo.GetRecentlyCompletedTasks(ctx, userID, int32(limit))
}

// BulkUpdateStatus updates the status of multiple tasks at once
func (s *taskService) BulkUpdateStatus(ctx context.Context, taskIDs []int32, status task.Status) error {
	if len(taskIDs) == 0 {
		return errors.InvalidInput("task IDs list cannot be empty")
	}
	if !isValidStatus(status) {
		return errors.InvalidInput("invalid status")
	}

	// Calculate if tasks should be completed based on status
	isCompleted := status == task.StatusDone

	return s.repo.BulkUpdateTaskStatus(ctx, taskIDs, status, isCompleted)
}

// GetAllTags retrieves all unique tags used by a user
func (s *taskService) GetAllTags(ctx context.Context, userID int64) ([]string, error) {
	if userID <= 0 {
		return nil, errors.InvalidInput("user ID must be positive")
	}

	return s.repo.GetAllTagsForUser(ctx, userID)
}

// Helper functions

// isValidStatus checks if a status is valid
func isValidStatus(status task.Status) bool {
	validStatuses := []task.Status{
		task.StatusTodo,
		task.StatusInProgress,
		task.StatusDone,
	}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}

	return false
}

// isValidPriority checks if a priority is valid
func isValidPriority(priority task.Priority) bool {
	validPriorities := []task.Priority{
		task.PriorityLow,
		task.PriorityMedium,
		task.PriorityHigh,
	}

	for _, validPriority := range validPriorities {
		if priority == validPriority {
			return true
		}
	}

	return false
}
