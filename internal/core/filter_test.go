package core_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/newbpydev/tusk/internal/core"
)

func TestFilterTasks(t *testing.T) {
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	dueTomorrow := now.Add(24 * time.Hour)
	dueNextWeek := now.Add(7 * 24 * time.Hour)

	t1, _ := core.NewTask(core.NewTaskParams{
		ID:       "t1",
		Title:    "Backend API service",
		Priority: core.PriorityUrgent,
		Tags:     []string{"backend", "api"},
		DueDate:  &dueTomorrow,
		Now:      now,
	})
	_ = t1.TransitionTo(core.StatusInProgress, now)

	t2, _ := core.NewTask(core.NewTaskParams{
		ID:          "t2",
		Title:       "Frontend UI dashboard",
		Description: "Build with Bubble Tea",
		Priority:    core.PriorityHigh,
		Tags:        []string{"frontend", "ui"},
		DueDate:     &dueNextWeek,
		Now:         now,
	})

	parentID := "t1"
	t3, _ := core.NewTask(core.NewTaskParams{
		ID:       "t3",
		Title:    "Database schema migration",
		Priority: core.PriorityMedium,
		ParentID: &parentID,
		Tags:     []string{"backend", "db"},
		Now:      now,
	})
	_ = t3.TransitionTo(core.StatusDone, now)

	tasks := []core.Task{*t1, *t2, *t3}

	// 1. Filter by Status
	got := core.FilterTasks(tasks, core.TaskFilter{
		Statuses: []core.Status{core.StatusDone},
	})
	if len(got) != 1 || got[0].ID != "t3" {
		t.Errorf("FilterTasks by StatusDone got %v, want [t3]", got)
	}

	// 2. Filter by Priority
	got = core.FilterTasks(tasks, core.TaskFilter{
		Priorities: []core.Priority{core.PriorityUrgent, core.PriorityHigh},
	})
	if len(got) != 2 || got[0].ID != "t1" || got[1].ID != "t2" {
		t.Errorf("FilterTasks by Priority got %v, want [t1, t2]", got)
	}

	// 3. Filter by Tags (must match all specified tags)
	got = core.FilterTasks(tasks, core.TaskFilter{
		Tags: []core.Tag{core.Tag("backend"), core.Tag("api")},
	})
	if len(got) != 1 || got[0].ID != "t1" {
		t.Errorf("FilterTasks by Tags [backend, api] got %v, want [t1]", got)
	}

	// 4. Filter by RootOnly
	got = core.FilterTasks(tasks, core.TaskFilter{
		RootOnly: true,
	})
	if len(got) != 2 || got[0].ID != "t1" || got[1].ID != "t2" {
		t.Errorf("FilterTasks RootOnly got %v, want [t1, t2]", got)
	}

	// 5. Filter by specific ParentID
	got = core.FilterTasks(tasks, core.TaskFilter{
		ParentID: &parentID,
	})
	if len(got) != 1 || got[0].ID != "t3" {
		t.Errorf("FilterTasks by ParentID=t1 got %v, want [t3]", got)
	}

	// 6. Filter by DueBefore / DueAfter
	dueCutoff := now.Add(48 * time.Hour)
	got = core.FilterTasks(tasks, core.TaskFilter{
		DueBefore: &dueCutoff,
	})
	if len(got) != 1 || got[0].ID != "t1" {
		t.Errorf("FilterTasks by DueBefore got %v, want [t1]", got)
	}

	// 7. Filter by SearchTerm (case-insensitive substring in title or description)
	got = core.FilterTasks(tasks, core.TaskFilter{
		SearchTerm: "bubble tea",
	})
	if len(got) != 1 || got[0].ID != "t2" {
		t.Errorf("FilterTasks by SearchTerm 'bubble tea' got %v, want [t2]", got)
	}

	// 8. Zero match query returns empty non-nil slice
	got = core.FilterTasks(tasks, core.TaskFilter{
		SearchTerm: "non-existent query",
	})
	if got == nil || len(got) != 0 {
		t.Errorf("FilterTasks with zero matches expected empty non-nil slice, got %v", got)
	}

	// 9. Conflicting criteria: RootOnly and ParentID both set returns empty slice
	got = core.FilterTasks(tasks, core.TaskFilter{
		RootOnly: true,
		ParentID: &parentID,
	})
	if got == nil || len(got) != 0 {
		t.Errorf("FilterTasks with conflicting RootOnly and ParentID expected empty, got %v", got)
	}

	// 10. Empty tasks slice returns empty non-nil slice
	empty := core.FilterTasks(nil, core.TaskFilter{})
	if empty == nil || len(empty) != 0 {
		t.Errorf("FilterTasks on nil tasks expected empty non-nil slice, got %v", empty)
	}
}

