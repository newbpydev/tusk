package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/newbpydev/tusk/internal/core"
	generated "github.com/newbpydev/tusk/internal/storage/sqlc"
)

func TestQueries_CRUDAndCounts(t *testing.T) {
	db, inv := migrationFixture(t)
	ctx := context.Background()
	if err := migrate(ctx, db, inv); err != nil {
		t.Fatal(err)
	}
	q := generated.New(db)
	row := generated.CreateTaskParams{ID: "opaque '; DROP TABLE tasks;--", Title: "title %_ Ä", Description: "notes", Status: "todo", Priority: 2, Progress: 0, Tags: "[]", CreatedAt: "2026-09-08T00:00:00.000000000Z", UpdatedAt: "2026-09-08T00:00:00.000000000Z"}
	if err := q.CreateTask(ctx, row); err != nil {
		t.Fatal(err)
	}
	if err := q.CreateTask(ctx, row); err == nil {
		t.Fatal("duplicate accepted")
	}
	got, err := q.GetTask(ctx, row.ID)
	if err != nil || got.Title != row.Title || got.ParentID.Valid || got.DueDate.Valid {
		t.Fatalf("roundtrip=%+v err=%v", got, err)
	}
	count, err := q.UpdateTask(ctx, generated.UpdateTaskParams{ID: row.ID, Title: "updated", Description: row.Description, Status: row.Status, Priority: row.Priority, Progress: 0, Tags: "[]", UpdatedAt: row.UpdatedAt})
	if err != nil || count != 1 {
		t.Fatalf("update=%d %v", count, err)
	}
	seq, err := q.AppendEvent(ctx, generated.AppendEventParams{TaskID: row.ID, Kind: "metadata", ChangedFields: "[\"title\"]", OccurredAt: row.UpdatedAt})
	if err != nil || seq <= 0 {
		t.Fatalf("event=%d %v", seq, err)
	}
	events, err := q.ListEvents(ctx, row.ID)
	if err != nil || len(events) != 1 || events[0].Sequence != seq {
		t.Fatalf("events=%v err=%v", events, err)
	}
	count, err = q.DeleteTask(ctx, row.ID)
	if err != nil || count != 1 {
		t.Fatalf("delete=%d %v", count, err)
	}
	count, err = q.DeleteTask(ctx, row.ID)
	if err != nil || count != 0 {
		t.Fatalf("missing delete=%d %v", count, err)
	}
	if _, err := q.GetTask(ctx, row.ID); err != sql.ErrNoRows {
		t.Fatalf("missing row=%v", err)
	}
	events, err = q.ListEvents(ctx, row.ID)
	if err != nil || events == nil || len(events) != 0 {
		t.Fatalf("empty events=%v err=%v", events, err)
	}
}

func TestQueries_TreeSets(t *testing.T) {
	db, inv := migrationFixture(t)
	ctx := context.Background()
	if err := migrate(ctx, db, inv); err != nil {
		t.Fatal(err)
	}
	q := generated.New(db)
	for i := 0; i < 12; i++ {
		parent := sql.NullString{}
		if i > 0 && i < 11 {
			parent = sql.NullString{String: fmt.Sprintf("node-%d", i-1), Valid: true}
		}
		if err := q.CreateTask(ctx, generated.CreateTaskParams{ID: fmt.Sprintf("node-%d", i), Title: "node", Status: "todo", Priority: 1, Tags: "[]", ParentID: parent, CreatedAt: "x", UpdatedAt: "x"}); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		id                 string
		subtree, ancestors int
	}{{"node-0", 11, 0}, {"node-5", 6, 5}, {"node-10", 1, 10}, {"node-11", 1, 0}} {
		sub, err := q.GetSubtree(ctx, tc.id)
		if err != nil || len(sub) != tc.subtree {
			t.Fatalf("subtree %s=%d %v", tc.id, len(sub), err)
		}
		anc, err := q.GetAncestors(ctx, tc.id)
		if err != nil || len(anc) != tc.ancestors {
			t.Fatalf("ancestors %s=%d %v", tc.id, len(anc), err)
		}
	}
	kids, err := q.ListChildren(ctx, sql.NullString{String: "node-0", Valid: true})
	if err != nil || len(kids) != 1 || kids[0].ID != "node-1" {
		t.Fatalf("children=%v err=%v", kids, err)
	}
	if _, err := db.Exec("UPDATE tasks SET parent_id='node-10' WHERE id='node-0'"); err != nil {
		t.Fatal(err)
	}
	bounded, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	sub, err := q.GetSubtree(bounded, "node-0")
	if err != nil || len(sub) != 11 {
		t.Fatalf("cyclic CTE failed to terminate: %d %v", len(sub), err)
	}
	anc, err := q.GetAncestors(bounded, "node-0")
	if err != nil || len(anc) != 11 {
		t.Fatalf("cyclic ancestor CTE failed: %d %v", len(anc), err)
	}
}

