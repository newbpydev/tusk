// Package services contains business logic for the TUI application.
package services

import (
	"context"
	"time"

	"github.com/newbpydev/tusk/internal/core/task"
	coreTaskService "github.com/newbpydev/tusk/internal/service/task"
)

// CoreTaskServiceAdapter adapts the core task.Service to the TUI TaskServiceInterface
type CoreTaskServiceAdapter struct {
	coreService coreTaskService.Service
}

// NewCoreTaskServiceAdapter creates a new adapter for the core task service
func NewCoreTaskServiceAdapter(coreService coreTaskService.Service) *CoreTaskServiceAdapter {
	return &CoreTaskServiceAdapter{
		coreService: coreService,
	}
}

// List retrieves all tasks for a user
func (a *CoreTaskServiceAdapter) List(ctx context.Context, userID int64) ([]task.Task, error) {
	return a.coreService.List(ctx, userID)
}

// Create creates a new task
func (a *CoreTaskServiceAdapter) Create(
	ctx context.Context,
	userID int64,
	parentID *int64,
	title, description string,
	dueDate *time.Time,
	priority task.Priority,
	tags []string,
) (task.Task, error) {
	return a.coreService.Create(ctx, userID, parentID, title, description, dueDate, priority, tags)
}

// Update updates an existing task
func (a *CoreTaskServiceAdapter) Update(
	ctx context.Context,
	taskID int64,
	title, description string,
	dueDate *time.Time,
	priority task.Priority,
	tags []string,
) (task.Task, error) {
	return a.coreService.Update(ctx, taskID, title, description, dueDate, priority, tags)
}

// Delete deletes a task
func (a *CoreTaskServiceAdapter) Delete(ctx context.Context, taskID int64) error {
	return a.coreService.Delete(ctx, taskID)
}

// Complete marks a task as completed
func (a *CoreTaskServiceAdapter) Complete(ctx context.Context, taskID int64) (task.Task, error) {
	return a.coreService.Complete(ctx, taskID)
}

// ChangeStatus changes the status of a task
func (a *CoreTaskServiceAdapter) ChangeStatus(
	ctx context.Context,
	taskID int64,
	status task.Status,
) (task.Task, error) {
	return a.coreService.ChangeStatus(ctx, taskID, status)
}
