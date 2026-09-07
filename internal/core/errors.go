package core

import "errors"

var (
	ErrTaskNotFound                  = errors.New("task not found")
	ErrEmptyTitle                    = errors.New("task title cannot be empty")
	ErrTitleTooLong                  = errors.New("task title exceeds maximum length of 255 characters")
	ErrInvalidStatus                 = errors.New("invalid task status")
	ErrInvalidPriority               = errors.New("invalid task priority")
	ErrInvalidStatusTransition       = errors.New("invalid status transition")
	ErrCyclicDependency              = errors.New("cyclic dependency detected: a task cannot be its own ancestor")
	ErrSelfParenting           error = selfParentingError{}
	ErrMaxDepthExceeded              = errors.New("maximum subtask hierarchy depth exceeded")
	ErrInvalidTag                    = errors.New("invalid tag format: tags must be alphanumeric with hyphens")
	ErrInvalidProgress               = errors.New("task progress must be an integer between 0 and 100")
	ErrInvalidTaskID                 = errors.New("invalid task id: id cannot be empty")
	ErrDuplicateTaskID               = errors.New("duplicate task id in hierarchy")
	ErrInvalidDepth                  = errors.New("invalid hierarchy depth: depth cannot be negative")
	ErrTraversalLimitExceeded        = errors.New("hierarchy traversal limit exceeded")
)

type selfParentingError struct{}

func (e selfParentingError) Error() string {
	return "task cannot reference itself as parent"
}

func (e selfParentingError) Is(target error) bool {
	return target == ErrCyclicDependency || target == ErrSelfParenting
}
