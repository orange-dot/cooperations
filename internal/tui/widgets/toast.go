// Package widgets provides TUI components.
package widgets

import (
	"strings"
	"time"

	"cooperations/internal/tui/styles"
	"github.com/charmbracelet/lipgloss"
)

// ToastLevel represents the severity level of a toast.
type ToastLevel int

const (
	ToastLevelInfo ToastLevel = iota
	ToastLevelSuccess
	ToastLevelWarning
	ToastLevelError
)

// Toast represents a notification toast.
type Toast struct {
	Message   string
	Level     ToastLevel
	Duration  time.Duration
	CreatedAt time.Time
	Width     int
}

// NewToast creates a new toast notification.
func NewToast(message string, level ToastLevel, duration time.Duration) Toast {
	return Toast{
		Message:   message,
		Level:     level,
		Duration:  duration,
		CreatedAt: time.Now(),
		Width:     40,
	}
}

// IsExpired returns true if the toast has expired.
func (t Toast) IsExpired() bool {
	return time.Since(t.CreatedAt) > t.Duration
}

// TimeRemaining returns the time until expiration.
func (t Toast) TimeRemaining() time.Duration {
	remaining := t.Duration - time.Since(t.CreatedAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// View renders the toast.
func (t Toast) View() string {
	var style lipgloss.Style
	var icon string

	switch t.Level {
	case ToastLevelSuccess:
		style = styles.ToastSuccess
		icon = "✓ "
	case ToastLevelWarning:
		style = styles.ToastWarning
		icon = "⚠ "
	case ToastLevelError:
		style = styles.ToastError
		icon = "✗ "
	default:
		style = styles.ToastInfo
		icon = "ℹ "
	}

	// Wrap message if too long
	msg := t.Message
	if len(msg) > t.Width-4 {
		msg = msg[:t.Width-7] + "..."
	}

	return style.Render(icon + msg)
}

// ToastStack manages multiple toasts.
type ToastStack struct {
	Toasts   []Toast
	MaxCount int
	Width    int
}

// NewToastStack creates a new toast stack.
func NewToastStack(maxCount, width int) ToastStack {
	return ToastStack{
		MaxCount: maxCount,
		Width:    width,
	}
}

// Push adds a new toast to the stack.
func (s *ToastStack) Push(message string, level ToastLevel, duration time.Duration) {
	toast := NewToast(message, level, duration)
	toast.Width = s.Width

	s.Toasts = append(s.Toasts, toast)

	// Remove oldest if over limit
	if len(s.Toasts) > s.MaxCount {
		s.Toasts = s.Toasts[len(s.Toasts)-s.MaxCount:]
	}
}

// PushInfo adds an info toast.
func (s *ToastStack) PushInfo(message string) {
	s.Push(message, ToastLevelInfo, 3*time.Second)
}

// PushSuccess adds a success toast.
func (s *ToastStack) PushSuccess(message string) {
	s.Push(message, ToastLevelSuccess, 3*time.Second)
}

// PushWarning adds a warning toast.
func (s *ToastStack) PushWarning(message string) {
	s.Push(message, ToastLevelWarning, 5*time.Second)
}

// PushError adds an error toast.
func (s *ToastStack) PushError(message string) {
	s.Push(message, ToastLevelError, 7*time.Second)
}

// Cleanup removes expired toasts.
func (s *ToastStack) Cleanup() {
	var active []Toast
	for _, t := range s.Toasts {
		if !t.IsExpired() {
			active = append(active, t)
		}
	}
	s.Toasts = active
}

// View renders the toast stack.
func (s ToastStack) View() string {
	if len(s.Toasts) == 0 {
		return ""
	}

	var lines []string
	for i := len(s.Toasts) - 1; i >= 0; i-- {
		toast := s.Toasts[i]
		if !toast.IsExpired() {
			lines = append(lines, toast.View())
		}
	}

	return strings.Join(lines, "\n")
}

// Count returns the number of active toasts.
func (s ToastStack) Count() int {
	count := 0
	for _, t := range s.Toasts {
		if !t.IsExpired() {
			count++
		}
	}
	return count
}

// Notification is a persistent notification with action.
type Notification struct {
	Title    string
	Message  string
	Level    ToastLevel
	Action   string
	ActionFn func()
	Width    int
}

// NewNotification creates a new notification.
func NewNotification(title, message string, level ToastLevel) Notification {
	return Notification{
		Title:   title,
		Message: message,
		Level:   level,
		Width:   50,
	}
}

// View renders the notification.
func (n Notification) View() string {
	var borderColor lipgloss.Color
	var icon string

	switch n.Level {
	case ToastLevelSuccess:
		borderColor = styles.Current.Success
		icon = "✓"
	case ToastLevelWarning:
		borderColor = styles.Current.Warning
		icon = "⚠"
	case ToastLevelError:
		borderColor = styles.Current.Error
		icon = "✗"
	default:
		borderColor = styles.Current.Info
		icon = "ℹ"
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(n.Width)

	titleStyle := lipgloss.NewStyle().
		Foreground(borderColor).
		Bold(true)

	msgStyle := lipgloss.NewStyle().
		Foreground(styles.Current.Foreground)

	content := titleStyle.Render(icon+" "+n.Title) + "\n" + msgStyle.Render(n.Message)

	if n.Action != "" {
		actionStyle := lipgloss.NewStyle().
			Foreground(styles.Current.Primary).
			Underline(true)
		content += "\n\n" + actionStyle.Render("["+n.Action+"]")
	}

	return style.Render(content)
}
