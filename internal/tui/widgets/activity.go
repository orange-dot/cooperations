// Package widgets provides TUI components.
package widgets

import (
	"fmt"
	"strings"
	"time"

	"cooperations/internal/tui/styles"
	"github.com/charmbracelet/lipgloss"
)

// LogLevel represents log severity.
type LogLevel int

const (
	LogDebug LogLevel = iota
	LogInfo
	LogWarn
	LogError
)

// LogEntry represents a single log entry.
type LogEntry struct {
	Timestamp time.Time
	Level     LogLevel
	Agent     string
	Message   string
}

// ActivityLog is a scrollable activity log widget.
type ActivityLog struct {
	Entries    []LogEntry
	MaxEntries int
	Width      int
	Height     int
	ScrollPos  int
	ShowTime   bool
	ShowAgent  bool
}

// NewActivityLog creates a new activity log widget.
func NewActivityLog(width, height int) ActivityLog {
	return ActivityLog{
		Entries:    make([]LogEntry, 0, 100),
		MaxEntries: 500,
		Width:      width,
		Height:     height,
		ShowTime:   true,
		ShowAgent:  true,
	}
}

// Add adds a new log entry.
func (a *ActivityLog) Add(level LogLevel, agent, message string) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Agent:     agent,
		Message:   message,
	}

	a.Entries = append(a.Entries, entry)

	// Trim old entries if over limit
	if len(a.Entries) > a.MaxEntries {
		a.Entries = a.Entries[len(a.Entries)-a.MaxEntries:]
	}

	// Auto-scroll to bottom
	if len(a.Entries) > a.Height {
		a.ScrollPos = len(a.Entries) - a.Height
	}
}

// Clear resets the activity log.
func (a *ActivityLog) Clear() {
	a.Entries = nil
	a.ScrollPos = 0
}

// AddInfo adds an info level log.
func (a *ActivityLog) AddInfo(agent, message string) {
	a.Add(LogInfo, agent, message)
}

// AddWarn adds a warning level log.
func (a *ActivityLog) AddWarn(agent, message string) {
	a.Add(LogWarn, agent, message)
}

// AddError adds an error level log.
func (a *ActivityLog) AddError(agent, message string) {
	a.Add(LogError, agent, message)
}

// AddDebug adds a debug level log.
func (a *ActivityLog) AddDebug(agent, message string) {
	a.Add(LogDebug, agent, message)
}

// ScrollUp scrolls the log up.
func (a *ActivityLog) ScrollUp(lines int) {
	a.ScrollPos -= lines
	if a.ScrollPos < 0 {
		a.ScrollPos = 0
	}
}

// ScrollDown scrolls the log down.
func (a *ActivityLog) ScrollDown(lines int) {
	maxScroll := len(a.Entries) - a.Height
	if maxScroll < 0 {
		maxScroll = 0
	}
	a.ScrollPos += lines
	if a.ScrollPos > maxScroll {
		a.ScrollPos = maxScroll
	}
}

// ScrollToTop jumps to the top of the log.
func (a *ActivityLog) ScrollToTop() {
	a.ScrollPos = 0
}

// ScrollToBottom jumps to the bottom of the log.
func (a *ActivityLog) ScrollToBottom() {
	maxScroll := len(a.Entries) - a.Height
	if maxScroll < 0 {
		maxScroll = 0
	}
	a.ScrollPos = maxScroll
}

// levelStyle returns the style for a log level.
func levelStyle(level LogLevel) lipgloss.Style {
	switch level {
	case LogDebug:
		return styles.LogDebug
	case LogInfo:
		return styles.LogInfo
	case LogWarn:
		return styles.LogWarn
	case LogError:
		return styles.LogError
	default:
		return styles.LogInfo
	}
}

// levelPrefix returns the prefix for a log level.
func levelPrefix(level LogLevel) string {
	switch level {
	case LogDebug:
		return "DBG"
	case LogInfo:
		return "INF"
	case LogWarn:
		return "WRN"
	case LogError:
		return "ERR"
	default:
		return "???"
	}
}

// View renders the activity log.
func (a ActivityLog) View() string {
	if a.Width <= 0 || a.Height <= 0 {
		return ""
	}
	if len(a.Entries) == 0 {
		return styles.MutedStyle.Render("No activity yet...")
	}

	var lines []string

	// Calculate visible range
	start := a.ScrollPos
	end := start + a.Height
	if end > len(a.Entries) {
		end = len(a.Entries)
	}

	for i := start; i < end; i++ {
		entry := a.Entries[i]
		var parts []string

		// Timestamp
		if a.ShowTime {
			timeStr := entry.Timestamp.Format("15:04:05")
			parts = append(parts, styles.MutedStyle.Render(timeStr))
		}

		// Level
		lvlStyle := levelStyle(entry.Level)
		parts = append(parts, lvlStyle.Render(levelPrefix(entry.Level)))

		// Agent
		if a.ShowAgent && entry.Agent != "" {
			agentStyle := styles.AgentStyle(entry.Agent)
			parts = append(parts, agentStyle.Render(fmt.Sprintf("[%s]", entry.Agent)))
		}

		// Message
		msgStyle := lipgloss.NewStyle().Foreground(styles.Current.Foreground)
		parts = append(parts, msgStyle.Render(entry.Message))

		line := strings.Join(parts, " ")

		// Truncate if too wide
		if lipgloss.Width(line) > a.Width {
			if a.Width <= 3 {
				line = line[:maxInt(a.Width, 0)]
			} else {
				line = line[:a.Width-3] + "..."
			}
		}

		lines = append(lines, line)
	}

	// Show scroll indicator if needed
	if len(a.Entries) > a.Height {
		scrollInfo := fmt.Sprintf(" [%d/%d]", a.ScrollPos+1, len(a.Entries)-a.Height+1)
		lines = append(lines, styles.MutedStyle.Render(scrollInfo))
	}

	return strings.Join(lines, "\n")
}
