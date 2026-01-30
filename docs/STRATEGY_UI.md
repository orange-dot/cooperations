# UI Strategy: Human-in-the-Mob

**Date:** 2026-01-30
**Status:** Proposed
**Updated:** 2026-01-30 - Changed from TUI to Full GUI (Gio)

## Goal

Design a futuristic widescreen GUI for the Cooperations mob programming workflow that enables humans to participate as active members of the mob alongside AI agents (Architect, Implementer, Reviewer).

**Requirements:**
- Pure Go - single binary, no CGO
- Futuristic visual design
- Widescreen optimized layout
- Full interactivity with human-in-the-loop

## Context

### Current Architecture

- Go CLI (`coop run/status/history`)
- 4 AI agent roles: Architect, Implementer, Reviewer, Navigator
- JSON-based handoff system between agents
- File storage in `.cooperations/`
- Navigator already has `NEXT: user` marker (currently unused)
- No HTTP/WebSocket interface

### Key Insight

The existing agent pattern naturally supports human insertion - humans can be another "agent" that receives handoffs and produces decisions. The `NEXT: user` marker in Navigator was designed for this purpose.

### Current Workflow

```
User → Router → Architect → Implementer → Reviewer → [loop or done]
                    ↓            ↓            ↓
                 Handoff     Handoff      Handoff
                  (JSON)      (JSON)       (JSON)
```

### Proposed Workflow with Human

```
User → Router → Architect → [Human?] → Implementer → Reviewer → [Human] → done
                    ↓           ↓            ↓            ↓          ↓
                 Handoff    Approval     Handoff      Handoff    Approval
```

## Approaches Evaluated

### Option A: Gio UI (GPU-Accelerated GUI)

