package service

import (
	"context"
	"github.com/newbpydev/tusk/internal/ports"
	"slices"
)

func (s *TaskService) GetTaskHistory(ctx context.Context, rawID string) ([]ports.TaskEvent, error) {
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	id, e := identifier(rawID)
	if e != nil {
		return nil, e
	}
	var out []ports.TaskEvent
	e = s.repo.WithRead(ctx, func(ctx context.Context, r ports.TaskReader) error {
		events, e := r.ListEvents(ctx, id)
		if e != nil {
			return e
		}
		out = make([]ports.TaskEvent, len(events))
		for i, v := range events {
			if e := ctx.Err(); e != nil {
				return e
			}
			v.ChangedFields = slices.Clone(v.ChangedFields)
			out[i] = v
		}
		return nil
	})
	if e != nil {
		return nil, e
	}
	return out, nil
}
