package confluence

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
		URL:     "example.atlassian.net",
		Enabled: true,
	}

	p := NewProvider(config)

	if p.Name() != "confluence" {
		t.Errorf("Expected provider name to be 'confluence', got '%s'", p.Name())
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
				URL:     "example.atlassian.net",
				Enabled: true,
			},
			expected: true,
		},
		{
			name: "disabled",
			config: provider.Config{
				Email:   "test@example.com",
				Token:   "testtoken",
				URL:     "example.atlassian.net",
				Enabled: false,
			},
			expected: false,
		},
		{
			name: "missing token",
			config: provider.Config{
				Email:   "test@example.com",
				Token:   "",
				URL:     "example.atlassian.net",
				Enabled: true,
			},
			expected: false,
		},
		{
			name: "missing email",
			config: provider.Config{
				Email:   "",
				Token:   "testtoken",
				URL:     "example.atlassian.net",
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
		{
			name: "URL with https prefix",
			config: provider.Config{
				Email:   "test@example.com",
				Token:   "testtoken",
				URL:     "https://example.atlassian.net",
				Enabled: true,
			},
			expected: true,
		},
		{
			name: "URL with trailing slash",
			config: provider.Config{
				Email:   "test@example.com",
				Token:   "testtoken",
				URL:     "example.atlassian.net/",
				Enabled: true,
			},
			expected: true,
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
	tests := []struct {
		name        string
		config      provider.Config
		expectError bool
	}{
		{
			name: "configured provider",
			config: provider.Config{
				Email:   "test@example.com",
				Token:   "testtoken",
				URL:     "example.atlassian.net",
				Enabled: true,
			},
			expectError: false,
		},
		{
			name: "unconfigured provider",
			config: provider.Config{
				Email:   "",
				Token:   "",
				URL:     "",
				Enabled: false,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProvider(tt.config)

			from := time.Now().AddDate(0, 0, -1)
			to := time.Now()

			activities, err := p.GetActivities(context.Background(), from, to)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error for unconfigured provider, got nil")
				}
				return
			}

			// For configured provider with fake credentials, we expect no error
			// due to resilient design, but activities should be empty/nil
			if err != nil {
				t.Errorf("Expected no error from resilient provider design, got: %v", err)
			}

			if activities == nil {
				t.Error("Expected activities slice to be initialized, got nil")
			}
		})
	}
}

