# Implementation Plan: Futuristic Gio GUI

**Date:** 2026-01-30
**Status:** Ready for Implementation
**Estimated Files:** ~25 new files

## Summary

Implement a futuristic GPU-accelerated GUI using Gio framework with human-in-the-loop mob programming support. Single binary, pure Go, widescreen optimized.

---

## Phase 1: Core Infrastructure

### Step 1.1: Add Dependencies

```bash
go get gioui.org@latest
go get gioui.org/x@latest
go get github.com/alecthomas/chroma/v2@latest
```

### Step 1.2: Create Theme

**File:** `internal/gui/theme.go`

```go
package gui

import "image/color"

type Theme struct {
    Background     color.NRGBA // #0a0e17
    PanelBg        color.NRGBA // #0d1520
    Border         color.NRGBA // #1a3a4a
    BorderActive   color.NRGBA // #00ffff
    TextPrimary    color.NRGBA // #ffffff
    TextSecondary  color.NRGBA // #8899aa
    Success        color.NRGBA // #00ff88
    Error          color.NRGBA // #ff4466
    Warning        color.NRGBA // #ffaa00
    Accent         color.NRGBA // #ff00ff
    Cyan           color.NRGBA // #00ffff
}

var DefaultTheme = &Theme{...}
```

### Step 1.3: Embed Fonts

**File:** `internal/gui/fonts/fonts.go`

```go
package fonts

import (
    _ "embed"
    "gioui.org/font"
    "gioui.org/font/opentype"
)

//go:embed JetBrainsMono-Regular.ttf
var jetBrainsMonoRegular []byte

//go:embed JetBrainsMono-Bold.ttf
var jetBrainsMonoBold []byte

func LoadFonts() []font.FontFace
```

### Step 1.4: Application State

**File:** `internal/gui/state.go`

```go
type AppState struct {
    sync.RWMutex

    Task             string
    WorkflowSteps    []WorkflowStepState
    CurrentRole      types.Role
    Artifacts        types.HArtifacts
    Handoffs         []HandoffEntry
    TokensUsed       int
    ElapsedTime      time.Duration

    AwaitingDecision bool
    DecisionRequest  *DecisionRequest

    ShowDesignDoc    bool
    ShowHistory      bool
    EditMode         bool

    Error            error
    Complete         bool
}
```

### Step 1.5: App Skeleton

**File:** `internal/gui/app.go`

```go
type App struct {
    window    *app.Window
    theme     *Theme
    state     *AppState
    stream    *WorkflowStream
    layout    *Layout
}

func NewApp() *App
func (a *App) Run(task string) error
func (a *App) eventLoop() error
func (a *App) handleStream()
```

---

## Phase 2: Custom Widgets

### Step 2.1: NeonProgress

**File:** `internal/gui/widgets/neon_progress.go`

```go
type NeonProgress struct {
    Progress  float32     // 0.0 to 1.0
    Color     color.NRGBA
    Glow      bool
    animating bool
    startTime time.Time
}

func (p *NeonProgress) Layout(gtx layout.Context) layout.Dimensions
```

Features:
- Gradient fill (cyan → magenta)
- Glow effect with alpha layers
- Smooth animation

### Step 2.2: WorkflowStep

**File:** `internal/gui/widgets/workflow_step.go`

```go
type StepStatus int
const (
    StepPending StepStatus = iota
    StepInProgress
    StepComplete
    StepWaiting
)

type WorkflowStep struct {
    Role     types.Role
    Status   StepStatus
    Progress float32
    Label    string
    Tokens   int
}
```

### Step 2.3: NeonButton

**File:** `internal/gui/widgets/neon_button.go`

```go
type NeonButton struct {
    Text     string
    Icon     string
    Color    color.NRGBA
    OnClick  func()

    clickable widget.Clickable
    hovered   bool
    pressed   bool
}
```

Features:
- Hover detection
- Scale animation (1.0 → 1.05)
- Glow on hover
- Press feedback

### Step 2.4: CodePanel

**File:** `internal/gui/widgets/code_panel.go`

```go
type CodePanel struct {
    Code     string
    Language string

    lexer    chroma.Lexer
    style    *chroma.Style
    tokens   []chroma.Token
    list     widget.List
}

func (c *CodePanel) SetCode(code, lang string)
func (c *CodePanel) Layout(gtx layout.Context) layout.Dimensions
```

Features:
- Chroma syntax highlighting
- Custom dark theme matching UI
- Scrollable viewport
- Line numbers

### Step 2.5: DecisionPanel

**File:** `internal/gui/widgets/decision_panel.go`

```go
type DecisionPanel struct {
    OnApprove func()
    OnReject  func(feedback string)
    OnEdit    func()

    approveBtn *NeonButton
    rejectBtn  *NeonButton
    editBtn    *NeonButton
    feedback   widget.Editor
    visible    bool
}
```

