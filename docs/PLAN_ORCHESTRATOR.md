# Orchestrator Implementation Plan

Build a TypeScript CLI orchestrator that coordinates Claude Opus 4.5 and Codex 5.2 as mob programming agents with role-based task routing and JSON file handoffs.

## Project Structure

```
cooperations/
├── package.json                    # Project config, dependencies
├── tsconfig.json                   # TypeScript configuration
├── .gitignore                      # Ignore node_modules, .env, etc.
├── src/
│   ├── index.ts                    # CLI entry point
│   ├── orchestrator/
│   │   ├── index.ts                # Main orchestrator class
│   │   ├── router.ts               # Task → Role routing logic
│   │   └── workflow.ts             # Workflow execution engine
│   ├── agents/
│   │   ├── index.ts                # Agent exports
│   │   ├── base.ts                 # Base agent interface
│   │   ├── architect.ts            # Architect role
│   │   ├── implementer.ts          # Implementer role
│   │   ├── reviewer.ts             # Reviewer role
│   │   └── navigator.ts            # Navigator role
│   ├── adapters/
│   │   ├── index.ts                # Adapter exports
│   │   ├── base.ts                 # Base adapter interface
│   │   ├── claude.ts               # Claude Opus 4.5 adapter
│   │   └── codex.ts                # Codex 5.2 adapter
│   ├── context/
│   │   ├── index.ts                # Context exports
│   │   ├── handoff.ts              # Handoff JSON schema & serialization
│   │   └── store.ts                # Local JSON file storage
│   └── types/
│       └── index.ts                # Shared TypeScript types
├── tests/
│   ├── orchestrator.test.ts        # Orchestrator unit tests
│   ├── router.test.ts              # Router unit tests
│   └── mocks/                      # Mock adapters for testing
└── examples/
    └── feature-request.json        # Example task input
```

## Implementation Phases

### Phase 1: Project Setup

**Goal:** Initialize TypeScript project with all dependencies.

**Tasks:**
- [ ] Initialize npm project with TypeScript
- [ ] Install dependencies: `commander` (CLI), `uuid`, `dotenv`, `zod`
- [ ] Install AI SDKs: `@anthropic-ai/sdk`, `openai`
- [ ] Configure tsconfig for Node.js ESM
- [ ] Set up `.gitignore` and `.env.example`

**Files:**
- `package.json`
- `tsconfig.json`
- `.gitignore`
- `.env.example`

---

### Phase 2: Type Definitions

**Goal:** Define all TypeScript types and interfaces.

**Types to define:**
```typescript
// Enums
enum Role { Architect, Implementer, Reviewer, Navigator }
enum Model { ClaudeOpus, Codex }

// Core interfaces
interface Handoff {
  task_id: string;
  timestamp: string;
  from_role: Role;
  to_role: Role;
  context: {
    task_description: string;
    requirements: string[];
    constraints: string[];
    files_in_scope: string[];
  };
  artifacts: {
    design_doc?: string;
    interfaces?: string[];
    code?: string;
    review_feedback?: string;
    notes?: string;
  };
  metadata: {
    tokens_used: number;
    model: string;
    duration_ms: number;
  };
}

interface Task {
  id: string;
  description: string;
  created_at: string;
  status: 'pending' | 'in_progress' | 'completed' | 'failed';
}

interface AgentResponse {
  content: string;
  artifacts: Record<string, any>;
  tokens_used: number;
  duration_ms: number;
  next_role?: Role;
}

interface WorkflowState {
  task: Task;
  handoffs: Handoff[];
  current_role: Role;
  review_cycles: number;
}
```

**Files:**
- `src/types/index.ts`

---

### Phase 3: Model Adapters

**Goal:** Create adapters for both AI models with unified interface.

**Interface:**
```typescript
interface BaseAdapter {
  model: Model;
  complete(prompt: string, context?: string): Promise<AdapterResponse>;
}

interface AdapterResponse {
  content: string;
  tokens_used: number;
  model: string;
}
```

**Implementations:**
| Adapter | Model | SDK |
|---------|-------|-----|
| `ClaudeAdapter` | Claude Opus 4.5 | `@anthropic-ai/sdk` |
| `CodexAdapter` | Codex 5.2 | `openai` or custom |

**Features:**
- Environment variable config (`ANTHROPIC_API_KEY`, `CODEX_API_KEY`)
- Error handling with retries (3 attempts, exponential backoff)
- Response normalization

**Files:**
- `src/adapters/base.ts`
- `src/adapters/claude.ts`
- `src/adapters/codex.ts`
- `src/adapters/index.ts`

