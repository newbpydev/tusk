package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/tusk/internal/ports"
	"github.com/newbpydev/tusk/internal/storage"
)

func TestCLIConditions_StartFailure(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "not-executable")
	if err := os.WriteFile(binary, []byte("not an executable"), 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestCLIConditions$")
	cmd.Env = append(os.Environ(), "TUSK_CLI_CONDITIONS=1", "TUSK_CLI_TIMING_DIAGNOSTIC=0", "TUSK_CLI_BINARY="+binary)
	out, err := cmd.CombinedOutput()
	if err == nil || strings.Contains(string(out), "panic:") || !strings.Contains(string(out), "first-use: fork/exec") {
		t.Fatalf("expected controlled launch failure: %v\n%s", err, out)
	}
}

type slowOutput struct{ bytes.Buffer }

func (w *slowOutput) Write(p []byte) (int, error) {
	time.Sleep(5 * time.Millisecond)
	return w.Buffer.Write(p)
}

// Separate observations, never substituted for the 100 reference samples.
func TestCLIConditions(t *testing.T) {
	if os.Getenv("TUSK_CLI_CONDITIONS") != "1" {
		t.Skip("make bench-cli-conditions")
	}
	binary, err := filepath.Abs(os.Getenv("TUSK_CLI_BINARY"))
	if err != nil {
		t.Fatal(err)
	}
	m, err := manifest(binary)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if os.Getenv("TUSK_CLI_TIMING_DIAGNOSTIC") == "1" {
		timingDiagnostic(t, binary, dir, m)
		return
	}
	type observation struct {
		Name        string
		DurationNS  int64
		Bytes, Exit int
		Error       string
	}
	var observations []observation
	runCase := func(name, path string, args []string, slow bool, want int) {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, binary, args...)
		cmd.Env = processEnv(dir, path)
		var out slowOutput
		var stderr bytes.Buffer
		if slow {
			cmd.Stdout = &out
		} else {
			cmd.Stdout = &out.Buffer
		}
		cmd.Stderr = &stderr
		start := time.Now()
		err := cmd.Run()
		elapsed := time.Since(start)
		code := cmd.ProcessState.ExitCode()
		o := observation{Name: name, DurationNS: elapsed.Nanoseconds(), Bytes: out.Len(), Exit: code, Error: stderr.String()}
		observations = append(observations, o)
		if code != want || ctx.Err() != nil {
			t.Fatalf("%s: %v %s", name, err, stderr.String())
		}
		if want == 0 && !json.Valid(out.Bytes()) {
			t.Fatalf("%s invalid JSON", name)
		}
	}
	runCase("first-use", filepath.Join(dir, "absent.db"), []string{"stats", "--json"}, false, 0)
	large := filepath.Join(dir, "large.db")
	if _, err := seed(large, 10000, 128); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"list", "tree", "stats"} {
		args := []string{name, "--json"}
		if name == "list" {
			args = append(args, "--all")
		}
		runCase("10000/"+name, large, args, false, 0)
	}
	notes := filepath.Join(dir, "notes.db")
	if _, err := seed(notes, 1, 1<<20); err != nil {
		t.Fatal(err)
	}
	runCase("1MiB-notes", notes, []string{"list", "--all", "--json"}, false, 0)
	runCase("throttled-output", large, []string{"list", "--all", "--json"}, true, 0)
	r, err := storage.Open(context.Background(), storage.Options{Path: notes})
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	ready, release, done := make(chan struct{}), make(chan struct{}), make(chan error, 1)
	go func() {
		done <- r.WithWrite(context.Background(), func(context.Context, ports.TaskWriter) error { close(ready); <-release; return nil })
	}()
	<-ready
	func() {
		defer func() {
			close(release)
			if err := <-done; err != nil {
				t.Error(err)
			}
		}()
		runCase("held-writer", notes, []string{"add", "blocked", "--json"}, false, 1)
	}()
	data, err := json.MarshalIndent(struct {
		Manifest     map[string]string
		Observations []observation
	}{m, observations}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(os.Getenv("TUSK_CLI_CONDITIONS_OUTPUT"), append(data, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(os.Stdout, bytes.NewReader(data))
	t.Log("bounded observations completed; latency violations remain visible")
}
