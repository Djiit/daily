# Implementation Plan: Adding Doctobriefing Features to Daily

## Overview

This document provides a step-by-step implementation plan for integrating three key features from `doctobriefing` into the `daily` CLI tool:

1. **Confluence Integration** (Provider)
2. **Local Hide Cache** (Filtering)
3. **WebUI with WebSocket** (Interface)

## Implementation Order

We'll implement features in order of complexity and value:

**Phase 1**: Confluence Provider + Local Hide Cache (1-2 weeks)
**Phase 2**: WebUI with WebSocket (2-3 weeks)

---

# Phase 1: Core Features

## Feature 1: Confluence Integration

### 🎯 Goal
Add Confluence as a new provider to aggregate mentions and page updates.

### 📋 Prerequisites
- Understanding of existing provider pattern
- Confluence API knowledge
- CQL (Confluence Query Language) basics

### 🛠 Implementation Steps

#### Step 1.1: Add Confluence Configuration
**File**: `internal/config/config.go`

```go
// Add Confluence to Config struct
type Config struct {
    GitHub     provider.Config `json:"github"`
    JIRA       provider.Config `json:"jira"`
    Obsidian   provider.Config `json:"obsidian"`
    Confluence provider.Config `json:"confluence"`  // NEW
}

// Update DefaultConfig()
func DefaultConfig() *Config {
    return &Config{
        // ... existing configs
        Confluence: provider.Config{
            Enabled: false,
        },
    }
}
```

#### Step 1.2: Create Confluence Provider
**File**: `internal/provider/confluence/confluence.go`

```go
package confluence

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "os"
    "strings"
    "time"

    "daily/internal/activity"
    "daily/internal/provider"
)

type Provider struct {
    config provider.Config
}

func New(config provider.Config) *Provider {
    return &Provider{config: config}
}

func (p *Provider) Name() string {
    return "confluence"
}

func (p *Provider) IsConfigured() bool {
    token := os.Getenv("JIRA_API_TOKEN")
    email := os.Getenv("ATLASSIAN_EMAIL")
    domain := os.Getenv("ATLASSIAN_DOMAIN")
    return p.config.Enabled && token != "" && email != "" && domain != ""
}

func (p *Provider) GetActivities(ctx context.Context, from, to time.Time) ([]activity.Activity, error) {
    // Calculate relative date for CQL
    since := calculateRelativeDate(from)

    // Get mentions using CQL
    mentions, err := p.getConfluenceMentions(ctx, since)
    if err != nil {
        return nil, fmt.Errorf("failed to get confluence mentions: %w", err)
    }

    var activities []activity.Activity
    for _, mention := range mentions {
        activities = append(activities, activity.Activity{
            ID:          mention.ID,
            Type:        activity.ActivityTypeNote, // or create ActivityTypeConfluence
            Title:       mention.Title,
            Description: fmt.Sprintf("Confluence %s", mention.Type),
            URL:         mention.URL,
            Platform:    "confluence",
            Timestamp:   mention.UpdatedAt,
            Tags:        []string{"mention"},
        })
    }

    return activities, nil
}

type confluenceMention struct {
    ID        string
    Title     string
    Type      string
    URL       string
    UpdatedAt time.Time
}

func (p *Provider) getConfluenceMentions(ctx context.Context, since string) ([]confluenceMention, error) {
    // Implementation similar to doctobriefing
    token := os.Getenv("JIRA_API_TOKEN")
    email := os.Getenv("ATLASSIAN_EMAIL")
    domain := os.Getenv("ATLASSIAN_DOMAIN")

    // CQL to find mentions
    cql := fmt.Sprintf("mention = currentUser() AND lastModified >= now(\"%s\")", since)

    // Build request
    apiURL := fmt.Sprintf("https://%s/wiki/rest/api/search", domain)
    params := url.Values{}
    params.Add("cql", cql)
    params.Add("limit", "50")
    fullURL := apiURL + "?" + params.Encode()

    req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    req.SetBasicAuth(email, token)
    req.Header.Set("Accept", "application/json")
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{Timeout: 30 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to execute request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %w", err)
    }

    var result confluenceSearchResult
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }

    var mentions []confluenceMention
    for _, item := range result.Results {
        mentions = append(mentions, confluenceMention{
            ID:        item.Content.ID,
            Title:     item.Content.Title,
            Type:      item.Content.Type,
            URL:       fmt.Sprintf("https://%s/wiki%s", domain, item.URL),
            UpdatedAt: time.Now(), // Confluence search doesn't provide exact timestamp
        })
    }

    return mentions, nil
}

type confluenceSearchResult struct {
    Results []struct {
        Content struct {
            ID    string `json:"id"`
            Title string `json:"title"`
            Type  string `json:"type"`
        } `json:"content"`
        URL string `json:"url"`
    } `json:"results"`
}

func calculateRelativeDate(from time.Time) string {
    duration := time.Since(from)
    days := int(duration.Hours() / 24)

    if days <= 1 {
        return "-1d"
    } else if days <= 7 {
        return fmt.Sprintf("-%dd", days)
    } else if days <= 30 {
        weeks := days / 7
        return fmt.Sprintf("-%dw", weeks)
    } else {
        months := days / 30
        return fmt.Sprintf("-%dm", months)
    }
}
```

