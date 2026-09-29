package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/muesli/termenv"
	"github.com/newbpydev/tusk/internal/cli"
	"github.com/newbpydev/tusk/internal/ports"
	"github.com/newbpydev/tusk/internal/service"
	"github.com/newbpydev/tusk/internal/service/dateparse"
	"github.com/newbpydev/tusk/internal/storage"
	"github.com/newbpydev/tusk/internal/tui"
)

func tuiRunner(input io.Reader, output, diagnostics io.Writer) func(context.Context, cli.Config) (cli.TUIResult, error) {
	return func(ctx context.Context, cfg cli.Config) (cli.TUIResult, error) {
		result, err := tui.Run(ctx, tui.RunOptions{Input: input, Output: output, Open: tuiFactory(cfg), Location: cfg.Location, Profile: tuiProfile(os.Getenv), DayBounds: dateparse.DayBounds, ParseDue: dateparse.ParseDue,
			CleanupProgress: func() { _, _ = io.WriteString(diagnostics, "tusk: cleanup in progress; waiting for active work\n") }})
		return cli.TUIResult{HadCommittedChanges: result.HadCommittedChanges, OutcomeUnknown: result.OutcomeUnknown}, err
	}
}

// Select color capability from declared terminal facts, without querying the
// terminal, changing a global renderer, or probing background appearance.
func tuiProfile(getenv func(string) string) termenv.Profile {
	term := getenv("TERM")
	if getenv("NO_COLOR") != "" || term == "" || term == "dumb" {
		return termenv.Ascii
	}
	if color := getenv("COLORTERM"); color == "truecolor" || color == "24bit" {
		return termenv.TrueColor
	}
	if strings.Contains(term, "256color") {
		return termenv.ANSI256
	}
	return termenv.ANSI
}

func tuiFactory(cfg cli.Config) tui.Factory {
	// Only the session's admitted command accesses this closure. Resolve once,
	// at first open, so recovery cannot redirect after environment/cwd changes.
	var path string
	return func(ctx context.Context) (ports.TaskService, func() error, error) {
		if path == "" {
			if err := ctx.Err(); err != nil {
				return nil, nil, err
			}
			candidate := os.Getenv("TUSK_DB_PATH")
			if candidate == "" {
				base := os.Getenv("XDG_DATA_HOME")
				if !filepath.IsAbs(base) {
					home, err := os.UserHomeDir()
					if err != nil || !filepath.IsAbs(home) {
						return nil, nil, ports.ErrStorage
					}
					base = filepath.Join(home, ".local", "share")
				}
				candidate = filepath.Join(base, "tusk", "tusk.db")
			}
			var err error
			path, err = filepath.Abs(candidate)
			if err != nil {
				return nil, nil, ports.ErrStorage
			}
		}
		return composeService(ctx, cfg, func(ctx context.Context, _ storage.Options) (*storage.Repository, error) {
			return storage.Open(ctx, storage.Options{Path: path})
		}, service.NewTaskService)
	}
}
