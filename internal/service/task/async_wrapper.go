package task

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/newbpydev/tusk/internal/core/task"
	"github.com/newbpydev/tusk/internal/ports/output"
	"github.com/newbpydev/tusk/internal/util/worker"
	"go.uber.org/zap"
)

// AsyncTaskService wraps the regular task service with asynchronous capabilities
type AsyncTaskService struct {
	taskService Service
	workerPool  *worker.Pool
	log         *zap.Logger
	cache       sync.Map // Used to cache recent operations for faster UI feedback
}

// NewAsyncTaskService creates a new async task service that wraps a regular task service
func NewAsyncTaskService(taskService Service, logger *zap.Logger) *AsyncTaskService {
	as := &AsyncTaskService{
		taskService: taskService,
		workerPool:  worker.NewPool(10), // 10 concurrent workers for better performance
		log:         logger.Named("async_task_service"),
	}

	// Start the worker pool
	as.workerPool.Start()

	// Set up result handler
	as.workerPool.CollectResults(func(err error) {
		if err != nil {
			as.log.Error("Background task failed", zap.Error(err))
		}
	})

	return as
}

// List retrieves all tasks for a user
func (s *AsyncTaskService) List(ctx context.Context, userID int64) ([]task.Task, error) {
	// First check if we have tasks in cache for this user
	if cachedTasks, ok := s.cache.Load("user_tasks_" + fmt.Sprintf("%d", userID)); ok {
		// Use the cached tasks while refreshing in the background
		tasks := cachedTasks.([]task.Task)

		// Refresh in background unless the context is already about to expire
		select {
		case <-ctx.Done():
			// Context is already done, don't start background refresh
		default:
			s.workerPool.Submit(func() error {
				bgCtx := context.Background()
				freshTasks, err := s.taskService.List(bgCtx, userID)
				if err == nil {
					s.cache.Store("user_tasks_"+fmt.Sprintf("%d", userID), freshTasks)
				}
				return err
			})
		}

		return tasks, nil
	}

	// If no cache hit, do the normal synchronous operation
	tasks, err := s.taskService.List(ctx, userID)
	if err == nil {
		// Cache the results for future use
		s.cache.Store("user_tasks_"+fmt.Sprintf("%d", userID), tasks)
	}
	return tasks, err
}

// Create creates a new task
func (s *AsyncTaskService) Create(
	ctx context.Context,
	userID int64,
	parentID *int64,
	title, description string,
	dueDate *time.Time,
	priority task.Priority,
	tags []string,
) (task.Task, error) {
	// Perform the synchronous operation first for immediate feedback
	createdTask, err := s.taskService.Create(ctx, userID, parentID, title, description, dueDate, priority, tags)
	if err != nil {
		return task.Task{}, err
	}

	// Cache the result
	s.cacheTask(createdTask)

	// Invalidate user task list cache to force refresh on next list fetch
	s.cache.Delete("user_tasks_" + fmt.Sprintf("%d", userID))

	// Submit background job to ensure all associated data is properly updated
	s.workerPool.Submit(func() error {
		refreshCtx := context.Background()
		_, err := s.taskService.Show(refreshCtx, int64(createdTask.ID))
		if err != nil {
			s.log.Error("Failed to refresh created task data",
				zap.Int32("task_id", createdTask.ID),
				zap.Error(err))
			return err
		}
		return nil
	})

	return createdTask, nil
}

// Update updates an existing task with the given parameters
func (s *AsyncTaskService) Update(
	ctx context.Context,
	taskID int64,
	title, description string,
	dueDate *time.Time,
	priority task.Priority,
	tags []string,
) (task.Task, error) {
	// Get task to determine its user ID for cache invalidation
	var userID int64
	if cachedTask, ok := s.cache.Load(taskID); ok {
		t := cachedTask.(task.Task)
		userID = int64(t.UserID)
	}

	updatedTask, err := s.taskService.Update(ctx, taskID, title, description, dueDate, priority, tags)
	if err != nil {
		return task.Task{}, err
	}

	// Cache the result
	s.cacheTask(updatedTask)

	// Invalidate user task list cache
	if userID > 0 {
		s.cache.Delete("user_tasks_" + fmt.Sprintf("%d", userID))
	}

	// Submit background job for any related updates
	s.workerPool.Submit(func() error {
		bgCtx := context.Background()

		// If userID wasn't found in cache, get it from the updated task
		if userID == 0 {
			userID = int64(updatedTask.UserID)
		}

		// Refresh the user's task list in the background
		if userID > 0 {
			_, err := s.taskService.List(bgCtx, userID)
			if err != nil {
				s.log.Error("Failed to refresh task list after update",
					zap.Int64("user_id", userID),
					zap.Error(err))
				return err
			}
		}
		return nil
	})

	return updatedTask, nil
}

