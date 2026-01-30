# GUI Implementation via Coop Tasks

**Date:** 2026-01-30
**Approach:** Use `coop run` to implement GUI incrementally

## Overview

Break GUI implementation into discrete tasks that can be executed through the Cooperations orchestrator. Each task produces one or more files that can be reviewed and approved.

## Prerequisites

Before starting coop tasks:
```bash
# Add Gio dependencies
go get gioui.org@latest
go get gioui.org/x@latest
go get github.com/alecthomas/chroma/v2@latest

# Download JetBrains Mono font
# Place in internal/gui/fonts/
```

---

## Task Sequence

### Task 1: Theme System
```bash
./coop run "implement internal/gui/theme.go with futuristic dark theme:
- Background #0a0e17, PanelBg #0d1520
- Neon accents: Cyan #00ffff, Magenta #ff00ff, Green #00ff88
- Theme struct with all colors
- DefaultTheme variable
- Helper functions for NRGBA from hex" -o internal/gui/theme.go
```

### Task 2: Font Embedding
```bash
./coop run "implement internal/gui/fonts/fonts.go:
- Embed JetBrainsMono-Regular.ttf and Bold.ttf using go:embed
- LoadFonts() returns []font.FontFace for Gio
- Support monospace and UI variants" -o internal/gui/fonts/fonts.go
```

### Task 3: Application State
```bash
./coop run "implement internal/gui/state.go:
- AppState struct with sync.RWMutex
- Fields: Task, WorkflowSteps, CurrentRole, Artifacts, Handoffs
- Fields: TokensUsed, ElapsedTime, AwaitingDecision
- Thread-safe Update methods
- Observable pattern for GUI updates" -o internal/gui/state.go
```

### Task 4: Stream Events
```bash
./coop run "implement internal/gui/stream/events.go:
- ProgressUpdate struct (Role, Status, Progress, Label, Tokens)
- CodeUpdate struct (Content, Language, Append bool)
- HandoffEvent struct (From, To, Timestamp)
- TokenUpdate struct (Total, Delta)
- DecisionRequest struct (Handoff, Artifacts)" -o internal/gui/stream/events.go
```

### Task 5: WorkflowStream
```bash
./coop run "implement internal/gui/stream/stream.go:
- WorkflowStream struct with channels for Progress, Code, Handoffs, Tokens
- Decision and Response channels for human interaction
- Done and Error channels
- NewWorkflowStream() constructor with buffered channels" -o internal/gui/stream/stream.go
```

### Task 6: HumanAgent
```bash
./coop run "implement internal/agents/human.go:
- HumanAgent struct with inputChan and outputChan
- HumanDecision struct (Action, Feedback, EditedCode)
- DecisionAction enum (Approve, Reject, Edit)
- Execute() blocks on channel, returns AgentResponse based on decision
- Approve returns nil NextRole (complete)
- Reject returns Implementer NextRole with feedback" -o internal/agents/human.go
```

### Task 7: Types Update
```bash
./coop run "update internal/types/types.go:
- Add RoleHuman Role = \"human\"
- Keep existing types unchanged" -o internal/types/types.go
```

### Task 8: NeonProgress Widget
```bash
./coop run "implement internal/gui/widgets/neon_progress.go using Gio:
- NeonProgress struct with Progress float32, Color, Glow bool
- Layout() draws gradient progress bar (cyan to magenta)
- Glow effect using multiple alpha layers
- Smooth animation support" -o internal/gui/widgets/neon_progress.go
```

### Task 9: NeonButton Widget
```bash
./coop run "implement internal/gui/widgets/neon_button.go using Gio:
- NeonButton struct with Text, Icon, Color, OnClick
- Hover detection using pointer.InputOp
- Scale animation on hover (1.0 to 1.05)
- Glow effect on hover
- Press feedback" -o internal/gui/widgets/neon_button.go
```

### Task 10: WorkflowStep Widget
```bash
./coop run "implement internal/gui/widgets/workflow_step.go using Gio:
- StepStatus enum (Pending, InProgress, Complete, Waiting)
- WorkflowStep struct with Role, Status, Progress, Label, Tokens
- Layout() shows icon, role name, progress bar, token count
- Pending: empty circle, InProgress: pulsing filled, Complete: checkmark" -o internal/gui/widgets/workflow_step.go
```

### Task 11: CodePanel Widget
```bash
./coop run "implement internal/gui/widgets/code_panel.go using Gio and Chroma:
- CodePanel struct with Code, Language strings
- Use alecthomas/chroma for syntax highlighting
- Custom dark style matching theme colors
- Scrollable viewport with widget.List
- Line numbers on left" -o internal/gui/widgets/code_panel.go
```

### Task 12: DecisionPanel Widget
```bash
./coop run "implement internal/gui/widgets/decision_panel.go using Gio:
- DecisionPanel with OnApprove, OnReject, OnEdit callbacks
- Three NeonButtons: Approve (green), Reject (red), Edit (cyan)
- Feedback text area using widget.Editor
- Visible bool to show/hide panel" -o internal/gui/widgets/decision_panel.go
```

### Task 13: HandoffLog Widget
```bash
./coop run "implement internal/gui/widgets/handoff_log.go using Gio:
- HandoffEntry struct with Timestamp, From, To roles
- HandoffLog struct with Entries slice and widget.List
- Scrollable list showing '19:45 architect -> implementer'
- New entries highlight briefly" -o internal/gui/widgets/handoff_log.go
```

