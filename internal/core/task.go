package core

import (
	"strings"
	"time"
)

type Task struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Status      Status     `json:"status"`
	Priority    Priority   `json:"priority"`
	ParentID    *string    `json:"parent_id,omitempty"`
	Progress    int        `json:"progress"` // 0 - 100
	Tags        []Tag      `json:"tags,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type NewTaskParams struct {
	ID          string
	Title       string
	Description string
	Priority    Priority
	ParentID    *string
	Tags        []string
	DueDate     *time.Time
	Now         time.Time
}

func NewTask(params NewTaskParams) (*Task, error) {
	title := strings.TrimSpace(params.Title)
	if title == "" {
		return nil, ErrEmptyTitle
	}
	if len(title) > 255 {
		return nil, ErrTitleTooLong
	}

	if params.ParentID != nil && *params.ParentID == params.ID {
		return nil, ErrSelfParenting
	}

	priority := params.Priority
	if !priority.IsValid() {
		priority = PriorityMedium
	}

	now := params.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}

	tags, err := NormalizeTags(params.Tags)
	if err != nil {
		return nil, err
	}

	task := &Task{
		ID:          params.ID,
		Title:       title,
		Description: strings.TrimSpace(params.Description),
		Status:      StatusTodo,
		Priority:    priority,
		ParentID:    params.ParentID,
		Progress:    0,
		Tags:        tags,
		DueDate:     params.DueDate,
		CreatedAt:   now,
		UpdatedAt:   now,
		CompletedAt: nil,
	}

	return task, nil
}

func (t *Task) TransitionTo(next Status, now time.Time) error {
	if !t.Status.CanTransitionTo(next) {
		return ErrInvalidStatusTransition
	}
	if t.Status == next {
		return nil
	}

	if now.IsZero() {
		now = time.Now().UTC()
	}

	t.Status = next
	t.UpdatedAt = now

	if next == StatusDone {
		completed := now
		t.CompletedAt = &completed
		t.Progress = 100
	} else if t.CompletedAt != nil {
		t.CompletedAt = nil
	}

	return nil
}

func (t *Task) Update(title, desc string, priority Priority, tags []Tag, dueDate *time.Time, now time.Time) error {
	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle == "" {
		return ErrEmptyTitle
	}
	if len(trimmedTitle) > 255 {
		return ErrTitleTooLong
	}
	if !priority.IsValid() {
		return ErrInvalidPriority
	}

	if now.IsZero() {
		now = time.Now().UTC()
	}

	t.Title = trimmedTitle
	t.Description = strings.TrimSpace(desc)
	t.Priority = priority
	t.Tags = tags
	t.DueDate = dueDate
	t.UpdatedAt = now

	return nil
}

func (t *Task) SetParent(parentID *string, now time.Time) error {
	if parentID != nil && *parentID == t.ID {
		return ErrSelfParenting
	}

	if now.IsZero() {
		now = time.Now().UTC()
	}

	t.ParentID = parentID
	t.UpdatedAt = now
	return nil
}

func (t *Task) SetProgress(progress int, now time.Time) error {
	if progress < 0 || progress > 100 {
		return ErrInvalidProgress
	}

	if now.IsZero() {
		now = time.Now().UTC()
	}

	t.Progress = progress
	t.UpdatedAt = now
	return nil
}

func (t *Task) IsRoot() bool {
	return t.ParentID == nil
}

func (t *Task) IsDone() bool {
	return t.Status == StatusDone
}
