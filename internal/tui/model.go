package tui

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cooperations/internal/tui/session"
	"cooperations/internal/tui/stream"
	"cooperations/internal/tui/styles"
	"cooperations/internal/tui/views"
	"cooperations/internal/tui/widgets"
)

// ViewMode represents the current view mode.
type ViewMode int

const (
	ViewModeDashboard ViewMode = iota
	ViewModeFocus
	ViewModeHelp
	ViewModeZen
)

// WorkflowState represents the current workflow state.
type WorkflowState int

const (
	WorkflowIdle WorkflowState = iota
	WorkflowRunning
	WorkflowPaused
	WorkflowComplete
	WorkflowError
)

// InputMode indicates what the input dialog is used for.
type InputMode int

const (
	InputModeNone InputMode = iota
	InputModeDecisionEdit
	InputModeOpenSession
)

// SearchTarget represents which view is being searched.
type SearchTarget int

const (
	SearchTargetNone SearchTarget = iota
	SearchTargetStreaming
	SearchTargetCode
	SearchTargetDiff
)

// Model is the main TUI application state.
type Model struct {
	// Dimensions
	Width  int
	Height int
	Ready  bool

	// View state
	ViewMode     ViewMode
	PreviousMode ViewMode
	ShowHelp     bool
	ShowDialog   bool

	// Views
	Dashboard *views.DashboardView
	Focus     *views.FocusView
	Help      *views.HelpView
	Zen       *views.ZenView

	// Workflow state
	WorkflowState     WorkflowState
	CurrentAgent      string
	CurrentTask       string
	TotalSteps        int
	CompletedSteps    int
	WorkflowStepIndex map[string]int

	// Stream for receiving updates
	Stream *stream.WorkflowStream

	// Dialogs
	DecisionDialog        *widgets.DecisionDialog
	ConfirmDialog         *widgets.ConfirmDialog
	InputDialog           *widgets.InputDialog
	PendingDecision       *stream.DecisionRequest
	PendingDecisionAction stream.DecisionAction
	InputMode             InputMode
	PendingAction         string // "skip", "kill", "quit" for confirm dialogs

	// Workflow control state
	StepMode bool // Auto-pause after each agent
	CanSkip  bool // Whether skip is available at current phase

	// Input state
	Keys          KeyMap
	SearchMode    bool
	SearchQuery   string
	SearchResults []int
	SearchIndex   int
	SearchTarget  SearchTarget

	// Timing
	StartTime    time.Time
	LastTick     time.Time
	TickInterval time.Duration

	// Session
	SessionID      string
	SessionName    string
	SessionDir     string
	RepoRoot       string
	SessionManager *session.Manager
	SessionInitErr error
	ReplayActive   bool
	ReplaySpeed    float64

	// Errors
	LastError error
}

// NewModel creates a new TUI model.
func NewModel(workflowStream *stream.WorkflowStream) Model {
	return NewModelWithTask(workflowStream, "")
}

// NewModelWithTask creates a new TUI model with an initial task label.
func NewModelWithTask(workflowStream *stream.WorkflowStream, task string) Model {
	sessionDir := os.Getenv("COOPERATIONS_DIR")
	if sessionDir == "" {
		sessionDir = ".cooperations"
	}
	sessionDir = filepath.Join(sessionDir, "tui_sessions")

	repoRoot, _ := os.Getwd()
	manager, err := session.NewManager(sessionDir)

	model := Model{
		Stream:            workflowStream,
		Keys:              DefaultKeyMap(),
		TickInterval:      100 * time.Millisecond,
		StartTime:         time.Now(),
		SessionDir:        sessionDir,
		RepoRoot:          repoRoot,
		SessionManager:    manager,
		SessionInitErr:    err,
		ReplaySpeed:       1.0,
		WorkflowStepIndex: map[string]int{},
	}

	if task != "" && manager != nil {
		current := manager.NewSession(task)
		model.SessionID = current.ID
		model.SessionName = current.Name
		model.CurrentTask = task
	}

	return model
}

