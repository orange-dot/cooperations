# Orchestrator Implementation Plan

Build a Go CLI orchestrator that coordinates Claude Opus 4.5 and Codex 5.2 as mob programming agents with role-based task routing and JSON file handoffs. The orchestrator runs locally, uses local storage, and has no managed cloud dependencies (Docker for local runs is OK).

## Project Structure

```
cooperations/
├── go.mod
├── go.sum
├── .gitignore                      # Ignore .cooperations/, .env, etc.
├── cmd/
│   └── coop/
│       └── main.go                 # CLI entry point
├── internal/
│   ├── orchestrator/
│   │   ├── orchestrator.go         # Main orchestrator
│   │   ├── router.go               # Task -> Role routing logic
│   │   └── workflow.go             # Workflow execution engine
│   ├── agents/
│   │   ├── agent.go                # Agent interface
│   │   ├── architect.go            # Architect role
│   │   ├── implementer.go          # Implementer role
│   │   ├── reviewer.go             # Reviewer role
│   │   └── navigator.go            # Navigator role
│   ├── adapters/
│   │   ├── adapter.go              # Adapter interface
│   │   ├── claude.go               # Claude Opus 4.5 adapter
│   │   └── codex.go                # Codex 5.2 adapter
│   ├── context/
│   │   ├── handoff.go              # Handoff schema + validation
│   │   └── store.go                # Local JSON file storage
│   ├── types/
│   │   └── types.go                # Shared types
│   └── logging/
│       └── logger.go               # slog setup + helpers
├── tests/
│   ├── router_test.go
│   ├── handoff_test.go
│   └── workflow_test.go
└── examples/
    └── feature-request.json        # Example task input
```

## Implementation Phases

### Phase 1: Project Setup

**Goal:** Initialize Go project with CLI, config, and logging.

**Tasks:**
- [ ] `go mod init` and `go mod tidy`
- [ ] Set up CLI framework (Cobra)
- [ ] Add dotenv loading for local runs
- [ ] Configure logging with `log/slog`
- [ ] Create `.env.example` and `.gitignore`

**Files:**
- `go.mod`
- `.gitignore`
- `.env.example`
- `cmd/coop/main.go`

---

### Phase 2: Type Definitions

**Goal:** Define all shared types and JSON structures.

**Types to define:**
```go
type Role string
const (
  RoleArchitect   Role = "architect"
  RoleImplementer Role = "implementer"
  RoleReviewer    Role = "reviewer"
  RoleNavigator   Role = "navigator"
)

type Model string
const (
  ModelClaude Model = "claude-opus-4-5"
  ModelCodex  Model = "codex-5-2"
)

type Handoff struct {
  TaskID    string    `json:"task_id" validate:"required"`
  Timestamp string    `json:"timestamp" validate:"required"`
  FromRole  Role      `json:"from_role" validate:"required"`
  ToRole    Role      `json:"to_role" validate:"required"`
  Context   HContext  `json:"context" validate:"required"`
  Artifacts HArtifacts `json:"artifacts"`
  Metadata  HMetadata `json:"metadata" validate:"required"`
}

type HContext struct {
  TaskDescription string   `json:"task_description"`
  Requirements    []string `json:"requirements"`
  Constraints     []string `json:"constraints"`
  FilesInScope    []string `json:"files_in_scope"`
}

type HArtifacts struct {
  DesignDoc      string   `json:"design_doc,omitempty"`
  Interfaces     []string `json:"interfaces,omitempty"`
  Code           string   `json:"code,omitempty"`
  ReviewFeedback string   `json:"review_feedback,omitempty"`
  Notes          string   `json:"notes,omitempty"`
}

type HMetadata struct {
  TokensUsed int    `json:"tokens_used"`
  Model      string `json:"model"`
  DurationMS int64  `json:"duration_ms"`
}

type Task struct {
  ID          string `json:"id"`
  Description string `json:"description"`
  CreatedAt   string `json:"created_at"`
  Status      string `json:"status"` // pending|in_progress|completed|failed
}

type AgentResponse struct {
  Content    string                 `json:"content"`
  Artifacts  map[string]any         `json:"artifacts"`
  TokensUsed int                    `json:"tokens_used"`
  DurationMS int64                  `json:"duration_ms"`
  NextRole   *Role                  `json:"next_role,omitempty"`
}

type WorkflowState struct {
  Task         Task     `json:"task"`
  Handoffs     []Handoff `json:"handoffs"`
  CurrentRole  Role     `json:"current_role"`
  ReviewCycles int      `json:"review_cycles"`
}
```

