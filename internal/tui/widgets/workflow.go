// Package widgets provides TUI components.
package widgets

import (
	"fmt"
	"strings"

	"cooperations/internal/tui/styles"
	"github.com/charmbracelet/lipgloss"
)

// StepStatus represents the status of a workflow step.
type StepStatus int

const (
	StepPending StepStatus = iota
	StepRunning
	StepComplete
	StepFailed
	StepSkipped
)

// WorkflowStep represents a single step in the workflow.
type WorkflowStep struct {
	Name        string
	Description string
	Status      StepStatus
	Agent       string
	Duration    string
	Error       string
}

// WorkflowSteps displays the workflow progress as steps.
type WorkflowSteps struct {
	Steps      []WorkflowStep
	Width      int
	Vertical   bool
	ShowDetail bool
}

// NewWorkflowSteps creates a new workflow steps widget.
func NewWorkflowSteps(width int) WorkflowSteps {
	return WorkflowSteps{
		Width:      width,
		Vertical:   true,
		ShowDetail: true,
	}
}

// Clear resets all workflow steps.
func (w *WorkflowSteps) Clear() {
	w.Steps = nil
}

// AddStep adds a new step.
func (w *WorkflowSteps) AddStep(name, description, agent string) {
	w.Steps = append(w.Steps, WorkflowStep{
		Name:        name,
		Description: description,
		Agent:       agent,
		Status:      StepPending,
	})
}

// SetStatus sets the status of a step.
func (w *WorkflowSteps) SetStatus(index int, status StepStatus) {
	if index >= 0 && index < len(w.Steps) {
		w.Steps[index].Status = status
	}
}

// SetDuration sets the duration of a step.
func (w *WorkflowSteps) SetDuration(index int, duration string) {
	if index >= 0 && index < len(w.Steps) {
		w.Steps[index].Duration = duration
	}
}

// SetError sets the error of a step.
func (w *WorkflowSteps) SetError(index int, err string) {
	if index >= 0 && index < len(w.Steps) {
		w.Steps[index].Error = err
		w.Steps[index].Status = StepFailed
	}
}

// CurrentStep returns the index of the currently running step.
func (w *WorkflowSteps) CurrentStep() int {
	for i, step := range w.Steps {
		if step.Status == StepRunning {
			return i
		}
	}
	return -1
}

// stepIcon returns the icon for a step status.
func stepIcon(status StepStatus) string {
	switch status {
	case StepPending:
		return "○"
	case StepRunning:
		return "◉"
	case StepComplete:
		return "✓"
	case StepFailed:
		return "✗"
	case StepSkipped:
		return "⊘"
	default:
		return "?"
	}
}

// stepStyle returns the style for a step status.
func stepStyle(status StepStatus) lipgloss.Style {
	switch status {
	case StepPending:
		return styles.MutedStyle
	case StepRunning:
		return styles.StatusRunning
	case StepComplete:
		return styles.StatusComplete
	case StepFailed:
		return styles.StatusError
	case StepSkipped:
		return styles.MutedStyle
	default:
		return styles.MutedStyle
	}
}

// View renders the workflow steps.
func (w WorkflowSteps) View() string {
	if len(w.Steps) == 0 {
		return styles.MutedStyle.Render("No workflow steps")
	}

	if w.Vertical {
		return w.viewVertical()
	}
	return w.viewHorizontal()
}

// viewVertical renders steps vertically.
func (w WorkflowSteps) viewVertical() string {
	var lines []string

	for i, step := range w.Steps {
		style := stepStyle(step.Status)
		icon := stepIcon(step.Status)

		// Main line: icon + name
		mainLine := style.Render(icon + " " + step.Name)

		// Add agent tag if present
		if step.Agent != "" {
			agentStyle := styles.AgentStyle(step.Agent)
			mainLine += " " + agentStyle.Render("["+step.Agent+"]")
		}

		// Add duration if present
		if step.Duration != "" {
			mainLine += " " + styles.MutedStyle.Render("("+step.Duration+")")
		}

		lines = append(lines, mainLine)

		// Add description for current step
		if w.ShowDetail && step.Status == StepRunning && step.Description != "" {
			descLine := "  " + styles.MutedStyle.Render(step.Description)
			lines = append(lines, descLine)
		}

		// Add error if present
		if step.Error != "" {
			errLine := "  " + styles.StatusError.Render("Error: "+step.Error)
			lines = append(lines, errLine)
		}

		// Add connector line (except for last)
		if i < len(w.Steps)-1 {
			connector := "│"
			if step.Status == StepComplete || step.Status == StepSkipped {
				connector = styles.MutedStyle.Render(connector)
			} else if step.Status == StepRunning {
				connector = styles.StatusRunning.Render(connector)
			} else {
				connector = styles.MutedStyle.Render(connector)
			}
			lines = append(lines, connector)
		}
	}

	return strings.Join(lines, "\n")
}

