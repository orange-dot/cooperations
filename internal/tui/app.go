package tui

import (
	"fmt"
	"strings"
	"time"

	"cooperations/internal/tui/stream"
	"cooperations/internal/tui/widgets"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// tickMsg is sent periodically for animations.
type tickMsg time.Time

// streamMsg wraps stream events for the update loop.
type streamMsg struct {
	event interface{}
}

// Init initializes the Bubble Tea program.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		tickCmd(m.TickInterval),
		listenForStreams(m.Stream),
	)
}

// tickCmd returns a command that sends tick messages.
func tickCmd(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// listenForStreams returns a command that listens for stream events.
func listenForStreams(s *stream.WorkflowStream) tea.Cmd {
	if s == nil {
		return nil
	}

	return func() tea.Msg {
		select {
		case token, ok := <-s.Tokens:
			if !ok {
				return nil
			}
			return streamMsg{event: token}

		case progress, ok := <-s.Progress:
			if !ok {
				return nil
			}
			return streamMsg{event: progress}

		case handoff, ok := <-s.Handoffs:
			if !ok {
				return nil
			}
			return streamMsg{event: handoff}

		case code, ok := <-s.Code:
			if !ok {
				return nil
			}
			return streamMsg{event: code}

		case diff, ok := <-s.FileDiff:
			if !ok {
				return nil
			}
			return streamMsg{event: diff}

		case tree, ok := <-s.FileTree:
			if !ok {
				return nil
			}
			return streamMsg{event: tree}

		case log, ok := <-s.AgentLog:
			if !ok {
				return nil
			}
			return streamMsg{event: log}

		case metrics, ok := <-s.Metrics:
			if !ok {
				return nil
			}
			return streamMsg{event: metrics}

		case thinking, ok := <-s.Thinking:
			if !ok {
				return nil
			}
			return streamMsg{event: thinking}

		case toast, ok := <-s.Toast:
			if !ok {
				return nil
			}
			return streamMsg{event: toast}

		case decision, ok := <-s.Decision:
			if !ok {
				return nil
			}
			return streamMsg{event: decision}

		case session, ok := <-s.Session:
			if !ok {
				return nil
			}
			return streamMsg{event: session}

		case <-s.Done:
			return streamMsg{event: "done"}

		case err, ok := <-s.Error:
			if !ok {
				return nil
			}
			return streamMsg{event: err}
		}
	}
}

// Update handles all messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		cmd := m.handleKeyPress(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		if !m.Ready {
			m.Initialize(msg.Width, msg.Height)
		} else {
			m.Resize(msg.Width, msg.Height)
		}

	case tickMsg:
		m.Tick()
		cmds = append(cmds, tickCmd(m.TickInterval))

	case streamMsg:
		cmd := m.handleStreamEvent(msg.event)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		// Keep listening for more stream events
		cmds = append(cmds, listenForStreams(m.Stream))
	}

	return m, tea.Batch(cmds...)
}