// Delete removes a task
func (s *AsyncTaskService) Delete(ctx context.Context, taskID int64) error {
	// First get the task to determine the user ID for cache invalidation
	var userID int64
	if cachedTask, ok := s.cache.Load(taskID); ok {
		t := cachedTask.(task.Task)
		userID = int64(t.UserID)
	} else {
		// Try to fetch the task first to get its user ID
		t, err := s.taskService.Show(ctx, taskID)
		if err == nil {
			userID = int64(t.UserID)
		}
	}

	// Delete the task
	err := s.taskService.Delete(ctx, taskID)
	if err != nil {
		return err
	}

	// Remove from cache if deletion was successful
	s.cache.Delete(taskID)

	// Invalidate user task list cache
	if userID > 0 {
		s.cache.Delete("user_tasks_" + fmt.Sprintf("%d", userID))
	}

	// Background refresh of user's task list if we know the user ID
	if userID > 0 {
		s.workerPool.Submit(func() error {
			bgCtx := context.Background()
			_, err := s.taskService.List(bgCtx, userID)
			return err
		})
	}

	return nil
}

// Complete marks a task as completed
func (s *AsyncTaskService) Complete(ctx context.Context, taskID int64) (task.Task, error) {
	// Perform the synchronous operation first for immediate feedback
	completedTask, err := s.taskService.Complete(ctx, taskID)
	if err != nil {
		return task.Task{}, err
	}

	// Cache the result
	s.cacheTask(completedTask)

	// Invalidate user task list cache
	userID := int64(completedTask.UserID)
	s.cache.Delete("user_tasks_" + fmt.Sprintf("%d", userID))

	// Submit background job to ensure all associated data is properly updated
	s.workerPool.Submit(func() error {
		refreshCtx := context.Background()
		_, err := s.taskService.Show(refreshCtx, int64(completedTask.ID))
		if err != nil {
			s.log.Error("Failed to refresh completed task data",
				zap.Int32("task_id", completedTask.ID),
				zap.Error(err))
			return err
		}

		// Also refresh the user's task list
		_, err = s.taskService.List(refreshCtx, userID)
		if err != nil {
			s.log.Error("Failed to refresh task list after completion",
				zap.Int64("user_id", userID),
				zap.Error(err))
			return err
		}

		return nil
	})

	return completedTask, nil
}

// ChangeStatus updates the status of a task
func (s *AsyncTaskService) ChangeStatus(ctx context.Context, taskID int64, status task.Status) (task.Task, error) {
	// First fetch the current task to check if it's already cached
	var cachedTask task.Task
	var userID int64

	if taskCache, ok := s.cache.Load(taskID); ok {
		cachedTask = taskCache.(task.Task)
		// If we already have the task cached and it has the requested status, return it immediately
		if cachedTask.Status == status {
			return cachedTask, nil
		}
		userID = int64(cachedTask.UserID)
	}

	// Perform the synchronous operation for immediate feedback
	updatedTask, err := s.taskService.ChangeStatus(ctx, taskID, status)
	if err != nil {
		return task.Task{}, err
	}

	// Cache the result
	s.cacheTask(updatedTask)

	// If we didn't know the user ID before, get it now
	if userID == 0 {
		userID = int64(updatedTask.UserID)
	}

	// Invalidate user task list cache
	s.cache.Delete("user_tasks_" + fmt.Sprintf("%d", userID))

	// Submit background job to ensure changes are properly propagated
	s.workerPool.Submit(func() error {
		refreshCtx := context.Background()
		freshTask, err := s.taskService.Show(refreshCtx, taskID)
		if err != nil {
			s.log.Error("Failed to refresh task status",
				zap.Int64("task_id", taskID),
				zap.String("status", string(status)),
				zap.Error(err))
			return err
		}

		// Update cache with fresh task data
		s.cacheTask(freshTask)

		// Also refresh the user's task list
		_, err = s.taskService.List(refreshCtx, userID)
		if err != nil {
			s.log.Error("Failed to refresh task list after status change",
				zap.Int64("user_id", userID),
				zap.Error(err))
			return err
		}

		return nil
	})

	return updatedTask, nil
}

