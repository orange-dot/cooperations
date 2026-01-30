// Package types defines shared types for the cooperations orchestrator.
package types

// Role represents an agent role in the mob programming workflow.
type Role string

const (
	RoleArchitect   Role = "architect"
	RoleImplementer Role = "implementer"
	RoleReviewer    Role = "reviewer"
	RoleNavigator   Role = "navigator"
)

// Model represents an AI model provider.
type Model string

const (
	ModelClaude Model = "claude-opus-4-5"
	ModelCodex  Model = "codex-5-2"
)

// Task represents a unit of work in the system.
type Task struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	Status      string `json:"status"` // pending, in_progress, completed, failed
}

// TaskStatus constants.
const (
	TaskStatusPending    = "pending"
	TaskStatusInProgress = "in_progress"
	TaskStatusCompleted  = "completed"
	TaskStatusFailed     = "failed"
)

// Handoff represents a context transfer between agents.
type Handoff struct {
	TaskID    string     `json:"task_id" validate:"required"`
	Timestamp string     `json:"timestamp" validate:"required"`
	FromRole  Role       `json:"from_role" validate:"required"`
	ToRole    Role       `json:"to_role" validate:"required"`
	Context   HContext   `json:"context" validate:"required"`
	Artifacts HArtifacts `json:"artifacts"`
	Metadata  HMetadata  `json:"metadata" validate:"required"`
}

// HContext holds the task context passed between agents.
type HContext struct {
	TaskDescription string   `json:"task_description"`
	Requirements    []string `json:"requirements"`
	Constraints     []string `json:"constraints"`
	FilesInScope    []string `json:"files_in_scope"`
}

// HArtifacts holds outputs produced by agents.
type HArtifacts struct {
	DesignDoc      string   `json:"design_doc,omitempty"`
	Interfaces     []string `json:"interfaces,omitempty"`
	Code           string   `json:"code,omitempty"`
	ReviewFeedback string   `json:"review_feedback,omitempty"`
	Notes          string   `json:"notes,omitempty"`
}

// HMetadata holds execution metadata.
type HMetadata struct {
	TokensUsed int    `json:"tokens_used"`
	Model      string `json:"model"`
	DurationMS int64  `json:"duration_ms"`
}

// AgentResponse is the output from an agent execution.
type AgentResponse struct {
	Content    string         `json:"content"`
	Artifacts  map[string]any `json:"artifacts"`
	TokensUsed int            `json:"tokens_used"`
	DurationMS int64          `json:"duration_ms"`
	NextRole   *Role          `json:"next_role,omitempty"`
}

// WorkflowState tracks the current state of a workflow execution.
type WorkflowState struct {
	Task         Task      `json:"task"`
	Handoffs     []Handoff `json:"handoffs"`
	CurrentRole  Role      `json:"current_role"`
	ReviewCycles int       `json:"review_cycles"`
}

// WorkflowResult is the final output of a workflow execution.
type WorkflowResult struct {
	Task      Task      `json:"task"`
	Handoffs  []Handoff `json:"handoffs"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
	Artifacts HArtifacts `json:"artifacts"`
}

// AdapterResponse is the normalized response from a model adapter.
type AdapterResponse struct {
	Content    string `json:"content"`
	TokensUsed int    `json:"tokens_used"`
	Model      string `json:"model"`
}
