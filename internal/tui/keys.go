package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all keyboard shortcuts for the TUI.
type KeyMap struct {
	// Navigation
	Left      key.Binding
	Right     key.Binding
	Up        key.Binding
	Down      key.Binding
	Tab       key.Binding
	ShiftTab  key.Binding
	Panel1    key.Binding
	Panel2    key.Binding
	Panel3    key.Binding

	// Scrolling
	PageUp    key.Binding
	PageDown  key.Binding
	HalfUp    key.Binding
	HalfDown  key.Binding
	Top       key.Binding
	Bottom    key.Binding

	// View modes
	ToggleCenter key.Binding
	ToggleRight  key.Binding
	FocusMode    key.Binding
	MetricsView  key.Binding
	DiffView     key.Binding
	ZenMode      key.Binding

	// Workflow control
	Pause     key.Binding
	Resume    key.Binding
	NextStep  key.Binding
	Skip      key.Binding
	Confirm   key.Binding
	Cancel    key.Binding

	// File operations
	Open      key.Binding
	Edit      key.Binding
	CopyPath  key.Binding
	Refresh   key.Binding

	// Search
	Search       key.Binding
	NextResult   key.Binding
	PrevResult   key.Binding
	ClearSearch  key.Binding

	// Session
	SaveSession  key.Binding
	OpenSession  key.Binding
	Replay       key.Binding

	// General
	Help      key.Binding
	Quit      key.Binding
	ForceQuit key.Binding
}

// DefaultKeyMap returns the default vim-style keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Navigation - vim style
		Left: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("h/←", "left panel"),
		),
		Right: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("l/→", "right panel"),
		),
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/↓", "down"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("Tab", "next panel"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("Shift+Tab", "prev panel"),
		),
		Panel1: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "panel 1"),
		),
		Panel2: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "panel 2"),
		),
		Panel3: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "panel 3"),
		),

		// Scrolling - vim style
		PageUp: key.NewBinding(
			key.WithKeys("ctrl+b", "pgup"),
			key.WithHelp("Ctrl+b", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("ctrl+f", "pgdown"),
			key.WithHelp("Ctrl+f", "page down"),
		),
		HalfUp: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("Ctrl+u", "half page up"),
		),
		HalfDown: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("Ctrl+d", "half page down"),
		),
		Top: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "go to top"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("G"),
			key.WithHelp("G", "go to bottom"),
		),

		// View modes
		ToggleCenter: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "cycle center mode"),
		),
		ToggleRight: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "cycle right mode"),
		),
		FocusMode: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "focus mode"),
		),
		MetricsView: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "metrics view"),
		),
		DiffView: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "diff view"),
		),
		ZenMode: key.NewBinding(
			key.WithKeys("z"),
			key.WithHelp("z", "zen mode"),
		),

		// Workflow control
		Pause: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("Space", "pause"),
		),
		Resume: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("Space", "resume"),
		),
		NextStep: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "next step"),
		),
		Skip: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "skip step"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("Enter", "confirm"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("Esc", "cancel"),
		),

		// File operations
		Open: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "open/expand"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit file"),
		),
		CopyPath: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "copy path"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "refresh"),
		),

		// Search
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		NextResult: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "next result"),
		),
		PrevResult: key.NewBinding(
			key.WithKeys("N"),
			key.WithHelp("N", "prev result"),
		),
		ClearSearch: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("Esc", "clear search"),
		),

		// Session
		SaveSession: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("Ctrl+s", "save session"),
		),
		OpenSession: key.NewBinding(
			key.WithKeys("ctrl+o"),
			key.WithHelp("Ctrl+o", "open session"),
		),
		Replay: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("Ctrl+r", "replay"),
		),

		// General
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
		ForceQuit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("Ctrl+c", "force quit"),
		),
	}
}

// ShortHelp returns a subset of bindings for the help bar.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Help,
		k.Tab,
		k.ToggleCenter,
		k.Pause,
		k.Quit,
	}
}

// FullHelp returns all bindings grouped by category.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		// Navigation
		{k.Left, k.Right, k.Up, k.Down, k.Tab, k.Panel1, k.Panel2, k.Panel3},
		// Scrolling
		{k.PageUp, k.PageDown, k.HalfUp, k.HalfDown, k.Top, k.Bottom},
		// Views
		{k.ToggleCenter, k.ToggleRight, k.FocusMode, k.MetricsView, k.DiffView, k.ZenMode},
		// Workflow
		{k.Pause, k.NextStep, k.Skip, k.Confirm, k.Cancel},
		// Files
		{k.Open, k.Edit, k.CopyPath, k.Refresh},
		// Search
		{k.Search, k.NextResult, k.PrevResult},
		// Session
		{k.SaveSession, k.OpenSession, k.Replay},
		// General
		{k.Help, k.Quit, k.ForceQuit},
	}
}
