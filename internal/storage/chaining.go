package storage

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// ChainedRequest represents a request that can extract variables from response
type ChainedRequest struct {
	Request        SavedRequest       `json:"request"`
	Extractions    []VariableExtract  `json:"extractions"`
	Assertions     []ResponseAssertion `json:"assertions,omitempty"`
	NextRequestID  string             `json:"next_request_id,omitempty"`
}

// VariableExtract defines how to extract a variable from response
type VariableExtract struct {
	Name     string `json:"name"`      // Variable name to save as
	JSONPath string `json:"json_path"` // Path like "data.user.id" or "items[0].name"
	Regex    string `json:"regex"`     // Alternative: regex pattern with capture group
	Header   string `json:"header"`    // Alternative: extract from response header
}

// ResponseAssertion defines a test to validate response
type ResponseAssertion struct {
	Type     string `json:"type"`     // "status_code", "json_path", "response_time", "header", "body_contains"
	Field    string `json:"field"`    // For json_path: path, for header: header name
	Operator string `json:"operator"` // "equals", "contains", "greater_than", "less_than", "exists"
	Value    string `json:"value"`    // Expected value
}

// RequestChain represents a sequence of requests
type RequestChain struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Requests    []ChainedRequest `json:"requests"`
}

// ExtractVariable extracts a value from JSON response using JSONPath-like syntax
func ExtractVariable(responseBody string, extraction VariableExtract) (string, error) {
	// If extracting from header, this should be handled by caller with response headers
	if extraction.Header != "" {
		return "", fmt.Errorf("header extraction must be handled separately")
	}

	// If using regex
	if extraction.Regex != "" {
		re, err := regexp.Compile(extraction.Regex)
		if err != nil {
			return "", fmt.Errorf("invalid regex pattern: %w", err)
		}
		matches := re.FindStringSubmatch(responseBody)
		if len(matches) > 1 {
			return matches[1], nil // Return first capture group
		}
		return "", fmt.Errorf("regex pattern did not match")
	}

	// JSONPath extraction
	if extraction.JSONPath != "" {
		var data interface{}
		if err := json.Unmarshal([]byte(responseBody), &data); err != nil {
			return "", fmt.Errorf("response is not valid JSON: %w", err)
		}

		value, err := extractJSONPath(data, extraction.JSONPath)
		if err != nil {
			return "", err
		}

		// Convert value to string
		switch v := value.(type) {
		case string:
			return v, nil
		case float64:
			return fmt.Sprintf("%v", v), nil
		case bool:
			return fmt.Sprintf("%v", v), nil
		case nil:
			return "null", nil
		default:
			// For objects/arrays, return JSON representation
			bytes, err := json.Marshal(v)
			if err != nil {
				return "", fmt.Errorf("failed to convert value to string: %w", err)
			}
			return string(bytes), nil
		}
	}

	return "", fmt.Errorf("no extraction method specified")
}

// extractJSONPath extracts a value from nested JSON using dot notation
// Supports: "data.user.id", "items[0].name", "data.items[1].id"
func extractJSONPath(data interface{}, path string) (interface{}, error) {
	if path == "" {
		return data, nil
	}

	parts := parseJSONPath(path)
	current := data

	for _, part := range parts {
		if part.isArray {
			// Handle array index access
			arr, ok := current.([]interface{})
			if !ok {
				return nil, fmt.Errorf("expected array at '%s', got %T", part.key, current)
			}
			if part.index < 0 || part.index >= len(arr) {
				return nil, fmt.Errorf("array index %d out of bounds (length: %d)", part.index, len(arr))
			}
			current = arr[part.index]
		} else {
			// Handle object key access
			obj, ok := current.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("expected object at '%s', got %T", part.key, current)
			}
			val, exists := obj[part.key]
			if !exists {
				return nil, fmt.Errorf("key '%s' not found", part.key)
			}
			current = val
		}
	}

	return current, nil
}

type pathPart struct {
	key     string
	isArray bool
	index   int
}

