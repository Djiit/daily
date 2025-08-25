package github

import (
	"context"
	"testing"
	"time"

	"daily/internal/provider"
)

func TestProvider_Name(t *testing.T) {
	config := provider.Config{
		Username: "testuser",
		Token:    "testtoken",
		Enabled:  true,
	}

	p := NewProvider(config)

	if p.Name() != "github" {
		t.Errorf("Expected provider name to be 'github', got '%s'", p.Name())
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
				Username: "testuser",
				Token:    "testtoken",
				Enabled:  true,
			},
			expected: true,
		},
		{
			name: "disabled",
			config: provider.Config{
				Username: "testuser",
				Token:    "testtoken",
				Enabled:  false,
			},
			expected: false,
		},
		{
			name: "missing token",
			config: provider.Config{
				Username: "testuser",
				Token:    "",
				Enabled:  true,
			},
			expected: false,
		},
		{
			name: "missing username",
			config: provider.Config{
				Username: "",
				Token:    "testtoken",
				Enabled:  true,
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
		Username: "testuser",
		Token:    "testtoken",
		Enabled:  true,
	}

	p := NewProvider(config)

	from := time.Now().AddDate(0, 0, -1)
	to := time.Now()

	// This test will fail with real API calls due to authentication
	// In a real scenario, we'd need either a valid token or a mock HTTP client
	activities, err := p.GetActivities(context.Background(), from, to)

	// The provider is designed to be resilient and return empty results instead of errors
	// This allows the aggregator to continue with other providers even if one fails
	if err != nil {
		t.Errorf("Expected no error from resilient provider design, got: %v", err)
	}

	// With fake credentials, we should get empty activities
	if activities == nil {
		t.Error("Expected empty activities slice, got nil")
	}
}

func TestProvider_GetActivities_NotConfigured(t *testing.T) {
	config := provider.Config{
		Username: "",
		Token:    "",
		Enabled:  false,
	}

	p := NewProvider(config)

	from := time.Now().AddDate(0, 0, -1)
	to := time.Now()

	_, err := p.GetActivities(context.Background(), from, to)

	if err == nil {
		t.Error("Expected error for unconfigured provider, got nil")
	}
}

func TestProvider_GetOpenPRs(t *testing.T) {
	tests := []struct {
		name           string
		config         provider.Config
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "configured provider",
			config: provider.Config{
				Username: "testuser",
				Token:    "testtoken",
				Enabled:  true,
			},
			expectError: true, // Will fail with fake credentials but should not panic
		},
		{
			name: "unconfigured provider",
			config: provider.Config{
				Username: "",
				Token:    "",
				Enabled:  false,
			},
			expectError:    true,
			expectedErrMsg: "GitHub provider not configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProvider(tt.config)

			todos, err := p.GetOpenPRs(context.Background())

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

func TestProvider_IsConfigured_WithFilter(t *testing.T) {
	config := provider.Config{
		Username: "testuser",
		Token:    "testtoken",
		Enabled:  true,
		Filter:   "repo:myorg/myrepo",
	}

	p := NewProvider(config)

	if !p.IsConfigured() {
		t.Error("Expected provider to be configured with filter")
	}
}

func TestProvider_GetPendingReviews(t *testing.T) {
	tests := []struct {
		name           string
		config         provider.Config
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "configured provider",
			config: provider.Config{
				Username: "testuser",
				Token:    "testtoken",
				Enabled:  true,
			},
			expectError: true, // Will fail with fake credentials but should not panic
		},
		{
			name: "unconfigured provider",
			config: provider.Config{
				Username: "",
				Token:    "",
				Enabled:  false,
			},
			expectError:    true,
			expectedErrMsg: "GitHub provider not configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProvider(tt.config)

			todos, err := p.GetPendingReviews(context.Background())

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
