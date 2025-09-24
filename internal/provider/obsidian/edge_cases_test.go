package obsidian

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"daily/internal/provider"
)

func TestProvider_EdgeCases_InvalidVaultPath(t *testing.T) {
	config := provider.Config{
		URL:     "/nonexistent/path",
		Enabled: true,
	}

	p := NewProvider(config)

	_, err := p.GetTasks(context.Background())
	if err == nil {
		t.Error("Expected error for nonexistent vault path, got nil")
	}
}

func TestProvider_EdgeCases_EmptyVault(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "obsidian-empty-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	config := provider.Config{
		URL:     tempDir,
		Enabled: true,
	}

	p := NewProvider(config)

	tasks, err := p.GetTasks(context.Background())
	if err != nil {
		t.Errorf("Expected no error for empty vault, got: %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks in empty vault, got %d", len(tasks))
	}
}

func TestProvider_EdgeCases_NonMarkdownFiles(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "obsidian-non-md-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create non-markdown files with task-like content
	testFiles := []struct {
		name    string
		content string
	}{
		{
			name:    "tasks.txt",
			content: "- [ ] This should not be parsed\n- [/] Neither should this",
		},
		{
			name:    "notes.doc",
			content: "- [ ] Document task\n- [x] Completed doc task",
		},
		{
			name:    "real-tasks.md",
			content: "# Real Markdown\n\n- [ ] This should be parsed\n- [/] This too",
		},
	}

	for _, tf := range testFiles {
		filePath := filepath.Join(tempDir, tf.name)
		err := os.WriteFile(filePath, []byte(tf.content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", tf.name, err)
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

	// Should only find tasks from .md files
	expectedTasks := 2 // Only from real-tasks.md
	if len(tasks) != expectedTasks {
		t.Errorf("Expected %d tasks (only from .md files), got %d", expectedTasks, len(tasks))
		for i, task := range tasks {
			t.Logf("Task %d: %s", i, task.Title)
		}
	}
}

func TestProvider_EdgeCases_CorruptedFiles(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "obsidian-corrupt-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create files with various problematic content
	testFiles := []struct {
		name    string
		content string
	}{
		{
			name:    "binary-data.md",
			content: "\x00\x01\x02- [ ] Task after binary data",
		},
		{
			name:    "extremely-long-line.md",
			content: "# Long Line\n\n- [ ] " + strings.Repeat("x", 10000) + " very long task",
		},
		{
			name:    "unicode.md",
			content: "# Unicode\n\n- [ ] Task with émojis 🚀 and special chars áéíóú\n- [/] 中文任务 with Chinese chars",
		},
	}

	for _, tf := range testFiles {
		filePath := filepath.Join(tempDir, tf.name)
		err := os.WriteFile(filePath, []byte(tf.content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", tf.name, err)
		}
	}

	config := provider.Config{
		URL:     tempDir,
		Enabled: true,
	}

	p := NewProvider(config)

	tasks, err := p.GetTasks(context.Background())
	if err != nil {
		t.Errorf("Expected no error handling corrupted files, got: %v", err)
	}

	// Should handle all files gracefully and extract valid tasks
	if len(tasks) == 0 {
		t.Error("Expected to find some tasks despite file issues")
	}

	// Verify all returned tasks are valid
	for i, task := range tasks {
		if task.Title == "" {
			t.Errorf("Task %d has empty title", i)
		}
		if task.ID == "" {
			t.Errorf("Task %d has empty ID", i)
		}
	}
}

func TestProvider_EdgeCases_DeepNestedDirectories(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "obsidian-nested-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create deeply nested directory structure
	deepPath := filepath.Join(tempDir, "level1", "level2", "level3", "level4")
	err = os.MkdirAll(deepPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create deep directory: %v", err)
	}

	// Create task file in deep directory
	taskFile := filepath.Join(deepPath, "deep-tasks.md")
	taskContent := `# Deep Tasks

- [ ] Task in deeply nested directory
- [/] Another deep task
`
	err = os.WriteFile(taskFile, []byte(taskContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create deep task file: %v", err)
	}

	config := provider.Config{
		URL:     tempDir,
		Enabled: true,
	}

	p := NewProvider(config)

	tasks, err := p.GetTasks(context.Background())
	if err != nil {
		t.Errorf("Expected no error with nested directories, got: %v", err)
	}

	// Should find tasks in nested directories
	expectedTasks := 2
	if len(tasks) != expectedTasks {
		t.Errorf("Expected %d tasks from nested directory, got %d", expectedTasks, len(tasks))
	}

	// Verify task IDs contain proper path
	for _, task := range tasks {
		if !strings.Contains(task.ID, "level1/level2/level3/level4/deep-tasks.md") {
			t.Errorf("Task ID should contain nested path, got: %s", task.ID)
		}
	}
}

func TestProvider_EdgeCases_LargeVault(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "obsidian-large-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create many files with tasks
	numFiles := 50
	tasksPerFile := 10

	for i := 0; i < numFiles; i++ {
		filePath := filepath.Join(tempDir, fmt.Sprintf("file%03d.md", i))

		var content strings.Builder
		content.WriteString(fmt.Sprintf("# File %d\n\n", i))

		for j := 0; j < tasksPerFile; j++ {
			if j%3 == 0 {
				content.WriteString(fmt.Sprintf("- [ ] Todo task %d-%d\n", i, j))
			} else if j%3 == 1 {
				content.WriteString(fmt.Sprintf("- [/] Ongoing task %d-%d\n", i, j))
			} else {
				content.WriteString(fmt.Sprintf("- [x] Completed task %d-%d\n", i, j))
			}
		}

		err := os.WriteFile(filePath, []byte(content.String()), 0644)
		if err != nil {
			t.Fatalf("Failed to create file %d: %v", i, err)
		}
	}

	config := provider.Config{
		URL:     tempDir,
		Enabled: true,
	}

	p := NewProvider(config)

	startTime := time.Now()
	tasks, err := p.GetTasks(context.Background())
	duration := time.Since(startTime)

	if err != nil {
		t.Errorf("Expected no error with large vault, got: %v", err)
	}

	// Each file should contribute ~6-7 incomplete tasks (todo + ongoing, excluding completed)
	expectedMinTasks := numFiles * 6
	expectedMaxTasks := numFiles * 8
	if len(tasks) < expectedMinTasks || len(tasks) > expectedMaxTasks {
		t.Errorf("Expected between %d and %d tasks, got %d", expectedMinTasks, expectedMaxTasks, len(tasks))
	}

	// Performance check - should complete reasonably quickly
	if duration > 5*time.Second {
		t.Errorf("Task parsing took too long: %v (consider optimization)", duration)
	}

	t.Logf("Parsed %d tasks from %d files in %v", len(tasks), numFiles, duration)
}
