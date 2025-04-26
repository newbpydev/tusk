package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sqlc "github.com/newbpydev/tusk/internal/adapters/db/sqlc"
	"github.com/newbpydev/tusk/internal/core/errors"
	"github.com/newbpydev/tusk/internal/core/task"
	"github.com/newbpydev/tusk/internal/ports/output"
	"github.com/newbpydev/tusk/internal/util/logging"
	"go.uber.org/zap"
)

// Ensure SQLTaskRepository implements output.TaskRepository interface
var _ output.TaskRepository = (*SQLTaskRepository)(nil)

// SQLTaskRepository implements the output.TaskRepository interface using SQLC and PostgreSQL
type SQLTaskRepository struct {
	q   *sqlc.Queries
	log *zap.Logger
}

// NewSQLTaskRepository creates a new SQLTaskRepository with the provided connection pool
func NewSQLTaskRepository(pool *pgxpool.Pool) *SQLTaskRepository {
	return &SQLTaskRepository{
		q:   sqlc.New(pool),
		log: logging.DBLogger.Named("task_repo"),
	}
}

// Create implements output.TaskRepository.Create
func (r *SQLTaskRepository) Create(ctx context.Context, t task.Task) (task.Task, error) {
	r.log.Debug("Creating new task in database",
		zap.Int32("user_id", t.UserID),
		zap.String("title_length", fmt.Sprintf("%d chars", len(t.Title))),
		zap.Bool("has_parent", t.ParentID != nil),
		zap.Bool("has_description", t.Description != nil),
		zap.Bool("has_due_date", t.DueDate != nil),
		zap.Int("tag_count", len(t.Tags)))

	// Convert domain model to db params
	params := sqlc.CreateTaskParams{
		UserID:   t.UserID,
		ParentID: intPtrToNullInt4(t.ParentID),
		Title:    t.Title,
		Description: pgtype.Text{
			String: stringPtrToString(t.Description),
			Valid:  t.Description != nil,
		},
		DueDate: timePtrToNullTimestamp(t.DueDate),
		IsCompleted: pgtype.Bool{
			Bool:  t.IsCompleted,
			Valid: true,
		},
		Status: pgtype.Text{
			String: string(t.Status),
			Valid:  true,
		},
		Priority: pgtype.Text{
			String: string(t.Priority),
			Valid:  true,
		},
		Tags: tagsToStringSlice(t.Tags),
		DisplayOrder: pgtype.Int4{
			Int32: int32(t.DisplayOrder),
			Valid: true,
		},
	}

	// Execute query
	startTime := time.Now()
	row, err := r.q.CreateTask(ctx, params)
	queryDuration := time.Since(startTime)

	if err != nil {
		r.log.Error("Failed to create task in database",
			zap.Int32("user_id", t.UserID),
			zap.Duration("duration_ms", queryDuration),
			zap.Error(err))
		return task.Task{}, errors.InternalError(fmt.Sprintf("failed to create task: %v", err))
	}

	r.log.Info("Task created successfully in database",
		zap.Int32("task_id", row.ID),
		zap.Int32("user_id", t.UserID),
		zap.Duration("duration_ms", queryDuration))

	// Convert result back to domain model
	result := mapDBTaskToDomain(row)
	return result, nil
}

