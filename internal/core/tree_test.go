package core_test

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/tusk/internal/core"
)

func TestDetectCycles_DirectSelf(t *testing.T) {
	selfID := "task-A"
	err := core.DetectCycles(selfID, &selfID, func(id string) (*string, error) {
		return nil, nil
	})
	if !errors.Is(err, core.ErrSelfParenting) {
		t.Errorf("expected ErrSelfParenting, got %v", err)
	}
}

func TestDetectCycles_RootPromotion(t *testing.T) {
	lookupCalled := false
	err := core.DetectCycles("task-A", nil, func(id string) (*string, error) {
		lookupCalled = true
		return nil, nil
	})
	if err != nil {
		t.Errorf("expected nil error on root promotion, got %v", err)
	}
	if lookupCalled {
		t.Errorf("expected lookupParent not to be called on root promotion")
	}
}

func TestDetectCycles_TwoNodeLoop(t *testing.T) {
	// Existing: taskB -> taskA (taskB's parent is taskA)
	parents := map[string]*string{
		"taskB": ptr("taskA"),
		"taskA": nil,
	}

	lookup := func(id string) (*string, error) {
		return parents[id], nil
	}

	// Propose: taskA -> taskB (setting taskA's parent to taskB)
	err := core.DetectCycles("taskA", ptr("taskB"), lookup)
	if !errors.Is(err, core.ErrCyclicDependency) {
		t.Errorf("expected ErrCyclicDependency, got %v", err)
	}
}

func TestDetectCycles_DeepLoop(t *testing.T) {
	// Chain: E -> D -> C -> B -> A (E parent is D, D parent is C, ...)
	parents := map[string]*string{
		"E": ptr("D"),
		"D": ptr("C"),
		"C": ptr("B"),
		"B": ptr("A"),
		"A": nil,
	}

	lookup := func(id string) (*string, error) {
		return parents[id], nil
	}

	// Propose: A.parent = E -> creates A -> E -> D -> C -> B -> A
	err := core.DetectCycles("A", ptr("E"), lookup)
	if !errors.Is(err, core.ErrCyclicDependency) {
		t.Errorf("expected ErrCyclicDependency, got %v", err)
	}

	// Propose valid: new task F.parent = E -> valid chain
	err = core.DetectCycles("F", ptr("E"), lookup)
	if err != nil {
		t.Errorf("expected valid attachment, got %v", err)
	}

	// 1001-node ring in DetectCycles must return ErrCyclicDependency
	err = core.DetectCycles("new-task", ptr("L1001"), makeCycleLookup(1001))
	if !errors.Is(err, core.ErrCyclicDependency) {
		t.Errorf("expected ErrCyclicDependency for 1001-node ring in DetectCycles, got %v", err)
	}

	// 1002-node ring in DetectCycles must return ErrCyclicDependency
	err = core.DetectCycles("new-task", ptr("L1002"), makeCycleLookup(1002))
	if !errors.Is(err, core.ErrCyclicDependency) {
		t.Errorf("expected ErrCyclicDependency for 1002-node ring in DetectCycles, got %v", err)
	}

	// 10001-node ring in DetectCycles must return ErrCyclicDependency
	err = core.DetectCycles("new-task", ptr("L10001"), makeCycleLookup(10001))
	if !errors.Is(err, core.ErrCyclicDependency) {
		t.Errorf("expected ErrCyclicDependency for 10001-node ring in DetectCycles, got %v", err)
	}

	// 20001-node ring in DetectCycles must return ErrCyclicDependency
	err = core.DetectCycles("new-task", ptr("L20001"), makeCycleLookup(20001))
	if !errors.Is(err, core.ErrCyclicDependency) {
		t.Errorf("expected ErrCyclicDependency for 20001-node ring in DetectCycles, got %v", err)
	}
}

