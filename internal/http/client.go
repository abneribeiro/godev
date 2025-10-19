package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Request struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    string
}

type Response struct {
	StatusCode   int
	Status       string
	Body         string
	Headers      map[string][]string
	ResponseTime time.Duration
	Size         int64
	Error        error
}

type Client struct {
	httpClient *http.Client
}

func NewClient(timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) Send(req Request) Response {
	startTime := time.Now()

	httpReq, err := http.NewRequest(req.Method, req.URL, bytes.NewBufferString(req.Body))
	if err != nil {
		return Response{
			Error:        fmt.Errorf("failed to create request: %w", err),
			ResponseTime: time.Since(startTime),
		}
	}

	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return Response{
			Error:        fmt.Errorf("request failed: %w", err),
			ResponseTime: time.Since(startTime),
		}
	}
	defer httpResp.Body.Close()

	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return Response{
			Error:        fmt.Errorf("failed to read response body: %w", err),
			ResponseTime: time.Since(startTime),
		}
	}

	responseTime := time.Since(startTime)
	bodyString := string(bodyBytes)

	formattedBody, err := formatJSON(bodyString)
	if err == nil {
		bodyString = formattedBody
	}

	return Response{
		StatusCode:   httpResp.StatusCode,
		Status:       httpResp.Status,
		Body:         bodyString,
		Headers:      httpResp.Header,
		ResponseTime: responseTime,
		Size:         int64(len(bodyBytes)),
		Error:        nil,
	}
}

func formatJSON(data string) (string, error) {
	var jsonData interface{}
	if err := json.Unmarshal([]byte(data), &jsonData); err != nil {
		return "", err
	}

	formatted, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return "", err
	}

	return string(formatted), nil
}

func FormatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func FormatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%dns", d.Nanoseconds())
	} else if d < time.Millisecond {
		return fmt.Sprintf("%.2fµs", float64(d.Microseconds()))
	} else if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

func RequestToCurl(req Request) string {
	var parts []string

	parts = append(parts, "curl")

	parts = append(parts, fmt.Sprintf("'%s'", req.URL))

	if req.Method != "GET" {
		parts = append(parts, "-X", req.Method)
	}

	for key, value := range req.Headers {
		parts = append(parts, "-H", fmt.Sprintf("'%s: %s'", key, value))
	}

	if req.Body != "" {
		escapedBody := req.Body
		parts = append(parts, "-d", fmt.Sprintf("'%s'", escapedBody))
	}

	return joinCurlParts(parts)
}

func joinCurlParts(parts []string) string {
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += " \\\n  " + parts[i]
	}
	return result
}
