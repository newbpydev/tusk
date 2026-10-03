package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestReleaseSelection_MeasurementUsesSuppliedBinary(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "accepted ü artifact")
	if err := os.WriteFile(binary, []byte("accepted immutable bytes"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TUSK_RELEASE_BINARY", binary)
	t.Setenv("PATH", t.TempDir())
	got, err := measurementBinaryPath()
	if err != nil || got != binary {
		t.Fatalf("selected %q, %v; want supplied %q", got, err, binary)
	}
	data, err := os.ReadFile(binary)
	if err != nil || string(data) != "accepted immutable bytes" {
		t.Fatalf("supplied artifact changed: %v", err)
	}
}

func TestReleaseSelection_RejectMissingOrDirectory(t *testing.T) {
	base := t.TempDir()
	t.Chdir(base)
	for _, selected := range []string{"missing-tusk", base} {
		t.Setenv("TUSK_RELEASE_BINARY", selected)
		if _, err := measurementBinaryPath(); err == nil {
			t.Fatalf("non-file supplied executable accepted: %q", selected)
		}
	}
}

func measurementBinaryPath() (string, error) {
	binary := os.Getenv("TUSK_RELEASE_BINARY")
	if binary == "" {
		binary = "../../bin/tusk"
	}
	binary, err := filepath.Abs(binary)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(binary)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("supplied executable must be a regular file: %s", binary)
	}
	return binary, nil
}

func releaseDirectory(t *testing.T) string {
	t.Helper()
	if root := os.Getenv("TUSK_RELEASE_FIXTURE_DIR"); root != "" {
		dir, err := os.MkdirTemp(root, "terminal-")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "case.txt"), []byte(t.Name()+"\n"), 0600); err != nil {
			t.Fatal(err)
		}
		return dir
	}
	return t.TempDir()
}
