package service

import (
	"context"
	"errors"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

func (s *TaskService) CompleteTask(ctx context.Context, cmd ports.TaskCommand) (*core.Task, error) {
	status := core.StatusDone
	return s.UpdateTask(ctx, ports.UpdateTaskCommand{ID: cmd.ID, Base: cmd.Base, Status: &status})
}
func (s *TaskService) ReopenTask(ctx context.Context, cmd ports.ReopenTaskCommand) (*core.Task, error) {
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	if cmd.Status != core.StatusTodo && cmd.Status != core.StatusInProgress {
		return nil, core.ErrInvalidStatusTransition
	}
	return s.UpdateTask(ctx, ports.UpdateTaskCommand{ID: cmd.ID, Base: cmd.Base, Status: &cmd.Status})
}
func (s *TaskService) UpdateTask(ctx context.Context, command ports.UpdateTaskCommand) (*core.Task, error) {
	cmd, err := prepareUpdate(ctx, command)
	if err != nil {
		return nil, err
	}
	now, err := referenceTime(s.options.Clock)
	if err != nil {
		return nil, err
	}
	due, err := s.parseDue(cmd.Due, now)
	if err != nil {
		return nil, err
	}
	var result *core.Task
	err = s.repo.WithWrite(ctx, func(ctx context.Context, w ports.TaskWriter) error {
		c := newChanges(ctx, w, now, s.options.AutoCompleteParent)
		v, e := c.get(cmd.ID)
		if e != nil {
			return e
		}
		if e = c.patch(v, cmd, due); e != nil {
			return e
		}
		if cmd.ParentID != nil || cmd.ClearParent {
			parent := cmd.ParentID
			if !equalPtr(v.ParentID, parent) {
				if e = c.validateMove(v, parent); e != nil {
					return e
				}
				if e = c.reparent(v, parent); e != nil {
					return e
				}
			}
		}
		if cmd.Status != nil {
			if e = c.changeStatus(v, *cmd.Status); e != nil {
				return e
			}
		}
		if e = c.recompute(); e != nil {
			return e
		}
		if e = c.flush(); e != nil {
			return e
		}
		copy := v.Clone()
		result = &copy
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
func (c *changes) subtree(id string) ([]string, error) {
	rows, e := c.writer.GetSubtree(c.ctx, id)
	if e != nil {
		return nil, e
	}
	ids := make([]string, 0, len(rows))
	seen := map[string]bool{}
	for _, v := range rows {
		if e := c.ctx.Err(); e != nil {
			return nil, e
		}
		if seen[v.ID] || validateBase(v.ID, &v) != nil {
			return nil, ports.ErrCorrupt
		}
		seen[v.ID] = true
		ids = append(ids, v.ID)
		if _, ok := c.current[v.ID]; !ok {
			o := v.Clone()
			n := v.Clone()
			c.original[v.ID] = &o
			c.current[v.ID] = &n
		}
	}
	if !seen[id] {
		return nil, ports.ErrCorrupt
	}
	return c.ordered(ids)
}
func (c *changes) validateMove(v *core.Task, parent *string) error {
	ids, e := c.subtree(v.ID)
	if e != nil {
		return e
	}
	depth, e := c.depth(v.ID)
	if e != nil {
		return e
	}
	height := 0
	for _, id := range ids {
		d, e := c.depth(id)
		if e != nil {
			return e
		}
		if d-depth > height {
			height = d - depth
		}
	}
	var lookupError error
	lookup := func(id string) (*string, error) {
		v, e := c.get(id)
		if e != nil {
			lookupError = e
			return nil, e
		}
		return copyPtr(v.ParentID), nil
	}
	if e = core.DetectCycles(v.ID, parent, lookup); e != nil {
		if lookupError != nil {
			return lookupError
		}
		if errors.Is(e, core.ErrSelfParenting) {
			return core.ErrSelfParenting
		}
		return core.ErrCyclicDependency
	}
	if parent == nil {
		if height+1 > core.MaxHierarchyDepth {
			return core.ErrMaxDepthExceeded
		}
		return nil
	}
	if e = core.ValidateHierarchyDepth(height, *parent, lookup); e != nil {
		if lookupError != nil {
			return lookupError
		}
		if errors.Is(e, core.ErrCyclicDependency) {
			return core.ErrCyclicDependency
		}
		return core.ErrMaxDepthExceeded
	}
	return nil
}
func (c *changes) changeStatus(v *core.Task, status core.Status) error {
	if status == core.StatusDone {
		ids, e := c.subtree(v.ID)
		if e != nil {
			return e
		}
		changed := false
		for _, id := range ids {
			n, e := c.get(id)
			if e != nil {
				return e
			}
			if n.Status == core.StatusDone {
				continue
			}
			if e = n.TransitionTo(core.StatusDone, c.now); e != nil {
				return e
			}
			c.statusProgress[id] = true
			changed = true
		}
		if changed && v.ParentID != nil {
			return c.affectChain(*v.ParentID)
		}
		return nil
	}
	c.explicitOpen[v.ID] = true
	if v.Status == status {
		return nil
	}
	wasDone := v.Status == core.StatusDone
	if e := v.TransitionTo(status, c.now); e != nil {
		return e
	}
	if wasDone {
		children, e := c.children(v.ID)
		if e != nil {
			return e
		}
		c.statusProgress[v.ID] = len(children) == 0
	}
	return c.affectChain(v.ID)
}
