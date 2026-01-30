package agents

import (
	"regexp"
	"strings"

	"cooperations/internal/types"
)

var nextRolePattern = regexp.MustCompile(`(?i)NEXT:\s*(architect|implementer|reviewer|navigator|done|user)`)

// parseNextRole extracts the next role from agent response content.
func parseNextRole(content string) *types.Role {
	matches := nextRolePattern.FindStringSubmatch(content)
	if len(matches) < 2 {
		return nil
	}

	roleStr := strings.ToLower(matches[1])
	switch roleStr {
	case "architect":
		role := types.RoleArchitect
		return &role
	case "implementer":
		role := types.RoleImplementer
		return &role
	case "reviewer":
		role := types.RoleReviewer
		return &role
	case "navigator":
		role := types.RoleNavigator
		return &role
	case "done", "user":
		// nil indicates workflow should end
		return nil
	default:
		return nil
	}
}
