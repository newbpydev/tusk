package service

import (
	"context"
	"fmt"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"github.com/newbpydev/tusk/internal/storage"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type measuredRepository struct {
	ports.TaskRepository
	reads, writes, snapshots int
	duration                 time.Duration
}
type measuredReader struct {
	ports.TaskReader
	owner *measuredRepository
}

func (r measuredReader) GetByID(ctx context.Context, id string) (*core.Task, error) {
	r.owner.reads++
	return r.TaskReader.GetByID(ctx, id)
}
func (r measuredReader) List(ctx context.Context, f core.TaskFilter) ([]core.Task, error) {
	r.owner.reads++
	return r.TaskReader.List(ctx, f)
}
func (r measuredReader) ListChildren(ctx context.Context, id string) ([]core.Task, error) {
	r.owner.reads++
	return r.TaskReader.ListChildren(ctx, id)
}
func (r measuredReader) GetSubtree(ctx context.Context, id string) ([]core.Task, error) {
	r.owner.reads++
	return r.TaskReader.GetSubtree(ctx, id)
}
func (r measuredReader) GetAncestors(ctx context.Context, id string) ([]core.Task, error) {
	r.owner.reads++
	return r.TaskReader.GetAncestors(ctx, id)
}
func (r measuredReader) ListEvents(ctx context.Context, id string) ([]ports.TaskEvent, error) {
	r.owner.reads++
	return r.TaskReader.ListEvents(ctx, id)
}

type measuredWriter struct {
	measuredReader
	writer ports.TaskWriter
}

func (w measuredWriter) Create(ctx context.Context, v *core.Task) error {
	w.owner.writes++
	return w.writer.Create(ctx, v)
}
func (w measuredWriter) Update(ctx context.Context, v *core.Task) error {
	w.owner.writes++
	return w.writer.Update(ctx, v)
}
func (w measuredWriter) Delete(ctx context.Context, id string, recursive bool) ([]string, error) {
	w.owner.writes++
	return w.writer.Delete(ctx, id, recursive)
}
func (w measuredWriter) AppendEvent(ctx context.Context, e ports.TaskEvent) (int64, error) {
	w.owner.writes++
	return w.writer.AppendEvent(ctx, e)
}
func (r *measuredRepository) WithRead(ctx context.Context, fn func(context.Context, ports.TaskReader) error) error {
	start := time.Now()
	defer func() { r.duration += time.Since(start) }()
	r.snapshots++
	return r.TaskRepository.WithRead(ctx, func(ctx context.Context, reader ports.TaskReader) error {
		return fn(ctx, measuredReader{TaskReader: reader, owner: r})
	})
}
func (r *measuredRepository) WithWrite(ctx context.Context, fn func(context.Context, ports.TaskWriter) error) error {
	start := time.Now()
	defer func() { r.duration += time.Since(start) }()
	r.snapshots++
	return r.TaskRepository.WithWrite(ctx, func(ctx context.Context, w ports.TaskWriter) error {
		return fn(ctx, measuredWriter{measuredReader: measuredReader{TaskReader: w, owner: r}, writer: w})
	})
}
func BenchmarkService(b *testing.B) {
	for _, n := range []int{0, 100, 1000, 10000} {
		for _, op := range []string{"Get", "List", "Tree", "Stats", "History", "Create", "Complete", "Reopen", "Move", "Delete"} {
			if n == 0 && op != "List" && op != "Tree" && op != "Stats" && op != "Create" {
				continue
			}
			b.Run(fmt.Sprintf("%s/tasks=%d", op, n), func(b *testing.B) {
				reads, writes, snapshots := 0, 0, 0
				var duration time.Duration
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					path := filepath.Join(b.TempDir(), "bench.db")
					r, e := storage.Open(context.Background(), storage.Options{Path: path})
					if e != nil {
						b.Fatal(e)
					}
					e = r.WithWrite(context.Background(), func(ctx context.Context, w ports.TaskWriter) error {
						for j := 0; j < n; j++ {
							parent := ""
							if j > 0 {
								parent = "t00000"
								if j < 10 {
									parent = fmt.Sprintf("t%05d", j-1)
								}
							}
							status := core.StatusTodo
							if op == "Reopen" {
								status = core.StatusDone
							}
							v := task(fmt.Sprintf("t%05d", j), parent, 0, status)
							v.Description = "benchmark notes " + strings.Repeat("x", 112)
							v.Tags = []core.Tag{"a", "b", "c"}
							if e := w.Create(ctx, &v); e != nil {
								return e
							}
						}
						return nil
					})
					if e != nil {
						r.Close()
						b.Fatal(e)
					}
					measured := &measuredRepository{TaskRepository: r}
					s, e := NewTaskService(measured, Options{Clock: fixedTime, NewID: func(time.Time) (string, error) { return "0198ff00-0000-7000-8000-000000000001", nil }, Location: time.UTC})
					if e != nil {
						b.Fatal(e)
					}
					// Read once before timing to separate warm service work from storage open.
					if _, e = r.List(context.Background(), core.TaskFilter{}); e != nil {
						b.Fatal(e)
					}
					b.StartTimer()
					switch op {
					case "Get":
						_, e = s.GetTask(context.Background(), "t00000")
					case "List":
						_, e = s.ListTasks(context.Background(), ports.TaskQuery{All: true})
					case "Tree":
						_, e = s.GetTaskTree(context.Background(), "")
					case "Stats":
						_, e = s.GetStats(context.Background())
					case "History":
						_, e = s.GetTaskHistory(context.Background(), "t00000")
					case "Create":
						parent := (*string)(nil)
						if n > 0 {
							parent = ptr("t00000")
						}
						_, e = s.CreateTask(context.Background(), ports.CreateTaskCommand{Title: "new", ParentID: parent})
					case "Complete":
						_, e = s.CompleteTask(context.Background(), ports.TaskCommand{ID: "t00000"})
					case "Reopen":
						_, e = s.ReopenTask(context.Background(), ports.ReopenTaskCommand{ID: "t00001", Status: core.StatusTodo})
					case "Move":
						_, e = s.UpdateTask(context.Background(), ports.UpdateTaskCommand{ID: "t00009", ClearParent: true})
					case "Delete":
						_, e = s.DeleteTask(context.Background(), ports.DeleteTaskCommand{ID: "t00000", Force: true, Recursive: true})
					}
					b.StopTimer()
					reads += measured.reads
					writes += measured.writes
					snapshots += measured.snapshots
					duration += measured.duration
					if e != nil {
						r.Close()
						b.Fatal(e)
					}
					if e = r.Close(); e != nil {
						b.Fatal(e)
					}
				}
				b.ReportMetric(float64(reads)/float64(b.N), "repo-reads/op")
				b.ReportMetric(float64(writes)/float64(b.N), "repo-writes/op")
				b.ReportMetric(float64(snapshots)/float64(b.N), "snapshots/op")
				b.ReportMetric(float64(duration.Nanoseconds())/float64(b.N), "snapshot-ns/op")
			})
		}
	}
}
