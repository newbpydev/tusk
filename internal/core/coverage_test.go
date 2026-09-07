package core_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/tusk/internal/core"
)

func TestCoverage_FilterAndSortFields(t *testing.T) {
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	dueEarly := now.Add(24 * time.Hour)

	t1, _ := core.NewTask(core.NewTaskParams{ID: "1", Title: "Zebra", DueDate: &dueEarly, Now: now.Add(1 * time.Hour)})
	t2, _ := core.NewTask(core.NewTaskParams{ID: "2", Title: "Alpha", Now: now.Add(2 * time.Hour)})

	// Sort by Title ASC
	tasks := []core.Task{*t1, *t2}
	core.SortTasks(tasks, []core.SortOrder{{Field: core.SortByTitle, Direction: core.SortAsc}})
	if tasks[0].ID != "2" || tasks[1].ID != "1" {
		t.Errorf("SortByTitle ASC failed: got %v", tasks)
	}

	// Sort by Title DESC
	core.SortTasks(tasks, []core.SortOrder{{Field: core.SortByTitle, Direction: core.SortDesc}})
	if tasks[0].ID != "1" || tasks[1].ID != "2" {
		t.Errorf("SortByTitle DESC failed: got %v", tasks)
	}

	// Sort by CreatedAt ASC
	core.SortTasks(tasks, []core.SortOrder{{Field: core.SortByCreatedAt, Direction: core.SortAsc}})
	if tasks[0].ID != "1" || tasks[1].ID != "2" {
		t.Errorf("SortByCreatedAt ASC failed: got %v", tasks)
	}

	// Sort by CreatedAt DESC
	core.SortTasks(tasks, []core.SortOrder{{Field: core.SortByCreatedAt, Direction: core.SortDesc}})
	if tasks[0].ID != "2" || tasks[1].ID != "1" {
		t.Errorf("SortByCreatedAt DESC failed: got %v", tasks)
	}

	// Sort by ID DESC
	core.SortTasks(tasks, []core.SortOrder{{Field: core.SortByID, Direction: core.SortDesc}})
	if tasks[0].ID != "2" || tasks[1].ID != "1" {
		t.Errorf("SortByID DESC failed: got %v", tasks)
	}

	// Unknown field or equal field
	core.SortTasks(tasks, []core.SortOrder{{Field: core.SortField("unknown"), Direction: core.SortAsc}})

	// Tasks with identical title and created at
	tIdentical1, _ := core.NewTask(core.NewTaskParams{ID: "A", Title: "Same", Now: now})
	tIdentical2, _ := core.NewTask(core.NewTaskParams{ID: "B", Title: "Same", Now: now})
	core.SortTasks([]core.Task{*tIdentical1, *tIdentical2}, []core.SortOrder{
		{Field: core.SortByTitle, Direction: core.SortAsc},
		{Field: core.SortByCreatedAt, Direction: core.SortAsc},
		{Field: core.SortByID, Direction: core.SortAsc},
	})

	// Sort with single task or empty tasks
	core.SortTasks(nil, nil)
	single := []core.Task{*t1}
	core.SortTasks(single, nil)

	// Filter with DueAfter
	dueCutoff := now.Add(12 * time.Hour)
	matched := core.FilterTasks(tasks, core.TaskFilter{DueAfter: &dueCutoff})
	if len(matched) != 1 || matched[0].ID != "1" {
		t.Errorf("FilterTasks DueAfter failed: got %v", matched)
	}
}

func TestCoverage_TaskFallbacksAndValidations(t *testing.T) {
	// Zero Now in NewTask uses time.Now()
	task, err := core.NewTask(core.NewTaskParams{
		ID:    "t-now",
		Title: "Zero Now Task",
	})
	if err != nil {
		t.Fatalf("NewTask with zero Now failed: %v", err)
	}
	if task.Priority != core.PriorityMedium {
		t.Errorf("expected PriorityMedium default, got %v", task.Priority)
	}
	if task.CreatedAt.IsZero() || task.UpdatedAt.IsZero() {
		t.Errorf("expected non-zero timestamps")
	}

	// NewTask with invalid tag returns error
	_, err = core.NewTask(core.NewTaskParams{
		ID:    "t-tag-err",
		Title: "Tag Err Task",
		Tags:  []string{"invalid tag"},
	})
	if !errors.Is(err, core.ErrInvalidTag) {
		t.Errorf("expected ErrInvalidTag on invalid tag in NewTask, got %v", err)
	}

	// TransitionTo with same status is no-op
	if err := task.TransitionTo(core.StatusTodo, time.Time{}); err != nil {
		t.Errorf("TransitionTo same status failed: %v", err)
	}

	// TransitionTo with zero Now uses time.Now()
	if err := task.TransitionTo(core.StatusInProgress, time.Time{}); err != nil {
		t.Errorf("TransitionTo zero Now failed: %v", err)
	}

	// Update with zero Now and invalid priority
	if err := task.Update("Title", "Desc", core.Priority(99), nil, nil, time.Time{}); !errors.Is(err, core.ErrInvalidPriority) {
		t.Errorf("expected ErrInvalidPriority, got %v", err)
	}
	if err := task.Update(strings.Repeat("x", 256), "Desc", core.PriorityLow, nil, nil, time.Time{}); !errors.Is(err, core.ErrTitleTooLong) {
		t.Errorf("expected ErrTitleTooLong, got %v", err)
	}
	if err := task.Update("Valid", "Desc", core.PriorityLow, nil, nil, time.Time{}); err != nil {
		t.Errorf("Update with zero Now failed: %v", err)
	}

	// SetParent with zero Now
	pID := "parent-1"
	if err := task.SetParent(&pID, time.Time{}); err != nil {
		t.Errorf("SetParent with zero Now failed: %v", err)
	}

	// SetProgress with zero Now
	if err := task.SetProgress(30, time.Time{}); err != nil {
		t.Errorf("SetProgress with zero Now failed: %v", err)
	}
}

