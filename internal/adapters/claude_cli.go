package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"cooperations/internal/types"
)

const (
	defaultClaudeCLIBinary  = "claude"
	defaultClaudeCLITimeout = 5 * time.Minute
)

// ClaudeCLI implements CLI interface for Claude Code CLI.
// Used for architect, reviewer, and navigator agents.
type ClaudeCLI struct {
	binaryPath string
	timeout    time.Duration
}

// NewClaudeCLI creates a new Claude CLI executor.
func NewClaudeCLI() (*ClaudeCLI, error) {
	binaryPath := os.Getenv("CLAUDE_CLI_PATH")
	if binaryPath == "" {
		binaryPath = defaultClaudeCLIBinary
	}

	// Verify CLI exists
	path, err := exec.LookPath(binaryPath)
	if err != nil {
		return nil, fmt.Errorf("claude CLI not found: %w (install from https://claude.ai/code)", err)
	}

	return &ClaudeCLI{
		binaryPath: path,
		timeout:    defaultClaudeCLITimeout,
	}, nil
}

// Name returns the CLI identifier.
func (c *ClaudeCLI) Name() string {
	return "claude-cli"
}

// Execute runs Claude CLI with the given prompt.
func (c *ClaudeCLI) Execute(ctx context.Context, prompt string) (types.CLIResponse, error) {
	var resp types.CLIResponse

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Build command:
	// claude -p "prompt" --output-format json --max-turns 1
	cmd := exec.CommandContext(execCtx, c.binaryPath,
		"-p", prompt,
		"--output-format", "json",
		"--max-turns", "1",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if execCtx.Err() == context.DeadlineExceeded {
			return resp, fmt.Errorf("claude CLI timed out after %v", c.timeout)
		}
		return resp, fmt.Errorf("claude CLI failed: %w\nstderr: %s", err, stderr.String())
	}

	// Parse JSON output
	output := stdout.Bytes()
	if len(output) == 0 {
		return resp, fmt.Errorf("claude CLI returned empty output")
	}

	var cliResp claudeCLIResponse
	if err := json.Unmarshal(output, &cliResp); err != nil {
		// Fallback: use raw output as content
		resp.Content = string(output)
		resp.Model = "claude-cli"
		resp.TokensUsed = 0
		return resp, nil
	}

	// Check for error response
	if cliResp.IsError || cliResp.Subtype == "error" {
		return resp, fmt.Errorf("claude CLI returned error: %s", cliResp.Result)
	}

	resp.Content = cliResp.Result
	resp.Model = "claude-cli"
	resp.TokensUsed = cliResp.totalTokens()

	return resp, nil
}

// claudeCLIResponse represents the JSON output from Claude Code CLI.
type claudeCLIResponse struct {
	Type        string         `json:"type"`
	Subtype     string         `json:"subtype"`
	IsError     bool           `json:"is_error"`
	Result      string         `json:"result"`
	DurationMS  int64          `json:"duration_ms"`
	NumTurns    int            `json:"num_turns"`
	SessionID   string         `json:"session_id,omitempty"`
	TotalCostUS float64        `json:"total_cost_usd"`
	Usage       claudeCLIUsage `json:"usage"`
}

type claudeCLIUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

func (r *claudeCLIResponse) totalTokens() int {
	return r.Usage.InputTokens + r.Usage.OutputTokens
}
