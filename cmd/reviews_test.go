package cmd

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"daily/internal/provider"
	"daily/internal/provider/github"
)

func TestGetGitHubReviews(t *testing.T) {
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
			expectedErrMsg: "failed to get user review requests: GitHub provider not configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := github.NewProvider(tt.config)

			reviews, err := getGitHubReviews(context.Background(), provider, false, false)

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
				if reviews.UserRequests == nil {
					t.Error("Expected non-nil UserRequests slice")
				}
				if reviews.TeamRequests == nil {
					t.Error("Expected non-nil TeamRequests slice")
				}
			}
		})
	}
}

func TestReviewsCmd_Creation(t *testing.T) {
	cmd := ReviewsCmd()

	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}

	if cmd.Use != "reviews" {
		t.Errorf("Expected command use to be 'reviews', got '%s'", cmd.Use)
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

	skipDetailsFlag := cmd.Flags().Lookup("skip-details")
	if skipDetailsFlag == nil {
		t.Error("Expected skip-details flag to be defined")
	}
}

func TestReviewsCmd_FlagValidation(t *testing.T) {
	cmd := ReviewsCmd()

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

func TestEnrichPRWithDetails(t *testing.T) {
	// This tests the structure of enrichPRWithDetails without making real API calls
	pr := github.TodoItem{
		ID:          "test-pr-1",
		Title:       "Test PR Title",
		Description: "Test PR Description",
		URL:         "https://github.com/owner/repo/pull/123",
		Number:      123,
		Repository:  "owner/repo",
		UpdatedAt:   time.Now(),
		Tags:        []string{"bug", "high-priority"},
	}

	// Test with unconfigured provider (should return error)
	config := provider.Config{
		Username: "",
		Token:    "",
		Enabled:  false,
	}
	provider := github.NewProvider(config)

	reviewItem, err := enrichPRWithDetails(context.Background(), provider, pr)

	// Should get an error due to unconfigured provider
	if err == nil {
		t.Error("Expected error for unconfigured provider, got nil")
	}

	// But should still return a valid review item structure
	if reviewItem.TodoItem.ID != pr.ID {
		t.Errorf("Expected ID %s, got %s", pr.ID, reviewItem.TodoItem.ID)
	}
	if reviewItem.TodoItem.Title != pr.Title {
		t.Errorf("Expected Title %s, got %s", pr.Title, reviewItem.TodoItem.Title)
	}
}

func TestConvertCheckRuns(t *testing.T) {
	githubChecks := []github.CheckRun{
		{
			Name:       "CI",
			Status:     "completed",
			Conclusion: "success",
			URL:        "https://github.com/owner/repo/runs/123",
		},
		{
			Name:       "Tests",
			Status:     "in_progress",
			Conclusion: "",
			URL:        "https://github.com/owner/repo/runs/124",
		},
	}

	outputChecks := convertCheckRuns(githubChecks)

	if len(outputChecks) != len(githubChecks) {
		t.Errorf("Expected %d checks, got %d", len(githubChecks), len(outputChecks))
	}

	for i, check := range outputChecks {
		if check.Name != githubChecks[i].Name {
			t.Errorf("Expected check %d name %s, got %s", i, githubChecks[i].Name, check.Name)
		}
		if check.Status != githubChecks[i].Status {
			t.Errorf("Expected check %d status %s, got %s", i, githubChecks[i].Status, check.Status)
		}
		if check.Conclusion != githubChecks[i].Conclusion {
			t.Errorf("Expected check %d conclusion %s, got %s", i, githubChecks[i].Conclusion, check.Conclusion)
		}
		if check.URL != githubChecks[i].URL {
			t.Errorf("Expected check %d URL %s, got %s", i, githubChecks[i].URL, check.URL)
		}
	}
}

func TestMinFunction(t *testing.T) {
	tests := []struct {
		a, b, expected int
	}{
		{5, 3, 3},
		{1, 10, 1},
		{7, 7, 7},
		{0, 5, 0},
		{-1, 2, -1},
	}

	for _, tt := range tests {
		result := min(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("min(%d, %d) = %d, expected %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestEnrichPRsConcurrently_EmptySlice(t *testing.T) {
	config := provider.Config{
		Username: "testuser",
		Token:    "testtoken",
		Enabled:  true,
	}
	provider := github.NewProvider(config)

	emptyPRs := []github.TodoItem{}
	result := enrichPRsConcurrently(context.Background(), provider, emptyPRs, "test", false)

	if len(result) != 0 {
		t.Errorf("Expected empty result for empty input, got %d items", len(result))
	}
}

func TestEnrichPRsConcurrently_ConcurrencyLimits(t *testing.T) {
	config := provider.Config{
		Username: "testuser",
		Token:    "testtoken",
		Enabled:  true,
	}
	provider := github.NewProvider(config)

	// Create a slice with more PRs than max workers
	prs := make([]github.TodoItem, 10)
	for i := 0; i < 10; i++ {
		prs[i] = github.TodoItem{
			ID:          "pr-" + string(rune(i+'1')),
			Title:       "Test PR " + string(rune(i+'1')),
			Description: "Test Description",
			URL:         "https://github.com/owner/repo/pull/" + string(rune(i+'1')),
			Number:      i + 1,
			Repository:  "owner/repo",
			UpdatedAt:   time.Now(),
		}
	}

	// This will fail with unconfigured credentials but should not panic
	// and should return the same number of items as input
	result := enrichPRsConcurrently(context.Background(), provider, prs, "test", false)

	if len(result) != len(prs) {
		t.Errorf("Expected %d results, got %d", len(prs), len(result))
	}

	// Verify that all items have the basic TodoItem structure preserved
	for i, item := range result {
		if item.TodoItem.ID != prs[i].ID {
			t.Errorf("Item %d: expected ID %s, got %s", i, prs[i].ID, item.TodoItem.ID)
		}
	}
}

func TestEnrichPRsConcurrently_RateLimiting(t *testing.T) {
	config := provider.Config{
		Username: "testuser",
		Token:    "testtoken",
		Enabled:  true,
	}
	provider := github.NewProvider(config)

	// Create a few PRs to test rate limiting timing
	prs := make([]github.TodoItem, 3)
	for i := 0; i < 3; i++ {
		prs[i] = github.TodoItem{
			ID:          "pr-" + string(rune(i+'1')),
			Title:       "Test PR " + string(rune(i+'1')),
			Description: "Test Description",
			URL:         "https://github.com/owner/repo/pull/" + string(rune(i+'1')),
			Number:      i + 1,
			Repository:  "owner/repo",
			UpdatedAt:   time.Now(),
		}
	}

	start := time.Now()
	result := enrichPRsConcurrently(context.Background(), provider, prs, "test", false)
	elapsed := time.Since(start)

	// With 200ms rate limiting, processing 3 items should take at least 400ms
	// (3 ticks minimum, but we give some tolerance for test timing)
	if elapsed < 300*time.Millisecond {
		t.Errorf("Expected rate limiting to cause delay of at least 300ms, got %v", elapsed)
	}

	if len(result) != len(prs) {
		t.Errorf("Expected %d results, got %d", len(prs), len(result))
	}
}

func TestEnrichPRsConcurrently_ContextCancellation(t *testing.T) {
	config := provider.Config{
		Username: "testuser",
		Token:    "testtoken",
		Enabled:  true,
	}
	provider := github.NewProvider(config)

	prs := make([]github.TodoItem, 5)
	for i := 0; i < 5; i++ {
		prs[i] = github.TodoItem{
			ID:          "pr-" + string(rune(i+'1')),
			Title:       "Test PR " + string(rune(i+'1')),
			Description: "Test Description",
			URL:         "https://github.com/owner/repo/pull/" + string(rune(i+'1')),
			Number:      i + 1,
			Repository:  "owner/repo",
			UpdatedAt:   time.Now(),
		}
	}

	// Create a context that will be cancelled quickly
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	result := enrichPRsConcurrently(ctx, provider, prs, "test", false)

	// Should still return results even with cancelled context
	// The workers might complete some items before context cancellation
	if len(result) != len(prs) {
		t.Errorf("Expected %d results even with cancelled context, got %d", len(prs), len(result))
	}
}

func TestGetGitHubReviews_SkipDetails(t *testing.T) {
	config := provider.Config{
		Username: "testuser",
		Token:    "testtoken",
		Enabled:  true,
	}
	provider := github.NewProvider(config)

	// This will fail due to fake credentials, but we test that skip-details path works
	_, err := getGitHubReviews(context.Background(), provider, false, true)

	// Should get error from the initial API calls, not from details fetching
	if err == nil {
		t.Error("Expected error due to fake credentials, got nil")
	}

	// Error should be from user review requests, not from enrichment
	if !strings.Contains(err.Error(), "failed to get user review requests") {
		t.Errorf("Expected error from user review requests, got: %v", err)
	}
}

// Mock helper for testing worker coordination
func TestWorkerPoolBehavior(t *testing.T) {
	// Test that worker pool properly coordinates work distribution
	const numWorkers = 3
	const numJobs = 10

	jobs := make(chan int, numJobs)
	results := make(chan int, numJobs)

	var wg sync.WaitGroup

	// Start workers
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for job := range jobs {
				// Simulate work
				time.Sleep(10 * time.Millisecond)
				results <- job * 2
			}
		}(w)
	}

	// Send jobs
	go func() {
		defer close(jobs)
		for i := 0; i < numJobs; i++ {
			jobs <- i
		}
	}()

	// Collect results
	received := make([]int, 0, numJobs)
	for i := 0; i < numJobs; i++ {
		received = append(received, <-results)
	}

	// Wait for workers to finish
	wg.Wait()

	if len(received) != numJobs {
		t.Errorf("Expected %d results, got %d", numJobs, len(received))
	}

	// Verify all jobs were processed (sum should be doubled)
	expectedSum := 0
	for i := 0; i < numJobs; i++ {
		expectedSum += i * 2
	}

	actualSum := 0
	for _, result := range received {
		actualSum += result
	}

	if actualSum != expectedSum {
		t.Errorf("Expected sum %d, got %d", expectedSum, actualSum)
	}
}
