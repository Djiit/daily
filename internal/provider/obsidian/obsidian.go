package obsidian

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"daily/internal/activity"
	"daily/internal/provider"
)

type Provider struct {
	config provider.Config
	vaultPath string
}

func NewProvider(config provider.Config) *Provider {
	return &Provider{
		config: config,
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