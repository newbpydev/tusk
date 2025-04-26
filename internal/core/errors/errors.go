package errors

import (
	stderrors "errors"

	pkgerrors "github.com/pkg/errors"
)

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
	Code    ErrorCode
	Message string
	Err     error
}

// Error implements the error interface for DomainError.
func (e DomainError) Error() string {
	if e.Err != nil {
		return string(e.Code) + ": " + e.Message + ": " + e.Err.Error()
	}
	return string(e.Code) + ": " + e.Message
}

// Unwrap returns the underlying error if any.
func (e DomainError) Unwrap() error {
	return e.Err
}

// Helper functions

// NotFound creates a new DomainError with the NotFound code.
// This is used when a resource is not found in the system.
func NotFound(msg string, err ...error) error {
	var wrapped error
	if len(err) > 0 {
		wrapped = err[0]
	}
	return DomainError{Code: CodeNotFound, Message: msg, Err: wrapped}
}

// Conflict creates a new DomainError with the Conflict code.
// This is used when there is a conflict in the system, such as a duplicate resource.
func Conflict(msg string, err ...error) error {
	var wrapped error
	if len(err) > 0 {
		wrapped = err[0]
	}
	return DomainError{Code: CodeConflict, Message: msg, Err: wrapped}
}

// InvalidInput creates a new DomainError with the InvalidInput code.
// This is used when the input provided to the system is invalid or malformed.
func InvalidInput(msg string, err ...error) error {
	var wrapped error
	if len(err) > 0 {
		wrapped = err[0]
	}
	return DomainError{Code: CodeInvalidInput, Message: msg, Err: wrapped}
}

// Unauthorized creates a new DomainError with the Unauthorized code.
// This is used when a user is not authenticated to perform an action.
func Unauthorized(msg string, err ...error) error {
	var wrapped error
	if len(err) > 0 {
		wrapped = err[0]
	}
	return DomainError{Code: CodeUnauthorized, Message: msg, Err: wrapped}
}

// InternalError creates a new DomainError with the InternalError code.
// This is used when an unexpected error occurs within the system.
func InternalError(msg string, err ...error) error {
	var wrapped error
	if len(err) > 0 {
		wrapped = err[0]
	}
	return DomainError{Code: CodeInternalError, Message: msg, Err: wrapped}
}

// Forbidden creates a new DomainError with the Forbidden code.
// This is used when a user is authenticated but not authorized to perform an action.
func Forbidden(msg string, err ...error) error {
	var wrapped error
	if len(err) > 0 {
		wrapped = err[0]
	}
	return DomainError{Code: CodeForbidden, Message: msg, Err: wrapped}
}

// Wrap wraps an error with a message and returns a new error.
// This function uses pkg/errors.Wrap for enhanced stack tracing.
func Wrap(err error, msg string) error {
	return pkgerrors.Wrap(err, msg)
}

// Wrapf wraps an error with a formatted message and returns a new error.
// This function uses pkg/errors.Wrapf for enhanced stack tracing.
func Wrapf(err error, format string, args ...interface{}) error {
	return pkgerrors.Wrapf(err, format, args...)
}

// WithStack annotates an error with a stack trace at the point WithStack was called.
// This function uses pkg/errors.WithStack for enhanced stack tracing.
func WithStack(err error) error {
	return pkgerrors.WithStack(err)
}

// Cause returns the underlying cause of the error, if possible.
// This function uses pkg/errors.Cause to unwrap errors.
func Cause(err error) error {
	return pkgerrors.Cause(err)
}

// Error type checking functions

// IsDomainError checks if the given error is a DomainError.
func IsDomainError(err error) bool {
	var domainErr DomainError
	return stderrors.As(err, &domainErr)
}

// IsNotFound checks if the given error is a NotFound error.
func IsNotFound(err error) bool {
	var domainErr DomainError
	if stderrors.As(err, &domainErr) {
		return domainErr.Code == CodeNotFound
	}
	return false
}

// IsConflict checks if the given error is a Conflict error.
func IsConflict(err error) bool {
	var domainErr DomainError
	if stderrors.As(err, &domainErr) {
		return domainErr.Code == CodeConflict
	}
	return false
}

// IsInvalidInput checks if the given error is an InvalidInput error.
func IsInvalidInput(err error) bool {
	var domainErr DomainError
	if stderrors.As(err, &domainErr) {
		return domainErr.Code == CodeInvalidInput
	}
	return false
}

// IsUnauthorized checks if the given error is an Unauthorized error.
func IsUnauthorized(err error) bool {
	var domainErr DomainError
	if stderrors.As(err, &domainErr) {
		return domainErr.Code == CodeUnauthorized
	}
	return false
}

// IsInternalError checks if the given error is an InternalError error.
func IsInternalError(err error) bool {
	var domainErr DomainError
	if stderrors.As(err, &domainErr) {
		return domainErr.Code == CodeInternalError
	}
	return false
}

// IsForbidden checks if the given error is a Forbidden error.
func IsForbidden(err error) bool {
	var domainErr DomainError
	if stderrors.As(err, &domainErr) {
		return domainErr.Code == CodeForbidden
	}
	return false
}
