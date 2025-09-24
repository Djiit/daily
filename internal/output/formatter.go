package output

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	catppuccin "github.com/catppuccin/go"
	"github.com/charmbracelet/lipgloss/v2"

	"daily/internal/activity"
	"daily/internal/tui"
	"daily/internal/tui/types"
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

// isDarkMode detects if the terminal is using a dark theme
func isDarkMode() bool {
	// Check for explicit dark mode environment variables
	if theme := os.Getenv("THEME"); theme == "dark" {
		return true
	}
	if theme := os.Getenv("TERMINAL_THEME"); theme == "dark" {
		return true
	}

	// Check environment variables that indicate dark mode
	if colorScheme := os.Getenv("COLORFGBG"); colorScheme != "" {
		// COLORFGBG format is usually "foreground;background"
		// Dark themes typically have light foreground on dark background
		parts := strings.Split(colorScheme, ";")
		if len(parts) >= 2 {
			bg := parts[len(parts)-1]
			// Background colors like 0-7 (especially 0, 1, 8) indicate dark themes
			return bg == "0" || bg == "1" || bg == "8"
		}
	}

	// Default to light mode if we can't determine
	return false
}

func NewFormatter() *Formatter {
	isDark := isDarkMode()

	if isDark {
		// Use Catppuccin Mocha colors for dark mode
		mocha := catppuccin.Mocha
		return &Formatter{
			titleStyle: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(mocha.Mauve().Hex)).
				MarginBottom(1),
			headerStyle: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(mocha.Blue().Hex)).
				MarginTop(1).
				MarginBottom(1),
			platformStyle: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(mocha.Green().Hex)).
				PaddingLeft(1).
				PaddingRight(1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(mocha.Green().Hex)),
			activityStyle: lipgloss.NewStyle().
				PaddingLeft(2).
				PaddingTop(1).
				MarginBottom(1),
			timeStyle: lipgloss.NewStyle().
				Foreground(lipgloss.Color(mocha.Subtext1().Hex)).
				Bold(true),
			descriptionStyle: lipgloss.NewStyle().
				Foreground(lipgloss.Color(mocha.Subtext0().Hex)).
				PaddingLeft(5).
				Italic(true),
			urlStyle: lipgloss.NewStyle().
				Foreground(lipgloss.Color(mocha.Sapphire().Hex)).
				PaddingLeft(5).
				Underline(true),
			tagStyle: lipgloss.NewStyle().
				Foreground(lipgloss.Color(mocha.Peach().Hex)).
				PaddingLeft(5).
				Italic(true),
			borderStyle: lipgloss.NewStyle().
				Foreground(lipgloss.Color(mocha.Surface2().Hex)),
		}
	}

	// Light mode colors (default - Catppuccin Latte)
	latte := catppuccin.Latte
	return &Formatter{
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(latte.Mauve().Hex)).
			MarginBottom(1),
		headerStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(latte.Blue().Hex)).
			MarginTop(1).
			MarginBottom(1),
		platformStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(latte.Green().Hex)).
			PaddingLeft(1).
			PaddingRight(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(latte.Green().Hex)),
		activityStyle: lipgloss.NewStyle().
			PaddingLeft(2).
			PaddingTop(1).
			MarginBottom(1),
		timeStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(latte.Subtext1().Hex)).
			Bold(true),
		descriptionStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(latte.Subtext0().Hex)).
			PaddingLeft(5).
			Italic(true),
		urlStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(latte.Sapphire().Hex)).
			PaddingLeft(5).
			Underline(true),
		tagStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(latte.Peach().Hex)).
			PaddingLeft(5).
			Italic(true),
		borderStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(latte.Surface2().Hex)),
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

