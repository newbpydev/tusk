package core

import (
	"fmt"
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

	now := normalizeTime(params.Now)

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
		now = normalizeTime(now)
		if next == StatusDone {
			if t.CompletedAt != nil && t.Progress == 100 {
				return nil
			}
			if t.CompletedAt == nil {
				t.CompletedAt = &now
			}
			t.Progress = 100
		} else if t.CompletedAt != nil {
			t.CompletedAt = nil
			if t.Progress < 0 || t.Progress > 99 {
				t.Progress = 0
			}
		} else {
			return nil
		}
		t.UpdatedAt = now
		return nil
	}

	now = normalizeTime(now)

	prevStatus := t.Status
	t.Status = next
	t.UpdatedAt = now

	if next == StatusDone {
		completed := now
		t.CompletedAt = &completed
		t.Progress = 100
	} else if prevStatus == StatusDone || t.CompletedAt != nil {
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

	now = normalizeTime(now)

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

	now = normalizeTime(now)

	t.UpdatedAt = now
	return nil
}

// SetProgress sets explicit manual progress (0-99 for non-done tasks, strictly 100 for done tasks).
// The task must carry a valid status enum; ResetLeaf inherits this check via delegation.
// Returns ErrInvalidStatus on invalid task status, ErrInvalidProgress if progress < 0,
// progress > 100, progress == 100 on a non-done task, or progress != 100 on a done task.
func (t *Task) SetProgress(progress int, now time.Time) error {
	if !t.Status.IsValid() {
		return fmt.Errorf("task %s has invalid status %q: %w", t.ID, string(t.Status), ErrInvalidStatus)
	}
	if progress < 0 || progress > 100 {
		return ErrInvalidProgress
	}
	if t.Status == StatusDone && progress != 100 {
		return ErrInvalidProgress
	}
	if t.Status != StatusDone && progress == 100 {
		return ErrInvalidProgress
	}

	t.Progress = progress
	t.UpdatedAt = normalizeTime(now)
	return nil
}

// SetRollupProgress sets progress computed by the rollup engine (0-100).
// The parent task must carry a valid status enum; for non-done tasks with
// subtasks, every subtask must carry a valid status enum and every non-done
// subtask must carry valid stored progress (0-100); progress must match
// CalculateProgress(*t, subtasks).
// CalculateProgress clamps corrupt values for display purposes only; this setter
// rejects corrupt child data so invalid rows cannot enter the persisted rollup lifecycle.
// Setting 100 on a non-done parent requires complete subtasks.
// Returns ErrInvalidStatus on invalid parent or child status enums,
// ErrInvalidProgress on out-of-bounds inputs, corrupt subtask progress,
// value mismatches with subtasks, or if attempting to set 100 on a non-done
// parent without complete subtasks.
func (t *Task) SetRollupProgress(progress int, subtasks []Task, now time.Time) error {
	if !t.Status.IsValid() {
		return fmt.Errorf("task %s has invalid status %q: %w", t.ID, string(t.Status), ErrInvalidStatus)
	}
	if progress < 0 || progress > 100 {
		return ErrInvalidProgress
	}
	if t.Status == StatusDone && progress != 100 {
		return ErrInvalidProgress
	}
	if t.Status != StatusDone {
		if len(subtasks) > 0 {
			for _, s := range subtasks {
				if !s.Status.IsValid() {
					return fmt.Errorf("subtask %s has invalid status %q: %w", s.ID, string(s.Status), ErrInvalidStatus)
				}
				if s.Status != StatusDone && (s.Progress < 0 || s.Progress > 100) {
					return fmt.Errorf("subtask %s has invalid progress %d: %w", s.ID, s.Progress, ErrInvalidProgress)
				}
			}
			expected := CalculateProgress(*t, subtasks)
			if progress != expected {
				return fmt.Errorf("provided progress %d does not match calculated rollup %d: %w", progress, expected, ErrInvalidProgress)
			}
		} else if progress == 100 {
			return fmt.Errorf("setting 100 on non-done task requires complete subtasks: %w", ErrInvalidProgress)
		}
	}

	t.Progress = progress
	t.UpdatedAt = normalizeTime(now)
	return nil
}

// ResetLeaf resets a task whose subtasks were removed back to leaf status with explicit manual progress.
// Subtasks must be provided and must be empty (len(subtasks) == 0).
// For non-done tasks, progress must be between 0 and 99; for done tasks, progress must be 100.
func (t *Task) ResetLeaf(manualProgress int, subtasks []Task, now time.Time) error {
	if len(subtasks) != 0 {
		return fmt.Errorf("cannot reset to leaf when task still has subtasks: %w", ErrInvalidProgress)
	}
	return t.SetProgress(manualProgress, now)
}

func normalizeTime(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now().UTC()
	}
	return t.UTC()
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
