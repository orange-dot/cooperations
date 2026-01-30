// Package widgets provides TUI components.
package widgets

import (
	"fmt"
	"strings"

	"cooperations/internal/tui/styles"
	"github.com/charmbracelet/lipgloss"
)

// CodeBlock displays code with line numbers and optional syntax highlighting.
type CodeBlock struct {
	Content    string
	Language   string
	Filename   string
	Width      int
	Height     int
	ScrollPos  int
	ShowLines  bool
	StartLine  int
	Highlights []int // Lines to highlight
}

// NewCodeBlock creates a new code block widget.
func NewCodeBlock(width, height int) CodeBlock {
	return CodeBlock{
		Width:     width,
		Height:    height,
		ShowLines: true,
		StartLine: 1,
	}
}

// SetContent sets the code content.
func (c *CodeBlock) SetContent(content, language, filename string) {
	c.Content = content
	c.Language = language
	c.Filename = filename
	c.ScrollPos = 0
}

// ScrollUp scrolls code up.
func (c *CodeBlock) ScrollUp(lines int) {
	c.ScrollPos -= lines
	if c.ScrollPos < 0 {
		c.ScrollPos = 0
	}
}

// ScrollDown scrolls code down.
func (c *CodeBlock) ScrollDown(lines int) {
	allLines := strings.Split(c.Content, "\n")
	maxScroll := len(allLines) - c.Height
	if maxScroll < 0 {
		maxScroll = 0
	}
	c.ScrollPos += lines
	if c.ScrollPos > maxScroll {
		c.ScrollPos = maxScroll
	}
}

// ScrollToLine scrolls so the given line index is visible near the top.
func (c *CodeBlock) ScrollToLine(line int) {
	allLines := strings.Split(c.Content, "\n")
	if len(allLines) == 0 {
		c.ScrollPos = 0
		return
	}
	if line < 0 {
		line = 0
	}
	if line > len(allLines)-1 {
		line = len(allLines) - 1
	}
	maxScroll := len(allLines) - c.Height
	if maxScroll < 0 {
		maxScroll = 0
	}
	c.ScrollPos = line
	if c.ScrollPos > maxScroll {
		c.ScrollPos = maxScroll
	}
}

// ScrollToTop jumps to the top of the code.
func (c *CodeBlock) ScrollToTop() {
	c.ScrollPos = 0
}

// ScrollToBottom jumps to the bottom of the code.
func (c *CodeBlock) ScrollToBottom() {
	allLines := strings.Split(c.Content, "\n")
	maxScroll := len(allLines) - c.Height
	if maxScroll < 0 {
		maxScroll = 0
	}
	c.ScrollPos = maxScroll
}

// AddHighlight adds a line to highlight.
func (c *CodeBlock) AddHighlight(line int) {
	c.Highlights = append(c.Highlights, line)
}

// ClearHighlights removes all highlights.
func (c *CodeBlock) ClearHighlights() {
	c.Highlights = nil
}

// isHighlighted checks if a line should be highlighted.
func (c *CodeBlock) isHighlighted(line int) bool {
	for _, h := range c.Highlights {
		if h == line {
			return true
		}
	}
	return false
}

// View renders the code block.
func (c CodeBlock) View() string {
	var result strings.Builder

	if c.Width <= 0 || c.Height <= 0 {
		return ""
	}

	// Header with filename
	if c.Filename != "" {
		headerStyle := lipgloss.NewStyle().
			Foreground(styles.Current.Primary).
			Bold(true)
		langStyle := lipgloss.NewStyle().
			Foreground(styles.Current.Muted)

		header := headerStyle.Render(c.Filename)
		if c.Language != "" {
			header += " " + langStyle.Render("["+c.Language+"]")
		}
		result.WriteString(header + "\n")
		result.WriteString(strings.Repeat("─", c.Width) + "\n")
	}

	if c.Content == "" {
		return result.String() + styles.MutedStyle.Render("No code to display")
	}

	lines := strings.Split(c.Content, "\n")

	// Calculate visible range
	start := c.ScrollPos
	end := start + c.Height - 2 // Account for header
	if end > len(lines) {
		end = len(lines)
	}

	// Calculate line number width
	maxLineNum := c.StartLine + len(lines) - 1
	lineNumWidth := len(fmt.Sprintf("%d", maxLineNum))

	lineNumStyle := lipgloss.NewStyle().Foreground(styles.Current.Muted)
	codeStyle := lipgloss.NewStyle().Foreground(styles.Current.Foreground)
	highlightStyle := lipgloss.NewStyle().
		Foreground(styles.Current.Foreground).
		Background(lipgloss.Color("#1a2332"))

	for i := start; i < end; i++ {
		lineNum := c.StartLine + i
		line := lines[i]

		// Truncate long lines
		availWidth := c.Width - lineNumWidth - 3 // " │ "
		if availWidth < 1 {
			availWidth = 1
		}
		if len(line) > availWidth {
			if availWidth <= 1 {
				line = "…"
			} else {
				line = line[:availWidth-1] + "…"
			}
		}

		// Apply styles
		var styledLine string
		if c.ShowLines {
			numStr := fmt.Sprintf("%*d", lineNumWidth, lineNum)
			styledLine = lineNumStyle.Render(numStr) + " │ "
		}

		if c.isHighlighted(lineNum) {
			styledLine += highlightStyle.Render(line)
		} else {
			styledLine += codeStyle.Render(line)
		}

		result.WriteString(styledLine)
		if i < end-1 {
			result.WriteString("\n")
		}
	}

	// Scroll indicator
	if len(lines) > c.Height-2 {
		scrollInfo := fmt.Sprintf("\n%s", styles.MutedStyle.Render(
			fmt.Sprintf("Lines %d-%d of %d", start+1, end, len(lines)),
		))
		result.WriteString(scrollInfo)
	}

	return result.String()
}

