package core_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/newbpydev/tusk/internal/core"
)

func TestErrorsExist(t *testing.T) {
	sentinels := []struct {
		name string
		err  error
		msg  string
	}{
		{"ErrTaskNotFound", core.ErrTaskNotFound, "task not found"},
		{"ErrEmptyTitle", core.ErrEmptyTitle, "task title cannot be empty"},
		{"ErrTitleTooLong", core.ErrTitleTooLong, "task title exceeds maximum length of 255 characters"},
		{"ErrInvalidStatus", core.ErrInvalidStatus, "invalid task status"},
		{"ErrInvalidPriority", core.ErrInvalidPriority, "invalid task priority"},
		{"ErrInvalidStatusTransition", core.ErrInvalidStatusTransition, "invalid status transition"},
		{"ErrCyclicDependency", core.ErrCyclicDependency, "cyclic dependency detected: a task cannot be its own ancestor"},
		{"ErrSelfParenting", core.ErrSelfParenting, "task cannot reference itself as parent"},
		{"ErrMaxDepthExceeded", core.ErrMaxDepthExceeded, "maximum subtask hierarchy depth exceeded"},
		{"ErrInvalidTag", core.ErrInvalidTag, "invalid tag format: tags must be alphanumeric with hyphens"},
		{"ErrInvalidProgress", core.ErrInvalidProgress, "task progress must be an integer between 0 and 100"},
		{"ErrInvalidTaskID", core.ErrInvalidTaskID, "invalid task id: id cannot be empty"},
		{"ErrDuplicateTaskID", core.ErrDuplicateTaskID, "duplicate task id in hierarchy"},
		{"ErrInvalidDepth", core.ErrInvalidDepth, "invalid hierarchy depth: depth cannot be negative"},
		{"ErrTraversalLimitExceeded", core.ErrTraversalLimitExceeded, "hierarchy traversal limit exceeded"},
	}

	seen := make(map[error]string)
	for _, tc := range sentinels {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err == nil {
				t.Fatalf("expected %s to be non-nil", tc.name)
			}
			if tc.err.Error() != tc.msg {
				t.Errorf("%s message = %q, want %q", tc.name, tc.err.Error(), tc.msg)
			}
			if other, exists := seen[tc.err]; exists {
				t.Errorf("%s and %s share identical error instance", tc.name, other)
			}
			seen[tc.err] = tc.name
		})
	}
}

func TestErrors_SentinelIntegrity(t *testing.T) {
	sentinels := []error{
		core.ErrTaskNotFound,
		core.ErrEmptyTitle,
		core.ErrTitleTooLong,
		core.ErrInvalidStatus,
		core.ErrInvalidPriority,
		core.ErrInvalidStatusTransition,
		core.ErrCyclicDependency,
		core.ErrSelfParenting,
		core.ErrMaxDepthExceeded,
		core.ErrInvalidTag,
		core.ErrInvalidProgress,
		core.ErrInvalidTaskID,
		core.ErrDuplicateTaskID,
		core.ErrInvalidDepth,
		core.ErrTraversalLimitExceeded,
	}

	for _, s := range sentinels {
		t.Run(s.Error(), func(t *testing.T) {
			if !errors.Is(s, s) {
				t.Errorf("errors.Is(s, s) failed for %v", s)
			}
			wrapped := fmt.Errorf("wrapped context: %w", s)
			if !errors.Is(wrapped, s) {
				t.Errorf("errors.Is(wrapped, s) failed for %v", s)
			}
		})
	}
}