// handleKeyPress handles keyboard input.
func (m *Model) handleKeyPress(msg tea.KeyMsg) tea.Cmd {
	// Handle dialog input first
	if m.ShowDialog {
		return m.handleDialogInput(msg)
	}

	// Handle search mode
	if m.SearchMode {
		return m.handleSearchInput(msg)
	}

	// Handle view-specific keys
	switch m.ViewMode {
	case ViewModeHelp:
		return m.handleHelpKeys(msg)
	case ViewModeFocus:
		return m.handleFocusKeys(msg)
	case ViewModeZen:
		return m.handleZenKeys(msg)
	}

	// Global keys
	switch {
	case key.Matches(msg, m.Keys.ForceQuit):
		return tea.Quit

	case key.Matches(msg, m.Keys.Quit):
		if m.WorkflowState == WorkflowRunning {
			// Show confirmation dialog
			m.ShowConfirm("Quit", "Workflow is still running. Are you sure you want to quit?", true)
			return nil
		}
		return tea.Quit

	case key.Matches(msg, m.Keys.Help):
		m.ToggleHelp()

	case key.Matches(msg, m.Keys.FocusMode):
		m.ToggleFocus()

	case key.Matches(msg, m.Keys.ZenMode):
		m.ToggleZen()

	case key.Matches(msg, m.Keys.Tab):
		if m.Dashboard != nil {
			m.Dashboard.ActivePanel = (m.Dashboard.ActivePanel + 1) % 3
		}

	case key.Matches(msg, m.Keys.ShiftTab):
		if m.Dashboard != nil {
			m.Dashboard.ActivePanel = (m.Dashboard.ActivePanel + 2) % 3
		}

	case key.Matches(msg, m.Keys.Panel1):
		if m.Dashboard != nil {
			m.Dashboard.FocusLeft()
		}

	case key.Matches(msg, m.Keys.Panel2):
		if m.Dashboard != nil {
			m.Dashboard.FocusCenter()
		}

	case key.Matches(msg, m.Keys.Panel3):
		if m.Dashboard != nil {
			m.Dashboard.FocusRight()
		}

	case key.Matches(msg, m.Keys.Left):
		if m.Dashboard != nil && m.Dashboard.ActivePanel > 0 {
			m.Dashboard.ActivePanel--
		}

	case key.Matches(msg, m.Keys.Right):
		if m.Dashboard != nil && m.Dashboard.ActivePanel < 2 {
			m.Dashboard.ActivePanel++
		}

	case key.Matches(msg, m.Keys.ToggleCenter):
		if m.Dashboard != nil {
			m.Dashboard.CycleCenter()
		}

	case key.Matches(msg, m.Keys.ToggleRight):
		if m.Dashboard != nil {
			m.Dashboard.CycleRight()
		}
	case key.Matches(msg, m.Keys.MetricsView):
		if m.Dashboard != nil {
			m.Dashboard.RightMode = 2
			m.Dashboard.ActivePanel = 2
		}
	case key.Matches(msg, m.Keys.DiffView):
		if m.ViewMode == ViewModeFocus && m.Focus != nil {
			m.Focus.Mode = 2
		} else if m.Dashboard != nil {
			m.Dashboard.CenterMode = 2
			m.Dashboard.ActivePanel = 1
		}

	case key.Matches(msg, m.Keys.Up):
		m.scrollUp(1)

	case key.Matches(msg, m.Keys.Down):
		m.scrollDown(1)

	case key.Matches(msg, m.Keys.HalfUp):
		m.scrollUp(m.Height / 2)

	case key.Matches(msg, m.Keys.HalfDown):
		m.scrollDown(m.Height / 2)

	case key.Matches(msg, m.Keys.PageUp):
		m.scrollUp(m.Height - 4)

	case key.Matches(msg, m.Keys.PageDown):
		m.scrollDown(m.Height - 4)

	case key.Matches(msg, m.Keys.Top):
		m.scrollToTop()

	case key.Matches(msg, m.Keys.Bottom):
		m.scrollToBottom()

	case key.Matches(msg, m.Keys.Pause):
		if m.WorkflowState == WorkflowRunning {
			m.SetWorkflowState(WorkflowPaused)
			m.ShowToast("Workflow paused", widgets.ToastLevelInfo)
			if m.Stream != nil {
				select {
				case m.Stream.Pause <- true:
				default:
				}
			}
		} else if m.WorkflowState == WorkflowPaused {
			m.SetWorkflowState(WorkflowRunning)
			m.ShowToast("Workflow resumed", widgets.ToastLevelInfo)
			if m.Stream != nil {
				select {
				case m.Stream.Pause <- false:
				default:
				}
			}
		}

	case key.Matches(msg, m.Keys.Search):
		m.SearchMode = true
		m.SearchQuery = ""

	case key.Matches(msg, m.Keys.Open):
		if m.Dashboard != nil && m.Dashboard.ActivePanel == 2 && m.Dashboard.RightMode == 1 {
			m.Dashboard.FileTree.Toggle()
		}
	case key.Matches(msg, m.Keys.CopyPath):
		if m.Dashboard != nil && m.Dashboard.ActivePanel == 2 && m.Dashboard.RightMode == 1 {
			path := m.Dashboard.FileTree.GetSelected()
			if path != "" {
				m.ShowToast("Selected: "+path, widgets.ToastLevelInfo)
			}
		}
	}

	return nil
}

