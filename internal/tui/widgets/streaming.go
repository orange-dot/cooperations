// Package widgets provides TUI components.
package widgets

import (
	"strings"

	"cooperations/internal/tui/styles"
	"github.com/charmbracelet/lipgloss"
)

// StreamingText displays real-time streaming text with cursor.
type StreamingText struct {
	Content     string
	Width       int
	Height      int
	ScrollPos   int
	ShowCursor  bool
	CursorChar  string
	AgentRole   string
	IsStreaming bool
	HighlightLines []int
}

// NewStreamingText creates a new streaming text widget.
func NewStreamingText(width, height int) StreamingText {
	return StreamingText{
		Width:      width,
		Height:     height,
		ShowCursor: true,
		CursorChar: "▌",
	}
}

// Append adds text to the stream.
func (s *StreamingText) Append(text string) {
	s.Content += text
	s.IsStreaming = true

	// Auto-scroll to bottom
	lines := strings.Split(s.Content, "\n")
	if len(lines) > s.Height {
		s.ScrollPos = len(lines) - s.Height
	}
}

// SetContent replaces all content.
func (s *StreamingText) SetContent(content string) {
	s.Content = content
	lines := strings.Split(s.Content, "\n")
	if len(lines) > s.Height {
		s.ScrollPos = len(lines) - s.Height
	}
}

// Clear resets the content.
func (s *StreamingText) Clear() {
	s.Content = ""
	s.ScrollPos = 0
	s.IsStreaming = false
}

// EndStream marks streaming as complete.
func (s *StreamingText) EndStream() {
	s.IsStreaming = false
}

// ScrollUp scrolls content up.
func (s *StreamingText) ScrollUp(lines int) {
	s.ScrollPos -= lines
	if s.ScrollPos < 0 {
		s.ScrollPos = 0
	}
}

// ScrollDown scrolls content down.
func (s *StreamingText) ScrollDown(lines int) {
	allLines := strings.Split(s.Content, "\n")
	maxScroll := len(allLines) - s.Height
	if maxScroll < 0 {
		maxScroll = 0
	}
	s.ScrollPos += lines
	if s.ScrollPos > maxScroll {
		s.ScrollPos = maxScroll
	}
}

// ScrollToLine scrolls so the given line index is visible near the top.
func (s *StreamingText) ScrollToLine(line int) {
	allLines := strings.Split(s.Content, "\n")
	if len(allLines) == 0 {
		s.ScrollPos = 0
		return
	}
	if line < 0 {
		line = 0
	}
	if line > len(allLines)-1 {
		line = len(allLines) - 1
	}
	maxScroll := len(allLines) - s.Height
	if maxScroll < 0 {
		maxScroll = 0
	}
	s.ScrollPos = line
	if s.ScrollPos > maxScroll {
		s.ScrollPos = maxScroll
	}
}

// SetHighlights sets highlighted line indices.
func (s *StreamingText) SetHighlights(lines []int) {
	s.HighlightLines = append([]int(nil), lines...)
}

// ClearHighlights clears highlighted lines.
func (s *StreamingText) ClearHighlights() {
	s.HighlightLines = nil
}

// ScrollToTop jumps to the top of the content.
func (s *StreamingText) ScrollToTop() {
	s.ScrollPos = 0
}

// ScrollToBottom jumps to the bottom of the content.
func (s *StreamingText) ScrollToBottom() {
	allLines := strings.Split(s.Content, "\n")
	maxScroll := len(allLines) - s.Height
	if maxScroll < 0 {
		maxScroll = 0
	}
	s.ScrollPos = maxScroll
}

// View renders the streaming text.
func (s StreamingText) View() string {
	if s.Width <= 0 || s.Height <= 0 {
		return ""
	}
	if s.Content == "" {
		if s.IsStreaming {
			cursorStyle := lipgloss.NewStyle().
				Foreground(styles.Current.Primary).
				Blink(true)
			return cursorStyle.Render(s.CursorChar)
		}
		return styles.MutedStyle.Render("Waiting for response...")
	}

	lines := strings.Split(s.Content, "\n")

	// Apply scroll
	start := s.ScrollPos
	end := start + s.Height
	if end > len(lines) {
		end = len(lines)
	}
	if start > len(lines) {
		start = len(lines)
	}

	visible := lines[start:end]

	// Style based on agent role
	var textStyle lipgloss.Style
	if s.AgentRole != "" {
		textStyle = styles.AgentStyle(s.AgentRole)
	} else {
		textStyle = lipgloss.NewStyle().Foreground(styles.Current.Foreground)
	}

	highlight := make(map[int]struct{}, len(s.HighlightLines))
	for _, line := range s.HighlightLines {
		highlight[line] = struct{}{}
	}

	// Build output
	var result strings.Builder
	for i, line := range visible {
		absolute := start + i
		// Truncate long lines
		if s.Width <= 1 {
			line = ""
		} else if len(line) > s.Width {
			line = line[:s.Width-1] + "…"
		}
		if _, ok := highlight[absolute]; ok {
			highlightStyle := lipgloss.NewStyle().
				Foreground(styles.Current.Foreground).
				Background(styles.Current.Accent)
			result.WriteString(highlightStyle.Render(line))
		} else {
			result.WriteString(textStyle.Render(line))
		}
		if i < len(visible)-1 {
			result.WriteString("\n")
		}
	}

	// Add cursor if streaming
	if s.IsStreaming && s.ShowCursor {
		cursorStyle := lipgloss.NewStyle().
			Foreground(styles.Current.Primary).
			Blink(true)
		result.WriteString(cursorStyle.Render(s.CursorChar))
	}

	return result.String()
}

// ThinkingIndicator shows an animated thinking indicator.
type ThinkingIndicator struct {
	Dots      int
	MaxDots   int
	Label     string
	AgentRole string
}

// NewThinkingIndicator creates a new thinking indicator.
func NewThinkingIndicator(label string) ThinkingIndicator {
	return ThinkingIndicator{
		MaxDots: 3,
		Label:   label,
	}
}

// Tick advances the animation.
func (t *ThinkingIndicator) Tick() {
	t.Dots = (t.Dots + 1) % (t.MaxDots + 1)
}

// View renders the thinking indicator.
func (t ThinkingIndicator) View() string {
	var style lipgloss.Style
	if t.AgentRole != "" {
		style = styles.AgentStyle(t.AgentRole)
	} else {
		style = lipgloss.NewStyle().Foreground(styles.Current.Primary)
	}

	dots := strings.Repeat(".", t.Dots) + strings.Repeat(" ", t.MaxDots-t.Dots)
	return style.Render("💭 " + t.Label + dots)
}
