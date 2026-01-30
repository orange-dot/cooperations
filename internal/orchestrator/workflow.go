package orchestrator

import (
	"context"
	"fmt"
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

		// Emit code update if code was generated
		if artifacts.Code != "" && state.CurrentRole == types.RoleImplementer {
			o.emitCode("generated/code.go", artifacts.Code, "go")
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
	case types.RoleArchitect, types.RoleReviewer:
		return types.ModelClaude
	case types.RoleImplementer:
		return types.ModelCodex
	default:
		return types.ModelClaude
	}
}
