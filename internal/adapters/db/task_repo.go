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
)

// Ensure SQLTaskRepository implements output.TaskRepository interface
var _ output.TaskRepository = (*SQLTaskRepository)(nil)

// SQLTaskRepository implements the output.TaskRepository interface using SQLC and PostgreSQL
type SQLTaskRepository struct {
	q *sqlc.Queries
}

// NewSQLTaskRepository creates a new SQLTaskRepository with the provided connection pool
func NewSQLTaskRepository(pool *pgxpool.Pool) *SQLTaskRepository {
	return &SQLTaskRepository{
		q: sqlc.New(pool),
	}
}

// Create implements output.TaskRepository.Create
func (r *SQLTaskRepository) Create(ctx context.Context, t task.Task) (task.Task, error) {
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
	row, err := r.q.CreateTask(ctx, params)
	if err != nil {
		return task.Task{}, errors.InternalError(fmt.Sprintf("failed to create task: %v", err))
	}

	// Convert result back to domain model
	result := mapDBTaskToDomain(row)
	return result, nil
}

// Update implements output.TaskRepository.Update
func (r *SQLTaskRepository) Update(ctx context.Context, t task.Task) error {
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

	err := r.q.UpdateTask(ctx, params)
	if err != nil {
		if err == pgx.ErrNoRows {
			return errors.NotFound(fmt.Sprintf("task %d not found", t.ID))
		}
		return errors.InternalError(fmt.Sprintf("failed to update task: %v", err))
	}
	return nil
}

// Delete implements output.TaskRepository.Delete
func (r *SQLTaskRepository) Delete(ctx context.Context, id int64) error {
	err := r.q.DeleteTask(ctx, int32(id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return errors.NotFound(fmt.Sprintf("task %d not found", id))
		}
		return errors.InternalError(fmt.Sprintf("failed to delete task: %v", err))
	}
	return nil
}

// GetByID implements output.TaskRepository.GetByID
func (r *SQLTaskRepository) GetByID(ctx context.Context, id int64) (task.Task, error) {
	row, err := r.q.GetTaskById(ctx, int32(id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return task.Task{}, errors.NotFound(fmt.Sprintf("task %d not found", id))
		}
		return task.Task{}, errors.InternalError(fmt.Sprintf("failed to get task: %v", err))
	}
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
	rows, err := r.q.ListTasksWithSubtasksRecursive(ctx, int32(rootID))
	if err != nil {
		return task.Task{}, errors.InternalError(fmt.Sprintf("failed to get task tree: %v", err))
	}

	if len(rows) == 0 {
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
