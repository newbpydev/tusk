package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/newbpydev/tusk/internal/core"
	"github.com/newbpydev/tusk/internal/ports"
	"modernc.org/sqlite"
)

func storageCause(err error) error {
	if err == nil {
		return nil
	}
	var safe error
	for _, sentinel := range []error{
		context.Canceled, context.DeadlineExceeded,
		core.ErrTaskNotFound, core.ErrDuplicateTaskID, core.ErrInvalidTaskID, core.ErrEmptyTitle, core.ErrTitleTooLong, core.ErrInvalidStatus, core.ErrInvalidPriority, core.ErrInvalidStatusTransition, core.ErrInvalidProgress, core.ErrInvalidTag, core.ErrSelfParenting, core.ErrCyclicDependency, core.ErrMaxDepthExceeded, core.ErrInvalidDepth,
		ports.ErrBusy, ports.ErrCorrupt, ports.ErrReadOnly, ports.ErrStorage, ports.ErrInvalidRecord, ports.ErrTransactionInUse, ports.ErrTransactionClosed, ports.ErrChildrenPresent, ports.ErrIncompatibleSchema, ports.ErrClosedRepository, ports.ErrInvalidCallback, ports.ErrNestedTransaction,
	} {
		if errors.Is(err, sentinel) {
			if safe == nil {
				safe = sentinel
			} else {
				safe = errors.Join(safe, sentinel)
			}
		}
	}
	if safe != nil {
		return safe
	}
	if errors.Is(err, sql.ErrNoRows) {
		return core.ErrTaskNotFound
	}
	var driverErr *sqlite.Error
	if errors.As(err, &driverErr) {
		switch driverErr.Code() {
		case 1555, 2067:
			return core.ErrDuplicateTaskID
		case 787:
			return core.ErrTaskNotFound
		}
		switch driverErr.Code() & 255 {
		case 5, 6:
			return ports.ErrBusy
		case 8:
			return ports.ErrReadOnly
		case 11, 26:
			return ports.ErrCorrupt
		}
	}
	return ports.ErrStorage
}

func corruptCause(err error) error {
	for _, sentinel := range []error{core.ErrInvalidTaskID, core.ErrSelfParenting, core.ErrCyclicDependency, core.ErrMaxDepthExceeded, core.ErrTaskNotFound, core.ErrDuplicateTaskID} {
		if errors.Is(err, sentinel) {
			return errors.Join(ports.ErrCorrupt, sentinel)
		}
	}
	return ports.ErrCorrupt
}

// schemaFailure preserves cancellation while classifying actual schema faults.
func schemaFailure(err, category error) error {
	for _, cause := range []error{context.Canceled, context.DeadlineExceeded} {
		if errors.Is(err, cause) {
			return cause
		}
	}
	return category
}

func openCause(err error) error {
	cause := storageCause(err)
	if errors.Is(err, errIncompatibleSchema) {
		cause = schemaFailure(err, ports.ErrIncompatibleSchema)
	} else if errors.Is(err, errCorruptSchema) {
		cause = schemaFailure(err, ports.ErrCorrupt)
	}
	if errors.Is(err, errMigrationOutcome) {
		return ports.NewTransactionError("migration", cause)
	}
	return cause
}
