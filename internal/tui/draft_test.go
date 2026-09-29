package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/service/dateparse"
)

func draftCommand(d *taskDraft) (mutationRequest, int, string) {
	return d.command(testOptions().Now(), time.UTC, dateparse.ParseDue)
}

func TestForm_DefaultsAndDetachedDirtyPatch(t *testing.T) {
	d := newTaskDraft(nil, false, time.UTC)
	d.fields[fieldTitle] = "  A useful task  "
	req, _, err := draftCommand(d)
	if err != "" || req.kind != mutationCreate || req.create.Title != "A useful task" || req.create.Priority != core.PriorityMedium || req.create.ParentID != nil || req.create.Description != "" {
		t.Fatalf("wrong defaults: %+v %s", req, err)
	}
	original := fixtureNode("one", "Original", core.PriorityHigh, nil).Task
	original.Description = "raw\ttext\x1b[31m"
	original.Tags = []core.Tag{"work"}
	due := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	original.DueDate = &due
	d = newTaskDraft(&original, false, time.UTC)
	d.fields[fieldTitle] = "Changed"
	req, _, err = draftCommand(d)
	if err != "" || req.update.Title == nil || req.update.Description != nil || req.update.Due != nil || req.update.ClearDue || req.update.Tags != nil {
		t.Fatal("unchanged raw fields were included")
	}
	wantDue := due
	original.Tags[0] = "mutated"
	*original.DueDate = due.Add(time.Hour)
	if req.update.Base.Description != "raw\ttext\x1b[31m" || req.update.Base.Tags[0] != "work" || !req.update.Base.DueDate.Equal(wantDue) {
		t.Fatal("Base aliases caller fields")
	}
	d.fields[fieldTitle] = "Other"
	if *req.update.Title != "Changed" {
		t.Fatal("queued patch aliases draft")
	}
}

func TestForm_NoChangesAndExplicitClears(t *testing.T) {
	parent := "parent"
	due := testOptions().Now()
	base := fixtureNode("one", "Original", 2, &due).Task
	base.ParentID = &parent
	base.Tags = []core.Tag{"design", "work"}
	d := newTaskDraft(&base, false, time.UTC)
	d.fields[fieldTitle] = " Original "
	d.fields[fieldTags] = "WORK, design"
	req, _, err := draftCommand(d)
	if err != "" || req.kind != mutationNone {
		t.Fatal("equal edit issued a mutation")
	}
	d.fields[fieldDue] = ""
	d.fields[fieldTags] = ""
	d.fields[fieldParent] = ""
	req, _, err = draftCommand(d)
	if err != "" || !req.update.ClearDue || !req.update.ClearParent || req.update.Tags == nil || len(*req.update.Tags) != 0 {
		t.Fatal("explicit clear lost intent")
	}
}

func TestForm_ValidationFocusAndProgressContracts(t *testing.T) {
	for _, tc := range []struct {
		field int
		value string
	}{{fieldTitle, ""}, {fieldTitle, strings.Repeat("界", 256)}, {fieldTitle, "nul\x00"}, {fieldTags, "bad tag"}, {fieldDue, "nonsense"}, {fieldPriority, "invalid"}, {fieldParent, "one"}, {fieldStatus, "invalid"}, {fieldProgress, "-1"}, {fieldProgress, "100"}, {fieldProgress, "text"}} {
		base := fixtureNode("one", "Original", 2, nil).Task
		d := newTaskDraft(&base, false, time.UTC)
		d.fields[tc.field] = tc.value
		_, field, err := draftCommand(d)
		if err == "" || field != tc.field {
			t.Errorf("field %d: got %d %q", tc.field, field, err)
		}
	}
	base := fixtureNode("one", "Original", 2, nil).Task
	for _, value := range []string{"0", "99"} {
		d := newTaskDraft(&base, false, time.UTC)
		d.fields[fieldProgress] = value
		if _, _, err := draftCommand(d); err != "" {
			t.Fatal(err)
		}
	}
	d := newTaskDraft(&base, false, time.UTC)
	d.fields[fieldProgress] = "40"
	d.fields[fieldStatus] = "in-progress"
	if _, field, err := draftCommand(d); err == "" || field != fieldProgress {
		t.Fatal("progress and status combined")
	}
	d = newTaskDraft(&base, true, time.UTC)
	d.fields[fieldProgress] = "40"
	if _, _, err := draftCommand(d); err == "" {
		t.Fatal("parent progress accepted")
	}
	base.Status = core.StatusDone
	base.Progress = 100
	d = newTaskDraft(&base, false, time.UTC)
	if !d.progressLocked {
		t.Fatal("done progress editable")
	}
}
