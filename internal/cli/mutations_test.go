package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"github.com/newbpydev/tusk/internal/service"
	"github.com/newbpydev/tusk/internal/storage"
)

type mutationSpy struct {
	ports.TaskService
	created ports.CreateTaskCommand
	updated ports.UpdateTaskCommand
	done    ports.TaskCommand
	calls   int
	err     error
}

func (s *mutationSpy) CreateTask(_ context.Context, c ports.CreateTaskCommand) (*core.Task, error) {
	s.calls++
	s.created = c
	t := jsonFixture()
	return &t, s.err
}
func (s *mutationSpy) UpdateTask(_ context.Context, c ports.UpdateTaskCommand) (*core.Task, error) {
	s.calls++
	s.updated = c
	t := jsonFixture()
	return &t, s.err
}
func (s *mutationSpy) CompleteTask(_ context.Context, c ports.TaskCommand) (*core.Task, error) {
	s.calls++
	s.done = c
	t := jsonFixture()
	return &t, s.err
}
func runSpy(args []string, s ports.TaskService) (int, string, string, int) {
	var out, errout bytes.Buffer
	opened := 0
	code := Run(context.Background(), args, Options{Stdout: &out, Stderr: &errout, OpenService: func(context.Context, Config) (ports.TaskService, func() error, error) {
		opened++
		return s, func() error { return nil }, nil
	}})
	return code, out.String(), errout.String(), opened
}

func TestAdd_ExactCommand(t *testing.T) {
	s := &mutationSpy{}
	code, out, errout, _ := runSpy([]string{"add", "界 quoted title", "-n", "notes\nline", "-p", "high", "-d", "tomorrow", "-t", "B,a", "-t", "b", "--parent", " opaque ", "--json"}, s)
	if code != 0 || s.calls != 1 || errout != "" {
		t.Fatalf("%d %q %q", code, out, errout)
	}
	c := s.created
	if c.Title != "界 quoted title" || c.Description != "notes\nline" || c.Priority != core.PriorityHigh || *c.Due != "tomorrow" || *c.ParentID != "opaque" || !reflect.DeepEqual(c.Tags, []string{"a", "b"}) {
		t.Fatalf("%+v", c)
	}
	decodeJSON(t, []byte(out))
	for _, args := range [][]string{{"add", "--", "-dash"}, {"add", "title", "--priority=low", "--priority=urgent"}, {"add", "title"}} {
		s = &mutationSpy{}
		code, _, _, _ = runSpy(args, s)
		if code != 0 || s.calls != 1 {
			t.Fatalf("%v: %d", args, code)
		}
	}
}
func TestEdit_OmittedVersusClear(t *testing.T) {
	s := &mutationSpy{}
	code, _, errout, _ := runSpy([]string{"edit", " opaque ", "--notes=", "--clear-tags", "--clear-due", "--root", "--json"}, s)
	c := s.updated
	if code != 0 || s.calls != 1 || c.ID != "opaque" || c.Description == nil || *c.Description != "" || c.Tags == nil || len(*c.Tags) != 0 || !c.ClearDue || !c.ClearParent || c.Title != nil || c.Base != nil {
		t.Fatalf("%d %+v %s", code, c, errout)
	}
	code, _, errout, _ = runSpy([]string{"edit", "id", "--title", "new", "--priority", "3", "--status", "done", "--parent", "parent", "--tags", "a,b", "--due", "today"}, s)
	if code != 0 || s.calls != 2 || *s.updated.Status != core.StatusDone || *s.updated.ParentID != "parent" || *s.updated.Priority != core.PriorityHigh {
		t.Fatalf("%d %+v %s", code, s.updated, errout)
	}
}

func TestEdit_DecimalProgress(t *testing.T) {
	for _, value := range []string{"010", "+010", "10", "00010"} {
		s := &mutationSpy{}
		code, _, stderr, _ := runSpy([]string{"edit", "id", "--progress=" + value}, s)
		if code != 0 || s.updated.Progress == nil || *s.updated.Progress != 10 {
			t.Errorf("%s: code=%d progress=%v %s", value, code, s.updated.Progress, stderr)
		}
	}
	for _, value := range []string{"0x10", "0o10", "0b10", "1_0", " 10", "9223372036854775808"} {
		s := &mutationSpy{}
		code, out, _, opened := runSpy([]string{"edit", "id", "--progress=" + value}, s)
		if code != 2 || opened != 0 || out != "" {
			t.Errorf("%s: code=%d opened=%d", value, code, opened)
		}
	}
}
func TestMutation_SyntaxAndDomains(t *testing.T) {
	for _, args := range [][]string{{"add"}, {"add", "a", "b"}, {"done"}, {"edit", "id"}, {"edit", "id", "--root=false"}, {"edit", "id", "--due=x", "--clear-due"}, {"edit", "id", "--tags=a", "--clear-tags"}, {"edit", "id", "--parent=x", "--root"}, {"edit", "id", "--progress=5", "--status=todo"}, {"edit", "id", "--progress=5", "--parent=x"}, {"edit", "id", "--progress=9223372036854775808"}, {"edit", "id", "--progress=x"}} {
		s := &mutationSpy{}
		code, out, _, opened := runSpy(args, s)
		if code != 2 || out != "" || opened != 0 || s.calls != 0 {
			t.Fatalf("%q: %d open %d", args, code, opened)
		}
	}
	for _, args := range [][]string{{"add", "bad\x00title"}, {"add", string([]byte{255})}, {"add", "x", "--priority=bad"}, {"add", "x", "--tags=a,,b"}, {"add", "x", "--parent="}, {"edit", "id", "--status=bad"}, {"edit", "id", "--priority=bad"}, {"edit", "id", "--progress=-1"}, {"edit", "id", "--progress=101"}, {"edit", "id", "--notes=bad\x00"}, {"done", " "}} {
		s := &mutationSpy{}
		code, out, _, opened := runSpy(args, s)
		if code != 1 || out != "" || opened != 0 {
			t.Fatalf("%q: %d open %d", args, code, opened)
		}
	}
	for _, args := range [][]string{{"add", "--help"}, {"edit", "--help"}, {"done", "--help"}} {
		code, _, _, opened := runSpy(args, &mutationSpy{})
		if code != 0 || opened != 0 {
			t.Fatalf("%q: %d", args, code)
		}
	}
}
func TestMutation_Failures(t *testing.T) {
	for _, args := range [][]string{{"add", "title"}, {"edit", "id", "--title=new"}, {"done", "id"}} {
		for _, err := range []error{ports.ErrConflict, ports.NewTransactionError("commit", context.Canceled), errors.New("PRIVATE")} {
			s := &mutationSpy{err: err}
			code, out, errout, _ := runSpy(args, s)
			if code != 1 || out != "" || s.calls != 1 || strings.Contains(errout, "PRIVATE") {
				t.Fatalf("%q: %d %q %q", args, code, out, errout)
			}
		}
	}
}

