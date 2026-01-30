// Package widgets provides TUI components.
package widgets

import (
	"fmt"
	"strings"

	"cooperations/internal/tui/styles"
	"github.com/charmbracelet/lipgloss"
)

// ProgressBar is a custom neon progress bar widget.
type ProgressBar struct {
	Percent float64
	Width   int
	Label   string
	Style   ProgressStyle
}

// ProgressStyle defines the visual style of a progress bar.
type ProgressStyle struct {
	Filled      string
	Empty       string
	LeftCap     string
	RightCap    string
	FilledColor lipgloss.Color
	EmptyColor  lipgloss.Color
	LabelColor  lipgloss.Color
}

// DefaultProgressStyle is the default neon style.
var DefaultProgressStyle = ProgressStyle{
	Filled:      "█",
	Empty:       "░",
	LeftCap:     "[",
	RightCap:    "]",
	FilledColor: styles.Current.Primary,
	EmptyColor:  styles.Current.Muted,
	LabelColor:  styles.Current.Foreground,
}

// NewProgressBar creates a new progress bar with default style.
func NewProgressBar(width int) ProgressBar {
	return ProgressBar{
		Width: width,
		Style: DefaultProgressStyle,
	}
}

// SetPercent updates the progress percentage.
func (p *ProgressBar) SetPercent(percent float64) {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	p.Percent = percent
}

// View renders the progress bar.
func (p ProgressBar) View() string {
	if p.Width <= 0 {
		return ""
	}
	filledStyle := lipgloss.NewStyle().Foreground(p.Style.FilledColor)
	emptyStyle := lipgloss.NewStyle().Foreground(p.Style.EmptyColor)
	labelStyle := lipgloss.NewStyle().Foreground(p.Style.LabelColor)

	filled := int(p.Percent / 100 * float64(p.Width))
	if filled > p.Width {
		filled = p.Width
	}

	bar := p.Style.LeftCap
	bar += filledStyle.Render(strings.Repeat(p.Style.Filled, filled))
	bar += emptyStyle.Render(strings.Repeat(p.Style.Empty, p.Width-filled))
	bar += p.Style.RightCap

	percentStr := fmt.Sprintf("%3.0f%%", p.Percent)

	if p.Label != "" {
		return fmt.Sprintf("%s %s %s", bar, labelStyle.Render(percentStr), labelStyle.Render(p.Label))
	}
	return fmt.Sprintf("%s %s", bar, labelStyle.Render(percentStr))
}

// MiniProgressBar is a compact progress bar for inline use.
type MiniProgressBar struct {
	Percent float64
	Width   int
	Color   lipgloss.Color
}

// NewMiniProgressBar creates a compact progress bar.
func NewMiniProgressBar(width int, color lipgloss.Color) MiniProgressBar {
	return MiniProgressBar{
		Width: width,
		Color: color,
	}
}

// View renders the mini progress bar.
func (p MiniProgressBar) View() string {
	if p.Width <= 0 {
		return ""
	}
	filled := int(p.Percent / 100 * float64(p.Width))
	if filled > p.Width {
		filled = p.Width
	}

	filledStyle := lipgloss.NewStyle().Foreground(p.Color)
	emptyStyle := lipgloss.NewStyle().Foreground(styles.Current.Muted)

	return filledStyle.Render(strings.Repeat("▓", filled)) +
		emptyStyle.Render(strings.Repeat("░", p.Width-filled))
}
