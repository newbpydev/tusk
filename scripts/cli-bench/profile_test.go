package main

import (
	"context"
	"io"
	"path/filepath"
	"testing"
	"time"

	"github.com/newbpydev/tusk/internal/cli"
	"github.com/newbpydev/tusk/internal/ports"
	"github.com/newbpydev/tusk/internal/service"
	"github.com/newbpydev/tusk/internal/storage"
)

// Diagnostic only: the acceptance runner measures fresh processes.
func BenchmarkCLIProfile(b *testing.B) {
	path := filepath.Join(b.TempDir(), "fixture.db")
	if _, err := seed(path, 1000, 128); err != nil {
		b.Fatal(err)
	}
	opts := cli.Options{Stdout: io.Discard, Stderr: io.Discard, Getenv: func(string) string { return "" }, Local: time.UTC,
		OpenService: func(ctx context.Context, _ cli.Config) (ports.TaskService, func() error, error) {
			r, err := storage.Open(ctx, storage.Options{Path: path})
			if err != nil {
				return nil, nil, err
			}
			s, err := service.NewTaskService(r, service.Options{Clock: time.Now, NewID: func(time.Time) (string, error) { return "unused", nil }, Location: time.UTC})
			return s, r.Close, err
		},
	}
	b.ResetTimer()
	for range b.N {
		if code := cli.Run(context.Background(), []string{"list", "--all", "--json"}, opts); code != 0 {
			b.Fatal(code)
		}
	}
}