---

### Phase 4: Agent Roles

**Goal:** Implement role-specific agents with tailored prompts.

**Base Agent:**
```typescript
abstract class BaseAgent {
  role: Role;
  model: Model;
  adapter: BaseAdapter;
  systemPrompt: string;

  abstract execute(context: Handoff): Promise<AgentResponse>;
}
```

**Role Implementations:**

| Agent | Model | Responsibilities |
|-------|-------|------------------|
| `ArchitectAgent` | Claude | System design, API contracts, patterns |
| `ImplementerAgent` | Codex | Code generation, refactoring |
| `ReviewerAgent` | Claude | Code review, security, improvements |
| `NavigatorAgent` | Either | Context tracking, next steps |

**System Prompts (summary):**
- **Architect:** "You are a software architect. Design systems, define interfaces, make structural decisions."
- **Implementer:** "You are a code implementer. Write clean, working code based on specifications."
- **Reviewer:** "You are a code reviewer. Find bugs, security issues, suggest improvements."
- **Navigator:** "You are a navigator. Track context, identify blockers, suggest next steps."

**Files:**
- `src/agents/base.ts`
- `src/agents/architect.ts`
- `src/agents/implementer.ts`
- `src/agents/reviewer.ts`
- `src/agents/navigator.ts`
- `src/agents/index.ts`

---

### Phase 5: Context Management

**Goal:** Handle handoff serialization and local file storage.

**Components:**

1. **Handoff Schema** (using Zod)
   - Validate incoming/outgoing handoffs
   - Provide type-safe parsing

2. **Context Store**
   - Storage location: `.cooperations/` in project root
   - Files: `tasks.json`, `handoffs/<task_id>.json`
   - Operations: `save()`, `load()`, `list()`, `getByTaskId()`

**Directory Structure:**
```
.cooperations/
├── tasks.json              # List of all tasks
└── handoffs/
    ├── abc123.json         # Handoff history for task abc123
    └── def456.json         # Handoff history for task def456
```

**Files:**
- `src/context/handoff.ts`
- `src/context/store.ts`
- `src/context/index.ts`

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
```typescript
class Router {
  route(task: string): Role {
    const lower = task.toLowerCase();

    if (/design|architect|plan|structure|api|interface/.test(lower)) {
      return Role.Architect;
    }
    if (/review|check|verify|audit|security/.test(lower)) {
      return Role.Reviewer;
    }
    if (/help|stuck|context|status|what|where/.test(lower)) {
      return Role.Navigator;
    }
    // Default to Implementer
    return Role.Implementer;
  }
}
```

**Logging:** Every routing decision is logged with rationale.

**Files:**
- `src/orchestrator/router.ts`

---

### Phase 7: Orchestrator Core

**Goal:** Main orchestration logic coordinating all components.

**Orchestrator Class:**
```typescript
class Orchestrator {
  private router: Router;
  private agents: Map<Role, BaseAgent>;
  private store: ContextStore;

  async run(taskDescription: string): Promise<WorkflowResult> {
    // 1. Create task
    const task = this.createTask(taskDescription);

    // 2. Route to initial role
    const role = this.router.route(taskDescription);

    // 3. Execute workflow
    return this.executeWorkflow(task, role);
  }

  private async executeWorkflow(task: Task, initialRole: Role): Promise<WorkflowResult> {
    let currentRole = initialRole;
    let context = this.createInitialContext(task);
    let reviewCycles = 0;
    const maxReviewCycles = 3;

    while (true) {
      // Execute current agent
      const agent = this.agents.get(currentRole);
      const response = await agent.execute(context);

      // Create handoff
      const handoff = this.createHandoff(currentRole, response);
      this.store.saveHandoff(task.id, handoff);

      // Determine next step
      if (response.next_role) {
        // Check review cycle limit
        if (response.next_role === Role.Reviewer) {
          reviewCycles++;
          if (reviewCycles > maxReviewCycles) {
            return this.escalateToUser(task, context);
          }
        }
        currentRole = response.next_role;
        context = this.updateContext(context, handoff);
      } else {
        // Workflow complete
        return this.completeWorkflow(task, context);
      }
    }
  }
}
```

**Workflow Types:**

1. **Feature Development:**
   ```
   Architect → Implementer → Reviewer → (loop or complete)
   ```

2. **Bug Fix:**
   ```
   Reviewer (analyze) → Architect (design fix) → Implementer → Reviewer (verify)
   ```

3. **Code Review:**
   ```
   Reviewer → (complete with feedback)
   ```

