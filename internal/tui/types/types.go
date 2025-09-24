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
