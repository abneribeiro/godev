package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abneribeiro/godev/internal/config"
	"github.com/abneribeiro/godev/internal/logging"
	"github.com/abneribeiro/godev/internal/ui"
)

func main() {
	// Load configuration
	cfg, err := config.LoadFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Ensure config directory exists
	if err := cfg.EnsureConfigDir(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create config directory: %v\n", err)
		os.Exit(1)
	}

	// Setup structured logging
	logConfig := &logging.Config{
		Level:  parseLogLevel(cfg.LogLevel),
		Format: cfg.LogFormat,
		Output: os.Stderr,
	}
	logging.Setup(logConfig)

	logger := logging.GetLogger()
	logger.Info("Starting godev application",
		"version", cfg.Version,
		"config_dir", cfg.ConfigDir,
	)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Shutdown signal received, initiating graceful shutdown")
		cancel()
	}()

	// Start UI application
	m := ui.NewModel()
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Run application in a goroutine
	done := make(chan error, 1)
	go func() {
		_, err := p.Run()
		done <- err
	}()

	// Wait for completion or shutdown signal
	select {
	case err := <-done:
		if err != nil {
			logger.Error("Application error", "error", err)
			fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		logger.Info("Shutting down application...")
		// Give the program time to cleanup
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		// Quit the program
		p.Quit()

		// Wait for shutdown or timeout
		select {
		case <-done:
			logger.Info("Application shutdown complete")
		case <-shutdownCtx.Done():
			logger.Warn("Shutdown timeout exceeded, forcing exit")
		}
	}

	logger.Info("Application shutdown complete")
}

func parseLogLevel(level string) logging.Level {
	switch level {
	case "debug":
		return logging.LevelDebug
	case "info":
		return logging.LevelInfo
	case "warn":
		return logging.LevelWarn
	case "error":
		return logging.LevelError
	default:
		return logging.LevelInfo
	}
}
