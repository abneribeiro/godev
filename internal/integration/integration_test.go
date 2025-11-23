// +build integration

package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	httpclient "github.com/abneribeiro/godev/internal/http"
	"github.com/abneribeiro/godev/internal/storage"
)

// TestHTTPRequestFlow tests the complete flow of creating and sending an HTTP request
func TestHTTPRequestFlow(t *testing.T) {
	// Create test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type header")
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"status":  "success",
			"message": "Request processed",
			"data":    map[string]string{"id": "123"},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create HTTP client
	client := httpclient.NewClient(30 * time.Second)

	// Build request
	req := httpclient.Request{
		Method: "POST",
		URL:    server.URL + "/api/test",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: `{"name": "test", "value": "123"}`,
	}

	// Send request
	resp := client.Send(req)

	// Verify response
	if resp.Error != nil {
		t.Fatalf("Request failed: %v", resp.Error)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	if resp.Body == "" {
		t.Error("Expected non-empty response body")
	}

	// Verify response contains expected data
	var responseData map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Body), &responseData); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if responseData["status"] != "success" {
		t.Errorf("Expected status 'success', got %v", responseData["status"])
	}
}

// TestRequestPersistence tests saving and loading requests
func TestRequestPersistence(t *testing.T) {
	// Create temporary directory for test storage
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Create storage
	store, err := storage.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Save request
	testURL := "https://api.example.com/users"
	testMethod := "GET"
	testName := "GET https://api.example.com/users"
	testHeaders := map[string]string{
		"Authorization": "Bearer token123",
		"Accept":        "application/json",
	}
	testBody := `{"filter": "active"}`
	testQueryParams := map[string]string{
		"page":  "1",
		"limit": "10",
	}

	err = store.SaveRequest(testName, testMethod, testURL, testHeaders, testBody, testQueryParams)
	if err != nil {
		t.Fatalf("Failed to save request: %v", err)
	}

	// Load requests
	requests := store.GetRequests()
	if len(requests) == 0 {
		t.Fatal("Expected at least one saved request")
	}

	// Verify saved request
	found := false
	for _, req := range requests {
		if req.Name == testName {
			found = true
			if req.Method != testMethod {
				t.Errorf("Expected method %s, got %s", testMethod, req.Method)
			}
			if req.URL != testURL {
				t.Errorf("Expected URL %s, got %s", testURL, req.URL)
			}
			if req.Headers["Authorization"] != testHeaders["Authorization"] {
				t.Error("Headers not preserved correctly")
			}
			if req.Body != testBody {
				t.Error("Body not preserved correctly")
			}
			if len(req.QueryParams) != 2 {
				t.Errorf("Expected 2 query params, got %d", len(req.QueryParams))
			}
			break
		}
	}

	if !found {
		t.Error("Saved request not found after reload")
	}
}

// TestEnvironmentVariableSubstitution tests environment variable replacement
func TestEnvironmentVariableSubstitution(t *testing.T) {
	// Create temporary directory for test storage
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Create storage
	store, err := storage.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Create environment
	err = store.AddEnvironment("test")
	if err != nil {
		t.Fatalf("Failed to create environment: %v", err)
	}

	// Add variables
	err = store.AddVariable("test", "API_URL", "https://api.test.com")
	if err != nil {
		t.Fatalf("Failed to add variable: %v", err)
	}
	err = store.AddVariable("test", "API_TOKEN", "test_token_123")
	if err != nil {
		t.Fatalf("Failed to add variable: %v", err)
	}

	// Set as active
	err = store.SetActiveEnvironment("test")
	if err != nil {
		t.Fatalf("Failed to set active environment: %v", err)
	}

	// Get active variables
	vars, err := store.GetActiveEnvironmentVariables()
	if err != nil {
		t.Fatalf("Failed to get active variables: %v", err)
	}

	// Test variable replacement
	testText := "{{API_URL}}/users?token={{API_TOKEN}}"
	result := storage.ReplaceVariables(testText, vars)

	expected := "https://api.test.com/users?token=test_token_123"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// Test undefined variable (should remain unchanged)
	testText2 := "{{API_URL}}/{{UNDEFINED}}"
	result2 := storage.ReplaceVariables(testText2, vars)
	expected2 := "https://api.test.com/{{UNDEFINED}}"
	if result2 != expected2 {
		t.Errorf("Expected %s, got %s", expected2, result2)
	}
}

// TestRequestHistory tests request history tracking
func TestRequestHistory(t *testing.T) {
	// Create temporary directory for test storage
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Create storage
	store, err := storage.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Add request to history
	err = store.AddToHistory(
		"GET",
		"https://api.example.com/users",
		map[string]string{"Accept": "application/json"},
		"",
		map[string]string{"page": "1"},
		200,
		"200 OK",
		`{"users": []}`,
		150,
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to add to history: %v", err)
	}

	// Get history
	history := store.GetHistory()
	if len(history) == 0 {
		t.Fatal("Expected at least one history item")
	}

	// Verify history item
	item := history[0]
	if item.Method != "GET" {
		t.Errorf("Expected method GET, got %s", item.Method)
	}
	if item.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", item.StatusCode)
	}
	if item.ResponseTime != 150 {
		t.Errorf("Expected response time 150ms, got %d", item.ResponseTime)
	}
}