func TestValidateHierarchyDepth_Subtree(t *testing.T) {
	// 8-level chain: N8 -> N7 -> ... -> N1 (root N1 depth = 1, N8 depth = 8)
	parents := map[string]*string{
		"N8": ptr("N7"),
		"N7": ptr("N6"),
		"N6": ptr("N5"),
		"N5": ptr("N4"),
		"N4": ptr("N3"),
		"N3": ptr("N2"),
		"N2": ptr("N1"),
		"N1": nil,
	}

	lookup := func(id string) (*string, error) {
		return parents[id], nil
	}

	// Proposing to attach a leaf (taskSubtreeDepth = 0) under N8:
	// parentDepth(8) + 1 + 0 = 9 <= 10 -> valid
	err := core.ValidateHierarchyDepth(0, "N8", lookup)
	if err != nil {
		t.Errorf("expected valid depth 9, got %v", err)
	}

	// Proposing to attach a 3-level subtree (taskSubtreeDepth = 2) under N8:
	// parentDepth(8) + 1 + 2 = 11 > 10 -> ErrMaxDepthExceeded
	err = core.ValidateHierarchyDepth(2, "N8", lookup)
	if !errors.Is(err, core.ErrMaxDepthExceeded) {
		t.Errorf("expected ErrMaxDepthExceeded, got %v", err)
	}

	// Negative subtree depth must be rejected with ErrInvalidDepth
	err = core.ValidateHierarchyDepth(-1, "N8", lookup)
	if !errors.Is(err, core.ErrInvalidDepth) {
		t.Errorf("expected ErrInvalidDepth for negative subtree depth, got %v", err)
	}

	// 11-node cyclic loop: L11 -> L10 -> ... -> L1 -> L11
	err = core.ValidateHierarchyDepth(0, "L11", makeCycleLookup(11))
	if !errors.Is(err, core.ErrCyclicDependency) {
		t.Errorf("expected ErrCyclicDependency for 11-node cyclic loop, got %v", err)
	}

	// 1001-node cyclic loop: L1001 -> L1000 -> ... -> L1 -> L1001
	err = core.ValidateHierarchyDepth(0, "L1001", makeCycleLookup(1001))
	if !errors.Is(err, core.ErrCyclicDependency) {
		t.Errorf("expected ErrCyclicDependency for 1001-node cyclic loop, got %v", err)
	}

	// Acyclic chain exceeding MaxHierarchyDepth returns ErrMaxDepthExceeded
	acyclicParents := make(map[string]*string)
	for i := 2; i <= 1005; i++ {
		acyclicParents[fmt.Sprintf("A%d", i)] = ptr(fmt.Sprintf("A%d", i-1))
	}
	acyclicParents["A1"] = nil
	lookupAcyclic := func(id string) (*string, error) {
		return acyclicParents[id], nil
	}
	err = core.ValidateHierarchyDepth(0, "A1005", lookupAcyclic)
	if !errors.Is(err, core.ErrMaxDepthExceeded) {
		t.Errorf("expected ErrMaxDepthExceeded for over-MaxHierarchyDepth chain, got %v", err)
	}

	// Infinite chain exceeding maxTraversalSteps returns ErrTraversalLimitExceeded
	lookupInfiniteDepth := func(id string) (*string, error) {
		var n int
		fmt.Sscanf(id, "node-%d", &n)
		next := fmt.Sprintf("node-%d", n+1)
		return &next, nil
	}
	err = core.ValidateHierarchyDepth(0, "node-1", lookupInfiniteDepth)
	if !errors.Is(err, core.ErrTraversalLimitExceeded) {
		t.Errorf("expected ErrTraversalLimitExceeded for infinite chain, got %v", err)
	}

	// Overflow guard: taskSubtreeDepth = math.MaxInt must return ErrMaxDepthExceeded
	err = core.ValidateHierarchyDepth(math.MaxInt, "N8", lookup)
	if !errors.Is(err, core.ErrMaxDepthExceeded) {
		t.Errorf("expected ErrMaxDepthExceeded for math.MaxInt taskSubtreeDepth, got %v", err)
	}

	// 1002-node ring in ValidateHierarchyDepth must return ErrCyclicDependency
	err = core.ValidateHierarchyDepth(0, "L1002", makeCycleLookup(1002))
	if !errors.Is(err, core.ErrCyclicDependency) {
		t.Errorf("expected ErrCyclicDependency for 1002-node ring in ValidateHierarchyDepth, got %v", err)
	}

	// 10001-node ring in ValidateHierarchyDepth must return ErrCyclicDependency
	err = core.ValidateHierarchyDepth(0, "L10001", makeCycleLookup(10001))
	if !errors.Is(err, core.ErrCyclicDependency) {
		t.Errorf("expected ErrCyclicDependency for 10001-node ring in ValidateHierarchyDepth, got %v", err)
	}

	// 20001-node ring in ValidateHierarchyDepth must return ErrCyclicDependency
	err = core.ValidateHierarchyDepth(0, "L20001", makeCycleLookup(20001))
	if !errors.Is(err, core.ErrCyclicDependency) {
		t.Errorf("expected ErrCyclicDependency for 20001-node ring in ValidateHierarchyDepth, got %v", err)
	}
	// When taskSubtreeDepth > MaxHierarchyDepth and parent is cyclic,
	// cycle detection takes precedence over depth rejection
	cyclicParent := map[string]*string{
		"B": ptr("B"),
	}
	lookupSelf := func(id string) (*string, error) {
		return cyclicParent[id], nil
	}
	err = core.ValidateHierarchyDepth(11, "B", lookupSelf)
	if !errors.Is(err, core.ErrCyclicDependency) {
		t.Errorf("expected ErrCyclicDependency to take precedence over oversized subtree depth, got %v", err)
	}
}

