package output

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss/v2"

	"daily/internal/activity"
)

type Formatter struct {
	// Styles for different components
	titleStyle       lipgloss.Style
	headerStyle      lipgloss.Style
	platformStyle    lipgloss.Style
	activityStyle    lipgloss.Style
	timeStyle        lipgloss.Style
	descriptionStyle lipgloss.Style
	urlStyle         lipgloss.Style
	tagStyle         lipgloss.Style
	borderStyle      lipgloss.Style
}

func NewFormatter() *Formatter {
	return &Formatter{
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1),
		headerStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			MarginTop(1).
			MarginBottom(1),
		platformStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			PaddingLeft(1).
			PaddingRight(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("86")),
		activityStyle: lipgloss.NewStyle().
			PaddingLeft(2).
			PaddingTop(1).
			MarginBottom(1),
		timeStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Bold(true),
		descriptionStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			PaddingLeft(5).
			Italic(true),
		urlStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("33")).
			PaddingLeft(5).
			Underline(true),
		tagStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("208")).
			PaddingLeft(5).
			Italic(true),
		borderStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
	}
}

func (f *Formatter) FormatSummary(summary *activity.Summary) string {
	if len(summary.Activities) == 0 {
		return f.headerStyle.Render("No activities found for this date.")
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

	// Title with styling
	title := fmt.Sprintf("📊 Daily Summary for %s", summary.Date.Format("January 2, 2006"))
	output.WriteString(f.titleStyle.Render(title))
	output.WriteString("\n")

	// Summary stats
	stats := fmt.Sprintf("Found %d activities across %d platforms", len(activities), len(groups))
	output.WriteString(f.headerStyle.Render(stats))
	output.WriteString("\n\n")

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

	// Platform header with icon and styling
	icon := f.getPlatformIcon(platform)
	platformHeader := fmt.Sprintf("%s %s (%d)", icon, strings.Title(platform), len(activities))
	section.WriteString(f.platformStyle.Render(platformHeader))
	section.WriteString("\n")

	// Styled border
	border := strings.Repeat("─", 60)
	section.WriteString(f.borderStyle.Render(border))
	section.WriteString("\n")

	for _, act := range activities {
		section.WriteString(f.formatActivity(act))
	}

	section.WriteString("\n")
	return section.String()
}

func (f *Formatter) formatActivity(act activity.Activity) string {
	var activityContent strings.Builder

	// Time and type with styling
	timeStr := f.timeStyle.Render(act.Timestamp.Format("15:04"))
	typeIcon := f.getTypeIcon(act.Type)

	// Main activity line
	mainLine := fmt.Sprintf("%s %s  %s", timeStr, typeIcon, act.Title)
	activityContent.WriteString(mainLine)
	activityContent.WriteString("\n")

	if act.Description != "" {
		description := f.descriptionStyle.Render(act.Description)
		activityContent.WriteString(description)
		activityContent.WriteString("\n")
	}

	if act.URL != "" {
		url := f.urlStyle.Render("🔗 " + act.URL)
		activityContent.WriteString(url)
		activityContent.WriteString("\n")
	}

	if len(act.Tags) > 0 {
		tags := f.tagStyle.Render("🏷️  " + strings.Join(act.Tags, ", "))
		activityContent.WriteString(tags)
		activityContent.WriteString("\n")
	}

	// Wrap the entire activity in the activity style
	return f.activityStyle.Render(activityContent.String())
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
		return f.headerStyle.Render("No activities found for this date.")
	}

	var output strings.Builder

	// Sort activities by timestamp
	activities := make([]activity.Activity, len(summary.Activities))
	copy(activities, summary.Activities)
	sort.Slice(activities, func(i, j int) bool {
		return activities[i].Timestamp.Before(activities[j].Timestamp)
	})

	// Header with styling
	header := fmt.Sprintf("Daily Summary - %d activities:", len(activities))
	output.WriteString(f.titleStyle.Render(header))
	output.WriteString("\n\n")

	for _, act := range activities {
		timeStr := f.timeStyle.Render(act.Timestamp.Format("15:04"))
		platformIcon := f.getPlatformIcon(act.Platform)
		typeIcon := f.getTypeIcon(act.Type)
		platformStr := fmt.Sprintf("%s %s", platformIcon, act.Platform)
		output.WriteString(fmt.Sprintf("%s %s %s %s\n", timeStr, typeIcon, platformStr, act.Title))
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
			Total      int            `json:"total"`
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
