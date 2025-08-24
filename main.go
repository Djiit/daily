package main

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"

	"daily/cmd"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "daily",
		Short: "Get a summary of your daily work activities",
		Long:  "Daily CLI gathers your activity data from JIRA, GitHub, and Obsidian to provide a comprehensive summary of your work.",
	}

	rootCmd.AddCommand(cmd.SumCmd())
	rootCmd.AddCommand(cmd.ConfigCmd())

	if err := fang.Execute(context.Background(), rootCmd); err != nil {
		os.Exit(1)
	}
}
