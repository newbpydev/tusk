package storage

import (
	"context"
	"database/sql"
	"errors"
	"github.com/newbpydev/tusk/internal/ports"
	"strings"
	"testing"
	"time"
)

type canceledLedgerReader struct{ *sql.DB }

func (r canceledLedgerReader) QueryContext(ctx context.Context, q string, args ...any) (*sql.Rows, error) {
	if strings.HasPrefix(q, "SELECT version, filename") {
		return nil, context.Canceled
	}
	return r.DB.QueryContext(ctx, q, args...)
}
func TestInspection_PreservesCancellationCause(t *testing.T) {
	db, inv := migrationFixture(t)
	if err := migrate(context.Background(), db, inv); err != nil {
		t.Fatal(err)
	}
	t.Run("ledger", func(t *testing.T) {
		_, err := inspectSchema(context.Background(), canceledLedgerReader{db}, inv)
		if !errors.Is(openCause(err), context.Canceled) {
			t.Fatalf("ledger cancellation mapped to %v", openCause(err))
		}
	})
	t.Run("deadline", func(t *testing.T) {
		ctx, cancel := context.WithDeadline(context.Background(), time.Unix(0, 0))
		defer cancel()
		_, err := expectedSchema(ctx, inv)
		if !errors.Is(openCause(err), context.DeadlineExceeded) {
			t.Fatalf("deadline mapped to %v", openCause(err))
		}
	})
	t.Run("canonical replay", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := expectedSchema(ctx, inv)
		if !errors.Is(openCause(err), context.Canceled) {
			t.Fatalf("canonical replay cancellation mapped to %v", openCause(err))
		}
	})
}

func TestOpenCause_RetainsMigrationUnknownOutcome(t *testing.T) {
	for cause, want := range map[error]error{errCorruptSchema: ports.ErrCorrupt, errIncompatibleSchema: ports.ErrIncompatibleSchema, context.Canceled: context.Canceled, context.DeadlineExceeded: context.DeadlineExceeded} {
		err := openCause(errors.Join(cause, errMigrationOutcome))
		var unknown interface{ Outcome() string }
		if !errors.As(err, &unknown) || unknown.Outcome() != "unknown" || !errors.Is(err, want) {
			t.Fatalf("unknown cleanup outcome hidden by %v", err)
		}
	}
}
