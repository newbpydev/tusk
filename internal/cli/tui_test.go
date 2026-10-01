package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/newbpydev/tusk/internal/ports"
)

func TestTUIAdmission_NoOpen(t *testing.T) {
	for _, tc := range []struct {
		args        []string
		in, out     bool
		term        string
		want, calls int
	}{
		{[]string{"tui"}, true, true, "xterm", 0, 1},
		{[]string{"tui"}, true, true, "", 0, 1},
		{[]string{"tui"}, true, false, "xterm", 1, 0},
		{[]string{"tui"}, false, true, "xterm", 1, 0},
		{[]string{"tui"}, false, false, "xterm", 1, 0},
		{[]string{"tui"}, true, true, "dumb", 1, 0},
		{[]string{"tui", "--json"}, true, true, "xterm", 2, 0},
		{[]string{"tui", "extra"}, true, true, "xterm", 2, 0},
		{[]string{"tui", "--wat"}, true, true, "xterm", 2, 0},
		{[]string{"tui", "--help"}, false, false, "dumb", 0, 0},
		{[]string{"help", "tui"}, false, false, "dumb", 0, 0},
	} {
		t.Run(strings.Join(tc.args, " ")+tc.term+string(rune(tc.want)), func(t *testing.T) {
			var out, errout bytes.Buffer
			calls := 0
			opts := Options{Stdout: &out, Stderr: &errout, Getenv: func(k string) string {
				if k == "TERM" {
					return tc.term
				}
				return ""
			}, Terminal: func() TerminalFacts { return TerminalFacts{In: tc.in, Out: tc.out} },
				OpenService: func(context.Context, Config) (ports.TaskService, func() error, error) {
					t.Fatal("CLI opened storage")
					return nil, nil, nil
				},
				RunTUI: func(context.Context, Config) (TUIResult, error) { calls++; return TUIResult{}, nil }}
			if got := Run(context.Background(), tc.args, opts); got != tc.want || calls != tc.calls {
				t.Fatalf("exit=%d calls=%d stderr=%q", got, calls, errout.String())
			}
			if tc.want == 1 && !strings.Contains(errout.String(), "tusk list --json") {
				t.Fatal("missing actionable terminal guidance")
			}
		})
	}
}

func TestTUI_Diagnostics(t *testing.T) {
	for _, tc := range []struct {
		result TUIResult
		err    error
		want   []string
	}{
		{TUIResult{}, errors.New("PRIVATE"), []string{"operation failed"}},
		{TUIResult{HadCommittedChanges: true}, ports.ErrConflict, []string{"earlier changes remain committed", ports.ErrConflict.Error()}},
		{TUIResult{HadCommittedChanges: true, OutcomeUnknown: true}, ports.ErrConflict, []string{"outcome unknown", "earlier changes remain committed"}},
		{TUIResult{}, ports.NewTransactionError("PRIVATE", ports.ErrBusy), []string{"outcome unknown"}},
	} {
		var out bytes.Buffer
		opts := Options{Stderr: &out, Terminal: func() TerminalFacts { return TerminalFacts{In: true, Out: true} }, RunTUI: func(context.Context, Config) (TUIResult, error) { return tc.result, tc.err }}
		if Run(context.Background(), []string{"tui"}, opts) != 1 {
			t.Fatal("exit")
		}
		for _, s := range tc.want {
			if !strings.Contains(out.String(), s) {
				t.Fatalf("missing %q: %s", s, &out)
			}
		}
		if strings.Contains(out.String(), "PRIVATE") {
			t.Fatal("private diagnostic")
		}
	}
}
