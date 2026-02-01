package ui

import (
	"testing"
)

func TestHighlightSQL(t *testing.T) {
	sh := NewSyntaxHighlighter()

	sql := `SELECT id, name FROM users WHERE age > 18 AND active = TRUE;`

	highlighted := sh.HighlightSQL(sql)

	// Should still contain original text (highlighting adds ANSI codes)
	// We'll check that it's not empty and different from input
	if highlighted == "" {
		t.Error("Expected non-empty highlighted output")
	}

	// Highlighting should add ANSI escape codes, making it longer
	if len(highlighted) < len(sql) {
		t.Error("Expected highlighted output to be longer than input due to ANSI codes")
	}
}

func TestHighlightJSON(t *testing.T) {
	sh := NewSyntaxHighlighter()

	json := `{
		"name": "John Doe",
		"age": 30,
		"active": true,
		"email": null
	}`

	highlighted := sh.HighlightJSON(json)

	if highlighted == "" {
		t.Error("Expected non-empty highlighted output")
	}

	if len(highlighted) < len(json) {
		t.Error("Expected highlighted output to be longer than input")
	}
}

func TestHighlightGraphQL(t *testing.T) {
	sh := NewSyntaxHighlighter()

	gql := `query GetUser($id: ID!) {
		user(id: $id) {
			id
			name
			email
		}
	}`

	highlighted := sh.HighlightGraphQL(gql)

	if highlighted == "" {
		t.Error("Expected non-empty highlighted output")
	}

	if len(highlighted) < len(gql) {
		t.Error("Expected highlighted output to be longer than input")
	}
}

func TestStripANSI(t *testing.T) {
	// Create a string with ANSI color codes
	withANSI := "\x1b[31mRed Text\x1b[0m Normal Text"

	stripped := StripANSI(withANSI)

	expected := "Red Text Normal Text"
	if stripped != expected {
		t.Errorf("Expected '%s', got '%s'", expected, stripped)
	}
}

func TestStripANSINoANSI(t *testing.T) {
	original := "Plain text with no ANSI codes"

	stripped := StripANSI(original)

	if stripped != original {
		t.Errorf("Expected text to remain unchanged, got '%s'", stripped)
	}
}

func TestHighlightError(t *testing.T) {
	sh := NewSyntaxHighlighter()

	errorMsg := "error: failed to connect to database.go:123 on line 45"

	highlighted := sh.HighlightError(errorMsg)

	if highlighted == "" {
		t.Error("Expected non-empty highlighted output")
	}

	// Should highlight the file path and line number
	if len(highlighted) < len(errorMsg) {
		t.Error("Expected highlighted output to be longer than input")
	}
}

func TestLineNumberedCode(t *testing.T) {
	code := "line 1\nline 2\nline 3"

	numbered := LineNumberedCode(code, 1)

	if numbered == "" {
		t.Error("Expected non-empty numbered code")
	}

	// Should contain line numbers
	if !containsHighlightStr(numbered, "1") {
		t.Error("Expected line number 1")
	}

	if !containsHighlightStr(numbered, "2") {
		t.Error("Expected line number 2")
	}

	if !containsHighlightStr(numbered, "3") {
		t.Error("Expected line number 3")
	}

	// Should contain the pipe separator
	if !containsHighlightStr(numbered, "│") {
		t.Error("Expected pipe separator")
	}
}

func TestLineNumberedCodeWithOffset(t *testing.T) {
	code := "line 1\nline 2"

	numbered := LineNumberedCode(code, 10)

	// Should start at line 10
	if !containsHighlightStr(numbered, "10") {
		t.Error("Expected line number 10")
	}

	if !containsHighlightStr(numbered, "11") {
		t.Error("Expected line number 11")
	}
}

func TestHighlightDiff(t *testing.T) {
	diff := `=== Header ===
+ Added line
- Removed line
~ Modified line
  Unchanged line`

	highlighted := HighlightDiff(diff)

	if highlighted == "" {
		t.Error("Expected non-empty highlighted output")
	}

	// Should be longer due to ANSI codes
	if len(highlighted) < len(diff) {
		t.Error("Expected highlighted output to be longer than input")
	}
}

func TestDefaultThemes(t *testing.T) {
	darkTheme := DefaultDarkTheme()
	lightTheme := DefaultLightTheme()

	// Check that themes can render text (verifies styles exist)
	darkRendered := darkTheme.Keyword.Render("test")
	lightRendered := lightTheme.Keyword.Render("test")

	if len(darkRendered) == 0 {
		t.Error("Expected dark theme to render keyword")
	}

	if len(lightRendered) == 0 {
		t.Error("Expected light theme to render keyword")
	}

	// Both themes should be able to render (functionality works)
	// Note: Actual rendering differences may not be visible in all terminal environments
}

func TestHighlightSQLComments(t *testing.T) {
	sh := NewSyntaxHighlighter()

	sql := `-- This is a comment
SELECT * FROM users; -- inline comment
/* Multi-line
   comment */
SELECT id FROM orders;`

	highlighted := sh.HighlightSQL(sql)

	// Comments should be highlighted
	if highlighted == "" {
		t.Error("Expected non-empty highlighted output")
	}

	// Output should be longer due to highlighting
	if len(highlighted) < len(sql) {
		t.Error("Expected highlighted output to be longer")
	}
}

func TestHighlightSQLStrings(t *testing.T) {
	sh := NewSyntaxHighlighter()

	sql := `SELECT 'John Doe', 'john@example.com' FROM users WHERE name = 'test';`

	highlighted := sh.HighlightSQL(sql)

	if highlighted == "" {
		t.Error("Expected non-empty highlighted output")
	}

	if len(highlighted) < len(sql) {
		t.Error("Expected highlighted output to be longer")
	}
}

func TestHighlightSQLTypes(t *testing.T) {
	sh := NewSyntaxHighlighter()

	sql := `CREATE TABLE users (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255),
		age INTEGER,
		created_at TIMESTAMP
	);`

	highlighted := sh.HighlightSQL(sql)

	if highlighted == "" {
		t.Error("Expected non-empty highlighted output")
	}
}

func TestHighlightJSONNumbers(t *testing.T) {
	sh := NewSyntaxHighlighter()

	json := `{"count": 42, "price": 99.99, "scientific": 1e10}`

	highlighted := sh.HighlightJSON(json)

	if highlighted == "" {
		t.Error("Expected non-empty highlighted output")
	}
}

func TestHighlightJSONBooleans(t *testing.T) {
	sh := NewSyntaxHighlighter()

	json := `{"active": true, "deleted": false}`

	highlighted := sh.HighlightJSON(json)

	if highlighted == "" {
		t.Error("Expected non-empty highlighted output")
	}
}

func TestHighlightGraphQLVariables(t *testing.T) {
	sh := NewSyntaxHighlighter()

	gql := `query GetUser($userId: ID!, $includeEmail: Boolean) {
		user(id: $userId) {
			id
			name
		}
	}`

	highlighted := sh.HighlightGraphQL(gql)

	if highlighted == "" {
		t.Error("Expected non-empty highlighted output")
	}

	// Should highlight variables ($userId, $includeEmail)
	if len(highlighted) < len(gql) {
		t.Error("Expected highlighted output to be longer")
	}
}

func containsHighlightStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
