// Package widgets provides TUI components.
package widgets

import (
	"fmt"
	"strings"

	"cooperations/internal/tui/styles"
	"github.com/charmbracelet/lipgloss"
)

// AgentStatus represents the current status of an agent.
type AgentStatus int

const (
	AgentIdle AgentStatus = iota
	AgentThinking
	AgentWorking
	AgentWaiting
	AgentComplete
	AgentError
)

// AgentInfo holds information about an agent.
type AgentInfo struct {
	Role        string
	Status      AgentStatus
	CurrentTask string
	TokensUsed  int
	Duration    string
	LastMessage string
}

// AgentCard displays a single agent's status.
type AgentCard struct {
	Info     AgentInfo
	Width    int
	Expanded bool
	Spinner  *AgentSpinner
}

// NewAgentCard creates a new agent card.
func NewAgentCard(role string, width int) AgentCard {
	spinner := NewAgentSpinner(role)
	return AgentCard{
		Info: AgentInfo{
			Role:   role,
			Status: AgentIdle,
		},
		Width:   width,
		Spinner: &spinner,
	}
}

// SetStatus updates the agent status.
func (a *AgentCard) SetStatus(status AgentStatus, task string) {
	a.Info.Status = status
	a.Info.CurrentTask = task
}

// SetMetrics updates agent metrics.
func (a *AgentCard) SetMetrics(tokens int, duration string) {
	a.Info.TokensUsed = tokens
	a.Info.Duration = duration
}

// Tick advances the spinner animation.
func (a *AgentCard) Tick() {
	if a.Spinner != nil {
		a.Spinner.Tick()
	}
}

// agentIcon returns the icon for an agent role.
func agentIcon(role string) string {
	switch role {
	case "architect":
		return "🏗️"
	case "implementer":
		return "⚡"
	case "reviewer":
		return "🔍"
	case "navigator":
		return "🧭"
	default:
		return "🤖"
	}
}

// statusText returns the text for an agent status.
func statusText(status AgentStatus) string {
	switch status {
	case AgentIdle:
		return "Idle"
	case AgentThinking:
		return "Thinking..."
	case AgentWorking:
		return "Working"
	case AgentWaiting:
		return "Waiting"
	case AgentComplete:
		return "Complete"
	case AgentError:
		return "Error"
	default:
		return "Unknown"
	}
}

// View renders the agent card.
func (a AgentCard) View() string {
	roleStyle := styles.AgentStyle(a.Info.Role)

	// Border color based on status
	var borderColor lipgloss.Color
	switch a.Info.Status {
	case AgentWorking, AgentThinking:
		borderColor = roleStyle.GetForeground().(lipgloss.Color)
	case AgentComplete:
		borderColor = styles.Current.Success
	case AgentError:
		borderColor = styles.Current.Error
	default:
		borderColor = styles.Current.Border
	}

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(a.Width)

	var lines []string

	// Header: icon + role + status
	header := agentIcon(a.Info.Role) + " " + roleStyle.Render(a.Info.Role)

	// Add spinner or status
	if a.Info.Status == AgentWorking || a.Info.Status == AgentThinking {
		if a.Spinner != nil {
			header += " " + a.Spinner.View()
		}
	} else {
		statusStyle := styles.MutedStyle
		switch a.Info.Status {
		case AgentComplete:
			statusStyle = styles.StatusComplete
		case AgentError:
			statusStyle = styles.StatusError
		case AgentWaiting:
			statusStyle = styles.StatusWaiting
		}
		header += " " + statusStyle.Render(statusText(a.Info.Status))
	}

	lines = append(lines, header)

	// Current task
	if a.Info.CurrentTask != "" {
		taskStyle := lipgloss.NewStyle().Foreground(styles.Current.Foreground)
		task := a.Info.CurrentTask
		if len(task) > a.Width-4 {
			task = task[:a.Width-7] + "..."
		}
		lines = append(lines, taskStyle.Render(task))
	}

	// Expanded details
	if a.Expanded {
		lines = append(lines, "")
		detailStyle := styles.MutedStyle

		if a.Info.TokensUsed > 0 {
			lines = append(lines, detailStyle.Render(fmt.Sprintf("Tokens: %s", formatNumber(a.Info.TokensUsed))))
		}
		if a.Info.Duration != "" {
			lines = append(lines, detailStyle.Render(fmt.Sprintf("Duration: %s", a.Info.Duration)))
		}
		if a.Info.LastMessage != "" {
			msg := a.Info.LastMessage
			if len(msg) > a.Width-4 {
				msg = msg[:a.Width-7] + "..."
			}
			lines = append(lines, detailStyle.Render("Last: "+msg))
		}
	}

	return cardStyle.Render(strings.Join(lines, "\n"))
}

