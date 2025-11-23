package http

import (
	"encoding/json"
	"fmt"
	"strings"
)

// DiffResult represents the comparison between two responses
type DiffResult struct {
	StatusCodeDiff   *ValueDiff
	HeadersDiff      map[string]*ValueDiff
	BodyDiff         *BodyDiff
	ResponseTimeDiff *TimeDiff
}

// ValueDiff represents a difference in a single value
type ValueDiff struct {
	Old    string
	New    string
	Changed bool
}

// BodyDiff represents differences in response bodies
type BodyDiff struct {
	Type    string // "json", "text", "binary"
	Changes []Change
	Summary string
}

// Change represents a single change in the body
type Change struct {
	Type     string // "added", "removed", "modified"
	Path     string // JSON path or line number for text
	OldValue string
	NewValue string
}

// TimeDiff represents difference in response times
type TimeDiff struct {
	OldMs      int64
	NewMs      int64
	DiffMs     int64
	DiffPercent float64
}

// CompareResponses compares two HTTP responses and returns differences
func CompareResponses(old, new Response) *DiffResult {
	result := &DiffResult{
		HeadersDiff: make(map[string]*ValueDiff),
	}

	// Compare status codes
	if old.StatusCode != new.StatusCode {
		result.StatusCodeDiff = &ValueDiff{
			Old:     fmt.Sprintf("%d", old.StatusCode),
			New:     fmt.Sprintf("%d", new.StatusCode),
			Changed: true,
		}
	}

	// Compare headers
	allHeaders := make(map[string]bool)
	for k := range old.Headers {
		allHeaders[k] = true
	}
	for k := range new.Headers {
		allHeaders[k] = true
	}

	for header := range allHeaders {
		oldVal := old.Headers[header]
		newVal := new.Headers[header]

		// Compare header values (they're slices, so we need to compare properly)
		oldStr := strings.Join(oldVal, ", ")
		newStr := strings.Join(newVal, ", ")

		if oldStr != newStr {
			result.HeadersDiff[header] = &ValueDiff{
				Old:     oldStr,
				New:     newStr,
				Changed: true,
			}
		}
	}

	// Compare bodies
	result.BodyDiff = compareBodies(old.Body, new.Body)

	// Compare response times
	if old.ResponseTime.Milliseconds() != new.ResponseTime.Milliseconds() {
		oldMs := old.ResponseTime.Milliseconds()
		newMs := new.ResponseTime.Milliseconds()
		diffMs := newMs - oldMs
		diffPercent := float64(0)
		if oldMs > 0 {
			diffPercent = (float64(diffMs) / float64(oldMs)) * 100
		}

		result.ResponseTimeDiff = &TimeDiff{
			OldMs:       oldMs,
			NewMs:       newMs,
			DiffMs:      diffMs,
			DiffPercent: diffPercent,
		}
	}

	return result
}

// compareBodies compares two response bodies
func compareBodies(old, new string) *BodyDiff {
	// Try to parse as JSON first
	var oldJSON, newJSON interface{}
	oldIsJSON := json.Unmarshal([]byte(old), &oldJSON) == nil
	newIsJSON := json.Unmarshal([]byte(new), &newJSON) == nil

	if oldIsJSON && newIsJSON {
		return compareJSON(oldJSON, newJSON, "")
	}

	// Fall back to text comparison
	return compareText(old, new)
}

// compareJSON compares two JSON structures
func compareJSON(old, new interface{}, path string) *BodyDiff {
	diff := &BodyDiff{
		Type:    "json",
		Changes: []Change{},
	}

	changes := findJSONDifferences(old, new, path)
	diff.Changes = changes

	// Generate summary
	added := 0
	removed := 0
	modified := 0
	for _, change := range changes {
		switch change.Type {
		case "added":
			added++
		case "removed":
			removed++
		case "modified":
			modified++
		}
	}

	diff.Summary = fmt.Sprintf("%d modified, %d added, %d removed", modified, added, removed)
	return diff
}

// findJSONDifferences recursively finds differences in JSON structures
func findJSONDifferences(old, new interface{}, path string) []Change {
	var changes []Change

	// Both nil
	if old == nil && new == nil {
		return changes
	}

	// One is nil
	if old == nil {
		changes = append(changes, Change{
			Type:     "added",
			Path:     path,
			OldValue: "null",
			NewValue: formatJSONValue(new),
		})
		return changes
	}
	if new == nil {
		changes = append(changes, Change{
			Type:     "removed",
			Path:     path,
			OldValue: formatJSONValue(old),
			NewValue: "null",
		})
		return changes
	}

	// Both are maps
	if oldMap, oldOk := old.(map[string]interface{}); oldOk {
		if newMap, newOk := new.(map[string]interface{}); newOk {
			// Find all keys
			allKeys := make(map[string]bool)
			for k := range oldMap {
				allKeys[k] = true
			}
			for k := range newMap {
				allKeys[k] = true
			}

			for key := range allKeys {
				newPath := path
				if newPath == "" {
					newPath = key
				} else {
					newPath = path + "." + key
				}

				oldVal, oldExists := oldMap[key]
				newVal, newExists := newMap[key]

				if !oldExists {
					changes = append(changes, Change{
						Type:     "added",
						Path:     newPath,
						OldValue: "",
						NewValue: formatJSONValue(newVal),
					})
				} else if !newExists {
					changes = append(changes, Change{
						Type:     "removed",
						Path:     newPath,
						OldValue: formatJSONValue(oldVal),
						NewValue: "",
					})
				} else {
					// Recursively compare
					subChanges := findJSONDifferences(oldVal, newVal, newPath)
					changes = append(changes, subChanges...)
				}
			}
			return changes
		}
	}

	// Both are arrays
	if oldArr, oldOk := old.([]interface{}); oldOk {
		if newArr, newOk := new.([]interface{}); newOk {
			maxLen := len(oldArr)
			if len(newArr) > maxLen {
				maxLen = len(newArr)
			}

			for i := 0; i < maxLen; i++ {
				newPath := fmt.Sprintf("%s[%d]", path, i)

				if i >= len(oldArr) {
					changes = append(changes, Change{
						Type:     "added",
						Path:     newPath,
						OldValue: "",
						NewValue: formatJSONValue(newArr[i]),
					})
				} else if i >= len(newArr) {
					changes = append(changes, Change{
						Type:     "removed",
						Path:     newPath,
						OldValue: formatJSONValue(oldArr[i]),
						NewValue: "",
					})
				} else {
					// Recursively compare
					subChanges := findJSONDifferences(oldArr[i], newArr[i], newPath)
					changes = append(changes, subChanges...)
				}
			}
			return changes
		}
	}

	// Primitive values - compare directly
	oldStr := formatJSONValue(old)
	newStr := formatJSONValue(new)
	if oldStr != newStr {
		changes = append(changes, Change{
			Type:     "modified",
			Path:     path,
			OldValue: oldStr,
			NewValue: newStr,
		})
	}

	return changes
}