// FormatTodo formats todo items for text output
func (f *Formatter) FormatTodo(todoItems TodoItems) string {
	var output strings.Builder

	// Title
	title := "📋 Todo Items"
	output.WriteString(f.titleStyle.Render(title))
	output.WriteString("\n")

	totalItems := len(todoItems.GitHub.OpenPRs) + len(todoItems.GitHub.PendingReviews) + len(todoItems.JIRA.AssignedTickets) + len(todoItems.Obsidian.Tasks)
	if totalItems == 0 {
		output.WriteString(f.headerStyle.Render("No pending items found."))
		output.WriteString("\n")
		return output.String()
	}

	stats := fmt.Sprintf("Found %d pending items", totalItems)
	output.WriteString(f.headerStyle.Render(stats))
	output.WriteString("\n\n")

	// GitHub Open PRs
	if len(todoItems.GitHub.OpenPRs) > 0 {
		output.WriteString(f.formatTodoSection("🐙 Open Pull Requests", todoItems.GitHub.OpenPRs))
	}

	// GitHub Pending Reviews
	if len(todoItems.GitHub.PendingReviews) > 0 {
		output.WriteString(f.formatTodoSection("👁️ Pending Reviews", todoItems.GitHub.PendingReviews))
	}

	// JIRA Assigned Tickets
	if len(todoItems.JIRA.AssignedTickets) > 0 {
		output.WriteString(f.formatTodoSection("🎫 Assigned Tickets", todoItems.JIRA.AssignedTickets))
	}

	// Obsidian Tasks
	if len(todoItems.Obsidian.Tasks) > 0 {
		output.WriteString(f.formatTodoSection("📝 Obsidian Tasks", todoItems.Obsidian.Tasks))
	}

	return output.String()
}

func (f *Formatter) formatTodoSection(sectionTitle string, items []TodoItem) string {
	var section strings.Builder

	// Section header
	section.WriteString(f.platformStyle.Render(fmt.Sprintf("%s (%d)", sectionTitle, len(items))))
	section.WriteString("\n")

	// Styled border
	border := strings.Repeat("─", 60)
	section.WriteString(f.borderStyle.Render(border))
	section.WriteString("\n")

	// Sort items by updated time (most recent first)
	sortedItems := make([]TodoItem, len(items))
	copy(sortedItems, items)
	sort.Slice(sortedItems, func(i, j int) bool {
		return sortedItems[i].UpdatedAt.After(sortedItems[j].UpdatedAt)
	})

	for _, item := range sortedItems {
		section.WriteString(f.formatTodoItem(item))
	}

	section.WriteString("\n")
	return section.String()
}

func (f *Formatter) formatTodoItem(item TodoItem) string {
	var itemContent strings.Builder

	// Updated time and title
	timeStr := f.timeStyle.Render(item.UpdatedAt.Format("Jan 2 15:04"))
	mainLine := fmt.Sprintf("%s  %s", timeStr, item.Title)
	itemContent.WriteString(mainLine)
	itemContent.WriteString("\n")

	if item.Description != "" {
		description := f.descriptionStyle.Render(item.Description)
		itemContent.WriteString(description)
		itemContent.WriteString("\n")
	}

	if item.URL != "" {
		url := f.urlStyle.Render("🔗 " + item.URL)
		itemContent.WriteString(url)
		itemContent.WriteString("\n")
	}

	if len(item.Tags) > 0 {
		tags := f.tagStyle.Render("🏷️  " + strings.Join(item.Tags, ", "))
		itemContent.WriteString(tags)
		itemContent.WriteString("\n")
	}

	// Wrap the entire item in the activity style
	return f.activityStyle.Render(itemContent.String())
}

