package cli

import (
	"testing"

	"github.com/newbpydev/tusk/internal/terminaltext"
)

func TestCLI_SanitizerCompatibility(t *testing.T) {
	for _, tc := range []struct{ raw, want string }{
		{"plain", "plain"}, {"界 é 👩‍💻", "界 é 👩‍💻"},
		{"\xff\x00\r\n\t\x7f\u0085", `�\u0000\r\n\t\u007f\u0085`},
		{"\x1b[31m\x1b]8;;url\a\u202e\u2069", `\u001b[31m\u001b]8;;url\u0007\u202e\u2069`},
	} {
		if got := sanitize(tc.raw); got != tc.want {
			t.Fatalf("CLI %q: %q want %q", tc.raw, got, tc.want)
		}
		if got := terminaltext.Scalar(tc.raw); got != tc.want {
			t.Fatalf("shared %q: %q want %q", tc.raw, got, tc.want)
		}
	}
}
