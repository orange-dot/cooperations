// Package adapters provides model API adapters for the orchestrator.
package adapters

import (
	"context"

	"cooperations/internal/types"
)

// Adapter is the interface for AI model adapters.
type Adapter interface {
	// Model returns the model identifier.
	Model() types.Model

	// Complete sends a prompt to the model and returns the response.
	Complete(ctx context.Context, prompt string, contextText string) (types.AdapterResponse, error)
}
