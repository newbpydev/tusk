package main

import (
	"testing"

	"github.com/muesli/termenv"
)

func TestTerminal_TUIProfile(t *testing.T) {
	for _, tc := range []struct {
		name, term, color, noColor string
		want                       termenv.Profile
	}{
		{"truecolor", "xterm-kitty", "truecolor", "", termenv.TrueColor},
		{"24bit", "xterm-256color", "24bit", "", termenv.TrueColor},
		{"256", "xterm-256color", "", "", termenv.ANSI256},
		{"basic", "vt100", "", "", termenv.ANSI},
		{"no_color", "xterm-kitty", "truecolor", "1", termenv.Ascii},
		{"dumb", "dumb", "truecolor", "", termenv.Ascii},
		{"missing", "", "truecolor", "", termenv.Ascii},
	} {
		t.Run(tc.name, func(t *testing.T) {
			env := map[string]string{"TERM": tc.term, "COLORTERM": tc.color, "NO_COLOR": tc.noColor}
			if got := tuiProfile(func(k string) string { return env[k] }); got != tc.want {
				t.Fatalf("profile=%v want %v", got, tc.want)
			}
		})
	}
}

func TestTerminal_NilTUIEnvironmentIsEmpty(t *testing.T) {
	if got := tuiProfile(nil); got != termenv.Ascii {
		t.Fatalf("nil environment profile %v", got)
	}
}
