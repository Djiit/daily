package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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
			Timeout: 30 * time.Second, // Reasonable timeout for API calls
		},
	}
}

func (p *Provider) Name() string {
	return "github"
}

func (p *Provider) IsConfigured() bool {
	return p.config.Enabled && p.config.Token != "" && p.config.Username != ""
}

func (p *Provider) GetActivities(ctx context.Context, from, to time.Time) ([]activity.Activity, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("GitHub provider not configured")
	}

	activities := make([]activity.Activity, 0)

	// Get commits - continue even if this fails
	commits, err := p.getCommits(ctx, from, to)
	if err != nil {
		// Log error but continue with pull requests - warning handled by aggregator
	} else {
		activities = append(activities, commits...)
	}

	// Get pull requests - continue even if this fails
	pullRequests, err := p.getPullRequests(ctx, from, to)
	if err != nil {
		// Log error but continue with partial results - warning handled by aggregator
	} else {
		activities = append(activities, pullRequests...)
	}

	return activities, nil
}

func (p *Provider) getCommits(ctx context.Context, from, to time.Time) ([]activity.Activity, error) {
	// Search for commits by the user in the specified time range
	// For single day: use just the date. For range: use from..to format
	var dateQuery string
	if from.Format("2006-01-02") == to.Add(-24*time.Hour).Format("2006-01-02") {
		// Single day query
		dateQuery = from.Format("2006-01-02")
	} else {
		// Date range query
		dateQuery = fmt.Sprintf("%s..%s", from.Format("2006-01-02"), to.Add(-time.Second).Format("2006-01-02"))
	}

	query := fmt.Sprintf("author:%s committer-date:%s", p.config.Username, dateQuery)

	// Add filter if configured
	if p.config.Filter != "" {
		query = fmt.Sprintf("%s %s", query, p.config.Filter)
	}

	searchURL := fmt.Sprintf("https://api.github.com/search/commits?q=%s&sort=committer-date&order=desc",
		url.QueryEscape(query))

	var searchResult struct {
		Items []struct {
			SHA    string `json:"sha"`
			Commit struct {
				Message   string `json:"message"`
				Committer struct {
					Date time.Time `json:"date"`
				} `json:"committer"`
			} `json:"commit"`
			Repository struct {
				Name     string `json:"name"`
				FullName string `json:"full_name"`
				HTMLURL  string `json:"html_url"`
			} `json:"repository"`
		} `json:"items"`
	}

	if err := p.makeRequestWithHeaders(ctx, searchURL, map[string]string{
		"Accept": "application/vnd.github.cloak-preview+json", // Required for commit search
	}, &searchResult); err != nil {
		return nil, err
	}

	var activities []activity.Activity
	for _, item := range searchResult.Items {
		// Only include commits from the specified time range
		if item.Commit.Committer.Date.Before(from) || item.Commit.Committer.Date.After(to) {
			continue
		}

		activities = append(activities, activity.Activity{
			ID:          fmt.Sprintf("github-commit-%s", item.SHA),
			Type:        activity.ActivityTypeCommit,
			Title:       item.Commit.Message,
			Description: fmt.Sprintf("Commit in %s", item.Repository.FullName),
			URL:         fmt.Sprintf("%s/commit/%s", item.Repository.HTMLURL, item.SHA),
			Platform:    "github",
			Timestamp:   item.Commit.Committer.Date,
			Tags:        []string{item.Repository.Name},
		})
	}

	return activities, nil
}

func (p *Provider) getPullRequests(ctx context.Context, from, to time.Time) ([]activity.Activity, error) {
	// Search for pull requests created or updated by the user in the specified time range
	var dateQuery string
	if from.Format("2006-01-02") == to.Add(-24*time.Hour).Format("2006-01-02") {
		// Single day query
		dateQuery = from.Format("2006-01-02")
	} else {
		// Date range query
		dateQuery = fmt.Sprintf("%s..%s", from.Format("2006-01-02"), to.Add(-time.Second).Format("2006-01-02"))
	}

	// Include type:pr in the query BEFORE URL encoding
	query := fmt.Sprintf("author:%s created:%s type:pr", p.config.Username, dateQuery)

	// Add filter if configured
	if p.config.Filter != "" {
		query = fmt.Sprintf("%s %s", query, p.config.Filter)
	}

	searchURL := fmt.Sprintf("https://api.github.com/search/issues?q=%s&sort=created&order=desc",
		url.QueryEscape(query))

	var searchResult struct {
		Items []struct {
			Number     int       `json:"number"`
			Title      string    `json:"title"`
			Body       string    `json:"body"`
			HTMLURL    string    `json:"html_url"`
			State      string    `json:"state"`
			CreatedAt  time.Time `json:"created_at"`
			UpdatedAt  time.Time `json:"updated_at"`
			Repository struct {
				Name     string `json:"name"`
				FullName string `json:"full_name"`
			} `json:"repository_url"`
		} `json:"items"`
	}

	if err := p.makeRequest(ctx, searchURL, &searchResult); err != nil {
		return nil, err
	}

	var activities []activity.Activity
	for _, item := range searchResult.Items {
		// Only include PRs from the specified time range
		if item.CreatedAt.Before(from) || item.CreatedAt.After(to) {
			continue
		}

		// Extract repository name from URL if needed
		repoName := fmt.Sprintf("PR #%d", item.Number)

		activities = append(activities, activity.Activity{
			ID:          fmt.Sprintf("github-pr-%d", item.Number),
			Type:        activity.ActivityTypePR,
			Title:       item.Title,
			Description: fmt.Sprintf("Pull request: %s", item.State),
			URL:         item.HTMLURL,
			Platform:    "github",
			Timestamp:   item.CreatedAt,
			Tags:        []string{repoName},
		})
	}

	return activities, nil
}

