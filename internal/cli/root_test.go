package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/tusk/internal/ports"
	"github.com/spf13/cobra"
)

type fakeService struct{ ports.TaskService }
type badWriter struct{ short bool }

func (w badWriter) Write(p []byte) (int, error) {
	if w.short {
		return len(p) - 1, nil
	}
	return 0, io.ErrClosedPipe
}

func TestRoot_NoStorageForHelp(t *testing.T) {
	for _, args := range [][]string{nil, {"-h"}, {"--help"}, {"help"}, {"-v"}, {"--version"}, {"version"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var out, errout bytes.Buffer
			opts := Options{Stdout: &out, Stderr: &errout, Version: "test", Getenv: func(string) string { t.Fatal("config read"); return "" }, OpenService: func(context.Context, Config) (ports.TaskService, func() error, error) {
				t.Fatal("opened")
				return nil, nil, nil
			}}
			if got := Run(context.Background(), args, opts); got != 0 || out.Len() == 0 || errout.Len() != 0 {
				t.Fatalf("exit %d, out %q, err %q", got, out.String(), errout.String())
			}
			opts.Stdout = badWriter{}
			if got := Run(context.Background(), args, opts); got != 1 {
				t.Fatalf("writer exit %d", got)
			}
		})
	}
}

func TestRoot_Syntax(t *testing.T) {
	for _, args := range [][]string{{"wat"}, {"--wat"}, {"--timezone"}, {"--auto-complete-parent=wat"}, {"version", "extra"}, {"version", "--json"}, {"help", "wat"}, {"--help", "--wat"}} {
		var out, errout bytes.Buffer
		got := Run(context.Background(), args, Options{Stdout: &out, Stderr: &errout})
		if got != 2 || out.Len() != 0 || !strings.Contains(errout.String(), "Usage:") {
			t.Fatalf("%q: exit %d out %q err %q", args, got, out.String(), errout.String())
		}
	}
}

func fixtureRun(ctx context.Context, args []string, opts Options, work func(ports.TaskService) ([]byte, bool, error)) int {
	inv := newInvocation(opts)
	c := &cobra.Command{Use: "probe <title>", Args: syntaxArgs(cobra.ExactArgs(1)), RunE: func(c *cobra.Command, _ []string) error { return inv.invoke(c, work) }}
	c.Flags().String("scalar", "", "")
	c.Flags().StringArray("items", nil, "")
	inv.root.AddCommand(c)
	return inv.execute(ctx, args)
}

func TestRun_CloseAndOutcome(t *testing.T) {
	for _, tc := range []struct {
		name              string
		workErr, closeErr error
		committed         bool
		writer            io.Writer
		want              int
		hint              string
	}{
		{name: "success"},
		{name: "service", workErr: errors.New("unknown flag PRIVATE"), want: 1},
		{name: "close", closeErr: errors.New("PRIVATE"), want: 1},
		{name: "committed close", closeErr: errors.New("PRIVATE"), committed: true, want: 1, hint: "committed"},
		{name: "committed writer", writer: badWriter{}, committed: true, want: 1, hint: "committed"},
		{name: "short", writer: badWriter{short: true}, want: 1},
		{name: "unknown", workErr: ports.NewTransactionError("PRIVATE", context.Canceled), want: 1, hint: "outcome unknown"},
		{name: "cancel", workErr: context.Canceled, want: 1, hint: "canceled"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out, errout bytes.Buffer
			writes := io.Writer(&out)
			if tc.writer != nil {
				writes = tc.writer
			}
			opened, closed, calls := 0, 0, 0
			opts := Options{Stdout: writes, Stderr: &errout, OpenService: func(context.Context, Config) (ports.TaskService, func() error, error) {
				opened++
				return fakeService{}, func() error { closed++; return tc.closeErr }, nil
			}}
			got := fixtureRun(context.Background(), []string{"probe", "title"}, opts, func(ports.TaskService) ([]byte, bool, error) {
				calls++
				return []byte("result\n"), tc.committed, tc.workErr
			})
			if got != tc.want || opened != 1 || closed != 1 || calls != 1 {
				t.Fatalf("got %d opened %d closed %d calls %d", got, opened, closed, calls)
			}
			if strings.Contains(errout.String(), "PRIVATE") || !strings.Contains(errout.String(), tc.hint) {
				t.Fatalf("diagnostic %q", errout.String())
			}
			if got != 0 && out.Len() != 0 {
				t.Fatalf("pending output leaked %q", out.String())
			}
		})
	}
}

