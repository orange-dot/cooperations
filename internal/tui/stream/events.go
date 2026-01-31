// Package stream provides event types for TUI workflow streaming.
package stream

import "time"

// TokenChunk represents a single token or small chunk from AI streaming response.
type TokenChunk struct {
	AgentRole string    `json:"agent_role"`
	Token     string    `json:"token"`
	Timestamp time.Time `json:"timestamp"`
	IsFinal   bool      `json:"is_final"`
}

// ThinkingUpdate indicates an agent is processing.
type ThinkingUpdate struct {
	AgentRole string        `json:"agent_role"`
	Stage     string        `json:"stage"` // "analyzing", "generating", "reviewing"
	Duration  time.Duration `json:"duration"`
}

// ProgressUpdate represents workflow progress.
type ProgressUpdate struct {
	Percent float64 `json:"percent"`
	Stage   string  `json:"stage"`
	Message string  `json:"message"`
}

// HandoffEvent represents a transition between agents.
type HandoffEvent struct {
	From      string    `json:"from"`
	To        string    `json:"to"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

// AgentLogEntry is a detailed log entry from an agent.
type AgentLogEntry struct {
	Timestamp time.Time      `json:"timestamp"`
	AgentRole string         `json:"agent_role"`
	Level     string         `json:"level"` // "info", "debug", "warn", "error"
	Message   string         `json:"message"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// CodeUpdate represents new or updated code content.
type CodeUpdate struct {
	Path     string `json:"path"`
	Content  string `json:"content"`
	Language string `json:"language"`
}

// FileDiff represents a git-style diff for a file.
type FileDiff struct {
	Path       string     `json:"path"`
	OldContent string     `json:"old_content"`
	NewContent string     `json:"new_content"`
	Hunks      []DiffHunk `json:"hunks"`
}

// DiffHunk represents a section of changes in a diff.
type DiffHunk struct {
	OldStart int        `json:"old_start"`
	OldCount int        `json:"old_count"`
	NewStart int        `json:"new_start"`
	NewCount int        `json:"new_count"`
	Lines    []DiffLine `json:"lines"`
}

// DiffLine represents a single line in a diff.
type DiffLine struct {
	Type    string `json:"type"` // "add", "remove", "context"
	Content string `json:"content"`
}

// FileTreeUpdate represents a change in the generated file tree.
type FileTreeUpdate struct {
	Action string `json:"action"` // "add", "modify", "delete"
	Path   string `json:"path"`
	IsDir  bool   `json:"is_dir"`
	Size   int64  `json:"size"`
}

// MetricsSnapshot contains live metrics data.
type MetricsSnapshot struct {
	TotalTokens      int           `json:"total_tokens"`
	PromptTokens     int           `json:"prompt_tokens"`
	CompletionTokens int           `json:"completion_tokens"`
	EstimatedCostUSD float64       `json:"estimated_cost_usd"`
	ElapsedTime      time.Duration `json:"elapsed_time"`
	APICallsCount    int           `json:"api_calls_count"`
	AgentCycles      int           `json:"agent_cycles"`
	CurrentAgent     string        `json:"current_agent"`
}

// ToastNotification is a non-blocking notification.
type ToastNotification struct {
	ID       string        `json:"id"`
	Level    string        `json:"level"` // "info", "success", "warning", "error"
	Title    string        `json:"title"`
	Message  string        `json:"message"`
	Duration time.Duration `json:"duration"`
}

// DecisionRequest asks a human to make a decision.
type DecisionRequest struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Prompt  string   `json:"prompt"`
	Options []string `json:"options"`
}

// DecisionAction describes what action was taken.
type DecisionAction string

const (
	DecisionApprove DecisionAction = "approve"
	DecisionReject  DecisionAction = "reject"
	DecisionEdit    DecisionAction = "edit"
)

// HumanDecision is the human's response to a DecisionRequest.
type HumanDecision struct {
	RequestID string         `json:"request_id"`
	Action    DecisionAction `json:"action"`
	Comment   string         `json:"comment"`
	Edited    string         `json:"edited"`
}

// SessionEvent represents session management events.
type SessionEvent struct {
	Type      string    `json:"type"` // "checkpoint", "save", "load", "replay_start", "replay_end"
	SessionID string    `json:"session_id"`
	Timestamp time.Time `json:"timestamp"`
	Data      any       `json:"data,omitempty"`
}

// ControlSignal represents workflow control commands from the UI.
type ControlSignal string

const (
	ControlStep   ControlSignal = "step"   // Execute one agent, then pause
	ControlSkip   ControlSignal = "skip"   // Skip current agent, go to next
	ControlKill   ControlSignal = "kill"   // Immediate abort with cleanup
	ControlPause  ControlSignal = "pause"  // Pause at next hook point
	ControlResume ControlSignal = "resume" // Resume execution
)

// ControlEvent represents a control signal from the UI to the orchestrator.
type ControlEvent struct {
	Signal    ControlSignal `json:"signal"`
	Timestamp time.Time     `json:"timestamp"`
	Reason    string        `json:"reason,omitempty"`
}

// HookPhase identifies when a hook runs (mirrors orchestrator.HookPhase for TUI).
type HookPhase string

const (
	HookPhaseWorkflowStart HookPhase = "workflow_start"
	HookPhasePreAgent      HookPhase = "pre_agent"
	HookPhaseMidAgent      HookPhase = "mid_agent"
	HookPhasePostAgent     HookPhase = "post_agent"
	HookPhasePreHandoff    HookPhase = "pre_handoff"
	HookPhasePostHandoff   HookPhase = "post_handoff"
	HookPhaseWorkflowEnd   HookPhase = "workflow_end"
)

// HookNotification represents a hook being triggered (for UI display).
type HookNotification struct {
	Phase     HookPhase `json:"phase"`
	TaskID    string    `json:"task_id"`
	Role      string    `json:"role"`
	Timestamp time.Time `json:"timestamp"`
	Paused    bool      `json:"paused"`
	CanSkip   bool      `json:"can_skip"`
}

// RVREvent represents RVR processing state for TUI display.
type RVREvent struct {
	Phase       string  `json:"phase"` // "layer1", "layer2", "synthesis"
	ChunkID     int     `json:"chunk_id"`
	Confidence  float64 `json:"confidence"`
	Threshold   float64 `json:"threshold"`
	Uncertainty string  `json:"uncertainty,omitempty"`
	Retrying    bool    `json:"retrying"`
}

// RVRResultEvent represents final RVR results for TUI display.
type RVRResultEvent struct {
	TaskType   string              `json:"task_type"`
	Overall    float64             `json:"overall"`
	Breakdown  []RVRBreakdownItem  `json:"breakdown"`
	Caveats    []string            `json:"caveats"`
}

// RVRBreakdownItem represents a single chunk's RVR result.
type RVRBreakdownItem struct {
	ChunkID     int     `json:"chunk_id"`
	Confidence  float64 `json:"confidence"`
	Uncertainty string  `json:"uncertainty"`
	Verified    bool    `json:"verified"`
}
