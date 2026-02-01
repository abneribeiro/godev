package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestFormatJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid JSON object",
			input:   `{"name":"Alice","age":30}`,
			wantErr: false,
		},
		{
			name:    "valid JSON array",
			input:   `[1,2,3]`,
			wantErr: false,
		},
		{
			name:    "already formatted JSON",
			input:   "{\n  \"name\": \"Alice\"\n}",
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   `{invalid}`,
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   ``,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := formatJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("formatJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == "" {
				t.Error("formatJSON() returned empty string for valid JSON")
			}
			// Check that valid JSON is properly indented
			if !tt.wantErr && !strings.Contains(result, "\n") {
				t.Error("formatJSON() result is not indented")
			}
		})
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"bytes", 100, "100 B"},
		{"kilobytes", 1024, "1.00 KB"},
		{"megabytes", 1024 * 1024, "1.00 MB"},
		{"gigabytes", 1024 * 1024 * 1024, "1.00 GB"},
		{"large KB", 5 * 1024, "5.00 KB"},
		{"fractional MB", 1536 * 1024, "1.50 MB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatSize(tt.bytes)
			if result != tt.expected {
				t.Errorf("FormatSize(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"nanoseconds", 500 * time.Nanosecond, "500ns"},
		{"microseconds", 500 * time.Microsecond, "500.00µs"},
		{"milliseconds", 250 * time.Millisecond, "250ms"},
		{"seconds", 2 * time.Second, "2.00s"},
		{"sub-millisecond", 100 * time.Microsecond, "100.00µs"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.duration)
			if result != tt.want {
				t.Errorf("FormatDuration(%v) = %q, want %q", tt.duration, result, tt.want)
			}
		})
	}
}

func TestRequestToCurl(t *testing.T) {
	tests := []struct {
		name     string
		request  Request
		contains []string
	}{
		{
			name: "simple GET request",
			request: Request{
				Method: "GET",
				URL:    "https://api.example.com/users",
			},
			contains: []string{"curl", "'https://api.example.com/users'"},
		},
		{
			name: "POST request with body",
			request: Request{
				Method: "POST",
				URL:    "https://api.example.com/users",
				Body:   `{"name":"Alice"}`,
			},
			contains: []string{"curl", "-X", "POST", "-d", `'{"name":"Alice"}'`},
		},
		{
			name: "request with headers",
			request: Request{
				Method: "GET",
				URL:    "https://api.example.com/users",
				Headers: map[string]string{
					"Authorization": "Bearer token123",
					"Content-Type":  "application/json",
				},
			},
			contains: []string{"-H", "'Authorization: Bearer token123'", "'Content-Type: application/json'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RequestToCurl(tt.request)
			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("RequestToCurl() result missing %q\nGot: %s", substr, result)
				}
			}
		})
	}
}

func TestClientSendInvalidURL(t *testing.T) {
	client := NewClient(5 * time.Second)

	req := Request{
		Method: "GET",
		URL:    "not a valid url",
	}

	resp := client.Send(req)
	if resp.Error == nil {
		t.Error("Expected error for invalid URL")
	}
	if !strings.Contains(resp.Error.Error(), "invalid URL") {
		t.Errorf("Error message should mention invalid URL, got: %v", resp.Error)
	}
}

func TestClientSendSuccess(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check method
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		// Check headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("Expected Content-Type header")
		}

		// Send response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	}))
	defer server.Close()

	client := NewClient(5 * time.Second)

	req := Request{
		Method: "POST",
		URL:    server.URL,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: `{"test":"data"}`,
	}

	resp := client.Send(req)

	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	if !strings.Contains(resp.Body, "success") {
		t.Errorf("Response body doesn't contain expected data: %s", resp.Body)
	}

	if resp.ResponseTime == 0 {
		t.Error("Response time should be measured")
	}
}

func TestClientSendJSONFormatting(t *testing.T) {
	// Create test server that returns compact JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"name":"Alice","age":30,"city":"NYC"}`))
	}))
	defer server.Close()

	client := NewClient(5 * time.Second)

	req := Request{
		Method: "GET",
		URL:    server.URL,
	}

	resp := client.Send(req)

	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	// Check that JSON is formatted (contains newlines)
	if !strings.Contains(resp.Body, "\n") {
		t.Error("Response JSON should be formatted with newlines")
	}

	// Check that it contains proper indentation
	if !strings.Contains(resp.Body, "  ") {
		t.Error("Response JSON should be indented")
	}
}

func TestClientSendTimeout(t *testing.T) {
	// Create test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create client with very short timeout
	client := NewClient(50 * time.Millisecond)

	req := Request{
		Method: "GET",
		URL:    server.URL,
	}

	resp := client.Send(req)

	if resp.Error == nil {
		t.Error("Expected timeout error")
	}
}

func TestClientSendNonJSONResponse(t *testing.T) {
	// Create test server that returns plain text
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Plain text response"))
	}))
	defer server.Close()

	client := NewClient(5 * time.Second)

	req := Request{
		Method: "GET",
		URL:    server.URL,
	}

	resp := client.Send(req)

	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	// Should return original text if not JSON
	if resp.Body != "Plain text response" {
		t.Errorf("Expected plain text response, got: %s", resp.Body)
	}
}

func TestClientSendStatusCodes(t *testing.T) {
	tests := []struct {
		name           string
		serverStatus   int
		expectedStatus int
	}{
		{"200 OK", http.StatusOK, 200},
		{"404 Not Found", http.StatusNotFound, 404},
		{"500 Internal Server Error", http.StatusInternalServerError, 500},
		{"201 Created", http.StatusCreated, 201},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
			}))
			defer server.Close()

			client := NewClient(5 * time.Second)
			req := Request{
				Method: "GET",
				URL:    server.URL,
			}

			resp := client.Send(req)

			if resp.Error != nil {
				t.Fatalf("Unexpected error: %v", resp.Error)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestClientSendExactlyMaxSize(t *testing.T) {
	// Create response exactly MaxResponseSize bytes
	// This should NOT cause an error (only >MaxResponseSize should)
	responseData := make([]byte, MaxResponseSize)
	for i := range responseData {
		responseData[i] = 'A'
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(responseData)
	}))
	defer server.Close()

	client := NewClient(5 * time.Second)
	req := Request{
		Method: "GET",
		URL:    server.URL,
	}

	resp := client.Send(req)

	// Should NOT error - exactly MaxResponseSize is OK
	if resp.Error != nil {
		t.Errorf("Unexpected error for response of exactly MaxResponseSize: %v", resp.Error)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestClientSendExceedsMaxSize(t *testing.T) {
	// Create response larger than MaxResponseSize
	// This SHOULD cause an error
	responseData := make([]byte, MaxResponseSize+1000)
	for i := range responseData {
		responseData[i] = 'B'
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(responseData)
	}))
	defer server.Close()

	client := NewClient(10 * time.Second)
	req := Request{
		Method: "GET",
		URL:    server.URL,
	}

	resp := client.Send(req)

	// Should error - exceeds MaxResponseSize
	if resp.Error == nil {
		t.Error("Expected error for response exceeding MaxResponseSize")
	}

	if !strings.Contains(resp.Error.Error(), "response too large") {
		t.Errorf("Error should mention 'response too large', got: %v", resp.Error)
	}
}