func diskService(t *testing.T) (ports.TaskService, Options) {
	t.Helper()
	repo, err := storage.Open(context.Background(), storage.Options{Path: filepath.Join(t.TempDir(), "data.db")})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := repo.Close(); err != nil {
			t.Error(err)
		}
	})
	seq := 0
	clock := func() time.Time { return time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC) }
	serviceOptions := service.Options{Clock: clock, NewID: func(time.Time) (string, error) { seq++; return fmt.Sprintf("01987654-1234-7000-8000-%012d", seq), nil }, Location: time.UTC}
	svc, err := service.NewTaskService(repo, serviceOptions)
	if err != nil {
		t.Fatal(err)
	}
	return svc, Options{Local: time.UTC, OpenService: func(_ context.Context, cfg Config) (ports.TaskService, func() error, error) {
		options := serviceOptions
		options.AutoCompleteParent = cfg.AutoCompleteParent
		options.Location = cfg.Location
		instance, err := service.NewTaskService(repo, options)
		return instance, func() error { return nil }, err
	}}
}
func TestDone_SubtreeAndReopen(t *testing.T) {
	svc, opts := diskService(t)
	var out, errout bytes.Buffer
	opts.Stdout = &out
	opts.Stderr = &errout
	run := func(args ...string) {
		t.Helper()
		out.Reset()
		errout.Reset()
		if code := Run(context.Background(), args, opts); code != 0 {
			t.Fatalf("%q: %d %s", args, code, errout.String())
		}
	}
	run("add", "root", "--json")
	root := decodeJSON(t, out.Bytes()).(map[string]any)["id"].(string)
	run("add", "child", "--parent", root, "--due", "tomorrow", "--json")
	child := decodeJSON(t, out.Bytes()).(map[string]any)["id"].(string)
	run("done", root)
	task, err := svc.GetTask(context.Background(), child)
	if err != nil || task.Status != core.StatusDone {
		t.Fatalf("%+v %v", task, err)
	}
	run("edit", root, "--status=todo")
	run("edit", child, "--root", "--clear-due", "--status=todo", "--notes=latest")
	task, err = svc.GetTask(context.Background(), child)
	if err != nil || task.ParentID != nil || task.DueDate != nil || task.Status != core.StatusTodo || task.Description != "latest" {
		t.Fatalf("%+v %v", task, err)
	}
	before, err := svc.GetTaskHistory(context.Background(), child)
	if err != nil {
		t.Fatal(err)
	}
	run("edit", child, "--notes=latest")
	after, err := svc.GetTaskHistory(context.Background(), child)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("no-op changed history")
	}
	run("edit", child, "--progress=+40")
	for _, title := range []string{" ", strings.Repeat("界", 256)} {
		if got := Run(context.Background(), []string{"add", title}, opts); got != 1 {
			t.Fatalf("invalid title exit %d", got)
		}
	}
	for _, title := range []string{"a", strings.Repeat("界", 255)} {
		run("add", title)
	}
}

func TestMutation_PolicyAndDates(t *testing.T) {
	for _, policy := range []string{"true", "false"} {
		t.Run(policy, func(t *testing.T) {
			svc, opts := diskService(t)
			opts.Getenv = func(k string) string {
				if k == "TUSK_AUTO_COMPLETE_PARENT" {
					return policy
				}
				return ""
			}
			var out bytes.Buffer
			opts.Stdout = &out
			run := func(args ...string) string {
				t.Helper()
				out.Reset()
				if code := Run(context.Background(), append(args, "--json"), opts); code != 0 {
					t.Fatalf("%q exit %d", args, code)
				}
				return decodeJSON(t, out.Bytes()).(map[string]any)["id"].(string)
			}
			root := run("add", "parent")
			child := run("add", "child", "--parent", root)
			run("done", child)
			parent, err := svc.GetTask(context.Background(), root)
			if err != nil || (parent.Status == core.StatusDone) != (policy == "true") {
				t.Fatalf("parent %+v err %v", parent, err)
			}
			for _, date := range []string{"today", "tomorrow", "tonight", "+1d", "+1w", "+1m", "2026-10-01", "2026-10-01T01:00:00-03:00"} {
				id := run("add", "dated", "--due", date, "--timezone=America/New_York")
				task, err := svc.GetTask(context.Background(), id)
				if err != nil || task.DueDate == nil {
					t.Fatalf("date %q %+v %v", date, task, err)
				}
			}
		})
	}
}
