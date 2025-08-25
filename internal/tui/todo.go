package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss/v2"

	"daily/internal/tui/types"
)

// TodoModel represents the state of the todo TUI
type TodoModel struct {
	todoItems    types.TodoItems
	currentView  int // 0: all, 1: open PRs, 2: pending reviews, 3: assigned tickets
	selectedItem int
	width        int
	height       int
	styles       *CommonStyles
	allItems     []TodoListItem // flattened list of all items for navigation
}

// TodoListItem represents an item in the navigation list
type TodoListItem struct {
	Item        types.TodoItem
	Type        string // "open_pr", "pending_review", "assigned_ticket"
	DisplayText string
}

// NewTodoModel creates a new todo TUI model
func NewTodoModel(todoItems types.TodoItems) TodoModel {
	model := TodoModel{
		todoItems: todoItems,
		styles:    NewCommonStyles(),
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
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			m.selectedItem = ClampCursor(m.selectedItem-1, 0, len(m.allItems)-1)
		case "down", "j":
			m.selectedItem = ClampCursor(m.selectedItem+1, 0, len(m.allItems)-1)
		case "home", "g":
			m.selectedItem = 0
		case "end", "G":
			m.selectedItem = len(m.allItems) - 1
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

func (m TodoModel) View() string {
	if len(m.allItems) == 0 {
		return m.styles.Base.Render(
			m.styles.Header.Render("📋 Todo Items") + "\n" +
				"No pending items found.\n\n" +
				m.styles.Help.Render("Press 'q' to quit"),
		)
	}

	var content strings.Builder

	// Header
	totalItems := len(m.allItems)
	header := fmt.Sprintf("📋 Todo Items (%d)", totalItems)
	content.WriteString(m.styles.Header.Render(header))
	content.WriteString("\n\n")

	// Calculate available height for content
	headerHeight := 4 // Header + spacing
	statusBarHeight := 2
	helpHeight := 1
	availableHeight := m.height - headerHeight - statusBarHeight - helpHeight

	// Calculate visible range
	visibleStart := 0
	visibleEnd := len(m.allItems)

	if availableHeight > 0 && len(m.allItems) > availableHeight {
		// Ensure selected item is visible
		if m.selectedItem < visibleStart {
			visibleStart = m.selectedItem
		} else if m.selectedItem >= visibleStart+availableHeight {
			visibleStart = m.selectedItem - availableHeight + 1
		}
		visibleEnd = visibleStart + availableHeight
		if visibleEnd > len(m.allItems) {
			visibleEnd = len(m.allItems)
			visibleStart = visibleEnd - availableHeight
			if visibleStart < 0 {
				visibleStart = 0
			}
		}
	}

	// Display items
	for i := visibleStart; i < visibleEnd; i++ {
		item := m.allItems[i]

		var style lipgloss.Style
		var prefix string

		if i == m.selectedItem {
			style = m.styles.Selected
			prefix = "▶ "
		} else {
			style = m.styles.Unselected
			prefix = "  "
		}

		// Format the item display
		timeStr := m.styles.Time.Render(item.Item.UpdatedAt.Format("Jan 2 15:04"))
		title := item.DisplayText
		if len(title) > 60 {
			title = title[:57] + "..."
		}

		line := fmt.Sprintf("%s%s  %s", prefix, timeStr, title)
		content.WriteString(style.Render(line))
		content.WriteString("\n")
	}

	// Show details for selected item if available
	if m.selectedItem < len(m.allItems) {
		selectedItem := m.allItems[m.selectedItem].Item
		content.WriteString("\n")
		content.WriteString(m.styles.SectionHeader.Render("Details:"))
		content.WriteString("\n")

		if selectedItem.Description != "" {
			content.WriteString(m.styles.Description.Render(selectedItem.Description))
			content.WriteString("\n")
		}

		if selectedItem.URL != "" {
			url := selectedItem.URL
			if len(url) > 80 {
				url = url[:77] + "..."
			}
			content.WriteString(m.styles.URL.Render("🔗 " + url))
			content.WriteString("\n")
		}

		if len(selectedItem.Tags) > 0 {
			tags := strings.Join(selectedItem.Tags, ", ")
			if len(tags) > 80 {
				tags = tags[:77] + "..."
			}
			content.WriteString(m.styles.Tags.Render("🏷️  " + tags))
			content.WriteString("\n")
		}
	}

	// Status bar
	content.WriteString("\n")
	statusText := fmt.Sprintf("Item %d of %d", m.selectedItem+1, totalItems)
	if len(m.allItems) > availableHeight && availableHeight > 0 {
		statusText += fmt.Sprintf(" | Showing %d-%d", visibleStart+1, visibleEnd)
	}
	statusBar := m.styles.StatusBar.Copy().Width(m.width - 2).Render(statusText)
	content.WriteString(statusBar)
	content.WriteString("\n")

	// Help text
	help := "↑/↓ navigate • g/G go to top/bottom • q quit"
	content.WriteString(m.styles.Help.Render(help))

	return m.styles.Base.Render(content.String())
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
