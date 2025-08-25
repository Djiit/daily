package types

import (
	"testing"
	"time"
)

func TestTodoItem(t *testing.T) {
	updatedAt := time.Now()
	item := TodoItem{
		ID:          "test-id",
		Title:       "Test Item",
		Description: "Test Description",
		URL:         "https://example.com",
		UpdatedAt:   updatedAt,
		Tags:        []string{"test", "example"},
	}

	if item.ID != "test-id" {
		t.Errorf("Expected ID to be 'test-id', got '%s'", item.ID)
	}

	if item.Title != "Test Item" {
		t.Errorf("Expected Title to be 'Test Item', got '%s'", item.Title)
	}

	if item.Description != "Test Description" {
		t.Errorf("Expected Description to be 'Test Description', got '%s'", item.Description)
	}

	if item.URL != "https://example.com" {
		t.Errorf("Expected URL to be 'https://example.com', got '%s'", item.URL)
	}

	if !item.UpdatedAt.Equal(updatedAt) {
		t.Errorf("Expected UpdatedAt to be %v, got %v", updatedAt, item.UpdatedAt)
	}

	if len(item.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(item.Tags))
	}

	if item.Tags[0] != "test" || item.Tags[1] != "example" {
		t.Errorf("Expected tags ['test', 'example'], got %v", item.Tags)
	}
}

func TestTodoItems_Structure(t *testing.T) {
	item := TodoItem{
		ID:    "test-item",
		Title: "Test",
	}

	todoItems := TodoItems{
		GitHub: GitHubTodos{
			OpenPRs:        []TodoItem{item},
			PendingReviews: []TodoItem{item},
		},
		JIRA: JIRATodos{
			AssignedTickets: []TodoItem{item},
		},
	}

	if len(todoItems.GitHub.OpenPRs) != 1 {
		t.Errorf("Expected 1 open PR, got %d", len(todoItems.GitHub.OpenPRs))
	}

	if len(todoItems.GitHub.PendingReviews) != 1 {
		t.Errorf("Expected 1 pending review, got %d", len(todoItems.GitHub.PendingReviews))
	}

	if len(todoItems.JIRA.AssignedTickets) != 1 {
		t.Errorf("Expected 1 assigned ticket, got %d", len(todoItems.JIRA.AssignedTickets))
	}

	if todoItems.GitHub.OpenPRs[0].ID != "test-item" {
		t.Errorf("Expected first open PR ID to be 'test-item', got '%s'", todoItems.GitHub.OpenPRs[0].ID)
	}
}
