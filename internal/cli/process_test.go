package cli

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"github.com/newbpydev/tusk/internal/service"
	"github.com/newbpydev/tusk/internal/storage"
)

type killRepository struct {
	ports.TaskRepository
	stage string
}

func killBarrier() {
	fmt.Fprintln(os.Stdout, "TUSK-CLI-BARRIER")
	var b [1]byte
	_, _ = os.Stdin.Read(b[:])
}
func (r killRepository) WithWrite(ctx context.Context, fn func(context.Context, ports.TaskWriter) error) error {
	err := r.TaskRepository.WithWrite(ctx, func(ctx context.Context, w ports.TaskWriter) error {
		if err := fn(ctx, w); err != nil {
			return err
		}
		if r.stage == "staged" {
			killBarrier()
		}
		return nil
	})
	if err == nil && r.stage == "committed" {
		killBarrier()
	}
	return err
}
func TestCLIProcessKillHelper(t *testing.T) {
	stage := os.Getenv("TUSK_CLI_KILL_STAGE")
	if stage == "" {
		t.Skip("child only")
	}
	opts := Options{Stdout: os.Stdout, Stderr: os.Stderr, OpenService: func(ctx context.Context, _ Config) (ports.TaskService, func() error, error) {
		r, err := storage.Open(ctx, storage.Options{Path: os.Getenv("TUSK_CLI_KILL_DB")})
		if err != nil {
			return nil, nil, err
		}
		s, err := service.NewTaskService(killRepository{r, stage}, service.Options{Clock: time.Now, NewID: func(now time.Time) (string, error) { return service.NewUUIDv7(now, bytes.NewReader(make([]byte, 10))) }, Location: time.UTC})
		return s, r.Close, err
	}}
	if code := Run(context.Background(), []string{"add", "crash recovery", "--json"}, opts); code != 0 {
		t.Fatal(code)
	}
}
func TestProcess_HardKillReadback(t *testing.T) {
	if testing.Short() {
		t.Skip("hard process kill")
	}
	for _, stage := range []string{"staged", "committed"} {
		t.Run(stage, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "tasks.db")
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestCLIProcessKillHelper$")
			cmd.Env = append(os.Environ(), "TUSK_CLI_KILL_STAGE="+stage, "TUSK_CLI_KILL_DB="+path)
			out, err := cmd.StdoutPipe()
			if err != nil {
				t.Fatal(err)
			}
			in, err := cmd.StdinPipe()
			if err != nil {
				t.Fatal(err)
			}
			defer in.Close()
			cmd.Stderr = os.Stderr
			if err = cmd.Start(); err != nil {
				t.Fatal(err)
			}
			defer func() { _ = cmd.Process.Kill(); _ = cmd.Wait() }()
			scanner := bufio.NewScanner(out)
			if !scanner.Scan() || scanner.Text() != "TUSK-CLI-BARRIER" {
				t.Fatal("barrier not reached")
			}
			if err = cmd.Process.Kill(); err != nil {
				t.Fatal(err)
			}
			if err = cmd.Wait(); err == nil {
				t.Fatal("killed process succeeded")
			}
			fresh, err := storage.Open(context.Background(), storage.Options{Path: path})
			if err != nil {
				t.Fatal(err)
			}
			defer fresh.Close()
			tasks, err := fresh.List(context.Background(), core.TaskFilter{})
			want := 0
			if stage == "committed" {
				want = 1
			}
			if err != nil || len(tasks) != want {
				t.Fatalf("tasks %d want %d: %v", len(tasks), want, err)
			}
			if want == 1 {
				events, err := fresh.ListEvents(context.Background(), tasks[0].ID)
				if err != nil || len(events) != 1 {
					t.Fatalf("events %d: %v", len(events), err)
				}
			}
		})
	}
}

type processFixture struct{ binary, dir, path string }

