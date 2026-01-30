package orchestrator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	ctx "cooperations/internal/context"
	"cooperations/internal/gui/stream"
	"cooperations/internal/logging"
	"cooperations/internal/types"
)

// WorkflowConfig holds workflow execution settings.
type WorkflowConfig struct {
	MaxReviewCycles int
}

// DefaultWorkflowConfig returns the default workflow configuration.
func DefaultWorkflowConfig() WorkflowConfig {
	return WorkflowConfig{
		MaxReviewCycles: 2,
	}
}

// emitProgress sends a progress update to the stream if available.
func (o *Orchestrator) emitProgress(stage string, percent float64, message string) {
	if o.stream == nil {
		return
	}
	select {
	case o.stream.Progress <- stream.ProgressUpdate{
		Stage:   stage,
		Percent: percent,
		Message: message,
	}:
	default:
		// Channel full, skip
	}
}

// emitHandoff sends a handoff event to the stream if available.
func (o *Orchestrator) emitHandoff(from, to, reason string) {
	if o.stream == nil {
		return
	}
	select {
	case o.stream.Handoffs <- stream.HandoffEvent{
		From:      from,
		To:        to,
		Reason:    reason,
		Timestamp: time.Now(),
	}:
	default:
	}
}

// emitTokens sends a token update to the stream if available.
func (o *Orchestrator) emitTokens(prompt, completion, total int) {
	if o.stream == nil {
		return
	}
	select {
	case o.stream.Tokens <- stream.TokenUpdate{
		PromptTokens:     prompt,
		CompletionTokens: completion,
		TotalTokens:      total,
	}:
	default:
	}
}

// emitCode sends a code update to the stream if available.
func (o *Orchestrator) emitCode(path, content, language string) {
	if o.stream == nil {
		return
	}
	select {
	case o.stream.Code <- stream.CodeUpdate{
		Path:     path,
		Content:  content,
		Language: language,
	}:
	default:
	}
}

// emitError sends an error to the stream if available.
func (o *Orchestrator) emitError(err error) {
	if o.stream == nil {
		return
	}
	select {
	case o.stream.Error <- err:
	default:
	}
}

// emitDone signals workflow completion to the stream if available.
func (o *Orchestrator) emitDone() {
	if o.stream == nil {
		return
	}
	select {
	case o.stream.Done <- struct{}{}:
	default:
	}
}