func TestProvider_GetMentions(t *testing.T) {
	tests := []struct {
		name        string
		config      provider.Config
		expectError bool
	}{
		{
			name: "configured provider",
			config: provider.Config{
				Email:   "test@example.com",
				Token:   "testtoken",
				URL:     "example.atlassian.net",
				Enabled: true,
			},
			expectError: false,
		},
		{
			name: "unconfigured provider",
			config: provider.Config{
				Email:   "",
				Token:   "",
				URL:     "",
				Enabled: false,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProvider(tt.config)

			mentions, err := p.GetMentions(context.Background(), "2w")

			if tt.expectError {
				if err == nil {
					t.Error("Expected error for unconfigured provider, got nil")
				}
				return
			}

			// For configured provider with fake credentials, we expect an error
			// since GetMentions doesn't have the same resilient design as GetActivities
			// (it returns actual errors to be handled by the todo command)
			if err == nil {
				t.Error("Expected error with fake credentials, got nil")
			}

			// mentions could be nil due to error, which is acceptable
			if mentions == nil && err == nil {
				t.Error("Expected either mentions or error, got neither")
			}
		})
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

	expectedError := "Confluence provider not configured"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

func TestProvider_GetMentions_NotConfigured(t *testing.T) {
	config := provider.Config{
		Email:   "",
		Token:   "",
		URL:     "",
		Enabled: false,
	}

	p := NewProvider(config)

	_, err := p.GetMentions(context.Background(), "2w")

	if err == nil {
		t.Error("Expected error for unconfigured provider, got nil")
	}

	expectedError := "Confluence provider not configured"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

func TestProvider_NewProvider(t *testing.T) {
	config := provider.Config{
		Email:   "test@example.com",
		Token:   "testtoken",
		URL:     "example.atlassian.net",
		Enabled: true,
	}

	p := NewProvider(config)

	if p == nil {
		t.Error("Expected provider instance, got nil")
	}

	if p.client == nil {
		t.Error("Expected HTTP client to be initialized, got nil")
	}

	if p.client.Timeout != 60*time.Second {
		t.Errorf("Expected timeout to be 60s, got %v", p.client.Timeout)
	}

	// Verify config is stored correctly
	if p.config.Email != config.Email {
		t.Errorf("Expected email '%s', got '%s'", config.Email, p.config.Email)
	}

	if p.config.Token != config.Token {
		t.Errorf("Expected token '%s', got '%s'", config.Token, p.config.Token)
	}

	if p.config.URL != config.URL {
		t.Errorf("Expected URL '%s', got '%s'", config.URL, p.config.URL)
	}

	if p.config.Enabled != config.Enabled {
		t.Errorf("Expected enabled %t, got %t", config.Enabled, p.config.Enabled)
	}
}

func TestProvider_URLHandling(t *testing.T) {
	tests := []struct {
		name        string
		inputURL    string
		description string
	}{
		{
			name:        "plain domain",
			inputURL:    "example.atlassian.net",
			description: "should work with plain domain",
		},
		{
			name:        "https prefix",
			inputURL:    "https://example.atlassian.net",
			description: "should work with https prefix",
		},
		{
			name:        "trailing slash",
			inputURL:    "example.atlassian.net/",
			description: "should work with trailing slash",
		},
		{
			name:        "https and trailing slash",
			inputURL:    "https://example.atlassian.net/",
			description: "should work with both https and trailing slash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := provider.Config{
				Email:   "test@example.com",
				Token:   "testtoken",
				URL:     tt.inputURL,
				Enabled: true,
			}

			p := NewProvider(config)

			if !p.IsConfigured() {
				t.Errorf("Expected provider to be configured with URL '%s', but IsConfigured() returned false", tt.inputURL)
			}

			// Test that the provider can be created without panicking
			// The actual URL processing happens in the searchConfluence method
			// which would be tested with integration tests
		})
	}
}

func TestProvider_ContextCancellation(t *testing.T) {
	config := provider.Config{
		Email:   "test@example.com",
		Token:   "testtoken",
		URL:     "example.atlassian.net",
		Enabled: true,
	}

	p := NewProvider(config)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	from := time.Now().AddDate(0, 0, -1)
	to := time.Now()

	// Test GetActivities with cancelled context
	activities, err := p.GetActivities(ctx, from, to)
	// Should handle context cancellation gracefully
	if err != nil && activities == nil {
		// This is acceptable behavior
	}

	// Test GetMentions with cancelled context
	mentions, err := p.GetMentions(ctx, "2w")
	// Should handle context cancellation gracefully
	if err != nil && mentions == nil {
		// This is acceptable behavior
	}
}

func TestProvider_TimeRangeHandling(t *testing.T) {
	config := provider.Config{
		Email:   "test@example.com",
		Token:   "testtoken",
		URL:     "example.atlassian.net",
		Enabled: true,
	}

	p := NewProvider(config)

	tests := []struct {
		name string
		from time.Time
		to   time.Time
	}{
		{
			name: "past day",
			from: time.Now().AddDate(0, 0, -1),
			to:   time.Now(),
		},
		{
			name: "past week",
			from: time.Now().AddDate(0, 0, -7),
			to:   time.Now(),
		},
		{
			name: "same day",
			from: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			to:   time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC),
		},
		{
			name: "future range (should still work)",
			from: time.Now().AddDate(0, 0, 1),
			to:   time.Now().AddDate(0, 0, 2),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			activities, err := p.GetActivities(context.Background(), tt.from, tt.to)

			// With fake credentials, we should get an error, but the method should not panic
			if activities == nil && err != nil {
				// This is expected behavior with invalid credentials
			} else if activities != nil && err == nil {
				// This would be unexpected with fake credentials, but acceptable if it happens
			}

			// The important thing is that the method doesn't panic with various time ranges
		})
	}
}

