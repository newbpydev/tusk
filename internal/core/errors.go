package core

import "errors"

var (
	ErrTaskNotFound            = errors.New("task not found")
	ErrEmptyTitle              = errors.New("task title cannot be empty")
	ErrTitleTooLong            = errors.New("task title exceeds maximum length of 255 characters")
	ErrInvalidStatus           = errors.New("invalid task status")
	ErrInvalidPriority         = errors.New("invalid task priority")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrCyclicDependency        = errors.New("cyclic dependency detected: a task cannot be its own ancestor")
	ErrSelfParenting           = errors.New("task cannot reference itself as parent")
	ErrMaxDepthExceeded        = errors.New("maximum subtask hierarchy depth exceeded")
	ErrInvalidTag              = errors.New("invalid tag format: tags must be alphanumeric with hyphens")
	ErrInvalidProgress         = errors.New("task progress must be an integer between 0 and 100")
)
