package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss/v2"

	"daily/internal/tui/types"
)

// ReviewsModel represents the state of the reviews TUI
type ReviewsModel struct {
	reviewItems   types.ReviewItems
	selectedItem  int
	width         int
	height        int
	styles        *CommonStyles
	allItems      []ReviewListItem // flattened list of all items for navigation
	leftViewport  viewportState
	rightViewport viewportState
	glamourStyle  *glamour.TermRenderer
}

// ReviewListItem represents an item in the navigation list
type ReviewListItem struct {
	Item        types.ReviewItem
	Type        string // "user_request", "team_request"
	DisplayText string
}

// NewReviewsModel creates a new reviews TUI model
func NewReviewsModel(reviewItems types.ReviewItems) ReviewsModel {
	// Initialize glamour renderer
	var glamourStyle *glamour.TermRenderer
	var glamourTheme string
	if isDarkMode() {
		glamourTheme = "dark"
	} else {
		glamourTheme = "light"
	}
	glamourStyle, err := glamour.NewTermRenderer(glamour.WithStandardStyle(glamourTheme), glamour.WithEmoji())
	if err != nil {
		glamourStyle = nil
	}

	model := ReviewsModel{
		reviewItems:  reviewItems,
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
	model.buildItemsList()
	return model
}

func (m *ReviewsModel) buildItemsList() {
	m.allItems = []ReviewListItem{}

	// Add user requests
	for _, item := range m.reviewItems.GitHub.UserRequests {
		m.allItems = append(m.allItems, ReviewListItem{
			Item:        item,
			Type:        "user_request",
			DisplayText: fmt.Sprintf("👤 %s", item.TodoItem.Title),
		})
	}

	// Add team requests
	for _, item := range m.reviewItems.GitHub.TeamRequests {
		m.allItems = append(m.allItems, ReviewListItem{
			Item:        item,
			Type:        "team_request",
			DisplayText: fmt.Sprintf("👥 %s", item.TodoItem.Title),
		})
	}

	// Sort by updated time (most recent first)
	sort.Slice(m.allItems, func(i, j int) bool {
		return m.allItems[i].Item.TodoItem.UpdatedAt.After(m.allItems[j].Item.TodoItem.UpdatedAt)
	})
}

func (m ReviewsModel) Init() tea.Cmd {
	return nil
}

func (m ReviewsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.leftViewport.height = msg.Height - 4  // Reserve space for header
		m.rightViewport.height = msg.Height - 4 // Reserve space for header
		m.updateLeftViewport()
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			m.selectedItem = ClampCursor(m.selectedItem-1, 0, len(m.allItems)-1)
			m.updateLeftViewport()
		case "down", "j":
			m.selectedItem = ClampCursor(m.selectedItem+1, 0, len(m.allItems)-1)
			m.updateLeftViewport()
		case "home", "g":
			m.selectedItem = 0
			m.updateLeftViewport()
		case "end", "G":
			m.selectedItem = len(m.allItems) - 1
			m.updateLeftViewport()
		case "enter", " ":
			if m.selectedItem < len(m.allItems) && m.allItems[m.selectedItem].Item.TodoItem.URL != "" {
				url := m.allItems[m.selectedItem].Item.TodoItem.URL
				return m, tea.Exec(urlCommand{url: url}, nil)
			}
			return m, nil
		}
	}
	return m, nil
}

func (m *ReviewsModel) updateLeftViewport() {
	if m.leftViewport.height <= 0 {
		return
	}

	// Ensure cursor is visible in viewport
	if m.selectedItem < m.leftViewport.offset {
		m.leftViewport.offset = m.selectedItem
	} else if m.selectedItem >= m.leftViewport.offset+m.leftViewport.height {
		m.leftViewport.offset = m.selectedItem - m.leftViewport.height + 1
	}

	// Ensure viewport doesn't exceed bounds
	m.leftViewport.offset = max(0, m.leftViewport.offset)
	maxOffset := max(0, len(m.allItems)-m.leftViewport.height)
	m.leftViewport.offset = min(m.leftViewport.offset, maxOffset)
}

func (m ReviewsModel) View() string {
	// Check if terminal is too small
	if !IsTerminalSizeAdequate(m.width, m.height) {
		return RenderTerminalTooSmallMessage(m.styles, m.width, m.height)
	}

	if len(m.allItems) == 0 {
		return m.styles.Base.Render(
			m.styles.Header.Render("👁️ Review Requests") + "\n" +
				"No pending review requests found.\n\n" +
				m.styles.Help.Render("Press 'q' to quit"),
		)
	}

	// Calculate panel dimensions
	dimensions := CalculatePanelDimensions(m.width)
	if dimensions.UseSingle {
		return m.renderSinglePanelView()
	}

	// Header
	title := fmt.Sprintf("👁️ Review Requests (%d)", len(m.allItems))
	header := RenderHeader(title, m.width)

	// Create left and right panels
	leftPanel := m.renderLeftPanel(dimensions.LeftWidth)
	rightPanel := m.renderRightPanel(dimensions.RightWidth)

	// Combine panels
	return lipgloss.JoinVertical(lipgloss.Top,
		header,
		lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel),
	)
}

