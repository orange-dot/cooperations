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
- Median time-to-completion improves by ≥ 25% vs. single-model baseline
- Cost per task is within a defined budget cap
- Audit logs are complete and reproducible for each run

## Orchestrator Decision Policy

- **Default**: rule-based routing using task type, size, and risk classification
- **Adaptation**: allow mid-task handoffs when confidence drops or blockers persist
- **Fallback**: when disagreement persists, escalate to Reviewer or Architect
- **Transparency**: every routing decision is logged with its rationale

## Conflict Resolution Policy (Draft)

| Situation | Primary Tiebreaker | Secondary |
|----------|--------------------|-----------|
| Safety or security concern | Reviewer | Architect |
| Architectural disagreement | Architect | Reviewer |
| Implementation detail | Implementer | Reviewer |
| Style or formatting | Project standards | Reviewer |
| Unclear requirements | Navigator | User |

## Minimal Viable Scope (MVP)

- Task types: small feature additions, localized refactors, and bug fixes
- Code size: single-file to small multi-file changes
- Languages: define 1–2 initial languages (e.g., TypeScript, Python)
- Interfaces: CLI-based orchestrator with file-based handoffs

## Rollout Plan (Phased)

1. Local CLI MVP for single-repo usage
2. Team pilot with shared configuration and logging
3. CI/CD integration for automated review and checks

## Risks & Mitigations

- **Cost growth**: enforce per-task budgets and early stopping rules
- **Context drift**: structured handoff templates + artifact tracking
- **Review deadlocks**: cap review loops and define escalation paths
- **Security leakage**: secret redaction and strict data handling policy

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

1. Resolve open questions (especially Codex API access)
2. Design orchestrator component (`/plan orchestrator`)
3. Build model adapters for both APIs
4. Implement core roles
5. Create example workflows
6. Test and iterate

---

*Document created: 2026-01-30*
*Status: Strategy defined, awaiting implementation planning*
