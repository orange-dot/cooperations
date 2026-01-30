// internal/gui/app.go
package gui

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"sync"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"cooperations/internal/gui/stream"
	"cooperations/internal/gui/widgets"
	"cooperations/internal/orchestrator"
)

// App is the main GUI application for the Cooperations workflow.
type App struct {
	window *app.Window
	theme  *material.Theme
	state  *AppState

	mu     sync.Mutex
	stream *stream.WorkflowStream

	// Panel widgets
	sidebar     *widgets.SidebarPanel
	mainPanel   *widgets.MainPanel
	bottomPanel *widgets.BottomPanel

	// Current code display
	currentCode     string
	currentCodeLang string
}

// NewApp creates a new App instance with a window and theme.
func NewApp() *App {
	w := new(app.Window)
	w.Option(
		app.Title("COOPERATIONS"),
		app.Size(unit.Dp(1400), unit.Dp(900)),
	)

	th := material.NewTheme()

	return &App{
		window:      w,
		theme:       th,
		state:       NewAppState(),
		sidebar:     widgets.NewSidebarPanel(),
		mainPanel:   widgets.NewMainPanel(),
		bottomPanel: widgets.NewBottomPanel(),
	}
}

// Run starts the GUI application with the given task description.
func (a *App) Run(task string) error {
	return a.RunWithDemo(task, true) // Default to demo mode
}

// RunWithDemo starts the GUI with optional demo/stub progress.
func (a *App) RunWithDemo(task string, demo bool) error {
	a.state.SetTaskDescription(task)
	a.state.SetTaskInProgress(true)

	// Initialize with default workflow steps
	a.state.SetWorkflowSteps([]WorkflowStepState{
		{ID: "analyze", Label: "Analyze Task", Status: "pending"},
		{ID: "architect", Label: "Architecture", Status: "pending"},
		{ID: "implement", Label: "Implementation", Status: "pending"},
		{ID: "review", Label: "Review", Status: "pending"},
		{ID: "human", Label: "Human Review", Status: "pending"},
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ws := stream.NewWorkflowStream()
	a.mu.Lock()
	a.stream = ws
	a.mu.Unlock()

	defer ws.Close()

	// Wire up bottom panel callbacks
	a.bottomPanel.OnApprove = func() {
		a.handleDecision(stream.DecisionActionApprove, "")
	}
	a.bottomPanel.OnReject = func() {
		a.handleDecision(stream.DecisionActionReject, "")
	}
	a.bottomPanel.OnEdit = func(comment string) {
		a.handleDecision(stream.DecisionActionEdit, comment)
	}

	// Start stream handler goroutine.
	go a.handleStream(ctx, ws)

	// Start workflow execution
	if demo {
		go a.runDemoProgress(ctx, ws, task)
	} else {
		go a.runRealWorkflow(ctx, ws, task)
	}

	// Run the event loop (blocks until window closes).
	err := a.eventLoop()

	// Ensure goroutines terminate promptly.
	cancel()

	return err
}

// Stream returns the current WorkflowStream for external producers to send events.
func (a *App) Stream() *stream.WorkflowStream {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.stream
}

func (a *App) handleDecision(action stream.DecisionAction, comment string) {
	a.mu.Lock()
	ws := a.stream
	a.mu.Unlock()

	if ws == nil {
		return
	}

	decision := stream.HumanDecision{
		RequestID: "", // Would be set from current decision request
		Action:    action,
		Comment:   comment,
	}

	select {
	case ws.Response <- decision:
		a.state.SetWaitingForInput(false)
		a.state.AddActivity(fmt.Sprintf("Decision: %s", action))
	default:
		// Channel full or closed
	}
}

func (a *App) eventLoop() error {
	var ops op.Ops

	for {
		switch e := a.window.Event().(type) {
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			a.layout(gtx)
			e.Frame(gtx.Ops)

		case app.DestroyEvent:
			return e.Err
		}
	}
}

func (a *App) handleStream(ctx context.Context, ws *stream.WorkflowStream) {
	for {
		select {
		case <-ctx.Done():
			return

		case prog, ok := <-ws.Progress:
			if !ok {
				return
			}
			a.state.SetStatusLine(fmt.Sprintf("%s: %.0f%%", prog.Stage, prog.Percent))
			a.state.AddActivity(prog.Message)
			a.window.Invalidate()

		case code, ok := <-ws.Code:
			if !ok {
				return
			}
			a.mu.Lock()
			a.currentCode = code.Content
			a.currentCodeLang = code.Language
			a.mu.Unlock()
			a.state.AddActivity(fmt.Sprintf("Code update: %s", code.Path))
			a.window.Invalidate()

		case handoff, ok := <-ws.Handoffs:
			if !ok {
				return
			}
			a.state.AddHandoff(handoff.From, handoff.To)
			a.state.AddActivity(fmt.Sprintf("Handoff: %s → %s", handoff.From, handoff.To))
			a.window.Invalidate()

		case tokens, ok := <-ws.Tokens:
			if !ok {
				return
			}
			a.state.AddActivity(fmt.Sprintf("Tokens: %d total", tokens.TotalTokens))
			a.window.Invalidate()

		case decision, ok := <-ws.Decision:
			if !ok {
				return
			}
			a.state.SetWaitingForInput(true)
			a.state.SetStatusLine(fmt.Sprintf("Decision needed: %s", decision.Title))
			a.bottomPanel.Title = decision.Title
			a.bottomPanel.Prompt = decision.Prompt
			a.bottomPanel.Options = decision.Options
			a.window.Invalidate()

		case err, ok := <-ws.Error:
			if !ok {
				return
			}
			a.state.SetError(err.Error())
			a.state.AddActivity(fmt.Sprintf("Error: %s", err.Error()))
			a.window.Invalidate()

		case <-ws.Done:
			a.state.SetTaskInProgress(false)
			a.state.SetCompleted(true)
			a.state.SetCompletionMessage("Workflow completed")
			a.state.AddActivity("Workflow completed")
			a.window.Invalidate()
			return
		}
	}
}

func (a *App) layout(gtx layout.Context) layout.Dimensions {
	snap := a.state.Snapshot()

	// Fill background
	paint.Fill(gtx.Ops, DefaultTheme.Background)

	// Update panel data from state
	a.updatePanelData(&snap)

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Header bar
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.layoutHeader(gtx, &snap)
		}),

		// Main content area (sidebar + main panel)
		layout.Flexed(1.0, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				// Sidebar (20% width)
				layout.Flexed(0.2, func(gtx layout.Context) layout.Dimensions {
					return a.sidebar.Layout(gtx, a.theme)
				}),

				// Main panel (80% width)
				layout.Flexed(0.8, func(gtx layout.Context) layout.Dimensions {
					return a.mainPanel.Layout(gtx, a.theme)
				}),
			)
		}),

		// Bottom panel (only when waiting for input)
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			a.bottomPanel.Visible = snap.WaitingForInput
			return a.bottomPanel.Layout(gtx, a.theme)
		}),
	)
}

