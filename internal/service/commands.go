package service

import (
	"context"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

func validText(v string) error {
	if !utf8.ValidString(v) || strings.ContainsRune(v, 0) {
		return ports.ErrInvalidText
	}
	return nil
}
func identifier(v string) (string, error) {
	if err := validText(v); err != nil {
		return "", err
	}
	v = strings.TrimSpace(v)
	if v == "" {
		return "", core.ErrInvalidTaskID
	}
	return v, nil
}
func title(v string) (string, error) {
	if err := validText(v); err != nil {
		return "", err
	}
	v = strings.TrimSpace(v)
	if v == "" {
		return "", core.ErrEmptyTitle
	}
	if utf8.RuneCountInString(v) > 255 {
		return "", core.ErrTitleTooLong
	}
	return v, nil
}
func copyPtr[T any](v *T) *T {
	if v == nil {
		return nil
	}
	x := *v
	return &x
}
func optionalText(v *string, id bool) (*string, error) {
	if v == nil {
		return nil, nil
	}
	s := strings.TrimSpace(*v)
	if err := validText(s); err != nil {
		return nil, err
	}
	if s == "" {
		if id {
			return nil, core.ErrInvalidTaskID
		}
		return nil, ports.ErrInvalidDate
	}
	return &s, nil
}
func tags(raw []string) ([]string, error) {
	for _, v := range raw {
		if err := validText(v); err != nil {
			return nil, err
		}
	}
	norm, err := core.NormalizeTags(raw)
	if err != nil {
		return nil, err
	}
	out := make([]string, len(norm))
	for i, v := range norm {
		out[i] = string(v)
	}
	return out, nil
}
func prepareCreate(ctx context.Context, c ports.CreateTaskCommand) (ports.CreateTaskCommand, error) {
	if err := ctx.Err(); err != nil {
		return ports.CreateTaskCommand{}, err
	}
	var err error
	if c.Title, err = title(c.Title); err != nil {
		return ports.CreateTaskCommand{}, err
	}
	if err = validText(c.Description); err != nil {
		return ports.CreateTaskCommand{}, err
	}
	if c.Priority == 0 {
		c.Priority = core.PriorityMedium
	}
	if !c.Priority.IsValid() {
		return ports.CreateTaskCommand{}, core.ErrInvalidPriority
	}
	if c.Tags, err = tags(c.Tags); err != nil {
		return ports.CreateTaskCommand{}, err
	}
	if c.ParentID, err = optionalText(c.ParentID, true); err != nil {
		return ports.CreateTaskCommand{}, err
	}
	if c.Due, err = optionalText(c.Due, false); err != nil {
		return ports.CreateTaskCommand{}, err
	}
	return c, nil
}
func prepareUpdate(ctx context.Context, c ports.UpdateTaskCommand) (ports.UpdateTaskCommand, error) {
	fail := func(err error) (ports.UpdateTaskCommand, error) { return ports.UpdateTaskCommand{}, err }
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	if c.Due != nil && c.ClearDue || c.ParentID != nil && c.ClearParent || c.Progress != nil && (c.Status != nil || c.ParentID != nil || c.ClearParent) {
		return fail(ports.ErrInvalidCommand)
	}
	var err error
	if c.ID, err = identifier(c.ID); err != nil {
		return fail(err)
	}
	if c.Title != nil {
		v, e := title(*c.Title)
		if e != nil {
			return fail(e)
		}
		c.Title = &v
	}
	if c.Description != nil {
		if err = validText(*c.Description); err != nil {
			return fail(err)
		}
		c.Description = copyPtr(c.Description)
	}
	if c.Priority != nil {
		if !c.Priority.IsValid() {
			return fail(core.ErrInvalidPriority)
		}
		c.Priority = copyPtr(c.Priority)
	}
	if c.Status != nil {
		if !c.Status.IsValid() {
			return fail(core.ErrInvalidStatus)
		}
		c.Status = copyPtr(c.Status)
	}
	if c.Progress != nil {
		if *c.Progress < 0 || *c.Progress > 100 {
			return fail(core.ErrInvalidProgress)
		}
		c.Progress = copyPtr(c.Progress)
	}
	if c.Tags != nil {
		v, e := tags(*c.Tags)
		if e != nil {
			return fail(e)
		}
		c.Tags = &v
	}
	if c.ParentID, err = optionalText(c.ParentID, true); err != nil {
		return fail(err)
	}
	if c.Due, err = optionalText(c.Due, false); err != nil {
		return fail(err)
	}
	if c.Base != nil {
		if err = validateBase(c.ID, c.Base); err != nil {
			return fail(err)
		}
		v := c.Base.Clone()
		c.Base = &v
	}
	return c, nil
}
func validateBase(id string, b *core.Task) error {
	if b.ID != id {
		return ports.ErrInvalidCommand
	}
	for _, s := range []string{b.ID, b.Title, b.Description} {
		if err := validText(s); err != nil {
			return err
		}
	}
	if v, err := title(b.Title); err != nil || v != b.Title {
		return ports.ErrInvalidCommand
	}
	if !b.Status.IsValid() || !b.Priority.IsValid() || b.Progress < 0 || b.Progress > 100 || b.CreatedAt.IsZero() || !validTime(b.CreatedAt) {
		return ports.ErrInvalidCommand
	}
	if b.ParentID != nil {
		v, err := identifier(*b.ParentID)
		if err != nil || v != *b.ParentID || v == id {
			return ports.ErrInvalidCommand
		}
	}
	for _, p := range []*time.Time{b.DueDate, b.CompletedAt} {
		if p != nil && !validTime(*p) {
			return ports.ErrInvalidCommand
		}
	}
	for _, tag := range b.Tags {
		if err := validText(string(tag)); err != nil {
			return err
		}
	}
	norm, err := core.NormalizeTagSlice(b.Tags)
	if err != nil || !slices.Equal(norm, b.Tags) {
		return ports.ErrInvalidCommand
	}
	return nil
}
func equalPtr[T comparable](a, b *T) bool {
	return a == nil && b == nil || a != nil && b != nil && *a == *b
}
func equalTime(a, b *time.Time) bool {
	return a == nil && b == nil || a != nil && b != nil && a.Equal(*b)
}
func checkBase(base, current *core.Task, leaf bool) error {
	if base == nil {
		return nil
	}
	if base.ID != current.ID || !base.CreatedAt.Equal(current.CreatedAt) || base.Title != current.Title || base.Description != current.Description || base.Priority != current.Priority || base.Status != current.Status || !equalPtr(base.ParentID, current.ParentID) || !slices.Equal(base.Tags, current.Tags) || !equalTime(base.DueDate, current.DueDate) || leaf && base.Progress != current.Progress {
		return ports.ErrConflict
	}
	return nil
}
func prepareQuery(ctx context.Context, q ports.TaskQuery) (ports.TaskQuery, error) {
	fail := func(err error) (ports.TaskQuery, error) { return ports.TaskQuery{}, err }
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	f := q.Filter
	if q.All && len(f.Statuses) > 0 || f.RootOnly && f.ParentID != nil {
		return fail(ports.ErrInvalidCommand)
	}
	if err := validText(f.SearchTerm); err != nil {
		return fail(err)
	}
	f.Statuses = slices.Clone(f.Statuses)
	f.Priorities = slices.Clone(f.Priorities)
	for _, s := range f.Statuses {
		if !s.IsValid() {
			return fail(core.ErrInvalidStatus)
		}
	}
	for _, p := range f.Priorities {
		if !p.IsValid() {
			return fail(core.ErrInvalidPriority)
		}
	}
	if len(f.Statuses) == 0 && !q.All {
		f.Statuses = []core.Status{core.StatusTodo, core.StatusInProgress, core.StatusBlocked}
	}
	for _, tag := range f.Tags {
		if err := validText(string(tag)); err != nil {
			return fail(err)
		}
	}
	var err error
	if f.Tags, err = core.NormalizeTagSlice(f.Tags); err != nil {
		return fail(err)
	}
	if f.ParentID, err = optionalText(f.ParentID, true); err != nil {
		return fail(err)
	}
	f.DueBefore = copyPtr(f.DueBefore)
	f.DueAfter = copyPtr(f.DueAfter)
	for _, p := range []*time.Time{f.DueBefore, f.DueAfter} {
		if p != nil && !validTime(*p) {
			return fail(ports.ErrInvalidDate)
		}
	}
	if q.Due, err = optionalText(q.Due, false); err != nil {
		return fail(err)
	}
	q.Filter = f
	return q, nil
}