#### Step 1.3: Add Confluence Tests
**File**: `internal/provider/confluence/confluence_test.go`

```go
package confluence

import (
    "context"
    "testing"
    "time"

    "daily/internal/provider"
)

func TestProvider_Name(t *testing.T) {
    p := New(provider.Config{})
    if got := p.Name(); got != "confluence" {
        t.Errorf("Name() = %v, want %v", got, "confluence")
    }
}

func TestProvider_IsConfigured(t *testing.T) {
    tests := []struct {
        name    string
        config  provider.Config
        env     map[string]string
        want    bool
    }{
        {
            name:   "not enabled",
            config: provider.Config{Enabled: false},
            env:    map[string]string{},
            want:   false,
        },
        {
            name:   "enabled but no env vars",
            config: provider.Config{Enabled: true},
            env:    map[string]string{},
            want:   false,
        },
        {
            name:   "enabled with all env vars",
            config: provider.Config{Enabled: true},
            env: map[string]string{
                "JIRA_API_TOKEN":    "token",
                "ATLASSIAN_EMAIL":   "email",
                "ATLASSIAN_DOMAIN":  "domain",
            },
            want: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Set environment variables
            for k, v := range tt.env {
                t.Setenv(k, v)
            }

            p := New(tt.config)
            if got := p.IsConfigured(); got != tt.want {
                t.Errorf("IsConfigured() = %v, want %v", got, tt.want)
            }
        })
    }
}

// Add more tests for GetActivities, API integration, etc.
```

#### Step 1.4: Integrate Confluence Provider
**File**: `cmd/sum.go`, `cmd/todo.go`, `cmd/reviews.go`

```go
// In all command files, add Confluence provider
import (
    "daily/internal/provider/confluence"
    // ... other imports
)

// In the command logic
confluenceProvider := confluence.New(config.Confluence)
aggregator := provider.NewAggregator(
    github.New(config.GitHub),
    jira.New(config.JIRA),
    obsidian.New(config.Obsidian),
    confluenceProvider, // NEW
)
```

#### Step 1.5: Update Documentation
**File**: `CLAUDE.md` (add Confluence section)

```markdown
### Confluence Configuration
- `JIRA_API_TOKEN` - Atlassian API token (same as JIRA)
- `ATLASSIAN_EMAIL` - Your Atlassian email
- `ATLASSIAN_DOMAIN` - Your domain (e.g., company.atlassian.net)

Example:
```bash
export JIRA_API_TOKEN="your-token"
export ATLASSIAN_EMAIL="you@company.com"
export ATLASSIAN_DOMAIN="company.atlassian.net"
./daily sum -v
```

The Confluence provider finds:
- Pages where you are mentioned
- Comments where you are mentioned
- Recent updates to pages you're mentioned in
```

---

## Feature 2: Local Hide Cache

### 🎯 Goal
Allow users to permanently hide/ignore specific activities across sessions.

### 📋 Prerequisites
- Understanding of daily's architecture
- File system operations knowledge

### 🛠 Implementation Steps

#### Step 2.1: Create Cache Package
**File**: `internal/cache/cache.go` (exists, extend it)

