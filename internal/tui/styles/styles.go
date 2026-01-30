package styles

import "github.com/charmbracelet/lipgloss"

// Pre-built styles using the current theme.
var (
	// Base styles
	BaseStyle = lipgloss.NewStyle().
			Background(Current.Background).
			Foreground(Current.Foreground)

	// Title and headers
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Current.Primary).
			Padding(0, 1)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Current.Primary).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(Current.Border)

	SubHeaderStyle = lipgloss.NewStyle().
			Foreground(Current.Secondary).
			Bold(true)

	// Panel styles
	PanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Current.Border).
			Padding(0, 1)

	ActivePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Current.Primary).
				Padding(0, 1)

	// Status styles
	StatusRunning = lipgloss.NewStyle().
			Foreground(Current.Primary).
			Bold(true)

	StatusComplete = lipgloss.NewStyle().
			Foreground(Current.Success).
			Bold(true)

	StatusError = lipgloss.NewStyle().
			Foreground(Current.Error).
			Bold(true)

	StatusWaiting = lipgloss.NewStyle().
			Foreground(Current.Warning).
			Bold(true)

	// Text styles
	MutedStyle = lipgloss.NewStyle().
			Foreground(Current.Muted)

	AccentStyle = lipgloss.NewStyle().
			Foreground(Current.Accent)

	PrimaryStyle = lipgloss.NewStyle().
			Foreground(Current.Primary)

	SecondaryStyle = lipgloss.NewStyle().
			Foreground(Current.Secondary)

	// Log level styles
	LogInfo = lipgloss.NewStyle().
		Foreground(Current.Info)

	LogWarn = lipgloss.NewStyle().
		Foreground(Current.Warning)

	LogError = lipgloss.NewStyle().
		Foreground(Current.Error)

	LogDebug = lipgloss.NewStyle().
		Foreground(Current.Muted)

	// Diff styles
	DiffAdd = lipgloss.NewStyle().
		Foreground(Current.Success)

	DiffRemove = lipgloss.NewStyle().
		Foreground(Current.Error)

	DiffContext = lipgloss.NewStyle().
			Foreground(Current.Muted)

	// Agent styles
	AgentArchitectStyle = lipgloss.NewStyle().
				Foreground(Current.AgentArchitect).
				Bold(true)

	AgentImplementerStyle = lipgloss.NewStyle().
				Foreground(Current.AgentImplementer).
				Bold(true)

	AgentReviewerStyle = lipgloss.NewStyle().
				Foreground(Current.AgentReviewer).
				Bold(true)

	AgentNavigatorStyle = lipgloss.NewStyle().
				Foreground(Current.AgentNavigator).
				Bold(true)

	// Button styles
	ButtonStyle = lipgloss.NewStyle().
			Foreground(Current.Foreground).
			Background(Current.Border).
			Padding(0, 2).
			Margin(0, 1)

	ButtonActiveStyle = lipgloss.NewStyle().
				Foreground(Current.Background).
				Background(Current.Primary).
				Padding(0, 2).
				Margin(0, 1).
				Bold(true)

	// Help style
	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(Current.Primary).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(Current.Muted)

	// Toast styles
	ToastInfo = lipgloss.NewStyle().
		Foreground(Current.Foreground).
		Background(Current.Info).
		Padding(0, 1)

	ToastSuccess = lipgloss.NewStyle().
			Foreground(Current.Background).
			Background(Current.Success).
			Padding(0, 1)

	ToastWarning = lipgloss.NewStyle().
			Foreground(Current.Background).
			Background(Current.Warning).
			Padding(0, 1)

	ToastError = lipgloss.NewStyle().
		Foreground(Current.Foreground).
		Background(Current.Error).
		Padding(0, 1)
)

// AgentStyle returns the style for a given agent role.
func AgentStyle(role string) lipgloss.Style {
	switch role {
	case "architect":
		return AgentArchitectStyle
	case "implementer":
		return AgentImplementerStyle
	case "reviewer":
		return AgentReviewerStyle
	case "navigator":
		return AgentNavigatorStyle
	default:
		return PrimaryStyle
	}
}

// ToastStyle returns the style for a given toast level.
func ToastStyle(level string) lipgloss.Style {
	switch level {
	case "success":
		return ToastSuccess
	case "warning":
		return ToastWarning
	case "error":
		return ToastError
	default:
		return ToastInfo
	}
}