func TestConfig_Precedence(t *testing.T) {
	for _, tc := range []struct {
		args []string
		env  map[string]string
		want bool
		zone string
		exit int
	}{
		{zone: "local"},
		{env: map[string]string{"TUSK_AUTO_COMPLETE_PARENT": "true"}, want: true, zone: "local"},
		{env: map[string]string{"TUSK_AUTO_COMPLETE_PARENT": "bad"}, exit: 1},
		{args: []string{"--auto-complete-parent=false"}, env: map[string]string{"TUSK_AUTO_COMPLETE_PARENT": "bad"}, zone: "local"},
		{args: []string{"--auto-complete-parent"}, want: true, zone: "local"},
		{args: []string{"--timezone=UTC"}, env: map[string]string{"TUSK_TIMEZONE": "bad"}, zone: "UTC"},
		{env: map[string]string{"TUSK_TIMEZONE": "America/Sao_Paulo"}, zone: "America/Sao_Paulo"},
		{args: []string{"--timezone="}, exit: 1},
		{args: []string{"--timezone=not-a-zone"}, exit: 1},
	} {
		calls := 0
		opts := Options{Stdout: io.Discard, Stderr: io.Discard, Local: time.FixedZone("local", 3600), Getenv: func(k string) string { return tc.env[k] }, OpenService: func(_ context.Context, c Config) (ports.TaskService, func() error, error) {
			calls++
			if c.AutoCompleteParent != tc.want || c.Location.String() != tc.zone {
				t.Fatalf("config %+v", c)
			}
			return fakeService{}, func() error { return nil }, nil
		}}
		got := fixtureRun(context.Background(), append([]string{"probe", "title"}, tc.args...), opts, func(ports.TaskService) ([]byte, bool, error) { return nil, false, nil })
		if got != tc.exit || (tc.exit != 0 && calls != 0) {
			t.Fatalf("%+v: exit %d calls %d", tc, got, calls)
		}
	}
}

func TestConfig_ActionablePrivateDiagnostics(t *testing.T) {
	for _, tc := range []struct{ key, value, hint string }{
		{"TUSK_AUTO_COMPLETE_PARENT", "PRIVATE", "TUSK_AUTO_COMPLETE_PARENT"},
		{"TUSK_TIMEZONE", "PRIVATE", "TUSK_TIMEZONE"},
	} {
		var stderr bytes.Buffer
		code := Run(context.Background(), []string{"list"}, Options{Stderr: &stderr, Getenv: func(k string) string {
			if k == tc.key {
				return tc.value
			}
			return ""
		}, OpenService: func(context.Context, Config) (ports.TaskService, func() error, error) {
			t.Fatal("opened for invalid configuration")
			return nil, nil, nil
		}})
		if code != 1 || !strings.Contains(stderr.String(), tc.hint) || strings.Contains(stderr.String(), tc.value) {
			t.Errorf("code %d, diagnostic %q", code, stderr.String())
		}
	}
	var stderr bytes.Buffer
	if code := Run(context.Background(), []string{"list", "--timezone="}, Options{Stderr: &stderr}); code != 1 || !strings.Contains(stderr.String(), "--timezone") {
		t.Fatalf("code %d, diagnostic %q", code, stderr.String())
	}
}

func TestDiagnostic_DeletionRemediation(t *testing.T) {
	for _, tc := range []struct {
		err  error
		hint string
	}{
		{ports.ErrChildrenPresent, "--recursive"},
		{errForceRequired, "restore terminal input/output"},
	} {
		code, message := diagnostic(tc.err, false)
		if code != 1 || !strings.Contains(message, tc.hint) {
			t.Errorf("%d %q", code, message)
		}
	}
}

