package core_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/tusk/internal/core"
)

func TestNewTask_Validation(t *testing.T) {
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)

	// Empty title
	_, err := core.NewTask(core.NewTaskParams{
		ID:    "task-1",
		Title: "   ",
		Now:   now,
	})
	if !errors.Is(err, core.ErrEmptyTitle) {
		t.Errorf("expected ErrEmptyTitle, got %v", err)
	}
	// Empty ID
	_, err = core.NewTask(core.NewTaskParams{
		ID:    "  ",
		Title: "Valid Title",
		Now:   now,
	})
	if !errors.Is(err, core.ErrInvalidTaskID) {
		t.Errorf("expected ErrInvalidTaskID, got %v", err)
	}

	// Empty parent ID
	emptyParent := "   "
	_, err = core.NewTask(core.NewTaskParams{
		ID:       "task-1",
		Title:    "Valid Title",
		ParentID: &emptyParent,
		Now:      now,
	})
	if !errors.Is(err, core.ErrInvalidTaskID) {
		t.Errorf("expected ErrInvalidTaskID for empty parent, got %v", err)
	}

	// Non-zero invalid priority
	_, err = core.NewTask(core.NewTaskParams{
		ID:       "task-1",
		Title:    "Valid Title",
		Priority: core.Priority(99),
		Now:      now,
	})
	if !errors.Is(err, core.ErrInvalidPriority) {
		t.Errorf("expected ErrInvalidPriority for Priority(99), got %v", err)
	}

	// Title too long (> 255 characters)
	_, err = core.NewTask(core.NewTaskParams{
		ID:    "task-1",
		Title: strings.Repeat("a", 256),
		Now:   now,
	})
	if !errors.Is(err, core.ErrTitleTooLong) {
		t.Errorf("expected ErrTitleTooLong, got %v", err)
	}

	// Multi-byte runes UTF-8: 255 non-ASCII runes should pass (takes > 255 bytes)
	multiRuneTitle := strings.Repeat("世", 255)
	_, err = core.NewTask(core.NewTaskParams{
		ID:    "task-multibyte",
		Title: multiRuneTitle,
		Now:   now,
	})
	if err != nil {
		t.Fatalf("expected 255-rune title to pass, got: %v", err)
	}

	// 256 multi-byte runes should fail
	_, err = core.NewTask(core.NewTaskParams{
		ID:    "task-multibyte-too-long",
		Title: strings.Repeat("世", 256),
		Now:   now,
	})
	if !errors.Is(err, core.ErrTitleTooLong) {
		t.Errorf("expected ErrTitleTooLong for 256 runes, got %v", err)
	}
	// Title exactly 255 characters (should pass)
	task, err := core.NewTask(core.NewTaskParams{
		ID:       "task-1",
		Title:    strings.Repeat("a", 255),
		Priority: core.PriorityHigh,
		Tags:     []string{"#backend", "API"},
		Now:      now,
	})
	if err != nil {
		t.Fatalf("unexpected error for 255-char title: %v", err)
	}
	if task.ID != "task-1" {
		t.Errorf("task.ID = %q, want task-1", task.ID)
	}
	if task.Status != core.StatusTodo {
		t.Errorf("task.Status = %v, want %v", task.Status, core.StatusTodo)
	}
	if task.Progress != 0 {
		t.Errorf("task.Progress = %d, want 0", task.Progress)
	}
	if !task.CreatedAt.Equal(now) || !task.UpdatedAt.Equal(now) {
		t.Errorf("task timestamps mismatch: created=%v, updated=%v", task.CreatedAt, task.UpdatedAt)
	}
	if task.CompletedAt != nil {
		t.Errorf("expected CompletedAt to be nil, got %v", task.CompletedAt)
	}
	if len(task.Tags) != 2 || task.Tags[0] != core.Tag("api") || task.Tags[1] != core.Tag("backend") {
		t.Errorf("unexpected tags: %v", task.Tags)
	}
	if !task.IsRoot() {
		t.Errorf("expected task without parent to be root")
	}

	// Self-parenting in NewTask
	selfID := "task-self"
	_, err = core.NewTask(core.NewTaskParams{
		ID:       selfID,
		Title:    "Self Parenting Task",
		ParentID: &selfID,
		Now:      now,
	})
	if !errors.Is(err, core.ErrSelfParenting) {
		t.Errorf("expected ErrSelfParenting, got %v", err)
	}
}

