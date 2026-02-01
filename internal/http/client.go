package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/abneribeiro/godev/internal/errors"
)

const (
	// MaxResponseSize limits response body to 100MB to prevent DoS
	MaxResponseSize = 100 * 1024 * 1024 // 100MB
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
	return c.SendWithContext(context.Background(), req)
}

func (c *Client) SendWithContext(ctx context.Context, req Request) Response {
	startTime := time.Now()
	logger := slog.With("method", req.Method, "url", req.URL)

	// Validate URL before sending
	if _, err := url.ParseRequestURI(req.URL); err != nil {
		logger.Error("Invalid URL", "error", err)
		return Response{
			Error:        errors.NewHTTPError("invalid URL", err),
			ResponseTime: time.Since(startTime),
		}
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, bytes.NewBufferString(req.Body))
	if err != nil {
		logger.Error("Failed to create request", "error", err)
		return Response{
			Error:        errors.NewHTTPError("failed to create request", err),
			ResponseTime: time.Since(startTime),
		}
	}

	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	logger.Debug("Sending HTTP request")
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		logger.Error("Request failed", "error", err)
		return Response{
			Error:        errors.NewHTTPError("request failed", err),
			ResponseTime: time.Since(startTime),
		}
	}
	defer httpResp.Body.Close()

	// Limit response size to prevent DoS attacks
	// Read up to MaxResponseSize + 1 to detect if response exceeds limit
	limitedReader := io.LimitReader(httpResp.Body, MaxResponseSize+1)
	bodyBytes, err := io.ReadAll(limitedReader)
	if err != nil {
		logger.Error("Failed to read response body", "error", err)
		return Response{
			Error:        errors.NewHTTPError("failed to read response body", err),
			ResponseTime: time.Since(startTime),
		}
	}

	// Check if response was truncated (read more than MaxResponseSize)
	if int64(len(bodyBytes)) > MaxResponseSize {
		err := fmt.Errorf("response too large (exceeds %d bytes)", MaxResponseSize)
		logger.Warn("Response too large", "max_size", MaxResponseSize, "actual_size", len(bodyBytes))
		return Response{
			Error:        errors.NewHTTPError("response too large", err),
			ResponseTime: time.Since(startTime),
		}
	}

	responseTime := time.Since(startTime)
	bodyString := string(bodyBytes)

	formattedBody, err := formatJSON(bodyString)
	if err == nil {
		bodyString = formattedBody
	}

	logger.Info("Request completed successfully",
		"status_code", httpResp.StatusCode,
		"response_time", responseTime,
		"response_size", len(bodyBytes),
	)

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

// formatJSON formats JSON using json.Indent for better performance
// This avoids the unnecessary unmarshal/marshal cycle
func formatJSON(data string) (string, error) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(data), "", "  "); err != nil {
		return "", err
	}
	return buf.String(), nil
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
