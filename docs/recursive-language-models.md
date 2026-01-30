# Recursive Language Models (RLMs)

> Based on MIT CSAIL paper: "Recursive Language Models" (arXiv:2512.24601v2)
> Authors: MIT CSAIL research team
> Models tested: GPT-5, Qwen3-Coder-480B-A35B
>
> **Source:** Dr Milan Milanovic (Twitter) - originally shared this research

## Overview

Recursive Language Models (RLMs) are a general-purpose inference paradigm that dramatically scales the effective input and output lengths of LLMs by treating the prompt as **part of the environment** rather than feeding it directly into the neural network.

**Key Results:**
| Task | Standard LLM | RLM | Improvement |
|------|-------------|-----|-------------|
| BrowseComp+ (1K docs) | 0% (context limit) | 91.3% | N/A |
| OOLONG | 44% | 56.5% | +28.4% |
| OOLONG-Pairs | 0.04% | 58% | +1450x |
| CodeQA | 24% | 62% | +158% |

- Handles inputs **100x beyond context window** (10M+ tokens)
- Comparable or lower costs than standard approaches
- Task-agnostic - works across domains without modification

## The Problem: Context Rot

**Context rot** is the phenomenon where LLM quality degrades steeply as prompts get longer, even within stated context limits.

Key insight from the paper:
> "The effective context window of an LLM cannot be understood independently of the specific task. More complex problems exhibit degradation at even shorter lengths than simpler ones."

**Task complexity scaling:**
- **O(1)** - Needle-in-haystack: finding one thing in large text
- **O(n)** - OOLONG: processing every line of input
- **O(n²)** - OOLONG-Pairs: processing all pairs of entries

Frontier models fail catastrophically on O(n) and O(n²) tasks even within their context windows.

## The Solution: Symbolic Recursion

### Core Insight

> "Arbitrarily long user prompts should not be fed into the neural network directly but should instead be treated as part of the environment that the LLM is tasked to **symbolically and recursively** interact with."

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    RLM Interface                             │
│  Input: Prompt P (arbitrary length)                          │
│  Output: Response Y (arbitrary length)                       │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    REPL Environment                          │
│  • context = P (as variable, not in LLM context)            │
│  • llm_query() function for sub-RLM calls                   │
│  • Variables persist across iterations                       │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    RLM Loop                                  │
│  1. LLM receives METADATA about prompt (length, prefix)     │
│  2. LLM generates CODE to interact with prompt              │
│  3. REPL executes code, returns metadata about output       │
│  4. Repeat until LLM sets Final variable                    │
└─────────────────────────────────────────────────────────────┘
```

## Algorithm

```python
def RLM(prompt: str, model: LLM) -> str:
    # Initialize REPL with prompt as variable (NOT in LLM context)
    state = InitREPL(prompt=prompt)

    # Add function for recursive sub-LLM calls
    state = AddFunction(state, sub_RLM)

    # LLM only sees METADATA (length, short prefix), not full prompt
    history = [Metadata(state)]

    while True:
        # LLM generates code to interact with prompt
        code = LLM(model, history)

        # Execute code in REPL (may include sub-RLM calls in loops)
        state, stdout = REPL(state, code)

        # Only add METADATA about output to history
        history = history + code + Metadata(stdout)

        # Return when LLM sets Final variable
        if state["Final"] is set:
            return state["Final"]
```

## Three Critical Design Choices

### 1. Symbolic Handle to Prompt

The prompt P is stored as a **variable in the REPL**, not placed in the LLM context window.

❌ **Bad (traditional):** Put prompt in context → inherits window limitations
✅ **Good (RLM):** Prompt as variable → model manipulates without copying

### 2. Symbolic Output

The output Y is built up in **REPL variables**, not generated autoregressively.

❌ **Bad:** Generate output directly → limited by output token limits
✅ **Good:** Build output in variables → unbounded output length

### 3. Symbolic Recursion

Code running inside the REPL can invoke the LLM on **programmatically constructed** inputs.

❌ **Bad:** Verbalize sub-tasks explicitly → limited delegation
✅ **Good:** Write code with loops calling LLM → O(n) or O(n²) semantic work

```python
# Example: Process every line with a sub-LLM call
for i, line in enumerate(context.split('\n')):
    result = llm_query(f"Classify this line: {line}")
    results.append(result)
```

## System Prompt (GPT-5 Version)

```
You are tasked with answering a query with associated context. You can access,
transform, and analyze this context interactively in a REPL environment that
can recursively query sub-LLMs, which you are strongly encouraged to use as
much as possible.

Your context is a {context_type} with {context_total_length} total characters.

The REPL environment is initialized with:
1. A `context` variable containing extremely important information about your
   query. Make sure you look through it sufficiently.
2. A `llm_query` function that allows you to query an LLM (can handle ~500K
   chars) inside your REPL environment.
3. The ability to use `print()` statements to view output and continue reasoning.

You will only see truncated outputs from the REPL, so use the query LLM function
on variables you want to analyze. Use these variables as buffers to build up
your final answer.

Example strategy: First look at context and figure out a chunking strategy,
break into smart chunks, query an LLM per chunk with a question, save answers
to a buffer, then query an LLM with all buffers to produce final answer.

When you want to execute Python code, wrap it in triple backticks with 'repl':
```repl
chunk = context[:10000]
answer = llm_query(f"What is the answer? Context: {chunk}")
print(answer)
```

