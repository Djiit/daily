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

// TodoModel represents the state of the todo TUI
type TodoModel struct {
	todoItems     types.TodoItems
	selectedItem  int
	width         int
	height        int
	styles        *CommonStyles
	allItems      []TodoListItem // flattened list of all items for navigation
	leftViewport  viewportState
	rightViewport viewportState
	glamourStyle  *glamour.TermRenderer
}

// TodoListItem represents an item in the navigation list
type TodoListItem struct {
	Item        types.TodoItem
	Type        string // "open_pr", "pending_review", "assigned_ticket"
	DisplayText string
}

// NewTodoModel creates a new todo TUI model
func NewTodoModel(todoItems types.TodoItems) TodoModel {
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

	model := TodoModel{
		todoItems:    todoItems,
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

func (m *TodoModel) buildItemsList() {
	m.allItems = []TodoListItem{}

	// Add open PRs
	for _, item := range m.todoItems.GitHub.OpenPRs {
		m.allItems = append(m.allItems, TodoListItem{
			Item:        item,
			Type:        "open_pr",
			DisplayText: fmt.Sprintf("🐙 %s", item.Title),
		})
	}

	// Add pending reviews
	for _, item := range m.todoItems.GitHub.PendingReviews {
		m.allItems = append(m.allItems, TodoListItem{
			Item:        item,
			Type:        "pending_review",
			DisplayText: fmt.Sprintf("👁️ %s", item.Title),
		})
	}

	// Add assigned tickets
	for _, item := range m.todoItems.JIRA.AssignedTickets {
		m.allItems = append(m.allItems, TodoListItem{
			Item:        item,
			Type:        "assigned_ticket",
			DisplayText: fmt.Sprintf("🎫 %s", item.Title),
		})
	}

	// Add Obsidian tasks
	for _, item := range m.todoItems.Obsidian.Tasks {
		m.allItems = append(m.allItems, TodoListItem{
			Item:        item,
			Type:        "obsidian_task",
			DisplayText: fmt.Sprintf("📝 %s", item.Title),
		})
	}

	// Sort by updated time (most recent first)
	sort.Slice(m.allItems, func(i, j int) bool {
		return m.allItems[i].Item.UpdatedAt.After(m.allItems[j].Item.UpdatedAt)
	})
}

func (m TodoModel) Init() tea.Cmd {
	return nil
}

func (m TodoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.selectedItem < len(m.allItems) && m.allItems[m.selectedItem].Item.URL != "" {
				url := m.allItems[m.selectedItem].Item.URL
				return m, tea.Exec(urlCommand{url: url}, nil)
			}
			return m, nil
		}
	}
	return m, nil
}

