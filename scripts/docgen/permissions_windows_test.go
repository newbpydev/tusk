package main

import (
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"testing"
)

func denyWindowsAccess(t *testing.T, path, access string) {
	t.Helper()
	owner, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	sid := "*" + owner.Uid
	command := func(args ...string) {
		t.Helper()
		if output, err := exec.Command("icacls", args...).CombinedOutput(); err != nil {
			t.Fatalf("native ACL fixture: %v: %s", err, output)
		}
	}
	t.Cleanup(func() { command(path, "/remove:d", sid) })
	command(path, "/deny", sid+":"+access)
}

// Windows ignores Unix mode bits; use actual owned-file ACLs instead.
func TestDocgen_WindowsUnreadableOutput(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "unreadable")
	if err := os.WriteFile(file, nil, 0600); err != nil {
		t.Fatal(err)
	}
	denyWindowsAccess(t, file, "(R)")
	if _, err := os.ReadFile(file); err == nil {
		t.Fatal("read-denial ACL fixture ineffective")
	}
	if _, err := readGroup(root); err == nil {
		t.Fatal("unreadable output accepted")
	}
}

func TestDocgen_WindowsUnwritableOutput(t *testing.T) {
	root := t.TempDir()
	denyWindowsAccess(t, root, "(W)")
	if err := os.WriteFile(filepath.Join(root, "probe"), nil, 0600); err == nil {
		t.Fatal("write-denial ACL fixture ineffective")
	}
	if err := execute(root, false, os.Rename); err == nil {
		t.Fatal("unwritable output accepted")
	}
}
