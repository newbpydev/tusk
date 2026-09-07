package core

import (
	"strings"
	"time"
	"unicode/utf8"
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
	id := strings.TrimSpace(params.ID)
	if id == "" {
		return nil, ErrInvalidTaskID
	}

	title := strings.TrimSpace(params.Title)
	if title == "" {
		return nil, ErrEmptyTitle
	}
	if utf8.RuneCountInString(title) > 255 {
		return nil, ErrTitleTooLong
	}

	var parentID *string
	if params.ParentID != nil {
		p := strings.TrimSpace(*params.ParentID)
		if p == "" {
			return nil, ErrInvalidTaskID
		}
		if p == id {
			return nil, ErrSelfParenting
		}
		parentID = &p
	}

	priority := params.Priority
	if priority == 0 {
		priority = PriorityMedium
	} else if !priority.IsValid() {
		return nil, ErrInvalidPriority
	}

	now := params.Now
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}

	tags, err := NormalizeTags(params.Tags)
	if err != nil {
		return nil, err
	}

	var dueDate *time.Time
	if params.DueDate != nil {
		d := params.DueDate.UTC()
		dueDate = &d
	}

	task := &Task{
		ID:          id,
		Title:       title,
		Description: params.Description,
		Status:      StatusTodo,
		Priority:    priority,
		ParentID:    parentID,
		Progress:    0,
		Tags:        tags,
		DueDate:     dueDate,
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
	} else {
		now = now.UTC()
	}

	t.Status = next
	t.UpdatedAt = now

	if next == StatusDone {
		completed := now
		t.CompletedAt = &completed
		t.Progress = 100
	} else if t.CompletedAt != nil {
		t.CompletedAt = nil
		t.Progress = 0
	}

	return nil
}

func (t *Task) Update(title, desc string, priority Priority, tags []Tag, dueDate *time.Time, now time.Time) error {
	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle == "" {
		return ErrEmptyTitle
	}
	if utf8.RuneCountInString(trimmedTitle) > 255 {
		return ErrTitleTooLong
	}
	if !priority.IsValid() {
		return ErrInvalidPriority
	}

	normTags, err := NormalizeTagSlice(tags)
	if err != nil {
		return err
	}

	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}

	var copiedDue *time.Time
	if dueDate != nil {
		d := dueDate.UTC()
		copiedDue = &d
	}

	t.Title = trimmedTitle
	t.Description = desc
	t.Priority = priority
	t.Tags = normTags
	t.DueDate = copiedDue
	t.UpdatedAt = now

	return nil
}

func (t *Task) SetParent(parentID *string, now time.Time) error {
	if parentID != nil {
		p := strings.TrimSpace(*parentID)
		if p == "" {
			return ErrInvalidTaskID
		}
		if p == t.ID {
			return ErrSelfParenting
		}
		t.ParentID = &p
	} else {
		t.ParentID = nil
	}

	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}

	t.UpdatedAt = now
	return nil
}

func (t *Task) SetProgress(progress int, now time.Time) error {
	if progress < 0 || progress > 100 {
		return ErrInvalidProgress
	}
	if t.Status == StatusDone && progress != 100 {
		return ErrInvalidProgress
	}
	if t.Status != StatusDone && progress == 100 {
		return ErrInvalidProgress
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}

	t.Progress = progress
	t.UpdatedAt = now
	return nil
}

// SetRollupProgress sets the rollup-calculated progress (0-100) on a task,
// allowing 100 on a non-done parent whose direct subtasks have all completed.
func (t *Task) SetRollupProgress(progress int, now time.Time) error {
	if progress < 0 || progress > 100 {
		return ErrInvalidProgress
	}
	if t.Status == StatusDone && progress != 100 {
		return ErrInvalidProgress
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
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

// Clone returns a deep copy of the Task with independent pointer and slice fields.
func (t Task) Clone() Task {
	clone := t
	if t.ParentID != nil {
		p := *t.ParentID
		clone.ParentID = &p
	}
	if t.DueDate != nil {
		d := *t.DueDate
		clone.DueDate = &d
	}
	if t.CompletedAt != nil {
		c := *t.CompletedAt
		clone.CompletedAt = &c
	}
	if t.Tags != nil {
		clone.Tags = make([]Tag, len(t.Tags))
		copy(clone.Tags, t.Tags)
	}
	return clone
}
