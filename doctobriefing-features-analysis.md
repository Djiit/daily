# Daily vs Doctobriefing: Feature Analysis

## Overview

This document analyzes the key features that `doctobriefing` has that `daily` currently lacks. After thorough analysis of both codebases, three major features emerge as valuable additions that could enhance `daily`'s functionality.

## Architecture Comparison

### Daily Architecture
- **Go CLI tool** with Cobra framework
- **Provider-based system** (GitHub, JIRA, Obsidian)
- **TUI interface** using Bubble Tea
- **JSON configuration** in `~/.config/daily/config.json`
- **Commands**: `sum`, `todo`, `reviews`, `config`

### Doctobriefing Architecture
- **Go CLI tool** with flag-based configuration
- **Multi-source integration** (Confluence, JIRA, GitHub)
- **WebUI with WebSocket** support
- **Local hide/ignore cache** mechanism
- **Single command** with multiple flags

## Key Missing Features in Daily

## 1. 🔗 Confluence Integration

### Current State in Daily
- ❌ **No Confluence support** whatsoever
- ❌ **No mention tracking** from Confluence
- ❌ **No Confluence activity aggregation**

### Implementation in Doctobriefing
```go
// Confluence API integration
func getConfluenceMentions(since string) ([]BriefingItem, error) {
    // Uses Confluence REST API: https://{domain}/wiki/rest/api/search
    // CQL Query: "mention = currentUser() AND lastModified >= now(\"-2w\")"
    // Environment variables: JIRA_API_TOKEN, ATLASSIAN_EMAIL, ATLASSIAN_DOMAIN
}
```

**Key Features:**
- **CQL (Confluence Query Language)** support
- **Mention detection** using `mention = currentUser()`
- **Relative date filtering** (`-1d`, `-2w`, `-1m`)
- **Priority assignment** (comments = high, pages = normal)
- **Basic authentication** with email + API token
- **URL construction** for clickable links

**Environment Variables Required:**
```bash
JIRA_API_TOKEN=your_token       # Same token works for Confluence
ATLASSIAN_EMAIL=your@email.com
ATLASSIAN_DOMAIN=company.atlassian.net
```

### Value Proposition
- **Mentions tracking**: Find when you're mentioned in Confluence pages/comments
- **Activity awareness**: Stay updated on Confluence content changes
- **Unified workflow**: Single tool for GitHub, JIRA, Obsidian, and Confluence

---

## 2. 🌐 WebUI with WebSocket Support

### Current State in Daily
- ❌ **CLI-only interface** (TUI available but no web interface)
- ❌ **No real-time updates**
- ❌ **No browser-based interaction**
- ❌ **No remote access capability**

### Implementation in Doctobriefing
```go
// WebSocket server architecture
type WebUIServer struct {
    hiddenDB *HiddenItemsDB
    upgrader websocket.Upgrader
}

// WebSocket message protocol
type WebSocketMessage struct {
    Type string      `json:"type"`    // "refresh|hide|briefing|error"
    Data interface{} `json:"data,omitempty"`
    SHA1 string      `json:"sha1,omitempty"`
    Since string     `json:"since,omitempty"`
}
```

**Key Features:**
- **HTTP server** on `localhost:8080`
- **WebSocket communication** for real-time updates
- **Single-page application** with embedded HTML/CSS/JS
- **Interactive hiding** of items
- **Time period selection** dropdown (1d, 3d, 1w, 2w, 1m, 3m)
- **Priority-based color coding**:
  - 🚨 **Urgent**: Red
  - 🔔 **High**: Orange
  - 📌 **Normal**: Blue
  - 💬 **Low**: Green
- **Dark theme** with professional styling
- **Automatic reconnection** on disconnect
- **Cross-platform browser support**

**Web Interface Benefits:**
- **Remote access**: View briefings from any device with web browser
- **Better UX**: Click to hide items, dropdown filtering
- **Real-time updates**: No manual refresh needed
- **Persistent sessions**: Reconnects automatically
- **Visual feedback**: Loading states, error messages

