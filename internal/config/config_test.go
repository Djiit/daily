package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("Expected config to be non-nil")
	}

	if config.GitHub.Enabled {
		t.Error("Expected GitHub to be disabled by default")
	}

	if config.JIRA.Enabled {
		t.Error("Expected JIRA to be disabled by default")
	}

	if config.Obsidian.Enabled {
		t.Error("Expected Obsidian to be disabled by default")
	}
}

func TestConfig_SaveAndLoad(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "daily-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override the config path for testing
	originalConfigPathFunc := configPathFunc
	testConfigPath := filepath.Join(tempDir, "config.json")
	configPathFunc = func() (string, error) {
		return testConfigPath, nil
	}
	defer func() { configPathFunc = originalConfigPathFunc }()

	// Create a test config
	originalConfig := DefaultConfig()
	originalConfig.GitHub.Enabled = true
	originalConfig.GitHub.Username = "testuser"
	originalConfig.GitHub.Token = "testtoken"
	originalConfig.JIRA.Enabled = true
	originalConfig.JIRA.URL = "https://test.atlassian.net"
	originalConfig.JIRA.Email = "test@example.com"
	originalConfig.JIRA.Token = "jiratoken"

	// Save the config
	err = originalConfig.Save()
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(testConfigPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Load the config
	loadedConfig, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify the loaded config matches the original
	if loadedConfig.GitHub.Enabled != originalConfig.GitHub.Enabled {
		t.Error("GitHub.Enabled mismatch")
	}
	if loadedConfig.GitHub.Username != originalConfig.GitHub.Username {
		t.Error("GitHub.Username mismatch")
	}
	if loadedConfig.GitHub.Token != originalConfig.GitHub.Token {
		t.Error("GitHub.Token mismatch")
	}
	if loadedConfig.JIRA.Enabled != originalConfig.JIRA.Enabled {
		t.Error("JIRA.Enabled mismatch")
	}
	if loadedConfig.JIRA.URL != originalConfig.JIRA.URL {
		t.Error("JIRA.URL mismatch")
	}
	if loadedConfig.JIRA.Email != originalConfig.JIRA.Email {
		t.Error("JIRA.Email mismatch")
	}
	if loadedConfig.JIRA.Token != originalConfig.JIRA.Token {
		t.Error("JIRA.Token mismatch")
	}
}

func TestLoad_CreatesDefaultConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "daily-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override the config path for testing
	originalConfigPathFunc := configPathFunc
	testConfigPath := filepath.Join(tempDir, "config.json")
	configPathFunc = func() (string, error) {
		return testConfigPath, nil
	}
	defer func() { configPathFunc = originalConfigPathFunc }()

	// Load config when file doesn't exist
	config, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify it created a default config
	if config.GitHub.Enabled {
		t.Error("Expected GitHub to be disabled in default config")
	}

	// Verify file was created
	if _, err := os.Stat(testConfigPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Verify file content is valid JSON
	data, err := os.ReadFile(testConfigPath)
	if err != nil {
		t.Fatalf("Failed to read created config file: %v", err)
	}

	var testConfig Config
	if err := json.Unmarshal(data, &testConfig); err != nil {
		t.Fatalf("Created config file is not valid JSON: %v", err)
	}
}