// Package widgets provides TUI components.
package widgets

import (
	"strings"

	"cooperations/internal/tui/styles"
	"github.com/charmbracelet/lipgloss"
)

// DecisionOption represents a choice in a decision dialog.
type DecisionOption struct {
	Key         string
	Label       string
	Description string
	Danger      bool
}

// DecisionDialog displays a decision prompt with options.
type DecisionDialog struct {
	Title    string
	Message  string
	Options  []DecisionOption
	Selected int
	Width    int
	ShowHelp bool
}

// NewDecisionDialog creates a new decision dialog.
func NewDecisionDialog(title, message string, width int) DecisionDialog {
	return DecisionDialog{
		Title:    title,
		Message:  message,
		Width:    width,
		ShowHelp: true,
	}
}

// AddOption adds an option to the dialog.
func (d *DecisionDialog) AddOption(key, label, description string, danger bool) {
	d.Options = append(d.Options, DecisionOption{
		Key:         key,
		Label:       label,
		Description: description,
		Danger:      danger,
	})
}

// MoveUp moves selection up.
func (d *DecisionDialog) MoveUp() {
	if d.Selected > 0 {
		d.Selected--
	}
}

// MoveDown moves selection down.
func (d *DecisionDialog) MoveDown() {
	if d.Selected < len(d.Options)-1 {
		d.Selected++
	}
}

// GetSelected returns the selected option key.
func (d *DecisionDialog) GetSelected() string {
	if d.Selected >= 0 && d.Selected < len(d.Options) {
		return d.Options[d.Selected].Key
	}
	return ""
}

// GetSelectedOption returns the currently selected option.
func (d *DecisionDialog) GetSelectedOption() *DecisionOption {
	if d.Selected >= 0 && d.Selected < len(d.Options) {
		return &d.Options[d.Selected]
	}
	return nil
}

// SelectByKey selects an option by its key and returns true if found.
func (d *DecisionDialog) SelectByKey(key string) bool {
	for i, opt := range d.Options {
		if opt.Key == key {
			d.Selected = i
			return true
		}
	}
	return false
}

// View renders the decision dialog.
func (d DecisionDialog) View() string {
	// Dialog container style
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(styles.Current.Secondary).
		Padding(1, 2).
		Width(d.Width)

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(styles.Current.Secondary).
		Bold(true).
		Underline(true)

	// Message
	msgStyle := lipgloss.NewStyle().
		Foreground(styles.Current.Foreground)

	var lines []string
	lines = append(lines, titleStyle.Render(d.Title))
	lines = append(lines, "")
	lines = append(lines, msgStyle.Render(d.Message))
	lines = append(lines, "")

	// Options
	for i, opt := range d.Options {
		var optStyle lipgloss.Style
		var prefix string

		if i == d.Selected {
			prefix = "▶ "
			if opt.Danger {
				optStyle = lipgloss.NewStyle().
					Foreground(styles.Current.Error).
					Bold(true).
					Reverse(true)
			} else {
				optStyle = lipgloss.NewStyle().
					Foreground(styles.Current.Primary).
					Bold(true).
					Reverse(true)
			}
		} else {
			prefix = "  "
			if opt.Danger {
				optStyle = lipgloss.NewStyle().
					Foreground(styles.Current.Error)
			} else {
				optStyle = lipgloss.NewStyle().
					Foreground(styles.Current.Foreground)
			}
		}

		keyStyle := lipgloss.NewStyle().
			Foreground(styles.Current.Primary).
			Bold(true)

		line := prefix + keyStyle.Render("["+opt.Key+"]") + " " + optStyle.Render(opt.Label)
		lines = append(lines, line)

		if opt.Description != "" && i == d.Selected {
			descStyle := lipgloss.NewStyle().
				Foreground(styles.Current.Muted).
				Italic(true)
			lines = append(lines, "     "+descStyle.Render(opt.Description))
		}
	}

	// Help
	if d.ShowHelp {
		lines = append(lines, "")
		helpStyle := styles.MutedStyle
		lines = append(lines, helpStyle.Render("↑/↓: navigate  Enter: select  Esc: cancel"))
	}

	return containerStyle.Render(strings.Join(lines, "\n"))
}

// ConfirmDialog is a simple yes/no confirmation dialog.
type ConfirmDialog struct {
	Title    string
	Message  string
	YesLabel string
	NoLabel  string
	Selected int // 0 = No, 1 = Yes
	Width    int
	Danger   bool
}

// NewConfirmDialog creates a new confirmation dialog.
func NewConfirmDialog(title, message string, width int) ConfirmDialog {
	return ConfirmDialog{
		Title:    title,
		Message:  message,
		YesLabel: "Yes",
		NoLabel:  "No",
		Width:    width,
	}
}