// FormatTodoJSON formats todo items for JSON output
func (f *Formatter) FormatTodoJSON(todoItems TodoItems) string {
	// Sort all items by updated time for consistent output
	sortTodoItems := func(items []TodoItem) []TodoItem {
		sorted := make([]TodoItem, len(items))
		copy(sorted, items)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].UpdatedAt.After(sorted[j].UpdatedAt)
		})
		return sorted
	}

	jsonOutput := struct {
		GitHub struct {
			OpenPRs        []TodoItem `json:"open_prs"`
			PendingReviews []TodoItem `json:"pending_reviews"`
		} `json:"github"`
		JIRA struct {
			AssignedTickets []TodoItem `json:"assigned_tickets"`
		} `json:"jira"`
		Obsidian struct {
			Tasks []TodoItem `json:"tasks"`
		} `json:"obsidian"`
		Summary struct {
			Total           int `json:"total"`
			OpenPRs         int `json:"open_prs"`
			PendingReviews  int `json:"pending_reviews"`
			AssignedTickets int `json:"assigned_tickets"`
			ObsidianTasks   int `json:"obsidian_tasks"`
		} `json:"summary"`
	}{}

	// Sort and assign items
	jsonOutput.GitHub.OpenPRs = sortTodoItems(todoItems.GitHub.OpenPRs)
	jsonOutput.GitHub.PendingReviews = sortTodoItems(todoItems.GitHub.PendingReviews)
	jsonOutput.JIRA.AssignedTickets = sortTodoItems(todoItems.JIRA.AssignedTickets)
	jsonOutput.Obsidian.Tasks = sortTodoItems(todoItems.Obsidian.Tasks)

	// Calculate summary
	jsonOutput.Summary.OpenPRs = len(todoItems.GitHub.OpenPRs)
	jsonOutput.Summary.PendingReviews = len(todoItems.GitHub.PendingReviews)
	jsonOutput.Summary.AssignedTickets = len(todoItems.JIRA.AssignedTickets)
	jsonOutput.Summary.ObsidianTasks = len(todoItems.Obsidian.Tasks)
	jsonOutput.Summary.Total = jsonOutput.Summary.OpenPRs + jsonOutput.Summary.PendingReviews + jsonOutput.Summary.AssignedTickets + jsonOutput.Summary.ObsidianTasks

	// Marshal to JSON with proper indentation
	jsonBytes, err := json.MarshalIndent(jsonOutput, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "Failed to marshal JSON: %s"}`, err.Error())
	}

	return string(jsonBytes) + "\n"
}

// FormatTodoTUI launches an interactive TUI for browsing todo items
func (f *Formatter) FormatTodoTUI(todoItems TodoItems) error {
	// Convert output types to tui types to avoid import cycle
	tuiTodoItems := f.convertToTUITypes(todoItems)
	return tui.RunTodoTUI(tuiTodoItems)
}

// convertToTUITypes converts output types to TUI types to avoid import cycles
func (f *Formatter) convertToTUITypes(todoItems TodoItems) types.TodoItems {
	convertTodoItems := func(items []TodoItem) []types.TodoItem {
		result := make([]types.TodoItem, len(items))
		for i, item := range items {
			result[i] = types.TodoItem{
				ID:          item.ID,
				Title:       item.Title,
				Description: item.Description,
				URL:         item.URL,
				UpdatedAt:   item.UpdatedAt,
				Tags:        item.Tags,
			}
		}
		return result
	}

	return types.TodoItems{
		GitHub: types.GitHubTodos{
			OpenPRs:        convertTodoItems(todoItems.GitHub.OpenPRs),
			PendingReviews: convertTodoItems(todoItems.GitHub.PendingReviews),
		},
		JIRA: types.JIRATodos{
			AssignedTickets: convertTodoItems(todoItems.JIRA.AssignedTickets),
		},
		Obsidian: types.ObsidianTodos{
			Tasks: convertTodoItems(todoItems.Obsidian.Tasks),
		},
	}
}

// FormatReview formats review items for text output
func (f *Formatter) FormatReview(reviewItems ReviewItems) string {
	var output strings.Builder

	// Title
	title := "👁️ Review Requests"
	output.WriteString(f.titleStyle.Render(title))
	output.WriteString("\n")

	totalItems := len(reviewItems.GitHub.UserRequests) + len(reviewItems.GitHub.TeamRequests)
	if totalItems == 0 {
		output.WriteString(f.headerStyle.Render("No review requests found."))
		output.WriteString("\n")
		return output.String()
	}

	stats := fmt.Sprintf("Found %d PRs awaiting review", totalItems)
	output.WriteString(f.headerStyle.Render(stats))
	output.WriteString("\n\n")

	// User Review Requests
	if len(reviewItems.GitHub.UserRequests) > 0 {
		output.WriteString(f.formatReviewSection("🫵 Direct Review Requests", reviewItems.GitHub.UserRequests))
	}

	// Team Review Requests
	if len(reviewItems.GitHub.TeamRequests) > 0 {
		output.WriteString(f.formatReviewSection("👥 Team Review Requests", reviewItems.GitHub.TeamRequests))
	}

	return output.String()
}

func (f *Formatter) formatReviewSection(sectionTitle string, items []ReviewItem) string {
	var section strings.Builder

	// Section header
	section.WriteString(f.platformStyle.Render(fmt.Sprintf("%s (%d)", sectionTitle, len(items))))
	section.WriteString("\n")

	// Styled border
	border := strings.Repeat("─", 60)
	section.WriteString(f.borderStyle.Render(border))
	section.WriteString("\n")

	// Sort items by updated time (most recent first)
	sortedItems := make([]ReviewItem, len(items))
	copy(sortedItems, items)
	sort.Slice(sortedItems, func(i, j int) bool {
		return sortedItems[i].TodoItem.UpdatedAt.After(sortedItems[j].TodoItem.UpdatedAt)
	})

	for _, item := range sortedItems {
		section.WriteString(f.formatReviewItem(item))
	}

	section.WriteString("\n")
	return section.String()
}

func (f *Formatter) formatReviewItem(item ReviewItem) string {
	var itemContent strings.Builder

	// Updated time and title
	timeStr := f.timeStyle.Render(item.TodoItem.UpdatedAt.Format("Jan 2 15:04"))

	// CI status indicator
	ciIcon := f.getCIStatusIcon(item.CIStatus.State)

	mainLine := fmt.Sprintf("%s %s %s", timeStr, ciIcon, item.TodoItem.Title)
	itemContent.WriteString(mainLine)
	itemContent.WriteString("\n")

	if item.TodoItem.Description != "" {
		description := f.descriptionStyle.Render(item.TodoItem.Description)
		itemContent.WriteString(description)
		itemContent.WriteString("\n")
	}

	// PR details
	if item.PRDetails.ChangedFiles > 0 {
		prStats := fmt.Sprintf("📊 +%d -%d files: %d",
			item.PRDetails.Additions, item.PRDetails.Deletions, item.PRDetails.ChangedFiles)
		prStatsStyled := f.descriptionStyle.Render(prStats)
		itemContent.WriteString(prStatsStyled)
		itemContent.WriteString("\n")
	}

	// CI status details
	if item.CIStatus.TotalCount > 0 {
		ciDetails := fmt.Sprintf("🔍 CI: %s (%d checks)", item.CIStatus.State, item.CIStatus.TotalCount)
		ciDetailsStyled := f.descriptionStyle.Render(ciDetails)
		itemContent.WriteString(ciDetailsStyled)
		itemContent.WriteString("\n")
	}

	if item.TodoItem.URL != "" {
		url := f.urlStyle.Render("🔗 " + item.TodoItem.URL)
		itemContent.WriteString(url)
		itemContent.WriteString("\n")
	}

	if len(item.TodoItem.Tags) > 0 {
		tags := f.tagStyle.Render("🏷️  " + strings.Join(item.TodoItem.Tags, ", "))
		itemContent.WriteString(tags)
		itemContent.WriteString("\n")
	}

	// Wrap the entire item in the activity style
	return f.activityStyle.Render(itemContent.String())
}

func (f *Formatter) getCIStatusIcon(state string) string {
	switch state {
	case "success":
		return "✅"
	case "failure":
		return "❌"
	case "pending":
		return "🟡"
	default:
		return "⚪"
	}
}

// FormatReviewJSON formats review items for JSON output
func (f *Formatter) FormatReviewJSON(reviewItems ReviewItems) string {
	// Sort all items by updated time for consistent output
	sortReviewItems := func(items []ReviewItem) []ReviewItem {
		sorted := make([]ReviewItem, len(items))
		copy(sorted, items)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].TodoItem.UpdatedAt.After(sorted[j].TodoItem.UpdatedAt)
		})
		return sorted
	}

	jsonOutput := struct {
		GitHub struct {
			UserRequests []ReviewItem `json:"user_requests"`
			TeamRequests []ReviewItem `json:"team_requests"`
		} `json:"github"`
		Summary struct {
			Total        int `json:"total"`
			UserRequests int `json:"user_requests"`
			TeamRequests int `json:"team_requests"`
		} `json:"summary"`
	}{}

	// Sort and assign items
	jsonOutput.GitHub.UserRequests = sortReviewItems(reviewItems.GitHub.UserRequests)
	jsonOutput.GitHub.TeamRequests = sortReviewItems(reviewItems.GitHub.TeamRequests)

	// Calculate summary
	jsonOutput.Summary.UserRequests = len(reviewItems.GitHub.UserRequests)
	jsonOutput.Summary.TeamRequests = len(reviewItems.GitHub.TeamRequests)
	jsonOutput.Summary.Total = jsonOutput.Summary.UserRequests + jsonOutput.Summary.TeamRequests

	// Marshal to JSON with proper indentation
	jsonBytes, err := json.MarshalIndent(jsonOutput, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "Failed to marshal JSON: %s"}`, err.Error())
	}

	return string(jsonBytes) + "\n"
}

