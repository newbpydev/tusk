package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

type querySpy struct {
	ports.TaskService
	query  ports.TaskQuery
	id     string
	calls  int
	err    error
	tasks  []core.Task
	nodes  []*core.TaskNode
	stats  ports.TaskStats
	events []ports.TaskEvent
}

func (s *querySpy) ListTasks(_ context.Context, q ports.TaskQuery) ([]core.Task, error) {
	s.query = q
	s.calls++
	return s.tasks, s.err
}
func (s *querySpy) GetTaskTree(_ context.Context, id string) ([]*core.TaskNode, error) {
	s.id = id
	s.calls++
	return s.nodes, s.err
}
func (s *querySpy) GetStats(context.Context) (ports.TaskStats, error) {
	s.calls++
	return s.stats, s.err
}
func (s *querySpy) GetTaskHistory(_ context.Context, id string) ([]ports.TaskEvent, error) {
	s.id = id
	s.calls++
	return s.events, s.err
}
func TestList_FilterContract(t *testing.T) {
	s := &querySpy{}
	code, out, errout, _ := runSpy([]string{"list", "-s", "todo", "-s", "blocked", "-p", "1", "-p", "HIGH", "-t", "B,a", "-t", "b", "--due", "today", "--search", "literal %界", "--parent", " opaque ", "--json"}, s)
	if code != 0 || out != "[]\n" || errout != "" || s.calls != 1 {
		t.Fatalf("%d %q %q", code, out, errout)
	}
	q := s.query
	if !reflect.DeepEqual(q.Filter.Statuses, []core.Status{core.StatusTodo, core.StatusBlocked}) || !reflect.DeepEqual(q.Filter.Priorities, []core.Priority{core.PriorityLow, core.PriorityHigh}) || !reflect.DeepEqual(q.Filter.Tags, []core.Tag{"a", "b"}) || q.Filter.SearchTerm != "literal %界" || *q.Filter.ParentID != "opaque" || *q.Due != "today" {
		t.Fatalf("%+v", q)
	}
	for _, args := range [][]string{{"list"}, {"list", "--all"}, {"list", "--root"}} {
		s = &querySpy{}
		code, _, _, _ = runSpy(args, s)
		if code != 0 {
			t.Fatal(code)
		}
	}
}
func TestQuery_SyntaxAndInvalid(t *testing.T) {
	for _, args := range [][]string{{"list", "extra"}, {"stats", "extra"}, {"tree", "one", "two"}, {"history"}, {"list", "--parent=x", "--root"}, {"list", "--all", "--status=done"}} {
		code, out, _, open := runSpy(args, &querySpy{})
		if code != 2 || out != "" || open != 0 {
			t.Fatalf("%q: %d open %d", args, code, open)
		}
	}
	for _, args := range [][]string{{"list", "--status=bad"}, {"list", "--priority=bad"}, {"list", "--tags=a,,b"}, {"list", "--parent="}, {"list", "--search=bad\x00"}, {"tree", " "}, {"history", " "}} {
		code, out, _, open := runSpy(args, &querySpy{})
		if code != 1 || out != "" || open != 0 {
			t.Fatalf("%q: %d open %d", args, code, open)
		}
	}
}
func TestTree_StoredParentAndDepth(t *testing.T) {
	parent := "stored-parent"
	task := jsonFixture()
	task.ParentID = &parent
	s := &querySpy{nodes: []*core.TaskNode{{Task: task, Depth: 1}}}
	code, out, _, _ := runSpy([]string{"tree", " opaque ", "--json"}, s)
	if code != 0 || s.id != "opaque" || s.calls != 1 {
		t.Fatalf("%d %+v", code, s)
	}
	node := decodeJSON(t, []byte(out)).([]any)[0].(map[string]any)
	if node["task"].(map[string]any)["parent_id"] != parent {
		t.Fatal(node)
	}
}
func TestQuery_Failures(t *testing.T) {
	for _, args := range [][]string{{"list", "--json"}, {"tree", "--json"}, {"stats", "--json"}, {"history", "id", "--json"}} {
		s := &querySpy{err: errors.New("PRIVATE"), tasks: []core.Task{jsonFixture()}}
		code, out, errout, _ := runSpy(args, s)
		if code != 1 || out != "" || s.calls != 1 || strings.Contains(errout, "PRIVATE") {
			t.Fatalf("%q: %d %q %q", args, code, out, errout)
		}
		var output bytes.Buffer
		s.err = nil
		opts := Options{Stdout: &output, OpenService: func(context.Context, Config) (ports.TaskService, func() error, error) {
			return s, func() error { return errors.New("close") }, nil
		}}
		if code := Run(context.Background(), args, opts); code != 1 || output.Len() != 0 {
			t.Fatalf("%q: %d %q", args, code, output.String())
		}
	}
}
func TestQuery_DiskReadModels(t *testing.T) {
	svc, opts := diskService(t)
	var out, errout bytes.Buffer
	opts.Stdout = &out
	opts.Stderr = &errout
	run := func(args ...string) any {
		t.Helper()
		out.Reset()
		errout.Reset()
		if code := Run(context.Background(), append(args, "--json"), opts); code != 0 {
			t.Fatalf("%q: %d %s", args, code, errout.String())
		}
		return decodeJSON(t, out.Bytes())
	}
	if len(run("list").([]any)) != 0 || len(run("tree").([]any)) != 0 {
		t.Fatal("nonempty")
	}
	root := run("add", "root", "--tags=a,b").(map[string]any)["id"].(string)
	child := run("add", "child", "--parent", root, "--due=today", "--notes=Literal %界", "--tags=a,b").(map[string]any)["id"].(string)
	if len(run("list", "--due=today", "--tags=a,b", "--search=%界", "--parent", root).([]any)) != 1 {
		t.Fatal("filter")
	}
	run("done", child)
	if len(run("list").([]any)) != 1 || len(run("list", "--all").([]any)) != 2 || len(run("list", "--status=done").([]any)) != 1 {
		t.Fatal("status filtering")
	}
	forest := run("tree", root).([]any)
	if len(forest) != 1 || len(forest[0].(map[string]any)["children"].([]any)) != 1 {
		t.Fatal("done node missing")
	}
	stats := run("stats").(map[string]any)
	if stats["completion_percent"].(json.Number) != "50" || stats["completed_last_7_days"].(json.Number) != "1" {
		t.Fatal(stats)
	}
	events := run("history", child).([]any)
	if len(events) < 2 {
		t.Fatal(events)
	}
	run("edit", child, "--status=todo")
	if run("stats").(map[string]any)["completed_last_7_days"].(json.Number) != "0" {
		t.Fatal("retained completions")
	}
	run("delete", root, "--force", "--recursive")
	if code := Run(context.Background(), []string{"history", child, "--json"}, opts); code != 1 {
		t.Fatal("deleted history accessible")
	}
	if _, err := svc.GetTask(context.Background(), root); !errors.Is(err, core.ErrTaskNotFound) {
		t.Fatal(err)
	}
}
func TestHistory_EmptyVersusMissing(t *testing.T) {
	s := &querySpy{}
	code, out, _, _ := runSpy([]string{"history", "id", "--json"}, s)
	if code != 0 || out != "[]\n" {
		t.Fatalf("%d %q", code, out)
	}
	s.err = core.ErrTaskNotFound
	code, out, _, _ = runSpy([]string{"history", "id", "--json"}, s)
	if code != 1 || out != "" {
		t.Fatalf("%d %q", code, out)
	}
}
