package database

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type SavedQuery struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Query     string    `json:"query"`
	CreatedAt time.Time `json:"created_at"`
	LastUsed  time.Time `json:"last_used"`
}

type QueryExecution struct {
	ID             string        `json:"id"`
	Timestamp      time.Time     `json:"timestamp"`
	Query          string        `json:"query"`
	RowsAffected   int64         `json:"rows_affected"`
	ExecutionTime  int64         `json:"execution_time_ms"`
	Error          string        `json:"error,omitempty"`
	ConnectionInfo string        `json:"connection_info"`
}

type DatabaseConfig struct {
	Version          string             `json:"version"`
	SavedQueries     []SavedQuery       `json:"saved_queries"`
	QueryHistory     []QueryExecution   `json:"query_history"`
	SavedConnections []ConnectionConfig `json:"saved_connections"`
}

type DatabaseStorage struct {
	configPath string
	config     *DatabaseConfig
}

const (
	databaseConfigFile = "database.json"
	dbConfigVersion    = "0.4.0"
	maxQueryHistory    = 100
)

func NewDatabaseStorage() (*DatabaseStorage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDirPath := filepath.Join(homeDir, ".godev")
	// Use secure directory permissions (0700 - only owner can access)
	if err := os.MkdirAll(configDirPath, 0700); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDirPath, databaseConfigFile)

	storage := &DatabaseStorage{
		configPath: configPath,
	}

	if err := storage.load(); err != nil {
		storage.config = &DatabaseConfig{
			Version:          dbConfigVersion,
			SavedQueries:     []SavedQuery{},
			QueryHistory:     []QueryExecution{},
			SavedConnections: []ConnectionConfig{},
		}
		if err := storage.save(); err != nil {
			return nil, fmt.Errorf("failed to initialize database config: %w", err)
		}
	}

	if storage.config.SavedQueries == nil {
		storage.config.SavedQueries = []SavedQuery{}
	}
	if storage.config.QueryHistory == nil {
		storage.config.QueryHistory = []QueryExecution{}
	}
	if storage.config.SavedConnections == nil {
		storage.config.SavedConnections = []ConnectionConfig{}
	}

	return storage, nil
}

func (s *DatabaseStorage) load() error {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return err
		}
		return fmt.Errorf("failed to read database config file: %w", err)
	}

	var config DatabaseConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse database config file: %w", err)
	}

	s.config = &config
	return nil
}

func (s *DatabaseStorage) save() error {
	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal database config: %w", err)
	}

	// Use secure file permissions (0600 - only owner can read/write)
	// This is critical as the file may contain database passwords
	if err := os.WriteFile(s.configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write database config file: %w", err)
	}

	return nil
}

func (s *DatabaseStorage) SaveQuery(name, query string) error {
	now := time.Now()

	savedQuery := SavedQuery{
		ID:        uuid.New().String(),
		Name:      name,
		Query:     query,
		CreatedAt: now,
		LastUsed:  now,
	}

	s.config.SavedQueries = append(s.config.SavedQueries, savedQuery)
	return s.save()
}

func (s *DatabaseStorage) GetQueries() []SavedQuery {
	return s.config.SavedQueries
}

func (s *DatabaseStorage) DeleteQuery(id string) error {
	for i := range s.config.SavedQueries {
		if s.config.SavedQueries[i].ID == id {
			s.config.SavedQueries = append(s.config.SavedQueries[:i], s.config.SavedQueries[i+1:]...)
			return s.save()
		}
	}
	return fmt.Errorf("query not found: %s", id)
}

func (s *DatabaseStorage) QueryExists(name string) bool {
	for _, query := range s.config.SavedQueries {
		if query.Name == name {
			return true
		}
	}
	return false
}

func (s *DatabaseStorage) FilterQueries(searchQuery string) []SavedQuery {
	if searchQuery == "" {
		return s.config.SavedQueries
	}

	searchQuery = strings.ToLower(searchQuery)
	filtered := []SavedQuery{}

	for _, query := range s.config.SavedQueries {
		if strings.Contains(strings.ToLower(query.Name), searchQuery) ||
			strings.Contains(strings.ToLower(query.Query), searchQuery) {
			filtered = append(filtered, query)
		}
	}

	return filtered
}

func (s *DatabaseStorage) AddToQueryHistory(query, connectionInfo string, rowsAffected int64, executionTimeMs int64, err error) error {
	execution := QueryExecution{
		ID:             uuid.New().String(),
		Timestamp:      time.Now(),
		Query:          query,
		RowsAffected:   rowsAffected,
		ExecutionTime:  executionTimeMs,
		ConnectionInfo: connectionInfo,
	}

	if err != nil {
		execution.Error = err.Error()
	}

	s.config.QueryHistory = append([]QueryExecution{execution}, s.config.QueryHistory...)

	if len(s.config.QueryHistory) > maxQueryHistory {
		s.config.QueryHistory = s.config.QueryHistory[:maxQueryHistory]
	}

	return s.save()
}

func (s *DatabaseStorage) GetQueryHistory() []QueryExecution {
	return s.config.QueryHistory
}

func (s *DatabaseStorage) ClearQueryHistory() error {
	s.config.QueryHistory = []QueryExecution{}
	return s.save()
}

func (s *DatabaseStorage) DeleteQueryHistoryItem(id string) error {
	for i := range s.config.QueryHistory {
		if s.config.QueryHistory[i].ID == id {
			s.config.QueryHistory = append(s.config.QueryHistory[:i], s.config.QueryHistory[i+1:]...)
			return s.save()
		}
	}
	return fmt.Errorf("query history item not found: %s", id)
}

func (s *DatabaseStorage) SaveConnection(config ConnectionConfig) error {
	for i, conn := range s.config.SavedConnections {
		if conn.Host == config.Host && conn.Port == config.Port && conn.Database == config.Database {
			s.config.SavedConnections[i] = config
			return s.save()
		}
	}

	s.config.SavedConnections = append(s.config.SavedConnections, config)
	return s.save()
}

func (s *DatabaseStorage) GetSavedConnections() []ConnectionConfig {
	return s.config.SavedConnections
}

func (s *DatabaseStorage) DeleteConnection(host string, port int, database string) error {
	for i := range s.config.SavedConnections {
		conn := s.config.SavedConnections[i]
		if conn.Host == host && conn.Port == port && conn.Database == database {
			s.config.SavedConnections = append(s.config.SavedConnections[:i], s.config.SavedConnections[i+1:]...)
			return s.save()
		}
	}
	return fmt.Errorf("connection not found")
}