// executeWorkflow runs the main workflow loop.
func (o *Orchestrator) executeWorkflow(c context.Context, task types.Task, initialRole types.Role) (types.WorkflowResult, error) {
	state := types.WorkflowState{
		Task:         task,
		Handoffs:     []types.Handoff{},
		CurrentRole:  initialRole,
		ReviewCycles: 0,
	}

	// Create initial handoff context
	handoffCtx := types.HContext{
		TaskDescription: task.Description,
		Requirements:    []string{},
		Constraints:     []string{},
		FilesInScope:    []string{},
	}
	artifacts := types.HArtifacts{}

	// Track total tokens for stream updates
	totalTokens := 0
	stepCount := 0

	// Emit initial progress
	o.emitProgress("Starting", 0, fmt.Sprintf("Starting workflow for task: %s", task.ID))
	o.emitHandoff("user", string(initialRole), "Initial routing")

	for {
		// Check for context cancellation
		select {
		case <-c.Done():
			return types.WorkflowResult{
				Task:     task,
				Handoffs: state.Handoffs,
				Success:  false,
				Error:    "workflow cancelled",
			}, c.Err()
		default:
		}

		// Get the agent for current role
		agent, ok := o.agents[state.CurrentRole]
		if !ok {
			return types.WorkflowResult{
				Task:     task,
				Handoffs: state.Handoffs,
				Success:  false,
				Error:    fmt.Sprintf("no agent for role: %s", state.CurrentRole),
			}, fmt.Errorf("no agent for role: %s", state.CurrentRole)
		}

		// Create handoff for this step
		handoff := ctx.NewHandoff(
			task.ID,
			state.CurrentRole, // from (will be updated after execution)
			state.CurrentRole, // to (current)
			handoffCtx,
			artifacts,
			types.HMetadata{},
		)

		// Execute the agent
		logging.AgentStart(string(state.CurrentRole), task.ID)
		stepCount++

		// Emit progress before execution
		roleLabel := roleToLabel(state.CurrentRole)
		o.emitProgress(roleLabel, float64(stepCount*20), fmt.Sprintf("%s is working...", roleLabel))

		response, err := agent.Execute(c, *handoff)
		if err != nil {
			logging.Error("agent execution failed", err, "role", state.CurrentRole, "task_id", task.ID)
			o.emitError(err)
			return types.WorkflowResult{
				Task:     task,
				Handoffs: state.Handoffs,
				Success:  false,
				Error:    err.Error(),
			}, err
		}

		logging.AgentComplete(string(state.CurrentRole), task.ID, response.DurationMS, response.TokensUsed)

		// Emit token update
		totalTokens += response.TokensUsed
		o.emitTokens(response.TokensUsed/2, response.TokensUsed/2, totalTokens)

		// Update handoff with execution metadata
		handoff.Metadata = types.HMetadata{
			TokensUsed: response.TokensUsed,
			Model:      string(o.getModelForRole(state.CurrentRole)),
			DurationMS: response.DurationMS,
		}

		// Merge artifacts
		artifacts = ctx.MergeArtifacts(artifacts, response.Artifacts)

		// Update handoff with merged artifacts
		handoff.Artifacts = artifacts

		// Save artifacts to generated folder
		if state.CurrentRole == types.RoleImplementer {
			files := extractFiles(response.Artifacts)
			if len(files) > 0 {
				for rawPath, content := range files {
					cleanPath := cleanRelativePath(rawPath)
					if cleanPath == "" {
						logging.Error("invalid file path in response", nil, "task_id", task.ID, "path", rawPath)
						continue
					}

					path, err := o.store.SaveGeneratedCode(task.ID, cleanPath, content)
					if err != nil {
						logging.Error("failed to save code artifact", err, "task_id", task.ID)
					} else {
						logging.Info("saved code artifact", "path", path)
						o.emitCode(cleanPath, content, detectLanguage(cleanPath))
					}

					if err := writeWorkspaceFile(cleanPath, content); err != nil {
						logging.Error("failed to write task output", err, "task_id", task.ID, "path", cleanPath)
					} else {
						logging.Info("wrote task output", "path", cleanPath)
					}
				}
			} else if artifacts.Code != "" {
				path, err := o.store.SaveGeneratedCode(task.ID, "", artifacts.Code)
				if err != nil {
					logging.Error("failed to save code artifact", err, "task_id", task.ID)
				} else {
					logging.Info("saved code artifact", "path", path)
					o.emitCode(path, artifacts.Code, "go")
				}

				if targetPath := extractTargetPath(task.Description); targetPath != "" {
					if err := writeWorkspaceFile(targetPath, artifacts.Code); err != nil {
						logging.Error("failed to write task output", err, "task_id", task.ID, "path", targetPath)
					} else {
						logging.Info("wrote task output", "path", targetPath)
					}
				}
			}
		}
		if artifacts.DesignDoc != "" && state.CurrentRole == types.RoleArchitect {
			path, err := o.store.SaveDesignDoc(task.ID, artifacts.DesignDoc)
			if err != nil {
				logging.Error("failed to save design doc", err, "task_id", task.ID)
			} else {
				logging.Info("saved design doc", "path", path)
			}
		}
		if artifacts.ReviewFeedback != "" && state.CurrentRole == types.RoleReviewer {
			path, err := o.store.SaveReviewFeedback(task.ID, artifacts.ReviewFeedback)
			if err != nil {
				logging.Error("failed to save review feedback", err, "task_id", task.ID)
			} else {
				logging.Info("saved review feedback", "path", path)
			}
		}

		// Determine next role
		var nextRole *types.Role
		if response.NextRole != nil {
			nextRole = response.NextRole
		}

		// Update handoff with next role
		if nextRole != nil {
			handoff.ToRole = *nextRole
		}

		// Save handoff
		state.Handoffs = append(state.Handoffs, *handoff)
		if err := o.store.SaveHandoff(task.ID, *handoff); err != nil {
			logging.Error("failed to save handoff", err, "task_id", task.ID)
		}

		// Check if workflow is complete
		if nextRole == nil {
			logging.WorkflowComplete(task.ID, true, state.ReviewCycles)
			o.emitProgress("Complete", 100, "Workflow completed successfully")
			o.emitDone()

			// Save task summary to generated folder
			task.Status = types.TaskStatusCompleted
			if path, err := o.store.SaveTaskSummary(task.ID, task, artifacts); err != nil {
				logging.Error("failed to save task summary", err, "task_id", task.ID)
			} else {
				logging.Info("saved task summary", "path", path)
			}

			return types.WorkflowResult{
				Task:      task,
				Handoffs:  state.Handoffs,
				Success:   true,
				Artifacts: artifacts,
			}, nil
		}

		// Check review cycle limit
		if *nextRole == types.RoleReviewer {
			state.ReviewCycles++
			if state.ReviewCycles > o.config.MaxReviewCycles {
				logging.WorkflowComplete(task.ID, false, state.ReviewCycles)
				return types.WorkflowResult{
					Task:      task,
					Handoffs:  state.Handoffs,
					Success:   false,
					Error:     fmt.Sprintf("exceeded max review cycles (%d)", o.config.MaxReviewCycles),
					Artifacts: artifacts,
				}, nil
			}
		}

		// Transition to next role
		logging.Handoff(string(state.CurrentRole), string(*nextRole), task.ID)

		// Emit handoff event
		o.emitHandoff(string(state.CurrentRole), string(*nextRole), fmt.Sprintf("Transitioning to %s", roleToLabel(*nextRole)))

		state.CurrentRole = *nextRole

		// Update context for next iteration
		handoffCtx.TaskDescription = response.Content
	}
}