// Other methods with similar implementations

// GetByID retrieves a task by ID, utilizing the cache when possible
func (s *AsyncTaskService) GetByID(ctx context.Context, taskID int64) (task.Task, error) {
	if taskCache, ok := s.cache.Load(taskID); ok {
		return taskCache.(task.Task), nil
	}

	t, err := s.taskService.Show(ctx, taskID)
	if err == nil {
		s.cacheTask(t)
	}
	return t, err
}

// Show retrieves a task by ID, utilizing cache when possible
func (s *AsyncTaskService) Show(ctx context.Context, taskID int64) (task.Task, error) {
	// Try cache first
	if t, ok := s.cache.Load(taskID); ok {
		return t.(task.Task), nil
	}
	// Delegate to underlying service
	result, err := s.taskService.Show(ctx, taskID)
	if err == nil {
		s.cacheTask(result)
	}
	return result, err
}

// Helper functions

// cacheTask stores a task in the cache
func (s *AsyncTaskService) cacheTask(t task.Task) {
	// Cache by ID for direct lookups
	s.cache.Store(int64(t.ID), t)
}

// Close shuts down the worker pool
func (s *AsyncTaskService) Close() {
	if s.workerPool != nil {
		s.workerPool.Stop()
	}
}

// Pass-through methods to underlying service
// These methods could be enhanced with caching and background operations as needed

func (s *AsyncTaskService) ChangePriority(ctx context.Context, taskID int64, priority task.Priority) (task.Task, error) {
	return s.taskService.ChangePriority(ctx, taskID, priority)
}

func (s *AsyncTaskService) Reorder(ctx context.Context, taskID int64, newOrder int) error {
	return s.taskService.Reorder(ctx, taskID, newOrder)
}

func (s *AsyncTaskService) SearchByTitle(ctx context.Context, userID int64, titlePattern string) ([]task.Task, error) {
	return s.taskService.SearchByTitle(ctx, userID, titlePattern)
}

func (s *AsyncTaskService) SearchByTag(ctx context.Context, userID int64, tag string) ([]task.Task, error) {
	return s.taskService.SearchByTag(ctx, userID, tag)
}

func (s *AsyncTaskService) ListByStatus(ctx context.Context, userID int64, status task.Status) ([]task.Task, error) {
	return s.taskService.ListByStatus(ctx, userID, status)
}

func (s *AsyncTaskService) ListByPriority(ctx context.Context, userID int64, priority task.Priority) ([]task.Task, error) {
	return s.taskService.ListByPriority(ctx, userID, priority)
}

func (s *AsyncTaskService) ListTasksDueToday(ctx context.Context, userID int64) ([]task.Task, error) {
	return s.taskService.ListTasksDueToday(ctx, userID)
}

func (s *AsyncTaskService) ListTasksDueSoon(ctx context.Context, userID int64) ([]task.Task, error) {
	return s.taskService.ListTasksDueSoon(ctx, userID)
}

func (s *AsyncTaskService) ListOverdueTasks(ctx context.Context, userID int64) ([]task.Task, error) {
	return s.taskService.ListOverdueTasks(ctx, userID)
}

func (s *AsyncTaskService) GetTaskCountsByStatus(ctx context.Context, userID int64) (output.TaskStatusCounts, error) {
	return s.taskService.GetTaskCountsByStatus(ctx, userID)
}

func (s *AsyncTaskService) GetTaskCountsByPriority(ctx context.Context, userID int64) (output.TaskPriorityCounts, error) {
	return s.taskService.GetTaskCountsByPriority(ctx, userID)
}

func (s *AsyncTaskService) GetRecentlyCompletedTasks(ctx context.Context, userID int64, limit int) ([]task.Task, error) {
	return s.taskService.GetRecentlyCompletedTasks(ctx, userID, limit)
}

func (s *AsyncTaskService) BulkUpdateStatus(ctx context.Context, taskIDs []int32, status task.Status) error {
	err := s.taskService.BulkUpdateStatus(ctx, taskIDs, status)

	// If successful, refresh the tasks in the background
	if err == nil {
		s.workerPool.Submit(func() error {
			return nil // Basic implementation, could be enhanced to refresh those tasks
		})
	}

	return err
}

func (s *AsyncTaskService) GetAllTags(ctx context.Context, userID int64) ([]string, error) {
	return s.taskService.GetAllTags(ctx, userID)
}
