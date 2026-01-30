# Recursive Verified Reasoning (RVR)

> A hybrid approach combining REPL-based recursion with explicit verification and confidence scoring.

## Inspiration

### MIT CSAIL: Recursive Language Models (RLM)
- **Paper:** arXiv:2512.24601v2
- **Key idea:** Prompt as a variable in REPL environment, not in LLM context
- **Mechanism:** Model writes code that recursively calls sub-LLM in loops
- **Result:** Scaling to 10M+ tokens, O(n²) semantic work

### Dr. Milan Milanovic (Twitter)
- **Key idea:** Explicit meta-cognition with confidence scores
- **Mechanism:** VERIFY step that checks logic, facts, bias
- **Result:** Retry mechanism when confidence drops below threshold

## Why Combine?

| Aspect | RLM | Dr. Milan Milanovic | RVR (proposed concept) |
|--------|-----|---------------------|------------------------|
| Scaling | 10M+ tokens | Limited by context | 10M+ tokens (theoretical) |
| Verification | Implicit (model decides) | Explicit (confidence) | Explicit + recursive (idea) |
| Retry | None | When confidence < 0.8 | Configurable per task type (idea) |
| Efficiency | High | Medium (double calls) | High (two-layer approach, theoretical) |

> **Note:** RLM and Dr. Milan Milanovic columns reflect documented results. RVR column represents proposed ideas combining both approaches - not yet validated.

---

## Architecture: Two Layers

```
┌─────────────────────────────────────────────────────────────────┐
│                         INPUT                                    │
│              (Prompt P, Task Config T)                          │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    LAYER 1: Fast Processing                      │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  for chunk in context:                                      │ │
│  │      result, confidence = llm_query_with_confidence(chunk)  │ │
│  │      buffer.append({result, confidence, chunk_id})          │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                  │
│  Output: Buffer with all results and confidence scores          │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    CONFIDENCE TRIAGE                             │
│  ┌──────────────┬──────────────┬──────────────┐                 │
│  │   CRITICAL   │     LOW      │    HIGH      │                 │
│  │   (< 0.4)    │  (0.4-0.8)   │   (≥ 0.8)    │                 │
│  │              │              │              │                 │
│  │  Retry with  │   Verify     │   Pass       │                 │
│  │  alternative │   specific   │   through    │                 │
│  │  approach    │   dimensions │              │                 │
│  └──────────────┴──────────────┴──────────────┘                 │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                 LAYER 2: Selective Verification                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  for item in buffer where confidence < threshold:           │ │
│  │      if confidence < critical_threshold:                    │ │
│  │          item = retry_with_alternative(item)                │ │
│  │      else:                                                  │ │
│  │          item = verify_dimensions(item, task.verify_fields) │ │
│  │      update_confidence(item)                                │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    SYNTHESIS                                     │
│  Combine results with weighted confidence                        │
│  Final confidence = Σ(result.confidence × weight) / Σ(weight)   │
└─────────────────────────────────────────────────────────────────┘
```

---

## Task Type Configuration

### Configuration Structure

```yaml
rvr_config:
  # Global defaults
  defaults:
    confidence_threshold: 0.8
    critical_threshold: 0.4
    retry_attempts: 2
    batch_size: 10  # Chunks per sub-LLM call

  # Task-specific configurations
  tasks:
    research:
      description: "Research and information synthesis"
      confidence_threshold: 0.7
      critical_threshold: 0.3
      retry_attempts: 2
      verify_fields:
        - facts
        - sources
        - completeness
      retry_strategies:
        - "rephrase_query"
        - "expand_context"

    code_generation:
      description: "Code generation and modification"
      confidence_threshold: 0.85
      critical_threshold: 0.5
      retry_attempts: 3
      verify_fields:
        - logic
        - syntax
        - security
        - edge_cases
      retry_strategies:
        - "step_by_step"
        - "test_driven"
        - "simplify"

    code_review:
      description: "Review and analysis of existing code"
      confidence_threshold: 0.75
      critical_threshold: 0.4
      retry_attempts: 2
      verify_fields:
        - logic
        - security
        - performance
        - maintainability
      retry_strategies:
        - "focus_on_critical"
        - "compare_patterns"

    decision_making:
      description: "Strategic decisions and trade-off analysis"
      confidence_threshold: 0.9
      critical_threshold: 0.6
      retry_attempts: 1
      verify_fields:
        - logic
        - bias
        - completeness
        - alternatives
      retry_strategies:
        - "devils_advocate"
        - "seek_counterexamples"

    summarization:
      description: "Summarizing long texts"
      confidence_threshold: 0.7
      critical_threshold: 0.4
      retry_attempts: 2
      verify_fields:
        - completeness
        - accuracy
      retry_strategies:
        - "chunk_smaller"
        - "hierarchical"

    translation:
      description: "Translation between languages or formats"
      confidence_threshold: 0.8
      critical_threshold: 0.5
      retry_attempts: 2
      verify_fields:
        - accuracy
        - fluency
        - terminology
      retry_strategies:
        - "back_translate"
        - "terminology_check"
```

