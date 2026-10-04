package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMarkdownPolicy(t *testing.T) {
	const ci = "https://github.com/newbpydev/tusk/actions/workflows/ci.yml/badge.svg?branch=main"
	const license = "https://img.shields.io/github/license/newbpydev/tusk"
	for _, tc := range []struct {
		name, source, diagnostic string
	}{
		{"empty", "", ""},
		{"continued approved", "[x]:\n  " + ci + "\n\n![OK][x]", ""},
		{"multiline approved", "[Foo\n bar]: " + license + "\n\n![OK][Foo bar]", ""},
		{"multiline inline approved", "![OK](\n<" + ci + ">\n 'Title ü')", ""},
		{"continued unknown", "[x]:\n  https://evil.example/status.svg\n\n![No][x]", "unverified badge"},
		{"multiline unknown", "[Foo\n bar]: https://evil.example/status.svg\n\n![No][Foo bar]", "unverified badge"},
		{"unused continued", "[unused]:\n  https://evil.example/status.svg", "unverified badge"},
		{"duplicate definition", "[x]: " + ci + "\n[x]: https://evil.example/status.svg\n\n![No][x]", "unverified badge"},
		{"normalized duplicate", "[Foo bar]: " + ci + "\n[FOO\n BAR]: https://evil.example/status.svg", "unverified badge"},
		{"quote reference", "> [x]:\n>   https://evil.example/status.svg\n>\n> ![No][x]", "unverified badge"},
		{"list reference", "- [x]:\n    https://evil.example/status.svg\n\n![No][x]", "unverified badge"},
		{"image", "![No](https://evil.example/status.svg)", "unverified badge"},
		{"protocol relative", "![No](//evil.example/a.svg)", "unverified badge"},
		{"data", "![No](data:image/svg+xml,anything)", "unverified badge"},
		{"absolute image", "![No](/image.svg)", "unverified badge"},
		{"local unicode", "![Yes](docs/图.png) [Path](https://example.com/路径)", ""},
		{"ordinary link", "[Other](https://example.com/?a=1&b=2 'Title ü')", ""},
		{"angle autolink", "<https://example.com/a%20b>", "literal value"},
		{"angle whitespace", "[Other](<https://example.com/a b>)", "literal value"},
		{"userinfo unicode host", "[Other](https://user@img。shields.io/a)", "literal ASCII"},
		{"unfinished literal link", "[Other](https://example.com/a", "unsupported link"},
		{"escaped", `![No](https://img\.shields\.io/a)`, "literal value"},
		{"entity", "[x]: &#104;ttps://evil.example/a", "literal value"},
		{"named entity", "[Other](https://example.com/?a=&amp;)", "literal value"},
		{"encoded path", "[Other](https://example.com/a%20b)", "literal value"},
		{"unicode host", "[Other](https://img。shields.io/a)", "literal ASCII"},
		{"entity authority", "[Other](https://example&shy.com/a)", "literal ASCII"},
		{"split image", "![No](https://evil.\nexample/a)", "unsupported image"},
		{"split angle image", "![No](<https://evil.\nexample/a>)", "unsupported image"},
		{"nested link", "![a [b](u1)](docs/image.png)", "unsupported image"},
		{"unfinished image", "![a [b](u1)", "unsupported image"},
		{"ordinary nested label", "[a [b](u1)", ""},
		{"escaped opener", `\![x] missing`, ""},
		{"escaped slash before opener", `\\![x] missing`, "unsupported image"},
		{"image in code", "`![No](https://example.com/a%20b)`", ""},
		{"indented code", "    ![No](https://example.com/a%20b)\n", ""},
		{"quoted code", "> ```text\n> ![No](https://example.com/a%20b)\n> ```", ""},
		{"list code", "- item\n\n        ![No](https://example.com/a%20b)", "literal value"},
		{"list fenced code", "- item\n  ```text\n  ![No](https://example.com/a%20b)\n  ```", "literal value"},
		{"deep list code", "- item\n\n            ![No](https://example.com/a%20b)", "literal value"},
		{"list nested fence", "- item\n  ~~~~text\n  ```\n  ![No](https://example.com/a%20b)\n  ```\n  ~~~~", "literal value"},
		{"wrapped code", "See `[x](https://example.com/a%20b\ntail)`.", "literal value"},
		{"wrapped valid code", "See `[x](https://example.com/a\n)`.", ""},
		{"quote tab", "> ```text\n> example\n>\t```\n> ![No](https://img。shields.io/a)", "literal ASCII"},
		{"footnote", "[^note]: https://example.com/a%20b", ""},
		{"plain url", "Example https://example.com/a%20b", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := check([]byte(tc.source))
			if tc.diagnostic == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tc.diagnostic) {
				t.Fatalf("want diagnostic %q, got %v", tc.diagnostic, err)
			}
		})
	}
}

func TestRun(t *testing.T) {
	dir := t.TempDir()
	valid := filepath.Join(dir, "valid.md")
	invalid := filepath.Join(dir, "invalid.md")
	for name, contents := range map[string]string{valid: "![Local](image.png)", invalid: "![No](https://evil.example/a)"} {
		if err := os.WriteFile(name, []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		args       []string
		status     int
		diagnostic string
	}{
		{nil, 2, "expected"},
		{[]string{valid, invalid}, 2, "expected"},
		{[]string{filepath.Join(dir, "missing")}, 1, "doccheck:"},
		{[]string{invalid}, 1, "unverified badge"},
		{[]string{valid}, 0, ""},
	} {
		var stderr bytes.Buffer
		if got := run(tc.args, &stderr); got != tc.status {
			t.Fatalf("%v: status %d, want %d", tc.args, got, tc.status)
		}
		if !strings.Contains(stderr.String(), tc.diagnostic) {
			t.Fatalf("%v: diagnostic %q", tc.args, stderr.String())
		}
	}
}