**Files:**
- `src/orchestrator/index.ts`
- `src/orchestrator/workflow.ts`

---

### Phase 8: CLI Interface

**Goal:** User-facing command-line interface.

**Commands:**

```bash
# Run a task
coop run "Add a login button to the header"
coop run --workflow=feature "Implement user authentication"
coop run --workflow=bugfix "Fix the null pointer in UserService"

# View status
coop status                    # Current/last task status
coop status <task_id>          # Specific task status

# View history
coop history                   # List all tasks
coop history --limit=10        # Last 10 tasks

# Utility
coop config                    # Show configuration
coop config set <key> <value>  # Set config value
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--verbose`, `-v` | Show detailed output including prompts |
| `--dry-run` | Show routing decision without executing |
| `--workflow` | Force specific workflow type |
| `--max-cycles` | Override max review cycles (default: 3) |

**Output Format:**
```
[ROUTE] Task routed to: Architect
[AGENT] Architect executing...
[HANDOFF] Architect → Implementer
[AGENT] Implementer executing...
[HANDOFF] Implementer → Reviewer
[AGENT] Reviewer executing...
[COMPLETE] Task completed successfully

Artifacts saved to: .cooperations/handoffs/abc123.json
```

**Files:**
- `src/index.ts`

---

### Phase 9: Testing

**Goal:** Comprehensive test coverage.

**Test Structure:**

```
tests/
├── unit/
│   ├── router.test.ts          # Routing logic tests
│   ├── handoff.test.ts         # Schema validation tests
│   └── store.test.ts           # File storage tests
├── integration/
│   └── workflow.test.ts        # Full workflow with mocks
└── mocks/
    ├── claude.mock.ts          # Mock Claude adapter
    └── codex.mock.ts           # Mock Codex adapter
```

**Test Cases:**

1. **Router Tests**
   - Routes "design a user system" → Architect
   - Routes "implement the login" → Implementer
   - Routes "review this code" → Reviewer
   - Routes unknown task → Implementer (default)

2. **Handoff Tests**
   - Valid handoff passes validation
   - Missing required fields fails
   - Serialization/deserialization roundtrip

3. **Workflow Tests**
   - Complete workflow executes all steps
   - Review cycle limit triggers escalation
   - Error in agent propagates correctly

**Files:**
- `tests/unit/*.test.ts`
- `tests/integration/*.test.ts`
- `tests/mocks/*.ts`

---

## Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Language | TypeScript | Type safety, good async support, familiar ecosystem |
| CLI Framework | Commander.js | Simple, well-documented, industry standard |
| Validation | Zod | Runtime type checking, great TypeScript integration |
| Testing | Vitest | Fast, ESM-native, Jest-compatible API |
| State Storage | Local JSON | Simple, debuggable, git-friendly, per decisions |
| Routing | Keyword regex | Simple, predictable, easy to debug and extend |
| Review Loop | Max 3 cycles | Prevent infinite loops without being too restrictive |

---

## Dependencies

```json
{
  "name": "cooperations",
  "version": "0.1.0",
  "type": "module",
  "scripts": {
    "build": "tsc",
    "start": "node dist/index.js",
    "dev": "tsx src/index.ts",
    "test": "vitest",
    "lint": "eslint src/"
  },
  "dependencies": {
    "@anthropic-ai/sdk": "^0.30.0",
    "openai": "^4.0.0",
    "commander": "^12.0.0",
    "uuid": "^9.0.0",
    "dotenv": "^16.0.0",
    "zod": "^3.22.0"
  },
  "devDependencies": {
    "typescript": "^5.3.0",
    "tsx": "^4.0.0",
    "vitest": "^1.0.0",
    "@types/node": "^20.0.0",
    "@types/uuid": "^9.0.0",
    "eslint": "^8.0.0",
    "@typescript-eslint/eslint-plugin": "^6.0.0"
  }
}
```

---

## Environment Variables

```bash
# .env.example
ANTHROPIC_API_KEY=sk-ant-...
CODEX_API_KEY=sk-...
COOPERATIONS_DIR=.cooperations
LOG_LEVEL=info
MAX_REVIEW_CYCLES=3
```

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| API key exposure | High | Use `.env`, add to `.gitignore`, document in README |
| Different API response formats | Medium | Normalize in adapters, comprehensive tests |
| Long-running tasks timeout | Medium | Add configurable timeout, show progress indicator |
| Codex API changes | Medium | Abstract behind adapter interface for easy swap |
| Infinite review loops | Low | Hard limit at 3 cycles, escalate to user |
| Context size overflow | Low | Deferred; truncate if needed |

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
