// Package views provides TUI view components.
package views

import (
	"fmt"
	"strings"

	"cooperations/internal/tui/styles"
	"cooperations/internal/tui/widgets"
	"github.com/charmbracelet/lipgloss"
)

// FocusMode represents what content is focused.
type FocusMode int

const (
	FocusModeStreaming FocusMode = iota
	FocusModeCode
	FocusModeDiff
	FocusModeActivity
)

// FocusView displays a single panel in full-screen mode.
type FocusView struct {
	Width  int
	Height int
	Mode   FocusMode

	// Content widgets
	StreamingText *widgets.StreamingText
	CodeBlock     *widgets.CodeBlock
	DiffBlock     *widgets.DiffBlock
	ActivityLog   *widgets.ActivityLog

	// Agent info
	ActiveAgent string
	AgentSpinner *widgets.AgentSpinner

	// Progress
	ProgressBar *widgets.ProgressBar

	// Mini metrics
	TokenCount int
	Duration   string
}

// NewFocusView creates a new focus view.
func NewFocusView(width, height int) *FocusView {
	contentHeight := height - 4 // header + footer

	streamingText := widgets.NewStreamingText(width-4, contentHeight-2)
	codeBlock := widgets.NewCodeBlock(width-4, contentHeight-2)
	diffBlock := widgets.NewDiffBlock(width-4, contentHeight-2)
	activityLog := widgets.NewActivityLog(width-4, contentHeight-2)
	progressBar := widgets.NewProgressBar(width - 30)

	return &FocusView{
		Width:         width,
		Height:        height,
		StreamingText: &streamingText,
		CodeBlock:     &codeBlock,
		DiffBlock:     &diffBlock,
		ActivityLog:   &activityLog,
		ProgressBar:   &progressBar,
	}
}

// Resize adjusts the view dimensions.
func (f *FocusView) Resize(width, height int) {
	f.Width = width
	f.Height = height

	contentHeight := height - 4
	contentWidth := width - 4

	f.StreamingText.Width = contentWidth
	f.StreamingText.Height = contentHeight - 2
	f.CodeBlock.Width = contentWidth
	f.CodeBlock.Height = contentHeight - 2
	f.DiffBlock.Width = contentWidth
	f.DiffBlock.Height = contentHeight - 2
	f.ActivityLog.Width = contentWidth
	f.ActivityLog.Height = contentHeight - 2
	f.ProgressBar.Width = width - 30
}

// SetMode sets the focus mode.
func (f *FocusView) SetMode(mode FocusMode) {
	f.Mode = mode
}

// SetActiveAgent sets the currently active agent.
func (f *FocusView) SetActiveAgent(role string) {
	f.ActiveAgent = role
	if role != "" {
		spinner := widgets.NewAgentSpinner(role)
		f.AgentSpinner = &spinner
	} else {
		f.AgentSpinner = nil
	}
}

// Tick advances animations.
func (f *FocusView) Tick() {
	if f.AgentSpinner != nil {
		f.AgentSpinner.Tick()
	}
}

// View renders the focus view.
func (f *FocusView) View() string {
	var result strings.Builder

	// Header
	result.WriteString(f.renderHeader())
	result.WriteString("\n")

	// Main content
	content := f.renderContent()
	result.WriteString(content)
	result.WriteString("\n")

	// Footer
	result.WriteString(f.renderFooter())

	return result.String()
}

// renderHeader renders the focus view header.
func (f *FocusView) renderHeader() string {
	// Mode indicator
	var modeLabel string
	var modeColor lipgloss.Color

	switch f.Mode {
	case FocusModeStreaming:
		modeLabel = "STREAMING"
		modeColor = styles.Current.Primary
	case FocusModeCode:
		modeLabel = "CODE"
		modeColor = styles.Current.Accent
	case FocusModeDiff:
		modeLabel = "DIFF"
		modeColor = styles.Current.Secondary
	case FocusModeActivity:
		modeLabel = "ACTIVITY"
		modeColor = styles.Current.Info
	}

	modeStyle := lipgloss.NewStyle().
		Foreground(styles.Current.Background).
		Background(modeColor).
		Bold(true).
		Padding(0, 2)

	header := modeStyle.Render(modeLabel)

	// Agent spinner
	if f.AgentSpinner != nil {
		header += "  " + f.AgentSpinner.View()
	}

	// Metrics on the right
	metricsStyle := styles.MutedStyle
	metrics := ""
	if f.TokenCount > 0 {
		metrics += metricsStyle.Render("Tokens: ") +
			lipgloss.NewStyle().Foreground(styles.Current.Info).Render(formatFocusNumber(f.TokenCount))
	}
	if f.Duration != "" {
		if metrics != "" {
			metrics += "  "
		}
		metrics += metricsStyle.Render("Time: ") +
			lipgloss.NewStyle().Foreground(styles.Current.Info).Render(f.Duration)
	}

	padding := f.Width - lipgloss.Width(header) - lipgloss.Width(metrics) - 2
	if padding < 0 {
		padding = 0
	}

	return header + strings.Repeat(" ", padding) + metrics
}

