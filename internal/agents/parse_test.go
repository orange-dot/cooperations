package agents

import (
	"testing"

	"cooperations/internal/types"
)

func TestParseNextRole(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected *types.Role
	}{
		{
			name:     "architect",
			content:  "Here is the design.\n\nNEXT: architect",
			expected: rolePtr(types.RoleArchitect),
		},
		{
			name:     "implementer",
			content:  "Ready for implementation.\n\nNEXT: implementer",
			expected: rolePtr(types.RoleImplementer),
		},
		{
			name:     "reviewer",
			content:  "Please review.\n\nNEXT: reviewer",
			expected: rolePtr(types.RoleReviewer),
		},
		{
			name:     "navigator",
			content:  "Need context.\n\nNEXT: navigator",
			expected: rolePtr(types.RoleNavigator),
		},
		{
			name:     "done",
			content:  "Task complete.\n\nNEXT: done",
			expected: nil,
		},
		{
			name:     "user",
			content:  "Need clarification.\n\nNEXT: user",
			expected: nil,
		},
		{
			name:     "no next role",
			content:  "Just some output without direction.",
			expected: nil,
		},
		{
			name:     "case insensitive",
			content:  "Done.\n\nnext: ARCHITECT",
			expected: rolePtr(types.RoleArchitect),
		},
		{
			name:     "with extra whitespace",
			content:  "Output.\n\nNEXT:   implementer  ",
			expected: rolePtr(types.RoleImplementer),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseNextRole(tt.content)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("parseNextRole(%q) = %v, want nil", tt.content, *result)
				}
			} else {
				if result == nil {
					t.Errorf("parseNextRole(%q) = nil, want %v", tt.content, *tt.expected)
				} else if *result != *tt.expected {
					t.Errorf("parseNextRole(%q) = %v, want %v", tt.content, *result, *tt.expected)
				}
			}
		})
	}
}

func rolePtr(r types.Role) *types.Role {
	return &r
}
