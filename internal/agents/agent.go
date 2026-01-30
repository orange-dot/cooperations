// Package agents implements role-specific agents for the orchestrator.
package agents

import (
	"context"
	"regexp"
	"strings"

	"cooperations/internal/types"
)

// Agent is the interface for role-specific agents.
type Agent interface {
	// Role returns the agent's role.
	Role() types.Role

	// Execute runs the agent with the given handoff context.
	Execute(ctx context.Context, handoff types.Handoff) (types.AgentResponse, error)
}

// buildClaudePrompt constructs a full prompt for Claude CLI agents.
// Includes system prompt, artifacts, and task.
func buildClaudePrompt(systemPrompt string, handoff types.Handoff) string {
	var b strings.Builder

	// System prompt
	b.WriteString(systemPrompt)
	b.WriteString("\n\n")

	// Previous artifacts
	if handoff.Artifacts.DesignDoc != "" {
		b.WriteString("## Design Document\n")
		b.WriteString(handoff.Artifacts.DesignDoc)
		b.WriteString("\n\n")
	}
	if handoff.Artifacts.Code != "" {
		b.WriteString("## Current Code\n```go\n")
		b.WriteString(handoff.Artifacts.Code)
		b.WriteString("\n```\n\n")
	}
	if handoff.Artifacts.ReviewFeedback != "" {
		b.WriteString("## Review Feedback\n")
		b.WriteString(handoff.Artifacts.ReviewFeedback)
		b.WriteString("\n\n")
	}
	if handoff.Artifacts.Notes != "" {
		b.WriteString("## Navigator Notes\n")
		b.WriteString(handoff.Artifacts.Notes)
		b.WriteString("\n\n")
	}

	// Requirements and constraints
	if len(handoff.Context.Requirements) > 0 {
		b.WriteString("## Requirements\n")
		for _, req := range handoff.Context.Requirements {
			b.WriteString("- ")
			b.WriteString(req)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	if len(handoff.Context.Constraints) > 0 {
		b.WriteString("## Constraints\n")
		for _, con := range handoff.Context.Constraints {
			b.WriteString("- ")
			b.WriteString(con)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Task description
	b.WriteString("## Task\n")
	b.WriteString(handoff.Context.TaskDescription)
	b.WriteString("\n\n")

	// Instruction for next step
	b.WriteString("After completing, indicate the next step with: NEXT: <role> (architect/implementer/reviewer/navigator) or NEXT: done")

	return b.String()
}

// buildCodexPrompt constructs a direct prompt for Codex CLI.
// Keep it simple - Codex works best with clear, direct instructions.
func buildCodexPrompt(handoff types.Handoff) string {
	var b strings.Builder

	// Direct task - no preamble
	b.WriteString(handoff.Context.TaskDescription)

	// Add design doc if available
	if handoff.Artifacts.DesignDoc != "" {
		b.WriteString("\n\nFollow this design:\n")
		b.WriteString(handoff.Artifacts.DesignDoc)
	}

	// Add review feedback if available
	if handoff.Artifacts.ReviewFeedback != "" {
		b.WriteString("\n\nAddress this feedback:\n")
		b.WriteString(handoff.Artifacts.ReviewFeedback)
	}

	// Instructions for Codex CLI
	b.WriteString("\n\nYou are running in Codex CLI with write access to this repo.")
	b.WriteString("\nMake the required changes directly in the workspace (create/modify files).")
	b.WriteString("\nIf the task mentions a file path, your response will be written to that path.")
	b.WriteString("\nIf you need to change multiple files, output them in this exact format:")
	b.WriteString("\nFILE: path/to/file.ext\n<full file contents>\nEND_FILE")
	b.WriteString("\n(repeat for each file). If only one file, output just the full file contents.")
	b.WriteString("\nIf you need to signal completion, add a final line: NEXT: done")
	b.WriteString("\nDo not mention sandbox or limitations.")
	b.WriteString("\nNo markdown fences, no explanations, no extra text.")

	return b.String()
}

var nextLinePattern = regexp.MustCompile(`(?i)^NEXT:\s*(architect|implementer|reviewer|navigator|done|user)\s*$`)
var fileHeaderPattern = regexp.MustCompile(`(?i)^FILE:\s*(.+)$`)
var endFilePattern = regexp.MustCompile(`(?i)^END_FILE\s*$`)

// sanitizeCodexOutput extracts code from fenced blocks and strips trailing NEXT lines.
func sanitizeCodexOutput(content string) string {
	if code, ok := extractFirstCodeBlock(content); ok {
		return strings.TrimSpace(code)
	}

	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return trimmed
	}

	lines := strings.Split(trimmed, "\n")
	if len(lines) > 0 && nextLinePattern.MatchString(strings.TrimSpace(lines[len(lines)-1])) {
		lines = lines[:len(lines)-1]
	}

	return strings.TrimSpace(strings.Join(lines, "\n"))
}

type fileBlock struct {
	path    string
	content string
}

func parseCodexFileBlocks(content string) []fileBlock {
	lines := strings.Split(content, "\n")
	blocks := make([]fileBlock, 0)
	var current *fileBlock

	flush := func() {
		if current == nil {
			return
		}
		current.content = strings.TrimRight(current.content, "\n")
		blocks = append(blocks, *current)
		current = nil
	}

	for _, line := range lines {
		if match := fileHeaderPattern.FindStringSubmatch(line); match != nil {
			flush()
			path := strings.TrimSpace(match[1])
			if path == "" {
				continue
			}
			current = &fileBlock{path: path}
			continue
		}
		if current != nil && endFilePattern.MatchString(strings.TrimSpace(line)) {
			flush()
			continue
		}
		if current != nil {
			current.content += line + "\n"
		}
	}

	flush()

	return blocks
}

// extractFirstCodeBlock returns the first fenced code block if present.
func extractFirstCodeBlock(content string) (string, bool) {
	start := strings.Index(content, "```")
	if start == -1 {
		return "", false
	}

	afterFence := start + 3
	if afterFence >= len(content) {
		return "", false
	}

	lineEnd := strings.Index(content[afterFence:], "\n")
	if lineEnd == -1 {
		return "", false
	}
	codeStart := afterFence + lineEnd + 1

	end := strings.Index(content[codeStart:], "```")
	if end == -1 {
		return "", false
	}

	return content[codeStart : codeStart+end], true
}
