package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"daily/internal/config"
	"daily/internal/output"
	"daily/internal/provider/github"
	"daily/internal/provider/jira"
	"daily/internal/provider/obsidian"
)

func TodoCmd() *cobra.Command {
	var verbose bool
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "todo",
		Short: "Get a list of pending work items",
		Long:  "Display open pull requests, pending reviews, and assigned JIRA tickets that need attention.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate output format
			if outputFormat != "text" && outputFormat != "json" && outputFormat != "tui" {
				return fmt.Errorf("invalid output format: %s (must be 'text', 'json', or 'tui')", outputFormat)
			}

			if outputFormat == "text" {
				fmt.Println("Gathering pending work items...")
			}

			// Load configuration
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			ctx := context.Background()
			showVerbose := verbose && outputFormat == "text"

			var todoItems output.TodoItems

			// Get GitHub todos
			if cfg.GitHub.Enabled {
				if showVerbose {
					fmt.Println("✓ GitHub provider enabled")
				}
				githubProvider := github.NewProvider(cfg.GitHub)
				if githubProvider.IsConfigured() {
					githubTodos, err := getGitHubTodos(ctx, githubProvider)
					if err != nil {
						if showVerbose {
							fmt.Printf("❌ GitHub todos failed: %v\n", err)
						}
					} else {
						todoItems.GitHub = githubTodos
						if showVerbose {
							fmt.Printf("✅ GitHub returned %d open PRs and %d pending reviews\n",
								len(githubTodos.OpenPRs), len(githubTodos.PendingReviews))
						}
					}
				} else if showVerbose {
					fmt.Println("⚠️  GitHub provider not configured")
				}
			} else if showVerbose {
				fmt.Println("✗ GitHub provider disabled")
			}

			// Get JIRA todos
			if cfg.JIRA.Enabled {
				if showVerbose {
					fmt.Println("✓ JIRA provider enabled")
				}
				jiraProvider := jira.NewProvider(cfg.JIRA)
				if jiraProvider.IsConfigured() {
					jiraTodos, err := getJIRATodos(ctx, jiraProvider)
					if err != nil {
						if showVerbose {
							fmt.Printf("❌ JIRA todos failed: %v\n", err)
						}
					} else {
						todoItems.JIRA = jiraTodos
						if showVerbose {
							fmt.Printf("✅ JIRA returned %d assigned tickets\n", len(jiraTodos.AssignedTickets))
						}
					}
				} else if showVerbose {
					fmt.Println("⚠️  JIRA provider not configured")
				}
			} else if showVerbose {
				fmt.Println("✗ JIRA provider disabled")
			}

			// Get Obsidian todos
			if cfg.Obsidian.Enabled {
				if showVerbose {
					fmt.Println("✓ Obsidian provider enabled")
				}
				obsidianProvider := obsidian.NewProvider(cfg.Obsidian)
				if obsidianProvider.IsConfigured() {
					obsidianTodos, err := getObsidianTodos(ctx, obsidianProvider)
					if err != nil {
						if showVerbose {
							fmt.Printf("❌ Obsidian todos failed: %v\n", err)
						}
					} else {
						todoItems.Obsidian = obsidianTodos
						if showVerbose {
							fmt.Printf("✅ Obsidian returned %d tasks\n", len(obsidianTodos.Tasks))
						}
					}
				} else if showVerbose {
					fmt.Println("⚠️  Obsidian provider not configured")
				}
			} else if showVerbose {
				fmt.Println("✗ Obsidian provider disabled")
			}

			if showVerbose {
				fmt.Println()
			}

			// Format and display results
			switch outputFormat {
			case "json":
				formatter := output.NewFormatter()
				result := formatter.FormatTodoJSON(todoItems)
				fmt.Print(result)
			case "tui":
				formatter := output.NewFormatter()
				return formatter.FormatTodoTUI(todoItems)
			case "text":
				formatter := output.NewFormatter()
				result := formatter.FormatTodo(todoItems)
				fmt.Print(result)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output for debugging (text mode only)")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "tui", "Output format: 'tui', 'text', or 'json'")

	return cmd
}

func getGitHubTodos(ctx context.Context, provider *github.Provider) (output.GitHubTodos, error) {
	var todos output.GitHubTodos

	// Get open PRs
	openPRs, err := provider.GetOpenPRs(ctx)
	if err != nil {
		return todos, fmt.Errorf("failed to get open PRs: %w", err)
	}

	// Convert from github.TodoItem to output.TodoItem
	todos.OpenPRs = make([]output.TodoItem, len(openPRs))
	for i, item := range openPRs {
		todos.OpenPRs[i] = output.TodoItem{
			ID:          item.ID,
			Title:       item.Title,
			Description: item.Description,
			URL:         item.URL,
			UpdatedAt:   item.UpdatedAt,
			Tags:        item.Tags,
		}
	}

	// Get pending reviews
	pendingReviews, err := provider.GetPendingReviews(ctx)
	if err != nil {
		return todos, fmt.Errorf("failed to get pending reviews: %w", err)
	}

	// Convert from github.TodoItem to output.TodoItem
	todos.PendingReviews = make([]output.TodoItem, len(pendingReviews))
	for i, item := range pendingReviews {
		todos.PendingReviews[i] = output.TodoItem{
			ID:          item.ID,
			Title:       item.Title,
			Description: item.Description,
			URL:         item.URL,
			UpdatedAt:   item.UpdatedAt,
			Tags:        item.Tags,
		}
	}

	return todos, nil
}

func getJIRATodos(ctx context.Context, provider *jira.Provider) (output.JIRATodos, error) {
	var todos output.JIRATodos

	// Get assigned tickets that are not done
	assignedTickets, err := provider.GetAssignedTickets(ctx)
	if err != nil {
		return todos, fmt.Errorf("failed to get assigned tickets: %w", err)
	}

	// Convert from jira.TodoItem to output.TodoItem
	todos.AssignedTickets = make([]output.TodoItem, len(assignedTickets))
	for i, item := range assignedTickets {
		todos.AssignedTickets[i] = output.TodoItem{
			ID:          item.ID,
			Title:       item.Title,
			Description: item.Description,
			URL:         item.URL,
			UpdatedAt:   item.UpdatedAt,
			Tags:        item.Tags,
		}
	}

	return todos, nil
}

func getObsidianTodos(ctx context.Context, provider *obsidian.Provider) (output.ObsidianTodos, error) {
	var todos output.ObsidianTodos

	// Get tasks from Obsidian vault
	tasks, err := provider.GetTasks(ctx)
	if err != nil {
		return todos, fmt.Errorf("failed to get Obsidian tasks: %w", err)
	}

	// Convert from obsidian.TodoItem to output.TodoItem
	todos.Tasks = make([]output.TodoItem, len(tasks))
	for i, item := range tasks {
		todos.Tasks[i] = output.TodoItem{
			ID:          item.ID,
			Title:       item.Title,
			Description: item.Description,
			URL:         item.URL,
			UpdatedAt:   item.UpdatedAt,
			Tags:        item.Tags,
		}
	}

	return todos, nil
}
