package ports

import (
	"errors"
	"strings"
	"testing"
)

func TestTransactionError_OptionalCause(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  TransactionError
		want error
	}{
		{"missing", NewTransactionError("commit", nil), ErrStorage},
		{"zero", TransactionError{}, ErrStorage},
		{"known", NewTransactionError("rollback", ErrCorrupt), ErrCorrupt},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(tc.err.Error(), tc.want.Error()) || !errors.Is(tc.err, tc.want) || tc.err.Outcome() != "unknown" {
				t.Fatalf("invalid transaction error: %v", tc.err)
			}
			if errors.Unwrap(tc.err) != nil {
				t.Fatal("cause is exposed by Unwrap")
			}
		})
	}
}
