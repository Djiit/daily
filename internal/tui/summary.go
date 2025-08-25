package tui

import (
	"fmt"
	"io"
	"math"
	"sort"
	"strings"
	"time"

	catppuccin "github.com/catppuccin/go"
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
	if len(m.activities) == 0 {
		return m.styles.Header.Render("No activities found for this date.") +
			"\n\nPress q to quit"
	}

	// Calculate panel dimensions
	leftWidth := int(math.Floor(float64(m.windowWidth) * 0.4)) // 40% for left panel
	rightWidth := m.windowWidth - leftWidth - 3                // Remaining for right panel (minus border)

	// Header
	title := fmt.Sprintf("📊 Daily Summary for %s", m.summary.Date.Format("January 2, 2006"))

	// Use appropriate Catppuccin colors based on theme
	var headerColor string
	if isDarkMode() {
		headerColor = catppuccin.Mocha.Mauve().Hex
	} else {
		headerColor = catppuccin.Latte.Mauve().Hex
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(headerColor)).
		Width(m.windowWidth).
		Align(lipgloss.Center).
		MarginBottom(1)

	header := headerStyle.Render(title)

	// Create left and right panels
	leftPanel := m.renderLeftPanel(leftWidth)
	rightPanel := m.renderRightPanel(rightWidth)

	// Combine panels
	return lipgloss.JoinVertical(lipgloss.Top,
		header,
		lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel),
	)
}

func (m summaryModel) renderLeftPanel(width int) string {
	// Use appropriate Catppuccin colors based on theme
	var borderColor string
	if isDarkMode() {
		borderColor = catppuccin.Mocha.Surface2().Hex
	} else {
		borderColor = catppuccin.Latte.Surface2().Hex
	}

	// Left panel style with border
	leftStyle := lipgloss.NewStyle().
		Width(width).
		Height(m.leftViewport.height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(1)

	var content strings.Builder

	// Navigation help
	helpText := "↑/↓ j/k: Navigate • Enter: Open URL • q: Quit"

	// Use appropriate Catppuccin colors based on theme
	var helpColor string
	if isDarkMode() {
		helpColor = catppuccin.Mocha.Subtext1().Hex
	} else {
		helpColor = catppuccin.Latte.Subtext1().Hex
	}

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(helpColor)).
		Italic(true).
		Width(width - 4)
	content.WriteString(helpStyle.Render(helpText))
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
		maxTitleWidth := max(5, width-15) // Account for time, icons, and padding, minimum 5 chars
		title := act.Title
		if len(title) > maxTitleWidth && maxTitleWidth > 3 {
			title = title[:maxTitleWidth-3] + "..."
		} else if len(title) > maxTitleWidth {
			// If maxTitleWidth is too small even for "...", just truncate
			title = title[:max(1, maxTitleWidth)]
		}

		var line strings.Builder
		line.WriteString(fmt.Sprintf("%s %s %s %s", timeStr, platformIcon, typeIcon, title))

		if act.URL != "" {
			line.WriteString(" 🔗")
		}

		// Apply selection styling
		if isSelected {
			var selectedFg, selectedBg string
			if isDarkMode() {
				selectedFg = catppuccin.Mocha.Base().Hex
				selectedBg = catppuccin.Mocha.Blue().Hex
			} else {
				selectedFg = catppuccin.Latte.Base().Hex
				selectedBg = catppuccin.Latte.Blue().Hex
			}

			style := lipgloss.NewStyle().
				Foreground(lipgloss.Color(selectedFg)).
				Background(lipgloss.Color(selectedBg)).
				Bold(true).
				Width(width - 4)
			content.WriteString(style.Render("> " + line.String()))
		} else {
			style := lipgloss.NewStyle().Width(width - 4)
			content.WriteString(style.Render("  " + line.String()))
		}

		content.WriteString("\n")
	}

	// Scroll indicator
	if len(m.activities) > m.leftViewport.height-4 {
		current := m.cursor + 1
		total := len(m.activities)
		scrollInfo := fmt.Sprintf("[%d/%d]", current, total)

		var scrollColor string
		if isDarkMode() {
			scrollColor = catppuccin.Mocha.Subtext1().Hex
		} else {
			scrollColor = catppuccin.Latte.Subtext1().Hex
		}

		scrollStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(scrollColor)).
			Align(lipgloss.Right).
			Width(width - 4)
		content.WriteString("\n")
		content.WriteString(scrollStyle.Render(scrollInfo))
	}

	return leftStyle.Render(content.String())
}

func (m summaryModel) renderRightPanel(width int) string {
	// Use appropriate Catppuccin colors based on theme
	var borderColor string
	if isDarkMode() {
		borderColor = catppuccin.Mocha.Surface2().Hex
	} else {
		borderColor = catppuccin.Latte.Surface2().Hex
	}

	// Right panel style with border
	rightStyle := lipgloss.NewStyle().
		Width(width).
		Height(m.rightViewport.height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(1)

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
		Width(width - 4) // Account for padding and border

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
	glamourStyle, err := glamour.NewTermRenderer(glamour.WithStandardStyle(glamourTheme))
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
