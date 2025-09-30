# Daily CLI

A Go CLI tool that aggregates work activities from multiple providers (GitHub, JIRA, Obsidian) and provides daily summaries and todo lists.

## Features

- **Multi-provider support**: GitHub, JIRA, and Obsidian integration
- **Daily summaries**: Get activities for specific dates or date ranges
- **Todo management**: View pending PRs, reviews, and assigned tickets
- **Flexible filtering**: Use provider-specific filters to focus on relevant content
- **Multiple output formats**: TUI (default), text, compact text, and JSON
- **Secure configuration**: Store credentials safely in local config files

## Installation

### Homebrew (Recommended)

```bash
brew tap djiit/tap
brew install daily
```

### GitHub Releases

Download the latest release for your platform from the [releases page](https://github.com/Djiit/daily/releases).

### From Source

```bash
git clone <repository-url>
cd daily
go build -o daily .
```

### Using Go Install

```bash
go install github.com/djiit/daily@latest
```

## Quick Start

1. **Initialize configuration**:
   ```bash
   ./daily config show
   ```

2. **Configure providers** (edit the config file at `~/.config/daily/config.json`):
   ```json
   {
     "github": {
       "username": "your-github-username",
       "token": "ghp_your-token-here",
       "enabled": true,
       "filter": "org:your-org"
     },
     "jira": {
       "email": "your-email@company.com",
       "token": "your-jira-api-token",
       "url": "https://your-company.atlassian.net",
       "enabled": true,
       "filter": "project in (PROJ1, PROJ2)"
     },
     "obsidian": {
       "url": "/path/to/your/obsidian/vault",
       "enabled": true
     }
   }
   ```

3. **Get today's summary** (interactive TUI):
   ```bash
   ./daily sum
   ```

4. **View pending work** (interactive TUI):
   ```bash
   ./daily todo
   ```

## Commands

### `sum` - Daily Summary

Get activities for a time range or specific date.

```bash
# Get activities from last day (new default)
./daily sum

# Get activities from last 2 weeks
./daily sum --since 2w

# Get activities from last 3 hours
./daily sum --since 3h

# Get specific date (legacy date-based query)
./daily sum -d yesterday
./daily sum -d today
./daily sum -d 2024-01-15

# Text output format
./daily sum -o text

# Verbose output (shows provider status, text mode only)
./daily sum -v

# Compact text output
./daily sum -c

# JSON output
./daily sum -o json
```

**Time Range Formats:**
- `1h`, `2h`, etc. - Hours
- `1d`, `2d`, etc. - Days
- `1w`, `2w`, etc. - Weeks
- `1m`, `2m`, etc. - Months

**Note:** Cannot use both `--since` and `--date` flags together.

### `todo` - Todo Management

View pending work items across all providers.

```bash
# Get pending items (default: 2 weeks lookback for Confluence mentions)
./daily todo

# Limit Confluence mentions to last week
./daily todo --since 1w

# Only recent Confluence mentions from last day
./daily todo --since 1d

# Text output format
./daily todo -o text

# Verbose output (shows provider status, text mode only)
./daily todo -v

# JSON output
./daily todo -o json
```

The todo command displays:
- **Open PRs**: Pull requests created by you that are still open
- **Pending Reviews**: Pull requests where you are requested as a reviewer
- **Assigned JIRA Tickets**: JIRA tickets assigned to you that are not done/closed/resolved
- **Confluence Mentions**: Confluence pages where you have been mentioned (controlled by `--since` flag, default: 2w)

### `config` - Configuration Management

Manage your configuration settings.

```bash
# View current configuration
./daily config show

# Show config file location
./daily config path
```

## Provider Configuration

### GitHub

Required fields:
- `username`: Your GitHub username
- `token`: GitHub Personal Access Token
- `enabled`: Set to `true` to enable the provider

Optional fields:
- `filter`: GitHub search filter (see [GitHub Search Filters](#github-search-filters))

#### GitHub Personal Access Token

1. Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Generate a new token with these scopes:
   - `repo` (for private repositories)
   - `read:user` (for user information)
   - `read:org` (for organization information)

### JIRA

Required fields:
- `email`: Your JIRA account email
- `token`: JIRA API Token
- `url`: Your JIRA instance URL (e.g., `https://company.atlassian.net`)
- `enabled`: Set to `true` to enable the provider

Optional fields:
- `filter`: JQL (JIRA Query Language) filter (see [JIRA Filters (JQL)](#jira-filters-jql))

#### JIRA API Token

1. Go to [Atlassian Account Settings](https://id.atlassian.com/manage-profile/security/api-tokens)
2. Create a new API token
3. Copy the token to your configuration

### Obsidian

Required fields:
- `url`: Path to your Obsidian vault directory
- `enabled`: Set to `true` to enable the provider

## Filtering

Filtering allows you to focus on specific repositories, projects, or content that's relevant to you.

### GitHub Search Filters

Use GitHub's search syntax to filter results:

```json
{
  "github": {
    "filter": "org:mycompany repo:mycompany/important-repo language:go"
  }
}
```

**Common GitHub filters:**
- `org:myorg` - Only repositories from specific organization
- `repo:owner/name` - Only specific repository
- `user:username` - Only repositories owned by specific user
- `language:go` - Only repositories with specific primary language
- `is:public` - Only public repositories
- `is:private` - Only private repositories
- `is:fork` - Only forked repositories
- `archived:false` - Exclude archived repositories

**Combining filters:**
```json
"filter": "org:mycompany -repo:mycompany/archived-repo language:typescript"
```

### JIRA Filters (JQL)

Use JQL (JIRA Query Language) to filter tickets:

```json
{
  "jira": {
    "filter": "project in (WEB, API) AND labels = backend"
  }
}
```

**Common JIRA filters:**
- `project = PROJ` - Only specific project
- `project in (PROJ1, PROJ2)` - Multiple projects
- `labels = urgent` - Only tickets with specific label
- `priority = High` - Only high priority tickets
- `status in ("In Progress", "Code Review")` - Specific statuses
- `component = Backend` - Only specific component
- `fixVersion = "1.2.0"` - Specific fix version
- `created >= -7d` - Created in last 7 days
- `assignee = currentUser()` - Assigned to current user (automatically included)

**Combining JQL conditions:**
```json
"filter": "project = WEB AND labels in (urgent, bug) AND status != Done"
```

### Filter Examples

#### Focus on specific team/project:
```json
{
  "github": {
    "filter": "org:mycompany repo:mycompany/frontend repo:mycompany/backend"
  },
  "jira": {
    "filter": "project = FRONTEND AND component in (UI, API)"
  }
}
```

#### Exclude certain repositories/projects:
```json
{
  "github": {
    "filter": "org:mycompany -repo:mycompany/deprecated-app archived:false"
  },
  "jira": {
    "filter": "project != DEPRECATED AND status != Closed"
  }
}
```

#### Language-specific work:
```json
{
  "github": {
    "filter": "language:go language:typescript"
  },
  "jira": {
    "filter": "labels in (backend, api)"
  }
}
```

## Output Formats

### TUI Output (Default)

Interactive Text User Interface for browsing activities and todo items. This is the default output format for both `sum` and `todo` commands.

**Summary TUI** (`./daily sum`):
- **Two-panel layout**: Activity list on left, details on right
- **Markdown rendering**: Rich formatting for activity descriptions
- **Navigation**: Use `↑/↓` or `j/k` to navigate, `g/G` for top/bottom
- **URL opening**: Press `Enter` or `Space` to open URLs in browser

**Todo TUI** (`./daily todo`):
- **Unified list**: All todo items in chronological order
- **Item details**: Full descriptions, URLs, and tags
- **Visual indicators**: Icons for different platforms and item types

### Text Output

Clean, colorized output suitable for terminal viewing:

```bash
./daily sum -o text
./daily todo -o text
```

**TUI Features:**
- **Navigation**: Use `↑/↓` or `j/k` to navigate through items
- **Quick jump**: Use `g` to go to top, `G` to go to bottom
- **Item details**: View full description, URL, and tags for selected item
- **Visual feedback**: Selected item is highlighted with detailed information
- **Scrolling**: Automatically scrolls through large lists
- **Responsive**: Adapts to terminal size

**TUI Controls:**
- `↑/↓` or `j/k` - Navigate up/down
- `g/G` - Jump to top/bottom
- `Enter/Space` - Select item (reserved for future features)
- `q` or `Ctrl+C` - Quit

### Compact Text Output

Minimal output with less spacing, useful for quick overviews:

```bash
./daily sum -c
./daily todo -c
```

### JSON Output

Structured JSON output for programmatic use:

```bash
./daily sum -o json
./daily todo -o json
```

## Configuration File

The configuration file is stored at `~/.config/daily/config.json` and is automatically created with default values on first run.

### Example Configuration

```json
{
  "github": {
    "username": "myusername",
    "token": "ghp_1234567890abcdef",
    "enabled": true,
    "filter": "org:mycompany -repo:mycompany/legacy archived:false"
  },
  "jira": {
    "email": "user@company.com",
    "token": "ATATT3xFfGF09WmR...",
    "url": "https://company.atlassian.net",
    "enabled": true,
    "filter": "project in (WEB, API, MOBILE) AND status != Done"
  },
  "obsidian": {
    "url": "/Users/username/Documents/Obsidian Vault",
    "enabled": false
  }
}
```

## Activity Types

The tool tracks different types of activities:

- **`commit`** - Git commits
- **`pull_request`** - GitHub pull requests
- **`issue`** - GitHub issues
- **`jira_ticket`** - JIRA tickets
- **`note`** - Obsidian notes

## Development

### Building

```bash
go build -o daily .
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Test specific package
go test ./internal/provider/github
```

### Code Quality

```bash
# Format code
go fmt ./...

# Run static analysis
go vet ./...

# Clean up dependencies
go mod tidy
```

## Architecture

### Core Components

- **`main.go`**: Entry point with Cobra CLI setup
- **`cmd/`**: Command implementations (sum, config, todo)
- **`internal/activity/`**: Core activity and summary data structures
- **`internal/provider/`**: Provider interface and aggregator
- **`internal/config/`**: Configuration management
- **`internal/output/`**: Output formatting (text and JSON)
- **`internal/tui/`**: TUI components using Bubble Tea framework

### Provider System

Each provider implements the `Provider` interface:

```go
type Provider interface {
    Name() string
    GetActivities(ctx context.Context, from, to time.Time) ([]Activity, error)
    IsConfigured() bool
}
```

Providers are located in `internal/provider/{github,jira,obsidian}/` and use a common `Config` struct for authentication and settings.

## Troubleshooting

### Common Issues

**GitHub API rate limiting:**
- Ensure you're using a personal access token
- Consider reducing the frequency of requests

**JIRA authentication errors:**
- Verify your email and API token are correct
- Check that your JIRA URL is properly formatted (with https://)

**Configuration not found:**
- Run `./daily config show` to create the default configuration
- Check the config path with `./daily config path`

**No activities found:**
- Verify your providers are enabled and properly configured
- Check that your filters aren't too restrictive
- Use verbose mode (`-v`) to see provider status

### Verbose Mode

Use the `-v` flag to see detailed information about provider status:

```bash
./daily sum -v
./daily todo -v
```

This will show:
- Which providers are enabled/disabled
- Authentication status
- Number of activities returned by each provider
- Any errors encountered

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
