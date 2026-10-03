package cli

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestReleaseLifecycle_ReplaceRemoveAndNewerSchema(t *testing.T) {
	if testing.Short() {
		t.Skip("actual executable lifecycle")
	}
	source := processBinary(t)
	root := processDirectory(t)
	path := filepath.Join(root, "space ü home", "tasks.db")
	copyBinary := filepath.Join(root, filepath.Base(source))
	binaryBytes, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	install := func() {
		t.Helper()
		if err := os.WriteFile(copyBinary, binaryBytes, 0755); err != nil {
			t.Fatal(err)
		}
	}
	install()
	p := processFixture{copyBinary, root, path}
	run := func(args ...string) []byte {
		t.Helper()
		code, out, stderr := p.run(args...)
		if code != 0 || stderr != "" {
			t.Fatalf("%v: %d %s", args, code, stderr)
		}
		return out
	}
	var task struct{ ID string }
	if err := json.Unmarshal(run("add", "Persistent release fixture", "--due", "tomorrow", "--json"), &task); err != nil {
		t.Fatal(err)
	}
	run("edit", task.ID, "--notes", "Keep history after replacement", "--json")
	beforeTasks := run("list", "--all", "--json")
	beforeEvents := run("history", task.ID, "--json")
	dataBefore := make(map[string][]byte)
	for _, suffix := range []string{"", "-wal", "-shm"} {
		data, err := os.ReadFile(path + suffix)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		dataBefore[suffix] = data
	}
	// Replacement/removal owns only this copy. The caller's binary is never removed.
	if err := os.Remove(copyBinary); err != nil {
		t.Fatal(err)
	}
	for _, suffix := range []string{"", "-wal", "-shm"} {
		want, existed := dataBefore[suffix]
		got, err := os.ReadFile(path + suffix)
		if !existed && os.IsNotExist(err) {
			continue
		}
		if !existed {
			t.Fatalf("unexpected sidecar %q: %v", suffix, err)
		}
		if err != nil || !bytes.Equal(got, want) {
			t.Fatalf("removal changed data %q: %v", suffix, err)
		}
	}
	install()
	if !bytes.Equal(run("list", "--all", "--json"), beforeTasks) || !bytes.Equal(run("history", task.ID, "--json"), beforeEvents) {
		t.Fatal("replacement recreated tasks/events")
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(context.Background(), "UPDATE schema_migrations SET version=999")
	closeErr := db.Close()
	if err != nil || closeErr != nil {
		t.Fatalf("newer fixture: %v %v", err, closeErr)
	}
	dataBefore = make(map[string][]byte)
	for _, suffix := range []string{"", "-wal", "-shm"} {
		data, err := os.ReadFile(path + suffix)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		dataBefore[suffix] = data
	}
	code, output, _ := p.run("list", "--all", "--json")
	if code != 1 || len(output) != 0 {
		t.Fatalf("newer schema accepted: %d %s", code, output)
	}
	for suffix, want := range dataBefore {
		got, err := os.ReadFile(path + suffix)
		if err != nil || !bytes.Equal(got, want) {
			t.Fatalf("newer refusal changed %q: %v", suffix, err)
		}
	}
	// A read-only WAL inspection may create SQLite's empty WAL/read cache.
	// Existing bytes above are immutable; a newly created WAL must contain no writes.
	if _, existed := dataBefore["-wal"]; !existed {
		wal, err := os.ReadFile(path + "-wal")
		if err != nil && !os.IsNotExist(err) || len(wal) != 0 {
			t.Fatalf("refusal wrote new WAL data: %v, %d bytes", err, len(wal))
		}
	}
	original, err := os.ReadFile(source)
	if err != nil || !bytes.Equal(original, binaryBytes) {
		t.Fatal("supplied executable changed")
	}
}