// Initialize sets up the views with the given dimensions.
func (m *Model) Initialize(width, height int) {
	m.Width = width
	m.Height = height
	m.Ready = true

	// Create views
	m.Dashboard = views.NewDashboardView(width, height)
	m.Focus = views.NewFocusView(width, height)
	m.Help = views.NewHelpView(width, height)
	m.Zen = views.NewZenView(width, height)

	// Set initial view mode
	m.ViewMode = ViewModeDashboard

	if m.SessionInitErr != nil {
		m.ShowToast(fmt.Sprintf("Session init failed: %v", m.SessionInitErr), widgets.ToastLevelWarning)
	}
}

// Resize updates all views with new dimensions.
func (m *Model) Resize(width, height int) {
	m.Width = width
	m.Height = height

	if m.Dashboard != nil {
		m.Dashboard.Resize(width, height)
	}
	if m.Focus != nil {
		m.Focus.Resize(width, height)
	}
	if m.Help != nil {
		m.Help.Resize(width, height)
	}
	if m.Zen != nil {
		m.Zen.Resize(width, height)
	}
}

// SetViewMode changes the current view mode.
func (m *Model) SetViewMode(mode ViewMode) {
	m.PreviousMode = m.ViewMode
	m.ViewMode = mode
}

// ToggleHelp toggles the help overlay.
func (m *Model) ToggleHelp() {
	if m.ViewMode == ViewModeHelp {
		m.ViewMode = m.PreviousMode
	} else {
		m.SetViewMode(ViewModeHelp)
	}
}

// ToggleFocus toggles focus mode.
func (m *Model) ToggleFocus() {
	if m.ViewMode == ViewModeFocus {
		m.ViewMode = ViewModeDashboard
	} else {
		m.SetViewMode(ViewModeFocus)
	}
}

// ToggleZen toggles zen mode.
func (m *Model) ToggleZen() {
	if m.ViewMode == ViewModeZen {
		m.ViewMode = ViewModeDashboard
	} else {
		m.SetViewMode(ViewModeZen)
	}
}

// SetWorkflowState updates the workflow state.
func (m *Model) SetWorkflowState(state WorkflowState) {
	m.WorkflowState = state
}

// SetCurrentAgent updates the active agent.
func (m *Model) SetCurrentAgent(role, task string) {
	m.CurrentAgent = role
	m.CurrentTask = task

	if m.Dashboard != nil {
		m.Dashboard.AgentPanel.SetStatus(role, widgets.AgentWorking, task)
	}
	if m.Focus != nil {
		m.Focus.SetActiveAgent(role)
	}
	if m.Zen != nil {
		m.Zen.AgentRole = role
	}
}

// ClearCurrentAgent clears the active agent.
func (m *Model) ClearCurrentAgent(role string, success bool) {
	if m.Dashboard != nil {
		if success {
			m.Dashboard.AgentPanel.SetStatus(role, widgets.AgentComplete, "")
		} else {
			m.Dashboard.AgentPanel.SetStatus(role, widgets.AgentError, "")
		}
	}

	if m.CurrentAgent == role {
		m.CurrentAgent = ""
		m.CurrentTask = ""
		if m.Focus != nil {
			m.Focus.SetActiveAgent("")
		}
		if m.Zen != nil {
			m.Zen.AgentRole = ""
		}
	}
}

// AppendStreamingContent adds content to the streaming display.
func (m *Model) AppendStreamingContent(content string) {
	if m.Dashboard != nil {
		m.Dashboard.StreamingText.Append(content)
	}
	if m.Focus != nil {
		m.Focus.StreamingText.Append(content)
	}
	if m.Zen != nil {
		m.Zen.Content += content
		m.Zen.ShowCursor = true
	}
}

// SetCodeContent sets the code display content.
func (m *Model) SetCodeContent(content, language, filename string) {
	if m.Dashboard != nil {
		m.Dashboard.CodeBlock.SetContent(content, language, filename)
	}
	if m.Focus != nil {
		m.Focus.CodeBlock.SetContent(content, language, filename)
	}
}

// SetDiffContent sets the diff display content.
func (m *Model) SetDiffContent(content, filename string) {
	if m.Dashboard != nil {
		m.Dashboard.DiffBlock.SetContent(content, filename)
	}
	if m.Focus != nil {
		m.Focus.DiffBlock.SetContent(content, filename)
	}
}

