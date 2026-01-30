# Open Questions for Refinement

This document tracks unresolved questions from the strategy that need answers before or during implementation. Update with decisions as they are made.

## Status Legend

| Status | Meaning |
|--------|---------|
| 🔴 OPEN | Needs discussion/decision |
| 🟡 IN PROGRESS | Being investigated |
| 🟢 RESOLVED | Decision made |
| ⚪ DEFERRED | Postponed to later phase |

---

## API and Access

### Q1: Codex 5.2 API

**Status:** 🟢 RESOLVED
**Priority:** Critical

**Question:** Confirm API availability, rate limits, authentication, and regional availability.

**Decision:** API access confirmed and available. No blockers.

---

### Q2: Claude Opus 4.5 API

**Status:** 🟢 RESOLVED
**Priority:** Critical

**Question:** Confirm quota constraints, latency expectations, and production readiness.

**Decision:** API access confirmed. No limitations for this experiment.

---

## Orchestration and Hosting

### Q3: Orchestration Hosting Model

**Status:** 🟢 RESOLVED
**Priority:** High

**Question:** Local CLI vs. server-based vs. CI/CD integration?

**Decision:** Local CLI.

---

### Q4: Execution Model

**Status:** 🟢 RESOLVED
**Priority:** Medium

**Question:** Long-running daemon vs. per-task execution?

**Decision:** Per-task process execution (simple, stateless).

---

### Q5: State Persistence

**Status:** 🟢 RESOLVED
**Priority:** Medium

**Question:** Where and how is run state stored (local files, DB, git commits)?

**Decision:** Local JSON files.

---

## Context and Handoffs

### Q6: Context Sharing Mechanism

**Status:** 🟢 RESOLVED
**Priority:** High

**Question:** File-based handoffs vs. shared memory vs. git commits?

**Decision:** Structured JSON files.

---

### Q7: Artifact Format

**Status:** 🟢 RESOLVED
**Priority:** High

**Question:** What is the canonical handoff schema (JSON, Markdown, YAML)?

**Decision:** JSON format with schema:
```json
{
  "task_id": "uuid",
  "timestamp": "ISO-8601",
  "from_role": "Architect",
  "to_role": "Implementer",
  "context": {
    "task_description": "...",
    "requirements": ["..."],
    "constraints": ["..."],
    "files_in_scope": ["..."]
  },
  "artifacts": {
    "design_doc": "path or inline",
    "interfaces": ["..."],
    "notes": "..."
  },
  "metadata": {
    "tokens_used": 1234,
    "model": "claude-opus-4-5",
    "duration_ms": 5000
  }
}
```

---

### Q8: Context Limits

**Status:** ⚪ DEFERRED
**Priority:** Medium

**Question:** How to compress or prune context between roles?

**Note:** Out of scope for initial experiment. Will address if context limits become a problem.

---

## Decision Policy and Disagreements

### Q9: Conflict Resolution

**Status:** ⚪ DEFERRED
**Priority:** High

**Question:** When models disagree, what's the tiebreaker?

**Note:** Out of scope for initial experiment. Draft policy exists in STRATEGY.md.

---

### Q10: Routing Policy

**Status:** ⚪ DEFERRED
**Priority:** Medium

**Question:** Rule-based vs. learned vs. hybrid; how is confidence modeled?

**Note:** Out of scope for initial experiment. Will use simple rule-based routing.

---

### Q11: Escalation Rules

**Status:** ⚪ DEFERRED
**Priority:** Medium

**Question:** When does the orchestrator hand off or stop a loop?

**Note:** Out of scope for initial experiment.

---

## Scope and Evaluation

### Q12: Initial Scope

**Status:** ⚪ DEFERRED
**Priority:** High

**Question:** Single-file tasks? Full projects? Language focus?

**Note:** Out of scope for initial experiment. Will define organically during development.

---

### Q13: Success Criteria

**Status:** ⚪ DEFERRED
**Priority:** High

**Question:** What metrics and thresholds define "better" than single-model?

**Note:** Out of scope for initial experiment. Draft criteria exist in STRATEGY.md.

---

### Q14: Baseline Tasks

**Status:** ⚪ DEFERRED
**Priority:** Medium

**Question:** What benchmark task set will be used for evaluation?

**Note:** Out of scope for initial experiment.

---

## Safety, Security, and Compliance

### Q15: Data Handling

**Status:** ⚪ DEFERRED
**Priority:** High

**Question:** How are secrets redacted and stored?

**Note:** Out of scope for initial experiment. Will address before production use.

---

### Q16: Prompt Injection

**Status:** ⚪ DEFERRED
**Priority:** High

**Question:** What defenses or filters are required?

**Note:** Out of scope for initial experiment. Will address before production use.

---

### Q17: Auditability

**Status:** ⚪ DEFERRED
**Priority:** Medium

**Question:** What logs and traces are required for compliance?

**Note:** Out of scope for initial experiment.

---

## Resolved Questions Summary

| Question | Decision |
|----------|----------|
| Q1: Codex 5.2 API | Available, confirmed |
| Q2: Claude Opus 4.5 API | Available, no limitations |
| Q3: Orchestration Hosting | Local CLI |
| Q4: Execution Model | Per-task process |
| Q5: State Persistence | Local JSON files |
| Q6: Context Sharing | Structured JSON files |
| Q7: Artifact Format | JSON with defined schema |

---

## Adding New Questions

When adding a new question:
1. Assign the next Q number (Q18, Q19, etc.)
2. Set initial status to 🔴 OPEN
3. Assess priority (Critical/High/Medium/Low)
4. Identify what it blocks
5. List options if known
6. Add action items

---

*Last updated: 2026-01-30*