// Toggle switches between yes and no.
func (c *ConfirmDialog) Toggle() {
	c.Selected = 1 - c.Selected
}

// IsYes returns true if Yes is selected.
func (c *ConfirmDialog) IsYes() bool {
	return c.Selected == 1
}

// View renders the confirmation dialog.
func (c ConfirmDialog) View() string {
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Current.Warning).
		Padding(1, 2).
		Width(c.Width)

	titleStyle := lipgloss.NewStyle().
		Foreground(styles.Current.Warning).
		Bold(true)

	msgStyle := lipgloss.NewStyle().
		Foreground(styles.Current.Foreground)

	var lines []string
	lines = append(lines, titleStyle.Render("⚠ "+c.Title))
	lines = append(lines, "")
	lines = append(lines, msgStyle.Render(c.Message))
	lines = append(lines, "")

	// Buttons
	var yesStyle, noStyle lipgloss.Style

	if c.Selected == 1 {
		if c.Danger {
			yesStyle = lipgloss.NewStyle().
				Foreground(styles.Current.Background).
				Background(styles.Current.Error).
				Bold(true).
				Padding(0, 2)
		} else {
			yesStyle = styles.ButtonActiveStyle
		}
		noStyle = styles.ButtonStyle
	} else {
		yesStyle = styles.ButtonStyle
		noStyle = styles.ButtonActiveStyle
	}

	buttons := noStyle.Render(c.NoLabel) + "  " + yesStyle.Render(c.YesLabel)
	lines = append(lines, buttons)

	return containerStyle.Render(strings.Join(lines, "\n"))
}

// InputDialog is a text input dialog.
type InputDialog struct {
	Title       string
	Prompt      string
	Value       string
	Placeholder string
	Width       int
	MaxLength   int
	CursorPos   int
}

// NewInputDialog creates a new input dialog.
func NewInputDialog(title, prompt string, width int) InputDialog {
	return InputDialog{
		Title:     title,
		Prompt:    prompt,
		Width:     width,
		MaxLength: 200,
	}
}

// Insert adds a character at cursor position.
func (i *InputDialog) Insert(char rune) {
	if len(i.Value) >= i.MaxLength {
		return
	}
	i.Value = i.Value[:i.CursorPos] + string(char) + i.Value[i.CursorPos:]
	i.CursorPos++
}

// Backspace removes character before cursor.
func (i *InputDialog) Backspace() {
	if i.CursorPos > 0 {
		i.Value = i.Value[:i.CursorPos-1] + i.Value[i.CursorPos:]
		i.CursorPos--
	}
}

// Delete removes character at cursor.
func (i *InputDialog) Delete() {
	if i.CursorPos < len(i.Value) {
		i.Value = i.Value[:i.CursorPos] + i.Value[i.CursorPos+1:]
	}
}

// MoveLeft moves cursor left.
func (i *InputDialog) MoveLeft() {
	if i.CursorPos > 0 {
		i.CursorPos--
	}
}

// MoveRight moves cursor right.
func (i *InputDialog) MoveRight() {
	if i.CursorPos < len(i.Value) {
		i.CursorPos++
	}
}

// View renders the input dialog.
func (i InputDialog) View() string {
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Current.Primary).
		Padding(1, 2).
		Width(i.Width)

	titleStyle := lipgloss.NewStyle().
		Foreground(styles.Current.Primary).
		Bold(true)

	promptStyle := lipgloss.NewStyle().
		Foreground(styles.Current.Foreground)

	var lines []string
	lines = append(lines, titleStyle.Render(i.Title))
	lines = append(lines, "")
	lines = append(lines, promptStyle.Render(i.Prompt))
	lines = append(lines, "")

	// Input field
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(styles.Current.Border).
		Padding(0, 1).
		Width(i.Width - 6)

	value := i.Value
	if value == "" && i.Placeholder != "" {
		value = styles.MutedStyle.Render(i.Placeholder)
	} else {
		// Insert cursor
		if i.CursorPos <= len(value) {
			before := value[:i.CursorPos]
			after := value[i.CursorPos:]
			cursorStyle := lipgloss.NewStyle().Reverse(true)
			cursor := cursorStyle.Render(" ")
			if len(after) > 0 {
				cursor = cursorStyle.Render(string(after[0]))
				after = after[1:]
			}
			value = before + cursor + after
		}
	}

	lines = append(lines, inputStyle.Render(value))
	lines = append(lines, "")
	lines = append(lines, styles.MutedStyle.Render("Enter: submit  Esc: cancel"))

	return containerStyle.Render(strings.Join(lines, "\n"))
}
