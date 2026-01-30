package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"cooperations/internal/types"
)

const (
	claudeAPIURL     = "https://api.anthropic.com/v1/messages"
	claudeModel      = "claude-opus-4-5-20250514"
	claudeAPIVersion = "2023-06-01"
)

// ClaudeAdapter implements Adapter for Claude Opus 4.5.
type ClaudeAdapter struct {
	apiKey     string
	httpClient *http.Client
	maxRetries int
}

// NewClaudeAdapter creates a new Claude adapter.
func NewClaudeAdapter() (*ClaudeAdapter, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
	}

	return &ClaudeAdapter{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		maxRetries: 3,
	}, nil
}

// Model returns the model identifier.
func (c *ClaudeAdapter) Model() types.Model {
	return types.ModelClaude
}

// Complete sends a prompt to Claude and returns the response.
func (c *ClaudeAdapter) Complete(ctx context.Context, prompt string, contextText string) (types.AdapterResponse, error) {
	fullPrompt := prompt
	if contextText != "" {
		fullPrompt = fmt.Sprintf("Context:\n%s\n\nTask:\n%s", contextText, prompt)
	}

	reqBody := claudeRequest{
		Model:     claudeModel,
		MaxTokens: 4096,
		Messages: []claudeMessage{
			{Role: "user", Content: fullPrompt},
		},
	}

	var resp types.AdapterResponse
	var lastErr error

	for attempt := 0; attempt < c.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			select {
			case <-ctx.Done():
				return resp, ctx.Err()
			case <-time.After(backoff):
			}
		}

		resp, lastErr = c.doRequest(ctx, reqBody)
		if lastErr == nil {
			return resp, nil
		}
	}

	return resp, fmt.Errorf("claude api failed after %d retries: %w", c.maxRetries, lastErr)
}

func (c *ClaudeAdapter) doRequest(ctx context.Context, reqBody claudeRequest) (types.AdapterResponse, error) {
	var resp types.AdapterResponse

	body, err := json.Marshal(reqBody)
	if err != nil {
		return resp, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", claudeAPIURL, bytes.NewReader(body))
	if err != nil {
		return resp, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", claudeAPIVersion)

	httpResp, err := c.httpClient.Do(req)
	if err != nil {
		return resp, fmt.Errorf("http request: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return resp, fmt.Errorf("read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return resp, fmt.Errorf("api error (status %d): %s", httpResp.StatusCode, string(respBody))
	}

	var claudeResp claudeResponse
	if err := json.Unmarshal(respBody, &claudeResp); err != nil {
		return resp, fmt.Errorf("unmarshal response: %w", err)
	}

	// Extract text content
	for _, block := range claudeResp.Content {
		if block.Type == "text" {
			resp.Content += block.Text
		}
	}

	resp.Model = claudeResp.Model
	resp.TokensUsed = claudeResp.Usage.InputTokens + claudeResp.Usage.OutputTokens

	return resp, nil
}

// Claude API types

type claudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []claudeMessage `json:"messages"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeResponse struct {
	Content []claudeContent `json:"content"`
	Model   string          `json:"model"`
	Usage   claudeUsage     `json:"usage"`
}

type claudeContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type claudeUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}
