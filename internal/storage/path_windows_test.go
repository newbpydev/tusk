package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestOpen_WindowsPathsAndReparse(t *testing.T) {
	base := t.TempDir()
	path := filepath.Join(base, "Unicode ü spaces.db")
	r, err := Open(context.Background(), Options{Path: path})
	if err != nil {
		t.Fatal(err)
	}
	r.Close()
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(context.Background(), Options{Path: filepath.Join(base, "NUL")}); err == nil {
		t.Fatal("reserved filename accepted")
	}
	link := filepath.Join(base, "link.db")
	if err := os.Symlink(path, link); err != nil {
		t.Skipf("Windows symlink privilege needed: %v", err)
	}
	if _, err := Open(context.Background(), Options{Path: link}); err == nil {
		t.Fatal("reparse target accepted")
	}
}
