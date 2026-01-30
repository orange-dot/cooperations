# Recursive Verified Reasoning (RVR)

> Hibridni pristup koji kombinuje REPL-baziranu rekurziju sa eksplicitnom verifikacijom i confidence scoring-om.

## Inspiracija

### MIT CSAIL: Recursive Language Models (RLM)
- **Paper:** arXiv:2512.24601v2
- **Kljucna ideja:** Prompt kao varijabla u REPL okruzenju, ne u LLM kontekstu
- **Mehanizam:** Model pise kod koji rekurzivno poziva sub-LLM u petljama
- **Rezultat:** Skaliranje do 10M+ tokena, O(n²) semanticki rad

### Dr Milan Milanovic (Twitter)
- **Kljucna ideja:** Eksplicitna meta-kognicija sa confidence scores
- **Mehanizam:** VERIFY korak koji proverava logiku, cinjenice, bias
- **Rezultat:** Retry mehanizam kad confidence padne ispod praga

## Zasto kombinovati?

| Aspekt | RLM | Dr Milan Milanovic | RVR (predlozeni koncept) |
|--------|-----|---------------------|--------------------------|
| Skaliranje | 10M+ tokena | Ograniceno kontekstom | 10M+ tokena (teoretski) |
| Verifikacija | Implicitna (model odlucuje) | Eksplicitna (confidence) | Eksplicitna + rekurzivna (ideja) |
| Retry | Nema | Kad confidence < 0.8 | Konfigurabilan po task tipu (ideja) |
| Efikasnost | Visoka | Srednja (dupli pozivi) | Visoka (dvoslojni pristup, teoretski) |

> **Napomena:** RLM i Dr Milan Milanovic kolone odrazavaju dokumentovane rezultate. RVR kolona predstavlja predlozene ideje koje kombinuju oba pristupa - jos uvek nije validirano.

---

## Arhitektura: Dva Sloja

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
│  Output: Buffer sa svim rezultatima i confidence scores         │
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
│  Kombinuj rezultate sa weighted confidence                       │
│  Final confidence = Σ(result.confidence × weight) / Σ(weight)   │
└─────────────────────────────────────────────────────────────────┘
```

---

## Konfiguracija Task Types

### Struktura konfiguracije

```yaml
rvr_config:
  # Globalni defaults
  defaults:
    confidence_threshold: 0.8
    critical_threshold: 0.4
    retry_attempts: 2
    batch_size: 10  # Koliko chunk-ova po sub-LLM pozivu

  # Task-specificne konfiguracije
  tasks:
    research:
      description: "Istrazivanje i sinteza informacija"
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
      description: "Generisanje i modifikacija koda"
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
      description: "Pregled i analiza postojeceg koda"
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
      description: "Strateske odluke i trade-off analiza"
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
      description: "Sazimanje dugackih tekstova"
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
      description: "Prevod izmedju jezika ili formata"
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
# Korisnik moze definisati custom task
tasks:
  my_custom_task:
    description: "Moj specificni use case"
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

| Dimension | Opis | Prompt hint | Tipicni taskovi |
|-----------|------|-------------|-----------------|
| `logic` | Da li zakljucak sledi iz premisa | "Is the reasoning logically valid?" | code, decision |
| `facts` | Da li su cinjenice tacne | "Are the stated facts accurate?" | research |
| `sources` | Da li su izvori pouzdani | "Are the sources reliable and cited?" | research |
| `completeness` | Da li su svi aspekti pokriveni | "Are all aspects of the question addressed?" | all |
| `bias` | Da li postoje skrivene pretpostavke | "Are there hidden assumptions or biases?" | decision |
| `syntax` | Da li je kod sintaksno ispravan | "Is the code syntactically correct?" | code |
| `security` | Da li postoje sigurnosni propusti | "Are there security vulnerabilities?" | code, review |
| `edge_cases` | Da li su edge cases pokriveni | "Are edge cases handled?" | code |
| `performance` | Da li ima performance problema | "Are there performance concerns?" | review |
| `maintainability` | Da li je kod odrziv | "Is the code maintainable?" | review |
| `accuracy` | Da li je rezultat tacan | "Is the result accurate?" | translation, summary |
| `fluency` | Da li je tekst tecan | "Is the text fluent and natural?" | translation |
| `alternatives` | Da li su alternative razmotrene | "Were alternatives considered?" | decision |

---

## Algoritam

### Pseudokod