IMPORTANT: When done, provide final answer using:
1. FINAL(your answer here) - provide answer directly
2. FINAL_VAR(variable_name) - return a variable as output
```

## Emergent Patterns

RLMs exhibit interesting behavior without explicit training:

### 1. Filtering with Model Priors
```python
# Use regex based on model's knowledge to filter relevant sections
import re
sections = re.findall(r'festival.*La Union', context)
```

### 2. Chunking and Recursive Sub-calls
```python
# Chunk by structure and process each with sub-LLM
for section in context.split('### '):
    summary = llm_query(f"Summarize: {section}")
    summaries.append(summary)
```

### 3. Building Long Outputs via Variables
```python
# Accumulate results in variables, not in LLM output
results = []
for pair in all_pairs:
    result = llm_query(f"Compare: {pair}")
    results.append(result)
final_output = "\n".join(results)  # Can be arbitrarily long
```

## Cost Analysis

| Method | Median Cost | 95th Percentile | Performance |
|--------|-------------|-----------------|-------------|
| Base LLM | $0.14 | $0.16 | 44% |
| Summary Agent | $0.13 | $0.14 | 46% |
| RLM | $0.43 | $0.85 | **56.5%** |

Key observations:
- RLM median cost comparable to base model
- High variance due to trajectory length differences
- Up to 3x cheaper than summary agents on large inputs
- Selective context viewing reduces total tokens processed

## When to Use RLMs

### Best For:
- **Long context** beyond model limits (10M+ tokens)
- **Information-dense tasks** requiring processing of most/all input
- **Complex reasoning** with O(n) or O(n²) semantic work
- **Long output generation** beyond output token limits

### Not Ideal For:
- **Short contexts** within model limits (slight overhead)
- **Simple tasks** like needle-in-haystack
- **Latency-critical** applications (iterative = slower)

## Implementation Considerations

### Sub-model Selection
- GPT-5 experiments used GPT-5-mini for sub-calls
- Balance capability vs cost for recursive calls

### Batching Sub-calls
```
IMPORTANT for Qwen3-Coder: Be careful about using `llm_query` as it incurs
high runtime costs. Always batch as much information as reasonably possible
into each call (aim for ~200k characters per call).
```

### Context Window Awareness
- Sub-LLMs have their own context limits
- Design chunks to fit comfortably in sub-model windows

## Fine-tuning RLMs

The paper shows that fine-tuning on RLM trajectories improves performance even on unrelated domains:

| Model | Base RLM | Fine-tuned RLM | Improvement |
|-------|----------|----------------|-------------|
| Qwen3-8B (CodeQA) | 26% | 32% | +23% |
| Qwen3-8B (BrowseComp+) | 2% | 14% | +600% |
| Qwen3-8B (OOLONG) | 24% | 32% | +33% |

Training insight:
> "Being an effective sub-call model is roughly similar to being a general purpose reasoning model, so we can make training much more tractable at small scale by focusing on improving the root model's ability to manipulate the REPL and launch recursive calls."

## Comparison with Other Approaches

| Approach | Handles Long Input | Long Output | Dense Processing |
|----------|-------------------|-------------|------------------|
| Base LLM | ❌ Context limit | ❌ Output limit | ❌ Degradation |
| Context Compaction | ⚠️ Lossy | ❌ Output limit | ❌ Forgets details |
| Retrieval (BM25) | ⚠️ Sparse | ❌ Output limit | ❌ Can't process all |
| CodeAct | ❌ Context limit | ❌ Output limit | ⚠️ Limited |
| CodeAct + Sub-calls | ❌ Context limit | ❌ Output limit | ⚠️ Limited |
| **RLM** | ✅ Unbounded | ✅ Unbounded | ✅ O(n²) capable |

## Limitations

1. **Latency**: Iterative approach slower than single-pass
2. **Variance**: High cost variance due to trajectory differences
3. **Recursion depth**: Paper used depth=1; deeper recursion unexplored
4. **Synchronous calls**: Async sub-calls could reduce runtime significantly
5. **Prompt engineering**: System prompt not tuned per task

## Integration with Cooperations Orchestrator

RLMs align well with the multi-agent architecture:

```go
type RLMAgent struct {
    Model      string
    SubModel   string  // For recursive calls
    REPL       *PythonREPL
    MaxIters   int
}

func (a *RLMAgent) Execute(ctx context.Context, prompt string) (string, error) {
    // Initialize REPL with prompt as variable
    a.REPL.SetVariable("context", prompt)
    a.REPL.AddFunction("llm_query", a.subQuery)

    history := []Message{{Role: "system", Content: RLM_SYSTEM_PROMPT}}
    history = append(history, Message{
        Role: "user",
        Content: fmt.Sprintf("Context length: %d chars. Query: ...", len(prompt)),
    })

    for i := 0; i < a.MaxIters; i++ {
        // Get code from LLM
        response := a.Model.Generate(history)
        code := extractREPLCode(response)

        // Execute in REPL
        stdout, final := a.REPL.Execute(code)

        if final != "" {
            return final, nil
        }

        // Add only metadata to history
        history = append(history, Message{
            Role: "assistant",
            Content: response,
        }, Message{
            Role: "user",
            Content: fmt.Sprintf("Output (%d chars): %s...", len(stdout), truncate(stdout, 500)),
        })
    }

    return "", errors.New("max iterations exceeded")
}
```

## References

- **Original source:** Dr Milan Milanovic (Twitter)
- **Paper:** arXiv:2512.24601v2 "Recursive Language Models"
- **Institution:** MIT CSAIL
- **Benchmarks:** BrowseComp+, OOLONG, LongBench-v2 CodeQA
- **Related work:** CodeAct, Context Compaction, Sub-agent delegation
