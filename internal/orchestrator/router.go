// Package orchestrator implements the core orchestration logic.
package orchestrator

import (
	"regexp"
	"strings"

	"cooperations/internal/logging"
	"cooperations/internal/types"
)

// Router determines which role should handle a task.
type Router struct{}

// NewRouter creates a new router.
func NewRouter() *Router {
	return &Router{}
}

// Routing patterns
var (
	architectPattern   = regexp.MustCompile(`(?i)\b(design|architect|plan|structure|api|interface|schema)\b`)
	reviewerPattern    = regexp.MustCompile(`(?i)\b(review|check|verify|audit|security|test)\b`)
	navigatorPattern   = regexp.MustCompile(`(?i)\b(help|stuck|context|status|what|where|why|how)\b`)
	implementerPattern = regexp.MustCompile(`(?i)\b(implement|code|build|create|write|add|fix|bug)\b`)
)

// Route determines the initial role for a task based on keywords.
func (r *Router) Route(task string) types.Role {
	lower := strings.ToLower(task)

	var role types.Role
	var reason string

	switch {
	case architectPattern.MatchString(lower):
		role = types.RoleArchitect
		reason = "matched design/architecture keywords"
	case reviewerPattern.MatchString(lower):
		role = types.RoleReviewer
		reason = "matched review/verification keywords"
	case navigatorPattern.MatchString(lower):
		role = types.RoleNavigator
		reason = "matched help/context keywords"
	case implementerPattern.MatchString(lower):
		role = types.RoleImplementer
		reason = "matched implementation keywords"
	default:
		role = types.RoleImplementer
		reason = "default routing (no specific keywords matched)"
	}

	logging.Route(task, string(role), reason)
	return role
}

// RouteWithConfidence returns the role and a confidence score (0-1).
func (r *Router) RouteWithConfidence(task string) (types.Role, float64) {
	lower := strings.ToLower(task)

	// Count keyword matches for confidence
	archMatches := len(architectPattern.FindAllString(lower, -1))
	reviewMatches := len(reviewerPattern.FindAllString(lower, -1))
	navMatches := len(navigatorPattern.FindAllString(lower, -1))
	implMatches := len(implementerPattern.FindAllString(lower, -1))

	total := archMatches + reviewMatches + navMatches + implMatches
	if total == 0 {
		return types.RoleImplementer, 0.3 // Low confidence default
	}

	// Find the highest match count
	maxMatches := max(archMatches, reviewMatches, navMatches, implMatches)
	confidence := float64(maxMatches) / float64(total)

	// Determine role based on highest matches
	switch maxMatches {
	case archMatches:
		return types.RoleArchitect, confidence
	case reviewMatches:
		return types.RoleReviewer, confidence
	case navMatches:
		return types.RoleNavigator, confidence
	default:
		return types.RoleImplementer, confidence
	}
}
