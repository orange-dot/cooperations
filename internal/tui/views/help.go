// Package views provides TUI view components.
package views

import (
	"strings"

	"cooperations/internal/tui/styles"
	"github.com/charmbracelet/lipgloss"
)

// KeyBinding represents a key binding with description.
type KeyBinding struct {
	Key         string
	Description string
	Category    string
}

// HelpView displays keyboard shortcuts and help information.
type HelpView struct {
	Width    int
	Height   int
	Bindings []KeyBinding
	Scroll   int
}

// NewHelpView creates a new help view with default keybindings.
func NewHelpView(width, height int) *HelpView {
	h := &HelpView{
		Width:  width,
		Height: height,
	}
	h.initBindings()
	return h
}

// initBindings sets up the default keybindings.
func (h *HelpView) initBindings() {
	h.Bindings = []KeyBinding{
		// Navigation
		{Key: "h/←", Description: "Focus left panel", Category: "Navigation"},
		{Key: "l/→", Description: "Focus right panel", Category: "Navigation"},
		{Key: "j/↓", Description: "Scroll down / Move down", Category: "Navigation"},
		{Key: "k/↑", Description: "Scroll up / Move up", Category: "Navigation"},
		{Key: "Tab", Description: "Cycle panel focus", Category: "Navigation"},
		{Key: "1-3", Description: "Jump to panel 1-3", Category: "Navigation"},

		// View modes
		{Key: "c", Description: "Toggle center panel mode", Category: "Views"},
		{Key: "r", Description: "Toggle right panel mode", Category: "Views"},
		{Key: "f", Description: "Toggle focus mode", Category: "Views"},
		{Key: "m", Description: "Toggle metrics view", Category: "Views"},
		{Key: "d", Description: "Toggle diff view", Category: "Views"},

		// Workflow control
		{Key: "Space", Description: "Pause/Resume workflow", Category: "Workflow"},
		{Key: "Enter", Description: "Confirm decision", Category: "Workflow"},
		{Key: "Esc", Description: "Cancel / Close dialog", Category: "Workflow"},
		{Key: "n", Description: "Next step (when paused)", Category: "Workflow"},
		{Key: "s", Description: "Skip current step", Category: "Workflow"},

		// File tree
		{Key: "o", Description: "Open/Expand file or folder", Category: "Files"},
		{Key: "e", Description: "Edit file (external)", Category: "Files"},
		{Key: "y", Description: "Copy file path", Category: "Files"},

		// Search
		{Key: "/", Description: "Search in current view", Category: "Search"},
		{Key: "n", Description: "Next search result", Category: "Search"},
		{Key: "N", Description: "Previous search result", Category: "Search"},
		{Key: "Esc", Description: "Clear search", Category: "Search"},

		// Session
		{Key: "Ctrl+s", Description: "Save session", Category: "Session"},
		{Key: "Ctrl+o", Description: "Open session", Category: "Session"},
		{Key: "Ctrl+r", Description: "Replay session", Category: "Session"},

		// General
		{Key: "?", Description: "Toggle help", Category: "General"},
		{Key: "q", Description: "Quit", Category: "General"},
		{Key: "Ctrl+c", Description: "Force quit", Category: "General"},
	}
}

// Resize adjusts the view dimensions.
func (h *HelpView) Resize(width, height int) {
	h.Width = width
	h.Height = height
}

// ScrollUp scrolls the help view up.
func (h *HelpView) ScrollUp(lines int) {
	h.Scroll -= lines
	if h.Scroll < 0 {
		h.Scroll = 0
	}
}

// ScrollDown scrolls the help view down.
func (h *HelpView) ScrollDown(lines int) {
	maxScroll := len(h.Bindings) - h.Height + 10
	if maxScroll < 0 {
		maxScroll = 0
	}
	h.Scroll += lines
	if h.Scroll > maxScroll {
		h.Scroll = maxScroll
	}
}

// View renders the help view.
func (h *HelpView) View() string {
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(styles.Current.Primary).
		Padding(1, 2).
		Width(h.Width - 4).
		Height(h.Height - 4)

	titleStyle := lipgloss.NewStyle().
		Foreground(styles.Current.Primary).
		Bold(true).
		Underline(true)

	categoryStyle := lipgloss.NewStyle().
		Foreground(styles.Current.Secondary).
		Bold(true)

	keyStyle := styles.HelpKeyStyle
	descStyle := styles.HelpDescStyle

	var lines []string
	lines = append(lines, titleStyle.Render("⌨️  Keyboard Shortcuts"))
	lines = append(lines, "")

	// Group by category
	categories := []string{"Navigation", "Views", "Workflow", "Files", "Search", "Session", "General"}
	categoryBindings := make(map[string][]KeyBinding)

	for _, b := range h.Bindings {
		categoryBindings[b.Category] = append(categoryBindings[b.Category], b)
	}

	for _, cat := range categories {
		bindings := categoryBindings[cat]
		if len(bindings) == 0 {
			continue
		}

		lines = append(lines, categoryStyle.Render("─── "+cat+" ───"))

		for _, b := range bindings {
			// Pad key to consistent width
			key := b.Key
			padding := 12 - len(key)
			if padding > 0 {
				key += strings.Repeat(" ", padding)
			}

			line := keyStyle.Render(key) + " " + descStyle.Render(b.Description)
			lines = append(lines, line)
		}

		lines = append(lines, "")
	}

	// Apply scrolling
	start := h.Scroll
	end := start + h.Height - 8
	if end > len(lines) {
		end = len(lines)
	}
	if start > len(lines) {
		start = len(lines)
	}

	visibleLines := lines[start:end]

	// Footer
	footer := styles.MutedStyle.Render("\nPress ? or Esc to close")

	return containerStyle.Render(strings.Join(visibleLines, "\n") + footer)
}

// QuickHelp displays a compact help bar.
type QuickHelp struct {
	Bindings []KeyBinding
	Width    int
}

// NewQuickHelp creates a quick help bar with essential bindings.
func NewQuickHelp(width int) QuickHelp {
	return QuickHelp{
		Width: width,
		Bindings: []KeyBinding{
			{Key: "?", Description: "Help"},
			{Key: "Tab", Description: "Focus"},
			{Key: "c/r", Description: "Mode"},
			{Key: "Space", Description: "Pause"},
			{Key: "q", Description: "Quit"},
		},
	}
}

// View renders the quick help bar.
func (q QuickHelp) View() string {
	var parts []string

	for _, b := range q.Bindings {
		key := styles.HelpKeyStyle.Render(b.Key)
		desc := styles.HelpDescStyle.Render(b.Description)
		parts = append(parts, key+" "+desc)
	}

	return strings.Join(parts, "  │  ")
}
