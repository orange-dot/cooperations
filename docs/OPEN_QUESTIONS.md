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

**Status:** 🟢 RESOLVED
**Priority:** Medium

**Question:** How to compress or prune context between roles?

**Decision:** Use structured handoff summaries with strict token budgets. Persist artifacts on disk and reference by path, keep last two handoffs verbatim, and summarize older context into a short decision + blockers list.

---

## Decision Policy and Disagreements

### Q9: Conflict Resolution

**Status:** 🟢 RESOLVED
**Priority:** High

**Question:** When models disagree, what's the tiebreaker?

**Decision:** Use a safety-first, role-based tiebreaker policy. Reviewer has veto on safety/security, Architect decides architecture, Implementer decides implementation details, Navigator routes requirement ambiguities to the user. If evidence is inconclusive, prefer the safer/simpler option and log the rationale.

---

### Q10: Routing Policy

**Status:** 🟢 RESOLVED
**Priority:** Medium

**Question:** Rule-based vs. learned vs. hybrid; how is confidence modeled?

**Decision:** Rule-based routing only (no learned components). Confidence is a simple heuristic using task size, file count, and risk keywords; low confidence triggers an Architect pass or a user clarification.

---

### Q11: Escalation Rules

**Status:** 🟢 RESOLVED
**Priority:** Medium

**Question:** When does the orchestrator hand off or stop a loop?

**Decision:** Limit to two review loops per task. Escalate to user when requirements are ambiguous or conflicts persist; stop when the same issue repeats twice without progress or budget is exceeded.

---

## Scope and Evaluation

### Q12: Initial Scope

**Status:** 🟢 RESOLVED
**Priority:** High

**Question:** Single-file tasks? Full projects? Language focus?

**Decision:** Local, single-repo tasks; single-file to small multi-file changes. Initial language focus: Go. No cloud dependencies.

---

### Q13: Success Criteria

**Status:** 🟢 RESOLVED
**Priority:** High

**Question:** What metrics and thresholds define "better" than single-model?

**Decision:** Targets are: >= 80% successful completion on the benchmark set, >= 20% median time-to-completion improvement vs. single-model baseline, and <= 1.2x cost vs. baseline. All measured locally with reproducible logs.

---

### Q14: Baseline Tasks

**Status:** 🟢 RESOLVED
**Priority:** Medium

**Question:** What benchmark task set will be used for evaluation?

**Decision:** Create a local benchmark set of 10-20 tasks covering: small features, bug fixes, refactors, and documentation updates in Go. Store as fixtures in `examples/` with expected outcomes.

---

## Safety, Security, and Compliance

### Q15: Data Handling

**Status:** 🟢 RESOLVED
**Priority:** High

**Question:** How are secrets redacted and stored?

**Decision:** Local-only storage. Secrets are never logged; redact known patterns and `.env` values before sending to models. Allowlist file access for context injection.

---

### Q16: Prompt Injection

**Status:** 🟢 RESOLVED
**Priority:** High

**Question:** What defenses or filters are required?

**Decision:** Treat repository content as untrusted. Strip instructions from file content, separate system directives, and require explicit user confirmation before executing risky operations. Log and flag prompt-injection indicators.

---

### Q17: Auditability

**Status:** 🟢 RESOLVED
**Priority:** Medium

**Question:** What logs and traces are required for compliance?

**Decision:** Write local JSONL logs per run with routing decisions, model outputs, artifacts referenced by path, and a hash of inputs/outputs to support reproducibility.

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
| Q8: Context Limits | Handoff summaries + token budgets |
| Q9: Conflict Resolution | Role-based tiebreaker + safety-first veto |
| Q10: Routing Policy | Rule-based + heuristic confidence |
| Q11: Escalation Rules | Two review loops + stop/escalate |
| Q12: Initial Scope | Local small tasks; Go only |
| Q13: Success Criteria | 80% success, 20% faster, <= 1.2x cost |
| Q14: Baseline Tasks | Local benchmark set in examples/ |
| Q15: Data Handling | Local-only + secret redaction |
| Q16: Prompt Injection | Treat repo content as untrusted |
| Q17: Auditability | Local JSONL logs + hashes |

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