// handleDialogInput handles input when a dialog is open.
func (m *Model) handleDialogInput(msg tea.KeyMsg) tea.Cmd {
	if m.InputDialog != nil {
		return m.handleInputDialog(msg)
	}
	if m.DecisionDialog != nil {
		return m.handleDecisionDialog(msg)
	}
	if m.ConfirmDialog != nil {
		return m.handleConfirmDialog(msg)
	}
	if key.Matches(msg, m.Keys.Cancel) {
		m.HideDialog()
	}
	return nil
}

func (m *Model) handleDecisionDialog(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.Keys.Cancel):
		m.sendDecision(stream.DecisionReject, "cancelled", "")
		return nil

	case key.Matches(msg, m.Keys.Up):
		m.DecisionDialog.MoveUp()

	case key.Matches(msg, m.Keys.Down):
		m.DecisionDialog.MoveDown()

	case msg.Type == tea.KeyRunes && len(msg.Runes) == 1:
		m.DecisionDialog.SelectByKey(string(msg.Runes[0]))

	case key.Matches(msg, m.Keys.Confirm):
		opt := m.DecisionDialog.GetSelectedOption()
		if opt == nil {
			m.HideDialog()
			return nil
		}

		action := decisionActionFromLabel(opt.Label)
		if action == stream.DecisionEdit {
			input := widgets.NewInputDialog("Edit output", "Provide edited output", m.Width/2)
			m.InputDialog = &input
			m.DecisionDialog = nil
			m.PendingDecisionAction = action
			return nil
		}

		m.sendDecision(action, opt.Label, "")
	}

	return nil
}

func (m *Model) handleConfirmDialog(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.Keys.Cancel):
		m.HideDialog()
	case msg.String() == "y":
		m.ConfirmDialog.Selected = 1
		m.HideDialog()
		return tea.Quit
	case msg.String() == "n":
		m.HideDialog()
	case key.Matches(msg, m.Keys.Left), key.Matches(msg, m.Keys.Right), key.Matches(msg, m.Keys.Tab):
		m.ConfirmDialog.Toggle()
	case key.Matches(msg, m.Keys.Confirm):
		isYes := m.ConfirmDialog.IsYes()
		m.HideDialog()
		if isYes {
			return tea.Quit
		}
	}
	return nil
}

func (m *Model) handleInputDialog(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyEsc:
		m.sendDecision(stream.DecisionReject, "edit cancelled", "")
	case tea.KeyEnter:
		action := m.PendingDecisionAction
		if action == "" {
			action = stream.DecisionEdit
		}
		m.sendDecision(action, "", m.InputDialog.Value)
	case tea.KeyBackspace:
		m.InputDialog.Backspace()
	case tea.KeyDelete:
		m.InputDialog.Delete()
	case tea.KeyLeft:
		m.InputDialog.MoveLeft()
	case tea.KeyRight:
		m.InputDialog.MoveRight()
	case tea.KeyRunes:
		for _, r := range msg.Runes {
			m.InputDialog.Insert(r)
		}
	}
	return nil
}

func (m *Model) sendDecision(action stream.DecisionAction, comment, edited string) {
	if m.PendingDecision == nil {
		m.HideDialog()
		return
	}
	if action == "" {
		action = stream.DecisionApprove
	}
	if m.Stream != nil {
		select {
		case m.Stream.Response <- stream.HumanDecision{
			RequestID: m.PendingDecision.ID,
			Action:    action,
			Comment:   comment,
			Edited:    edited,
		}:
		default:
		}
	}
	m.PendingDecision = nil
	m.PendingDecisionAction = ""
	m.HideDialog()
}

func decisionActionFromLabel(label string) stream.DecisionAction {
	lower := strings.ToLower(label)
	switch {
	case strings.Contains(lower, "reject"), strings.Contains(lower, "deny"), strings.Contains(lower, "change"), strings.Contains(lower, "request"):
		return stream.DecisionReject
	case strings.Contains(lower, "edit"):
		return stream.DecisionEdit
	default:
		return stream.DecisionApprove
	}
}

// handleSearchInput handles input in search mode.
func (m *Model) handleSearchInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.SearchMode = false
		m.SearchQuery = ""
	case "enter":
		m.SearchMode = false
		// TODO: Perform search
	case "backspace":
		if len(m.SearchQuery) > 0 {
			m.SearchQuery = m.SearchQuery[:len(m.SearchQuery)-1]
		}
	default:
		if len(msg.String()) == 1 {
			m.SearchQuery += msg.String()
		}
	}
	return nil
}