// formatJSONValue formats a JSON value as a string
func formatJSONValue(v interface{}) string {
	if v == nil {
		return "null"
	}

	switch val := v.(type) {
	case string:
		return fmt.Sprintf("%q", val)
	case float64:
		return fmt.Sprintf("%v", val)
	case bool:
		return fmt.Sprintf("%v", val)
	default:
		bytes, _ := json.Marshal(v)
		return string(bytes)
	}
}

// compareText compares two text strings line by line
func compareText(old, new string) *BodyDiff {
	diff := &BodyDiff{
		Type:    "text",
		Changes: []Change{},
	}

	oldLines := strings.Split(old, "\n")
	newLines := strings.Split(new, "\n")

	// Simple line-by-line comparison
	maxLen := len(oldLines)
	if len(newLines) > maxLen {
		maxLen = len(newLines)
	}

	for i := 0; i < maxLen; i++ {
		path := fmt.Sprintf("line %d", i+1)

		if i >= len(oldLines) {
			diff.Changes = append(diff.Changes, Change{
				Type:     "added",
				Path:     path,
				OldValue: "",
				NewValue: newLines[i],
			})
		} else if i >= len(newLines) {
			diff.Changes = append(diff.Changes, Change{
				Type:     "removed",
				Path:     path,
				OldValue: oldLines[i],
				NewValue: "",
			})
		} else if oldLines[i] != newLines[i] {
			diff.Changes = append(diff.Changes, Change{
				Type:     "modified",
				Path:     path,
				OldValue: oldLines[i],
				NewValue: newLines[i],
			})
		}
	}

	// Generate summary
	diff.Summary = fmt.Sprintf("%d lines changed", len(diff.Changes))
	return diff
}

// FormatDiff returns a human-readable diff report
func FormatDiff(diff *DiffResult) string {
	var output strings.Builder

	output.WriteString("Response Comparison\n")
	output.WriteString("===================\n\n")

	// Status code
	if diff.StatusCodeDiff != nil {
		output.WriteString(fmt.Sprintf("Status Code: %s -> %s\n\n",
			diff.StatusCodeDiff.Old, diff.StatusCodeDiff.New))
	}

	// Headers
	if len(diff.HeadersDiff) > 0 {
		output.WriteString("Header Changes:\n")
		for header, valueDiff := range diff.HeadersDiff {
			if valueDiff.Old == "" {
				output.WriteString(fmt.Sprintf("  + %s: %s\n", header, valueDiff.New))
			} else if valueDiff.New == "" {
				output.WriteString(fmt.Sprintf("  - %s: %s\n", header, valueDiff.Old))
			} else {
				output.WriteString(fmt.Sprintf("  ~ %s: %s -> %s\n", header, valueDiff.Old, valueDiff.New))
			}
		}
		output.WriteString("\n")
	}

	// Response time
	if diff.ResponseTimeDiff != nil {
		output.WriteString(fmt.Sprintf("Response Time: %dms -> %dms (%+dms, %+.1f%%)\n\n",
			diff.ResponseTimeDiff.OldMs,
			diff.ResponseTimeDiff.NewMs,
			diff.ResponseTimeDiff.DiffMs,
			diff.ResponseTimeDiff.DiffPercent))
	}

	// Body changes
	if diff.BodyDiff != nil && len(diff.BodyDiff.Changes) > 0 {
		output.WriteString(fmt.Sprintf("Body Changes (%s):\n", diff.BodyDiff.Summary))
		output.WriteString("\n")

		for _, change := range diff.BodyDiff.Changes {
			switch change.Type {
			case "added":
				output.WriteString(fmt.Sprintf("  + %s: %s\n", change.Path, change.NewValue))
			case "removed":
				output.WriteString(fmt.Sprintf("  - %s: %s\n", change.Path, change.OldValue))
			case "modified":
				output.WriteString(fmt.Sprintf("  ~ %s: %s -> %s\n", change.Path, change.OldValue, change.NewValue))
			}
		}
	}

	return output.String()
}

// HasDifferences checks if there are any differences between responses
func (d *DiffResult) HasDifferences() bool {
	return d.StatusCodeDiff != nil ||
		len(d.HeadersDiff) > 0 ||
		(d.BodyDiff != nil && len(d.BodyDiff.Changes) > 0) ||
		d.ResponseTimeDiff != nil
}
