package cli

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// Rehearse the public examples and the closed-file backup recipe on owned data.
// Every process has exited before copy; no owner is alive at copy.
func TestDocsExamples_QuickStartAndClosedBackup(t *testing.T) {
	if testing.Short() {
		t.Skip("actual executable and closed backup")
	}
	root := t.TempDir()
	binary := filepath.Join(root, "tusk")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	build := exec.Command("make", "build", "BUILD_OUTPUT="+binary)
	build.Dir = "../.."
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v: %s", err, output)
	}
	path := filepath.Join(root, "original", "tusk.db")
	run := func(args ...string) []byte {
		t.Helper()
		fixture := processFixture{binary: binary, dir: root, path: path}
		code, output, errs := fixture.run(args...)
		if code != 0 || errs != "" {
			t.Fatalf("%v: exit %d, %s", args, code, errs)
		}
		return output
	}
	run("--version")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("version opened storage: %v", err)
	}
	var parent, child struct{ ID string }
	if err := json.Unmarshal(run("add", "Plan the week", "--priority", "high", "--due", "tomorrow", "--json"), &parent); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(run("add", "Review the plan", "--parent", parent.ID, "--due", "+1w", "--json"), &child); err != nil {
		t.Fatal(err)
	}
	run("edit", child.ID, "--status", "in-progress", "--json")
	run("list")
	run("tree")
	before := map[string][]byte{
		"list":    run("list", "--all", "--json"),
		"tree":    run("tree", "--json"),
		parent.ID: run("history", parent.ID, "--json"),
		child.ID:  run("history", child.ID, "--json"),
	}
	for name, data := range before {
		if !json.Valid(data) {
			t.Fatalf("unclean JSON for %s: %q", name, data)
		}
	}
	backup := filepath.Join(root, "backup")
	restored := filepath.Join(root, "restored")
	for _, directory := range []string{backup, restored} {
		if err := os.Mkdir(directory, 0700); err != nil {
			t.Fatal(err)
		}
	}
	for _, suffix := range []string{"", "-wal", "-shm"} {
		data, err := os.ReadFile(path + suffix)
		if suffix != "" && os.IsNotExist(err) {
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		for _, directory := range []string{backup, restored} {
			if err := os.WriteFile(filepath.Join(directory, "tusk.db"+suffix), data, 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	originalPath := path
	originalBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	path = filepath.Join(restored, "tusk.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	var integrity string
	err = db.QueryRowContext(t.Context(), "PRAGMA integrity_check").Scan(&integrity)
	closeErr := db.Close()
	if err != nil || closeErr != nil || integrity != "ok" {
		t.Fatalf("restored integrity: %q, %v, %v", integrity, err, closeErr)
	}
	for name, want := range before {
		var got []byte
		switch name {
		case "list":
			got = run("list", "--all", "--json")
		case "tree":
			got = run("tree", "--json")
		default:
			got = run("history", name, "--json")
		}
		if !bytes.Equal(want, got) {
			t.Fatalf("restored tasks/events changed: %s", name)
		}
	}
	unchanged, err := os.ReadFile(originalPath)
	if err != nil || !bytes.Equal(originalBytes, unchanged) {
		t.Fatalf("restore rehearsal modified original: %v", err)
	}
}
