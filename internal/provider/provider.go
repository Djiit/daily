package provider

import (
	"context"
	"fmt"
	"time"

	"daily/internal/activity"
)

// Provider defines the interface that all activity providers must implement
type Provider interface {
	// Name returns the name of the provider (e.g., "github", "jira", "obsidian")
	Name() string

	// GetActivities retrieves activities for the specified date range
	GetActivities(ctx context.Context, from, to time.Time) ([]activity.Activity, error)

	// IsConfigured returns true if the provider is properly configured
	IsConfigured() bool
}

// Config holds common configuration for providers
type Config struct {
	// Common fields that providers might need
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
	Token    string `json:"token,omitempty"`
	URL      string `json:"url,omitempty"`
	Enabled  bool   `json:"enabled"`
	Filter   string `json:"filter,omitempty"` // Additional filter string for customizing queries
}

// Aggregator collects activities from multiple providers
type Aggregator struct {
	providers []Provider
}

// NewAggregator creates a new activity aggregator
func NewAggregator(providers ...Provider) *Aggregator {
	return &Aggregator{
		providers: providers,
	}
}

// AddProvider adds a provider to the aggregator
func (a *Aggregator) AddProvider(provider Provider) {
	a.providers = append(a.providers, provider)
}

// GetSummary retrieves activities from all configured providers for the given date
func (a *Aggregator) GetSummary(ctx context.Context, date time.Time) (*activity.Summary, error) {
	// Get activities for the full day
	from := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	to := from.Add(24 * time.Hour)

	var allActivities []activity.Activity

	for _, provider := range a.providers {
		if !provider.IsConfigured() {
			continue
		}

		activities, err := provider.GetActivities(ctx, from, to)
		if err != nil {
			// Continue with other providers but could add logging here
			continue
		}

		allActivities = append(allActivities, activities...)
	}

	return &activity.Summary{
		Date:       date,
		Activities: allActivities,
	}, nil
}

// GetSummaryWithVerbose is like GetSummary but with verbose logging
func (a *Aggregator) GetSummaryWithVerbose(ctx context.Context, date time.Time, verbose bool) (*activity.Summary, error) {
	// Get activities for the full day
	from := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	to := from.Add(24 * time.Hour)

	var allActivities []activity.Activity

	for _, provider := range a.providers {
		if !provider.IsConfigured() {
			if verbose {
				fmt.Printf("⚠️  %s provider not configured, skipping\n", provider.Name())
			}
			continue
		}

		if verbose {
			fmt.Printf("🔍 Querying %s provider...\n", provider.Name())
		}

		activities, err := provider.GetActivities(ctx, from, to)
		if err != nil {
			if verbose {
				fmt.Printf("❌ %s provider failed: %v\n", provider.Name(), err)
			}
			continue
		}

		if verbose {
			fmt.Printf("✅ %s provider returned %d activities\n", provider.Name(), len(activities))
		}

		allActivities = append(allActivities, activities...)
	}

	return &activity.Summary{
		Date:       date,
		Activities: allActivities,
	}, nil
}
