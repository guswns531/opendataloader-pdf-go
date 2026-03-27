# LOOPS.md — OpenClaw × Codex Porting Loop

## Purpose

This repository is operated as a **parity-driven Java → Go porting system**, not as a generic feature factory.

The primary goal is to increase **behavioral parity** between the existing Java implementation and the Go implementation with measurable evidence.

In this operating model:

- **OpenClaw** is the loop orchestrator, backlog manager, and status summarizer.
- **Codex** is the single-task implementation worker.
- **Ralph-style harness** means each iteration starts with **fresh context** and a narrowly scoped task packet.
- **autoresearch-style evaluator** decides whether a result should be **kept, discarded, or split into follow-up gaps**.

---

## North Star

> Use the Java implementation as the reference behavior and systematically drive the Go implementation toward parity in output, correctness, and benchmark quality.

Parity includes:

1. **Output parity** — text, markdown, JSON, HTML, ordering, structure
2. **Behavior parity** — same handling of tricky PDF cases and edge cases
3. **Option parity** — CLI/config behavior stays aligned where intended
4. **Quality parity** — benchmark metrics do not regress and should trend upward
5. **Regression safety** — fixing one gap must not silently break another area

---

## Roles

### OpenClaw

OpenClaw is responsible for:

- selecting the next gap to work on
- preparing the task packet for Codex
- keeping loop context small and fresh
- collecting results
- invoking evaluators / running checks
- deciding keep / discard / split
- updating plan state and summarizing progress to humans

OpenClaw should behave like a **loop designer + monitor + summarizer**, not as the main hand-coder for large changes.

### Codex

Codex is responsible for:

- implementing one scoped change
- porting one behavior or one micro-unit from Java to Go
- adding or improving tests
- running local validation commands within the task scope
- reporting what changed and what remains uncertain

Codex should be given a **single scoped packet**, not the whole project strategy.

### Ralph-style harness

Ralph-style operation means:

- each loop starts fresh
- prompts include only the current gap and necessary files
- avoid large accumulated context from prior failures
- if a task is too large, split it instead of extending context indefinitely
- prefer many small verified loops over one giant run

### autoresearch-style evaluator

The evaluator is the gatekeeper.

A change is **not accepted** merely because code was written.
A change is accepted only if evidence shows parity improved or was preserved.

Evaluator outputs:

- **KEEP** — improvement accepted
- **DISCARD** — revert or ignore candidate
- **SPLIT** — candidate revealed sub-gaps that need narrower follow-up work
- **NEEDS-HUMAN-REVIEW** — evidence ambiguous or trade-off unclear

---

## Loop Rules

1. **One loop = one gap or one micro-port unit**
2. **Every loop must have measurable acceptance criteria**
3. **Implementation alone is not completion; evaluator approval is completion**
4. **Java behavior is the reference unless explicitly documented otherwise**
5. **Stale docs do not override code/test evidence**
6. **If the task grows, split it immediately**
7. **Fresh context beats long chat history**
8. **Prefer fixture and golden diff checks before full benchmark runs**

---

## Preferred Work Unit: Gap Packet

Every active task should be expressed as a gap packet with these fields:

- **Gap ID** — stable identifier (example: `GAP-TEXT-TJ-OFFSET`)
- **Category** — text / reading-order / table / heading / serializer / cli / infra
- **Reference behavior** — what Java currently does
- **Current Go behavior** — how Go differs today
- **Impact** — why this matters (user-visible quality, benchmark metric, regression risk)
- **Fixtures** — PDFs, expected outputs, failing cases
- **Relevant files** — Java files, Go files, tests, docs
- **Acceptance criteria** — specific checks required for completion
- **Evaluation commands** — exact commands used to validate keep/discard
- **Non-goals** — what this loop must not attempt to solve

---

## Canonical Loop Sequence

### 1. Select

OpenClaw selects one gap based on:

- highest user value
- benchmark weakness
- known stale task (for example T16/T17)
- unblock value for later work
- low ambiguity / high tractability

### 2. Packetize

OpenClaw creates a short task packet containing:

- the exact gap
- the relevant Java and Go files
- reproduction fixture(s)
- acceptance criteria
- required commands
- explicit non-goals

### 3. Execute

Codex receives the packet and does one thing:

