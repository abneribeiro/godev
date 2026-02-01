package storage

import (
	"testing"
)

func TestGetBuiltInTemplates(t *testing.T) {
	templates := GetBuiltInTemplates()

	if len(templates) == 0 {
		t.Error("Expected at least one built-in template")
	}

	// Check that all templates have required fields
	for _, tmpl := range templates {
		if tmpl.ID == "" {
			t.Error("Template ID should not be empty")
		}
		if tmpl.Name == "" {
			t.Error("Template name should not be empty")
		}
		if tmpl.Method == "" {
			t.Error("Template method should not be empty")
		}
		if tmpl.Category == "" {
			t.Error("Template category should not be empty")
		}
	}
}

func TestGetTemplatesByCategory(t *testing.T) {
	tests := []struct {
		category     string
		minTemplates int
	}{
		{"REST", 4},       // GET, POST, PUT, PATCH, DELETE
		{"Auth", 2},       // Bearer, API Key
		{"GraphQL", 2},    // Query, Mutation
		{"Pagination", 2}, // Offset, Cursor
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			templates := GetTemplatesByCategory(tt.category)

			if len(templates) < tt.minTemplates {
				t.Errorf("Expected at least %d %s templates, got %d",
					tt.minTemplates, tt.category, len(templates))
			}

			// Verify all templates are from correct category
			for _, tmpl := range templates {
				if tmpl.Category != tt.category {
					t.Errorf("Expected category %s, got %s", tt.category, tmpl.Category)
				}
			}
		})
	}
}

func TestGetTemplateCategories(t *testing.T) {
	categories := GetTemplateCategories()

	if len(categories) == 0 {
		t.Error("Expected at least one category")
	}

	expectedCategories := map[string]bool{
		"REST":       true,
		"Auth":       true,
		"GraphQL":    true,
		"Pagination": true,
	}

	for expectedCat := range expectedCategories {
		found := false
		for _, cat := range categories {
			if cat == expectedCat {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find category %s", expectedCat)
		}
	}
}

func TestApplyTemplate(t *testing.T) {
	templates := GetBuiltInTemplates()

	// Find REST POST template
	var postTemplate RequestTemplate
	for _, tmpl := range templates {
		if tmpl.ID == "rest-post" {
			postTemplate = tmpl
			break
		}
	}

	if postTemplate.ID == "" {
		t.Fatal("Could not find rest-post template")
	}

	// Apply template with variables
	variables := map[string]string{
		"API_URL":        "https://api.example.com",
		"RESOURCE_NAME":  "John Doe",
		"RESOURCE_VALUE": "test@example.com",
	}

	request := ApplyTemplate(postTemplate, variables)

	// Check that variables were replaced
	if !containsSubstring(request.URL, "https://api.example.com") {
		t.Errorf("Expected URL to contain 'https://api.example.com', got %s", request.URL)
	}

	if !containsSubstring(request.Body, "John Doe") {
		t.Errorf("Expected body to contain 'John Doe', got %s", request.Body)
	}

	if !containsSubstring(request.Body, "test@example.com") {
		t.Errorf("Expected body to contain 'test@example.com', got %s", request.Body)
	}

	// Check headers
	if request.Headers["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type header to be 'application/json', got '%s'",
			request.Headers["Content-Type"])
	}
}

func TestApplyTemplateWithMissingVariables(t *testing.T) {
	template := RequestTemplate{
		ID:       "test",
		Name:     "Test",
		Category: "Test",
		Method:   "GET",
		URL:      "{{API_URL}}/{{RESOURCE}}",
		Headers:  make(map[string]string),
		Body:     "",
	}

	// Apply with partial variables
	variables := map[string]string{
		"API_URL": "https://api.example.com",
		// RESOURCE is missing
	}

	request := ApplyTemplate(template, variables)

	// API_URL should be replaced, RESOURCE should remain as placeholder
	expected := "https://api.example.com/{{RESOURCE}}"
	if request.URL != expected {
		t.Errorf("Expected URL '%s', got '%s'", expected, request.URL)
	}
}

