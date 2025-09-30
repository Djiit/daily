package cmd

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"daily/internal/activity"
	"daily/internal/cache"
	"daily/internal/config"
	"daily/internal/output"
	"daily/internal/provider"
	"daily/internal/provider/confluence"
	"daily/internal/provider/github"
	"daily/internal/provider/jira"
	"daily/internal/provider/obsidian"
	"daily/internal/tui"
)

func SumCmd() *cobra.Command {
	var date string
	var since string
	var compact bool
	var verbose bool
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "sum",
		Short: "Get a summary of your daily work activities",
		Long:  "Gather activity data from JIRA, GitHub, and Obsidian to provide a comprehensive summary of your work for the specified date.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate output format
			if outputFormat != "text" && outputFormat != "json" && outputFormat != "tui" {
				return fmt.Errorf("invalid output format: %s (must be 'text', 'json', or 'tui')", outputFormat)
			}

			// Handle --since and --date mutual exclusivity
			if since != "" && date != "" {
				return fmt.Errorf("cannot use both --since and --date flags")
			}

			// Default to --since 1d if neither flag is provided
			if since == "" && date == "" {
				since = "1d"
			}

			// Determine if we're using since-based or date-based querying
			var usingSince bool
			var fromTime, toTime time.Time
			var targetDate time.Time

			if since != "" {
				usingSince = true
				var err error
				fromTime, err = parseSinceDuration(since)
				if err != nil {
					return fmt.Errorf("invalid since format: %w", err)
				}
				toTime = time.Now()
				targetDate = fromTime // Use from time as the summary date

				if outputFormat == "text" {
					fmt.Printf("Gathering activities since %s (%s to now)...\n", since, fromTime.Format("2006-01-02 15:04"))
				}
			} else {
				usingSince = false
				var err error
				targetDate, err = parseDate(date)
				if err != nil {
					return fmt.Errorf("invalid date format: %w", err)
				}

				if outputFormat == "text" {
					fmt.Printf("Gathering activities for %s...\n", targetDate.Format("2006-01-02"))
				}
			}

			// Initialize cache
			summaryCache, err := cache.NewCache()
			if err != nil {
				return fmt.Errorf("failed to initialize cache: %w", err)
			}

			// Check cache first for historical dates (only when using date-based queries)
			if !usingSince && summaryCache.ShouldCache(targetDate) {
				if cachedSummary, err := summaryCache.Get(targetDate); err != nil {
					if outputFormat == "text" && verbose {
						fmt.Printf("Cache read error (proceeding with fresh data): %v\n", err)
					}
				} else if cachedSummary != nil {
					if outputFormat == "text" && verbose {
						fmt.Printf("📋 Using cached summary for %s\n\n", targetDate.Format("2006-01-02"))
					}
					// Format and display cached results
					switch outputFormat {
					case "tui":
						err := tui.RunTUI(cachedSummary)
						if err != nil {
							// Fallback to text output if TUI fails
							formatter := output.NewFormatter()
							result := formatter.FormatSummary(cachedSummary)
							fmt.Print(result)
						}
						return nil
					case "json":
						formatter := output.NewFormatter()
						result := formatter.FormatJSON(cachedSummary)
						fmt.Print(result)
					case "text":
						formatter := output.NewFormatter()
						var result string
						if compact {
							result = formatter.FormatCompactSummary(cachedSummary)
						} else {
							result = formatter.FormatSummary(cachedSummary)
						}
						fmt.Print(result)
					}
					return nil
				}
			}

			// Load configuration
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create providers
			aggregator := provider.NewAggregator()

			showVerbose := verbose && outputFormat == "text"

			if cfg.GitHub.Enabled {
				if showVerbose {
					fmt.Println("✓ GitHub provider enabled")
				}
				aggregator.AddProvider(github.NewProvider(cfg.GitHub))
			} else if showVerbose {
				fmt.Println("✗ GitHub provider disabled")
			}

			if cfg.JIRA.Enabled {
				if showVerbose {
					fmt.Println("✓ JIRA provider enabled")
				}
				aggregator.AddProvider(jira.NewProvider(cfg.JIRA))
			} else if showVerbose {
				fmt.Println("✗ JIRA provider disabled")
			}

			if cfg.Obsidian.Enabled {
				if showVerbose {
					fmt.Println("✓ Obsidian provider enabled")
				}
				aggregator.AddProvider(obsidian.NewProvider(cfg.Obsidian))
			} else if showVerbose {
				fmt.Println("✗ Obsidian provider disabled")
			}

			if cfg.Confluence.Enabled {
				if showVerbose {
					fmt.Println("✓ Confluence provider enabled")
				}
				aggregator.AddProvider(confluence.NewProvider(cfg.Confluence))
			} else if showVerbose {
				fmt.Println("✗ Confluence provider disabled")
			}

			// Get summary
			ctx := context.Background()
			if showVerbose {
				fmt.Println()
			}

			var summary *activity.Summary

			if usingSince {
				// Use time range method for --since
				summary, err = aggregator.GetSummaryByTimeRange(ctx, fromTime, toTime, showVerbose)
				if err != nil {
					return fmt.Errorf("failed to get activity summary: %w", err)
				}
			} else {
				// Use date-based method for --date
				summary, err = aggregator.GetSummaryWithVerbose(ctx, targetDate, showVerbose)
				if err != nil {
					return fmt.Errorf("failed to get activity summary: %w", err)
				}
			}

			if showVerbose {
				fmt.Printf("\n📊 Retrieved %d total activities\n\n", len(summary.Activities))
			}

			// Cache the summary if it's for a historical date (only for date-based queries)
			if !usingSince && summaryCache.ShouldCache(targetDate) {
				if err := summaryCache.Set(targetDate, summary); err != nil {
					if outputFormat == "text" && verbose {
						fmt.Printf("Warning: Failed to cache summary: %v\n", err)
					}
				} else if outputFormat == "text" && verbose {
					fmt.Printf("💾 Cached summary for future use\n\n")
				}
			}

			// Format and display results
			switch outputFormat {
			case "tui":
				err := tui.RunTUI(summary)
				if err != nil {
					// Fallback to text output if TUI fails
					formatter := output.NewFormatter()
					result := formatter.FormatSummary(summary)
					fmt.Print(result)
				}
				return nil
			case "json":
				formatter := output.NewFormatter()
				result := formatter.FormatJSON(summary)
				fmt.Print(result)
			case "text":
				formatter := output.NewFormatter()
				var result string
				if compact {
					result = formatter.FormatCompactSummary(summary)
				} else {
					result = formatter.FormatSummary(summary)
				}
				fmt.Print(result)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&date, "date", "d", "", "Date to get summary for (yesterday, today, or YYYY-MM-DD)")
	cmd.Flags().StringVarP(&since, "since", "s", "", "Time range to look back (e.g., 1h, 1d, 2w, 1m). Default: 1d")
	cmd.Flags().BoolVarP(&compact, "compact", "c", false, "Use compact output format (text mode only)")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output for debugging (text mode only)")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "tui", "Output format: 'tui', 'text', or 'json'")

	return cmd
}