// FormatReviewTUI launches an interactive TUI for browsing review items
func (f *Formatter) FormatReviewTUI(reviewItems ReviewItems) error {
	// Convert output.ReviewItems to types.ReviewItems
	typesReviewItems := types.ReviewItems{
		GitHub: types.GitHubReviews{
			UserRequests: convertReviewItems(reviewItems.GitHub.UserRequests),
			TeamRequests: convertReviewItems(reviewItems.GitHub.TeamRequests),
		},
	}
	return tui.RunReviewsTUI(typesReviewItems)
}

func convertReviewItems(items []ReviewItem) []types.ReviewItem {
	result := make([]types.ReviewItem, len(items))
	for i, item := range items {
		result[i] = types.ReviewItem{
			TodoItem: types.TodoItem{
				ID:          item.TodoItem.ID,
				Title:       item.TodoItem.Title,
				Description: item.TodoItem.Description,
				URL:         item.TodoItem.URL,
				UpdatedAt:   item.TodoItem.UpdatedAt,
				Tags:        item.TodoItem.Tags,
			},
			CIStatus: types.CIStatus{
				State:      item.CIStatus.State,
				TotalCount: item.CIStatus.TotalCount,
				Checks:     convertCheckRuns(item.CIStatus.Checks),
			},
			PRDetails: types.PRDetails{
				Additions:    item.PRDetails.Additions,
				Deletions:    item.PRDetails.Deletions,
				ChangedFiles: item.PRDetails.ChangedFiles,
			},
		}
	}
	return result
}