// handleHelpKeys handles keys in help view.
func (m *Model) handleHelpKeys(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.Keys.Cancel), key.Matches(msg, m.Keys.Help):
		m.ViewMode = m.PreviousMode

	case key.Matches(msg, m.Keys.Up):
		if m.Help != nil {
			m.Help.ScrollUp(1)
		}

	case key.Matches(msg, m.Keys.Down):
		if m.Help != nil {
			m.Help.ScrollDown(1)
		}

	case key.Matches(msg, m.Keys.ForceQuit):
		return tea.Quit
	}

	return nil
}

// handleFocusKeys handles keys in focus view.
func (m *Model) handleFocusKeys(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.Keys.Cancel):
		m.ViewMode = ViewModeDashboard

	case key.Matches(msg, m.Keys.ToggleCenter):
		if m.Focus != nil {
			m.Focus.Mode = (m.Focus.Mode + 1) % 4
		}

	case key.Matches(msg, m.Keys.DiffView):
		if m.Focus != nil {
			m.Focus.Mode = 2 // Diff mode
		}

	case key.Matches(msg, m.Keys.Up):
		m.scrollFocusUp(1)

	case key.Matches(msg, m.Keys.Down):
		m.scrollFocusDown(1)

	case key.Matches(msg, m.Keys.ForceQuit):
		return tea.Quit

	case key.Matches(msg, m.Keys.Quit):
		m.ViewMode = ViewModeDashboard
	}

	return nil
}

// handleZenKeys handles keys in zen view.
func (m *Model) handleZenKeys(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.Keys.Cancel), key.Matches(msg, m.Keys.Quit):
		m.ViewMode = ViewModeDashboard

	case key.Matches(msg, m.Keys.ForceQuit):
		return tea.Quit
	}

	return nil
}

// scrollUp scrolls the active panel up.
func (m *Model) scrollUp(lines int) {
	if m.Dashboard == nil {
		return
	}

	switch m.Dashboard.ActivePanel {
	case 1: // Center
		switch m.Dashboard.CenterMode {
		case 0:
			m.Dashboard.StreamingText.ScrollUp(lines)
		case 1:
			m.Dashboard.CodeBlock.ScrollUp(lines)
		case 2:
			m.Dashboard.DiffBlock.ScrollUp(lines)
		}
	case 2: // Right
		switch m.Dashboard.RightMode {
		case 0:
			m.Dashboard.ActivityLog.ScrollUp(lines)
		case 1:
			m.Dashboard.FileTree.MoveUp()
		}
	}
}

// scrollDown scrolls the active panel down.
func (m *Model) scrollDown(lines int) {
	if m.Dashboard == nil {
		return
	}

	switch m.Dashboard.ActivePanel {
	case 1: // Center
		switch m.Dashboard.CenterMode {
		case 0:
			m.Dashboard.StreamingText.ScrollDown(lines)
		case 1:
			m.Dashboard.CodeBlock.ScrollDown(lines)
		case 2:
			m.Dashboard.DiffBlock.ScrollDown(lines)
		}
	case 2: // Right
		switch m.Dashboard.RightMode {
		case 0:
			m.Dashboard.ActivityLog.ScrollDown(lines)
		case 1:
			m.Dashboard.FileTree.MoveDown()
		}
	}
}

// scrollToTop scrolls the active panel to the top.
func (m *Model) scrollToTop() {
	if m.ViewMode == ViewModeFocus && m.Focus != nil {
		switch m.Focus.Mode {
		case 0:
			m.Focus.StreamingText.ScrollToTop()
		case 1:
			m.Focus.CodeBlock.ScrollToTop()
		case 2:
			m.Focus.DiffBlock.ScrollToTop()
		case 3:
			m.Focus.ActivityLog.ScrollToTop()
		}
		return
	}

	if m.Dashboard == nil {
		return
	}

	switch m.Dashboard.ActivePanel {
	case 1:
		switch m.Dashboard.CenterMode {
		case 0:
			m.Dashboard.StreamingText.ScrollToTop()
		case 1:
			m.Dashboard.CodeBlock.ScrollToTop()
		case 2:
			m.Dashboard.DiffBlock.ScrollToTop()
		}
	case 2:
		switch m.Dashboard.RightMode {
		case 0:
			m.Dashboard.ActivityLog.ScrollToTop()
		case 1:
			m.Dashboard.FileTree.ScrollToTop()
		}
	}
}

