package main

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/muesli/termenv"
	"github.com/newbpydev/tusk/internal/cli"
	"github.com/newbpydev/tusk/internal/ports"
	"github.com/newbpydev/tusk/internal/service"
	"github.com/newbpydev/tusk/internal/storage"
	"github.com/newbpydev/tusk/internal/tui"
)

func tuiRunner(input io.Reader, output, diagnostics io.Writer) func(context.Context, cli.Config) (cli.TUIResult, error) {
	return func(ctx context.Context, cfg cli.Config) (cli.TUIResult, error) {
		profile := termenv.Ascii
		if os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "" {
			profile = termenv.ANSI
		}
		result, err := tui.Run(ctx, tui.RunOptions{Input: input, Output: output, Open: tuiFactory(cfg), Location: cfg.Location, Profile: profile,
			CleanupProgress: func() { _, _ = io.WriteString(diagnostics, "tusk: cleanup in progress; waiting for active work\n") }})
		return cli.TUIResult{HadCommittedChanges: result.HadCommittedChanges, OutcomeUnknown: result.OutcomeUnknown}, err
	}
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
