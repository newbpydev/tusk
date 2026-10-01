// Package terminaltext renders untrusted text as inert terminal content.
// Display values must never replace the raw values used for persistence.
package terminaltext

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Escape terminal controls, directional overrides and the explicit blank fillers
// below. Shaping marks (CGJ, Khmer vowels, Mongolian selectors), variation
// selectors and ZWJ/ZWNJ stay intact: they can select meaningful glyphs. This
// display boundary does not canonicalize text identity; task IDs identify tasks.
func control(r rune) bool {
	return r < 32 || r >= 127 && r <= 159 || r == 0x00ad || r == 0x115f || r == 0x1160 || r == 0x180e || r == 0x3164 || r == 0xffa0 || r == 0x200b || r >= 0x2060 && r <= 0x206f || r >= 0xfff9 && r <= 0xfffb || r == 0xe0001 || r >= 0xe0020 && r <= 0xe007f || r == 0x061c || r == 0x200e || r == 0x200f || r == 0xfeff || r == 0x2028 || r == 0x2029 || r >= 0x202a && r <= 0x202e
}

// Scalar preserves the CLI's visible escaping, including literal \n, \r and \t.
func Scalar(text string) string { return display(text, false) }

// Multiline retains line feeds, expands each tab to four spaces and escapes
// every other control. Invalid UTF-8 is rendered as replacement characters.
func Multiline(text string) string { return display(text, true) }

func display(text string, multiline bool) string {
	first := strings.IndexFunc(text, func(r rune) bool { return r == utf8.RuneError || control(r) })
	if first < 0 {
		return text
	}
	var out strings.Builder
	out.Grow(len(text))
	out.WriteString(text[:first])
	for _, r := range text[first:] {
		switch r {
		case '\n':
			if multiline {
				out.WriteByte('\n')
			} else {
				out.WriteString(`\n`)
			}
		case '\r':
			out.WriteString(`\r`)
		case '\t':
			if multiline {
				out.WriteString("    ")
			} else {
				out.WriteString(`\t`)
			}
		default:
			if control(r) {
				fmt.Fprintf(&out, `\u%04x`, r)
			} else {
				out.WriteRune(r)
			}
		}
	}
	return out.String()
}