func TestQueries_CandidateSuperset(t *testing.T) {
	db, inv := migrationFixture(t)
	ctx := context.Background()
	if err := migrate(ctx, db, inv); err != nil {
		t.Fatal(err)
	}
	q := generated.New(db)
	now := time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)
	corpus := []core.Task{}
	for i := 0; i < 16; i++ {
		status := core.StatusTodo
		if i%2 == 1 {
			status = core.StatusBlocked
		}
		task := core.Task{ID: fmt.Sprintf("id-%02d", i), Title: "Ä %_ '); DROP TABLE tasks;--", Status: status, Priority: core.Priority(i%4 + 1), CreatedAt: now, UpdatedAt: now}
		if i%3 == 0 {
			due := now.Add(time.Duration(i) * time.Hour)
			task.DueDate = &due
		}
		if i > 0 && i%2 == 0 {
			parent := "id-00"
			task.ParentID = &parent
		}
		p := generated.CreateTaskParams{ID: task.ID, Title: task.Title, Status: string(status), Priority: int64(task.Priority), Tags: "[]", CreatedAt: now.UTC().Format(dateLayout), UpdatedAt: now.UTC().Format(dateLayout)}
		if task.ParentID != nil {
			p.ParentID = sql.NullString{String: *task.ParentID, Valid: true}
		}
		if task.DueDate != nil {
			p.DueDate = sql.NullString{String: task.DueDate.UTC().Format(dateLayout), Valid: true}
		}
		if err := q.CreateTask(ctx, p); err != nil {
			t.Fatal(err)
		}
		corpus = append(corpus, task)
	}
	parent := "id-00"
	after := now.Add(3 * time.Hour)
	before := now.Add(12 * time.Hour)
	for _, statuses := range [][]core.Status{nil, {core.StatusTodo}, {"bad", core.StatusBlocked}, {"invalid"}} {
		for _, priorities := range [][]core.Priority{nil, {core.PriorityLow}, {0, core.PriorityUrgent}, {5}} {
			for _, filter := range []core.TaskFilter{{}, {RootOnly: true}, {ParentID: &parent}, {RootOnly: true, ParentID: &parent}, {DueAfter: &after}, {DueBefore: &before}, {SearchTerm: "Ä %_"}} {
				filter.Statuses = statuses
				filter.Priorities = priorities
				s, p := "[]", "[]"
				if len(statuses) > 0 {
					b, _ := json.Marshal(statuses)
					s = string(b)
				}
				if len(priorities) > 0 {
					b, _ := json.Marshal(priorities)
					p = string(b)
				}
				args := generated.ListCandidatesParams{Statuses: s, Priorities: p}
				if filter.RootOnly {
					args.RootOnly = 1
				}
				if filter.ParentID != nil {
					args.ParentID = sql.NullString{String: *filter.ParentID, Valid: true}
				}
				if filter.DueBefore != nil {
					args.DueBefore = sql.NullString{String: filter.DueBefore.UTC().Format(dateLayout), Valid: true}
				}
				if filter.DueAfter != nil {
					args.DueAfter = sql.NullString{String: filter.DueAfter.UTC().Format(dateLayout), Valid: true}
				}
				rows, err := q.ListCandidates(ctx, args)
				if err != nil {
					t.Fatal(err)
				}
				ids := map[string]bool{}
				for _, r := range rows {
					ids[r.ID] = true
				}
				for _, want := range core.FilterTasks(corpus, filter) {
					if !ids[want.ID] {
						t.Fatalf("candidate dropped %s for %+v", want.ID, filter)
					}
				}
			}
		}
	}
}

func TestQueries_NullableCandidateParameters(t *testing.T) {
	// These assignments require concrete nullable strings at compile time.
	args := generated.ListCandidatesParams{}
	var parent, before, after sql.NullString = args.ParentID, args.DueBefore, args.DueAfter
	if parent.Valid || before.Valid || after.Valid {
		t.Fatal("zero parameters must not constrain candidates")
	}
}
