package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Users       map[string]string `yaml:"users"` // username: password (plaintext in MVP)
	BrowserPath string            `yaml:"browserPath"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if len(cfg.Users) == 0 {
		return nil, errors.New("no users defined in config")
	}

	return &cfg, nil
}

// ValidateUser checks if username and password match
func (c *Config) ValidateUser(username, password string) bool {
	storedPassword, exists := c.Users[username]
	return exists && storedPassword == password
}
