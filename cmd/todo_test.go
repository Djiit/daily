package cmd

import (
	"context"
	"strings"
	"testing"

	"daily/internal/output"
	"daily/internal/provider"
	"daily/internal/provider/github"
	"daily/internal/provider/jira"
)

func TestGetGitHubTodos(t *testing.T) {
	tests := []struct {
		name           string
		config         provider.Config
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "unconfigured provider",
			config: provider.Config{
				Username: "",
				Token:    "",
				Enabled:  false,
			},
			expectError:    true,
			expectedErrMsg: "failed to get open PRs: GitHub provider not configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := github.NewProvider(tt.config)

			todos, err := getGitHubTodos(context.Background(), provider)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if tt.expectedErrMsg != "" && !strings.Contains(err.Error(), tt.expectedErrMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.expectedErrMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				// Verify the structure
				if todos.OpenPRs == nil {
					t.Error("Expected non-nil OpenPRs slice")
				}
				if todos.PendingReviews == nil {
					t.Error("Expected non-nil PendingReviews slice")
				}
			}
		})
	}
}

func TestGetJIRATodos(t *testing.T) {
	tests := []struct {
		name           string
		config         provider.Config
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "unconfigured provider",
			config: provider.Config{
				Email:   "",
				Token:   "",
				URL:     "",
				Enabled: false,
			},
			expectError:    true,
			expectedErrMsg: "failed to get assigned tickets: JIRA provider not configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := jira.NewProvider(tt.config)

			todos, err := getJIRATodos(context.Background(), provider)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if tt.expectedErrMsg != "" && !strings.Contains(err.Error(), tt.expectedErrMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.expectedErrMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				// Verify the structure
				if todos.AssignedTickets == nil {
					t.Error("Expected non-nil AssignedTickets slice")
				}
			}
		})
	}
}

func TestTodoCmd_Creation(t *testing.T) {
	cmd := TodoCmd()

	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}

	if cmd.Use != "todo" {
		t.Errorf("Expected command use to be 'todo', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected non-empty short description")
	}

	if cmd.Long == "" {
		t.Error("Expected non-empty long description")
	}

	// Check flags
	verboseFlag := cmd.Flags().Lookup("verbose")
	if verboseFlag == nil {
		t.Error("Expected verbose flag to be defined")
	}

	outputFlag := cmd.Flags().Lookup("output")
	if outputFlag == nil {
		t.Error("Expected output flag to be defined")
	}
}

func TestTodoItemConversion(t *testing.T) {
	// Test that we can convert between different TodoItem types
	githubTodoItem := github.TodoItem{
		ID:          "test-id",
		Title:       "Test Title",
		Description: "Test Description",
		URL:         "https://example.com",
		Tags:        []string{"tag1", "tag2"},
	}

	outputTodoItem := output.TodoItem{
		ID:          githubTodoItem.ID,
		Title:       githubTodoItem.Title,
		Description: githubTodoItem.Description,
		URL:         githubTodoItem.URL,
		UpdatedAt:   githubTodoItem.UpdatedAt,
		Tags:        githubTodoItem.Tags,
	}

	if outputTodoItem.ID != githubTodoItem.ID {
		t.Errorf("Expected ID %s, got %s", githubTodoItem.ID, outputTodoItem.ID)
	}

	if outputTodoItem.Title != githubTodoItem.Title {
		t.Errorf("Expected Title %s, got %s", githubTodoItem.Title, outputTodoItem.Title)
	}

	if outputTodoItem.Description != githubTodoItem.Description {
		t.Errorf("Expected Description %s, got %s", githubTodoItem.Description, outputTodoItem.Description)
	}

	if outputTodoItem.URL != githubTodoItem.URL {
		t.Errorf("Expected URL %s, got %s", githubTodoItem.URL, outputTodoItem.URL)
	}

	if len(outputTodoItem.Tags) != len(githubTodoItem.Tags) {
		t.Errorf("Expected %d tags, got %d", len(githubTodoItem.Tags), len(outputTodoItem.Tags))
	}
}

func TestTodoCmd_FlagValidation(t *testing.T) {
	cmd := TodoCmd()

	// Test with invalid output format
	cmd.SetArgs([]string{"--output", "invalid"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid output format, got nil")
	}

	expectedErrMsg := "invalid output format: invalid"
	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("Expected error message to contain '%s', got '%s'", expectedErrMsg, err.Error())
	}
}
