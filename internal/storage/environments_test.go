package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestReplaceVariables(t *testing.T) {
	variables := []Variable{
		{Key: "API_URL", Value: "https://api.example.com"},
		{Key: "API_TOKEN", Value: "secret123"},
		{Key: "PORT", Value: "8080"},
	}

	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "single variable",
			text:     "{{API_URL}}/users",
			expected: "https://api.example.com/users",
		},
		{
			name:     "multiple variables",
			text:     "{{API_URL}}/users?token={{API_TOKEN}}",
			expected: "https://api.example.com/users?token=secret123",
		},
		{
			name:     "variable with spaces",
			text:     "{{ API_URL }}/users",
			expected: "https://api.example.com/users",
		},
		{
			name:     "undefined variable",
			text:     "{{UNDEFINED}}/users",
			expected: "{{UNDEFINED}}/users",
		},
		{
			name:     "no variables",
			text:     "https://api.example.com/users",
			expected: "https://api.example.com/users",
		},
		{
			name:     "variable in JSON",
			text:     `{"url":"{{API_URL}}","token":"{{API_TOKEN}}"}`,
			expected: `{"url":"https://api.example.com","token":"secret123"}`,
		},
		{
			name:     "same variable multiple times",
			text:     "{{API_URL}}/users and {{API_URL}}/posts",
			expected: "https://api.example.com/users and https://api.example.com/posts",
		},
		{
			name:     "empty text",
			text:     "",
			expected: "",
		},
		{
			name:     "malformed brackets",
			text:     "{API_URL}",
			expected: "{API_URL}",
		},
		{
			name:     "nested-looking brackets",
			text:     "{{{{API_URL}}}}",
			expected: "{{{{API_URL}}}}", // Regex doesn't match nested brackets
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReplaceVariables(tt.text, variables)
			if result != tt.expected {
				t.Errorf("ReplaceVariables(%q) = %q, want %q", tt.text, result, tt.expected)
			}
		})
	}
}

func TestReplaceVariablesPerformance(t *testing.T) {
	// Create many variables to test map lookup performance
	variables := make([]Variable, 100)
	for i := 0; i < 100; i++ {
		variables[i] = Variable{
			Key:   fmt.Sprintf("VAR_%d", i),
			Value: fmt.Sprintf("value_%d", i),
		}
	}
	// Add specific test variables
	variables = append(variables, Variable{Key: "A", Value: "valueA"})
	variables = append(variables, Variable{Key: "B", Value: "valueB"})
	variables = append(variables, Variable{Key: "C", Value: "valueC"})

	text := "{{A}} {{B}} {{C}} {{NOTFOUND}}"

	// Should be fast with map lookup
	result := ReplaceVariables(text, variables)

	expected := "valueA valueB valueC {{NOTFOUND}}"
	if result != expected {
		t.Errorf("ReplaceVariables() = %q, want %q", result, expected)
	}
}

func TestStorageSaveEnvironments(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Set HOME to temp directory
	os.Setenv("HOME", tmpDir)

	storage := &Storage{}

	config := &EnvironmentConfig{
		Version: "0.4.0",
		Environments: []Environment{
			{
				Name: "dev",
				Variables: []Variable{
					{Key: "API_URL", Value: "https://dev.api.com"},
				},
			},
		},
		ActiveEnvironment: "dev",
	}

	err := storage.SaveEnvironments(config)
	if err != nil {
		t.Fatalf("SaveEnvironments() error = %v", err)
	}

	// Check file was created
	envPath := filepath.Join(tmpDir, ".godev", "environments.json")
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		t.Fatal("environments.json was not created")
	}

	// Check file permissions
	info, err := os.Stat(envPath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("File permissions = %o, want 0600", info.Mode().Perm())
	}

	// Load back and verify
	loaded, err := storage.LoadEnvironments()
	if err != nil {
		t.Fatalf("LoadEnvironments() error = %v", err)
	}

	if loaded.ActiveEnvironment != "dev" {
		t.Errorf("ActiveEnvironment = %q, want %q", loaded.ActiveEnvironment, "dev")
	}

	if len(loaded.Environments) != 1 {
		t.Fatalf("Expected 1 environment, got %d", len(loaded.Environments))
	}

	if loaded.Environments[0].Name != "dev" {
		t.Errorf("Environment name = %q, want %q", loaded.Environments[0].Name, "dev")
	}
}

