package tui

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"daily/internal/activity"
)

type urlCommand struct {
	url string
}

func (c urlCommand) Run() error {
	return OpenURL(c.url)
}

func (c urlCommand) SetStdout(w io.Writer) {}
func (c urlCommand) SetStderr(w io.Writer) {}
func (c urlCommand) SetStdin(r io.Reader)  {}

type summaryModel struct {
	summary       *activity.Summary
	activities    []activity.Activity
	cursor        int
	leftViewport  viewportState
	rightViewport viewportState
	windowHeight  int
	windowWidth   int
	styles        *CommonStyles
	glamourStyle  *glamour.TermRenderer
}

type viewportState struct {
	offset int
	height int
}

func (m summaryModel) Init() tea.Cmd {
	return nil
}

func (m summaryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.updateLeftViewport()
			}
		case "down", "j":
			if m.cursor < len(m.activities)-1 {
				m.cursor++
				m.updateLeftViewport()
			}
		case "enter", " ":
			if m.cursor < len(m.activities) && m.activities[m.cursor].URL != "" {
				url := m.activities[m.cursor].URL
				return m, tea.Exec(urlCommand{url: url}, nil)
			}
		case "home", "g":
			m.cursor = 0
			m.updateLeftViewport()
		case "end", "G":
			m.cursor = len(m.activities) - 1
			m.updateLeftViewport()
		}

	case tea.WindowSizeMsg:
		m.windowHeight = msg.Height
		m.windowWidth = msg.Width
		m.leftViewport.height = msg.Height - 4  // Reserve space for header
		m.rightViewport.height = msg.Height - 4 // Reserve space for header
		m.updateLeftViewport()
	}

	return m, nil
}

func (m *summaryModel) updateLeftViewport() {
	if m.leftViewport.height <= 0 {
		return
	}

	// Ensure cursor is visible in viewport
	if m.cursor < m.leftViewport.offset {
		m.leftViewport.offset = m.cursor
	} else if m.cursor >= m.leftViewport.offset+m.leftViewport.height {
		m.leftViewport.offset = m.cursor - m.leftViewport.height + 1
	}

	// Ensure viewport doesn't exceed bounds
	m.leftViewport.offset = max(0, m.leftViewport.offset)
	maxOffset := max(0, len(m.activities)-m.leftViewport.height)
	m.leftViewport.offset = min(m.leftViewport.offset, maxOffset)
}

func (m summaryModel) View() string {
	// Check if terminal is too small
	if !IsTerminalSizeAdequate(m.windowWidth, m.windowHeight) {
		return RenderTerminalTooSmallMessage(m.styles, m.windowWidth, m.windowHeight)
	}

	if len(m.activities) == 0 {
		return m.styles.Header.Render("No activities found for this date.") +
			"\n\nPress q to quit"
	}

	// Calculate panel dimensions
	dimensions := CalculatePanelDimensions(m.windowWidth)
	if dimensions.UseSingle {
		return m.renderSinglePanelView()
	}

	// Header
	title := fmt.Sprintf("📊 Daily Summary for %s", m.summary.Date.Format("January 2, 2006"))
	header := RenderHeader(title, m.windowWidth)

	// Create left and right panels
	leftPanel := m.renderLeftPanel(dimensions.LeftWidth)
	rightPanel := m.renderRightPanel(dimensions.RightWidth)

	// Combine panels
	return lipgloss.JoinVertical(lipgloss.Top,
		header,
		lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel),
	)
}

