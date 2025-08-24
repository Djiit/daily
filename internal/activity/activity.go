package activity

import (
	"time"
)

// ActivityType represents different types of work activities
type ActivityType string

const (
	ActivityTypeCommit     ActivityType = "commit"
	ActivityTypePR         ActivityType = "pull_request"
	ActivityTypeIssue      ActivityType = "issue"
	ActivityTypeJiraTicket ActivityType = "jira_ticket"
	ActivityTypeNote       ActivityType = "note"
)

// Activity represents a single work activity
type Activity struct {
	ID          string       `json:"id"`
	Type        ActivityType `json:"type"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	URL         string       `json:"url,omitempty"`
	Platform    string       `json:"platform"`
	Timestamp   time.Time    `json:"timestamp"`
	Tags        []string     `json:"tags,omitempty"`
}

// Summary represents a collection of activities for a specific date
type Summary struct {
	Date       time.Time  `json:"date"`
	Activities []Activity `json:"activities"`
}

// GroupByPlatform groups activities by their platform
func (s *Summary) GroupByPlatform() map[string][]Activity {
	groups := make(map[string][]Activity)
	for _, activity := range s.Activities {
		groups[activity.Platform] = append(groups[activity.Platform], activity)
	}
	return groups
}

// GroupByType groups activities by their type
func (s *Summary) GroupByType() map[ActivityType][]Activity {
	groups := make(map[ActivityType][]Activity)
	for _, activity := range s.Activities {
		groups[activity.Type] = append(groups[activity.Type], activity)
	}
	return groups
}
