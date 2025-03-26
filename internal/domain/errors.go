package domain

import "fmt"

// Common domain errors
var (
	ErrNotFound       = NewError("not found", 404)
	ErrUnauthorized   = NewError("unauthorized", 401)
	ErrInvalidInput   = NewError("invalid input", 400)
	ErrInternal       = NewError("internal server error", 500)
	ErrNotImplemented = NewError("not implemented", 501)
)

// Error represents a domain error with an associated code.
type Error struct {
	Message string
	Code    int
}

// Error returns the error message.
func (e *Error) Error() string {
	return e.Message
}

// NewError creates a new domain error with the given message and code.
func NewError(message string, code int) *Error {
	return &Error{
		Message: message,
		Code:    code,
	}
}

// ResourceNotFoundError indicates that a requested resource was not found.
type ResourceNotFoundError struct {
	URI string
	Err *Error
}

// Error returns the error message.
func (e *ResourceNotFoundError) Error() string {
	return e.Err.Error()
}

// NewResourceNotFoundError creates a new ResourceNotFoundError.
func NewResourceNotFoundError(uri string) *ResourceNotFoundError {
	return &ResourceNotFoundError{
		URI: uri,
		Err: NewError(
			fmt.Sprintf("resource with URI %s not found", uri),
			404,
		),
	}
}

// ToolNotFoundError indicates that a requested tool was not found.
type ToolNotFoundError struct {
	Name string
	Err  *Error
}

// Error returns the error message.
func (e *ToolNotFoundError) Error() string {
	return e.Err.Error()
}

// NewToolNotFoundError creates a new ToolNotFoundError.
func NewToolNotFoundError(name string) *ToolNotFoundError {
	return &ToolNotFoundError{
		Name: name,
		Err: NewError(
			fmt.Sprintf("tool with name %s not found", name),
			404,
		),
	}
}

// PromptNotFoundError indicates that a requested prompt was not found.
type PromptNotFoundError struct {
	Name string
	Err  *Error
}

// Error returns the error message.
func (e *PromptNotFoundError) Error() string {
	return e.Err.Error()
}

// NewPromptNotFoundError creates a new PromptNotFoundError.
func NewPromptNotFoundError(name string) *PromptNotFoundError {
	return &PromptNotFoundError{
		Name: name,
		Err: NewError(
			fmt.Sprintf("prompt with name %s not found", name),
			404,
		),
	}
}

// SessionNotFoundError indicates that a requested session was not found.
type SessionNotFoundError struct {
	ID  string
	Err *Error
}

// Error returns the error message.
func (e *SessionNotFoundError) Error() string {
	return e.Err.Error()
}

// NewSessionNotFoundError creates a new SessionNotFoundError.
func NewSessionNotFoundError(id string) *SessionNotFoundError {
	return &SessionNotFoundError{
		ID: id,
		Err: NewError(
			fmt.Sprintf("session with ID %s not found", id),
			404,
		),
	}
}

// ValidationError indicates that input validation failed.
type ValidationError struct {
	Field   string
	Message string
	Err     *Error
}

// Error returns the error message.
func (e *ValidationError) Error() string {
	return e.Err.Error()
}

// NewValidationError creates a new ValidationError.
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
		Err: NewError(
			fmt.Sprintf("validation failed for field %s: %s", field, message),
			400,
		),
	}
}