### Technical Implementation
```go
// Launch WebUI mode
./doctobriefing --webui

// Server endpoints
http.HandleFunc("/", serveHome)
http.HandleFunc("/ws", handleWebSocket)
```

---

## 3. 💾 Local Ignore/Hide Cache Mechanism

### Current State in Daily
- ❌ **No item filtering** capability
- ❌ **No persistent ignore functionality**
- ❌ **Cannot hide recurring irrelevant items**
- ❌ **No user customization** of displayed content

### Implementation in Doctobriefing
```go
// Hidden items database
type HiddenItemsDB struct {
    filePath   string
    hiddenSHA1 map[string]bool // O(1) lookup
}

// Core functionality
func (db *HiddenItemsDB) HideItem(title string) error
func (db *HiddenItemsDB) IsHidden(title string) bool
func (db *HiddenItemsDB) FilterBriefingItems(items []BriefingItem) []BriefingItem
```

**Key Features:**
- **SHA1-based identification**: Uses title hashing for unique IDs
- **Local storage**: `~/.doctobriefing_hidden_db.txt`
- **In-memory caching**: `map[string]bool` for O(1) lookup
- **Persistent across sessions**: Automatically loads on startup
- **Atomic operations**: Proper error handling and rollback
- **WebSocket integration**: Real-time hiding from web UI

**File Format:**
```
# ~/.doctobriefing_hidden_db.txt
da39a3ee5e6b4b0d3255bfef95601890afd80709
356a192b7913b04c54574d18c28d46e6395428ab
...
```

**User Workflow:**
1. **View briefing** with all items
2. **Hide irrelevant items** (CLI or WebUI)
3. **Items stay hidden** permanently
4. **Cleaner briefings** in future runs

### Value Proposition
- **Noise reduction**: Hide recurring irrelevant items
- **Personalization**: Customize what you see
- **Productivity**: Focus on what matters
- **Cross-interface**: Works in both CLI and WebUI

---

## Integration Complexity Assessment

### 1. Confluence Integration - **EASY** ⭐⭐
- **Effort**: ~1-2 days
- **Complexity**: Low - follows existing provider pattern
- **Dependencies**: None (uses standard library HTTP)
- **Breaking changes**: None

### 2. Local Hide Cache - **EASY** ⭐⭐
- **Effort**: ~1-2 days
- **Complexity**: Low - simple file-based storage
- **Dependencies**: None (uses standard library)
- **Breaking changes**: None

### 3. WebUI with WebSocket - **MODERATE** ⭐⭐⭐⭐
- **Effort**: ~1-2 weeks
- **Complexity**: High - new HTTP server, WebSocket handling, frontend
- **Dependencies**: `github.com/gorilla/websocket`
- **Breaking changes**: None (new optional feature)

---

## Recommendations

### Phase 1: Quick Wins (1-2 days each)
1. **Add Confluence provider** - immediate value for Atlassian users
2. **Add local hide cache** - significant UX improvement

### Phase 2: Advanced Features (1-2 weeks)
3. **Add WebUI with WebSocket** - major new capability

### Phase 3: Integration & Polish
4. **TUI integration** with hide functionality
5. **Configuration management** for new features
6. **Documentation and examples**

---

## Technical Considerations

### Confluence Provider Integration
- Reuse existing `provider.Provider` interface
- Add Confluence config to main config struct
- Follow same pattern as GitHub/JIRA providers
- Add to `NewAggregator()` initialization

### Hide Cache Integration
- Add as optional feature with config flag
- Integrate with existing output formatters
- Add CLI commands for managing hidden items
- Consider TUI integration

### WebUI Integration
- Add as new command: `daily webui`
- Reuse existing provider aggregation logic
- Consider making port configurable
- Add proper shutdown handling

## Conclusion

All three features would significantly enhance `daily`'s value proposition:

1. **Confluence integration** fills a major gap for Atlassian users
2. **Local hide cache** dramatically improves UX by reducing noise
3. **WebUI with WebSocket** provides modern, accessible interface

The first two features are relatively simple to implement and provide immediate value. The WebUI is more complex but offers substantial long-term benefits for user adoption and ease of use.