### Custom Task Definition

```yaml
# Users can define custom tasks
tasks:
  my_custom_task:
    description: "My specific use case"
    confidence_threshold: 0.75
    critical_threshold: 0.35
    retry_attempts: 4
    verify_fields:
      - custom_field_1
      - custom_field_2
    retry_strategies:
      - "my_strategy"

    # Custom verification prompts
    verification_prompts:
      custom_field_1: |
        Verify that the result satisfies custom requirement X.
        Rate confidence 0.0-1.0.
      custom_field_2: |
        Check if Y condition is met.
        Rate confidence 0.0-1.0.
```

---

## Verification Dimensions

| Dimension | Description | Prompt hint | Typical tasks |
|-----------|-------------|-------------|---------------|
| `logic` | Does conclusion follow from premises | "Is the reasoning logically valid?" | code, decision |
| `facts` | Are stated facts accurate | "Are the stated facts accurate?" | research |
| `sources` | Are sources reliable | "Are the sources reliable and cited?" | research |
| `completeness` | Are all aspects covered | "Are all aspects of the question addressed?" | all |
| `bias` | Are there hidden assumptions | "Are there hidden assumptions or biases?" | decision |
| `syntax` | Is code syntactically correct | "Is the code syntactically correct?" | code |
| `security` | Are there security vulnerabilities | "Are there security vulnerabilities?" | code, review |
| `edge_cases` | Are edge cases handled | "Are edge cases handled?" | code |
| `performance` | Are there performance issues | "Are there performance concerns?" | review |
| `maintainability` | Is code maintainable | "Is the code maintainable?" | review |
| `accuracy` | Is the result accurate | "Is the result accurate?" | translation, summary |
| `fluency` | Is text fluent | "Is the text fluent and natural?" | translation |
| `alternatives` | Were alternatives considered | "Were alternatives considered?" | decision |

---

## Algorithm

### Implementation (Go)

```go
// internal/rvr/types.go
package rvr

type BufferItem struct {
    ChunkID       int                    `json:"chunk_id"`
    Result        string                 `json:"result"`
    Confidence    float64                `json:"confidence"`
    Uncertainty   string                 `json:"uncertainty"`
    OriginalChunk string                 `json:"original_chunk"`
    RetryStrategy string                 `json:"retry_strategy,omitempty"`
    Verifications map[string]Verification `json:"verifications,omitempty"`
}

type Verification struct {
    Dimension  string  `json:"dimension"`
    Valid      string  `json:"valid"` // "yes", "no", "partial"
    Confidence float64 `json:"confidence"`
    Issues     string  `json:"issues"`
}

type RVRResult struct {
    Answer     string       `json:"answer"`
    Confidence float64      `json:"confidence"`
    Caveats    []string     `json:"caveats"`
    Breakdown  []BufferItem `json:"breakdown"`
}

type TriageResult struct {
    Critical []BufferItem
    Low      []BufferItem
    High     []BufferItem
}
```

