package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// variableRegex is compiled once for better performance
var variableRegex = regexp.MustCompile(`\{\{([^}]+)\}\}`)

type Variable struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Environment struct {
	Name      string     `json:"name"`
	Variables []Variable `json:"variables"`
}

type EnvironmentConfig struct {
	Version           string        `json:"version"`
	Environments      []Environment `json:"environments"`
	ActiveEnvironment string        `json:"active_environment"`
}

const (
	envConfigFile    = "environments.json"
	envConfigVersion = "0.4.0"
)

func (s *Storage) LoadEnvironments() (*EnvironmentConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	envPath := filepath.Join(homeDir, ".godev", envConfigFile)

	data, err := os.ReadFile(envPath)
	if err != nil {
		if os.IsNotExist(err) {
			defaultConfig := &EnvironmentConfig{
				Version:           envConfigVersion,
				Environments:      []Environment{},
				ActiveEnvironment: "",
			}
			if err := s.SaveEnvironments(defaultConfig); err != nil {
				return nil, err
			}
			return defaultConfig, nil
		}
		return nil, fmt.Errorf("failed to read environment config: %w", err)
	}

	var config EnvironmentConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse environment config: %w", err)
	}

	return &config, nil
}

func (s *Storage) SaveEnvironments(config *EnvironmentConfig) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".godev")
	// Use secure directory permissions (0700 - only owner can access)
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	envPath := filepath.Join(configDir, envConfigFile)

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal environment config: %w", err)
	}

	// Use secure file permissions (0600 - only owner can read/write)
	// This is critical as the file contains API keys and sensitive environment variables
	if err := os.WriteFile(envPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write environment config: %w", err)
	}

	return nil
}

func (s *Storage) AddEnvironment(name string) error {
	config, err := s.LoadEnvironments()
	if err != nil {
		return err
	}

	for _, env := range config.Environments {
		if env.Name == name {
			return fmt.Errorf("environment already exists: %s", name)
		}
	}

	newEnv := Environment{
		Name:      name,
		Variables: []Variable{},
	}

	config.Environments = append(config.Environments, newEnv)

	if config.ActiveEnvironment == "" {
		config.ActiveEnvironment = name
	}

	return s.SaveEnvironments(config)
}

func (s *Storage) DeleteEnvironment(name string) error {
	config, err := s.LoadEnvironments()
	if err != nil {
		return err
	}

	for i, env := range config.Environments {
		if env.Name == name {
			config.Environments = append(config.Environments[:i], config.Environments[i+1:]...)

			if config.ActiveEnvironment == name {
				if len(config.Environments) > 0 {
					config.ActiveEnvironment = config.Environments[0].Name
				} else {
					config.ActiveEnvironment = ""
				}
			}

			return s.SaveEnvironments(config)
		}
	}

	return fmt.Errorf("environment not found: %s", name)
}

func (s *Storage) SetActiveEnvironment(name string) error {
	config, err := s.LoadEnvironments()
	if err != nil {
		return err
	}

	found := false
	for _, env := range config.Environments {
		if env.Name == name {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("environment not found: %s", name)
	}

	config.ActiveEnvironment = name
	return s.SaveEnvironments(config)
}

func (s *Storage) AddVariable(envName, key, value string) error {
	config, err := s.LoadEnvironments()
	if err != nil {
		return err
	}

	for i, env := range config.Environments {
		if env.Name == envName {
			for j, v := range env.Variables {
				if v.Key == key {
					config.Environments[i].Variables[j].Value = value
					return s.SaveEnvironments(config)
				}
			}

			config.Environments[i].Variables = append(config.Environments[i].Variables, Variable{
				Key:   key,
				Value: value,
			})
			return s.SaveEnvironments(config)
		}
	}

	return fmt.Errorf("environment not found: %s", envName)
}

func (s *Storage) DeleteVariable(envName, key string) error {
	config, err := s.LoadEnvironments()
	if err != nil {
		return err
	}

	for i, env := range config.Environments {
		if env.Name == envName {
			for j, v := range env.Variables {
				if v.Key == key {
					config.Environments[i].Variables = append(
						config.Environments[i].Variables[:j],
						config.Environments[i].Variables[j+1:]...,
					)
					return s.SaveEnvironments(config)
				}
			}
			return fmt.Errorf("variable not found: %s", key)
		}
	}

	return fmt.Errorf("environment not found: %s", envName)
}

// ReplaceVariables replaces {{VARIABLE}} placeholders with their values
// Uses a pre-compiled regex and map for O(1) lookups instead of O(n)
func ReplaceVariables(text string, variables []Variable) string {
	// Build a map for O(1) lookup instead of O(n) linear search
	varMap := make(map[string]string, len(variables))
	for _, v := range variables {
		varMap[v.Key] = v.Value
	}

	// Use pre-compiled regex for better performance
	result := variableRegex.ReplaceAllStringFunc(text, func(match string) string {
		// Extract variable name (remove {{ and }})
		varName := strings.TrimSpace(match[2 : len(match)-2])

		// O(1) map lookup instead of O(n) loop
		if value, ok := varMap[varName]; ok {
			return value
		}

		// Keep original if variable not found
		return match
	})

	return result
}

func (s *Storage) GetActiveEnvironmentVariables() ([]Variable, error) {
	config, err := s.LoadEnvironments()
	if err != nil {
		return nil, err
	}

	if config.ActiveEnvironment == "" {
		return []Variable{}, nil
	}

	for _, env := range config.Environments {
		if env.Name == config.ActiveEnvironment {
			return env.Variables, nil
		}
	}

	return []Variable{}, nil
}
