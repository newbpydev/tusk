package service

import (
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"slices"
	"time"
)

func taskEvents(before, after *core.Task, now time.Time, statusProgress, manualProgress bool) []ports.TaskEvent {
	out := []ports.TaskEvent{}
	appendEvent := func(kind ports.EventKind, fields []string) {
		if len(fields) == 0 {
			return
		}
		slices.Sort(fields)
		out = append(out, ports.TaskEvent{TaskID: after.ID, Kind: kind, ChangedFields: fields, OccurredAt: now})
	}
	if before == nil {
		appendEvent(ports.EventCreate, []string{"completed_at", "description", "due_date", "parent_id", "priority", "progress", "status", "tags", "title"})
		return out
	}
	metadata := []string{}
	if before.Title != after.Title {
		metadata = append(metadata, "title")
	}
	if before.Description != after.Description {
		metadata = append(metadata, "description")
	}
	if before.Priority != after.Priority {
		metadata = append(metadata, "priority")
	}
	if !slices.Equal(before.Tags, after.Tags) {
		metadata = append(metadata, "tags")
	}
	if !equalTime(before.DueDate, after.DueDate) {
		metadata = append(metadata, "due_date")
	}
	appendEvent(ports.EventMetadata, metadata)
	if !equalPtr(before.ParentID, after.ParentID) {
		appendEvent(ports.EventMove, []string{"parent_id"})
	}
	status := []string{}
	if before.Status != after.Status {
		status = append(status, "status")
	}
	if !equalTime(before.CompletedAt, after.CompletedAt) {
		status = append(status, "completed_at")
	}
	changedProgress := before.Progress != after.Progress
	if changedProgress && statusProgress {
		status = append(status, "progress")
	}
	appendEvent(ports.EventStatus, status)
	if changedProgress && !statusProgress {
		kind := ports.EventRollup
		if manualProgress {
			kind = ports.EventProgress
		}
		appendEvent(kind, []string{"progress"})
	}
	return out
}
