package service

import (
	"github.com/newbpydev/tusk/internal/core"
)

func (c *changes) recompute() error {
	ids := make([]string, 0, len(c.affected))
	for id := range c.affected {
		ids = append(ids, id)
	}
	ids, err := c.ordered(ids)
	if err != nil {
		return err
	}
	for _, id := range ids {
		v, e := c.get(id)
		if e != nil {
			return e
		}
		children, e := c.children(id)
		if e != nil {
			return e
		}
		if len(children) == 0 {
			// Only a changed child membership resets former parent progress. A leaf
			// merely visited for explicit status intent retains its manual progress.
			if !c.lostChild[id] {
				continue
			}
			if v.Status == core.StatusDone {
				e = v.ResetLeaf(100, children, c.now)
			} else {
				e = v.ResetLeaf(0, children, c.now)
			}
			if e != nil {
				return e
			}
			continue
		}
		allDone := true
		for _, child := range children {
			if child.Status != core.StatusDone {
				allDone = false
			}
		}
		if v.Status == core.StatusDone && !allDone {
			if e = v.TransitionTo(core.StatusInProgress, c.now); e != nil {
				return e
			}
		}
		if c.auto && allDone && !c.explicitOpen[id] {
			if e = v.TransitionTo(core.StatusDone, c.now); e != nil {
				return e
			}
		}
		if e = v.SetRollupProgress(core.CalculateProgress(*v, children), children, c.now); e != nil {
			return e
		}
	}
	return nil
}
