package agents

import (
	"context"
	"time"

	"cooperations/internal/adapters"
	"cooperations/internal/types"
)

const architectSystemPrompt = `You are a software architect in a mob programming team.

Your responsibilities:
- Design systems and high-level structure
- Define API contracts and interfaces
- Select appropriate patterns and enforce consistency
- Document technical decisions

Output format:
Provide your design in a clear, structured format including:
1. Overview of the approach
2. Key interfaces/types (in Go)
3. File structure if applicable
4. Any important constraints or considerations`

// ArchitectAgent handles system design tasks using Claude CLI.
type ArchitectAgent struct {
	cli *adapters.ClaudeCLI
}

// NewArchitectAgent creates a new Architect agent with Claude CLI.
func NewArchitectAgent(cli *adapters.ClaudeCLI) *ArchitectAgent {
	return &ArchitectAgent{cli: cli}
}

// Role returns the agent's role.
func (a *ArchitectAgent) Role() types.Role {
	return types.RoleArchitect
}

// Execute runs the architect agent.
func (a *ArchitectAgent) Execute(ctx context.Context, handoff types.Handoff) (types.AgentResponse, error) {
	start := time.Now()

	prompt := buildClaudePrompt(architectSystemPrompt, handoff)

	resp, err := a.cli.Execute(ctx, prompt)
	if err != nil {
		return types.AgentResponse{}, err
	}

	nextRole := parseNextRole(resp.Content)

	return types.AgentResponse{
		Content:    resp.Content,
		Artifacts:  map[string]any{"design_doc": resp.Content},
		TokensUsed: resp.TokensUsed,
		DurationMS: time.Since(start).Milliseconds(),
		NextRole:   nextRole,
	}, nil
}
