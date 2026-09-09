package service

import (
	"context"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"github.com/newbpydev/tusk/internal/service/dateparse"
	"time"
)

func (s *TaskService) CreateTask(ctx context.Context, command ports.CreateTaskCommand) (*core.Task, error) {
	cmd, err := prepareCreate(ctx, command)
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
	id, err := s.options.NewID(now)
	if err != nil || !validUUIDv7(id) {
		return nil, ports.ErrIdentityGeneration
	}
	var result *core.Task
	err = s.repo.WithWrite(ctx, func(ctx context.Context, w ports.TaskWriter) error {
		c := newChanges(ctx, w, now, s.options.AutoCompleteParent)
		if cmd.ParentID != nil {
			depth, e := c.depth(*cmd.ParentID)
			if e != nil {
				return e
			}
			if depth >= core.MaxHierarchyDepth {
				return core.ErrMaxDepthExceeded
			}
			if e = c.affectChain(*cmd.ParentID); e != nil {
				return e
			}
			if _, e = c.children(*cmd.ParentID); e != nil {
				return e
			}
		}
		if _, exists := c.current[id]; exists {
			return core.ErrDuplicateTaskID
		}
		v, e := core.NewTask(core.NewTaskParams{ID: id, Title: cmd.Title, Description: cmd.Description, Priority: cmd.Priority, Tags: cmd.Tags, ParentID: cmd.ParentID, DueDate: due, Now: now})
		if e != nil {
			return e
		}
		c.current[id] = v
		if cmd.ParentID != nil {
			c.members[*cmd.ParentID] = append(c.members[*cmd.ParentID], id)
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
func (s *TaskService) parseDue(expression *string, now time.Time) (*time.Time, error) {
	if expression == nil {
		return nil, nil
	}
	v, e := dateparse.ParseDue(*expression, now, s.options.Location)
	if e != nil {
		return nil, e
	}
	return &v, nil
}