// AddLogEntry adds an entry to the activity log.
func (m *Model) AddLogEntry(level widgets.LogLevel, agent, message string) {
	if m.Dashboard != nil {
		m.Dashboard.ActivityLog.Add(level, agent, message)
	}
	if m.Focus != nil {
		m.Focus.ActivityLog.Add(level, agent, message)
	}
}

// AddFile adds a file to the file tree.
func (m *Model) AddFile(path string, status widgets.FileStatus) {
	if m.Dashboard != nil {
		m.Dashboard.FileTree.AddFile(path, status)
	}
}

// RefreshFileTree reloads the file tree from disk or refreshes known entries.
func (m *Model) RefreshFileTree() {
	if m.Dashboard == nil || m.Dashboard.FileTree == nil {
		return
	}

	root := m.RepoRoot
	if root == "" {
		if cwd, err := os.Getwd(); err == nil {
			root = cwd
		}
	}
	if root == "" {
		m.ShowToast("Unable to determine workspace root", widgets.ToastLevelWarning)
		return
	}

	entries := m.Dashboard.FileTree.Snapshot()
	if len(entries) > 0 {
		m.Dashboard.FileTree.Clear()
		for _, entry := range entries {
			absPath := filepath.Join(root, filepath.FromSlash(entry.Path))
			info, err := os.Stat(absPath)
			if err != nil {
				m.Dashboard.FileTree.AddPath(entry.Path, widgets.FileStatusDeleted, entry.IsDir)
				continue
			}
			status := entry.Status
			m.Dashboard.FileTree.AddPath(entry.Path, status, info.IsDir())
		}
		m.ShowToast("File tree refreshed", widgets.ToastLevelInfo)
		return
	}

	// If no existing entries, build from disk with a reasonable cap.
	const maxFiles = 2000
	skips := map[string]struct{}{
		".git":           {},
		".cooperations":  {},
		".claude":        {},
		"node_modules":   {},
	}

	scanned := 0
	startRoot := root
	generated := filepath.Join(root, "generated")
	if info, err := os.Stat(generated); err == nil && info.IsDir() {
		startRoot = generated
	}

	m.Dashboard.FileTree.Clear()
	_ = filepath.WalkDir(startRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if path == startRoot {
			return nil
		}
		name := d.Name()
		if _, ok := skips[name]; ok && d.IsDir() {
			return filepath.SkipDir
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			return nil
		}
		m.Dashboard.FileTree.AddPath(rel, widgets.FileStatusNone, false)
		scanned++
		if scanned >= maxFiles {
			return fs.SkipAll
		}
		return nil
	})

	if scanned == 0 {
		m.ShowToast("No files found to display", widgets.ToastLevelInfo)
	} else {
		m.ShowToast(fmt.Sprintf("Loaded %d files", scanned), widgets.ToastLevelInfo)
	}
}

// ResolvePath returns an absolute path for a workspace-relative path.
func (m *Model) ResolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	root := m.RepoRoot
	if root == "" {
		if cwd, err := os.Getwd(); err == nil {
			root = cwd
		}
	}
	if root == "" {
		return path
	}
	return filepath.Join(root, filepath.FromSlash(path))
}

// UpdateProgress updates the progress bar.
func (m *Model) UpdateProgress(percent float64, label string) {
	if m.Dashboard != nil {
		m.Dashboard.ProgressBar.SetPercent(percent)
		m.Dashboard.ProgressBar.Label = label
		m.Dashboard.ShowProgress = true
	}
	if m.Focus != nil {
		m.Focus.ProgressBar.SetPercent(percent)
		m.Focus.ProgressBar.Label = label
	}
}

// UpdateMetrics updates the cost tracker.
func (m *Model) UpdateMetrics(input, output int) {
	if m.Dashboard != nil {
		m.Dashboard.CostTracker.Update(input, output)
	}
	if m.Focus != nil && m.Dashboard != nil {
		m.Focus.TokenCount = m.Dashboard.CostTracker.TotalTokens
	}
}