func TestBuildTree_Forest(t *testing.T) {
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)

	// Two roots: Root1 has Child1, Child1 has Grandchild1. Root2 has Child2.
	r1, _ := core.NewTask(core.NewTaskParams{ID: "r1", Title: "Root 1", Now: now})
	c1, _ := core.NewTask(core.NewTaskParams{ID: "c1", Title: "Child 1", ParentID: ptr("r1"), Now: now})
	g1, _ := core.NewTask(core.NewTaskParams{ID: "g1", Title: "Grandchild 1", ParentID: ptr("c1"), Now: now})

	r2, _ := core.NewTask(core.NewTaskParams{ID: "r2", Title: "Root 2", Now: now})
	c2, _ := core.NewTask(core.NewTaskParams{ID: "c2", Title: "Child 2", ParentID: ptr("r2"), Now: now})

	tasks := []core.Task{*g1, *c2, *r1, *c1, *r2} // unordered
	forest, err := core.BuildTree(tasks)
	if err != nil {
		t.Fatalf("BuildTree failed: %v", err)
	}

	if len(forest) != 2 {
		t.Fatalf("expected 2 root nodes, got %d", len(forest))
	}

	// Verify depths
	for _, root := range forest {
		if root.Depth != 1 {
			t.Errorf("root %s depth = %d, want 1", root.Task.ID, root.Depth)
		}
		if root.Task.ID == "r1" {
			if len(root.Children) != 1 {
				t.Fatalf("r1 expected 1 child, got %d", len(root.Children))
			}
			child := root.Children[0]
			if child.Depth != 2 || child.Task.ID != "c1" {
				t.Errorf("expected child c1 depth 2, got %s depth %d", child.Task.ID, child.Depth)
			}
			if len(child.Children) != 1 {
				t.Fatalf("c1 expected 1 grandchild, got %d", len(child.Children))
			}
			grandchild := child.Children[0]
			if grandchild.Depth != 3 || grandchild.Task.ID != "g1" {
				t.Errorf("expected grandchild g1 depth 3, got %s depth %d", grandchild.Task.ID, grandchild.Depth)
			}
		}
	}
}