func TestSortTasks_MultiKey(t *testing.T) {
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	dueEarly := now.Add(24 * time.Hour)
	dueLate := now.Add(72 * time.Hour)

	// Tasks:
	// A: PriorityUrgent, DueDate = nil, CreatedAt = now
	// B: PriorityHigh, DueDate = dueLate, CreatedAt = now+1h
	// C: PriorityHigh, DueDate = dueEarly, CreatedAt = now+2h
	// D: PriorityLow, DueDate = nil, CreatedAt = now+3h
	tA, _ := core.NewTask(core.NewTaskParams{ID: "A", Title: "A", Priority: core.PriorityUrgent, Now: now})
	tB, _ := core.NewTask(core.NewTaskParams{ID: "B", Title: "B", Priority: core.PriorityHigh, DueDate: &dueLate, Now: now.Add(1 * time.Hour)})
	tC, _ := core.NewTask(core.NewTaskParams{ID: "C", Title: "C", Priority: core.PriorityHigh, DueDate: &dueEarly, Now: now.Add(2 * time.Hour)})
	tD, _ := core.NewTask(core.NewTaskParams{ID: "D", Title: "D", Priority: core.PriorityLow, Now: now.Add(3 * time.Hour)})

	// 1. Sort by Priority DESC
	tasks := []core.Task{*tD, *tB, *tA, *tC}
	core.SortTasks(tasks, []core.SortOrder{
		{Field: core.SortByPriority, Direction: core.SortDesc},
	})
	// Priority order: A (Urgent=4), B & C (High=3), D (Low=1)
	// B & C tie on priority -> deterministic ID tie-breaker places B before C
	wantIDs := []string{"A", "B", "C", "D"}
	gotIDs := extractIDs(tasks)
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Errorf("SortTasks by Priority DESC = %v, want %v", gotIDs, wantIDs)
	}

	// 2. Sort by DueDate ASC (nil DueDate sorts last)
	tasks = []core.Task{*tA, *tD, *tB, *tC}
	core.SortTasks(tasks, []core.SortOrder{
		{Field: core.SortByDueDate, Direction: core.SortAsc},
	})
	// DueDate ASC: C (dueEarly), B (dueLate), then nil dates: A, D (tied on nil due date -> ID tie-breaker places A before D)
	wantDueIDs := []string{"C", "B", "A", "D"}
	gotDueIDs := extractIDs(tasks)
	if !reflect.DeepEqual(gotDueIDs, wantDueIDs) {
		t.Errorf("SortTasks by DueDate ASC = %v, want %v", gotDueIDs, wantDueIDs)
	}

	// 3. Sort by DueDate DESC (nil DueDate sorts first on DESC)
	tasks = []core.Task{*tC, *tB, *tA, *tD}
	core.SortTasks(tasks, []core.SortOrder{
		{Field: core.SortByDueDate, Direction: core.SortDesc},
	})
	// Nil dates first: A, D (tied -> ID tie-breaker: A, D), then B (dueLate), then C (dueEarly)
	wantDueDescIDs := []string{"A", "D", "B", "C"}
	gotDueDescIDs := extractIDs(tasks)
	if !reflect.DeepEqual(gotDueDescIDs, wantDueDescIDs) {
		t.Errorf("SortTasks by DueDate DESC = %v, want %v", gotDueDescIDs, wantDueDescIDs)
	}
	// 4. Deterministic multi-run stability: run 100 consecutive sorts
	wantMultiIDs := []string{"A", "C", "B", "D"}
	for i := 0; i < 100; i++ {
		shuffled := []core.Task{*tC, *tA, *tD, *tB}
		core.SortTasks(shuffled, []core.SortOrder{
			{Field: core.SortByPriority, Direction: core.SortDesc},
			{Field: core.SortByDueDate, Direction: core.SortAsc},
		})
		runIDs := extractIDs(shuffled)
		if !reflect.DeepEqual(runIDs, wantMultiIDs) {
			t.Fatalf("run %d produced non-deterministic sort order: %v, want %v", i, runIDs, wantMultiIDs)
		}
	}
}

