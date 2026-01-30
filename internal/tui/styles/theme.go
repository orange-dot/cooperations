// Package styles provides theming for the TUI.
package styles

import "github.com/charmbracelet/lipgloss"

// Theme defines the color scheme and styles for TUI.
type Theme struct {
	// Base colors
	Background    lipgloss.Color
	Foreground    lipgloss.Color
	Muted         lipgloss.Color
	Border        lipgloss.Color

	// Accent colors
	Primary       lipgloss.Color
	Secondary     lipgloss.Color
	Accent        lipgloss.Color

	// Status colors
	Success       lipgloss.Color
	Warning       lipgloss.Color
	Error         lipgloss.Color
	Info          lipgloss.Color

	// Agent colors
	AgentArchitect   lipgloss.Color
	AgentImplementer lipgloss.Color
	AgentReviewer    lipgloss.Color
	AgentNavigator   lipgloss.Color
}

// Neon is the default cyberpunk neon theme.
var Neon = Theme{
	Background:    lipgloss.Color("#0a0e17"),
	Foreground:    lipgloss.Color("#e0e0e0"),
	Muted:         lipgloss.Color("#4a5568"),
	Border:        lipgloss.Color("#1a2332"),

	Primary:       lipgloss.Color("#00ffff"), // Cyan
	Secondary:     lipgloss.Color("#ff00ff"), // Magenta
	Accent:        lipgloss.Color("#00ff88"), // Neon green

	Success:       lipgloss.Color("#00ff88"),
	Warning:       lipgloss.Color("#ffaa00"),
	Error:         lipgloss.Color("#ff4466"),
	Info:          lipgloss.Color("#00aaff"),

	AgentArchitect:   lipgloss.Color("#00ffff"),
	AgentImplementer: lipgloss.Color("#00ff88"),
	AgentReviewer:    lipgloss.Color("#ffaa00"),
	AgentNavigator:   lipgloss.Color("#ff00ff"),
}

// Current is the active theme.
var Current = Neon