### Task 14: Sidebar Panel
```bash
./coop run "implement internal/gui/panels/sidebar.go using Gio:
- Sidebar struct with WorkflowProgress, TaskContext, HandoffLog
- Layout() stacks three panels vertically
- Each panel has header with title
- Fixed width 250dp" -o internal/gui/panels/sidebar.go
```

### Task 15: MainPanel
```bash
./coop run "implement internal/gui/panels/main_panel.go using Gio:
- MainPanel with CodePanel and DesignPanel
- ShowDesign bool to toggle between views
- Flexible width, fills remaining space
- Header shows 'LIVE OUTPUT' or 'DESIGN DOCUMENT'" -o internal/gui/panels/main_panel.go
```

### Task 16: BottomPanel
```bash
./coop run "implement internal/gui/panels/bottom_panel.go using Gio:
- BottomPanel with DecisionPanel
- Visible bool controls show/hide
- Fixed height 150dp when visible
- Header shows 'HUMAN DECISION'" -o internal/gui/panels/bottom_panel.go
```

### Task 17: Layout Manager
```bash
./coop run "implement internal/gui/panels/layout.go using Gio:
- Layout struct with Sidebar, MainPanel, BottomPanel, theme
- Layout() arranges panels in 3-column design
- Header bar at top with title and stats
- Keyboard shortcut bar at bottom" -o internal/gui/panels/layout.go
```

### Task 18: Keyboard Handlers
```bash
./coop run "implement internal/gui/handlers/keyboard.go:
- HandleKeyboard(gtx, state, callbacks) function
- A key: approve, R key: reject, E key: edit
- D key: toggle design doc, H key: toggle history
- Q/Escape: quit with confirmation" -o internal/gui/handlers/keyboard.go
```

### Task 19: Main App
```bash
./coop run "implement internal/gui/app.go using Gio:
- App struct with window, theme, state, stream, layout
- NewApp() creates window titled 'COOPERATIONS'
- Run(task string) starts workflow goroutine and event loop
- handleStream() updates state from WorkflowStream channels
- eventLoop() processes Gio events and calls layout" -o internal/gui/app.go
```

### Task 20: CLI Command
```bash
./coop run "update cmd/coop/main.go:
- Add guiCmd cobra.Command with Use: 'gui <task>'
- RunE creates gui.NewApp() and calls app.Run(args[0])
- Import internal/gui package
- Register with rootCmd.AddCommand(guiCmd)" -o cmd/coop/main.go
```

### Task 21: Streaming Workflow
```bash
./coop run "implement internal/orchestrator/workflow_stream.go:
- RunWithStream(ctx, task, stream) method on Orchestrator
- Same logic as Run() but sends updates to stream channels
- After each agent: send ProgressUpdate, CodeUpdate, HandoffEvent
- When HumanAgent next: send DecisionRequest, wait for Response
- Send to Done channel when complete" -o internal/orchestrator/workflow_stream.go
```

### Task 22: Parse Update
```bash
./coop run "update internal/agents/parse.go:
- In parseNextRole() case 'user': return &types.RoleHuman instead of nil
- This routes to HumanAgent instead of terminating workflow" -o internal/agents/parse.go
```

### Task 23: Orchestrator Update
```bash
./coop run "update internal/orchestrator/orchestrator.go:
- Add humanAgent field to Orchestrator struct
- Add WithHumanAgent(agent) option function
- In getAgentForRole(): return humanAgent for RoleHuman
- Register human agent in initialization" -o internal/orchestrator/orchestrator.go
```

---

## Execution Script

Create a shell script to run all tasks:

```bash
#!/bin/bash
# gui_build.sh - Build GUI via coop tasks

set -e

echo "=== Phase 1: Infrastructure ==="
./coop run "implement theme..." -o internal/gui/theme.go
./coop run "implement fonts..." -o internal/gui/fonts/fonts.go
./coop run "implement state..." -o internal/gui/state.go

echo "=== Phase 2: Streaming ==="
./coop run "implement events..." -o internal/gui/stream/events.go
./coop run "implement stream..." -o internal/gui/stream/stream.go

echo "=== Phase 3: HumanAgent ==="
./coop run "implement human agent..." -o internal/agents/human.go

echo "=== Phase 4: Widgets ==="
./coop run "implement neon_progress..." -o internal/gui/widgets/neon_progress.go
./coop run "implement neon_button..." -o internal/gui/widgets/neon_button.go
# ... etc

echo "=== Phase 5: Panels ==="
# ...

echo "=== Phase 6: Integration ==="
# ...

echo "=== Build ==="
go build -o coop.exe ./cmd/coop

echo "Done! Run: ./coop gui 'your task'"
```

---

## Human Review Points

After each coop task, human reviews and can:
1. **Approve** - File accepted, continue to next task
2. **Reject** - Provide feedback, re-run task
3. **Edit** - Manually fix issues, continue

Recommended review after:
- Task 3 (state.go) - Core data structures
- Task 6 (human.go) - Critical for human-in-loop
- Task 11 (code_panel.go) - Complex widget
- Task 19 (app.go) - Main integration point

---

## Verification

After all tasks:
```bash
# Build
go build -o coop.exe ./cmd/coop

# Test GUI launch
./coop gui "implement hello world in Go"

# Verify:
# - Window opens with dark theme
# - Workflow progress shows
# - Code appears with syntax highlighting
# - Human decision panel activates
# - Approve completes workflow
```

---

*Plan for meta-implementation: using coop to build coop GUI*
