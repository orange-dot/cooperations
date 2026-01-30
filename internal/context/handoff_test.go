package context

import (
	"testing"

	"cooperations/internal/types"
)

func TestValidateHandoff(t *testing.T) {
	tests := []struct {
		name    string
		handoff types.Handoff
		wantErr bool
	}{
		{
			name: "valid handoff",
			handoff: types.Handoff{
				TaskID:    "123",
				Timestamp: "2024-01-01T00:00:00Z",
				FromRole:  types.RoleArchitect,
				ToRole:    types.RoleImplementer,
			},
			wantErr: false,
		},
		{
			name: "missing task_id",
			handoff: types.Handoff{
				Timestamp: "2024-01-01T00:00:00Z",
				FromRole:  types.RoleArchitect,
				ToRole:    types.RoleImplementer,
			},
			wantErr: true,
		},
		{
			name: "missing timestamp",
			handoff: types.Handoff{
				TaskID:   "123",
				FromRole: types.RoleArchitect,
				ToRole:   types.RoleImplementer,
			},
			wantErr: true,
		},
		{
			name: "missing from_role",
			handoff: types.Handoff{
				TaskID:    "123",
				Timestamp: "2024-01-01T00:00:00Z",
				ToRole:    types.RoleImplementer,
			},
			wantErr: true,
		},
		{
			name: "missing to_role",
			handoff: types.Handoff{
				TaskID:    "123",
				Timestamp: "2024-01-01T00:00:00Z",
				FromRole:  types.RoleArchitect,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHandoff(&tt.handoff)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateHandoff() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMarshalUnmarshalHandoff(t *testing.T) {
	original := &types.Handoff{
		TaskID:    "test-123",
		Timestamp: "2024-01-01T12:00:00Z",
		FromRole:  types.RoleArchitect,
		ToRole:    types.RoleImplementer,
		Context: types.HContext{
			TaskDescription: "Build a login feature",
			Requirements:    []string{"Must use OAuth"},
			Constraints:     []string{"No external dependencies"},
			FilesInScope:    []string{"auth.go", "login.go"},
		},
		Artifacts: types.HArtifacts{
			DesignDoc: "Design doc content",
			Notes:     "Some notes",
		},
		Metadata: types.HMetadata{
			TokensUsed: 1500,
			Model:      "claude-opus-4-5",
			DurationMS: 5000,
		},
	}

	// Marshal
	data, err := MarshalHandoff(original)
	if err != nil {
		t.Fatalf("MarshalHandoff() error = %v", err)
	}

	// Unmarshal
	result, err := UnmarshalHandoff(data)
	if err != nil {
		t.Fatalf("UnmarshalHandoff() error = %v", err)
	}

	// Compare
	if result.TaskID != original.TaskID {
		t.Errorf("TaskID = %v, want %v", result.TaskID, original.TaskID)
	}
	if result.FromRole != original.FromRole {
		t.Errorf("FromRole = %v, want %v", result.FromRole, original.FromRole)
	}
	if result.ToRole != original.ToRole {
		t.Errorf("ToRole = %v, want %v", result.ToRole, original.ToRole)
	}
	if result.Context.TaskDescription != original.Context.TaskDescription {
		t.Errorf("TaskDescription = %v, want %v", result.Context.TaskDescription, original.Context.TaskDescription)
	}
	if result.Metadata.TokensUsed != original.Metadata.TokensUsed {
		t.Errorf("TokensUsed = %v, want %v", result.Metadata.TokensUsed, original.Metadata.TokensUsed)
	}
}

func TestMergeArtifacts(t *testing.T) {
	existing := types.HArtifacts{
		DesignDoc: "Original design",
		Notes:     "Original notes",
	}

	response := map[string]any{
		"code":            "func main() {}",
		"review_feedback": "Looks good",
	}

	result := MergeArtifacts(existing, response)

	// Check preserved fields
	if result.DesignDoc != "Original design" {
		t.Errorf("DesignDoc = %v, want 'Original design'", result.DesignDoc)
	}
	if result.Notes != "Original notes" {
		t.Errorf("Notes = %v, want 'Original notes'", result.Notes)
	}

	// Check merged fields
	if result.Code != "func main() {}" {
		t.Errorf("Code = %v, want 'func main() {}'", result.Code)
	}
	if result.ReviewFeedback != "Looks good" {
		t.Errorf("ReviewFeedback = %v, want 'Looks good'", result.ReviewFeedback)
	}
}
