package agents

import (
	"context"
	"time"

	"cooperations/internal/adapters"
	"cooperations/internal/types"
)

const navigatorSystemPrompt = `You are a navigator in a mob programming team.

Your responsibilities:
- Track context across the session
- Suggest next steps when the team is stuck
- Identify blockers and dependencies
- Maintain focus on the task goal

Guidelines:
- Summarize the current state clearly
- Identify what's blocking progress
- Suggest concrete next actions
- Ask clarifying questions if requirements are unclear

Output format:
1. Current State: What has been done
2. Blockers: What's preventing progress (if any)
3. Next Steps: Recommended actions
4. Questions: Any clarifications needed from the user`

// NavigatorAgent handles context tracking using Claude CLI.
type NavigatorAgent struct {
	cli *adapters.ClaudeCLI
}

// NewNavigatorAgent creates a new Navigator agent with Claude CLI.
func NewNavigatorAgent(cli *adapters.ClaudeCLI) *NavigatorAgent {
	return &NavigatorAgent{cli: cli}
}

// Role returns the agent's role.
func (a *NavigatorAgent) Role() types.Role {
	return types.RoleNavigator
}

// Execute runs the navigator agent.
func (a *NavigatorAgent) Execute(ctx context.Context, handoff types.Handoff) (types.AgentResponse, error) {
	start := time.Now()

	prompt := buildClaudePrompt(navigatorSystemPrompt, handoff)

	resp, err := a.cli.Execute(ctx, prompt)
	if err != nil {
		return types.AgentResponse{}, err
	}

	nextRole := parseNextRole(resp.Content)

	return types.AgentResponse{
		Content:    resp.Content,
		Artifacts:  map[string]any{"notes": resp.Content},
		TokensUsed: resp.TokensUsed,
		DurationMS: time.Since(start).Milliseconds(),
		NextRole:   nextRole,
	}, nil
}
