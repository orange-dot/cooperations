// Package widgets provides TUI components.
package widgets

import (
	"time"

	"cooperations/internal/tui/styles"
	"github.com/charmbracelet/lipgloss"
)

// SpinnerFrames defines different spinner animations.
var SpinnerFrames = map[string][]string{
	"dots": {"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	"line": {"-", "\\", "|", "/"},
	"arc":  {"◜", "◠", "◝", "◞", "◡", "◟"},
	"neon": {"◐", "◓", "◑", "◒"},
	"pulse": {"█", "▓", "▒", "░", "▒", "▓"},
	"wave": {"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█", "▇", "▆", "▅", "▄", "▃", "▂"},
}

// Spinner is an animated spinner widget.
type Spinner struct {
	Frames    []string
	Frame     int
	Color     lipgloss.Color
	Label     string
	LastTick  time.Time
	Interval  time.Duration
}

// NewSpinner creates a new spinner with default neon style.
func NewSpinner() Spinner {
	return Spinner{
		Frames:   SpinnerFrames["dots"],
		Color:    styles.Current.Primary,
		Interval: 80 * time.Millisecond,
	}
}

// NewSpinnerWithStyle creates a spinner with specific frames and color.
func NewSpinnerWithStyle(frames []string, color lipgloss.Color) Spinner {
	return Spinner{
		Frames:   frames,
		Color:    color,
		Interval: 80 * time.Millisecond,
	}
}

// Tick advances the spinner frame.
func (s *Spinner) Tick() {
	s.Frame = (s.Frame + 1) % len(s.Frames)
	s.LastTick = time.Now()
}

// ShouldTick returns true if enough time has passed for next frame.
func (s *Spinner) ShouldTick() bool {
	return time.Since(s.LastTick) >= s.Interval
}

// View renders the spinner.
func (s Spinner) View() string {
	style := lipgloss.NewStyle().Foreground(s.Color)
	spinner := style.Render(s.Frames[s.Frame])

	if s.Label != "" {
		labelStyle := lipgloss.NewStyle().Foreground(styles.Current.Foreground)
		return spinner + " " + labelStyle.Render(s.Label)
	}
	return spinner
}

// AgentSpinner is a spinner with agent role styling.
type AgentSpinner struct {
	Spinner
	Role string
}

// NewAgentSpinner creates a spinner styled for a specific agent role.
func NewAgentSpinner(role string) AgentSpinner {
	color := styles.Current.Primary
	switch role {
	case "architect":
		color = styles.Current.AgentArchitect
	case "implementer":
		color = styles.Current.AgentImplementer
	case "reviewer":
		color = styles.Current.AgentReviewer
	case "navigator":
		color = styles.Current.AgentNavigator
	}

	return AgentSpinner{
		Spinner: Spinner{
			Frames:   SpinnerFrames["pulse"],
			Color:    color,
			Interval: 100 * time.Millisecond,
		},
		Role: role,
	}
}

// View renders the agent spinner with role indicator.
func (s AgentSpinner) View() string {
	style := lipgloss.NewStyle().Foreground(s.Color).Bold(true)
	spinner := style.Render(s.Frames[s.Frame])

	roleStyle := lipgloss.NewStyle().Foreground(s.Color)
	roleText := roleStyle.Render("[" + s.Role + "]")

	if s.Label != "" {
		labelStyle := lipgloss.NewStyle().Foreground(styles.Current.Foreground)
		return spinner + " " + roleText + " " + labelStyle.Render(s.Label)
	}
	return spinner + " " + roleText
}