func TestApplyTemplateAuthBearer(t *testing.T) {
	templates := GetBuiltInTemplates()

	var bearerTemplate RequestTemplate
	for _, tmpl := range templates {
		if tmpl.ID == "auth-bearer" {
			bearerTemplate = tmpl
			break
		}
	}

	if bearerTemplate.ID == "" {
		t.Fatal("Could not find auth-bearer template")
	}

	variables := map[string]string{
		"API_URL": "https://api.example.com",
		"TOKEN":   "abc123token",
	}

	request := ApplyTemplate(bearerTemplate, variables)

	authHeader := request.Headers["Authorization"]
	if authHeader != "Bearer abc123token" {
		t.Errorf("Expected Authorization header 'Bearer abc123token', got '%s'", authHeader)
	}
}

func TestApplyTemplateGraphQLQuery(t *testing.T) {
	templates := GetBuiltInTemplates()

	var gqlTemplate RequestTemplate
	for _, tmpl := range templates {
		if tmpl.ID == "graphql-query" {
			gqlTemplate = tmpl
			break
		}
	}

	if gqlTemplate.ID == "" {
		t.Fatal("Could not find graphql-query template")
	}

	variables := map[string]string{
		"GRAPHQL_URL": "https://api.example.com/graphql",
		"QUERY_NAME":  "users",
		"FIELDS":      "id name email",
	}

	request := ApplyTemplate(gqlTemplate, variables)

	if request.URL != "https://api.example.com/graphql" {
		t.Errorf("Expected URL 'https://api.example.com/graphql', got '%s'", request.URL)
	}

	if !containsSubstring(request.Body, "users") {
		t.Errorf("Expected body to contain 'users', got %s", request.Body)
	}

	if request.Headers["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", request.Headers["Content-Type"])
	}
}

func TestApplyTemplatePagination(t *testing.T) {
	templates := GetBuiltInTemplates()

	var paginationTemplate RequestTemplate
	for _, tmpl := range templates {
		if tmpl.ID == "pagination-offset" {
			paginationTemplate = tmpl
			break
		}
	}

	if paginationTemplate.ID == "" {
		t.Fatal("Could not find pagination-offset template")
	}

	variables := map[string]string{
		"API_URL": "https://api.example.com",
		"OFFSET":  "20",
		"LIMIT":   "10",
	}

	request := ApplyTemplate(paginationTemplate, variables)

	if request.QueryParams["offset"] != "20" {
		t.Errorf("Expected offset param '20', got '%s'", request.QueryParams["offset"])
	}

	if request.QueryParams["limit"] != "10" {
		t.Errorf("Expected limit param '10', got '%s'", request.QueryParams["limit"])
	}
}

func TestExportOpenAPISpec(t *testing.T) {
	collection := CreateCollection("Test API", "Test API collection")

	// Add a few requests
	req1 := SavedRequest{
		Name:   "Get Users",
		Method: "GET",
		URL:    "https://api.example.com/users",
		Headers: map[string]string{
			"Accept": "application/json",
		},
	}

	req2 := SavedRequest{
		Name:   "Create User",
		Method: "POST",
		URL:    "https://api.example.com/users",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: `{"name": "John Doe"}`,
	}

	AddRequestToCollection(&collection, req1)
	AddRequestToCollection(&collection, req2)

	spec, err := ExportOpenAPISpec(&collection)
	if err != nil {
		t.Fatalf("Failed to export OpenAPI spec: %v", err)
	}

	specStr := string(spec)

	// Check for required OpenAPI fields
	if !containsSubstring(specStr, "openapi") {
		t.Error("Expected spec to contain 'openapi' field")
	}

	if !containsSubstring(specStr, "Test API") {
		t.Error("Expected spec to contain 'Test API'")
	}

	if !containsSubstring(specStr, "paths") {
		t.Error("Expected spec to contain 'paths'")
	}
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