// Update implements output.TaskRepository.Update
func (r *SQLTaskRepository) Update(ctx context.Context, t task.Task) error {
	r.log.Debug("Updating task in database",
		zap.Int32("task_id", t.ID),
		zap.Int32("user_id", t.UserID),
		zap.String("status", string(t.Status)),
		zap.Bool("is_completed", t.IsCompleted))

	params := sqlc.UpdateTaskParams{
		ID:       int32(t.ID),
		UserID:   t.UserID,
		ParentID: intPtrToNullInt4(t.ParentID),
		Title:    t.Title,
		Description: pgtype.Text{
			String: stringPtrToString(t.Description),
			Valid:  t.Description != nil,
		},
		DueDate: timePtrToNullTimestamp(t.DueDate),
		IsCompleted: pgtype.Bool{
			Bool:  t.IsCompleted,
			Valid: true,
		},
		Status: pgtype.Text{
			String: string(t.Status),
			Valid:  true,
		},
		Priority: pgtype.Text{
			String: string(t.Priority),
			Valid:  true,
		},
		Tags: tagsToStringSlice(t.Tags),
		DisplayOrder: pgtype.Int4{
			Int32: int32(t.DisplayOrder),
			Valid: true,
		},
	}

	startTime := time.Now()
	err := r.q.UpdateTask(ctx, params)
	queryDuration := time.Since(startTime)

	if err != nil {
		if err == pgx.ErrNoRows {
			r.log.Warn("Failed to update task - not found",
				zap.Int32("task_id", t.ID),
				zap.Duration("duration_ms", queryDuration))
			return errors.NotFound(fmt.Sprintf("task %d not found", t.ID))
		}
		r.log.Error("Failed to update task",
			zap.Int32("task_id", t.ID),
			zap.Duration("duration_ms", queryDuration),
			zap.Error(err))
		return errors.InternalError(fmt.Sprintf("failed to update task: %v", err))
	}

	r.log.Info("Task updated successfully",
		zap.Int32("task_id", t.ID),
		zap.Int32("user_id", t.UserID),
		zap.Duration("duration_ms", queryDuration))

	return nil
}

// Delete implements output.TaskRepository.Delete
func (r *SQLTaskRepository) Delete(ctx context.Context, id int64) error {
	r.log.Info("Deleting task",
		zap.Int64("task_id", id))

	startTime := time.Now()
	err := r.q.DeleteTask(ctx, int32(id))
	queryDuration := time.Since(startTime)

	if err != nil {
		if err == pgx.ErrNoRows {
			r.log.Warn("Failed to delete task - not found",
				zap.Int64("task_id", id),
				zap.Duration("duration_ms", queryDuration))
			return errors.NotFound(fmt.Sprintf("task %d not found", id))
		}
		r.log.Error("Failed to delete task",
			zap.Int64("task_id", id),
			zap.Duration("duration_ms", queryDuration),
			zap.Error(err))
		return errors.InternalError(fmt.Sprintf("failed to delete task: %v", err))
	}

	r.log.Info("Task deleted successfully",
		zap.Int64("task_id", id),
		zap.Duration("duration_ms", queryDuration))

	return nil
}

// GetByID implements output.TaskRepository.GetByID
func (r *SQLTaskRepository) GetByID(ctx context.Context, id int64) (task.Task, error) {
	r.log.Debug("Fetching task by ID",
		zap.Int64("task_id", id))

	startTime := time.Now()
	row, err := r.q.GetTaskById(ctx, int32(id))
	queryDuration := time.Since(startTime)

	if err != nil {
		if err == pgx.ErrNoRows {
			r.log.Warn("Task not found",
				zap.Int64("task_id", id),
				zap.Duration("duration_ms", queryDuration))
			return task.Task{}, errors.NotFound(fmt.Sprintf("task %d not found", id))
		}
		r.log.Error("Failed to get task by ID",
			zap.Int64("task_id", id),
			zap.Duration("duration_ms", queryDuration),
			zap.Error(err))
		return task.Task{}, errors.InternalError(fmt.Sprintf("failed to get task: %v", err))
	}

	r.log.Debug("Task fetched successfully",
		zap.Int64("task_id", id),
		zap.Int32("user_id", row.UserID),
		zap.Duration("duration_ms", queryDuration))

	return mapDBTaskToDomain(row), nil
}

// ListRootTasks implements output.TaskRepository.ListRootTasks
func (r *SQLTaskRepository) ListRootTasks(ctx context.Context, userID int64) ([]task.Task, error) {
	rows, err := r.q.ListRootTasksByUserId(ctx, int32(userID))
	if err != nil {
		return nil, errors.InternalError(fmt.Sprintf("failed to list root tasks: %v", err))
	}

	tasks := make([]task.Task, len(rows))
	for i, row := range rows {
		tasks[i] = mapDBTaskToDomain(row)
	}
	return tasks, nil
}

// ListSubTasks implements output.TaskRepository.ListSubTasks
func (r *SQLTaskRepository) ListSubTasks(ctx context.Context, parentID int64) ([]task.Task, error) {
	pid := pgtype.Int4{
		Int32: int32(parentID),
		Valid: true,
	}
	rows, err := r.q.GetSubtasksByParentId(ctx, pid)
	if err != nil {
		return nil, errors.InternalError(fmt.Sprintf("failed to list subtasks: %v", err))
	}

	tasks := make([]task.Task, len(rows))
	for i, row := range rows {
		tasks[i] = mapDBTaskToDomain(row)
	}
	return tasks, nil
}

