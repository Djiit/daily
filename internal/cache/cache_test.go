package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"daily/internal/activity"
)

func TestCache(t *testing.T) {
	// Create a temporary cache directory for testing
	tempDir := t.TempDir()
	cache := &Cache{cacheDir: tempDir}

	// Test data
	testDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	testSummary := &activity.Summary{
		Date: testDate,
		Activities: []activity.Activity{
			{
				ID:          "test-1",
				Type:        activity.ActivityTypeCommit,
				Title:       "Test commit",
				Description: "A test commit",
				Platform:    "github",
				Timestamp:   testDate,
			},
		},
	}

	// Test Set (should cache historical date)
	err := cache.Set(testDate, testSummary)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	// Verify file was created
	expectedFile := filepath.Join(tempDir, "summary_2024-01-01.json")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Fatalf("Cache file was not created: %s", expectedFile)
	}

	// Test Get
	cachedSummary, err := cache.Get(testDate)
	if err != nil {
		t.Fatalf("Failed to get cached data: %v", err)
	}

	if cachedSummary == nil {
		t.Fatal("Expected cached summary, got nil")
	}

	if len(cachedSummary.Activities) != 1 {
		t.Fatalf("Expected 1 activity, got %d", len(cachedSummary.Activities))
	}

	if cachedSummary.Activities[0].ID != "test-1" {
		t.Fatalf("Expected activity ID 'test-1', got '%s'", cachedSummary.Activities[0].ID)
	}
}

func TestShouldCache(t *testing.T) {
	cache := &Cache{}

	// Historical date should be cached
	yesterday := time.Now().AddDate(0, 0, -1)
	if !cache.ShouldCache(yesterday) {
		t.Error("Expected yesterday to be cacheable")
	}

	// Today should not be cached
	today := time.Now()
	if cache.ShouldCache(today) {
		t.Error("Expected today to not be cacheable")
	}

	// Future date should not be cached
	tomorrow := time.Now().AddDate(0, 0, 1)
	if cache.ShouldCache(tomorrow) {
		t.Error("Expected tomorrow to not be cacheable")
	}
}

func TestSetTodayNotCached(t *testing.T) {
	// Create a temporary cache directory for testing
	tempDir := t.TempDir()
	cache := &Cache{cacheDir: tempDir}

	// Test that today's date is not cached
	today := time.Now()
	testSummary := &activity.Summary{
		Date:       today,
		Activities: []activity.Activity{},
	}

	err := cache.Set(today, testSummary)
	if err != nil {
		t.Fatalf("Set should not fail for today: %v", err)
	}

	// Verify no file was created
	expectedFile := filepath.Join(tempDir, cache.getFilename(today))
	if _, err := os.Stat(expectedFile); !os.IsNotExist(err) {
		t.Fatal("Today's summary should not be cached")
	}
}

func TestGetNonExistentCache(t *testing.T) {
	// Create a temporary cache directory for testing
	tempDir := t.TempDir()
	cache := &Cache{cacheDir: tempDir}

	// Test getting non-existent cache
	testDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	summary, err := cache.Get(testDate)
	if err != nil {
		t.Fatalf("Get should not fail for non-existent cache: %v", err)
	}

	if summary != nil {
		t.Error("Expected nil summary for non-existent cache")
	}
}
