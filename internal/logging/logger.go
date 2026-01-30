// Package logging provides structured logging for the cooperations orchestrator.
package logging

import (
	"log/slog"
	"os"
	"strings"
)

// Setup initializes the global logger with the specified level.
func Setup(level string) {
	var logLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn", "warning":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	})
	slog.SetDefault(slog.New(handler))
}

// Route logs a routing decision.
func Route(task string, role string, reason string) {
	slog.Info("task routed",
		"role", role,
		"reason", reason,
		"task_preview", truncate(task, 50),
	)
}

// AgentStart logs the start of agent execution.
func AgentStart(role string, taskID string) {
	slog.Info("agent executing",
		"role", role,
		"task_id", taskID,
	)
}

// AgentComplete logs completion of agent execution.
func AgentComplete(role string, taskID string, durationMS int64, tokens int) {
	slog.Info("agent completed",
		"role", role,
		"task_id", taskID,
		"duration_ms", durationMS,
		"tokens_used", tokens,
	)
}

// Handoff logs a context handoff between agents.
func Handoff(from string, to string, taskID string) {
	slog.Info("handoff",
		"from", from,
		"to", to,
		"task_id", taskID,
	)
}

// WorkflowComplete logs workflow completion.
func WorkflowComplete(taskID string, success bool, cycles int) {
	slog.Info("workflow completed",
		"task_id", taskID,
		"success", success,
		"review_cycles", cycles,
	)
}

// Error logs an error with context.
func Error(msg string, err error, attrs ...any) {
	args := append([]any{"error", err}, attrs...)
	slog.Error(msg, args...)
}

// Info logs an informational message with context.
func Info(msg string, attrs ...any) {
	slog.Info(msg, attrs...)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