**Files:**
- `internal/types/types.go`

---

### Phase 3: Model Adapters

**Goal:** Create adapters for both AI models with a unified interface.

**Interface:**
```go
type Adapter interface {
  Model() Model
  Complete(ctx context.Context, prompt string, contextText string) (AdapterResponse, error)
}

type AdapterResponse struct {
  Content    string
  TokensUsed int
  Model      string
}
```

**Implementations:**
- `ClaudeAdapter`: Anthropic API (SDK or direct HTTP)
- `CodexAdapter`: OpenAI/Codex API (SDK or direct HTTP)

**Features:**
- Environment variable config (`ANTHROPIC_API_KEY`, `CODEX_API_KEY`)
- Error handling with retries (3 attempts, exponential backoff)
- Response normalization

**Files:**
- `internal/adapters/adapter.go`
- `internal/adapters/claude.go`
- `internal/adapters/codex.go`

---

### Phase 4: Agent Roles

**Goal:** Implement role-specific agents with tailored prompts.

**Agent interface:**
```go
type Agent interface {
  Role() Role
  Execute(ctx context.Context, handoff Handoff) (AgentResponse, error)
}
```

**Role Implementations:**

| Agent | Model | Responsibilities |
|-------|-------|------------------|
| Architect | Claude | System design, API contracts, patterns |
| Implementer | Codex | Code generation, refactoring |
| Reviewer | Claude | Code review, security, improvements |
| Navigator | Either | Context tracking, next steps |

**System Prompts (summary):**
- Architect: "You are a software architect. Design systems, define interfaces, make structural decisions."
- Implementer: "You are a code implementer. Write clean, working code based on specifications."
- Reviewer: "You are a code reviewer. Find bugs, security issues, suggest improvements."
- Navigator: "You are a navigator. Track context, identify blockers, suggest next steps."

**Files:**
- `internal/agents/agent.go`
- `internal/agents/architect.go`
- `internal/agents/implementer.go`
- `internal/agents/reviewer.go`
- `internal/agents/navigator.go`

---

### Phase 5: Context Management

**Goal:** Handle handoff serialization and local file storage.

**Components:**
1. **Handoff Validation**
   - Use struct validation and manual checks
   - Enforce required fields and timestamps

2. **Context Store**
   - Storage location: `.cooperations/` in project root
   - Files: `tasks.json`, `handoffs/<task_id>.json`
   - Operations: `Save()`, `Load()`, `List()`, `GetByTaskID()`

**Directory Structure:**
```
.cooperations/
├── tasks.json              # List of all tasks
└── handoffs/
    ├── abc123.json         # Handoff history for task abc123
    └── def456.json         # Handoff history for task def456
```

**Files:**
- `internal/context/handoff.go`
- `internal/context/store.go`

---

### Phase 6: Router

**Goal:** Route tasks to appropriate roles based on keywords.

**Routing Rules:**

| Keywords | Target Role |
|----------|-------------|
| `design`, `architect`, `plan`, `structure`, `api`, `interface` | Architect |
| `implement`, `code`, `build`, `create`, `write`, `add`, `fix` | Implementer |
| `review`, `check`, `verify`, `test`, `audit`, `security` | Reviewer |
| `help`, `stuck`, `context`, `status`, `what`, `where` | Navigator |
| (default) | Implementer |

**Implementation:**
```go
type Router struct {}

func (r *Router) Route(task string) Role {
  lower := strings.ToLower(task)
  switch {
  case regexp.MustCompile(`design|architect|plan|structure|api|interface`).MatchString(lower):
    return RoleArchitect
  case regexp.MustCompile(`review|check|verify|audit|security`).MatchString(lower):
    return RoleReviewer
  case regexp.MustCompile(`help|stuck|context|status|what|where`).MatchString(lower):
    return RoleNavigator
  default:
    return RoleImplementer
  }
}
```

**Logging:** Every routing decision is logged with rationale.

**Files:**
- `internal/orchestrator/router.go`

---

### Phase 7: Orchestrator Core

**Goal:** Main orchestration logic coordinating all components.

**Orchestrator:**
```go
type Orchestrator struct {
  router *Router
  agents map[Role]Agent
  store  *ContextStore
}

func (o *Orchestrator) Run(ctx context.Context, taskDescription string) (WorkflowResult, error) {
  task := o.createTask(taskDescription)
  role := o.router.Route(taskDescription)
  return o.executeWorkflow(ctx, task, role)
}
```

