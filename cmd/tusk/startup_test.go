package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCLIStartup_NoSyntaxRegistryInitialization(t *testing.T) {
	if testing.Short() {
		t.Skip("actual executable initialization")
	}
	if _, err := exec.LookPath("make"); err != nil {
		t.Skip("startup executable check requires canonical make build")
	}
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("startup test source location unavailable")
	}
	root := filepath.Dir(filepath.Dir(filepath.Dir(source)))
	if _, err := os.Stat(filepath.Join(root, "Makefile")); err != nil {
		t.Skip("startup build requires the source checkout")
	}
	binary := filepath.Join(t.TempDir(), "tusk")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	build := exec.Command("make", "build", "BUILD_OUTPUT="+binary)
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	for _, arg := range []string{"--help", "--version"} {
		cmd := exec.Command(binary, arg)
		cmd.Env = append(os.Environ(), "GODEBUG=inittrace=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%s: %v: %s", arg, err, out)
		}
		if !strings.Contains(string(out), "init runtime") {
			t.Fatal("initialization trace missing")
		}
		for _, pkg := range []string{"github.com/alecthomas/chroma/", "github.com/microcosm-cc/bluemonday", "github.com/gorilla/css"} {
			if strings.Contains(string(out), "init "+pkg) {
				t.Fatalf("%s eagerly initializes rendering registries (%s) before CLI routing", arg, pkg)
			}
		}
	}
}
