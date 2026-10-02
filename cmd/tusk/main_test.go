package main

import (
	"bytes"
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestVersion_ImmutableDevelopmentFile(t *testing.T) {
	if Version != "dev" {
		t.Fatalf("ordinary source version must be clearly unreleased: %q", Version)
	}
	file, err := parser.ParseFile(token.NewFileSet(), "version.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(file.Decls) != 1 {
		t.Fatal("version file must contain only the immutable version constant")
	}
	decl, ok := file.Decls[0].(*ast.GenDecl)
	if !ok || decl.Tok != token.CONST || len(decl.Specs) != 1 {
		t.Fatal("mutable or unrelated release metadata")
	}
	value, ok := decl.Specs[0].(*ast.ValueSpec)
	if !ok || len(value.Names) != 1 || value.Names[0].Name != "Version" || len(value.Values) != 1 {
		t.Fatal("expected exactly const Version")
	}
	literal, ok := value.Values[0].(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING || literal.Value != `"dev"` {
		t.Fatal("version is not an immutable development literal")
	}
}

func TestRunVersion(t *testing.T) {
	var buf bytes.Buffer
	exitCode := run([]string{"tusk", "--version"}, &buf)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	output := buf.String()
	if !strings.Contains(output, Version) {
		t.Fatalf("expected version %q in output, got %q", Version, output)
	}
}

func TestRunHelp(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "flag -h", args: []string{"tusk", "-h"}},
		{name: "flag --help", args: []string{"tusk", "--help"}},
		{name: "subcommand help", args: []string{"tusk", "help"}},
		{name: "no arguments", args: []string{"tusk"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			exitCode := run(tc.args, &buf)
			if exitCode != 0 {
				t.Fatalf("expected exit code 0, got %d", exitCode)
			}

			output := buf.String()
			if !strings.Contains(output, "Tusk - Zero-friction terminal task management system") {
				t.Fatalf("expected help banner in output, got %q", output)
			}
		})
	}
}

func TestMainExecution(t *testing.T) {
	if os.Getenv("TEST_MAIN_EXEC") == "1" {
		os.Args = []string{"tusk"}
		main()
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=TestMainExecution")
	cmd.Env = append(os.Environ(), "TEST_MAIN_EXEC=1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("process execution failed: %v, output: %s", err, string(output))
	}
	if !strings.Contains(string(output), "Tusk - Zero-friction terminal task management system") {
		t.Errorf("expected help banner in main output, got %q", string(output))
	}
}

func TestRoot_UnknownCommand(t *testing.T) {
	var out bytes.Buffer
	if got := run([]string{"tusk", "nonsense"}, &out); got != 2 {
		t.Fatalf("unknown command exit = %d, want 2", got)
	}
	if out.Len() != 0 {
		t.Fatalf("unexpected stdout: %q", out.String())
	}
}
