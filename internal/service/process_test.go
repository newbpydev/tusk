package service

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"github.com/newbpydev/tusk/internal/storage"
	"os"
	"os/exec"
	"testing"
	"time"
)

type processRepository struct {
	ports.TaskRepository
	stage string
}
type processWriter struct {
	ports.TaskWriter
	stage  string
	writes *int
}

func processBarrier() { fmt.Println("TUSK-SERVICE-READY"); var b [1]byte; _, _ = os.Stdin.Read(b[:]) }
func (w processWriter) Update(ctx context.Context, v *core.Task) error {
	*w.writes++
	if *w.writes == 1 && w.stage == "before" {
		processBarrier()
	}
	e := w.TaskWriter.Update(ctx, v)
	if e == nil && *w.writes == 1 && w.stage == "partial" {
		processBarrier()
	}
	return e
}
func (r processRepository) WithWrite(ctx context.Context, fn func(context.Context, ports.TaskWriter) error) error {
	writes := 0
	return r.TaskRepository.WithWrite(ctx, func(ctx context.Context, w ports.TaskWriter) error {
		return fn(ctx, processWriter{TaskWriter: w, stage: r.stage, writes: &writes})
	})
}
func TestServiceProcessHelper(t *testing.T) {
	stage := os.Getenv("TUSK_SERVICE_CRASH")
	if stage == "" {
		t.Skip("subprocess helper")
	}
	r, e := storage.Open(context.Background(), storage.Options{Path: os.Getenv("TUSK_SERVICE_CRASH_DB")})
	if e != nil {
		t.Fatal(e)
	}
	defer r.Close()
	s := serviceFor(t, processRepository{TaskRepository: r, stage: stage}, false)
	if _, e = s.CompleteTask(context.Background(), ports.TaskCommand{ID: "p"}); e != nil {
		t.Fatal(e)
	}
	if stage == "committed" {
		processBarrier()
	}
}
func TestService_ProcessTermination(t *testing.T) {
	if testing.Short() {
		t.Skip("process kill and recovery")
	}
	for _, stage := range []string{"before", "partial", "committed"} {
		t.Run(stage, func(t *testing.T) {
			path, a, b := twoOwners(t, task("p", "", 0, core.StatusTodo), task("a", "p", 0, core.StatusTodo), task("b", "p", 0, core.StatusTodo))
			if e := a.Close(); e != nil {
				t.Fatal(e)
			}
			if e := b.Close(); e != nil {
				t.Fatal(e)
			}
			cmd := exec.Command(os.Args[0], "-test.run=^TestServiceProcessHelper$")
			cmd.Env = append(os.Environ(), "TUSK_SERVICE_CRASH="+stage, "TUSK_SERVICE_CRASH_DB="+path)
			stdout, e := cmd.StdoutPipe()
			if e != nil {
				t.Fatal(e)
			}
			stdin, e := cmd.StdinPipe()
			if e != nil {
				t.Fatal(e)
			}
			defer stdin.Close()
			cmd.Stderr = os.Stderr
			if e = cmd.Start(); e != nil {
				t.Fatal(e)
			}
			ready := make(chan bool, 1)
			go func() {
				scanner := bufio.NewScanner(stdout)
				for scanner.Scan() {
					if scanner.Text() == "TUSK-SERVICE-READY" {
						ready <- true
						return
					}
				}
				ready <- false
			}()
			select {
			case ok := <-ready:
				if !ok {
					_ = cmd.Process.Kill()
					_ = cmd.Wait()
					t.Fatal("child did not reach barrier")
				}
			case <-time.After(20 * time.Second):
				_ = cmd.Process.Kill()
				_ = cmd.Wait()
				t.Fatal("child barrier timeout")
			}
			if e = cmd.Process.Kill(); e != nil {
				t.Fatal(e)
			}
			_ = cmd.Wait()
			fresh, e := storage.Open(context.Background(), storage.Options{Path: path})
			if e != nil {
				t.Fatal(e)
			}
			defer fresh.Close()
			for _, id := range []string{"p", "a", "b"} {
				v := mustGet(t, fresh, id)
				events, e := fresh.ListEvents(context.Background(), id)
				if e != nil || (v.Status == core.StatusDone) != (stage == "committed") || len(events) != map[bool]int{true: 1, false: 0}[stage == "committed"] {
					t.Fatalf("%+v %v %v", v, events, e)
				}
			}
			db, e := sql.Open("sqlite", path)
			if e != nil {
				t.Fatal(e)
			}
			defer db.Close()
			var integrity, mode string
			if e = db.QueryRow("PRAGMA integrity_check").Scan(&integrity); e != nil || integrity != "ok" {
				t.Fatalf("%s %v", integrity, e)
			}
			if e = db.QueryRow("PRAGMA journal_mode").Scan(&mode); e != nil || mode != "wal" {
				t.Fatalf("%s %v", mode, e)
			}
			rows, e := db.Query("PRAGMA foreign_key_check")
			if e != nil {
				t.Fatal(e)
			}
			defer rows.Close()
			if rows.Next() || rows.Err() != nil {
				t.Fatal("foreign key corruption")
			}
		})
	}
}
