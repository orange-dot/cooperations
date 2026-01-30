// Package widgets provides TUI components.
package widgets

import (
	"fmt"
	"strings"

	"cooperations/internal/tui/styles"
	"github.com/charmbracelet/lipgloss"
)

// MetricCard displays a single metric value.
type MetricCard struct {
	Label    string
	Value    string
	Unit     string
	Color    lipgloss.Color
	Icon     string
	Trend    int // -1 down, 0 stable, 1 up
	SubValue string
}

// NewMetricCard creates a new metric card.
func NewMetricCard(label, value, unit string) MetricCard {
	return MetricCard{
		Label: label,
		Value: value,
		Unit:  unit,
		Color: styles.Current.Primary,
	}
}

// View renders the metric card.
func (m MetricCard) View() string {
	labelStyle := styles.MutedStyle
	valueStyle := lipgloss.NewStyle().
		Foreground(m.Color).
		Bold(true)
	unitStyle := styles.MutedStyle

	var icon string
	if m.Icon != "" {
		icon = m.Icon + " "
	}

	var trend string
	switch m.Trend {
	case 1:
		trend = styles.StatusComplete.Render(" ↑")
	case -1:
		trend = styles.StatusError.Render(" ↓")
	}

	line1 := labelStyle.Render(icon + m.Label)
	line2 := valueStyle.Render(m.Value) + unitStyle.Render(" "+m.Unit) + trend

	if m.SubValue != "" {
		line2 += " " + styles.MutedStyle.Render("("+m.SubValue+")")
	}

	return line1 + "\n" + line2
}

// MetricsPanel displays multiple metrics in a grid.
type MetricsPanel struct {
	Metrics []MetricCard
	Width   int
	Columns int
}

// NewMetricsPanel creates a new metrics panel.
func NewMetricsPanel(width, columns int) MetricsPanel {
	return MetricsPanel{
		Width:   width,
		Columns: columns,
	}
}

// AddMetric adds a metric to the panel.
func (p *MetricsPanel) AddMetric(m MetricCard) {
	p.Metrics = append(p.Metrics, m)
}

// Clear removes all metrics.
func (p *MetricsPanel) Clear() {
	p.Metrics = nil
}

// View renders the metrics panel.
func (p MetricsPanel) View() string {
	if len(p.Metrics) == 0 {
		return styles.MutedStyle.Render("No metrics")
	}

	colWidth := p.Width / p.Columns

	var rows []string
	for i := 0; i < len(p.Metrics); i += p.Columns {
		var cols []string
		for j := 0; j < p.Columns && i+j < len(p.Metrics); j++ {
			metric := p.Metrics[i+j]
			rendered := metric.View()

			// Pad to column width
			lines := strings.Split(rendered, "\n")
			var paddedLines []string
			for _, line := range lines {
				padding := colWidth - lipgloss.Width(line)
				if padding > 0 {
					line += strings.Repeat(" ", padding)
				}
				paddedLines = append(paddedLines, line)
			}
			cols = append(cols, strings.Join(paddedLines, "\n"))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cols...))
	}

	return strings.Join(rows, "\n\n")
}

// CostTracker tracks and displays cost information.
type CostTracker struct {
	TotalTokens    int
	InputTokens    int
	OutputTokens   int
	EstimatedCost  float64
	CostPerMToken  float64
	SessionBudget  float64
	Width          int
}

// NewCostTracker creates a new cost tracker.
func NewCostTracker(width int) CostTracker {
	return CostTracker{
		Width:         width,
		CostPerMToken: 15.0, // Default Claude pricing estimate
	}
}

// Update updates token counts and recalculates cost.
func (c *CostTracker) Update(input, output int) {
	c.InputTokens += input
	c.OutputTokens += output
	c.TotalTokens = c.InputTokens + c.OutputTokens
	c.EstimatedCost = float64(c.TotalTokens) / 1_000_000 * c.CostPerMToken
}

