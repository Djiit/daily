package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"daily/internal/config"
	"daily/internal/output"
	"daily/internal/provider"
	"daily/internal/provider/github"
	"daily/internal/provider/jira"
	"daily/internal/provider/obsidian"
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
			if outputFormat != "text" && outputFormat != "json" {
				return fmt.Errorf("invalid output format: %s (must be 'text' or 'json')", outputFormat)
			}

			targetDate, err := parseDate(date)
			if err != nil {
				return fmt.Errorf("invalid date format: %w", err)
			}

			if outputFormat == "text" {
				fmt.Printf("Gathering activities for %s...\n", targetDate.Format("2006-01-02"))
			}
			
			// Load configuration
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create providers
			aggregator := provider.NewAggregator()
			
			if cfg.GitHub.Enabled {
				if verbose && outputFormat == "text" {
					fmt.Println("✓ GitHub provider enabled")
				}
				aggregator.AddProvider(github.NewProvider(cfg.GitHub))
			} else if verbose && outputFormat == "text" {
				fmt.Println("✗ GitHub provider disabled")
			}
			
			if cfg.JIRA.Enabled {
				if verbose && outputFormat == "text" {
					fmt.Println("✓ JIRA provider enabled")
				}
				aggregator.AddProvider(jira.NewProvider(cfg.JIRA))
			} else if verbose && outputFormat == "text" {
				fmt.Println("✗ JIRA provider disabled")
			}
			
			if cfg.Obsidian.Enabled {
				if verbose && outputFormat == "text" {
					fmt.Println("✓ Obsidian provider enabled")
				}
				aggregator.AddProvider(obsidian.NewProvider(cfg.Obsidian))
			} else if verbose && outputFormat == "text" {
				fmt.Println("✗ Obsidian provider disabled")
			}

			// Get summary
			ctx := context.Background()
			if verbose && outputFormat == "text" {
				fmt.Println()
			}
			summary, err := aggregator.GetSummaryWithVerbose(ctx, targetDate, verbose && outputFormat == "text")
			if err != nil {
				return fmt.Errorf("failed to get activity summary: %w", err)
			}
			
			if verbose && outputFormat == "text" {
				fmt.Printf("\n📊 Retrieved %d total activities\n\n", len(summary.Activities))
			}

			// Format and display results
			formatter := output.NewFormatter()
			var result string
			
			switch outputFormat {
			case "json":
				result = formatter.FormatJSON(summary)
			case "text":
				if compact {
					result = formatter.FormatCompactSummary(summary)
				} else {
					result = formatter.FormatSummary(summary)
				}
			}
			
			fmt.Print(result)
			
			return nil
		},
	}

	cmd.Flags().StringVarP(&date, "date", "d", "yesterday", "Date to get summary for (yesterday, today, or YYYY-MM-DD)")
	cmd.Flags().BoolVarP(&compact, "compact", "c", false, "Use compact output format (text mode only)")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output for debugging (text mode only)")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Output format: 'text' or 'json'")

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