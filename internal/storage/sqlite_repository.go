package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"slices"

	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	generated "github.com/newbpydev/tusk/internal/storage/sqlc"
)

var _ ports.TaskRepository = (*Repository)(nil)

func repositoryRead[T any](r *Repository, ctx context.Context, fn func(context.Context, ports.TaskReader) (T, error)) (value T, err error) {
	err = r.WithRead(ctx, func(c context.Context, reader ports.TaskReader) error {
		var e error
		value, e = fn(c, reader)
		return e
	})
	if err != nil {
		var zero T
		return zero, err
	}
	return value, nil
}
func (r *Repository) GetByID(ctx context.Context, id string) (*core.Task, error) {
	return repositoryRead(r, ctx, func(c context.Context, h ports.TaskReader) (*core.Task, error) { return h.GetByID(c, id) })
}
func (r *Repository) List(ctx context.Context, f core.TaskFilter) ([]core.Task, error) {
	return repositoryRead(r, ctx, func(c context.Context, h ports.TaskReader) ([]core.Task, error) { return h.List(c, f) })
}
func (r *Repository) ListChildren(ctx context.Context, id string) ([]core.Task, error) {
	return repositoryRead(r, ctx, func(c context.Context, h ports.TaskReader) ([]core.Task, error) { return h.ListChildren(c, id) })
}
func (r *Repository) GetSubtree(ctx context.Context, id string) ([]core.Task, error) {
	return repositoryRead(r, ctx, func(c context.Context, h ports.TaskReader) ([]core.Task, error) { return h.GetSubtree(c, id) })
}
func (r *Repository) GetAncestors(ctx context.Context, id string) ([]core.Task, error) {
	return repositoryRead(r, ctx, func(c context.Context, h ports.TaskReader) ([]core.Task, error) { return h.GetAncestors(c, id) })
}
func (r *Repository) ListEvents(ctx context.Context, id string) ([]ports.TaskEvent, error) {
	return repositoryRead(r, ctx, func(c context.Context, h ports.TaskReader) ([]ports.TaskEvent, error) { return h.ListEvents(c, id) })
}

func (h *taskHandle) get(ctx context.Context, id string) (*core.Task, error) {
	if !validID(id) {
		return nil, core.ErrInvalidTaskID
	}
	row, err := h.q.GetTask(ctx, id)
	if err != nil {
		return nil, storageCause(err)
	}
	return decodeTask(row)
}
func (h *taskHandle) GetByID(ctx context.Context, id string) (*core.Task, error) {
	return handleCall(h, ctx, false, func(c context.Context) (*core.Task, error) { return h.get(c, id) })
}

func decodeTasks(rows []generated.Task) ([]core.Task, error) {
	tasks := make([]core.Task, 0, len(rows))
	for _, row := range rows {
		task, err := decodeTask(row)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *task)
	}
	return tasks, nil
}
func sortTasks(tasks []core.Task) {
	core.SortTasks(tasks, []core.SortOrder{{Field: core.SortByPriority, Direction: core.SortDesc}, {Field: core.SortByDueDate, Direction: core.SortAsc}, {Field: core.SortByCreatedAt, Direction: core.SortAsc}})
}

func (h *taskHandle) List(ctx context.Context, f core.TaskFilter) ([]core.Task, error) {
	return handleCall(h, ctx, false, func(c context.Context) ([]core.Task, error) {
		p := generated.ListCandidatesParams{Statuses: "[]", Priorities: "[]"}
		if len(f.Statuses) > 0 {
			b, _ := json.Marshal(f.Statuses)
			p.Statuses = string(b)
		}
		if len(f.Priorities) > 0 {
			b, _ := json.Marshal(f.Priorities)
			p.Priorities = string(b)
		}
		if f.RootOnly {
			p.RootOnly = 1
		}
		if f.ParentID != nil {
			p.ParentID = *f.ParentID
		}
		if f.DueBefore != nil && validTime(*f.DueBefore) {
			p.DueBefore = f.DueBefore.UTC().Format(dateLayout)
		}
		if f.DueAfter != nil && validTime(*f.DueAfter) {
			p.DueAfter = f.DueAfter.UTC().Format(dateLayout)
		}
		rows, err := h.q.ListCandidates(c, p)
		if err != nil {
			return nil, storageCause(err)
		}
		tasks, err := decodeTasks(rows)
		if err != nil {
			return nil, err
		}
		tasks = core.FilterTasks(tasks, f)
		sortTasks(tasks)
		return tasks, nil
	})
}

// tree materializes the requested rows and full ancestor chain, validates them
// as a complete forest, then projects without rewriting stored ParentID values.
func (h *taskHandle) tree(ctx context.Context, id string, subtree bool) (*core.TaskNode, map[string]*core.TaskNode, error) {
	target, err := h.get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	ancestors, err := h.q.GetAncestors(ctx, id)
	if err != nil {
		return nil, nil, storageCause(err)
	}
	rows := ancestors
	if subtree {
		descendants, err := h.q.GetSubtree(ctx, id)
		if err != nil {
			return nil, nil, storageCause(err)
		}
		rows = append(rows, descendants...)
	}
	tasks, err := decodeTasks(rows)
	if err != nil {
		return nil, nil, err
	}
	all := []core.Task{*target}
	seen := map[string]bool{id: true}
	for _, task := range tasks {
		if !seen[task.ID] {
			all = append(all, task)
			seen[task.ID] = true
		}
	}
	sortTasks(all)
	forest, err := core.BuildTree(all)
	if err != nil {
		return nil, nil, corruptCause(err)
	}
	nodes := make(map[string]*core.TaskNode, len(all))
	queue := forest
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		nodes[node.Task.ID] = node
		queue = append(queue, node.Children...)
	}
	return nodes[id], nodes, nil
}