func (a *App) layoutHeader(gtx layout.Context, snap *StateSnapshot) layout.Dimensions {
	headerHeight := gtx.Dp(unit.Dp(56))
	headerBg := DefaultTheme.PanelBg
	borderColor := DefaultTheme.Border

	// Draw header background
	size := image.Pt(gtx.Constraints.Max.X, headerHeight)
	defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, headerBg)

	// Draw bottom border
	borderRect := image.Rect(0, headerHeight-1, size.X, headerHeight)
	st := clip.Rect(borderRect).Push(gtx.Ops)
	paint.Fill(gtx.Ops, borderColor)
	st.Pop()

	// Header content
	inset := layout.Inset{
		Left:  unit.Dp(20),
		Right: unit.Dp(20),
		Top:   unit.Dp(12),
	}

	inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
			// Title
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.H6(a.theme, "COOPERATIONS")
				lbl.Color = DefaultTheme.Cyan
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(24)}.Layout),

			// Task description
			layout.Flexed(1.0, func(gtx layout.Context) layout.Dimensions {
				title := snap.TaskDescription
				if title == "" {
					title = "No task"
				}
				lbl := material.Body1(a.theme, title)
				lbl.Color = DefaultTheme.TextPrimary
				return lbl.Layout(gtx)
			}),

			// Status indicator
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.layoutStatusBadge(gtx, snap)
			}),
		)
	})

	return layout.Dimensions{Size: size}
}

