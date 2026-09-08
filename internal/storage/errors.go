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
	for _, sentinel := range []error{
		context.Canceled, context.DeadlineExceeded,
		core.ErrTaskNotFound, core.ErrDuplicateTaskID, core.ErrInvalidTaskID, core.ErrEmptyTitle, core.ErrTitleTooLong, core.ErrInvalidStatus, core.ErrInvalidPriority, core.ErrInvalidProgress, core.ErrInvalidTag, core.ErrSelfParenting, core.ErrCyclicDependency, core.ErrMaxDepthExceeded,
		ports.ErrBusy, ports.ErrCorrupt, ports.ErrReadOnly, ports.ErrStorage, ports.ErrInvalidRecord, ports.ErrTransactionInUse, ports.ErrTransactionClosed, ports.ErrChildrenPresent,
	} {
		if errors.Is(err, sentinel) {
			return sentinel
		}
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

func openCause(err error) error {
	if errors.Is(err, errIncompatibleSchema) {
		return ports.ErrIncompatibleSchema
	}
	if errors.Is(err, errCorruptSchema) {
		return ports.ErrCorrupt
	}
	if errors.Is(err, errMigrationOutcome) {
		return ports.NewTransactionError("migration", ports.ErrStorage)
	}
	return storageCause(err)
}