func (m summaryModel) renderLeftPanel(width int) string {
	// Create bordered panel with theme-appropriate colors
	_, borderColor, _, _, _, _ := GetThemeColors()
	leftStyle := CreateBorderedPanel(width, m.leftViewport.height, borderColor)

	var content strings.Builder

	// Navigation help
	helpText := "↑/↓ j/k: Navigate • Enter: Open URL • q: Quit"
	adjustedWidth := max(20, width) // Same adjustment as in CreateBorderedPanel
	content.WriteString(RenderHelpText(helpText, adjustedWidth-4))
	content.WriteString("\n\n")

	// Activities list
	end := min(len(m.activities), m.leftViewport.offset+m.leftViewport.height-4) // Account for help text and padding

	for i := m.leftViewport.offset; i < end; i++ {
		act := m.activities[i]
		isSelected := i == m.cursor

		// Create activity display
		timeStr := act.Timestamp.Format("15:04")
		platformIcon := getPlatformIcon(act.Platform)
		typeIcon := getTypeIcon(act.Type)

		// Truncate title to fit width
		maxTitleWidth := max(5, adjustedWidth-15) // Account for time, icons, and padding, minimum 5 chars
		title := TruncateText(act.Title, maxTitleWidth)

		var line strings.Builder
		line.WriteString(fmt.Sprintf("%s %s %s %s", timeStr, platformIcon, typeIcon, title))

		if act.URL != "" {
			line.WriteString(" 🔗")
		}

		// Apply selection styling
		content.WriteString(ApplySelectionStyle(line.String(), isSelected, adjustedWidth-4))

		content.WriteString("\n")
	}

	// Scroll indicator
	if len(m.activities) > m.leftViewport.height-4 {
		content.WriteString("\n")
		content.WriteString(RenderScrollIndicator(m.cursor+1, len(m.activities), adjustedWidth-4))
	}

	return leftStyle.Render(content.String())
}

// renderSinglePanelView renders a simplified single-panel view for narrow terminals
func (m summaryModel) renderSinglePanelView() string {
	var content strings.Builder

	// Header
	title := fmt.Sprintf("📊 Daily Summary for %s", m.summary.Date.Format("January 2, 2006"))
	content.WriteString(RenderHeader(title, m.windowWidth))
	content.WriteString("\n")

	// Navigation help
	helpText := "↑/↓ j/k: Navigate • Enter: Open URL • q: Quit"
	content.WriteString(RenderHelpText(helpText, m.windowWidth))
	content.WriteString("\n\n")

	// Activities list (simplified)
	availableHeight := m.windowHeight - 6 // Account for header and help
	start := max(0, m.cursor-availableHeight/2)
	end := min(len(m.activities), start+availableHeight)

	// Adjust start if end reached the limit
	if end == len(m.activities) && end-start < availableHeight {
		start = max(0, end-availableHeight)
	}

	for i := start; i < end; i++ {
		act := m.activities[i]
		isSelected := i == m.cursor

		// Simple activity line
		timeStr := act.Timestamp.Format("15:04")
		platformIcon := getPlatformIcon(act.Platform)
		typeIcon := getTypeIcon(act.Type)

		// Truncate title to fit
		maxTitleWidth := max(5, m.windowWidth-15)
		title := TruncateText(act.Title, maxTitleWidth)

		line := fmt.Sprintf("%s %s %s %s", timeStr, platformIcon, typeIcon, title)
		if act.URL != "" {
			line += " 🔗"
		}

		content.WriteString(ApplySelectionStyle(line, isSelected, m.windowWidth))
		content.WriteString("\n")
	}

	// Show current item details if space available
	if m.cursor < len(m.activities) && m.windowHeight > 15 {
		act := m.activities[m.cursor]
		content.WriteString("\n")
		if act.Description != "" {
			desc := TruncateText(act.Description, m.windowWidth-4)
			_, _, _, _, _, scrollColor := GetThemeColors() // Reuse scroll color for description
			descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(scrollColor)).Italic(true)
			content.WriteString(descStyle.Render(desc))
			content.WriteString("\n")
		}
	}

	// Scroll indicator
	if len(m.activities) > availableHeight {
		content.WriteString("\n")
		content.WriteString(RenderScrollIndicator(m.cursor+1, len(m.activities), m.windowWidth))
	}

	return content.String()
}

func (m summaryModel) renderRightPanel(width int) string {
	// Create bordered panel with theme-appropriate colors
	_, borderColor, _, _, _, _ := GetThemeColors()
	rightStyle := CreateBorderedPanel(width, m.rightViewport.height, borderColor)
	adjustedWidth := max(30, width) // Same adjustment as in CreateBorderedPanel

	if m.cursor >= len(m.activities) {
		return rightStyle.Render("Select an activity to view details")
	}

	act := m.activities[m.cursor]

	// Create markdown content for the selected activity
	markdown := m.createMarkdownContent(act)

	// Render markdown using glamour if available
	var rendered string
	if m.glamourStyle != nil {
		var err error
		rendered, err = m.glamourStyle.Render(markdown)
		if err != nil {
			rendered = markdown // Fallback to plain text
		}
	} else {
		rendered = markdown // No glamour available, use plain text
	}

	// Wrap content to fit width
	contentStyle := lipgloss.NewStyle().
		Width(max(10, adjustedWidth-4)) // Account for padding and border

	return rightStyle.Render(contentStyle.Render(rendered))
}