func TestStorageAddEnvironment(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	os.Setenv("HOME", tmpDir)

	storage := &Storage{}

	// Add first environment
	err := storage.AddEnvironment("dev")
	if err != nil {
		t.Fatalf("AddEnvironment() error = %v", err)
	}

	// Add second environment
	err = storage.AddEnvironment("prod")
	if err != nil {
		t.Fatalf("AddEnvironment() error = %v", err)
	}

	// Try to add duplicate
	err = storage.AddEnvironment("dev")
	if err == nil {
		t.Error("Expected error when adding duplicate environment")
	}

	// Load and verify
	config, err := storage.LoadEnvironments()
	if err != nil {
		t.Fatalf("LoadEnvironments() error = %v", err)
	}

	if len(config.Environments) != 2 {
		t.Errorf("Expected 2 environments, got %d", len(config.Environments))
	}

	// First environment should be active
	if config.ActiveEnvironment != "dev" {
		t.Errorf("ActiveEnvironment = %q, want %q", config.ActiveEnvironment, "dev")
	}
}

func TestStorageAddVariable(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	os.Setenv("HOME", tmpDir)

	storage := &Storage{}

	// Create environment first
	err := storage.AddEnvironment("dev")
	if err != nil {
		t.Fatalf("AddEnvironment() error = %v", err)
	}

	// Add variable
	err = storage.AddVariable("dev", "API_URL", "https://api.dev.com")
	if err != nil {
		t.Fatalf("AddVariable() error = %v", err)
	}

	// Add another variable
	err = storage.AddVariable("dev", "API_TOKEN", "secret123")
	if err != nil {
		t.Fatalf("AddVariable() error = %v", err)
	}

	// Update existing variable
	err = storage.AddVariable("dev", "API_URL", "https://api.updated.com")
	if err != nil {
		t.Fatalf("AddVariable() error = %v", err)
	}

	// Load and verify
	config, err := storage.LoadEnvironments()
	if err != nil {
		t.Fatalf("LoadEnvironments() error = %v", err)
	}

	env := config.Environments[0]
	if len(env.Variables) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(env.Variables))
	}

	// Check updated value
	for _, v := range env.Variables {
		if v.Key == "API_URL" && v.Value != "https://api.updated.com" {
			t.Errorf("Variable value = %q, want %q", v.Value, "https://api.updated.com")
		}
	}
}

func TestStorageDeleteVariable(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	os.Setenv("HOME", tmpDir)

	storage := &Storage{}

	// Setup
	storage.AddEnvironment("dev")
	storage.AddVariable("dev", "VAR1", "value1")
	storage.AddVariable("dev", "VAR2", "value2")

	// Delete variable
	err := storage.DeleteVariable("dev", "VAR1")
	if err != nil {
		t.Fatalf("DeleteVariable() error = %v", err)
	}

	// Verify
	config, err := storage.LoadEnvironments()
	if err != nil {
		t.Fatalf("LoadEnvironments() error = %v", err)
	}

	env := config.Environments[0]
	if len(env.Variables) != 1 {
		t.Errorf("Expected 1 variable, got %d", len(env.Variables))
	}

	if env.Variables[0].Key != "VAR2" {
		t.Errorf("Remaining variable = %q, want %q", env.Variables[0].Key, "VAR2")
	}
}

func TestStorageSetActiveEnvironment(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	os.Setenv("HOME", tmpDir)

	storage := &Storage{}

	// Create environments
	storage.AddEnvironment("dev")
	storage.AddEnvironment("prod")

	// Set active
	err := storage.SetActiveEnvironment("prod")
	if err != nil {
		t.Fatalf("SetActiveEnvironment() error = %v", err)
	}

	// Verify
	config, err := storage.LoadEnvironments()
	if err != nil {
		t.Fatalf("LoadEnvironments() error = %v", err)
	}

	if config.ActiveEnvironment != "prod" {
		t.Errorf("ActiveEnvironment = %q, want %q", config.ActiveEnvironment, "prod")
	}

	// Try to set non-existent environment
	err = storage.SetActiveEnvironment("nonexistent")
	if err == nil {
		t.Error("Expected error when setting non-existent environment")
	}
}

func TestStorageGetActiveEnvironmentVariables(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	os.Setenv("HOME", tmpDir)

	storage := &Storage{}

	// Setup
	storage.AddEnvironment("dev")
	storage.AddVariable("dev", "VAR1", "value1")
	storage.AddVariable("dev", "VAR2", "value2")
	storage.SetActiveEnvironment("dev")

	// Get variables
	vars, err := storage.GetActiveEnvironmentVariables()
	if err != nil {
		t.Fatalf("GetActiveEnvironmentVariables() error = %v", err)
	}

	if len(vars) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(vars))
	}
}
