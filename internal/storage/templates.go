package storage

import (
	"encoding/json"
	"fmt"
)

// RequestTemplate represents a pre-configured request template
type RequestTemplate struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Category    string            `json:"category"` // "REST", "GraphQL", "Auth", "CRUD", etc.
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	Headers     map[string]string `json:"headers"`
	Body        string            `json:"body"`
	QueryParams map[string]string `json:"query_params"`
	Variables   []string          `json:"variables"` // List of variable names to be filled
}

// GetBuiltInTemplates returns a list of pre-configured templates
func GetBuiltInTemplates() []RequestTemplate {
	return []RequestTemplate{
		{
			ID:          "rest-get",
			Name:        "REST - GET Request",
			Description: "Basic GET request to fetch data",
			Category:    "REST",
			Method:      "GET",
			URL:         "{{API_URL}}/resources",
			Headers: map[string]string{
				"Accept": "application/json",
			},
			Body:        "",
			QueryParams: make(map[string]string),
			Variables:   []string{"API_URL"},
		},
		{
			ID:          "rest-post",
			Name:        "REST - POST Request",
			Description: "Create a new resource",
			Category:    "REST",
			Method:      "POST",
			URL:         "{{API_URL}}/resources",
			Headers: map[string]string{
				"Accept":       "application/json",
				"Content-Type": "application/json",
			},
			Body: `{
  "name": "{{RESOURCE_NAME}}",
  "value": "{{RESOURCE_VALUE}}"
}`,
			QueryParams: make(map[string]string),
			Variables:   []string{"API_URL", "RESOURCE_NAME", "RESOURCE_VALUE"},
		},
		{
			ID:          "rest-put",
			Name:        "REST - PUT Request",
			Description: "Update an existing resource",
			Category:    "REST",
			Method:      "PUT",
			URL:         "{{API_URL}}/resources/{{RESOURCE_ID}}",
			Headers: map[string]string{
				"Accept":       "application/json",
				"Content-Type": "application/json",
			},
			Body: `{
  "name": "{{RESOURCE_NAME}}",
  "value": "{{RESOURCE_VALUE}}"
}`,
			QueryParams: make(map[string]string),
			Variables:   []string{"API_URL", "RESOURCE_ID", "RESOURCE_NAME", "RESOURCE_VALUE"},
		},
		{
			ID:          "rest-patch",
			Name:        "REST - PATCH Request",
			Description: "Partially update a resource",
			Category:    "REST",
			Method:      "PATCH",
			URL:         "{{API_URL}}/resources/{{RESOURCE_ID}}",
			Headers: map[string]string{
				"Accept":       "application/json",
				"Content-Type": "application/json",
			},
			Body: `{
  "{{FIELD_NAME}}": "{{FIELD_VALUE}}"
}`,
			QueryParams: make(map[string]string),
			Variables:   []string{"API_URL", "RESOURCE_ID", "FIELD_NAME", "FIELD_VALUE"},
		},
		{
			ID:          "rest-delete",
			Name:        "REST - DELETE Request",
			Description: "Delete a resource",
			Category:    "REST",
			Method:      "DELETE",
			URL:         "{{API_URL}}/resources/{{RESOURCE_ID}}",
			Headers: map[string]string{
				"Accept": "application/json",
			},
			Body:        "",
			QueryParams: make(map[string]string),
			Variables:   []string{"API_URL", "RESOURCE_ID"},
		},
		{
			ID:          "auth-bearer",
			Name:        "Auth - Bearer Token",
			Description: "Request with Bearer token authentication",
			Category:    "Auth",
			Method:      "GET",
			URL:         "{{API_URL}}/protected",
			Headers: map[string]string{
				"Accept":        "application/json",
				"Authorization": "Bearer {{TOKEN}}",
			},
			Body:        "",
			QueryParams: make(map[string]string),
			Variables:   []string{"API_URL", "TOKEN"},
		},
		{
			ID:          "auth-api-key",
			Name:        "Auth - API Key",
			Description: "Request with API key authentication",
			Category:    "Auth",
			Method:      "GET",
			URL:         "{{API_URL}}/protected",
			Headers: map[string]string{
				"Accept":  "application/json",
				"X-API-Key": "{{API_KEY}}",
			},
			Body:        "",
			QueryParams: make(map[string]string),
			Variables:   []string{"API_URL", "API_KEY"},
		},
		{
			ID:          "graphql-query",
			Name:        "GraphQL - Query",
			Description: "GraphQL query request",
			Category:    "GraphQL",
			Method:      "POST",
			URL:         "{{GRAPHQL_URL}}",
			Headers: map[string]string{
				"Accept":       "application/json",
				"Content-Type": "application/json",
			},
			Body: `{
  "query": "query { {{QUERY_NAME}} { {{FIELDS}} } }",
  "variables": {}
}`,
			QueryParams: make(map[string]string),
			Variables:   []string{"GRAPHQL_URL", "QUERY_NAME", "FIELDS"},
		},
		{
			ID:          "graphql-mutation",
			Name:        "GraphQL - Mutation",
			Description: "GraphQL mutation request",
			Category:    "GraphQL",
			Method:      "POST",
			URL:         "{{GRAPHQL_URL}}",
			Headers: map[string]string{
				"Accept":       "application/json",
				"Content-Type": "application/json",
			},
			Body: `{
  "query": "mutation { {{MUTATION_NAME}}(input: $input) { {{FIELDS}} } }",
  "variables": {
    "input": {}
  }
}`,
			QueryParams: make(map[string]string),
			Variables:   []string{"GRAPHQL_URL", "MUTATION_NAME", "FIELDS"},
		},
		{
			ID:          "pagination-offset",
			Name:        "Pagination - Offset/Limit",
			Description: "Paginated request using offset and limit",
			Category:    "Pagination",
			Method:      "GET",
			URL:         "{{API_URL}}/resources",
			Headers: map[string]string{
				"Accept": "application/json",
			},
			Body: "",
			QueryParams: map[string]string{
				"offset": "{{OFFSET}}",
				"limit":  "{{LIMIT}}",
			},
			Variables: []string{"API_URL", "OFFSET", "LIMIT"},
		},
		{
			ID:          "pagination-cursor",
			Name:        "Pagination - Cursor-based",
			Description: "Paginated request using cursor",
			Category:    "Pagination",
			Method:      "GET",
			URL:         "{{API_URL}}/resources",
			Headers: map[string]string{
				"Accept": "application/json",
			},
			Body: "",
			QueryParams: map[string]string{
				"cursor": "{{CURSOR}}",
				"limit":  "{{LIMIT}}",
			},
			Variables: []string{"API_URL", "CURSOR", "LIMIT"},
		},
		{
			ID:          "json-rpc",
			Name:        "JSON-RPC Request",
			Description: "JSON-RPC 2.0 request",
			Category:    "RPC",
			Method:      "POST",
			URL:         "{{RPC_URL}}",
			Headers: map[string]string{
				"Accept":       "application/json",
				"Content-Type": "application/json",
			},
			Body: `{
  "jsonrpc": "2.0",
  "method": "{{METHOD_NAME}}",
  "params": [],
  "id": 1
}`,
			QueryParams: make(map[string]string),
			Variables:   []string{"RPC_URL", "METHOD_NAME"},
		},
		{
			ID:          "file-upload",
			Name:        "File Upload (multipart/form-data)",
			Description: "Upload file using multipart form data",
			Category:    "File",
			Method:      "POST",
			URL:         "{{API_URL}}/upload",
			Headers: map[string]string{
				"Accept": "application/json",
			},
			Body:        "Note: Use multipart/form-data encoding for file uploads",
			QueryParams: make(map[string]string),
			Variables:   []string{"API_URL"},
		},
		{
			ID:          "search-query",
			Name:        "Search Query",
			Description: "Search with query parameters",
			Category:    "Search",
			Method:      "GET",
			URL:         "{{API_URL}}/search",
			Headers: map[string]string{
				"Accept": "application/json",
			},
			Body: "",
			QueryParams: map[string]string{
				"q":    "{{SEARCH_QUERY}}",
				"sort": "{{SORT_FIELD}}",
			},
			Variables: []string{"API_URL", "SEARCH_QUERY", "SORT_FIELD"},
		},
	}
}

