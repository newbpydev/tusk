package service

import (
	"context"
	"github.com/newbpydev/tusk/internal/ports"
	"slices"
)

func (s *TaskService) PreviewDeleteTask(ctx context.Context, rawID string) (ports.DeletePreview, error) {
	if e := ctx.Err(); e != nil {
		return ports.DeletePreview{}, e
	}
	id, e := identifier(rawID)
	if e != nil {
		return ports.DeletePreview{}, e
	}
	var result ports.DeletePreview
	e = s.repo.WithRead(ctx, func(ctx context.Context, r ports.TaskReader) error {
		v, e := r.GetByID(ctx, id)
		if e != nil {
			return e
		}
		rows, e := r.GetSubtree(ctx, id)
		if e != nil {
			return e
		}
		ids := make([]string, 0, len(rows))
		for _, v := range rows {
			if e := ctx.Err(); e != nil {
				return e
			}
			ids = append(ids, v.ID)
		}
		slices.Sort(ids)
		result = ports.DeletePreview{Target: v.Clone(), IDs: ids}
		return validatePreview(id, &result)
	})
	if e != nil {
		return ports.DeletePreview{}, e
	}
	return result, nil
}
func validatePreview(id string, p *ports.DeletePreview) error {
	if e := validateBase(id, &p.Target); e != nil {
		return e
	}
	if len(p.IDs) == 0 || !slices.IsSorted(p.IDs) || !slices.Contains(p.IDs, id) {
		return ports.ErrInvalidCommand
	}
	for i, v := range p.IDs {
		canonical, e := identifier(v)
		if e != nil || canonical != v || i > 0 && p.IDs[i-1] == v {
			return ports.ErrInvalidCommand
		}
	}
	return nil
}
func (s *TaskService) DeleteTask(ctx context.Context, command ports.DeleteTaskCommand) (ports.DeleteResult, error) {
	fail := func(e error) (ports.DeleteResult, error) { return ports.DeleteResult{}, e }
	if e := ctx.Err(); e != nil {
		return fail(e)
	}
	cmd := command
	if cmd.Force && cmd.Expected != nil {
		return fail(ports.ErrInvalidCommand)
	}
	id, e := identifier(cmd.ID)
	if e != nil {
		return fail(e)
	}
	if !cmd.Force && cmd.Expected == nil {
		return fail(ports.ErrConfirmationRequired)
	}
	if cmd.Expected != nil {
		if e = validatePreview(id, cmd.Expected); e != nil {
			return fail(e)
		}
		p := ports.DeletePreview{Target: cmd.Expected.Target.Clone(), IDs: slices.Clone(cmd.Expected.IDs)}
		cmd.Expected = &p
	}
	now, e := referenceTime(s.options.Clock)
	if e != nil {
		return fail(e)
	}
	var result ports.DeleteResult
	e = s.repo.WithWrite(ctx, func(ctx context.Context, w ports.TaskWriter) error {
		c := newChanges(ctx, w, now, s.options.AutoCompleteParent)
		v, e := c.get(id)
		if e != nil {
			return e
		}
		ids, e := c.subtree(id)
		if e != nil {
			return e
		}
		slices.Sort(ids)
		if cmd.Expected != nil {
			if e = checkBase(&cmd.Expected.Target, v, len(ids) == 1); e != nil {
				return e
			}
			if !slices.Equal(ids, cmd.Expected.IDs) {
				return ports.ErrConflict
			}
		}
		if len(ids) > 1 && !cmd.Recursive {
			return ports.ErrChildrenPresent
		}
		if v.ParentID != nil {
			p := *v.ParentID
			if e = c.affectChain(p); e != nil {
				return e
			}
			if _, e = c.children(p); e != nil {
				return e
			}
			c.members[p] = slices.DeleteFunc(c.members[p], func(child string) bool { return child == id })
			c.lostChild[p] = true
		}
		deleted, e := w.Delete(ctx, id, cmd.Recursive)
		if e != nil {
			return e
		}
		if !slices.Equal(deleted, ids) {
			return ports.ErrCorrupt
		}
		for _, id := range ids {
			delete(c.current, id)
			delete(c.original, id)
			delete(c.members, id)
			delete(c.affected, id)
		}
		if e = c.recompute(); e != nil {
			return e
		}
		if e = c.flush(); e != nil {
			return e
		}
		result = ports.DeleteResult{ID: id, DeletedIDs: slices.Clone(ids), DeletedCount: len(ids), Deleted: true}
		return nil
	})
	if e != nil {
		return fail(e)
	}
	return result, nil
}