```go
// Add to existing cache.go
import (
    "crypto/sha1"
    "fmt"
)

// Add hide functionality to existing Cache struct
func (c *Cache) IsHidden(title string) bool {
    hash := c.generateSHA1(title)
    c.mu.RLock()
    defer c.mu.RUnlock()

    // Check if hash exists in hiddenItems set
    return c.data["hidden_items"].(map[string]bool)[hash]
}

func (c *Cache) HideItem(title string) error {
    hash := c.generateSHA1(title)
    c.mu.Lock()
    defer c.mu.Unlock()

    // Initialize hidden_items if it doesn't exist
    if c.data["hidden_items"] == nil {
        c.data["hidden_items"] = make(map[string]bool)
    }

    hiddenItems := c.data["hidden_items"].(map[string]bool)
    hiddenItems[hash] = true

    return c.Save()
}

func (c *Cache) ShowItem(title string) error {
    hash := c.generateSHA1(title)
    c.mu.Lock()
    defer c.mu.Unlock()

    if c.data["hidden_items"] != nil {
        hiddenItems := c.data["hidden_items"].(map[string]bool)
        delete(hiddenItems, hash)
    }

    return c.Save()
}

func (c *Cache) FilterActivities(activities []activity.Activity) []activity.Activity {
    var filtered []activity.Activity
    for _, act := range activities {
        if !c.IsHidden(act.Title) {
            filtered = append(filtered, act)
        }
    }
    return filtered
}

func (c *Cache) GetHiddenCount() int {
    c.mu.RLock()
    defer c.mu.RUnlock()

    if c.data["hidden_items"] == nil {
        return 0
    }
    return len(c.data["hidden_items"].(map[string]bool))
}

func (c *Cache) generateSHA1(title string) string {
    h := sha1.New()
    h.Write([]byte(strings.TrimSpace(title)))
    return fmt.Sprintf("%x", h.Sum(nil))
}
```

#### Step 2.2: Add Hide Command
**File**: `cmd/hide.go` (new file)

```go
package cmd

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"

    "daily/internal/cache"
    "daily/internal/config"
)

func HideCmd() *cobra.Command {
    var unhide bool
    var list bool

    cmd := &cobra.Command{
        Use:   "hide [title]",
        Short: "Hide or unhide activities by title",
        Long: `Hide activities to prevent them from showing in future summaries.

Examples:
  daily hide "Some annoying notification"  # Hide this title
  daily hide --list                       # List hidden items
  daily hide --unhide "Previously hidden" # Unhide this title`,
        Args: cobra.MaximumNArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            config, err := config.Load()
            if err != nil {
                return fmt.Errorf("failed to load config: %w", err)
            }

            cache, err := cache.New()
            if err != nil {
                return fmt.Errorf("failed to load cache: %w", err)
            }

            if list {
                return listHiddenItems(cache)
            }

            if len(args) == 0 {
                return fmt.Errorf("title is required unless using --list")
            }

            title := args[0]

            if unhide {
                if err := cache.ShowItem(title); err != nil {
                    return fmt.Errorf("failed to unhide item: %w", err)
                }
                fmt.Printf("✓ Unhidden: %s\n", title)
            } else {
                if err := cache.HideItem(title); err != nil {
                    return fmt.Errorf("failed to hide item: %w", err)
                }
                fmt.Printf("✓ Hidden: %s\n", title)
                fmt.Printf("  This item will not appear in future summaries.\n")
            }

            return nil
        },
    }

    cmd.Flags().BoolVar(&unhide, "unhide", false, "Unhide the specified title")
    cmd.Flags().BoolVar(&list, "list", false, "List all hidden items")

    return cmd
}

func listHiddenItems(cache *cache.Cache) error {
    count := cache.GetHiddenCount()
    fmt.Printf("Hidden items: %d\n", count)

    if count > 0 {
        fmt.Printf("Use 'daily hide --unhide \"<title>\"' to unhide items\n")
    }

    return nil
}
```

#### Step 2.3: Integrate Hide Functionality
**File**: `main.go`

```go
// Add hide command
rootCmd.AddCommand(cmd.HideCmd())
```

**Files**: `cmd/sum.go`, `cmd/todo.go`, `cmd/reviews.go`

```go
// In each command, after getting activities but before display:
cache, err := cache.New()
if err != nil {
    // Log warning but continue
    fmt.Fprintf(os.Stderr, "Warning: failed to load cache: %v\n", err)
} else {
    // Filter hidden activities
    for i, group := range summary.GroupByPlatform() {
        filteredActivities := cache.FilterActivities(group)
        // Update summary with filtered activities
        // This may require refactoring Summary to allow modification
    }
}
```

#### Step 2.4: Add Cache Configuration
**File**: `internal/config/config.go`

```go
type Config struct {
    GitHub     provider.Config `json:"github"`
    JIRA       provider.Config `json:"jira"`
    Obsidian   provider.Config `json:"obsidian"`
    Confluence provider.Config `json:"confluence"`
    Cache      CacheConfig     `json:"cache"`        // NEW
}

