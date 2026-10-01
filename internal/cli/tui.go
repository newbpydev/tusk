package cli

import (
	"errors"

	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"github.com/spf13/cobra"
)

// TUIResult preserves session-wide receipts after terminal restoration.
type TUIResult struct{ HadCommittedChanges, OutcomeUnknown bool }

const errTUITerminal core.Error = "tui requires terminal input and output; use tusk list --json for pipes"

func (i *invocation) tuiCommand() *cobra.Command {
	return &cobra.Command{Use: "tui", Short: "Open the interactive task application", Args: syntaxArgs(cobra.NoArgs), RunE: func(c *cobra.Command, _ []string) error {
		cfg, err := i.config(c)
		if err != nil {
			return err
		}
		facts := i.terminal()
		if !facts.In || !facts.Out || (i.options.Getenv != nil && i.options.Getenv("TERM") == "dumb") {
			return errTUITerminal
		}
		if err = c.Context().Err(); err != nil {
			return err
		}
		if i.options.RunTUI == nil {
			return ports.ErrInvalidServiceOptions
		}
		result, err := i.options.RunTUI(c.Context(), cfg)
		i.tuiResult = &result
		return err
	}}
}

func tuiDiagnostic(err error, result TUIResult) (int, string) {
	code, message := diagnostic(err, false)
	var value ports.TransactionError
	var pointer *ports.TransactionError
	if result.OutcomeUnknown || errors.As(err, &value) || errors.As(err, &pointer) {
		code, message = 1, "outcome unknown; reopen and read tasks and history before considering another mutation"
	}
	if result.HadCommittedChanges {
		message += "; earlier changes remain committed"
	}
	return code, message
}
