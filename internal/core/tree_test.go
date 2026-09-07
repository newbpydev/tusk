package core_test

import (
	"errors"
	"fmt"
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