func TestBuildTree_Errors(t *testing.T) {
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)

	// Orphan task: references parent "missing-parent" not in slice
	orphan, _ := core.NewTask(core.NewTaskParams{ID: "orphan", Title: "Orphan", ParentID: ptr("missing-parent"), Now: now})
	_, err := core.BuildTree([]core.Task{*orphan})
	if !errors.Is(err, core.ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound for orphan, got %v", err)
	}

	// Cyclic loop with no roots: A -> B -> A
	taskA, _ := core.NewTask(core.NewTaskParams{ID: "A", Title: "A", ParentID: ptr("B"), Now: now})
	taskB, _ := core.NewTask(core.NewTaskParams{ID: "B", Title: "B", ParentID: ptr("A"), Now: now})
	_, err = core.BuildTree([]core.Task{*taskA, *taskB})
	if !errors.Is(err, core.ErrCyclicDependency) {
		t.Errorf("expected ErrCyclicDependency for unrooted cycle, got %v", err)
	}

	// Duplicate task ID in tasks slice
	tDup1, _ := core.NewTask(core.NewTaskParams{ID: "dup-id", Title: "Task 1", Now: now})
	tDup2, _ := core.NewTask(core.NewTaskParams{ID: "dup-id", Title: "Task 2", Now: now})
	_, err = core.BuildTree([]core.Task{*tDup1, *tDup2})
	if !errors.Is(err, core.ErrDuplicateTaskID) {
		t.Errorf("expected ErrDuplicateTaskID, got %v", err)
	}

	// Chain of 11 tasks exceeding MaxHierarchyDepth (10)
	var chainTasks []core.Task
	root, _ := core.NewTask(core.NewTaskParams{ID: "d1", Title: "Depth 1", Now: now})
	chainTasks = append(chainTasks, *root)
	for i := 2; i <= 11; i++ {
		parentID := fmt.Sprintf("d%d", i-1)
		currID := fmt.Sprintf("d%d", i)
		tChild, _ := core.NewTask(core.NewTaskParams{ID: currID, Title: currID, ParentID: &parentID, Now: now})
		chainTasks = append(chainTasks, *tChild)
	}
	_, err = core.BuildTree(chainTasks)
	if !errors.Is(err, core.ErrMaxDepthExceeded) {
		t.Errorf("expected ErrMaxDepthExceeded for 11-level chain, got %v", err)
	}

	// Graph containing BOTH an 11-level rooted chain AND a disconnected cycle
	// Cycle detection must take precedence over depth limits (returns ErrCyclicDependency, not ErrMaxDepthExceeded)
	chainWithCycle := make([]core.Task, len(chainTasks))
	copy(chainWithCycle, chainTasks)
	c1P := "cycle-2"
	c2P := "cycle-1"
	chainWithCycle = append(chainWithCycle,
		core.Task{ID: "cycle-1", Title: "C1", ParentID: &c1P},
		core.Task{ID: "cycle-2", Title: "C2", ParentID: &c2P},
	)
	_, err = core.BuildTree(chainWithCycle)
	if !errors.Is(err, core.ErrCyclicDependency) {
		t.Errorf("expected ErrCyclicDependency when both deep chain and cycle are present, got %v", err)
	}

	// Empty ID in BuildTree
	_, err = core.BuildTree([]core.Task{{ID: ""}})
	if !errors.Is(err, core.ErrInvalidTaskID) {
		t.Errorf("expected ErrInvalidTaskID for empty task ID, got %v", err)
	}

	// Empty ParentID in BuildTree
	emptyP := "   "
	_, err = core.BuildTree([]core.Task{{ID: "valid", ParentID: &emptyP}})
	if !errors.Is(err, core.ErrInvalidTaskID) {
		t.Errorf("expected ErrInvalidTaskID for empty parent ID, got %v", err)
	}

	// Self parent in BuildTree
	selfP := "self"
	_, err = core.BuildTree([]core.Task{{ID: "self", ParentID: &selfP}})
	if !errors.Is(err, core.ErrCyclicDependency) || !errors.Is(err, core.ErrSelfParenting) {
		t.Errorf("expected ErrCyclicDependency/ErrSelfParenting, got %v", err)
	}

	// Whitespace-padded ID in BuildTree
	_, err = core.BuildTree([]core.Task{{ID: " root "}})
	if !errors.Is(err, core.ErrInvalidTaskID) {
		t.Errorf("expected ErrInvalidTaskID for whitespace-padded ID, got %v", err)
	}

	// Whitespace-padded ParentID in BuildTree
	padP := " parent "
	_, err = core.BuildTree([]core.Task{{ID: "child", ParentID: &padP}})
	if !errors.Is(err, core.ErrInvalidTaskID) {
		t.Errorf("expected ErrInvalidTaskID for whitespace-padded ParentID, got %v", err)
	}
}

func TestBuildTree_DeepCopy(t *testing.T) {
	parentID := "root"
	dueDate := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
	completedAt := time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)
	tags := []core.Tag{core.Tag("work")}

	tasks := []core.Task{
		{
			ID:    "root",
			Title: "Root",
		},
		{
			ID:          "child",
			Title:       "Child",
			ParentID:    &parentID,
			DueDate:     &dueDate,
			CompletedAt: &completedAt,
			Tags:        tags,
		},
	}

	forest, err := core.BuildTree(tasks)
	if err != nil {
		t.Fatalf("BuildTree failed: %v", err)
	}

	if len(forest) != 1 || len(forest[0].Children) != 1 {
		t.Fatalf("expected 1 root with 1 child")
	}

	childNode := forest[0].Children[0]

	// Mutate input tasks pointers/slices
	parentID = "mutated"
	dueDate = dueDate.Add(24 * time.Hour)
	completedAt = completedAt.Add(24 * time.Hour)
	tags[0] = core.Tag("mutated")

	// Verify childNode.Task remained untouched
	if *childNode.Task.ParentID != "root" {
		t.Errorf("childNode.Task.ParentID mutated: got %s, want root", *childNode.Task.ParentID)
	}
	if !childNode.Task.DueDate.Equal(time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("childNode.Task.DueDate mutated: got %v", childNode.Task.DueDate)
	}
	if !childNode.Task.CompletedAt.Equal(time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("childNode.Task.CompletedAt mutated: got %v", childNode.Task.CompletedAt)
	}
	if childNode.Task.Tags[0] != core.Tag("work") {
		t.Errorf("childNode.Task.Tags mutated: got %s, want work", childNode.Task.Tags[0])
	}
}