func parseDate(dateStr string) (time.Time, error) {
	now := time.Now()

	switch dateStr {
	case "today":
		return now, nil
	case "yesterday":
		return now.AddDate(0, 0, -1), nil
	default:
		return time.Parse("2006-01-02", dateStr)
	}
}

// parseSinceDuration parses a "since" duration string (e.g., "1d", "2w", "3h", "1m")
// and returns the "from" time (now - duration)
func parseSinceDuration(since string) (time.Time, error) {
	// Match format: number + unit (h/d/w/m)
	re := regexp.MustCompile(`^(\d+)([hdwm])$`)
	matches := re.FindStringSubmatch(since)

	if matches == nil {
		return time.Time{}, fmt.Errorf("invalid since format: %s (expected format: 1h, 1d, 1w, or 1m)", since)
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid since value: %s", matches[1])
	}

	unit := matches[2]
	now := time.Now()

	switch unit {
	case "h":
		return now.Add(-time.Duration(value) * time.Hour), nil
	case "d":
		return now.AddDate(0, 0, -value), nil
	case "w":
		return now.AddDate(0, 0, -value*7), nil
	case "m":
		return now.AddDate(0, -value, 0), nil
	default:
		return time.Time{}, fmt.Errorf("invalid since unit: %s (expected h, d, w, or m)", unit)
	}
}
