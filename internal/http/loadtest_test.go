package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLoadTestBasic(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	client := NewClient(5 * time.Second)
	config := LoadTestConfig{
		Request: Request{
			Method: "GET",
			URL:    server.URL,
		},
		Concurrency:   2,
		TotalRequests: 10,
	}

	result, err := RunLoadTest(client, config, nil)
	if err != nil {
		t.Fatalf("LoadTest failed: %v", err)
	}

	if result.TotalRequests != 10 {
		t.Errorf("Expected 10 total requests, got %d", result.TotalRequests)
	}

	if result.SuccessfulRequests != 10 {
		t.Errorf("Expected 10 successful requests, got %d", result.SuccessfulRequests)
	}

	if result.FailedRequests != 0 {
		t.Errorf("Expected 0 failed requests, got %d", result.FailedRequests)
	}

	if result.StatusCodes[200] != 10 {
		t.Errorf("Expected 10 requests with status 200, got %d", result.StatusCodes[200])
	}
}

func TestLoadTestStatistics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(5 * time.Second)
	config := LoadTestConfig{
		Request: Request{
			Method: "GET",
			URL:    server.URL,
		},
		Concurrency:   2,
		TotalRequests: 20,
	}

	result, err := RunLoadTest(client, config, nil)
	if err != nil {
		t.Fatalf("LoadTest failed: %v", err)
	}

	// Check that statistics were calculated
	if result.MinResponseTime == 0 {
		t.Error("Expected non-zero min response time")
	}

	if result.MaxResponseTime == 0 {
		t.Error("Expected non-zero max response time")
	}

	if result.AvgResponseTime == 0 {
		t.Error("Expected non-zero avg response time")
	}

	if result.MedianResponseTime == 0 {
		t.Error("Expected non-zero median response time")
	}

	if result.P95ResponseTime == 0 {
		t.Error("Expected non-zero P95 response time")
	}

	if result.P99ResponseTime == 0 {
		t.Error("Expected non-zero P99 response time")
	}

	if result.RequestsPerSecond == 0 {
		t.Error("Expected non-zero requests per second")
	}
}

func TestLoadTestWithErrors(t *testing.T) {
	successCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		successCount++
		if successCount%3 == 0 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	client := NewClient(5 * time.Second)
	config := LoadTestConfig{
		Request: Request{
			Method: "GET",
			URL:    server.URL,
		},
		Concurrency:   1,
		TotalRequests: 9,
	}

	result, err := RunLoadTest(client, config, nil)
	if err != nil {
		t.Fatalf("LoadTest failed: %v", err)
	}

	if result.TotalRequests != 9 {
		t.Errorf("Expected 9 total requests, got %d", result.TotalRequests)
	}

	// Should have mix of 200 and 500 status codes
	if result.StatusCodes[200]+result.StatusCodes[500] != 9 {
		t.Errorf("Expected 9 requests total across all status codes")
	}
}

func TestLoadTestProgressCallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(5 * time.Second)
	config := LoadTestConfig{
		Request: Request{
			Method: "GET",
			URL:    server.URL,
		},
		Concurrency:   2,
		TotalRequests: 10,
	}

	callbackCalls := 0
	progressCallback := func(completed, total int) {
		callbackCalls++
		if completed > total {
			t.Errorf("Completed (%d) should not exceed total (%d)", completed, total)
		}
	}

	result, err := RunLoadTest(client, config, progressCallback)
	if err != nil {
		t.Fatalf("LoadTest failed: %v", err)
	}

	if callbackCalls != result.TotalRequests {
		t.Errorf("Expected callback to be called %d times, got %d", result.TotalRequests, callbackCalls)
	}
}

func TestLoadTestInvalidConfig(t *testing.T) {
	client := NewClient(5 * time.Second)

	// No total requests and no duration
	config := LoadTestConfig{
		Request: Request{
			Method: "GET",
			URL:    "http://example.com",
		},
		Concurrency: 1,
	}

	_, err := RunLoadTest(client, config, nil)
	if err == nil {
		t.Error("Expected error for invalid config")
	}
}

func TestFormatLoadTestResult(t *testing.T) {
	result := &LoadTestResult{
		TotalRequests:      100,
		SuccessfulRequests: 95,
		FailedRequests:     5,
		TotalDuration:      10 * time.Second,
		MinResponseTime:    10 * time.Millisecond,
		MaxResponseTime:    500 * time.Millisecond,
		AvgResponseTime:    100 * time.Millisecond,
		MedianResponseTime: 90 * time.Millisecond,
		P95ResponseTime:    200 * time.Millisecond,
		P99ResponseTime:    300 * time.Millisecond,
		RequestsPerSecond:  10.0,
		StatusCodes: map[int]int{
			200: 95,
			500: 5,
		},
		Errors: map[string]int{
			"connection timeout": 3,
			"server error":       2,
		},
	}

	formatted := FormatLoadTestResult(result)

	// Check that formatted string contains key information
	expectedStrings := []string{
		"Load Test Results",
		"Total Requests:",
		"Successful:",
		"Failed:",
		"Response Times:",
		"Min:",
		"Max:",
		"Avg:",
		"Median",
		"95th Percentile:",
		"99th Percentile:",
		"Status Codes:",
		"Errors:",
		"Requests/sec:",
	}

	for _, expected := range expectedStrings {
		if !containsString(formatted, expected) {
			t.Errorf("Expected formatted output to contain '%s'", expected)
		}
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) != -1
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestLoadTestDurationBased(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(5 * time.Second)
	config := LoadTestConfig{
		Request: Request{
			Method: "GET",
			URL:    server.URL,
		},
		Concurrency: 2,
		Duration:    500 * time.Millisecond,
	}

	result, err := RunLoadTest(client, config, nil)
	if err != nil {
		t.Fatalf("LoadTest failed: %v", err)
	}

	// Should have completed some requests in 500ms
	if result.TotalRequests == 0 {
		t.Error("Expected at least some requests to complete")
	}

	// Duration should be approximately 500ms (allow up to 700ms for system variance)
	if result.TotalDuration < 400*time.Millisecond || result.TotalDuration > 700*time.Millisecond {
		t.Errorf("Expected duration around 500ms, got %v", result.TotalDuration)
	}
}