// GetTaskTree implements output.TaskRepository.GetTaskTree
func (r *SQLTaskRepository) GetTaskTree(ctx context.Context, rootID int64) (task.Task, error) {
	r.log.Debug("Fetching task tree",
		zap.Int64("root_id", rootID))

	startTime := time.Now()
	rows, err := r.q.ListTasksWithSubtasksRecursive(ctx, int32(rootID))
	queryDuration := time.Since(startTime)

	if err != nil {
		r.log.Error("Failed to get task tree",
			zap.Int64("root_id", rootID),
			zap.Duration("duration_ms", queryDuration),
			zap.Error(err))
		return task.Task{}, errors.InternalError(fmt.Sprintf("failed to get task tree: %v", err))
	}

	if len(rows) == 0 {
		r.log.Warn("Task tree not found",
			zap.Int64("root_id", rootID),
			zap.Duration("duration_ms", queryDuration))
		return task.Task{}, errors.NotFound(fmt.Sprintf("task %d not found", rootID))
	}

	// Map rows to domain tasks
	domainTasks := make([]task.Task, len(rows))
	for i, row := range rows {
		domainTasks[i] = mapRecursiveRowToDomain(row)
	}

	// Build tree
	tree := buildTaskTree(domainTasks)

	// Compute metrics
	computeTaskMetrics(&tree)

	r.log.Debug("Task tree fetched successfully",
		zap.Int64("root_id", rootID),
		zap.Int("node_count", len(rows)),
		zap.Duration("duration_ms", queryDuration))

	return tree, nil
}

// ReorderTask implements output.TaskRepository.ReorderTask
func (r *SQLTaskRepository) ReorderTask(ctx context.Context, taskID int64, newOrder int) error {
	params := sqlc.ReorderTaskParams{
		ID: int32(taskID),
		DisplayOrder: pgtype.Int4{
			Int32: int32(newOrder),
			Valid: true,
		},
	}

	err := r.q.ReorderTask(ctx, params)
	if err != nil {
		if err == pgx.ErrNoRows {
			return errors.NotFound(fmt.Sprintf("task %d not found", taskID))
		}
		return errors.InternalError(fmt.Sprintf("failed to reorder task: %v", err))
	}
	return nil
}

// SearchTasksByTitle implements output.TaskRepository.SearchTasksByTitle
func (r *SQLTaskRepository) SearchTasksByTitle(ctx context.Context, userID int64, titlePattern string) ([]task.Task, error) {
	rows, err := r.q.SearchTasksByTitle(ctx, sqlc.SearchTasksByTitleParams{
		UserID: int32(userID),
		Title:  fmt.Sprintf("%%%s%%", titlePattern), // Add wildcards for ILIKE
	})
	if err != nil {
		return nil, errors.InternalError(fmt.Sprintf("failed to search tasks by title: %v", err))
	}

	tasks := make([]task.Task, len(rows))
	for i, row := range rows {
		tasks[i] = mapDBTaskToDomain(row)
	}
	return tasks, nil
}

// SearchTasksByTag implements output.TaskRepository.SearchTasksByTag
func (r *SQLTaskRepository) SearchTasksByTag(ctx context.Context, userID int64, tag string) ([]task.Task, error) {
	rows, err := r.q.SearchTasksByTag(ctx, sqlc.SearchTasksByTagParams{
		UserID: int32(userID),
		Tags:   []string{tag}, // Changed from Tag to Tags and passing as a slice
	})
	if err != nil {
		return nil, errors.InternalError(fmt.Sprintf("failed to search tasks by tag: %v", err))
	}

	tasks := make([]task.Task, len(rows))
	for i, row := range rows {
		tasks[i] = mapDBTaskToDomain(row)
	}
	return tasks, nil
}

