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
func TestCalculateProgress_Mutations(t *testing.T) {
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	parent, _ := core.NewTask(core.NewTaskParams{ID: "parent", Title: "Parent", Now: now})
	_ = parent.TransitionTo(core.StatusInProgress, now)
	_ = parent.SetProgress(20, now) // initial manual progress 20%

	// 1. Initial state without subtasks: preserves manual progress
	if got := core.CalculateProgress(*parent, nil); got != 20 {
		t.Errorf("expected 20%% manual progress, got %d", got)
	}

	// 2. Add subtask 1 (in-progress, 60%)
	c1, _ := core.NewTask(core.NewTaskParams{ID: "c1", Title: "C1", Now: now})
	_ = c1.TransitionTo(core.StatusInProgress, now)
	_ = c1.SetProgress(60, now)
	subtasks := []core.Task{*c1}
	if got := core.CalculateProgress(*parent, subtasks); got != 60 {
		t.Errorf("after adding c1: got %d, want 60", got)
	}

	// 3. Add subtask 2 (todo, 0%)
	c2, _ := core.NewTask(core.NewTaskParams{ID: "c2", Title: "C2", Now: now})
	subtasks = append(subtasks, *c2)
	if got := core.CalculateProgress(*parent, subtasks); got != 30 {
		t.Errorf("after adding c2: got %d, want 30", got)
	}

	// 4. Mark both subtasks done -> strictly 100%
	_ = c1.TransitionTo(core.StatusDone, now)
	_ = c2.TransitionTo(core.StatusDone, now)
	subtasks = []core.Task{*c1, *c2}
	if got := core.CalculateProgress(*parent, subtasks); got != 100 {
		t.Errorf("after marking all done: got %d, want 100", got)
	}

	// 5. Reopen subtask 1 to in-progress -> progress resets to 0, parent drops to 50%
	_ = c1.TransitionTo(core.StatusInProgress, now)
	subtasks = []core.Task{*c1, *c2}
	if got := core.CalculateProgress(*parent, subtasks); got != 50 {
		t.Errorf("after reopening c1: got %d, want 50", got)
	}

	// 6. Delete/remove subtask 1 -> remaining subtask c2 is done -> parent becomes 100%
	subtasks = []core.Task{*c2}
	if got := core.CalculateProgress(*parent, subtasks); got != 100 {
		t.Errorf("after removing c1: got %d, want 100", got)
	}

	// 7. Delete/remove all subtasks -> parent preserves manual progress
	subtasks = []core.Task{}
	if got := core.CalculateProgress(*parent, subtasks); got != 20 {
		t.Errorf("after removing all subtasks: got %d, want 20", got)
	}
}

func TestCalculateProgress_NestedHierarchy100(t *testing.T) {
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	grandparent, _ := core.NewTask(core.NewTaskParams{ID: "gp", Title: "Grandparent", Now: now})

	// Intermediate parent whose subtasks are done: has progress 100 via rollup
	parent, _ := core.NewTask(core.NewTaskParams{ID: "p", Title: "Parent", Now: now})
	_ = parent.TransitionTo(core.StatusInProgress, now)
	_ = parent.SetRollupProgress(100, now)

	// Direct child that is done
	childDone, _ := core.NewTask(core.NewTaskParams{ID: "c-done", Title: "Child Done", Now: now})
	_ = childDone.TransitionTo(core.StatusDone, now)

	// Grandparent subtasks are [parent (progress 100, in-progress), childDone (done)]
	gpSubtasks := []core.Task{*parent, *childDone}
	got := core.CalculateProgress(*grandparent, gpSubtasks)
	if got != 100 {
		t.Errorf("expected grandparent CalculateProgress = 100 with 100%% intermediate subtasks, got %d", got)
	}
}
