package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	oldConfigDir = ".devscope"
	configDir    = ".godev"
	configFile   = "config.json"
	version      = "0.4.0"
)

type RequestExecution struct {
	ID           string              `json:"id"`
	Timestamp    time.Time           `json:"timestamp"`
	Method       string              `json:"method"`
	URL          string              `json:"url"`
	Headers      map[string]string   `json:"headers"`
	Body         string              `json:"body"`
	QueryParams  map[string]string   `json:"query_params"`
	StatusCode   int                 `json:"status_code"`
	Status       string              `json:"status"`
	ResponseBody string              `json:"response_body"`
	ResponseTime int64               `json:"response_time_ms"`
	Error        string              `json:"error,omitempty"`
}

type SavedRequest struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	Headers     map[string]string `json:"headers"`
	Body        string            `json:"body"`
	QueryParams map[string]string `json:"query_params"`
	CreatedAt   time.Time         `json:"created_at"`
	LastUsed    time.Time         `json:"last_used"`
}

type Config struct {
	Version  string             `json:"version"`
	Requests []SavedRequest     `json:"requests"`
	History  []RequestExecution `json:"history"`
}

type Storage struct {
	configPath string
	config     *Config
}

func NewStorage() (*Storage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDirPath := filepath.Join(homeDir, configDir)
	oldConfigDirPath := filepath.Join(homeDir, oldConfigDir)

	if err := migrateOldConfig(oldConfigDirPath, configDirPath); err != nil {
		fmt.Printf("Warning: Migration from .devscope failed: %v\n", err)
	}

	// Use secure directory permissions (0700 - only owner can access)
	if err := os.MkdirAll(configDirPath, 0700); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDirPath, configFile)

	storage := &Storage{
		configPath: configPath,
	}

	if err := storage.load(); err != nil {
		storage.config = &Config{
			Version:  version,
			Requests: []SavedRequest{},
			History:  []RequestExecution{},
		}
		if err := storage.save(); err != nil {
			return nil, fmt.Errorf("failed to initialize config: %w", err)
		}
	}

	if storage.config.History == nil {
		storage.config.History = []RequestExecution{}
	}

	return storage, nil
}

func migrateOldConfig(oldDir, newDir string) error {
	oldConfigPath := filepath.Join(oldDir, configFile)
	newConfigPath := filepath.Join(newDir, configFile)

	if _, err := os.Stat(newConfigPath); err == nil {
		return nil
	}

	if _, err := os.Stat(oldConfigPath); os.IsNotExist(err) {
		return nil
	}

	// Use secure directory permissions
	if err := os.MkdirAll(newDir, 0700); err != nil {
		return err
	}

	data, err := os.ReadFile(oldConfigPath)
	if err != nil {
		return err
	}

	// Use secure file permissions during migration
	if err := os.WriteFile(newConfigPath, data, 0600); err != nil {
		return err
	}

	fmt.Println("✓ Successfully migrated config from ~/.devscope to ~/.godev")
	return nil
}

func (s *Storage) load() error {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return err
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	s.config = &config
	return nil
}

func (s *Storage) save() error {
	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Use secure file permissions (0600 - only owner can read/write)
	// This is critical as the file may contain API tokens and sensitive data
	if err := os.WriteFile(s.configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (s *Storage) SaveRequest(name, method, url string, headers map[string]string, body string, queryParams map[string]string) error {
	now := time.Now()

	request := SavedRequest{
		ID:          uuid.New().String(),
		Name:        name,
		Method:      method,
		URL:         url,
		Headers:     headers,
		Body:        body,
		QueryParams: queryParams,
		CreatedAt:   now,
		LastUsed:    now,
	}

	s.config.Requests = append(s.config.Requests, request)
	return s.save()
}

func (s *Storage) GetRequests() []SavedRequest {
	return s.config.Requests
}

func (s *Storage) GetRequest(id string) (*SavedRequest, error) {
	for i := range s.config.Requests {
		if s.config.Requests[i].ID == id {
			return &s.config.Requests[i], nil
		}
	}
	return nil, fmt.Errorf("request not found: %s", id)
}

func (s *Storage) UpdateLastUsed(id string) error {
	for i := range s.config.Requests {
		if s.config.Requests[i].ID == id {
			s.config.Requests[i].LastUsed = time.Now()
			return s.save()
		}
	}
	return fmt.Errorf("request not found: %s", id)
}

func (s *Storage) DeleteRequest(id string) error {
	for i := range s.config.Requests {
		if s.config.Requests[i].ID == id {
			s.config.Requests = append(s.config.Requests[:i], s.config.Requests[i+1:]...)
			return s.save()
		}
	}
	return fmt.Errorf("request not found: %s", id)
}

func (s *Storage) RequestExists(name string) bool {
	for _, req := range s.config.Requests {
		if req.Name == name {
			return true
		}
	}
	return false
}

const maxHistorySize = 100

func (s *Storage) AddToHistory(method, url string, headers map[string]string, body string, queryParams map[string]string, statusCode int, status, responseBody string, responseTimeMs int64, err error) error {
	execution := RequestExecution{
		ID:           uuid.New().String(),
		Timestamp:    time.Now(),
		Method:       method,
		URL:          url,
		Headers:      headers,
		Body:         body,
		QueryParams:  queryParams,
		StatusCode:   statusCode,
		Status:       status,
		ResponseBody: responseBody,
		ResponseTime: responseTimeMs,
	}

	if err != nil {
		execution.Error = err.Error()
	}

	s.config.History = append([]RequestExecution{execution}, s.config.History...)

	if len(s.config.History) > maxHistorySize {
		s.config.History = s.config.History[:maxHistorySize]
	}

	return s.save()
}

func (s *Storage) GetHistory() []RequestExecution {
	return s.config.History
}

func (s *Storage) ClearHistory() error {
	s.config.History = []RequestExecution{}
	return s.save()
}

func (s *Storage) DeleteHistoryItem(id string) error {
	for i := range s.config.History {
		if s.config.History[i].ID == id {
			s.config.History = append(s.config.History[:i], s.config.History[i+1:]...)
			return s.save()
		}
	}
	return fmt.Errorf("history item not found: %s", id)
}

func (s *Storage) FilterRequests(query string) []SavedRequest {
	if query == "" {
		return s.config.Requests
	}

	query = strings.ToLower(query)
	filtered := []SavedRequest{}

	for _, req := range s.config.Requests {
		if strings.Contains(strings.ToLower(req.Name), query) ||
			strings.Contains(strings.ToLower(req.Method), query) ||
			strings.Contains(strings.ToLower(req.URL), query) {
			filtered = append(filtered, req)
		}
	}

	return filtered
}
