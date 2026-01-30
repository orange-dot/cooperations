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
4. Any important constraints or considerations

After your design, indicate the next step:
- If implementation is needed, say "NEXT: implementer"
- If review is needed first, say "NEXT: reviewer"
- If the task is complete, say "NEXT: done"`

// ArchitectAgent handles system design tasks.
type ArchitectAgent struct {
	BaseAgent
}

// NewArchitectAgent creates a new Architect agent.
func NewArchitectAgent(adapter adapters.Adapter) *ArchitectAgent {
	return &ArchitectAgent{
		BaseAgent: BaseAgent{
			role:         types.RoleArchitect,
			adapter:      adapter,
			systemPrompt: architectSystemPrompt,
		},
	}
}

// Execute runs the architect agent.
func (a *ArchitectAgent) Execute(ctx context.Context, handoff types.Handoff) (types.AgentResponse, error) {
	start := time.Now()

	prompt := a.buildPrompt(handoff)
	contextText := a.buildContext(handoff)

	resp, err := a.adapter.Complete(ctx, prompt, contextText)
	if err != nil {
		return types.AgentResponse{}, err
	}

	// Parse next role from response
	nextRole := parseNextRole(resp.Content)

	return types.AgentResponse{
		Content:    resp.Content,
		Artifacts:  map[string]any{"design_doc": resp.Content},
		TokensUsed: resp.TokensUsed,
		DurationMS: time.Since(start).Milliseconds(),
		NextRole:   nextRole,
	}, nil
}