func TestRoot_HelpPrecedenceAndBoundaries(t *testing.T) {
	for _, args := range [][]string{{"probe", "--help"}, {"probe", "--timezone=bad", "--help"}} {
		if got := fixtureRun(context.Background(), args, Options{Stdout: io.Discard}, nil); got != 0 {
			t.Fatalf("%q: %d", args, got)
		}
	}
	for _, args := range [][]string{{"probe"}, {"probe", "a", "b"}, {"probe", "--help", "--auto-complete-parent=bad"}} {
		if got := fixtureRun(context.Background(), args, Options{}, nil); got != 2 {
			t.Fatalf("%q: %d", args, got)
		}
	}
	for _, args := range [][]string{{"probe", "multi word"}, {"probe", "--", "-dash"}, {"probe", "title", "--scalar=a", "--scalar=b", "--items=a", "--items=b"}} {
		opts := Options{OpenService: func(context.Context, Config) (ports.TaskService, func() error, error) {
			return fakeService{}, func() error { return nil }, nil
		}}
		if got := fixtureRun(context.Background(), args, opts, func(ports.TaskService) ([]byte, bool, error) { return nil, false, nil }); got != 0 {
			t.Fatalf("%q: %d", args, got)
		}
	}
}

func TestRoot_Independent(t *testing.T) {
	for range 20 {
		t.Run("independent", func(t *testing.T) {
			t.Parallel()
			var out bytes.Buffer
			if Run(context.Background(), []string{"--version"}, Options{Stdout: &out, Version: "one"}) != 0 || out.String() != "tusk version one\n" {
				t.Fatalf("%q", out.String())
			}
		})
	}
}

func TestRun_FactoryAndCancellation(t *testing.T) {
	for _, tc := range []struct {
		name      string
		service   ports.TaskService
		close     bool
		err       error
		cancel    bool
		wantClose int
	}{
		{name: "factory error", err: errors.New("PRIVATE")},
		{name: "nil service", close: true, wantClose: 1},
		{name: "nil closer", service: fakeService{}},
		{name: "canceled before open", cancel: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if tc.cancel {
				cancel()
			}
			closed := 0
			opts := Options{OpenService: func(context.Context, Config) (ports.TaskService, func() error, error) {
				if tc.cancel {
					t.Fatal("opened after cancel")
				}
				var closeOwner func() error
				if tc.close {
					closeOwner = func() error { closed++; return nil }
				}
				return tc.service, closeOwner, tc.err
			}}
			got := fixtureRun(ctx, []string{"probe", "title"}, opts, func(ports.TaskService) ([]byte, bool, error) { t.Fatal("work admitted"); return nil, false, nil })
			if got != 1 || closed != tc.wantClose {
				t.Fatalf("exit %d close %d", got, closed)
			}
		})
	}
	if got := fixtureRun(context.Background(), []string{"probe", "title"}, Options{}, nil); got != 1 {
		t.Fatalf("missing factory: %d", got)
	}
}

func TestDiagnostic_UnknownPointer(t *testing.T) {
	unknown := ports.NewTransactionError("PRIVATE", ports.ErrConflict)
	code, text := diagnostic(errors.Join(context.Canceled, &unknown), false)
	if code != 1 || !strings.Contains(text, "outcome unknown") || strings.Contains(text, "PRIVATE") {
		t.Fatalf("%d %q", code, text)
	}
	code, text = diagnostic(ports.ErrBusy, false)
	if code != 1 || text != ports.ErrBusy.Error() {
		t.Fatalf("%d %q", code, text)
	}
}

func TestRun_CloseBeforePublish(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var errout bytes.Buffer
	closed := false
	opts := Options{Stdout: checkWriter{t: t, closed: &closed}, Stderr: &errout, OpenService: func(context.Context, Config) (ports.TaskService, func() error, error) {
		return fakeService{}, func() error { closed = true; return nil }, nil
	}}
	got := fixtureRun(ctx, []string{"probe", "title"}, opts, func(ports.TaskService) ([]byte, bool, error) { cancel(); return []byte("confirmed\n"), true, nil })
	if got != 0 || !closed || errout.Len() != 0 {
		t.Fatalf("exit %d closed %v err %q", got, closed, errout.String())
	}
}

type checkWriter struct {
	t      *testing.T
	closed *bool
}

func (w checkWriter) Write(p []byte) (int, error) {
	if !*w.closed {
		w.t.Fatal("published before close")
	}
	return len(p), nil
}