type CacheConfig struct {
    EnableHiding bool `json:"enable_hiding"`
}

// Update DefaultConfig()
func DefaultConfig() *Config {
    return &Config{
        // ... existing configs
        Cache: CacheConfig{
            EnableHiding: true,
        },
    }
}
```

#### Step 2.5: Add Tests
**File**: `internal/cache/cache_test.go` (extend existing)

```go
func TestCache_HideItem(t *testing.T) {
    // Test hide functionality
}

func TestCache_IsHidden(t *testing.T) {
    // Test hidden check
}

func TestCache_FilterActivities(t *testing.T) {
    // Test activity filtering
}
```

---

# Phase 2: WebUI with WebSocket

## Feature 3: WebUI Implementation

### 🎯 Goal
Add a web interface with real-time updates for viewing and managing daily activities.

### 📋 Prerequisites
- HTTP server knowledge
- WebSocket implementation
- Frontend development (HTML/CSS/JS)

### 🛠 Implementation Steps

#### Step 3.1: Add Dependencies
**File**: `go.mod`

```go
require (
    // ... existing dependencies
    github.com/gorilla/websocket v1.5.3
)
```

#### Step 3.2: Create WebUI Package
**File**: `internal/webui/server.go`

```go
package webui

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/gorilla/websocket"

    "daily/internal/activity"
    "daily/internal/cache"
    "daily/internal/config"
    "daily/internal/provider"
)

type Server struct {
    config     *config.Config
    cache      *cache.Cache
    aggregator *provider.Aggregator
    upgrader   websocket.Upgrader
}

type WebSocketMessage struct {
    Type    string      `json:"type"`
    Data    interface{} `json:"data,omitempty"`
    SHA1    string      `json:"sha1,omitempty"`
    Message string      `json:"message,omitempty"`
}

func NewServer(config *config.Config, cache *cache.Cache, aggregator *provider.Aggregator) *Server {
    return &Server{
        config:     config,
        cache:      cache,
        aggregator: aggregator,
        upgrader: websocket.Upgrader{
            CheckOrigin: func(r *http.Request) bool {
                return r.Host == "localhost:8080" || r.Host == "127.0.0.1:8080"
            },
        },
    }
}

func (s *Server) Start(port int) error {
    http.HandleFunc("/", s.serveHome)
    http.HandleFunc("/ws", s.handleWebSocket)
    http.HandleFunc("/api/summary", s.handleAPISummary)

    addr := fmt.Sprintf("127.0.0.1:%d", port)
    fmt.Printf("🚀 WebUI server starting on http://%s\n", addr)
    return http.ListenAndServe(addr, nil)
}

func (s *Server) serveHome(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        http.NotFound(w, r)
        return
    }

    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.Write([]byte(getHomeHTML()))
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := s.upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("WebSocket upgrade failed: %v", err)
        return
    }
    defer conn.Close()

    // Send initial data
    if err := s.sendSummary(conn, time.Now().AddDate(0, 0, -1)); err != nil {
        log.Printf("Failed to send initial summary: %v", err)
        return
    }

    // Handle incoming messages
    for {
        var msg WebSocketMessage
        err := conn.ReadJSON(&msg)
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Printf("WebSocket error: %v", err)
            }
            break
        }

        switch msg.Type {
        case "hide":
            s.handleHide(conn, msg.SHA1)
        case "refresh":
            date := time.Now().AddDate(0, 0, -1) // Default to yesterday
            s.sendSummary(conn, date)
        }
    }
}

