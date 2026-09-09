package service

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

func ptr[T any](v T) *T { return &v }
func TestCommandValidation_PatchIntent(t *testing.T) {
	for _, c := range []ports.UpdateTaskCommand{
		{ID: "id", Due: ptr("today"), ClearDue: true},
		{ID: "id", ParentID: ptr("p"), ClearParent: true},
		{ID: "id", Progress: ptr(0), Status: ptr(core.StatusTodo)},
		{ID: "id", Progress: ptr(0), ClearParent: true},
		{ID: "id", Progress: ptr(0), ParentID: ptr("p")},
	} {
		if _, err := prepareUpdate(context.Background(), c); !errors.Is(err, ports.ErrInvalidCommand) {
			t.Fatalf("%+v: %v", c, err)
		}
	}
	if _, err := prepareUpdate(context.Background(), ports.UpdateTaskCommand{ID: "id"}); err != nil {
		t.Fatal(err)
	}
}
func TestCommandValidation_Metadata(t *testing.T) {
	for _, tc := range []struct {
		title string
		want  error
	}{{"", core.ErrEmptyTitle}, {"  ", core.ErrEmptyTitle}, {strings.Repeat("界", 255), nil}, {strings.Repeat("界", 256), core.ErrTitleTooLong}, {"a\x00", ports.ErrInvalidText}, {"\xff", ports.ErrInvalidText}} {
		_, err := prepareCreate(context.Background(), ports.CreateTaskCommand{Title: tc.title})
		if !errors.Is(err, tc.want) {
			t.Fatalf("title validation: %v want %v", err, tc.want)
		}
	}
	c, err := prepareCreate(context.Background(), ports.CreateTaskCommand{Title: " task ", Description: "exact\n\x1b[0m", Tags: []string{" #Work ", "work"}, ParentID: ptr(" opaque ")})
	if err != nil {
		t.Fatal(err)
	}
	if c.Title != "task" || c.Description != "exact\n\x1b[0m" || c.Priority != core.PriorityMedium || *c.ParentID != "opaque" || !reflect.DeepEqual(c.Tags, []string{"work"}) {
		t.Fatalf("normalization: %+v", c)
	}
	for _, c := range []ports.CreateTaskCommand{{Title: "t", Description: "\x00"}, {Title: "t", Priority: 99}, {Title: "t", Tags: []string{""}}, {Title: "t", ParentID: ptr("")}, {Title: "t", Due: ptr("")}} {
		if _, err := prepareCreate(context.Background(), c); err == nil {
			t.Fatalf("accepted %+v", c)
		}
	}
	for _, c := range []ports.UpdateTaskCommand{{ID: ""}, {ID: "\x00"}, {ID: "id", Title: ptr("")}, {ID: "id", Description: ptr("\xff")}, {ID: "id", Priority: ptr(core.Priority(0))}, {ID: "id", Status: ptr(core.Status("bad"))}, {ID: "id", Progress: ptr(-1)}, {ID: "id", Progress: ptr(101)}, {ID: "id", Tags: ptr([]string{""})}, {ID: "id", Due: ptr("")}, {ID: "id", ParentID: ptr("")}} {
		if _, err := prepareUpdate(context.Background(), c); err == nil {
			t.Fatalf("accepted %+v", c)
		}
	}
}
func TestCommandValidation_DetachesInput(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	base := core.Task{ID: "id", Title: "t", Status: core.StatusTodo, Priority: core.PriorityMedium, CreatedAt: now, UpdatedAt: now, Tags: []core.Tag{"work"}}
	c := ports.UpdateTaskCommand{ID: " id ", Title: ptr(" t "), Description: ptr("n"), Priority: ptr(core.PriorityHigh), Status: ptr(core.StatusBlocked), Tags: ptr([]string{" #Work "}), Due: ptr(" today "), ParentID: ptr(" p "), Base: &base}
	got, err := prepareUpdate(context.Background(), c)
	if err != nil {
		t.Fatal(err)
	}
	*got.Title = "changed"
	*got.Description = "changed"
	*got.Priority = core.PriorityLow
	*got.Status = core.StatusTodo
	(*got.Tags)[0] = "changed"
	*got.Due = "changed"
	*got.ParentID = "changed"
	got.Base.Tags[0] = "changed"
	if *c.Title != " t " || *c.Description != "n" || *c.Priority != core.PriorityHigh || *c.Status != core.StatusBlocked || (*c.Tags)[0] != " #Work " || *c.Due != " today " || *c.ParentID != " p " || base.Tags[0] != "work" {
		t.Fatal("caller input mutated")
	}
}
func TestBaseComparison(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	base := core.Task{ID: "id", Title: "title", Description: "notes", Status: core.StatusTodo, Priority: core.PriorityMedium, Progress: 4, CreatedAt: now, UpdatedAt: now, DueDate: &now, Tags: []core.Tag{"tag"}, ParentID: ptr("parent")}
	current := base.Clone()
	current.UpdatedAt = now.Add(time.Hour)
	current.Progress = 80
	current.CreatedAt = now.In(time.FixedZone("zone", 3600))
	if err := checkBase(&base, &current, false); err != nil {
		t.Fatal(err)
	}
	if err := checkBase(&base, &current, true); !errors.Is(err, ports.ErrConflict) {
		t.Fatal(err)
	}
	if err := checkBase(nil, &current, true); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*core.Task){func(v *core.Task) { v.ID = "other" }, func(v *core.Task) { v.Title = "other" }, func(v *core.Task) { v.Description = "other" }, func(v *core.Task) { v.Priority = core.PriorityHigh }, func(v *core.Task) { v.Status = core.StatusBlocked }, func(v *core.Task) { v.ParentID = nil }, func(v *core.Task) { v.Tags = nil }, func(v *core.Task) { v.DueDate = nil }, func(v *core.Task) { v.CreatedAt = now.Add(time.Second) }} {
		v := base.Clone()
		change(&v)
		if err := checkBase(&base, &v, false); !errors.Is(err, ports.ErrConflict) {
			t.Fatal(err)
		}
	}
	bad := base.Clone()
	bad.ID = "other"
	if _, err := prepareUpdate(context.Background(), ports.UpdateTaskCommand{ID: "id", Base: &bad}); !errors.Is(err, ports.ErrInvalidCommand) {
		t.Fatal(err)
	}
	bad = base.Clone()
	bad.Description = "\x00"
	if _, err := prepareUpdate(context.Background(), ports.UpdateTaskCommand{ID: "id", Base: &bad}); err == nil {
		t.Fatal("invalid snapshot")
	}
}
func TestQueryValidation(t *testing.T) {
	for _, q := range []ports.TaskQuery{{All: true, Filter: core.TaskFilter{Statuses: []core.Status{core.StatusDone}}}, {Filter: core.TaskFilter{RootOnly: true, ParentID: ptr("p")}}, {Filter: core.TaskFilter{Statuses: []core.Status{"bad"}}}, {Filter: core.TaskFilter{Priorities: []core.Priority{99}}}, {Filter: core.TaskFilter{Tags: []core.Tag{""}}}, {Filter: core.TaskFilter{SearchTerm: "\x00"}}, {Filter: core.TaskFilter{ParentID: ptr("")}}, {Due: ptr("")}} {
		if _, err := prepareQuery(context.Background(), q); err == nil {
			t.Fatalf("accepted %+v", q)
		}
	}
	q := ports.TaskQuery{Filter: core.TaskFilter{Tags: []core.Tag{"#Work"}, ParentID: ptr(" p ")}}
	got, err := prepareQuery(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Filter.Statuses) != 3 || got.Filter.Tags[0] != "work" || *got.Filter.ParentID != "p" {
		t.Fatalf("%+v", got)
	}
	got.Filter.Tags[0] = "changed"
	*got.Filter.ParentID = "changed"
	if q.Filter.Tags[0] != "#Work" || *q.Filter.ParentID != " p " {
		t.Fatal("aliased")
	}
}

