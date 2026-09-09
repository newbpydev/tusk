package service

import (
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"time"
)

// patch applies metadata/manual progress to the latest row. Status and parent
// dispatch belongs to the public mutation orchestrator.
func (c *changes) patch(v *core.Task, cmd ports.UpdateTaskCommand, due *time.Time) error {
	if cmd.Base != nil || cmd.Progress != nil {
		children, e := c.children(v.ID)
		if e != nil {
			return e
		}
		if e = checkBase(cmd.Base, v, len(children) == 0); e != nil {
			return e
		}
		if cmd.Progress != nil && len(children) != 0 {
			return core.ErrInvalidProgress
		}
	}
	title, description, priority, tags, date := v.Title, v.Description, v.Priority, v.Tags, v.DueDate
	if cmd.Title != nil {
		title = *cmd.Title
	}
	if cmd.Description != nil {
		description = *cmd.Description
	}
	if cmd.Priority != nil {
		priority = *cmd.Priority
	}
	if cmd.Tags != nil {
		tags = make([]core.Tag, len(*cmd.Tags))
		for i, t := range *cmd.Tags {
			tags[i] = core.Tag(t)
		}
	}
	if cmd.Due != nil {
		date = due
	} else if cmd.ClearDue {
		date = nil
	}
	if e := v.Update(title, description, priority, tags, date, c.now); e != nil {
		return e
	}
	if cmd.Progress != nil {
		old := v.Progress
		if e := v.SetProgress(*cmd.Progress, c.now); e != nil {
			return e
		}
		c.manualProgress[v.ID] = true
		if old != v.Progress && v.ParentID != nil {
			if e := c.affectChain(*v.ParentID); e != nil {
				return e
			}
		}
	}
	return nil
}
