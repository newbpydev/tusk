package ansi

import (
	"strings"

	"golang.org/x/net/html"
)

// RenderContext holds the current rendering options and state.
type RenderContext struct {
	options Options

	blockStack *BlockStack
	table      *TableElement
}

// NewRenderContext returns a new RenderContext.
func NewRenderContext(options Options) RenderContext {
	return RenderContext{
		options:    options,
		blockStack: &BlockStack{},
		table:      &TableElement{},
	}
}

// SanitizeHTML extracts text only. No tags or attributes survive into terminal
// output; script/style contents are omitted. A CSS/URL policy registry is not
// needed because this renderer never emits HTML. Tusk also sanitizes terminal
// controls after Markdown rendering, including decoded character references.
func (ctx RenderContext) SanitizeHTML(s string, trimSpaces bool) string {
	tokens := html.NewTokenizer(strings.NewReader(s))
	var out strings.Builder
	var skip string
	for {
		kind := tokens.Next()
		if kind == html.ErrorToken {
			break
		}
		switch kind {
		// Self-closing raw-text tags (<script/>) still enter rawtext mode in
		// x/net/html, so the skip must arm on both token kinds.
		case html.StartTagToken, html.SelfClosingTagToken:
			name, _ := tokens.TagName()
			if string(name) == "script" || string(name) == "style" {
				skip = string(name)
			}
		case html.EndTagToken:
			name, _ := tokens.TagName()
			if string(name) == skip {
				skip = ""
			}
		case html.TextToken:
			if skip == "" {
				out.Write(tokens.Text())
			}
		}
	}
	s = out.String()
	if trimSpaces {
		s = strings.TrimSpace(s)
	}

	return s
}
