package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

const (
	configDir  = ".devscope"
	configFile = "config.json"
	version    = "0.1.0"
)

type SavedRequest struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Method    string            `json:"method"`
	URL       string            `json:"url"`
	Headers   map[string]string `json:"headers"`
	Body      string            `json:"body"`
	CreatedAt time.Time         `json:"created_at"`
	LastUsed  time.Time         `json:"last_used"`
}

type Config struct {
	Version  string          `json:"version"`
	Requests []SavedRequest  `json:"requests"`
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
	if err := os.MkdirAll(configDirPath, 0755); err != nil {
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
		}
		if err := storage.save(); err != nil {
			return nil, fmt.Errorf("failed to initialize config: %w", err)
		}
	}

	return storage, nil
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

	if err := os.WriteFile(s.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (s *Storage) SaveRequest(name, method, url string, headers map[string]string, body string) error {
	now := time.Now()

	request := SavedRequest{
		ID:        uuid.New().String(),
		Name:      name,
		Method:    method,
		URL:       url,
		Headers:   headers,
		Body:      body,
		CreatedAt: now,
		LastUsed:  now,
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
