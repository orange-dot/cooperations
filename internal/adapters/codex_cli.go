package adapters

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"cooperations/internal/types"
)

const (
	defaultCodexCLIBinary  = "codex"
	defaultCodexCLITimeout = 10 * time.Minute // Longer timeout for agentic tasks
)

// CodexCLI implements CLI interface for Codex CLI with full agentic access.
// Used for implementer agent with full repo access.
type CodexCLI struct {
	binaryPath string
	timeout    time.Duration
	workDir    string // Repository root directory
}

// NewCodexCLI creates a new Codex CLI executor with full agentic access.
func NewCodexCLI(workDir string) (*CodexCLI, error) {
	binaryPath := os.Getenv("CODEX_CLI_PATH")
	if binaryPath == "" {
		binaryPath = defaultCodexCLIBinary
	}

	// Verify CLI exists
	path, err := exec.LookPath(binaryPath)
	if err != nil {
		return nil, fmt.Errorf("codex CLI not found: %w (install with: npm install -g @openai/codex)", err)
	}

	// Use current directory if workDir not specified
	if workDir == "" {
		workDir, _ = os.Getwd()
	}

	return &CodexCLI{
		binaryPath: path,
		timeout:    defaultCodexCLITimeout,
		workDir:    workDir,
	}, nil
}

// Name returns the CLI identifier.
func (c *CodexCLI) Name() string {
	return "codex-cli"
}

// Execute runs Codex CLI with the given prompt in full agentic mode.
func (c *CodexCLI) Execute(ctx context.Context, prompt string) (types.CLIResponse, error) {
	var resp types.CLIResponse

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Build command with write access and no approvals:
	// codex exec "prompt" --json --full-auto --sandbox workspace-write --ask-for-approval never -C workdir
	cmd := exec.CommandContext(execCtx, c.binaryPath,
		"exec", prompt,
		"--json",
		"--full-auto",
		"--sandbox", "workspace-write",
		"--ask-for-approval", "never",
		"-C", c.workDir,                              // Working directory = repo root
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if execCtx.Err() == context.DeadlineExceeded {
			return resp, fmt.Errorf("codex CLI timed out after %v", c.timeout)
		}
		return resp, fmt.Errorf("codex CLI failed: %w\nstderr: %s", err, stderr.String())
	}

	// Parse JSONL output (one JSON object per line)
	output := stdout.String()
	if output == "" {
		return resp, fmt.Errorf("codex CLI returned empty output")
	}

	var content strings.Builder
	var tokensUsed int

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var event codexCLIEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue // Skip malformed lines
		}

		switch event.Type {
		case "item.completed":
			// Extract agent message content
			if event.Item.Type == "agent_message" {
				if content.Len() > 0 {
					content.WriteString("\n")
				}
				content.WriteString(event.Item.Text)
			}
		case "turn.completed":
			// Extract token usage
			tokensUsed = event.Usage.InputTokens + event.Usage.OutputTokens
		}
	}

	if content.Len() == 0 {
		// Fallback: return raw output if no agent_message found
		resp.Content = output
	} else {
		resp.Content = content.String()
	}
	resp.Model = "codex-cli"
	resp.TokensUsed = tokensUsed

	return resp, nil
}

// Codex CLI JSONL event types

type codexCLIEvent struct {
	Type     string        `json:"type"`
	ThreadID string        `json:"thread_id,omitempty"`
	Item     codexCLIItem  `json:"item,omitempty"`
	Usage    codexCLIUsage `json:"usage,omitempty"`
}

type codexCLIItem struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type"` // "reasoning", "agent_message", etc.
	Text string `json:"text,omitempty"`
}

type codexCLIUsage struct {
	InputTokens       int `json:"input_tokens"`
	CachedInputTokens int `json:"cached_input_tokens"`
	OutputTokens      int `json:"output_tokens"`
}
