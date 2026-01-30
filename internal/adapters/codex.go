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
	codexAPIURL = "https://api.openai.com/v1/chat/completions"
	codexModel  = "gpt-5.2-2025-12-11" // Codex 5.2
)

// CodexAdapter implements Adapter for Codex 5.2.
type CodexAdapter struct {
	apiKey     string
	httpClient *http.Client
	maxRetries int
}

// NewCodexAdapter creates a new Codex adapter.
func NewCodexAdapter() (*CodexAdapter, error) {
	apiKey := os.Getenv("CODEX_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("CODEX_API_KEY environment variable not set")
	}

	return &CodexAdapter{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		maxRetries: 3,
	}, nil
}

// Model returns the model identifier.
func (c *CodexAdapter) Model() types.Model {
	return types.ModelCodex
}

// Complete sends a prompt to Codex and returns the response.
func (c *CodexAdapter) Complete(ctx context.Context, prompt string, contextText string) (types.AdapterResponse, error) {
	fullPrompt := prompt
	if contextText != "" {
		fullPrompt = fmt.Sprintf("Context:\n%s\n\nTask:\n%s", contextText, prompt)
	}

	reqBody := openAIRequest{
		Model: codexModel,
		Messages: []openAIMessage{
			{Role: "user", Content: fullPrompt},
		},
		MaxCompletionTokens: 4096,
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

	return resp, fmt.Errorf("codex api failed after %d retries: %w", c.maxRetries, lastErr)
}

func (c *CodexAdapter) doRequest(ctx context.Context, reqBody openAIRequest) (types.AdapterResponse, error) {
	var resp types.AdapterResponse

	body, err := json.Marshal(reqBody)
	if err != nil {
		return resp, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", codexAPIURL, bytes.NewReader(body))
	if err != nil {
		return resp, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

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

	var openAIResp openAIResponse
	if err := json.Unmarshal(respBody, &openAIResp); err != nil {
		return resp, fmt.Errorf("unmarshal response: %w", err)
	}

	if len(openAIResp.Choices) > 0 {
		resp.Content = openAIResp.Choices[0].Message.Content
	}

	resp.Model = openAIResp.Model
	resp.TokensUsed = openAIResp.Usage.TotalTokens

	return resp, nil
}

// OpenAI API types

type openAIRequest struct {
	Model               string          `json:"model"`
	Messages            []openAIMessage `json:"messages"`
	MaxCompletionTokens int             `json:"max_completion_tokens"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []openAIChoice `json:"choices"`
	Model   string         `json:"model"`
	Usage   openAIUsage    `json:"usage"`
}

type openAIChoice struct {
	Message openAIMessage `json:"message"`
}

type openAIUsage struct {
	TotalTokens int `json:"total_tokens"`
}
