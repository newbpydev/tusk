package tui

import (
	"context"
	"errors"
	"strings"

	"github.com/charmbracelet/glamour"
	glamouransi "github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
	"github.com/newbpydev/tusk/internal/terminaltext"
)

// RenderMarkdown owns a fresh renderer. Neither environment-selected styles nor
// automatic terminal discovery are admitted; links and images remain text.
func RenderMarkdown(ctx context.Context, source string, width int, profile termenv.Profile) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if width < 1 {
		return "", errors.New("invalid render width")
	}
	source = terminaltext.Multiline(source)
	if !withinMarkdownBudget(source) {
		return "", errors.New("plain text formatting budget")
	}
	renderer, err := glamour.NewTermRenderer(glamour.WithStyles(markdownStyles()), glamour.WithWordWrap(width), glamour.WithColorProfile(profile))
	if err != nil {
		return "", err
	}
	out, err := renderer.Render(source)
	if err != nil {
		return "", err
	}
	if err = ctx.Err(); err != nil {
		return "", err
	}
	// The renderer may decode entities or emit its own control sequences. Apply
	// the output boundary too, before wrapping and handing text to the compositor.
	return ansi.Hardwrap(safeMarkdownOutput(out, profile), width, true), nil
}

// Start from the fixed built-in grammar styles; replace values locally without
// changing any shared style pointer. The viewport supplies its own padding.
func markdownStyles() glamouransi.StyleConfig {
	s := styles.DarkStyleConfig
	fg, accent, muted, bg := textColor, accentColor, mutedColor, dialogColor
	zero := uint(0)
	bold := true
	s.Document.Color = &fg
	s.Document.Margin = &zero
	s.Heading.Color = &accent
	for _, h := range []*glamouransi.StyleBlock{&s.H1, &s.H2, &s.H3, &s.H4, &s.H5, &s.H6} {
		h.Color = &accent
		h.BackgroundColor = nil
		h.Prefix = ""
		h.Suffix = ""
		h.Bold = &bold
	}
	s.Link.Color = &accent
	s.LinkText.Color = &accent
	s.Image.Color = &muted
	s.ImageText.Color = &fg
	s.Code.Color = &accent
	s.Code.BackgroundColor = &bg
	if s.CodeBlock.Chroma != nil {
		chroma := *s.CodeBlock.Chroma
		chroma.Comment.Color = &muted
		s.CodeBlock.Chroma = &chroma
	}
	return s
}

func withinMarkdownBudget(source string) bool {
	if len(source) > 32*1024 {
		return false
	}
	for _, word := range strings.Fields(source) {
		if len(word) > 2048 {
			return false
		}
	}
	return true
}

type MarkdownRenderer func(context.Context, string, int, termenv.Profile) (string, error)

func renderNoteText(ctx context.Context, render MarkdownRenderer, source string, width int, profile termenv.Profile) (lines []string, plain bool) {
	defer func() {
		if recover() != nil {
			lines = strings.Split(ansi.Hardwrap(terminaltext.Multiline(source), max(1, width), true), "\n")
			plain = true
		}
	}()
	out, err := render(ctx, source, width, profile)
	if err != nil {
		out = terminaltext.Multiline(source)
		plain = true
	} else {
		out = safeMarkdownOutput(out, profile)
	}
	return strings.Split(ansi.Hardwrap(out, max(1, width), true), "\n"), plain
}

func safeMarkdownOutput(s string, profile termenv.Profile) string {
	var out strings.Builder
	for len(s) > 0 {
		i := strings.IndexByte(s, '\x1b')
		if i < 0 {
			out.WriteString(terminaltext.Multiline(s))
			break
		}
		out.WriteString(terminaltext.Multiline(s[:i]))
		s = s[i:]
		seq, _, n, _ := ansi.DecodeSequence(s, 0, nil)
		if profile != termenv.Ascii && safeSGR(seq) {
			out.WriteString(seq)
		}
		s = s[max(1, n):]
	}
	return out.String()
}

func safeSGR(s string) bool {
	if len(s) < 3 || !strings.HasPrefix(s, "\x1b[") || s[len(s)-1] != 'm' {
		return false
	}
	for _, c := range s[2 : len(s)-1] {
		if c != ';' && c != ':' && (c < '0' || c > '9') {
			return false
		}
	}
	return true
}
