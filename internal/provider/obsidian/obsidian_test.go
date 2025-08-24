package obsidian

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"daily/internal/provider"
)

func TestProvider_Name(t *testing.T) {
	config := provider.Config{
		URL:     "/path/to/vault",
		Enabled: true,
	}

	p := NewProvider(config)

	if p.Name() != "obsidian" {
		t.Errorf("Expected provider name to be 'obsidian', got '%s'", p.Name())
	}
}

func TestProvider_IsConfigured(t *testing.T) {
	tests := []struct {
		name     string
		config   provider.Config
		expected bool
	}{
		{
			name: "fully configured",
			config: provider.Config{
				URL:     "/path/to/vault",
				Enabled: true,
			},
			expected: true,
		},
		{
			name: "disabled",
			config: provider.Config{
				URL:     "/path/to/vault",
				Enabled: false,
			},
			expected: false,
		},
		{
			name: "missing vault path",
			config: provider.Config{
				URL:     "",
				Enabled: true,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProvider(tt.config)

			if p.IsConfigured() != tt.expected {
				t.Errorf("Expected IsConfigured() to return %t, got %t", tt.expected, p.IsConfigured())
			}
		})
	}
}

func TestProvider_GetActivities(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "obsidian-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test markdown files
	testFiles := []struct {
		name    string
		content string
		modTime time.Time
	}{
		{
			name:    "note1.md",
			content: "# Test Note 1\nThis is a test note",
			modTime: time.Now().Add(-2 * time.Hour),
		},
		{
			name:    "note2.md",
			content: "# Test Note 2\nAnother test note",
			modTime: time.Now().Add(-1 * time.Hour),
		},
		{
			name:    "old-note.md",
			content: "# Old Note\nThis is an old note",
			modTime: time.Now().Add(-25 * time.Hour), // Outside our range
		},
	}

	for _, tf := range testFiles {
		filePath := filepath.Join(tempDir, tf.name)
		err := os.WriteFile(filePath, []byte(tf.content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", tf.name, err)
		}

		// Set modification time
		err = os.Chtimes(filePath, tf.modTime, tf.modTime)
		if err != nil {
			t.Fatalf("Failed to set mod time for %s: %v", tf.name, err)
		}
	}

	config := provider.Config{
		URL:     tempDir,
		Enabled: true,
	}

	p := NewProvider(config)

	// Test with last 24 hours
	from := time.Now().Add(-24 * time.Hour)
	to := time.Now()

	activities, err := p.GetActivities(context.Background(), from, to)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if activities == nil {
		t.Error("Expected activities slice to be initialized, got nil")
	}

	// Should find 2 notes (note1.md and note2.md), but not old-note.md
	if len(activities) != 2 {
		t.Errorf("Expected 2 activities, got %d", len(activities))
	}

	// Check that all activities are notes
	for _, act := range activities {
		if act.Platform != "obsidian" {
			t.Errorf("Expected platform 'obsidian', got '%s'", act.Platform)
		}
	}
}

func TestProvider_GetActivities_NotConfigured(t *testing.T) {
	config := provider.Config{
		URL:     "",
		Enabled: false,
	}

	p := NewProvider(config)

	from := time.Now().AddDate(0, 0, -1)
	to := time.Now()

	_, err := p.GetActivities(context.Background(), from, to)

	if err == nil {
		t.Error("Expected error for unconfigured provider, got nil")
	}
}
