package storage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

// Fixed fixture: 128-byte notes, three tags, groups of one depth-ten chain.
// All population, cleanup and path allocation are outside the timed loops.
func BenchmarkStorage(b *testing.B) {
	ctx := context.Background()
	for _, size := range []int{0, 100, 1000, 10000} {
		b.Run(fmt.Sprintf("tasks=%d", size), func(b *testing.B) {
			path := filepath.Join(b.TempDir(), "benchmark.db")
			r, err := Open(ctx, Options{Path: path})
			if err != nil {
				b.Fatal(err)
			}
			err = r.WithWrite(ctx, func(c context.Context, w ports.TaskWriter) error {
				for i := 0; i < size; i++ {
					task := taskFixture(fmt.Sprintf("task-%05d", i))
					task.Description = strings.Repeat("n", 128)
					task.Tags = []core.Tag{"a", "m", "z"}
					if i%10 != 0 {
						parent := fmt.Sprintf("task-%05d", i-1)
						task.ParentID = &parent
					}
					if err := w.Create(c, task); err != nil {
						return err
					}
				}
				return nil
			})
			if err != nil {
				b.Fatal(err)
			}
			if err := r.Close(); err != nil {
				b.Fatal(err)
			}
			r, err = Open(ctx, Options{Path: path})
			if err != nil {
				b.Fatal(err)
			}
			defer r.Close()
			b.Run("GetByID", func(b *testing.B) {
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_, err := r.GetByID(ctx, "task-00000")
					if err != nil && !(size == 0 && errors.Is(err, core.ErrTaskNotFound)) {
						b.Fatal(err)
					}
				}
			})
			b.Run("List", func(b *testing.B) {
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					if _, err := r.List(ctx, core.TaskFilter{}); err != nil {
						b.Fatal(err)
					}
				}
			})
			b.Run("GetSubtree", func(b *testing.B) {
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_, err := r.GetSubtree(ctx, "task-00000")
					if err != nil && !(size == 0 && errors.Is(err, core.ErrTaskNotFound)) {
						b.Fatal(err)
					}
				}
			})
			b.Run("OpenCurrent", func(b *testing.B) {
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					next, err := Open(ctx, Options{Path: path})
					if err != nil {
						b.Fatal(err)
					}
					if err := next.Close(); err != nil {
						b.Fatal(err)
					}
				}
			})
		})
	}
	b.Run("FirstCreate", func(b *testing.B) {
		dir := b.TempDir()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			path := filepath.Join(dir, fmt.Sprintf("new-%d.db", i))
			b.StartTimer()
			r, err := Open(ctx, Options{Path: path})
			if err != nil {
				b.Fatal(err)
			}
			if err := r.Close(); err != nil {
				b.Fatal(err)
			}
			b.StopTimer()
			if err := os.Remove(path); err != nil {
				b.Fatal(err)
			}
			b.StartTimer()
		}
	})
}
