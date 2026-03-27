# GAP-RO-SIDEBAR-MIXED-LAYOUT

## Category

reading-order-parity

## Why this gap exists

Not all multi-region pages are pure two-column layouts. Some PDFs include sidebars, callouts, captions, floating notes, or mixed block structures that can confuse column-aware ordering. A robust reading-order system needs an explicit strategy for these mixed layouts.

This is an advanced sub-gap within `T17`.

---

## Reference behavior (Java)

The Java reference implementation should provide more stable handling for mixed-layout pages, whether through XYCut++ heuristics or fallback ordering logic.

---

## Current Go behavior

Go may incorrectly classify sidebars or floating regions as a full column, causing:

- main text interruption
- side content inserted too early
- local regions read before parent flow is complete

---

## Impact

- poor human readability on magazines, reports, and papers with callouts
- reading-order benchmark degradation on mixed layouts
- brittle behavior if basic column heuristics are over-applied

---

## Candidate source files

### Go
- `go/internal/processors/xycut_plus_plus_sorter.go`
- related segmentation and ordering helpers

### Java reference
- corresponding Java layout partitioning and reading-order logic

---

## Fixtures / Reproduction

Use pages with:

- sidebars
- pull quotes
- captions close to body text
- floating blocks near column boundaries

Compare Java vs Go ordering at the points where side content enters the sequence.

---

## Acceptance criteria

1. Mixed-layout fixtures move closer to Java ordering.
2. Basic two-column cases do not regress.
3. Reading order remains deterministic across reruns.
4. Relevant tests pass and `go build ./...` succeeds.

---

## Evaluation commands

```bash
cd go && go test ./internal/processors/... -v
cd go && go build ./...
```

Use fixture comparisons against Java output where possible.

---

## Non-goals

Out of scope here:

- basic TJ / text extraction fixes
- unrelated serializer formatting issues
- broad benchmark optimization without mixed-layout evidence

---

## Keep / Discard guidance

### KEEP

- side content placement becomes more human-correct and closer to Java
- ordinary two-column pages remain stable

### DISCARD

- mixed-layout handling remains ad hoc or harms simpler layouts
