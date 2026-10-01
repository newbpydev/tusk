package terminaltext_test

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/newbpydev/tusk/internal/terminaltext"
)

func TestTerminalText_ControlAndUnicode(t *testing.T) {
	for _, tc := range []struct{ raw, scalar, multiline string }{
		{"", "", ""},
		{"界 e\u0301 👩‍💻", "界 e\u0301 👩‍💻", "界 e\u0301 👩‍💻"},
		{"a\n\t\rb", `a\n\t\rb`, "a\n    \\rb"},
		{"a\xffb", "a�b", "a�b"},
		{"\x1b]52;c;payload\a", `\u001b]52;c;payload\u0007`, `\u001b]52;c;payload\u0007`},
		{"\x1b]8;;https://invalid\x1b\\link\x1b]8;;\a", `\u001b]8;;https://invalid\u001b\link\u001b]8;;\u0007`, `\u001b]8;;https://invalid\u001b\link\u001b]8;;\u0007`},
	} {
		if got := terminaltext.Scalar(tc.raw); got != tc.scalar {
			t.Errorf("Scalar(%q)=%q want %q", tc.raw, got, tc.scalar)
		}
		if got := terminaltext.Multiline(tc.raw); got != tc.multiline {
			t.Errorf("Multiline(%q)=%q want %q", tc.raw, got, tc.multiline)
		}
	}
	var controls []rune
	for r := rune(0); r < 32; r++ {
		controls = append(controls, r)
	}
	for r := rune(127); r <= 159; r++ {
		controls = append(controls, r)
	}
	controls = append(controls, 0x061c, 0x200e, 0x200f, 0xfeff, 0x2028, 0x2029, 0x202a, 0x202b, 0x202c, 0x202d, 0x202e, 0x2066, 0x2067, 0x2068, 0x2069)
	for _, r := range controls {
		raw := "a" + string(r) + "b"
		want := fmt.Sprintf(`a\u%04xb`, r)
		switch r {
		case '\n':
			want = `a\nb`
		case '\r':
			want = `a\rb`
		case '\t':
			want = `a\tb`
		}
		if got := terminaltext.Scalar(raw); got != want {
			t.Errorf("%U: %q want %q", r, got, want)
		}
		if r == '\n' {
			want = "a\nb"
		}
		if r == '\t' {
			want = "a    b"
		}
		if got := terminaltext.Multiline(raw); got != want {
			t.Errorf("multiline %U: %q want %q", r, got, want)
		}
	}
	for _, s := range []string{"ordinary task", "界 e\u0301 👩‍💻", "quotes \" & < >"} {
		if allocs := testing.AllocsPerRun(100, func() {
			if terminaltext.Scalar(s) != s {
				t.Fatal("changed safe text")
			}
		}); allocs != 0 {
			t.Fatalf("safe text allocates: %v", allocs)
		}
	}
}

func FuzzTerminalText(f *testing.F) {
	for _, s := range []string{"界 é 👩‍💻", "\x1b]52;c;payload\a", "\xff\n\t\u202e"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		for _, out := range []string{terminaltext.Scalar(s), terminaltext.Multiline(s)} {
			if !utf8.ValidString(out) {
				t.Fatal("invalid UTF-8")
			}
			for _, r := range out {
				if (r < 32 && r != '\n') || (r >= 127 && r <= 159) || strings.ContainsRune("\u061c\u200e\u200f\ufeff\u2028\u2029\u202a\u202b\u202c\u202d\u202e\u2066\u2067\u2068\u2069", r) {
					t.Fatalf("active control %U", r)
				}
			}
		}
	})
}

func TestPR5_InvisibleFormattingIsVisible(t *testing.T) {
	for _, r := range []rune{0x115f, 0x1160, 0x180e, 0x3164, 0xffa0, 0x2065, 0x206a, 0x206f, 0x200b, 0x00ad, 0x2060, 0x2061, 0x2062, 0x2063, 0x2064, 0xfff9, 0xfffa, 0xfffb, 0xe0001, 0xe0020, 0xe007f} {
		raw := "a" + string(r) + "b"
		want := fmt.Sprintf(`a\u%04xb`, r)
		if terminaltext.Scalar(raw) != want || terminaltext.Multiline(raw) != want {
			t.Errorf("invisible %U not shown", r)
		}
	}
	if terminaltext.Scalar("한글 👩‍💻 ا\u200cب") != "한글 👩‍💻 ا\u200cب" {
		t.Fatal("ZWJ emoji broken")
	}
}

func TestPR5_ShapingAndVariationExemptions(t *testing.T) {
	for _, r := range []rune{0x034f, 0x17b4, 0x17b5, 0x180b, 0x180c, 0x180d, 0x180f, 0xfe00, 0xfe0f, 0xe0100, 0xe01ef, 0x200c, 0x200d} {
		raw := "a" + string(r) + "b"
		if terminaltext.Scalar(raw) != raw || terminaltext.Multiline(raw) != raw {
			t.Errorf("shaping exemption %U altered", r)
		}
	}
	for _, raw := range []string{"☑️", "👩‍💻", "葛\U000e0100", "ᠠ\u180b", "a\u034f\u0301", "ក\u17b4", "ا\u200cب"} {
		if terminaltext.Scalar(raw) != raw || terminaltext.Multiline(raw) != raw {
			t.Errorf("composed text altered: %q", raw)
		}
	}
}
