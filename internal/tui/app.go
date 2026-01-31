package tui

import (
	"fmt"
	"strings"
	"time"

	"cooperations/internal/tui/session"
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

type sessionSavedMsg struct {
	ID  string
	Err error
}

type sessionLoadedMsg struct {
	Session *session.Session
	Err     error
}

type replayDoneMsg struct {
	Err error
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

		case hookNotify, ok := <-s.HookNotify:
			if !ok {
				return nil
			}
			return streamMsg{event: hookNotify}

		case rvrEvent, ok := <-s.RVR:
			if !ok {
				return nil
			}
			return streamMsg{event: rvrEvent}

		case rvrResult, ok := <-s.RVRResult:
			if !ok {
				return nil
			}
			return streamMsg{event: rvrResult}
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

	case sessionSavedMsg:
		if msg.Err != nil {
			m.ShowToast(fmt.Sprintf("Session save failed: %v", msg.Err), widgets.ToastLevelError)
		} else {
			m.ShowToast(fmt.Sprintf("Session saved (%s)", msg.ID), widgets.ToastLevelSuccess)
			m.AddLogEntry(widgets.LogInfo, "session", fmt.Sprintf("Saved %s", msg.ID))
		}

	case sessionLoadedMsg:
		if msg.Err != nil {
			m.ShowToast(fmt.Sprintf("Session load failed: %v", msg.Err), widgets.ToastLevelError)
			break
		}
		if msg.Session != nil {
			m.SessionID = msg.Session.ID
			m.SessionName = msg.Session.Name
			m.CurrentTask = msg.Session.Task
			m.ShowToast(fmt.Sprintf("Session loaded (%s)", msg.Session.ID), widgets.ToastLevelSuccess)
			m.AddLogEntry(widgets.LogInfo, "session", fmt.Sprintf("Loaded %s", msg.Session.ID))
		}

	case replayDoneMsg:
		m.ReplayActive = false
		if msg.Err != nil {
			m.ShowToast(fmt.Sprintf("Replay failed: %v", msg.Err), widgets.ToastLevelError)
			break
		}
		if m.SessionManager != nil && m.SessionManager.Current != nil {
			switch strings.ToLower(m.SessionManager.Current.Status) {
			case "complete":
				m.SetWorkflowState(WorkflowComplete)
			case "error":
				m.SetWorkflowState(WorkflowError)
			case "paused":
				m.SetWorkflowState(WorkflowPaused)
			default:
				m.SetWorkflowState(WorkflowRunning)
			}
		}
		m.ShowToast("Replay finished", widgets.ToastLevelSuccess)
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

	// Search navigation (works in dashboard and focus)
	if (m.ViewMode == ViewModeDashboard || m.ViewMode == ViewModeFocus) && m.SearchQuery != "" {
		switch {
		case key.Matches(msg, m.Keys.NextResult):
			m.jumpSearch(1)
			return nil
		case key.Matches(msg, m.Keys.PrevResult):
			m.jumpSearch(-1)
			return nil
		}
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
			if m.SessionManager != nil {
				m.SessionManager.SetStatus("paused")
			}
			m.ShowToast("Workflow paused", widgets.ToastLevelInfo)
			if m.Stream != nil {
				select {
				case m.Stream.Pause <- true:
				default:
				}
			}
		} else if m.WorkflowState == WorkflowPaused {
			m.SetWorkflowState(WorkflowRunning)
			if m.SessionManager != nil {
				m.SessionManager.SetStatus("running")
			}
			m.ShowToast("Workflow resumed", widgets.ToastLevelInfo)
			if m.Stream != nil {
				select {
				case m.Stream.Pause <- false:
				default:
				}
			}
		}

	case key.Matches(msg, m.Keys.NextStep):
		if m.WorkflowState == WorkflowPaused || m.WorkflowState == WorkflowRunning {
			// Send step signal - advances one agent then auto-pauses
			if m.Stream != nil {
				m.Stream.SendControl(stream.ControlStep, "user requested step")
			}
			m.SetWorkflowState(WorkflowRunning)
			if m.SessionManager != nil {
				m.SessionManager.SetStatus("stepping")
			}
			m.ShowToast("Stepping to next agent...", widgets.ToastLevelInfo)
		} else {
			m.ShowToast("No workflow running", widgets.ToastLevelInfo)
		}

	case key.Matches(msg, m.Keys.Skip):
		if m.WorkflowState == WorkflowRunning || m.WorkflowState == WorkflowPaused {
			// Show confirmation before skipping
			m.ShowConfirm("Skip Agent",
				fmt.Sprintf("Skip current agent (%s) and advance to next?", m.CurrentAgent),
				false)
			m.PendingAction = "skip"
		} else {
			m.ShowToast("No workflow running", widgets.ToastLevelWarning)
		}

	case key.Matches(msg, m.Keys.Kill):
		if m.WorkflowState == WorkflowRunning || m.WorkflowState == WorkflowPaused {
			// Show confirmation before killing
			m.ShowConfirm("Kill Workflow",
				"This will immediately abort the workflow. Continue?",
				true)
			m.PendingAction = "kill"
		} else {
			m.ShowToast("No workflow to kill", widgets.ToastLevelWarning)
		}

	case key.Matches(msg, m.Keys.Search):
		m.SearchMode = true
		m.SearchQuery = ""

	case key.Matches(msg, m.Keys.ClearSearch):
		if m.ViewMode == ViewModeDashboard && m.SearchQuery != "" {
			m.runSearch("")
			m.SearchMode = false
		}

	case key.Matches(msg, m.Keys.SaveSession):
		if m.SessionManager == nil {
			m.ShowToast("Session manager unavailable", widgets.ToastLevelWarning)
			return nil
		}
		if m.CurrentTask != "" {
			m.ensureSession(m.CurrentTask)
		}
		return m.saveSessionCmd()

	case key.Matches(msg, m.Keys.OpenSession):
		if m.SessionManager == nil {
			m.ShowToast("Session manager unavailable", widgets.ToastLevelWarning)
			return nil
		}
		dialog := widgets.NewInputDialog("Open session", "Enter session ID", m.Width/2)
		dialog.Placeholder = "session_..."
		m.InputDialog = &dialog
		m.DecisionDialog = nil
		m.ConfirmDialog = nil
		m.InputMode = InputModeOpenSession
		m.ShowDialog = true

	case key.Matches(msg, m.Keys.Replay):
		if m.SessionManager == nil {
			m.ShowToast("Session manager unavailable", widgets.ToastLevelWarning)
			return nil
		}
		if m.SessionManager.Current == nil {
			m.ShowToast("Load a session first (Ctrl+O)", widgets.ToastLevelWarning)
			return nil
		}
		m.ReplayActive = true
		m.resetForReplay()
		m.SetWorkflowState(WorkflowRunning)
		m.ShowToast(fmt.Sprintf("Replaying %s", m.SessionManager.Current.ID), widgets.ToastLevelInfo)
		return m.replaySessionCmd()

	case key.Matches(msg, m.Keys.Open):
		if m.Dashboard != nil && m.Dashboard.ActivePanel == 2 && m.Dashboard.RightMode == 1 {
			m.Dashboard.FileTree.Toggle()
		}
	case key.Matches(msg, m.Keys.CopyPath):
		if m.Dashboard != nil && m.Dashboard.ActivePanel == 2 && m.Dashboard.RightMode == 1 {
			path := m.Dashboard.FileTree.GetSelected()
			if path == "" {
				m.ShowToast("No file selected", widgets.ToastLevelWarning)
				return nil
			}
			absPath := m.ResolvePath(path)
			if err := copyToClipboard(absPath); err != nil {
				m.ShowToast("Copy failed: "+err.Error(), widgets.ToastLevelWarning)
			} else {
				m.ShowToast("Copied: "+path, widgets.ToastLevelSuccess)
			}
		}

	case key.Matches(msg, m.Keys.Edit):
		if m.Dashboard != nil && m.Dashboard.ActivePanel == 2 && m.Dashboard.RightMode == 1 {
			path := m.Dashboard.FileTree.GetSelected()
			if path == "" {
				m.ShowToast("No file selected", widgets.ToastLevelWarning)
				return nil
			}
			absPath := m.ResolvePath(path)
			if err := openInEditor(absPath); err != nil {
				m.ShowToast("Open failed: "+err.Error(), widgets.ToastLevelWarning)
			} else {
				m.ShowToast("Opened: "+path, widgets.ToastLevelInfo)
			}
		}

	case key.Matches(msg, m.Keys.Refresh):
		m.RefreshFileTree()
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
		m.PendingAction = ""
		m.HideDialog()
	case msg.String() == "y":
		m.ConfirmDialog.Selected = 1
		return m.executeConfirmedAction()
	case msg.String() == "n":
		m.PendingAction = ""
		m.HideDialog()
	case key.Matches(msg, m.Keys.Left), key.Matches(msg, m.Keys.Right), key.Matches(msg, m.Keys.Tab):
		m.ConfirmDialog.Toggle()
	case key.Matches(msg, m.Keys.Confirm):
		if m.ConfirmDialog.IsYes() {
			return m.executeConfirmedAction()
		}
		m.PendingAction = ""
		m.HideDialog()
	}
	return nil
}