```go
// internal/rvr/processor.go
package rvr

import (
    "context"
    "fmt"
    "strings"
)

type Processor struct {
    config RVRConfig
    cli    CLI
}

func NewProcessor(config RVRConfig, cli CLI) *Processor {
    return &Processor{config: config, cli: cli}
}

func (p *Processor) Process(ctx context.Context, prompt, taskType string) (*RVRResult, error) {
    task := p.getTaskConfig(taskType)

    // LAYER 1: Fast Processing
    buffer, err := p.layer1Process(ctx, prompt, task)
    if err != nil {
        return nil, fmt.Errorf("layer1: %w", err)
    }

    // Triage by confidence
    triage := p.triage(buffer, task)

    // LAYER 2: Selective Verification
    verified, err := p.layer2Verify(ctx, triage, task)
    if err != nil {
        return nil, fmt.Errorf("layer2: %w", err)
    }

    // Synthesis
    return p.synthesize(ctx, verified, task)
}

func (p *Processor) layer1Process(ctx context.Context, prompt string, task TaskConfig) ([]BufferItem, error) {
    chunks := p.chunkContext(prompt, task.BatchSize*1000)
    buffer := make([]BufferItem, 0, len(chunks))

    for i, chunk := range chunks {
        query := fmt.Sprintf(`Task: %s
Context: %s

Respond in this format:
ANSWER: <your answer>
CONFIDENCE: <0.0-1.0>
UNCERTAINTY: <what you're unsure about, if any>`, task.Description, chunk)

        resp, err := p.cli.Execute(ctx, query)
        if err != nil {
            return nil, err
        }

        parsed := p.parseResponse(resp.Content)
        buffer = append(buffer, BufferItem{
            ChunkID:       i,
            Result:        parsed.Answer,
            Confidence:    parsed.Confidence,
            Uncertainty:   parsed.Uncertainty,
            OriginalChunk: chunk,
        })
    }

    return buffer, nil
}

func (p *Processor) triage(buffer []BufferItem, task TaskConfig) TriageResult {
    var result TriageResult

    for _, item := range buffer {
        switch {
        case item.Confidence < task.CriticalThreshold:
            result.Critical = append(result.Critical, item)
        case item.Confidence < task.ConfidenceThreshold:
            result.Low = append(result.Low, item)
        default:
            result.High = append(result.High, item)
        }
    }

    return result
}

func (p *Processor) layer2Verify(ctx context.Context, triage TriageResult, task TaskConfig) ([]BufferItem, error) {
    verified := make([]BufferItem, 0)

    // High confidence items pass through
    verified = append(verified, triage.High...)

    // Critical items: retry with alternative approach
    for _, item := range triage.Critical {
        retried := item
        for _, strategy := range task.RetryStrategies {
            newResult, err := p.retryWithStrategy(ctx, item, strategy, task)
            if err != nil {
                continue
            }
            if newResult.Confidence >= task.CriticalThreshold {
                retried = *newResult
                break
            }
        }
        verified = append(verified, retried)
    }

    // Low confidence items: verify specific dimensions
    for _, item := range triage.Low {
        item.Verifications = make(map[string]Verification)
        for _, dimension := range task.VerifyFields {
            v, err := p.verifyDimension(ctx, item, dimension)
            if err != nil {
                continue
            }
            item.Confidence = (item.Confidence + v.Confidence) / 2
            item.Verifications[dimension] = *v
        }
        verified = append(verified, item)
    }

    return verified, nil
}

func (p *Processor) verifyDimension(ctx context.Context, item BufferItem, dimension string) (*Verification, error) {
    contextPreview := item.OriginalChunk
    if len(contextPreview) > 500 {
        contextPreview = contextPreview[:500] + "..."
    }

    query := fmt.Sprintf(`Original question context: %s
Answer to verify: %s

Dimension to verify: %s

Evaluate this dimension and respond:
VALID: <yes/no/partial>
CONFIDENCE: <0.0-1.0>
ISSUES: <any issues found, or "none">`, contextPreview, item.Result, dimension)

    resp, err := p.cli.Execute(ctx, query)
    if err != nil {
        return nil, err
    }

    parsed := p.parseVerification(resp.Content)
    return &Verification{
        Dimension:  dimension,
        Valid:      parsed.Valid,
        Confidence: parsed.Confidence,
        Issues:     parsed.Issues,
    }, nil
}

var retryStrategies = map[string]string{
    "rephrase_query":      "Rephrase the question differently",
    "expand_context":      "Consider broader context",
    "step_by_step":        "Break down into smaller steps",
    "simplify":            "Simplify the approach",
    "devils_advocate":     "Argue the opposite position first",
    "seek_counterexamples": "Find counterexamples before concluding",
}

