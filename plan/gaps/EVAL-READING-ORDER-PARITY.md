# EVAL-READING-ORDER-PARITY

## Category

evaluator

## Purpose

Provide the core evaluation checklist for reading-order loops in the Go port.

This evaluator is intended to judge candidate patches related to the `T17` family, including:

- basic two-column ordering
- balanced-column behavior
- mixed-layout / sidebar handling
- XYCut recursion stability

---

## Reference

The Java implementation is the primary behavioral reference.

---

## Core questions

1. Does the candidate move Go ordering closer to Java output on targeted fixtures?
2. Does it improve reading order without regressing simpler single-column cases?
3. Are deterministic ordering and recursion stability preserved?
4. Is the patch scoped to the specific gap?
5. Should the result be KEPT, DISCARDED, or split into narrower gaps?

---

## Minimum evidence

A reading-order candidate should ideally provide:

- targeted XYCut / processor test results
- one or more fixture comparisons
- note of any remaining ambiguity in layout heuristics

---

## Suggested checks

```bash
cd go && go test ./internal/processors/... -v
cd go && go build ./...
```

For larger changes, compare targeted fixture ordering against Java output and inspect any relevant benchmark subset signals such as NID.

---

## KEEP when

- targeted reading-order fixtures clearly improve
- single-column behavior does not regress
- no instability or recursion failure appears

## DISCARD when

- ordering remains materially wrong
- simpler layouts regress
- patch is brittle and overfit to one sample

## SPLIT when

- one candidate mixes column detection, merge order, and stability concerns
- different fixture families need different heuristics

## NEEDS-HUMAN-REVIEW when

- Java ordering is itself ambiguous on a layout
- heuristic trade-offs cannot be decided from current evidence
