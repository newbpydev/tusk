package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/newbpydev/tusk/internal/service"
	"github.com/newbpydev/tusk/internal/storage"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

type deleteSpy struct {
	ports.TaskService
	preview               ports.DeletePreview
	previewErr, deleteErr error
	previews, deletes     int
	command               ports.DeleteTaskCommand
}

func (s *deleteSpy) PreviewDeleteTask(context.Context, string) (ports.DeletePreview, error) {
	s.previews++
	return s.preview, s.previewErr
}
func (s *deleteSpy) DeleteTask(_ context.Context, c ports.DeleteTaskCommand) (ports.DeleteResult, error) {
	s.deletes++
	s.command = c
	return ports.DeleteResult{ID: c.ID, DeletedIDs: []string{c.ID}, DeletedCount: 1, Deleted: true}, s.deleteErr
}
func TestDelete_ForceNeverImpliesRecursive(t *testing.T) {
	s := &deleteSpy{}
	code, out, errout, _ := runSpy([]string{"delete", "id", "--force", "--json"}, s)
	if code != 0 || s.previews != 0 || s.deletes != 1 || !s.command.Force || s.command.Recursive {
		t.Fatalf("%d %q %q %+v", code, out, errout, s)
	}
	decodeJSON(t, []byte(out))
}
func TestDelete_NonTTYNeverReads(t *testing.T) {
	for _, facts := range []TerminalFacts{{}, {In: true, Out: true}, {In: true, Err: true}, {Out: true, Err: true}, {In: true, Out: true, Err: true}} {
		s := &deleteSpy{}
		var errout bytes.Buffer
		opened := 0
		opts := Options{Stderr: &errout, Terminal: func() TerminalFacts { return facts }, Confirm: func(context.Context) (bool, error) { t.Fatal("read"); return false, nil }, OpenService: func(context.Context, Config) (ports.TaskService, func() error, error) {
			opened++
			return s, func() error { return nil }, nil
		}}
		args := []string{"delete", "id"}
		if facts.In && facts.Out && facts.Err {
			args = append(args, "--json")
		}
		if got := Run(context.Background(), args, opts); got != 1 || opened != 0 || !strings.Contains(errout.String(), "--force") {
			t.Fatalf("%d %q open %d", got, errout.String(), opened)
		}
	}
}
func TestDelete_Confirmation(t *testing.T) {
	for _, tc := range []struct {
		name                              string
		yes                               bool
		confirmErr, previewErr, deleteErr error
		ids                               []string
		recursive                         bool
		writerFail                        bool
		exit, calls                       int
	}{
		{name: "yes", yes: true, ids: []string{"id"}, calls: 1},
		{name: "decline", ids: []string{"id"}},
		{name: "read error", ids: []string{"id"}, confirmErr: errors.New("PRIVATE"), exit: 1},
		{name: "cancel", ids: []string{"id"}, confirmErr: context.Canceled, exit: 1},
		{name: "preview error", previewErr: core.ErrTaskNotFound, exit: 1},
		{name: "children", ids: []string{"id", "child"}, yes: true, exit: 1},
		{name: "recursive", ids: []string{"id", "child"}, yes: true, recursive: true, calls: 1},
		{name: "stale", ids: []string{"id"}, yes: true, deleteErr: ports.ErrConflict, exit: 1, calls: 1},
		{name: "unknown", ids: []string{"id"}, yes: true, deleteErr: ports.NewTransactionError("commit", ports.ErrStorage), exit: 1, calls: 1},
		{name: "prompt write", ids: []string{"id"}, yes: true, writerFail: true, exit: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out, errout bytes.Buffer
			s := &deleteSpy{preview: ports.DeletePreview{Target: core.Task{ID: "id", Title: "bad\x1b]52;\a"}, IDs: tc.ids}, previewErr: tc.previewErr, deleteErr: tc.deleteErr}
			closed, reads := 0, 0
			opts := Options{Stdout: &out, Stderr: &errout, Terminal: func() TerminalFacts { return TerminalFacts{In: true, Out: true, Err: true, Width: 80} }, Confirm: func(context.Context) (bool, error) { reads++; return tc.yes, tc.confirmErr }, OpenService: func(context.Context, Config) (ports.TaskService, func() error, error) {
				return s, func() error { closed++; return nil }, nil
			}}
			if tc.writerFail {
				opts.Stderr = badWriter{}
			}
			args := []string{"delete", "id"}
			if tc.recursive {
				args = append(args, "--recursive")
			}
			got := Run(context.Background(), args, opts)
			if got != tc.exit || s.deletes != tc.calls || closed != 1 || strings.Contains(errout.String(), "PRIVATE") || strings.ContainsRune(errout.String(), 27) {
				t.Fatalf("%d close %d delete %d stderr %q", got, closed, s.deletes, errout.String())
			}
			if s.deletes > 0 && !reflect.DeepEqual(s.command.Expected, &s.preview) {
				t.Fatal("consent changed")
			}
			if tc.writerFail && reads != 0 {
				t.Fatal("read after failed prompt")
			}
			if tc.calls == 0 && out.Len() != 0 {
				t.Fatal("unexpected output")
			}
		})
	}
}
func TestDelete_StalePreview(t *testing.T) {
	for _, change := range []string{"membership", "target"} {
		t.Run(change, func(t *testing.T) {
			svc, opts := diskService(t)
			root, err := svc.CreateTask(context.Background(), ports.CreateTaskCommand{Title: "parent"})
			if err != nil {
				t.Fatal(err)
			}
			opts.Terminal = func() TerminalFacts { return TerminalFacts{In: true, Out: true, Err: true, Width: 80} }
			opts.Confirm = func(context.Context) (bool, error) {
				if change == "membership" {
					_, err = svc.CreateTask(context.Background(), ports.CreateTaskCommand{Title: "new child", ParentID: &root.ID})
				} else {
					title := "changed"
					_, err = svc.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: root.ID, Title: &title})
				}
				if err != nil {
					t.Fatal(err)
				}
				return true, nil
			}
			if code := Run(context.Background(), []string{"delete", root.ID, "--recursive"}, opts); code != 1 {
				t.Fatalf("stale exit %d", code)
			}
			if _, err = svc.GetTask(context.Background(), root.ID); err != nil {
				t.Fatalf("deleted after stale preview: %v", err)
			}
			if code := Run(context.Background(), []string{"delete", root.ID, "--force", "--recursive"}, opts); code != 0 {
				t.Fatal(code)
			}
		})
	}
}

