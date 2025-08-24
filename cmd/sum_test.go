package cmd

import (
	"testing"
	"time"
)

func TestParseDate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		input    string
		expected time.Time
		hasError bool
	}{
		{
			input:    "today",
			expected: now,
			hasError: false,
		},
		{
			input:    "yesterday",
			expected: now.AddDate(0, 0, -1),
			hasError: false,
		},
		{
			input:    "2023-12-25",
			expected: time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC),
			hasError: false,
		},
		{
			input:    "invalid-date",
			expected: time.Time{},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseDate(tt.input)

			if tt.hasError && err == nil {
				t.Errorf("Expected error for input %s, but got none", tt.input)
			}

			if !tt.hasError && err != nil {
				t.Errorf("Expected no error for input %s, but got: %v", tt.input, err)
			}

			if !tt.hasError {
				// For relative dates, compare only the date part
				if tt.input == "today" || tt.input == "yesterday" {
					if result.Year() != tt.expected.Year() || result.Month() != tt.expected.Month() || result.Day() != tt.expected.Day() {
						t.Errorf("Expected date %s, got %s", tt.expected.Format("2006-01-02"), result.Format("2006-01-02"))
					}
				} else {
					if !result.Equal(tt.expected) {
						t.Errorf("Expected %v, got %v", tt.expected, result)
					}
				}
			}
		})
	}
}
