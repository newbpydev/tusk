package storage

import (
	"context"
	"strings"
	"testing"
)

func TestSchema_RejectsInvalidRows(t *testing.T) {
	db, inv := migrationFixture(t)
	if err := migrate(context.Background(), db, inv); err != nil {
		t.Fatal(err)
	}
	base := []any{"task", "title", "", "todo", 2, 0, nil, "[]", "2026-09-08T00:00:00.000000000Z", "2026-09-08T00:00:00.000000000Z", nil, nil}
	query := "INSERT INTO tasks(id,title,description,status,priority,progress,parent_id,tags,created_at,updated_at,due_date,completed_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)"
	for _, tc := range []struct {
		index int
		value any
	}{
		{0, nil}, {0, ""}, {1, ""}, {1, strings.Repeat("x", 256)}, {2, nil}, {3, "invalid"}, {4, 0}, {4, 5}, {5, -1}, {5, 101}, {6, "task"}, {6, "missing"}, {7, "{}"}, {7, "bad JSON"}, {8, nil}, {9, nil}, {11, "2026-09-08T00:00:00.000000000Z"}, {3, "done"},
	} {
		args := append([]any(nil), base...)
		args[tc.index] = tc.value
		if _, err := db.Exec(query, args...); err == nil {
			t.Fatalf("accepted invalid field %d=%v", tc.index, tc.value)
		}
	}
	base[5] = 100 // A valid open parent may have complete subtask progress.
	if _, err := db.Exec(query, base...); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(query, base...); err == nil {
		t.Fatal("duplicate accepted")
	}
	base[0], base[6] = "child", "task"
	if _, err := db.Exec(query, base...); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("INSERT INTO task_events(task_id,kind,changed_fields,occurred_at) VALUES('child','create','[\"title\"]','2026-09-08T00:00:00.000000000Z')"); err != nil {
		t.Fatal(err)
	}
	for _, stmt := range []string{
		"INSERT INTO task_events(task_id,kind,changed_fields,occurred_at) VALUES('missing','create','[\"title\"]','x')",
		"INSERT INTO task_events(task_id,kind,changed_fields,occurred_at) VALUES('task','delete','[\"title\"]','x')",
		"INSERT INTO task_events(task_id,kind,changed_fields,occurred_at) VALUES('task','create','[]','x')",
	} {
		if _, err := db.Exec(stmt); err == nil {
			t.Fatal("invalid event accepted")
		}
	}
	if _, err := db.Exec("DELETE FROM tasks WHERE id='task'"); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"tasks", "task_events"} {
		var count int
		if err := db.QueryRow("SELECT count(*) FROM " + table).Scan(&count); err != nil || count != 0 {
			t.Fatalf("cascade %s=%d err=%v", table, count, err)
		}
	}
}