func (p *Processor) retryWithStrategy(ctx context.Context, item BufferItem, strategy string, task TaskConfig) (*BufferItem, error) {
    strategyDesc := retryStrategies[strategy]
    if strategyDesc == "" {
        strategyDesc = strategy
    }

    query := fmt.Sprintf(`Previous attempt had low confidence.
Strategy: %s

Original context: %s
Previous answer: %s
Previous confidence: %.2f
Uncertainty: %s

Try again with the suggested strategy.

ANSWER: <new answer>
CONFIDENCE: <0.0-1.0>
UNCERTAINTY: <remaining uncertainties>`, strategyDesc, item.OriginalChunk, item.Result, item.Confidence, item.Uncertainty)

    resp, err := p.cli.Execute(ctx, query)
    if err != nil {
        return nil, err
    }

    parsed := p.parseResponse(resp.Content)
    return &BufferItem{
        ChunkID:       item.ChunkID,
        Result:        parsed.Answer,
        Confidence:    parsed.Confidence,
        Uncertainty:   parsed.Uncertainty,
        OriginalChunk: item.OriginalChunk,
        RetryStrategy: strategy,
    }, nil
}

func (p *Processor) synthesize(ctx context.Context, verified []BufferItem, task TaskConfig) (*RVRResult, error) {
    // Weighted average confidence
    var totalWeight, weightedSum float64
    for _, item := range verified {
        totalWeight += item.Confidence
        weightedSum += item.Confidence * item.Confidence
    }

    overallConfidence := 0.0
    if totalWeight > 0 {
        overallConfidence = weightedSum / totalWeight
    }

    // Format results for synthesis
    var resultsBuilder strings.Builder
    for _, item := range verified {
        fmt.Fprintf(&resultsBuilder, "- [Confidence: %.2f] %s\n", item.Confidence, item.Result)
    }

    query := fmt.Sprintf(`Synthesize these partial results into a final answer:

%s

Weight higher-confidence results more heavily.

FINAL_ANSWER: <synthesized answer>
OVERALL_CONFIDENCE: <0.0-1.0>
CAVEATS: <important caveats or limitations>`, resultsBuilder.String())

    resp, err := p.cli.Execute(ctx, query)
    if err != nil {
        return nil, err
    }

    parsed := p.parseSynthesis(resp.Content)
    return &RVRResult{
        Answer:     parsed.FinalAnswer,
        Confidence: overallConfidence,
        Caveats:    parsed.Caveats,
        Breakdown:  verified,
    }, nil
}

func (p *Processor) chunkContext(content string, chunkSize int) []string {
    if len(content) <= chunkSize {
        return []string{content}
    }

    var chunks []string
    for i := 0; i < len(content); i += chunkSize {
        end := i + chunkSize
        if end > len(content) {
            end = len(content)
        }
        chunks = append(chunks, content[i:end])
    }
    return chunks
}

func (p *Processor) getTaskConfig(taskType string) TaskConfig {
    if task, ok := p.config.Tasks[taskType]; ok {
        return task
    }
    return p.config.Defaults
}
```

---

## Usage Examples

### Example 1: Research Task

```go
func main() {
    config := rvr.LoadConfig("config.yaml")
    cli := adapters.NewClaudeCLI()
    processor := rvr.NewProcessor(config, cli)

    result, err := processor.Process(
        context.Background(),
        largeDocumentCorpus,  // 5M tokens
        "research",
    )
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Answer: %s\n", result.Answer)
    fmt.Printf("Confidence: %.2f\n", result.Confidence)
    fmt.Printf("Caveats: %v\n", result.Caveats)

    // Output:
    // Answer: Based on analysis of 1,247 documents...
    // Confidence: 0.82
    // Caveats: [Some sources from 2019 may be outdated]
}
```

### Example 2: Code Generation with High Standards

```go
result, err := processor.Process(
    ctx,
    "Implement a thread-safe cache with LRU eviction",
    "code_generation",
)

// Threshold is 0.85, so more chunks go to Layer 2
// Verify fields: logic, syntax, security, edge_cases