// GetTemplatesByCategory returns templates filtered by category
func GetTemplatesByCategory(category string) []RequestTemplate {
	templates := GetBuiltInTemplates()
	var filtered []RequestTemplate

	for _, tmpl := range templates {
		if tmpl.Category == category {
			filtered = append(filtered, tmpl)
		}
	}

	return filtered
}

// GetTemplateCategories returns list of all template categories
func GetTemplateCategories() []string {
	categoriesMap := make(map[string]bool)
	templates := GetBuiltInTemplates()

	for _, tmpl := range templates {
		categoriesMap[tmpl.Category] = true
	}

	var categories []string
	for cat := range categoriesMap {
		categories = append(categories, cat)
	}

	return categories
}

// ApplyTemplate creates a SavedRequest from a template
func ApplyTemplate(template RequestTemplate, variableValues map[string]string) SavedRequest {
	// Replace variables in URL, headers, and body
	url := template.URL
	body := template.Body
	headers := make(map[string]string)
	queryParams := make(map[string]string)

	// Replace in URL
	for varName, varValue := range variableValues {
		placeholder := fmt.Sprintf("{{%s}}", varName)
		url = replaceAll(url, placeholder, varValue)
		body = replaceAll(body, placeholder, varValue)
	}

	// Replace in headers
	for key, value := range template.Headers {
		newKey := key
		newValue := value
		for varName, varValue := range variableValues {
			placeholder := fmt.Sprintf("{{%s}}", varName)
			newKey = replaceAll(newKey, placeholder, varValue)
			newValue = replaceAll(newValue, placeholder, varValue)
		}
		headers[newKey] = newValue
	}

	// Replace in query params
	for key, value := range template.QueryParams {
		newKey := key
		newValue := value
		for varName, varValue := range variableValues {
			placeholder := fmt.Sprintf("{{%s}}", varName)
			newKey = replaceAll(newKey, placeholder, varValue)
			newValue = replaceAll(newValue, placeholder, varValue)
		}
		queryParams[newKey] = newValue
	}

	return SavedRequest{
		ID:          "",
		Name:        template.Name,
		Method:      template.Method,
		URL:         url,
		Headers:     headers,
		Body:        body,
		QueryParams: queryParams,
	}
}

