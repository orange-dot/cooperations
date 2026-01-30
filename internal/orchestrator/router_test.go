package orchestrator

import (
	"testing"

	"cooperations/internal/types"
)

func TestRouter_Route(t *testing.T) {
	router := NewRouter()

	tests := []struct {
		name     string
		task     string
		expected types.Role
	}{
		// Architect keywords
		{"design keyword", "design a user authentication system", types.RoleArchitect},
		{"architect keyword", "architect the API layer", types.RoleArchitect},
		{"plan keyword", "plan the database schema", types.RoleArchitect},
		{"api keyword", "define the API interface", types.RoleArchitect},

		// Reviewer keywords
		{"review keyword", "review this code for bugs", types.RoleReviewer},
		{"check keyword", "check for security issues", types.RoleReviewer},
		{"verify keyword", "verify the implementation", types.RoleReviewer},
		{"audit keyword", "audit the codebase", types.RoleReviewer},
		{"security keyword", "security analysis needed", types.RoleReviewer},

		// Navigator keywords
		{"help keyword", "help me understand this code", types.RoleNavigator},
		{"stuck keyword", "I'm stuck on this problem", types.RoleNavigator},
		{"context keyword", "what's the context here", types.RoleNavigator},
		{"status keyword", "what's the status", types.RoleNavigator},

		// Implementer keywords
		{"implement keyword", "implement the login feature", types.RoleImplementer},
		{"code keyword", "code a new endpoint", types.RoleImplementer},
		{"build keyword", "build the authentication module", types.RoleImplementer},
		{"fix keyword", "fix this bug in the parser", types.RoleImplementer},
		{"add keyword", "add error handling", types.RoleImplementer},

		// Default to Implementer
		{"no keywords", "do something with the database", types.RoleImplementer},
		{"empty task", "", types.RoleImplementer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := router.Route(tt.task)
			if result != tt.expected {
				t.Errorf("Route(%q) = %v, want %v", tt.task, result, tt.expected)
			}
		})
	}
}

func TestRouter_RouteWithConfidence(t *testing.T) {
	router := NewRouter()

	tests := []struct {
		name           string
		task           string
		expectedRole   types.Role
		minConfidence  float64
	}{
		{"strong architect signal", "design and architect the system", types.RoleArchitect, 0.5},
		{"weak signal", "do something", types.RoleImplementer, 0.0},
		{"mixed signals", "design and implement", types.RoleArchitect, 0.3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role, confidence := router.RouteWithConfidence(tt.task)
			if role != tt.expectedRole {
				t.Errorf("RouteWithConfidence(%q) role = %v, want %v", tt.task, role, tt.expectedRole)
			}
			if confidence < tt.minConfidence {
				t.Errorf("RouteWithConfidence(%q) confidence = %v, want >= %v", tt.task, confidence, tt.minConfidence)
			}
		})
	}
}
