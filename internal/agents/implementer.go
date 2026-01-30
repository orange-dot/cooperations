package agents

import (
	"context"
	"strings"
	"time"

	"cooperations/internal/adapters"
	"cooperations/internal/types"
)

// ImplementerAgent handles code implementation using Codex CLI with full agentic access.
type ImplementerAgent struct {
	cli *adapters.CodexCLI
}

// NewImplementerAgent creates a new Implementer agent with Codex CLI.
func NewImplementerAgent(cli *adapters.CodexCLI) *ImplementerAgent {
	return &ImplementerAgent{cli: cli}
}

// Role returns the agent's role.
func (a *ImplementerAgent) Role() types.Role {
	return types.RoleImplementer
}

// Execute runs the implementer agent with full repo access.
func (a *ImplementerAgent) Execute(ctx context.Context, handoff types.Handoff) (types.AgentResponse, error) {
	start := time.Now()

	prompt := buildCodexPrompt(handoff)

	resp, err := a.cli.Execute(ctx, prompt)
	if err != nil {
		return types.AgentResponse{}, err
	}

	nextRole := parseNextRole(resp.Content)
	fileBlocks := parseCodexFileBlocks(resp.Content)
	cleanCode := sanitizeCodexOutput(resp.Content)
	files := map[string]string{}
	if len(fileBlocks) > 0 {
		cleanCode = strings.TrimSpace(fileBlocks[0].content)
		for _, block := range fileBlocks {
			if block.path == "" {
				continue
			}
			files[block.path] = strings.TrimRight(block.content, "\n")
		}
	}

	artifacts := map[string]any{"code": cleanCode}
	if len(files) > 0 {
		artifacts["files"] = files
	}

	return types.AgentResponse{
		Content:    resp.Content,
		Artifacts:  artifacts,
		TokensUsed: resp.TokensUsed,
		DurationMS: time.Since(start).Milliseconds(),
		NextRole:   nextRole,
	}, nil
}
