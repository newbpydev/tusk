package service

import (
	"context"
	"errors"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"slices"
	"sort"
	"time"
)

type changes struct {
	ctx                                                    context.Context
	writer                                                 ports.TaskWriter
	now                                                    time.Time
	auto                                                   bool
	original                                               map[string]*core.Task
	current                                                map[string]*core.Task
	members                                                map[string][]string
	affected, explicitOpen, statusProgress, manualProgress map[string]bool
	lostChild                                              map[string]bool
}

func newChanges(ctx context.Context, w ports.TaskWriter, now time.Time, auto bool) *changes {
	return &changes{ctx: ctx, writer: w, now: now, auto: auto, original: map[string]*core.Task{}, current: map[string]*core.Task{}, members: map[string][]string{}, affected: map[string]bool{}, explicitOpen: map[string]bool{}, statusProgress: map[string]bool{}, manualProgress: map[string]bool{}, lostChild: map[string]bool{}}
}
func (c *changes) get(id string) (*core.Task, error) {
	if e := c.ctx.Err(); e != nil {
		return nil, e
	}
	if v, ok := c.current[id]; ok {
		return v, nil
	}
	v, e := c.writer.GetByID(c.ctx, id)
	if e != nil {
		return nil, e
	}
	if v == nil || v.ID != id || validateBase(id, v) != nil {
		return nil, ports.ErrCorrupt
	}
	original := v.Clone()
	current := v.Clone()
	c.original[id] = &original
	c.current[id] = &current
	return &current, nil
}
func (c *changes) children(id string) ([]core.Task, error) {
	if _, ok := c.members[id]; !ok {
		rows, e := c.writer.ListChildren(c.ctx, id)
		if e != nil {
			return nil, e
		}
		ids := make([]string, 0, len(rows))
		seen := map[string]bool{}
		for _, v := range rows {
			if e := c.ctx.Err(); e != nil {
				return nil, e
			}
			if seen[v.ID] || v.ParentID == nil || *v.ParentID != id || validateBase(v.ID, &v) != nil {
				return nil, ports.ErrCorrupt
			}
			seen[v.ID] = true
			ids = append(ids, v.ID)
			if _, ok := c.current[v.ID]; !ok {
				original := v.Clone()
				current := v.Clone()
				c.original[v.ID] = &original
				c.current[v.ID] = &current
			}
		}
		c.members[id] = ids
	}
	out := make([]core.Task, 0, len(c.members[id]))
	for _, child := range c.members[id] {
		v, e := c.get(child)
		if e != nil {
			return nil, e
		}
		out = append(out, v.Clone())
	}
	return out, nil
}
func (c *changes) depth(id string) (int, error) {
	seen := map[string]bool{}
	depth := 0
	for id != "" {
		if seen[id] {
			return 0, errors.Join(ports.ErrCorrupt, core.ErrCyclicDependency)
		}
		seen[id] = true
		v, e := c.get(id)
		if e != nil {
			return 0, e
		}
		depth++
		if depth > core.MaxHierarchyDepth {
			return 0, errors.Join(ports.ErrCorrupt, core.ErrMaxDepthExceeded)
		}
		id = ""
		if v.ParentID != nil {
			id = *v.ParentID
		}
	}
	return depth, nil
}
func (c *changes) affectChain(id string) error {
	seen := map[string]bool{}
	for id != "" {
		if seen[id] {
			return errors.Join(ports.ErrCorrupt, core.ErrCyclicDependency)
		}
		seen[id] = true
		v, e := c.get(id)
		if e != nil {
			return e
		}
		c.affected[id] = true
		id = ""
		if v.ParentID != nil {
			id = *v.ParentID
		}
	}
	return nil
}
func (c *changes) ordered(ids []string) ([]string, error) {
	depths := map[string]int{}
	for _, id := range ids {
		d, e := c.depth(id)
		if e != nil {
			return nil, e
		}
		depths[id] = d
	}
	sort.Slice(ids, func(i, j int) bool {
		if depths[ids[i]] != depths[ids[j]] {
			return depths[ids[i]] > depths[ids[j]]
		}
		return ids[i] < ids[j]
	})
	return ids, nil
}
func (c *changes) reparent(v *core.Task, parent *string) error {
	if equalPtr(v.ParentID, parent) {
		return nil
	}
	for _, p := range []*string{v.ParentID, parent} {
		if p != nil {
			if e := c.affectChain(*p); e != nil {
				return e
			}
			if _, e := c.children(*p); e != nil {
				return e
			}
		}
	}
	if v.ParentID != nil {
		p := *v.ParentID
		c.lostChild[p] = true
		c.members[p] = slices.DeleteFunc(c.members[p], func(id string) bool { return id == v.ID })
	}
	if parent != nil {
		c.members[*parent] = append(c.members[*parent], v.ID)
	}
	return v.SetParent(parent, c.now)
}
func (c *changes) flush() error {
	ids := make([]string, 0, len(c.current))
	events := map[string][]ports.TaskEvent{}
	for id, v := range c.current {
		e := taskEvents(c.original[id], v, c.now, c.statusProgress[id], c.manualProgress[id])
		if len(e) == 0 {
			if old := c.original[id]; old != nil {
				v.UpdatedAt = old.UpdatedAt
			}
			continue
		}
		events[id] = e
		ids = append(ids, id)
	}
	ids, err := c.ordered(ids)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if err := c.ctx.Err(); err != nil {
			return err
		}
		v := c.current[id]
		v.UpdatedAt = c.now
		var err error
		if c.original[id] == nil {
			err = c.writer.Create(c.ctx, v)
		} else {
			err = c.writer.Update(c.ctx, v)
		}
		if err != nil {
			return err
		}
		for _, e := range events[id] {
			if _, err = c.writer.AppendEvent(c.ctx, e); err != nil {
				return err
			}
		}
	}
	return c.ctx.Err()
}
