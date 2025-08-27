package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"daily/internal/activity"
)

// Cache manages cached summaries for historical dates
type Cache struct {
	cacheDir string
}

// NewCache creates a new cache instance
func NewCache() (*Cache, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	cacheDir := filepath.Join(homeDir, ".config", "daily", "cache")
	return &Cache{cacheDir: cacheDir}, nil
}

// Get retrieves a cached summary for the given date if it exists
func (c *Cache) Get(date time.Time) (*activity.Summary, error) {
	filename := c.getFilename(date)
	filePath := filepath.Join(c.cacheDir, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, nil // Not found, not an error
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	var summary activity.Summary
	if err := json.Unmarshal(data, &summary); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached summary: %w", err)
	}

	return &summary, nil
}

// Set stores a summary in the cache for the given date
// Only caches summaries for dates before today
func (c *Cache) Set(date time.Time, summary *activity.Summary) error {
	// Only cache historical dates (before today)
	today := time.Now().Truncate(24 * time.Hour)
	if !date.Truncate(24 * time.Hour).Before(today) {
		return nil // Don't cache today or future dates
	}

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(c.cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	filename := c.getFilename(date)
	filePath := filepath.Join(c.cacheDir, filename)

	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal summary: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// ShouldCache determines if a date should be cached
func (c *Cache) ShouldCache(date time.Time) bool {
	today := time.Now().Truncate(24 * time.Hour)
	return date.Truncate(24 * time.Hour).Before(today)
}

// getFilename generates a filename for the given date
func (c *Cache) getFilename(date time.Time) string {
	return fmt.Sprintf("summary_%s.json", date.Format("2006-01-02"))
}

// Clear removes all cached files (useful for testing or manual cleanup)
func (c *Cache) Clear() error {
	if _, err := os.Stat(c.cacheDir); os.IsNotExist(err) {
		return nil // Directory doesn't exist, nothing to clear
	}

	entries, err := os.ReadDir(c.cacheDir)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		filePath := filepath.Join(c.cacheDir, entry.Name())
		if err := os.Remove(filePath); err != nil {
			return fmt.Errorf("failed to remove cache file %s: %w", entry.Name(), err)
		}
	}

	return nil
}
