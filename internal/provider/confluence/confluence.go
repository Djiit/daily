package confluence

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"daily/internal/activity"
	"daily/internal/provider"
)

type Provider struct {
	config provider.Config
	client *http.Client
}

func NewProvider(config provider.Config) *Provider {
	return &Provider{
		config: config,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (p *Provider) Name() string {
	return "confluence"
}

func (p *Provider) IsConfigured() bool {
	return p.config.Enabled &&
		p.config.Token != "" &&
		p.config.Email != "" &&
		p.config.URL != ""
}

// GetActivities retrieves pages that the user contributed to (for summary)
func (p *Provider) GetActivities(ctx context.Context, from, to time.Time) ([]activity.Activity, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("Confluence provider not configured")
	}

	activities := make([]activity.Activity, 0)

	// Get pages contributed to by the user in the time range
	contributions, err := p.getContributions(ctx, from, to)
	if err != nil {
		// Log error but continue with empty results - warning handled by aggregator
		fmt.Printf("Confluence: error fetching contributions: %v", err)
	} else {
		activities = append(activities, contributions...)
	}

	return activities, nil
}

// GetMentions retrieves pages that mention the user (for todos)
func (p *Provider) GetMentions(ctx context.Context, since string) ([]TodoItem, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("Confluence provider not configured")
	}

	// Ensure since has "-" prefix for CQL format
	if !strings.HasPrefix(since, "-") {
		since = "-" + since
	}

	// CQL to find mentions of current user
	cql := fmt.Sprintf("mention = currentUser() AND lastModified >= now(\"%s\")", since)

	searchResults, err := p.searchConfluence(ctx, cql)
	if err != nil {
		return nil, fmt.Errorf("failed to search for mentions: %w", err)
	}

	var mentions []TodoItem
	for _, result := range searchResults.Results {
		priority := "normal"
		if result.Content.Type == "comment" {
			priority = "high" // Comments usually need responses
		}

		mentions = append(mentions, TodoItem{
			ID:          result.Content.ID,
			Title:       result.Content.Title,
			Description: fmt.Sprintf("Type: %s", strings.Title(result.Content.Type)),
			URL:         fmt.Sprintf("%s/wiki%s", p.getBaseURL(), result.URL),
			UpdatedAt:   time.Now(), // Confluence search doesn't provide lastModified in this format
			Tags:        []string{priority},
		})
	}

	return mentions, nil
}

// getContributions retrieves pages that the user contributed to
func (p *Provider) getContributions(ctx context.Context, from, to time.Time) ([]activity.Activity, error) {
	// CQL to find pages contributed to by current user in date range
	cql := fmt.Sprintf("contributor = currentUser() AND lastModified >= \"%s\" AND lastModified < \"%s\"",
		from.Format("2006-01-02"),
		to.Format("2006-01-02"))

	searchResults, err := p.searchConfluence(ctx, cql)
	if err != nil {
		return nil, fmt.Errorf("failed to search for contributions: %w", err)
	}

	var activities []activity.Activity
	for _, result := range searchResults.Results {
		activities = append(activities, activity.Activity{
			ID:          result.Content.ID,
			Type:        activity.ActivityTypeConfluenceContribution,
			Title:       result.Content.Title,
			Description: fmt.Sprintf("Modified %s", strings.ToLower(result.Content.Type)),
			URL:         fmt.Sprintf("%s/wiki%s", p.getBaseURL(), result.URL),
			Platform:    "confluence",
			Timestamp:   time.Now(), // Will be updated when we can parse lastModified properly
			Tags:        []string{result.Content.Type},
		})
	}

	return activities, nil
}

// getBaseURL returns the properly formatted base URL with https prefix
func (p *Provider) getBaseURL() string {
	baseURL := strings.TrimSuffix(p.config.URL, "/")
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}
	return baseURL
}

// searchConfluence performs a CQL search against Confluence
func (p *Provider) searchConfluence(ctx context.Context, cql string) (*ConfluenceSearchResult, error) {
	// Properly encode URL parameters
	apiURL := fmt.Sprintf("%s/wiki/rest/api/search", p.getBaseURL())
	params := url.Values{}
	params.Add("cql", cql)
	params.Add("limit", "50")
	fullURL := apiURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Confluence request: %w", err)
	}

	// Basic auth with email + API token
	req.SetBasicAuth(p.config.Email, p.config.Token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute Confluence request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log the close error but don't override the main error
			fmt.Printf("Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Confluence API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Confluence response: %w", err)
	}

	var result ConfluenceSearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse Confluence response: %w", err)
	}

	return &result, nil
}

// ConfluenceSearchResult represents Confluence search API response
type ConfluenceSearchResult struct {
	Results []struct {
		Content struct {
			ID    string `json:"id"`
			Title string `json:"title"`
			Type  string `json:"type"`
		} `json:"content"`
		URL string `json:"url"`
	} `json:"results"`
}

// TodoItem represents a single todo item (avoiding import cycles)
type TodoItem struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
	Tags        []string  `json:"tags,omitempty"`
}
