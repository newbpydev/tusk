package tui

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	glamouransi "github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
)

func TestMarkdown_HTMLTextOnly(t *testing.T) {
	ctx := glamouransi.NewRenderContext(glamouransi.Options{})
	for _, tc := range []struct{ source, want string }{
		{"<b>Hello</b> &amp; <i>界</i>", "Hello & 界"},
		{"before<script>alert('bad')</script><style>css</style>after", "beforeafter"},
		{"<a href='https://invalid' onclick='bad'>label</a><img src='invalid'>", "label"},
		{"<!-- hidden -->visible", "visible"},
		{"<script>unclosed", ""},
		{"&lt;tag&gt; &amp; &#27;", "<tag> & \x1b"},
	} {
		if got := ctx.SanitizeHTML(tc.source, false); got != tc.want {
			t.Errorf("%q: got %q, want %q", tc.source, got, tc.want)
		}
	}
	if got := ctx.SanitizeHTML("  <b> text </b>  ", true); got != "text" {
		t.Fatalf("trimmed: %q", got)
	}
}

func TestMarkdown_ControlsAndNoExternalEffects(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { requests.Add(1) }))
	defer server.Close()
	dir := t.TempDir()
	style := filepath.Join(dir, "style.json")
	if err := os.WriteFile(style, []byte("invalid and must never be loaded"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GLAMOUR_STYLE", style)
	source := "# Heading\n\n- readable list\n\n| A | B |\n|---|---|\n| one | two |\n\n```sh\ntouch " + filepath.Join(dir, "executed") + "\n```\n\n[Link](" + server.URL + ") ![Image](" + server.URL + "/image)\n\n\x1b]52;c;bad\a\x1b[2J\u202e\n&#27;]52;c;entity&#7;\n<script>never execute</script>"
	for _, profile := range []termenv.Profile{termenv.Ascii, termenv.TrueColor} {
		out, err := RenderMarkdown(context.Background(), source, 40, profile)
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{"Heading", "readable", "one", "touch", "Link", "Image"} {
			if !strings.Contains(ansi.Strip(out), want) {
				t.Fatalf("missing %q", want)
			}
		}
		if strings.ContainsAny(out, "\a\u202e") || strings.Contains(out, "\x1b]") || strings.Contains(out, "\x1b[2J") {
			t.Fatalf("unsafe rendered output: %q", out)
		}
		if profile == termenv.Ascii && strings.Contains(out, "\x1b") {
			t.Fatal("plain Markdown has escapes")
		}
		for _, line := range strings.Split(out, "\n") {
			if ansi.StringWidth(line) > 40 {
				t.Fatal("render exceeds width")
			}
		}
	}
	if requests.Load() != 0 {
		t.Fatal("renderer fetched a URL")
	}
	if _, err := os.Stat(filepath.Join(dir, "executed")); !os.IsNotExist(err) {
		t.Fatal("renderer executed code")
	}
}

func TestMarkdown_GeneratedOutputAllowsOnlySGR(t *testing.T) {
	source := "\x1b[31mred\x1b[0m\x1b]8;;https://example.com\x1b\\link\x1b]8;;\x1b\\\x1b[2J\x1b[3A\x1bPpayload\x1b\\\a\r\u202e\xff"
	out := safeMarkdownOutput(source, termenv.TrueColor)
	if !strings.Contains(out, "\x1b[31mred\x1b[0m") || strings.ContainsAny(out, "\a\r\u202e") || strings.Contains(out, "\x1b]") || strings.Contains(out, "payload") || strings.Contains(out, "\x1b[2J") {
		t.Fatalf("unsafe output %q", out)
	}
	if strings.Contains(safeMarkdownOutput(source, termenv.Ascii), "\x1b") {
		t.Fatal("plain output has styles")
	}
}

func TestMarkdown_LongUnicodeAndCancellation(t *testing.T) {
	source := strings.Repeat("界é👩‍💻", 65536)
	lines, plain := renderNoteText(context.Background(), RenderMarkdown, source, 42, termenv.Ascii)
	if !plain {
		t.Fatal("large note bypassed formatting budget")
	}
	out := strings.Join(lines, "\n")
	if !strings.Contains(out, "é") || strings.Count(out, "界") != 65536 {
		t.Fatal("long note lost content")
	}
	for _, line := range strings.Split(out, "\n") {
		if ansi.StringWidth(line) > 42 {
			t.Fatal("wide note overflow")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := RenderMarkdown(ctx, "unused", 42, termenv.Ascii); err == nil {
		t.Fatal("canceled render did work")
	}
	if _, err := RenderMarkdown(context.Background(), "unused", 0, termenv.Ascii); err == nil {
		t.Fatal("invalid width accepted")
	}
}

func TestMarkdown_BudgetBoundariesAndGeneratedSGRGrammar(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   bool
	}{{strings.Repeat("a", 2048), true}, {"prefix " + strings.Repeat("a", 2048), true}, {strings.Repeat("a", 2049), false}, {strings.Repeat("a ", 16384), true}, {strings.Repeat("a ", 16385), false}} {
		if got := withinMarkdownBudget(tc.source); got != tc.want {
			t.Fatalf("budget (%d bytes): %v", len(tc.source), got)
		}
	}
	if safeSGR("\x1b[?25m") || safeSGR("\x1b[31x") || safeSGR("\x1b[mgarbage") {
		t.Fatal("invalid SGR sequence admitted")
	}
	if !safeSGR("\x1b[38:2:1:2:3m") {
		t.Fatal("truecolor SGR rejected")
	}
	lines, plain := renderNoteText(context.Background(), RenderMarkdown, strings.Repeat("a", 2049), 44, termenv.Ascii)
	if !plain || strings.Count(strings.Join(lines, ""), "a") != 2049 {
		t.Fatal("oversized word not preserved in fallback")
	}
}

func TestMarkdown_StyleMatchesWorkspaceWithoutSharedMutation(t *testing.T) {
	first := markdownStyles()
	second := markdownStyles()
	if first.H1.BackgroundColor != nil || first.H1.Color == nil || *first.H1.Color != accentColor || first.Document.Margin == nil || *first.Document.Margin != 0 {
		t.Fatal("Markdown does not use the workspace palette and padding")
	}
	*first.H1.Color = "changed"
	if *second.H1.Color != accentColor {
		t.Fatal("renderers share mutable style state")
	}
}
