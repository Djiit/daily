package output

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"daily/internal/activity"
)

func TestFormatter_FormatSummary(t *testing.T) {
	formatter := NewFormatter()
	
	date := time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)
	activities := []activity.Activity{
		{
			ID:          "1",
			Type:        activity.ActivityTypeCommit,
			Title:       "Fix bug in authentication",
			Description: "Commit in user-service",
			URL:         "https://github.com/user/repo/commit/123",
			Platform:    "github",
			Timestamp:   time.Date(2023, 12, 25, 9, 30, 0, 0, time.UTC),
			Tags:        []string{"user-service"},
		},
		{
			ID:          "2",
			Type:        activity.ActivityTypeJiraTicket,
			Title:       "PROJ-123: Implement user login",
			Description: "Status: In Progress",
			URL:         "https://company.atlassian.net/browse/PROJ-123",
			Platform:    "jira",
			Timestamp:   time.Date(2023, 12, 25, 14, 15, 0, 0, time.UTC),
			Tags:        []string{"PROJ-123", "In Progress"},
		},
	}

	summary := &activity.Summary{
		Date:       date,
		Activities: activities,
	}

	result := formatter.FormatSummary(summary)

	// Check that the output contains expected elements
	if !strings.Contains(result, "Daily Summary") {
		t.Error("Output should contain 'Daily Summary'")
	}
	
	if !strings.Contains(result, "December 25, 2023") {
		t.Error("Output should contain formatted date")
	}
	
	if !strings.Contains(result, "2 activities") {
		t.Error("Output should show activity count")
	}
	
	if !strings.Contains(result, "Fix bug in authentication") {
		t.Error("Output should contain first activity title")
	}
	
	if !strings.Contains(result, "PROJ-123: Implement user login") {
		t.Error("Output should contain second activity title")
	}
	
	if !strings.Contains(result, "🐙 Github") {
		t.Error("Output should contain GitHub section header")
	}
	
	if !strings.Contains(result, "🎫 Jira") {
		t.Error("Output should contain JIRA section header")
	}
}

func TestFormatter_FormatSummary_Empty(t *testing.T) {
	formatter := NewFormatter()
	
	summary := &activity.Summary{
		Date:       time.Now(),
		Activities: []activity.Activity{},
	}

	result := formatter.FormatSummary(summary)

	expected := "No activities found for this date."
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestFormatter_FormatCompactSummary(t *testing.T) {
	formatter := NewFormatter()
	
	activities := []activity.Activity{
		{
			ID:        "1",
			Type:      activity.ActivityTypeCommit,
			Title:     "Fix bug",
			Platform:  "github",
			Timestamp: time.Date(2023, 12, 25, 9, 30, 0, 0, time.UTC),
		},
		{
			ID:        "2",
			Type:      activity.ActivityTypeNote,
			Title:     "Meeting notes",
			Platform:  "obsidian",
			Timestamp: time.Date(2023, 12, 25, 10, 0, 0, 0, time.UTC),
		},
	}

	summary := &activity.Summary{
		Date:       time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC),
		Activities: activities,
	}

	result := formatter.FormatCompactSummary(summary)

	if !strings.Contains(result, "Daily Summary - 2 activities") {
		t.Error("Output should contain activity count")
	}
	
	if !strings.Contains(result, "09:30 [github] Fix bug") {
		t.Error("Output should contain first activity in compact format")
	}
	
	if !strings.Contains(result, "10:00 [obsidian] Meeting notes") {
		t.Error("Output should contain second activity in compact format")
	}
}

func TestFormatter_GetPlatformIcon(t *testing.T) {
	formatter := NewFormatter()
	
	tests := []struct {
		platform string
		expected string
	}{
		{"github", "🐙"},
		{"jira", "🎫"},
		{"obsidian", "📝"},
		{"unknown", "📌"},
	}

	for _, tt := range tests {
		t.Run(tt.platform, func(t *testing.T) {
			result := formatter.getPlatformIcon(tt.platform)
			if result != tt.expected {
				t.Errorf("Expected icon '%s' for platform '%s', got '%s'", tt.expected, tt.platform, result)
			}
		})
	}
}

func TestFormatter_GetTypeIcon(t *testing.T) {
	formatter := NewFormatter()
	
	tests := []struct {
		actType  activity.ActivityType
		expected string
	}{
		{activity.ActivityTypeCommit, "💾"},
		{activity.ActivityTypePR, "🔀"},
		{activity.ActivityTypeJiraTicket, "🎯"},
		{activity.ActivityTypeNote, "📄"},
		{activity.ActivityType("unknown"), "📋"},
	}

	for _, tt := range tests {
		t.Run(string(tt.actType), func(t *testing.T) {
			result := formatter.getTypeIcon(tt.actType)
			if result != tt.expected {
				t.Errorf("Expected icon '%s' for type '%s', got '%s'", tt.expected, tt.actType, result)
			}
		})
	}
}

func TestFormatter_FormatJSON(t *testing.T) {
	formatter := NewFormatter()
	
	date := time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)
	activities := []activity.Activity{
		{
			ID:          "1",
			Type:        activity.ActivityTypeCommit,
			Title:       "Fix bug in authentication",
			Description: "Commit in user-service",
			URL:         "https://github.com/user/repo/commit/123",
			Platform:    "github",
			Timestamp:   time.Date(2023, 12, 25, 9, 30, 0, 0, time.UTC),
			Tags:        []string{"user-service"},
		},
		{
			ID:          "2",
			Type:        activity.ActivityTypeJiraTicket,
			Title:       "PROJ-123: Implement user login",
			Description: "Status: In Progress",
			URL:         "https://company.atlassian.net/browse/PROJ-123",
			Platform:    "jira",
			Timestamp:   time.Date(2023, 12, 25, 14, 15, 0, 0, time.UTC),
			Tags:        []string{"PROJ-123", "In Progress"},
		},
	}

	summary := &activity.Summary{
		Date:       date,
		Activities: activities,
	}

	result := formatter.FormatJSON(summary)

	// Check that the output is valid JSON
	if !strings.Contains(result, `"date": "2023-12-25"`) {
		t.Error("JSON output should contain formatted date")
	}
	
	if !strings.Contains(result, `"total": 2`) {
		t.Error("JSON output should contain total count")
	}
	
	if !strings.Contains(result, `"Fix bug in authentication"`) {
		t.Error("JSON output should contain activity titles")
	}
	
	if !strings.Contains(result, `"by_platform"`) {
		t.Error("JSON output should contain platform summary")
	}
	
	if !strings.Contains(result, `"by_type"`) {
		t.Error("JSON output should contain type summary")
	}
	
	// Validate it's actually valid JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(result), &jsonData); err != nil {
		t.Errorf("Output should be valid JSON: %v", err)
	}
}

func TestFormatter_FormatJSON_Empty(t *testing.T) {
	formatter := NewFormatter()
	
	summary := &activity.Summary{
		Date:       time.Now(),
		Activities: []activity.Activity{},
	}

	result := formatter.FormatJSON(summary)

	// Check for empty activity list
	if !strings.Contains(result, `"activities": []`) {
		t.Error("JSON output should contain empty activities array")
	}
	
	if !strings.Contains(result, `"total": 0`) {
		t.Error("JSON output should show zero total")
	}
}