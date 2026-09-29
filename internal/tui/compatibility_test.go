package tui

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
)

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