// ListTasksByStatus implements output.TaskRepository.ListTasksByStatus
func (r *SQLTaskRepository) ListTasksByStatus(ctx context.Context, userID int64, status task.Status) ([]task.Task, error) {
	rows, err := r.q.ListTasksByStatus(ctx, sqlc.ListTasksByStatusParams{
		UserID: int32(userID),
		Status: pgtype.Text{
			String: string(status),
			Valid:  true,
		},
	})
	if err != nil {
		return nil, errors.InternalError(fmt.Sprintf("failed to list tasks by status: %v", err))
	}

	tasks := make([]task.Task, len(rows))
	for i, row := range rows {
		tasks[i] = mapDBTaskToDomain(row)
	}
	return tasks, nil
}

// ListTasksByPriority implements output.TaskRepository.ListTasksByPriority
func (r *SQLTaskRepository) ListTasksByPriority(ctx context.Context, userID int64, priority task.Priority) ([]task.Task, error) {
	rows, err := r.q.ListTasksByPriority(ctx, sqlc.ListTasksByPriorityParams{
		UserID: int32(userID),
		Priority: pgtype.Text{
			String: string(priority),
			Valid:  true,
		},
	})
	if err != nil {
		return nil, errors.InternalError(fmt.Sprintf("failed to list tasks by priority: %v", err))
	}

	tasks := make([]task.Task, len(rows))
	for i, row := range rows {
		tasks[i] = mapDBTaskToDomain(row)
	}
	return tasks, nil
}

// ListTasksDueToday implements output.TaskRepository.ListTasksDueToday
func (r *SQLTaskRepository) ListTasksDueToday(ctx context.Context, userID int64) ([]task.Task, error) {
	r.log.Debug("Listing tasks due today",
		zap.Int64("user_id", userID))

	startTime := time.Now()
	rows, err := r.q.ListTasksDueToday(ctx, int32(userID))
	queryDuration := time.Since(startTime)

	if err != nil {
		r.log.Error("Failed to list tasks due today",
			zap.Int64("user_id", userID),
			zap.Duration("duration_ms", queryDuration),
			zap.Error(err))
		return nil, errors.InternalError(fmt.Sprintf("failed to list tasks due today: %v", err))
	}

	tasks := make([]task.Task, len(rows))
	for i, row := range rows {
		tasks[i] = mapDBTaskToDomain(row)
	}

	r.log.Debug("Successfully listed tasks due today",
		zap.Int64("user_id", userID),
		zap.Int("task_count", len(tasks)),
		zap.Duration("duration_ms", queryDuration))

	return tasks, nil
}

// ListTasksDueSoon implements output.TaskRepository.ListTasksDueSoon
func (r *SQLTaskRepository) ListTasksDueSoon(ctx context.Context, userID int64) ([]task.Task, error) {
	rows, err := r.q.ListTasksDueSoon(ctx, int32(userID))
	if err != nil {
		return nil, errors.InternalError(fmt.Sprintf("failed to list tasks due soon: %v", err))
	}

	tasks := make([]task.Task, len(rows))
	for i, row := range rows {
		tasks[i] = mapDBTaskToDomain(row)
	}
	return tasks, nil
}

// ListOverdueTasks implements output.TaskRepository.ListOverdueTasks
func (r *SQLTaskRepository) ListOverdueTasks(ctx context.Context, userID int64) ([]task.Task, error) {
	rows, err := r.q.ListOverdueTasks(ctx, int32(userID))
	if err != nil {
		return nil, errors.InternalError(fmt.Sprintf("failed to list overdue tasks: %v", err))
	}

	tasks := make([]task.Task, len(rows))
	for i, row := range rows {
		tasks[i] = mapDBTaskToDomain(row)
	}
	return tasks, nil
}

// GetTaskCountsByStatus implements output.TaskRepository.GetTaskCountsByStatus
func (r *SQLTaskRepository) GetTaskCountsByStatus(ctx context.Context, userID int64) (output.TaskStatusCounts, error) {
	row, err := r.q.GetTaskCountsByStatus(ctx, int32(userID))
	if err != nil {
		return output.TaskStatusCounts{}, errors.InternalError(fmt.Sprintf("failed to get task counts by status: %v", err))
	}

	return output.TaskStatusCounts{
		TodoCount:       int(row.TodoCount),
		InProgressCount: int(row.InProgressCount),
		DoneCount:       int(row.DoneCount),
		TotalCount:      int(row.TotalCount),
	}, nil
}

