package errors

import "fmt"

// ErrorCode represents the error code for a domain error.
// It is a string that identifies the type of error that occurred.
type ErrorCode string

const (
	// CodeNotFound is used when a requested resource cannot be found
	CodeNotFound ErrorCode = "NOTE_FOUND"

	// CodeConflict is used when there is a conflict with existing resources
	CodeConflict ErrorCode = "NOTE_CONFLICT"

	// CodeInvalidInput is used when the provided input is invalid or malformed
	CodeInvalidInput ErrorCode = "INVALID_INPUT"

	// CodeUnauthorized is used when authentication is required but not provided
	CodeUnauthorized ErrorCode = "UNAUTHORIZED"

	// CodeInternalError is used when an unexpected internal error occurs
	CodeInternalError ErrorCode = "INTERNAL_ERROR"

	// CodeForbidden is used when the authenticated user lacks necessary permissions
	CodeForbidden ErrorCode = "FORBIDDEN"
)

// DomainError represents a domain-specific error with a code and message.
type DomainError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Err     error     `json:"-"`
}

// Error implements the error interface for DomainError.
func (e DomainError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error if any.
func (e DomainError) Unwrap() error {
	return e.Err
}

// Helper functions

// NotFound creates a new DomainError with the NotFound code.
// This is used when a resource is not found in the system.
func NotFound(msg string) error {
	return DomainError{Code: CodeNotFound, Message: msg}
}

// Conflict creates a new DomainError with the Conflict code.
// This is used when there is a conflict in the system, such as a duplicate resource.
func Conflict(msg string) error {
	return DomainError{Code: CodeConflict, Message: msg}
}

// InvalidInput creates a new DomainError with the InvalidInput code.
// This is used when the input provided to the system is invalid or malformed.
func InvalidInput(msg string) error {
	return DomainError{Code: CodeInvalidInput, Message: msg}
}

// Unauthorized creates a new DomainError with the Unauthorized code.
// This is used when a user is not authenticated to perform an action.
func Unauthorized(msg string) error {
	return DomainError{Code: CodeUnauthorized, Message: msg}
}

// InternalError creates a new DomainError with the InternalError code.
// This is used when an unexpected error occurs within the system.
func InternalError(msg string) error {
	return DomainError{Code: CodeInternalError, Message: msg}
}

// Forbidden creates a new DomainError with the Forbidden code.
// This is used when a user is authenticated but not authorized to perform an action.
func Forbidden(msg string) error {
	return DomainError{Code: CodeForbidden, Message: msg}
}
