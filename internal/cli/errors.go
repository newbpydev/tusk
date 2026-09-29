package cli

import (
	"context"
	"errors"

	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
)

const (
	errAutoCompleteConfiguration core.Error = "invalid TUSK_AUTO_COMPLETE_PARENT; use true or false"
	errTimezoneConfiguration     core.Error = "invalid timezone; use an IANA zone such as UTC; --timezone overrides TUSK_TIMEZONE"
)

type syntaxError struct{ cause error }

func (e syntaxError) Error() string { return "invalid invocation" }
func (e syntaxError) Unwrap() error { return e.cause }

func diagnostic(err error, committed bool) (int, string) {
	var syntax syntaxError
	if errors.As(err, &syntax) {
		return 2, "invalid invocation\nUsage: tusk [command] [flags]\nRun tusk help for usage."
	}
	if committed {
		return 1, "change committed, but output or cleanup failed; read tasks and history before retrying"
	}
	var value ports.TransactionError
	var pointer *ports.TransactionError
	if errors.As(err, &value) || errors.As(err, &pointer) {
		return 1, "outcome unknown; reopen and read tasks (list --all --json) and history before considering another mutation"
	}
	if errors.Is(err, ports.ErrChildrenPresent) {
		return 1, "task has children; use --recursive to delete the whole subtree"
	}
	// Only trusted constants cross the diagnostic boundary, never wrapper text.
	for _, safe := range []error{
		context.Canceled, context.DeadlineExceeded, errAutoCompleteConfiguration, errTimezoneConfiguration, errForceRequired,
		core.ErrTaskNotFound, core.ErrEmptyTitle, core.ErrTitleTooLong, core.ErrInvalidStatus, core.ErrInvalidPriority, core.ErrInvalidStatusTransition, core.ErrSelfParenting, core.ErrCyclicDependency, core.ErrMaxDepthExceeded, core.ErrInvalidTag, core.ErrInvalidProgress, core.ErrInvalidTaskID, core.ErrDuplicateTaskID, core.ErrInvalidDepth,
		ports.ErrInvalidRecord, ports.ErrCorrupt, ports.ErrIncompatibleSchema, ports.ErrBusy, ports.ErrStorage, ports.ErrReadOnly, ports.ErrClosedRepository, ports.ErrInvalidCallback, ports.ErrNestedTransaction, ports.ErrTransactionClosed, ports.ErrTransactionInUse,
		ports.ErrInvalidCommand, ports.ErrInvalidText, ports.ErrInvalidDate, ports.ErrInvalidReferenceTime, ports.ErrIdentityGeneration, ports.ErrConflict, ports.ErrConfirmationRequired, ports.ErrInvalidServiceOptions,
	} {
		if errors.Is(err, safe) {
			return 1, safe.Error()
		}
	}
	return 1, "operation failed"
}
