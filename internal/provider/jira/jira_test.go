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

func TestProvider_IsConfigured_WithFilter(t *testing.T) {
	config := provider.Config{
		Email:   "test@example.com",
		Token:   "testtoken",
		URL:     "https://example.atlassian.net",
		Enabled: true,
		Filter:  "project = PROJ",
	}

	p := NewProvider(config)

	if !p.IsConfigured() {
		t.Error("Expected provider to be configured with filter")
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

func TestProvider_GetAssignedTickets(t *testing.T) {
	tests := []struct {
		name           string
		config         provider.Config
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "unconfigured provider",
			config: provider.Config{
				Email:   "",
				Token:   "",
				URL:     "",
				Enabled: false,
			},
			expectError:    true,
			expectedErrMsg: "JIRA provider not configured",
		},
		{
			name: "missing URL",
			config: provider.Config{
				Email:   "test@example.com",
				Token:   "testtoken",
				URL:     "",
				Enabled: true,
			},
			expectError:    true,
			expectedErrMsg: "JIRA provider not configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProvider(tt.config)

			todos, err := p.GetAssignedTickets(context.Background())

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if tt.expectedErrMsg != "" && err.Error() != tt.expectedErrMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.expectedErrMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if todos == nil {
					t.Error("Expected non-nil todos slice")
				}
			}
		})
	}
}

func TestProvider_ParseJIRATime(t *testing.T) {
	config := provider.Config{
		Email:   "test@example.com",
		Token:   "testtoken",
		URL:     "https://example.atlassian.net",
		Enabled: true,
	}

	p := NewProvider(config)

	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{
			name:      "RFC3339 format",
			input:     "2023-12-25T10:30:00Z",
			expectErr: false,
		},
		{
			name:      "JIRA format with timezone",
			input:     "2023-12-25T10:30:00.000+0200",
			expectErr: false,
		},
		{
			name:      "JIRA format with milliseconds",
			input:     "2023-12-25T10:30:00.540Z",
			expectErr: false,
		},
		{
			name:      "invalid format",
			input:     "not-a-date",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.parseJIRATime(tt.input)

			if tt.expectErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}
