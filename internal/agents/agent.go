// Package agents implements role-specific agents for the orchestrator.
package agents

import (
	"context"

	"cooperations/internal/adapters"
	"cooperations/internal/types"
)

// Agent is the interface for role-specific agents.
type Agent interface {
	// Role returns the agent's role.
	Role() types.Role

	// Execute runs the agent with the given handoff context.
	Execute(ctx context.Context, handoff types.Handoff) (types.AgentResponse, error)
}

// BaseAgent provides common functionality for agents.
type BaseAgent struct {
	role         types.Role
	adapter      adapters.Adapter
	systemPrompt string
}

// Role returns the agent's role.
func (a *BaseAgent) Role() types.Role {
	return a.role
}

// buildPrompt constructs the full prompt from handoff context.
func (a *BaseAgent) buildPrompt(handoff types.Handoff) string {
	return handoff.Context.TaskDescription
}

// buildContext constructs the context string from handoff.
func (a *BaseAgent) buildContext(handoff types.Handoff) string {
	ctx := a.systemPrompt + "\n\n"

	if len(handoff.Context.Requirements) > 0 {
		ctx += "Requirements:\n"
		for _, req := range handoff.Context.Requirements {
			ctx += "- " + req + "\n"
		}
		ctx += "\n"
	}

	if len(handoff.Context.Constraints) > 0 {
		ctx += "Constraints:\n"
		for _, con := range handoff.Context.Constraints {
			ctx += "- " + con + "\n"
		}
		ctx += "\n"
	}

	if len(handoff.Context.FilesInScope) > 0 {
		ctx += "Files in scope:\n"
		for _, f := range handoff.Context.FilesInScope {
			ctx += "- " + f + "\n"
		}
		ctx += "\n"
	}

	// Include previous artifacts
	if handoff.Artifacts.DesignDoc != "" {
		ctx += "Design Document:\n" + handoff.Artifacts.DesignDoc + "\n\n"
	}
	if handoff.Artifacts.Code != "" {
		ctx += "Current Code:\n" + handoff.Artifacts.Code + "\n\n"
	}
	if handoff.Artifacts.ReviewFeedback != "" {
		ctx += "Review Feedback:\n" + handoff.Artifacts.ReviewFeedback + "\n\n"
	}
	if handoff.Artifacts.Notes != "" {
		ctx += "Notes:\n" + handoff.Artifacts.Notes + "\n\n"
	}

	return ctx
}
