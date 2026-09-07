package core_test

import (
	"testing"
	"time"

	"github.com/newbpydev/tusk/internal/core"
)

func TestCalculateProgress_Leaf(t *testing.T) {
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)

	// Leaf task with StatusTodo, progress 0
	taskTodo, _ := core.NewTask(core.NewTaskParams{
		ID:    "t-1",
		Title: "Leaf Todo",
		Now:   now,
	})
	if got := core.CalculateProgress(*taskTodo, nil); got != 0 {
		t.Errorf("CalculateProgress(todo, nil) = %d, want 0", got)
	}

	// Leaf task with StatusInProgress and manual progress 50
	taskInProgress, _ := core.NewTask(core.NewTaskParams{
		ID:    "t-2",
		Title: "Leaf InProgress",
		Now:   now,
	})
	_ = taskInProgress.TransitionTo(core.StatusInProgress, now)
	_ = taskInProgress.SetProgress(50, now)
	if got := core.CalculateProgress(*taskInProgress, nil); got != 50 {
		t.Errorf("CalculateProgress(in-progress, nil) = %d, want 50", got)
	}

	// Leaf task with StatusDone -> strictly 100
	taskDone, _ := core.NewTask(core.NewTaskParams{
		ID:    "t-3",
		Title: "Leaf Done",
		Now:   now,
	})
	_ = taskDone.TransitionTo(core.StatusDone, now)
	if got := core.CalculateProgress(*taskDone, []core.Task{}); got != 100 {
		t.Errorf("CalculateProgress(done, empty) = %d, want 100", got)
	}
}

func TestCalculateProgress_Subtasks(t *testing.T) {
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	parent, _ := core.NewTask(core.NewTaskParams{
		ID:    "parent",
		Title: "Parent Task",
		Now:   now,
	})

	child1, _ := core.NewTask(core.NewTaskParams{ID: "c1", Title: "C1", Now: now})
	_ = child1.TransitionTo(core.StatusDone, now) // 100%

	child2, _ := core.NewTask(core.NewTaskParams{ID: "c2", Title: "C2", Now: now})
	_ = child2.TransitionTo(core.StatusInProgress, now)
	_ = child2.SetProgress(50, now) // 50%

	child3, _ := core.NewTask(core.NewTaskParams{ID: "c3", Title: "C3", Now: now}) // 0%

	subtasks := []core.Task{*child1, *child2, *child3}
	got := core.CalculateProgress(*parent, subtasks)
	if got != 50 {
		t.Errorf("CalculateProgress with [100, 50, 0] = %d, want 50", got)
	}
}

func TestCalculateProgress_FloorRounding(t *testing.T) {
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	parent, _ := core.NewTask(core.NewTaskParams{ID: "parent", Title: "Parent", Now: now})

	child1, _ := core.NewTask(core.NewTaskParams{ID: "c1", Title: "C1", Now: now})
	_ = child1.TransitionTo(core.StatusDone, now) // 100%

	child2, _ := core.NewTask(core.NewTaskParams{ID: "c2", Title: "C2", Now: now}) // 0%
	child3, _ := core.NewTask(core.NewTaskParams{ID: "c3", Title: "C3", Now: now}) // 0%

	subtasks := []core.Task{*child1, *child2, *child3}
	got := core.CalculateProgress(*parent, subtasks)
	// 100 / 3 = 33.333... -> floor is 33
	if got != 33 {
		t.Errorf("CalculateProgress with [100, 0, 0] = %d, want 33", got)
	}
}

func TestCalculateProgress_AllDone(t *testing.T) {
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	parent, _ := core.NewTask(core.NewTaskParams{ID: "parent", Title: "Parent", Now: now})

	child1, _ := core.NewTask(core.NewTaskParams{ID: "c1", Title: "C1", Now: now})
	_ = child1.TransitionTo(core.StatusDone, now)

	child2, _ := core.NewTask(core.NewTaskParams{ID: "c2", Title: "C2", Now: now})
	_ = child2.TransitionTo(core.StatusDone, now)

	subtasks := []core.Task{*child1, *child2}
	got := core.CalculateProgress(*parent, subtasks)
	if got != 100 {
		t.Errorf("CalculateProgress with all done = %d, want 100", got)
	}

	// Reopen child2 to in-progress with 40%
	_ = child2.TransitionTo(core.StatusInProgress, now)
	_ = child2.SetProgress(40, now)

	reopenedSubtasks := []core.Task{*child1, *child2}
	gotReopened := core.CalculateProgress(*parent, reopenedSubtasks)
	// (100 + 40) / 2 = 70
	if gotReopened != 70 {
		t.Errorf("CalculateProgress after reopen = %d, want 70", gotReopened)
	}
}

func TestCalculateProgress_NonDoneSubtaskAt100(t *testing.T) {
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	parent, _ := core.NewTask(core.NewTaskParams{ID: "parent", Title: "Parent", Now: now})

	child, _ := core.NewTask(core.NewTaskParams{ID: "c1", Title: "C1", Now: now})
	_ = child.TransitionTo(core.StatusInProgress, now)
	_ = child.SetProgress(100, now)

	// Single non-done subtask with progress 100 must produce 99%, never 100%
	got := core.CalculateProgress(*parent, []core.Task{*child})
	if got != 99 {
		t.Errorf("CalculateProgress with non-done subtask at 100 = %d, want 99", got)
	}
}
