# Cooperations: AI Mob Programming Workflow

A greenfield project combining multiple AI models (Claude Opus 4.5 and Codex 5.2) as collaborative "mob programmers" for software development.

## Vision

Traditional mob programming involves multiple developers working together - one "driver" types while others "navigate" and review. This project applies that concept to AI models, leveraging each model's unique strengths in a coordinated workflow.

## Model Characteristics

| Model | Strengths | Best For |
|-------|-----------|----------|
| Claude Opus 4.5 | Reasoning, planning, nuanced understanding, code review | Architecture, design decisions, review, complex debugging |
| Codex 5.2 | Code generation, completion, rapid prototyping | Implementation, boilerplate, refactoring, broad coverage |

## Architecture: Role-Specialized Agents with Orchestrator

We use a role-based system where specialized agents can be filled by the most appropriate model. A lightweight orchestrator routes work and manages shared context.

```
┌─────────────────────────────────────────────────────────┐
│                     ORCHESTRATOR                        │
│  - Routes tasks to appropriate roles                    │
│  - Maintains shared context                             │
│  - Manages handoffs between agents                      │
│  - Tracks progress and artifacts                        │
└─────────────────────────────────────────────────────────┘
         │              │              │              │
         ▼              ▼              ▼              ▼
   ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐
   │ARCHITECT │  │IMPLEMENTER│  │ REVIEWER │  │NAVIGATOR │
   │(Opus 4.5)│  │(Codex 5.2)│  │(Opus 4.5)│  │ (Either) │
   └──────────┘  └──────────┘  └──────────┘  └──────────┘
```

### Roles

#### Architect (Primary: Opus 4.5)
- System design and high-level structure
- API contracts and interfaces
- Pattern selection and enforcement
- Technical decision documentation

#### Implementer (Primary: Codex 5.2)
- Code generation from specifications
- Boilerplate and scaffolding
- Refactoring existing code
- Rapid prototyping

#### Reviewer (Primary: Opus 4.5)
- Code review for correctness and style
- Security vulnerability detection
- Performance analysis
- Suggesting improvements

#### Navigator (Either model)
- Context tracking across the session
- Suggesting next steps
- Identifying blockers
- Maintaining task focus

## Workflow Examples

### Example 1: New Feature Development

```
1. User Request → Orchestrator
2. Orchestrator → Architect: "Design the feature"
3. Architect produces: Design doc, interfaces, file structure
4. Orchestrator → Implementer: "Build to this spec"
5. Implementer produces: Working code
6. Orchestrator → Reviewer: "Review this implementation"
7. Reviewer produces: Feedback, issues, suggestions
8. Orchestrator → Implementer: "Address review feedback"
9. Loop until approved
10. Final artifact delivered
```

### Example 2: Bug Fix

```
1. Bug Report → Orchestrator
2. Orchestrator → Reviewer: "Analyze and locate the bug"
3. Reviewer produces: Root cause analysis, affected files
4. Orchestrator → Architect: "Design the fix approach"
5. Architect produces: Fix strategy, risk assessment
6. Orchestrator → Implementer: "Implement the fix"
7. Implementer produces: Fixed code
8. Orchestrator → Reviewer: "Verify the fix"
9. Final verification and delivery
```

## Alternatives Considered

### Option A: Sequential Handoff Pipeline
Tasks flow through a fixed pipeline: Plan → Implement → Review → Revise.

- **Rejected because**: Too rigid, loses real-time collaborative benefits, context degrades between handoffs.

### Option B: Parallel Voting/Consensus
Both models work simultaneously, then merge or select the best output.

- **Rejected because**: Higher costs (2x API calls), complex merge logic for code, conflicting approaches hard to reconcile.

## Open Questions

All open questions are tracked in `docs/OPEN_QUESTIONS.md` for refinement.

## Verification Strategy

### Automated Testing
- Unit tests for orchestrator routing logic
- Integration tests with mock model responses
- End-to-end task completion tests

### Quality Metrics
- Compare output quality: solo model vs. collaborative
- Measure time-to-completion for standard tasks
- Track revision cycles needed

### Cost Tracking
- Monitor API usage per role
- Ensure collaborative approach remains cost-effective
- Identify optimization opportunities

