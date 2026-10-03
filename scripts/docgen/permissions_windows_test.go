package main

import (
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/sys/windows"
)

// Go opens read handles with backup semantics. Limit privileges on this test's
// OS thread so an elevated runner cannot bypass its owned-file deny ACLs.
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
	if err := os.WriteFile(file, []byte("owned output"), 0600); err != nil {
		t.Fatal(err)
	}
	denyWindowsAccess(t, file, "(R)")
	defer withoutWindowsPrivileges(t)()
	if _, err := os.ReadFile(file); err == nil {
		t.Fatal("read-denial ACL fixture ineffective")
	}
	if _, err := readGroup(root); err == nil {
		t.Fatal("unreadable output accepted")
	}
}

func TestDocgen_WindowsUnwritableOutput(t *testing.T) {
	root := t.TempDir()
	denyWindowsAccess(t, root, "(AD)")
	defer withoutWindowsPrivileges(t)()
	if _, err := os.MkdirTemp(root, "probe-"); err == nil {
		t.Fatal("directory-creation deny ACL fixture ineffective")
	}
	if err := execute(root, false, os.Rename); err == nil {
		t.Fatal("unwritable output accepted")
	}
}
