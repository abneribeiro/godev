// Package errors provides custom error types and utilities for the godev application
package errors

import (
	"errors"
	"fmt"
)

// Error types for different components
var (
	ErrInvalidConfig      = errors.New("invalid configuration")
	ErrNetworkTimeout     = errors.New("network timeout")
	ErrDatabaseConnection = errors.New("database connection failed")
	ErrInvalidQuery       = errors.New("invalid query")
	ErrPermissionDenied   = errors.New("permission denied")
	ErrResourceNotFound   = errors.New("resource not found")
	ErrInvalidRequest     = errors.New("invalid request")
)

// AppError represents an application-specific error with context
type AppError struct {
	Type      string
	Message   string
	Cause     error
	Operation string
	Context   map[string]interface{}
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

// NewAppError creates a new application error
func NewAppError(errorType, message, operation string, cause error) *AppError {
	return &AppError{
		Type:      errorType,
		Message:   message,
		Cause:     cause,
		Operation: operation,
		Context:   make(map[string]interface{}),
	}
}

// WithContext adds context to an error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	e.Context[key] = value
	return e
}

// HTTP error types
func NewHTTPError(message string, cause error) *AppError {
	return NewAppError("HTTP_ERROR", message, "http_request", cause)
}

// Database error types
func NewDatabaseError(message string, cause error) *AppError {
	return NewAppError("DATABASE_ERROR", message, "database_operation", cause)
}

// Storage error types
func NewStorageError(message string, cause error) *AppError {
	return NewAppError("STORAGE_ERROR", message, "storage_operation", cause)
}

// Configuration error types
func NewConfigError(message string, cause error) *AppError {
	return NewAppError("CONFIG_ERROR", message, "configuration", cause)
}

// Validation error types
func NewValidationError(message string, cause error) *AppError {
	return NewAppError("VALIDATION_ERROR", message, "validation", cause)
}

// IsType checks if an error is of a specific type
func IsType(err error, errorType string) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == errorType
	}
	return false
}

// IsHTTPError checks if an error is an HTTP error
func IsHTTPError(err error) bool {
	return IsType(err, "HTTP_ERROR")
}

// IsDatabaseError checks if an error is a database error
func IsDatabaseError(err error) bool {
	return IsType(err, "DATABASE_ERROR")
}

// IsStorageError checks if an error is a storage error
func IsStorageError(err error) bool {
	return IsType(err, "STORAGE_ERROR")
}

// Wrap wraps an error with additional context
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// Wrapf wraps an error with a formatted message
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(format+": %w", append(args, err)...)
}
