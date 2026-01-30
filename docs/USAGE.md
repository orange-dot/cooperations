# Cooperations CLI Usage

## Quick Start

```bash
# Build the CLI
go build -o coop ./cmd/coop

# Configure API keys in .env
ANTHROPIC_API_KEY=sk-ant-...
CODEX_API_KEY=sk-...

# Run a task
./coop run "implement a health check endpoint"
```

## Commands

### run

Execute a task through the mob programming workflow.

```bash
coop run <task> [flags]
```

**Flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--output` | `-o` | Write generated code to file |
| `--dry-run` | | Show routing decision without executing |
| `--verbose` | `-v` | Show detailed output including generated code |
| `--max-cycles` | | Override max review cycles (default: 2) |
| `--workflow` | | Force workflow type: `feature`, `bugfix`, `review` |

**Examples:**
```bash
# Basic usage - auto-routes based on keywords
coop run "design a user authentication system"
coop run "implement JWT token validation"
coop run "review the login handler for security issues"

# Write generated code to file
coop run -o dijkstra.go "implement dijkstra algorithm"
coop run --output api.go "implement REST API handler"

# Combine flags
coop run -v -o output.go --max-cycles 2 "implement binary search"

# Dry run - preview routing without API calls
coop run --dry-run "fix the nil pointer in UserService"

# Verbose - see full output including generated code
coop run -v "add input validation to the API"

# Override review cycles
coop run --max-cycles 3 "refactor the database layer"
```

### status

Show task status.

```bash
coop status [task_id]
```

**Examples:**
```bash
# Show most recent task
coop status

# Show specific task with handoff history
coop status 1706620800000000000
```

**Output:**
```
Task: 1706620800000000000
Status: completed
Created: 2024-01-30T12:00:00Z
Description: implement a health check endpoint

Handoffs: 3
  1. implementer -> reviewer (codex-5-2, 1200 tokens)
  2. reviewer -> implementer (claude-opus-4-5, 800 tokens)
  3. implementer -> done (codex-5-2, 600 tokens)
```

### history

List past tasks.

```bash
coop history [flags]
```

**Flags:**
| Flag | Description |
|------|-------------|
| `--limit` | Number of tasks to show (default: 10) |

**Examples:**
```bash
# Show last 10 tasks
coop history

# Show last 20 tasks
coop history --limit=20
```

**Output:**
```
Recent tasks (showing 10 of 25):

  1706620800000000000  [completed]  implement a health check endpoint
  1706620900000000000  [completed]  add logging to the API
  1706621000000000000  [failed]     design a caching layer
```

## Routing Rules

Tasks are automatically routed to the appropriate agent based on keywords:

| Keywords | Target Role | Model |
|----------|-------------|-------|
| `design`, `architect`, `plan`, `structure`, `api`, `interface`, `schema` | Architect | Claude Opus 4.5 |
| `implement`, `code`, `build`, `create`, `write`, `add`, `fix`, `bug` | Implementer | Codex 5.2 |
| `review`, `check`, `verify`, `test`, `audit`, `security` | Reviewer | Claude Opus 4.5 |
| `help`, `stuck`, `context`, `status`, `what`, `where`, `why`, `how` | Navigator | Claude Opus 4.5 |
| (no match) | Implementer | Codex 5.2 |

Use `--dry-run` to preview routing:
```bash
$ coop run --dry-run "design a REST API"
[DRY-RUN] Task would be routed to: architect (confidence: 100%)
```

## Workflow

### Feature Development
```
User Request
    ↓
Architect (design)
    ↓
Implementer (code)
    ↓
Reviewer (review)
    ↓ (if changes needed, loop back to Implementer)
Done
```

### Bug Fix
```
User Request
    ↓
Reviewer (analyze bug)
    ↓
Architect (design fix)
    ↓
Implementer (implement fix)
    ↓
Reviewer (verify fix)
    ↓
Done
```

### Review Cycle Limit

By default, the workflow allows **2 review cycles** before stopping. This prevents infinite loops when the Reviewer and Implementer disagree.

Override with `--max-cycles`:
```bash
coop run --max-cycles 3 "implement complex feature"
```

## Output Files

### Writing Code to File

Use `-o` / `--output` to write generated code directly to a file:

```bash
coop run -o mycode.go "implement a stack data structure"
```

Output:
```
[START] Running task: implement a stack data structure
[COMPLETE] Task 1706620800000000000 completed successfully
Artifacts saved to: .cooperations/handoffs/1706620800000000000.json
Code written to: mycode.go
```

The CLI automatically extracts code from markdown code blocks if present.

### Task Data

All task data is stored in `.cooperations/`:

```
.cooperations/
├── tasks.json              # List of all tasks
└── handoffs/
    ├── 170662080000.json   # Handoff history for task
    └── 170662090000.json   # Handoff history for task
```

Generated artifacts are stored in `generated/` by default:

```
generated/
└── 1706620800000000000/
    ├── README.md
    ├── design.md
    ├── review.md
    └── code/
        └── main.go
```

### Handoff Format

Each handoff is a JSON object:
```json
{
  "task_id": "1706620800000000000",
  "timestamp": "2024-01-30T12:00:00Z",
  "from_role": "implementer",
  "to_role": "reviewer",
  "context": {
    "task_description": "implement a health check endpoint",
    "requirements": [],
    "constraints": [],
    "files_in_scope": []
  },
  "artifacts": {
    "code": "func HealthCheck(w http.ResponseWriter, r *http.Request) { ... }"
  },
  "metadata": {
    "tokens_used": 1200,
    "model": "codex-5-2",
    "duration_ms": 3500
  }
}
```

## Environment Variables

Configure in `.env` file:

| Variable | Default | Description |
|----------|---------|-------------|
| `ANTHROPIC_API_KEY` | (required) | Claude Opus 4.5 API key |
| `CODEX_API_KEY` | (required) | Codex 5.2 API key |
| `COOPERATIONS_DIR` | `.cooperations` | Storage directory |
| `COOPERATIONS_GENERATED_DIR` | `generated` | Generated artifacts directory |
| `LOG_LEVEL` | `info` | Logging level: `debug`, `info`, `warn`, `error` |
| `MAX_REVIEW_CYCLES` | `2` | Default max review cycles |

## Troubleshooting

### "ANTHROPIC_API_KEY environment variable not set"
Ensure `.env` file exists and contains valid API key.

### "exceeded max review cycles"
The Reviewer requested changes more times than allowed. Either:
- Increase `--max-cycles`
- Simplify the task
- Check if requirements are clear

### Task stuck in loop
The workflow stops automatically after max review cycles. Check the handoff history to see what the Reviewer is requesting.

```bash
coop status <task_id>
```

### Debug logging
Set `LOG_LEVEL=debug` in `.env` to see detailed routing and execution info.