func extractIDs(tasks []core.Task) []string {
	ids := make([]string, len(tasks))
	for i, t := range tasks {
		ids[i] = t.ID
	}
	return ids
}

func TestFilterTasks_DeepCopy(t *testing.T) {
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	parentID := "p1"
	dueDate := now.Add(24 * time.Hour)
	completedAt := now.Add(12 * time.Hour)
	tags := []core.Tag{core.Tag("work")}

	task := core.Task{
		ID:          "task-1",
		Title:       "Original",
		Status:      core.StatusTodo,
		Priority:    core.PriorityHigh,
		ParentID:    &parentID,
		Tags:        tags,
		DueDate:     &dueDate,
		CompletedAt: &completedAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	filtered := core.FilterTasks([]core.Task{task}, core.TaskFilter{})
	if len(filtered) != 1 {
		t.Fatalf("expected 1 filtered task, got %d", len(filtered))
	}

	// Mutate filtered task reference fields
	*filtered[0].ParentID = "mutated-parent"
	*filtered[0].DueDate = dueDate.Add(48 * time.Hour)
	*filtered[0].CompletedAt = completedAt.Add(48 * time.Hour)
	filtered[0].Tags[0] = core.Tag("mutated-tag")

	// Verify original task fields are untouched
	if *task.ParentID != "p1" {
		t.Errorf("original ParentID was mutated to %s", *task.ParentID)
	}
	if !task.DueDate.Equal(now.Add(24 * time.Hour)) {
		t.Errorf("original DueDate was mutated to %v", task.DueDate)
	}
	if !task.CompletedAt.Equal(now.Add(12 * time.Hour)) {
		t.Errorf("original CompletedAt was mutated to %v", task.CompletedAt)
	}
	if task.Tags[0] != core.Tag("work") {
		t.Errorf("original Tag was mutated to %s", task.Tags[0])
	}

	// Exercise nil and empty reference fields
	minimalTask := core.Task{
		ID:        "min-1",
		Title:     "Minimal",
		Status:    core.StatusTodo,
		Priority:  core.PriorityLow,
		CreatedAt: now,
		UpdatedAt: now,
	}
	emptyTagsTask := minimalTask
	emptyTagsTask.ID = "empty-tags"
	emptyTagsTask.Tags = []core.Tag{}

	filteredBatch := core.FilterTasks([]core.Task{minimalTask, emptyTagsTask}, core.TaskFilter{})
	if len(filteredBatch) != 2 {
		t.Fatalf("expected 2 filtered tasks, got %d", len(filteredBatch))
	}

	// Assert nil fields remain nil
	if filteredBatch[0].ParentID != nil || filteredBatch[0].DueDate != nil || filteredBatch[0].CompletedAt != nil || filteredBatch[0].Tags != nil {
		t.Errorf("expected nil reference fields on minimal task, got %+v", filteredBatch[0])
	}
	// Assert empty slice remains empty non-nil
	if filteredBatch[1].Tags == nil || len(filteredBatch[1].Tags) != 0 {
		t.Errorf("expected empty non-nil Tags, got %v", filteredBatch[1].Tags)
	}
}