// parseJSONPath parses a JSON path into parts
// Examples: "data.user.id" -> [{data}, {user}, {id}]
//          "items[0].name" -> [{items}, {[0]}, {name}]
func parseJSONPath(path string) []pathPart {
	var parts []pathPart

	// Split by dots, but handle array brackets specially
	segments := strings.Split(path, ".")

	for _, segment := range segments {
		// Check if segment contains array index
		if strings.Contains(segment, "[") {
			// Split into key and array parts
			bracketIdx := strings.Index(segment, "[")
			key := segment[:bracketIdx]

			// Add the key part if not empty
			if key != "" {
				parts = append(parts, pathPart{key: key, isArray: false})
			}

			// Extract all array indices from this segment
			remainder := segment[bracketIdx:]
			arrayPattern := regexp.MustCompile(`\[(\d+)\]`)
			matches := arrayPattern.FindAllStringSubmatch(remainder, -1)

			for _, match := range matches {
				var index int
				fmt.Sscanf(match[1], "%d", &index)
				parts = append(parts, pathPart{key: match[1], isArray: true, index: index})
			}
		} else {
			parts = append(parts, pathPart{key: segment, isArray: false})
		}
	}

	return parts
}

// ValidateAssertion checks if a response matches an assertion
func ValidateAssertion(assertion ResponseAssertion, statusCode int, responseBody string, responseHeaders map[string]string, responseTimeMs int64) error {
	switch assertion.Type {
	case "status_code":
		expected := 0
		fmt.Sscanf(assertion.Value, "%d", &expected)
		if statusCode != expected {
			return fmt.Errorf("expected status %d, got %d", expected, statusCode)
		}
		return nil

	case "response_time":
		var maxTime int64
		fmt.Sscanf(assertion.Value, "%d", &maxTime)

		switch assertion.Operator {
		case "less_than":
			if responseTimeMs >= maxTime {
				return fmt.Errorf("response time %dms not less than %dms", responseTimeMs, maxTime)
			}
		case "greater_than":
			if responseTimeMs <= maxTime {
				return fmt.Errorf("response time %dms not greater than %dms", responseTimeMs, maxTime)
			}
		}
		return nil

	case "header":
		headerValue, exists := responseHeaders[assertion.Field]
		if !exists {
			if assertion.Operator == "exists" {
				return fmt.Errorf("header '%s' does not exist", assertion.Field)
			}
			return fmt.Errorf("header '%s' not found", assertion.Field)
		}

		switch assertion.Operator {
		case "equals":
			if headerValue != assertion.Value {
				return fmt.Errorf("header '%s' expected '%s', got '%s'", assertion.Field, assertion.Value, headerValue)
			}
		case "contains":
			if !strings.Contains(headerValue, assertion.Value) {
				return fmt.Errorf("header '%s' does not contain '%s'", assertion.Field, assertion.Value)
			}
		}
		return nil

	case "body_contains":
		if !strings.Contains(responseBody, assertion.Value) {
			return fmt.Errorf("response body does not contain '%s'", assertion.Value)
		}
		return nil

	case "json_path":
		var data interface{}
		if err := json.Unmarshal([]byte(responseBody), &data); err != nil {
			return fmt.Errorf("response is not valid JSON: %w", err)
		}

		value, err := extractJSONPath(data, assertion.Field)
		if err != nil {
			if assertion.Operator == "exists" {
				return fmt.Errorf("JSON path '%s' does not exist", assertion.Field)
			}
			return err
		}

		if assertion.Operator == "exists" {
			return nil // Value exists
		}

		// Convert value to string for comparison
		valueStr := fmt.Sprintf("%v", value)

		switch assertion.Operator {
		case "equals":
			if valueStr != assertion.Value {
				return fmt.Errorf("JSON path '%s' expected '%s', got '%s'", assertion.Field, assertion.Value, valueStr)
			}
		case "contains":
			if !strings.Contains(valueStr, assertion.Value) {
				return fmt.Errorf("JSON path '%s' value '%s' does not contain '%s'", assertion.Field, valueStr, assertion.Value)
			}
		}
		return nil

	default:
		return fmt.Errorf("unknown assertion type: %s", assertion.Type)
	}
}