// roleToLabel converts a role to a human-readable label.
func roleToLabel(role types.Role) string {
	switch role {
	case types.RoleArchitect:
		return "Architect"
	case types.RoleImplementer:
		return "Implementer"
	case types.RoleReviewer:
		return "Reviewer"
	case types.RoleNavigator:
		return "Navigator"
	case types.RoleHuman:
		return "Human"
	default:
		return string(role)
	}
}

// getModelForRole returns the model used by a role.
func (o *Orchestrator) getModelForRole(role types.Role) types.Model {
	switch role {
	case types.RoleArchitect, types.RoleReviewer, types.RoleNavigator:
		return types.ModelClaudeCLI
	case types.RoleImplementer:
		return types.ModelCodexCLI
	default:
		return types.ModelClaudeCLI
	}
}

var taskPathPattern = regexp.MustCompile(`(?i)([A-Za-z0-9][A-Za-z0-9_./\\-]*\.[A-Za-z0-9]{1,6})`)

func extractTargetPath(task string) string {
	match := taskPathPattern.FindStringSubmatch(task)
	if len(match) < 2 {
		return ""
	}

	candidate := strings.Trim(match[1], "`\"'()[]{}<>.,;:")
	return cleanRelativePath(candidate)
}

func extractFiles(artifacts map[string]any) map[string]string {
	raw, ok := artifacts["files"]
	if !ok {
		return nil
	}

	switch v := raw.(type) {
	case map[string]string:
		return v
	case map[string]any:
		out := make(map[string]string, len(v))
		for key, value := range v {
			str, ok := value.(string)
			if !ok {
				continue
			}
			out[key] = str
		}
		return out
	default:
		return nil
	}
}

func cleanRelativePath(candidate string) string {
	if candidate == "" {
		return ""
	}

	clean := filepath.Clean(candidate)
	if clean == "." || clean == ".." {
		return ""
	}
	if filepath.IsAbs(clean) {
		return ""
	}
	if strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return ""
	}
	if strings.Contains(clean, ":") {
		return ""
	}

	return clean
}

func detectLanguage(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".go":
		return "go"
	case ".js", ".mjs", ".cjs":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".py":
		return "python"
	case ".md":
		return "markdown"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".toml":
		return "toml"
	default:
		return "text"
	}
}

func writeWorkspaceFile(relPath string, content string) error {
	dir := filepath.Dir(relPath)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create target directory: %w", err)
		}
	}

	if err := os.WriteFile(relPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write target file: %w", err)
	}

	return nil
}
