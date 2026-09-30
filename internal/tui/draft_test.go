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

func TestForm_EditDiffPatchesNotesAndPriority(t *testing.T) {
	for _, tc := range []struct {
		name                    string
		notes, priority         string
		wantNotes, wantPriority bool
	}{
		{"notes only", "fresh notes", "medium", true, false},
		{"priority only", "stored notes", "high", false, true},
		{"notes and priority", "rewritten notes", "urgent", true, true},
		{"unchanged edit", "stored notes", "medium", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			base := fixtureNode("one", "Original", core.PriorityMedium, nil).Task
			base.Description = "stored notes"
			d := newTaskDraft(&base, false, time.UTC)
			d.fields[fieldNotes] = tc.notes
			d.fields[fieldPriority] = tc.priority
			req, field, message := draftCommand(d)
			if message != "" || field != 0 {
				t.Fatalf("unexpected failure: field %d %q", field, message)
			}
			patch := req.update
			if tc.wantNotes {
				if req.kind != mutationEdit || patch.Description == nil || *patch.Description != tc.notes {
					t.Fatalf("notes edit dropped: kind %d %+v", req.kind, patch.Description)
				}
			} else if patch.Description != nil {
				t.Fatalf("unchanged notes patched: %q", *patch.Description)
			}
			priority, err := core.ParsePriority(tc.priority)
			if err != nil {
				t.Fatal(err)
			}
			if tc.wantPriority {
				if req.kind != mutationEdit || patch.Priority == nil || *patch.Priority != priority {
					t.Fatalf("priority edit dropped: kind %d %+v", req.kind, patch.Priority)
				}
			} else if patch.Priority != nil {
				t.Fatalf("unchanged priority patched: %v", *patch.Priority)
			}
			if patch.Title != nil || patch.Tags != nil || patch.Due != nil || patch.ClearDue || patch.ParentID != nil || patch.ClearParent || patch.Status != nil || patch.Progress != nil {
				t.Fatalf("spurious patch entries: %+v", patch)
			}
			if !tc.wantNotes && !tc.wantPriority {
				if req.kind != mutationNone {
					t.Fatalf("unchanged edit issued a mutation: %+v", req)
				}
				return
			}
			if req.kind != mutationEdit || patch.ID != base.ID || patch.Base == nil || patch.Base.Description != "stored notes" || patch.Base.Priority != core.PriorityMedium {
				t.Fatalf("patch targets the wrong base: %+v", patch)
			}
		})
	}
}

func TestForm_DueWithoutParserIsRejected(t *testing.T) {
	d := newTaskDraft(nil, false, time.UTC)
	d.fields[fieldTitle] = "Task"
	d.fields[fieldDue] = "tomorrow"
	_, field, message := d.command(testOptions().Now(), time.UTC, nil)
	if field != fieldDue || message != "Due-date parsing is unavailable." {
		t.Fatalf("nil parser accepted a due date: field %d %q", field, message)
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
