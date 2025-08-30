package obsidian

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"daily/internal/provider"
)

func TestProvider_GetTasks(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "obsidian-tasks-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create test markdown files with tasks
	testFiles := []struct {
		name    string
		content string
		modTime time.Time
	}{
		{
			name: "tasks.md",
			content: `# Test Tasks

## Todo
- [ ] Review pull request
- [/] Fix authentication bug
- [x] Update documentation (completed, should not appear)

## Numbered Tasks
1. [ ] First numbered task
2. [/] Second ongoing task
3. [x] Third completed task (should not appear)

## Mixed Content
Some regular text here.

- [ ] Another todo task
- Regular bullet point (not a task)

` + "```markdown\n- [ ] Task in code block (should not appear)\n```" + `

> - [ ] Task in blockquote (should not appear)
`,
			modTime: time.Now(),
		},
		{
			name: "empty.md",
			content: `# Empty Note

No tasks here.
`,
			modTime: time.Now(),
		},
		{
			name: "only-completed.md",
			content: `# Completed Tasks Only

- [x] Done task 1
- [x] Done task 2
`,
			modTime: time.Now(),
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

	tasks, err := p.GetTasks(context.Background())

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if tasks == nil {
		t.Error("Expected tasks slice to be initialized, got nil")
	}

	// Should find 5 tasks: 2 from first file (review PR, fix auth), 2 numbered tasks, 1 additional todo
	expectedTasks := 5
	if len(tasks) != expectedTasks {
		t.Errorf("Expected %d tasks, got %d", expectedTasks, len(tasks))
		for i, task := range tasks {
			t.Logf("Task %d: %s", i, task.Title)
		}
	}

	// Verify task content
	found := make(map[string]bool)
	for _, task := range tasks {
		found[task.Title] = true
		// Verify all tasks have proper IDs
		if !strings.HasPrefix(task.ID, "obsidian-task-") {
			t.Errorf("Task ID should start with 'obsidian-task-', got: %s", task.ID)
		}
		// Verify URL format
		if !strings.Contains(task.URL, "obsidian://open") {
			t.Errorf("Task URL should contain 'obsidian://open', got: %s", task.URL)
		}
	}

	// Check specific tasks exist
	expectedTaskTitles := []string{
		"Review pull request",
		"Fix authentication bug",
		"First numbered task",
		"Second ongoing task",
		"Another todo task",
	}

	for _, expectedTitle := range expectedTaskTitles {
		if !found[expectedTitle] {
			t.Errorf("Expected to find task '%s', but it was not found", expectedTitle)
		}
	}

	// Ensure completed tasks are not included
	completedTasks := []string{
		"Update documentation (completed, should not appear)",
		"Third completed task (should not appear)",
	}

	for _, completedTitle := range completedTasks {
		if found[completedTitle] {
			t.Errorf("Completed task '%s' should not be included in todo list", completedTitle)
		}
	}
}

func TestProvider_GetTasks_NotConfigured(t *testing.T) {
	config := provider.Config{
		URL:     "",
		Enabled: false,
	}

	p := NewProvider(config)

	_, err := p.GetTasks(context.Background())

	if err == nil {
		t.Error("Expected error for unconfigured provider, got nil")
	}
}

func TestProvider_parseTasksFromFile(t *testing.T) {
	// Create a temporary file for testing
	tempDir, err := os.MkdirTemp("", "obsidian-parse-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	testContent := `# Test File

## Tasks
- [ ] Todo task
- [/] Ongoing task
- [x] Completed task

## Numbered
1. [ ] Numbered todo
2. [/] Numbered ongoing
3. [x] Numbered completed

## Should be ignored
` + "```\n- [ ] Task in code block\n```" + `

> - [ ] Task in blockquote

- Regular bullet point
`

	filePath := filepath.Join(tempDir, "test.md")
	err = os.WriteFile(filePath, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Failed to stat test file: %v", err)
	}

	config := provider.Config{
		URL:     tempDir,
		Enabled: true,
	}
	p := NewProvider(config)

	tasks, err := p.parseTasksFromFile(filePath, fileInfo)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Should find 4 tasks: 2 regular + 2 numbered (excluding completed ones)
	expected := 4
	if len(tasks) != expected {
		t.Errorf("Expected %d tasks, got %d", expected, len(tasks))
		for i, task := range tasks {
			t.Logf("Task %d: %s", i, task.Title)
		}
	}

	// Verify specific tasks
	found := make(map[string]bool)
	for _, task := range tasks {
		found[task.Title] = true
	}

	expectedTasks := []string{"Todo task", "Ongoing task", "Numbered todo", "Numbered ongoing"}
	for _, expectedTask := range expectedTasks {
		if !found[expectedTask] {
			t.Errorf("Expected to find task '%s'", expectedTask)
		}
	}

	// Verify completed tasks are not included
	completedTasks := []string{"Completed task", "Numbered completed"}
	for _, completedTask := range completedTasks {
		if found[completedTask] {
			t.Errorf("Completed task '%s' should not be included", completedTask)
		}
	}
}

func TestProvider_GetTasks_WithTestData(t *testing.T) {
	// Use our test fixtures
	testDataDir := "./testdata"

	// Check if testdata directory exists
	if _, err := os.Stat(testDataDir); os.IsNotExist(err) {
		t.Skip("testdata directory not found, skipping fixture tests")
	}

	config := provider.Config{
		URL:     testDataDir,
		Enabled: true,
	}

	p := NewProvider(config)

	tasks, err := p.GetTasks(context.Background())

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// We should find tasks from our test fixtures
	if len(tasks) == 0 {
		t.Error("Expected to find tasks in test fixtures, got 0")
	}

	// Verify all tasks have required fields
	for i, task := range tasks {
		if task.ID == "" {
			t.Errorf("Task %d has empty ID", i)
		}
		if task.Title == "" {
			t.Errorf("Task %d has empty title", i)
		}
		if task.Description == "" {
			t.Errorf("Task %d has empty description", i)
		}
		if !strings.Contains(task.URL, "obsidian://") {
			t.Errorf("Task %d URL should contain obsidian://, got: %s", i, task.URL)
		}
		if task.UpdatedAt.IsZero() {
			t.Errorf("Task %d has zero UpdatedAt time", i)
		}
	}
}

func TestProvider_extractTags(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "no-tags",
			text:     "Simple task with no tags",
			expected: []string{},
		},
		{
			name:     "hashtag",
			text:     "Task with #work tag",
			expected: []string{"work"},
		},
		{
			name:     "multiple-hashtags",
			text:     "Task with #work and #important tags",
			expected: []string{"work", "important"},
		},
		{
			name:     "priority-fire",
			text:     "🔥 Critical task",
			expected: []string{"high-priority"},
		},
		{
			name:     "priority-star",
			text:     "⭐ Important task",
			expected: []string{"starred"},
		},
		{
			name:     "scheduled-calendar",
			text:     "Task with 📅 due date",
			expected: []string{"scheduled"},
		},
		{
			name:     "mixed-tags",
			text:     "🔥 #urgent task with ⭐ priority",
			expected: []string{"urgent", "high-priority", "starred"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTags(tt.text)

			// Check length
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d tags, got %d: %v", len(tt.expected), len(result), result)
				return
			}

			// Check each expected tag is present
			for _, expectedTag := range tt.expected {
				found := false
				for _, actualTag := range result {
					if actualTag == expectedTag {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected tag '%s' not found in result: %v", expectedTag, result)
				}
			}
		})
	}
}

func TestProvider_createTodoItem(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "obsidian-item-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	filePath := filepath.Join(tempDir, "test.md")
	err = os.WriteFile(filePath, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Failed to stat test file: %v", err)
	}

	config := provider.Config{
		URL:     tempDir,
		Enabled: true,
	}
	p := NewProvider(config)

	taskText := "Review #urgent document with 🔥 priority"
	lineNum := 5

	item := p.createTodoItem(taskText, filePath, fileInfo, lineNum)

	// Verify basic fields
	if item.Title != taskText {
		t.Errorf("Expected title '%s', got '%s'", taskText, item.Title)
	}

	expectedID := "obsidian-task-test.md:5"
	if item.ID != expectedID {
		t.Errorf("Expected ID '%s', got '%s'", expectedID, item.ID)
	}

	if item.Description != "Task in test" {
		t.Errorf("Expected description 'Task in test', got '%s'", item.Description)
	}

	if !strings.Contains(item.URL, "obsidian://open") {
		t.Errorf("Expected URL to contain 'obsidian://open', got '%s'", item.URL)
	}

	if item.UpdatedAt != fileInfo.ModTime() {
		t.Errorf("Expected UpdatedAt to match file mod time")
	}

	// Verify tags were extracted
	expectedTags := []string{"urgent", "high-priority"}
	if len(item.Tags) != len(expectedTags) {
		t.Errorf("Expected %d tags, got %d: %v", len(expectedTags), len(item.Tags), item.Tags)
	}
}

func TestProvider_parseTasksFromFile_EdgeCases(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "obsidian-edge-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	tests := []struct {
		name        string
		content     string
		expectedLen int
		description string
	}{
		{
			name:        "empty-file",
			content:     "",
			expectedLen: 0,
			description: "Empty file should have no tasks",
		},
		{
			name:        "no-tasks",
			content:     "# Regular Note\n\nJust some text with no tasks.",
			expectedLen: 0,
			description: "File with no tasks should return empty list",
		},
		{
			name:        "only-completed",
			content:     "# Done\n\n- [x] Done task 1\n- [x] Done task 2",
			expectedLen: 0,
			description: "File with only completed tasks should return empty list",
		},
		{
			name:        "tasks-in-code-blocks",
			content:     "# Code\n\n```\n- [ ] Task in code block\n```\n\n- [ ] Real task",
			expectedLen: 1,
			description: "Should ignore tasks in code blocks",
		},
		{
			name:        "tasks-in-blockquotes",
			content:     "# Quotes\n\n> - [ ] Task in blockquote\n\n- [ ] Real task",
			expectedLen: 1,
			description: "Should ignore tasks in blockquotes",
		},
		{
			name:        "mixed-bullet-types",
			content:     "# Mixed\n\n- [ ] Dash task\n* [ ] Asterisk task\n+ [ ] Plus task",
			expectedLen: 3,
			description: "Should support all bullet types",
		},
		{
			name:        "indented-tasks",
			content:     "# Nested\n\n- [ ] Main task\n  - [ ] Nested task\n    - [/] Deep nested ongoing",
			expectedLen: 3,
			description: "Should support indented tasks",
		},
		{
			name:        "numbered-tasks-only",
			content:     "# Numbered\n\n1. [ ] First task\n2. [/] Second task\n3. [x] Completed task",
			expectedLen: 2,
			description: "Should support numbered tasks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tempDir, tt.name+".md")
			err := os.WriteFile(filePath, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			fileInfo, err := os.Stat(filePath)
			if err != nil {
				t.Fatalf("Failed to stat test file: %v", err)
			}

			config := provider.Config{
				URL:     tempDir,
				Enabled: true,
			}
			p := NewProvider(config)

			tasks, err := p.parseTasksFromFile(filePath, fileInfo)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if len(tasks) != tt.expectedLen {
				t.Errorf("%s: Expected %d tasks, got %d", tt.description, tt.expectedLen, len(tasks))
				for i, task := range tasks {
					t.Logf("Task %d: %s", i, task.Title)
				}
			}
		})
	}
}

func TestProvider_GetTasks_WithTestFixtures(t *testing.T) {
	// Use our test fixtures
	testDataDir := "./testdata"

	// Check if testdata directory exists
	if _, err := os.Stat(testDataDir); os.IsNotExist(err) {
		t.Skip("testdata directory not found, skipping fixture tests")
	}

	config := provider.Config{
		URL:     testDataDir,
		Enabled: true,
	}

	p := NewProvider(config)

	tasks, err := p.GetTasks(context.Background())

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// We should find tasks from our test fixtures
	if len(tasks) == 0 {
		t.Error("Expected to find tasks in test fixtures, got 0")
	}

	// Verify all tasks have required fields
	for i, task := range tasks {
		if task.ID == "" {
			t.Errorf("Task %d has empty ID", i)
		}
		if task.Title == "" {
			t.Errorf("Task %d has empty title", i)
		}
		if task.Description == "" {
			t.Errorf("Task %d has empty description", i)
		}
		if !strings.Contains(task.URL, "obsidian://") {
			t.Errorf("Task %d URL should contain obsidian://, got: %s", i, task.URL)
		}
		if task.UpdatedAt.IsZero() {
			t.Errorf("Task %d has zero UpdatedAt time", i)
		}
	}

	// Log all found tasks for debugging
	t.Logf("Found %d tasks:", len(tasks))
	for i, task := range tasks {
		t.Logf("  %d: %s (from %s)", i, task.Title, task.Description)
	}
}

func TestExtractTags(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "no-tags",
			text:     "Simple task with no tags",
			expected: []string{},
		},
		{
			name:     "hashtag",
			text:     "Task with #work tag",
			expected: []string{"work"},
		},
		{
			name:     "multiple-hashtags",
			text:     "Task with #work and #important tags",
			expected: []string{"work", "important"},
		},
		{
			name:     "priority-fire",
			text:     "🔥 Critical task",
			expected: []string{"high-priority"},
		},
		{
			name:     "priority-star",
			text:     "⭐ Important task",
			expected: []string{"starred"},
		},
		{
			name:     "scheduled-calendar",
			text:     "Task with 📅 due date",
			expected: []string{"scheduled"},
		},
		{
			name:     "mixed-tags",
			text:     "🔥 #urgent task with ⭐ priority",
			expected: []string{"urgent", "high-priority", "starred"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTags(tt.text)

			// Check length
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d tags, got %d: %v", len(tt.expected), len(result), result)
				return
			}

			// Check each expected tag is present
			for _, expectedTag := range tt.expected {
				found := false
				for _, actualTag := range result {
					if actualTag == expectedTag {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected tag '%s' not found in result: %v", expectedTag, result)
				}
			}
		})
	}
}