func (m *TodoModel) updateLeftViewport() {
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

func (m TodoModel) View() string {
	// Check if terminal is too small
	if !IsTerminalSizeAdequate(m.width, m.height) {
		return RenderTerminalTooSmallMessage(m.styles, m.width, m.height)
	}

	if len(m.allItems) == 0 {
		return m.styles.Base.Render(
			m.styles.Header.Render("📋 Todo Items") + "\n" +
				"No pending items found.\n\n" +
				m.styles.Help.Render("Press 'q' to quit"),
		)
	}

	// Calculate panel dimensions
	dimensions := CalculatePanelDimensions(m.width)
	if dimensions.UseSingle {
		return m.renderSinglePanelView()
	}

	// Header
	title := fmt.Sprintf("📋 Todo Items (%d)", len(m.allItems))
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

func (m TodoModel) renderLeftPanel(width int) string {
	// Create bordered panel with theme-appropriate colors
	_, borderColor, _, _, _, _ := GetThemeColors()
	leftStyle := CreateBorderedPanel(width, m.leftViewport.height, borderColor)

	var content strings.Builder

	// Navigation help
	helpText := "↑/↓ j/k: Navigate • Enter: Open URL • q: Quit"
	adjustedWidth := max(20, width) // Same adjustment as in CreateBorderedPanel
	content.WriteString(RenderHelpText(helpText, adjustedWidth-4))
	content.WriteString("\n\n")

	// Todo items list
	end := min(len(m.allItems), m.leftViewport.offset+m.leftViewport.height-4) // Account for help text and padding

	for i := m.leftViewport.offset; i < end; i++ {
		item := m.allItems[i]
		isSelected := i == m.selectedItem

		// Create todo item display
		timeStr := item.Item.UpdatedAt.Format("Jan 2")

		// Get appropriate icon for item type
		var icon string
		switch item.Type {
		case "open_pr":
			icon = "🔀"
		case "pending_review":
			icon = "👁️"
		case "assigned_ticket":
			icon = "🎯"
		default:
			icon = "📋"
		}

		// Truncate title to fit width
		maxTitleWidth := max(5, adjustedWidth-15) // Account for time, icons, and padding
		title := TruncateText(item.Item.Title, maxTitleWidth)

		var line strings.Builder
		line.WriteString(fmt.Sprintf("%s %s %s", timeStr, icon, title))

		if item.Item.URL != "" {
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

func (m TodoModel) renderRightPanel(width int) string {
	// Create bordered panel with theme-appropriate colors
	_, borderColor, _, _, _, _ := GetThemeColors()
	rightStyle := CreateBorderedPanel(width, m.rightViewport.height, borderColor)
	adjustedWidth := max(30, width) // Same adjustment as in CreateBorderedPanel

	if m.selectedItem >= len(m.allItems) {
		return rightStyle.Render("Select a todo item to view details")
	}

	selectedItem := m.allItems[m.selectedItem]

	// Create markdown content for the selected todo item
	markdown := m.createTodoMarkdownContent(selectedItem)

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

func (m TodoModel) createTodoMarkdownContent(item TodoListItem) string {
	var md strings.Builder

	// Title
	md.WriteString(fmt.Sprintf("# %s\n\n", item.Item.Title))

	// Metadata table
	md.WriteString("## Details\n\n")
	md.WriteString("| Field | Value |\n")
	md.WriteString("|-------|-------|\n")
	md.WriteString(fmt.Sprintf("| **Updated** | %s |\n", item.Item.UpdatedAt.Format("Jan 2, 2006 15:04")))

	// Type-specific information
	switch item.Type {
	case "open_pr":
		md.WriteString("| **Type** | 🔀 Open Pull Request |\n")
	case "pending_review":
		md.WriteString("| **Type** | 👁️ Pending Review |\n")
	case "assigned_ticket":
		md.WriteString("| **Type** | 🎯 Assigned Ticket |\n")
	default:
		md.WriteString("| **Type** | 📋 Todo Item |\n")
	}

	if item.Item.URL != "" {
		md.WriteString(fmt.Sprintf("| **URL** | [🔗 Open Link](%s) |\n", item.Item.URL))
	}

	// Description
	if item.Item.Description != "" {
		md.WriteString("\n## Description\n\n")
		md.WriteString(item.Item.Description)
		md.WriteString("\n\n")
	}

	// Tags
	if len(item.Item.Tags) > 0 {
		md.WriteString("## Tags\n\n")
		for _, tag := range item.Item.Tags {
			md.WriteString(fmt.Sprintf("- `%s`\n", tag))
		}
		md.WriteString("\n")
	}

	// Additional metadata
	md.WriteString("## Metadata\n\n")
	md.WriteString(fmt.Sprintf("- **ID**: `%s`\n", item.Item.ID))

	return md.String()
}

func (m TodoModel) renderSinglePanelView() string {
	var content strings.Builder

	// Header
	title := fmt.Sprintf("📋 Todo Items (%d)", len(m.allItems))
	content.WriteString(RenderHeader(title, m.width))
	content.WriteString("\n")

	// Navigation help
	helpText := "↑/↓ j/k: Navigate • Enter: Open URL • q: Quit"
	content.WriteString(RenderHelpText(helpText, m.width))
	content.WriteString("\n\n")

	// Todo items list (simplified)
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

		// Simple todo item line
		timeStr := item.Item.UpdatedAt.Format("Jan 2")

		// Get appropriate icon for item type
		var icon string
		switch item.Type {
		case "open_pr":
			icon = "🔀"
		case "pending_review":
			icon = "👁️"
		case "assigned_ticket":
			icon = "🎯"
		default:
			icon = "📋"
		}

		// Truncate title to fit
		maxTitleWidth := max(5, m.width-15)
		title := TruncateText(item.Item.Title, maxTitleWidth)

		line := fmt.Sprintf("%s %s %s", timeStr, icon, title)
		if item.Item.URL != "" {
			line += " 🔗"
		}

		content.WriteString(ApplySelectionStyle(line, isSelected, m.width))
		content.WriteString("\n")
	}

	// Show current item details if space available
	if m.selectedItem < len(m.allItems) && m.height > 15 {
		item := m.allItems[m.selectedItem]
		content.WriteString("\n")
		if item.Item.Description != "" {
			desc := TruncateText(item.Item.Description, m.width-4)
			_, _, _, _, _, scrollColor := GetThemeColors() // Reuse scroll color for description
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

// RunTodoTUI starts the todo TUI application
func RunTodoTUI(todoItems types.TodoItems) error {
	model := NewTodoModel(todoItems)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, err := p.Run()
	return err
}