// TestCompleteWorkflow tests a complete end-to-end workflow
func TestCompleteWorkflow(t *testing.T) {
	// Setup test environment
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "success"})
	}))
	defer server.Close()

	// 1. Create storage
	store, err := storage.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// 2. Create environment
	err = store.AddEnvironment("dev")
	if err != nil {
		t.Fatalf("Failed to create environment: %v", err)
	}

	// 3. Add variable
	err = store.AddVariable("dev", "BASE_URL", server.URL)
	if err != nil {
		t.Fatalf("Failed to add variable: %v", err)
	}

	// 4. Set active environment
	err = store.SetActiveEnvironment("dev")
	if err != nil {
		t.Fatalf("Failed to set active environment: %v", err)
	}

	// 5. Get variables and replace in URL
	vars, err := store.GetActiveEnvironmentVariables()
	if err != nil {
		t.Fatalf("Failed to get variables: %v", err)
	}

	url := "{{BASE_URL}}/api/test"
	finalURL := storage.ReplaceVariables(url, vars)

	// 6. Create and send request
	client := httpclient.NewClient(30 * time.Second)
	req := httpclient.Request{
		Method:  "GET",
		URL:     finalURL,
		Headers: map[string]string{"Accept": "application/json"},
		Body:    "",
	}

	resp := client.Send(req)
	if resp.Error != nil {
		t.Fatalf("Request failed: %v", resp.Error)
	}

	// 7. Save to history
	err = store.AddToHistory(
		req.Method,
		req.URL,
		req.Headers,
		req.Body,
		map[string]string{},
		resp.StatusCode,
		resp.Status,
		resp.Body,
		resp.ResponseTime.Milliseconds(),
		resp.Error,
	)
	if err != nil {
		t.Fatalf("Failed to add to history: %v", err)
	}

	// 8. Save request
	err = store.SaveRequest(
		"Test API Request",
		req.Method,
		url, // Save with template variable
		req.Headers,
		req.Body,
		map[string]string{},
	)
	if err != nil {
		t.Fatalf("Failed to save request: %v", err)
	}

	// 9. Verify everything was saved
	requests := store.GetRequests()
	if len(requests) == 0 {
		t.Error("Expected saved request")
	}

	history := store.GetHistory()
	if len(history) == 0 {
		t.Error("Expected history item")
	}

	config, err := store.LoadEnvironments()
	if err != nil {
		t.Fatalf("Failed to load environments: %v", err)
	}
	if len(config.Environments) == 0 {
		t.Error("Expected environment")
	}

	t.Log("Complete workflow test passed successfully!")
}

// TestCurlExport tests cURL command export
func TestCurlExport(t *testing.T) {
	req := httpclient.Request{
		Method: "POST",
		URL:    "https://api.example.com/users",
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer token123",
		},
		Body: `{"name": "John", "email": "john@example.com"}`,
	}

	curlCmd := httpclient.RequestToCurl(req)

	// Print for debugging
	t.Logf("Generated cURL command:\n%s", curlCmd)

	// Verify cURL command contains expected parts
	if curlCmd == "" {
		t.Fatal("Expected non-empty cURL command")
	}

	// Check for key components (simplified check)
	if !contains(curlCmd, "curl") {
		t.Error("cURL command missing 'curl'")
	}
	if !contains(curlCmd, "https://api.example.com/users") {
		t.Error("cURL command missing URL")
	}
	if !contains(curlCmd, "POST") {
		t.Error("cURL command missing POST method")
	}
	if !contains(curlCmd, "Content-Type") {
		t.Error("cURL command missing Content-Type header")
	}
	if !contains(curlCmd, "Authorization") {
		t.Error("cURL command missing Authorization header")
	}
	if !contains(curlCmd, "John") {
		t.Error("cURL command missing body content")
	}
}

// Helper function
func contains(s, substr string) bool {
	// Simple contains check
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// TestStorageDirectoryPermissions tests that config directories have correct permissions
func TestStorageDirectoryPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Create storage
	_, err := storage.NewStorage()
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Check directory permissions
	configDir := filepath.Join(tmpDir, ".godev")
	info, err := os.Stat(configDir)
	if err != nil {
		t.Fatalf("Failed to stat config directory: %v", err)
	}

	perm := info.Mode().Perm()
	expected := os.FileMode(0700)
	if perm != expected {
		t.Errorf("Expected directory permissions %o, got %o", expected, perm)
	}

	// Check file permissions
	configFile := filepath.Join(configDir, "config.json")
	fileInfo, err := os.Stat(configFile)
	if err != nil {
		t.Fatalf("Failed to stat config file: %v", err)
	}

	filePerm := fileInfo.Mode().Perm()
	expectedFile := os.FileMode(0600)
	if filePerm != expectedFile {
		t.Errorf("Expected file permissions %o, got %o", expectedFile, filePerm)
	}
}