// scrollToBottom scrolls the active panel to the bottom.
func (m *Model) scrollToBottom() {
	if m.ViewMode == ViewModeFocus && m.Focus != nil {
		switch m.Focus.Mode {
		case 0:
			m.Focus.StreamingText.ScrollToBottom()
		case 1:
			m.Focus.CodeBlock.ScrollToBottom()
		case 2:
			m.Focus.DiffBlock.ScrollToBottom()
		case 3:
			m.Focus.ActivityLog.ScrollToBottom()
		}
		return
	}

	if m.Dashboard == nil {
		return
	}

	switch m.Dashboard.ActivePanel {
	case 1:
		switch m.Dashboard.CenterMode {
		case 0:
			m.Dashboard.StreamingText.ScrollToBottom()
		case 1:
			m.Dashboard.CodeBlock.ScrollToBottom()
		case 2:
			m.Dashboard.DiffBlock.ScrollToBottom()
		}
	case 2:
		switch m.Dashboard.RightMode {
		case 0:
			m.Dashboard.ActivityLog.ScrollToBottom()
		case 1:
			m.Dashboard.FileTree.ScrollToBottom()
		}
	}
}

// scrollFocusUp scrolls focus view up.
func (m *Model) scrollFocusUp(lines int) {
	if m.Focus == nil {
		return
	}

	switch m.Focus.Mode {
	case 0:
		m.Focus.StreamingText.ScrollUp(lines)
	case 1:
		m.Focus.CodeBlock.ScrollUp(lines)
	case 2:
		m.Focus.DiffBlock.ScrollUp(lines)
	case 3:
		m.Focus.ActivityLog.ScrollUp(lines)
	}
}

// scrollFocusDown scrolls focus view down.
func (m *Model) scrollFocusDown(lines int) {
	if m.Focus == nil {
		return
	}

	switch m.Focus.Mode {
	case 0:
		m.Focus.StreamingText.ScrollDown(lines)
	case 1:
		m.Focus.CodeBlock.ScrollDown(lines)
	case 2:
		m.Focus.DiffBlock.ScrollDown(lines)
	case 3:
		m.Focus.ActivityLog.ScrollDown(lines)
	}
}

