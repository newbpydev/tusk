package tui

import (
	"reflect"
	"testing"
	"time"

	"github.com/newbpydev/tusk/internal/core"
)

func fixtureNode(id, title string, priority core.Priority, due *time.Time, children ...*core.TaskNode) *core.TaskNode {
	return &core.TaskNode{Task: core.Task{ID: id, Title: title, Status: core.StatusTodo, Priority: priority, DueDate: due, CreatedAt: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)}, Children: children}
}

func TestProjection_GroupsAndCanonicalOrder(t *testing.T) {
	now := time.Date(2026, 9, 29, 12, 0, 0, 0, time.UTC)
	earlier := now.Add(-48 * time.Hour)
	later := now.Add(48 * time.Hour)
	child := fixtureNode("child", "Done child", core.PriorityLow, nil)
	child.Task.Status = core.StatusDone
	done := fixtureNode("done", "Done root", core.PriorityUrgent, &earlier)
	done.Task.Status = core.StatusDone
	forest := []*core.TaskNode{done, fixtureNode("backlog", "Backlog", core.PriorityHigh, nil), fixtureNode("upcoming", "Later", core.PriorityHigh, &later), fixtureNode("today-low", "Today low", core.PriorityLow, &earlier, child), fixtureNode("today-high", "Today high", core.PriorityHigh, &earlier)}
	before := snapshot(forest)
	rows := project(forest, core.TaskFilter{}, nil, nil, nil, now, time.UTC)
	var got []string
	for _, row := range rows {
		got = append(got, row.group+":"+row.node.Task.ID)
	}
	want := []string{"Today:today-high", "Today:today-low", "Today:child", "Upcoming:upcoming", "Backlog:backlog", "Completed:done"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%v want %v", got, want)
	}
	if snapshot(forest) != before {
		t.Fatal("projection mutated accepted forest")
	}
	if len(project(nil, core.TaskFilter{}, nil, nil, nil, now, time.UTC)) != 0 {
		t.Fatal("empty forest produced rows")
	}
}

func TestSearch_ContextCollapseAndPredicates(t *testing.T) {
	child := fixtureNode("child", "界 NEEDLE", core.PriorityHigh, nil)
	child.Task.Tags = []core.Tag{"work", "design"}
	other := fixtureNode("other", "Unrelated", core.PriorityLow, nil)
	root := fixtureNode("root", "Parent", core.PriorityMedium, nil, child, other)
	forest := []*core.TaskNode{root}
	collapsed := map[string]bool{"root": true}
	if got := project(forest, core.TaskFilter{}, nil, nil, collapsed, time.Now(), time.UTC); len(got) != 1 {
		t.Fatal("collapse ignored")
	}
	f := core.TaskFilter{SearchTerm: " needle ", Statuses: []core.Status{core.StatusTodo, core.StatusDone}, Priorities: []core.Priority{core.PriorityHigh, core.PriorityUrgent}, Tags: []core.Tag{"work", "design"}}
	rows := project(forest, f, nil, nil, collapsed, time.Now(), time.UTC)
	if len(rows) != 2 || !rows[0].context || rows[1].context || rows[1].depth != 1 || rows[1].node.Task.ID != "child" {
		t.Fatalf("bad context projection: %+v", rows)
	}
	f.SearchTerm = "parent"
	f.Statuses = nil
	f.Priorities = nil
	f.Tags = nil
	if rows = project(forest, f, nil, nil, collapsed, time.Now(), time.UTC); len(rows) != 1 || rows[0].context {
		t.Fatal("matching parent included unrelated children")
	}
	f.SearchTerm = "absent"
	if rows = project(forest, f, nil, nil, collapsed, time.Now(), time.UTC); len(rows) != 0 {
		t.Fatal("missing query has rows")
	}
	if !collapsed["root"] {
		t.Fatal("search changed saved collapse state")
	}
}

func TestProjection_DueDayDST(t *testing.T) {
	zone, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, 3, 8, 0, 0, 0, 0, zone)
	end := start.AddDate(0, 0, 1)
	if end.Sub(start) != 23*time.Hour {
		t.Fatal("fixture isn't a DST boundary")
	}
	last := end.Add(-time.Nanosecond)
	forest := []*core.TaskNode{fixtureNode("at-start", "Start", 1, &start), fixtureNode("last", "Last", 1, &last), fixtureNode("end", "End", 1, &end), fixtureNode("missing", "Missing", 1, nil)}
	rows := project(forest, core.TaskFilter{}, &start, &end, nil, start, zone)
	if len(rows) != 2 || rows[0].node.Task.ID != "at-start" || rows[1].node.Task.ID != "last" {
		t.Fatalf("half-open day failed: %+v", rows)
	}
	rows = project(forest, core.TaskFilter{}, nil, nil, nil, start, zone)
	if rows[2].group != "Upcoming" || rows[3].group != "Backlog" {
		t.Fatal("local day grouping failed")
	}
	rows = project(forest, core.TaskFilter{}, nil, nil, nil, end, zone)
	if rows[2].group != "Today" {
		t.Fatal("group clock did not advance")
	}
}
