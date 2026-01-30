package agents

import (
	"context"
	"time"

	"cooperations/internal/adapters"
	"cooperations/internal/types"
)

const reviewerSystemPrompt = `You are a code reviewer in a mob programming team.

Your responsibilities:
- Review code for correctness and style
- Identify security vulnerabilities
- Analyze performance implications
- Suggest improvements

Guidelines:
- Be specific about issues and their locations
- Explain why something is a problem
- Provide concrete suggestions for fixes
- Prioritize issues by severity (high/medium/low)

Output format:
Structure your review as:
1. Summary (overall assessment)
2. Issues (if any, with severity)
3. Suggestions (improvements)
4. Verdict: APPROVED or CHANGES_NEEDED

After your review:
- If changes are needed, say "NEXT: implementer"
- If approved, say "NEXT: done"`

// ReviewerAgent handles code review tasks.
type ReviewerAgent struct {
	BaseAgent
}

// NewReviewerAgent creates a new Reviewer agent.
func NewReviewerAgent(adapter adapters.Adapter) *ReviewerAgent {
	return &ReviewerAgent{
		BaseAgent: BaseAgent{
			role:         types.RoleReviewer,
			adapter:      adapter,
			systemPrompt: reviewerSystemPrompt,
		},
	}
}

// Execute runs the reviewer agent.
func (a *ReviewerAgent) Execute(ctx context.Context, handoff types.Handoff) (types.AgentResponse, error) {
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
		Artifacts:  map[string]any{"review_feedback": resp.Content},
		TokensUsed: resp.TokensUsed,
		DurationMS: time.Since(start).Milliseconds(),
		NextRole:   nextRole,
	}, nil
}
