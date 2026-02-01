package http

import (
	"testing"
	"time"
)

func TestCompareResponsesStatusCode(t *testing.T) {
	old := Response{StatusCode: 200}
	new := Response{StatusCode: 404}

	diff := CompareResponses(old, new)

	if diff.StatusCodeDiff == nil {
		t.Fatal("Expected status code diff")
	}

	if diff.StatusCodeDiff.Old != "200" {
		t.Errorf("Expected old status '200', got '%s'", diff.StatusCodeDiff.Old)
	}

	if diff.StatusCodeDiff.New != "404" {
		t.Errorf("Expected new status '404', got '%s'", diff.StatusCodeDiff.New)
	}
}

func TestCompareResponsesNoStatusCodeDiff(t *testing.T) {
	old := Response{StatusCode: 200}
	new := Response{StatusCode: 200}

	diff := CompareResponses(old, new)

	if diff.StatusCodeDiff != nil {
		t.Error("Expected no status code diff for identical status codes")
	}
}

func TestCompareResponsesHeaders(t *testing.T) {
	old := Response{
		Headers: map[string][]string{
			"Content-Type":  {"application/json"},
			"X-Old-Header":  {"old-value"},
			"Shared-Header": {"value1"},
		},
	}

	new := Response{
		Headers: map[string][]string{
			"Content-Type":  {"application/xml"},
			"X-New-Header":  {"new-value"},
			"Shared-Header": {"value1"},
		},
	}

	diff := CompareResponses(old, new)

	// Content-Type changed
	if diff.HeadersDiff["Content-Type"] == nil {
		t.Error("Expected Content-Type diff")
	}
	if diff.HeadersDiff["Content-Type"].Old != "application/json" {
		t.Error("Expected old Content-Type 'application/json'")
	}
	if diff.HeadersDiff["Content-Type"].New != "application/xml" {
		t.Error("Expected new Content-Type 'application/xml'")
	}

	// X-Old-Header removed
	if diff.HeadersDiff["X-Old-Header"] == nil {
		t.Error("Expected X-Old-Header diff")
	}

	// X-New-Header added
	if diff.HeadersDiff["X-New-Header"] == nil {
		t.Error("Expected X-New-Header diff")
	}

	// Shared-Header unchanged - should not be in diff
	if diff.HeadersDiff["Shared-Header"] != nil {
		t.Error("Expected no diff for unchanged header")
	}
}

func TestCompareResponsesJSON(t *testing.T) {
	old := Response{
		Body: `{"name": "John", "age": 30, "city": "NYC"}`,
	}

	new := Response{
		Body: `{"name": "John", "age": 31, "country": "USA"}`,
	}

	diff := CompareResponses(old, new)

	if diff.BodyDiff == nil {
		t.Fatal("Expected body diff")
	}

	if diff.BodyDiff.Type != "json" {
		t.Errorf("Expected diff type 'json', got '%s'", diff.BodyDiff.Type)
	}

	if len(diff.BodyDiff.Changes) == 0 {
		t.Error("Expected at least one change")
	}

	// Check for specific changes
	hasAgeChange := false
	hasCityRemoval := false
	hasCountryAddition := false

	for _, change := range diff.BodyDiff.Changes {
		if change.Path == "age" && change.Type == "modified" {
			hasAgeChange = true
		}
		if change.Path == "city" && change.Type == "removed" {
			hasCityRemoval = true
		}
		if change.Path == "country" && change.Type == "added" {
			hasCountryAddition = true
		}
	}

	if !hasAgeChange {
		t.Error("Expected to find age modification")
	}
	if !hasCityRemoval {
		t.Error("Expected to find city removal")
	}
	if !hasCountryAddition {
		t.Error("Expected to find country addition")
	}
}

func TestCompareResponsesNestedJSON(t *testing.T) {
	old := Response{
		Body: `{"user": {"name": "John", "email": "john@old.com"}}`,
	}

	new := Response{
		Body: `{"user": {"name": "John", "email": "john@new.com"}}`,
	}

	diff := CompareResponses(old, new)

	if len(diff.BodyDiff.Changes) == 0 {
		t.Error("Expected at least one change")
	}

	// Should detect change in user.email
	hasEmailChange := false
	for _, change := range diff.BodyDiff.Changes {
		if change.Path == "user.email" && change.Type == "modified" {
			hasEmailChange = true
		}
	}

	if !hasEmailChange {
		t.Error("Expected to find user.email modification")
	}
}

