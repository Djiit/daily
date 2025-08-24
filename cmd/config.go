package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"daily/internal/config"
)

func ConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration settings",
		Long:  "View and manage configuration settings for daily CLI providers.",
	}

	cmd.AddCommand(configShowCmd())
	cmd.AddCommand(configPathCmd())

	return cmd
}

func configShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Long:  "Display the current configuration settings for all providers.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			fmt.Println("Current Configuration:")
			fmt.Printf("\nGitHub:")
			fmt.Printf("\n  Enabled: %t", cfg.GitHub.Enabled)
			fmt.Printf("\n  Username: %s", cfg.GitHub.Username)
			fmt.Printf("\n  Token: %s", maskToken(cfg.GitHub.Token))

			fmt.Printf("\n\nJIRA:")
			fmt.Printf("\n  Enabled: %t", cfg.JIRA.Enabled)
			fmt.Printf("\n  URL: %s", cfg.JIRA.URL)
			fmt.Printf("\n  Email: %s", cfg.JIRA.Email)
			fmt.Printf("\n  Token: %s", maskToken(cfg.JIRA.Token))

			fmt.Printf("\n\nObsidian:")
			fmt.Printf("\n  Enabled: %t", cfg.Obsidian.Enabled)
			fmt.Printf("\n  Vault Path: %s", cfg.Obsidian.URL)
			fmt.Println()

			return nil
		},
	}
}

func configPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Show configuration file path",
		Long:  "Display the path to the configuration file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := config.GetConfigPath()
			if err != nil {
				return fmt.Errorf("failed to get config path: %w", err)
			}

			fmt.Printf("Configuration file: %s\n", path)
			return nil
		},
	}
}

func maskToken(token string) string {
	if token == "" {
		return "(not set)"
	}
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}
