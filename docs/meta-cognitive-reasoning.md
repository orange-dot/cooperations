# Meta-Cognitive Reasoning Technique

## Overview

Meta-Cognitive Reasoning is a prompting technique based on MIT CSAIL's "Recursive Language Models" research. The technique dramatically improves AI accuracy by having the model recursively verify its own reasoning instead of answering in a single pass.

**Key Results:**
- Complex reasoning tasks: 0.04% → 58% accuracy (GPT-5)
- Research tasks: 51% → 91% accuracy
- Handles inputs 100x beyond normal context windows
- No significant cost increase

## The Problem: Context Rot

Even the best language models degrade as problems become complex:
- Model sounds confident but accuracy drops
- Single-pass answers provide no uncertainty indication
- No mechanism for self-verification
- False confidence masks actual uncertainty

## The Solution: Recursive Self-Verification

The technique mirrors how experienced engineers think:

1. **Don't solve hard problems in one pass**
2. **Decompose into verifiable sub-problems**
3. **Verify each piece independently**
4. **Combine with weighted confidence**
5. **Retry weak areas until confident**

## The 5-Step Framework

### Step 1: DECOMPOSE
Break the problem into independent sub-problems.

```
Input: Complex multi-faceted question
Output: List of atomic sub-problems that can be solved independently
```

**Guidelines:**
- Each sub-problem should be self-contained
- Identify dependencies between sub-problems
- Order by dependency (solve foundations first)

### Step 2: SOLVE
Address each sub-problem with an explicit confidence score (0.0–1.0).

```
For each sub-problem:
  - Generate solution
  - Assign confidence: 0.0 (no confidence) to 1.0 (certain)
  - Document reasoning chain
```

**Confidence Scale:**
| Score | Meaning | Action |
|-------|---------|--------|
| 0.0–0.4 | Low confidence | Reject, retry with different approach |
| 0.4–0.8 | Medium confidence | Flag uncertainty, proceed with caution |
| 0.8–1.0 | High confidence | Trust, proceed |

### Step 3: VERIFY
Check each solution for:
- **Logic**: Is the reasoning valid?
- **Facts**: Are stated facts accurate?
- **Completeness**: Are all aspects addressed?
- **Bias**: Are there hidden assumptions?

```
Verification Checklist:
[ ] Logical consistency
[ ] Factual accuracy
[ ] No missing considerations
[ ] Assumptions made explicit
[ ] Edge cases considered
```

### Step 4: SYNTHESIZE
Combine sub-solutions using weighted confidence.

```
Final Confidence = Σ(sub_confidence × weight) / Σ(weight)

Where weight is based on:
- Importance of sub-problem to final answer
- Dependencies (foundational problems weighted higher)
```

### Step 5: REFLECT
If overall confidence < 0.8:
1. Identify the weakest sub-problem
2. Analyze why confidence is low
3. Retry that specific area with alternative approach
4. Re-synthesize

```
While overall_confidence < 0.8:
    weakest = find_lowest_confidence_subproblem()
    alternative_solution = retry_with_different_approach(weakest)
    if alternative_solution.confidence > weakest.confidence:
        replace(weakest, alternative_solution)
    recalculate_overall_confidence()
```

## Prompt Template

```
Adopt the role of a Meta-Cognitive Reasoning Expert

For complex problems:
1. DECOMPOSE: Break into sub-problems
2. SOLVE: Address each with confidence (0.0–1.0)
3. VERIFY: Check logic, facts, completeness, bias
4. SYNTHESIZE: Combine using weighted confidence
5. REFLECT: If confidence <0.8, identify weakness and retry

For simple questions, answer directly.

Output: Clear answer, confidence level, key caveats

---

[Your question/problem here]
```

## Implementation Patterns

### Pattern A: Research Tasks

```markdown
## Task: [Research Question]

### Decomposition
1. Sub-question A: [specific aspect]
2. Sub-question B: [specific aspect]
3. Sub-question C: [specific aspect]

### Solutions
**A**: [Answer] | Confidence: 0.85
**B**: [Answer] | Confidence: 0.72
**C**: [Answer] | Confidence: 0.91

### Verification
- Logic: ✓ All reasoning chains valid
- Facts: ✓ Cross-referenced sources
- Completeness: ⚠ Missing edge case X
- Bias: ✓ Multiple perspectives considered

### Synthesis
Combined answer with overall confidence: 0.82

### Reflection
Confidence above threshold. Proceeding with noted caveat about edge case X.
```

### Pattern B: Technical Analysis

```markdown
## Problem: [Technical Challenge]

### Decomposition
1. Architecture considerations
2. Performance implications
3. Security concerns
4. Maintainability factors

### Per-Component Analysis
[Each with confidence score and verification]

### Synthesis Matrix
| Component | Confidence | Weight | Weighted |
|-----------|------------|--------|----------|
| Architecture | 0.9 | 0.3 | 0.27 |
| Performance | 0.7 | 0.25 | 0.175 |
| Security | 0.85 | 0.25 | 0.2125 |
| Maintainability | 0.8 | 0.2 | 0.16 |
| **Total** | | | **0.8175** |

### Final Recommendation
[Synthesized answer with confidence 0.82]
```

### Pattern C: Strategy Decisions

```markdown
## Decision: [Strategic Choice]

### Option Analysis
For each option:
- Pros (with confidence)
- Cons (with confidence)
- Risk assessment (with confidence)

### Cross-Verification
- Market assumptions verified: [Y/N + confidence]
- Technical feasibility verified: [Y/N + confidence]
- Resource requirements verified: [Y/N + confidence]

### Weighted Recommendation
Option X recommended with confidence 0.78

### Reflection Loop
Confidence below 0.8. Investigating uncertainty in [specific area].
[Re-analysis of weak point]
Updated confidence: 0.83
```

## Adaptive Complexity

The framework scales automatically:

| Question Type | Response |
|---------------|----------|
| Simple factual | Direct answer |
| Moderate complexity | 2-3 step decomposition |
| High complexity | Full 5-step framework |
| Novel/uncertain | Multiple iterations with reflection |

**Complexity Detection Signals:**
- Multiple interacting factors
- Requires domain expertise
- Has significant consequences
- Contains ambiguity
- Needs multi-step reasoning

## Best Practices

### Do:
- Be explicit about confidence levels
- Flag uncertainty rather than hiding it
- Use reflection loops for weak areas
- Document reasoning chains
- Identify assumptions

### Don't:
- Over-engineer simple questions
- Accept low-confidence answers as final
- Skip verification steps
- Ignore bias checking
- Combine solutions without weighting

## Integration with Cooperations

This technique can be integrated into the orchestrator workflow:

```go
type MetaCognitiveStep struct {
    Decomposition []SubProblem
    Solutions     []SolutionWithConfidence
    Verification  VerificationResult
    Synthesis     SynthesisResult
    Reflection    *ReflectionLoop // nil if confidence >= 0.8
}

type SolutionWithConfidence struct {
    Content    string
    Confidence float64 // 0.0 to 1.0
    Reasoning  string
}

func (o *Orchestrator) RunWithMetaCognition(ctx context.Context, task string) {
    // 1. Decompose via Architect
    // 2. Solve sub-problems via Implementer
    // 3. Verify via Reviewer
    // 4. Synthesize results
    // 5. Reflect and retry if needed
}
```

## References

- MIT CSAIL: "Recursive Language Models" paper
- Psychology: Meta-cognition research (Flavell, 1979)
- Engineering: Structured problem decomposition methods
