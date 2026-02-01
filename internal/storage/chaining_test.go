package storage

import (
	"testing"
)

func TestExtractVariableJSONPath(t *testing.T) {
	responseBody := `{
		"data": {
			"user": {
				"id": 123,
				"name": "John Doe",
				"email": "john@example.com"
			},
			"items": [
				{"id": 1, "name": "Item 1"},
				{"id": 2, "name": "Item 2"}
			]
		}
	}`

	tests := []struct {
		name        string
		extraction  VariableExtract
		expected    string
		shouldError bool
	}{
		{
			name: "simple path",
			extraction: VariableExtract{
				Name:     "user_id",
				JSONPath: "data.user.id",
			},
			expected:    "123",
			shouldError: false,
		},
		{
			name: "nested path",
			extraction: VariableExtract{
				Name:     "user_email",
				JSONPath: "data.user.email",
			},
			expected:    "john@example.com",
			shouldError: false,
		},
		{
			name: "array index",
			extraction: VariableExtract{
				Name:     "first_item_name",
				JSONPath: "data.items[0].name",
			},
			expected:    "Item 1",
			shouldError: false,
		},
		{
			name: "array second element",
			extraction: VariableExtract{
				Name:     "second_item_id",
				JSONPath: "data.items[1].id",
			},
			expected:    "2",
			shouldError: false,
		},
		{
			name: "non-existent path",
			extraction: VariableExtract{
				Name:     "missing",
				JSONPath: "data.missing.field",
			},
			expected:    "",
			shouldError: true,
		},
		{
			name: "array out of bounds",
			extraction: VariableExtract{
				Name:     "invalid_index",
				JSONPath: "data.items[10].name",
			},
			expected:    "",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractVariable(responseBody, tt.extraction)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected '%s', got '%s'", tt.expected, result)
				}
			}
		})
	}
}

func TestExtractVariableRegex(t *testing.T) {
	responseBody := `<html><body><div id="token">abc123def</div></body></html>`

	extraction := VariableExtract{
		Name:  "token",
		Regex: `<div id="token">([^<]+)</div>`,
	}

	result, err := ExtractVariable(responseBody, extraction)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result != "abc123def" {
		t.Errorf("Expected 'abc123def', got '%s'", result)
	}
}

func TestExtractVariableInvalidRegex(t *testing.T) {
	extraction := VariableExtract{
		Name:  "test",
		Regex: `[invalid(`,
	}

	_, err := ExtractVariable("test", extraction)
	if err == nil {
		t.Error("Expected error for invalid regex")
	}
}

func TestValidateAssertionStatusCode(t *testing.T) {
	assertion := ResponseAssertion{
		Type:     "status_code",
		Operator: "equals",
		Value:    "200",
	}

	err := ValidateAssertion(assertion, 200, "", nil, 0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	err = ValidateAssertion(assertion, 404, "", nil, 0)
	if err == nil {
		t.Error("Expected error for mismatched status code")
	}
}

func TestValidateAssertionResponseTime(t *testing.T) {
	tests := []struct {
		name        string
		assertion   ResponseAssertion
		responseMs  int64
		shouldError bool
	}{
		{
			name: "response time less than max",
			assertion: ResponseAssertion{
				Type:     "response_time",
				Operator: "less_than",
				Value:    "500",
			},
			responseMs:  200,
			shouldError: false,
		},
		{
			name: "response time exceeds max",
			assertion: ResponseAssertion{
				Type:     "response_time",
				Operator: "less_than",
				Value:    "100",
			},
			responseMs:  200,
			shouldError: true,
		},
		{
			name: "response time greater than min",
			assertion: ResponseAssertion{
				Type:     "response_time",
				Operator: "greater_than",
				Value:    "100",
			},
			responseMs:  200,
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAssertion(tt.assertion, 200, "", nil, tt.responseMs)

			if tt.shouldError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateAssertionHeader(t *testing.T) {
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer token123",
	}

	tests := []struct {
		name        string
		assertion   ResponseAssertion
		shouldError bool
	}{
		{
			name: "header equals",
			assertion: ResponseAssertion{
				Type:     "header",
				Field:    "Content-Type",
				Operator: "equals",
				Value:    "application/json",
			},
			shouldError: false,
		},
		{
			name: "header contains",
			assertion: ResponseAssertion{
				Type:     "header",
				Field:    "Authorization",
				Operator: "contains",
				Value:    "Bearer",
			},
			shouldError: false,
		},
		{
			name: "header exists",
			assertion: ResponseAssertion{
				Type:     "header",
				Field:    "Content-Type",
				Operator: "exists",
			},
			shouldError: false,
		},
		{
			name: "header missing",
			assertion: ResponseAssertion{
				Type:     "header",
				Field:    "X-Missing",
				Operator: "exists",
			},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAssertion(tt.assertion, 200, "", headers, 0)

			if tt.shouldError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateAssertionBodyContains(t *testing.T) {
	responseBody := `{"message": "Hello, World!", "status": "success"}`

	tests := []struct {
		name        string
		value       string
		shouldError bool
	}{
		{
			name:        "body contains string",
			value:       "Hello, World!",
			shouldError: false,
		},
		{
			name:        "body contains key",
			value:       "status",
			shouldError: false,
		},
		{
			name:        "body missing string",
			value:       "missing text",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertion := ResponseAssertion{
				Type:  "body_contains",
				Value: tt.value,
			}

			err := ValidateAssertion(assertion, 200, responseBody, nil, 0)

			if tt.shouldError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateAssertionJSONPath(t *testing.T) {
	responseBody := `{
		"user": {
			"id": 123,
			"name": "John Doe",
			"active": true
		}
	}`

	tests := []struct {
		name        string
		assertion   ResponseAssertion
		shouldError bool
	}{
		{
			name: "json path equals",
			assertion: ResponseAssertion{
				Type:     "json_path",
				Field:    "user.name",
				Operator: "equals",
				Value:    "John Doe",
			},
			shouldError: false,
		},
		{
			name: "json path contains",
			assertion: ResponseAssertion{
				Type:     "json_path",
				Field:    "user.name",
				Operator: "contains",
				Value:    "John",
			},
			shouldError: false,
		},
		{
			name: "json path exists",
			assertion: ResponseAssertion{
				Type:     "json_path",
				Field:    "user.id",
				Operator: "exists",
			},
			shouldError: false,
		},
		{
			name: "json path missing",
			assertion: ResponseAssertion{
				Type:     "json_path",
				Field:    "user.missing",
				Operator: "exists",
			},
			shouldError: true,
		},
		{
			name: "json path value mismatch",
			assertion: ResponseAssertion{
				Type:     "json_path",
				Field:    "user.name",
				Operator: "equals",
				Value:    "Jane Doe",
			},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAssertion(tt.assertion, 200, responseBody, nil, 0)

			if tt.shouldError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestParseJSONPath(t *testing.T) {
	tests := []struct {
		path     string
		expected int // number of parts
	}{
		{"data.user.id", 3},
		{"items[0].name", 3},    // items, [0], name
		{"data.items[1].id", 4}, // data, items, [1], id
		{"simple", 1},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			parts := parseJSONPath(tt.path)
			if len(parts) != tt.expected {
				t.Errorf("Expected %d parts, got %d", tt.expected, len(parts))
			}
		})
	}
}
