package cli

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestReleaseSelection_PreservesSuppliedBinary(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "accepted ü artifact")
	if err := os.WriteFile(binary, []byte("accepted immutable bytes"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TUSK_RELEASE_BINARY", binary)
	t.Setenv("PATH", t.TempDir()) // A selector must not invoke a build tool.
	if got := processBinary(t); got != binary {
		t.Fatalf("selected %q; want supplied %q", got, binary)
	}
	data, err := os.ReadFile(binary)
	if err != nil || string(data) != "accepted immutable bytes" {
		t.Fatalf("supplied artifact changed: %v", err)
	}
}

// processBinary uses caller-selected artifacts only in the explicit release lane.
// Ordinary canonical tests retain their owned source build.
func processBinary(t *testing.T) string {
	t.Helper()
	if binary := os.Getenv("TUSK_RELEASE_BINARY"); binary != "" {
		absolute, err := filepath.Abs(binary)
		if err != nil {
			t.Fatal(err)
		}
		st, err := os.Stat(absolute)
		if err != nil || !st.Mode().IsRegular() {
			t.Fatalf("invalid supplied artifact: %v", err)
		}
		return absolute
	}
	binary := filepath.Join(t.TempDir(), "tusk")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	build := exec.CommandContext(ctx, "make", "build", "BUILD_OUTPUT="+binary)
	build.Dir = "../.."
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v: %s", err, out)
	}
	return binary
}

func processDirectory(t *testing.T) string {
	t.Helper()
	if root := os.Getenv("TUSK_RELEASE_FIXTURE_DIR"); root != "" {
		dir, err := os.MkdirTemp(root, "process-")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "case.txt"), []byte(t.Name()+"\n"), 0600); err != nil {
			t.Fatal(err)
		}
		return dir // Native QA retains owned fixtures on both success and failure.
	}
	return t.TempDir()
}