func (s *Server) handleHide(conn *websocket.Conn, sha1Hash string) {
    // Find the activity by SHA1 and hide it
    // This requires extending cache to support SHA1-based hiding
    if err := s.cache.HideItemBySHA1(sha1Hash); err != nil {
        s.sendError(conn, fmt.Sprintf("Failed to hide item: %v", err))
        return
    }

    msg := WebSocketMessage{
        Type: "hidden",
        SHA1: sha1Hash,
    }
    conn.WriteJSON(msg)
}

func (s *Server) sendSummary(conn *websocket.Conn, date time.Time) error {
    summary, err := s.aggregator.GetSummary(context.Background(), date)
    if err != nil {
        return err
    }

    // Filter hidden items
    activities := s.cache.FilterActivities(summary.Activities)

    // Add SHA1 hashes for hiding functionality
    for i := range activities {
        activities[i].SHA1 = s.cache.GenerateSHA1(activities[i].Title)
    }

    msg := WebSocketMessage{
        Type: "summary",
        Data: activities,
    }

    return conn.WriteJSON(msg)
}

func (s *Server) sendError(conn *websocket.Conn, message string) {
    msg := WebSocketMessage{
        Type:    "error",
        Message: message,
    }
    conn.WriteJSON(msg)
}

func (s *Server) handleAPISummary(w http.ResponseWriter, r *http.Request) {
    date := time.Now().AddDate(0, 0, -1)
    summary, err := s.aggregator.GetSummary(r.Context(), date)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    activities := s.cache.FilterActivities(summary.Activities)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(activities)
}
```

#### Step 3.3: Create HTML Template
**File**: `internal/webui/template.go`

```go
package webui

