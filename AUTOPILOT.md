# AUTOPILOT.md — Self-Driving Parity Mode

## Purpose

This document defines how OpenClaw should continue parity work in this repository **without waiting for new human prompts**.

The goal is not blind activity. The goal is to keep producing **measurable Java → Go parity gains** through repeatable small loops.

---

## Core operating rule

If no human response is available, OpenClaw should:

1. find the highest-value remaining parity gap,
2. package it into a small task,
3. run a fresh-context Codex loop,
4. evaluate the result,
5. keep/discard/split,
6. commit and push meaningful wins,
7. immediately continue to the next gap.

In other words:

> No idle waiting. If a credible next parity step exists, take it.

---

## Autopilot priorities

When choosing the next task, use this order:

### Priority 1 — Existing drafted gaps with narrow scope

Examples:

- `GAP-TEXT-TOUNICODE`
- `GAP-TEXT-WINANSI`
- `GAP-TEXT-TRIMSPACE`
- `GAP-TEXT-TJ-OFFSET`
- `GAP-RO-TWO-COLUMN-BASIC`

### Priority 2 — Follow-up gaps discovered during evaluation

Examples:

- candidate patch fixed one symptom but exposed a narrower root cause
- a regression test reveals an adjacent parity issue
- Java behavior suggests a missing cleanup or merge stage

### Priority 3 — Gap discovery from evidence

Create a new gap when one of these happens:

1. Java and Go outputs diverge on a reproducible fixture
2. benchmark prediction visibly differs from ground truth in a recurring pattern
3. Codex reports a residual mismatch that does not fit the current gap cleanly
4. a patch is rejected because it bundles multiple root causes
5. a stale milestone task is too broad and needs sub-gap decomposition

---

## Gap discovery rules

When creating a new gap, OpenClaw should prefer names that reveal the real failure mode.

Good examples:

- `GAP-TEXT-LEADING-BOUNDARY-WHITESPACE`
- `GAP-RO-COLUMN-SPLIT-DETECTION`
- `GAP-RO-COLUMN-MERGE-ORDER`
- `GAP-TEXT-TOUNICODE-LIGATURE-MAP`

Bad examples:

- `GAP-MISC-FIXES`
- `GAP-TEXT-CLEANUP`
- `GAP-T17-RETRY`

A discovered gap should be added under `plan/gaps/` when it is likely to recur or when it needs explicit tracking.

---

## Loop lifecycle

Every autonomous loop should follow this exact sequence:

### 1. Select

Choose one gap only.

### 2. Packetize

Create or update a packet under `harness/packets/`.

### 3. Execute

Run Codex on that packet with fresh context.

### 4. Evaluate

Run the required validation outside Codex if Codex sandbox restrictions prevent trustworthy verification.

### 5. Decide

Possible outcomes:

- **KEEP**
- **DISCARD**
- **SPLIT**
- **NEEDS-HUMAN-REVIEW**

### 6. Record

Write a loop report under `reports/loops/`.

### 7. Commit / Push

If KEEP and meaningful, commit and push immediately.

### 8. Continue

Move to the next best gap without waiting for a human reply.

---

## When to commit

Commit when all are true:

1. evaluator result is **KEEP**
2. targeted tests pass
3. build passes
4. the change is scoped and understandable

Do **not** wait for a giant batch. Small verified parity wins should land incrementally.

---

## When to push

Push after every meaningful KEEP commit unless there is a specific reason to batch.

Reasons to delay push:

- immediate follow-up fix is required to restore green state
- the current commit is only an intermediate checkpoint and not evaluator-clean

Default rule:

> KEEP → commit → push

---

## When to split a gap

Split instead of stretching the current loop when:

- the patch starts touching multiple subsystems
- there is more than one plausible root cause
- a test reveals a different adjacent failure mode
- acceptance criteria become ambiguous
- Codex starts exploring broadly instead of implementing narrowly

Examples:

- `GAP-TEXT-TOUNICODE` → `GAP-TEXT-TOUNICODE-LIGATURES`, `GAP-TEXT-TOUNICODE-MULTIBYTE-MAP`
- `GAP-RO-TWO-COLUMN-BASIC` → `GAP-RO-COLUMN-SPLIT-DETECTION`, `GAP-RO-COLUMN-MERGE-ORDER`

---

## When to discard

Discard a candidate when:

- tests fail
- parity signal does not improve
- the patch is heuristic and cannot be defended against Java behavior
- one fixture improves but a nearby regression appears without a strong reason

Discard means:

- do not commit it as a parity win
- extract any useful learnings
- create follow-up gaps if needed

---

## When to stop and ask the human

Autopilot should pause only when:

- external side effects are required beyond normal repo work
- Java behavior and Go behavior are both ambiguous
- multiple trade-offs are plausible and the choice is product-level, not parity-level
- a destructive or risky operation is needed
- credentials, access, or missing infrastructure block trustworthy evaluation

If none of those apply, keep going.

---

## Evaluation checklist

Every loop should answer:

1. What exact mismatch did this loop target?
2. What evidence shows the mismatch existed before?
3. What evidence shows it improved now?
4. What tests/build commands passed?
5. What residual uncertainty remains?
6. What is the next best follow-up gap?

---

## Practical validator rule

If Codex cannot reliably run validation because of sandbox or network limits:

- let Codex produce the smallest plausible patch + tests,
- then run validation from the outer OpenClaw execution environment,
- use that result for the actual KEEP / DISCARD decision.

This is the default pattern for this repo when Go module/network access differs between runtimes.

---

## Completion rules for a gap family

A family like `T16` or `T17` is only considered substantially complete when:

1. all known drafted child gaps are resolved, split, or explicitly waived
2. no high-signal benchmark/fixture evidence still points to the same family
3. evaluator reports no obvious uncovered root cause in that family

This means a milestone can move from:

- pending
- active
- mostly-covered
- validated

without pretending every edge case in the universe is solved.

---

## Default next-gap strategy for this repository

Until better evidence overrides it:

1. finish `T16` text-parity family first
2. then move into `T17` reading-order family
3. then use benchmark diffs to discover the next highest-value parity gaps

Preferred sequence:

- `GAP-TEXT-TJ-OFFSET`
- `GAP-TEXT-TRIMSPACE`
- `GAP-TEXT-WINANSI`
- `GAP-TEXT-TOUNICODE`
- `GAP-RO-TWO-COLUMN-BASIC`
- `GAP-RO-BALANCED-COLUMNS`
- `GAP-RO-SIDEBAR-MIXED-LAYOUT`
- `GAP-RO-XYCUT-DEPTH-STABILITY`

---

## One-line autopilot promise

> Keep turning Java-vs-Go evidence into small verified parity wins until no credible next gap remains or human intervention is truly required.
