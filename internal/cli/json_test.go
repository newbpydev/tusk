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
	for _, v := range []any{&bad, (*core.Task)(nil), 42, []*core.TaskNode{nil}, []*core.TaskNode{{Task: jsonFixture(), Children: []*core.TaskNode{nil}}}} {
		data, err := encodeJSON(v)
		if err == nil || len(data) != 0 {
			t.Fatalf("%T: %q %v", v, data, err)
		}
	}
}

func TestJSON_EncodingAllocation(t *testing.T) {
	tasks := make([]core.Task, 100)
	for i := range tasks {
		tasks[i] = jsonFixture()
		tasks[i].Title = strings.Repeat("t", 64)
		tasks[i].Description = strings.Repeat("n", 128)
		tasks[i].CreatedAt = tasks[i].CreatedAt.UTC()
		tasks[i].Tags = []core.Tag{"alpha", "beta", "gamma"}
		tasks[i].DueDate = &tasks[i].CreatedAt
		tasks[i].CompletedAt = &tasks[i].UpdatedAt
	}
	allocs := testing.AllocsPerRun(100, func() {
		if data, err := encodeJSON(tasks); err != nil || len(data) == 0 {
			t.Fatalf("encode: %v", err)
		}
	})
	if allocs > 20 {
		t.Fatalf("100-task JSON needs %.0f allocations; want at most 20", allocs)
	}
	data, err := encodeJSON(tasks)
	if err != nil {
		t.Fatal(err)
	}
	measured := testing.Benchmark(func(b *testing.B) {
		for range b.N {
			if _, err := encodeJSON(tasks); err != nil {
				b.Fatal(err)
			}
		}
	})
	if n := measured.AllocedBytesPerOp(); n > int64(len(data)*5/4) {
		t.Fatalf("wire encoding allocates %d bytes for %d output bytes; want <= 125%%", n, len(data))
	}
}

func TestJSON_TreeEncodingAllocation(t *testing.T) {
	nodes := make([]core.TaskNode, 1000)
	roots := make([]*core.TaskNode, len(nodes))
	for i := range nodes {
		nodes[i] = core.TaskNode{Task: jsonFixture(), Depth: 1}
		nodes[i].Task.Title = strings.Repeat("t", 32)
		nodes[i].Task.Description = strings.Repeat("n", 128)
		nodes[i].Task.CreatedAt = time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC)
		nodes[i].Task.UpdatedAt = nodes[i].Task.CreatedAt
		roots[i] = &nodes[i]
	}
	data, err := encodeJSON(roots)
	if err != nil {
		t.Fatal(err)
	}
	measured := testing.Benchmark(func(b *testing.B) {
		for range b.N {
			if _, err := encodeJSON(roots); err != nil {
				b.Fatal(err)
			}
		}
	})
	if n := measured.AllocedBytesPerOp(); n > int64(len(data)*11/10) {
		t.Fatalf("tree JSON reserves %d bytes for %d output bytes; want <= 110%%", n, len(data))
	}
}

func TestJSON_EncodingParity(t *testing.T) {
	check := func(value any) {
		t.Helper()
		dto, err := referenceJSONDTO(value)
		if err != nil {
			t.Fatal(err)
		}
		want, wantErr := json.Marshal(dto)
		got, gotErr := encodeJSON(value)
		if (wantErr == nil) != (gotErr == nil) || wantErr == nil && !bytes.Equal(got, append(want, '\n')) {
			t.Fatalf("wire mismatch: got %q / %v; want %q / %v", got, gotErr, want, wantErr)
		}
	}
	for c := range 256 {
		task := jsonFixture()
		task.Description = "prefix" + string([]byte{byte(c)}) + "suffix"
		task.Title = "界 👩‍💻 é \u2028\u2029 \u202e <&> \" \\"
		task.Tags = []core.Tag{"alpha", "beta"}
		task.ParentID = &task.ID
		task.DueDate = &task.CreatedAt
		check(&task)
		check([]core.Task{task, task})
		check([]*core.TaskNode{{Task: task, Depth: 1, Children: []*core.TaskNode{{Task: task, Depth: 2}}}})
	}
	for _, year := range []int{-1, 0, 1, 9999, 10000} {
		for _, field := range []string{"created", "updated", "due", "completed"} {
			task := jsonFixture()
			tm := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
			switch field {
			case "created":
				task.CreatedAt = tm
			case "updated":
				task.UpdatedAt = tm
			case "due":
				task.DueDate = &tm
			case "completed":
				task.CompletedAt = &tm
			}
			check(&task)
			check([]core.Task{jsonFixture(), task})
			check([]*core.TaskNode{{Task: task, Depth: 1}})
			check([]*core.TaskNode{{Task: jsonFixture(), Depth: 1, Children: []*core.TaskNode{{Task: task, Depth: 2}}}})
		}
	}
}

func BenchmarkJSON(b *testing.B) {
	nodes := make([]core.TaskNode, 1000)
	roots := make([]*core.TaskNode, len(nodes))
	for i := range nodes {
		nodes[i] = core.TaskNode{Task: jsonFixture(), Depth: 1}
		nodes[i].Task.Title = strings.Repeat("t", 64)
		nodes[i].Task.Description = strings.Repeat("n", 128)
		roots[i] = &nodes[i]
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if _, err := encodeJSON(roots); err != nil {
			b.Fatal(err)
		}
	}
}

// Standard-library oracle for the explicit task schema, independent of the
// production append encoder. Batch DTO construction is only needed in tests.
type nodeJSON struct {
	Task     taskJSON   `json:"task"`
	Children []nodeJSON `json:"children"`
	Depth    int        `json:"depth"`
}

func treeDTO(nodes []*core.TaskNode) ([]nodeJSON, error) {
	out := make([]nodeJSON, len(nodes))
	for n, node := range nodes {
		if node == nil {
			return nil, ports.ErrInvalidRecord
		}
		children, err := treeDTO(node.Children)
		if err != nil {
			return nil, err
		}
		out[n] = nodeJSON{taskDTO(node.Task), children, node.Depth}
	}
	return out, nil
}
func referenceJSONDTO(value any) (any, error) {
	switch v := value.(type) {
	case *core.Task:
		if v == nil {
			return nil, ports.ErrInvalidRecord
		}
		return taskDTO(*v), nil
	case []core.Task:
		out := make([]taskJSON, len(v))
		for n, t := range v {
			out[n] = taskDTO(t)
		}
		return out, nil
	case []*core.TaskNode:
		return treeDTO(v)
	default:
		return jsonDTO(value)
	}
}