| Aspect | Details |
|--------|---------|
| **Description** | Pure Go GPU-accelerated immediate mode GUI using [Gio](https://gioui.org/). Custom futuristic design with animations, glow effects, widescreen layout. |
| **Pros** | Pure Go; GPU accelerated; Custom shaders; Full visual control; Single binary; Animations |
| **Cons** | Steeper learning curve; Less documentation than alternatives |
| **Effort** | High |
| **Futuristic Potential** | Excellent |

### Option B: Ebitengine (Game Engine)

| Aspect | Details |
|--------|---------|
| **Description** | 2D game engine repurposed for UI. Total rendering control. |
| **Pros** | Pure Go on Windows; Excellent performance; Complete control |
| **Cons** | Need to build all UI primitives; Game-oriented API |
| **Effort** | Very High |
| **Futuristic Potential** | Excellent (but more work) |

### Option C: Fyne

| Aspect | Details |
|--------|---------|
| **Description** | Material Design inspired cross-platform toolkit. |
| **Pros** | Easier to use; Better documentation |
| **Cons** | Requires CGO; Material design constraints; Performance issues on macOS |
| **Effort** | Medium |
| **Futuristic Potential** | Limited |

### Option D: Terminal UI (Bubble Tea)

| Aspect | Details |
|--------|---------|
| **Description** | Rich terminal interface using Bubble Tea. |
| **Pros** | Pure Go; Works over SSH; Low latency |
| **Cons** | Terminal only; Limited visuals; Not futuristic |
| **Effort** | Medium |
| **Futuristic Potential** | Minimal |

## Recommendation

**Option A: Gio UI** - Pure Go GPU-accelerated GUI framework.

### Rationale

1. **Pure Go** - No CGO, single binary distribution
2. **GPU Accelerated** - Smooth animations, gradients, glow effects
3. **Immediate Mode** - Perfect for real-time workflow updates
4. **Custom Shaders** - Can achieve futuristic neon/glow aesthetics
5. **Widescreen Ready** - Flexible layout system (Flex, Stack)
6. **Cross-Platform** - Windows, macOS, Linux from same codebase

## Proposed GUI Design

### Visual Style

- **Dark theme** - Deep navy/black background (#0a0e17)
- **Neon accents** - Cyan (#00ffff), Magenta (#ff00ff), Green (#00ff88)
- **Glow effects** - Subtle bloom on active elements
- **Smooth animations** - Progress bars, transitions, typing effect
- **Monospace code** - Syntax highlighted with Chroma
- **Widescreen optimized** - 3-column layout for 1920x1080+

### Layout (Widescreen 1920x1080)

```
┌────────────────────────────────────────────────────────────────────────────────┐
│ ▀▀▀ COOPERATIONS ▀▀▀                                    ◉ 12,450 tokens  ⏱ 45s │
├────────────────────────────────────────────────────────────────────────────────┤
│                                                                                │
│  ┌─────────────────────┐  ┌──────────────────────────────────────────────────┐ │
│  │    MOB PROGRESS     │  │                  LIVE OUTPUT                     │ │
│  │                     │  │                                                  │ │
│  │  ◉ ARCHITECT        │  │  package auth                                    │ │
│  │    ████████████ ✓   │  │                                                  │ │
│  │                     │  │  // Service handles authentication               │ │
│  │  ◉ IMPLEMENTER      │  │  type Service struct {                           │ │
│  │    ████████░░░░ ●   │  │      store    TokenStore                         │ │
│  │    generating...    │  │      expiry   time.Duration                      │ │
│  │                     │  │  }                                               │ │
│  │  ○ REVIEWER         │  │                                                  │ │
│  │    ░░░░░░░░░░░░     │  │  func NewService(store TokenStore) *Service {    │ │
│  │                     │  │      return &Service{                            │ │
│  │  ○ HUMAN            │  │          store:  store,                          │ │
│  │    awaiting...      │  │          expiry: 24 * time.Hour,                 │ │
│  │                     │  │      }                                           │ │
│  └─────────────────────┘  │  }                                               │ │
│                           │                                                  │ │
│  ┌─────────────────────┐  │  func (s *Service) Authenticate(ctx context.Con │ │
│  │   TASK CONTEXT      │  │      token string) (*User, error) {              │ │
│  │                     │  │                                                  │ │
│  │  implement user     │  │      // Validate token format                    │ │
│  │  authentication     │  │      if !isValidFormat(token) {                  │ │
│  │  service with JWT   │  │          return nil, ErrInvalidToken             │ │
│  │  tokens             │  │      }                                           │ │
│  │                     │  │  █                                               │ │
│  │  ──────────────     │  └──────────────────────────────────────────────────┘ │
│  │  Review: 1/2        │                                                       │
│  │  Model: codex-5.2   │  ┌──────────────────────────────────────────────────┐ │
│  └─────────────────────┘  │              HUMAN DECISION                      │ │
│                           │                                                  │ │
│  ┌─────────────────────┐  │   ┌──────────┐  ┌──────────┐  ┌──────────────┐   │ │
│  │    HANDOFF LOG      │  │   │ APPROVE  │  │  REJECT  │  │ EDIT & SEND  │   │ │
│  │                     │  │   │    ✓     │  │    ✗     │  │      ✎       │   │ │
│  │  19:45 architect→   │  │   └──────────┘  └──────────┘  └──────────────┘   │ │
│  │        implementer  │  │                                                  │ │
│  │  19:46 implementer→ │  │   Feedback: ____________________________________│ │
│  │        reviewer     │  │   _____________________________________________ │ │
│  │  19:47 reviewer→    │  │                                                  │ │
│  │        human        │  └──────────────────────────────────────────────────┘ │
│  └─────────────────────┘                                                       │
│                                                                                │
├────────────────────────────────────────────────────────────────────────────────┤
│  [A] Approve   [R] Reject   [E] Edit   [D] Design Doc   [H] History   [Q] Quit │
└────────────────────────────────────────────────────────────────────────────────┘
```

### Color Palette

| Element | Color | Hex |
|---------|-------|-----|
| Background | Deep Navy | #0a0e17 |
| Panel Background | Dark Blue | #0d1520 |
| Border | Dim Cyan | #1a3a4a |
| Active Border | Bright Cyan | #00ffff |
| Text Primary | White | #ffffff |
| Text Secondary | Gray | #8899aa |
| Success/Approve | Neon Green | #00ff88 |
| Error/Reject | Neon Red | #ff4466 |
| Warning | Neon Orange | #ffaa00 |
| Accent | Neon Magenta | #ff00ff |
| Progress Bar | Gradient Cyan→Magenta | #00ffff → #ff00ff |

### Animations

| Element | Animation | Duration |
|---------|-----------|----------|
| Progress bars | Smooth fill | 300ms ease-out |
| Panel focus | Border glow pulse | 1s loop |
| Code output | Typing effect | Real-time |
| Button hover | Scale + glow | 150ms |
| Transitions | Fade in/out | 200ms |

## Implementation Components

### 1. HumanAgent

New agent implementation that integrates with GUI:

```go
type HumanAgent struct {
    inputChan  chan HumanDecision
    outputChan chan Handoff
}

func (h *HumanAgent) Execute(ctx context.Context, handoff Handoff) (AgentResult, error) {
    // Send handoff to GUI for display
    h.outputChan <- handoff

    // Block waiting for human decision
    select {
    case decision := <-h.inputChan:
        return h.processDecision(decision, handoff)
    case <-ctx.Done():
        return AgentResult{}, ctx.Err()
    }
}
```

### 2. Custom Gio Widgets

```go
// Futuristic progress bar with glow effect
type NeonProgress struct {
    Progress float32      // 0.0 to 1.0
    Color    color.NRGBA  // Primary color
    Glow     bool         // Enable glow effect
}

func (p *NeonProgress) Layout(gtx layout.Context) layout.Dimensions {
    // GPU-accelerated gradient + glow rendering
}

// Syntax-highlighted code panel
type CodePanel struct {
    Code      string
    Language  string
    Lexer     chroma.Lexer
    Theme     *chroma.Style
}

// Human decision panel with animated buttons
type DecisionPanel struct {
    OnApprove func()
    OnReject  func(feedback string)
    OnEdit    func()
    Feedback  *widget.Editor
}

// Animated workflow progress indicator
type WorkflowProgress struct {
    Steps    []WorkflowStep
    Current  int
    Animated bool
}
```

### 3. Real-time Streaming

```go
// Workflow streams updates to GUI via channels
type WorkflowStream struct {
    Progress   chan ProgressUpdate  // Step completion
    CodeOutput chan string          // Live code generation
    Handoffs   chan Handoff         // Agent transitions
    Decision   chan HumanDecision   // Human responses
    Tokens     chan int             // Token counter
}

// GUI listens and updates immediately (60 FPS)
func (app *App) handleStream(stream *WorkflowStream) {
    for {
        select {
        case p := <-stream.Progress:
            app.updateProgress(p)
            app.window.Invalidate() // Trigger redraw
        case code := <-stream.CodeOutput:
            app.appendCode(code)
            app.window.Invalidate()
        // ...
        }
    }
}
```

### 2. Approval Checkpoints

| Checkpoint | Trigger | Options |
|------------|---------|---------|
| Post-Design | After Architect completes | Approve / Edit requirements / Reject |
| Post-Review | After Reviewer approves | Accept code / Request changes / Reject |
| On Blocker | Navigator detects issue | Clarify / Skip / Abort |

### 3. Extended Context

```go
type HContext struct {
    TaskDescription string
    Requirements    []string
    Constraints     []string
    FilesInScope    []string
    // New fields for human interaction
    HumanDecision   string   // approve/reject/edit
    HumanFeedback   string   // Free-form feedback text
    HumanEditedCode string   // If human modified code directly
}
```

### 4. GUI Interactions

| Input | Action |
|-------|--------|
| Click APPROVE button | Approve current artifact, continue workflow |
| Click REJECT button | Open feedback input, route back to Implementer |
| Click EDIT button | Open code editor panel |
| `A` key | Approve (keyboard shortcut) |
| `R` key | Reject (keyboard shortcut) |
| `E` key | Edit (keyboard shortcut) |
| `D` key | Toggle design document panel |
| `H` key | Toggle handoff history panel |
| `Esc` / `Q` | Quit (with save prompt) |
| Mouse wheel | Scroll code/feedback panels |
| Click panel | Focus panel |

## Verification Strategy

### Automated Tests

- **Unit tests**: HumanAgent correctly queues/receives decisions
- **Widget tests**: Gio widgets render correctly
- **Integration test**: Full workflow with simulated human input via channel
- **Timeout test**: Workflow handles human not responding

### Manual Testing

1. Run `coop gui "implement stack in Go"`
2. Verify futuristic theme renders correctly
3. Verify workflow progress animates
4. Verify code panel shows syntax highlighting
5. Test approve flow - verify workflow continues
6. Test reject flow - verify routes back to Implementer with feedback
7. Test keyboard shortcuts (A, R, E, D, H)
8. Test window resize behavior
9. Test quit and resume

### Acceptance Criteria

- [ ] Futuristic dark theme with neon accents
- [ ] Smooth animations at 60 FPS
- [ ] Human can see design before implementation starts
- [ ] Human can approve/reject code after review
- [ ] Human can provide feedback that routes back to Implementer
- [ ] Syntax highlighting in code panel
- [ ] Workflow state persists on close (resume support)
- [ ] Token usage displayed in real-time
- [ ] Handoff history viewable
- [ ] Single binary deployment (no external assets)

## Open Questions

### 1. Approval Granularity

Should human approval be required:
- After every agent? (verbose but safe)
- Only after Reviewer? (minimal friction)
- Configurable via `--approve-after` flag? (flexible)

**Recommendation:** Configurable, default to post-review only.

### 2. Edit Capability

Should humans be able to:
- Just approve/reject with comments?
- Edit code directly in TUI?
- Provide text feedback only?

**Recommendation:** Start with approve/reject + feedback. Add inline editing in v2.

### 3. Resume Support

If user quits mid-workflow:
- Auto-save state and allow `coop resume <task_id>`?
- Prompt before quit?

**Recommendation:** Auto-save on quit, add `coop resume` command.

### 4. Multi-Human (Future)

Multiple humans in same mob for pair programming?

**Recommendation:** Out of scope for v1. Architecture should not preclude it.

## Dependencies

| Dependency | Purpose | Version |
|------------|---------|---------|
| [gioui.org](https://gioui.org/) | GUI framework (pure Go, GPU accelerated) | latest |
| [gioui.org/x](https://pkg.go.dev/gioui.org/x) | Extended Gio widgets | latest |
| [github.com/alecthomas/chroma](https://github.com/alecthomas/chroma) | Syntax highlighting (pure Go) | v2 |
| Embedded font | JetBrains Mono or Fira Code | - |

All dependencies are pure Go - no CGO required.

## Architecture Benefits

The Gio implementation creates reusable components:

1. **HumanAgent** - Decoupled from UI, communicates via channels
2. **Approval checkpoints** - Same workflow logic as CLI
3. **Extended context** - Same data model
4. **Streaming interface** - Can be reused for web UI later

Future web UI would reuse:
- HumanAgent and channel-based communication
- Workflow streaming interface
- Approval checkpoint logic
- Only replace Gio rendering with HTML/WebSocket

## File Structure

```
internal/
├── gui/
│   ├── app.go           # Main Gio application
│   ├── theme.go         # Futuristic color palette & styles
│   ├── fonts.go         # Embedded fonts (JetBrains Mono)
│   ├── widgets/
│   │   ├── progress.go  # NeonProgress bar
│   │   ├── code.go      # CodePanel with syntax highlighting
│   │   ├── decision.go  # Human decision panel
│   │   ├── workflow.go  # Workflow progress indicator
│   │   ├── handoff.go   # Handoff history log
│   │   └── button.go    # Futuristic animated buttons
│   ├── panels/
│   │   ├── sidebar.go   # Left sidebar (progress, context, log)
│   │   ├── main.go      # Main content (code output)
│   │   └── bottom.go    # Bottom panel (human decision)
│   └── stream.go        # Real-time workflow streaming
├── agents/
│   └── human.go         # HumanAgent implementation
cmd/
└── coop/
    └── gui.go           # `coop gui` command
```

## Next Steps

1. `/plan gui` - Detailed implementation plan
2. Add Gio dependencies (`go get gioui.org@latest`)
3. Create theme and color palette
4. Implement basic window with panels
5. Create custom widgets (NeonProgress, CodePanel)
6. Implement HumanAgent with channel communication
7. Wire up workflow streaming
8. Add approval checkpoints
9. Add `coop gui` command
10. Test and iterate

---

*Strategy by: Claude Opus 4.5 + Human*
