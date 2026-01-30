// Package adapters provides CLI adapters for the orchestrator.
package adapters

import (
	"context"

	"cooperations/internal/types"
)

// CLI is the interface for CLI-based AI tool execution.
type CLI interface {
	// Execute runs the CLI with the given prompt and returns the response.
	Execute(ctx context.Context, prompt string) (types.CLIResponse, error)

	// Name returns the CLI identifier.
	Name() string
}
