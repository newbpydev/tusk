//go:build !windows

package storage

import (
	"context"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestOpen_POSIXPermissions(t *testing.T) {
	base := t.TempDir()
	path := filepath.Join(base, "private", "tasks.db")
	r, err := Open(context.Background(), Options{Path: path})
	if err != nil {
		t.Fatal(err)
	}
	r.Close()
	for p, max := range map[string]os.FileMode{filepath.Dir(path): 0700, path: 0600} {
		info, err := os.Stat(p)
		if err != nil || info.Mode().Perm() & ^max != 0 {
			t.Fatalf("permissions for %s: %v", p, err)
		}
	}
	if err := os.Chmod(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(base, "alias")
	if err := os.Symlink(filepath.Dir(path), link); err != nil {
		t.Fatal(err)
	}
	r, err = Open(context.Background(), Options{Path: filepath.Join(link, "tasks.db")})
	if err != nil {
		t.Fatal(err)
	}
	r.Close()
	for p, want := range map[string]os.FileMode{filepath.Dir(path): 0755, path: 0644} {
		info, err := os.Stat(p)
		if err != nil || info.Mode().Perm() != want {
			t.Fatalf("existing permissions changed: %s", p)
		}
	}
	fifo := filepath.Join(base, "fifo")
	if err := syscall.Mkfifo(fifo, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(context.Background(), Options{Path: fifo}); err == nil {
		t.Fatal("FIFO accepted")
	}
}