func TestTask_TransitionToDone_And_Reopen(t *testing.T) {
	t0 := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	task, err := core.NewTask(core.NewTaskParams{
		ID:    "task-1",
		Title: "Test Transition",
		Now:   t0,
	})
	if err != nil {
		t.Fatalf("NewTask failed: %v", err)
	}

	// Transition to in-progress
	t1 := t0.Add(1 * time.Hour)
	if err := task.TransitionTo(core.StatusInProgress, t1); err != nil {
		t.Fatalf("TransitionTo(StatusInProgress) failed: %v", err)
	}
	if task.Status != core.StatusInProgress || !task.UpdatedAt.Equal(t1) || task.CompletedAt != nil {
		t.Errorf("unexpected state after in-progress transition: %+v", task)
	}

	// Transition to done
	t2 := t1.Add(1 * time.Hour)
	if err := task.TransitionTo(core.StatusDone, t2); err != nil {
		t.Fatalf("TransitionTo(StatusDone) failed: %v", err)
	}
	if task.Status != core.StatusDone {
		t.Errorf("expected StatusDone, got %v", task.Status)
	}
	if task.CompletedAt == nil || !task.CompletedAt.Equal(t2) {
		t.Errorf("expected CompletedAt = %v, got %v", t2, task.CompletedAt)
	}
	if !task.IsDone() {
		t.Errorf("expected IsDone() to be true")
	}

	// Invalid transition from done to blocked
	t3 := t2.Add(1 * time.Hour)
	err = task.TransitionTo(core.StatusBlocked, t3)
	if !errors.Is(err, core.ErrInvalidStatusTransition) {
		t.Errorf("expected ErrInvalidStatusTransition for done -> blocked, got %v", err)
	}

	// Reopen from done to in-progress
	t4 := t3.Add(1 * time.Hour)
	if err := task.TransitionTo(core.StatusInProgress, t4); err != nil {
		t.Fatalf("reopening to in-progress failed: %v", err)
	}
	if task.Status != core.StatusInProgress {
		t.Errorf("expected StatusInProgress, got %v", task.Status)
	}
	if task.CompletedAt != nil {
		t.Errorf("expected CompletedAt to be cleared to nil, got %v", task.CompletedAt)
	}
	if task.Progress != 0 {
		t.Errorf("expected Progress to be reset to 0 on reopen, got %d", task.Progress)
	}
	if !task.UpdatedAt.Equal(t4) {
		t.Errorf("expected UpdatedAt = %v, got %v", t4, task.UpdatedAt)
	}
}

func TestTask_SetParent(t *testing.T) {
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	task, err := core.NewTask(core.NewTaskParams{
		ID:    "task-child",
		Title: "Child Task",
		Now:   now,
	})
	if err != nil {
		t.Fatalf("NewTask failed: %v", err)
	}

	// Self-parenting
	selfID := "task-child"
	err = task.SetParent(&selfID, now.Add(10*time.Minute))
	if !errors.Is(err, core.ErrSelfParenting) {
		t.Errorf("expected ErrSelfParenting, got %v", err)
	}

	// Valid reparenting
	parentID := "task-parent"
	t1 := now.Add(20 * time.Minute)
	if err := task.SetParent(&parentID, t1); err != nil {
		t.Fatalf("SetParent failed: %v", err)
	}
	if task.ParentID == nil || *task.ParentID != parentID {
		t.Errorf("task.ParentID = %v, want %s", task.ParentID, parentID)
	}
	if !task.UpdatedAt.Equal(t1) {
		t.Errorf("task.UpdatedAt = %v, want %v", task.UpdatedAt, t1)
	}
	if task.IsRoot() {
		t.Errorf("expected task with parent to not be root")
	}

	// Unparent back to root
	t2 := t1.Add(10 * time.Minute)
	if err := task.SetParent(nil, t2); err != nil {
		t.Fatalf("SetParent(nil) failed: %v", err)
	}
	if task.ParentID != nil {
		t.Errorf("expected ParentID to be nil, got %v", task.ParentID)
	}
	if !task.IsRoot() {
		t.Errorf("expected task after SetParent(nil) to be root")
	}
}

func TestTask_SetProgress_Validation(t *testing.T) {
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	task, err := core.NewTask(core.NewTaskParams{
		ID:    "task-1",
		Title: "Progress Task",
		Now:   now,
	})
	if err != nil {
		t.Fatalf("NewTask failed: %v", err)
	}

	// Out of bounds < 0
	err = task.SetProgress(-1, now.Add(5*time.Minute))
	if !errors.Is(err, core.ErrInvalidProgress) {
		t.Errorf("expected ErrInvalidProgress for -1, got %v", err)
	}

	// Out of bounds > 100
	err = task.SetProgress(101, now.Add(5*time.Minute))
	if !errors.Is(err, core.ErrInvalidProgress) {
		t.Errorf("expected ErrInvalidProgress for 101, got %v", err)
	}

	// Valid progress
	t1 := now.Add(10 * time.Minute)
	if err := task.SetProgress(75, t1); err != nil {
		t.Fatalf("SetProgress(75) unexpected error: %v", err)
	}
	if task.Progress != 75 {
		t.Errorf("task.Progress = %d, want 75", task.Progress)
	}
	if !task.UpdatedAt.Equal(t1) {
		t.Errorf("task.UpdatedAt = %v, want %v", task.UpdatedAt, t1)
	}

	// Invariant: Once done, progress cannot be set below 100
	t2 := t1.Add(10 * time.Minute)
	_ = task.TransitionTo(core.StatusDone, t2)
	err = task.SetProgress(50, t2.Add(5*time.Minute))
	if !errors.Is(err, core.ErrInvalidProgress) {
		t.Errorf("expected ErrInvalidProgress when setting <100 on done task, got %v", err)
	}
	// Setting 100 on done task succeeds
	err = task.SetProgress(100, t2.Add(10*time.Minute))
	if err != nil {
		t.Errorf("expected SetProgress(100) on done task to succeed, got %v", err)
	}
}

