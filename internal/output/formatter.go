package output

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"daily/internal/activity"
)

type Formatter struct{}

func NewFormatter() *Formatter {
	return &Formatter{}
}

func (f *Formatter) FormatSummary(summary *activity.Summary) string {
	if len(summary.Activities) == 0 {
		return "No activities found for this date."
	}

	var output strings.Builder
	
	// Sort activities by timestamp
	activities := make([]activity.Activity, len(summary.Activities))
	copy(activities, summary.Activities)
	sort.Slice(activities, func(i, j int) bool {
		return activities[i].Timestamp.Before(activities[j].Timestamp)
	})

	// Group by platform
	groups := make(map[string][]activity.Activity)
	for _, act := range activities {
		groups[act.Platform] = append(groups[act.Platform], act)
	}

	output.WriteString(fmt.Sprintf("📊 Daily Summary for %s\n", summary.Date.Format("January 2, 2006")))
	output.WriteString(fmt.Sprintf("Found %d activities across %d platforms\n\n", len(activities), len(groups)))

	// Display by platform
	platforms := []string{"github", "jira", "obsidian"}
	for _, platform := range platforms {
		platformActivities, exists := groups[platform]
		if !exists || len(platformActivities) == 0 {
			continue
		}

		output.WriteString(f.formatPlatformSection(platform, platformActivities))
	}

	// Add any other platforms not in the main list
	for platform, platformActivities := range groups {
		if platform != "github" && platform != "jira" && platform != "obsidian" {
			output.WriteString(f.formatPlatformSection(platform, platformActivities))
		}
	}

	return output.String()
}

func (f *Formatter) formatPlatformSection(platform string, activities []activity.Activity) string {
	var section strings.Builder
	
	// Platform header with icon
	icon := f.getPlatformIcon(platform)
	section.WriteString(fmt.Sprintf("%s %s (%d)\n", icon, strings.Title(platform), len(activities)))
	section.WriteString(strings.Repeat("─", 50) + "\n")

	for _, act := range activities {
		section.WriteString(f.formatActivity(act))
		section.WriteString("\n")
	}

	section.WriteString("\n")
	return section.String()
}

func (f *Formatter) formatActivity(act activity.Activity) string {
	var output strings.Builder
	
	// Time and type
	timeStr := act.Timestamp.Format("15:04")
	typeIcon := f.getTypeIcon(act.Type)
	
	output.WriteString(fmt.Sprintf("  %s %s  %s\n", timeStr, typeIcon, act.Title))
	
	if act.Description != "" {
		output.WriteString(fmt.Sprintf("     %s\n", act.Description))
	}
	
	if act.URL != "" {
		output.WriteString(fmt.Sprintf("     🔗 %s\n", act.URL))
	}
	
	if len(act.Tags) > 0 {
		output.WriteString(fmt.Sprintf("     🏷️  %s\n", strings.Join(act.Tags, ", ")))
	}
	
	return output.String()
}

func (f *Formatter) getPlatformIcon(platform string) string {
	icons := map[string]string{
		"github":   "🐙",
		"jira":     "🎫",
		"obsidian": "📝",
	}
	
	if icon, exists := icons[platform]; exists {
		return icon
	}
	return "📌"
}

func (f *Formatter) getTypeIcon(actType activity.ActivityType) string {
	icons := map[activity.ActivityType]string{
		activity.ActivityTypeCommit:     "💾",
		activity.ActivityTypePR:         "🔀",
		activity.ActivityTypeIssue:      "🐛",
		activity.ActivityTypeJiraTicket: "🎯",
		activity.ActivityTypeNote:       "📄",
	}
	
	if icon, exists := icons[actType]; exists {
		return icon
	}
	return "📋"
}

func (f *Formatter) FormatCompactSummary(summary *activity.Summary) string {
	if len(summary.Activities) == 0 {
		return "No activities found for this date."
	}

	var output strings.Builder
	
	// Sort activities by timestamp
	activities := make([]activity.Activity, len(summary.Activities))
	copy(activities, summary.Activities)
	sort.Slice(activities, func(i, j int) bool {
		return activities[i].Timestamp.Before(activities[j].Timestamp)
	})

	output.WriteString(fmt.Sprintf("Daily Summary - %d activities:\n\n", len(activities)))

	for _, act := range activities {
		timeStr := act.Timestamp.Format("15:04")
		output.WriteString(fmt.Sprintf("%s [%s] %s\n", timeStr, act.Platform, act.Title))
	}

	return output.String()
}

func (f *Formatter) FormatJSON(summary *activity.Summary) string {
	// Sort activities by timestamp for consistent output
	activities := make([]activity.Activity, len(summary.Activities))
	copy(activities, summary.Activities)
	sort.Slice(activities, func(i, j int) bool {
		return activities[i].Timestamp.Before(activities[j].Timestamp)
	})

	// Create a JSON-friendly structure
	jsonOutput := struct {
		Date       string              `json:"date"`
		Activities []activity.Activity `json:"activities"`
		Summary    struct {
			Total     int            `json:"total"`
			ByPlatform map[string]int `json:"by_platform"`
			ByType     map[string]int `json:"by_type"`
		} `json:"summary"`
	}{
		Date:       summary.Date.Format("2006-01-02"),
		Activities: activities,
	}

	// Calculate summary statistics
	jsonOutput.Summary.Total = len(activities)
	jsonOutput.Summary.ByPlatform = make(map[string]int)
	jsonOutput.Summary.ByType = make(map[string]int)

	for _, act := range activities {
		jsonOutput.Summary.ByPlatform[act.Platform]++
		jsonOutput.Summary.ByType[string(act.Type)]++
	}

	// Marshal to JSON with proper indentation
	jsonBytes, err := json.MarshalIndent(jsonOutput, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "Failed to marshal JSON: %s"}`, err.Error())
	}

	return string(jsonBytes) + "\n"
}