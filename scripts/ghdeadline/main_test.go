package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// The owned child is a real process on every native OS, without a shell or
// network. No test ever terminates an unrelated process.
func TestMain(m *testing.M) {
	if os.Getenv("TUSK_GH_DEADLINE_CHILD") == "1" {
		switch os.Args[1] {
		case "echo":
			_ = json.NewEncoder(os.Stdout).Encode(os.Args[2:])
		case "stdin":
			_, _ = io.Copy(os.Stdout, os.Stdin)
		case "exit":
			fmt.Fprintln(os.Stderr, "child diagnostic")
			os.Exit(47)
		case "stall":
			time.Sleep(time.Hour)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func fakeGH(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	name := "gh"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), data, 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	t.Setenv("TUSK_GH_DEADLINE_CHILD", "1")
}

func TestGHDeadlineStreamsAndExit(t *testing.T) {
	fakeGH(t)
	var out, diagnostic bytes.Buffer
	args := []string{"echo", "spaces and ü", "$(touch marker)", "--input", "a#b"}
	if code := run(context.Background(), "", args, nil, &out, &diagnostic); code != 0 {
		t.Fatalf("echo: %d %s", code, &diagnostic)
	}
	var got []string
	if err := json.Unmarshal(out.Bytes(), &got); err != nil || strings.Join(got, "\x00") != strings.Join(args[1:], "\x00") {
		t.Fatalf("arguments changed: %q, %v", got, err)
	}
	out.Reset()
	if code := run(context.Background(), "5m", []string{"stdin"}, strings.NewReader("raw body"), &out, &diagnostic); code != 0 || out.String() != "raw body" {
		t.Fatalf("stdin: %d %q", code, out.String())
	}
	if code := run(context.Background(), "5s", []string{"exit"}, nil, &out, &diagnostic); code != 47 || !strings.Contains(diagnostic.String(), "child diagnostic") {
		t.Fatalf("exit/stderr: %d %s", code, &diagnostic)
	}
}

func TestGHDeadlineRejectsUnboundedValues(t *testing.T) {
	for _, timeout := range []string{"bad", "0", "-1s", "5m1s"} {
		var diagnostic bytes.Buffer
		if code := run(context.Background(), timeout, nil, nil, io.Discard, &diagnostic); code != 2 || diagnostic.Len() == 0 {
			t.Fatalf("timeout %q: %d %s", timeout, code, &diagnostic)
		}
	}
}

func TestGHDeadlineKillsAndJoinsStalledChild(t *testing.T) {
	fakeGH(t)
	var diagnostic bytes.Buffer
	start := time.Now()
	if code := run(context.Background(), "100ms", []string{"stall"}, nil, io.Discard, &diagnostic); code != 124 || !strings.Contains(diagnostic.String(), "deadline") {
		t.Fatalf("stalled child: %d %s", code, &diagnostic)
	}
	if time.Since(start) > 5*time.Second {
		t.Fatal("child was not joined within the deadline and pipe grace period")
	}
}

func TestGHDeadlineCancellationAndMissingCommand(t *testing.T) {
	fakeGH(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var diagnostic bytes.Buffer
	if code := run(ctx, "5s", []string{"stall"}, nil, io.Discard, &diagnostic); code != 130 {
		t.Fatalf("cancel: %d %s", code, &diagnostic)
	}
	t.Setenv("PATH", t.TempDir())
	if code := run(context.Background(), "5s", nil, nil, io.Discard, &diagnostic); code != 127 {
		t.Fatalf("missing gh: %d %s", code, &diagnostic)
	}
}

func TestGHDeadlineCommandBoundary(t *testing.T) {
	fakeGH(t)
	t.Setenv("GH_REQUEST_TIMEOUT", "5s")
	if code := commandMain([]string{"echo"}); code != 0 {
		t.Fatalf("command boundary: %d", code)
	}
}
