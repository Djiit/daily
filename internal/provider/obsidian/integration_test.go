package obsidian

import (
	"context"
	"os"
	"strings"
	"testing"

	"daily/internal/provider"
)

func TestProvider_Integration_VaultStructure(t *testing.T) {
	// Use the realistic vault structure we created
	vaultPath := "./testdata/vault-structure"

	// Check if vault structure exists
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		t.Skip("vault-structure testdata not found, skipping integration tests")
	}

	config := provider.Config{
		URL:     vaultPath,
		Enabled: true,
	}

	p := NewProvider(config)

	if !p.IsConfigured() {
		t.Fatal("Provider should be configured with vault path")
	}

	// Test task retrieval
	tasks, err := p.GetTasks(context.Background())
	if err != nil {
		t.Fatalf("Expected no error getting tasks, got: %v", err)
	}

	// We should find multiple tasks across different directories
	if len(tasks) == 0 {
		t.Error("Expected to find tasks in vault structure, got 0")
	}

	// Verify we have tasks from different files
	foundFiles := make(map[string]bool)
	for _, task := range tasks {
		// Extract file name from description
		if strings.Contains(task.Description, "Task in ") {
			fileName := strings.TrimPrefix(task.Description, "Task in ")
			foundFiles[fileName] = true
		}
	}

	// We should have tasks from multiple files
	expectedFiles := []string{"2024-08-30", "Web App Redesign", "API Migration"}
	for _, expectedFile := range expectedFiles {
		found := false
		for fileName := range foundFiles {
			if strings.Contains(fileName, expectedFile) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find tasks from file containing '%s'", expectedFile)
		}
	}

	// Verify task structure
	for i, task := range tasks {
		if task.ID == "" {
			t.Errorf("Task %d has empty ID", i)
		}
		if task.Title == "" {
			t.Errorf("Task %d has empty title", i)
		}
		if !strings.HasPrefix(task.ID, "obsidian-task-") {
			t.Errorf("Task %d ID should start with 'obsidian-task-', got: %s", i, task.ID)
		}
		if !strings.Contains(task.URL, "obsidian://open") {
			t.Errorf("Task %d URL should contain 'obsidian://open', got: %s", i, task.URL)
		}
	}

	// Log found tasks for debugging
	t.Logf("Integration test found %d tasks:", len(tasks))
	for i, task := range tasks {
		t.Logf("  %d: %s (%s)", i, task.Title, task.Description)
	}
}

func TestProvider_Integration_TaskTypes(t *testing.T) {
	vaultPath := "./testdata/vault-structure"

	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		t.Skip("vault-structure testdata not found, skipping integration tests")
	}

	config := provider.Config{
		URL:     vaultPath,
		Enabled: true,
	}

	p := NewProvider(config)
	tasks, err := p.GetTasks(context.Background())
	if err != nil {
		t.Fatalf("Expected no error getting tasks, got: %v", err)
	}

	// Count different task types based on expected content
	todoCount := 0
	ongoingCount := 0

	taskTypeCounts := make(map[string]int)
	for _, task := range tasks {
		// We can't directly test the checkbox type since we only store incomplete tasks
		// But we can verify the tasks we expect to be found
		taskTypeCounts[task.Title]++
	}

	// Count expected tasks
	expectedTodoTasks := []string{
		"Morning standup meeting",
		"Update API documentation",
		"Client call at 3 PM",
		"High-fidelity mockups",
		"Prototype development",
		"Usability testing",
		"Write migration documentation",
		"Plan rollout strategy",
		"Mobile app integration",
		"Web dashboard updates",
		"Third-party webhook migrations",
		"Finalize design system",
		"Create component library",
	}

	expectedOngoingTasks := []string{
		"Code review for feature-auth branch",
		"User research interviews",
		"Update client applications",
		"Update brand guidelines",
	}

	for _, expectedTask := range expectedTodoTasks {
		if taskTypeCounts[expectedTask] > 0 {
			todoCount++
		}
	}

	for _, expectedTask := range expectedOngoingTasks {
		if taskTypeCounts[expectedTask] > 0 {
			ongoingCount++
		}
	}

	// Verify we found both types of tasks
	if todoCount == 0 {
		t.Error("Expected to find some todo tasks [ ]")
	}
	if ongoingCount == 0 {
		t.Error("Expected to find some ongoing tasks [/]")
	}

	t.Logf("Found %d todo tasks and %d ongoing tasks", todoCount, ongoingCount)
}
