package jira

import (
	"context"
	"encoding/json"
	"fmt"
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
			Timeout: 60 * time.Second, // Increased timeout for API calls
		},
	}
}

func (p *Provider) Name() string {
	return "jira"
}

func (p *Provider) IsConfigured() bool {
	return p.config.Enabled &&
		p.config.Token != "" &&
		p.config.Email != "" &&
		p.config.URL != ""
}

func (p *Provider) GetActivities(ctx context.Context, from, to time.Time) ([]activity.Activity, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("JIRA provider not configured")
	}

	activities := make([]activity.Activity, 0)

	// Get issues updated in the time range - continue even if this fails
	issues, err := p.getUpdatedIssues(ctx, from, to)
	if err != nil {
		// Log error but continue with empty results - warning handled by aggregator
		fmt.Printf("JIRA: error fetching updated issues: %v", err)
	} else {
		activities = append(activities, issues...)
	}

	return activities, nil
}

func (p *Provider) getUpdatedIssues(ctx context.Context, from, to time.Time) ([]activity.Activity, error) {
	// Build JQL query to find issues updated in the time range
	// Use proper from and to dates - to is exclusive so it's the next day
	jql := fmt.Sprintf("assignee = currentUser() AND updated >= \"%s\" AND updated < \"%s\"",
		from.Format("2006-01-02"),
		to.Format("2006-01-02"))

	// Add filter if configured
	if p.config.Filter != "" {
		jql = fmt.Sprintf("%s AND (%s)", jql, p.config.Filter)
	}

	jql = fmt.Sprintf("%s ORDER BY updated DESC", jql)

	// URL encode the JQL query
	searchURL := fmt.Sprintf("%s/rest/api/3/search?jql=%s&fields=key,summary,status,updated,assignee",
		strings.TrimSuffix(p.config.URL, "/"),
		url.QueryEscape(jql))

	var searchResult struct {
		Issues []struct {
			Key    string `json:"key"`
			Fields struct {
				Summary string `json:"summary"`
				Updated string `json:"updated"` // Keep as string to handle different timezone formats
				Status  struct {
					Name string `json:"name"`
				} `json:"status"`
				Assignee struct {
					DisplayName string `json:"displayName"`
				} `json:"assignee"`
			} `json:"fields"`
		} `json:"issues"`
		Total int `json:"total"`
	}

	if err := p.makeRequest(ctx, searchURL, &searchResult); err != nil {
		return nil, err
	}

	var activities []activity.Activity
	for _, issue := range searchResult.Issues {
		// Parse the updated time with flexible timezone formats
		updatedTime, err := p.parseJIRATime(issue.Fields.Updated)
		if err != nil {
			continue // Skip issues with unparseable times
		}

		// Double-check the time range (API might return broader results)
		if updatedTime.Before(from) || updatedTime.After(to) {
			continue
		}

		activities = append(activities, activity.Activity{
			ID:          fmt.Sprintf("jira-%s", issue.Key),
			Type:        activity.ActivityTypeJiraTicket,
			Title:       fmt.Sprintf("%s: %s", issue.Key, issue.Fields.Summary),
			Description: fmt.Sprintf("Status: %s", issue.Fields.Status.Name),
			URL:         fmt.Sprintf("%s/browse/%s", strings.TrimSuffix(p.config.URL, "/"), issue.Key),
			Platform:    "jira",
			Timestamp:   updatedTime,
			Tags:        []string{issue.Key, issue.Fields.Status.Name},
		})
	}

	return activities, nil
}

func (p *Provider) parseJIRATime(timeStr string) (time.Time, error) {
	// Try different time formats that JIRA might use
	formats := []string{
		"2006-01-02T15:04:05.000Z0700", // "2025-08-20T18:41:17.540+0200"
		"2006-01-02T15:04:05.000-0700", // "2025-08-20T18:41:17.540-0200"
		time.RFC3339,                   // "2006-01-02T15:04:05Z07:00"
		"2006-01-02T15:04:05.000Z",     // "2025-08-20T18:41:17.540Z"
		"2006-01-02T15:04:05Z",         // "2025-08-20T18:41:17Z"
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", timeStr)
}

func (p *Provider) makeRequest(ctx context.Context, url string, result any) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	// JIRA uses basic auth with email and API token
	req.SetBasicAuth(p.config.Email, p.config.Token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JIRA API request failed with status %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

// GetAssignedTickets retrieves JIRA tickets assigned to the current user that are not done
func (p *Provider) GetAssignedTickets(ctx context.Context) ([]TodoItem, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("JIRA provider not configured")
	}

	// JQL query to find tickets assigned to current user that are not in done/closed states
	jql := "assignee = currentUser() AND status NOT IN (Done, Closed, Resolved)"

	// Add filter if configured
	if p.config.Filter != "" {
		jql = fmt.Sprintf("%s AND (%s)", jql, p.config.Filter)
	}

	jql = fmt.Sprintf("%s ORDER BY updated DESC", jql)

	// URL encode the JQL query
	searchURL := fmt.Sprintf("%s/rest/api/3/search?jql=%s&fields=key,summary,status,updated,assignee&maxResults=50",
		strings.TrimSuffix(p.config.URL, "/"),
		url.QueryEscape(jql))

	var searchResult struct {
		Issues []struct {
			Key    string `json:"key"`
			Fields struct {
				Summary string `json:"summary"`
				Updated string `json:"updated"`
				Status  struct {
					Name string `json:"name"`
				} `json:"status"`
			} `json:"fields"`
		} `json:"issues"`
	}

	if err := p.makeRequest(ctx, searchURL, &searchResult); err != nil {
		return nil, err
	}

	var todos []TodoItem
	for _, issue := range searchResult.Issues {
		// Parse the updated time
		updatedTime, err := p.parseJIRATime(issue.Fields.Updated)
		if err != nil {
			// Use current time if parsing fails
			updatedTime = time.Now()
		}

		todos = append(todos, TodoItem{
			ID:          fmt.Sprintf("jira-%s", issue.Key),
			Title:       fmt.Sprintf("%s: %s", issue.Key, issue.Fields.Summary),
			Description: fmt.Sprintf("Status: %s", issue.Fields.Status.Name),
			URL:         fmt.Sprintf("%s/browse/%s", strings.TrimSuffix(p.config.URL, "/"), issue.Key),
			UpdatedAt:   updatedTime,
			Tags:        []string{issue.Key, issue.Fields.Status.Name},
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
