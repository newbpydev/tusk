package main

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/newbpydev/tusk/internal/cli"
	"github.com/newbpydev/tusk/internal/ports"
	"github.com/newbpydev/tusk/internal/storage"
	"github.com/newbpydev/tusk/internal/tui"
)

// The TUI factory must open exactly the path the storage layer resolves; a
// second precedence implementation here can silently fork the data space.
func TestTUIFactory_PathMatchesStorageResolution(t *testing.T) {
	for _, mode := range []string{"explicit", "relative", "xdg", "relative xdg", "home"} {
		t.Run(mode, func(t *testing.T) {
			base := t.TempDir()
			t.Chdir(base)
			t.Setenv("HOME", base)
			t.Setenv("TUSK_DB_PATH", "")
			t.Setenv("XDG_DATA_HOME", "")
			switch mode {
			case "explicit":
				t.Setenv("TUSK_DB_PATH", filepath.Join(base, "tasks.db"))
			case "relative":
				t.Setenv("TUSK_DB_PATH", "tasks.db")
			case "xdg":
				t.Setenv("XDG_DATA_HOME", filepath.Join(base, "xdg"))
			case "relative xdg":
				t.Setenv("XDG_DATA_HOME", "ignored")
			}
			want, err := storage.Resolve("")
			if err != nil {
				t.Fatal(err)
			}
			_, closeOwner, err := tuiFactory(cli.Config{Location: time.UTC})(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if err = closeOwner(); err != nil {
				t.Fatal(err)
			}
			if _, err = os.Stat(want); err != nil {
				t.Fatalf("factory opened a path other than the storage resolution %s: %v", want, err)
			}
		})
	}
}

func TestSession_ReopenSamePath(t *testing.T) {
	base := t.TempDir()
	t.Chdir(base)
	t.Setenv("TUSK_DB_PATH", "original.db")
	factory := tuiFactory(cli.Config{Location: time.UTC})
	s := tui.NewSession(context.Background(), factory)
	_, err := s.Call(context.Background(), true, func(ctx context.Context, svc ports.TaskService) (any, error) {
		return svc.CreateTask(ctx, ports.CreateTaskCommand{Title: "persistent"})
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("TUSK_DB_PATH", filepath.Join(base, "must-not-exist.db"))
	t.Chdir(t.TempDir())
	forest, err := s.Reopen(context.Background())
	if err != nil || len(forest) != 1 || forest[0].Task.Title != "persistent" {
		t.Fatalf("wrong readback: %v %v", forest, err)
	}
	if _, err := s.Close(nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(base, "must-not-exist.db")); !os.IsNotExist(err) {
		t.Fatal("recovery resampled path")
	}
}

func TestSession_PathPrecedence(t *testing.T) {
	for _, mode := range []string{"explicit", "relative", "xdg", "relative xdg", "home", "invalid"} {
		t.Run(mode, func(t *testing.T) {
			base := t.TempDir()
			t.Chdir(base)
			t.Setenv("HOME", base)
			t.Setenv("TUSK_DB_PATH", "")
			t.Setenv("XDG_DATA_HOME", "")
			want := filepath.Join(base, ".local", "share", "tusk", "tusk.db")
			switch mode {
			case "explicit":
				want = filepath.Join(base, "tasks.db")
				t.Setenv("TUSK_DB_PATH", want)
			case "relative":
				want = filepath.Join(base, "tasks.db")
				t.Setenv("TUSK_DB_PATH", "tasks.db")
			case "xdg":
				t.Setenv("XDG_DATA_HOME", filepath.Join(base, "xdg"))
				want = filepath.Join(base, "xdg", "tusk", "tusk.db")
			case "relative xdg":
				t.Setenv("XDG_DATA_HOME", "ignore")
			case "invalid":
				t.Setenv("TUSK_DB_PATH", base)
			}
			factory := tuiFactory(cli.Config{Location: time.UTC})
			_, closeOwner, err := factory(context.Background())
			if mode == "invalid" {
				if err == nil {
					t.Fatal("opened directory")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if err = closeOwner(); err != nil {
				t.Fatal(err)
			}
			if _, err = os.Stat(want); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSession_ComposedRunner(t *testing.T) {
	t.Setenv("TERM", "xterm-kitty")
	t.Setenv("NO_COLOR", "")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := tuiRunner(os.Stdin, io.Discard, io.Discard, os.Getenv)(ctx, cli.Config{Location: time.UTC})
	if err == nil {
		t.Fatal("canceled runner succeeded")
	}
	t.Setenv("TUSK_DB_PATH", filepath.Join(t.TempDir(), "missing", "tasks.db"))
	_, _, err = tuiFactory(cli.Config{Location: time.UTC})(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestSession_PathErrors(t *testing.T) {
	for _, mode := range []string{"missing home", "relative home", "missing cwd"} {
		t.Run(mode, func(t *testing.T) {
			t.Chdir(t.TempDir())
			t.Setenv("XDG_DATA_HOME", "")
			t.Setenv("TUSK_DB_PATH", "")
			t.Setenv("HOME", "")
			if mode == "relative home" {
				t.Setenv("HOME", "relative")
			}
			if mode == "missing cwd" {
				dir := t.TempDir()
				t.Chdir(dir)
				if err := os.Remove(dir); err != nil {
					t.Fatal(err)
				}
				t.Setenv("TUSK_DB_PATH", "relative.db")
			}
			_, closeOwner, err := tuiFactory(cli.Config{Location: time.UTC})(context.Background())
			if closeOwner != nil {
				_ = closeOwner()
			}
			if err == nil {
				t.Fatal("invalid path inputs accepted")
			}
		})
	}
}
