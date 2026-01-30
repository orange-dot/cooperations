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
4. Questions: Any clarifications needed from the user

After your analysis:
- If design is needed, say "NEXT: architect"
- If implementation is needed, say "NEXT: implementer"
- If review is needed, say "NEXT: reviewer"
- If user input is needed, say "NEXT: user"
- If the task is complete, say "NEXT: done"`

// NavigatorAgent handles context tracking and navigation.
type NavigatorAgent struct {
	BaseAgent
}

// NewNavigatorAgent creates a new Navigator agent.
func NewNavigatorAgent(adapter adapters.Adapter) *NavigatorAgent {
	return &NavigatorAgent{
		BaseAgent: BaseAgent{
			role:         types.RoleNavigator,
			adapter:      adapter,
			systemPrompt: navigatorSystemPrompt,
		},
	}
}

// Execute runs the navigator agent.
func (a *NavigatorAgent) Execute(ctx context.Context, handoff types.Handoff) (types.AgentResponse, error) {
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
		Artifacts:  map[string]any{"notes": resp.Content},
		TokensUsed: resp.TokensUsed,
		DurationMS: time.Since(start).Milliseconds(),
		NextRole:   nextRole,
	}, nil
}