// View renders the cost tracker.
func (c CostTracker) View() string {
	tokenStyle := lipgloss.NewStyle().Foreground(styles.Current.Info)
	costStyle := lipgloss.NewStyle().Foreground(styles.Current.Warning).Bold(true)
	labelStyle := styles.MutedStyle

	lines := []string{
		labelStyle.Render("Tokens"),
		fmt.Sprintf("  %s %s",
			tokenStyle.Render(formatNumber(c.TotalTokens)),
			labelStyle.Render(fmt.Sprintf("(in:%d out:%d)", c.InputTokens, c.OutputTokens)),
		),
		"",
		labelStyle.Render("Estimated Cost"),
		fmt.Sprintf("  %s", costStyle.Render(fmt.Sprintf("$%.4f", c.EstimatedCost))),
	}

	if c.SessionBudget > 0 {
		remaining := c.SessionBudget - c.EstimatedCost
		var remainingStyle lipgloss.Style
		if remaining < c.SessionBudget*0.1 {
			remainingStyle = styles.StatusError
		} else if remaining < c.SessionBudget*0.3 {
			remainingStyle = styles.StatusWaiting
		} else {
			remainingStyle = styles.StatusComplete
		}

		pct := (c.EstimatedCost / c.SessionBudget) * 100
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("Budget"))
		lines = append(lines, fmt.Sprintf("  %s / $%.2f (%s)",
			remainingStyle.Render(fmt.Sprintf("$%.4f", remaining)),
			c.SessionBudget,
			labelStyle.Render(fmt.Sprintf("%.1f%%", pct)),
		))
	}

	return strings.Join(lines, "\n")
}

// formatNumber formats a number with thousands separators.
func formatNumber(n int) string {
	str := fmt.Sprintf("%d", n)
	if len(str) <= 3 {
		return str
	}

	var result strings.Builder
	for i, c := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(c)
	}
	return result.String()
}

// AgentMetrics displays metrics for a specific agent.
type AgentMetrics struct {
	Role       string
	Calls      int
	Tokens     int
	Duration   string
	Success    int
	Errors     int
}

// View renders agent metrics.
func (a AgentMetrics) View() string {
	roleStyle := styles.AgentStyle(a.Role)
	valueStyle := lipgloss.NewStyle().Foreground(styles.Current.Foreground)

	header := roleStyle.Render("[" + a.Role + "]")

	stats := []string{
		fmt.Sprintf("Calls: %d", a.Calls),
		fmt.Sprintf("Tokens: %s", formatNumber(a.Tokens)),
	}

	if a.Duration != "" {
		stats = append(stats, fmt.Sprintf("Time: %s", a.Duration))
	}

	if a.Success > 0 || a.Errors > 0 {
		successRate := float64(a.Success) / float64(a.Success+a.Errors) * 100
		stats = append(stats, fmt.Sprintf("Success: %.0f%%", successRate))
	}

	return header + "\n  " + valueStyle.Render(strings.Join(stats, " │ "))
}

// TimingBar displays a horizontal timing bar.
type TimingBar struct {
	Label      string
	Current    float64 // Current time in seconds
	Total      float64 // Total/expected time
	Width      int
	ShowValues bool
}

// View renders the timing bar.
func (t TimingBar) View() string {
	labelStyle := styles.MutedStyle
	barStyle := styles.PrimaryStyle

	pct := t.Current / t.Total
	if pct > 1 {
		pct = 1
	}

	filled := int(pct * float64(t.Width))
	empty := t.Width - filled

	bar := barStyle.Render(strings.Repeat("▓", filled)) +
		labelStyle.Render(strings.Repeat("░", empty))

	result := labelStyle.Render(t.Label) + "\n" + bar

	if t.ShowValues {
		values := fmt.Sprintf(" %.1fs / %.1fs", t.Current, t.Total)
		result += labelStyle.Render(values)
	}

	return result
}
