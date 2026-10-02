package main

import (
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"golang.org/x/sys/windows"
)

func withoutWindowsPrivileges(t *testing.T) func() {
	t.Helper()
	runtime.LockOSThread()
	if err := windows.ImpersonateSelf(windows.SecurityImpersonation); err != nil {
		runtime.UnlockOSThread()
		t.Fatal(err)
	}
	restore := func() {
		if err := windows.RevertToSelf(); err != nil {
			t.Fatal(err)
		}
		runtime.UnlockOSThread()
	}
	var token windows.Token
	if err := windows.OpenThreadToken(windows.CurrentThread(), windows.TOKEN_ADJUST_PRIVILEGES|windows.TOKEN_QUERY, true, &token); err != nil {
		restore()
		t.Fatal(err)
	}
	err := windows.AdjustTokenPrivileges(token, true, nil, 0, nil, nil)
	closeErr := token.Close()
	if err != nil || closeErr != nil {
		restore()
		t.Fatalf("limit owned test thread: %v; close: %v", err, closeErr)
	}
	return restore
}

func TestFilesystem_WindowsReadFailures(t *testing.T) {
	root := t.TempDir()
	payload, err := payloadFiles("../..")
	if err != nil {
		t.Fatal(err)
	}
	for name, input := range payload {
		if strings.HasPrefix(name, "completions/") || strings.HasPrefix(name, "man/") {
			name = "docs/" + name
		}
		filename := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filename, input.Data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := payloadFiles(root); err != nil {
		t.Fatalf("complete readable payload control: %v", err)
	}
	file := filepath.Join(root, "README.md")
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
	defer withoutWindowsPrivileges(t)()
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
