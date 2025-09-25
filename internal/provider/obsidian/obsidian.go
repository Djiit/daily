package obsidian

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"daily/internal/activity"
	"daily/internal/provider"
)

type Provider struct {
	config    provider.Config
	vaultPath string
}

func NewProvider(config provider.Config) *Provider {
	return &Provider{
		config:    config,
		vaultPath: config.URL, // Using URL field to store vault path
	}
}

func (p *Provider) Name() string {
	return "obsidian"
}

func (p *Provider) IsConfigured() bool {
	return p.config.Enabled && p.vaultPath != ""
}

func (p *Provider) GetActivities(ctx context.Context, from, to time.Time) ([]activity.Activity, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("Obsidian provider not configured")
	}

	var activities []activity.Activity

	// Find notes created or modified in the time range
	notes, err := p.findRecentNotes(from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to find recent notes: %w", err)
	}
	activities = append(activities, notes...)

	// Find tasks created or modified in the time range
	tasks, err := p.findRecentTasks(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to find recent tasks: %w", err)
	}
	activities = append(activities, tasks...)

	return activities, nil
}

func (p *Provider) findRecentNotes(from, to time.Time) ([]activity.Activity, error) {
	var activities []activity.Activity

	err := filepath.Walk(p.vaultPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process .md files
		if !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		// Check if file was modified in our time range
		if info.ModTime().Before(from) || info.ModTime().After(to) {
			return nil
		}

		// Create activity for this note
		relPath, _ := filepath.Rel(p.vaultPath, path)
		title := strings.TrimSuffix(info.Name(), ".md")

		activities = append(activities, activity.Activity{
			ID:          fmt.Sprintf("obsidian-%s", relPath),
			Type:        activity.ActivityTypeNote,
			Title:       title,
			Description: fmt.Sprintf("Note: %s", relPath),
			Platform:    "obsidian",
			Timestamp:   info.ModTime(),
		})

		return nil
	})

	return activities, err
}

// findRecentTasks finds tasks that were created or modified within the specified time range
func (p *Provider) findRecentTasks(ctx context.Context, from, to time.Time) ([]activity.Activity, error) {
	var activities []activity.Activity

	err := filepath.Walk(p.vaultPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process .md files
		if !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		// Check if file was modified in our time range
		if info.ModTime().Before(from) || info.ModTime().After(to) {
			return nil
		}

		// Parse tasks from this file and convert to activities
		fileTasks, err := p.parseTasksFromFile(path, info)
		if err != nil {
			return nil // Skip files we can't read
		}

		// Convert TodoItems to Activities
		for _, task := range fileTasks {
			activities = append(activities, activity.Activity{
				ID:          task.ID,
				Type:        activity.ActivityTypeTask,
				Title:       task.Title,
				Description: task.Description,
				URL:         task.URL,
				Platform:    "obsidian",
				Timestamp:   task.UpdatedAt,
				Tags:        task.Tags,
			})
		}

		return nil
	})

	return activities, err
}

// GetTasks retrieves pending tasks from Obsidian markdown files
func (p *Provider) GetTasks(ctx context.Context) ([]TodoItem, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("Obsidian provider not configured")
	}

	var tasks []TodoItem

	err := filepath.Walk(p.vaultPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process .md files
		if !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		// Parse tasks from this file
		fileTasks, err := p.parseTasksFromFile(path, info)
		if err != nil {
			return nil // Skip files we can't read
		}

		tasks = append(tasks, fileTasks...)
		return nil
	})

	return tasks, err
}

// parseTasksFromFile extracts incomplete tasks from a markdown file
func (p *Provider) parseTasksFromFile(filePath string, fileInfo os.FileInfo) ([]TodoItem, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := file.Close(); err != nil {
			// Log the error or handle it appropriately
			// Since this is in a defer, we can't return the error
		}
	}()

	var tasks []TodoItem
	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Regex patterns for supported task formats: [ ], [/], [x]
	todoTaskPattern := regexp.MustCompile(`^\s*[-*+]\s*\[\s\]\s*(.+)$`)
	ongoingTaskPattern := regexp.MustCompile(`^\s*[-*+]\s*\[/\]\s*(.+)$`)
	numberedTodoPattern := regexp.MustCompile(`^\s*\d+\.\s*\[\s\]\s*(.+)$`)
	numberedOngoingPattern := regexp.MustCompile(`^\s*\d+\.\s*\[/\]\s*(.+)$`)
	inCodeBlock := false
	inBlockQuote := false

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		// Skip tasks in code blocks
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock {
			continue
		}

		// Check if we're in a blockquote
		inBlockQuote = strings.HasPrefix(strings.TrimSpace(line), ">")
		if inBlockQuote {
			continue
		}

		// Match todo tasks (- [ ] or * [ ] or + [ ])
		if matches := todoTaskPattern.FindStringSubmatch(line); len(matches) > 1 {
			taskText := strings.TrimSpace(matches[1])
			tasks = append(tasks, p.createTodoItem(taskText, filePath, fileInfo, lineNum))
		}

		// Match ongoing tasks (- [/] or * [/] or + [/])
		if matches := ongoingTaskPattern.FindStringSubmatch(line); len(matches) > 1 {
			taskText := strings.TrimSpace(matches[1])
			tasks = append(tasks, p.createTodoItem(taskText, filePath, fileInfo, lineNum))
		}

		// Match numbered todo tasks (1. [ ])
		if matches := numberedTodoPattern.FindStringSubmatch(line); len(matches) > 1 {
			taskText := strings.TrimSpace(matches[1])
			tasks = append(tasks, p.createTodoItem(taskText, filePath, fileInfo, lineNum))
		}

		// Match numbered ongoing tasks (1. [/])
		if matches := numberedOngoingPattern.FindStringSubmatch(line); len(matches) > 1 {
			taskText := strings.TrimSpace(matches[1])
			tasks = append(tasks, p.createTodoItem(taskText, filePath, fileInfo, lineNum))
		}
	}

	return tasks, scanner.Err()
}

// createTodoItem creates a TodoItem from task text and file info
func (p *Provider) createTodoItem(taskText, filePath string, fileInfo os.FileInfo, lineNum int) TodoItem {
	relPath, _ := filepath.Rel(p.vaultPath, filePath)
	fileName := strings.TrimSuffix(fileInfo.Name(), ".md")

	// Extract tags from task text
	tags := extractTags(taskText)

	return TodoItem{
		ID:          fmt.Sprintf("obsidian-task-%s:%d", relPath, lineNum),
		Title:       taskText,
		Description: fmt.Sprintf("Task in %s", fileName),
		URL:         fmt.Sprintf("obsidian://open?vault=%s&file=%s", filepath.Base(p.vaultPath), relPath),
		UpdatedAt:   fileInfo.ModTime(),
		Tags:        tags,
	}
}

// extractTags extracts hashtags and other markers from task text
func extractTags(text string) []string {
	var tags []string

	// Extract hashtags
	hashtagPattern := regexp.MustCompile(`#([\w-]+)`)
	hashtagMatches := hashtagPattern.FindAllStringSubmatch(text, -1)
	for _, match := range hashtagMatches {
		tags = append(tags, match[1])
	}

	// Extract priority indicators
	if strings.Contains(text, "🔥") || strings.Contains(text, "urgent") {
		tags = append(tags, "high-priority")
	}
	if strings.Contains(text, "⭐") {
		tags = append(tags, "starred")
	}
	if strings.Contains(text, "📅") || strings.Contains(text, "⏰") || strings.Contains(text, "🕐") {
		tags = append(tags, "scheduled")
	}

	return tags
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
