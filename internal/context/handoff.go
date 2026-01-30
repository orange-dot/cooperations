// Package context handles handoff serialization and storage.
package context

import (
	"encoding/json"
	"fmt"
	"time"

	"cooperations/internal/types"
)

// ValidateHandoff validates a handoff struct.
func ValidateHandoff(h *types.Handoff) error {
	if h.TaskID == "" {
		return fmt.Errorf("task_id is required")
	}
	if h.Timestamp == "" {
		return fmt.Errorf("timestamp is required")
	}
	if h.FromRole == "" {
		return fmt.Errorf("from_role is required")
	}
	if h.ToRole == "" {
		return fmt.Errorf("to_role is required")
	}
	return nil
}

// NewHandoff creates a new handoff with the given parameters.
func NewHandoff(taskID string, from, to types.Role, ctx types.HContext, artifacts types.HArtifacts, meta types.HMetadata) *types.Handoff {
	return &types.Handoff{
		TaskID:    taskID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		FromRole:  from,
		ToRole:    to,
		Context:   ctx,
		Artifacts: artifacts,
		Metadata:  meta,
	}
}

// MarshalHandoff serializes a handoff to JSON.
func MarshalHandoff(h *types.Handoff) ([]byte, error) {
	return json.MarshalIndent(h, "", "  ")
}

// UnmarshalHandoff deserializes a handoff from JSON.
func UnmarshalHandoff(data []byte) (*types.Handoff, error) {
	var h types.Handoff
	if err := json.Unmarshal(data, &h); err != nil {
		return nil, fmt.Errorf("unmarshal handoff: %w", err)
	}
	return &h, nil
}

// MergeArtifacts combines artifacts from a response into existing artifacts.
func MergeArtifacts(existing types.HArtifacts, response map[string]any) types.HArtifacts {
	result := existing

	if v, ok := response["design_doc"].(string); ok && v != "" {
		result.DesignDoc = v
	}
	if v, ok := response["code"].(string); ok && v != "" {
		result.Code = v
	}
	if v, ok := response["review_feedback"].(string); ok && v != "" {
		result.ReviewFeedback = v
	}
	if v, ok := response["notes"].(string); ok && v != "" {
		result.Notes = v
	}
	if v, ok := response["interfaces"].([]string); ok {
		result.Interfaces = v
	}

	return result
}
