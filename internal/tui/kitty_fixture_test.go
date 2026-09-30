package tui

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/muesli/termenv"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"github.com/newbpydev/tusk/internal/service"
	"github.com/newbpydev/tusk/internal/service/dateparse"
	"github.com/newbpydev/tusk/internal/storage"
)

// This entry point runs only the real interactive adapter with a disk-backed
// service and one controlled read fault. It exits before the testing runner can
// print results. Build in Bash; use the executable solely as an app in Kitty.
func TestTUIInteractiveFixture(t *testing.T) {
	dir := os.Getenv("TUSK_TUI_KITTY_FIXTURE")
	if dir == "" {
		t.Skip("manual application fixture; not automated terminal evidence")
	}
	dir = filepath.Clean(dir)
	if filepath.Dir(dir) != os.TempDir() || !strings.HasPrefix(filepath.Base(dir), "tusk-005-") {
		fmt.Fprintln(os.Stderr, "fixture requires an owned temporary directory")
		os.Exit(2)
	}
	mode := os.Getenv("TUSK_TUI_KITTY_MODE")
	injected := false
	factory := func(ctx context.Context) (ports.TaskService, func() error, error) {
		repo, err := storage.Open(ctx, storage.Options{Path: filepath.Join(dir, "fault-app.db")})
		if err != nil {
			return nil, nil, err
		}
		svc, err := service.NewTaskService(repo, service.Options{Clock: time.Now, NewID: func(now time.Time) (string, error) { return service.NewUUIDv7(now, rand.Reader) }, Location: time.UTC})
		if err != nil {
			repo.Close()
			return nil, nil, err
		}
		closeOwner := func() error {
			err := repo.Close()
			if mode == "retire-failure" && injected {
				return ports.ErrStorage
			}
			return err
		}
		return &kittyReadFault{TaskService: svc, marker: filepath.Join(dir, "hold-read-fault"), mode: mode, injected: &injected}, closeOwner, nil
	}
	profile := termenv.TrueColor
	if os.Getenv("NO_COLOR") != "" {
		profile = termenv.Ascii
	}
	result, err := Run(context.Background(), RunOptions{Input: os.Stdin, Output: os.Stdout, Open: factory, Location: time.UTC, Profile: profile, DayBounds: dateparse.DayBounds, ParseDue: dateparse.ParseDue})
	fmt.Printf("App closed. Committed changes: %t; uncertain outcome: %t\n", result.HadCommittedChanges, result.OutcomeUnknown)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

type kittyReadFault struct {
	ports.TaskService
	marker   string
	saved    bool
	mode     string
	injected *bool
}

func (s *kittyReadFault) CreateTask(ctx context.Context, c ports.CreateTaskCommand) (*core.Task, error) {
	task, err := s.TaskService.CreateTask(ctx, c)
	if err == nil {
		s.saved = true
		if !*s.injected && (s.mode == "unknown-create" || s.mode == "retire-failure") {
			*s.injected = true
			return nil, ports.NewTransactionError("injected-create", ports.ErrBusy)
		}
	}
	return task, err
}

func (s *kittyReadFault) DeleteTask(ctx context.Context, c ports.DeleteTaskCommand) (ports.DeleteResult, error) {
	result, err := s.TaskService.DeleteTask(ctx, c)
	if err == nil && !*s.injected && s.mode == "unknown-delete" {
		*s.injected = true
		return ports.DeleteResult{}, ports.NewTransactionError("injected-delete", ports.ErrStorage)
	}
	return result, err
}
func (s *kittyReadFault) GetTaskTree(ctx context.Context, id string) ([]*core.TaskNode, error) {
	if s.saved {
		if _, err := os.Stat(s.marker); err == nil {
			return nil, ports.ErrStorage
		}
	}
	return s.TaskService.GetTaskTree(ctx, id)
}
