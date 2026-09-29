package cli

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func terminalControl(r rune) bool {
	return r < 32 || r >= 127 && r <= 159 || r == 0x061c || r == 0x200e || r == 0x200f || r == 0xfeff || r == 0x2028 || r == 0x2029 || r >= 0x202a && r <= 0x202e || r >= 0x2066 && r <= 0x2069
}

func sanitize(text string) string {
	first := strings.IndexFunc(text, func(r rune) bool { return r == utf8.RuneError || terminalControl(r) })
	if first < 0 {
		return text
	}
	var out strings.Builder
	out.Grow(len(text))
	out.WriteString(text[:first])
	for _, r := range text[first:] {
		switch r {
		case '\n':
			out.WriteString(`\n`)
		case '\r':
			out.WriteString(`\r`)
		case '\t':
			out.WriteString(`\t`)
		default:
			if terminalControl(r) {
				fmt.Fprintf(&out, `\u%04x`, r)
			} else {
				out.WriteRune(r)
			}
		}
	}
	return out.String()
}
