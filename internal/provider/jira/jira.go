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

	// Get issues updated in the time range
	issues, err := p.getUpdatedIssues(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated issues: %w", err)
	}
	activities = append(activities, issues...)

	return activities, nil
}

func (p *Provider) getUpdatedIssues(ctx context.Context, from, to time.Time) ([]activity.Activity, error) {
	// Build JQL query to find issues updated in the time range
	// Use proper from and to dates - to is exclusive so it's the next day
	jql := fmt.Sprintf("assignee = currentUser() AND updated >= \"%s\" AND updated < \"%s\" ORDER BY updated DESC",
		from.Format("2006-01-02"),
		to.Format("2006-01-02"))

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
		"2006-01-02T15:04:05.000Z0700",    // "2025-08-20T18:41:17.540+0200"
		"2006-01-02T15:04:05.000-0700",    // "2025-08-20T18:41:17.540-0200"
		time.RFC3339,                      // "2006-01-02T15:04:05Z07:00"
		"2006-01-02T15:04:05.000Z",        // "2025-08-20T18:41:17.540Z"
		"2006-01-02T15:04:05Z",            // "2025-08-20T18:41:17Z"
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JIRA API request failed with status %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(result)
}