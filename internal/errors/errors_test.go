package errors

import (
	"errors"
	"testing"
)

func TestAppError(t *testing.T) {
	tests := []struct {
		name          string
		errorType     string
		message       string
		operation     string
		cause         error
		expectedError string
	}{
		{
			name:          "error without cause",
			errorType:     "TEST_ERROR",
			message:       "test message",
			operation:     "test_operation",
			cause:         nil,
			expectedError: "TEST_ERROR: test message",
		},
		{
			name:          "error with cause",
			errorType:     "TEST_ERROR",
			message:       "test message",
			operation:     "test_operation",
			cause:         errors.New("underlying error"),
			expectedError: "TEST_ERROR: test message (caused by: underlying error)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAppError(tt.errorType, tt.message, tt.operation, tt.cause)

			if err.Error() != tt.expectedError {
				t.Errorf("Expected error message %q, got %q", tt.expectedError, err.Error())
			}

			if err.Type != tt.errorType {
				t.Errorf("Expected error type %q, got %q", tt.errorType, err.Type)
			}

			if err.Message != tt.message {
				t.Errorf("Expected error message %q, got %q", tt.message, err.Message)
			}

			if err.Operation != tt.operation {
				t.Errorf("Expected operation %q, got %q", tt.operation, err.Operation)
			}

			// Test Unwrap
			if tt.cause != nil {
				if !errors.Is(err, tt.cause) {
					t.Error("Expected error to wrap the cause")
				}
			}
		})
	}
}

func TestWithContext(t *testing.T) {
	err := NewAppError("TEST_ERROR", "test message", "test_operation", nil)
	err.WithContext("key1", "value1").WithContext("key2", 42)

	if err.Context["key1"] != "value1" {
		t.Errorf("Expected context key1 to be 'value1', got %v", err.Context["key1"])
	}

	if err.Context["key2"] != 42 {
		t.Errorf("Expected context key2 to be 42, got %v", err.Context["key2"])
	}
}

func TestErrorTypeCheckers(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		isHTTP    bool
		isDB      bool
		isStorage bool
	}{
		{
			name:      "HTTP error",
			err:       NewHTTPError("test", nil),
			isHTTP:    true,
			isDB:      false,
			isStorage: false,
		},
		{
			name:      "Database error",
			err:       NewDatabaseError("test", nil),
			isHTTP:    false,
			isDB:      true,
			isStorage: false,
		},
		{
			name:      "Storage error",
			err:       NewStorageError("test", nil),
			isHTTP:    false,
			isDB:      false,
			isStorage: true,
		},
		{
			name:      "Regular error",
			err:       errors.New("regular error"),
			isHTTP:    false,
			isDB:      false,
			isStorage: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if IsHTTPError(tt.err) != tt.isHTTP {
				t.Errorf("IsHTTPError() = %v, expected %v", IsHTTPError(tt.err), tt.isHTTP)
			}

			if IsDatabaseError(tt.err) != tt.isDB {
				t.Errorf("IsDatabaseError() = %v, expected %v", IsDatabaseError(tt.err), tt.isDB)
			}

			if IsStorageError(tt.err) != tt.isStorage {
				t.Errorf("IsStorageError() = %v, expected %v", IsStorageError(tt.err), tt.isStorage)
			}
		})
	}
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")

	tests := []struct {
		name     string
		err      error
		message  string
		expected string
		isNil    bool
	}{
		{
			name:     "wrap error",
			err:      originalErr,
			message:  "wrapped",
			expected: "wrapped: original error",
			isNil:    false,
		},
		{
			name:     "wrap nil error",
			err:      nil,
			message:  "wrapped",
			expected: "",
			isNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Wrap(tt.err, tt.message)

			if tt.isNil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
			} else {
				if result == nil {
					t.Fatal("Expected non-nil error")
				}

				if result.Error() != tt.expected {
					t.Errorf("Expected %q, got %q", tt.expected, result.Error())
				}

				if !errors.Is(result, originalErr) {
					t.Error("Expected wrapped error to contain original error")
				}
			}
		})
	}
}

func BenchmarkAppErrorCreation(b *testing.B) {
	cause := errors.New("underlying error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewAppError("TEST_ERROR", "test message", "test_operation", cause)
	}
}

func BenchmarkErrorTypeCheck(b *testing.B) {
	err := NewHTTPError("test error", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsHTTPError(err)
	}
}
