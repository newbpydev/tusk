package core

import (
	"fmt"
	"strings"
)

const (
	MaxHierarchyDepth = 10
	// maxTraversalSteps bounds cycle and depth traversal.
	// Valid hierarchies are bounded by MaxHierarchyDepth (10 lookups).
	// A 10,000-step bound ensures long cycles (e.g. 1000+ nodes) in corrupt or imported graphs
	// are fully detected and classified as ErrCyclicDependency before falling back to ErrTraversalLimitExceeded.
	maxTraversalSteps = 10000
)

type TaskNode struct {
	Task     Task        `json:"task"`
	Children []*TaskNode `json:"children,omitempty"`
	Depth    int         `json:"depth"`
}

// DetectCycles traverses the ancestor chain of proposedParentID using lookupParent
// to verify that attaching taskID under proposedParentID does not create a cycle.
//
// Rules:
// 1. If proposedParentID == nil, returns nil immediately (root promotion is always cycle-free).
// 2. If taskID == *proposedParentID, returns ErrSelfParenting.
// 3. Upward traversal from *proposedParentID: if an ancestor matches taskID, returns ErrCyclicDependency.
func DetectCycles(taskID string, proposedParentID *string, lookupParent func(id string) (*string, error)) error {
	if proposedParentID == nil {
		return nil
	}

	parentID := *proposedParentID
	if taskID == parentID {
		return ErrSelfParenting
	}
	currentID := parentID
	var visitedBuf [16]string
	visitedSlice := visitedBuf[:0]
	visitedSlice = append(visitedSlice, currentID)
	var visitedMap map[string]struct{}

	for step := 0; step < maxTraversalSteps; step++ {
		ancestorID, err := lookupParent(currentID)
		if err != nil {
			return fmt.Errorf("parent lookup failed for %s: %w", currentID, err)
		}
		if ancestorID == nil {
			return nil
		}

		anc := *ancestorID
		if anc == taskID {
			return ErrCyclicDependency
		}

		if visitedMap != nil {
			if _, loop := visitedMap[anc]; loop {
				return ErrCyclicDependency
			}
			visitedMap[anc] = struct{}{}
		} else {
			for _, v := range visitedSlice {
				if v == anc {
					return ErrCyclicDependency
				}
			}
			if len(visitedSlice) < cap(visitedBuf) {
				visitedSlice = append(visitedSlice, anc)
			} else {
				visitedMap = make(map[string]struct{}, 32)
				for _, v := range visitedSlice {
					visitedMap[v] = struct{}{}
				}
				visitedMap[anc] = struct{}{}
			}
		}

		currentID = anc
	}

	return ErrTraversalLimitExceeded
}

// ValidateHierarchyDepth validates that attaching a task (with descendant depth taskSubtreeDepth)
// under proposedParentID will not violate MaxHierarchyDepth.
//
// Depth of root = 1.
// parentDepth = depth of proposedParentID from its root.
// Invariant: parentDepth + 1 + taskSubtreeDepth <= MaxHierarchyDepth.
func ValidateHierarchyDepth(taskSubtreeDepth int, proposedParentID string, lookupParent func(id string) (*string, error)) error {
	if taskSubtreeDepth < 0 {
		return ErrInvalidDepth
	}
	parentDepth := 1
	currentID := proposedParentID

	visited := make(map[string]struct{})
	visited[currentID] = struct{}{}

	for step := 0; step < maxTraversalSteps; step++ {
		ancestorID, err := lookupParent(currentID)
		if err != nil {
			return fmt.Errorf("parent lookup failed for %s: %w", currentID, err)
		}
		if ancestorID == nil {
			break
		}
		if _, loop := visited[*ancestorID]; loop {
			return ErrCyclicDependency
		}
		parentDepth++
		currentID = *ancestorID
		visited[currentID] = struct{}{}

		if step == maxTraversalSteps-1 {
			return ErrTraversalLimitExceeded
		}
	}

	if parentDepth > MaxHierarchyDepth || taskSubtreeDepth > MaxHierarchyDepth || taskSubtreeDepth > MaxHierarchyDepth-parentDepth-1 {
		return ErrMaxDepthExceeded
	}

	return nil
}

