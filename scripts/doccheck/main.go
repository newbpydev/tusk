// Doccheck applies README URL policy to parsed Markdown, without rendering it.
package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"unicode"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

const ciBadge = "https://github.com/newbpydev/tusk/actions/workflows/ci.yml/badge.svg?branch=main"
const licenseBadge = "https://img.shields.io/github/license/newbpydev/tusk"

func main() { os.Exit(run(os.Args[1:], os.Stderr)) }

func run(args []string, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "doccheck: expected one README path")
		return 2
	}
	source, err := os.ReadFile(args[0])
	if err == nil {
		err = check(source)
	}
	if err != nil {
		fmt.Fprintln(stderr, "doccheck:", err)
		return 1
	}
	return 0
}

// Goldmark resolves the first definition for each normalized label. Policy
// also checks unused and shadowed definitions, so retain every AddReference.
type referenceContext struct {
	parser.Context
	definitions []parser.Reference
}

func (c *referenceContext) AddReference(ref parser.Reference) {
	c.definitions = append(c.definitions, ref)
	c.Context.AddReference(ref)
}

func check(source []byte) error {
	return checkMarkdown(source, false)
}

func checkMarkdown(source []byte, inspectBlocks bool) error {
	context := &referenceContext{Context: parser.NewContext()}
	markdown := goldmark.New(goldmark.WithExtensions(extension.Footnote))
	document := markdown.Parser().Parse(text.NewReader(source), parser.WithContext(context))
	for _, ref := range context.definitions {
		if err := destination(string(ref.Destination()), true); err != nil {
			return err
		}
	}
	return ast.Walk(document, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		var err error
		switch n := node.(type) {
		case *ast.Image:
			err = destination(string(n.Destination), true)
		case *ast.Link:
			if ancestor(n, ast.KindImage) {
				err = fmt.Errorf("README unsupported image label: nested link")
			} else {
				err = destination(string(n.Destination), false)
			}
		case *ast.AutoLink:
			err = destination(string(n.URL(source)), false)
		case *ast.Text:
			err = malformedTarget(n.Value(source))
		case *ast.CodeSpan:
			var content []byte
			for child := n.FirstChild(); child != nil; child = child.NextSibling() {
				content = append(content, child.(*ast.Text).Value(source)...)
			}
			// Only single-line inline examples are excluded by README policy.
			if bytes.ContainsAny(content, "\r\n") {
				err = checkMarkdown(content, true)
			}
			return ast.WalkSkipChildren, err
		case *ast.CodeBlock, *ast.FencedCodeBlock:
			// Preserve the declared conservative tripwire in list-contained
			// code and quote-tab code. Goldmark owns all block boundaries.
			content := node.Lines().Value(source)
			quoteTab := false
			if ancestor(node, ast.KindBlockquote) {
				for i := 0; i < node.Lines().Len(); i++ {
					line := node.Lines().At(i)
					start := bytes.LastIndexByte(source[:line.Start], '\n') + 1
					quoteTab = quoteTab || bytes.Contains(source[start:line.Start], []byte(">\t"))
				}
			}
			if inspectBlocks || ancestor(node, ast.KindList) || quoteTab {
				err = checkMarkdown(content, true)
			}
			return ast.WalkSkipChildren, err
		}
		return ast.WalkContinue, err
	})
}

func ancestor(node ast.Node, kind ast.NodeKind) bool {
	for parent := node.Parent(); parent != nil; parent = parent.Parent() {
		if parent.Kind() == kind {
			return true
		}
	}
	return false
}

// Malformed resource markup must not certify successfully merely because a
// renderer treats it as text. This checks active Text nodes, never code spans.
func malformedTarget(value []byte) error {
	for i := 0; i+1 < len(value); i++ {
		if value[i] == '\\' {
			i++
			continue
		}
		if value[i] == '!' && value[i+1] == '[' {
			return fmt.Errorf("README unsupported image target or label")
		}
		if value[i] == ']' && value[i+1] == '(' {
			target := strings.TrimLeft(string(value[i+2:]), " \t<")
			if end := strings.IndexAny(target, " \t\r\n)>"); end >= 0 {
				target = target[:end]
			}
			if needsLiteral(target) {
				if err := literal(target); err != nil {
					return err
				}
				return fmt.Errorf("README unsupported link target syntax")
			}
		}
	}
	return nil
}

func needsLiteral(value string) bool {
	lower := strings.ToLower(value)
	encoded, _ := regexp.MatchString(`&#|&[[:alpha:]][[:alnum:]]*;|%[[:xdigit:]][[:xdigit:]]`, value)
	return strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "http://") ||
		strings.HasPrefix(lower, "//") || encoded || strings.Contains(value, `\`)
}

func destination(value string, image bool) error {
	if image || needsLiteral(value) {
		if err := literal(value); err != nil {
			return err
		}
	}
	remote, _ := regexp.MatchString(`^[[:alpha:]][[:alnum:]+.-]*:|^/`, value)
	if image && remote && value != ciBadge && value != licenseBadge {
		return fmt.Errorf("README advertises an unverified badge: remote image/reference target %s", value)
	}
	return nil
}

func literal(value string) error {
	encoded, _ := regexp.MatchString(`&#|&[[:alpha:]][[:alnum:]]*;|%[[:xdigit:]][[:xdigit:]]`, value)
	if value == "" || strings.ContainsFunc(value, unicode.IsSpace) || strings.Contains(value, `\`) || encoded {
		return fmt.Errorf("README URL target requires a literal value: %s", value)
	}
	lower := strings.ToLower(value)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "//") {
		host := value[strings.Index(value, "//")+2:]
		if end := strings.IndexAny(host, "/?#"); end >= 0 {
			host = host[:end]
		}
		if user := strings.LastIndexByte(host, '@'); user >= 0 {
			host = host[user+1:]
		}
		if strings.ContainsFunc(host, func(r rune) bool { return r < 32 || r > 126 || r == '&' }) {
			return fmt.Errorf("README URL host requires literal ASCII: %s", host)
		}
	}
	return nil
}
