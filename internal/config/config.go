package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"daily/internal/provider"
)

type Config struct {
	GitHub   provider.Config `json:"github"`
	JIRA     provider.Config `json:"jira"`
	Obsidian provider.Config `json:"obsidian"`
}

func DefaultConfig() *Config {
	return &Config{
		GitHub: provider.Config{
			Enabled: false,
		},
		JIRA: provider.Config{
			Enabled: false,
		},
		Obsidian: provider.Config{
			Enabled: false,
		},
	}
}

func Load() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	// If config file doesn't exist, create default
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config := DefaultConfig()
		if err := config.Save(); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return config, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func (c *Config) Save() error {
	configPath, err := getConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// configPathFunc is a function variable to allow testing with different paths
var configPathFunc = defaultConfigPath

func defaultConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".config", "daily", "config.json"), nil
}

func getConfigPath() (string, error) {
	return configPathFunc()
}

func GetConfigPath() (string, error) {
	return getConfigPath()
}