func TestCompareResponsesArrays(t *testing.T) {
	old := Response{
		Body: `{"items": [1, 2, 3]}`,
	}

	new := Response{
		Body: `{"items": [1, 2, 3, 4]}`,
	}

	diff := CompareResponses(old, new)

	if len(diff.BodyDiff.Changes) == 0 {
		t.Error("Expected at least one change")
	}

	// Should detect new item at index 3
	hasNewItem := false
	for _, change := range diff.BodyDiff.Changes {
		if change.Path == "items[3]" && change.Type == "added" {
			hasNewItem = true
		}
	}

	if !hasNewItem {
		t.Error("Expected to find new array item")
	}
}

func TestCompareResponsesTextDiff(t *testing.T) {
	old := Response{
		Body: "Line 1\nLine 2\nLine 3",
	}

	new := Response{
		Body: "Line 1\nLine 2 modified\nLine 3\nLine 4",
	}

	diff := CompareResponses(old, new)

	if diff.BodyDiff.Type != "text" {
		t.Errorf("Expected diff type 'text', got '%s'", diff.BodyDiff.Type)
	}

	if len(diff.BodyDiff.Changes) < 2 {
		t.Errorf("Expected at least 2 changes, got %d", len(diff.BodyDiff.Changes))
	}
}

func TestCompareResponsesResponseTime(t *testing.T) {
	old := Response{
		ResponseTime: 100 * time.Millisecond,
	}

	new := Response{
		ResponseTime: 200 * time.Millisecond,
	}

	diff := CompareResponses(old, new)

	if diff.ResponseTimeDiff == nil {
		t.Fatal("Expected response time diff")
	}

	if diff.ResponseTimeDiff.OldMs != 100 {
		t.Errorf("Expected old time 100ms, got %dms", diff.ResponseTimeDiff.OldMs)
	}

	if diff.ResponseTimeDiff.NewMs != 200 {
		t.Errorf("Expected new time 200ms, got %dms", diff.ResponseTimeDiff.NewMs)
	}

	if diff.ResponseTimeDiff.DiffMs != 100 {
		t.Errorf("Expected diff 100ms, got %dms", diff.ResponseTimeDiff.DiffMs)
	}

	if diff.ResponseTimeDiff.DiffPercent != 100.0 {
		t.Errorf("Expected diff percent 100%%, got %.1f%%", diff.ResponseTimeDiff.DiffPercent)
	}
}

func TestHasDifferences(t *testing.T) {
	tests := []struct {
		name     string
		old      Response
		new      Response
		expected bool
	}{
		{
			name:     "identical responses",
			old:      Response{StatusCode: 200, Body: "same"},
			new:      Response{StatusCode: 200, Body: "same"},
			expected: false,
		},
		{
			name:     "different status codes",
			old:      Response{StatusCode: 200},
			new:      Response{StatusCode: 404},
			expected: true,
		},
		{
			name: "different headers",
			old: Response{
				Headers: map[string][]string{"X-Test": {"old"}},
			},
			new: Response{
				Headers: map[string][]string{"X-Test": {"new"}},
			},
			expected: true,
		},
		{
			name:     "different bodies",
			old:      Response{Body: "old"},
			new:      Response{Body: "new"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := CompareResponses(tt.old, tt.new)
			hasDiff := diff.HasDifferences()

			if hasDiff != tt.expected {
				t.Errorf("Expected HasDifferences() to be %v, got %v", tt.expected, hasDiff)
			}
		})
	}
}

func TestFormatDiff(t *testing.T) {
	old := Response{
		StatusCode: 200,
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body:         `{"status": "ok"}`,
		ResponseTime: 100 * time.Millisecond,
	}

	new := Response{
		StatusCode: 201,
		Headers: map[string][]string{
			"Content-Type": {"application/xml"},
		},
		Body:         `{"status": "created"}`,
		ResponseTime: 150 * time.Millisecond,
	}

	diff := CompareResponses(old, new)
	formatted := FormatDiff(diff)

	// Check that formatted output contains expected sections
	expectedStrings := []string{
		"Response Comparison",
		"Status Code:",
		"Header Changes:",
		"Response Time:",
		"Body Changes",
	}

	for _, expected := range expectedStrings {
		if !containsStr(formatted, expected) {
			t.Errorf("Expected formatted output to contain '%s'", expected)
		}
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