// BuildTree converts a flat slice of Task entities into a hierarchical forest of TaskNode pointers.
//
// Rules:
// 1. Every non-root task must reference a ParentID present in tasks; missing parent returns ErrTaskNotFound.
// 2. Roots (ParentID == nil) are placed at Depth 1.
// 3. Children inherit parent.Depth + 1.
// 4. If any task is part of an unrooted circular loop, returns ErrCyclicDependency.
func BuildTree(tasks []Task) ([]*TaskNode, error) {
	if len(tasks) == 0 {
		return []*TaskNode{}, nil
	}

	taskMap := make(map[string]*TaskNode, len(tasks))
	for _, t := range tasks {
		id := strings.TrimSpace(t.ID)
		if id == "" || t.ID != id {
			return nil, ErrInvalidTaskID
		}
		if t.ParentID != nil {
			parentID := strings.TrimSpace(*t.ParentID)
			if parentID == "" || *t.ParentID != parentID {
				return nil, ErrInvalidTaskID
			}
			if parentID == id {
				return nil, fmt.Errorf("task %s references itself as parent: %w", id, ErrSelfParenting)
			}
		}
		if _, exists := taskMap[id]; exists {
			return nil, fmt.Errorf("task with id %s already exists: %w", id, ErrDuplicateTaskID)
		}
		taskMap[id] = &TaskNode{
			Task:     t.Clone(),
			Children: []*TaskNode{},
			Depth:    1,
		}
	}

	var roots []*TaskNode
	for _, t := range tasks {
		id := strings.TrimSpace(t.ID)
		node := taskMap[id]
		if t.ParentID == nil {
			roots = append(roots, node)
		} else {
			parentID := strings.TrimSpace(*t.ParentID)
			parent, exists := taskMap[parentID]
			if !exists {
				return nil, fmt.Errorf("task %s references unknown parent %s: %w", id, parentID, ErrTaskNotFound)
			}
			parent.Children = append(parent.Children, node)
		}
	}

	// Cycle check: verify the entire graph is acyclic before evaluating component depth limits.
	// 0 = unvisited, 1 = visiting (in current ancestor stack), 2 = visited (known acyclic)
	cycleState := make(map[string]int, len(tasks))
	for _, t := range tasks {
		id := strings.TrimSpace(t.ID)
		if cycleState[id] == 2 {
			continue
		}
		curr := id
		for curr != "" {
			st := cycleState[curr]
			if st == 1 {
				return nil, ErrCyclicDependency
			}
			if st == 2 {
				break
			}
			cycleState[curr] = 1
			node := taskMap[curr]
			if node.Task.ParentID == nil {
				break
			}
			curr = strings.TrimSpace(*node.Task.ParentID)
		}
		// Mark all nodes traversed in this chain as visited (2)
		curr = id
		for curr != "" && cycleState[curr] == 1 {
			cycleState[curr] = 2
			node := taskMap[curr]
			if node.Task.ParentID == nil {
				break
			}
			curr = strings.TrimSpace(*node.Task.ParentID)
		}
	}

	visitedCount := 0
	var setDepth func(n *TaskNode, depth int) error
	visitedNodes := make(map[string]struct{}, len(tasks))

	setDepth = func(n *TaskNode, depth int) error {
		if depth > MaxHierarchyDepth {
			return ErrMaxDepthExceeded
		}
		// Defense-in-depth: cycleState check above already guarantees graph is acyclic.
		if _, seen := visitedNodes[n.Task.ID]; seen {
			return ErrCyclicDependency
		}
		visitedNodes[n.Task.ID] = struct{}{}
		visitedCount++
		n.Depth = depth

		for _, child := range n.Children {
			if err := setDepth(child, depth+1); err != nil {
				return err
			}
		}
		return nil
	}

	for _, root := range roots {
		if err := setDepth(root, 1); err != nil {
			return nil, err
		}
	}

	// Defense-in-depth: cycleState and parent existence guarantee all tasks are reachable from roots.
	if visitedCount != len(tasks) {
		return nil, ErrCyclicDependency
	}

	return roots, nil
}