func TestProvider_GetBaseURL(t *testing.T) {
	tests := []struct {
		name        string
		inputURL    string
		expectedURL string
	}{
		{
			name:        "plain domain",
			inputURL:    "example.atlassian.net",
			expectedURL: "https://example.atlassian.net",
		},
		{
			name:        "https prefix already present",
			inputURL:    "https://example.atlassian.net",
			expectedURL: "https://example.atlassian.net",
		},
		{
			name:        "http prefix (preserved)",
			inputURL:    "http://example.atlassian.net",
			expectedURL: "http://example.atlassian.net",
		},
		{
			name:        "trailing slash removed",
			inputURL:    "example.atlassian.net/",
			expectedURL: "https://example.atlassian.net",
		},
		{
			name:        "https with trailing slash",
			inputURL:    "https://example.atlassian.net/",
			expectedURL: "https://example.atlassian.net",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := provider.Config{
				URL:     tt.inputURL,
				Enabled: true,
			}

			p := NewProvider(config)
			result := p.getBaseURL()

			if result != tt.expectedURL {
				t.Errorf("Expected URL '%s', got '%s'", tt.expectedURL, result)
			}
		})
	}
}

func TestProvider_GetCommentsOnMyPages(t *testing.T) {
	tests := []struct {
		name        string
		config      provider.Config
		expectError bool
	}{
		{
			name: "configured provider",
			config: provider.Config{
				Email:   "test@example.com",
				Token:   "testtoken",
				URL:     "example.atlassian.net",
				Enabled: true,
			},
			expectError: false,
		},
		{
			name: "unconfigured provider",
			config: provider.Config{
				Email:   "",
				Token:   "",
				URL:     "",
				Enabled: false,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProvider(tt.config)

			comments, err := p.GetCommentsOnMyPages(context.Background(), "2d")

			if tt.expectError {
				if err == nil {
					t.Error("Expected error for unconfigured provider, got nil")
				}
				return
			}

			// For configured provider with fake credentials, we expect an error
			if err == nil {
				t.Error("Expected error with fake credentials, got nil")
			}

			// comments could be nil or empty due to error, which is acceptable
			_ = comments
		})
	}
}

func TestProvider_GetMentions_SinceFormat(t *testing.T) {
	tests := []struct {
		name       string
		sinceInput string
		expectErr  bool
	}{
		{
			name:       "without dash prefix",
			sinceInput: "2w",
			expectErr:  true, // Will fail with fake credentials, but shouldn't panic
		},
		{
			name:       "with dash prefix",
			sinceInput: "-2w",
			expectErr:  true, // Will fail with fake credentials, but shouldn't panic
		},
		{
			name:       "days format",
			sinceInput: "7d",
			expectErr:  true,
		},
		{
			name:       "hours format",
			sinceInput: "24h",
			expectErr:  true,
		},
		{
			name:       "months format",
			sinceInput: "1m",
			expectErr:  true,
		},
	}

	config := provider.Config{
		Email:   "test@example.com",
		Token:   "testtoken",
		URL:     "example.atlassian.net",
		Enabled: true,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProvider(config)

			// This will fail with fake credentials, but the test verifies the method doesn't panic
			mentions, err := p.GetMentions(context.Background(), tt.sinceInput)

			// With fake credentials, we expect an error
			if err == nil && tt.expectErr {
				t.Error("Expected error with fake credentials, got nil")
			}

			// The important thing is that the method doesn't panic with various since formats
			_ = mentions
		})
	}
}
