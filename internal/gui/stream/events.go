// File: internal/gui/stream/events.go
package stream

import "time"

// DecisionAction describes what a human chose to do in response to a DecisionRequest.
type DecisionAction string

const (
	DecisionActionApprove DecisionAction = "approve"
	DecisionActionReject  DecisionAction = "reject"
	DecisionActionEdit    DecisionAction = "edit"
)

// ProgressUpdate represents an incremental update for a long-running operation.
type ProgressUpdate struct {
	Percent float64 `json:"percent"`
	Stage   string  `json:"stage"`
	Message string  `json:"message"`
}

// CodeUpdate represents new or updated code content for a file.
type CodeUpdate struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Language string `json:"language"`
}

// HandoffEvent represents a transition event in the workflow (e.g., agent-to-human or phase changes).
type HandoffEvent struct {
	From      string    `json:"from"`
	To        string    `json:"to"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

// TokenUpdate provides token usage information for streaming displays.
type TokenUpdate struct {
	PromptTokens     int `json:"promptTokens"`
	CompletionTokens int `json:"completionTokens"`
	TotalTokens      int `json:"totalTokens"`
}

// DecisionRequest asks a human to make a decision with optional suggested options.
type DecisionRequest struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Prompt  string   `json:"prompt"`
	Options []string `json:"options"`
}

// HumanDecision represents the human's response to a DecisionRequest.
type HumanDecision struct {
	RequestID string         `json:"requestId"`
	Action    DecisionAction `json:"action"`
	Comment   string         `json:"comment"`
	Edited    string         `json:"edited"`
}