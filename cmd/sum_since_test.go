package cmd

import (
	"testing"
	"time"
)

func TestParseSinceDuration(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{
			name:      "1 hour",
			input:     "1h",
			wantError: false,
		},
		{
			name:      "24 hours",
			input:     "24h",
			wantError: false,
		},
		{
			name:      "1 day",
			input:     "1d",
			wantError: false,
		},
		{
			name:      "7 days",
			input:     "7d",
			wantError: false,
		},
		{
			name:      "1 week",
			input:     "1w",
			wantError: false,
		},
		{
			name:      "2 weeks",
			input:     "2w",
			wantError: false,
		},
		{
			name:      "1 month",
			input:     "1m",
			wantError: false,
		},
		{
			name:      "invalid format - no unit",
			input:     "5",
			wantError: true,
		},
		{
			name:      "invalid format - no number",
			input:     "d",
			wantError: true,
		},
		{
			name:      "invalid unit",
			input:     "1x",
			wantError: true,
		},
		{
			name:      "empty string",
			input:     "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseSinceDuration(tt.input)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error for input '%s', got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for input '%s': %v", tt.input, err)
				return
			}

			// Verify result is in the past
			now := time.Now()
			if result.After(now) {
				t.Errorf("Expected result to be in the past, but got %v (now is %v)", result, now)
			}

			// Verify result is reasonable (within expected range)
			// This is a sanity check, not an exact duration check
			diff := now.Sub(result)
			if diff < 0 {
				t.Errorf("Expected positive duration, got %v", diff)
			}
		})
	}
}

func TestParseSinceDuration_Accuracy(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  time.Duration
		tolerance time.Duration
	}{
		{
			name:      "1 hour",
			input:     "1h",
			expected:  1 * time.Hour,
			tolerance: 1 * time.Second,
		},
		{
			name:      "1 day",
			input:     "1d",
			expected:  24 * time.Hour,
			tolerance: 1 * time.Second,
		},
		{
			name:      "1 week",
			input:     "1w",
			expected:  7 * 24 * time.Hour,
			tolerance: 1 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			result, err := parseSinceDuration(tt.input)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Calculate the actual duration
			actualDuration := before.Sub(result)

			// Check if the duration is within tolerance
			diff := actualDuration - tt.expected
			if diff < 0 {
				diff = -diff
			}

			if diff > tt.tolerance {
				t.Errorf("Duration mismatch for '%s': expected %v, got %v (diff: %v)", tt.input, tt.expected, actualDuration, diff)
			}

			// Verify the result is in the past
			if result.After(before) {
				t.Errorf("Result should be before function call time")
			}
		})
	}
}
