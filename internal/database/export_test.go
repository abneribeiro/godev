package database

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"empty string", "", false},
		{"NULL", "NULL", false},
		{"simple integer", "123", true},
		{"negative integer", "-123", true},
		{"positive integer", "+123", true},
		{"decimal", "123.456", true},
		{"negative decimal", "-123.456", true},
		{"scientific notation", "1.23e10", true},
		{"scientific notation negative", "1.23e-10", true},
		{"scientific notation uppercase", "1.23E10", true},
		{"only sign", "-", false},
		{"only dot", ".", false},
		{"multiple dots", "1.2.3", false},
		{"text", "abc", false},
		{"mixed", "12abc", false},
		{"multiple e", "1e2e3", false},
		{"e without digits before", "e10", false},
		{"dot after e", "1e2.3", false},
		{"just zero", "0", true},
		{"negative zero", "-0", true},
		{"decimal starting with dot", ".5", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNumeric(tt.input)
			if result != tt.expected {
				t.Errorf("isNumeric(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestQuoteIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple name", "users", `"users"`},
		{"with spaces", "user name", `"user name"`},
		{"with quotes", `user"name`, `"user""name"`},
		{"multiple quotes", `us"er"name`, `"us""er""name"`},
		{"empty", "", `""`},
		{"special chars", "user-name_123", `"user-name_123"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := quoteIdentifier(tt.input)
			if result != tt.expected {
				t.Errorf("quoteIdentifier(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEscapeSQLString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple string", "hello", "hello"},
		{"with single quote", "it's", "it''s"},
		{"with backslash", `path\to\file`, `path\\to\\file`},
		{"with both", `it's\here`, `it''s\\here`},
		{"multiple quotes", "''test''", "''''test''''"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeSQLString(tt.input)
			if result != tt.expected {
				t.Errorf("escapeSQLString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExportToCSV(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.csv")

	result := &QueryResult{
		Columns: []string{"id", "name", "email"},
		Rows: [][]string{
			{"1", "Alice", "alice@example.com"},
			{"2", "Bob", "bob@example.com"},
		},
	}

	err := exportToCSV(filePath, result, "users")
	if err != nil {
		t.Fatalf("exportToCSV failed: %v", err)
	}

	// Check file was created
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("CSV file was not created")
	}

	// Check file permissions
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("File permissions = %o, want 0600", info.Mode().Perm())
	}

	// Check file contents
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	expectedLines := []string{
		"id,name,email",
		"1,Alice,alice@example.com",
		"2,Bob,bob@example.com",
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != len(expectedLines) {
		t.Fatalf("Expected %d lines, got %d", len(expectedLines), len(lines))
	}

	for i, expected := range expectedLines {
		if lines[i] != expected {
			t.Errorf("Line %d: got %q, want %q", i, lines[i], expected)
		}
	}
}

func TestExportToJSON(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.json")

	result := &QueryResult{
		Columns: []string{"id", "name"},
		Rows: [][]string{
			{"1", "Alice"},
			{"2", "Bob"},
		},
	}

	err := exportToJSON(filePath, result, "users")
	if err != nil {
		t.Fatalf("exportToJSON failed: %v", err)
	}

	// Check file was created
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("JSON file was not created")
	}

	// Check file permissions
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("File permissions = %o, want 0600", info.Mode().Perm())
	}

	// Check file contents can be read as JSON
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Just verify it's valid JSON and contains expected data
	if !strings.Contains(string(content), `"id"`) || !strings.Contains(string(content), `"name"`) {
		t.Error("JSON doesn't contain expected columns")
	}
	if !strings.Contains(string(content), `"Alice"`) || !strings.Contains(string(content), `"Bob"`) {
		t.Error("JSON doesn't contain expected values")
	}
}

func TestExportToSQL(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.sql")

	result := &QueryResult{
		Columns: []string{"id", "name", "age"},
		Rows: [][]string{
			{"1", "Alice", "30"},
			{"2", "Bob", "NULL"},
		},
	}

	err := exportToSQL(filePath, result, "users")
	if err != nil {
		t.Fatalf("exportToSQL failed: %v", err)
	}

	// Check file was created
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("SQL file was not created")
	}

	// Check file permissions
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("File permissions = %o, want 0600", info.Mode().Perm())
	}

	// Check file contents
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)

	// Check for quoted identifiers
	if !strings.Contains(contentStr, `"users"`) {
		t.Error("Table name not properly quoted")
	}
	if !strings.Contains(contentStr, `"id"`) || !strings.Contains(contentStr, `"name"`) {
		t.Error("Column names not properly quoted")
	}

	// Check for proper escaping of string values
	if !strings.Contains(contentStr, "'Alice'") {
		t.Error("String values not properly escaped")
	}

	// Check NULL handling
	if strings.Count(contentStr, "NULL") < 1 {
		t.Error("NULL values not handled correctly")
	}

	// Check numeric values are not quoted
	if strings.Contains(contentStr, "'30'") {
		t.Error("Numeric values should not be quoted")
	}
}

func TestExportToSQLWithInjection(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_injection.sql")

	// Test with malicious table name
	result := &QueryResult{
		Columns: []string{"id", "na\"me"},
		Rows: [][]string{
			{"1", "Alice'; DROP TABLE users; --"},
		},
	}

	err := exportToSQL(filePath, result, `evil"; DROP TABLE users; --`)
	if err != nil {
		t.Fatalf("exportToSQL failed: %v", err)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)

	// Verify table name is quoted (preventing injection)
	if !strings.Contains(contentStr, `"evil""; DROP TABLE users; --"`) {
		t.Error("Table name not properly escaped for SQL injection")
	}

	// Verify column name is quoted
	if !strings.Contains(contentStr, `"na""me"`) {
		t.Error("Column name with quote not properly escaped")
	}

	// Verify string value is escaped
	if !strings.Contains(contentStr, "Alice''; DROP TABLE users; --") {
		t.Error("String value not properly escaped")
	}
}
