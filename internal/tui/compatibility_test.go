package tui

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
)

func TestTUI_DependencyCheckoutNeedsNoLFS(t *testing.T) {
	for _, module := range []string{"bubbletea", "glamour"} {
		t.Run(module, func(t *testing.T) {
			dir := t.TempDir()
			checkout := t.TempDir()
			if err := os.CopyFS(dir, os.DirFS("../../third_party/"+module)); err != nil {
				t.Fatal(err)
			}
			for _, args := range [][]string{
				{"init", "--quiet"},
				{"-c", "filter.lfs.process=", "-c", "filter.lfs.clean=cat", "add", "."},
				// A required failing smudge detects any dependency on the LFS
				// service without installing Git LFS or contacting a network.
				{"-c", "filter.lfs.process=", "-c", "filter.lfs.clean=cat", "-c", "filter.lfs.smudge=false", "-c", "filter.lfs.required=true", "checkout-index", "--prefix=" + checkout + string(os.PathSeparator), "--all"},
			} {
				// Each Git operation gets its own resource budget; staging the
				// gallery must not consume the checkout operation's allowance.
				ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
				started := time.Now()
				cmd := exec.CommandContext(ctx, "git", args...)
				cmd.Dir = dir
				output, err := cmd.CombinedOutput()
				contextErr := ctx.Err()
				cancel()
				if err != nil {
					t.Fatalf("dependency checkout must work without Git LFS (%v): %v (context: %v, elapsed: %s)\n%s", args, err, contextErr, time.Since(started), output)
				}
			}
			if err := filepath.WalkDir(checkout, func(path string, entry fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if entry.IsDir() {
					return nil
				}
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				if bytes.HasPrefix(data, []byte("version https://git-lfs.github.com/spec/v1\n")) {
					return fmt.Errorf("dependency contains an unresolved Git LFS pointer: %s", path)
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestTUI_DependencyPatchIntegrity(t *testing.T) {
	for _, tc := range []struct {
		module, version string
		files           []string
	}{
		{"bubbletea", "v1.3.10", []string{"tea_init.go"}},
		{"glamour", "v0.9.1", []string{
			"ansi/codeblock.go", "ansi/context.go", ".gitattributes",
			"styles/gallery/ascii.png", "styles/gallery/auto.png",
			"styles/gallery/dark.png", "styles/gallery/dracula.png",
			"styles/gallery/light.png", "styles/gallery/notty.png",
			"styles/gallery/pink.png", "styles/gallery/tokyo-night.png",
		}},
	} {
		t.Run(tc.module, func(t *testing.T) {
			checkDependencyPatch(t, "../../third_party/"+tc.module, tc.version, tc.files)
		})
	}
}

func TestTUI_DependencyPatchIntegrityRejectsAdditionalSource(t *testing.T) {
	for _, module := range []string{"bubbletea", "glamour"} {
		t.Run(module, func(t *testing.T) {
			root := t.TempDir()
			cwd := filepath.Join(root, "internal", "tui")
			if err := os.MkdirAll(cwd, 0700); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"bubbletea", "glamour"} {
				if err := os.CopyFS(filepath.Join(root, "third_party", name), os.DirFS("../../third_party/"+name)); err != nil {
					t.Fatal(err)
				}
			}
			// A new Go file participates in the replaced module's next build,
			// even when every file already in the manifest has its pinned hash.
			path := filepath.Join(root, "third_party", module, "unrecorded.go")
			if err := os.WriteFile(path, []byte("package "+map[string]string{"bubbletea": "tea", "glamour": "glamour"}[module]+"\nfunc init() {}\n"), 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestTUI_DependencyPatchIntegrity$")
			cmd.Dir = cwd
			output, err := cmd.CombinedOutput()
			if ctx.Err() != nil {
				t.Fatal("integrity child timed out", ctx.Err())
			}
			if err == nil {
				t.Fatalf("dependency integrity guard passed with added %s source", module)
			}
			if !strings.Contains(string(output), "unrecorded dependency source: unrecorded.go") {
				t.Fatalf("guard failed for another reason: %v\n%s", err, output)
			}
		})
	}
}

func checkDependencyPatch(t *testing.T, dir, version string, files []string) {
	t.Helper()
	data, err := os.ReadFile(dir + "/TUSK-PATCH.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Version  string
		Upstream map[string]string `json:"upstream_sha256"`
		Patches  map[string]string `json:"patches"`
	}
	if err = json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Version != version || len(manifest.Patches) != len(files) {
		t.Fatal("unexpected dependency patch")
	}
	for _, name := range files {
		if _, ok := manifest.Patches[name]; !ok {
			t.Fatalf("missing declared dependency patch: %s", name)
		}
	}
	if err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		name, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		name = filepath.ToSlash(name)
		if !entry.Type().IsRegular() {
			return fmt.Errorf("nonregular dependency source: %s", name)
		}
		if _, ok := manifest.Upstream[name]; !ok && name != "TUSK-PATCH.json" {
			return fmt.Errorf("unrecorded dependency source: %s", name)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	for name, want := range manifest.Upstream {
		if patched, ok := manifest.Patches[name]; ok {
			allowed := false
			for _, file := range files {
				allowed = allowed || name == file
			}
			if !allowed {
				t.Fatal("unapproved patch")
			}
			want = patched
		}
		data, err = os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != want {
			t.Errorf("dependency source drift: %s", name)
		}
	}
}

func TestTUI_Architecture(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(token.NewFileSet(), name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range f.Imports {
			if strings.Contains(imp.Path.Value, "internal/storage") || strings.Contains(imp.Path.Value, "internal/cli") {
				t.Errorf("forbidden import %s", imp.Path.Value)
			}
		}
		for _, decl := range f.Decls {
			if d, ok := decl.(*ast.GenDecl); ok && d.Tok == token.VAR {
				t.Errorf("mutable package state in %s", name)
			}
			if d, ok := decl.(*ast.FuncDecl); ok && d.Name.Name == "View" {
				if len(d.Body.List) != 1 {
					t.Fatal("View must only return its frame")
				}
				r, ok := d.Body.List[0].(*ast.ReturnStmt)
				if !ok || len(r.Results) != 1 {
					t.Fatal("View must only return its frame")
				}
				s, ok := r.Results[0].(*ast.SelectorExpr)
				if !ok || s.Sel.Name != "frame" {
					t.Fatal("View must return prepared frame")
				}
			}
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if _, ok := n.(*ast.GoStmt); ok {
				t.Errorf("model goroutine in %s", name)
			}
			return true
		})
	}
}

func TestTUI_V1Compatibility(t *testing.T) {
	var _ tea.Model = (*Model)(nil)
	var _ tea.Msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	if _, err := glamour.NewTermRenderer(glamour.WithStandardStyle("notty"), glamour.WithWordWrap(40)); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("../../go.mod")
	if err != nil {
		t.Fatal(err)
	}
	for _, pin := range []string{"bubbletea v1.3.10", "bubbles v0.21.0", "lipgloss v1.1.0", "glamour v0.9.1", "x/ansi v0.10.1", "modernc.org/sqlite v1.58.0", "modernc.org/libc v1.75.6", "go 1.25.0"} {
		if !strings.Contains(string(data), pin) {
			t.Errorf("missing pin %s", pin)
		}
	}
}
