// Package cli adapts the task service to a fresh, invocation-owned command tree.
package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/newbpydev/tusk/internal/ports"
	"github.com/spf13/cobra"
)

type Options struct {
	Stdout, Stderr io.Writer
	Version        string
	Getenv         func(string) string
	Local          *time.Location
	OpenService    func(context.Context, Config) (ports.TaskService, func() error, error)
	Terminal       func() TerminalFacts
	Confirm        Confirmation
	RunTUI         func(context.Context, Config) (TUIResult, error)
}

type invocation struct {
	options         Options
	root            *cobra.Command
	output          bytes.Buffer
	result          []byte
	autoComplete    bool
	timezone        string
	committed       bool
	configValue     Config
	terminalFacts   TerminalFacts
	terminalSampled bool
	tuiResult       *TUIResult
}

func Run(ctx context.Context, args []string, options Options) int {
	return newInvocation(options).execute(ctx, args)
}

func newInvocation(options Options) *invocation {
	if options.Stdout == nil {
		options.Stdout = io.Discard
	}
	if options.Stderr == nil {
		options.Stderr = io.Discard
	}
	i := &invocation{options: options}
	root := &cobra.Command{Use: "tusk", Short: "Tusk - Zero-friction terminal task management system", SilenceErrors: true, SilenceUsage: true, DisableSuggestions: true, Args: syntaxArgs(cobra.NoArgs), Version: options.Version}
	i.root = root
	root.SetOut(&i.output)
	root.SetErr(&i.output)
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error { return syntaxError{err} })
	root.CompletionOptions.DisableDefaultCmd = true
	root.PersistentFlags().BoolVar(&i.autoComplete, "auto-complete-parent", false, "Complete parents when all children are done")
	root.PersistentFlags().StringVar(&i.timezone, "timezone", "", "Date timezone (IANA name)")
	root.Flags().BoolP("version", "v", false, "Print version")
	// Cobra's built-in version path requires a nonempty Version string.
	if root.Version == "" {
		root.Version = "development"
	}
	root.SetVersionTemplate("tusk version {{.Version}}\n")
	root.RunE = func(c *cobra.Command, _ []string) error { return c.Help() }
	version := &cobra.Command{Use: "version", Short: "Print the version of tusk", Args: syntaxArgs(cobra.NoArgs), RunE: func(*cobra.Command, []string) error {
		fmt.Fprintf(&i.output, "tusk version %s\n", root.Version)
		return nil
	}}
	root.AddCommand(version)
	root.AddCommand(i.addCommand(), i.editCommand(), i.doneCommand(), i.deleteCommand())
	root.AddCommand(i.listCommand(), i.treeCommand(), i.statsCommand(), i.historyCommand())
	root.AddCommand(i.tuiCommand())
	root.SetHelpCommand(&cobra.Command{Use: "help [command]", Short: "Help about any command", RunE: func(c *cobra.Command, args []string) error {
		target, remaining, err := root.Find(args)
		if err != nil || len(remaining) != 0 {
			return syntaxError{err}
		}
		return target.Help()
	}})
	return i
}

func syntaxArgs(validate cobra.PositionalArgs) cobra.PositionalArgs {
	return func(c *cobra.Command, args []string) error {
		if err := validate(c, args); err != nil {
			return syntaxError{err}
		}
		return nil
	}
}

func (i *invocation) execute(ctx context.Context, args []string) int {
	_, _, err := i.root.Find(args)
	if err != nil {
		err = syntaxError{err}
	} else {
		i.root.SetArgs(args)
		_, err = i.root.ExecuteContextC(ctx)
	}
	if err == nil {
		data := i.result
		if data == nil {
			data = i.output.Bytes()
		}
		err = writeAll(i.options.Stdout, data)
	}
	if err == nil {
		return 0
	}
	code, message := diagnostic(err, i.committed)
	if i.tuiResult != nil {
		code, message = tuiDiagnostic(err, *i.tuiResult)
	}
	// A diagnostic write failure is terminal; never recursively report it.
	_ = writeAll(i.options.Stderr, []byte("tusk: "+message+"\n"))
	return code
}

func writeAll(w io.Writer, p []byte) error {
	if len(p) == 0 {
		return nil
	}
	n, err := w.Write(p)
	if err == nil && n != len(p) {
		return io.ErrShortWrite
	}
	return err
}

func (i *invocation) invoke(c *cobra.Command, work func(ports.TaskService) ([]byte, bool, error)) error {
	cfg, err := i.config(c)
	if err != nil {
		return err
	}
	i.configValue = cfg
	if err = c.Context().Err(); err != nil {
		return err
	}
	if i.options.OpenService == nil {
		return ports.ErrInvalidServiceOptions
	}
	svc, closeOwner, err := i.options.OpenService(c.Context(), cfg)
	// On factory error the factory retains responsibility for partial resources.
	if err != nil {
		return err
	}
	if svc == nil || closeOwner == nil {
		if closeOwner != nil {
			_ = closeOwner()
		}
		return ports.ErrInvalidServiceOptions
	}
	data, committed, workErr := work(svc)
	// Work reports confirmed mutation success independently of serialization.
	i.committed = committed
	closeErr := closeOwner()
	if workErr != nil || closeErr != nil {
		return errors.Join(workErr, closeErr)
	}
	i.result = data
	return nil
}