for _, item := range result.Breakdown {
    if len(item.Verifications) > 0 {
        fmt.Printf("Chunk %d verified: %v\n", item.ChunkID, item.Verifications)
    }
}
```

### Example 3: Custom Task

```yaml
# config.yaml
tasks:
  legal_review:
    confidence_threshold: 0.95  # High standard for legal documents
    critical_threshold: 0.7
    retry_attempts: 3
    verify_fields:
      - legal_accuracy
      - jurisdiction
      - precedent
    verification_prompts:
      legal_accuracy: "Is this legally accurate per current law?"
      jurisdiction: "Is the jurisdiction correctly identified?"
      precedent: "Are relevant precedents cited?"
```

```go
result, err := processor.Process(ctx, contractText, "legal_review")
if err != nil {
    log.Fatal(err)
}

if result.Confidence < 0.95 {
    log.Printf("Warning: Legal review confidence below threshold: %.2f", result.Confidence)
    for _, item := range result.Breakdown {
        if item.Confidence < 0.7 {
            log.Printf("  Low confidence section: %s", item.Result[:100])
        }
    }
}
```

---

## Integration with Cooperations Orchestrator

```go
// internal/rvr/config.go
type TaskConfig struct {
    Description         string   `yaml:"description"`
    ConfidenceThreshold float64  `yaml:"confidence_threshold"`
    CriticalThreshold   float64  `yaml:"critical_threshold"`
    RetryAttempts       int      `yaml:"retry_attempts"`
    VerifyFields        []string `yaml:"verify_fields"`
    RetryStrategies     []string `yaml:"retry_strategies"`
}

type RVRConfig struct {
    Defaults TaskConfig            `yaml:"defaults"`
    Tasks    map[string]TaskConfig `yaml:"tasks"`
}

// internal/rvr/processor.go
type RVRProcessor struct {
    config  RVRConfig
    cli     adapters.CLI
}

func (p *RVRProcessor) Process(ctx context.Context, prompt string, taskType string) (*RVRResult, error) {
    task := p.getTaskConfig(taskType)

    // Layer 1: Fast processing
    buffer, err := p.layer1Process(ctx, prompt, task)
    if err != nil {
        return nil, err
    }

    // Triage
    critical, low, high := p.triage(buffer, task)

    // Layer 2: Selective verification
    verified, err := p.layer2Verify(ctx, critical, low, high, task)
    if err != nil {
        return nil, err
    }

    // Synthesize
    return p.synthesize(ctx, verified, task)
}

// internal/agents/architect.go - integration
func (a *ArchitectAgent) Execute(ctx context.Context, handoff types.Handoff) (types.AgentResponse, error) {
    // Use RVR for complex design decisions
    rvr := rvr.NewProcessor(a.rvrConfig, a.cli)

    result, err := rvr.Process(ctx, handoff.Payload, "decision_making")
    if err != nil {
        return types.AgentResponse{}, err
    }

    // Include confidence in response
    return types.AgentResponse{
        Content:    result.Answer,
        Confidence: result.Confidence,
        Caveats:    result.Caveats,
        Artifacts: map[string]any{
            "rvr_breakdown": result.Breakdown,
        },
    }, nil
}
```

---

## Metrics and Monitoring

```yaml
# Recommended metrics for tracking RVR performance
metrics:
  - name: layer1_pass_rate
    description: "% of items passing Layer 1 without verification"
    target: "> 60%"  # If lower, threshold may be too high

  - name: critical_rate
    description: "% of items falling into critical category"
    target: "< 10%"  # If higher, task may be too difficult

  - name: retry_success_rate
    description: "% of retry attempts that raise confidence above threshold"
    target: "> 50%"

  - name: avg_confidence_lift
    description: "Average confidence increase after Layer 2"
    target: "> 0.15"

  - name: verification_agreement
    description: "% of verifications where model confirms original answer"
    target: "70-90%"  # Too low = poor model, too high = rubber stamp
```

---

## Conclusion

RVR combines the best of both worlds:

1. **From RLM:** Scalability, REPL approach, recursive sub-calls
2. **From Milan:** Explicit verification, confidence scores, retry logic

The key advantage of the two-layer approach is **efficiency** - only items that truly need additional attention go through the more expensive verification process.

---

## References

- **RLM:** MIT CSAIL, arXiv:2512.24601v2 "Recursive Language Models"
- **Verification approach:** Dr. Milan Milanovic (Twitter)
- **Combined concept:** RVR (Recursive Verified Reasoning)