func getHomeHTML() string {
    return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>🌅 Daily - Work Activity Dashboard</title>
    <style>
        /* Copy and adapt the CSS from doctobriefing */
        /* Customize for daily's branding and needs */
    </style>
</head>
<body>
    <div class="header">
        <h1>🌅 Daily - Work Activity Dashboard</h1>
        <div class="status" id="status">Connecting...</div>
    </div>

    <div class="controls">
        <button onclick="refresh()">🔄 Refresh</button>
        <select id="dateSelector" onchange="changeDate()">
            <option value="0">Today</option>
            <option value="-1" selected>Yesterday</option>
            <option value="-2">2 days ago</option>
            <option value="-7">1 week ago</option>
        </select>
    </div>

    <div id="content">
        <div class="loading">Loading activities...</div>
    </div>

    <script>
        // Copy and adapt the JavaScript from doctobriefing
        // Customize for daily's data structure and needs
    </script>
</body>
</html>`
}
```

#### Step 3.4: Add WebUI Command
**File**: `cmd/webui.go`

```go
package cmd

import (
    "fmt"

    "github.com/spf13/cobra"

    "daily/internal/cache"
    "daily/internal/config"
    "daily/internal/provider"
    "daily/internal/provider/github"
    "daily/internal/provider/jira"
    "daily/internal/provider/obsidian"
    "daily/internal/provider/confluence"
    "daily/internal/webui"
)

func WebUICmd() *cobra.Command {
    var port int

    cmd := &cobra.Command{
        Use:   "webui",
        Short: "Launch web interface for daily activities",
        Long:  `Start a web server with real-time interface for viewing and managing daily activities.`,
        RunE: func(cmd *cobra.Command, args []string) error {
            config, err := config.Load()
            if err != nil {
                return fmt.Errorf("failed to load config: %w", err)
            }

            cache, err := cache.New()
            if err != nil {
                return fmt.Errorf("failed to load cache: %w", err)
            }

            // Create providers
            githubProvider := github.New(config.GitHub)
            jiraProvider := jira.New(config.JIRA)
            obsidianProvider := obsidian.New(config.Obsidian)
            confluenceProvider := confluence.New(config.Confluence)

            aggregator := provider.NewAggregator(
                githubProvider,
                jiraProvider,
                obsidianProvider,
                confluenceProvider,
            )

            server := webui.NewServer(config, cache, aggregator)
            return server.Start(port)
        },
    }

    cmd.Flags().IntVar(&port, "port", 8080, "Port to serve web interface")

    return cmd
}
```

#### Step 3.5: Integrate WebUI Command
**File**: `main.go`

```go
// Add webui command
rootCmd.AddCommand(cmd.WebUICmd())
```

#### Step 3.6: Update Configuration
**File**: `internal/config/config.go`

```go
type Config struct {
    GitHub     provider.Config `json:"github"`
    JIRA       provider.Config `json:"jira"`
    Obsidian   provider.Config `json:"obsidian"`
    Confluence provider.Config `json:"confluence"`
    Cache      CacheConfig     `json:"cache"`
    WebUI      WebUIConfig     `json:"webui"`        // NEW
}

type WebUIConfig struct {
    Enabled bool `json:"enabled"`
    Port    int  `json:"port"`
}

// Update DefaultConfig()
func DefaultConfig() *Config {
    return &Config{
        // ... existing configs
        WebUI: WebUIConfig{
            Enabled: true,
            Port:    8080,
        },
    }
}
```

---

## Testing Strategy

### Unit Tests
1. **Confluence Provider Tests**
   - API integration mocking
   - CQL query generation
   - Date calculation
   - Error handling

2. **Cache Functionality Tests**
   - Hide/show operations
   - SHA1 generation
   - Activity filtering
   - Persistence

3. **WebUI Server Tests**
   - HTTP endpoints
   - WebSocket message handling
   - Error scenarios

### Integration Tests
1. **End-to-end provider tests**
2. **WebUI with real WebSocket client**
3. **Cache persistence across restarts**

### Manual Testing
1. **Confluence API with real credentials**
2. **WebUI in different browsers**
3. **WebSocket reconnection scenarios**

---

## Documentation Updates

### CLAUDE.md Updates
```markdown
## New Features

### Confluence Integration
- Tracks mentions in Confluence pages and comments
- Uses same Atlassian credentials as JIRA
- Configurable via environment variables

### Activity Hiding
- Hide irrelevant activities permanently
- SHA1-based identification for consistency
- CLI and WebUI support

### Web Interface
- Real-time dashboard on localhost:8080
- Interactive hiding and filtering
- Cross-device access

## New Commands
- `daily hide <title>` - Hide activities by title
- `daily hide --list` - List hidden items
- `daily hide --unhide <title>` - Unhide activities
- `daily webui` - Launch web interface
```

### README Updates
- Add Confluence setup instructions
- Add WebUI usage examples
- Update environment variables list

---

## Migration Guide

### For Existing Users
1. **Automatic config migration** - new fields added with safe defaults
2. **Backward compatibility** - all existing functionality unchanged
3. **Optional features** - new features disabled by default if needed

### For New Users
1. **Enhanced onboarding** - setup wizard including Confluence
2. **Quick start guide** - get running with WebUI quickly

---

## Deployment Considerations

### Dependencies
- Add `github.com/gorilla/websocket` to go.mod
- No breaking changes to existing dependencies

### Configuration
- Environment variables for Confluence (same as JIRA)
- Optional features can be disabled
- Configurable WebUI port

### Security
- WebUI only binds to localhost by default
- CORS protection for WebSocket
- No external network exposure

---

## Success Metrics

### Phase 1 Success
- [ ] Confluence provider returns valid activities
- [ ] Hide cache persists across sessions
- [ ] Hidden activities don't appear in CLI output
- [ ] All existing functionality works unchanged

### Phase 2 Success
- [ ] WebUI serves on localhost:8080
- [ ] WebSocket provides real-time updates
- [ ] Activities can be hidden from web interface
- [ ] Responsive design works on different screen sizes

### Overall Success
- [ ] All existing tests pass
- [ ] New features have >= 80% test coverage
- [ ] Documentation is complete and accurate
- [ ] No breaking changes for existing users
- [ ] Performance impact < 10% for CLI usage

---

## Timeline Estimate

### Phase 1 (2 weeks)
- **Week 1**: Confluence provider + tests
- **Week 2**: Hide cache + integration

### Phase 2 (3 weeks)
- **Week 1**: WebUI server + WebSocket
- **Week 2**: Frontend interface + styling
- **Week 3**: Integration, testing, documentation

**Total**: 5 weeks for complete implementation

---

## Questions for Consideration

1. Should Confluence be enabled by default or opt-in?
2. Should hidden items be configurable per-command or global?
3. Should WebUI support multiple date ranges like doctobriefing?
4. Should we add authentication to WebUI for multi-user scenarios?
5. Should we support custom CSS themes for WebUI?

This implementation plan provides a roadmap for systematically adding all three key features from doctobriefing to daily while maintaining backward compatibility and following daily's existing architectural patterns.