// UpdateMetricsSnapshot updates the metrics panel with a live snapshot.
func (m *Model) UpdateMetricsSnapshot(snapshot stream.MetricsSnapshot) {
	if m.Dashboard != nil {
		m.Dashboard.CostTracker.SetSnapshot(snapshot.PromptTokens, snapshot.CompletionTokens, snapshot.EstimatedCostUSD)
		m.Dashboard.Metrics.Clear()
		m.Dashboard.Metrics.AddMetric(widgets.NewMetricCard("Tokens", fmt.Sprintf("%d", snapshot.TotalTokens), ""))
		m.Dashboard.Metrics.AddMetric(widgets.NewMetricCard("Prompt", fmt.Sprintf("%d", snapshot.PromptTokens), ""))
		m.Dashboard.Metrics.AddMetric(widgets.NewMetricCard("Completion", fmt.Sprintf("%d", snapshot.CompletionTokens), ""))
		cost := widgets.NewMetricCard("Cost", fmt.Sprintf("$%.4f", snapshot.EstimatedCostUSD), "")
		cost.Color = styles.Current.Warning
		m.Dashboard.Metrics.AddMetric(cost)
		if snapshot.ElapsedTime > 0 {
			duration := widgets.NewMetricCard("Elapsed", snapshot.ElapsedTime.Round(time.Second).String(), "")
			duration.Color = styles.Current.Info
			m.Dashboard.Metrics.AddMetric(duration)
		}
		if snapshot.APICallsCount > 0 {
			m.Dashboard.Metrics.AddMetric(widgets.NewMetricCard("API Calls", fmt.Sprintf("%d", snapshot.APICallsCount), ""))
		}
		if snapshot.AgentCycles > 0 {
			m.Dashboard.Metrics.AddMetric(widgets.NewMetricCard("Cycles", fmt.Sprintf("%d", snapshot.AgentCycles), ""))
		}
	}
	if m.Focus != nil && snapshot.ElapsedTime > 0 {
		m.Focus.Duration = snapshot.ElapsedTime.Round(time.Second).String()
	}
	if m.Focus != nil && snapshot.TotalTokens > 0 {
		m.Focus.TokenCount = snapshot.TotalTokens
	}
}

// ShowToast displays a toast notification.
func (m *Model) ShowToast(message string, level widgets.ToastLevel) {
	if m.Dashboard != nil {
		switch level {
		case widgets.ToastLevelSuccess:
			m.Dashboard.ToastStack.PushSuccess(message)
		case widgets.ToastLevelWarning:
			m.Dashboard.ToastStack.PushWarning(message)
		case widgets.ToastLevelError:
			m.Dashboard.ToastStack.PushError(message)
		default:
			m.Dashboard.ToastStack.PushInfo(message)
		}
	}
}

// ShowDecision displays a decision dialog.
func (m *Model) ShowDecision(title, message string, options []widgets.DecisionOption) {
	dialog := widgets.NewDecisionDialog(title, message, m.Width/2)
	for _, opt := range options {
		dialog.AddOption(opt.Key, opt.Label, opt.Description, opt.Danger)
	}
	m.DecisionDialog = &dialog
	m.ConfirmDialog = nil
	m.InputDialog = nil
	m.ShowDialog = true
}

// ShowConfirm displays a confirmation dialog.
func (m *Model) ShowConfirm(title, message string, danger bool) {
	dialog := widgets.NewConfirmDialog(title, message, m.Width/2)
	dialog.Danger = danger
	m.ConfirmDialog = &dialog
	m.DecisionDialog = nil
	m.InputDialog = nil
	m.ShowDialog = true
}

// HideDialog hides any open dialog.
func (m *Model) HideDialog() {
	m.DecisionDialog = nil
	m.ConfirmDialog = nil
	m.InputDialog = nil
	m.InputMode = InputModeNone
	m.ShowDialog = false
}

// Tick advances all animations.
func (m *Model) Tick() {
	m.LastTick = time.Now()

	if m.Dashboard != nil {
		m.Dashboard.AgentPanel.TickAll()
		m.Dashboard.ToastStack.Cleanup()
	}
	if m.Focus != nil {
		m.Focus.Tick()
	}
	if m.Zen != nil {
		m.Zen.ToggleCursor()
	}
}