// GetTaskCountsByPriority implements output.TaskRepository.GetTaskCountsByPriority
func (r *SQLTaskRepository) GetTaskCountsByPriority(ctx context.Context, userID int64) (output.TaskPriorityCounts, error) {
	row, err := r.q.GetTaskCountsByPriority(ctx, int32(userID))
	if err != nil {
		return output.TaskPriorityCounts{}, errors.InternalError(fmt.Sprintf("failed to get task counts by priority: %v", err))
	}

	return output.TaskPriorityCounts{
		LowCount:    int(row.LowPriorityCount),
		MediumCount: int(row.MediumPriorityCount),
		HighCount:   int(row.HighPriorityCount),
	}, nil
}

// GetRecentlyCompletedTasks implements output.TaskRepository.GetRecentlyCompletedTasks
func (r *SQLTaskRepository) GetRecentlyCompletedTasks(ctx context.Context, userID int64, limit int32) ([]task.Task, error) {
	rows, err := r.q.GetRecentlyCompletedTasks(ctx, sqlc.GetRecentlyCompletedTasksParams{
		UserID: int32(userID),
		Limit:  limit,
	})
	if err != nil {
		return nil, errors.InternalError(fmt.Sprintf("failed to get recently completed tasks: %v", err))
	}

	tasks := make([]task.Task, len(rows))
	for i, row := range rows {
		tasks[i] = mapDBTaskToDomain(row)
	}
	return tasks, nil
}

// BulkUpdateTaskStatus implements output.TaskRepository.BulkUpdateTaskStatus
func (r *SQLTaskRepository) BulkUpdateTaskStatus(ctx context.Context, taskIDs []int32, status task.Status, isCompleted bool) error {
	r.log.Info("Bulk updating task status",
		zap.Int("task_count", len(taskIDs)),
		zap.String("new_status", string(status)),
		zap.Bool("is_completed", isCompleted))

	startTime := time.Now()
	err := r.q.BulkUpdateTaskStatus(ctx, sqlc.BulkUpdateTaskStatusParams{
		Column1: taskIDs,
		Status: pgtype.Text{
			String: string(status),
			Valid:  true,
		},
		IsCompleted: pgtype.Bool{
			Bool:  isCompleted,
			Valid: true,
		},
	})
	queryDuration := time.Since(startTime)

	if err != nil {
		r.log.Error("Failed to bulk update task status",
			zap.Int("task_count", len(taskIDs)),
			zap.String("status", string(status)),
			zap.Duration("duration_ms", queryDuration),
			zap.Error(err))
		return errors.InternalError(fmt.Sprintf("failed to bulk update task status: %v", err))
	}

	r.log.Info("Bulk task status update successful",
		zap.Int("task_count", len(taskIDs)),
		zap.String("new_status", string(status)),
		zap.Duration("duration_ms", queryDuration))

	return nil
}

// GetAllTagsForUser implements output.TaskRepository.GetAllTagsForUser
func (r *SQLTaskRepository) GetAllTagsForUser(ctx context.Context, userID int64) ([]string, error) {
	rows, err := r.q.GetAllTagsForUser(ctx, int32(userID))
	if err != nil {
		return nil, errors.InternalError(fmt.Sprintf("failed to get all tags for user: %v", err))
	}

	// Convert the interface{} values to strings
	tags := make([]string, 0, len(rows))
	for _, row := range rows {
		// Convert the interface{} to string
		if str, ok := row.(string); ok {
			tags = append(tags, str)
		}
	}

	return tags, nil
}

// Mapping functions

// mapDBTaskToDomain maps a sqlc.Task to a task.Task
func mapDBTaskToDomain(dbt sqlc.Task) task.Task {
	return task.Task{
		ID:           int32(dbt.ID),
		UserID:       dbt.UserID,
		ParentID:     nullInt4ToIntPtr(dbt.ParentID),
		Title:        dbt.Title,
		Description:  nullTextToStringPtr(dbt.Description),
		CreatedAt:    dbt.CreatedAt.Time,
		UpdatedAt:    dbt.UpdatedAt.Time,
		DueDate:      nullTimestampToTimePtr(dbt.DueDate),
		IsCompleted:  dbt.IsCompleted.Bool,
		Status:       task.Status(dbt.Status.String),
		Priority:     task.Priority(dbt.Priority.String),
		Tags:         stringSliceToTags(dbt.Tags),
		DisplayOrder: int(dbt.DisplayOrder.Int32),
	}
}

