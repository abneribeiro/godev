// Package storage provides a unified storage abstraction for the godev application
package storage

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/abneribeiro/godev/internal/errors"
)

const (
	configDirName = ".godev"
	fileMode      = 0o600
	dirMode       = 0o700
	appVersion    = "0.4.0"
)

// FileStorage provides a unified interface for file-based storage operations
type FileStorage struct {
	baseDir string
}

// NewFileStorage creates a new file storage instance
func NewFileStorage() (*FileStorage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, errors.NewStorageError("failed to get home directory", err)
	}

	baseDir := filepath.Join(homeDir, configDirName)
	fs := &FileStorage{baseDir: baseDir}

	// Ensure base directory exists
	if err := fs.ensureDir(baseDir); err != nil {
		return nil, err
	}

	return fs, nil
}

// ensureDir ensures directory exists with proper permissions
func (fs *FileStorage) ensureDir(dir string) error {
	if err := os.MkdirAll(dir, dirMode); err != nil {
		return errors.NewStorageError("failed to create directory", err)
	}
	return nil
}

// SaveJSON saves data as JSON to a file with atomic write
func (fs *FileStorage) SaveJSON(ctx context.Context, filename string, data interface{}) error {
	logger := slog.With("filename", filename)

	filePath := filepath.Join(fs.baseDir, filename)
	tempPath := filePath + ".tmp"

	// Marshal to JSON with indentation for readability
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		logger.Error("Failed to marshal JSON", "error", err)
		return errors.NewStorageError("failed to marshal JSON", err)
	}

	// Write to temporary file first (atomic operation)
	if err := os.WriteFile(tempPath, jsonData, fileMode); err != nil {
		logger.Error("Failed to write temporary file", "error", err)
		return errors.NewStorageError("failed to write temporary file", err)
	}

	// Rename to final destination (atomic operation on most filesystems)
	if err := os.Rename(tempPath, filePath); err != nil {
		os.Remove(tempPath) // Clean up temp file
		logger.Error("Failed to rename temporary file", "error", err)
		return errors.NewStorageError("failed to rename temporary file", err)
	}

	logger.Debug("Successfully saved JSON file")
	return nil
}

// LoadJSON loads JSON data from a file
func (fs *FileStorage) LoadJSON(ctx context.Context, filename string, target interface{}) error {
	logger := slog.With("filename", filename)

	filePath := filepath.Join(fs.baseDir, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return errors.NewStorageError("file not found", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		logger.Error("Failed to read file", "error", err)
		return errors.NewStorageError("failed to read file", err)
	}

	if err := json.Unmarshal(data, target); err != nil {
		logger.Error("Failed to unmarshal JSON", "error", err)
		return errors.NewStorageError("failed to unmarshal JSON", err)
	}

	logger.Debug("Successfully loaded JSON file")
	return nil
}

// Exists checks if a file exists
func (fs *FileStorage) Exists(filename string) bool {
	filePath := filepath.Join(fs.baseDir, filename)
	_, err := os.Stat(filePath)
	return err == nil
}

// Delete removes a file
func (fs *FileStorage) Delete(ctx context.Context, filename string) error {
	logger := slog.With("filename", filename)

	filePath := filepath.Join(fs.baseDir, filename)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		logger.Error("Failed to delete file", "error", err)
		return errors.NewStorageError("failed to delete file", err)
	}

	logger.Debug("Successfully deleted file")
	return nil
}

// GetBaseDir returns the base storage directory
func (fs *FileStorage) GetBaseDir() string {
	return fs.baseDir
}

// MigrateFromOld migrates data from old directory structure
func (fs *FileStorage) MigrateFromOld(oldDirName string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return errors.NewStorageError("failed to get home directory", err)
	}

	oldDir := filepath.Join(homeDir, oldDirName)
	if _, err := os.Stat(oldDir); os.IsNotExist(err) {
		return nil // Nothing to migrate
	}

	logger := slog.With("old_dir", oldDir, "new_dir", fs.baseDir)
	logger.Info("Migrating configuration from old directory")

	// List files in old directory
	entries, err := os.ReadDir(oldDir)
	if err != nil {
		return errors.NewStorageError("failed to read old directory", err)
	}

	// Copy files to new location
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		oldPath := filepath.Join(oldDir, entry.Name())
		newPath := filepath.Join(fs.baseDir, entry.Name())

		data, err := os.ReadFile(oldPath)
		if err != nil {
			logger.Warn("Failed to read old file", "file", entry.Name(), "error", err)
			continue
		}

		if err := os.WriteFile(newPath, data, fileMode); err != nil {
			logger.Warn("Failed to write new file", "file", entry.Name(), "error", err)
			continue
		}

		logger.Debug("Migrated file", "file", entry.Name())
	}

	logger.Info("Migration completed successfully")
	return nil
}