// Elapsed returns the elapsed time since start.
func (m *Model) Elapsed() time.Duration {
	return time.Since(m.StartTime)
}

// ElapsedString returns formatted elapsed time.
func (m *Model) ElapsedString() string {
	elapsed := m.Elapsed()
	if elapsed < time.Minute {
		return elapsed.Round(time.Second).String()
	}
	minutes := int(elapsed.Minutes())
	seconds := int(elapsed.Seconds()) % 60
	if minutes < 60 {
		return (time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second).String()
	}
	return elapsed.Round(time.Minute).String()
}

func (m *Model) ensureSession(task string) {
	if m.SessionManager == nil || m.ReplayActive {
		return
	}
	if m.SessionManager.Current != nil {
		return
	}
	if task == "" {
		task = "unspecified"
	}
	current := m.SessionManager.NewSession(task)
	m.SessionID = current.ID
	m.SessionName = current.Name
}

func (m *Model) recordStreamEvent(event interface{}) {
	if m.SessionManager == nil || m.ReplayActive {
		return
	}
	m.SessionManager.RecordStreamEvent(event)
}

func (m *Model) searchTargetForView() SearchTarget {
	if m.ViewMode == ViewModeFocus && m.Focus != nil {
		switch m.Focus.Mode {
		case 0:
			return SearchTargetStreaming
		case 1:
			return SearchTargetCode
		case 2:
			return SearchTargetDiff
		default:
			return SearchTargetNone
		}
	}
	if m.Dashboard == nil {
		return SearchTargetNone
	}
	switch m.Dashboard.CenterMode {
	case 0:
		return SearchTargetStreaming
	case 1:
		return SearchTargetCode
	case 2:
		return SearchTargetDiff
	default:
		return SearchTargetNone
	}
}

func (m *Model) searchContent(target SearchTarget) string {
	if m.Dashboard == nil {
		return ""
	}
	switch target {
	case SearchTargetStreaming:
		return m.Dashboard.StreamingText.Content
	case SearchTargetCode:
		return m.Dashboard.CodeBlock.Content
	case SearchTargetDiff:
		return m.Dashboard.DiffBlock.Content
	default:
		return ""
	}
}

func (m *Model) applySearchHighlights(target SearchTarget, results []int) {
	if m.Dashboard == nil {
		return
	}
	switch target {
	case SearchTargetStreaming:
		m.Dashboard.StreamingText.SetHighlights(results)
		if m.Focus != nil {
			m.Focus.StreamingText.SetHighlights(results)
		}
	case SearchTargetCode:
		m.Dashboard.CodeBlock.ClearHighlights()
		if m.Focus != nil {
			m.Focus.CodeBlock.ClearHighlights()
		}
		for _, line := range results {
			m.Dashboard.CodeBlock.AddHighlight(line + 1)
			if m.Focus != nil {
				m.Focus.CodeBlock.AddHighlight(line + 1)
			}
		}
	case SearchTargetDiff:
		m.Dashboard.DiffBlock.SetHighlights(results)
		if m.Focus != nil {
			m.Focus.DiffBlock.SetHighlights(results)
		}
	}
}

func (m *Model) clearSearchHighlights() {
	if m.Dashboard == nil {
		return
	}
	m.Dashboard.StreamingText.ClearHighlights()
	m.Dashboard.DiffBlock.ClearHighlights()
	m.Dashboard.CodeBlock.ClearHighlights()
	if m.Focus != nil {
		m.Focus.StreamingText.ClearHighlights()
		m.Focus.DiffBlock.ClearHighlights()
		m.Focus.CodeBlock.ClearHighlights()
	}
}