func TestDelete_CommittedOutputFailure(t *testing.T) {
	svc, opts := diskService(t)
	task, err := svc.CreateTask(context.Background(), ports.CreateTaskCommand{Title: "remove"})
	if err != nil {
		t.Fatal(err)
	}
	var errout bytes.Buffer
	opts.Stdout = badWriter{}
	opts.Stderr = &errout
	if code := Run(context.Background(), []string{"delete", task.ID, "--force", "--json"}, opts); code != 1 || !strings.Contains(errout.String(), "committed") {
		t.Fatalf("%d %s", code, errout.String())
	}
	if _, err = svc.GetTask(context.Background(), task.ID); !errors.Is(err, core.ErrTaskNotFound) {
		t.Fatalf("delete rolled back: %v", err)
	}
}

func TestDelete_SecondOwnerBarrier(t *testing.T) {
	for _, change := range []string{"add", "remove", "move", "metadata"} {
		t.Run(change, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "data.db")
			one, err := storage.Open(context.Background(), storage.Options{Path: path})
			if err != nil {
				t.Fatal(err)
			}
			defer one.Close()
			two, err := storage.Open(context.Background(), storage.Options{Path: path})
			if err != nil {
				t.Fatal(err)
			}
			defer two.Close()
			seq := 0
			config := service.Options{Location: time.UTC, Clock: func() time.Time { return time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC) }, NewID: func(time.Time) (string, error) { seq++; return fmt.Sprintf("01987654-1234-7000-8000-%012d", seq), nil }}
			first, err := service.NewTaskService(one, config)
			if err != nil {
				t.Fatal(err)
			}
			second, err := service.NewTaskService(two, config)
			if err != nil {
				t.Fatal(err)
			}
			root, err := first.CreateTask(context.Background(), ports.CreateTaskCommand{Title: "root"})
			if err != nil {
				t.Fatal(err)
			}
			child, err := first.CreateTask(context.Background(), ports.CreateTaskCommand{Title: "child", ParentID: &root.ID})
			if err != nil {
				t.Fatal(err)
			}
			opts := Options{Terminal: func() TerminalFacts { return TerminalFacts{In: true, Out: true, Err: true, Width: 80} }, OpenService: func(context.Context, Config) (ports.TaskService, func() error, error) {
				return first, func() error { return nil }, nil
			}, Confirm: func(context.Context) (bool, error) {
				switch change {
				case "add":
					_, err = second.CreateTask(context.Background(), ports.CreateTaskCommand{Title: "new", ParentID: &root.ID})
				case "remove":
					_, err = second.DeleteTask(context.Background(), ports.DeleteTaskCommand{ID: child.ID, Force: true})
				case "move":
					_, err = second.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: child.ID, ClearParent: true})
				case "metadata":
					title := "changed"
					_, err = second.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: root.ID, Title: &title})
				}
				if err != nil {
					t.Fatal(err)
				}
				return true, nil
			}}
			if code := Run(context.Background(), []string{"delete", root.ID, "--recursive"}, opts); code != 1 {
				t.Fatalf("stale consent exit %d", code)
			}
			if _, err = second.GetTask(context.Background(), root.ID); err != nil {
				t.Fatal(err)
			}
		})
	}
}