func (p processFixture) command(ctx context.Context, args ...string) *exec.Cmd {
	c := exec.CommandContext(ctx, p.binary, args...)
	c.Dir = p.dir
	c.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + p.dir, "USERPROFILE=" + p.dir, "TUSK_DB_PATH=" + p.path, "TUSK_TIMEZONE=UTC", "TZ=UTC", "TERM=dumb"}
	return c
}
func (p processFixture) run(args ...string) (int, []byte, string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	c := p.command(ctx, args...)
	var out, errout bytes.Buffer
	c.Stdout = &out
	c.Stderr = &errout
	err := c.Run()
	if err == nil {
		return 0, out.Bytes(), errout.String()
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return exit.ExitCode(), out.Bytes(), errout.String()
	}
	return -1, nil, err.Error()
}
func TestProcess_Workflow(t *testing.T) {
	if testing.Short() {
		t.Skip("actual executable")
	}
	binary := processBinary(t)
	t.Run("isolation", func(t *testing.T) {
		home := processDirectory(t)
		p := processFixture{binary, home, filepath.Join(home, "missing", "db")}
		for _, args := range [][]string{{"--help"}, {"--version"}, {"unknown"}, {"add"}, {"edit", "id", "--progress=bad"}} {
			code, out, errout := p.run(args...)
			want := 2
			if args[0] == "--help" || args[0] == "--version" {
				want = 0
			}
			if code != want {
				t.Fatalf("%q %d %s %s", args, code, out, errout)
			}
		}
		if _, err := os.Stat(filepath.Dir(p.path)); !os.IsNotExist(err) {
			t.Fatalf("syntax opened storage: %v", err)
		}
	})
	t.Run("durable workflow", func(t *testing.T) {
		home := processDirectory(t)
		p := processFixture{binary, home, filepath.Join(home, "literal ?#% ' 界.db")}
		run := func(args ...string) any {
			t.Helper()
			code, data, errout := p.run(append(args, "--json")...)
			if code != 0 || errout != "" {
				t.Fatalf("%q %d %s", args, code, errout)
			}
			return decodeJSON(t, data)
		}
		root := run("add", "root").(map[string]any)["id"].(string)
		child := run("add", "child", "--parent", root, "--notes=complete notes", "--tags=a,b", "--due=tomorrow").(map[string]any)["id"].(string)
		if len(run("list").([]any)) != 2 || len(run("tree").([]any)) != 1 {
			t.Fatal("read count")
		}
		run("done", root)
		run("edit", root, "--status=todo")
		run("edit", child, "--root", "--clear-due", "--status=todo")
		if len(run("history", child).([]any)) < 3 {
			t.Fatal("history")
		}
		run("delete", root, "--force", "--recursive")
		run("delete", child, "--force")
		if run("stats").(map[string]any)["total"] != json.Number("0") {
			t.Fatal("remaining tasks")
		}
	})
	t.Run("broken pipe preserves commit", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Unix SIGPIPE contract")
		}
		for _, brokenStderr := range []bool{false, true} {
			home := processDirectory(t)
			p := processFixture{binary, home, filepath.Join(home, "tasks.db")}
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			if err = r.Close(); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			c := p.command(ctx, "add", "committed once", "--json")
			c.Stdout = w
			var stderr bytes.Buffer
			if brokenStderr {
				c.Stderr = w
			} else {
				c.Stderr = &stderr
			}
			err = c.Run()
			cancel()
			if e := w.Close(); e != nil {
				t.Fatal(e)
			}
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 1 {
				t.Fatalf("pipe exit %v %s", err, stderr.String())
			}
			code, data, errout := p.run("list", "--all", "--json")
			if code != 0 || len(decodeJSON(t, data).([]any)) != 1 {
				t.Fatalf("readback %d %s %s", code, data, errout)
			}
		}
	})
	t.Run("concurrent writers", func(t *testing.T) {
		home := processDirectory(t)
		p := processFixture{binary, home, filepath.Join(home, "tasks.db")}
		create := func(args ...string) string {
			t.Helper()
			code, data, errout := p.run(append(args, "--json")...)
			if code != 0 {
				t.Fatalf("%d %s", code, errout)
			}
			return decodeJSON(t, data).(map[string]any)["id"].(string)
		}
		root := create("add", "root")
		a := create("add", "a", "--parent", root)
		b := create("add", "b", "--parent", root)
		results := make(chan string, 2)
		for _, id := range []string{a, b} {
			go func(id string) {
				code, out, stderr := p.run("done", id, "--json")
				if code != 0 {
					results <- fmt.Sprintf("%d %s %s", code, out, stderr)
				} else {
					results <- ""
				}
			}(id)
		}
		for range 3 {
			code, data, stderr := p.run("tree", "--json")
			if code != 0 || len(decodeJSON(t, data).([]any)) != 1 {
				t.Fatalf("reader %d %s", code, stderr)
			}
		}
		for range 2 {
			if err := <-results; err != "" {
				t.Fatal(err)
			}
		}
		code, data, stderr := p.run("tree", "--json")
		if code != 0 {
			t.Fatal(stderr)
		}
		node := decodeJSON(t, data).([]any)[0].(map[string]any)
		if node["task"].(map[string]any)["progress"] != json.Number("100") {
			t.Fatal(node)
		}
	})
	t.Run("unsafe paths", func(t *testing.T) {
		for _, kind := range []string{"directory", "corrupt", "symlink", "readonly", "newer"} {
			home := processDirectory(t)
			path := filepath.Join(home, "PRIVATE.db")
			switch kind {
			case "directory":
				if err := os.Mkdir(path, 0700); err != nil {
					t.Fatal(err)
				}
			case "corrupt":
				if err := os.WriteFile(path, []byte("PRIVATE invalid SQLite"), 0600); err != nil {
					t.Fatal(err)
				}
			case "symlink":
				if runtime.GOOS == "windows" {
					continue
				}
				if err := os.Symlink(filepath.Join(home, "missing-target"), path); err != nil {
					t.Fatal(err)
				}
			case "readonly", "newer":
				if kind == "readonly" && runtime.GOOS == "windows" {
					continue
				}
				p := processFixture{binary, home, path}
				if code, _, stderr := p.run("stats"); code != 0 {
					t.Fatal(stderr)
				}
				if kind == "newer" {
					db, err := sql.Open("sqlite", path)
					if err != nil {
						t.Fatal(err)
					}
					_, err = db.Exec("UPDATE schema_migrations SET version=999")
					if e := db.Close(); err != nil || e != nil {
						t.Fatalf("%v %v", err, e)
					}
				} else {
					if err := os.Chmod(path, 0400); err != nil {
						t.Fatal(err)
					}
					defer os.Chmod(path, 0600)
				}
			}
			p := processFixture{binary, home, path}
			code, data, stderr := p.run("add", "must fail", "--json")
			if code != 1 || len(data) != 0 || strings.Contains(stderr, "PRIVATE") {
				t.Fatalf("%s: %d %s %s", kind, code, data, stderr)
			}
		}
	})
}