// formatFocusNumber formats a number with commas.
func formatFocusNumber(n int) string {
	if n == 0 {
		return "0"
	}

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

// renderContent renders the main focused content.
func (f *FocusView) renderContent() string {
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Current.Primary).
		Padding(1, 2).
		Width(f.Width - 4).
		Height(f.Height - 6)

	var content string

	switch f.Mode {
	case FocusModeStreaming:
		content = f.StreamingText.View()
	case FocusModeCode:
		content = f.CodeBlock.View()
	case FocusModeDiff:
		content = f.DiffBlock.View()
	case FocusModeActivity:
		content = f.ActivityLog.View()
	}

	return panelStyle.Render(content)
}

// renderFooter renders the focus view footer.
func (f *FocusView) renderFooter() string {
	// Progress bar
	progress := f.ProgressBar.View()

	// Help hint
	helpHint := styles.MutedStyle.Render("Esc: exit focus  c: code  d: diff  j/k: scroll")

	padding := f.Width - lipgloss.Width(progress) - lipgloss.Width(helpHint) - 4
	if padding < 0 {
		padding = 0
	}

	return progress + strings.Repeat(" ", padding) + helpHint
}

// ZenView is a minimal distraction-free view.
type ZenView struct {
	Width  int
	Height int

	Content       string
	AgentRole     string
	ShowCursor    bool
	CursorVisible bool
}

// NewZenView creates a new zen view.
func NewZenView(width, height int) *ZenView {
	return &ZenView{
		Width:      width,
		Height:     height,
		ShowCursor: true,
	}
}

// Resize adjusts the view dimensions.
func (z *ZenView) Resize(width, height int) {
	z.Width = width
	z.Height = height
}

// ToggleCursor toggles cursor visibility (for blinking).
func (z *ZenView) ToggleCursor() {
	z.CursorVisible = !z.CursorVisible
}

// View renders the zen view.
func (z *ZenView) View() string {
	// Centered content with generous padding
	contentWidth := z.Width * 2 / 3
	if contentWidth > 100 {
		contentWidth = 100
	}

	var textStyle lipgloss.Style
	if z.AgentRole != "" {
		textStyle = styles.AgentStyle(z.AgentRole)
	} else {
		textStyle = lipgloss.NewStyle().Foreground(styles.Current.Foreground)
	}

	// Wrap and center content
	lines := strings.Split(z.Content, "\n")
	var wrappedLines []string

	for _, line := range lines {
		if len(line) > contentWidth {
			// Simple word wrap
			words := strings.Fields(line)
			currentLine := ""
			for _, word := range words {
				if len(currentLine)+len(word)+1 > contentWidth {
					wrappedLines = append(wrappedLines, currentLine)
					currentLine = word
				} else {
					if currentLine != "" {
						currentLine += " "
					}
					currentLine += word
				}
			}
			if currentLine != "" {
				wrappedLines = append(wrappedLines, currentLine)
			}
		} else {
			wrappedLines = append(wrappedLines, line)
		}
	}

	// Add cursor to last line
	if z.ShowCursor && z.CursorVisible && len(wrappedLines) > 0 {
		cursorStyle := lipgloss.NewStyle().Reverse(true)
		wrappedLines[len(wrappedLines)-1] += cursorStyle.Render(" ")
	}

	content := textStyle.Render(strings.Join(wrappedLines, "\n"))

	// Center horizontally
	contentStyle := lipgloss.NewStyle().
		Width(z.Width).
		Align(lipgloss.Center).
		Padding((z.Height-len(wrappedLines))/2, 0)

	return contentStyle.Render(content)
}