// executeConfirmedAction handles the action after confirm dialog is accepted.
func (m *Model) executeConfirmedAction() tea.Cmd {
	action := m.PendingAction
	m.PendingAction = ""
	m.HideDialog()

	switch action {
	case "skip":
		if m.Stream != nil {
			m.Stream.SendControl(stream.ControlSkip, "user skipped agent")
		}
		m.ShowToast("Skipping current agent...", widgets.ToastLevelWarning)
		return nil

	case "kill":
		if m.Stream != nil {
			m.Stream.SendControl(stream.ControlKill, "user killed workflow")
		}
		m.SetWorkflowState(WorkflowError)
		m.ShowToast("Workflow killed", widgets.ToastLevelError)
		return nil

	case "quit", "":
		// Default behavior - quit
		return tea.Quit
	}

	return nil
}

func (m *Model) handleInputDialog(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyEsc:
		if m.InputMode == InputModeOpenSession {
			m.HideDialog()
			return nil
		}
		m.sendDecision(stream.DecisionReject, "edit cancelled", "")
	case tea.KeyEnter:
		if m.InputMode == InputModeOpenSession {
			sessionID := strings.TrimSpace(m.InputDialog.Value)
			if sessionID == "" {
				m.ShowToast("Session ID required", widgets.ToastLevelWarning)
				return nil
			}
			m.HideDialog()
			return m.loadSessionCmd(sessionID)
		}
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
	if m.SessionManager != nil {
		m.SessionManager.RecordEvent("decision_response", stream.HumanDecision{
			RequestID: m.PendingDecision.ID,
			Action:    action,
			Comment:   comment,
			Edited:    edited,
		})
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
	switch msg.Type {
	case tea.KeyEsc:
		m.SearchMode = false
		m.runSearch("")
	case tea.KeyEnter:
		m.SearchMode = false
		found := m.runSearch(m.SearchQuery)
		if found {
			m.ShowToast(fmt.Sprintf("Found %d matches", len(m.SearchResults)), widgets.ToastLevelInfo)
		} else if m.SearchQuery != "" {
			m.ShowToast("No matches", widgets.ToastLevelWarning)
		}
	case tea.KeyBackspace:
		if len(m.SearchQuery) > 0 {
			m.SearchQuery = m.SearchQuery[:len(m.SearchQuery)-1]
		}
	case tea.KeyRunes:
		for _, r := range msg.Runes {
			m.SearchQuery += string(r)
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

	case key.Matches(msg, m.Keys.Search):
		m.SearchMode = true
		m.SearchQuery = ""

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

func (m *Model) stepKeyForName(name string) (key, label string, isRole bool) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", "", false
	}
	lower := strings.ToLower(trimmed)
	switch lower {
	case "architect", "implementer", "reviewer", "navigator", "human":
		return "role:" + lower, titleCase(lower), true
	case "user":
		return "role:user", "User", true
	default:
		return "stage:" + lower, trimmed, false
	}
}

func titleCase(value string) string {
	if value == "" {
		return value
	}
	return strings.ToUpper(value[:1]) + value[1:]
}

func (m *Model) ensureWorkflowStep(key, name, description, agent string) int {
	if m.Dashboard == nil || m.Dashboard.WorkflowSteps == nil {
		return -1
	}
	if m.WorkflowStepIndex == nil {
		m.WorkflowStepIndex = map[string]int{}
	}
	if idx, ok := m.WorkflowStepIndex[key]; ok {
		step := &m.Dashboard.WorkflowSteps.Steps[idx]
		if name != "" {
			step.Name = name
		}
		if description != "" {
			step.Description = description
		}
		if agent != "" {
			step.Agent = agent
		}
		return idx
	}

	m.Dashboard.WorkflowSteps.AddStep(name, description, agent)
	idx := len(m.Dashboard.WorkflowSteps.Steps) - 1
	m.WorkflowStepIndex[key] = idx
	m.TotalSteps = len(m.Dashboard.WorkflowSteps.Steps)
	return idx
}

func (m *Model) updateWorkflowFromProgress(update stream.ProgressUpdate) {
	if m.Dashboard == nil || m.Dashboard.WorkflowSteps == nil {
		return
	}

	stage := strings.TrimSpace(update.Stage)
	key, label, isRole := m.stepKeyForName(stage)
	if key == "" {
		if update.Message == "" {
			return
		}
		key = "message:" + strings.ToLower(update.Message)
		label = update.Message
	}

	idx := m.ensureWorkflowStep(key, label, update.Message, "")
	if idx < 0 {
		return
	}

	if isRole {
		m.Dashboard.WorkflowSteps.Steps[idx].Agent = label
	}
	if update.Message != "" {
		m.Dashboard.WorkflowSteps.Steps[idx].Description = update.Message
	}

	isComplete := strings.EqualFold(stage, "complete") || update.Percent >= 100
	if isComplete {
		current := m.Dashboard.WorkflowSteps.CurrentStep()
		if current != -1 && current != idx {
			m.Dashboard.WorkflowSteps.SetStatus(current, widgets.StepComplete)
		}
		m.Dashboard.WorkflowSteps.SetStatus(idx, widgets.StepComplete)
		return
	}

	current := m.Dashboard.WorkflowSteps.CurrentStep()
	if current != -1 && current != idx {
		m.Dashboard.WorkflowSteps.SetStatus(current, widgets.StepComplete)
	}
	m.Dashboard.WorkflowSteps.SetStatus(idx, widgets.StepRunning)
}

func (m *Model) updateWorkflowFromHandoff(event stream.HandoffEvent) {
	if m.Dashboard == nil || m.Dashboard.WorkflowSteps == nil {
		return
	}

	if event.From != "" && strings.ToLower(event.From) != "user" {
		key, label, _ := m.stepKeyForName(event.From)
		if key == "" {
			key = "role:" + strings.ToLower(event.From)
			label = titleCase(event.From)
		}
		idx := m.ensureWorkflowStep(key, label, "", label)
		if idx >= 0 {
			m.Dashboard.WorkflowSteps.SetStatus(idx, widgets.StepComplete)
		}
	}

	if event.To != "" {
		key, label, _ := m.stepKeyForName(event.To)
		if key == "" {
			key = "role:" + strings.ToLower(event.To)
			label = titleCase(event.To)
		}
		idx := m.ensureWorkflowStep(key, label, event.Reason, label)
		if idx >= 0 {
			if event.Reason != "" {
				m.Dashboard.WorkflowSteps.Steps[idx].Description = event.Reason
			}
			m.Dashboard.WorkflowSteps.SetStatus(idx, widgets.StepRunning)
		}
	}
}

func extractTaskFromProgress(message string) string {
	lower := strings.ToLower(message)
	const runningPrefix = "running task:"
	const startingPrefix = "starting workflow for task:"
	switch {
	case strings.HasPrefix(lower, runningPrefix):
		return strings.TrimSpace(message[len(runningPrefix):])
	case strings.HasPrefix(lower, startingPrefix):
		return strings.TrimSpace(message[len(startingPrefix):])
	default:
		return ""
	}
}

// handleStreamEvent processes events from the workflow stream.
func (m *Model) handleStreamEvent(event interface{}) tea.Cmd {
	if progress, ok := event.(stream.ProgressUpdate); ok {
		if task := extractTaskFromProgress(progress.Message); task != "" {
			m.CurrentTask = task
		}
	}

	if m.CurrentTask != "" {
		m.ensureSession(m.CurrentTask)
	}
	m.recordStreamEvent(event)

	switch e := event.(type) {
	case stream.TokenChunk:
		m.AppendStreamingContent(e.Token)
		m.SetWorkflowState(WorkflowRunning)
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
		m.SetWorkflowState(WorkflowRunning)
		m.updateWorkflowFromProgress(e)

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
		m.updateWorkflowFromHandoff(e)
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

	case stream.HookNotification:
		// Update UI state based on hook
		m.CanSkip = e.CanSkip
		if e.Paused {
			m.SetWorkflowState(WorkflowPaused)
		}
		// Log the hook event
		m.AddLogEntry(widgets.LogDebug, "hook", fmt.Sprintf("%s: %s", e.Phase, e.Role))

	case stream.RVREvent:
		// RVR processing event
		if e.Confidence < e.Threshold {
			level := widgets.ToastLevelWarning
			if e.Confidence < 0.4 {
				level = widgets.ToastLevelError
			}
			m.ShowToast(fmt.Sprintf("RVR confidence %.0f%% (chunk %d)", e.Confidence*100, e.ChunkID), level)
		}
		m.AddLogEntry(widgets.LogDebug, "rvr", fmt.Sprintf("%s: chunk=%d conf=%.2f", e.Phase, e.ChunkID, e.Confidence))

	case stream.RVRResultEvent:
		// RVR final results
		level := widgets.ToastLevelSuccess
		if e.Overall < 0.8 {
			level = widgets.ToastLevelWarning
		}
		if e.Overall < 0.4 {
			level = widgets.ToastLevelError
		}
		m.ShowToast(fmt.Sprintf("RVR overall: %.0f%%", e.Overall*100), level)
		if len(e.Caveats) > 0 {
			m.AddLogEntry(widgets.LogWarn, "rvr", fmt.Sprintf("Caveats: %v", e.Caveats))
		}

	case string:
		if e == "done" {
			m.SetWorkflowState(WorkflowComplete)
			if m.SessionManager != nil {
				m.SessionManager.SetStatus("complete")
			}
			m.ShowToast("Workflow completed", widgets.ToastLevelSuccess)
		}

	case error:
		m.LastError = e
		m.SetWorkflowState(WorkflowError)
		if m.SessionManager != nil {
			m.SessionManager.SetStatus("error")
		}
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

func (m *Model) saveSessionCmd() tea.Cmd {
	return func() tea.Msg {
		if m.SessionManager == nil {
			return sessionSavedMsg{Err: fmt.Errorf("session manager unavailable")}
		}
		if m.SessionManager.Current == nil {
			return sessionSavedMsg{Err: fmt.Errorf("no active session")}
		}
		sessionID := m.SessionManager.Current.ID
		err := m.SessionManager.Save()
		return sessionSavedMsg{ID: sessionID, Err: err}
	}
}

func (m *Model) loadSessionCmd(sessionID string) tea.Cmd {
	return func() tea.Msg {
		if m.SessionManager == nil {
			return sessionLoadedMsg{Err: fmt.Errorf("session manager unavailable")}
		}
		sessionID = strings.TrimSpace(sessionID)
		sessionID = strings.TrimSuffix(sessionID, ".json")
		loaded, err := m.SessionManager.Load(sessionID)
		return sessionLoadedMsg{Session: loaded, Err: err}
	}
}

func (m *Model) replaySessionCmd() tea.Cmd {
	return func() tea.Msg {
		if m.SessionManager == nil || m.Stream == nil {
			return replayDoneMsg{Err: fmt.Errorf("session replay unavailable")}
		}
		if m.SessionManager.Current == nil {
			return replayDoneMsg{Err: fmt.Errorf("no session loaded")}
		}
		err := m.SessionManager.Replay(m.SessionManager.Current, m.Stream, m.ReplaySpeed)
		return replayDoneMsg{Err: err}
	}
}

// Run starts the TUI application.
func Run(workflowStream *stream.WorkflowStream) error {
	return RunWithTask(workflowStream, "")
}

// RunWithTask starts the TUI application with an initial task label.
func RunWithTask(workflowStream *stream.WorkflowStream, task string) error {
	model := NewModelWithTask(workflowStream, task)
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}