func (m ReviewsModel) renderLeftPanel(width int) string {
	// Create bordered panel with theme-appropriate colors
	_, borderColor, _, _, _, _ := GetThemeColors()
	leftStyle := CreateBorderedPanel(width, m.leftViewport.height, borderColor)

	var content strings.Builder

	// Navigation help
	helpText := "↑/↓ j/k: Navigate • Enter: Open URL • q: Quit"
	adjustedWidth := max(20, width) // Same adjustment as in CreateBorderedPanel
	content.WriteString(RenderHelpText(helpText, adjustedWidth-4))
	content.WriteString("\n\n")

	// Review items list
	end := min(len(m.allItems), m.leftViewport.offset+m.leftViewport.height-4) // Account for help text and padding

	for i := m.leftViewport.offset; i < end; i++ {
		item := m.allItems[i]
		isSelected := i == m.selectedItem

		// Create review item display
		timeStr := item.Item.TodoItem.UpdatedAt.Format("Jan 2")

		// Get appropriate icon for item type
		var icon string
		switch item.Type {
		case "user_request":
			icon = "👤"
		case "team_request":
			icon = "👥"
		default:
			icon = "👁️"
		}

		// Add CI status indicator
		ciIcon := getCIStatusIcon(item.Item.CIStatus)

		// Truncate title to fit width
		maxTitleWidth := max(5, adjustedWidth-20) // Account for time, icons, and padding
		title := TruncateText(item.Item.TodoItem.Title, maxTitleWidth)

		var line strings.Builder
		line.WriteString(fmt.Sprintf("%s %s %s %s", timeStr, icon, ciIcon, title))

		if item.Item.TodoItem.URL != "" {
			line.WriteString(" 🔗")
		}

		// Apply selection styling
		content.WriteString(ApplySelectionStyle(line.String(), isSelected, adjustedWidth-4))

		content.WriteString("\n")
	}

	// Scroll indicator
	if len(m.allItems) > m.leftViewport.height-4 {
		content.WriteString("\n")
		content.WriteString(RenderScrollIndicator(m.selectedItem+1, len(m.allItems), adjustedWidth-4))
	}

	return leftStyle.Render(content.String())
}