func convertCheckRuns(checks []CheckRun) []types.CheckRun {
	result := make([]types.CheckRun, len(checks))
	for i, check := range checks {
		result[i] = types.CheckRun{
			Name:       check.Name,
			Status:     check.Status,
			Conclusion: check.Conclusion,
			URL:        check.URL,
		}
	}
	return result
}

// TodoItem represents a single todo item (avoiding import cycles)
type TodoItem struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
	Tags        []string  `json:"tags,omitempty"`
}

// TodoItems represents all pending work items
type TodoItems struct {
	GitHub   GitHubTodos   `json:"github"`
	JIRA     JIRATodos     `json:"jira"`
	Obsidian ObsidianTodos `json:"obsidian"`
}

// GitHubTodos represents pending GitHub work items
type GitHubTodos struct {
	OpenPRs        []TodoItem `json:"open_prs"`
	PendingReviews []TodoItem `json:"pending_reviews"`
}

// JIRATodos represents pending JIRA work items
type JIRATodos struct {
	AssignedTickets []TodoItem `json:"assigned_tickets"`
}

// ObsidianTodos represents pending Obsidian work items
type ObsidianTodos struct {
	Tasks []TodoItem `json:"tasks"`
}

// ReviewItems represents all review items
type ReviewItems struct {
	GitHub GitHubReviews `json:"github"`
}

// GitHubReviews represents review items from GitHub
type GitHubReviews struct {
	UserRequests []ReviewItem `json:"user_requests"`
	TeamRequests []ReviewItem `json:"team_requests"`
}

// ReviewItem represents a pull request awaiting review with additional details
type ReviewItem struct {
	TodoItem  TodoItem  `json:"todo_item"`
	CIStatus  CIStatus  `json:"ci_status"`
	PRDetails PRDetails `json:"pr_details"`
}

// CIStatus represents CI check status for a PR
type CIStatus struct {
	State      string     `json:"state"` // success, failure, pending
	TotalCount int        `json:"total_count"`
	Checks     []CheckRun `json:"checks"`
}

// CheckRun represents a single CI check
type CheckRun struct {
	Name       string `json:"name"`
	Status     string `json:"status"`     // completed, in_progress, queued
	Conclusion string `json:"conclusion"` // success, failure, cancelled, etc.
	URL        string `json:"url,omitempty"`
}

// PRDetails represents additional PR information
type PRDetails struct {
	Additions    int `json:"additions"`
	Deletions    int `json:"deletions"`
	ChangedFiles int `json:"changed_files"`
}
