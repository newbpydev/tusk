package core

import (
	"fmt"
	"strings"
)

const MaxHierarchyDepth = 10

type TaskNode struct {
	Task     Task        `json:"task"`
	Children []*TaskNode `json:"children,omitempty"`
	Depth    int         `json:"depth"`
}

// DetectCycles traverses the ancestor chain of proposedParentID using lookupParent
// to verify that attaching taskID under proposedParentID does not create a cycle.
//
// Graph Finiteness Assumption:
// In storage-backed operation, the task graph is finite; traversal is guaranteed to
// terminate either by reaching a root (ancestorID == nil) or by revisiting an ancestor (ErrCyclicDependency).
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

	for {
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
}

// ValidateHierarchyDepth validates that attaching a task (with descendant depth taskSubtreeDepth)
// under proposedParentID will not violate MaxHierarchyDepth.
//
// Graph Finiteness Assumption:
// In storage-backed operation, the task graph is finite; traversal is guaranteed to
// terminate either by reaching a root (ancestorID == nil) or by revisiting an ancestor (ErrCyclicDependency).
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

	for {
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

	taskMap := make(map[string]int, len(tasks))
	nodes := make([]TaskNode, len(tasks))
	for i, t := range tasks {
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
		nodes[i] = TaskNode{
			Task:     t.Clone(),
			Children: []*TaskNode{},
			Depth:    1,
		}
		taskMap[id] = i
	}

	var roots []*TaskNode
	parents := make([]int, len(tasks))
	for i, t := range tasks {
		id := t.ID
		node := &nodes[i]
		parents[i] = -1
		if t.ParentID == nil {
			roots = append(roots, node)
		} else {
			parentID := strings.TrimSpace(*t.ParentID)
			parent, exists := taskMap[parentID]
			if !exists {
				return nil, fmt.Errorf("task %s references unknown parent %s: %w", id, parentID, ErrTaskNotFound)
			}
			parents[i] = parent
			nodes[parent].Children = append(nodes[parent].Children, node)
		}
	}

	// Cycle check: verify the entire graph is acyclic before evaluating component depth limits.
	// 0 = unvisited, 1 = visiting (in current ancestor stack), 2 = visited (known acyclic)
	cycleState := make([]uint8, len(tasks))
	for i := range tasks {
		if cycleState[i] == 2 {
			continue
		}
		curr := i
		for curr >= 0 {
			if cycleState[curr] == 1 {
				return nil, ErrCyclicDependency
			}
			if cycleState[curr] == 2 {
				break
			}
			cycleState[curr] = 1
			curr = parents[curr]
		}
		for curr = i; curr >= 0 && cycleState[curr] == 1; curr = parents[curr] {
			cycleState[curr] = 2
		}
	}

	visitedCount := 0
	var setDepth func(n *TaskNode, depth int) error

	setDepth = func(n *TaskNode, depth int) error {
		if depth > MaxHierarchyDepth {
			return ErrMaxDepthExceeded
		}
		// IDs are unique, each task has at most one parent, and the graph
		// is already acyclic: traversal cannot encounter a node twice.
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
