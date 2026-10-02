package main

import (
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"testing"
)

func TestFilesystem_WindowsReadFailures(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "README.md")
	if err := os.WriteFile(file, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	command(t, root, nil, "git", "init", "--quiet")
	command(t, root, nil, "git", "add", "README.md")
	owner, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	sid := "*" + owner.Uid
	acl := func(args ...string) {
		t.Helper()
		if out, err := exec.Command("icacls", args...).CombinedOutput(); err != nil {
			t.Fatalf("native ACL: %v: %s", err, out)
		}
	}
	t.Cleanup(func() { acl(file, "/remove:d", sid) })
	acl(file, "/deny", sid+":(R)")
	if _, err := os.ReadFile(file); err == nil {
		t.Fatal("read-denial ACL ineffective")
	}
	if _, err := payloadFiles(root); err == nil {
		t.Fatal("unreadable Windows payload")
	}
	if _, err := sourceFiles(root, "p/"); err == nil {
		t.Fatal("unreadable Windows source")
	}
}