func (a *App) layoutStatusBadge(gtx layout.Context, snap *StateSnapshot) layout.Dimensions {
	var statusText string
	var statusColor color.NRGBA

	if snap.ErrorMessage != "" {
		statusText = "ERROR"
		statusColor = DefaultTheme.Error
	} else if snap.Completed {
		statusText = "COMPLETE"
		statusColor = DefaultTheme.Success
	} else if snap.WaitingForInput {
		statusText = "WAITING"
		statusColor = DefaultTheme.Warning
	} else if snap.TaskInProgress {
		statusText = "RUNNING"
		statusColor = DefaultTheme.Cyan
	} else {
		statusText = "READY"
		statusColor = DefaultTheme.TextSecondary
	}

	// Badge background
	badgePadding := unit.Dp(8)
	return layout.Inset{
		Left:   badgePadding,
		Right:  badgePadding,
		Top:    unit.Dp(4),
		Bottom: unit.Dp(4),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		lbl := material.Caption(a.theme, statusText)
		lbl.Color = statusColor
		return lbl.Layout(gtx)
	})
}

func (a *App) updatePanelData(snap *StateSnapshot) {
	// Update sidebar with workflow steps
	steps := make([]widgets.WorkflowStep, len(snap.WorkflowSteps))
	for i, s := range snap.WorkflowSteps {
		steps[i] = widgets.WorkflowStep{
			ID:       s.ID,
			Label:    s.Label,
			Status:   s.Status,
			Progress: s.Progress,
			Subtext:  s.Subtext,
		}
	}
	a.sidebar.SetSteps(steps)
	a.sidebar.SetCurrentStep(snap.CurrentStep)

	// Update sidebar with handoff history
	handoffs := make([]widgets.HandoffEntry, len(snap.HandoffHistory))
	for i, h := range snap.HandoffHistory {
		handoffs[i] = widgets.HandoffEntry{
			FromRole:  h.FromRole,
			ToRole:    h.ToRole,
			Timestamp: h.Timestamp,
		}
	}
	a.sidebar.SetHandoffs(handoffs)

	// Update main panel
	a.mainPanel.SetActivityLog(snap.ActivityLog)

	// Update code display
	a.mu.Lock()
	code := a.currentCode
	lang := a.currentCodeLang
	a.mu.Unlock()
	a.mainPanel.SetCode(code, lang)
}

// runRealWorkflow executes the actual orchestrator workflow with stream events.
func (a *App) runRealWorkflow(ctx context.Context, ws *stream.WorkflowStream, task string) {
	// Create orchestrator with stream
	config := orchestrator.DefaultWorkflowConfig()
	orch, err := orchestrator.NewWithStream(config, ws)
	if err != nil {
		select {
		case ws.Error <- fmt.Errorf("failed to create orchestrator: %w", err):
		case <-ctx.Done():
		}
		return
	}

	// Run the workflow
	result, err := orch.Run(ctx, task)
	if err != nil {
		select {
		case ws.Error <- err:
		case <-ctx.Done():
		}
		return
	}

	// Handle result
	if !result.Success {
		select {
		case ws.Error <- fmt.Errorf("workflow failed: %s", result.Error):
		case <-ctx.Done():
		}
	}
}