func BenchmarkTreeTraversal(b *testing.B) {
	// 10-level chain: L10 -> L9 -> ... -> L1 (root)
	parents := make(map[string]*string)
	parents["L1"] = nil
	for i := 2; i <= 10; i++ {
		p := ptr(parentsKey(i - 1))
		parents[parentsKey(i)] = p
	}

	lookup := func(id string) (*string, error) {
		return parents[id], nil
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = core.DetectCycles("new-task", ptr("L10"), lookup)
	}
}

func BenchmarkBuildTree(b *testing.B) {
	// 100 tasks: 10 roots each with a 9-level subtree
	tasks := make([]core.Task, 0, 100)
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	tags := []core.Tag{core.Tag("core"), core.Tag("benchmark")}

	for r := range 10 {
		rootID := fmt.Sprintf("root-%d", r)
		tasks = append(tasks, core.Task{
			ID:        rootID,
			Title:     "Root Task",
			Status:    core.StatusTodo,
			Priority:  core.PriorityHigh,
			Tags:      tags,
			CreatedAt: now,
			UpdatedAt: now,
		})
		for c := 1; c < 10; c++ {
			childID := fmt.Sprintf("child-%d-%d", r, c)
			var parentID string
			if c == 1 {
				parentID = rootID
			} else {
				parentID = fmt.Sprintf("child-%d-%d", r, c-1)
			}
			dueDate := now.Add(time.Duration(c) * 24 * time.Hour)
			p := parentID
			tasks = append(tasks, core.Task{
				ID:        childID,
				Title:     "Child Task",
				ParentID:  &p,
				Status:    core.StatusInProgress,
				Priority:  core.PriorityMedium,
				Tags:      tags,
				DueDate:   &dueDate,
				CreatedAt: now,
				UpdatedAt: now,
			})
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_, err := core.BuildTree(tasks)
		if err != nil {
			b.Fatalf("BuildTree failed: %v", err)
		}
	}
}

func makeCycleLookup(size int) func(id string) (*string, error) {
	parents := make(map[string]*string, size)
	for i := 1; i <= size; i++ {
		var next string
		if i == 1 {
			next = fmt.Sprintf("L%d", size)
		} else {
			next = fmt.Sprintf("L%d", i-1)
		}
		parents[fmt.Sprintf("L%d", i)] = ptr(next)
	}
	return func(id string) (*string, error) {
		return parents[id], nil
	}
}

func ptr(s string) *string {
	return &s
}

func parentsKey(n int) string {
	switch n {
	case 1:
		return "L1"
	case 2:
		return "L2"
	case 3:
		return "L3"
	case 4:
		return "L4"
	case 5:
		return "L5"
	case 6:
		return "L6"
	case 7:
		return "L7"
	case 8:
		return "L8"
	case 9:
		return "L9"
	case 10:
		return "L10"
	default:
		return ""
	}
}

func TestDetectCycles_TraversalLimitExceeded(t *testing.T) {
	// Chain longer than 1000 steps without cycles
	lookupInfinite := func(id string) (*string, error) {
		var n int
		fmt.Sscanf(id, "node-%d", &n)
		next := fmt.Sprintf("node-%d", n+1)
		return &next, nil
	}

	start := "node-1"
	err := core.DetectCycles("target", &start, lookupInfinite)
	if !errors.Is(err, core.ErrTraversalLimitExceeded) {
		t.Errorf("expected ErrTraversalLimitExceeded, got %v", err)
	}
}

func TestTree_LookupParentError(t *testing.T) {
	dbErr := errors.New("database connection failed")
	lookupErr := func(id string) (*string, error) {
		return nil, dbErr
	}

	// DetectCycles wraps lookupParent error with failing ancestor ID
	err := core.DetectCycles("task-1", ptr("parent-1"), lookupErr)
	if !errors.Is(err, dbErr) || !strings.Contains(err.Error(), "parent-1") {
		t.Errorf("expected wrapped db error with parent-1 in DetectCycles, got %v", err)
	}

	// ValidateHierarchyDepth wraps lookupParent error with failing ancestor ID
	err = core.ValidateHierarchyDepth(0, "parent-1", lookupErr)
	if !errors.Is(err, dbErr) || !strings.Contains(err.Error(), "parent-1") {
		t.Errorf("expected wrapped db error with parent-1 in ValidateHierarchyDepth, got %v", err)
	}
}