```python
class RVR:
    def __init__(self, config: RVRConfig):
        self.config = config

    def process(self, prompt: str, task_type: str) -> RVRResult:
        task = self.config.tasks[task_type]

        # REPL setup (od RLM-a)
        repl = REPL()
        repl.set_variable("context", prompt)
        repl.set_variable("task_config", task)

        # LAYER 1: Fast Processing
        buffer = self.layer1_fast_process(repl, task)

        # Triage by confidence
        critical, low, high = self.triage(buffer, task)

        # LAYER 2: Selective Verification
        verified = self.layer2_verify(critical, low, high, task)

        # Synthesis
        return self.synthesize(verified, task)

    def layer1_fast_process(self, repl, task) -> List[BufferItem]:
        """Brzi prolaz sa inline confidence"""
        buffer = []

        chunks = repl.execute(f"""
            # Chunk context based on task
            chunk_size = {task.batch_size} * 1000
            chunks = [context[i:i+chunk_size]
                      for i in range(0, len(context), chunk_size)]
            chunks
        """)

        for i, chunk in enumerate(chunks):
            result = repl.llm_query(f"""
                Task: {task.description}
                Context: {chunk}

                Respond in this format:
                ANSWER: <your answer>
                CONFIDENCE: <0.0-1.0>
                UNCERTAINTY: <what you're unsure about, if any>
            """)

            parsed = self.parse_response(result)
            buffer.append(BufferItem(
                chunk_id=i,
                result=parsed.answer,
                confidence=parsed.confidence,
                uncertainty=parsed.uncertainty,
                original_chunk=chunk
            ))

        return buffer

    def triage(self, buffer, task) -> Tuple[List, List, List]:
        """Razvrstaj po confidence nivou"""
        critical = []  # < critical_threshold
        low = []       # critical_threshold <= x < confidence_threshold
        high = []      # >= confidence_threshold

        for item in buffer:
            if item.confidence < task.critical_threshold:
                critical.append(item)
            elif item.confidence < task.confidence_threshold:
                low.append(item)
            else:
                high.append(item)

        return critical, low, high

    def layer2_verify(self, critical, low, high, task) -> List[BufferItem]:
        """Selektivna verifikacija"""
        verified = list(high)  # High confidence prolazi

        # Critical items: retry sa alternativnim pristupom
        for item in critical:
            for strategy in task.retry_strategies:
                new_result = self.retry_with_strategy(item, strategy, task)
                if new_result.confidence >= task.critical_threshold:
                    item = new_result
                    break
            verified.append(item)

        # Low confidence items: verify specific dimensions
        for item in low:
            for dimension in task.verify_fields:
                verification = self.verify_dimension(item, dimension)
                item.confidence = (item.confidence + verification.confidence) / 2
                item.verifications[dimension] = verification
            verified.append(item)

        return verified

    def verify_dimension(self, item, dimension: str) -> Verification:
        """Verifikuj specifican aspekt"""
        prompt = f"""
            Original question context: {item.original_chunk[:500]}...
            Answer to verify: {item.result}

            Dimension to verify: {dimension}

            Evaluate this dimension and respond:
            VALID: <yes/no/partial>
            CONFIDENCE: <0.0-1.0>
            ISSUES: <any issues found, or "none">
        """

        result = self.llm_query(prompt)
        return Verification(
            dimension=dimension,
            valid=result.valid,
            confidence=result.confidence,
            issues=result.issues
        )

    def retry_with_strategy(self, item, strategy: str, task) -> BufferItem:
        """Retry sa specificnom strategijom"""
        strategies = {
            "rephrase_query": "Rephrase the question differently",
            "expand_context": "Consider broader context",
            "step_by_step": "Break down into smaller steps",
            "simplify": "Simplify the approach",
            "devils_advocate": "Argue the opposite position first",
            "seek_counterexamples": "Find counterexamples before concluding",
        }

        prompt = f"""
            Previous attempt had low confidence.
            Strategy: {strategies.get(strategy, strategy)}

            Original context: {item.original_chunk}
            Previous answer: {item.result}
            Previous confidence: {item.confidence}
            Uncertainty: {item.uncertainty}

            Try again with the suggested strategy.

            ANSWER: <new answer>
            CONFIDENCE: <0.0-1.0>
            UNCERTAINTY: <remaining uncertainties>
        """

        result = self.llm_query(prompt)
        return BufferItem(
            chunk_id=item.chunk_id,
            result=result.answer,
            confidence=result.confidence,
            uncertainty=result.uncertainty,
            original_chunk=item.original_chunk,
            retry_strategy=strategy
        )

    def synthesize(self, verified: List[BufferItem], task) -> RVRResult:
        """Kombinuj rezultate sa weighted confidence"""

        # Weighted average confidence
        total_weight = sum(item.confidence for item in verified)
        if total_weight == 0:
            overall_confidence = 0
        else:
            overall_confidence = sum(
                item.confidence ** 2 for item in verified
            ) / total_weight

        # Combine results
        combined = self.llm_query(f"""
            Synthesize these partial results into a final answer:

            {self._format_results(verified)}

            Weight higher-confidence results more heavily.

            FINAL_ANSWER: <synthesized answer>
            OVERALL_CONFIDENCE: <0.0-1.0>
            CAVEATS: <important caveats or limitations>
        """)

        return RVRResult(
            answer=combined.final_answer,
            confidence=overall_confidence,
            caveats=combined.caveats,
            breakdown=verified
        )
```

