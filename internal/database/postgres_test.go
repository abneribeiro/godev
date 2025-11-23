package database

import (
	"testing"
)

func TestConnectionConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  ConnectionConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: ConnectionConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				User:     "testuser",
				SSLMode:  "disable",
			},
			wantErr: false,
		},
		{
			name: "empty host",
			config: ConnectionConfig{
				Port:     5432,
				Database: "testdb",
				User:     "testuser",
			},
			wantErr: true,
			errMsg:  "host cannot be empty",
		},
		{
			name: "invalid port - too low",
			config: ConnectionConfig{
				Host:     "localhost",
				Port:     0,
				Database: "testdb",
				User:     "testuser",
			},
			wantErr: true,
			errMsg:  "invalid port",
		},
		{
			name: "invalid port - too high",
			config: ConnectionConfig{
				Host:     "localhost",
				Port:     70000,
				Database: "testdb",
				User:     "testuser",
			},
			wantErr: true,
			errMsg:  "invalid port",
		},
		{
			name: "empty database",
			config: ConnectionConfig{
				Host: "localhost",
				Port: 5432,
				User: "testuser",
			},
			wantErr: true,
			errMsg:  "database name cannot be empty",
		},
		{
			name: "empty user",
			config: ConnectionConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
			},
			wantErr: true,
			errMsg:  "user cannot be empty",
		},
		{
			name: "invalid sslmode",
			config: ConnectionConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				User:     "testuser",
				SSLMode:  "invalid",
			},
			wantErr: true,
			errMsg:  "invalid sslmode",
		},
		{
			name: "default sslmode",
			config: ConnectionConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				User:     "testuser",
				SSLMode:  "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, should contain %q", err, tt.errMsg)
				}
			}
			// Check that empty SSLMode is set to "disable"
			if tt.name == "default sslmode" && tt.config.SSLMode != "disable" {
				t.Errorf("Validate() should set default SSLMode to 'disable', got %q", tt.config.SSLMode)
			}
		})
	}
}

func TestIsReadOnlyQuery(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{"simple select", "SELECT * FROM users", true},
		{"select with whitespace", "  SELECT * FROM users  ", true},
		{"select with newlines", "\n\nSELECT * FROM users\n", true},
		{"with clause", "WITH cte AS (SELECT 1) SELECT * FROM cte", true},
		{"explain", "EXPLAIN SELECT * FROM users", true},
		{"describe", "DESCRIBE table_name", true},
		{"show", "SHOW TABLES", true},
		{"insert", "INSERT INTO users VALUES (1)", false},
		{"update", "UPDATE users SET name='test'", false},
		{"delete", "DELETE FROM users", false},
		{"drop", "DROP TABLE users", false},
		{"create", "CREATE TABLE users (id INT)", false},
		{"lowercase select", "select * from users", true},
		{"mixed case", "SeLeCt * FROM users", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isReadOnlyQuery(tt.query)
			if result != tt.expected {
				t.Errorf("isReadOnlyQuery(%q) = %v, want %v", tt.query, result, tt.expected)
			}
		})
	}
}

func TestRemoveComments(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{
			name:     "single line comment",
			query:    "SELECT * FROM users -- this is a comment",
			expected: "SELECT * FROM users",
		},
		{
			name:     "multiple single line comments",
			query:    "SELECT * FROM users -- comment1\nWHERE id = 1 -- comment2",
			expected: "SELECT * FROM users \nWHERE id = 1", // Note: trailing space preserved
		},
		{
			name:     "multi-line comment single line",
			query:    "SELECT /* inline comment */ * FROM users",
			expected: "SELECT   * FROM users",
		},
		{
			name:     "multi-line comment multiple lines",
			query:    "SELECT /* this is\na multi-line\ncomment */ * FROM users",
			expected: "SELECT   * FROM users",
		},
		{
			name:     "mixed comments",
			query:    "SELECT /* block */ * FROM users -- line comment",
			expected: "SELECT   * FROM users",
		},
		{
			name:     "no comments",
			query:    "SELECT * FROM users WHERE id = 1",
			expected: "SELECT * FROM users WHERE id = 1",
		},
		{
			name:     "comment at start",
			query:    "-- comment\nSELECT * FROM users",
			expected: "SELECT * FROM users",
		},
		{
			name:     "multiple block comments",
			query:    "SELECT /* c1 */ * /* c2 */ FROM users",
			expected: "SELECT   *   FROM users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeComments(tt.query)
			if result != tt.expected {
				t.Errorf("removeComments(%q) = %q, want %q", tt.query, result, tt.expected)
			}
		})
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"nil", nil, "NULL"},
		{"string", "hello", "hello"},
		{"int", 42, "42"},
		{"int64", int64(42), "42"},
		{"float64", 3.14, "3.14"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"byte slice", []byte("test"), "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValue(tt.value)
			if result != tt.expected {
				t.Errorf("formatValue(%v) = %q, want %q", tt.value, result, tt.expected)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