// viewHorizontal renders steps horizontally.
func (w WorkflowSteps) viewHorizontal() string {
	var parts []string

	for i, step := range w.Steps {
		style := stepStyle(step.Status)
		icon := stepIcon(step.Status)

		part := style.Render(icon + " " + step.Name)
		parts = append(parts, part)

		// Add connector (except for last)
		if i < len(w.Steps)-1 {
			connector := " → "
			if step.Status == StepComplete {
				connector = styles.StatusComplete.Render(connector)
			} else {
				connector = styles.MutedStyle.Render(connector)
			}
			parts = append(parts, connector)
		}
	}

	return strings.Join(parts, "")
}

// HandoffIndicator shows a handoff between agents.
type HandoffIndicator struct {
	FromAgent string
	ToAgent   string
	Message   string
	Width     int
}

// NewHandoffIndicator creates a new handoff indicator.
func NewHandoffIndicator(from, to, message string, width int) HandoffIndicator {
	return HandoffIndicator{
		FromAgent: from,
		ToAgent:   to,
		Message:   message,
		Width:     width,
	}
}

// View renders the handoff indicator.
func (h HandoffIndicator) View() string {
	fromStyle := styles.AgentStyle(h.FromAgent)
	toStyle := styles.AgentStyle(h.ToAgent)
	arrowStyle := lipgloss.NewStyle().Foreground(styles.Current.Secondary)

	line1 := fromStyle.Render("["+h.FromAgent+"]") +
		arrowStyle.Render(" ─── handoff ───► ") +
		toStyle.Render("["+h.ToAgent+"]")

	if h.Message != "" {
		msgStyle := lipgloss.NewStyle().
			Foreground(styles.Current.Foreground).
			Italic(true)
		line2 := "  " + msgStyle.Render(`"`+h.Message+`"`)
		return line1 + "\n" + line2
	}

	return line1
}

// PhaseIndicator shows the current workflow phase.
type PhaseIndicator struct {
	Phase       string
	Description string
	Progress    float64
}

// View renders the phase indicator.
func (p PhaseIndicator) View() string {
	phaseStyle := lipgloss.NewStyle().
		Foreground(styles.Current.Primary).
		Bold(true).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Current.Primary)

	descStyle := styles.MutedStyle

	result := phaseStyle.Render(p.Phase)
	if p.Description != "" {
		result += " " + descStyle.Render(p.Description)
	}

	if p.Progress > 0 {
		progressBar := NewProgressBar(20)
		progressBar.SetPercent(p.Progress)
		result += "\n" + progressBar.View()
	}

	return result
}

// StepSummary shows a compact summary of workflow progress.
type StepSummary struct {
	Total     int
	Completed int
	Failed    int
	Current   string
}

// View renders the step summary.
func (s StepSummary) View() string {
	completedStyle := styles.StatusComplete
	failedStyle := styles.StatusError
	totalStyle := styles.MutedStyle

	summary := fmt.Sprintf("%s/%s",
		completedStyle.Render(fmt.Sprintf("%d", s.Completed)),
		totalStyle.Render(fmt.Sprintf("%d", s.Total)),
	)

	if s.Failed > 0 {
		summary += " " + failedStyle.Render(fmt.Sprintf("(%d failed)", s.Failed))
	}

	if s.Current != "" {
		currentStyle := styles.StatusRunning
		summary += " " + currentStyle.Render("→ "+s.Current)
	}

	return summary
}