func (m summaryModel) createMarkdownContent(act activity.Activity) string {
	var md strings.Builder

	// Title
	md.WriteString(fmt.Sprintf("# %s\n\n", act.Title))

	// Metadata table
	md.WriteString("## Details\n\n")
	md.WriteString("| Field | Value |\n")
	md.WriteString("|-------|-------|\n")
	md.WriteString(fmt.Sprintf("| **Time** | %s |\n", act.Timestamp.Format("15:04:05")))
	md.WriteString(fmt.Sprintf("| **Platform** | %s %s |\n", getPlatformIcon(act.Platform), act.Platform))
	md.WriteString(fmt.Sprintf("| **Type** | %s %s |\n", getTypeIcon(act.Type), string(act.Type)))

	if act.URL != "" {
		md.WriteString(fmt.Sprintf("| **URL** | [🔗 Open Link](%s) |\n", act.URL))
	}

	// Description
	if act.Description != "" {
		md.WriteString("\n## Description\n\n")
		md.WriteString(act.Description)
		md.WriteString("\n\n")
	}

	// Tags
	if len(act.Tags) > 0 {
		md.WriteString("## Tags\n\n")
		for _, tag := range act.Tags {
			md.WriteString(fmt.Sprintf("- `%s`\n", tag))
		}
		md.WriteString("\n")
	}

	// Additional metadata
	md.WriteString("## Metadata\n\n")
	md.WriteString(fmt.Sprintf("- **ID**: `%s`\n", act.ID))
	md.WriteString(fmt.Sprintf("- **Full Timestamp**: `%s`\n", act.Timestamp.Format(time.RFC3339)))

	return md.String()
}

// RunTUIForced starts the TUI for the given summary, bypassing TTY checks (for testing)
func RunTUIForced(summary *activity.Summary) error {
	return runTUIInternal(summary, true)
}

// RunTUI starts the TUI for the given summary
func RunTUI(summary *activity.Summary) error {
	return runTUIInternal(summary, false)
}

func runTUIInternal(summary *activity.Summary, force bool) error {
	// Check if we're running in a terminal that supports TUI (unless forced)
	if !force && !IsTerminalCapable() {
		// Not in a TTY, fall back to text output
		// We'll handle the fallback in the calling function
		return fmt.Errorf("terminal does not support TUI")
	}

	// Sort activities by timestamp
	activities := make([]activity.Activity, len(summary.Activities))
	copy(activities, summary.Activities)
	sort.Slice(activities, func(i, j int) bool {
		return activities[i].Timestamp.Before(activities[j].Timestamp)
	})

	// Initialize glamour renderer with simple fallback
	var glamourStyle *glamour.TermRenderer
	var glamourTheme string
	if isDarkMode() {
		glamourTheme = "dark"
	} else {
		glamourTheme = "light"
	}
	glamourStyle, err := glamour.NewTermRenderer(glamour.WithStandardStyle(glamourTheme), glamour.WithEmoji())
	if err != nil {
		// If glamour fails completely, we'll handle this in the render function
		glamourStyle = nil
	}

	m := summaryModel{
		summary:      summary,
		activities:   activities,
		cursor:       0,
		styles:       NewCommonStyles(),
		glamourStyle: glamourStyle,
		leftViewport: viewportState{
			offset: 0,
			height: 20, // Default height, will be updated on window size msg
		},
		rightViewport: viewportState{
			offset: 0,
			height: 20, // Default height, will be updated on window size msg
		},
	}

	// Run the TUI
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	if err != nil {
		// If TUI fails for any reason, return error so caller can handle fallback
		return err
	}
	return nil
}

// Icon functions for activities and platforms
func getPlatformIcon(platform string) string {
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

func getTypeIcon(actType activity.ActivityType) string {
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
