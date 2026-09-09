package ports

import "errors"

type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrInvalidRecord      Error = "storage: invalid record"
	ErrCorrupt            Error = "storage: corrupt data"
	ErrIncompatibleSchema Error = "storage: incompatible database schema"
	ErrBusy               Error = "storage: database busy"
	ErrStorage            Error = "storage: operation failed"
	ErrReadOnly           Error = "storage: read-only operation"
	ErrClosedRepository   Error = "storage: repository closed"
	ErrInvalidCallback    Error = "storage: callback is nil"
	ErrNestedTransaction  Error = "storage: nested transaction"
	ErrTransactionClosed  Error = "storage: transaction closed"
	ErrTransactionInUse   Error = "storage: transaction already in use"
	ErrChildrenPresent    Error = "storage: task has children"
)

// TransactionError reports an unconfirmed commit/cleanup outcome. Its cause is
// a sanitized domain/context category, matched through errors.Is. It deliberately
// has no Unwrap method: errors.As must not expose a concrete driver cause.
type TransactionError struct {
	operation string
	cause     error
}

// NewTransactionError requires an already sanitized cause. A nil cause, like the
// zero value of TransactionError, matches ErrStorage.
func NewTransactionError(operation string, cause error) TransactionError {
	return TransactionError{operation, cause}
}
func (e TransactionError) category() error {
	if e.cause == nil {
		return ErrStorage
	}
	return e.cause
}
func (e TransactionError) Error() string {
	return "storage: " + e.operation + " outcome unknown (" + e.category().Error() + ")"
}
func (e TransactionError) Is(target error) bool { return errors.Is(e.category(), target) }
func (e TransactionError) Outcome() string      { return "unknown" }
