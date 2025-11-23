package http

import (
	"encoding/json"
	"fmt"
	"strings"
)

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string                 `json:"operationName,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data   interface{}            `json:"data,omitempty"`
	Errors []GraphQLError         `json:"errors,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message    string                 `json:"message"`
	Locations  []GraphQLLocation      `json:"locations,omitempty"`
	Path       []interface{}          `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// GraphQLLocation represents error location in query
type GraphQLLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// GraphQLSchema represents a GraphQL schema from introspection
type GraphQLSchema struct {
	QueryType        *GraphQLType   `json:"queryType"`
	MutationType     *GraphQLType   `json:"mutationType,omitempty"`
	SubscriptionType *GraphQLType   `json:"subscriptionType,omitempty"`
	Types            []GraphQLType  `json:"types"`
	Directives       []GraphQLDirective `json:"directives"`
}

// GraphQLType represents a GraphQL type
type GraphQLType struct {
	Kind        string             `json:"kind"`
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Fields      []GraphQLField     `json:"fields,omitempty"`
	InputFields []GraphQLInputValue `json:"inputFields,omitempty"`
	Interfaces  []GraphQLTypeRef   `json:"interfaces,omitempty"`
	EnumValues  []GraphQLEnumValue `json:"enumValues,omitempty"`
	PossibleTypes []GraphQLTypeRef `json:"possibleTypes,omitempty"`
}

// GraphQLField represents a field in a GraphQL type
type GraphQLField struct {
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Args        []GraphQLInputValue `json:"args"`
	Type        GraphQLTypeRef      `json:"type"`
}

// GraphQLInputValue represents an input value or argument
type GraphQLInputValue struct {
	Name         string         `json:"name"`
	Description  string         `json:"description,omitempty"`
	Type         GraphQLTypeRef `json:"type"`
	DefaultValue string         `json:"defaultValue,omitempty"`
}

// GraphQLTypeRef represents a reference to a type
type GraphQLTypeRef struct {
	Kind   string          `json:"kind"`
	Name   string          `json:"name,omitempty"`
	OfType *GraphQLTypeRef `json:"ofType,omitempty"`
}

// GraphQLEnumValue represents an enum value
type GraphQLEnumValue struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// GraphQLDirective represents a GraphQL directive
type GraphQLDirective struct {
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Locations   []string            `json:"locations"`
	Args        []GraphQLInputValue `json:"args"`
}

// IntrospectionQuery is the standard GraphQL introspection query
const IntrospectionQuery = `
query IntrospectionQuery {
  __schema {
    queryType { name }
    mutationType { name }
    subscriptionType { name }
    types {
      ...FullType
    }
    directives {
      name
      description
      locations
      args {
        ...InputValue
      }
    }
  }
}

fragment FullType on __Type {
  kind
  name
  description
  fields(includeDeprecated: true) {
    name
    description
    args {
      ...InputValue
    }
    type {
      ...TypeRef
    }
  }
  inputFields {
    ...InputValue
  }
  interfaces {
    ...TypeRef
  }
  enumValues(includeDeprecated: true) {
    name
    description
  }
  possibleTypes {
    ...TypeRef
  }
}

fragment InputValue on __InputValue {
  name
  description
  type { ...TypeRef }
  defaultValue
}