func TestTask_Update(t *testing.T) {
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	task, err := core.NewTask(core.NewTaskParams{
		ID:    "task-1",
		Title: "Initial Title",
		Now:   now,
	})
	if err != nil {
		t.Fatalf("NewTask failed: %v", err)
	}

	due := now.Add(48 * time.Hour)
	t1 := now.Add(30 * time.Minute)
	err = task.Update("Updated Title", "Updated Description", core.PriorityUrgent, []core.Tag{core.Tag("ops")}, &due, t1)
	if err != nil {
		t.Fatalf("Update unexpected error: %v", err)
	}
	if task.Title != "Updated Title" {
		t.Errorf("task.Title = %q, want Updated Title", task.Title)
	}
	if task.Description != "Updated Description" {
		t.Errorf("task.Description = %q, want Updated Description", task.Description)
	}
	if task.Priority != core.PriorityUrgent {
		t.Errorf("task.Priority = %v, want %v", task.Priority, core.PriorityUrgent)
	}
	if len(task.Tags) != 1 || task.Tags[0] != core.Tag("ops") {
		t.Errorf("task.Tags = %v, want [ops]", task.Tags)
	}
	if task.DueDate == nil || !task.DueDate.Equal(due) {
		t.Errorf("task.DueDate = %v, want %v", task.DueDate, due)
	}
	if !task.UpdatedAt.Equal(t1) {
		t.Errorf("task.UpdatedAt = %v, want %v", task.UpdatedAt, t1)
	}

	// Update with empty title fails
	err = task.Update("  ", "", core.PriorityLow, nil, nil, t1)
	if !errors.Is(err, core.ErrEmptyTitle) {
		t.Errorf("expected ErrEmptyTitle, got %v", err)
	}

	// Update with invalid tags fails
	err = task.Update("Valid", "", core.PriorityLow, []core.Tag{core.Tag("invalid tag")}, nil, t1)
	if !errors.Is(err, core.ErrInvalidTag) {
		t.Errorf("expected ErrInvalidTag for update, got %v", err)
	}

	// SetParent with empty string fails
	emptyP := "   "
	err = task.SetParent(&emptyP, t1)
	if !errors.Is(err, core.ErrInvalidTaskID) {
		t.Errorf("expected ErrInvalidTaskID for SetParent empty, got %v", err)
	}
}

func TestTask_DefensiveCopiesAndUTC(t *testing.T) {
	// Local timezone
	loc := time.FixedZone("EST", -5*3600)
	localTime := time.Date(2026, 9, 6, 12, 0, 0, 0, loc)

	parentOriginal := "parent-1"
	parentPtr := &parentOriginal
	dueOriginal := localTime.Add(24 * time.Hour)
	duePtr := &dueOriginal

	task, err := core.NewTask(core.NewTaskParams{
		ID:       "task-defensive",
		Title:    "Defensive Task",
		ParentID: parentPtr,
		DueDate:  duePtr,
		Now:      localTime,
	})
	if err != nil {
		t.Fatalf("NewTask failed: %v", err)
	}

	// Mutate external pointers
	parentOriginal = "mutated-parent"
	*duePtr = localTime.Add(48 * time.Hour)

	if *task.ParentID != "parent-1" {
		t.Errorf("task.ParentID was mutated externally to %q", *task.ParentID)
	}
	if !task.DueDate.Equal(localTime.Add(24 * time.Hour)) {
		t.Errorf("task.DueDate was mutated externally to %v", task.DueDate)
	}

	// Invariant: Timestamps must be in UTC
	if task.CreatedAt.Location() != time.UTC {
		t.Errorf("expected CreatedAt in UTC, got location: %v", task.CreatedAt.Location())
	}
	if task.UpdatedAt.Location() != time.UTC {
		t.Errorf("expected UpdatedAt in UTC, got location: %v", task.UpdatedAt.Location())
	}
	if task.DueDate.Location() != time.UTC {
		t.Errorf("expected DueDate in UTC, got location: %v", task.DueDate.Location())
	}
}
