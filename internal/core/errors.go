package core

// Error represents an immutable domain error sentinel.
type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	ErrTaskNotFound            = Error("task not found")
	ErrEmptyTitle              = Error("task title cannot be empty")
	ErrTitleTooLong            = Error("task title exceeds maximum length of 255 characters")
	ErrInvalidStatus           = Error("invalid task status")
	ErrInvalidPriority         = Error("invalid task priority")
	ErrInvalidStatusTransition = Error("invalid status transition")
	ErrCyclicDependency        = Error("cyclic dependency detected: a task cannot be its own ancestor")
	ErrSelfParenting           = selfParentingError("task cannot reference itself as parent")
	ErrMaxDepthExceeded        = Error("maximum subtask hierarchy depth exceeded")
	ErrInvalidTag              = Error("invalid tag format: tags must be alphanumeric with hyphens")
	ErrInvalidProgress         = Error("task progress must be an integer between 0 and 100")
	ErrInvalidTaskID           = Error("invalid task id: id cannot be empty")
	ErrDuplicateTaskID         = Error("duplicate task id in hierarchy")
	ErrInvalidDepth            = Error("invalid hierarchy depth: depth cannot be negative")
	ErrTraversalLimitExceeded  = Error("hierarchy traversal limit exceeded")
)

type selfParentingError string

func (e selfParentingError) Error() string {
	return string(e)
}

func (e selfParentingError) Is(target error) bool {
	return target == ErrCyclicDependency || target == ErrSelfParenting
}
