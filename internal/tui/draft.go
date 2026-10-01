package tui

import (
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

const (
	fieldTitle = iota
	fieldNotes
	fieldPriority
	fieldDue
	fieldTags
	fieldParent
	fieldStatus
	fieldProgress
	fieldCount
)

type mutationKind uint8

const (
	mutationNone mutationKind = iota
	mutationCreate
	mutationEdit
	mutationComplete
	mutationReopen
	mutationDelete
)

type mutationRequest struct {
	kind     mutationKind
	formID   uint64
	create   ports.CreateTaskCommand
	update   ports.UpdateTaskCommand
	task     ports.TaskCommand
	deleteID uint64
	deletion ports.DeleteTaskCommand
}
type DueParser func(string, time.Time, *time.Location) (time.Time, error)

func (r mutationRequest) baseTask() *core.Task {
	switch r.kind {
	case mutationEdit:
		return r.update.Base
	case mutationComplete, mutationReopen:
		return r.task.Base
	case mutationDelete:
		if r.deletion.Expected != nil {
			return &r.deletion.Expected.Target
		}
	}
	return nil
}

type taskDraft struct {
	fields, original [fieldCount]string
	base             *core.Task
	id               uint64
	progressLocked   bool
}

func newTaskDraft(base *core.Task, children bool, location *time.Location) *taskDraft {
	d := &taskDraft{}
	d.fields[fieldPriority] = "medium"
	if base != nil {
		detached := base.Clone()
		d.base = &detached
		d.fields[fieldTitle] = base.Title
		d.fields[fieldNotes] = base.Description
		d.fields[fieldPriority] = base.Priority.String()
		d.fields[fieldStatus] = string(base.Status)
		d.fields[fieldProgress] = strconv.Itoa(base.Progress)
		d.progressLocked = children || base.Status == core.StatusDone
		if base.DueDate != nil {
			d.fields[fieldDue] = base.DueDate.In(location).Format(time.RFC3339Nano)
		}
		if base.ParentID != nil {
			d.fields[fieldParent] = *base.ParentID
		}
		tags := make([]string, len(base.Tags))
		for i, tag := range base.Tags {
			tags[i] = string(tag)
		}
		d.fields[fieldTags] = strings.Join(tags, ", ")
	}
	d.original = d.fields
	return d
}

func (d *taskDraft) dirty() bool { return d.fields != d.original }

func (d *taskDraft) command(now time.Time, zone *time.Location, parse DueParser) (mutationRequest, int, string) {
	fail := func(field int, message string) (mutationRequest, int, string) {
		return mutationRequest{}, field, message
	}
	for i, value := range d.fields {
		if !utf8.ValidString(value) || strings.ContainsRune(value, 0) {
			return fail(i, "Text must be valid UTF-8 without NUL characters.")
		}
	}
	title := strings.TrimSpace(d.fields[fieldTitle])
	if title == "" {
		return fail(fieldTitle, "Enter a title.")
	}
	if utf8.RuneCountInString(title) > 255 {
		return fail(fieldTitle, "Title must be at most 255 characters.")
	}
	priority, err := core.ParsePriority(d.fields[fieldPriority])
	if err != nil {
		return fail(fieldPriority, "Choose a priority.")
	}
	var tags []core.Tag
	if strings.TrimSpace(d.fields[fieldTags]) != "" {
		tags, err = core.NormalizeTags(strings.Split(d.fields[fieldTags], ","))
		if err != nil {
			return fail(fieldTags, "Use comma-separated tags, such as work, design.")
		}
	}
	tagNames := make([]string, len(tags))
	for i, tag := range tags {
		tagNames[i] = string(tag)
	}
	parent := strings.TrimSpace(d.fields[fieldParent])
	if d.base != nil && parent == d.base.ID {
		return fail(fieldParent, "A task cannot be its own parent.")
	}
	due := strings.TrimSpace(d.fields[fieldDue])
	dueChanged := d.fields[fieldDue] != d.original[fieldDue]
	var parsed time.Time
	if due != "" && (d.base == nil || dueChanged) {
		if parse == nil {
			return fail(fieldDue, "Due-date parsing is unavailable.")
		}
		parsed, err = parse(due, now, zone)
		if err != nil {
			return fail(fieldDue, "Use today, tomorrow, YYYY-MM-DD or a timestamp.")
		}
	}
	request := mutationRequest{formID: d.id}
	if d.base == nil {
		request.kind = mutationCreate
		request.create = ports.CreateTaskCommand{Title: title, Description: d.fields[fieldNotes], Priority: priority, Tags: tagNames}
		if parent != "" {
			request.create.ParentID = &parent
		}
		if due != "" {
			request.create.Due = &due
		}
		return request, 0, ""
	}
	base := d.base.Clone()
	patch := ports.UpdateTaskCommand{ID: base.ID, Base: &base}
	if title != base.Title {
		patch.Title = &title
	}
	if d.fields[fieldNotes] != base.Description {
		value := d.fields[fieldNotes]
		patch.Description = &value
	}
	if priority != base.Priority {
		patch.Priority = &priority
	}
	if !slices.Equal(tags, base.Tags) {
		patch.Tags = &tagNames
	}
	if dueChanged {
		if due == "" {
			patch.ClearDue = base.DueDate != nil
		} else if base.DueDate == nil || !parsed.Equal(*base.DueDate) {
			patch.Due = &due
		}
	}
	if parent == "" {
		patch.ClearParent = base.ParentID != nil
	} else if base.ParentID == nil || parent != *base.ParentID {
		patch.ParentID = &parent
	}
	status, err := core.ParseStatus(d.fields[fieldStatus])
	if err != nil {
		return fail(fieldStatus, "Choose a status.")
	}
	if status != base.Status {
		patch.Status = &status
	}
	if d.fields[fieldProgress] != d.original[fieldProgress] {
		if d.progressLocked {
			return fail(fieldProgress, "Progress is derived for parents and completed tasks.")
		}
		progress, err := strconv.Atoi(d.fields[fieldProgress])
		if err != nil || progress < 0 || progress > 99 {
			return fail(fieldProgress, "Open leaf progress must be a number from 0 to 99.")
		}
		if progress != base.Progress {
			patch.Progress = &progress
		}
	}
	if patch.Progress != nil && (patch.Status != nil || patch.ParentID != nil || patch.ClearParent) {
		return fail(fieldProgress, "Save progress separately from status or parent changes.")
	}
	if patch.Title != nil || patch.Description != nil || patch.Priority != nil || patch.Status != nil || patch.Progress != nil || patch.Tags != nil || patch.Due != nil || patch.ClearDue || patch.ParentID != nil || patch.ClearParent {
		request.kind = mutationEdit
		request.update = patch
	}
	return request, 0, ""
}
