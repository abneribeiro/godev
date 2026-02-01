package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSendGraphQLRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Verify Content-Type
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}

		// Return GraphQL response
		response := GraphQLResponse{
			Data: map[string]interface{}{
				"user": map[string]interface{}{
					"id":   "123",
					"name": "John Doe",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(5 * time.Second)
	query := `query { user(id: "123") { id name } }`

	resp, err := SendGraphQLRequest(client, server.URL, query, nil)
	if err != nil {
		t.Fatalf("SendGraphQLRequest failed: %v", err)
	}

	if resp.Data == nil {
		t.Error("Expected non-nil data")
	}

	if len(resp.Errors) != 0 {
		t.Errorf("Expected no errors, got %d", len(resp.Errors))
	}
}

func TestSendGraphQLRequestWithVariables(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse request body
		var gqlReq GraphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&gqlReq); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		// Verify variables were sent
		if gqlReq.Variables == nil {
			t.Error("Expected variables in request")
		}

		if gqlReq.Variables["userId"] != "123" {
			t.Errorf("Expected userId variable '123', got '%v'", gqlReq.Variables["userId"])
		}

		response := GraphQLResponse{
			Data: map[string]interface{}{"result": "ok"},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(5 * time.Second)
	query := `query($userId: ID!) { user(id: $userId) { id } }`
	variables := map[string]interface{}{
		"userId": "123",
	}

	_, err := SendGraphQLRequest(client, server.URL, query, variables)
	if err != nil {
		t.Fatalf("SendGraphQLRequest failed: %v", err)
	}
}

func TestSendGraphQLRequestWithErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := GraphQLResponse{
			Errors: []GraphQLError{
				{
					Message: "Field 'nonexistent' doesn't exist",
					Locations: []GraphQLLocation{
						{Line: 1, Column: 10},
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(5 * time.Second)
	query := `query { nonexistent }`

	resp, err := SendGraphQLRequest(client, server.URL, query, nil)
	if err != nil {
		t.Fatalf("SendGraphQLRequest failed: %v", err)
	}

	if len(resp.Errors) == 0 {
		t.Error("Expected errors in response")
	}

	if resp.Errors[0].Message != "Field 'nonexistent' doesn't exist" {
		t.Errorf("Expected error message about nonexistent field, got '%s'", resp.Errors[0].Message)
	}
}

func TestFormatGraphQLType(t *testing.T) {
	tests := []struct {
		name     string
		typeRef  GraphQLTypeRef
		expected string
	}{
		{
			name:     "scalar type",
			typeRef:  GraphQLTypeRef{Kind: "SCALAR", Name: "String"},
			expected: "String",
		},
		{
			name: "non-null type",
			typeRef: GraphQLTypeRef{
				Kind: "NON_NULL",
				OfType: &GraphQLTypeRef{
					Kind: "SCALAR",
					Name: "String",
				},
			},
			expected: "String!",
		},
		{
			name: "list type",
			typeRef: GraphQLTypeRef{
				Kind: "LIST",
				OfType: &GraphQLTypeRef{
					Kind: "SCALAR",
					Name: "String",
				},
			},
			expected: "[String]",
		},
		{
			name: "non-null list",
			typeRef: GraphQLTypeRef{
				Kind: "NON_NULL",
				OfType: &GraphQLTypeRef{
					Kind: "LIST",
					OfType: &GraphQLTypeRef{
						Kind: "SCALAR",
						Name: "String",
					},
				},
			},
			expected: "[String]!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatGraphQLType(tt.typeRef)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestFormatGraphQLError(t *testing.T) {
	tests := []struct {
		name     string
		error    GraphQLError
		contains []string
	}{
		{
			name: "simple error",
			error: GraphQLError{
				Message: "Syntax error",
			},
			contains: []string{"Syntax error"},
		},
		{
			name: "error with location",
			error: GraphQLError{
				Message: "Field error",
				Locations: []GraphQLLocation{
					{Line: 5, Column: 10},
				},
			},
			contains: []string{"Field error", "line 5", "column 10"},
		},
		{
			name: "error with path",
			error: GraphQLError{
				Message: "Null value",
				Path:    []interface{}{"user", "email"},
			},
			contains: []string{"Null value", "user.email"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted := FormatGraphQLError(tt.error)

			for _, expected := range tt.contains {
				if !containsGraphQLStr(formatted, expected) {
					t.Errorf("Expected formatted error to contain '%s', got '%s'", expected, formatted)
				}
			}
		})
	}
}

func TestValidateGraphQLQuery(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		shouldError bool
	}{
		{
			name:        "valid query",
			query:       "query { user { id name } }",
			shouldError: false,
		},
		{
			name:        "valid mutation",
			query:       "mutation { createUser(name: \"John\") { id } }",
			shouldError: false,
		},
		{
			name:        "valid shorthand query",
			query:       "{ user { id } }",
			shouldError: false,
		},
		{
			name:        "empty query",
			query:       "",
			shouldError: true,
		},
		{
			name:        "unbalanced braces - extra opening",
			query:       "query { user { id }",
			shouldError: true,
		},
		{
			name:        "unbalanced braces - extra closing",
			query:       "query { user { id } } }",
			shouldError: true,
		},
		{
			name:        "no keyword or braces",
			query:       "just some text",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGraphQLQuery(tt.query)

			if tt.shouldError && err == nil {
				t.Error("Expected error for invalid query")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestGenerateGraphQLQuery(t *testing.T) {
	// Create a simple schema
	schema := &GraphQLSchema{
		Types: []GraphQLType{
			{
				Kind: "OBJECT",
				Name: "User",
				Fields: []GraphQLField{
					{
						Name: "id",
						Type: GraphQLTypeRef{Kind: "SCALAR", Name: "ID"},
					},
					{
						Name: "name",
						Type: GraphQLTypeRef{Kind: "SCALAR", Name: "String"},
					},
					{
						Name: "email",
						Type: GraphQLTypeRef{Kind: "SCALAR", Name: "String"},
					},
				},
			},
		},
	}

	query, err := GenerateGraphQLQuery(schema, "User", 1)
	if err != nil {
		t.Fatalf("GenerateGraphQLQuery failed: %v", err)
	}

	// Check that query contains expected elements
	if !containsGraphQLStr(query, "query") {
		t.Error("Expected query to contain 'query' keyword")
	}

	if !containsGraphQLStr(query, "User") {
		t.Error("Expected query to contain 'User' type")
	}

	if !containsGraphQLStr(query, "id") {
		t.Error("Expected query to contain 'id' field")
	}

	if !containsGraphQLStr(query, "name") {
		t.Error("Expected query to contain 'name' field")
	}
}

func TestGenerateGraphQLQueryNonExistentType(t *testing.T) {
	schema := &GraphQLSchema{
		Types: []GraphQLType{},
	}

	_, err := GenerateGraphQLQuery(schema, "NonExistent", 1)
	if err == nil {
		t.Error("Expected error for non-existent type")
	}
}

func containsGraphQLStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
