package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/newbpydev/tusk/internal/cli"
	"github.com/newbpydev/tusk/internal/ports"
	"github.com/newbpydev/tusk/internal/service"
	"github.com/newbpydev/tusk/internal/storage"
)

func TestOpenService(t *testing.T) {
	t.Setenv("TUSK_DB_PATH", filepath.Join(t.TempDir(), "data.db"))
	svc, closeOwner, err := openService(context.Background(), cli.Config{Location: time.UTC})
	if err != nil {
		t.Fatal(err)
	}
	task, err := svc.CreateTask(context.Background(), ports.CreateTaskCommand{Title: "real composition"})
	if err != nil || task == nil {
		t.Fatalf("task %v err %v", task, err)
	}
	if err = closeOwner(); err != nil {
		t.Fatal(err)
	}
	if _, err = svc.GetTask(context.Background(), task.ID); !errors.Is(err, ports.ErrClosedRepository) {
		t.Fatalf("closed: %v", err)
	}
}

func TestOpenService_Failures(t *testing.T) {
	t.Setenv("TUSK_DB_PATH", filepath.Join(t.TempDir(), "db"))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := openService(ctx, cli.Config{Location: time.UTC}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	var captured *storage.Repository
	_, _, err := composeService(context.Background(), cli.Config{}, func(ctx context.Context, o storage.Options) (*storage.Repository, error) {
		r, e := storage.Open(ctx, o)
		captured = r
		return r, e
	}, service.NewTaskService)
	if !errors.Is(err, ports.ErrInvalidServiceOptions) {
		t.Fatal(err)
	}
	if err = captured.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestProcessContext(t *testing.T) {
	ctx, stop := processContext()
	if ctx.Err() != nil {
		t.Fatal(ctx.Err())
	}
	stop()
	if ctx.Err() == nil {
		t.Fatal("not stopped")
	}
}

func TestOpenService_PathPrecedence(t *testing.T) {
	for _, name := range []string{"explicit", "relative explicit", "xdg", "relative xdg", "home"} {
		t.Run(name, func(t *testing.T) {
			base := t.TempDir()
			t.Chdir(base)
			t.Setenv("HOME", base)
			t.Setenv("TUSK_DB_PATH", "")
			t.Setenv("XDG_DATA_HOME", "")
			want := filepath.Join(base, ".local", "share", "tusk", "tusk.db")
			switch name {
			case "explicit":
				want = filepath.Join(base, "literal ?#%.db")
				t.Setenv("TUSK_DB_PATH", want)
			case "relative explicit":
				want = filepath.Join(base, "relative.db")
				t.Setenv("TUSK_DB_PATH", "relative.db")
			case "xdg":
				t.Setenv("XDG_DATA_HOME", filepath.Join(base, "xdg"))
				want = filepath.Join(base, "xdg", "tusk", "tusk.db")
			case "relative xdg":
				t.Setenv("XDG_DATA_HOME", "ignored-relative")
			}
			_, closeOwner, err := openService(context.Background(), cli.Config{Location: time.UTC})
			if err != nil {
				t.Fatal(err)
			}
			if err = closeOwner(); err != nil {
				t.Fatal(err)
			}
			if _, err = os.Stat(want); err != nil {
				t.Fatalf("expected database %s: %v", want, err)
			}
		})
	}
}

func TestOpenService_NoFallback(t *testing.T) {
	base := t.TempDir()
	t.Setenv("HOME", base)
	t.Setenv("XDG_DATA_HOME", filepath.Join(base, "fallback"))
	t.Setenv("TUSK_DB_PATH", base)
	if _, _, err := openService(context.Background(), cli.Config{Location: time.UTC}); err == nil {
		t.Fatal("opened directory")
	}
	if _, err := os.Stat(filepath.Join(base, "fallback")); !os.IsNotExist(err) {
		t.Fatalf("fallback touched: %v", err)
	}
}

func TestTerminalFacts(t *testing.T) {
	facts := terminalFacts()
	if facts.WidthError == nil && facts.Width < 1 {
		t.Fatalf("invalid width %+v", facts)
	}
}
