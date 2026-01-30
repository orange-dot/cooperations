package orchestrator

import (
	"context"
	"fmt"

	ctx "cooperations/internal/context"
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
		response, err := agent.Execute(c, *handoff)
		if err != nil {
			logging.Error("agent execution failed", err, "role", state.CurrentRole, "task_id", task.ID)
			return types.WorkflowResult{
				Task:     task,
				Handoffs: state.Handoffs,
				Success:  false,
				Error:    err.Error(),
			}, err
		}

		logging.AgentComplete(string(state.CurrentRole), task.ID, response.DurationMS, response.TokensUsed)

		// Update handoff with execution metadata
		handoff.Metadata = types.HMetadata{
			TokensUsed: response.TokensUsed,
			Model:      string(o.getModelForRole(state.CurrentRole)),
			DurationMS: response.DurationMS,
		}

		// Merge artifacts
		artifacts = ctx.MergeArtifacts(artifacts, response.Artifacts)

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
		state.CurrentRole = *nextRole

		// Update context for next iteration
		handoffCtx.TaskDescription = response.Content
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