**Workflow Types:**
1. Feature Development:
   ```
   Architect -> Implementer -> Reviewer -> (loop or complete)
   ```
2. Bug Fix:
   ```
   Reviewer (analyze) -> Architect (design fix) -> Implementer -> Reviewer (verify)
   ```
3. Code Review:
   ```
   Reviewer -> (complete with feedback)
   ```

**Files:**
- `internal/orchestrator/orchestrator.go`
- `internal/orchestrator/workflow.go`

---

### Phase 8: CLI Interface

**Goal:** User-facing command-line interface.

**Commands:**
```bash
# Run a task
coop run "Add a login button to the header"
coop run --workflow=feature "Implement user authentication"
coop run --workflow=bugfix "Fix the nil pointer in UserService"

# View status
coop status
coop status <task_id>

# View history
coop history
coop history --limit=10

# Utility
coop config
coop config set <key> <value>
```

**Flags:**
| Flag | Description |
|------|-------------|
| `--verbose`, `-v` | Show detailed output including prompts |
| `--dry-run` | Show routing decision without executing |
| `--workflow` | Force specific workflow type |
| `--max-cycles` | Override max review cycles (default: 2) |

**Output Format:**
```
[ROUTE] Task routed to: Architect
[AGENT] Architect executing...
[HANDOFF] Architect -> Implementer
[AGENT] Implementer executing...
[HANDOFF] Implementer -> Reviewer
[AGENT] Reviewer executing...
[COMPLETE] Task completed successfully

Artifacts saved to: .cooperations/handoffs/abc123.json
```

**Files:**
- `cmd/coop/main.go`

---

### Phase 9: Testing

**Goal:** Comprehensive test coverage.

**Test Structure:**
```
internal/orchestrator/router_test.go
internal/context/handoff_test.go
internal/context/store_test.go
internal/orchestrator/workflow_test.go
internal/adapters/*_test.go
```

**Test Cases:**
1. Router tests (keyword routing + defaults)
2. Handoff validation tests (required fields, roundtrip)
3. Workflow tests (review cycle limit triggers escalation)
4. Adapter tests (mocked HTTP responses)

---

## Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Language | Go | Simple deployment, fast CLI, stdlib support |
| CLI Framework | Cobra | Reliable, common Go CLI pattern |
| Validation | go-playground/validator | Lightweight struct validation |
| Testing | go test + testify (optional) | Standard tooling |
| State Storage | Local JSON | Simple, debuggable, git-friendly |
| Routing | Keyword regex | Simple, predictable, easy to extend |
| Review Loop | Max 2 cycles | Prevents churn and deadlocks |

---

## Dependencies

- `github.com/spf13/cobra` (CLI)
- `github.com/spf13/viper` (config/env, optional)
- `github.com/joho/godotenv` (local .env)
- `github.com/google/uuid` (task IDs)
- `github.com/go-playground/validator/v10` (validation)
- `github.com/cenkalti/backoff/v4` (retries)

---

## Environment Variables

```bash
# .env.example
ANTHROPIC_API_KEY=sk-ant-...
CODEX_API_KEY=sk-...
COOPERATIONS_DIR=.cooperations
LOG_LEVEL=info
MAX_REVIEW_CYCLES=2
```

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| API key exposure | High | Use `.env`, add to `.gitignore`, document in README |
| Different API response formats | Medium | Normalize in adapters, comprehensive tests |
| Long-running tasks timeout | Medium | Add configurable timeout, show progress indicator |
| Codex API changes | Medium | Abstract behind adapter interface for easy swap |
| Infinite review loops | Low | Hard limit at 2 cycles, escalate to user |
| Context size overflow | Low | Enforce token budgets, summarize older context |

---

## Milestones

| Milestone | Deliverable | Status |
|-----------|-------------|--------|
| M1 | Project setup, types, adapters | Pending |
| M2 | Agents and context management | Pending |
| M3 | Router and orchestrator core | Pending |
| M4 | CLI interface | Pending |
| M5 | Tests and documentation | Pending |

---

## Next Steps

1. Run `/implement orchestrator` to begin Phase 1
2. Or implement phases incrementally with `/implement phase-1-setup`

---

*Document created: 2026-01-30*
*Status: Plan complete, ready for implementation*
