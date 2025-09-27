package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

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

			targetDate, err := parseDate(date)
			if err != nil {
				return fmt.Errorf("invalid date format: %w", err)
			}

			if outputFormat == "text" {
				fmt.Printf("Gathering activities for %s...\n", targetDate.Format("2006-01-02"))
			}

			// Initialize cache
			summaryCache, err := cache.NewCache()
			if err != nil {
				return fmt.Errorf("failed to initialize cache: %w", err)
			}

			// Check cache first for historical dates
			if summaryCache.ShouldCache(targetDate) {
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
			summary, err := aggregator.GetSummaryWithVerbose(ctx, targetDate, showVerbose)
			if err != nil {
				return fmt.Errorf("failed to get activity summary: %w", err)
			}

			if showVerbose {
				fmt.Printf("\n📊 Retrieved %d total activities\n\n", len(summary.Activities))
			}

			// Cache the summary if it's for a historical date
			if summaryCache.ShouldCache(targetDate) {
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

	cmd.Flags().StringVarP(&date, "date", "d", "yesterday", "Date to get summary for (yesterday, today, or YYYY-MM-DD)")
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
