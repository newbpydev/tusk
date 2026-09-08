package storage

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

// A real child process shares only the temporary database and protocol pipes.
func TestStorageProcessHelper(t *testing.T) {
	path := os.Getenv("TUSK_STORAGE_HELPER")
	if path == "" {
		t.Skip("child entry point")
	}
	r, err := Open(context.Background(), Options{Path: path})
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	input := bufio.NewReader(os.Stdin)
	wait := func() {
		if _, err := input.ReadString('\n'); err != nil {
			t.Fatal(err)
		}
	}
	fmt.Println("ready")
	wait()
	mode := os.Getenv("TUSK_STORAGE_MODE")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	count := 1
	if mode == "increment" {
		count = 20
	}
	for i := 0; i < count; i++ {
		err = r.WithWrite(ctx, func(c context.Context, w ports.TaskWriter) error {
			task, err := w.GetByID(c, "root")
			if err != nil {
				return err
			}
			if mode == "increment" {
				task.Progress++
			} else {
				task.Title = "committed change"
			}
			if err := w.Update(c, task); err != nil {
				return err
			}
			if _, err := w.AppendEvent(c, ports.TaskEvent{TaskID: "root", Kind: ports.EventMetadata, ChangedFields: []string{"title"}, OccurredAt: task.UpdatedAt}); err != nil {
				return err
			}
			if mode == "before" || mode == "hold" {
				fmt.Println("staged")
				wait()
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	fmt.Println("committed")
	wait()
}

type storageChild struct {
	cmd    *exec.Cmd
	input  io.WriteCloser
	output *bufio.Scanner
	stderr bytes.Buffer
}

func startStorageChild(t *testing.T, path, mode string) *storageChild {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	c := &storageChild{cmd: exec.CommandContext(ctx, executable, "-test.run=^TestStorageProcessHelper$")}
	c.cmd.Env = append(os.Environ(), "TUSK_STORAGE_HELPER="+path, "TUSK_STORAGE_MODE="+mode)
	c.input, err = c.cmd.StdinPipe()
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	output, err := c.cmd.StdoutPipe()
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	c.output = bufio.NewScanner(output)
	c.cmd.Stderr = &c.stderr
	if err := c.cmd.Start(); err != nil {
		cancel()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if c.cmd.ProcessState == nil {
			_ = c.cmd.Process.Kill()
			_ = c.cmd.Wait()
		}
		_ = c.input.Close()
		cancel()
	})
	c.expect(t, "ready")
	return c
}
func (c *storageChild) expect(t *testing.T, want string) {
	t.Helper()
	if !c.output.Scan() || c.output.Text() != want {
		t.Fatalf("child expected %s; got %q (%v)", want, c.output.Text(), c.output.Err())
	}
}
func (c *storageChild) send(t *testing.T) {
	t.Helper()
	if _, err := io.WriteString(c.input, "continue\n"); err != nil {
		t.Fatal(err)
	}
}
func (c *storageChild) finish(t *testing.T) {
	t.Helper()
	c.send(t)
	for c.output.Scan() {
	}
	if err := c.cmd.Wait(); err != nil {
		t.Fatalf("child: %v %s", err, c.stderr.String())
	}
}

func TestDiskWriters_NoLostUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("disk process acceptance")
	}
	r, path := diskRepository(t)
	createFixture(t, r, taskFixture("root"))
	a, b := startStorageChild(t, path, "increment"), startStorageChild(t, path, "increment")
	a.send(t)
	b.send(t)
	a.expect(t, "committed")
	b.expect(t, "committed")
	a.finish(t)
	b.finish(t)
	task, err := r.GetByID(context.Background(), "root")
	if err != nil || task.Progress != 40 {
		t.Fatalf("lost update: %+v %v", task, err)
	}
	events, err := r.ListEvents(context.Background(), "root")
	if err != nil || len(events) != 40 {
		t.Fatalf("acknowledged events: %d %v", len(events), err)
	}
}

func TestDiskReaders_SnapshotAcrossWriterCommit(t *testing.T) {
	if testing.Short() {
		t.Skip("disk process acceptance")
	}
	r, path := diskRepository(t)
	createFixture(t, r, taskFixture("root"))
	child := startStorageChild(t, path, "after")
	var journal string
	if err := r.reader.QueryRow("PRAGMA journal_mode").Scan(&journal); err != nil || journal != "wal" {
		t.Fatalf("journal=%s %v", journal, err)
	}
	err := r.WithRead(context.Background(), func(c context.Context, rd ports.TaskReader) error {
		before, err := rd.GetByID(c, "root")
		if err != nil {
			return err
		}
		child.send(t)
		child.expect(t, "committed")
		after, err := rd.GetByID(c, "root")
		if err != nil {
			return err
		}
		if before.Title != after.Title {
			return errors.New("snapshot changed")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	child.finish(t)
	task, err := r.GetByID(context.Background(), "root")
	if err != nil || task.Title != "committed change" {
		t.Fatalf("new snapshot=%+v %v", task, err)
	}
}

func TestDiskWriter_BusyAndCancelRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("five-second external lock acceptance")
	}
	r, path := diskRepository(t)
	createFixture(t, r, taskFixture("root"))
	child := startStorageChild(t, path, "hold")
	child.send(t)
	child.expect(t, "staged")
	callback := func(context.Context, ports.TaskWriter) error { return errors.New("callback must not run while locked") }
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	err := r.WithWrite(ctx, callback)
	elapsed := time.Since(start)
	canceledElapsed := elapsed
	if !errors.Is(err, context.DeadlineExceeded) || elapsed >= 5*time.Second {
		t.Fatalf("cancel=%v elapsed=%s", err, elapsed)
	}
	start = time.Now()
	err = r.WithWrite(context.Background(), callback)
	elapsed = time.Since(start)
	if !errors.Is(err, ports.ErrBusy) || elapsed < 4500*time.Millisecond || elapsed > 7*time.Second {
		t.Fatalf("busy=%v elapsed=%s", err, elapsed)
	}
	t.Logf("external writer: canceled in %s; busy bound %s", canceledElapsed, elapsed)
	child.send(t)
	child.expect(t, "committed")
	child.finish(t)
	if err := r.WithWrite(context.Background(), func(c context.Context, w ports.TaskWriter) error { return w.Create(c, taskFixture("recovery")) }); err != nil {
		t.Fatal(err)
	}
	if _, err := r.GetByID(context.Background(), "absent"); !errors.Is(err, core.ErrTaskNotFound) {
		t.Fatal(err)
	}
}
