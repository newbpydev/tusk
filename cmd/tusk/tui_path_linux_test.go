package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/newbpydev/tusk/internal/cli"
)

// Linux reports an unavailable cwd after unlinking it. Darwin may still return
// its former path; Windows refuses the unlink. The injected resolver error is
// tested independently on every platform in internal/storage.
func TestTUIFactory_UnavailableCWD(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	t.Setenv("TUSK_DB_PATH", "relative.db")
	if err := os.Remove(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Getwd(); err == nil {
		t.Fatal("fixture did not make cwd unavailable")
	}
	_, closeOwner, err := tuiFactory(cli.Config{Location: time.UTC})(context.Background())
	if closeOwner != nil {
		_ = closeOwner()
	}
	if err == nil {
		t.Fatal("unavailable cwd accepted")
	}
}