// AgentPanel displays all agents in a grid.
type AgentPanel struct {
	Agents   map[string]*AgentCard
	Order    []string
	Width    int
	Height   int
	Columns  int
	Expanded bool
}

// NewAgentPanel creates a new agent panel.
func NewAgentPanel(width, height, columns int) AgentPanel {
	return AgentPanel{
		Agents:  make(map[string]*AgentCard),
		Order:   []string{"architect", "implementer", "reviewer", "navigator"},
		Width:   width,
		Height:  height,
		Columns: columns,
	}
}

// InitAgents initializes the standard agents.
func (p *AgentPanel) InitAgents() {
	cardWidth := p.Width/p.Columns - 2
	for _, role := range p.Order {
		card := NewAgentCard(role, cardWidth)
		p.Agents[role] = &card
	}
}

// GetAgent returns an agent card by role.
func (p *AgentPanel) GetAgent(role string) *AgentCard {
	return p.Agents[role]
}

// SetStatus sets the status of an agent.
func (p *AgentPanel) SetStatus(role string, status AgentStatus, task string) {
	if agent, ok := p.Agents[role]; ok {
		agent.SetStatus(status, task)
	}
}

// TickAll advances all agent spinners.
func (p *AgentPanel) TickAll() {
	for _, agent := range p.Agents {
		agent.Tick()
	}
}

// View renders the agent panel.
func (p AgentPanel) View() string {
	if len(p.Agents) == 0 {
		return styles.MutedStyle.Render("No agents")
	}

	var rows []string

	for i := 0; i < len(p.Order); i += p.Columns {
		var cols []string
		for j := 0; j < p.Columns && i+j < len(p.Order); j++ {
			role := p.Order[i+j]
			if agent, ok := p.Agents[role]; ok {
				agent.Expanded = p.Expanded
				cols = append(cols, agent.View())
			}
		}
		if len(cols) > 0 {
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cols...))
		}
	}

	return strings.Join(rows, "\n")
}

// ActiveAgents returns a list of currently active agent roles.
func (p AgentPanel) ActiveAgents() []string {
	var active []string
	for role, agent := range p.Agents {
		if agent.Info.Status == AgentWorking || agent.Info.Status == AgentThinking {
			active = append(active, role)
		}
	}
	return active
}

// AgentBadge is a compact agent indicator.
type AgentBadge struct {
	Role   string
	Active bool
}

// View renders the agent badge.
func (b AgentBadge) View() string {
	style := styles.AgentStyle(b.Role)
	if !b.Active {
		style = style.Faint(true)
	}

	icon := agentIcon(b.Role)
	return style.Render(icon + " " + b.Role)
}

// AgentBadgeBar displays all agent badges in a row.
type AgentBadgeBar struct {
	Agents map[string]bool // role -> active
}

// NewAgentBadgeBar creates a new badge bar.
func NewAgentBadgeBar() AgentBadgeBar {
	return AgentBadgeBar{
		Agents: map[string]bool{
			"architect":   false,
			"implementer": false,
			"reviewer":    false,
			"navigator":   false,
		},
	}
}

// SetActive sets an agent as active or inactive.
func (b *AgentBadgeBar) SetActive(role string, active bool) {
	b.Agents[role] = active
}

// View renders the badge bar.
func (b AgentBadgeBar) View() string {
	order := []string{"architect", "implementer", "reviewer", "navigator"}
	var badges []string

	for _, role := range order {
		badge := AgentBadge{Role: role, Active: b.Agents[role]}
		badges = append(badges, badge.View())
	}

	return strings.Join(badges, " │ ")
}
