package core_test

import (
	"errors"
	"reflect"
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

	// Reopening a done task hydrated without CompletedAt (nil CompletedAt) still resets Progress to 0
	hydratedDoneTask := core.Task{
		ID:          "done-nil-completed",
		Title:       "Done Hydrated",
		Status:      core.StatusDone,
		Progress:    100,
		CompletedAt: nil,
	}
	if err := hydratedDoneTask.TransitionTo(core.StatusInProgress, t4); err != nil {
		t.Fatalf("reopening hydrated done task failed: %v", err)
	}
	if hydratedDoneTask.Progress != 0 {
		t.Errorf("expected Progress to reset to 0 when reopening hydrated done task with nil CompletedAt, got %d", hydratedDoneTask.Progress)
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

	// Invariant: Non-done task cannot have progress 100
	nonDoneTask, _ := core.NewTask(core.NewTaskParams{ID: "nd-1", Title: "Non-Done", Now: now})
	err = nonDoneTask.SetProgress(100, now)
	if !errors.Is(err, core.ErrInvalidProgress) {
		t.Errorf("expected ErrInvalidProgress when setting 100 on non-done task, got %v", err)
	}
	_ = nonDoneTask.TransitionTo(core.StatusInProgress, now)
	err = nonDoneTask.SetProgress(100, now)
	if !errors.Is(err, core.ErrInvalidProgress) {
		t.Errorf("expected ErrInvalidProgress when setting 100 on in-progress task, got %v", err)
	}
	_ = nonDoneTask.TransitionTo(core.StatusBlocked, now)
	err = nonDoneTask.SetProgress(100, now)
	if !errors.Is(err, core.ErrInvalidProgress) {
		t.Errorf("expected ErrInvalidProgress when setting 100 on blocked task, got %v", err)
	}
	// Invariant: task with invalid status enum is rejected before mutation
	badStatusTask := core.Task{ID: "bad-1", Title: "Bad", Status: core.Status("archived"), Progress: 10, UpdatedAt: now}
	err = badStatusTask.SetProgress(50, now.Add(time.Hour))
	if !errors.Is(err, core.ErrInvalidStatus) {
		t.Errorf("expected ErrInvalidStatus when task has invalid status, got %v", err)
	}
	if badStatusTask.Progress != 10 || !badStatusTask.UpdatedAt.Equal(now) {
		t.Errorf("failed SetProgress mutated state: Progress=%d, UpdatedAt=%v", badStatusTask.Progress, badStatusTask.UpdatedAt)
	}
}

func TestTask_SetRollupProgress(t *testing.T) {
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	parent, _ := core.NewTask(core.NewTaskParams{ID: "parent", Title: "Parent", Now: now})
	_ = parent.TransitionTo(core.StatusInProgress, now)

	child1, _ := core.NewTask(core.NewTaskParams{ID: "c1", Title: "C1", Now: now})
	_ = child1.TransitionTo(core.StatusDone, now)
	subtasks := []core.Task{*child1}

	// Out of bounds < 0
	err := parent.SetRollupProgress(-1, subtasks, now)
	if !errors.Is(err, core.ErrInvalidProgress) {
		t.Errorf("expected ErrInvalidProgress for -1, got %v", err)
	}

	// Out of bounds > 100
	err = parent.SetRollupProgress(101, subtasks, now)
	if !errors.Is(err, core.ErrInvalidProgress) {
		t.Errorf("expected ErrInvalidProgress for 101, got %v", err)
	}

	// Setting 100 on non-done parent without subtasks fails
	err = parent.SetRollupProgress(100, nil, now)
	if !errors.Is(err, core.ErrInvalidProgress) {
		t.Errorf("expected ErrInvalidProgress when setting 100 without subtasks, got %v", err)
	}

	// Setting 100 on non-done parent with incomplete subtasks fails
	childIncomplete, _ := core.NewTask(core.NewTaskParams{ID: "c2", Title: "C2", Now: now})
	_ = childIncomplete.TransitionTo(core.StatusInProgress, now)
	err = parent.SetRollupProgress(100, []core.Task{*childIncomplete}, now)
	if !errors.Is(err, core.ErrInvalidProgress) {
		t.Errorf("expected ErrInvalidProgress when subtasks incomplete, got %v", err)
	}

	// Subtask with corrupt progress > 100 (e.g. 150) must be rejected
	childOver, _ := core.NewTask(core.NewTaskParams{ID: "c-over", Title: "Over", Now: now})
	_ = childOver.TransitionTo(core.StatusInProgress, now)
	childOver.Progress = 150
	err = parent.SetRollupProgress(100, []core.Task{*childOver}, now)
	if !errors.Is(err, core.ErrInvalidProgress) {
		t.Errorf("expected ErrInvalidProgress when subtask has corrupt progress > 100, got %v", err)
	}

	// When subtasks are provided, progress must match CalculateProgress
	// childWith20 has 20% progress; passing 99% must be rejected
	childWith20, _ := core.NewTask(core.NewTaskParams{ID: "c-20", Title: "C20", Now: now})
	_ = childWith20.TransitionTo(core.StatusInProgress, now)
	_ = childWith20.SetProgress(20, now)
	err = parent.SetRollupProgress(99, []core.Task{*childWith20}, now)
	if !errors.Is(err, core.ErrInvalidProgress) {
		t.Errorf("expected ErrInvalidProgress when progress 99 does not match subtasks rollup 20, got %v", err)
	}

	// Passing matching progress 20 succeeds
	err = parent.SetRollupProgress(20, []core.Task{*childWith20}, now)
	if err != nil {
		t.Errorf("expected matching rollup progress 20 to succeed, got %v", err)
	}
	if parent.Progress != 20 {
		t.Errorf("parent.Progress = %d, want 20", parent.Progress)
	}

	// Non-done child with corrupt progress > 100 is rejected outright: the setter
	// validates child ranges before consulting CalculateProgress, so neither 100
	// nor the clamped 99 may be persisted from corrupt child data.
	childCorrupt, _ := core.NewTask(core.NewTaskParams{ID: "c-corrupt", Title: "Corrupt", Now: now})
	_ = childCorrupt.TransitionTo(core.StatusInProgress, now)
	childCorrupt.Progress = 150
	err = parent.SetRollupProgress(100, []core.Task{*childCorrupt}, now)
	if !errors.Is(err, core.ErrInvalidProgress) {
		t.Errorf("expected ErrInvalidProgress when setting 100 on corrupt subtask, got %v", err)
	}
	err = parent.SetRollupProgress(99, []core.Task{*childCorrupt}, now)
	if !errors.Is(err, core.ErrInvalidProgress) {
		t.Errorf("expected ErrInvalidProgress when setting clamped 99 on corrupt subtask, got %v", err)
	}

	// Child with invalid status enum and Progress == 100 is rejected with ErrInvalidStatus
	childBadStatus := core.Task{ID: "c-badstatus", Title: "Bad", Status: core.Status("archived"), Progress: 100}
	err = parent.SetRollupProgress(100, []core.Task{childBadStatus}, now)
	if !errors.Is(err, core.ErrInvalidStatus) {
		t.Errorf("expected ErrInvalidStatus when child has invalid status, got %v", err)
	}
	// Setting 100 on non-done parent with complete subtasks succeeds
	tUpdate := now.Add(5 * time.Minute)
	err = parent.SetRollupProgress(100, subtasks, tUpdate)
	if err != nil {
		t.Errorf("expected SetRollupProgress(100) on non-done parent to succeed, got %v", err)
	}
	if parent.Progress != 100 {
		t.Errorf("parent.Progress = %d, want 100", parent.Progress)
	}
	if !parent.UpdatedAt.Equal(tUpdate) {
		t.Errorf("parent.UpdatedAt = %v, want %v", parent.UpdatedAt, tUpdate)
	}

	// Zero time normalizes to current UTC time
	err = parent.SetRollupProgress(80, nil, time.Time{})
	if err != nil {
		t.Errorf("SetRollupProgress with zero time failed: %v", err)
	}
	if parent.UpdatedAt.IsZero() || parent.UpdatedAt.Location() != time.UTC {
		t.Errorf("expected non-zero UTC UpdatedAt, got %v", parent.UpdatedAt)
	}

	// On a done task, progress must be 100
	_ = parent.TransitionTo(core.StatusDone, now.Add(10*time.Minute))
	err = parent.SetRollupProgress(50, nil, now.Add(15*time.Minute))
	if !errors.Is(err, core.ErrInvalidProgress) {
		t.Errorf("expected ErrInvalidProgress when setting <100 on done task, got %v", err)
	}

	// On a done task, SetRollupProgress(100, incompleteSubtasks, now) succeeds
	// even if subtasks are incomplete (a done task is definitionally 100)
	err = parent.SetRollupProgress(100, []core.Task{*childWith20}, now.Add(20*time.Minute))
	if err != nil {
		t.Errorf("expected SetRollupProgress(100) on done task with incomplete subtasks to succeed, got %v", err)
	}
	if parent.Progress != 100 {
		t.Errorf("expected parent.Progress = 100 on done task, got %d", parent.Progress)
	}

	// Parent with invalid status enum is rejected with ErrInvalidStatus
	badParent := core.Task{ID: "bad", Title: "Bad", Status: core.Status("archived"), Progress: 50}
	err = badParent.SetRollupProgress(50, subtasks, now.Add(25*time.Minute))
	if !errors.Is(err, core.ErrInvalidStatus) {
		t.Errorf("expected ErrInvalidStatus when parent has invalid status, got %v", err)
	}
}

func TestTask_RollupSubtasksRemoved(t *testing.T) {
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	parent, _ := core.NewTask(core.NewTaskParams{ID: "p", Title: "Parent", Now: now})
	_ = parent.TransitionTo(core.StatusInProgress, now)
	_ = parent.SetProgress(20, now)

	// Add completed child and update rollup progress to 100
	child, _ := core.NewTask(core.NewTaskParams{ID: "c", Title: "Child", Now: now})
	_ = child.TransitionTo(core.StatusDone, now)
	_ = parent.SetRollupProgress(100, []core.Task{*child}, now.Add(5*time.Minute))
	if parent.Progress != 100 {
		t.Fatalf("expected rollup progress 100, got %d", parent.Progress)
	}
	// Setting 100 without subtasks on non-done parent fails
	tReset := now.Add(10 * time.Minute)
	err := parent.SetRollupProgress(100, nil, tReset)
	if !errors.Is(err, core.ErrInvalidProgress) {
		t.Errorf("expected ErrInvalidProgress when setting 100 without subtasks, got %v", err)
	}

	// Child is removed: SetRollupProgress with nil/empty subtasks resets leaf progress
	err = parent.SetRollupProgress(20, nil, tReset)
	if err != nil {
		t.Fatalf("SetRollupProgress with empty subtasks failed: %v", err)
	}
	if parent.Progress != 20 {
		t.Errorf("parent.Progress after child removal = %d, want 20", parent.Progress)
	}
	if !parent.UpdatedAt.Equal(tReset) {
		t.Errorf("parent.UpdatedAt = %v, want %v", parent.UpdatedAt, tReset)
	}
	if got := core.CalculateProgress(*parent, nil); got != 20 {
		t.Errorf("CalculateProgress after child removal = %d, want 20", got)
	}
}

func TestTask_ResetLeaf(t *testing.T) {
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	parent, _ := core.NewTask(core.NewTaskParams{ID: "p", Title: "Parent", Now: now})
	_ = parent.TransitionTo(core.StatusInProgress, now)

	// Happy path: set leaf progress to 50 with nil/empty subtasks
	t1 := now.Add(5 * time.Minute)
	err := parent.ResetLeaf(50, nil, t1)
	if err != nil {
		t.Fatalf("ResetLeaf(50) failed: %v", err)
	}
	if parent.Progress != 50 {
		t.Errorf("parent.Progress = %d, want 50", parent.Progress)
	}
	if !parent.UpdatedAt.Equal(t1) {
		t.Errorf("parent.UpdatedAt = %v, want %v", parent.UpdatedAt, t1)
	}

	// Negative path: cannot reset to leaf when task still has subtasks
	child, _ := core.NewTask(core.NewTaskParams{ID: "c", Title: "Child", Now: now})
	err = parent.ResetLeaf(50, []core.Task{*child}, now.Add(6*time.Minute))
	if !errors.Is(err, core.ErrInvalidProgress) {
		t.Errorf("expected ErrInvalidProgress when resetting with non-empty subtasks, got %v", err)
	}
	// Assert failed call left state untouched
	if parent.Progress != 50 || !parent.UpdatedAt.Equal(t1) {
		t.Errorf("failed ResetLeaf mutated state: Progress=%d, UpdatedAt=%v", parent.Progress, parent.UpdatedAt)
	}

	// Negative path: setting 100 on non-done task fails
	err = parent.ResetLeaf(100, nil, now.Add(7*time.Minute))
	if !errors.Is(err, core.ErrInvalidProgress) {
		t.Errorf("expected ErrInvalidProgress when setting 100 on non-done task via ResetLeaf, got %v", err)
	}
	// Assert failed call left state untouched
	if parent.Progress != 50 || !parent.UpdatedAt.Equal(t1) {
		t.Errorf("failed ResetLeaf mutated state: Progress=%d, UpdatedAt=%v", parent.Progress, parent.UpdatedAt)
	}

	// Negative path: out of bounds < 0 or > 100
	if err := parent.ResetLeaf(-1, nil, now); !errors.Is(err, core.ErrInvalidProgress) {
		t.Errorf("expected ErrInvalidProgress for -1, got %v", err)
	}
	if err := parent.ResetLeaf(101, nil, now); !errors.Is(err, core.ErrInvalidProgress) {
		t.Errorf("expected ErrInvalidProgress for 101, got %v", err)
	}
	if parent.Progress != 50 || !parent.UpdatedAt.Equal(t1) {
		t.Errorf("failed ResetLeaf mutated state: Progress=%d, UpdatedAt=%v", parent.Progress, parent.UpdatedAt)
	}

	// On done task: must be exactly 100
	_ = parent.TransitionTo(core.StatusDone, now.Add(10*time.Minute))
	if err := parent.ResetLeaf(50, nil, now); !errors.Is(err, core.ErrInvalidProgress) {
		t.Errorf("expected ErrInvalidProgress for < 100 on done task via ResetLeaf, got %v", err)
	}
	tDone := now.Add(15 * time.Minute)
	if err := parent.ResetLeaf(100, nil, tDone); err != nil {
		t.Errorf("expected ResetLeaf(100) on done task to succeed, got %v", err)
	}
	if parent.Progress != 100 {
		t.Errorf("parent.Progress = %d, want 100", parent.Progress)
	}
	if !parent.UpdatedAt.Equal(tDone) {
		t.Errorf("parent.UpdatedAt = %v, want %v", parent.UpdatedAt, tDone)
	}
}

func TestTask_DescriptionPreservesMarkdownIndentation(t *testing.T) {
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	markdownDesc := "    code block\n    continued"

	task, err := core.NewTask(core.NewTaskParams{
		ID:          "task-md",
		Title:       "Markdown Task",
		Description: markdownDesc,
		Now:         now,
	})
	if err != nil {
		t.Fatalf("NewTask failed: %v", err)
	}
	if task.Description != markdownDesc {
		t.Errorf("expected Description to preserve indentation, got %q", task.Description)
	}

	updatedMD := "  * list item\n  * nested"
	err = task.Update("Title", updatedMD, core.PriorityMedium, nil, nil, now)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if task.Description != updatedMD {
		t.Errorf("expected Update to preserve indentation, got %q", task.Description)
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

func TestTask_Clone(t *testing.T) {
	parentID := "p1"
	dueDate := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
	completedAt := time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)
	tags := []core.Tag{core.Tag("work")}

	task := core.Task{
		ID:          "task-1",
		Title:       "Title",
		Description: "Desc",
		Status:      core.StatusTodo,
		Priority:    core.PriorityHigh,
		ParentID:    &parentID,
		Progress:    50,
		Tags:        tags,
		DueDate:     &dueDate,
		CreatedAt:   time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC),
		CompletedAt: &completedAt,
	}

	clone := task.Clone()
	// 1. Fidelity: clone equals the original pre-mutation
	if !reflect.DeepEqual(clone, task) {
		t.Errorf("clone does not equal original task: got %+v, want %+v", clone, task)
	}

	// 2. Distinct pointer / slice backing memory
	if clone.ParentID == task.ParentID {
		t.Errorf("clone.ParentID shares pointer address with original")
	}
	if clone.DueDate == task.DueDate {
		t.Errorf("clone.DueDate shares pointer address with original")
	}
	if clone.CompletedAt == task.CompletedAt {
		t.Errorf("clone.CompletedAt shares pointer address with original")
	}
	if len(clone.Tags) > 0 && &clone.Tags[0] == &task.Tags[0] {
		t.Errorf("clone.Tags shares slice backing array with original")
	}

	// 3. Independence: mutate task and verify clone is untouched
	parentID = "p2"
	dueDate = dueDate.Add(24 * time.Hour)
	completedAt = completedAt.Add(24 * time.Hour)
	tags[0] = core.Tag("home")
	if *clone.ParentID != "p1" {
		t.Errorf("clone.ParentID mutated: got %s, want p1", *clone.ParentID)
	}
	if !clone.DueDate.Equal(time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("clone.DueDate mutated: got %v", clone.DueDate)
	}
	if !clone.CompletedAt.Equal(time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("clone.CompletedAt mutated: got %v", clone.CompletedAt)
	}
	if clone.Tags[0] != core.Tag("work") {
		t.Errorf("clone.Tags mutated: got %s, want work", clone.Tags[0])
	}
}

func TestTask_Clone_NilFields(t *testing.T) {
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	minimalTask := core.Task{
		ID:        "min-1",
		Title:     "Minimal",
		Status:    core.StatusTodo,
		Priority:  core.PriorityLow,
		CreatedAt: now,
		UpdatedAt: now,
		// ParentID, DueDate, CompletedAt, Tags are all nil
	}

	clone := minimalTask.Clone()

	if clone.ParentID != nil {
		t.Errorf("expected clone.ParentID to be nil, got %v", clone.ParentID)
	}
	if clone.DueDate != nil {
		t.Errorf("expected clone.DueDate to be nil, got %v", clone.DueDate)
	}
	if clone.CompletedAt != nil {
		t.Errorf("expected clone.CompletedAt to be nil, got %v", clone.CompletedAt)
	}
	if clone.Tags != nil {
		t.Errorf("expected clone.Tags to be nil, got %v", clone.Tags)
	}
	if !reflect.DeepEqual(clone, minimalTask) {
		t.Errorf("clone does not equal minimal task: got %+v, want %+v", clone, minimalTask)
	}

	// Also verify empty (non-nil) Tags preservation:
	emptyTagsTask := minimalTask
	emptyTagsTask.Tags = []core.Tag{}
	cloneEmpty := emptyTagsTask.Clone()
	if cloneEmpty.Tags == nil {
		t.Errorf("expected cloneEmpty.Tags to be empty non-nil slice, got nil")
	}
	if len(cloneEmpty.Tags) != 0 {
		t.Errorf("expected cloneEmpty.Tags length 0, got %d", len(cloneEmpty.Tags))
	}
	if !reflect.DeepEqual(cloneEmpty, emptyTagsTask) {
		t.Errorf("cloneEmpty does not equal emptyTagsTask: got %+v, want %+v", cloneEmpty, emptyTagsTask)
	}
}

func TestTask_Clone_FieldExhaustiveness(t *testing.T) {
	taskType := reflect.TypeOf(core.Task{})
	expectedFields := map[string]reflect.Kind{
		"ID":          reflect.String,
		"Title":       reflect.String,
		"Description": reflect.String,
		"Status":      reflect.String,
		"Priority":    reflect.Int,
		"ParentID":    reflect.Pointer,
		"Progress":    reflect.Int,
		"Tags":        reflect.Slice,
		"DueDate":     reflect.Pointer,
		"CreatedAt":   reflect.Struct,
		"UpdatedAt":   reflect.Struct,
		"CompletedAt": reflect.Pointer,
	}

	if taskType.NumField() != len(expectedFields) {
		t.Fatalf("Task has %d fields, expected %d. If you added a new field, ensure Task.Clone() explicitly handles it to prevent accidental aliasing!",
			taskType.NumField(), len(expectedFields))
	}

	for i := range taskType.NumField() {
		field := taskType.Field(i)
		expectedKind, ok := expectedFields[field.Name]
		if !ok {
			t.Errorf("unexpected field %s in Task", field.Name)
			continue
		}
		if field.Type.Kind() != expectedKind {
			t.Errorf("field %s kind is %v, expected %v", field.Name, field.Type.Kind(), expectedKind)
		}
		switch field.Type.Kind() {
		case reflect.Pointer, reflect.Slice, reflect.Map, reflect.Interface, reflect.Chan, reflect.Func:
			switch field.Name {
			case "ParentID", "DueDate", "CompletedAt", "Tags":
				// Explicitly cloned in Task.Clone()
			default:
				t.Errorf("field %s is a reference type (%v) not accounted for in Task.Clone() deep-copy logic", field.Name, field.Type.Kind())
			}
		}
	}
}