// DiffBlock displays a unified diff with colors.
type DiffBlock struct {
	Content   string
	Filename  string
	Width     int
	Height    int
	ScrollPos int
	HighlightLines []int
}

// NewDiffBlock creates a new diff block widget.
func NewDiffBlock(width, height int) DiffBlock {
	return DiffBlock{
		Width:  width,
		Height: height,
	}
}

// SetContent sets the diff content.
func (d *DiffBlock) SetContent(content, filename string) {
	d.Content = content
	d.Filename = filename
	d.ScrollPos = 0
}

// ScrollUp scrolls diff up.
func (d *DiffBlock) ScrollUp(lines int) {
	d.ScrollPos -= lines
	if d.ScrollPos < 0 {
		d.ScrollPos = 0
	}
}

// ScrollDown scrolls diff down.
func (d *DiffBlock) ScrollDown(lines int) {
	allLines := strings.Split(d.Content, "\n")
	maxScroll := len(allLines) - d.Height
	if maxScroll < 0 {
		maxScroll = 0
	}
	d.ScrollPos += lines
	if d.ScrollPos > maxScroll {
		d.ScrollPos = maxScroll
	}
}

// ScrollToLine scrolls so the given line index is visible near the top.
func (d *DiffBlock) ScrollToLine(line int) {
	allLines := strings.Split(d.Content, "\n")
	if len(allLines) == 0 {
		d.ScrollPos = 0
		return
	}
	if line < 0 {
		line = 0
	}
	if line > len(allLines)-1 {
		line = len(allLines) - 1
	}
	maxScroll := len(allLines) - d.Height
	if maxScroll < 0 {
		maxScroll = 0
	}
	d.ScrollPos = line
	if d.ScrollPos > maxScroll {
		d.ScrollPos = maxScroll
	}
}

// SetHighlights sets highlighted line indices.
func (d *DiffBlock) SetHighlights(lines []int) {
	d.HighlightLines = append([]int(nil), lines...)
}

// ClearHighlights clears highlighted lines.
func (d *DiffBlock) ClearHighlights() {
	d.HighlightLines = nil
}

// ScrollToTop jumps to the top of the diff.
func (d *DiffBlock) ScrollToTop() {
	d.ScrollPos = 0
}

// ScrollToBottom jumps to the bottom of the diff.
func (d *DiffBlock) ScrollToBottom() {
	allLines := strings.Split(d.Content, "\n")
	maxScroll := len(allLines) - d.Height
	if maxScroll < 0 {
		maxScroll = 0
	}
	d.ScrollPos = maxScroll
}

// View renders the diff block.
func (d DiffBlock) View() string {
	var result strings.Builder

	if d.Width <= 0 || d.Height <= 0 {
		return ""
	}

	// Header
	if d.Filename != "" {
		headerStyle := lipgloss.NewStyle().
			Foreground(styles.Current.Secondary).
			Bold(true)
		result.WriteString(headerStyle.Render("Diff: "+d.Filename) + "\n")
		result.WriteString(strings.Repeat("─", d.Width) + "\n")
	}

	if d.Content == "" {
		return result.String() + styles.MutedStyle.Render("No changes")
	}

	lines := strings.Split(d.Content, "\n")

	// Calculate visible range
	start := d.ScrollPos
	end := start + d.Height - 2
	if end > len(lines) {
		end = len(lines)
	}

	highlight := make(map[int]struct{}, len(d.HighlightLines))
	for _, line := range d.HighlightLines {
		highlight[line] = struct{}{}
	}

	for i := start; i < end; i++ {
		line := lines[i]

		// Truncate
		if d.Width <= 1 {
			line = ""
		} else if len(line) > d.Width {
			line = line[:d.Width-1] + "…"
		}

		// Color based on prefix
		var styledLine string
		switch {
		case strings.HasPrefix(line, "+"):
			styledLine = styles.DiffAdd.Render(line)
		case strings.HasPrefix(line, "-"):
			styledLine = styles.DiffRemove.Render(line)
		case strings.HasPrefix(line, "@@"):
			styledLine = styles.SecondaryStyle.Render(line)
		default:
			styledLine = styles.DiffContext.Render(line)
		}

		if _, ok := highlight[i]; ok {
			highlightStyle := lipgloss.NewStyle().
				Foreground(styles.Current.Foreground).
				Background(styles.Current.Accent)
			result.WriteString(highlightStyle.Render(line))
		} else {
			result.WriteString(styledLine)
		}
		if i < end-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}