// runDemoProgress simulates workflow execution with stub progress events.
func (a *App) runDemoProgress(ctx context.Context, ws *stream.WorkflowStream, task string) {
	steps := []struct {
		role     string
		label    string
		duration time.Duration
	}{
		{"navigator", "Analyze Task", 800 * time.Millisecond},
		{"architect", "Architecture", 1200 * time.Millisecond},
		{"implementer", "Implementation", 2000 * time.Millisecond},
		{"reviewer", "Review", 1000 * time.Millisecond},
		{"human", "Human Review", 500 * time.Millisecond},
	}

	send := func(ch interface{}, val interface{}) bool {
		select {
		case <-ctx.Done():
			return false
		default:
		}

		switch c := ch.(type) {
		case chan stream.ProgressUpdate:
			select {
			case c <- val.(stream.ProgressUpdate):
			case <-ctx.Done():
				return false
			}
		case chan stream.HandoffEvent:
			select {
			case c <- val.(stream.HandoffEvent):
			case <-ctx.Done():
				return false
			}
		case chan stream.CodeUpdate:
			select {
			case c <- val.(stream.CodeUpdate):
			case <-ctx.Done():
				return false
			}
		case chan stream.TokenUpdate:
			select {
			case c <- val.(stream.TokenUpdate):
			case <-ctx.Done():
				return false
			}
		case chan stream.DecisionRequest:
			select {
			case c <- val.(stream.DecisionRequest):
			case <-ctx.Done():
				return false
			}
		}
		return true
	}

	// Initial delay
	select {
	case <-time.After(300 * time.Millisecond):
	case <-ctx.Done():
		return
	}

	if !send(ws.Progress, stream.ProgressUpdate{
		Stage:   "Starting",
		Percent: 0,
		Message: fmt.Sprintf("Starting workflow for: %s", task),
	}) {
		return
	}

	prevRole := "user"
	totalTokens := 0

	for i, step := range steps {
		// Update workflow step status
		a.state.UpdateStepStatus(i, "inprogress")
		a.state.SetCurrentStep(i)

		// Progress start
		if !send(ws.Progress, stream.ProgressUpdate{
			Stage:   step.label,
			Percent: float64(i) / float64(len(steps)) * 100,
			Message: fmt.Sprintf("%s is working...", step.role),
		}) {
			return
		}

		// Handoff event
		if !send(ws.Handoffs, stream.HandoffEvent{
			From:      prevRole,
			To:        step.role,
			Reason:    fmt.Sprintf("Starting %s phase", step.label),
			Timestamp: time.Now(),
		}) {
			return
		}

		// Simulate work with progress updates
		subSteps := 5
		for j := 1; j <= subSteps; j++ {
			select {
			case <-time.After(step.duration / time.Duration(subSteps)):
			case <-ctx.Done():
				return
			}

			progress := (float64(i) + float64(j)/float64(subSteps)) / float64(len(steps)) * 100
			a.state.UpdateStepProgress(i, float32(j)/float32(subSteps))

			if !send(ws.Progress, stream.ProgressUpdate{
				Stage:   step.label,
				Percent: progress,
				Message: fmt.Sprintf("%s: processing step %d/%d", step.role, j, subSteps),
			}) {
				return
			}
		}

		// Token update
		tokens := 500 + (i * 200)
		totalTokens += tokens
		if !send(ws.Tokens, stream.TokenUpdate{
			PromptTokens:     tokens / 2,
			CompletionTokens: tokens / 2,
			TotalTokens:      totalTokens,
		}) {
			return
		}

		// Code update for implementer
		if step.role == "implementer" {
			if !send(ws.Code, stream.CodeUpdate{
				Path:     "internal/example/generated.go",
				Language: "go",
				Content: `package example

// Generated code from workflow
func ProcessTask(input string) (string, error) {
    // Implementation goes here
    result := fmt.Sprintf("Processed: %s", input)
    return result, nil
}
`,
			}) {
				return
			}
		}

		// Mark step complete
		a.state.UpdateStepStatus(i, "complete")
		a.state.UpdateStepProgress(i, 1.0)
		prevRole = step.role
	}

	// Human decision prompt
	if !send(ws.Decision, stream.DecisionRequest{
		ID:      "final-review",
		Title:   "Approve Workflow Result",
		Prompt:  "The workflow has completed. Please review the generated code and approve or request changes.",
		Options: []string{"Approve", "Request Changes", "Reject"},
	}) {
		return
	}

	// Wait for human decision
	select {
	case decision := <-ws.Response:
		if !send(ws.Progress, stream.ProgressUpdate{
			Stage:   "Complete",
			Percent: 100,
			Message: fmt.Sprintf("Human decision: %s - %s", decision.Action, decision.Comment),
		}) {
			return
		}
	case <-ctx.Done():
		return
	case <-time.After(30 * time.Second):
		// Timeout - auto-complete
		if !send(ws.Progress, stream.ProgressUpdate{
			Stage:   "Complete",
			Percent: 100,
			Message: "Workflow completed (no human response)",
		}) {
			return
		}
	}

	// Signal completion
	select {
	case ws.Done <- struct{}{}:
	case <-ctx.Done():
	}
}