// Simple string replace all helper
func replaceAll(s, old, new string) string {
	result := ""
	for {
		idx := indexOf(s, old)
		if idx == -1 {
			result += s
			break
		}
		result += s[:idx] + new
		s = s[idx+len(old):]
	}
	return result
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// ExportOpenAPISpec generates an OpenAPI specification from a collection
func ExportOpenAPISpec(collection *Collection) ([]byte, error) {
	spec := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]string{
			"title":       collection.Name,
			"description": collection.Description,
			"version":     "1.0.0",
		},
		"paths": make(map[string]interface{}),
	}

	paths := spec["paths"].(map[string]interface{})

	for _, req := range collection.Requests {
		// Extract path from URL (simplified - assumes no query params in URL)
		path := req.URL
		if len(path) > 0 && path[0] != '/' {
			// Try to extract path from full URL
			if idx := indexOf(path, "://"); idx != -1 {
				path = path[idx+3:]
				if idx := indexOf(path, "/"); idx != -1 {
					path = path[idx:]
				}
			}
		}

		if _, exists := paths[path]; !exists {
			paths[path] = make(map[string]interface{})
		}

		pathItem := paths[path].(map[string]interface{})
		method := methodToLower(req.Method)

		operation := map[string]interface{}{
			"summary": req.Name,
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Successful response",
				},
			},
		}

		// Add request body if present
		if req.Body != "" {
			operation["requestBody"] = map[string]interface{}{
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"example": parseJSONOrString(req.Body),
					},
				},
			}
		}

		pathItem[method] = operation
	}

	return json.MarshalIndent(spec, "", "  ")
}

func methodToLower(method string) string {
	result := ""
	for _, ch := range method {
		if ch >= 'A' && ch <= 'Z' {
			result += string(ch + 32)
		} else {
			result += string(ch)
		}
	}
	return result
}

func parseJSONOrString(s string) interface{} {
	var result interface{}
	if err := json.Unmarshal([]byte(s), &result); err == nil {
		return result
	}
	return s
}
