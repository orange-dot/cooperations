# UI Strategy: Human-in-the-Mob

**Date:** 2026-01-30
**Status:** Proposed

## Goal

Design a UI for the Cooperations mob programming workflow that enables humans to participate as active members of the mob alongside AI agents (Architect, Implementer, Reviewer).

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

### Option A: Terminal UI (TUI) with Bubble Tea

| Aspect | Details |
|--------|---------|
| **Description** | Rich terminal interface using [Bubble Tea](https://github.com/charmbracelet/bubbletea) Go library. Live display of workflow progress, agent outputs, and interactive approval prompts inline in terminal. |
| **Pros** | Pure Go, no additional stack; Works over SSH/remote; Low latency; Matches CLI workflow; Single binary distribution |
| **Cons** | Limited rich content (no images); Harder to display long code; Not shareable/collaborative |
| **Effort** | Medium |

### Option B: Local Web UI (Embedded HTTP Server)

| Aspect | Details |
|--------|---------|
| **Description** | Embedded HTTP server in Go binary serving a React/Svelte SPA. WebSocket for real-time updates. UI shows handoff chain, code diffs, approval buttons. |
| **Pros** | Rich rendering (syntax highlighting, markdown, diffs); Proper code formatting; Familiar web UX; Could evolve to hosted version |
| **Cons** | Frontend build complexity; Requires browser; More moving parts; CORS/security considerations |
| **Effort** | High |

### Option C: Hybrid - CLI with Browser Preview

| Aspect | Details |
|--------|---------|
| **Description** | Keep CLI as primary interface but add `coop ui` command that opens browser to view current task. Approvals in CLI, browser is read-only dashboard. |
| **Pros** | Minimal changes to core; Best of both worlds; Incremental path to web UI |
| **Cons** | Split attention; Approvals still CLI-based; Two systems to sync |
| **Effort** | Low-Medium |

## Recommendation

**Option A: Terminal UI (TUI)** for v1, with architecture supporting Option B later.

### Rationale

1. **Matches user profile** - Developers using CLI are comfortable in terminal
2. **Single binary** - No additional dependencies or build steps
3. **Fast iteration** - Bubble Tea is mature and well-documented
4. **Human-in-mob pattern** - TUI can block for input naturally
5. **Foundation for web** - Same backend (HumanAgent, approval checkpoints) can later expose HTTP API

## Proposed TUI Design

```
┌─────────────────────────────────────────────────────────────┐
│  Cooperations                              [tokens: 12,450] │
├─────────────────────────────────────────────────────────────┤
│ Task: implement user authentication service                 │
│ Status: Awaiting human approval          Review cycles: 1/2 │
├─────────────────────────────────────────────────────────────┤
│ ┌─ Workflow Progress ─────────────────────────────────────┐ │
│ │ ✓ Architect    design complete              845 tokens  │ │
│ │ ✓ Implementer  code generated             2,460 tokens  │ │
│ │ ✓ Reviewer     approved with comments     1,315 tokens  │ │
│ │ ● Human        awaiting decision                    ... │ │
│ └─────────────────────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│ ┌─ Review Feedback ───────────────────────────────────────┐ │
│ │ APPROVED with minor suggestions:                        │ │
│ │                                                         │ │
│ │ 1. Consider adding rate limiting to login endpoint      │ │
│ │ 2. Token expiry could be configurable                   │ │
│ │                                                         │ │
│ │ Code quality: Good                                      │ │
│ │ Security: No blocking issues                            │ │
│ └─────────────────────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│ ┌─ Generated Code (auth/service.go) ──────────────────────┐ │
│ │ package auth                                            │ │
│ │                                                         │ │
│ │ import (                                                │ │
│ │     "context"                                           │ │
│ │     "time"                                              │ │
│ │ )                                                       │ │
│ │                                                         │ │
│ │ type Service struct {                                   │ │
│ │     store    TokenStore                                 │ │
│ │     expiry   time.Duration                              │ │
│ │ }                                                       │ │
│ │ ...                                               [1/3] │ │
│ └─────────────────────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│ [a]pprove  [r]eject + feedback  [v]iew full  [q]uit        │
└─────────────────────────────────────────────────────────────┘
```

## Implementation Components

### 1. HumanAgent

New agent implementation that integrates with TUI:

```go
type HumanAgent struct {
    inputChan  chan HumanDecision
    outputChan chan Handoff
}

func (h *HumanAgent) Execute(ctx context.Context, handoff Handoff) (AgentResult, error) {
    // Send handoff to TUI for display
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

### 4. TUI Commands

| Key | Action |
|-----|--------|
| `a` | Approve current artifact, continue workflow |
| `r` | Reject with feedback prompt, route back to Implementer |
| `e` | Open editor for direct code modification |
| `v` | View full artifact (scrollable) |
| `d` | View design document |
| `c` | View generated code |
| `f` | View review feedback |
| `h` | View handoff history |
| `q` | Quit (with save prompt) |
| `?` | Help |

## Verification Strategy

### Automated Tests

- **Unit tests**: HumanAgent correctly queues/receives decisions
- **Integration test**: Full workflow with simulated human input via channel
- **Timeout test**: Workflow handles human not responding

### Manual Testing

1. Run `coop tui "implement stack in Go"`
2. Verify design is displayed after Architect
3. Approve design, verify Implementer runs
4. Verify code and review displayed
5. Test reject flow - verify routes back to Implementer
6. Test quit and resume

### Acceptance Criteria

- [ ] Human can see design before implementation starts
- [ ] Human can approve/reject code after review
- [ ] Human can provide feedback that routes back to Implementer
- [ ] Workflow state persists if terminal closes (resume support)
- [ ] Token usage displayed in real-time
- [ ] Handoff history viewable

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
| [bubbletea](https://github.com/charmbracelet/bubbletea) | TUI framework | latest |
| [lipgloss](https://github.com/charmbracelet/lipgloss) | TUI styling | latest |
| [bubbles](https://github.com/charmbracelet/bubbles) | TUI components (viewport, textinput) | latest |

## Migration Path to Web UI

The TUI implementation creates reusable components:

1. **HumanAgent** - Same interface works with HTTP handler
2. **Approval checkpoints** - Same workflow logic
3. **Extended context** - Same data model

Web UI would add:
- HTTP server with WebSocket for real-time updates
- REST endpoints for handoff retrieval
- Frontend SPA consuming same data structures

## Next Steps

1. `/plan tui` - Detailed implementation plan
2. Add Bubble Tea dependencies
3. Implement HumanAgent
4. Create TUI views (progress, code, feedback)
5. Wire up approval checkpoints
6. Add `coop tui` command
7. Test and iterate

---

*Strategy by: Claude Opus 4.5 + Human*