type uncertainRepository struct {
	ports.TaskRepository
	after bool
	calls int
}

func (r *uncertainRepository) WithWrite(ctx context.Context, fn func(context.Context, ports.TaskWriter) error) error {
	r.calls++
	if r.after {
		if err := r.TaskRepository.WithWrite(ctx, fn); err != nil {
			return err
		}
	}
	return ports.NewTransactionError("commit", context.Canceled)
}
func TestProcess_UnknownOutcomeReadback(t *testing.T) {
	for _, after := range []bool{false, true} {
		t.Run(fmt.Sprint(after), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "data.db")
			repo, err := storage.Open(context.Background(), storage.Options{Path: path})
			if err != nil {
				t.Fatal(err)
			}
			uncertain := &uncertainRepository{TaskRepository: repo, after: after}
			svc, err := service.NewTaskService(uncertain, service.Options{Clock: time.Now, NewID: func(now time.Time) (string, error) { return service.NewUUIDv7(now, bytes.NewReader(make([]byte, 10))) }, Location: time.UTC})
			if err != nil {
				t.Fatal(err)
			}
			closed := 0
			var out, stderr bytes.Buffer
			opts := Options{Stdout: &out, Stderr: &stderr, OpenService: func(context.Context, Config) (ports.TaskService, func() error, error) {
				return svc, func() error { closed++; return repo.Close() }, nil
			}}
			if code := Run(context.Background(), []string{"add", "uncertain", "--json"}, opts); code != 1 || out.Len() != 0 || closed != 1 || uncertain.calls != 1 || !strings.Contains(stderr.String(), "outcome unknown") {
				t.Fatalf("%d close %d calls %d %s", code, closed, uncertain.calls, stderr.String())
			}
			fresh, err := storage.Open(context.Background(), storage.Options{Path: path})
			if err != nil {
				t.Fatal(err)
			}
			defer fresh.Close()
			tasks, err := fresh.List(context.Background(), core.TaskFilter{})
			want := 0
			if after {
				want = 1
			}
			if err != nil || len(tasks) != want {
				t.Fatalf("%d tasks %v", len(tasks), err)
			}
			if after {
				events, err := fresh.ListEvents(context.Background(), tasks[0].ID)
				if err != nil || len(events) != 1 {
					t.Fatalf("%d events %v", len(events), err)
				}
			}
		})
	}
}
