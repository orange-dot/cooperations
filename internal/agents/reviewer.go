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
4. Verdict: APPROVED or CHANGES_NEEDED`

// ReviewerAgent handles code review using Claude CLI.
type ReviewerAgent struct {
	cli *adapters.ClaudeCLI
}

// NewReviewerAgent creates a new Reviewer agent with Claude CLI.
func NewReviewerAgent(cli *adapters.ClaudeCLI) *ReviewerAgent {
	return &ReviewerAgent{cli: cli}
}

// Role returns the agent's role.
func (a *ReviewerAgent) Role() types.Role {
	return types.RoleReviewer
}

// Execute runs the reviewer agent.
func (a *ReviewerAgent) Execute(ctx context.Context, handoff types.Handoff) (types.AgentResponse, error) {
	start := time.Now()

	prompt := buildClaudePrompt(reviewerSystemPrompt, handoff)

	resp, err := a.cli.Execute(ctx, prompt)
	if err != nil {
		return types.AgentResponse{}, err
	}

	nextRole := parseNextRole(resp.Content)

	return types.AgentResponse{
		Content:    resp.Content,
		Artifacts:  map[string]any{"review_feedback": resp.Content},
		TokensUsed: resp.TokensUsed,
		DurationMS: time.Since(start).Milliseconds(),
		NextRole:   nextRole,
	}, nil
}
