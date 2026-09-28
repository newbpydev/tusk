package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

func decodeJSON(t *testing.T, data []byte) any {
	t.Helper()
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	var v any
	if err := d.Decode(&v); err != nil {
		t.Fatal(err)
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		t.Fatalf("extra JSON: %v", err)
	}
	if len(data) == 0 || data[len(data)-1] != '\n' || bytes.Count(data, []byte{'\n'}) != 1 {
		t.Fatalf("not compact LF: %q", data)
	}
	return v
}
func jsonFixture() core.Task {
	return core.Task{ID: "opaque-complete-identifier", Title: "界 👩‍💻", Status: core.StatusTodo, Priority: core.PriorityMedium, CreatedAt: time.Date(2026, 1, 2, 3, 4, 5, 123, time.FixedZone("offset", 3600)), UpdatedAt: time.Date(2026, 1, 2, 3, 4, 5, 123, time.UTC)}
}

func TestJSON_ExplicitFields(t *testing.T) {
	task := jsonFixture()
	data, err := encodeJSON(&task)
	if err != nil {
		t.Fatal(err)
	}
	m := decodeJSON(t, data).(map[string]any)
	keys := []string{"id", "title", "description", "status", "priority", "parent_id", "progress", "tags", "due_date", "created_at", "updated_at", "completed_at"}
	if len(m) != len(keys) {
		t.Fatalf("keys %v", m)
	}
	for _, k := range keys {
		if _, ok := m[k]; !ok {
			t.Fatalf("missing %s", k)
		}
	}
	if m["description"] != "" || m["parent_id"] != nil || m["due_date"] != nil || m["completed_at"] != nil || len(m["tags"].([]any)) != 0 || m["priority"] != json.Number("2") || m["created_at"] != "2026-01-02T02:04:05.000000123Z" {
		t.Fatalf("fields %v", m)
	}
	task.Description = "quotes \" \\ <script>\n\x1b\x00 " + strings.Repeat("long", 10000)
	task.Tags = []core.Tag{"alpha"}
	parent := "full-parent"
	task.ParentID = &parent
	task.DueDate = &task.CreatedAt
	task.CompletedAt = &task.CreatedAt
	data, err = encodeJSON(&task)
	if err != nil {
		t.Fatal(err)
	}
	m = decodeJSON(t, data).(map[string]any)
	if m["description"] != task.Description || m["due_date"] != "2026-01-02T02:04:05.000000123Z" || bytes.Contains(data, []byte{27}) {
		t.Fatalf("bad roundtrip %v", err)
	}
}

func TestJSON_Collections(t *testing.T) {
	for _, v := range []any{[]core.Task(nil), []*core.TaskNode(nil), []ports.TaskEvent(nil)} {
		data, err := encodeJSON(v)
		if err != nil || string(data) != "[]\n" {
			t.Fatalf("%T: %q %v", v, data, err)
		}
	}
	task := jsonFixture()
	data, err := encodeJSON([]core.Task{task, task})
	if err != nil {
		t.Fatal(err)
	}
	if len(decodeJSON(t, data).([]any)) != 2 {
		t.Fatal("list count")
	}
	leaf := &core.TaskNode{Task: task, Depth: 10}
	root := leaf
	for d := 9; d >= 1; d-- {
		root = &core.TaskNode{Task: task, Children: []*core.TaskNode{root}, Depth: d}
	}
	data, err = encodeJSON([]*core.TaskNode{root})
	if err != nil {
		t.Fatal(err)
	}
	node := decodeJSON(t, data).([]any)[0].(map[string]any)
	for d := 1; d <= 10; d++ {
		n, _ := node["depth"].(json.Number).Int64()
		if n != int64(d) {
			t.Fatal(n)
		}
		kids := node["children"].([]any)
		if d == 10 {
			if len(kids) != 0 {
				t.Fatal(kids)
			}
		} else {
			node = kids[0].(map[string]any)
		}
	}
}

func TestJSON_StatsDeleteHistory(t *testing.T) {
	for _, total := range []int{0, 10} {
		data, err := encodeJSON(ports.TaskStats{Total: total, ByStatus: map[core.Status]int{core.StatusDone: 3}})
		if err != nil {
			t.Fatal(err)
		}
		m := decodeJSON(t, data).(map[string]any)
		if len(m) != 6 || len(m["by_status"].(map[string]any)) != 4 {
			t.Fatal(m)
		}
	}
	ids := []string{"z", "a"}
	data, err := encodeJSON(ports.DeleteResult{ID: "root", DeletedIDs: ids, DeletedCount: 2, Deleted: true})
	if err != nil {
		t.Fatal(err)
	}
	m := decodeJSON(t, data).(map[string]any)
	if len(m) != 4 || !reflect.DeepEqual(m["deleted_ids"], []any{"a", "z"}) || ids[0] != "z" {
		t.Fatal(m)
	}
	data, err = encodeJSON(ports.DeleteResult{})
	if err != nil {
		t.Fatal(err)
	}
	if len(decodeJSON(t, data).(map[string]any)["deleted_ids"].([]any)) != 0 {
		t.Fatal("null IDs")
	}
	fields := []string{"title", "description"}
	data, err = encodeJSON([]ports.TaskEvent{{Sequence: 9007199254740993, TaskID: "opaque", Kind: ports.EventMetadata, ChangedFields: fields, OccurredAt: jsonFixture().CreatedAt}})
	if err != nil {
		t.Fatal(err)
	}
	e := decodeJSON(t, data).([]any)[0].(map[string]any)
	if len(e) != 5 || e["sequence"] != json.Number("9007199254740993") || !reflect.DeepEqual(e["changed_fields"], []any{"description", "title"}) || fields[0] != "title" {
		t.Fatal(e)
	}
}

func TestJSON_Invalid(t *testing.T) {
	bad := jsonFixture()
	bad.CreatedAt = time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, v := range []any{&bad, (*core.Task)(nil), 42, []*core.TaskNode{nil}} {
		data, err := encodeJSON(v)
		if err == nil || len(data) != 0 {
			t.Fatalf("%T: %q %v", v, data, err)
		}
	}
}
