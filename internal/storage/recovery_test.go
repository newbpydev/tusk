package storage

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/newbpydev/tusk/internal/ports"
)

func assertDiskIntegrity(t *testing.T, r *Repository) {
	t.Helper()
	var result string
	if err := r.reader.QueryRow("PRAGMA integrity_check").Scan(&result); err != nil || result != "ok" {
		t.Fatalf("integrity=%s %v", result, err)
	}
	rows, err := r.reader.Query("PRAGMA foreign_key_check")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	if rows.Next() || rows.Err() != nil {
		t.Fatalf("foreign key violation: %v", rows.Err())
	}
}
func assertRecoveryState(t *testing.T, path string, changed bool) {
	t.Helper()
	r, err := Open(context.Background(), Options{Path: path})
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	task, err := r.GetByID(context.Background(), "root")
	if err != nil {
		t.Fatal(err)
	}
	events, err := r.ListEvents(context.Background(), "root")
	if err != nil {
		t.Fatal(err)
	}
	title, count := "title", 0
	if changed {
		title, count = "committed change", 1
	}
	if task.Title != title || len(events) != count {
		t.Fatalf("partial recovery: title=%s events=%d", task.Title, len(events))
	}
	assertDiskIntegrity(t, r)
}
func TestRecovery_KilledWriterAtomic(t *testing.T) {
	if testing.Short() {
		t.Skip("process death acceptance")
	}
	for _, mode := range []string{"before", "after"} {
		t.Run(mode, func(t *testing.T) {
			r, path := diskRepository(t)
			createFixture(t, r, taskFixture("root"))
			if err := r.Close(); err != nil {
				t.Fatal(err)
			}
			child := startStorageChild(t, path, mode)
			child.send(t)
			if mode == "before" {
				child.expect(t, "staged")
			} else {
				child.expect(t, "committed")
			}
			if err := child.cmd.Process.Kill(); err != nil {
				t.Fatal(err)
			}
			if err := child.cmd.Wait(); err == nil {
				t.Fatal("killed process succeeded")
			}
			assertRecoveryState(t, path, mode == "after")
		})
	}
}
func TestRecovery_AtomicStateAfterFailure(t *testing.T) {
	for _, stage := range []string{"statement", "event", "commit-before", "commit-after", "rollback"} {
		t.Run(stage, func(t *testing.T) {
			r, path := diskRepository(t)
			createFixture(t, r, taskFixture("root"))
			if strings.HasPrefix(stage, "commit") || stage == "rollback" {
				installTransactionFault(t, r, path, stage, false, nil, nil, nil)
			}
			calls := 0
			err := r.WithWrite(context.Background(), func(c context.Context, w ports.TaskWriter) error {
				calls++
				task, err := w.GetByID(c, "root")
				if err != nil {
					return err
				}
				task.Title = "committed change"
				if err := w.Update(c, task); err != nil {
					return err
				}
				if stage == "statement" {
					_ = w.Create(c, taskFixture("root"))
					return nil
				}
				event := ports.TaskEvent{TaskID: "root", Kind: ports.EventMetadata, ChangedFields: []string{"title"}, OccurredAt: task.UpdatedAt}
				if stage == "event" {
					event.TaskID = "missing"
				}
				_, err = w.AppendEvent(c, event)
				if stage == "event" {
					return nil
				}
				if err != nil {
					return err
				}
				if stage == "rollback" {
					return errors.New("callback failure")
				}
				return nil
			})
			if err == nil || calls != 1 {
				t.Fatalf("failure=%v callbacks=%d", err, calls)
			}
			if err := r.Close(); err != nil {
				t.Fatal(err)
			}
			assertRecoveryState(t, path, stage == "commit-after")
		})
	}
}

func TestDiskLifecycle_IntegrityAndCleanup(t *testing.T) {
	if testing.Short() {
		t.Skip("automatic checkpoint acceptance")
	}
	r, path := diskRepository(t)
	createFixture(t, r, taskFixture("root"))
	read, err := r.reader.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer read.Rollback()
	var title string
	if err := read.QueryRow("SELECT title FROM tasks").Scan(&title); err != nil {
		t.Fatal(err)
	}
	write := func(i int) {
		t.Helper()
		err := r.WithWrite(context.Background(), func(c context.Context, w ports.TaskWriter) error {
			task, err := w.GetByID(c, "root")
			if err != nil {
				return err
			}
			task.Description = strings.Repeat(string(rune('a'+i%26)), 32768)
			return w.Update(c, task)
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	header := func() (uint32, int64) {
		t.Helper()
		f, err := os.Open(path + "-wal")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		var h [32]byte
		if _, err := f.ReadAt(h[:], 0); err != nil {
			t.Fatal(err)
		}
		info, err := f.Stat()
		if err != nil {
			t.Fatal(err)
		}
		return binary.BigEndian.Uint32(h[12:16]), info.Size()
	}
	for i := 0; i < 160; i++ {
		write(i)
	}
	heldSequence, heldSize := header()
	if heldSize < 4*1024*1024 {
		t.Fatalf("fixture did not exceed automatic threshold: %d", heldSize)
	}
	if err := read.Rollback(); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 320; i++ {
		write(i)
	}
	recycledSequence, recycledSize := header()
	if recycledSequence <= heldSequence || recycledSize > heldSize*2 {
		t.Fatalf("automatic WAL recycling absent: sequence %d -> %d size %d -> %d", heldSequence, recycledSequence, heldSize, recycledSize)
	}
	t.Logf("automatic WAL restart sequence %d -> %d; size %d -> %d (no explicit checkpoint)", heldSequence, recycledSequence, heldSize, recycledSize)
	assertDiskIntegrity(t, r)
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		next, err := Open(context.Background(), Options{Path: path})
		if err != nil {
			t.Fatal(err)
		}
		assertDiskIntegrity(t, next)
		if err := next.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
}

func TestStorageRunbookReplay_OfflineBackup(t *testing.T) {
	r, path := diskRepository(t)
	createFixture(t, r, taskFixture("root"))
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	for _, suffix := range []string{"-wal", "-shm"} {
		if _, err := os.Stat(path + suffix); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("normal close left sidecar %s: %v", suffix, err)
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	backup := path + ".backup"
	if err := os.WriteFile(backup, data, 0600); err != nil {
		t.Fatal(err)
	}
	assertRecoveryState(t, backup, false)
	original, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(data, original) {
		t.Fatal("backup replay changed original")
	}
}
