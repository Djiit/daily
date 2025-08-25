package output

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss/v2"

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

	// Strip ANSI codes for testing
	plainResult := lipgloss.NewStyle().Render(result)
	plainResult = strings.ReplaceAll(plainResult, "\x1b[", "")

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

	// Check that the styled result contains the expected text
	if !strings.Contains(result, "No activities found for this date.") {
		t.Errorf("Output should contain 'No activities found for this date.', got: %s", result)
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

	if !strings.Contains(result, "Fix bug") {
		t.Error("Output should contain first activity in compact format")
	}

	if !strings.Contains(result, "Meeting notes") {
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

func TestFormatter_FormatTodo(t *testing.T) {
	formatter := NewFormatter()

	todoItems := TodoItems{
		GitHub: GitHubTodos{
			OpenPRs: []TodoItem{
				{
					ID:          "github-pr-123",
					Title:       "Fix authentication bug",
					Description: "Open PR in user-service",
					URL:         "https://github.com/user/repo/pull/123",
					UpdatedAt:   time.Date(2023, 12, 25, 10, 30, 0, 0, time.UTC),
					Tags:        []string{"user-service", "open"},
				},
			},
			PendingReviews: []TodoItem{
				{
					ID:          "github-review-456",
					Title:       "Add user registration",
					Description: "Review requested in auth-service",
					URL:         "https://github.com/user/auth/pull/456",
					UpdatedAt:   time.Date(2023, 12, 25, 11, 45, 0, 0, time.UTC),
					Tags:        []string{"auth-service", "review-requested"},
				},
			},
		},
		JIRA: JIRATodos{
			AssignedTickets: []TodoItem{
				{
					ID:          "jira-PROJ-789",
					Title:       "PROJ-789: Implement OAuth",
					Description: "Status: In Progress",
					URL:         "https://company.atlassian.net/browse/PROJ-789",
					UpdatedAt:   time.Date(2023, 12, 25, 9, 15, 0, 0, time.UTC),
					Tags:        []string{"PROJ-789", "In Progress"},
				},
			},
		},
	}

	result := formatter.FormatTodo(todoItems)

	// Check for basic structure
	if !strings.Contains(result, "Todo Items") {
		t.Error("Output should contain 'Todo Items' header")
	}

	if !strings.Contains(result, "Found 3 pending items") {
		t.Error("Output should show correct count of pending items")
	}

	if !strings.Contains(result, "Open Pull Requests") {
		t.Error("Output should contain 'Open Pull Requests' section")
	}

	if !strings.Contains(result, "Pending Reviews") {
		t.Error("Output should contain 'Pending Reviews' section")
	}

	if !strings.Contains(result, "Assigned Tickets") {
		t.Error("Output should contain 'Assigned Tickets' section")
	}

	if !strings.Contains(result, "Fix authentication bug") {
		t.Error("Output should contain PR title")
	}

	if !strings.Contains(result, "Add user registration") {
		t.Error("Output should contain review title")
	}

	if !strings.Contains(result, "PROJ-789: Implement OAuth") {
		t.Error("Output should contain JIRA ticket title")
	}
}

func TestFormatter_FormatTodo_Empty(t *testing.T) {
	formatter := NewFormatter()

	todoItems := TodoItems{
		GitHub: GitHubTodos{
			OpenPRs:        []TodoItem{},
			PendingReviews: []TodoItem{},
		},
		JIRA: JIRATodos{
			AssignedTickets: []TodoItem{},
		},
	}

	result := formatter.FormatTodo(todoItems)

	if !strings.Contains(result, "No pending items found") {
		t.Error("Output should show 'No pending items found' for empty todo list")
	}
}

func TestFormatter_FormatTodoJSON(t *testing.T) {
	formatter := NewFormatter()

	todoItems := TodoItems{
		GitHub: GitHubTodos{
			OpenPRs: []TodoItem{
				{
					ID:          "github-pr-123",
					Title:       "Fix authentication bug",
					Description: "Open PR in user-service",
					URL:         "https://github.com/user/repo/pull/123",
					UpdatedAt:   time.Date(2023, 12, 25, 10, 30, 0, 0, time.UTC),
					Tags:        []string{"user-service", "open"},
				},
			},
			PendingReviews: []TodoItem{},
		},
		JIRA: JIRATodos{
			AssignedTickets: []TodoItem{
				{
					ID:          "jira-PROJ-789",
					Title:       "PROJ-789: Implement OAuth",
					Description: "Status: In Progress",
					URL:         "https://company.atlassian.net/browse/PROJ-789",
					UpdatedAt:   time.Date(2023, 12, 25, 9, 15, 0, 0, time.UTC),
					Tags:        []string{"PROJ-789", "In Progress"},
				},
			},
		},
	}

	result := formatter.FormatTodoJSON(todoItems)

	// Parse JSON to verify it's valid
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	// Check summary section
	summary, ok := parsed["summary"].(map[string]interface{})
	if !ok {
		t.Fatal("JSON should contain summary section")
	}

	if summary["total"] != float64(2) {
		t.Errorf("Expected total 2, got %v", summary["total"])
	}

	if summary["open_prs"] != float64(1) {
		t.Errorf("Expected 1 open PR, got %v", summary["open_prs"])
	}

	if summary["pending_reviews"] != float64(0) {
		t.Errorf("Expected 0 pending reviews, got %v", summary["pending_reviews"])
	}

	if summary["assigned_tickets"] != float64(1) {
		t.Errorf("Expected 1 assigned ticket, got %v", summary["assigned_tickets"])
	}

	// Check GitHub section
	github, ok := parsed["github"].(map[string]interface{})
	if !ok {
		t.Fatal("JSON should contain github section")
	}

	openPRs, ok := github["open_prs"].([]interface{})
	if !ok {
		t.Fatal("GitHub section should contain open_prs array")
	}

	if len(openPRs) != 1 {
		t.Errorf("Expected 1 open PR, got %d", len(openPRs))
	}

	// Check JIRA section
	jira, ok := parsed["jira"].(map[string]interface{})
	if !ok {
		t.Fatal("JSON should contain jira section")
	}

	assignedTickets, ok := jira["assigned_tickets"].([]interface{})
	if !ok {
		t.Fatal("JIRA section should contain assigned_tickets array")
	}

	if len(assignedTickets) != 1 {
		t.Errorf("Expected 1 assigned ticket, got %d", len(assignedTickets))
	}
}

func TestFormatter_FormatTodoJSON_Empty(t *testing.T) {
	formatter := NewFormatter()

	todoItems := TodoItems{
		GitHub: GitHubTodos{
			OpenPRs:        []TodoItem{},
			PendingReviews: []TodoItem{},
		},
		JIRA: JIRATodos{
			AssignedTickets: []TodoItem{},
		},
	}

	result := formatter.FormatTodoJSON(todoItems)

	// Parse JSON to verify it's valid
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	// Check summary section for empty state
	summary, ok := parsed["summary"].(map[string]interface{})
	if !ok {
		t.Fatal("JSON should contain summary section")
	}

	if summary["total"] != float64(0) {
		t.Errorf("Expected total 0, got %v", summary["total"])
	}
}
