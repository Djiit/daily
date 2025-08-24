package jira

import (
	"context"
	"testing"
	"time"

	"daily/internal/provider"
)

func TestProvider_Name(t *testing.T) {
	config := provider.Config{
		Email:   "test@example.com",
		Token:   "testtoken",
		URL:     "https://example.atlassian.net",
		Enabled: true,
	}
	
	p := NewProvider(config)
	
	if p.Name() != "jira" {
		t.Errorf("Expected provider name to be 'jira', got '%s'", p.Name())
	}
}

func TestProvider_IsConfigured(t *testing.T) {
	tests := []struct {
		name     string
		config   provider.Config
		expected bool
	}{
		{
			name: "fully configured",
			config: provider.Config{
				Email:   "test@example.com",
				Token:   "testtoken",
				URL:     "https://example.atlassian.net",
				Enabled: true,
			},
			expected: true,
		},
		{
			name: "disabled",
			config: provider.Config{
				Email:   "test@example.com",
				Token:   "testtoken",
				URL:     "https://example.atlassian.net",
				Enabled: false,
			},
			expected: false,
		},
		{
			name: "missing token",
			config: provider.Config{
				Email:   "test@example.com",
				Token:   "",
				URL:     "https://example.atlassian.net",
				Enabled: true,
			},
			expected: false,
		},
		{
			name: "missing email",
			config: provider.Config{
				Email:   "",
				Token:   "testtoken",
				URL:     "https://example.atlassian.net",
				Enabled: true,
			},
			expected: false,
		},
		{
			name: "missing URL",
			config: provider.Config{
				Email:   "test@example.com",
				Token:   "testtoken",
				URL:     "",
				Enabled: true,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProvider(tt.config)
			
			if p.IsConfigured() != tt.expected {
				t.Errorf("Expected IsConfigured() to return %t, got %t", tt.expected, p.IsConfigured())
			}
		})
	}
}

func TestProvider_GetActivities(t *testing.T) {
	config := provider.Config{
		Email:   "test@example.com",
		Token:   "testtoken",
		URL:     "https://example.atlassian.net",
		Enabled: true,
	}
	
	p := NewProvider(config)
	
	from := time.Now().AddDate(0, 0, -1)
	to := time.Now()
	
	activities, err := p.GetActivities(context.Background(), from, to)
	
	// Since we're not making real API calls yet, we expect no error and empty activities
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	
	if activities == nil {
		t.Error("Expected activities slice to be initialized, got nil")
	}
	
	if len(activities) != 0 {
		t.Errorf("Expected 0 activities (placeholder implementation), got %d", len(activities))
	}
}

func TestProvider_GetActivities_NotConfigured(t *testing.T) {
	config := provider.Config{
		Email:   "",
		Token:   "",
		URL:     "",
		Enabled: false,
	}
	
	p := NewProvider(config)
	
	from := time.Now().AddDate(0, 0, -1)
	to := time.Now()
	
	_, err := p.GetActivities(context.Background(), from, to)
	
	if err == nil {
		t.Error("Expected error for unconfigured provider, got nil")
	}
}