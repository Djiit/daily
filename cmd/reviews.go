package cmd

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"daily/internal/config"
	"daily/internal/output"
	"daily/internal/provider/github"
)

func ReviewsCmd() *cobra.Command {
	var verbose bool
	var outputFormat string
	var skipDetails bool

	cmd := &cobra.Command{
		Use:   "reviews",
		Short: "Get PRs awaiting review from you and your teams",
		Long:  "Display pull requests that are awaiting review from you or your teams, including CI status and PR details. Uses concurrent processing with rate limiting for optimal performance. Use --verbose to see detailed progress.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate output format
			if outputFormat != "text" && outputFormat != "json" && outputFormat != "tui" {
				return fmt.Errorf("invalid output format: %s (must be 'text', 'json', or 'tui')", outputFormat)
			}

			if outputFormat == "text" {
				fmt.Println("Gathering review requests...")
			}

			// Load configuration
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			ctx := context.Background()
			showVerbose := verbose && outputFormat == "text"

			var reviewItems output.ReviewItems

			// Get GitHub review requests
			if cfg.GitHub.Enabled {
				if showVerbose {
					fmt.Println("✓ GitHub provider enabled")
				}
				githubProvider := github.NewProvider(cfg.GitHub)
				if githubProvider.IsConfigured() {
					githubReviews, err := getGitHubReviews(ctx, githubProvider, showVerbose, skipDetails)
					if err != nil {
						if showVerbose {
							fmt.Printf("❌ GitHub reviews failed: %v\n", err)
						}
					} else {
						reviewItems.GitHub = githubReviews
						if showVerbose {
							totalPRs := len(githubReviews.UserRequests) + len(githubReviews.TeamRequests)
							fmt.Printf("✅ GitHub returned %d PRs awaiting review\n", totalPRs)
						}
					}
				} else if showVerbose {
					fmt.Println("⚠️  GitHub provider not configured")
				}
			} else if showVerbose {
				fmt.Println("✗ GitHub provider disabled")
			}

			if showVerbose {
				fmt.Println()
			}

			// Format and display results
			switch outputFormat {
			case "json":
				formatter := output.NewFormatter()
				result := formatter.FormatReviewJSON(reviewItems)
				fmt.Print(result)
			case "tui":
				formatter := output.NewFormatter()
				return formatter.FormatReviewTUI(reviewItems)
			case "text":
				formatter := output.NewFormatter()
				result := formatter.FormatReview(reviewItems)
				fmt.Print(result)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output for debugging (text mode only)")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "tui", "Output format: 'tui', 'text', or 'json'")
	cmd.Flags().BoolVar(&skipDetails, "skip-details", false, "Skip fetching CI status and PR details for faster execution")

	return cmd
}

func getGitHubReviews(ctx context.Context, provider *github.Provider, verbose bool, skipDetails bool) (output.GitHubReviews, error) {
	var reviews output.GitHubReviews

	// Get user review requests
	userRequests, err := provider.GetUserReviewRequests(ctx)
	if err != nil {
		return reviews, fmt.Errorf("failed to get user review requests: %w", err)
	}

	// Get team review requests
	teamRequests, err := provider.GetTeamReviewRequests(ctx)
	if err != nil {
		return reviews, fmt.Errorf("failed to get team review requests: %w", err)
	}

	// Convert and enrich with CI status and PR details
	reviews.UserRequests = make([]output.ReviewItem, len(userRequests))
	if verbose && !skipDetails && len(userRequests) > 0 {
		fmt.Printf("🔄 Fetching additional details for %d user review requests (concurrent)...\n", len(userRequests))
	}

	if skipDetails {
		// Fast path: just convert without enrichment
		for i, pr := range userRequests {
			reviews.UserRequests[i] = output.ReviewItem{
				TodoItem: output.TodoItem{
					ID:          pr.ID,
					Title:       pr.Title,
					Description: pr.Description,
					URL:         pr.URL,
					UpdatedAt:   pr.UpdatedAt,
					Tags:        pr.Tags,
				},
			}
		}
	} else {
		// Concurrent enrichment
		reviews.UserRequests = enrichPRsConcurrently(ctx, provider, userRequests, "user", verbose)
	}

	reviews.TeamRequests = make([]output.ReviewItem, len(teamRequests))
	if verbose && !skipDetails && len(teamRequests) > 0 {
		fmt.Printf("🔄 Fetching additional details for %d team review requests (concurrent)...\n", len(teamRequests))
	}

	if skipDetails {
		// Fast path: just convert without enrichment
		for i, pr := range teamRequests {
			reviews.TeamRequests[i] = output.ReviewItem{
				TodoItem: output.TodoItem{
					ID:          pr.ID,
					Title:       pr.Title,
					Description: pr.Description,
					URL:         pr.URL,
					UpdatedAt:   pr.UpdatedAt,
					Tags:        pr.Tags,
				},
			}
		}
	} else {
		// Concurrent enrichment
		reviews.TeamRequests = enrichPRsConcurrently(ctx, provider, teamRequests, "team", verbose)
	}

	if verbose && !skipDetails {
		totalPRs := len(userRequests) + len(teamRequests)
		if totalPRs > 0 {
			fmt.Printf("✅ Completed fetching additional details for all %d PRs\n", totalPRs)
		}
	}

	return reviews, nil
}

func enrichPRWithDetails(ctx context.Context, provider *github.Provider, pr github.TodoItem) (output.ReviewItem, error) {
	reviewItem := output.ReviewItem{
		TodoItem: output.TodoItem{
			ID:          pr.ID,
			Title:       pr.Title,
			Description: pr.Description,
			URL:         pr.URL,
			UpdatedAt:   pr.UpdatedAt,
			Tags:        pr.Tags,
		},
	}

	// Get CI status
	ciStatus, err := provider.GetPRCIStatus(ctx, pr.Repository, pr.Number)
	if err == nil {
		// Convert github.CIStatus to output.CIStatus
		reviewItem.CIStatus = output.CIStatus{
			State:      ciStatus.State,
			TotalCount: ciStatus.TotalCount,
			Checks:     convertCheckRuns(ciStatus.Checks),
		}
	}

	// Get PR details (additions, deletions, changed files)
	prDetails, err2 := provider.GetPRDetails(ctx, pr.Repository, pr.Number)
	if err2 == nil {
		// Convert github.PRDetails to output.PRDetails
		reviewItem.PRDetails = output.PRDetails{
			Additions:    prDetails.Additions,
			Deletions:    prDetails.Deletions,
			ChangedFiles: prDetails.ChangedFiles,
		}
	}

	// Return the first error encountered, if any
	if err != nil {
		return reviewItem, err
	}
	if err2 != nil {
		return reviewItem, err2
	}

	return reviewItem, nil
}

func convertCheckRuns(githubChecks []github.CheckRun) []output.CheckRun {
	checks := make([]output.CheckRun, len(githubChecks))
	for i, check := range githubChecks {
		checks[i] = output.CheckRun{
			Name:       check.Name,
			Status:     check.Status,
			Conclusion: check.Conclusion,
			URL:        check.URL,
		}
	}
	return checks
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// enrichPRsConcurrently processes PRs concurrently with rate limiting
func enrichPRsConcurrently(ctx context.Context, provider *github.Provider, prs []github.TodoItem, requestType string, verbose bool) []output.ReviewItem {
	if len(prs) == 0 {
		return make([]output.ReviewItem, 0)
	}

	// Create a rate limiter: max 5 concurrent requests, 1 request every 200ms
	const maxWorkers = 5
	const rateLimitDelay = 200 * time.Millisecond

	// Channels for work distribution and results
	type prJob struct {
		index int
		pr    github.TodoItem
	}

	type prResult struct {
		index      int
		reviewItem output.ReviewItem
		err        error
	}

	jobs := make(chan prJob, len(prs))
	results := make(chan prResult, len(prs))

	// Rate limiting ticker
	ticker := time.NewTicker(rateLimitDelay)
	defer ticker.Stop()

	// Start worker goroutines
	var wg sync.WaitGroup
	for w := 0; w < min(maxWorkers, len(prs)); w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for job := range jobs {
				// Wait for rate limit
				<-ticker.C

				if verbose {
					fmt.Printf("  ⏳ [%d/%d] Worker %d processing PR #%d: %s...\n",
						job.index+1, len(prs), workerID+1, job.pr.Number,
						job.pr.Title[:min(50, len(job.pr.Title))])
				}

				reviewItem, err := enrichPRWithDetails(ctx, provider, job.pr)
				if err != nil {
					if verbose {
						fmt.Printf("    ⚠️  Worker %d failed to enrich PR %s: %v\n", workerID+1, job.pr.ID, err)
					}
					// Create fallback item
					reviewItem = output.ReviewItem{
						TodoItem: output.TodoItem{
							ID:          job.pr.ID,
							Title:       job.pr.Title,
							Description: job.pr.Description,
							URL:         job.pr.URL,
							UpdatedAt:   job.pr.UpdatedAt,
							Tags:        job.pr.Tags,
						},
					}
				}

				results <- prResult{
					index:      job.index,
					reviewItem: reviewItem,
					err:        err,
				}
			}
		}(w)
	}

	// Send jobs
	go func() {
		defer close(jobs)
		for i, pr := range prs {
			jobs <- prJob{index: i, pr: pr}
		}
	}()

	// Collect results
	reviewItems := make([]output.ReviewItem, len(prs))
	successCount := 0

	for i := 0; i < len(prs); i++ {
		result := <-results
		reviewItems[result.index] = result.reviewItem
		if result.err == nil {
			successCount++
		}
	}

	// Wait for all workers to complete
	wg.Wait()

	if verbose {
		fmt.Printf("  ✅ Completed %s requests: %d successful, %d failed\n",
			requestType, successCount, len(prs)-successCount)
	}

	return reviewItems
}
