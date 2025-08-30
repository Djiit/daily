package types

import "time"

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
