package activity

import (
	"testing"
	"time"
)

func TestSummary_GroupByPlatform(t *testing.T) {
	date := time.Now()
	activities := []Activity{
		{ID: "1", Platform: "github", Type: ActivityTypeCommit, Title: "Fix bug", Timestamp: date},
		{ID: "2", Platform: "github", Type: ActivityTypePR, Title: "Add feature", Timestamp: date},
		{ID: "3", Platform: "jira", Type: ActivityTypeJiraTicket, Title: "PROJ-123", Timestamp: date},
		{ID: "4", Platform: "obsidian", Type: ActivityTypeNote, Title: "Meeting notes", Timestamp: date},
	}

	summary := Summary{
		Date:       date,
		Activities: activities,
	}

	groups := summary.GroupByPlatform()

	if len(groups) != 3 {
		t.Errorf("Expected 3 platform groups, got %d", len(groups))
	}

	if len(groups["github"]) != 2 {
		t.Errorf("Expected 2 GitHub activities, got %d", len(groups["github"]))
	}

	if len(groups["jira"]) != 1 {
		t.Errorf("Expected 1 JIRA activity, got %d", len(groups["jira"]))
	}

	if len(groups["obsidian"]) != 1 {
		t.Errorf("Expected 1 Obsidian activity, got %d", len(groups["obsidian"]))
	}
}

func TestSummary_GroupByType(t *testing.T) {
	date := time.Now()
	activities := []Activity{
		{ID: "1", Platform: "github", Type: ActivityTypeCommit, Title: "Fix bug", Timestamp: date},
		{ID: "2", Platform: "github", Type: ActivityTypeCommit, Title: "Add tests", Timestamp: date},
		{ID: "3", Platform: "github", Type: ActivityTypePR, Title: "Add feature", Timestamp: date},
		{ID: "4", Platform: "jira", Type: ActivityTypeJiraTicket, Title: "PROJ-123", Timestamp: date},
	}

	summary := Summary{
		Date:       date,
		Activities: activities,
	}

	groups := summary.GroupByType()

	if len(groups) != 3 {
		t.Errorf("Expected 3 activity type groups, got %d", len(groups))
	}

	if len(groups[ActivityTypeCommit]) != 2 {
		t.Errorf("Expected 2 commit activities, got %d", len(groups[ActivityTypeCommit]))
	}

	if len(groups[ActivityTypePR]) != 1 {
		t.Errorf("Expected 1 PR activity, got %d", len(groups[ActivityTypePR]))
	}

	if len(groups[ActivityTypeJiraTicket]) != 1 {
		t.Errorf("Expected 1 JIRA ticket activity, got %d", len(groups[ActivityTypeJiraTicket]))
	}
}