func (p *Provider) makeRequest(ctx context.Context, url string, result any) error {
	return p.makeRequestWithHeaders(ctx, url, nil, result)
}

func (p *Provider) makeRequestWithHeaders(ctx context.Context, url string, extraHeaders map[string]string, result any) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "token "+p.config.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "daily-cli/1.0")

	// Add any extra headers
	for key, value := range extraHeaders {
		req.Header.Set(key, value)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API request failed with status %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

// GetOpenPRs retrieves open pull requests created by the user
func (p *Provider) GetOpenPRs(ctx context.Context) ([]TodoItem, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("GitHub provider not configured")
	}

	query := fmt.Sprintf("author:%s state:open type:pr", p.config.Username)

	// Add filter if configured
	if p.config.Filter != "" {
		query = fmt.Sprintf("%s %s", query, p.config.Filter)
	}

	searchURL := fmt.Sprintf("https://api.github.com/search/issues?q=%s&sort=updated&order=desc&per_page=50",
		url.QueryEscape(query))

	var searchResult struct {
		Items []struct {
			Number     int       `json:"number"`
			Title      string    `json:"title"`
			HTMLURL    string    `json:"html_url"`
			UpdatedAt  time.Time `json:"updated_at"`
			Repository struct {
				Name     string `json:"name"`
				FullName string `json:"full_name"`
			} `json:"repository"`
		} `json:"items"`
	}

	// Note: GitHub Search API sometimes returns repository data in different fields
	// We'll need to extract the repo name from the HTML URL if needed
	if err := p.makeRequest(ctx, searchURL, &searchResult); err != nil {
		return nil, err
	}

	var todos []TodoItem
	for _, item := range searchResult.Items {
		// Extract repository name from URL if repository field is not available
		repoName := fmt.Sprintf("PR #%d", item.Number)
		if item.Repository.FullName != "" {
			repoName = item.Repository.FullName
		} else if item.Repository.Name != "" {
			repoName = item.Repository.Name
		}

		todos = append(todos, TodoItem{
			ID:          fmt.Sprintf("github-pr-%d", item.Number),
			Title:       item.Title,
			Description: fmt.Sprintf("Open PR in %s", repoName),
			URL:         item.HTMLURL,
			UpdatedAt:   item.UpdatedAt,
			Tags:        []string{repoName, "open"},
		})
	}

	return todos, nil
}

// GetPendingReviews retrieves pull requests where the user is requested as a reviewer
func (p *Provider) GetPendingReviews(ctx context.Context) ([]TodoItem, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("GitHub provider not configured")
	}

	query := fmt.Sprintf("review-requested:%s state:open type:pr", p.config.Username)

	// Add filter if configured
	if p.config.Filter != "" {
		query = fmt.Sprintf("%s %s", query, p.config.Filter)
	}

	searchURL := fmt.Sprintf("https://api.github.com/search/issues?q=%s&sort=updated&order=desc&per_page=50",
		url.QueryEscape(query))

	var searchResult struct {
		Items []struct {
			Number     int       `json:"number"`
			Title      string    `json:"title"`
			HTMLURL    string    `json:"html_url"`
			UpdatedAt  time.Time `json:"updated_at"`
			Repository struct {
				Name     string `json:"name"`
				FullName string `json:"full_name"`
			} `json:"repository"`
		} `json:"items"`
	}

	if err := p.makeRequest(ctx, searchURL, &searchResult); err != nil {
		return nil, err
	}

	var todos []TodoItem
	for _, item := range searchResult.Items {
		// Extract repository name from URL if repository field is not available
		repoName := fmt.Sprintf("PR #%d", item.Number)
		if item.Repository.FullName != "" {
			repoName = item.Repository.FullName
		} else if item.Repository.Name != "" {
			repoName = item.Repository.Name
		}

		todos = append(todos, TodoItem{
			ID:          fmt.Sprintf("github-review-%d", item.Number),
			Title:       item.Title,
			Description: fmt.Sprintf("Review requested in %s", repoName),
			URL:         item.HTMLURL,
			UpdatedAt:   item.UpdatedAt,
			Tags:        []string{repoName, "review-requested"},
		})
	}

	return todos, nil
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
