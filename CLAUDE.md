# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Daily is a Go CLI tool that aggregates work activities from multiple providers (GitHub, JIRA, Obsidian, Confluence) and provides summaries. The application uses a provider-based architecture where each data source implements the `Provider` interface.

## Development Commands

### Building and Running
- `go build -o daily .` - Build the binary
- `go run main.go` - Run directly with Go
- `./daily --help` - Show CLI help
- `./daily sum --help` - Show sum command help
- `./daily config --help` - Show config command help
- `./daily todo --help` - Show todo command help

### Testing
- `go test ./...` - Run all tests
- `go test -v ./...` - Run tests with verbose output
- `go test ./internal/provider/...` - Test specific package tree

### Code Quality
- `go fmt ./...` - Format code
- `go vet ./...` - Run static analysis
- `go mod tidy` - Clean up dependencies

## Architecture

### Core Components
- **main.go**: Entry point with Cobra CLI setup using charmbracelet/fang
- **cmd/**: Command implementations (sum, config, todo)
- **internal/activity/**: Core activity and summary data structures
- **internal/provider/**: Provider interface and aggregator
- **internal/config/**: Configuration management (JSON-based, stored in ~/.config/daily/)
- **internal/output/**: Output formatting (text and JSON)
- **internal/tui/**: TUI (Terminal User Interface) components using Bubble Tea

### Provider System
Each provider implements the `Provider` interface:
```go
type Provider interface {
    Name() string
    GetActivities(ctx context.Context, from, to time.Time) ([]Activity, error)
    IsConfigured() bool
}
```

Providers are located in `internal/provider/{github,jira,obsidian,confluence}/` and use a common `Config` struct for authentication and settings.

### Activity Types
- `commit` - Git commits
- `pull_request` - GitHub PRs
- `issue` - GitHub issues
- `jira_ticket` - JIRA tickets
- `note` - Obsidian notes
- `confluence_contribution` - Confluence page contributions

## Configuration

Configuration is stored as JSON in `~/.config/daily/config.json`. Use:
- `./daily config show` - View current config
- `./daily config path` - Show config file location

Each provider has `enabled`, `username`, `email`, `token`, and `url` fields as needed.

## Key CLI Usage Patterns

### Daily Summary
- `./daily sum` - Get yesterday's summary (default TUI output)
- `./daily sum -d today` - Get today's summary
- `./daily sum -d 2024-01-15` - Get specific date summary
- `./daily sum -v` - Verbose output showing provider status (text mode only)
- `./daily sum -c` - Compact text output
- `./daily sum -o json` - JSON output format
- `./daily sum -o text` - Text output format
- `./daily sum -o tui` - TUI output format (default)

### Todo Management
- `./daily todo` - Get pending work items (default TUI output)
- `./daily todo -v` - Verbose output showing provider status (text mode only)
- `./daily todo -o json` - JSON output format
- `./daily todo -o text` - Text output format
- `./daily todo -o tui` - TUI output format (default)

The todo command displays:
- **Open PRs**: Pull requests created by you that are still open
- **Pending Reviews**: Pull requests where you are requested as a reviewer
- **Assigned JIRA Tickets**: JIRA tickets assigned to you that are not done/closed/resolved
- **Confluence Mentions**: Confluence pages where you have been mentioned