fragment TypeRef on __Type {
  kind
  name
  ofType {
    kind
    name
    ofType {
      kind
      name
      ofType {
        kind
        name
        ofType {
          kind
          name
          ofType {
            kind
            name
            ofType {
              kind
              name
              ofType {
                kind
                name
              }
            }
          }
        }
      }
    }
  }
}
`

// SendGraphQLRequest sends a GraphQL request
func SendGraphQLRequest(client *Client, endpoint string, query string, variables map[string]interface{}) (*GraphQLResponse, error) {
	gqlReq := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	bodyBytes, err := json.Marshal(gqlReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal GraphQL request: %w", err)
	}

	req := Request{
		Method: "POST",
		URL:    endpoint,
		Headers: map[string]string{
			"Content-Type": "application/json",
			"Accept":       "application/json",
		},
		Body: string(bodyBytes),
	}

	resp := client.Send(req)
	if resp.Error != nil {
		return nil, resp.Error
	}

	var gqlResp GraphQLResponse
	if err := json.Unmarshal([]byte(resp.Body), &gqlResp); err != nil {
		return nil, fmt.Errorf("failed to parse GraphQL response: %w", err)
	}

	return &gqlResp, nil
}

// IntrospectSchema performs introspection on a GraphQL endpoint
func IntrospectSchema(client *Client, endpoint string) (*GraphQLSchema, error) {
	gqlResp, err := SendGraphQLRequest(client, endpoint, IntrospectionQuery, nil)
	if err != nil {
		return nil, err
	}

	if len(gqlResp.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL errors: %v", gqlResp.Errors[0].Message)
	}

	// Parse schema from introspection result
	schemaData, ok := gqlResp.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid introspection response")
	}

	schemaObj, ok := schemaData["__schema"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing __schema in introspection response")
	}

	schemaBytes, err := json.Marshal(schemaObj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}

	var schema GraphQLSchema
	if err := json.Unmarshal(schemaBytes, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema: %w", err)
	}

	return &schema, nil
}

// FormatGraphQLType formats a GraphQL type reference as a string
func FormatGraphQLType(typeRef GraphQLTypeRef) string {
	switch typeRef.Kind {
	case "NON_NULL":
		if typeRef.OfType != nil {
			return FormatGraphQLType(*typeRef.OfType) + "!"
		}
	case "LIST":
		if typeRef.OfType != nil {
			return "[" + FormatGraphQLType(*typeRef.OfType) + "]"
		}
	default:
		return typeRef.Name
	}
	return ""
}

// GenerateGraphQLQuery generates a query string from a type
func GenerateGraphQLQuery(schema *GraphQLSchema, typeName string, maxDepth int) (string, error) {
	// Find the type
	var targetType *GraphQLType
	for i := range schema.Types {
		if schema.Types[i].Name == typeName {
			targetType = &schema.Types[i]
			break
		}
	}

	if targetType == nil {
		return "", fmt.Errorf("type not found: %s", typeName)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("query {\n  %s {\n", typeName))

	// Generate fields
	fields := generateFields(schema, targetType, 2, maxDepth, make(map[string]bool))
	sb.WriteString(fields)

	sb.WriteString("  }\n}\n")

	return sb.String(), nil
}

// generateFields recursively generates field selection
func generateFields(schema *GraphQLSchema, gqlType *GraphQLType, indent int, maxDepth int, visited map[string]bool) string {
	if maxDepth <= 0 || visited[gqlType.Name] {
		return ""
	}

	visited[gqlType.Name] = true
	defer func() { visited[gqlType.Name] = false }()

	var sb strings.Builder
	indentStr := strings.Repeat(" ", indent)

	for _, field := range gqlType.Fields {
		// Skip __typename and other introspection fields
		if strings.HasPrefix(field.Name, "__") {
			continue
		}

		sb.WriteString(fmt.Sprintf("%s%s", indentStr, field.Name))

		// Get the actual type (unwrap NON_NULL and LIST)
		fieldType := unwrapType(field.Type)

		// Find the field's type definition
		var fieldTypeDef *GraphQLType
		for i := range schema.Types {
			if schema.Types[i].Name == fieldType.Name {
				fieldTypeDef = &schema.Types[i]
				break
			}
		}

		// If it's an OBJECT type, recursively add its fields
		if fieldTypeDef != nil && fieldTypeDef.Kind == "OBJECT" && maxDepth > 1 {
			sb.WriteString(" {\n")
			subFields := generateFields(schema, fieldTypeDef, indent+2, maxDepth-1, visited)
			sb.WriteString(subFields)
			sb.WriteString(fmt.Sprintf("%s}", indentStr))
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

// unwrapType unwraps NON_NULL and LIST wrappers to get the actual type
func unwrapType(typeRef GraphQLTypeRef) GraphQLTypeRef {
	if typeRef.Kind == "NON_NULL" || typeRef.Kind == "LIST" {
		if typeRef.OfType != nil {
			return unwrapType(*typeRef.OfType)
		}
	}
	return typeRef
}

// FormatGraphQLError formats GraphQL errors for display
func FormatGraphQLError(err GraphQLError) string {
	msg := err.Message

	if len(err.Locations) > 0 {
		loc := err.Locations[0]
		msg += fmt.Sprintf(" (line %d, column %d)", loc.Line, loc.Column)
	}

	if len(err.Path) > 0 {
		pathStr := ""
		for i, p := range err.Path {
			if i > 0 {
				pathStr += "."
			}
			pathStr += fmt.Sprintf("%v", p)
		}
		msg += fmt.Sprintf(" at path: %s", pathStr)
	}

	return msg
}

// ValidateGraphQLQuery performs basic syntax validation on a GraphQL query
func ValidateGraphQLQuery(query string) error {
	query = strings.TrimSpace(query)

	if query == "" {
		return fmt.Errorf("query cannot be empty")
	}

	// Check for balanced braces
	braceCount := 0
	for _, ch := range query {
		if ch == '{' {
			braceCount++
		} else if ch == '}' {
			braceCount--
		}
		if braceCount < 0 {
			return fmt.Errorf("unbalanced braces in query")
		}
	}

	if braceCount != 0 {
		return fmt.Errorf("unbalanced braces in query (missing %d closing braces)", braceCount)
	}

	// Check for query/mutation/subscription keyword
	lowerQuery := strings.ToLower(query)
	hasKeyword := strings.Contains(lowerQuery, "query") ||
		strings.Contains(lowerQuery, "mutation") ||
		strings.Contains(lowerQuery, "subscription") ||
		strings.HasPrefix(lowerQuery, "{")

	if !hasKeyword {
		return fmt.Errorf("query must start with 'query', 'mutation', 'subscription', or '{'")
	}

	return nil
}