---

## Primeri Upotrebe

### Primer 1: Research Task

```python
rvr = RVR(config)

result = rvr.process(
    prompt=large_document_corpus,  # 5M tokena
    task_type="research"
)

# Output:
# {
#   "answer": "Based on analysis of 1,247 documents...",
#   "confidence": 0.82,
#   "caveats": ["Some sources from 2019 may be outdated"],
#   "breakdown": [
#     {"chunk_id": 0, "confidence": 0.91, ...},
#     {"chunk_id": 1, "confidence": 0.73, "verified": ["facts", "sources"]},
#     ...
#   ]
# }
```

### Primer 2: Code Generation sa visokim standardom

```python
result = rvr.process(
    prompt="Implement a thread-safe cache with LRU eviction",
    task_type="code_generation"
)

# Threshold je 0.85, pa ce vise chunk-ova ici na Layer 2
# Verify fields: logic, syntax, security, edge_cases
```

### Primer 3: Custom Task

```yaml
# config.yaml
tasks:
  legal_review:
    confidence_threshold: 0.95  # Visok standard za legalne dokumente
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

```python
result = rvr.process(
    prompt=contract_text,
    task_type="legal_review"
)
```

---

## Integracija sa Cooperations Orchestratorom

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

// internal/agents/architect.go - integracija
func (a *ArchitectAgent) Execute(ctx context.Context, handoff types.Handoff) (types.AgentResponse, error) {
    // Koristi RVR za kompleksne design odluke
    rvr := rvr.NewProcessor(a.rvrConfig, a.cli)

    result, err := rvr.Process(ctx, handoff.Payload, "decision_making")
    if err != nil {
        return types.AgentResponse{}, err
    }

    // Ukljuci confidence u response
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

## Metrike i Monitoring

```yaml
# Preporucene metrike za pracenje RVR performansi
metrics:
  - name: layer1_pass_rate
    description: "% items koji prodju Layer 1 bez verifikacije"
    target: "> 60%"  # Ako je manje, threshold mozda previsok

  - name: critical_rate
    description: "% items koji padnu u critical kategoriju"
    target: "< 10%"  # Ako je vise, mozda je task pretezan

  - name: retry_success_rate
    description: "% retry pokusaja koji podignu confidence iznad threshold"
    target: "> 50%"

  - name: avg_confidence_lift
    description: "Prosecno povecanje confidence nakon Layer 2"
    target: "> 0.15"

  - name: verification_agreement
    description: "% verifikacija gde model potvrdi svoj prvobitni odgovor"
    target: "70-90%"  # Previse nisko = los model, previse visoko = rubber stamp
```

---

## Zakljucak

RVR kombinuje najbolje od oba sveta:

1. **Od RLM-a:** Skalabilnost, REPL pristup, rekurzivni sub-pozivi
2. **Od Milana:** Eksplicitna verifikacija, confidence scores, retry logika

Kljucna prednost dvoslojnog pristupa je **efikasnost** - samo items koji zaista trebaju dodatnu paznju prolaze kroz skuplji verification proces.

---

## Reference

- **RLM:** MIT CSAIL, arXiv:2512.24601v2 "Recursive Language Models"
- **Verification approach:** Dr Milan Milanovic (Twitter)
- **Kombinovani koncept:** RVR (Recursive Verified Reasoning)