func (m *Model) scrollToSearchResult(target SearchTarget, line int) {
	if m.ViewMode == ViewModeFocus && m.Focus != nil {
		switch target {
		case SearchTargetStreaming:
			m.Focus.StreamingText.ScrollToLine(line)
		case SearchTargetCode:
			m.Focus.CodeBlock.ScrollToLine(line)
		case SearchTargetDiff:
			m.Focus.DiffBlock.ScrollToLine(line)
		}
		return
	}
	if m.Dashboard == nil {
		return
	}
	switch target {
	case SearchTargetStreaming:
		m.Dashboard.StreamingText.ScrollToLine(line)
	case SearchTargetCode:
		m.Dashboard.CodeBlock.ScrollToLine(line)
	case SearchTargetDiff:
		m.Dashboard.DiffBlock.ScrollToLine(line)
	}
}

func (m *Model) runSearch(query string) bool {
	query = strings.TrimSpace(query)
	if query == "" {
		m.clearSearchHighlights()
		m.SearchQuery = ""
		m.SearchResults = nil
		m.SearchIndex = 0
		m.SearchTarget = SearchTargetNone
		return false
	}
	target := m.searchTargetForView()
	if target == SearchTargetNone {
		return false
	}
	content := m.searchContent(target)
	lines := strings.Split(content, "\n")
	results := make([]int, 0)
	needle := strings.ToLower(query)
	for i, line := range lines {
		if strings.Contains(strings.ToLower(line), needle) {
			results = append(results, i)
		}
	}
	m.SearchQuery = query
	m.SearchResults = results
	m.SearchIndex = 0
	m.SearchTarget = target
	m.applySearchHighlights(target, results)
	if len(results) > 0 {
		m.scrollToSearchResult(target, results[0])
	}
	return len(results) > 0
}

func (m *Model) jumpSearch(delta int) {
	if m.SearchQuery == "" {
		m.ShowToast("No active search", widgets.ToastLevelInfo)
		return
	}
	currentTarget := m.searchTargetForView()
	if currentTarget == SearchTargetNone {
		m.ShowToast("Search not available in this view", widgets.ToastLevelWarning)
		return
	}
	if m.SearchTarget != currentTarget {
		m.runSearch(m.SearchQuery)
	}
	if len(m.SearchResults) == 0 {
		m.ShowToast("No matches", widgets.ToastLevelWarning)
		return
	}
	m.SearchIndex = (m.SearchIndex + delta) % len(m.SearchResults)
	if m.SearchIndex < 0 {
		m.SearchIndex = len(m.SearchResults) - 1
	}
	line := m.SearchResults[m.SearchIndex]
	m.scrollToSearchResult(m.SearchTarget, line)
	m.ShowToast(fmt.Sprintf("Match %d/%d", m.SearchIndex+1, len(m.SearchResults)), widgets.ToastLevelInfo)
}

func (m *Model) resetForReplay() {
	m.CurrentAgent = ""
	m.CurrentTask = ""
	m.TotalSteps = 0
	m.CompletedSteps = 0
	m.WorkflowStepIndex = nil
	m.clearSearchHighlights()
	m.SearchQuery = ""
	m.SearchResults = nil
	m.SearchIndex = 0
	m.SearchTarget = SearchTargetNone
	m.SearchMode = false

	if m.Dashboard != nil {
		m.Dashboard.StreamingText.Clear()
		m.Dashboard.CodeBlock.SetContent("", "", "")
		m.Dashboard.DiffBlock.SetContent("", "")
		m.Dashboard.ActivityLog.Clear()
		m.Dashboard.FileTree.Clear()
		m.Dashboard.Metrics.Clear()
		m.Dashboard.WorkflowSteps.Clear()
		m.Dashboard.ProgressBar.SetPercent(0)
		m.Dashboard.ProgressBar.Label = ""
		m.Dashboard.ShowProgress = false
	}

	if m.Focus != nil {
		m.Focus.StreamingText.Clear()
		m.Focus.CodeBlock.SetContent("", "", "")
		m.Focus.DiffBlock.SetContent("", "")
		m.Focus.ActivityLog.Clear()
		m.Focus.ProgressBar.SetPercent(0)
		m.Focus.ProgressBar.Label = ""
		m.Focus.TokenCount = 0
		m.Focus.Duration = ""
	}

	if m.Zen != nil {
		m.Zen.Content = ""
		m.Zen.AgentRole = ""
	}
}
