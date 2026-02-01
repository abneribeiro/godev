package logging

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Level != LevelInfo {
		t.Errorf("Expected default level to be Info, got %v", config.Level)
	}

	if config.Format != "text" {
		t.Errorf("Expected default format to be text, got %s", config.Format)
	}

	if config.Output != os.Stderr {
		t.Errorf("Expected default output to be stderr")
	}
}

func TestSetup(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name:   "default config",
			config: nil,
		},
		{
			name: "json format",
			config: &Config{
				Level:  LevelDebug,
				Format: "json",
				Output: os.Stderr,
			},
		},
		{
			name: "text format",
			config: &Config{
				Level:  LevelWarn,
				Format: "text",
				Output: os.Stderr,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Setup(tt.config)

			// Test that logger is properly configured
			logger := GetLogger()
			if logger == nil {
				t.Error("Expected logger to be configured")
			}
		})
	}
}

func TestWithContext(t *testing.T) {
	// Redirect output to capture logs
	var buf bytes.Buffer
	config := &Config{
		Level:  LevelInfo,
		Format: "text",
		Output: &buf,
	}
	Setup(config)

	logger := WithContext("component", "test", "operation", "validation")
	logger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "component=test") {
		t.Error("Expected output to contain context field component=test")
	}
	if !strings.Contains(output, "operation=validation") {
		t.Error("Expected output to contain context field operation=validation")
	}
}

func BenchmarkLoggingInfo(b *testing.B) {
	var buf bytes.Buffer
	config := &Config{
		Level:  LevelInfo,
		Format: "text",
		Output: &buf,
	}
	Setup(config)

	logger := GetLogger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message", "iteration", i)
	}
}

func BenchmarkLoggingJSON(b *testing.B) {
	var buf bytes.Buffer
	config := &Config{
		Level:  LevelInfo,
		Format: "json",
		Output: &buf,
	}
	Setup(config)

	logger := GetLogger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message", "iteration", i)
	}
}
