package cli

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/newbpydev/tusk/internal/ports"
)

func staticOptions(t *testing.T, out, errout *bytes.Buffer) Options {
	t.Helper()
	return Options{
		Stdout: out, Stderr: errout,
		Getenv:   func(string) string { t.Error("configuration read during completion"); return "bad" },
		Terminal: func() TerminalFacts { t.Error("terminal sampled during completion"); return TerminalFacts{} },
		Confirm: func(context.Context) (bool, error) {
			t.Error("confirmation during completion")
			return false, ports.ErrStorage
		},
		OpenService: func(context.Context, Config) (ports.TaskService, func() error, error) {
			t.Error("service opened during completion")
			return nil, nil, ports.ErrStorage
		},
		RunTUI: func(context.Context, Config) (TUIResult, error) {
			t.Error("TUI started during completion")
			return TUIResult{}, ports.ErrStorage
		},
	}
}

func TestCompletion_StorageFreeScripts(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish"} {
		t.Run(shell, func(t *testing.T) {
			var out, errout bytes.Buffer
			opts := staticOptions(t, &out, &errout)
			if code := Run(t.Context(), []string{"completion", shell, "--timezone=invalid"}, opts); code != 0 || out.Len() == 0 || errout.Len() != 0 {
				t.Fatalf("code %d, script %q, stderr %q", code, out.String(), errout.String())
			}
			if !strings.Contains(out.String(), "tusk") || strings.Contains(out.String(), "tusk version") {
				t.Fatalf("unexpected script output %q", out.String())
			}
			for _, writer := range []badWriter{{}, {short: true}} {
				opts.Stdout = writer
				errout.Reset()
				if code := Run(t.Context(), []string{"completion", shell}, opts); code != 1 || strings.Count(errout.String(), "tusk:") != 1 {
					t.Fatalf("output failure: code %d stderr %q", code, errout.String())
				}
			}
		})
	}
}

func TestCompletion_Usage(t *testing.T) {
	for _, args := range [][]string{{"completion"}, {"completion", "pwsh"}, {"completion", "bash", "extra"}, {"completion", "fish", "--json"}, {"completion", "bash", "--wat"}} {
		var out, errout bytes.Buffer
		if code := Run(t.Context(), args, staticOptions(t, &out, &errout)); code != 2 || out.Len() != 0 || strings.Count(errout.String(), "tusk:") != 1 {
			t.Errorf("%v: code %d stdout %q stderr %q", args, code, out.String(), errout.String())
		}
	}
}

func TestCompletion_StaticRequests(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"__complete", ""}, "completion"},
		{[]string{"__complete", "list", "--"}, "--json"},
		{[]string{"__complete", "list", "--status", ""}, "in-progress"},
		{[]string{"__complete", "add", "--priority", ""}, "urgent"},
		{[]string{"__completeNoDesc", "edit", "id", "--priority", ""}, "medium"},
		{[]string{"__complete", "delete", ""}, ":4"},
		{[]string{"__complete", "add", "--tags", ""}, ":4"},
	} {
		var out, errout bytes.Buffer
		if code := Run(t.Context(), tc.args, staticOptions(t, &out, &errout)); code != 0 || !strings.Contains(out.String(), tc.want) || strings.Contains(out.String(), "Completion ended") || errout.Len() != 0 {
			t.Errorf("%v: code %d stdout %q stderr %q", tc.args, code, out.String(), errout.String())
		}
	}
}

func TestCompletion_ConcurrentFreshInvocations(t *testing.T) {
	var reference, errout bytes.Buffer
	if code := Run(t.Context(), []string{"completion", "bash"}, staticOptions(t, &reference, &errout)); code != 0 {
		t.Fatalf("reference exit %d", code)
	}
	var wait sync.WaitGroup
	for range 8 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			var out, errs bytes.Buffer
			if code := Run(t.Context(), []string{"completion", "bash"}, staticOptions(t, &out, &errs)); code != 0 || out.String() != reference.String() || errs.Len() != 0 {
				t.Errorf("independent script drift or failure: %d %s", code, errs.String())
			}
		}()
	}
	wait.Wait()
}

func TestCompletion_OrdinaryConstructionDoesNotRegisterGlobalCallbacks(t *testing.T) {
	root := newInvocation(Options{}).root
	list, _, err := root.Find([]string{"list"})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := list.GetFlagCompletionFunc("status"); ok {
		t.Fatal("ordinary startup registered callbacks in Cobra global state")
	}
}