func TestCoverage_RollupClamping(t *testing.T) {
	// Leaf task with progress >= 100 but not StatusDone clamps to 99
	taskOver := core.Task{Status: core.StatusInProgress, Progress: 120}
	if got := core.CalculateProgress(taskOver, nil); got != 99 {
		t.Errorf("CalculateProgress leaf over 100 = %d, want 99", got)
	}

	// Leaf task with negative progress clamps to 0
	taskUnder := core.Task{Status: core.StatusInProgress, Progress: -10}
	if got := core.CalculateProgress(taskUnder, nil); got != 0 {
		t.Errorf("CalculateProgress leaf under 0 = %d, want 0", got)
	}

	// Subtask progress clamped
	parent := core.Task{Status: core.StatusInProgress}
	sub1 := core.Task{Status: core.StatusInProgress, Progress: -20} // clamped to 0
	sub2 := core.Task{Status: core.StatusInProgress, Progress: 200} // non-done clamped to 99
	if got := core.CalculateProgress(parent, []core.Task{sub1, sub2}); got != 49 {
		t.Errorf("CalculateProgress clamped subtasks = %d, want 49", got)
	}
}

func TestCoverage_PriorityStringAndParse(t *testing.T) {
	if s := core.Priority(99).String(); s != "" {
		t.Errorf("expected empty string for Priority(99), got %q", s)
	}

	// Numeric parse path for Priority valid
	p, err := core.ParsePriority("2")
	if err != nil || p != core.PriorityMedium {
		t.Errorf("ParsePriority('2') = %v, %v, want PriorityMedium", p, err)
	}

	// Numeric parse path for Priority invalid integer
	_, err = core.ParsePriority("42")
	if !errors.Is(err, core.ErrInvalidPriority) {
		t.Errorf("expected ErrInvalidPriority for '42', got %v", err)
	}
}

func TestCoverage_TreeErrorsAndLoops(t *testing.T) {
	// Empty tasks BuildTree
	empty, err := core.BuildTree(nil)
	if err != nil || len(empty) != 0 {
		t.Errorf("BuildTree(nil) failed: %v, %v", empty, err)
	}

	// lookupParent error in DetectCycles
	errExpected := errors.New("db error")
	err = core.DetectCycles("A", ptr("B"), func(id string) (*string, error) {
		return nil, errExpected
	})
	if !errors.Is(err, errExpected) {
		t.Errorf("expected db error in DetectCycles, got %v", err)
	}

	// Existing loop in DetectCycles
	err = core.DetectCycles("A", ptr("B"), func(id string) (*string, error) {
		return ptr("B"), nil // B points to B
	})
	if !errors.Is(err, core.ErrCyclicDependency) {
		t.Errorf("expected ErrCyclicDependency on existing loop, got %v", err)
	}

	// lookupParent error in ValidateHierarchyDepth
	err = core.ValidateHierarchyDepth(0, "B", func(id string) (*string, error) {
		return nil, errExpected
	})
	if !errors.Is(err, errExpected) {
		t.Errorf("expected db error in ValidateHierarchyDepth, got %v", err)
	}

	// Loop in ValidateHierarchyDepth
	err = core.ValidateHierarchyDepth(0, "B", func(id string) (*string, error) {
		return ptr("B"), nil
	})
	if !errors.Is(err, core.ErrCyclicDependency) {
		t.Errorf("expected ErrCyclicDependency on loop in ValidateHierarchyDepth, got %v", err)
	}

	// Parent depth exceeding MaxHierarchyDepth in ValidateHierarchyDepth
	count := 0
	err = core.ValidateHierarchyDepth(0, "P", func(id string) (*string, error) {
		count++
		if count > 15 {
			return nil, nil
		}
		next := parentsKey(count)
		return &next, nil
	})
	if !errors.Is(err, core.ErrMaxDepthExceeded) {
		t.Errorf("expected ErrMaxDepthExceeded on deep parent, got %v", err)
	}
}
