// Package config provides configuration management for the godev application
package config

import (
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/abneribeiro/godev/internal/errors"
)

// Config holds the application configuration
type Config struct {
	// Application settings
	Version string
	AppName string

	// Storage settings
	ConfigDir string

	// HTTP settings
	HTTPTimeout time.Duration
	MaxRetries  int

	// Database settings
	DBConnectTimeout time.Duration
	DBMaxConnections int
	DBMaxIdle        int
	DBConnLifetime   time.Duration

	// Logging settings
	LogLevel  string
	LogFormat string

	// Performance settings
	MaxResponseSize int64
	MaxRowsInMemory int

	// UI settings
	EnableColors bool
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Version: "0.4.0",
		AppName: "godev",

		// Use XDG Base Directory spec
		ConfigDir: getConfigDir(),

		// HTTP defaults
		HTTPTimeout: 30 * time.Second,
		MaxRetries:  3,

		// Database defaults
		DBConnectTimeout: 10 * time.Second,
		DBMaxConnections: 25,
		DBMaxIdle:        5,
		DBConnLifetime:   5 * time.Minute,

		// Logging defaults
		LogLevel:  "info",
		LogFormat: "text",

		// Performance defaults
		MaxResponseSize: 100 * 1024 * 1024, // 100MB
		MaxRowsInMemory: 10000,

		// UI defaults
		EnableColors: true,
	}
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() (*Config, error) {
	config := DefaultConfig()

	// Override with environment variables if present
	if timeout := os.Getenv("GODEV_HTTP_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			config.HTTPTimeout = d
		}
	}

	if retries := os.Getenv("GODEV_MAX_RETRIES"); retries != "" {
		if r, err := strconv.Atoi(retries); err == nil {
			config.MaxRetries = r
		}
	}

	if dbTimeout := os.Getenv("GODEV_DB_TIMEOUT"); dbTimeout != "" {
		if d, err := time.ParseDuration(dbTimeout); err == nil {
			config.DBConnectTimeout = d
		}
	}

	if maxConns := os.Getenv("GODEV_DB_MAX_CONNECTIONS"); maxConns != "" {
		if m, err := strconv.Atoi(maxConns); err == nil {
			config.DBMaxConnections = m
		}
	}

	if logLevel := os.Getenv("GODEV_LOG_LEVEL"); logLevel != "" {
		config.LogLevel = logLevel
	}

	if logFormat := os.Getenv("GODEV_LOG_FORMAT"); logFormat != "" {
		config.LogFormat = logFormat
	}

	if maxSize := os.Getenv("GODEV_MAX_RESPONSE_SIZE"); maxSize != "" {
		if m, err := strconv.ParseInt(maxSize, 10, 64); err == nil {
			config.MaxResponseSize = m
		}
	}

	if colors := os.Getenv("GODEV_ENABLE_COLORS"); colors != "" {
		config.EnableColors = colors != "false" && colors != "0"
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.HTTPTimeout <= 0 {
		return errors.NewConfigError("HTTP timeout must be positive", nil)
	}

	if c.MaxRetries < 0 {
		return errors.NewConfigError("max retries cannot be negative", nil)
	}

	if c.DBConnectTimeout <= 0 {
		return errors.NewConfigError("database connect timeout must be positive", nil)
	}

	if c.DBMaxConnections <= 0 {
		return errors.NewConfigError("database max connections must be positive", nil)
	}

	if c.MaxResponseSize <= 0 {
		return errors.NewConfigError("max response size must be positive", nil)
	}

	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if !validLogLevels[c.LogLevel] {
		return errors.NewConfigError("invalid log level", nil)
	}

	validLogFormats := map[string]bool{
		"text": true,
		"json": true,
	}

	if !validLogFormats[c.LogFormat] {
		return errors.NewConfigError("invalid log format", nil)
	}

	return nil
}

// getConfigDir returns the configuration directory following XDG Base Directory spec
func getConfigDir() string {
	// Try XDG_CONFIG_HOME first
	if configHome := os.Getenv("XDG_CONFIG_HOME"); configHome != "" {
		return filepath.Join(configHome, "godev")
	}

	// Fall back to ~/.config/godev
	if homeDir, err := os.UserHomeDir(); err == nil {
		return filepath.Join(homeDir, ".config", "godev")
	}

	// Ultimate fallback to current directory
	return ".godev"
}

// EnsureConfigDir ensures the configuration directory exists
func (c *Config) EnsureConfigDir() error {
	if err := os.MkdirAll(c.ConfigDir, 0o700); err != nil {
		return errors.NewConfigError("failed to create config directory", err)
	}
	return nil
}