func TestCommandValidation_MalformedSnapshots(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	base := core.Task{ID: "id", Title: "t", Status: core.StatusTodo, Priority: core.PriorityMedium, CreatedAt: now}
	for _, change := range []func(*core.Task){
		func(b *core.Task) { b.Title = " " }, func(b *core.Task) { b.Status = "bad" }, func(b *core.Task) { b.Priority = 0 }, func(b *core.Task) { b.Progress = -1 }, func(b *core.Task) { b.CreatedAt = time.Time{} }, func(b *core.Task) { b.ParentID = ptr("id") }, func(b *core.Task) { b.ParentID = ptr(" p ") }, func(b *core.Task) { b.DueDate = ptr(time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC)) }, func(b *core.Task) { b.Tags = []core.Tag{"\x00"} }, func(b *core.Task) { b.Tags = []core.Tag{"bad tag"} }, func(b *core.Task) { b.Tags = []core.Tag{"z", "a"} },
	} {
		b := base.Clone()
		change(&b)
		if err := validateBase("id", &b); err == nil {
			t.Fatalf("accepted malformed base: %+v", b)
		}
	}
	for _, v := range []string{"\xff", "\x00"} {
		if _, err := prepareCreate(context.Background(), ports.CreateTaskCommand{Title: "t", Tags: []string{v}}); !errors.Is(err, ports.ErrInvalidText) {
			t.Fatal(err)
		}
		if _, err := prepareCreate(context.Background(), ports.CreateTaskCommand{Title: "t", Due: &v}); !errors.Is(err, ports.ErrInvalidText) {
			t.Fatal(err)
		}
	}
	if _, err := prepareUpdate(context.Background(), ports.UpdateTaskCommand{ID: "id", Progress: ptr(99)}); err != nil {
		t.Fatal(err)
	}
	badTime := time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, q := range []ports.TaskQuery{{Filter: core.TaskFilter{Tags: []core.Tag{"\x00"}}}, {Filter: core.TaskFilter{DueBefore: &badTime}}, {Filter: core.TaskFilter{DueAfter: &badTime}}} {
		if _, err := prepareQuery(context.Background(), q); err == nil {
			t.Fatal("accepted invalid query")
		}
	}
	q := ports.TaskQuery{All: true, Filter: core.TaskFilter{Priorities: []core.Priority{core.PriorityHigh}, DueBefore: &now, DueAfter: &now}}
	got, err := prepareQuery(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	got.Filter.Priorities[0] = core.PriorityLow
	*got.Filter.DueBefore = now.Add(time.Hour)
	*got.Filter.DueAfter = now.Add(time.Hour)
	if q.Filter.Priorities[0] != core.PriorityHigh || !q.Filter.DueBefore.Equal(now) || !q.Filter.DueAfter.Equal(now) {
		t.Fatal("query alias")
	}
	if got, err := prepareQuery(context.Background(), ports.TaskQuery{Filter: core.TaskFilter{Statuses: []core.Status{core.StatusDone}}}); err != nil || len(got.Filter.Statuses) != 1 {
		t.Fatalf("explicit statuses: %+v %v", got, err)
	}
}