### Step 2.6: HandoffLog

**File:** `internal/gui/widgets/handoff_log.go`

```go
type HandoffEntry struct {
    Timestamp time.Time
    From      types.Role
    To        types.Role
}

type HandoffLog struct {
    Entries []HandoffEntry
    list    widget.List
}
```

---

## Phase 3: Panel Layout

### Step 3.1: Main Layout

**File:** `internal/gui/panels/layout.go`

```go
type Layout struct {
    Sidebar     *Sidebar
    MainPanel   *MainPanel
    BottomPanel *BottomPanel
    theme       *Theme
}

func (l *Layout) Layout(gtx layout.Context) layout.Dimensions
```

Layout structure:
```
┌────────────────────────────────────────────────────┐
│                    Header                          │
├──────────┬─────────────────────────────────────────┤
│          │                                         │
│ Sidebar  │           MainPanel                     │
│  250px   │           (flexible)                    │
│          │                                         │
│          ├─────────────────────────────────────────┤
│          │           BottomPanel                   │
│          │           (150px when visible)          │
└──────────┴─────────────────────────────────────────┘
```

### Step 3.2: Sidebar

**File:** `internal/gui/panels/sidebar.go`

```go
type Sidebar struct {
    WorkflowProgress []WorkflowStep
    TaskContext      TaskContextPanel
    HandoffLog       *HandoffLog
}
```

Contains:
- MOB PROGRESS panel
- TASK CONTEXT panel
- HANDOFF LOG panel

### Step 3.3: MainPanel

**File:** `internal/gui/panels/main_panel.go`

```go
type MainPanel struct {
    CodePanel    *CodePanel
    DesignPanel  *CodePanel
    ShowDesign   bool
}
```

### Step 3.4: BottomPanel

**File:** `internal/gui/panels/bottom_panel.go`

```go
type BottomPanel struct {
    Decision *DecisionPanel
    visible  bool
    height   unit.Dp
}
```

---

## Phase 4: HumanAgent

### Step 4.1: Create HumanAgent

**File:** `internal/agents/human.go`

```go
type HumanAgent struct {
    inputChan  chan HumanDecision
    outputChan chan types.Handoff
}

type HumanDecision struct {
    Action     DecisionAction
    Feedback   string
    EditedCode string
}

type DecisionAction int
const (
    DecisionApprove DecisionAction = iota
    DecisionReject
    DecisionEdit
)

func NewHumanAgent(input chan HumanDecision, output chan types.Handoff) *HumanAgent

func (h *HumanAgent) Role() types.Role
func (h *HumanAgent) Execute(ctx context.Context, handoff types.Handoff) (types.AgentResponse, error)
```

### Step 4.2: Modify Types

**File:** `internal/types/types.go`

Add:
```go
const RoleHuman Role = "human"
```

### Step 4.3: Modify Parser

**File:** `internal/agents/parse.go`

Change:
```go
case "user":
    role := types.RoleHuman
    return &role  // Was: return nil
```

### Step 4.4: Modify Orchestrator

**File:** `internal/orchestrator/orchestrator.go`

- Add `HumanAgent` field
- Add `WithHumanAgent()` option
- Register in agents map

---

## Phase 5: Streaming System

### Step 5.1: Event Types

**File:** `internal/gui/stream/events.go`

```go
type ProgressUpdate struct {
    Role     types.Role
    Status   StepStatus
    Progress float32
    Label    string
    Tokens   int
}

type CodeUpdate struct {
    Content  string
    Language string
    Append   bool
}

type HandoffEvent struct {
    From      types.Role
    To        types.Role
    Timestamp time.Time
}

type TokenUpdate struct {
    Total int
    Delta int
}

type DecisionRequest struct {
    Handoff   types.Handoff
    Artifacts types.HArtifacts
}
```

### Step 5.2: WorkflowStream

**File:** `internal/gui/stream/stream.go`

```go
type WorkflowStream struct {
    Progress  chan ProgressUpdate
    Code      chan CodeUpdate
    Handoffs  chan HandoffEvent
    Tokens    chan TokenUpdate
    Decision  chan DecisionRequest
    Response  chan HumanDecision
    Done      chan struct{}
    Error     chan error
}

func NewWorkflowStream() *WorkflowStream
```

### Step 5.3: Streaming Workflow

**File:** `internal/orchestrator/workflow_stream.go`

```go
func (o *Orchestrator) RunWithStream(ctx context.Context, task string, stream *WorkflowStream) error
```

Sends updates to stream channels during execution.

---

## Phase 6: Integration

### Step 6.1: Complete App

**File:** `internal/gui/app.go` (complete)