// handleStreamEvent processes events from the workflow stream.
func (m *Model) handleStreamEvent(event interface{}) tea.Cmd {
	switch e := event.(type) {
	case stream.TokenChunk:
		m.AppendStreamingContent(e.Token)
		if e.AgentRole != "" && e.AgentRole != m.CurrentAgent {
			m.SetCurrentAgent(e.AgentRole, "Generating response...")
		}
		if e.IsFinal {
			if m.Dashboard != nil {
				m.Dashboard.StreamingText.EndStream()
			}
			if m.Focus != nil {
				m.Focus.StreamingText.EndStream()
			}
			if m.Zen != nil {
				m.Zen.ShowCursor = false
			}
		}

	case stream.ProgressUpdate:
		m.UpdateProgress(e.Percent, e.Message)

	case stream.HandoffEvent:
		if m.Dashboard != nil {
			// Mark previous agent as complete
			if e.From != "" {
				m.ClearCurrentAgent(e.From, true)
			}
			// Set new agent as active
			if e.To != "" {
				m.SetCurrentAgent(e.To, e.Reason)
			}
		}
		m.AddLogEntry(widgets.LogInfo, e.From, fmt.Sprintf("Handoff to %s: %s", e.To, e.Reason))

	case stream.CodeUpdate:
		m.SetCodeContent(e.Content, e.Language, e.Path)
		if e.Path != "" {
			m.AddFile(e.Path, widgets.FileStatusModified)
		}

	case stream.FileDiff:
		// Generate unified diff from hunks
		var diffContent string
		for _, hunk := range e.Hunks {
			for _, line := range hunk.Lines {
				switch line.Type {
				case "add":
					diffContent += "+" + line.Content + "\n"
				case "remove":
					diffContent += "-" + line.Content + "\n"
				default:
					diffContent += " " + line.Content + "\n"
				}
			}
		}
		m.SetDiffContent(diffContent, e.Path)
		m.AddFile(e.Path, widgets.FileStatusModified)

	case stream.FileTreeUpdate:
		switch strings.ToLower(e.Action) {
		case "delete":
			if m.Dashboard != nil {
				m.Dashboard.FileTree.RemoveFile(e.Path)
			}
		default:
			status := widgets.FileStatusModified
			if strings.ToLower(e.Action) == "add" {
				status = widgets.FileStatusAdded
			}
			if m.Dashboard != nil {
				m.Dashboard.FileTree.AddPath(e.Path, status, e.IsDir)
			}
		}
		if e.Path != "" {
			m.AddLogEntry(widgets.LogInfo, "filesystem", fmt.Sprintf("%s %s", e.Action, e.Path))
		}

	case stream.AgentLogEntry:
		level := widgets.LogInfo
		switch e.Level {
		case "debug":
			level = widgets.LogDebug
		case "warn":
			level = widgets.LogWarn
		case "error":
			level = widgets.LogError
		}
		m.AddLogEntry(level, e.AgentRole, e.Message)

	case stream.MetricsSnapshot:
		m.UpdateMetrics(e.PromptTokens, e.CompletionTokens)
		m.UpdateMetricsSnapshot(e)

	case stream.ThinkingUpdate:
		if m.Dashboard != nil {
			m.Dashboard.AgentPanel.SetStatus(e.AgentRole, widgets.AgentThinking, e.Stage)
		}
		if m.Focus != nil && e.AgentRole != "" {
			m.Focus.SetActiveAgent(e.AgentRole)
		}

	case stream.ToastNotification:
		level := widgets.ToastLevelInfo
		switch e.Level {
		case "success":
			level = widgets.ToastLevelSuccess
		case "warning":
			level = widgets.ToastLevelWarning
		case "error":
			level = widgets.ToastLevelError
		}
		m.ShowToast(e.Message, level)

	case stream.DecisionRequest:
		reqCopy := e
		m.PendingDecision = &reqCopy
		m.PendingDecisionAction = ""
		var options []widgets.DecisionOption
		for i, opt := range e.Options {
			options = append(options, widgets.DecisionOption{
				Key:   fmt.Sprintf("%d", i+1),
				Label: opt,
			})
		}
		m.ShowDecision(e.Title, e.Prompt, options)

	case stream.SessionEvent:
		m.AddLogEntry(widgets.LogInfo, "session", fmt.Sprintf("%s (%s)", e.Type, e.SessionID))

	case string:
		if e == "done" {
			m.SetWorkflowState(WorkflowComplete)
			m.ShowToast("Workflow completed", widgets.ToastLevelSuccess)
		}

	case error:
		m.LastError = e
		m.SetWorkflowState(WorkflowError)
		m.ShowToast(fmt.Sprintf("Error: %v", e), widgets.ToastLevelError)
		m.AddLogEntry(widgets.LogError, "", e.Error())
	}

	return nil
}

// View renders the current view.
func (m Model) View() string {
	if !m.Ready {
		return "Initializing..."
	}

	// Render base view
	var content string
	switch m.ViewMode {
	case ViewModeHelp:
		content = m.Help.View()
	case ViewModeFocus:
		content = m.Focus.View()
	case ViewModeZen:
		content = m.Zen.View()
	default:
		content = m.Dashboard.View()
	}

	// Overlay dialog if showing
	if m.ShowDialog {
		switch {
		case m.InputDialog != nil:
			content += "\n" + m.InputDialog.View()
		case m.DecisionDialog != nil:
			content += "\n" + m.DecisionDialog.View()
		case m.ConfirmDialog != nil:
			content += "\n" + m.ConfirmDialog.View()
		}
	}

	// Search bar if active
	if m.SearchMode {
		searchBar := fmt.Sprintf("/%s█", m.SearchQuery)
		content += "\n" + searchBar
	}

	return content
}

// Run starts the TUI application.
func Run(workflowStream *stream.WorkflowStream) error {
	model := NewModel(workflowStream)
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}