## Definition of Done (DoD)

- Orchestrator routes tasks with ≥ 90% correct role selection on a curated test set
- End-to-end workflows complete successfully on ≥ 80% of defined scenarios
- Median time-to-completion improves by ≥ 20% vs. single-model baseline
- Cost per task is ≤ 1.2x the single-model baseline
- Audit logs are complete and reproducible for each run

## Orchestrator Decision Policy

- **Default**: rule-based routing using task type, size, and risk classification
- **Confidence**: heuristic score based on task size, file count, and risk keywords
- **Adaptation**: allow mid-task handoffs when confidence drops or blockers persist
- **Fallback**: when disagreement persists, escalate to Reviewer or Architect
- **Transparency**: every routing decision is logged with its rationale

## Context Management and Limits

- Use structured handoff summaries (JSON) that capture goals, constraints, files in scope, decisions, and next actions
- Store large artifacts on disk and reference by path instead of embedding content
- Enforce token budgets per handoff (e.g., 25% persistent context, 50% working set, 25% recent summary)
- Prune by keeping the last two handoffs verbatim and summarizing older context into decisions + blockers

## Conflict Resolution Policy

| Situation | Primary Tiebreaker | Secondary |
|----------|--------------------|-----------|
| Safety or security concern | Reviewer | Architect |
| Architectural disagreement | Architect | Reviewer |
| Implementation detail | Implementer | Reviewer |
| Style or formatting | Project standards | Reviewer |
| Unclear requirements | Navigator | User |

Decision protocol:
- Prefer objective evidence (tests, logs, benchmarks) over preferences
- If evidence is inconclusive, choose the safer/simpler option
- Escalate ambiguous requirements to the user
- Record the decision and rationale in the run log

## Escalation Rules

- Cap at two review loops per task
- Escalate to the user when requirements are unclear or conflicting
- Stop the loop if the same issue repeats twice without progress
- Stop if the per-task budget is exceeded

## Minimal Viable Scope (MVP)

- Task types: small feature additions, localized refactors, and bug fixes
- Code size: single-file to small multi-file changes
- Languages: Go only
- Interfaces: CLI-based orchestrator with file-based handoffs
- Dependencies: local-only; no cloud services (Docker is allowed for local runs)

## Baselines and Benchmarks

- Maintain a local benchmark set of 10–20 tasks under `examples/`
- Cover: small features, bug fixes, refactors, and docs in Go
- Track expected outcomes and regression checks per task

## Rollout Plan (Phased)

1. Local CLI MVP for single-repo usage
2. Team pilot with shared configuration and logging
3. Self-hosted CI/CD integration for automated review and checks (optional)

## Risks & Mitigations

- **Cost growth**: enforce per-task budgets and early stopping rules
- **Context drift**: structured handoff templates + artifact tracking
- **Review deadlocks**: cap review loops and define escalation paths
- **Security leakage**: secret redaction and strict data handling policy

## Security and Auditability

- **Data handling**: local-only storage; never log secrets; redact `.env` values and common secret patterns
- **Prompt injection**: treat repo content as untrusted; strip embedded instructions; require explicit user confirmation for risky actions
- **Auditability**: write JSONL run logs with routing decisions, artifacts referenced by path, and input/output hashes

## Project Structure (Proposed)

```
cooperations/
├── docs/
│   ├── STRATEGY.md          # This document
│   ├── ARCHITECTURE.md      # Detailed technical design
│   └── API.md               # API contracts
├── src/
│   ├── orchestrator/        # Core orchestration logic
│   ├── agents/              # Role implementations
│   ├── adapters/            # Model API adapters
│   └── context/             # Shared context management
├── tests/
│   ├── unit/
│   └── integration/
└── examples/                # Usage examples
```

## Next Steps

1. Design orchestrator component (`/plan orchestrator`)
2. Build model adapters for both APIs
3. Implement core roles and routing rules
4. Create local benchmark tasks in `examples/`
5. Add logging/redaction and audit trails
6. Test and iterate

---

*Document created: 2026-01-30*
*Status: Strategy defined, awaiting implementation planning*