```go
func (a *App) Run(task string) error {
    // Create channels
    humanInput := make(chan HumanDecision, 1)
    humanOutput := make(chan types.Handoff, 1)

    // Create HumanAgent
    humanAgent := agents.NewHumanAgent(humanInput, humanOutput)

    // Create orchestrator with HumanAgent
    orch := orchestrator.New(
        orchestrator.WithHumanAgent(humanAgent),
    )

    // Create stream
    a.stream = stream.NewWorkflowStream()

    // Start workflow goroutine
    go func() {
        err := orch.RunWithStream(context.Background(), task, a.stream)
        if err != nil {
            a.stream.Error <- err
        }
        close(a.stream.Done)
    }()

    // Start stream handler
    go a.handleStream()

    // Run event loop
    return a.eventLoop()
}
```

### Step 6.2: Keyboard Handlers

**File:** `internal/gui/handlers/keyboard.go`

```go
func (a *App) handleKeyboard(gtx layout.Context) {
    for {
        event, ok := gtx.Event(key.Filter{...})
        if !ok {
            break
        }
        if e, ok := event.(key.Event); ok && e.State == key.Press {
            switch e.Name {
            case "A": a.approve()
            case "R": a.reject()
            case "E": a.edit()
            case "D": a.toggleDesign()
            case "H": a.toggleHistory()
            case key.NameEscape, "Q": a.quit()
            }
        }
    }
}
```

### Step 6.3: Add CLI Command

**File:** `cmd/coop/main.go`

Add:
```go
var guiCmd = &cobra.Command{
    Use:   "gui <task>",
    Short: "Launch futuristic GUI for mob programming",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        app := gui.NewApp()
        return app.Run(args[0])
    },
}

func init() {
    rootCmd.AddCommand(guiCmd)
}
```

---

## Phase 7: Testing

### Unit Tests

| File | Tests |
|------|-------|
| `widgets/neon_progress_test.go` | Progress rendering, animation |
| `widgets/code_panel_test.go` | Syntax highlighting, scrolling |
| `agents/human_test.go` | Decision handling, channel communication |
| `stream/stream_test.go` | Channel operations, no deadlocks |

### Integration Tests

| File | Tests |
|------|-------|
| `gui/app_test.go` | Full workflow with mock stream |
| `orchestrator/workflow_stream_test.go` | Streaming with mock agents |

### Manual Test Checklist

- [ ] Theme renders correctly (dark background, neon accents)
- [ ] Animations smooth at 60 FPS
- [ ] Workflow progress updates in real-time
- [ ] Code panel syntax highlighting works
- [ ] Approve flow completes workflow
- [ ] Reject flow returns to Implementer
- [ ] Keyboard shortcuts work (A, R, E, D, H, Q)
- [ ] Window resize handles gracefully
- [ ] Quit prompts to save state

---

## File Summary

### New Files (25)

```
internal/gui/
├── app.go
├── state.go
├── theme.go
├── fonts/
│   ├── fonts.go
│   ├── JetBrainsMono-Regular.ttf
│   └── JetBrainsMono-Bold.ttf
├── widgets/
│   ├── neon_progress.go
│   ├── workflow_step.go
│   ├── neon_button.go
│   ├── code_panel.go
│   ├── decision_panel.go
│   ├── handoff_log.go
│   └── text_editor.go
├── panels/
│   ├── layout.go
│   ├── sidebar.go
│   ├── main_panel.go
│   └── bottom_panel.go
├── stream/
│   ├── events.go
│   └── stream.go
└── handlers/
    └── keyboard.go

internal/agents/
└── human.go

internal/orchestrator/
└── workflow_stream.go
```

### Modified Files (4)

```
internal/types/types.go          # Add RoleHuman
internal/agents/parse.go         # Return RoleHuman for "user"
internal/orchestrator/orchestrator.go  # Add HumanAgent support
cmd/coop/main.go                 # Add gui command
```

---

## Implementation Order

| Priority | Phase | Description | Dependencies |
|----------|-------|-------------|--------------|
| 1 | 1.1 | Add Go dependencies | None |
| 2 | 1.2-1.3 | Theme + Fonts | Dependencies |
| 3 | 4.1-4.4 | HumanAgent | Theme |
| 4 | 5.1-5.3 | Streaming system | HumanAgent |
| 5 | 1.4-1.5 | State + App skeleton | Streaming |
| 6 | 2.1-2.6 | Custom widgets | Theme |
| 7 | 3.1-3.4 | Panel layout | Widgets |
| 8 | 6.1-6.3 | Integration | All above |
| 9 | 7 | Testing | Integration |

---

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Gio learning curve | Start with simple widgets, iterate |
| Font embedding size | Use subset of JetBrains Mono glyphs |
| Channel deadlocks | Buffered channels, timeouts, careful testing |
| Animation performance | Profile early, use op.InvalidateOp sparingly |
| Cross-platform differences | Test on Windows, macOS, Linux early |

---

*Plan by: Claude Opus 4.5*