- implement the missing behavior
- add tests or strengthen regression coverage
- run the defined checks

### 4. Evaluate

OpenClaw or a dedicated evaluator checks:

- tests
- fixture outputs
- golden diffs
- benchmark subset
- regressions in adjacent behavior

### 5. Decide

Possible outcomes:

- **KEEP** — merge into active branch / preserve commit
- **DISCARD** — drop the candidate
- **SPLIT** — create narrower follow-up gaps
- **ESCALATE** — human judgment required

### 6. Update State

If accepted, update:

- backlog status
- remaining gaps
- known caveats
- loop report

---

## Evaluation Ladder

Use the lightest evaluator that gives trustworthy signal.

### Level 1 — Fixture / Golden Diff

Use for most loops.

Examples:

- text output parity against Java baseline
- markdown ordering parity
- JSON element ordering or structural parity
- specific bug reproduction samples

### Level 2 — Targeted Regression Suite

Use when the gap affects a subsystem.

Examples:

- text extraction regression tests
- XYCut / reading order subset
- table processor subset
- serializer schema tests

### Level 3 — Benchmark Subset

Use before accepting larger behavior changes.

Metrics from project conventions include:

- **NID** — reading order
- **TEDS** — table structure
- **MHS** — heading structure
- **Table Detection F1**
- **Speed**

### Level 4 — Full Benchmark / Release Gate

Use sparingly:

- before milestone completion
- before broader branch integration
- before declaring parity in a major area

---

## Keep / Discard Heuristics

### KEEP when

- failing fixture now matches Java behavior
- targeted tests pass
- benchmark subset improves or holds
- no meaningful regression is introduced
- implementation remains understandable and scoped

### DISCARD when

- tests still fail
- parity did not improve
- behavior improved in one fixture but regressed elsewhere without justification
- patch is overly ad hoc and cannot be defended against the Java reference

### SPLIT when

- one task actually contains multiple separable gaps
- implementation uncertainty remains high
- distinct causes are entangled (for example encoding + spacing + layout)

---

## Fresh-Context Prompting Rules

When preparing Codex tasks:

- include only the current gap and relevant files
- include exact commands to run
- include expected output or acceptance checks
- explicitly state non-goals
- do not paste broad project philosophy into every task
- do not rely on Codex remembering prior loops

A good Codex packet should be understandable even if the agent has never seen the repo before.

---

## Branching and Commit Strategy

Recommended:

- keep a long-lived integration branch for parity work
- let each loop produce one small logical commit when it succeeds
- do not mix unrelated gaps in one commit
- document evaluator evidence in commit messages or loop reports when useful

Suggested commit style:

- `gap(text): recover spaces from TJ kerning offsets`
- `gap(reading-order): split two-column layouts before Y sort`
- `eval(text): add WinAnsi parity fixtures`

Avoid committing generated or binary artifacts by accident.

---

## Suggested OpenClaw Workflow

For each loop, OpenClaw should be able to answer:

1. What exact gap are we solving?
2. How do we know Java is correct here?
3. What evidence currently shows Go is wrong?
4. What commands will prove improvement?
5. If this works, what status changes?
6. If it fails, how do we split it smaller?

If those answers are unclear, the task is not ready for Codex yet.

---

## What Not To Do

Avoid these anti-patterns:

- giant “implement T16” prompts
- mixing multiple unrelated parity gaps in one run
- accepting changes without explicit evaluation evidence
- treating stale plan docs as source-of-truth over tests and code
- running full benchmarks for every tiny edit
- letting long iterative context replace a crisp packet

---

## Initial Priority Areas

Immediate candidates already identified in this repo:

- text extraction parity (`T16` family)
  - TJ spacing
  - TrimSpace behavior
  - WinAnsi / encoding handling
  - ToUnicode edge cases
- reading order parity (`T17` family)
  - two-column layout sequencing
  - balanced column ordering
  - sidebar / mixed-layout behavior
  - XYCut recursion stability

These should be treated as **gap families**, not monolithic tasks.

---

## Definition of Done

A loop is done only when:

- the scoped gap is clearly described
- the Go behavior improved relative to the Java reference
- the evaluator result is KEEP
- tests or fixtures preserve the gain
- backlog state is updated

Until then, the work is still exploratory.
