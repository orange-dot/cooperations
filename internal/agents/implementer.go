package agents

import (
	"context"
	"time"

	"cooperations/internal/adapters"
	"cooperations/internal/types"
)

const implementerSystemPrompt = `You are a code implementer in a mob programming team.

Your responsibilities:
- Write clean, working code based on specifications
- Generate boilerplate and scaffolding
- Refactor existing code when needed
- Implement features according to the design

Guidelines:
- Write idiomatic Go code
- Include error handling
- Keep functions focused and testable
- Follow existing patterns in the codebase

Output format:
Provide your implementation as complete, runnable code.
Include comments only where the logic isn't self-evident.

After your implementation, indicate the next step:
- If review is needed, say "NEXT: reviewer"
- If the task is complete, say "NEXT: done"`

// ImplementerAgent handles code implementation tasks.
type ImplementerAgent struct {
	BaseAgent
}

// NewImplementerAgent creates a new Implementer agent.
func NewImplementerAgent(adapter adapters.Adapter) *ImplementerAgent {
	return &ImplementerAgent{
		BaseAgent: BaseAgent{
			role:         types.RoleImplementer,
			adapter:      adapter,
			systemPrompt: implementerSystemPrompt,
		},
	}
}

// Execute runs the implementer agent.
func (a *ImplementerAgent) Execute(ctx context.Context, handoff types.Handoff) (types.AgentResponse, error) {
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
		Artifacts:  map[string]any{"code": resp.Content},
		TokensUsed: resp.TokensUsed,
		DurationMS: time.Since(start).Milliseconds(),
		NextRole:   nextRole,
	}, nil
}