// mapRecursiveRowToDomain maps a sqlc.ListTasksWithSubtasksRecursiveRow to a task.Task
func mapRecursiveRowToDomain(row sqlc.ListTasksWithSubtasksRecursiveRow) task.Task {
	return task.Task{
		ID:           int32(row.ID),
		UserID:       row.UserID,
		ParentID:     nullInt4ToIntPtr(row.ParentID),
		Title:        row.Title,
		Description:  nullTextToStringPtr(row.Description),
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
		DueDate:      nullTimestampToTimePtr(row.DueDate),
		IsCompleted:  row.IsCompleted.Bool,
		Status:       task.Status(row.Status.String),
		Priority:     task.Priority(row.Priority.String),
		Tags:         stringSliceToTags(row.Tags),
		DisplayOrder: int(row.DisplayOrder.Int32),
		SubTasks:     []task.Task{}, // Initialize empty slice for subtasks
	}
}

// buildTaskTree builds a task tree from a flat list of tasks
func buildTaskTree(tasks []task.Task) task.Task {
	// Create a map of task ID to task pointer for quick lookups
	taskMap := make(map[int32]*task.Task, len(tasks))
	var root task.Task

	// First pass: create map entries
	for i := range tasks {
		taskMap[tasks[i].ID] = &tasks[i]
		if i == 0 {
			root = tasks[i] // The first task is the root
		}
	}

	// Second pass: connect parents and children
	for i := range tasks {
		// Skip the root task
		if tasks[i].ParentID == nil {
			continue
		}

		// Find the parent and append this task as a child
		if parent, found := taskMap[*tasks[i].ParentID]; found {
			parent.SubTasks = append(parent.SubTasks, tasks[i])
		}
	}

	return root
}

// computeTaskMetrics recursively computes metrics for a task and its subtasks
func computeTaskMetrics(t *task.Task) {
	totalCount := len(t.SubTasks)
	completedCount := 0

	// Process subtasks recursively
	for i := range t.SubTasks {
		computeTaskMetrics(&t.SubTasks[i])
		totalCount += t.SubTasks[i].TotalCount
		completedCount += t.SubTasks[i].CompletedCount
	}

	// Count completed tasks
	for _, subtask := range t.SubTasks {
		if subtask.IsCompleted {
			completedCount++
		}
	}

	// Update metrics
	t.TotalCount = totalCount
	t.CompletedCount = completedCount

	// Calculate progress (avoid division by zero)
	if totalCount > 0 {
		t.Progress = float64(completedCount) / float64(totalCount)
	} else {
		t.Progress = 0
	}
}

// Type conversion helper functions

// nullInt4ToIntPtr converts pgtype.Int4 to *int32
func nullInt4ToIntPtr(n pgtype.Int4) *int32 {
	if !n.Valid {
		return nil
	}
	return &n.Int32
}

// intPtrToNullInt4 converts *int32 to pgtype.Int4
func intPtrToNullInt4(i *int32) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: *i, Valid: true}
}

// nullTextToStringPtr converts pgtype.Text to *string
func nullTextToStringPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

// stringPtrToString safely dereferences a string pointer
func stringPtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// nullTimestampToTimePtr converts pgtype.Timestamp to *time.Time
func nullTimestampToTimePtr(ts pgtype.Timestamp) *time.Time {
	if !ts.Valid {
		return nil
	}
	return &ts.Time
}

// timePtrToNullTimestamp converts *time.Time to pgtype.Timestamp
func timePtrToNullTimestamp(t *time.Time) pgtype.Timestamp {
	if t == nil {
		return pgtype.Timestamp{Valid: false}
	}
	return pgtype.Timestamp{Time: *t, Valid: true}
}

// stringSliceToTags converts a slice of strings to a slice of task.Tag
func stringSliceToTags(ss []string) []task.Tag {
	tags := make([]task.Tag, len(ss))
	for i, s := range ss {
		tags[i] = task.Tag{Name: s}
	}
	return tags
}

// tagsToStringSlice converts a slice of task.Tag to a slice of strings
func tagsToStringSlice(tags []task.Tag) []string {
	ss := make([]string, len(tags))
	for i, tag := range tags {
		ss[i] = tag.Name
	}
	return ss
}