func (m ReviewsModel) renderRightPanel(width int) string {
	// Create bordered panel with theme-appropriate colors
	_, borderColor, _, _, _, _ := GetThemeColors()
	rightStyle := CreateBorderedPanel(width, m.rightViewport.height, borderColor)
	adjustedWidth := max(30, width) // Same adjustment as in CreateBorderedPanel

	if m.selectedItem >= len(m.allItems) {
		return rightStyle.Render("Select a review request to view details")
	}

	selectedItem := m.allItems[m.selectedItem]

	// Create markdown content for the selected review item
	markdown := m.createReviewMarkdownContent(selectedItem)

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

func (m ReviewsModel) createReviewMarkdownContent(item ReviewListItem) string {
	var md strings.Builder

	// Title
	md.WriteString(fmt.Sprintf("# %s\n\n", item.Item.TodoItem.Title))

	// Metadata table
	md.WriteString("## Details\n\n")
	md.WriteString("| Field | Value |\n")
	md.WriteString("|-------|-------|\n")
	md.WriteString(fmt.Sprintf("| **Updated** | %s |\n", item.Item.TodoItem.UpdatedAt.Format("Jan 2, 2006 15:04")))

	// Type-specific information
	switch item.Type {
	case "user_request":
		md.WriteString("| **Type** | 👤 User Review Request |\n")
	case "team_request":
		md.WriteString("| **Type** | 👥 Team Review Request |\n")
	default:
		md.WriteString("| **Type** | 👁️ Review Request |\n")
	}

	// CI Status
	ciStatus := item.Item.CIStatus
	if ciStatus.State != "" {
		icon := getCIStatusIcon(ciStatus)
		md.WriteString(fmt.Sprintf("| **CI Status** | %s %s |\n", icon, strings.Title(ciStatus.State)))
	}

	// PR Details
	prDetails := item.Item.PRDetails
	if prDetails.Additions > 0 || prDetails.Deletions > 0 || prDetails.ChangedFiles > 0 {
		md.WriteString(fmt.Sprintf("| **Changes** | +%d -%d (%d files) |\n",
			prDetails.Additions, prDetails.Deletions, prDetails.ChangedFiles))
	}

	if item.Item.TodoItem.URL != "" {
		md.WriteString(fmt.Sprintf("| **URL** | [🔗 Open PR](%s) |\n", item.Item.TodoItem.URL))
	}

	// Description
	if item.Item.TodoItem.Description != "" {
		md.WriteString("\n## Description\n\n")
		md.WriteString(item.Item.TodoItem.Description)
		md.WriteString("\n\n")
	}

	// CI Checks details
	if len(ciStatus.Checks) > 0 {
		md.WriteString("## CI Checks\n\n")
		for _, check := range ciStatus.Checks {
			checkIcon := getCheckIcon(check.Status, check.Conclusion)
			md.WriteString(fmt.Sprintf("- %s **%s**", checkIcon, check.Name))
			if check.URL != "" {
				md.WriteString(fmt.Sprintf(" ([link](%s))", check.URL))
			}
			md.WriteString("\n")
		}
		md.WriteString("\n")
	}

	// Tags
	if len(item.Item.TodoItem.Tags) > 0 {
		md.WriteString("## Tags\n\n")
		for _, tag := range item.Item.TodoItem.Tags {
			md.WriteString(fmt.Sprintf("- `%s`\n", tag))
		}
		md.WriteString("\n")
	}

	// Additional metadata
	md.WriteString("## Metadata\n\n")
	md.WriteString(fmt.Sprintf("- **ID**: `%s`\n", item.Item.TodoItem.ID))

	return md.String()
}

func (m ReviewsModel) renderSinglePanelView() string {
	var content strings.Builder

	// Header
	title := fmt.Sprintf("👁️ Review Requests (%d)", len(m.allItems))
	content.WriteString(RenderHeader(title, m.width))
	content.WriteString("\n")

	// Navigation help
	helpText := "↑/↓ j/k: Navigate • Enter: Open URL • q: Quit"
	content.WriteString(RenderHelpText(helpText, m.width))
	content.WriteString("\n\n")

	// Review items list (simplified)
	availableHeight := m.height - 6 // Account for header and help
	start := max(0, m.selectedItem-availableHeight/2)
	end := min(len(m.allItems), start+availableHeight)

	// Adjust start if end reached the limit
	if end == len(m.allItems) && end-start < availableHeight {
		start = max(0, end-availableHeight)
	}

	for i := start; i < end; i++ {
		item := m.allItems[i]
		isSelected := i == m.selectedItem

		// Simple review item line
		timeStr := item.Item.TodoItem.UpdatedAt.Format("Jan 2")

		// Get appropriate icon for item type
		var icon string
		switch item.Type {
		case "user_request":
			icon = "👤"
		case "team_request":
			icon = "👥"
		default:
			icon = "👁️"
		}

		// Add CI status indicator
		ciIcon := getCIStatusIcon(item.Item.CIStatus)

		// Truncate title to fit
		maxTitleWidth := max(5, m.width-20)
		title := TruncateText(item.Item.TodoItem.Title, maxTitleWidth)

		line := fmt.Sprintf("%s %s %s %s", timeStr, icon, ciIcon, title)
		if item.Item.TodoItem.URL != "" {
			line += " 🔗"
		}

		content.WriteString(ApplySelectionStyle(line, isSelected, m.width))
		content.WriteString("\n")
	}

	// Show current item details if space available
	if m.selectedItem < len(m.allItems) && m.height > 15 {
		item := m.allItems[m.selectedItem]
		content.WriteString("\n")

		// Show PR details if available
		prDetails := item.Item.PRDetails
		if prDetails.Additions > 0 || prDetails.Deletions > 0 {
			changes := fmt.Sprintf("+%d -%d (%d files)",
				prDetails.Additions, prDetails.Deletions, prDetails.ChangedFiles)
			_, _, _, _, _, scrollColor := GetThemeColors()
			changesStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(scrollColor)).Italic(true)
			content.WriteString(changesStyle.Render(changes))
			content.WriteString("\n")
		}

		if item.Item.TodoItem.Description != "" {
			desc := TruncateText(item.Item.TodoItem.Description, m.width-4)
			_, _, _, _, _, scrollColor := GetThemeColors()
			descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(scrollColor)).Italic(true)
			content.WriteString(descStyle.Render(desc))
			content.WriteString("\n")
		}
	}

	// Scroll indicator
	if len(m.allItems) > availableHeight {
		content.WriteString("\n")
		content.WriteString(RenderScrollIndicator(m.selectedItem+1, len(m.allItems), m.width))
	}

	return content.String()
}

// getCIStatusIcon returns an appropriate icon for CI status
func getCIStatusIcon(status types.CIStatus) string {
	switch status.State {
	case "success":
		return "✅"
	case "failure":
		return "❌"
	case "pending":
		return "⏳"
	default:
		return "⚪"
	}
}

// getCheckIcon returns an appropriate icon for individual check status
func getCheckIcon(status, conclusion string) string {
	if status == "completed" {
		switch conclusion {
		case "success":
			return "✅"
		case "failure":
			return "❌"
		case "cancelled":
			return "⚪"
		default:
			return "❓"
		}
	}
	if status == "in_progress" {
		return "⏳"
	}
	if status == "queued" {
		return "⏸️"
	}
	return "⚪"
}

// RunReviewsTUI starts the reviews TUI application
func RunReviewsTUI(reviewItems types.ReviewItems) error {
	model := NewReviewsModel(reviewItems)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, err := p.Run()
	return err
}