func flatten(node *core.TaskNode) []core.Task {
	result := []core.Task{node.Task.Clone()}
	for _, child := range node.Children {
		result = append(result, flatten(child)...)
	}
	return result
}
func (h *taskHandle) GetSubtree(ctx context.Context, id string) ([]core.Task, error) {
	return handleCall(h, ctx, false, func(c context.Context) ([]core.Task, error) {
		node, _, err := h.tree(c, id, true)
		if err != nil {
			return nil, err
		}
		return flatten(node), nil
	})
}
func (h *taskHandle) GetAncestors(ctx context.Context, id string) ([]core.Task, error) {
	return handleCall(h, ctx, false, func(c context.Context) ([]core.Task, error) {
		node, nodes, err := h.tree(c, id, false)
		if err != nil {
			return nil, err
		}
		result := []core.Task{}
		for node.Task.ParentID != nil {
			node = nodes[*node.Task.ParentID]
			result = append(result, node.Task.Clone())
		}
		return result, nil
	})
}
func (h *taskHandle) ListChildren(ctx context.Context, id string) ([]core.Task, error) {
	return handleCall(h, ctx, false, func(c context.Context) ([]core.Task, error) {
		node, _, err := h.tree(c, id, false)
		if err != nil {
			return nil, err
		}
		rows, err := h.q.ListChildren(c, sql.NullString{String: id, Valid: true})
		if err != nil {
			return nil, storageCause(err)
		}
		result, err := decodeTasks(rows)
		if err != nil {
			return nil, err
		}
		if len(result) > 0 && node.Depth >= core.MaxHierarchyDepth {
			return nil, corruptCause(core.ErrMaxDepthExceeded)
		}
		sortTasks(result)
		return result, nil
	})
}

func (h *taskHandle) ListEvents(ctx context.Context, id string) ([]ports.TaskEvent, error) {
	return handleCall(h, ctx, false, func(c context.Context) ([]ports.TaskEvent, error) {
		if _, err := h.get(c, id); err != nil {
			return nil, err
		}
		rows, err := h.q.ListEvents(c, id)
		if err != nil {
			return nil, storageCause(err)
		}
		result := make([]ports.TaskEvent, 0, len(rows))
		for _, row := range rows {
			event, err := decodeEvent(row)
			if err != nil {
				return nil, err
			}
			result = append(result, event)
		}
		return result, nil
	})
}

func (h *taskHandle) Create(ctx context.Context, task *core.Task) error {
	_, err := handleCall(h, ctx, true, func(c context.Context) (struct{}, error) {
		p, err := encodeTask(task)
		if err != nil {
			return struct{}{}, err
		}
		return struct{}{}, storageCause(h.q.CreateTask(c, p))
	})
	return err
}
func (h *taskHandle) Update(ctx context.Context, task *core.Task) error {
	_, err := handleCall(h, ctx, true, func(c context.Context) (struct{}, error) {
		p, err := encodeTask(task)
		if err != nil {
			return struct{}{}, err
		}
		stored, err := h.get(c, task.ID)
		if err != nil {
			return struct{}{}, err
		}
		if !stored.CreatedAt.Equal(task.CreatedAt) {
			return struct{}{}, ports.ErrInvalidRecord
		}
		_, err = h.q.UpdateTask(c, generated.UpdateTaskParams{ID: p.ID, Title: p.Title, Description: p.Description, Status: p.Status, Priority: p.Priority, Progress: p.Progress, ParentID: p.ParentID, Tags: p.Tags, UpdatedAt: p.UpdatedAt, DueDate: p.DueDate, CompletedAt: p.CompletedAt})
		return struct{}{}, storageCause(err)
	})
	return err
}
func (h *taskHandle) Delete(ctx context.Context, id string, recursive bool) ([]string, error) {
	return handleCall(h, ctx, true, func(c context.Context) ([]string, error) {
		node, _, err := h.tree(c, id, true)
		if err != nil {
			return nil, err
		}
		if len(node.Children) > 0 && !recursive {
			return nil, ports.ErrChildrenPresent
		}
		tasks := flatten(node)
		ids := make([]string, 0, len(tasks))
		for _, task := range tasks {
			ids = append(ids, task.ID)
		}
		slices.Sort(ids)
		if _, err := h.q.DeleteTask(c, id); err != nil {
			return nil, storageCause(err)
		}
		return ids, nil
	})
}
func (h *taskHandle) AppendEvent(ctx context.Context, event ports.TaskEvent) (int64, error) {
	return handleCall(h, ctx, true, func(c context.Context) (int64, error) {
		if event.Sequence != 0 {
			return 0, ports.ErrInvalidRecord
		}
		if err := validateEvent(event); err != nil {
			return 0, err
		}
		fields, _ := json.Marshal(event.ChangedFields)
		sequence, err := h.q.AppendEvent(c, generated.AppendEventParams{TaskID: event.TaskID, Kind: string(event.Kind), ChangedFields: string(fields), OccurredAt: event.OccurredAt.UTC().Format(dateLayout)})
		return sequence, storageCause(err)
	})
}
