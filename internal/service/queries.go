package service

import (
	"context"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"github.com/newbpydev/tusk/internal/service/dateparse"
	"math/bits"
	"time"
)

var _ ports.TaskService = (*TaskService)(nil)

func defaultOrder() []core.SortOrder {
	return []core.SortOrder{{Field: core.SortByPriority, Direction: core.SortDesc}, {Field: core.SortByDueDate, Direction: core.SortAsc}, {Field: core.SortByCreatedAt, Direction: core.SortAsc}}
}
func (s *TaskService) GetTask(ctx context.Context, rawID string) (*core.Task, error) {
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	id, e := identifier(rawID)
	if e != nil {
		return nil, e
	}
	var out *core.Task
	e = s.repo.WithRead(ctx, func(ctx context.Context, r ports.TaskReader) error {
		v, e := r.GetByID(ctx, id)
		if e != nil {
			return e
		}
		copy := v.Clone()
		out = &copy
		return nil
	})
	if e != nil {
		return nil, e
	}
	return out, nil
}
func (s *TaskService) ListTasks(ctx context.Context, query ports.TaskQuery) ([]core.Task, error) {
	q, e := prepareQuery(ctx, query)
	if e != nil {
		return nil, e
	}
	var start, end time.Time
	if q.Due != nil {
		now, err := referenceTime(s.options.Clock)
		if err != nil {
			return nil, err
		}
		start, end, e = dateparse.DayBounds(*q.Due, now, s.options.Location)
		if e != nil {
			return nil, e
		}
	}
	var out []core.Task
	e = s.repo.WithRead(ctx, func(ctx context.Context, r ports.TaskReader) error {
		rows, e := r.List(ctx, q.Filter)
		if e != nil {
			return e
		}
		out = make([]core.Task, 0, len(rows))
		for _, v := range rows {
			if e := ctx.Err(); e != nil {
				return e
			}
			if q.Due != nil && (v.DueDate == nil || v.DueDate.Before(start) || !v.DueDate.Before(end)) {
				continue
			}
			out = append(out, v.Clone())
		}
		core.SortTasks(out, defaultOrder())
		return nil
	})
	if e != nil {
		return nil, e
	}
	return out, nil
}
func (s *TaskService) GetTaskTree(ctx context.Context, rawID string) ([]*core.TaskNode, error) {
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	id := rawID
	var e error
	if id != "" {
		id, e = identifier(id)
		if e != nil {
			return nil, e
		}
	}
	var out []*core.TaskNode
	e = s.repo.WithRead(ctx, func(ctx context.Context, r ports.TaskReader) error {
		var rows []core.Task
		var e error
		if id == "" {
			rows, e = r.List(ctx, core.TaskFilter{})
		} else {
			rows, e = r.GetSubtree(ctx, id)
			if e == nil {
				var ancestors []core.Task
				ancestors, e = r.GetAncestors(ctx, id)
				rows = append(rows, ancestors...)
			}
		}
		if e != nil {
			return e
		}
		forest, e := core.BuildTree(rows)
		if e != nil {
			return e
		}
		if id == "" {
			out = forest
		} else {
			var find func([]*core.TaskNode) *core.TaskNode
			find = func(nodes []*core.TaskNode) *core.TaskNode {
				for _, v := range nodes {
					if v.Task.ID == id {
						return v
					}
					if found := find(v.Children); found != nil {
						return found
					}
				}
				return nil
			}
			node := find(forest)
			if node == nil {
				return ports.ErrCorrupt
			}
			out = []*core.TaskNode{node}
		}
		return orderTree(ctx, out, 1)
	})
	if e != nil {
		return nil, e
	}
	return out, nil
}
func orderTree(ctx context.Context, nodes []*core.TaskNode, depth int) error {
	tasks := make([]core.Task, len(nodes))
	byID := make(map[string]*core.TaskNode, len(nodes))
	for i, n := range nodes {
		tasks[i] = n.Task
		byID[n.Task.ID] = n
	}
	core.SortTasks(tasks, defaultOrder())
	for i, v := range tasks {
		if e := ctx.Err(); e != nil {
			return e
		}
		n := byID[v.ID]
		nodes[i] = n
		n.Depth = depth
		if e := orderTree(ctx, n.Children, depth+1); e != nil {
			return e
		}
	}
	return nil
}
func (s *TaskService) GetStats(ctx context.Context) (ports.TaskStats, error) {
	if e := ctx.Err(); e != nil {
		return ports.TaskStats{}, e
	}
	now, e := referenceTime(s.options.Clock)
	if e != nil {
		return ports.TaskStats{}, e
	}
	var out ports.TaskStats
	e = s.repo.WithRead(ctx, func(ctx context.Context, r ports.TaskReader) error {
		rows, e := r.List(ctx, core.TaskFilter{})
		if e != nil {
			return e
		}
		out.ByStatus = map[core.Status]int{core.StatusTodo: 0, core.StatusInProgress: 0, core.StatusBlocked: 0, core.StatusDone: 0}
		out.Total = len(rows)
		for _, v := range rows {
			if e := ctx.Err(); e != nil {
				return e
			}
			out.ByStatus[v.Status]++
			if v.Status == core.StatusDone {
				out.Done++
				if v.CompletedAt != nil && v.CompletedAt.After(now.Add(-168*time.Hour)) && !v.CompletedAt.After(now) {
					out.CompletedLast7Days++
				}
			} else if v.DueDate != nil && v.DueDate.Before(now) {
				out.Overdue++
			}
		}
		if out.Total > 0 {
			hi, lo := bits.Mul64(uint64(out.Done), 100)
			q, _ := bits.Div64(hi, lo, uint64(out.Total))
			out.CompletionPercent = int(q)
		}
		return nil
	})
	if e != nil {
		return ports.TaskStats{}, e
	}
	return out, nil
}
